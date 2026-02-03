package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/httputil"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// CompletionHandler handles project completion lifecycle endpoints.
type CompletionHandler struct {
	completionService service.CompletionServicer
	notificationSvc   service.NotificationServicer
	directorySvc      service.DirectoryServicer
}

// NewCompletionHandler creates a new CompletionHandler.
func NewCompletionHandler(cs service.CompletionServicer, ns service.NotificationServicer, ds service.DirectoryServicer) *CompletionHandler {
	return &CompletionHandler{
		completionService: cs,
		notificationSvc:   ns,
		directorySvc:      ds,
	}
}

// CompleteProjectRequest is the request body for POST /projects/{id}/complete.
type CompleteProjectRequest struct {
	Notes string `json:"notes"`
}

// CompleteProject handles POST /api/v1/projects/{id}/complete.
// Transitions a project to Completed status and generates a CompletionReport.
func (h *CompletionHandler) CompleteProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	// Parse optional request body
	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var req CompleteProjectRequest
	// Body is optional — ignore decode errors for empty bodies
	_ = json.NewDecoder(r.Body).Decode(&req)

	// L7: Cap notes length to prevent abuse (10KB is generous for completion notes)
	const maxNotesLen = 10_000
	if len(req.Notes) > maxNotesLen {
		http.Error(w, "Notes must be under 10,000 characters", http.StatusBadRequest)
		return
	}

	report, err := h.completionService.CompleteProject(ctx, projectID, orgID, userID, req.Notes)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, types.ErrConflict) {
			http.Error(w, "Project is not in Active status", http.StatusConflict)
			return
		}
		slog.Error("completion: failed to complete project", "error", err, "project_id", projectID)
		http.Error(w, "Failed to complete project", http.StatusInternalServerError)
		return
	}

	// Async: Notify project manager via email (non-blocking)
	go h.notifyProjectManager(projectID, orgID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(report)
}

// GetCompletionReport handles GET /api/v1/projects/{id}/completion-report.
func (h *CompletionHandler) GetCompletionReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	orgID, err := getAuthOrgID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	report, err := h.completionService.GetCompletionReport(ctx, projectID, orgID)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			http.Error(w, "Completion report not found", http.StatusNotFound)
			return
		}
		slog.Error("completion: failed to get report", "error", err, "project_id", projectID)
		http.Error(w, "Failed to get completion report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(report)
}

// notifyProjectManager sends an email notification to the project manager.
// Runs asynchronously — logs on failure but does not block the response.
func (h *CompletionHandler) notifyProjectManager(projectID, orgID uuid.UUID) {
	if h.notificationSvc == nil || h.directorySvc == nil {
		return
	}

	// L7: Bound async goroutine with timeout to prevent leak
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pm, err := h.directorySvc.GetProjectManager(ctx, projectID, orgID)
	if err != nil {
		slog.Warn("completion: could not find project manager for notification",
			"project_id", projectID, "error", err)
		return
	}

	if pm.Email == "" {
		return
	}

	subject := "Project Completed"
	body := "Your project has been marked as completed. A completion report has been generated and is available in the system."

	if err := h.notificationSvc.SendEmail(pm.Email, subject, body); err != nil {
		slog.Error("completion: failed to send PM notification email",
			"project_id", projectID, "email", pm.Email, "error", err)
	}
}
