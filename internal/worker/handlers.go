package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/colton/futurebuild/internal/agents"
	"github.com/colton/futurebuild/internal/futureshade"
	"github.com/colton/futurebuild/internal/futureshade/gateway"
	"github.com/colton/futurebuild/internal/futureshade/skills"
	"github.com/colton/futurebuild/pkg/clock"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkerHandler struct {
	focusAgent       *agents.DailyFocusAgent
	procurementAgent *agents.ProcurementAgent
	db               *pgxpool.Pool
	clock            clock.Clock
	// FutureShade Action Bridge fields (optional - initialized via WithSkillExecution)
	skillRegistry     *skills.Registry
	executionRepo     *gateway.Repository
	futureShadeConfig futureshade.Config
}

func NewWorkerHandler(focusAgent *agents.DailyFocusAgent, procurementAgent *agents.ProcurementAgent, db *pgxpool.Pool, clk clock.Clock) *WorkerHandler {
	return &WorkerHandler{
		focusAgent:       focusAgent,
		procurementAgent: procurementAgent,
		db:               db,
		clock:            clk,
	}
}

// WithSkillExecution configures the handler for FutureShade skill execution.
// See specs/FUTURESHADE_AGENTS_SPEC.md Section 4 (Action Bridge)
func (h *WorkerHandler) WithSkillExecution(registry *skills.Registry, repo *gateway.Repository, config futureshade.Config) *WorkerHandler {
	h.skillRegistry = registry
	h.executionRepo = repo
	h.futureShadeConfig = config
	return h
}

// HandleDailyBriefing executes the daily briefing agent logic.
func (h *WorkerHandler) HandleDailyBriefing(ctx context.Context, task *asynq.Task) error {
	slog.Info("Handling Daily Briefing Task...")
	if err := h.focusAgent.Execute(ctx); err != nil {
		return fmt.Errorf("daily briefing agent failed: %w", err)
	}
	slog.Info("Daily Briefing Task Completed Successfully.")
	return nil
}

// HandleProcurementCheck executes the procurement agent logic.
// See PRODUCTION_PLAN.md Step 46
func (h *WorkerHandler) HandleProcurementCheck(ctx context.Context, task *asynq.Task) error {
	slog.Info("Handling Procurement Check Task...")
	if err := h.procurementAgent.Execute(ctx); err != nil {
		return fmt.Errorf("procurement agent failed: %w", err)
	}
	slog.Info("Procurement Check Task Completed Successfully.")
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

	logArgs := []any{"project_id", payload.ProjectID}
	slog.Info("Handling Hydrate Project Task", logArgs...)
	if err := h.procurementAgent.HydrateProject(ctx, payload.ProjectID); err != nil {
		return fmt.Errorf("hydrate project failed: %w", err)
	}
	slog.Info("Hydrate Project Task Completed Successfully", logArgs...)
	return nil
}

// HandleProcurementNotification processes async procurement notifications.
// P1 Performance Fix: Sidecar pattern - notifications are queued instead of written synchronously.
// This handler performs dampening check and logs to communication_logs.
func (h *WorkerHandler) HandleProcurementNotification(ctx context.Context, task *asynq.Task) error {
	var payload ProcurementNotificationPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("invalid notification payload: %w", err)
	}

	slog.Info("Handling Procurement Notification", "item_id", payload.ItemID)

	// Dampening check: Skip if notification was sent in last 72 hours
	shouldSend, err := h.shouldSendNotification(ctx, payload.ItemID, payload.Timestamp)
	if err != nil {
		return fmt.Errorf("dampening check failed: %w", err)
	}
	if !shouldSend {
		slog.Info("Notification dampened (already sent within 72h)", "item_id", payload.ItemID)
		return nil
	}

	// Log notification to communication_logs
	if err := h.logNotification(ctx, payload); err != nil {
		return fmt.Errorf("log notification failed: %w", err)
	}

	slog.Info("Procurement Notification sent", "item_id", payload.ItemID)
	return nil
}

// shouldSendNotification checks communication_logs for recent alerts.
// See User Amendment #4: 72-hour dampening
func (h *WorkerHandler) shouldSendNotification(ctx context.Context, itemID interface{}, timestamp interface{}) (bool, error) {
	query := `
		SELECT COUNT(*) FROM communication_logs
		WHERE related_entity_id = $1
		  AND timestamp > $2 - INTERVAL '72 hours'
		  AND direction = 'Outbound'
	`
	var count int
	err := h.db.QueryRow(ctx, query, itemID, timestamp).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// logNotification persists the alert to communication_logs.
func (h *WorkerHandler) logNotification(ctx context.Context, payload ProcurementNotificationPayload) error {
	query := `
		INSERT INTO communication_logs (
			project_id, direction, content, channel, timestamp,
			related_entity_id, related_entity_type
		)
		SELECT p.id, 'Outbound', $1, 'Chat', $2, $3, 'procurement_item'
		FROM procurement_items pi
		JOIN project_tasks pt ON pi.project_task_id = pt.id
		JOIN projects p ON pt.project_id = p.id
		WHERE pi.id = $4
	`
	content := fmt.Sprintf("[PROCUREMENT ALERT] %s", payload.Message)
	_, err := h.db.Exec(ctx, query, content, payload.Timestamp, payload.ItemID, payload.ItemID)
	return err
}

// HandleSkillExecution processes FutureShade skill execution tasks.
// Implements idempotency (skips if not PENDING) and circuit breaker (skips if disabled).
// See specs/FUTURESHADE_AGENTS_SPEC.md Section 4 (Action Bridge)
func (h *WorkerHandler) HandleSkillExecution(ctx context.Context, task *asynq.Task) error {
	// Parse payload
	var payload SkillExecutionPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("invalid skill execution payload: %w", err)
	}

	logArgs := []any{
		"execution_id", payload.ExecutionID,
		"decision_id", payload.DecisionID,
		"skill_id", payload.SkillID,
	}

	// Circuit Breaker: Check if FutureShade is enabled
	if !h.futureShadeConfig.Enabled {
		slog.Info("FutureShade disabled, skipping skill execution", logArgs...)
		return nil
	}

	// Validate handler was configured
	if h.skillRegistry == nil || h.executionRepo == nil {
		slog.Error("skill execution handler not configured", logArgs...)
		return fmt.Errorf("skill execution handler not configured")
	}

	// Idempotency Check: Skip if not PENDING
	status, err := h.executionRepo.GetStatus(ctx, payload.ExecutionID)
	if err != nil {
		slog.Error("failed to get execution status", append(logArgs, "error", err)...)
		return fmt.Errorf("get execution status: %w", err)
	}
	if status != gateway.StatusPending {
		slog.Info("skill execution already processed, skipping",
			append(logArgs, "current_status", status)...)
		return nil
	}

	// Mark as RUNNING
	if err := h.executionRepo.MarkRunning(ctx, payload.ExecutionID); err != nil {
		slog.Error("failed to mark execution as running", append(logArgs, "error", err)...)
		return fmt.Errorf("mark running: %w", err)
	}

	slog.Info("Starting skill execution", logArgs...)

	// Get skill from registry
	skill, ok := h.skillRegistry.Get(payload.SkillID)
	if !ok {
		errMsg := fmt.Sprintf("skill %q not found in registry", payload.SkillID)
		if err := h.executionRepo.UpdateExecutionStatus(ctx, payload.ExecutionID,
			gateway.StatusFailed, nil, &errMsg); err != nil {
			slog.Error("failed to update execution status", append(logArgs, "error", err)...)
		}
		return fmt.Errorf("skill %q not found in registry", payload.SkillID)
	}

	// Execute the skill
	result, execErr := skill.Execute(ctx, payload.Params)

	// Update execution status based on result
	if execErr != nil {
		errMsg := execErr.Error()
		if err := h.executionRepo.UpdateExecutionStatus(ctx, payload.ExecutionID,
			gateway.StatusFailed, &result.Summary, &errMsg); err != nil {
			slog.Error("failed to update execution status on failure",
				append(logArgs, "error", err)...)
		}
		slog.Error("Skill execution failed", append(logArgs, "error", execErr)...)
		return fmt.Errorf("skill execution failed: %w", execErr)
	}

	// Success
	if err := h.executionRepo.UpdateExecutionStatus(ctx, payload.ExecutionID,
		gateway.StatusCompleted, &result.Summary, nil); err != nil {
		slog.Error("failed to update execution status on success",
			append(logArgs, "error", err)...)
	}

	slog.Info("Skill execution completed successfully",
		append(logArgs, "summary", result.Summary)...)

	return nil
}
