package agents

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/colton/futurebuild/pkg/clock"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProcurementAgent monitors long-lead items and calculates order dates.
// See PRODUCTION_PLAN.md Step 46, BACKEND_SCOPE.md Section 2.5
// Refactored for deterministic simulation: PRODUCTION_PLAN.md Step 49
type ProcurementAgent struct {
	db      *pgxpool.Pool
	weather types.WeatherService
	clock   clock.Clock
}

// NewProcurementAgent creates a new agent instance.
// Clock is required for deterministic time simulation (Step 49).
func NewProcurementAgent(db *pgxpool.Pool, weather types.WeatherService, clk clock.Clock) *ProcurementAgent {
	return &ProcurementAgent{
		db:      db,
		weather: weather,
		clock:   clk,
	}
}

// procurementRow represents the joined data for a single procurement item.
// See BACKEND_SCOPE.md Section 2.5 (Long-Lead Items)
type procurementRow struct {
	ID            uuid.UUID
	Name          string
	LeadTimeWeeks int
	Status        types.ProcurementAlertStatus
	EarlyStart    *time.Time
	// IsExterior could be derived from WBS Phase, but for MVP we assume all long-lead are exterior-related.
	// ZipCode for weather lookup (stubbed for MVP)
	ZipCode       string
	ProjectTaskID uuid.UUID
}

// alertResult holds the calculated status for a procurement item.
type alertResult struct {
	ID                  uuid.UUID
	NewStatus           types.ProcurementAlertStatus
	CalculatedOrderDate time.Time
	ShouldNotify        bool
	Message             string
}

// ItemProcessor is a callback function for processing procurement items one-by-one.
// Uses cursor-based iteration to prevent OOM at scale (P1 Scalability Fix).
type ItemProcessor func(item procurementRow) error

// Execute runs the procurement analysis for all active projects.
// See PRODUCTION_PLAN.md Step 46
// P1 Scalability Fix: Uses streaming iteration instead of loading all items into memory.
// P1 Performance Fix: Hydration is now event-driven (HydrateProject), not cron-swept.
func (a *ProcurementAgent) Execute(ctx context.Context) error {
	slog.Info("Starting Procurement Agent...")

	// NOTE: Hydration is now event-driven via HandleHydrateProject task.
	// See implementation_plan.md: "Event-Driven Hydration"

	now := a.clock.Now().Truncate(24 * time.Hour)
	var processed int

	// Stream items one-by-one to avoid unbounded memory allocation
	err := a.streamItems(ctx, func(item procurementRow) error {
		result := a.analyzeItem(item, now)
		if err := a.updateItem(ctx, result); err != nil {
			slog.Error("failed to update item", "id", item.ID, "error", err)
			// Continue processing other items
			return nil
		}

		// Notification Dampening & Delivery
		// See User Amendment #4
		if result.ShouldNotify {
			shouldSend, err := a.shouldSendNotification(ctx, result.ID)
			if err != nil {
				slog.Error("notification check failed", "id", item.ID, "error", err)
				return nil
			}
			if shouldSend {
				a.logNotification(ctx, result)
			}
		}
		processed++
		return nil
	})

	if err != nil {
		return fmt.Errorf("streaming items failed: %w", err)
	}

	slog.Info("Procurement Agent completed", "items_processed", processed)
	return nil
}

// HydrateProject populates procurement_items for a specific project.
// Called via event-driven task (TypeHydrateProject) on project creation.
// P1 Performance Fix: Replaces cron-swept hydrateItems with project-scoped operation.
// See User Amendment #2, Critical Blocker A.3 Remediation
func (a *ProcurementAgent) HydrateProject(ctx context.Context, projectID uuid.UUID) error {
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
	_, err := a.db.Exec(ctx, query, projectID)
	if err != nil {
		return fmt.Errorf("hydrate project %s: %w", projectID, err)
	}
	slog.Info("Project hydration completed", "project_id", projectID)
	return nil
}

// streamItems iterates through procurement items one-by-one via callback.
// P1 Scalability Fix: Uses cursor-based iteration to prevent OOM at scale.
// See User Amendment #3, BACKEND_SCOPE.md Section 2.5
func (a *ProcurementAgent) streamItems(ctx context.Context, process ItemProcessor) error {
	// Single optimized query per L7 performance standard
	query := `
		SELECT 
			pi.id,
			pi.name,
			pi.lead_time_weeks,
			pi.status,
			pt.early_start,
			COALESCE(pc.zip_code, '78701') as zip_code,
			pi.project_task_id
		FROM procurement_items pi
		JOIN project_tasks pt ON pi.project_task_id = pt.id
		JOIN projects p ON pt.project_id = p.id
		LEFT JOIN project_context pc ON pc.project_id = p.id
		WHERE p.status = 'Active'
		  AND pi.status NOT IN ('ok')
	`
	rows, err := a.db.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r procurementRow
		if err := rows.Scan(&r.ID, &r.Name, &r.LeadTimeWeeks, &r.Status, &r.EarlyStart, &r.ZipCode, &r.ProjectTaskID); err != nil {
			return fmt.Errorf("scan row: %w", err)
		}
		if err := process(r); err != nil {
			return fmt.Errorf("process item %s: %w", r.ID, err)
		}
	}
	return rows.Err()
}

// analyzeItem calculates the order date and determines the alert status.
// See PRODUCTION_PLAN.md Step 46 Requirements
func (a *ProcurementAgent) analyzeItem(item procurementRow, now time.Time) alertResult {
	// stagingBufferDays: Time for jobsite staging before work can begin
	// See PRODUCTION_PLAN.md Step 46: "Safety Buffer: 2 days for onsite staging"
	const stagingBufferDays = 2
	const warningThresholdDays = 3

	result := alertResult{
		ID:        item.ID,
		NewStatus: types.ProcurementAlertOK,
	}

	if item.EarlyStart == nil {
		// No schedule data yet
		result.NewStatus = types.ProcurementAlertPending
		return result
	}

	earlyStart := item.EarlyStart.Truncate(24 * time.Hour)
	leadTimeDays := item.LeadTimeWeeks * 7
	weatherBuffer := 0

	// Weather interaction (SWIM integration)
	// See PRODUCTION_PLAN.md Step 46: Weather Buffer
	// For MVP, we check if precipitation probability > 50% near the start date
	// P0 FIX: Do NOT use hardcoded fallback coordinates. If geocoding is unavailable,
	// we default to 0 weather buffer rather than use incorrect location data.
	// TODO: Inject GeocodingService and use item.ZipCode for location-specific weather.
	if a.weather != nil && item.ZipCode != "" {
		// ZipCode is available but geocoding service is not yet wired.
		// Log a metric indicating geocoding is needed for accurate weather data.
		// FAIL SAFE: Skip weather buffer calculation until geocoding is implemented.
		slog.Warn("weather buffer skipped: geocoding not implemented",
			"item_id", item.ID,
			"zip_code", item.ZipCode,
			"action", "using_zero_buffer",
		)
		// weatherBuffer remains 0 - no incorrect location data will be used
	}

	// MRP Feedback Loop Calculations (PRODUCTION_PLAN.md Step 46):
	// NeedByDate = EarlyStart - stagingBuffer (material must arrive 2 days before installation)
	// CalculatedOrderDate = NeedByDate - leadTime - weatherBuffer
	needByDate := earlyStart.AddDate(0, 0, -stagingBufferDays)
	mustOrderDate := needByDate.AddDate(0, 0, -(leadTimeDays + weatherBuffer))
	result.CalculatedOrderDate = mustOrderDate

	// State Detection
	// See PRODUCTION_PLAN.md Step 46
	daysUntilMustOrder := int(mustOrderDate.Sub(now).Hours() / 24)

	switch {
	case now.After(mustOrderDate) && item.Status != types.ProcurementAlertCritical:
		result.NewStatus = types.ProcurementAlertCritical
		result.ShouldNotify = true
		result.Message = fmt.Sprintf("⚠️ ACTION REQUIRED: Order %s by %s to avoid schedule slip.",
			item.Name, mustOrderDate.Format("Jan 02, 2006"))
	case daysUntilMustOrder <= warningThresholdDays && daysUntilMustOrder > 0:
		result.NewStatus = types.ProcurementAlertWarning
		if item.Status == types.ProcurementAlertPending || item.Status == types.ProcurementAlertOK {
			result.ShouldNotify = true
			result.Message = fmt.Sprintf("⏰ HEADS UP: Order %s soon (by %s).",
				item.Name, mustOrderDate.Format("Jan 02, 2006"))
		}
	case daysUntilMustOrder > warningThresholdDays:
		result.NewStatus = types.ProcurementAlertOK
	default:
		// Already critical, check for nag mode (> 3 days since last alert)
		if item.Status == types.ProcurementAlertCritical {
			result.NewStatus = types.ProcurementAlertCritical
			result.ShouldNotify = true // Will be filtered by dampening
			result.Message = fmt.Sprintf("🚨 OVERDUE: Order %s immediately!", item.Name)
		}
	}

	return result
}

// updateItem persists the calculated status and order date.
// See PRODUCTION_PLAN.md Step 49: Uses injected clock for deterministic simulation.
func (a *ProcurementAgent) updateItem(ctx context.Context, result alertResult) error {
	query := `
		UPDATE procurement_items
		SET status = $1, calculated_order_date = $2, last_checked_at = $3
		WHERE id = $4
	`
	_, err := a.db.Exec(ctx, query, string(result.NewStatus), result.CalculatedOrderDate, a.clock.Now(), result.ID)
	return err
}

// shouldSendNotification checks communication_logs for recent alerts.
// See User Amendment #4: 72-hour dampening, Optimized via Migration 000046
// See PRODUCTION_PLAN.md Step 49: Uses injected clock for deterministic simulation.
func (a *ProcurementAgent) shouldSendNotification(ctx context.Context, itemID uuid.UUID) (bool, error) {
	// Check for alerts in the last 72 hours linked to this specific entity
	// Uses 'related_entity_id' column added in migration 000046
	query := `
		SELECT COUNT(*) FROM communication_logs
		WHERE related_entity_id = $1
		  AND timestamp > $2 - INTERVAL '72 hours'
		  AND direction = 'Outbound'
	`
	var count int
	err := a.db.QueryRow(ctx, query, itemID, a.clock.Now()).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// logNotification persists the alert to communication_logs.
// See PRODUCTION_PLAN.md Step 49: Uses injected clock for deterministic simulation.
func (a *ProcurementAgent) logNotification(ctx context.Context, result alertResult) {
	// Insert into communication_logs with structured entity data
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
	_, err := a.db.Exec(ctx, query, content, a.clock.Now(), result.ID, result.ID)
	if err != nil {
		slog.Error("failed to log notification", "id", result.ID, "error", err)
	} else {
		slog.Info("Notification logged", "item_id", result.ID, "message", result.Message)
	}
}
