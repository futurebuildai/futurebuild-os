// Package agents provides AI agent implementations for FutureBuild.
// This file implements ProcurementRepository using PostgreSQL via pgxpool.
// See PRODUCTION_PLAN.md: Testing Strategy & CI Reliability Remediation
package agents

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgProcurementRepository implements ProcurementRepository using PostgreSQL.
// This is the production implementation that wraps pgxpool.Pool.
type PgProcurementRepository struct {
	db *pgxpool.Pool
}

// NewPgProcurementRepository creates a new PostgreSQL-backed repository.
func NewPgProcurementRepository(db *pgxpool.Pool) *PgProcurementRepository {
	return &PgProcurementRepository{db: db}
}

// StreamItems iterates through procurement items one-by-one via callback.
// P1 Scalability Fix: Uses cursor-based iteration to prevent OOM at scale.
// Items are re-evaluated every run to allow status transitions (ok → warning → critical).
// Notification dampening is handled separately in LogNotification.
func (r *PgProcurementRepository) StreamItems(ctx context.Context, process ItemProcessor) error {
	query := `
		SELECT 
			pi.id,
			pi.name,
			pi.lead_time_weeks,
			pi.status,
			pt.early_start,
			pc.zip_code,
			pi.project_task_id
		FROM procurement_items pi
		JOIN project_tasks pt ON pi.project_task_id = pt.id
		JOIN projects p ON pt.project_id = p.id
		LEFT JOIN project_context pc ON pc.project_id = p.id
		WHERE p.status = 'Active'
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item procurementRow
		if err := rows.Scan(&item.ID, &item.Name, &item.LeadTimeWeeks, &item.Status, &item.EarlyStart, &item.ZipCode, &item.ProjectTaskID); err != nil {
			return fmt.Errorf("scan row: %w", err)
		}
		if err := process(item); err != nil {
			return fmt.Errorf("process item %s: %w", item.ID, err)
		}
	}
	return rows.Err()
}

// UpdateBatch sends all UPDATE queries in a single pgx Batch round-trip.
// P1 Performance Fix: Reduces N database round-trips to 1 per batch.
func (r *PgProcurementRepository) UpdateBatch(ctx context.Context, now time.Time, batch []alertResult) error {
	if len(batch) == 0 {
		return nil
	}

	b := &pgx.Batch{}
	updateQuery := `
		UPDATE procurement_items
		SET status = $1, calculated_order_date = $2, last_checked_at = $3
		WHERE id = $4
	`
	for _, result := range batch {
		b.Queue(updateQuery, string(result.NewStatus), result.CalculatedOrderDate, now, result.ID)
	}

	br := r.db.SendBatch(ctx, b)
	defer br.Close()

	for i := 0; i < len(batch); i++ {
		if _, err := br.Exec(); err != nil {
			slog.Error("batch update failed", "index", i, "id", batch[i].ID, "error", err)
			// Continue - don't fail entire batch for one item
		}
	}

	slog.Debug("batch flushed", "count", len(batch))
	return nil
}

// HydrateProject populates procurement_items for a specific project.
func (r *PgProcurementRepository) HydrateProject(ctx context.Context, projectID uuid.UUID) error {
	query := `
		INSERT INTO procurement_items (project_task_id, name, lead_time_weeks)
		SELECT pt.id, pt.name, COALESCE(wt.lead_time_weeks_min, 4)
		FROM project_tasks pt
		LEFT JOIN procurement_items pi ON pi.project_task_id = pt.id
		JOIN wbs_tasks wt ON pt.wbs_code = wt.code
		WHERE wt.is_long_lead = true 
		  AND pi.id IS NULL
		  AND pt.project_id = $1
		ON CONFLICT DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, projectID)
	if err != nil {
		return fmt.Errorf("hydrate project %s: %w", projectID, err)
	}
	slog.Info("Project hydration completed", "project_id", projectID)
	return nil
}

// ShouldSendNotification checks communication_logs for recent alerts.
// Uses 72-hour dampening window per User Amendment #4.
func (r *PgProcurementRepository) ShouldSendNotification(ctx context.Context, itemID uuid.UUID, now time.Time) (bool, error) {
	query := `
		SELECT COUNT(*) FROM communication_logs
		WHERE related_entity_id = $1
		  AND timestamp > ($2::timestamptz - INTERVAL '72 hours')
		  AND direction = 'Outbound'
	`
	var count int
	err := r.db.QueryRow(ctx, query, itemID, now).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// LogNotification persists the alert to communication_logs.
func (r *PgProcurementRepository) LogNotification(ctx context.Context, result alertResult, now time.Time) error {
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
	content := fmt.Sprintf("[PROCUREMENT ALERT] %s", result.Message)
	_, err := r.db.Exec(ctx, query, content, now, result.ID, result.ID)
	if err != nil {
		slog.Error("failed to log notification", "id", result.ID, "error", err)
		return err
	}
	slog.Info("Notification logged", "item_id", result.ID, "message", result.Message)
	return nil
}
