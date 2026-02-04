package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/httputil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserHandler handles user profile endpoints.
// See LAUNCH_PLAN.md Section: User Profile Update Endpoint (P0).
type UserHandler struct {
	db          *pgxpool.Pool
	userService service.UserServicer
}

// NewUserHandler creates a new user handler.
func NewUserHandler(db *pgxpool.Pool, userService service.UserServicer) *UserHandler {
	return &UserHandler{db: db, userService: userService}
}

// UpdateProfileRequest is the request body for updating a user profile.
type UpdateProfileRequest struct {
	Name string `json:"name"`
}

// UserProfileResponse is the response for user profile operations.
type UserProfileResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	OrgID     string `json:"org_id"`
	CreatedAt string `json:"created_at"`
}

// GetProfile handles GET /api/v1/users/me.
// Returns the current authenticated user's profile.
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get claims from context
	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		slog.Warn("user: unauthorized - no claims in context", "error", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		slog.Error("user: invalid user_id in claims", "error", err)
		http.Error(w, "Invalid user", http.StatusInternalServerError)
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		slog.Error("user: invalid org_id in claims", "error", err)
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}

	// Fetch user from database
	var user models.User
	err = h.db.QueryRow(ctx, `
		SELECT id, org_id, email, name, role, created_at
		FROM users
		WHERE id = $1 AND org_id = $2
	`, userID, orgID).Scan(
		&user.ID,
		&user.OrgID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		slog.Error("user: failed to fetch user", "error", err, "user_id", userID)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(UserProfileResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		Role:      string(user.Role),
		OrgID:     user.OrgID.String(),
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// UpdateProfile handles PUT /api/v1/users/me.
// Updates the current authenticated user's profile.
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get claims from context
	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		slog.Warn("user: unauthorized - no claims in context", "error", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		slog.Error("user: invalid user_id in claims", "error", err)
		http.Error(w, "Invalid user", http.StatusInternalServerError)
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		slog.Error("user: invalid org_id in claims", "error", err)
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}

	// Parse request body (L7: limit body size)
	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("user: invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate name
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Update user in database
	var user models.User
	err = h.db.QueryRow(ctx, `
		UPDATE users
		SET name = $1
		WHERE id = $2 AND org_id = $3
		RETURNING id, org_id, email, name, role, created_at
	`, req.Name, userID, orgID).Scan(
		&user.ID,
		&user.OrgID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		slog.Error("user: failed to update user", "error", err, "user_id", userID)
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	slog.Info("user: profile updated", "user_id", userID, "name", req.Name)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(UserProfileResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		Role:      string(user.Role),
		OrgID:     user.OrgID.String(),
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// ListMembers handles GET /api/v1/org/members.
// Returns all members of the authenticated user's organization.
func (h *UserHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		slog.Warn("user: unauthorized - no claims in context", "error", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orgIdentifier := claims.OrgID
	if orgIdentifier == "" {
		// Fallback: resolve org from user's external_id in the database.
		// Clerk JWTs may omit org_id depending on the JWT template configuration.
		resolved, resolveErr := h.userService.ResolveUserOrg(ctx, claims.UserID)
		if resolveErr != nil {
			slog.Error("user: no org_id in JWT and failed to resolve from user",
				"error", resolveErr, "user_id", claims.UserID)
			http.Error(w, "No organization context", http.StatusBadRequest)
			return
		}
		orgIdentifier = resolved
	}

	// Pass raw claim org_id — service resolves UUID or Clerk external_id
	members, err := h.userService.ListOrgMembers(ctx, orgIdentifier)
	if err != nil {
		slog.Error("user: failed to list org members", "error", err, "claim_org_id", claims.OrgID)
		http.Error(w, "Failed to list members", http.StatusInternalServerError)
		return
	}

	resp := make([]UserProfileResponse, 0, len(members))
	for _, m := range members {
		resp = append(resp, UserProfileResponse{
			ID:        m.ID.String(),
			Email:     m.Email,
			Name:      m.Name,
			Role:      string(m.Role),
			OrgID:     m.OrgID.String(),
			CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
