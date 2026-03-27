package worker

import (
	"context"
	"log/slog"

	"github.com/colton/futurebuild/internal/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// HandleCorporateRollup iterates all organizations and triggers budget rollup.
// Phase 18: See BACKEND_SCOPE.md Section 20.1 (Corporate Financials)
// Scheduled daily at 23:00 UTC.
func (h *WorkerHandler) HandleCorporateRollup(ctx context.Context, _ *asynq.Task) error {
	slog.Info("corporate_rollup: starting daily rollup")

	if h.corporateFinancialsSvc == nil {
		slog.Warn("corporate_rollup: service not configured, skipping")
		return nil
	}

	// Get all org IDs that have projects
	rows, err := h.db.Query(ctx, `SELECT DISTINCT org_id FROM projects WHERE deleted_at IS NULL`)
	if err != nil {
		slog.Error("corporate_rollup: failed to query org IDs", "error", err)
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

	// Determine current fiscal year and quarter
	now := h.clock.Now()
	fiscalYear := now.Year()
	quarter := (int(now.Month())-1)/3 + 1

	var successCount, errorCount int
	for _, orgID := range orgIDs {
		if _, err := h.corporateFinancialsSvc.RollupCorporateBudget(ctx, orgID, fiscalYear, quarter); err != nil {
			slog.Warn("corporate_rollup: failed for org", "org_id", orgID, "error", err)
			errorCount++
		} else {
			successCount++
		}
	}

	slog.Info("corporate_rollup: completed",
		"orgs_processed", successCount,
		"orgs_failed", errorCount,
		"fiscal_year", fiscalYear,
		"quarter", quarter,
	)

	return nil
}

// CorporateFinancialsServicer is the subset of service.CorporateFinancialsServicer needed by the worker.
type CorporateFinancialsServicer interface {
	RollupCorporateBudget(ctx context.Context, orgID uuid.UUID, fiscalYear, quarter int) (*models.CorporateBudget, error)
}
