package worker

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Task Types
// See BACKEND_SCOPE.md Section 3.5 (Action Engine)
const (
	TypeDailyBriefing    = "task:daily_briefing"
	TypeProcurementCheck = "task:procurement_check"
	TypeHydrateProject   = "task:hydrate_project"
)

// HydrateProjectPayload contains the project ID for scoped hydration.
// See implementation_plan.md: "Event-Driven Hydration"
type HydrateProjectPayload struct {
	ProjectID uuid.UUID `json:"project_id"`
}

// NewDailyBriefingTask creates a task for daily briefing.
// No payload needed for a global daily briefing trigger.
func NewDailyBriefingTask() *asynq.Task {
	return asynq.NewTask(TypeDailyBriefing, nil)
}

// NewProcurementCheckTask creates a task for procurement analysis.
// See PRODUCTION_PLAN.md Step 46
func NewProcurementCheckTask() *asynq.Task {
	return asynq.NewTask(TypeProcurementCheck, nil)
}

// NewHydrateProjectTask creates a task for project-scoped hydration.
// Enqueued when a project is created to populate procurement_items.
// P1 Performance Fix: Replaces cron-swept hydration.
func NewHydrateProjectTask(projectID uuid.UUID) (*asynq.Task, error) {
	payload, err := json.Marshal(HydrateProjectPayload{ProjectID: projectID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeHydrateProject, payload), nil
}
