package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
)

// OnboardingHandler handles onboarding-related HTTP requests.
type OnboardingHandler struct {
	interrogator *service.InterrogatorService
}

// NewOnboardingHandler creates a new onboarding handler.
func NewOnboardingHandler(svc *service.InterrogatorService) *OnboardingHandler {
	return &OnboardingHandler{interrogator: svc}
}

// HandleOnboard processes the /api/v1/agent/onboard endpoint.
// L7: Input validation, structured errors, observability.
// @route POST /api/v1/agent/onboard
func (h *OnboardingHandler) HandleOnboard(w http.ResponseWriter, r *http.Request) {
	// L7: Limit request body size (prevent DoS)
	r.Body = http.MaxBytesReader(w, r.Body, 1*1024*1024) // 1MB max

	var req models.OnboardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// L7: User-friendly error, no stack traces
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// L7: Input validation
	if err := validateOnboardRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get user context from JWT (middleware should have set this)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "Missing user_id in context", http.StatusUnauthorized)
		return
	}

	tenantID, ok := r.Context().Value("tenant_id").(string)
	if !ok {
		http.Error(w, "Missing tenant_id in context", http.StatusUnauthorized)
		return
	}

	// L7: Set timeout for AI operations (Gemini can be slow)
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	resp, err := h.interrogator.ProcessMessage(ctx, userID, tenantID, &req)
	if err != nil {
		// L7: Log error for debugging
		slog.Error("onboarding_request_failed",
			"error", err.Error(),
			"user_id", userID,
			"tenant_id", tenantID,
			"session_id", req.SessionID,
		)
		// Don't expose internal errors to user
		http.Error(w, "Failed to process request. Please try again.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// L7: Response encoding failed (rare, log it)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// validateOnboardRequest validates input fields.
// L7: Prevents abuse and ensures data integrity.
func validateOnboardRequest(req *models.OnboardRequest) error {
	// Session ID required
	if req.SessionID == "" {
		return fmt.Errorf("session_id is required")
	}

	// Session ID length limit (prevent memory abuse)
	if len(req.SessionID) > 100 {
		return fmt.Errorf("session_id too long (max 100 characters)")
	}

	// Message length limit (prevent token exhaustion)
	if len(req.Message) > 10000 {
		return fmt.Errorf("message too long (max 10,000 characters)")
	}

	// Document URL validation (if provided)
	if req.DocumentURL != "" {
		if len(req.DocumentURL) > 2000 {
			return fmt.Errorf("document_url too long")
		}
		// Basic URL format check
		if _, err := url.Parse(req.DocumentURL); err != nil {
			return fmt.Errorf("invalid document_url format")
		}
	}

	// Current state must be provided (can be empty map)
	if req.CurrentState == nil {
		return fmt.Errorf("current_state is required (use {} for empty state)")
	}

	// Current state size limit (prevent memory abuse)
	if len(req.CurrentState) > 50 {
		return fmt.Errorf("current_state has too many fields (max 50)")
	}

	return nil
}
