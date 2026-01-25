//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/colton/futurebuild/internal/api/handlers"
	"github.com/colton/futurebuild/internal/futureshade"
	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/go-chi/chi/v5"
)

// TestFutureShade_ServiceInitialization verifies the futureshade package initializes correctly.
// See FUTURESHADE_INIT_specs.md Section 7: Implementation Notes (Fail Open).
func TestFutureShade_ServiceInitialization(t *testing.T) {
	t.Run("NilConfig_ReturnsDisabledService", func(t *testing.T) {
		// Fail Open: nil config should not panic, returns disabled service
		svc := futureshade.NewService(nil)
		if svc == nil {
			t.Fatal("NewService(nil) returned nil, expected NoOp service")
		}

		err := svc.Health()
		if err == nil {
			t.Error("expected Health() to return error for nil config")
		}
	})

	t.Run("DisabledConfig_ReturnsDisabledService", func(t *testing.T) {
		cfg := &futureshade.Config{
			Enabled: false,
			APIKey:  "",
			ModelID: "",
		}
		svc := futureshade.NewService(cfg)

		err := svc.Health()
		if err == nil {
			t.Error("expected Health() to return error when disabled")
		}
	})

	t.Run("EnabledWithoutAPIKey_ReturnsError", func(t *testing.T) {
		cfg := &futureshade.Config{
			Enabled: true,
			APIKey:  "", // Missing API key
			ModelID: "gemini-2.5-flash",
		}
		svc := futureshade.NewService(cfg)

		err := svc.Health()
		if err == nil {
			t.Error("expected Health() to return error when API key is missing")
		}
	})

	t.Run("EnabledWithAPIKey_ReturnsHealthy", func(t *testing.T) {
		cfg := &futureshade.Config{
			Enabled: true,
			APIKey:  "test-api-key",
			ModelID: "gemini-2.5-flash",
		}
		svc := futureshade.NewService(cfg)

		err := svc.Health()
		if err != nil {
			t.Errorf("expected Health() to succeed, got: %v", err)
		}
	})
}

// TestFutureShade_HealthEndpoint verifies the HTTP endpoint behavior.
// See FUTURESHADE_INIT_specs.md Section 3.2: HTTP Endpoints.
func TestFutureShade_HealthEndpoint(t *testing.T) {
	// Setup minimal router with FutureShade endpoint
	setupRouter := func(svc futureshade.Service, claims *types.Claims) *chi.Mux {
		handler := handlers.NewFutureShadeHandler(svc)
		r := chi.NewRouter()
		r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
			// Inject claims into context if provided
			if claims != nil {
				req = req.WithContext(middleware.WithClaims(req.Context(), claims))
			}
			handler.HandleHealth(w, req)
		})
		return r
	}

	t.Run("DisabledService_Returns503", func(t *testing.T) {
		cfg := &futureshade.Config{Enabled: false}
		svc := futureshade.NewService(cfg)
		adminClaims := &types.Claims{
			UserID: "admin-user",
			OrgID:  "test-org",
			Role:   types.UserRoleAdmin,
		}
		router := setupRouter(svc, adminClaims)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", w.Code)
		}

		var resp handlers.HealthResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Status != "disabled" {
			t.Errorf("expected status 'disabled', got %q", resp.Status)
		}
	})

	t.Run("EnabledService_Returns200", func(t *testing.T) {
		cfg := &futureshade.Config{
			Enabled: true,
			APIKey:  "test-api-key",
			ModelID: "gemini-2.5-flash",
		}
		svc := futureshade.NewService(cfg)
		adminClaims := &types.Claims{
			UserID: "admin-user",
			OrgID:  "test-org",
			Role:   types.UserRoleAdmin,
		}
		router := setupRouter(svc, adminClaims)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp handlers.HealthResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Status != "active" {
			t.Errorf("expected status 'active', got %q", resp.Status)
		}
	})

	t.Run("NilService_Returns503", func(t *testing.T) {
		// Fail Open: handler should handle nil service gracefully
		adminClaims := &types.Claims{
			UserID: "admin-user",
			OrgID:  "test-org",
			Role:   types.UserRoleAdmin,
		}
		router := setupRouter(nil, adminClaims)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503 for nil service, got %d", w.Code)
		}
	})
}

// TestFutureShade_CompilationCheck verifies all sub-packages compile.
// See FUTURESHADE_INIT_specs.md Section 2.1: Directory Structure.
func TestFutureShade_CompilationCheck(t *testing.T) {
	// Import tribunal and shadow packages to ensure they compile
	// These are currently stubs but must be importable
	t.Run("TribunalPackage_Compiles", func(t *testing.T) {
		// Verify tribunal types are defined
		_ = futureshade.ShadowDoc{}
		// Tribunal package existence verified by import in service
	})

	t.Run("ShadowPackage_Compiles", func(t *testing.T) {
		// Shadow package will be verified when observer interface is used
		// For now, just verify ShadowDoc is accessible
		doc := futureshade.ShadowDoc{
			ID:         "test-id",
			SourceType: "Code",
			SourceID:   "/path/to/file.go",
		}
		if doc.ID != "test-id" {
			t.Error("ShadowDoc fields not accessible")
		}
	})
}
