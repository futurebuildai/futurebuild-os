package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// VisionHandler handles vision verification endpoints.
// See PRODUCTION_PLAN.md Step 40
type VisionHandler struct {
	visionService *service.VisionService
	db            *pgxpool.Pool
}

// NewVisionHandler creates a new VisionHandler.
func NewVisionHandler(vs *service.VisionService, db *pgxpool.Pool) *VisionHandler {
	return &VisionHandler{
		visionService: vs,
		db:            db,
	}
}

// VerifyTaskRequest represents the request body for POST /tasks/{task_id}/verify.
// See PRODUCTION_PLAN.md Step 40
type VerifyTaskRequest struct {
	ImageURL        string `json:"image_url"`
	TaskDescription string `json:"task_description"`
}

// VerifyTaskResponse represents the response for POST /tasks/{task_id}/verify.
type VerifyTaskResponse struct {
	IsVerified bool    `json:"is_verified"`
	Confidence float64 `json:"confidence"`
}

// VerifyTask handles POST /api/v1/projects/{id}/tasks/{task_id}/verify.
// Completes the "Site Photo Verification Flow" from Step 40.
// See PRODUCTION_PLAN.md Step 40, DATA_SPINE_SPEC.md Section 3.3
func (h *VisionHandler) VerifyTask(w http.ResponseWriter, r *http.Request) {
	// 1. Extract IDs and enforce multi-tenancy
	projectID, orgID, err := extractProjectAndOrgIDs(r)
	if err != nil {
		slog.Warn("vision: invalid project/org IDs", "error", err, "method", r.Method)
		http.Error(w, "Invalid project or organization ID", http.StatusBadRequest)
		return
	}

	taskIDStr := chi.URLParam(r, "task_id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		slog.Warn("vision: invalid task ID", "raw_task_id", taskIDStr, "error", err)
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// 2. Parse request body (L7: limit body size)
	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var req VerifyTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("vision: invalid request body", "error", err, "task_id", taskID)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ImageURL == "" {
		slog.Warn("vision: image_url missing", "task_id", taskID)
		http.Error(w, "image_url is required", http.StatusBadRequest)
		return
	}

	if req.TaskDescription == "" {
		slog.Warn("vision: task_description missing", "task_id", taskID)
		http.Error(w, "task_description is required", http.StatusBadRequest)
		return
	}

	slog.Info("vision: verifying task",
		"task_id", taskID, "project_id", projectID, "org_id", orgID)

	// 3. Call VisionService with Persistence
	// This satisfies CTO Audit requirement: "Database is State"
	isVerified, confidence, err := h.visionService.VerifyAndPersistTask(
		r.Context(),
		h.db,
		taskID,
		projectID,
		orgID,
		req.ImageURL,
		req.TaskDescription,
	)
	if err != nil {
		slog.Error("vision: verification failed",
			"task_id", taskID, "project_id", projectID, "error", err)
		http.Error(w, "Vision verification failed", http.StatusInternalServerError)
		return
	}

	slog.Info("vision: verification completed",
		"task_id", taskID, "is_verified", isVerified, "confidence", confidence)

	// 4. Return result
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(VerifyTaskResponse{
		IsVerified: isVerified,
		Confidence: confidence,
	})
}
