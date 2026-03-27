package service

import (
	"context"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CorporateFinancialsService implements CorporateFinancialsServicer.
// See BACKEND_SCOPE.md Section 20.1
type CorporateFinancialsService struct {
	db *pgxpool.Pool
}

func NewCorporateFinancialsService(db *pgxpool.Pool) *CorporateFinancialsService {
	return &CorporateFinancialsService{db: db}
}

// RollupCorporateBudget aggregates all project_budget rows for the org into a corporate snapshot.
// Deterministic: pure SQL aggregation, no AI.
func (s *CorporateFinancialsService) RollupCorporateBudget(ctx context.Context, orgID uuid.UUID, fiscalYear, quarter int) (*models.CorporateBudget, error) {
	// Aggregate project-level budgets into org-level totals
	aggQuery := `
		SELECT
			COALESCE(SUM(pb.estimated_amount_cents), 0),
			COALESCE(SUM(pb.committed_amount_cents), 0),
			COALESCE(SUM(pb.actual_amount_cents), 0),
			COUNT(DISTINCT p.id)
		FROM project_budgets pb
		INNER JOIN projects p ON pb.project_id = p.id
		WHERE p.org_id = $1`

	var totalEstimated, totalCommitted, totalActual int64
	var projectCount int
	err := s.db.QueryRow(ctx, aggQuery, orgID).Scan(
		&totalEstimated, &totalCommitted, &totalActual, &projectCount,
	)
	if err != nil {
		return nil, err
	}

	// Upsert into corporate_budgets
	upsertQuery := `
		INSERT INTO corporate_budgets (org_id, fiscal_year, quarter, total_estimated_cents, total_committed_cents, total_actual_cents, project_count, last_rollup_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (org_id, fiscal_year, quarter)
		DO UPDATE SET
			total_estimated_cents = EXCLUDED.total_estimated_cents,
			total_committed_cents = EXCLUDED.total_committed_cents,
			total_actual_cents = EXCLUDED.total_actual_cents,
			project_count = EXCLUDED.project_count,
			last_rollup_at = NOW()
		RETURNING id, org_id, fiscal_year, quarter, total_estimated_cents, total_committed_cents, total_actual_cents, project_count, last_rollup_at, created_at, updated_at`

	var budget models.CorporateBudget
	err = s.db.QueryRow(ctx, upsertQuery, orgID, fiscalYear, quarter, totalEstimated, totalCommitted, totalActual, projectCount).Scan(
		&budget.ID, &budget.OrgID, &budget.FiscalYear, &budget.Quarter,
		&budget.TotalEstimatedCents, &budget.TotalCommittedCents, &budget.TotalActualCents,
		&budget.ProjectCount, &budget.LastRollupAt, &budget.CreatedAt, &budget.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &budget, nil
}

// GetCorporateBudget retrieves a stored corporate budget rollup.
func (s *CorporateFinancialsService) GetCorporateBudget(ctx context.Context, orgID uuid.UUID, fiscalYear, quarter int) (*models.CorporateBudget, error) {
	query := `
		SELECT id, org_id, fiscal_year, quarter, total_estimated_cents, total_committed_cents, total_actual_cents, project_count, last_rollup_at, created_at, updated_at
		FROM corporate_budgets
		WHERE org_id = $1 AND fiscal_year = $2 AND quarter = $3`

	var budget models.CorporateBudget
	err := s.db.QueryRow(ctx, query, orgID, fiscalYear, quarter).Scan(
		&budget.ID, &budget.OrgID, &budget.FiscalYear, &budget.Quarter,
		&budget.TotalEstimatedCents, &budget.TotalCommittedCents, &budget.TotalActualCents,
		&budget.ProjectCount, &budget.LastRollupAt, &budget.CreatedAt, &budget.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &budget, nil
}

// CalculateARAging computes current AR aging from invoices by age bucket.
// Deterministic: pure date math against invoice records.
func (s *CorporateFinancialsService) CalculateARAging(ctx context.Context, orgID uuid.UUID) (*models.ARAgingSnapshot, error) {
	now := time.Now()
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN i.invoice_date >= $2::date - interval '30 days' THEN i.amount_cents ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN i.invoice_date < $2::date - interval '30 days' AND i.invoice_date >= $2::date - interval '60 days' THEN i.amount_cents ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN i.invoice_date < $2::date - interval '60 days' AND i.invoice_date >= $2::date - interval '90 days' THEN i.amount_cents ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN i.invoice_date < $2::date - interval '90 days' THEN i.amount_cents ELSE 0 END), 0),
			COALESCE(SUM(i.amount_cents), 0)
		FROM invoices i
		INNER JOIN projects p ON i.project_id = p.id
		WHERE p.org_id = $1 AND i.status IN ('Draft', 'Pending')`

	var currentCents, days30, days60, days90Plus, total int64
	err := s.db.QueryRow(ctx, query, orgID, now).Scan(
		&currentCents, &days30, &days60, &days90Plus, &total,
	)
	if err != nil {
		return nil, err
	}

	// Upsert the snapshot
	upsertQuery := `
		INSERT INTO ar_aging_snapshots (org_id, snapshot_date, current_cents, days_30_cents, days_60_cents, days_90_plus_cents, total_receivable_cents)
		VALUES ($1, $2::date, $3, $4, $5, $6, $7)
		ON CONFLICT (org_id, snapshot_date)
		DO UPDATE SET
			current_cents = EXCLUDED.current_cents,
			days_30_cents = EXCLUDED.days_30_cents,
			days_60_cents = EXCLUDED.days_60_cents,
			days_90_plus_cents = EXCLUDED.days_90_plus_cents,
			total_receivable_cents = EXCLUDED.total_receivable_cents
		RETURNING id, org_id, snapshot_date, current_cents, days_30_cents, days_60_cents, days_90_plus_cents, total_receivable_cents, created_at`

	var snap models.ARAgingSnapshot
	err = s.db.QueryRow(ctx, upsertQuery, orgID, now, currentCents, days30, days60, days90Plus, total).Scan(
		&snap.ID, &snap.OrgID, &snap.SnapshotDate,
		&snap.CurrentCents, &snap.Days30Cents, &snap.Days60Cents, &snap.Days90PlusCents,
		&snap.TotalReceivableCents, &snap.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &snap, nil
}

// CreateGLSyncLog records a new GL sync operation for audit trail.
func (s *CorporateFinancialsService) CreateGLSyncLog(ctx context.Context, orgID uuid.UUID, syncType string) (*models.GLSyncLog, error) {
	query := `
		INSERT INTO gl_sync_logs (org_id, sync_type, status)
		VALUES ($1, $2, 'pending')
		RETURNING id, org_id, sync_type, status, records_synced, error_message, synced_at, created_at`

	var log models.GLSyncLog
	err := s.db.QueryRow(ctx, query, orgID, syncType).Scan(
		&log.ID, &log.OrgID, &log.SyncType, &log.Status,
		&log.RecordsSynced, &log.ErrorMessage, &log.SyncedAt, &log.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &log, nil
}

// ListGLSyncLogs returns recent GL sync operations for the org.
func (s *CorporateFinancialsService) ListGLSyncLogs(ctx context.Context, orgID uuid.UUID) ([]models.GLSyncLog, error) {
	query := `
		SELECT id, org_id, sync_type, status, records_synced, error_message, synced_at, created_at
		FROM gl_sync_logs
		WHERE org_id = $1
		ORDER BY created_at DESC
		LIMIT 100`

	rows, err := s.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.GLSyncLog
	for rows.Next() {
		var l models.GLSyncLog
		if err := rows.Scan(&l.ID, &l.OrgID, &l.SyncType, &l.Status,
			&l.RecordsSynced, &l.ErrorMessage, &l.SyncedAt, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	return logs, nil
}
