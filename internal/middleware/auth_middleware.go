package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/colton/futurebuild/internal/api/response"
	"github.com/colton/futurebuild/internal/auth"
	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const claimsContextKey contextKey = "claims"

type AuthMiddleware struct {
	cfg  *config.Config
	jwks jwt.Keyfunc
	db   *pgxpool.Pool
}

// NewAuthMiddlewareWithKeyfunc creates an AuthMiddleware with a custom JWT keyfunc.
// Used for testing when a JWKS endpoint is not available.
func NewAuthMiddlewareWithKeyfunc(cfg *config.Config, kf jwt.Keyfunc) *AuthMiddleware {
	return &AuthMiddleware{cfg: cfg, jwks: kf, db: nil}
}

// NewAuthMiddleware creates a new AuthMiddleware with JWKS-based validation.
// Phase 12: Replaced HMAC (HS256) with Clerk JWKS (RS256).
// See STEP_78_AUTH_PROVIDER.md Section 2 and STEP_79_MIDDLEWARE_SWAP.md.
func NewAuthMiddleware(cfg *config.Config, db *pgxpool.Pool) *AuthMiddleware {
	m := &AuthMiddleware{cfg: cfg, db: db}

	// Initialize JWKS keyfunc from Clerk's well-known endpoint.
	// keyfunc handles caching and background refresh.
	jwksURL := cfg.ClerkIssuerURL + "/.well-known/jwks.json"
	k, err := keyfunc.NewDefaultCtx(context.Background(), []string{jwksURL})
	if err != nil {
		slog.Error("auth: failed to initialize JWKS keyfunc", "error", err, "url", jwksURL)
		// Fallback: return middleware that will reject all requests
		m.jwks = func(_ *jwt.Token) (interface{}, error) {
			return nil, fmt.Errorf("JWKS not initialized")
		}
		return m
	}

	m.jwks = k.Keyfunc
	return m
}

// RequireAuth validates a Clerk-issued JWT from the Authorization header.
// Extracts claims and populates types.Claims in the request context.
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.JSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			response.JSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Parse JWT with JWKS validation (RS256).
		// Clerk JWTs use RSA signing, verified against the JWKS endpoint.
		// See STEP_79_MIDDLEWARE_SWAP.md Section 1.2.
		parserOpts := []jwt.ParserOption{
			jwt.WithValidMethods([]string{"RS256"}),
			jwt.WithIssuer(m.cfg.ClerkIssuerURL),
		}
		if m.cfg.ClerkAudience != "" {
			parserOpts = append(parserOpts, jwt.WithAudience(m.cfg.ClerkAudience))
		}

		token, err := jwt.Parse(tokenString, m.jwks, parserOpts...)
		if err != nil {
			slog.Debug("auth: JWT validation failed", "error", err)
			response.JSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		if !token.Valid {
			response.JSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		mapClaims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.JSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Map Clerk JWT claims to internal types.Claims struct.
		// Clerk JWTs contain: sub (user ID), org_id, org_role, email, name
		// Custom claims come from the JWT template configured in Clerk Dashboard.
		claims := &types.Claims{
			UserID:      getStringClaim(mapClaims, "sub"),
			OrgID:       getStringClaim(mapClaims, "org_id"),
			Role:        mapClerkRoleToInternal(getStringClaim(mapClaims, "org_role")),
			Email:       getStringClaim(mapClaims, "email"),
			Name:        getStringClaim(mapClaims, "name"),
			SubjectType: types.SubjectTypeUser,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: getStringClaim(mapClaims, "sub"),
			},
		}

		// Reject tokens with empty subject claim.
		// Downstream handlers assume UserID is populated.
		if claims.UserID == "" {
			slog.Warn("auth: JWT has empty sub claim — rejecting")
			response.JSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Enrich sparse JWT claims from the database.
		// Clerk session tokens may omit org_id, org_role, email, and name
		// depending on JWT template configuration.
		if m.db != nil && claims.UserID != "" {
			m.enrichClaimsFromDB(r.Context(), claims)
		}

		if claims.OrgID == "" {
			slog.Debug("auth: no org_id after enrichment", "user_id", claims.UserID)
		}

		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) RequireRole(allowedRoles ...types.UserRole) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := GetClaims(r.Context())
			if err != nil {
				response.JSONError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			for _, role := range allowedRoles {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			response.JSONError(w, http.StatusForbidden, "Forbidden")
		})
	}
}

// RequirePermission returns middleware that checks if the authenticated user's role
// has the given scope permission. Uses the RBAC matrix from internal/auth.
// See STEP_81_ROLE_MAPPING.md Section 3.1.
func (m *AuthMiddleware) RequirePermission(scope auth.Scope) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := GetClaims(r.Context())
			if err != nil {
				response.JSONError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			if !auth.Can(claims.Role, scope) {
				slog.Debug("auth: permission denied",
					"user_id", claims.UserID,
					"role", claims.Role,
					"scope", scope,
				)
				response.JSONError(w, http.StatusForbidden, "Forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetClaims(ctx context.Context) (*types.Claims, error) {
	claims, ok := ctx.Value(claimsContextKey).(*types.Claims)
	if !ok {
		return nil, fmt.Errorf("claims not found in context")
	}
	return claims, nil
}

// WithClaims returns a new context with the given claims attached.
// This is primarily used for testing to inject mock claims.
func WithClaims(ctx context.Context, claims *types.Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

// mapClerkRoleToInternal converts a Clerk organization role to an internal UserRole.
// Clerk roles: "org:admin", "org:member" (or just "admin", "member" depending on template)
// Internal roles: Admin, Builder, Client, Subcontractor
func mapClerkRoleToInternal(clerkRole string) types.UserRole {
	// Normalize: strip "org:" prefix if present
	role := strings.TrimPrefix(clerkRole, "org:")

	switch role {
	case "admin":
		return types.UserRoleAdmin
	case "member", "basic_member":
		return types.UserRoleBuilder
	case "pm":
		return types.UserRolePM
	case "viewer", "guest":
		return types.UserRoleViewer
	default:
		// L7: Default to Viewer (least privilege) for unknown roles
		slog.Warn("auth: unknown Clerk role, defaulting to Viewer", "clerk_role", clerkRole)
		return types.UserRoleViewer
	}
}

// enrichClaimsFromDB fills in missing JWT claims by looking up the user in the database.
// The JWT sub claim contains the Clerk user ID, which maps to users.external_id.
// Also replaces UserID with the internal UUID so downstream handlers can use uuid.Parse.
func (m *AuthMiddleware) enrichClaimsFromDB(ctx context.Context, claims *types.Claims) {
	slog.Info("auth: enrichClaimsFromDB called",
		"clerk_sub", claims.UserID,
		"jwt_org_id", claims.OrgID,
		"jwt_role", claims.Role,
		"jwt_email", claims.Email,
		"jwt_name", claims.Name)

	var userID, orgID, email, name, role string
	err := m.db.QueryRow(ctx,
		`SELECT u.id, u.org_id, u.email, u.name, u.role
		 FROM users u
		 WHERE u.external_id = $1`,
		claims.UserID,
	).Scan(&userID, &orgID, &email, &name, &role)
	if err != nil {
		slog.Warn("auth: enrichClaimsFromDB — user NOT FOUND in DB",
			"external_id", claims.UserID, "error", err)
		return
	}

	slog.Info("auth: enrichClaimsFromDB — user found",
		"external_id", claims.UserID,
		"db_user_id", userID, "db_org_id", orgID,
		"db_email", email, "db_name", name, "db_role", role)

	// Replace Clerk external_id with internal UUID for downstream handlers.
	if userID != "" {
		claims.UserID = userID
		claims.RegisteredClaims.Subject = userID
	}
	if claims.OrgID == "" && orgID != "" {
		claims.OrgID = orgID
	}
	// DB role is source of truth — always override JWT-derived role when DB has a value.
	// This prevents Clerk's generic org:member → Builder from overriding a DB-stored PM role.
	if role != "" {
		claims.Role = types.UserRole(role)
	}
	if claims.Email == "" && email != "" {
		claims.Email = email
	}
	if claims.Name == "" && name != "" {
		claims.Name = name
	}

	slog.Info("auth: enrichClaimsFromDB — final claims",
		"user_id", claims.UserID, "org_id", claims.OrgID,
		"role", claims.Role, "email", claims.Email, "name", claims.Name)
}

// getStringClaim safely extracts a string claim from JWT MapClaims.
func getStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// ---- Portal JWT Middleware ----
// Portal contacts authenticate via HS256 JWTs issued by AuthService (magic-link flow).
// This is separate from the Clerk RS256 middleware used by internal users.

// portalContextKey is the context key type for portal-specific values.
type portalContextKey string

const (
	portalContactIDKey   portalContextKey = "portal_contact_id"
	portalContactNameKey portalContextKey = "portal_contact_name"
)

// RequirePortalAuth validates a portal JWT (HS256, issuer "futurebuild") from the
// Authorization header. Extracts contact_id from claims and sets it in context.
// Only allows tokens with subject_type = "contact".
func (m *AuthMiddleware) RequirePortalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.JSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			response.JSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Portal JWTs use HS256 with the app's JWT secret (not Clerk JWKS).
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(m.cfg.JWTSecret), nil
		},
			jwt.WithValidMethods([]string{"HS256"}),
			jwt.WithIssuer("futurebuild"),
		)
		if err != nil {
			slog.Debug("portal auth middleware: JWT validation failed", "error", err)
			response.JSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		if !token.Valid {
			response.JSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		mapClaims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.JSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Verify this is a contact token, not a user token
		subjectType := getStringClaim(mapClaims, "subject_type")
		if subjectType != string(types.SubjectTypeContact) {
			slog.Warn("portal auth middleware: non-contact token used on portal endpoint",
				"subject_type", subjectType)
			response.JSONError(w, http.StatusForbidden, "Forbidden")
			return
		}

		contactIDStr := getStringClaim(mapClaims, "user_id")
		if contactIDStr == "" {
			contactIDStr = getStringClaim(mapClaims, "sub")
		}
		if contactIDStr == "" {
			response.JSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		contactID, err := uuid.Parse(contactIDStr)
		if err != nil {
			slog.Warn("portal auth middleware: invalid contact UUID", "value", contactIDStr)
			response.JSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		contactName := getStringClaim(mapClaims, "name")

		// Also check RBAC: portal endpoints require one of the portal roles
		role := types.UserRole(getStringClaim(mapClaims, "role"))
		if role != types.UserRoleClient && role != types.UserRoleSubcontractor {
			slog.Warn("portal auth middleware: non-portal role on portal endpoint",
				"contact_id", contactID, "role", role)
			response.JSONError(w, http.StatusForbidden, "Forbidden")
			return
		}

		ctx := context.WithValue(r.Context(), portalContactIDKey, contactID)
		ctx = context.WithValue(ctx, portalContactNameKey, contactName)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetPortalContactID extracts the portal contact ID from the request context.
func GetPortalContactID(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(portalContactIDKey).(uuid.UUID)
	return v, ok
}
