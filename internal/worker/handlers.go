package worker

import (
	"context"
	"fmt"
	"log"

	"github.com/colton/futurebuild/internal/agents"
	"github.com/hibiken/asynq"
)

type WorkerHandler struct {
	focusAgent *agents.DailyFocusAgent
}

func NewWorkerHandler(focusAgent *agents.DailyFocusAgent) *WorkerHandler {
	return &WorkerHandler{
		focusAgent: focusAgent,
	}
}

// HandleDailyBriefing executes the daily briefing agent logic.
func (h *WorkerHandler) HandleDailyBriefing(ctx context.Context, task *asynq.Task) error {
	log.Println("Handling Daily Briefing Task...")
	if err := h.focusAgent.Execute(ctx); err != nil {
		return fmt.Errorf("daily briefing agent failed: %w", err)
	}
	log.Println("Daily Briefing Task Completed Successfully.")
	return nil
}
