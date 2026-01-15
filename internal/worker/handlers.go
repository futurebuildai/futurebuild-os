package worker

import (
	"context"
	"encoding/json"
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

// HandleHydrateProject handles project-scoped hydration of procurement items.
// Triggered when a new project is created (event-driven).
// P1 Performance Fix: Replaces inefficient cron-swept hydration.
func (h *WorkerHandler) HandleHydrateProject(ctx context.Context, task *asynq.Task) error {
	var payload HydrateProjectPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("invalid hydrate project payload: %w", err)
	}

	log.Printf("Handling Hydrate Project Task for project: %s", payload.ProjectID)
	if err := h.procurementAgent.HydrateProject(ctx, payload.ProjectID); err != nil {
		return fmt.Errorf("hydrate project failed: %w", err)
	}
	log.Printf("Hydrate Project Task Completed Successfully for project: %s", payload.ProjectID)
	return nil
}
