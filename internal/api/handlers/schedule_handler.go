package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/colton/futurebuild/internal/service"
)

// ScheduleHandler handles schedule-related API endpoints.
// See STEP_89_DEPENDENCY_ARROWS.md (Phase 14: Gantt data for frontend)
type ScheduleHandler struct {
	scheduleService service.ScheduleServicer
}

// NewScheduleHandler creates a new ScheduleHandler instance.
func NewScheduleHandler(ss service.ScheduleServicer) *ScheduleHandler {
	return &ScheduleHandler{scheduleService: ss}
}

// GetSchedule returns the full GanttData for a project.
// GET /api/v1/projects/{id}/schedule
func (h *ScheduleHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	projectID, orgID, err := extractProjectAndOrgIDs(r)
	if err != nil {
		slog.Warn("schedule: invalid project/org IDs", "error", err)
		http.Error(w, "Invalid project or organization ID", http.StatusBadRequest)
		return
	}

	ganttData, err := h.scheduleService.GetGanttData(r.Context(), projectID, orgID)
	if err != nil {
		slog.Error("schedule: failed to fetch gantt data", "project_id", projectID, "error", err)
		http.Error(w, "Failed to retrieve schedule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ganttData)
}

// RecalculateSchedule triggers a CPM recalculation and returns updated GanttData.
// POST /api/v1/projects/{id}/schedule/recalculate
func (h *ScheduleHandler) RecalculateSchedule(w http.ResponseWriter, r *http.Request) {
	projectID, orgID, err := extractProjectAndOrgIDs(r)
	if err != nil {
		slog.Warn("schedule: invalid project/org IDs", "error", err)
		http.Error(w, "Invalid project or organization ID", http.StatusBadRequest)
		return
	}

	// Trigger recalculation first
	_, err = h.scheduleService.RecalculateSchedule(r.Context(), projectID, orgID)
	if err != nil {
		slog.Error("schedule: recalculation failed", "project_id", projectID, "error", err)
		http.Error(w, "Schedule recalculation failed", http.StatusInternalServerError)
		return
	}

	// Fetch the updated GanttData
	ganttData, err := h.scheduleService.GetGanttData(r.Context(), projectID, orgID)
	if err != nil {
		slog.Error("schedule: failed to fetch updated gantt data", "project_id", projectID, "error", err)
		http.Error(w, "Failed to retrieve updated schedule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ganttData)
}
