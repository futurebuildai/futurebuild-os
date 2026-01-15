package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/colton/futurebuild/internal/api/response"
	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const claimsContextKey contextKey = "claims"

type AuthMiddleware struct {
	cfg *config.Config
}

func NewAuthMiddleware(cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{cfg: cfg}
}

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

		token, err := jwt.ParseWithClaims(tokenString, &types.Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.cfg.JWTSecret), nil
		})

		if err != nil {
			response.JSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		if claims, ok := token.Claims.(*types.Claims); ok && token.Valid {
			// CRITICAL: Multi-Tenancy Enforcement
			if claims.OrgID == "" {
				response.JSONError(w, http.StatusUnauthorized, "Unauthorized: Missing OrgID")
				return
			}

			ctx := context.WithValue(r.Context(), claimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			response.JSONError(w, http.StatusUnauthorized, "Unauthorized")
		}
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
