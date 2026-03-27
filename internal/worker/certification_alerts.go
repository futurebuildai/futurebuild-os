package worker

import (
	"context"
	"log/slog"

	"github.com/colton/futurebuild/internal/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// EmployeeCertServicer is the subset of EmployeeServicer needed for cert alerts.
type EmployeeCertServicer interface {
	GetExpiringCertifications(ctx context.Context, orgID uuid.UUID, withinDays int) ([]models.Certification, error)
}

// HandleCertificationAlerts checks for expiring certifications and updates status.
// Phase 18: See BACKEND_SCOPE.md Section 20.2 (HR/Employee Management)
// Scheduled daily at 08:00 UTC.
func (h *WorkerHandler) HandleCertificationAlerts(ctx context.Context, _ *asynq.Task) error {
	slog.Info("certification_alerts: starting daily check")

	if h.employeeSvc == nil {
		slog.Warn("certification_alerts: service not configured, skipping")
		return nil
	}

	// Get all org IDs that have employees
	rows, err := h.db.Query(ctx, `SELECT DISTINCT org_id FROM employees WHERE status = 'active'`)
	if err != nil {
		slog.Error("certification_alerts: failed to query org IDs", "error", err)
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

	var totalExpiring int
	for _, orgID := range orgIDs {
		certs, err := h.employeeSvc.GetExpiringCertifications(ctx, orgID, 30)
		if err != nil {
			slog.Warn("certification_alerts: failed for org", "org_id", orgID, "error", err)
			continue
		}
		totalExpiring += len(certs)

		// Update cert statuses to 'expiring_soon' where applicable
		for _, cert := range certs {
			if cert.Status != "expiring_soon" {
				_, updateErr := h.db.Exec(ctx,
					`UPDATE certifications SET status = 'expiring_soon', updated_at = NOW() WHERE id = $1`,
					cert.ID,
				)
				if updateErr != nil {
					slog.Warn("certification_alerts: failed to update cert status",
						"cert_id", cert.ID, "error", updateErr)
				}
			}
		}
	}

	slog.Info("certification_alerts: completed",
		"orgs_checked", len(orgIDs),
		"total_expiring", totalExpiring,
	)

	return nil
}
