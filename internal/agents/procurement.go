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

// NotificationEnqueuer defines the interface for async notification delivery.
// P1 Performance Fix: Enables sidecar pattern for procurement notifications.
type NotificationEnqueuer interface {
	EnqueueNotification(ctx context.Context, itemID uuid.UUID, message string, ts time.Time) error
}

// ProcurementAgent monitors long-lead items and calculates order dates.
// See PRODUCTION_PLAN.md Step 46, BACKEND_SCOPE.md Section 2.5
// Refactored for deterministic simulation: PRODUCTION_PLAN.md Step 49
// P1 Performance Fix: Uses batching and async notifications to reduce DB round-trips
type ProcurementAgent struct {
	repo      ProcurementRepository
	weather   types.WeatherService
	clock     clock.Clock
	notifier  NotificationEnqueuer
	batchSize int
}

// DefaultBatchSize is the number of items to batch before flushing.
// Tunable based on workload (100-500 recommended for production).
const DefaultBatchSize = 100

// NewProcurementAgent creates a new agent instance with repository abstraction.
// Clock is required for deterministic time simulation (Step 49).
// notifier is optional - if nil, notifications are logged but not queued.
// See PRODUCTION_PLAN.md: Testing Strategy remediation (Repository Pattern).
func NewProcurementAgent(repo ProcurementRepository, weather types.WeatherService, clk clock.Clock) *ProcurementAgent {
	return &ProcurementAgent{
		repo:      repo,
		weather:   weather,
		clock:     clk,
		notifier:  nil, // Default: no async notifications
		batchSize: DefaultBatchSize,
	}
}

// NewProcurementAgentWithDB is a convenience constructor using the default PostgreSQL repository.
// Maintains backward compatibility with existing callers.
func NewProcurementAgentWithDB(db *pgxpool.Pool, weather types.WeatherService, clk clock.Clock) *ProcurementAgent {
	return NewProcurementAgent(NewPgProcurementRepository(db), weather, clk)
}

// WithNotificationEnqueuer sets the notification enqueuer for async delivery.
// P1 Performance Fix: Enables sidecar pattern.
func (a *ProcurementAgent) WithNotificationEnqueuer(notifier NotificationEnqueuer) *ProcurementAgent {
	a.notifier = notifier
	return a
}

// WithBatchSize sets a custom batch size.
func (a *ProcurementAgent) WithBatchSize(size int) *ProcurementAgent {
	if size > 0 {
		a.batchSize = size
	}
	return a
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
	// ZipCode for weather lookup - nil indicates missing location data (requires ConfigurationError)
	ZipCode       *string
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
// P1 Performance Fix: Uses batched updates and async notifications to reduce DB round-trips.
func (a *ProcurementAgent) Execute(ctx context.Context) error {
	slog.Info("Starting Procurement Agent...")

	// NOTE: Hydration is now event-driven via HandleHydrateProject task.
	// See implementation_plan.md: "Event-Driven Hydration"

	now := a.clock.Now().Truncate(24 * time.Hour)
	var processed int
	batch := make([]alertResult, 0, a.batchSize)

	// Stream items one-by-one to avoid unbounded memory allocation
	err := a.streamItems(ctx, func(item procurementRow) error {
		result := a.analyzeItem(item, now)
		batch = append(batch, result)

		// P1 Performance Fix: Flush batch when full
		if len(batch) >= a.batchSize {
			if err := a.flushBatch(ctx, batch); err != nil {
				slog.Error("failed to flush batch", "error", err)
				// Continue processing - don't fail entire run
			}
			processed += len(batch)
			batch = batch[:0] // Reset slice, keep capacity
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("streaming items failed: %w", err)
	}

	// Flush remaining items
	if len(batch) > 0 {
		if err := a.flushBatch(ctx, batch); err != nil {
			slog.Error("failed to flush final batch", "error", err)
		}
		processed += len(batch)
	}

	slog.Info("Procurement Agent completed", "items_processed", processed)
	return nil
}

// HydrateProject populates procurement_items for a specific project.
// Called via event-driven task (TypeHydrateProject) on project creation.
// P1 Performance Fix: Replaces cron-swept hydrateItems with project-scoped operation.
// See User Amendment #2, Critical Blocker A.3 Remediation
func (a *ProcurementAgent) HydrateProject(ctx context.Context, projectID uuid.UUID) error {
	return a.repo.HydrateProject(ctx, projectID)
}

// flushBatch sends all UPDATE queries via repository.
// P1 Performance Fix: Reduces N database round-trips to 1 per batch.
// Notifications are enqueued asynchronously via NotificationEnqueuer (sidecar pattern).
func (a *ProcurementAgent) flushBatch(ctx context.Context, batch []alertResult) error {
	if len(batch) == 0 {
		return nil
	}

	now := a.clock.Now()

	// Delegate batch update to repository
	if err := a.repo.UpdateBatch(ctx, now, batch); err != nil {
		return err
	}

	// Process notifications: log to communication_logs and optionally enqueue async
	for _, result := range batch {
		if result.ShouldNotify {
			// Check dampening window (72 hours) before sending
			shouldSend, err := a.shouldSendNotification(ctx, result.ID)
			if err != nil {
				slog.Error("failed to check notification dampening", "id", result.ID, "error", err)
				continue
			}
			if !shouldSend {
				slog.Debug("notification dampened", "id", result.ID)
				continue
			}

			// Log to communication_logs for audit trail
			a.logNotification(ctx, result)

			// Enqueue async notification if sidecar is configured
			if a.notifier != nil {
				if err := a.notifier.EnqueueNotification(ctx, result.ID, result.Message, now); err != nil {
					slog.Error("failed to enqueue notification", "id", result.ID, "error", err)
					// Continue - notifications are best-effort
				}
			}
		}
	}

	slog.Debug("batch flushed", "count", len(batch))
	return nil
}

// streamItems iterates through procurement items via repository callback.
// P1 Scalability Fix: Uses cursor-based iteration to prevent OOM at scale.
// See User Amendment #3, BACKEND_SCOPE.md Section 2.5
func (a *ProcurementAgent) streamItems(ctx context.Context, process ItemProcessor) error {
	return a.repo.StreamItems(ctx, process)
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

	// FAANG Standard: "Fail Loudly" for Location Data
	// If location data is missing, schedule calculation is BLOCKED (ConfigError).
	// No data is better than wrong data - an Alaskan project using Texas weather
	// could cause weeks of schedule drift and financial damages.
	// See User Amendment: Data Integrity & "Physics" Accuracy
	if item.ZipCode == nil || *item.ZipCode == "" {
		result.NewStatus = types.ProcurementAlertConfigError
		result.ShouldNotify = true
		result.Message = fmt.Sprintf("⚠️ CONFIGURATION REQUIRED: %s is missing location data. "+
			"Please add a zip code to the project to enable accurate schedule calculation.", item.Name)
		slog.Warn("config error: missing zip code for procurement item",
			"item_id", item.ID,
			"item_name", item.Name,
			"action", "blocking_schedule_calculation",
		)
		return result
	}

	// Weather interaction (SWIM integration)
	// See PRODUCTION_PLAN.md Step 46: Weather Buffer
	// For MVP, we check if precipitation probability > 50% near the start date
	// TODO: Inject GeocodingService and use item.ZipCode for location-specific weather.
	if a.weather != nil {
		// ZipCode is available but geocoding service is not yet wired.
		// Log a metric indicating geocoding is needed for accurate weather data.
		// FAIL SAFE: Skip weather buffer calculation until geocoding is implemented.
		slog.Warn("weather buffer skipped: geocoding not implemented",
			"item_id", item.ID,
			"zip_code", *item.ZipCode,
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

// shouldSendNotification checks communication_logs for recent alerts via repository.
// See User Amendment #4: 72-hour dampening, Optimized via Migration 000046
// See PRODUCTION_PLAN.md Step 49: Uses injected clock for deterministic simulation.
func (a *ProcurementAgent) shouldSendNotification(ctx context.Context, itemID uuid.UUID) (bool, error) {
	return a.repo.ShouldSendNotification(ctx, itemID, a.clock.Now())
}

// logNotification persists the alert to communication_logs via repository.
// See PRODUCTION_PLAN.md Step 49: Uses injected clock for deterministic simulation.
func (a *ProcurementAgent) logNotification(ctx context.Context, result alertResult) {
	if err := a.repo.LogNotification(ctx, result, a.clock.Now()); err != nil {
		slog.Error("failed to log notification", "id", result.ID, "error", err)
	}
}
