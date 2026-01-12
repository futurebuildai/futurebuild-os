package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// TaskHandler handles task-related API endpoints.
// See PRODUCTION_PLAN.md Step 32
type TaskHandler struct {
	scheduleService *service.ScheduleService
}

// NewTaskHandler creates a new TaskHandler instance.
func NewTaskHandler(ss *service.ScheduleService) *TaskHandler {
	return &TaskHandler{scheduleService: ss}
}

// UpdateTaskRequest represents the request body for PUT /tasks/{task_id}.
// See BACKEND_SCOPE.md Section 5.2 (API Structure)
type UpdateTaskRequest struct {
	ManualOverrideDays *float64 `json:"manual_override_days"`
	OverrideReason     string   `json:"override_reason"`
}

// UpdateTaskResponse represents the response for PUT /tasks/{task_id}.
type UpdateTaskResponse struct {
	Task         *models.ProjectTask `json:"task"`
	Recalculated bool                `json:"recalculated"`
}

// UpdateTask handles PUT /api/v1/projects/{id}/tasks/{task_id}.
// Updates a task's manual override duration and triggers schedule recalculation.
// See PRODUCTION_PLAN.md Step 32 (Trigger Point 1)
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	projectID, orgID, err := extractProjectAndOrgIDs(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	taskIDStr := chi.URLParam(r, "task_id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify task exists and belongs to project
	task, err := h.scheduleService.GetTask(r.Context(), taskID, projectID, orgID)
	if err != nil {
		http.Error(w, "Task not found or access denied", http.StatusNotFound)
		return
	}

	// Update the task's manual override if provided
	recalculated := false
	if req.ManualOverrideDays != nil {
		if err := h.scheduleService.UpdateTaskDuration(r.Context(), taskID, projectID, orgID, *req.ManualOverrideDays, req.OverrideReason); err != nil {
			http.Error(w, "Failed to update task", http.StatusInternalServerError)
			return
		}

		// Trigger schedule recalculation
		// See PRODUCTION_PLAN.md Step 32
		_, err = h.scheduleService.RecalculateSchedule(r.Context(), projectID, orgID)
		if err != nil {
			http.Error(w, "Schedule recalculation failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		recalculated = true

		// Fetch updated task
		task, _ = h.scheduleService.GetTask(r.Context(), taskID, projectID, orgID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UpdateTaskResponse{
		Task:         task,
		Recalculated: recalculated,
	})
}

// ProgressRequest represents the request body for POST /tasks/{task_id}/progress.
// See CPM_RES_MODEL_SPEC.md Section 20.2
type ProgressRequest struct {
	PercentComplete int    `json:"percent_complete"`
	Notes           string `json:"notes"`
	// UserID removed from body per /CTO Audit Correction #1
}

// ProgressResponse represents the response for POST /tasks/{task_id}/progress.
type ProgressResponse struct {
	Task          *models.ProjectTask `json:"task"`
	StatusChanged bool                `json:"status_changed"`
	Recalculated  bool                `json:"recalculated"`
}

// RecordProgress handles POST /api/v1/projects/{id}/tasks/{task_id}/progress.
// Records task progress and triggers recalculation if task completes.
// See PRODUCTION_PLAN.md Step 32 (Trigger Point 2)
func (h *TaskHandler) RecordProgress(w http.ResponseWriter, r *http.Request) {
	projectID, orgID, err := extractProjectAndOrgIDs(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	taskIDStr := chi.URLParam(r, "task_id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req ProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify task exists
	task, err := h.scheduleService.GetTask(r.Context(), taskID, projectID, orgID)
	if err != nil {
		http.Error(w, "Task not found or access denied", http.StatusNotFound)
		return
	}

	// Security Fix: Derive UserID from context, not request body
	// See /CTO Audit Correction #1
	// For now, using a placeholder until Auth middleware provides it
	userID := uuid.Nil
	if val := r.Context().Value("user_id"); val != nil {
		if id, ok := val.(uuid.UUID); ok {
			userID = id
		}
	}

	if err := h.scheduleService.CreateTaskProgress(r.Context(), projectID, taskID, userID, req.PercentComplete, req.Notes); err != nil {
		http.Error(w, "Failed to persist progress log", http.StatusInternalServerError)
		return
	}

	statusChanged := false
	recalculated := false

	// Transition to Completed if 100%
	// See CPM_RES_MODEL_SPEC.md Section 20.2 (Task Status Transitions)
	if req.PercentComplete >= 100 && task.Status != types.TaskStatusCompleted {
		if err := h.scheduleService.UpdateTaskStatus(r.Context(), taskID, projectID, orgID, types.TaskStatusCompleted); err != nil {
			http.Error(w, "Failed to update task status", http.StatusInternalServerError)
			return
		}
		statusChanged = true

		// Trigger recalculation on status change
		_, err = h.scheduleService.RecalculateSchedule(r.Context(), projectID, orgID)
		if err != nil {
			http.Error(w, "Schedule recalculation failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		recalculated = true

		// Fetch updated task
		task, _ = h.scheduleService.GetTask(r.Context(), taskID, projectID, orgID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ProgressResponse{
		Task:          task,
		StatusChanged: statusChanged,
		Recalculated:  recalculated,
	})
}

// InspectionRequest represents the request body for POST /tasks/{task_id}/inspection.
// See BACKEND_SCOPE.md Section 5.2 (Record Inspection Result)
type InspectionRequest struct {
	Result         types.InspectionResult `json:"result"`
	InspectorName  string                 `json:"inspector_name"`
	InspectionDate string                 `json:"inspection_date"` // YYYY-MM-DD
	Notes          string                 `json:"notes"`
}

// InspectionResponse represents the response for POST /tasks/{task_id}/inspection.
type InspectionResponse struct {
	Task           *models.ProjectTask    `json:"task"`
	Result         types.InspectionResult `json:"result"`
	UnblockedTasks []string               `json:"unblocked_tasks,omitempty"`
	Recalculated   bool                   `json:"recalculated"`
}

// RecordInspection handles POST /api/v1/projects/{id}/tasks/{task_id}/inspection.
// Records inspection result and triggers recalculation if passed.
// See PRODUCTION_PLAN.md Step 32 (Trigger Point 3)
func (h *TaskHandler) RecordInspection(w http.ResponseWriter, r *http.Request) {
	projectID, orgID, err := extractProjectAndOrgIDs(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	taskIDStr := chi.URLParam(r, "task_id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req InspectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate result enum via model-level logic
	// See /CTO Audit Correction #3
	validResults := map[types.InspectionResult]bool{
		types.InspectionResultPending:     true,
		types.InspectionResultPassed:      true,
		types.InspectionResultFailed:      true,
		types.InspectionResultConditional: true,
	}
	if !validResults[req.Result] {
		http.Error(w, "Invalid inspection result", http.StatusBadRequest)
		return
	}

	// Parse inspection date
	inspectionDate, err := time.Parse("2006-01-02", req.InspectionDate)
	if err != nil {
		http.Error(w, "Invalid inspection_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// Verify task exists and is an inspection
	task, err := h.scheduleService.GetTask(r.Context(), taskID, projectID, orgID)
	if err != nil {
		http.Error(w, "Task not found or access denied", http.StatusNotFound)
		return
	}

	if !task.IsInspection {
		http.Error(w, "This task is not an inspection task", http.StatusBadRequest)
		return
	}

	// Add audit trail persistence
	if err := h.scheduleService.CreateInspectionRecord(r.Context(), projectID, taskID, req.InspectorName, string(req.Result), req.Notes, inspectionDate); err != nil {
		http.Error(w, "Failed to persist inspection record", http.StatusInternalServerError)
		return
	}

	recalculated := false

	// If inspection passed, mark task complete and trigger recalculation
	// See CPM_RES_MODEL_SPEC.md Section 19.1 (Inspection Gate Rule)
	if req.Result == types.InspectionResultPassed {
		if err := h.scheduleService.UpdateTaskStatus(r.Context(), taskID, projectID, orgID, types.TaskStatusCompleted); err != nil {
			http.Error(w, "Failed to update task status", http.StatusInternalServerError)
			return
		}

		// Trigger recalculation to unblock dependent tasks
		_, err = h.scheduleService.RecalculateSchedule(r.Context(), projectID, orgID)
		if err != nil {
			http.Error(w, "Schedule recalculation failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		recalculated = true

		// Fetch updated task
		task, _ = h.scheduleService.GetTask(r.Context(), taskID, projectID, orgID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(InspectionResponse{
		Task:         task,
		Result:       req.Result,
		Recalculated: recalculated,
	})
}

// extractProjectAndOrgIDs extracts project ID from URL param and org ID from header.
func extractProjectAndOrgIDs(r *http.Request) (projectID, orgID uuid.UUID, err error) {
	projectIDStr := chi.URLParam(r, "id")
	projectID, err = uuid.Parse(projectIDStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid project ID: %w", err)
	}

	orgIDStr := r.Header.Get("X-Org-ID")
	if orgIDStr == "" {
		return uuid.Nil, uuid.Nil, fmt.Errorf("X-Org-ID header is required")
	}
	orgID, err = uuid.Parse(orgIDStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	return projectID, orgID, nil
}
