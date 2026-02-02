package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/colton/futurebuild/internal/api/response"
	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const claimsContextKey contextKey = "claims"

type AuthMiddleware struct {
	cfg  *config.Config
	jwks jwt.Keyfunc
}

// NewAuthMiddlewareWithKeyfunc creates an AuthMiddleware with a custom JWT keyfunc.
// Used for testing when a JWKS endpoint is not available.
func NewAuthMiddlewareWithKeyfunc(cfg *config.Config, kf jwt.Keyfunc) *AuthMiddleware {
	return &AuthMiddleware{cfg: cfg, jwks: kf}
}

// NewAuthMiddleware creates a new AuthMiddleware with JWKS-based validation.
// Phase 12: Replaced HMAC (HS256) with Clerk JWKS (RS256).
// See STEP_78_AUTH_PROVIDER.md Section 2 and STEP_79_MIDDLEWARE_SWAP.md.
func NewAuthMiddleware(cfg *config.Config) *AuthMiddleware {
	m := &AuthMiddleware{cfg: cfg}

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

		// Multi-Tenancy Enforcement: OrgID is required for all authenticated requests.
		// Users without an organization should be prompted to create/join one.
		if claims.OrgID == "" {
			// Allow requests without org_id — user may not have joined an org yet.
			// Handlers that require org context should check this themselves.
			slog.Debug("auth: JWT has no org_id claim", "user_id", claims.UserID)
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
	default:
		// Default to Builder for unknown roles
		return types.UserRoleBuilder
	}
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
