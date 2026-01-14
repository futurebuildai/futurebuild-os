package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to generate a token with specific modifications
func generateTestToken(secret string, expiry time.Duration, modifyClaims func(*types.Claims)) string {
	now := time.Now().UTC()
	claims := types.Claims{
		UserID: "user-123",
		OrgID:  "org-456",
		Role:   types.UserRoleAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "futurebuild",
		},
	}

	if modifyClaims != nil {
		modifyClaims(&claims)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, _ := token.SignedString([]byte(secret))
	return ss
}

func TestRequireAuth(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	mw := NewAuthMiddleware(cfg)
	handler := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify context injection
		claims, err := GetClaims(r.Context())
		if err != nil {
			t.Errorf("Claims not found in context")
			return
		}
		if claims.OrgID != "org-456" {
			t.Errorf("Expected OrgID org-456, got %s", claims.OrgID)
		}
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("OK"))
		require.NoError(t, err)
	}))

	t.Run("Valid Token + OrgID", func(t *testing.T) {
		token := generateTestToken(cfg.JWTSecret, time.Hour, nil)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "OK", rec.Body.String())
	})

	t.Run("Valid Token - OrgID (Multi-Tenancy Guard)", func(t *testing.T) {
		token := generateTestToken(cfg.JWTSecret, time.Hour, func(c *types.Claims) {
			c.OrgID = "" // REMOVE OrgID to trigger failure
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var body map[string]string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "Unauthorized: Missing OrgID", body["error"])
	})

	t.Run("No Header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var body map[string]string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "Unauthorized", body["error"])
	})

	t.Run("Invalid Token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var body map[string]string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "Unauthorized", body["error"])
	})

	t.Run("Expired Token", func(t *testing.T) {
		token := generateTestToken(cfg.JWTSecret, -time.Hour, nil) // Expired 1 hour ago
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var body map[string]string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "Unauthorized", body["error"])
	})
}

func TestRequireRole(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	mw := NewAuthMiddleware(cfg)

	// Admin-only handler
	handler := mw.RequireAuth(mw.RequireRole(types.UserRoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Access Granted"))
		require.NoError(t, err)
	})))

	t.Run("Admin User Accessing Admin Route -> Allow", func(t *testing.T) {
		token := generateTestToken(cfg.JWTSecret, time.Hour, func(c *types.Claims) {
			c.Role = types.UserRoleAdmin
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Access Granted", rec.Body.String())
	})

	t.Run("Client User Accessing Admin Route -> Forbidden", func(t *testing.T) {
		token := generateTestToken(cfg.JWTSecret, time.Hour, func(c *types.Claims) {
			c.Role = types.UserRoleClient
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		var body map[string]string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "Forbidden", body["error"])
	})

	t.Run("Builder User Accessing Builder Route -> Allow", func(t *testing.T) {
		// Create a separate handler for Builder
		builderHandler := mw.RequireAuth(mw.RequireRole(types.UserRoleBuilder)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("Builder Access"))
			require.NoError(t, err)
		})))

		token := generateTestToken(cfg.JWTSecret, time.Hour, func(c *types.Claims) {
			c.Role = types.UserRoleBuilder
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		builderHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Builder Access", rec.Body.String())
	})
}
