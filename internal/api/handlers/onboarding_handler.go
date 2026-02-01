package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/colton/futurebuild/internal/middleware"
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
// Accepts both JSON and multipart/form-data (Step 77: Magic Upload Trigger).
// L7: Input validation, structured errors, observability.
// @route POST /api/v1/agent/onboard
func (h *OnboardingHandler) HandleOnboard(w http.ResponseWriter, r *http.Request) {
	var req models.OnboardRequest
	var err error

	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Step 77: Multipart file upload path
		// L7: 50MB limit for blueprint uploads
		r.Body = http.MaxBytesReader(w, r.Body, 50*1024*1024)
		req, err = parseMultipartOnboard(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// Existing JSON path
		// L7: 1MB limit for text-only requests
		r.Body = http.MaxBytesReader(w, r.Body, 1*1024*1024)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}
	}

	// L7: Input validation
	if err := validateOnboardRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// C1 Fix: Use typed context key via middleware.GetClaims (matches all other handlers)
	claims, err := middleware.GetClaims(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID
	tenantID := claims.OrgID

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

// parseMultipartOnboard extracts OnboardRequest fields from multipart form data.
// Step 77: Reads the uploaded file and form fields into the request model.
func parseMultipartOnboard(r *http.Request) (models.OnboardRequest, error) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		return models.OnboardRequest{}, fmt.Errorf("failed to parse form: invalid or too large")
	}
	// C4 Fix: Clean up multipart temp files when done
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	file, header, err := r.FormFile("file")
	if err != nil {
		return models.OnboardRequest{}, fmt.Errorf("file field is required for multipart uploads")
	}
	defer file.Close()

	// Validate MIME type
	mimeType := header.Header.Get("Content-Type")
	if !isValidBlueprintMIME(mimeType) {
		return models.OnboardRequest{}, fmt.Errorf("invalid file type: %s (PDF, PNG, JPG only)", mimeType)
	}

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		return models.OnboardRequest{}, fmt.Errorf("failed to read uploaded file")
	}

	req := models.OnboardRequest{
		SessionID:           r.FormValue("session_id"),
		Message:             r.FormValue("message"),
		CurrentState:        parseCurrentState(r.FormValue("current_state")),
		DocumentData:        fileData,
		DocumentContentType: mimeType,
		DocumentFileName:    header.Filename,
	}

	return req, nil
}

// parseCurrentState parses a JSON string into a map, returning empty map on failure.
func parseCurrentState(jsonStr string) map[string]interface{} {
	if jsonStr == "" {
		return make(map[string]interface{})
	}
	var state map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &state); err != nil {
		return make(map[string]interface{})
	}
	return state
}

// isValidBlueprintMIME validates file MIME types for blueprint uploads.
// C6 Fix: Uses exact match after stripping parameters to prevent bypass
// (e.g., "image/jpeg; evil=true" was previously accepted via HasPrefix).
func isValidBlueprintMIME(mimeType string) bool {
	// Strip parameters: "image/jpeg; charset=utf-8" → "image/jpeg"
	normalized := strings.TrimSpace(strings.SplitN(mimeType, ";", 2)[0])
	allowed := map[string]bool{
		"image/jpeg":      true,
		"image/jpg":       true,
		"image/png":       true,
		"image/webp":      true,
		"application/pdf": true,
	}
	return allowed[normalized]
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

	// Mutual exclusion: document_url and inline file data
	if req.DocumentURL != "" && len(req.DocumentData) > 0 {
		return fmt.Errorf("provide either document_url or file upload, not both")
	}

	// Document URL validation (if provided)
	// C7 Fix: Strict validation - require http/https scheme and valid host.
	// url.Parse accepts javascript:, data:, and empty-scheme URIs which are unsafe.
	if req.DocumentURL != "" {
		if len(req.DocumentURL) > 2000 {
			return fmt.Errorf("document_url too long")
		}
		parsedURL, err := url.Parse(req.DocumentURL)
		if err != nil {
			return fmt.Errorf("invalid document_url format")
		}
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return fmt.Errorf("document_url must use http or https scheme")
		}
		if parsedURL.Host == "" {
			return fmt.Errorf("document_url must include a host")
		}
	}

	// Inline file validation (if provided)
	if len(req.DocumentData) > 0 {
		const maxFileSize = 50 * 1024 * 1024 // 50MB
		if len(req.DocumentData) > maxFileSize {
			return fmt.Errorf("file too large (max 50MB)")
		}
		if req.DocumentContentType == "" {
			return fmt.Errorf("file content type is required")
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
