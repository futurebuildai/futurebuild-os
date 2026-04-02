package oidc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
)

// Compile-time interface check.
var _ op.Storage = (*Storage)(nil)

// Storage implements the op.Storage interface using PostgreSQL (pgx).
type Storage struct {
	pool         *pgxpool.Pool
	logger       *slog.Logger
	loginURLFunc func(string) string
}

// NewStorage creates a new PostgreSQL-backed OIDC storage.
func NewStorage(pool *pgxpool.Pool, logger *slog.Logger) *Storage {
	return &Storage{
		pool:   pool,
		logger: logger,
		loginURLFunc: func(id string) string {
			return "/login?authRequestID=" + id
		},
	}
}

// ─── AuthStorage ────────────────────────────────────────────────────────────

// CreateAuthRequest persists a new authorization request.
func (s *Storage) CreateAuthRequest(ctx context.Context, authReq *oidc.AuthRequest, userID string) (op.AuthRequest, error) {
	request := authRequestFromOIDC(authReq, userID)
	request.ID_ = uuid.NewString()

	var codeChallenge, codeChallengeMethod *string
	if request.CodeChallenge_ != nil {
		c := request.CodeChallenge_.Challenge
		m := string(request.CodeChallenge_.Method)
		codeChallenge = &c
		codeChallengeMethod = &m
	}

	_, err := s.pool.Exec(ctx,
		`INSERT INTO oidc_auth_requests
			(id, client_id, redirect_uri, scopes, state, nonce, code_challenge,
			 code_challenge_method, response_type, response_mode, user_id, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NULLIF($11, ''), now() + interval '10 minutes')`,
		request.ID_,
		request.ApplicationID,
		request.CallbackURI,
		request.Scopes_,
		request.TransferState,
		request.Nonce_,
		codeChallenge,
		codeChallengeMethod,
		string(request.ResponseType_),
		stringOrNil(string(request.ResponseMode_)),
		request.UserID,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting auth request: %w", err)
	}

	return request, nil
}

// AuthRequestByID retrieves an authorization request by ID.
func (s *Storage) AuthRequestByID(ctx context.Context, id string) (op.AuthRequest, error) {
	return s.getAuthRequest(ctx, "id", id)
}

// AuthRequestByCode retrieves an authorization request by authorization code.
func (s *Storage) AuthRequestByCode(ctx context.Context, code string) (op.AuthRequest, error) {
	return s.getAuthRequest(ctx, "auth_code", code)
}

func (s *Storage) getAuthRequest(ctx context.Context, column, value string) (*AuthRequest, error) {
	query := fmt.Sprintf(
		`SELECT id, client_id, redirect_uri, scopes, state, nonce,
		        code_challenge, code_challenge_method, response_type, response_mode,
		        user_id, authenticated, auth_time
		 FROM oidc_auth_requests
		 WHERE %s = $1 AND expires_at > now()`, column)

	row := s.pool.QueryRow(ctx, query, value)

	var (
		ar                AuthRequest
		codeChallenge     *string
		codeChallengeMethod *string
		responseType      string
		responseMode      *string
		userID            *string
		authTime          *time.Time
	)

	err := row.Scan(
		&ar.ID_, &ar.ApplicationID, &ar.CallbackURI, &ar.Scopes_,
		&ar.TransferState, &ar.Nonce_, &codeChallenge, &codeChallengeMethod,
		&responseType, &responseMode, &userID, &ar.Done_, &authTime,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("auth request not found")
		}
		return nil, fmt.Errorf("querying auth request: %w", err)
	}

	ar.ResponseType_ = oidc.ResponseType(responseType)
	if responseMode != nil {
		ar.ResponseMode_ = oidc.ResponseMode(*responseMode)
	}
	if userID != nil {
		ar.UserID = *userID
	}
	if authTime != nil {
		ar.AuthTime_ = *authTime
	}
	if codeChallenge != nil {
		method := oidc.CodeChallengeMethodS256
		if codeChallengeMethod != nil && *codeChallengeMethod == "plain" {
			method = oidc.CodeChallengeMethodPlain
		}
		ar.CodeChallenge_ = &oidc.CodeChallenge{
			Challenge: *codeChallenge,
			Method:    method,
		}
	}

	return &ar, nil
}

// SaveAuthCode saves the authorization code for a request.
func (s *Storage) SaveAuthCode(ctx context.Context, id string, code string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE oidc_auth_requests SET auth_code = $1 WHERE id = $2`,
		code, id,
	)
	return err
}

// DeleteAuthRequest removes an authorization request after token exchange.
func (s *Storage) DeleteAuthRequest(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM oidc_auth_requests WHERE id = $1`, id,
	)
	return err
}

// CreateAccessToken creates a new access token.
func (s *Storage) CreateAccessToken(ctx context.Context, request op.TokenRequest) (string, time.Time, error) {
	var applicationID string
	if ar, ok := request.(*AuthRequest); ok {
		applicationID = ar.ApplicationID
	}

	tokenID := uuid.NewString()
	expiration := time.Now().Add(1 * time.Hour)

	_, err := s.pool.Exec(ctx,
		`INSERT INTO oidc_access_tokens (id, client_id, user_id, audience, scopes, expires_at)
		 VALUES ($1, $2, $3::uuid, $4, $5, $6)`,
		tokenID, applicationID, request.GetSubject(), request.GetAudience(), request.GetScopes(), expiration,
	)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("inserting access token: %w", err)
	}

	return tokenID, expiration, nil
}

// CreateAccessAndRefreshTokens creates both access and refresh tokens.
func (s *Storage) CreateAccessAndRefreshTokens(ctx context.Context, request op.TokenRequest, currentRefreshToken string) (string, string, time.Time, error) {
	applicationID, authTime, amr := getInfoFromRequest(request)

	tokenID := uuid.NewString()
	expiration := time.Now().Add(1 * time.Hour)
	refreshTokenID := uuid.NewString()
	refreshExpiration := time.Now().Add(24 * time.Hour)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// If refreshing, revoke the old refresh token
	if currentRefreshToken != "" {
		_, err = tx.Exec(ctx,
			`UPDATE oidc_refresh_tokens SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL`,
			currentRefreshToken,
		)
		if err != nil {
			return "", "", time.Time{}, fmt.Errorf("revoking old refresh token: %w", err)
		}
	}

	// Create access token
	_, err = tx.Exec(ctx,
		`INSERT INTO oidc_access_tokens (id, client_id, user_id, refresh_token_id, audience, scopes, expires_at)
		 VALUES ($1, $2, $3::uuid, $4, $5, $6, $7)`,
		tokenID, applicationID, request.GetSubject(), refreshTokenID,
		request.GetAudience(), request.GetScopes(), expiration,
	)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("inserting access token: %w", err)
	}

	// Create refresh token
	_, err = tx.Exec(ctx,
		`INSERT INTO oidc_refresh_tokens (id, token_hash, client_id, user_id, access_token_id, scopes, amr, audience, auth_time, expires_at)
		 VALUES ($1, $2, $3, $4::uuid, $5, $6, $7, $8, $9, $10)`,
		refreshTokenID, refreshTokenID, applicationID, request.GetSubject(), tokenID,
		request.GetScopes(), amr, request.GetAudience(), authTime, refreshExpiration,
	)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("inserting refresh token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", time.Time{}, fmt.Errorf("committing transaction: %w", err)
	}

	return tokenID, refreshTokenID, expiration, nil
}

// TokenRequestByRefreshToken loads a refresh token for renewal.
func (s *Storage) TokenRequestByRefreshToken(ctx context.Context, refreshToken string) (op.RefreshTokenRequest, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, token_hash, client_id, user_id, scopes, amr, audience, auth_time, expires_at, access_token_id
		 FROM oidc_refresh_tokens
		 WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()`,
		refreshToken,
	)

	var rt RefreshToken
	err := row.Scan(
		&rt.ID, &rt.Token, &rt.ApplicationID, &rt.UserID,
		&rt.Scopes, &rt.AMR, &rt.Audience, &rt.AuthTime,
		&rt.Expiration, &rt.AccessToken,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, op.ErrInvalidRefreshToken
		}
		return nil, fmt.Errorf("querying refresh token: %w", err)
	}

	return &RefreshTokenRequest{&rt}, nil
}

// TerminateSession removes tokens for a user/client session.
func (s *Storage) TerminateSession(ctx context.Context, userID string, clientID string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM oidc_access_tokens WHERE user_id = $1::uuid AND client_id = $2`, userID, clientID,
	)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx,
		`UPDATE oidc_refresh_tokens SET revoked_at = now() WHERE user_id = $1::uuid AND client_id = $2 AND revoked_at IS NULL`,
		userID, clientID,
	)
	return err
}

// RevokeToken revokes a specific token.
func (s *Storage) RevokeToken(ctx context.Context, tokenIDOrToken string, userID string, clientID string) *oidc.Error {
	// Try as access token
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM oidc_access_tokens WHERE id = $1::uuid AND client_id = $2`, tokenIDOrToken, clientID,
	)
	if err == nil && tag.RowsAffected() > 0 {
		return nil
	}

	// Try as refresh token
	_, err = s.pool.Exec(ctx,
		`UPDATE oidc_refresh_tokens SET revoked_at = now() WHERE token_hash = $1 AND client_id = $2 AND revoked_at IS NULL`,
		tokenIDOrToken, clientID,
	)
	if err != nil {
		s.logger.Error("revoking token", "error", err)
	}

	// Per RFC 7009: always return success
	return nil
}

// GetRefreshTokenInfo returns user and token ID for a refresh token.
func (s *Storage) GetRefreshTokenInfo(ctx context.Context, clientID string, token string) (string, string, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT user_id, id FROM oidc_refresh_tokens WHERE token_hash = $1 AND client_id = $2 AND revoked_at IS NULL`,
		token, clientID,
	)
	var userID, tokenID string
	if err := row.Scan(&userID, &tokenID); err != nil {
		return "", "", op.ErrInvalidRefreshToken
	}
	return userID, tokenID, nil
}

// ─── Key Management ─────────────────────────────────────────────────────────

// SigningKey returns the active RSA private key for JWT signing.
func (s *Storage) SigningKey(ctx context.Context) (op.SigningKey, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT kid, private_key_pem FROM oidc_signing_keys WHERE is_active = true ORDER BY created_at DESC LIMIT 1`,
	)

	var kid, privPEM string
	if err := row.Scan(&kid, &privPEM); err != nil {
		return nil, fmt.Errorf("no active signing key found: %w", err)
	}

	privKey, err := parseRSAPrivateKey(privPEM)
	if err != nil {
		return nil, fmt.Errorf("parsing signing key: %w", err)
	}

	return &signingKey{
		id:        kid,
		algorithm: jose.RS256,
		key:       privKey,
	}, nil
}

// SignatureAlgorithms returns supported signature algorithms.
func (s *Storage) SignatureAlgorithms(ctx context.Context) ([]jose.SignatureAlgorithm, error) {
	return []jose.SignatureAlgorithm{jose.RS256}, nil
}

// KeySet returns all active public keys for the JWKS endpoint.
func (s *Storage) KeySet(ctx context.Context) ([]op.Key, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT kid, public_key_pem FROM oidc_signing_keys WHERE is_active = true`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying signing keys: %w", err)
	}
	defer rows.Close()

	var keys []op.Key
	for rows.Next() {
		var kid, pubPEM string
		if err := rows.Scan(&kid, &pubPEM); err != nil {
			return nil, fmt.Errorf("scanning key row: %w", err)
		}
		pubKey, err := parseRSAPublicKey(pubPEM)
		if err != nil {
			s.logger.Warn("skipping malformed public key", "kid", kid, "error", err)
			continue
		}
		keys = append(keys, &publicKey{
			id:        kid,
			algorithm: jose.RS256,
			key:       pubKey,
		})
	}

	return keys, nil
}

// ─── OPStorage ──────────────────────────────────────────────────────────────

// GetClientByClientID loads an OIDC client from the database.
func (s *Storage) GetClientByClientID(ctx context.Context, clientID string) (op.Client, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT client_id, COALESCE(client_secret_hash, ''), client_name, redirect_uris,
		        grant_types, response_types, token_endpoint_auth, scopes, is_active, dev_mode
		 FROM oidc_clients
		 WHERE client_id = $1 AND is_active = true`,
		clientID,
	)

	var (
		id, secretHash, name, tokenAuth string
		redirectURIs, grantTypes, responseTypes, scopes []string
		isActive, devMode bool
	)

	if err := row.Scan(&id, &secretHash, &name, &redirectURIs, &grantTypes, &responseTypes, &tokenAuth, &scopes, &isActive, &devMode); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("client not found")
		}
		return nil, fmt.Errorf("querying client: %w", err)
	}

	return newClientFromDB(id, secretHash, name, redirectURIs, grantTypes, responseTypes, tokenAuth, scopes, isActive, devMode, s.loginURLFunc), nil
}

// AuthorizeClientIDSecret validates client credentials.
func (s *Storage) AuthorizeClientIDSecret(ctx context.Context, clientID, clientSecret string) error {
	// For public clients using PKCE, this is not called.
	// For confidential clients, compare bcrypt hash.
	row := s.pool.QueryRow(ctx,
		`SELECT client_secret_hash FROM oidc_clients WHERE client_id = $1 AND is_active = true`,
		clientID,
	)
	var hash string
	if err := row.Scan(&hash); err != nil {
		return fmt.Errorf("client not found")
	}
	if hash == "" {
		return fmt.Errorf("client has no secret configured")
	}
	// TODO: use bcrypt.CompareHashAndPassword when confidential clients are needed
	return fmt.Errorf("client secret validation not implemented yet")
}

// SetUserinfoFromScopes sets user info claims based on requested scopes.
func (s *Storage) SetUserinfoFromScopes(ctx context.Context, userinfo *oidc.UserInfo, userID, clientID string, scopes []string) error {
	return s.setUserinfo(ctx, userinfo, userID, clientID, scopes)
}

// SetUserinfoFromToken sets user info from an access token.
func (s *Storage) SetUserinfoFromToken(ctx context.Context, userinfo *oidc.UserInfo, tokenID, subject, origin string) error {
	// Verify token exists and is not expired
	row := s.pool.QueryRow(ctx,
		`SELECT client_id, scopes FROM oidc_access_tokens WHERE id = $1::uuid AND user_id = $2::uuid AND expires_at > now()`,
		tokenID, subject,
	)
	var clientID string
	var scopes []string
	if err := row.Scan(&clientID, &scopes); err != nil {
		return fmt.Errorf("token invalid or expired")
	}
	return s.setUserinfo(ctx, userinfo, subject, clientID, scopes)
}

// SetIntrospectionFromToken sets introspection response from a token.
func (s *Storage) SetIntrospectionFromToken(ctx context.Context, introspection *oidc.IntrospectionResponse, tokenID, subject, clientID string) error {
	row := s.pool.QueryRow(ctx,
		`SELECT client_id, scopes, expires_at FROM oidc_access_tokens WHERE id = $1::uuid AND user_id = $2::uuid`,
		tokenID, subject,
	)
	var appID string
	var scopes []string
	var expiresAt time.Time
	if err := row.Scan(&appID, &scopes, &expiresAt); err != nil {
		return fmt.Errorf("token not found")
	}

	introspection.Expiration = oidc.FromTime(expiresAt)
	if expiresAt.Before(time.Now()) {
		return fmt.Errorf("token expired")
	}

	userInfo := new(oidc.UserInfo)
	if err := s.setUserinfo(ctx, userInfo, subject, clientID, scopes); err != nil {
		return err
	}
	introspection.SetUserInfo(userInfo)
	introspection.Scope = scopes
	introspection.ClientID = appID
	return nil
}

// GetPrivateClaimsFromScopes returns custom JWT claims based on scopes.
// This is where we inject org_id, role, and plan_tier into access tokens.
func (s *Storage) GetPrivateClaimsFromScopes(ctx context.Context, userID, clientID string, scopes []string) (map[string]any, error) {
	claims := make(map[string]any)

	for _, scope := range scopes {
		if scope == "org" {
			// Look up org membership for this user
			row := s.pool.QueryRow(ctx,
				`SELECT om.org_id, om.role, o.plan_tier
				 FROM org_memberships om
				 JOIN organizations o ON o.id = om.org_id
				 WHERE om.user_id = $1::uuid
				 ORDER BY om.created_at ASC
				 LIMIT 1`,
				userID,
			)
			var orgID, role, planTier string
			if err := row.Scan(&orgID, &role, &planTier); err == nil {
				claims["org_id"] = orgID
				claims["role"] = role
				claims["plan_tier"] = planTier
			}
		}
	}

	return claims, nil
}

// GetKeyByIDAndClientID returns a public key for JWT profile grant validation.
func (s *Storage) GetKeyByIDAndClientID(ctx context.Context, keyID, clientID string) (*jose.JSONWebKey, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT public_key_pem FROM oidc_signing_keys WHERE kid = $1 AND is_active = true`,
		keyID,
	)
	var pubPEM string
	if err := row.Scan(&pubPEM); err != nil {
		return nil, fmt.Errorf("key not found")
	}
	pubKey, err := parseRSAPublicKey(pubPEM)
	if err != nil {
		return nil, err
	}
	return &jose.JSONWebKey{
		KeyID: keyID,
		Use:   "sig",
		Key:   pubKey,
	}, nil
}

// ValidateJWTProfileScopes validates scopes for JWT profile grants.
func (s *Storage) ValidateJWTProfileScopes(ctx context.Context, userID string, scopes []string) ([]string, error) {
	allowed := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		if scope == oidc.ScopeOpenID {
			allowed = append(allowed, scope)
		}
	}
	return allowed, nil
}

// ─── Health ─────────────────────────────────────────────────────────────────

// Health checks database connectivity.
func (s *Storage) Health(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// ─── Authentication Helper ──────────────────────────────────────────────────

// CheckUsernamePassword authenticates a user and marks the auth request as done.
// For Sprint 0, this uses a simple email lookup (magic link will be added later).
func (s *Storage) CheckUsernamePassword(username, password, id string) error {
	ctx := context.Background()

	// Look up user by email
	row := s.pool.QueryRow(ctx,
		`SELECT id FROM hub_users WHERE email = $1`, username,
	)
	var userID string
	if err := row.Scan(&userID); err != nil {
		return fmt.Errorf("user not found")
	}

	// Mark auth request as authenticated
	_, err := s.pool.Exec(ctx,
		`UPDATE oidc_auth_requests SET user_id = $1::uuid, authenticated = true, auth_time = now() WHERE id = $2`,
		userID, id,
	)
	if err != nil {
		return fmt.Errorf("updating auth request: %w", err)
	}

	return nil
}

// ─── Private Helpers ────────────────────────────────────────────────────────

// setUserinfo populates UserInfo based on user data and scopes.
func (s *Storage) setUserinfo(ctx context.Context, userInfo *oidc.UserInfo, userID, clientID string, scopes []string) error {
	row := s.pool.QueryRow(ctx,
		`SELECT id, email, display_name FROM hub_users WHERE id = $1::uuid`, userID,
	)
	var id, email, displayName string
	if err := row.Scan(&id, &email, &displayName); err != nil {
		return fmt.Errorf("user not found")
	}

	for _, scope := range scopes {
		switch scope {
		case oidc.ScopeOpenID:
			userInfo.Subject = id
		case oidc.ScopeEmail:
			userInfo.Email = email
			userInfo.EmailVerified = oidc.Bool(true)
		case oidc.ScopeProfile:
			userInfo.Name = displayName
			userInfo.PreferredUsername = email
		case "org":
			// Append org claims via UserInfo AppendClaims
			orgRow := s.pool.QueryRow(ctx,
				`SELECT om.org_id, om.role, o.plan_tier
				 FROM org_memberships om
				 JOIN organizations o ON o.id = om.org_id
				 WHERE om.user_id = $1::uuid
				 ORDER BY om.created_at ASC LIMIT 1`,
				userID,
			)
			var orgID, role, planTier string
			if err := orgRow.Scan(&orgID, &role, &planTier); err == nil {
				userInfo.AppendClaims("org_id", orgID)
				userInfo.AppendClaims("role", role)
				userInfo.AppendClaims("plan_tier", planTier)
			}
		}
	}

	return nil
}

// getInfoFromRequest extracts clientID, authTime, and AMR from a TokenRequest.
func getInfoFromRequest(req op.TokenRequest) (string, time.Time, []string) {
	if ar, ok := req.(*AuthRequest); ok {
		return ar.ApplicationID, ar.AuthTime_, ar.GetAMR()
	}
	if rr, ok := req.(*RefreshTokenRequest); ok {
		return rr.ApplicationID, rr.AuthTime, rr.AMR
	}
	return "", time.Time{}, nil
}

func stringOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// EnsureSigningKey creates a default signing key if none exists.
// Called at startup to ensure the JWKS endpoint has at least one key.
func (s *Storage) EnsureSigningKey(ctx context.Context) error {
	var count int
	err := s.pool.QueryRow(ctx,
		`SELECT count(*) FROM oidc_signing_keys WHERE is_active = true`,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("checking signing keys: %w", err)
	}

	if count > 0 {
		s.logger.Info("signing key exists", "active_keys", count)
		return nil
	}

	// Generate a new RSA key pair for development
	s.logger.Info("generating new RSA signing key for OIDC")
	privatePEM, publicPEM, err := GenerateRSAKeyPair()
	if err != nil {
		return fmt.Errorf("generating key pair: %w", err)
	}

	kid := "brain-key-2026-04"
	_, err = s.pool.Exec(ctx,
		`INSERT INTO oidc_signing_keys (kid, algorithm, private_key_pem, public_key_pem, is_active)
		 VALUES ($1, 'RS256', $2, $3, true)`,
		kid, privatePEM, publicPEM,
	)
	if err != nil {
		return fmt.Errorf("inserting signing key: %w", err)
	}

	s.logger.Info("signing key created", "kid", kid)
	return nil
}

// Compile-time check: Storage satisfies the authenticate interface for login handler.
var _ interface {
	CheckUsernamePassword(username, password, id string) error
} = (*Storage)(nil)
