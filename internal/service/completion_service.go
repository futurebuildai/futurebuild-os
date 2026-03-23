package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CompletionService handles project completion lifecycle operations.
type CompletionService struct {
	db          *pgxpool.Pool
	asynqClient *asynq.Client
}

// NewCompletionService creates a new CompletionService.
func NewCompletionService(db *pgxpool.Pool) *CompletionService {
	return &CompletionService{db: db}
}

// WithAsynqClient injects the asynq client for enqueuing post-completion calibration tasks.
func (s *CompletionService) WithAsynqClient(c *asynq.Client) *CompletionService {
	s.asynqClient = c
	return s
}

// CompleteProject transitions a project to Completed status and generates a CompletionReport.
// Runs within a single transaction to ensure atomicity.
func (s *CompletionService) CompleteProject(ctx context.Context, projectID, orgID, userID uuid.UUID, notes string) (*models.CompletionReport, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 1. Validate project exists, belongs to org, and is Active
	var currentStatus models.ProjectStatus
	err = tx.QueryRow(ctx,
		`SELECT status FROM projects WHERE id = $1 AND org_id = $2 FOR UPDATE`,
		projectID, orgID,
	).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: project %s", types.ErrNotFound, projectID)
		}
		return nil, fmt.Errorf("failed to fetch project: %w", err)
	}
	if currentStatus != models.ProjectStatusActive {
		return nil, fmt.Errorf("%w: project must be Active to complete (current: %s)", types.ErrConflict, currentStatus)
	}

	// 2. Aggregate schedule data
	scheduleSummary, err := s.aggregateSchedule(ctx, tx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate schedule: %w", err)
	}

	// 3. Aggregate budget data
	budgetSummary, err := s.aggregateBudget(ctx, tx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate budget: %w", err)
	}

	// 4. Aggregate weather impact (optional — nil if no data)
	weatherSummary := s.aggregateWeatherImpact(ctx, tx, projectID)

	// 5. Aggregate procurement (optional — nil if no data)
	procurementSummary := s.aggregateProcurement(ctx, tx, projectID)

	// 6. Marshal JSONB columns
	scheduleSummaryJSON, err := json.Marshal(scheduleSummary)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schedule_summary: %w", err)
	}
	budgetSummaryJSON, err := json.Marshal(budgetSummary)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal budget_summary: %w", err)
	}

	var weatherJSON, procurementJSON []byte
	if weatherSummary != nil {
		weatherJSON, _ = json.Marshal(weatherSummary)
	}
	if procurementSummary != nil {
		procurementJSON, _ = json.Marshal(procurementSummary)
	}

	// 7. INSERT completion report
	report := &models.CompletionReport{
		ProjectID:            projectID,
		GeneratedBy:          &userID,
		ScheduleSummary:      *scheduleSummary,
		BudgetSummary:        *budgetSummary,
		WeatherImpactSummary: weatherSummary,
		ProcurementSummary:   procurementSummary,
		Notes:                notes,
	}

	err = tx.QueryRow(ctx,
		`INSERT INTO completion_reports (project_id, generated_by, schedule_summary, budget_summary, weather_impact_summary, procurement_summary, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at`,
		projectID, userID, scheduleSummaryJSON, budgetSummaryJSON, weatherJSON, procurementJSON, notes,
	).Scan(&report.ID, &report.CreatedAt)
	if err != nil {
		// L7: Handle unique constraint violation (concurrent double-complete race)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("%w: project already has a completion report", types.ErrConflict)
		}
		return nil, fmt.Errorf("failed to insert completion report: %w", err)
	}

	// 8. UPDATE project status
	_, err = tx.Exec(ctx,
		`UPDATE projects SET status = $1, completed_at = now(), completed_by = $2 WHERE id = $3`,
		models.ProjectStatusCompleted, userID, projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update project status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("project completed", "project_id", projectID, "completed_by", userID, "report_id", report.ID)

	// Enqueue post-completion calibration to update org-level duration multipliers.
	// Non-blocking: calibration failure should not affect completion.
	if s.asynqClient != nil {
		payload, err := json.Marshal(struct {
			ProjectID uuid.UUID `json:"project_id"`
			OrgID     uuid.UUID `json:"org_id"`
		}{ProjectID: projectID, OrgID: orgID})
		if err == nil {
			task := asynq.NewTask("task:calibrate_on_completion", payload, asynq.Queue("default"))
			if _, err := s.asynqClient.Enqueue(task); err != nil {
				slog.Warn("failed to enqueue calibration task", "project_id", projectID, "error", err)
			}
		}
	}

	return report, nil
}

// GetCompletionReport retrieves the completion report for a project.
// Multi-tenancy enforced via JOIN on projects.org_id.
func (s *CompletionService) GetCompletionReport(ctx context.Context, projectID, orgID uuid.UUID) (*models.CompletionReport, error) {
	query := `
		SELECT cr.id, cr.project_id, cr.generated_by,
			cr.schedule_summary, cr.budget_summary, cr.weather_impact_summary, cr.procurement_summary,
			cr.notes, cr.created_at
		FROM completion_reports cr
		JOIN projects p ON cr.project_id = p.id
		WHERE cr.project_id = $1 AND p.org_id = $2
	`

	var report models.CompletionReport
	var scheduleSummaryJSON, budgetSummaryJSON []byte
	var weatherJSON, procurementJSON []byte

	err := s.db.QueryRow(ctx, query, projectID, orgID).Scan(
		&report.ID, &report.ProjectID, &report.GeneratedBy,
		&scheduleSummaryJSON, &budgetSummaryJSON, &weatherJSON, &procurementJSON,
		&report.Notes, &report.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: completion report for project %s", types.ErrNotFound, projectID)
		}
		return nil, fmt.Errorf("failed to get completion report: %w", err)
	}

	if err := json.Unmarshal(scheduleSummaryJSON, &report.ScheduleSummary); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schedule_summary: %w", err)
	}
	if err := json.Unmarshal(budgetSummaryJSON, &report.BudgetSummary); err != nil {
		return nil, fmt.Errorf("failed to unmarshal budget_summary: %w", err)
	}
	if weatherJSON != nil {
		var ws models.WeatherImpactSummary
		if err := json.Unmarshal(weatherJSON, &ws); err == nil {
			report.WeatherImpactSummary = &ws
		}
	}
	if procurementJSON != nil {
		var ps models.ProcurementSummary
		if err := json.Unmarshal(procurementJSON, &ps); err == nil {
			report.ProcurementSummary = &ps
		}
	}

	return &report, nil
}

// aggregateSchedule gathers task counts and on-time metrics within a transaction.
func (s *CompletionService) aggregateSchedule(ctx context.Context, tx pgx.Tx, projectID uuid.UUID) (*models.ScheduleSummary, error) {
	summary := &models.ScheduleSummary{}

	err := tx.QueryRow(ctx,
		`SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'Completed'),
			COALESCE(SUM(duration_days), 0)::int
		 FROM project_tasks WHERE project_id = $1`,
		projectID,
	).Scan(&summary.TotalTasks, &summary.CompletedTasks, &summary.TotalDurationDays)
	if err != nil {
		return nil, err
	}

	if summary.TotalTasks > 0 {
		// Note: This is task completion rate (completed/total), not late-vs-on-time metric.
		// A true on-time metric requires comparing actual_finish vs early_finish per task.
		summary.OnTimePercent = float64(summary.CompletedTasks) / float64(summary.TotalTasks) * 100
	}

	// Calculate actual duration from project start to now
	var actualDays *int
	_ = tx.QueryRow(ctx,
		`SELECT EXTRACT(DAY FROM now() - MIN(early_start))::int
		 FROM project_tasks WHERE project_id = $1 AND early_start IS NOT NULL`,
		projectID,
	).Scan(&actualDays)
	if actualDays != nil {
		summary.ActualDurationDays = *actualDays
	}

	return summary, nil
}

// aggregateBudget gathers financial metrics within a transaction.
func (s *CompletionService) aggregateBudget(ctx context.Context, tx pgx.Tx, projectID uuid.UUID) (*models.BudgetSummary, error) {
	summary := &models.BudgetSummary{}

	// Sum from invoices table: approved invoices represent actual spend
	err := tx.QueryRow(ctx,
		`SELECT
			COALESCE(SUM(CASE WHEN status = 'Approved' THEN amount_cents ELSE 0 END), 0),
			COALESCE(SUM(amount_cents), 0)
		 FROM invoices WHERE project_id = $1`,
		projectID,
	).Scan(&summary.ActualCents, &summary.CommittedCents)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate invoice data: %w", err)
	}

	// Best-effort: read estimated budget from project_budgets if the table exists.
	// Swallow errors — table may not exist in all deployments.
	_ = tx.QueryRow(ctx,
		`SELECT COALESCE(SUM(estimated_cents), 0)
		 FROM project_budgets WHERE project_id = $1`,
		projectID,
	).Scan(&summary.EstimatedCents)

	summary.VarianceCents = summary.ActualCents - summary.EstimatedCents
	return summary, nil
}

// aggregateWeatherImpact attempts to gather weather delay data. Returns nil if no data.
func (s *CompletionService) aggregateWeatherImpact(ctx context.Context, tx pgx.Tx, projectID uuid.UUID) *models.WeatherImpactSummary {
	var summary models.WeatherImpactSummary
	err := tx.QueryRow(ctx,
		`SELECT COALESCE(SUM(delay_days), 0), COUNT(DISTINCT phase_code)
		 FROM weather_delays WHERE project_id = $1`,
		projectID,
	).Scan(&summary.TotalDelayDays, &summary.PhasesAffected)
	if err != nil || (summary.TotalDelayDays == 0 && summary.PhasesAffected == 0) {
		return nil
	}
	return &summary
}

// aggregateProcurement gathers procurement metrics. Returns nil if no data.
func (s *CompletionService) aggregateProcurement(ctx context.Context, tx pgx.Tx, projectID uuid.UUID) *models.ProcurementSummary {
	var summary models.ProcurementSummary
	err := tx.QueryRow(ctx,
		`SELECT COUNT(*), COALESCE(SUM(cost_cents), 0), COUNT(DISTINCT vendor_name)
		 FROM procurement_items pi
		 JOIN project_tasks pt ON pi.project_task_id = pt.id
		 WHERE pt.project_id = $1`,
		projectID,
	).Scan(&summary.TotalItems, &summary.TotalSpendCents, &summary.VendorCount)
	if err != nil || summary.TotalItems == 0 {
		return nil
	}
	return &summary
}
