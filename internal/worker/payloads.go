package worker

import (
	"github.com/hibiken/asynq"
)

// Task Types
// See BACKEND_SCOPE.md Section 3.5 (Action Engine)
const (
	TypeDailyBriefing    = "task:daily_briefing"
	TypeProcurementCheck = "task:procurement_check"
)

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
