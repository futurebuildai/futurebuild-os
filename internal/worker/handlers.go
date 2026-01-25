package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/colton/futurebuild/internal/agents"
	"github.com/colton/futurebuild/pkg/clock"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkerHandler struct {
	focusAgent       *agents.DailyFocusAgent
	procurementAgent *agents.ProcurementAgent
	db               *pgxpool.Pool
	clock            clock.Clock
}

func NewWorkerHandler(focusAgent *agents.DailyFocusAgent, procurementAgent *agents.ProcurementAgent, db *pgxpool.Pool, clk clock.Clock) *WorkerHandler {
	return &WorkerHandler{
		focusAgent:       focusAgent,
		procurementAgent: procurementAgent,
		db:               db,
		clock:            clk,
	}
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
