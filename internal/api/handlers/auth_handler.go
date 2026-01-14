package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/types"
)

type AuthHandler struct {
	authService         *service.AuthService
	notificationService types.NotificationService
	baseURL             string
}

func NewAuthHandler(authService *service.AuthService, notificationService types.NotificationService, baseURL string) *AuthHandler {
	return &AuthHandler{
		authService:         authService,
		notificationService: notificationService,
		baseURL:             baseURL,
	}
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req types.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Lookup: Resolve identity across USERS (internal) and CONTACTS (external)
	identity, err := h.authService.LookupIdentityByEmail(r.Context(), req.Email)

	// Security Note: If the identity does not exist, return 200 OK to prevent email enumeration attacks.
	if err != nil {
		h.respondGenericSuccess(w)
		return
	}

	// Generate: Create a raw token and its hash.
	rawToken, err := h.authService.GenerateToken()
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}
	tokenHash := h.authService.HashToken(rawToken)

	// Persist: Transactionally save the hash based on identity type.
	if identity.IsInternal() {
		if err := h.authService.StoreToken(r.Context(), identity.GetID(), tokenHash); err != nil {
			http.Error(w, "failed to store token", http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.authService.StorePortalToken(r.Context(), identity.GetID(), tokenHash); err != nil {
			http.Error(w, "failed to store portal token", http.StatusInternalServerError)
			return
		}
	}

	// Notify: Call NotificationService with the magic link.
	magicLink := h.authService.ConstructLink(h.baseURL, rawToken)
	subject := "Your FutureBuild Magic Link"
	body := fmt.Sprintf("Click here to login: %s", magicLink)

	// For Contacts, we might want to prioritize SMS if it's their preference,
	// but for now, we'll stick to email for the magic link delivery consistency.
	if err := h.notificationService.SendEmail(identity.GetEmail(), subject, body); err != nil {
		http.Error(w, "failed to send email", http.StatusInternalServerError)
		return
	}

	h.respondGenericSuccess(w)
}

// Verify handles GET /api/v1/auth/verify?token={token}
func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	identity, err := h.authService.VerifyToken(r.Context(), token)
	if err != nil {
		// Log error internally if needed, but return generic unauthorized
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Output: Generate JWT and return TokenResponse
	tokenResp, err := h.authService.GenerateJWT(identity)
	if err != nil {
		http.Error(w, "failed to generate authentication token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tokenResp)
}

func (h *AuthHandler) respondGenericSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(types.AuthResponse{
		Message: "If this user exists, a login link has been sent.",
	})
}
