package worker

import (
	"context"
	"fmt"
	"log"

	"github.com/colton/futurebuild/internal/agents"
	"github.com/hibiken/asynq"
)

type WorkerHandler struct {
	focusAgent       *agents.DailyFocusAgent
	procurementAgent *agents.ProcurementAgent
}

func NewWorkerHandler(focusAgent *agents.DailyFocusAgent, procurementAgent *agents.ProcurementAgent) *WorkerHandler {
	return &WorkerHandler{
		focusAgent:       focusAgent,
		procurementAgent: procurementAgent,
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

// HandleProcurementCheck executes the procurement agent logic.
// See PRODUCTION_PLAN.md Step 46
func (h *WorkerHandler) HandleProcurementCheck(ctx context.Context, task *asynq.Task) error {
	log.Println("Handling Procurement Check Task...")
	if err := h.procurementAgent.Execute(ctx); err != nil {
		return fmt.Errorf("procurement agent failed: %w", err)
	}
	log.Println("Procurement Check Task Completed Successfully.")
	return nil
}
