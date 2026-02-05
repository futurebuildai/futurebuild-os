package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/httputil"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// InviteHandler handles user invitation endpoints.
// See LAUNCH_STRATEGY.md Task B2: User Invite Flow.
type InviteHandler struct {
	inviteService       *service.InviteService
	notificationService types.NotificationService
	baseURL             string
}

// NewInviteHandler creates a new invite handler.
func NewInviteHandler(
	inviteService *service.InviteService,
	notificationService types.NotificationService,
	baseURL string,
) *InviteHandler {
	return &InviteHandler{
		inviteService:       inviteService,
		notificationService: notificationService,
		baseURL:             baseURL,
	}
}

// CreateInviteRequest is the request body for creating an invitation.
type CreateInviteRequest struct {
	Email string         `json:"email"`
	Role  types.UserRole `json:"role"`
}

// InviteResponse is the response for invite operations.
type InviteResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

// CreateInvite handles POST /api/v1/admin/invites.
// Creates an invitation and sends the invite email.
func (h *InviteHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get claims from context
	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		slog.Warn("invite: unauthorized - no claims in context", "error", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body (L7: limit body size)
	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var req CreateInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("invite: invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate email
	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Parse org ID and user ID from claims
	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		slog.Error("invite: invalid org_id in claims", "error", err)
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}

	createdBy, err := uuid.Parse(claims.UserID)
	if err != nil {
		slog.Error("invite: invalid user_id in claims", "error", err)
		http.Error(w, "Invalid user", http.StatusInternalServerError)
		return
	}

	// Default role to builder if not specified
	role := req.Role
	if role == "" {
		role = types.UserRoleBuilder
	}

	// Create invitation
	rawToken, inv, err := h.inviteService.CreateInvitation(ctx, service.CreateInvitationInput{
		OrgID:     orgID,
		Email:     req.Email,
		Role:      role,
		CreatedBy: createdBy,
	})
	if err != nil {
		slog.Warn("invite: failed to create invitation", "error", err, "email", req.Email)
		http.Error(w, "Failed to create invitation", http.StatusBadRequest)
		return
	}

	// Send invitation email
	inviteLink := fmt.Sprintf("%s/invite/accept?token=%s", h.baseURL, rawToken)
	subject := "You're Invited to FutureBuild"
	body := fmt.Sprintf(
		"You've been invited to join a construction project on FutureBuild.\n\n"+
			"Click here to accept: %s\n\n"+
			"This invitation expires in 24 hours.",
		inviteLink,
	)

	if err := h.notificationService.SendEmail(req.Email, subject, body); err != nil {
		slog.Error("invite: failed to send invitation email", "error", err, "email", req.Email)
		// Don't fail - invitation was created, admin can resend
		slog.Warn("invite: invitation created but email failed - admin should resend")
	}

	slog.Info("invite: invitation created and sent", "email", req.Email, "role", role, "created_by", createdBy)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(InviteResponse{
		ID:        inv.ID.String(),
		Email:     inv.Email,
		Role:      string(inv.Role),
		ExpiresAt: inv.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt: inv.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// ListInvites handles GET /api/v1/admin/invites.
// Returns all pending invitations for the organization.
func (h *InviteHandler) ListInvites(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get claims from context
	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		slog.Warn("invite: unauthorized - no claims in context", "error", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		slog.Error("invite: invalid org_id in claims", "error", err)
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}

	invitations, err := h.inviteService.ListInvitations(ctx, orgID)
	if err != nil {
		slog.Error("invite: failed to list invitations", "error", err)
		http.Error(w, "Failed to list invitations", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	responses := make([]InviteResponse, len(invitations))
	for i, inv := range invitations {
		responses[i] = InviteResponse{
			ID:        inv.ID.String(),
			Email:     inv.Email,
			Role:      string(inv.Role),
			ExpiresAt: inv.ExpiresAt.Format("2006-01-02T15:04:05Z"),
			CreatedAt: inv.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(responses)
}

// RevokeInvite handles DELETE /api/v1/admin/invites/{id}.
// Revokes a pending invitation.
func (h *InviteHandler) RevokeInvite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get invite ID from URL
	inviteIDStr := chi.URLParam(r, "id")
	inviteID, err := uuid.Parse(inviteIDStr)
	if err != nil {
		http.Error(w, "Invalid invitation ID", http.StatusBadRequest)
		return
	}

	// Get claims from context
	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		slog.Warn("invite: unauthorized - no claims in context", "error", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		slog.Error("invite: invalid org_id in claims", "error", err)
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}

	if err := h.inviteService.RevokeInvitation(ctx, inviteID, orgID); err != nil {
		slog.Warn("invite: failed to revoke invitation", "error", err, "invite_id", inviteID)
		http.Error(w, "Failed to revoke invitation", http.StatusBadRequest)
		return
	}

	slog.Info("invite: invitation revoked", "invite_id", inviteID)
	w.WriteHeader(http.StatusNoContent)
}

// AcceptInviteRequest is the request body for accepting an invitation.
type AcceptInviteRequest struct {
	Token    string `json:"token"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// AcceptInvite handles POST /api/v1/invites/accept.
// Accepts an invitation and creates the user account.
func (h *InviteHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var req AcceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("invite: invalid accept request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	user, err := h.inviteService.AcceptInvitation(ctx, req.Token, req.Name, req.Password)
	if err != nil {
		slog.Warn("invite: failed to accept invitation", "error", err)
		http.Error(w, "Invalid or expired invitation", http.StatusBadRequest)
		return
	}

	slog.Info("invite: invitation accepted", "user_id", user.ID, "email", user.Email)

	// Return user info (frontend can then initiate a login flow)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Account created successfully. Please log in.",
		"email":   user.Email,
	})
}

// GetInviteInfo handles GET /api/v1/invites/info?token=xxx.
// Returns invitation details for the profile setup page.
func (h *InviteHandler) GetInviteInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	inv, err := h.inviteService.GetInvitationByToken(ctx, token)
	if err != nil {
		slog.Warn("invite: failed to get invitation info", "error", err)
		http.Error(w, "Invalid or expired invitation", http.StatusBadRequest)
		return
	}

	// Check if expired
	if inv.AcceptedAt != nil {
		http.Error(w, "Invitation already used", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"email":      inv.Email,
		"role":       string(inv.Role),
		"expires_at": inv.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	})
}
