package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthService struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewAuthService(db *pgxpool.Pool, cfg *config.Config) *AuthService {
	return &AuthService{
		db:  db,
		cfg: cfg,
	}
}

// GenerateToken creates a cryptographically secure 32-byte random string.
func (s *AuthService) GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashToken returns the SHA-256 hash of a plaintext token.
func (s *AuthService) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// StorePortalToken transactionally saves the token hash for a CONTACT.
func (s *AuthService) StorePortalToken(ctx context.Context, contactID uuid.UUID, tokenHash string) error {
	expiresAt := time.Now().UTC().Add(15 * time.Minute)
	query := `
		INSERT INTO portal_tokens (token_hash, contact_id, expires_at, used)
		VALUES ($1, $2, $3, false)
	`
	_, err := s.db.Exec(ctx, query, tokenHash, contactID, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to store portal token: %w", err)
	}
	return nil
}

// VerifyToken validates a token, marks it as used, and returns the Identity.
// OPTIMIZATION: Uses single UNION query to check both users and contacts.
// See PRODUCTION_PLAN.md Task 4 (N+1 Remediation).
func (s *AuthService) VerifyToken(ctx context.Context, plaintextToken string) (models.Identity, error) {
	tokenHash := s.HashToken(plaintextToken)

	// Single UNION query to search both tables in one round-trip
	// Note: role columns cast to text for UNION compatibility (user_role_type vs contact_role_type)
	query := `
		SELECT 'user' as identity_type, u.id, u.org_id, u.email, u.name, u.role::text, u.created_at,
		       t.expires_at, t.used, 'auth_tokens' as token_table
		FROM users u
		JOIN auth_tokens t ON u.id = t.user_id
		WHERE t.token_hash = $1
		UNION ALL
		SELECT 'contact', c.id, c.org_id, c.email, c.name, c.role::text, c.created_at,
		       t.expires_at, t.used, 'portal_tokens'
		FROM contacts c
		JOIN portal_tokens t ON c.id = t.contact_id
		WHERE t.token_hash = $1
		LIMIT 1
	`

	var identityType, tokenTable string
	var id uuid.UUID
	var orgID uuid.UUID
	var email, name, role string
	var createdAt, expiresAt time.Time
	var used bool

	err := s.db.QueryRow(ctx, query, tokenHash).Scan(
		&identityType, &id, &orgID, &email, &name, &role, &createdAt,
		&expiresAt, &used, &tokenTable,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token")
	}

	// Validate token status
	if used {
		return nil, fmt.Errorf("token already used")
	}
	if time.Now().After(expiresAt) {
		return nil, fmt.Errorf("token expired")
	}

	// Mark token as used
	// L7 Hardening: Explicit table selection avoids static analyzer flags (gosec G201)
	// See PRODUCTION_PLAN.md Step 61.1
	var updateQuery string
	switch tokenTable {
	case "auth_tokens":
		updateQuery = "UPDATE auth_tokens SET used = true WHERE token_hash = $1"
	case "portal_tokens":
		updateQuery = "UPDATE portal_tokens SET used = true WHERE token_hash = $1"
	default:
		return nil, fmt.Errorf("invalid token table: %s", tokenTable)
	}
	_, err = s.db.Exec(ctx, updateQuery, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to invalidate token: %w", err)
	}

	// Return appropriate identity type
	switch identityType {
	case "user":
		return &models.User{
			ID:        id,
			OrgID:     orgID,
			Email:     email,
			Name:      name,
			Role:      models.UserRole(role),
			CreatedAt: createdAt,
		}, nil
	case "contact":
		emailPtr := &email
		return &models.Contact{
			ID:        id,
			OrgID:     orgID,
			Email:     emailPtr,
			Name:      name,
			Role:      models.UserRole(role),
			CreatedAt: createdAt,
		}, nil
	default:
		return nil, fmt.Errorf("unknown identity type: %s", identityType)
	}
}

// LookupUserByEmail finds a user by their email address.
func (s *AuthService) LookupUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, org_id, email, name, role FROM users WHERE email = $1`
	err := s.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.OrgID, &user.Email, &user.Name, &user.Role)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// LookupContactByEmail finds a contact by their email address.
func (s *AuthService) LookupContactByEmail(ctx context.Context, email string) (*models.Contact, error) {
	var contact models.Contact
	query := `SELECT id, org_id, name, email, role, created_at FROM contacts WHERE email = $1`
	err := s.db.QueryRow(ctx, query, email).Scan(
		&contact.ID, &contact.OrgID, &contact.Name, &contact.Email, &contact.Role, &contact.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

// LookupIdentityByEmail attempts to find an identity in Users, then Contacts.
func (s *AuthService) LookupIdentityByEmail(ctx context.Context, email string) (models.Identity, error) {
	user, err := s.LookupUserByEmail(ctx, email)
	if err == nil {
		return user, nil
	}

	contact, err := s.LookupContactByEmail(ctx, email)
	if err == nil {
		return contact, nil
	}

	return nil, fmt.Errorf("identity not found")
}

// ConstructPortalLink formats the magic link URL for portal contacts.
// Points to frontend /portal/verify route.
func (s *AuthService) ConstructPortalLink(baseURL string, rawToken string) string {
	return fmt.Sprintf("%s/portal/verify?token=%s", baseURL, rawToken)
}

// GenerateJWT creates a signed token response for an identity.
func (s *AuthService) GenerateJWT(identity models.Identity) (*types.TokenResponse, error) {
	now := time.Now().UTC()
	expiry := now.Add(s.cfg.JWTExpiry)

	subType := types.SubjectTypeUser
	if !identity.IsInternal() {
		subType = types.SubjectTypeContact
	}

	claims := types.Claims{
		UserID:      identity.GetID().String(),
		OrgID:       identity.GetOrgID().String(),
		Role:        types.UserRole(identity.GetRole()),
		SubjectType: subType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   identity.GetID().String(),
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "futurebuild",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return &types.TokenResponse{
		AccessToken: ss,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.cfg.JWTExpiry.Seconds()),
		Principal: &types.Principal{
			ID:          identity.GetID().String(),
			OrgID:       identity.GetOrgID().String(),
			Email:       identity.GetEmail(),
			Name:        identity.GetName(),
			Role:        types.UserRole(identity.GetRole()),
			SubjectType: subType,
			CreatedAt:   identity.GetCreatedAt().Format(time.RFC3339),
		},
	}, nil
}
