package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/api/response"
	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testRSAKey is a shared RSA key pair for all tests in this file.
var testRSAKey *rsa.PrivateKey

func init() {
	var err error
	testRSAKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("failed to generate test RSA key: " + err.Error())
	}
}

// testKeyfunc returns the public key for RS256 validation in tests.
func testKeyfunc(_ *jwt.Token) (interface{}, error) {
	return &testRSAKey.PublicKey, nil
}

const testIssuer = "https://test.clerk.accounts.dev"

// generateTestTokenRS256 creates an RS256-signed JWT for testing.
func generateTestTokenRS256(expiry time.Duration, modifyClaims func(*jwt.MapClaims)) string {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub":      "user-123",
		"org_id":   "org-456",
		"org_role": "org:admin",
		"email":    "test@example.com",
		"name":     "Test User",
		"iss":      testIssuer,
		"exp":      jwt.NewNumericDate(now.Add(expiry)),
		"iat":      jwt.NewNumericDate(now),
		"nbf":      jwt.NewNumericDate(now),
	}

	if modifyClaims != nil {
		modifyClaims(&claims)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	ss, _ := token.SignedString(testRSAKey)
	return ss
}

// newTestMiddleware creates an AuthMiddleware with the test RSA keyfunc.
func newTestMiddleware() *AuthMiddleware {
	cfg := &config.Config{ClerkIssuerURL: testIssuer}
	return NewAuthMiddlewareWithKeyfunc(cfg, testKeyfunc)
}

func TestRequireAuth(t *testing.T) {
	mw := newTestMiddleware()
	handler := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := GetClaims(r.Context())
		if err != nil {
			t.Errorf("Claims not found in context")
			return
		}
		if claims.OrgID != "org-456" {
			t.Errorf("Expected OrgID org-456, got %s", claims.OrgID)
		}
		if claims.Email != "test@example.com" {
			t.Errorf("Expected Email test@example.com, got %s", claims.Email)
		}
		if claims.Name != "Test User" {
			t.Errorf("Expected Name Test User, got %s", claims.Name)
		}
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("OK"))
		require.NoError(t, err)
	}))

	t.Run("Valid Token + OrgID", func(t *testing.T) {
		token := generateTestTokenRS256(time.Hour, nil)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "OK", rec.Body.String())
	})

	t.Run("Valid Token without OrgID (allowed with debug log)", func(t *testing.T) {
		// Phase 12: Missing OrgID is allowed — user may not have joined an org yet.
		// Handlers that require org context should check this themselves.
		// Use a dedicated handler that doesn't assert OrgID.
		noOrgHandler := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := GetClaims(r.Context())
			if err != nil {
				t.Errorf("Claims not found in context")
				return
			}
			assert.Empty(t, claims.OrgID, "OrgID should be empty when not in token")
			w.WriteHeader(http.StatusOK)
		}))

		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			delete(*c, "org_id")
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		noOrgHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("No Header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var body response.ErrorEnvelope
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "Unauthorized", body.Error.Message)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var body response.ErrorEnvelope
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "Unauthorized", body.Error.Message)
	})

	t.Run("Expired Token", func(t *testing.T) {
		token := generateTestTokenRS256(-time.Hour, nil)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var body response.ErrorEnvelope
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "Unauthorized", body.Error.Message)
	})

	t.Run("Wrong Issuer", func(t *testing.T) {
		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["iss"] = "https://wrong-issuer.example.com"
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Malformed Bearer prefix", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Basic abc123")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Role mapping — org:admin maps to Admin", func(t *testing.T) {
		mwLocal := newTestMiddleware()
		h := mwLocal.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, _ := GetClaims(r.Context())
			assert.Equal(t, types.UserRoleAdmin, claims.Role)
			w.WriteHeader(http.StatusOK)
		}))

		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["org_role"] = "org:admin"
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Role mapping — org:member maps to Builder", func(t *testing.T) {
		mwLocal := newTestMiddleware()
		h := mwLocal.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, _ := GetClaims(r.Context())
			assert.Equal(t, types.UserRoleBuilder, claims.Role)
			w.WriteHeader(http.StatusOK)
		}))

		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["org_role"] = "org:member"
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestRequireRole(t *testing.T) {
	mw := newTestMiddleware()

	// Admin-only handler
	handler := mw.RequireAuth(mw.RequireRole(types.UserRoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Access Granted"))
		require.NoError(t, err)
	})))

	t.Run("Admin User Accessing Admin Route -> Allow", func(t *testing.T) {
		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["org_role"] = "org:admin"
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Access Granted", rec.Body.String())
	})

	t.Run("Member User Accessing Admin Route -> Forbidden", func(t *testing.T) {
		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["org_role"] = "org:member"
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		var body response.ErrorEnvelope
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "Forbidden", body.Error.Message)
	})

	t.Run("Builder User Accessing Builder Route -> Allow", func(t *testing.T) {
		builderHandler := mw.RequireAuth(mw.RequireRole(types.UserRoleBuilder)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("Builder Access"))
			require.NoError(t, err)
		})))

		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["org_role"] = "org:member" // member maps to Builder
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		builderHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Builder Access", rec.Body.String())
	})
}
