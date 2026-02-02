package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/httputil"
	"github.com/colton/futurebuild/pkg/types"
)

// PortalAuthHandler handles authentication for portal contacts.
// Portal contacts use magic-link email auth (separate from Clerk).
// See LAUNCH_PLAN.md P2: Field Portal (Mobile).
type PortalAuthHandler struct {
	authService         *service.AuthService
	notificationService types.NotificationService
	baseURL             string
}

// NewPortalAuthHandler creates a new PortalAuthHandler.
func NewPortalAuthHandler(authService *service.AuthService, notificationService types.NotificationService, baseURL string) *PortalAuthHandler {
	return &PortalAuthHandler{
		authService:         authService,
		notificationService: notificationService,
		baseURL:             baseURL,
	}
}

// Login handles POST /api/v1/portal/auth/login
// Sends a magic link to a contact's email address.
func (h *PortalAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var req types.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(types.AuthResponse{Message: "Invalid request body"})
		return
	}

	if req.Email == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(types.AuthResponse{Message: "Email is required"})
		return
	}

	// Always return success to prevent email enumeration attacks.
	successMsg := types.AuthResponse{Message: "If that email is registered, a login link has been sent."}

	// Lookup contact by email
	contact, err := h.authService.LookupContactByEmail(r.Context(), req.Email)
	if err != nil {
		slog.Debug("portal auth: contact not found", "email", req.Email)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(successMsg)
		return
	}

	// Generate and store token
	rawToken, err := h.authService.GenerateToken()
	if err != nil {
		slog.Error("portal auth: failed to generate token", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(types.AuthResponse{Message: "Internal server error"})
		return
	}

	tokenHash := h.authService.HashToken(rawToken)
	if err := h.authService.StorePortalToken(r.Context(), contact.ID, tokenHash); err != nil {
		slog.Error("portal auth: failed to store token", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(types.AuthResponse{Message: "Internal server error"})
		return
	}

	// Construct portal magic link and send
	link := h.authService.ConstructPortalLink(h.baseURL, rawToken)

	if contact.Email != nil && *contact.Email != "" {
		body := fmt.Sprintf(
			"Click this link to log in to the FutureBuild portal:\n\n%s\n\nThis link expires in 15 minutes.",
			link,
		)
		if err := h.notificationService.SendEmail(*contact.Email, "FutureBuild Portal Login", body); err != nil {
			slog.Error("portal auth: failed to send email", "error", err, "contact_id", contact.ID)
			// Don't reveal the failure to prevent information leakage
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(successMsg)
}

// Verify handles GET /api/v1/portal/auth/verify
// Validates a magic link token and returns a JWT for the portal session.
func (h *PortalAuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(types.AuthResponse{Message: "Token is required"})
		return
	}

	identity, err := h.authService.VerifyToken(r.Context(), token)
	if err != nil {
		slog.Debug("portal auth: token verification failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(types.AuthResponse{Message: "Invalid or expired token"})
		return
	}

	tokenResponse, err := h.authService.GenerateJWT(identity)
	if err != nil {
		slog.Error("portal auth: failed to generate JWT", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(types.AuthResponse{Message: "Internal server error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tokenResponse)
}
