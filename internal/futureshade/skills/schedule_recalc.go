package skills

import (
	"context"
	"errors"
	"fmt"

	"github.com/colton/futurebuild/internal/service"
	"github.com/google/uuid"
)

// ScheduleRecalcSkill wraps ScheduleService.RecalculateSchedule for Tribunal-triggered execution.
// See specs/FUTURESHADE_AGENTS_SPEC.md Section 4.3 (Skill Wrappers)
type ScheduleRecalcSkill struct {
	scheduleService *service.ScheduleService
}

// NewScheduleRecalcSkill creates a new ScheduleRecalcSkill.
func NewScheduleRecalcSkill(svc *service.ScheduleService) *ScheduleRecalcSkill {
	return &ScheduleRecalcSkill{scheduleService: svc}
}

// ID returns the skill identifier.
func (s *ScheduleRecalcSkill) ID() string {
	return "schedule_recalc"
}

// Execute recalculates the schedule for a project.
// Required params: project_id (string UUID), org_id (string UUID)
func (s *ScheduleRecalcSkill) Execute(ctx context.Context, params map[string]any) (Result, error) {
	// Extract and validate required parameters
	projectIDStr, ok := params["project_id"].(string)
	if !ok || projectIDStr == "" {
		return Result{
			Success: false,
			Summary: "Missing required parameter: project_id",
		}, errors.New("missing required parameter: project_id")
	}

	orgIDStr, ok := params["org_id"].(string)
	if !ok || orgIDStr == "" {
		return Result{
			Success: false,
			Summary: "Missing required parameter: org_id",
		}, errors.New("missing required parameter: org_id")
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return Result{
			Success: false,
			Summary: fmt.Sprintf("Invalid project_id format: %v", err),
		}, fmt.Errorf("invalid project_id: %w", err)
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return Result{
			Success: false,
			Summary: fmt.Sprintf("Invalid org_id format: %v", err),
		}, fmt.Errorf("invalid org_id: %w", err)
	}

	// Execute schedule recalculation
	result, err := s.scheduleService.RecalculateSchedule(ctx, projectID, orgID)
	if err != nil {
		return Result{
			Success: false,
			Summary: fmt.Sprintf("Schedule recalculation failed: %v", err),
		}, err
	}

	// Build success result with CPM data
	return Result{
		Success: true,
		Summary: fmt.Sprintf("Schedule recalculated: %d tasks, %d on critical path, project ends %s",
			len(result.Tasks), len(result.CriticalPath), result.ProjectEnd.Format("2006-01-02")),
		Data: map[string]any{
			"task_count":          len(result.Tasks),
			"critical_path_count": len(result.CriticalPath),
			"project_end":         result.ProjectEnd,
		},
	}, nil
}
