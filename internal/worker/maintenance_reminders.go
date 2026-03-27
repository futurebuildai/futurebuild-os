package worker

import (
	"context"
	"log/slog"

	"github.com/colton/futurebuild/internal/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// FleetMaintenanceServicer is the subset of FleetServicer needed for maintenance reminders.
type FleetMaintenanceServicer interface {
	GetUpcomingMaintenance(ctx context.Context, orgID uuid.UUID, withinDays int) ([]models.MaintenanceLog, error)
}

// HandleMaintenanceReminders checks for upcoming equipment maintenance.
// Phase 18: See BACKEND_SCOPE.md Section 20.3 (Equipment/Fleet Management)
// Scheduled weekly Monday 07:00 UTC.
func (h *WorkerHandler) HandleMaintenanceReminders(ctx context.Context, _ *asynq.Task) error {
	slog.Info("maintenance_reminders: starting weekly check")

	if h.fleetSvc == nil {
		slog.Warn("maintenance_reminders: service not configured, skipping")
		return nil
	}

	// Get all org IDs that have fleet assets
	rows, err := h.db.Query(ctx, `SELECT DISTINCT org_id FROM fleet_assets WHERE status != 'retired'`)
	if err != nil {
		slog.Error("maintenance_reminders: failed to query org IDs", "error", err)
		return err
	}
	defer rows.Close()

	var orgIDs []uuid.UUID
	for rows.Next() {
		var orgID uuid.UUID
		if err := rows.Scan(&orgID); err != nil {
			continue
		}
		orgIDs = append(orgIDs, orgID)
	}

	var totalUpcoming int
	for _, orgID := range orgIDs {
		logs, err := h.fleetSvc.GetUpcomingMaintenance(ctx, orgID, 14)
		if err != nil {
			slog.Warn("maintenance_reminders: failed for org", "org_id", orgID, "error", err)
			continue
		}
		totalUpcoming += len(logs)
	}

	slog.Info("maintenance_reminders: completed",
		"orgs_checked", len(orgIDs),
		"total_upcoming", totalUpcoming,
	)

	return nil
}
