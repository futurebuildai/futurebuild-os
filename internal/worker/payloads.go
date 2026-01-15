package worker

import (
	"github.com/hibiken/asynq"
)

// Task Types
// See BACKEND_SCOPE.md Section 3.5 (Action Engine)
const (
	TypeDailyBriefing = "task:daily_briefing"
)

// NewDailyBriefingTask creates a task for daily briefing.
// No payload needed for a global daily briefing trigger.
func NewDailyBriefingTask() *asynq.Task {
	return asynq.NewTask(TypeDailyBriefing, nil)
}
