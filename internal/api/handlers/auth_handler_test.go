package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/pkg/types"
)

func TestAuthHandler_Me(t *testing.T) {
	handler := NewAuthHandler()

	t.Run("returns principal from JWT claims", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)

		// Inject mock claims into context
		claims := &types.Claims{
			UserID:      "user_abc123",
			OrgID:       "org_xyz789",
			Role:        types.UserRoleAdmin,
			SubjectType: types.SubjectTypeUser,
			Email:       "admin@futurebuild.ai",
			Name:        "Test Admin",
		}
		req = req.WithContext(middleware.WithClaims(req.Context(), claims))

		rr := httptest.NewRecorder()
		handler.Me(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}

		var principal types.Principal
		if err := json.Unmarshal(rr.Body.Bytes(), &principal); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if principal.ID != "user_abc123" {
			t.Errorf("expected ID user_abc123, got %s", principal.ID)
		}
		if principal.OrgID != "org_xyz789" {
			t.Errorf("expected OrgID org_xyz789, got %s", principal.OrgID)
		}
		if principal.Role != types.UserRoleAdmin {
			t.Errorf("expected role Admin, got %s", string(principal.Role))
		}
		if principal.SubjectType != types.SubjectTypeUser {
			t.Errorf("expected subject_type user, got %s", string(principal.SubjectType))
		}
		if principal.Email != "admin@futurebuild.ai" {
			t.Errorf("expected email admin@futurebuild.ai, got %s", principal.Email)
		}
		if principal.Name != "Test Admin" {
			t.Errorf("expected name Test Admin, got %s", principal.Name)
		}
	})

	t.Run("returns 401 without claims", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
		rr := httptest.NewRecorder()
		handler.Me(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rr.Code)
		}
	})
}
