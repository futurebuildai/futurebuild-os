package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/httputil"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// PortalHandler handles portal API endpoints.
// See LAUNCH_PLAN.md P2: Field Portal (Mobile).
type PortalHandler struct {
	portalService *service.PortalService
}

// NewPortalHandler creates a new PortalHandler.
func NewPortalHandler(portalService *service.PortalService) *PortalHandler {
	return &PortalHandler{
		portalService: portalService,
	}
}

// CreateActionLinkRequest represents a request to create an action link.
type CreateActionLinkRequest struct {
	ContactID  string `json:"contact_id"`
	ProjectID  string `json:"project_id"`
	TaskID     string `json:"task_id"`
	ActionType string `json:"action_type"` // status_update, photo_upload, view
}

// CreateActionLinkResponse represents the response for creating an action link.
type CreateActionLinkResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// HandleCreateActionLink creates and sends a one-time action link via SMS.
// POST /api/v1/admin/portal/link
func (h *PortalHandler) HandleCreateActionLink(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var req CreateActionLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("portal: invalid request body", "error", err)
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Parse UUIDs
	contactID, err := uuid.Parse(req.ContactID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid contact_id")
		return
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project_id")
		return
	}

	taskID, err := uuid.Parse(req.TaskID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task_id")
		return
	}

	// Validate action type
	var actionType service.ActionType
	switch req.ActionType {
	case "status_update":
		actionType = service.ActionTypeStatusUpdate
	case "photo_upload":
		actionType = service.ActionTypePhotoUpload
	case "view":
		actionType = service.ActionTypeView
	default:
		respondError(w, http.StatusBadRequest, "invalid action_type")
		return
	}

	// Send action link
	err = h.portalService.SendActionLink(r.Context(), contactID, projectID, taskID, actionType)
	if err != nil {
		slog.Error("portal: failed to send action link", "error", err)
		respondError(w, http.StatusInternalServerError, "Failed to send action link")
		return
	}

	slog.Info("portal: action link sent", "contact_id", contactID, "task_id", taskID, "action_type", actionType)
	respondJSON(w, http.StatusOK, CreateActionLinkResponse{
		Success: true,
		Message: "Action link sent successfully",
	})
}

// ActionContextResponse represents the context returned when verifying an action token.
type ActionContextResponse struct {
	ActionType string          `json:"action_type"`
	Contact    *ContactSummary `json:"contact"`
	Project    *ProjectSummary `json:"project"`
	Task       *TaskSummary    `json:"task"`
}

// ContactSummary is a minimal contact representation for portal UI.
type ContactSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ProjectSummary is a minimal project representation for portal UI.
type ProjectSummary struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

// TaskSummary is a minimal task representation for portal UI.
type TaskSummary struct {
	ID        string `json:"id"`
	WBSCode   string `json:"wbs_code"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

// HandleVerifyActionToken verifies an action token and returns the context.
// GET /api/v1/portal/action/:token
func (h *PortalHandler) HandleVerifyActionToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "missing token")
		return
	}

	ctx, err := h.portalService.VerifyActionToken(r.Context(), token)
	if err != nil {
		slog.Warn("portal: token verification failed", "error", err)
		respondError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	// Build response
	response := ActionContextResponse{
		ActionType: string(ctx.Token.ActionType),
		Contact: &ContactSummary{
			ID:   ctx.Contact.ID.String(),
			Name: ctx.Contact.Name,
		},
		Project: &ProjectSummary{
			ID:      ctx.Project.ID.String(),
			Name:    ctx.Project.Name,
			Address: ctx.Project.Address,
		},
		Task: &TaskSummary{
			ID:      ctx.Task.ID.String(),
			WBSCode: ctx.Task.WBSCode,
			Name:    ctx.Task.Name,
			Status:  string(ctx.Task.Status),
		},
	}

	if ctx.Task.PlannedStart != nil {
		response.Task.StartDate = ctx.Task.PlannedStart.Format("2006-01-02")
	}
	if ctx.Task.PlannedEnd != nil {
		response.Task.EndDate = ctx.Task.PlannedEnd.Format("2006-01-02")
	}

	respondJSON(w, http.StatusOK, response)
}

// SubmitActionRequest represents a request to submit an action.
type SubmitActionRequest struct {
	Status  *string `json:"status,omitempty"`   // For status updates
	PhotoID *string `json:"photo_id,omitempty"` // For photo uploads
}

// HandleSubmitAction submits the action for a token and marks it as used.
// POST /api/v1/portal/action/:token
func (h *PortalHandler) HandleSubmitAction(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "missing token")
		return
	}

	// Verify token first
	ctx, err := h.portalService.VerifyActionToken(r.Context(), token)
	if err != nil {
		slog.Warn("portal: token verification failed", "error", err)
		respondError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	// Parse request (L7: limit body size)
	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var req SubmitActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Handle action based on type
	switch ctx.Token.ActionType {
	case service.ActionTypeStatusUpdate:
		if req.Status == nil {
			respondError(w, http.StatusBadRequest, "status is required")
			return
		}
		// Validate status
		status, err := parseTaskStatus(*req.Status)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid task status")
			return
		}
		// Update task status
		if err := h.portalService.UpdateTaskStatus(r.Context(), ctx.Token.TaskID, ctx.Token.ProjectID, status); err != nil {
			slog.Error("portal: failed to update task status", "error", err)
			respondError(w, http.StatusInternalServerError, "Failed to update task status")
			return
		}
		slog.Info("portal: task status updated", "task_id", ctx.Token.TaskID, "status", status)

	case service.ActionTypePhotoUpload:
		if req.PhotoID == nil {
			respondError(w, http.StatusBadRequest, "photo_id is required")
			return
		}
		// Photo upload is handled separately via S3; just validate the ID exists
		// TODO: Verify photo exists and link to task
		slog.Info("portal: photo upload action", "task_id", ctx.Token.TaskID, "photo_id", *req.PhotoID)

	case service.ActionTypeView:
		// View-only action, nothing to submit
		slog.Info("portal: view action", "task_id", ctx.Token.TaskID)
	}

	// Mark token as used
	if err := h.portalService.UseActionToken(r.Context(), ctx.Token.ID); err != nil {
		slog.Error("portal: failed to mark token as used", "error", err)
		respondError(w, http.StatusInternalServerError, "Failed to complete action")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"success": "true", "message": "Action submitted successfully"})
}

// parseTaskStatus converts a string to TaskStatus.
func parseTaskStatus(s string) (types.TaskStatus, error) {
	switch s {
	case "pending":
		return types.TaskStatusPending, nil
	case "ready":
		return types.TaskStatusReady, nil
	case "in_progress":
		return types.TaskStatusInProgress, nil
	case "inspection_pending":
		return types.TaskStatusInspectionPending, nil
	case "completed", "complete":
		return types.TaskStatusCompleted, nil
	case "blocked":
		return types.TaskStatusBlocked, nil
	case "delayed":
		return types.TaskStatusDelayed, nil
	default:
		return "", fmt.Errorf("invalid status: %s", s)
	}
}

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("portal: failed to encode response", "error", err)
	}
}

// respondError writes an error JSON response.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// HandleUploadVoiceMemo accepts an audio file upload and metadata.
// POST /api/v1/portal/voice-memos
// Phase 18: See FRONTEND_SCOPE.md §15.2 (Voice-First Field Portal)
func (h *PortalHandler) HandleUploadVoiceMemo(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 25 << 20 // 25MB max for voice memos

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		slog.Warn("portal: voice memo upload too large or invalid", "error", err)
		respondError(w, http.StatusBadRequest, "file too large or invalid form (max 25MB)")
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		respondError(w, http.StatusBadRequest, "missing audio field")
		return
	}
	defer file.Close()

	// Validate MIME type
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "audio/webm"
	}
	validAudioTypes := []string{"audio/webm", "audio/wav", "audio/mp4", "audio/ogg", "audio/mpeg"}
	isValid := false
	for _, t := range validAudioTypes {
		if t == mimeType {
			isValid = true
			break
		}
	}
	if !isValid {
		respondError(w, http.StatusBadRequest, "unsupported audio type; accepted: webm, wav, mp4, ogg, mpeg")
		return
	}

	// Read audio bytes
	audioBytes, err := io.ReadAll(file)
	if err != nil {
		slog.Error("portal: failed to read voice memo", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to read audio file")
		return
	}

	// Parse optional metadata
	metadataStr := r.FormValue("metadata")
	var metadata map[string]string
	if metadataStr != "" {
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
			respondError(w, http.StatusBadRequest, "invalid metadata JSON")
			return
		}
	}

	memoID := uuid.New()
	slog.Info("portal: voice memo uploaded",
		"memo_id", memoID,
		"size_bytes", len(audioBytes),
		"mime_type", mimeType,
	)

	// TODO Phase 18: Store audio in S3 and enqueue Asynq voice transcription job.
	// For now, acknowledge receipt. The worker (voice_transcription.go) will
	// download from S3, send to Vertex AI, and save the transcript.

	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"id":      memoID.String(),
		"status":  "queued",
		"message": "Voice memo received and queued for transcription",
	})
}
