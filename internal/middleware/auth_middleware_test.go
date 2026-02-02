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
	"github.com/colton/futurebuild/internal/auth"
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

// newTestMiddlewareWithAudience creates an AuthMiddleware with audience validation enabled.
func newTestMiddlewareWithAudience(audience string) *AuthMiddleware {
	cfg := &config.Config{
		ClerkIssuerURL: testIssuer,
		ClerkAudience:  audience,
	}
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

	t.Run("Empty sub claim rejected", func(t *testing.T) {
		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["sub"] = ""
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var body response.ErrorEnvelope
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "Unauthorized", body.Error.Message)
	})

	t.Run("Missing sub claim rejected", func(t *testing.T) {
		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			delete(*c, "sub")
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Audience validation — valid aud accepted", func(t *testing.T) {
		const testAudience = "futurebuild-app"
		mwAud := newTestMiddlewareWithAudience(testAudience)
		h := mwAud.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["aud"] = testAudience
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Audience validation — wrong aud rejected", func(t *testing.T) {
		mwAud := newTestMiddlewareWithAudience("futurebuild-app")
		h := mwAud.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["aud"] = "other-app"
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Audience validation — missing aud rejected when configured", func(t *testing.T) {
		mwAud := newTestMiddlewareWithAudience("futurebuild-app")
		h := mwAud.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		token := generateTestTokenRS256(time.Hour, nil)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Audience validation — skipped when config empty", func(t *testing.T) {
		token := generateTestTokenRS256(time.Hour, nil)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
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

// Step 81: RequirePermission tests
func TestRequirePermission(t *testing.T) {
	mw := newTestMiddleware()

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	t.Run("Admin has all permissions", func(t *testing.T) {
		handler := mw.RequireAuth(mw.RequirePermission(auth.ScopeProjectDelete)(okHandler))

		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["org_role"] = "org:admin"
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Builder can create projects", func(t *testing.T) {
		handler := mw.RequireAuth(mw.RequirePermission(auth.ScopeProjectCreate)(okHandler))

		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["org_role"] = "org:member"
		})
		req := httptest.NewRequest("POST", "/projects", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Builder cannot delete projects", func(t *testing.T) {
		handler := mw.RequireAuth(mw.RequirePermission(auth.ScopeProjectDelete)(okHandler))

		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["org_role"] = "org:member"
		})
		req := httptest.NewRequest("DELETE", "/projects/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("Viewer can read projects", func(t *testing.T) {
		handler := mw.RequireAuth(mw.RequirePermission(auth.ScopeProjectRead)(okHandler))

		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["org_role"] = "org:viewer"
		})
		req := httptest.NewRequest("GET", "/projects", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Viewer cannot write to chat", func(t *testing.T) {
		handler := mw.RequireAuth(mw.RequirePermission(auth.ScopeChatWrite)(okHandler))

		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["org_role"] = "org:viewer"
		})
		req := httptest.NewRequest("POST", "/chat", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("Unauthenticated request returns 401", func(t *testing.T) {
		handler := mw.RequireAuth(mw.RequirePermission(auth.ScopeProjectRead)(okHandler))

		req := httptest.NewRequest("GET", "/projects", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

// Step 81: Viewer role mapping test
func TestMapClerkRoleToInternal_Viewer(t *testing.T) {
	mw := newTestMiddleware()

	t.Run("org:viewer maps to Viewer", func(t *testing.T) {
		h := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, _ := GetClaims(r.Context())
			assert.Equal(t, types.UserRoleViewer, claims.Role)
			w.WriteHeader(http.StatusOK)
		}))

		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["org_role"] = "org:viewer"
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("org:guest maps to Viewer", func(t *testing.T) {
		h := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, _ := GetClaims(r.Context())
			assert.Equal(t, types.UserRoleViewer, claims.Role)
			w.WriteHeader(http.StatusOK)
		}))

		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["org_role"] = "org:guest"
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

// L7 Audit: E2E auth pipeline test — validates JWT → Role Mapping → RBAC → Route in a single chain
func TestE2E_AuthPipeline_ViewerBlockedFromWrite(t *testing.T) {
	mw := newTestMiddleware()

	writeHandler := mw.RequireAuth(
		mw.RequirePermission(auth.ScopeChatWrite)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("WRITE_OK"))
			}),
		),
	)

	readHandler := mw.RequireAuth(
		mw.RequirePermission(auth.ScopeProjectRead)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("READ_OK"))
			}),
		),
	)

	// Viewer JWT hits write endpoint → 403
	t.Run("Viewer blocked from write endpoint", func(t *testing.T) {
		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["org_role"] = "org:viewer"
		})
		req := httptest.NewRequest("POST", "/chat", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		writeHandler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	// Viewer JWT hits read endpoint → 200
	t.Run("Viewer allowed on read endpoint", func(t *testing.T) {
		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["org_role"] = "org:viewer"
		})
		req := httptest.NewRequest("GET", "/projects/123", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		readHandler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "READ_OK", rec.Body.String())
	})

	// Builder JWT hits write endpoint → 200
	t.Run("Builder allowed on write endpoint", func(t *testing.T) {
		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["org_role"] = "org:member"
		})
		req := httptest.NewRequest("POST", "/chat", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		writeHandler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "WRITE_OK", rec.Body.String())
	})

	// No JWT → 401
	t.Run("No auth returns 401", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/chat", nil)
		rec := httptest.NewRecorder()
		writeHandler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

// L7 Audit: Tenant isolation — claims from Org A cannot be used to imply access to Org B
func TestTenantIsolation_ClaimsOrgID(t *testing.T) {
	mw := newTestMiddleware()

	orgA := "org_aaaa-aaaa-aaaa-aaaa"
	orgB := "org_bbbb-bbbb-bbbb-bbbb"

	handler := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := GetClaims(r.Context())
		assert.NoError(t, err)
		// The handler receives only the org from the JWT — no cross-org access
		w.Header().Set("X-Org-ID", claims.OrgID)
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("OrgA token gets OrgA claims", func(t *testing.T) {
		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["org_id"] = orgA
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, orgA, rec.Header().Get("X-Org-ID"))
	})

	t.Run("OrgB token gets OrgB claims", func(t *testing.T) {
		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["org_id"] = orgB
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, orgB, rec.Header().Get("X-Org-ID"))
	})

	t.Run("Header-based org override is ignored", func(t *testing.T) {
		// Token has OrgA, but attacker sets X-Org-ID header to OrgB
		token := generateTestTokenRS256(time.Hour, func(c *jwt.MapClaims) {
			(*c)["org_id"] = orgA
		})
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-Org-ID", orgB) // Attacker tries to override
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		// Claims should still reflect OrgA from JWT, not the header
		assert.Equal(t, orgA, rec.Header().Get("X-Org-ID"))
	})
}
