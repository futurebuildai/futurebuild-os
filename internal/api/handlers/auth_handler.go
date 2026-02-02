package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/pkg/types"
)

// AuthHandler handles authentication-related endpoints.
// Phase 12: Login and Verify removed — Clerk handles sign-in.
// Only the /auth/me endpoint remains to return the principal from JWT claims.
type AuthHandler struct{}

// NewAuthHandler creates a new AuthHandler.
// Phase 12: Dependencies (authService, notificationService) removed — Clerk manages auth.
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// Me handles GET /api/v1/auth/me
// Returns the authenticated principal extracted from the Clerk JWT claims.
// This endpoint validates that the Clerk JWT is accepted by the backend
// and provides the frontend with the mapped internal identity.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	principal := types.Principal{
		ID:          claims.UserID,
		OrgID:       claims.OrgID,
		Email:       claims.Email,
		Name:        claims.Name,
		Role:        claims.Role,
		SubjectType: claims.SubjectType,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(principal)
}
