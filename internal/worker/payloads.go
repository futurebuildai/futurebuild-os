package worker

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Task Types
// See BACKEND_SCOPE.md Section 3.5 (Action Engine)
const (
	TypeDailyBriefing           = "task:daily_briefing"
	TypeProcurementCheck        = "task:procurement_check"
	TypeHydrateProject          = "task:hydrate_project"
	TypeProcurementNotification = "task:procurement_notification"
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

// ProcurementNotificationPayload contains notification data for async delivery.
// P1 Performance Fix: Enables sidecar notification pattern to reduce DB round-trips.
type ProcurementNotificationPayload struct {
	ItemID    uuid.UUID `json:"item_id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// NewProcurementNotificationTask creates a task for async notification delivery.
// P1 Performance Fix: Notifications are queued instead of written synchronously.
func NewProcurementNotificationTask(itemID uuid.UUID, message string, ts time.Time) (*asynq.Task, error) {
	payload, err := json.Marshal(ProcurementNotificationPayload{
		ItemID:    itemID,
		Message:   message,
		Timestamp: ts,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeProcurementNotification, payload, asynq.Queue("default")), nil
}
