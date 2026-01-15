package agents

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProcurementAgent monitors long-lead items and calculates order dates.
// See PRODUCTION_PLAN.md Step 46, BACKEND_SCOPE.md Section 2.5
type ProcurementAgent struct {
	db      *pgxpool.Pool
	weather types.WeatherService
}

// NewProcurementAgent creates a new agent instance.
func NewProcurementAgent(db *pgxpool.Pool, weather types.WeatherService) *ProcurementAgent {
	return &ProcurementAgent{
		db:      db,
		weather: weather,
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

// Execute runs the procurement analysis for all active projects.
// See PRODUCTION_PLAN.md Step 46
func (a *ProcurementAgent) Execute(ctx context.Context) error {
	slog.Info("Starting Procurement Agent...")

	// 1. Discovery Step (Auto-Hydration)
	// See User Amendment #2: Idempotent population of procurement_items
	if err := a.hydrateItems(ctx); err != nil {
		return fmt.Errorf("hydration failed: %w", err)
	}

	// 2. Fetch all items with single optimized query
	// See User Amendment #3: No N+1
	items, err := a.fetchItems(ctx)
	if err != nil {
		return fmt.Errorf("fetch failed: %w", err)
	}

	now := time.Now().Truncate(24 * time.Hour)

	// 3. Process each item
	for _, item := range items {
		result := a.analyzeItem(item, now)
		if err := a.updateItem(ctx, result); err != nil {
			slog.Error("failed to update item", "id", item.ID, "error", err)
			continue
		}

		// 4. Notification Dampening & Delivery
		// See User Amendment #4
		if result.ShouldNotify {
			shouldSend, err := a.shouldSendNotification(ctx, result.ID)
			if err != nil {
				slog.Error("notification check failed", "id", item.ID, "error", err)
				continue
			}
			if shouldSend {
				a.logNotification(ctx, result)
			}
		}
	}

	slog.Info("Procurement Agent completed", "items_processed", len(items))
	return nil
}

// hydrateItems populates procurement_items for tasks marked as is_long_lead.
// See User Amendment #2
func (a *ProcurementAgent) hydrateItems(ctx context.Context) error {
	query := `
		INSERT INTO procurement_items (project_task_id, name, lead_time_weeks)
		SELECT pt.id, pt.name, 4
		FROM project_tasks pt
		LEFT JOIN procurement_items pi ON pi.project_task_id = pt.id
		JOIN projects p ON pt.project_id = p.id
		JOIN wbs_tasks wt ON pt.wbs_code = wt.code
		WHERE wt.is_long_lead = true 
		  AND pi.id IS NULL
		  AND p.status IN ('Active', 'Preconstruction')
		ON CONFLICT DO NOTHING
	`
	_, err := a.db.Exec(ctx, query)
	return err
}

// fetchItems retrieves all relevant items with a single JOIN query.
// See User Amendment #3, BACKEND_SCOPE.md Section 2.5
func (a *ProcurementAgent) fetchItems(ctx context.Context) ([]procurementRow, error) {
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
		return nil, err
	}
	defer rows.Close()

	var items []procurementRow
	for rows.Next() {
		var r procurementRow
		if err := rows.Scan(&r.ID, &r.Name, &r.LeadTimeWeeks, &r.Status, &r.EarlyStart, &r.ZipCode, &r.ProjectTaskID); err != nil {
			return nil, err
		}
		items = append(items, r)
	}
	return items, rows.Err()
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
	// Stub: Using default lat/long for weather check
	if a.weather != nil {
		forecast, err := a.weather.GetForecast(30.2672, -97.7431) // Austin, TX default
		if err == nil && forecast.PrecipitationProbability > 0.5 {
			weatherBuffer = 2
		}
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
func (a *ProcurementAgent) updateItem(ctx context.Context, result alertResult) error {
	query := `
		UPDATE procurement_items
		SET status = $1, calculated_order_date = $2, last_checked_at = NOW()
		WHERE id = $3
	`
	_, err := a.db.Exec(ctx, query, string(result.NewStatus), result.CalculatedOrderDate, result.ID)
	return err
}

// shouldSendNotification checks communication_logs for recent alerts.
// See User Amendment #4: 72-hour dampening, Optimized via Migration 000046
func (a *ProcurementAgent) shouldSendNotification(ctx context.Context, itemID uuid.UUID) (bool, error) {
	// Check for alerts in the last 72 hours linked to this specific entity
	// Uses 'related_entity_id' column added in migration 000046
	query := `
		SELECT COUNT(*) FROM communication_logs
		WHERE related_entity_id = $1
		  AND timestamp > NOW() - INTERVAL '72 hours'
		  AND direction = 'Outbound'
	`
	var count int
	err := a.db.QueryRow(ctx, query, itemID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// logNotification persists the alert to communication_logs.
func (a *ProcurementAgent) logNotification(ctx context.Context, result alertResult) {
	// Insert into communication_logs with structured entity data
	query := `
		INSERT INTO communication_logs (
			project_id, direction, content, channel, timestamp, 
			related_entity_id, related_entity_type
		)
		SELECT p.id, 'Outbound', $1, 'Chat', NOW(), $2, 'procurement_item'
		FROM procurement_items pi
		JOIN project_tasks pt ON pi.project_task_id = pt.id
		JOIN projects p ON pt.project_id = p.id
		WHERE pi.id = $3
	`
	content := fmt.Sprintf("[PROCUREMENT ALERT] %s", result.Message)
	_, err := a.db.Exec(ctx, query, content, result.ID, result.ID)
	if err != nil {
		slog.Error("failed to log notification", "id", result.ID, "error", err)
	} else {
		slog.Info("Notification logged", "item_id", result.ID, "message", result.Message)
	}
}
