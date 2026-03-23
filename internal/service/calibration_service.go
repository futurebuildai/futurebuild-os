package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CalibrationEntry records the actual-vs-predicted ratio for a single task.
type CalibrationEntry struct {
	ID            uuid.UUID `json:"id" db:"id"`
	OrgID         uuid.UUID `json:"org_id" db:"org_id"`
	ProjectID     uuid.UUID `json:"project_id" db:"project_id"`
	WBSCode       string    `json:"wbs_code" db:"wbs_code"`
	PredictedDays float64   `json:"predicted_days" db:"predicted_days"`
	ActualDays    float64   `json:"actual_days" db:"actual_days"`
	Ratio         float64   `json:"ratio" db:"ratio"` // actual / predicted
}

// OrgMultiplier is a blended duration multiplier for a specific WBS code.
type OrgMultiplier struct {
	WBSCode    string  `json:"wbs_code"`
	Multiplier float64 `json:"multiplier"`
	SampleSize int     `json:"sample_size"`
}

// CalibrationService manages org-level schedule calibration from historical project data.
// After a project is completed, actual task durations are compared against predicted durations
// to build org-specific multipliers that improve future schedule accuracy.
type CalibrationService struct {
	db *pgxpool.Pool
}

// NewCalibrationService creates a new CalibrationService.
func NewCalibrationService(db *pgxpool.Pool) *CalibrationService {
	return &CalibrationService{db: db}
}

// CalibrateFromCompletion calculates actual/predicted duration ratios for all completed tasks
// in a project and stores them as calibration entries. Should be called after CompleteProject.
func (s *CalibrationService) CalibrateFromCompletion(ctx context.Context, projectID, orgID uuid.UUID) (int, error) {
	// Query completed tasks that have both predicted (calculated_duration_days) and actual durations.
	// actual_days = actual_end - actual_start; predicted_days = calculated_duration_days
	query := `
		SELECT wbs_code, calculated_duration_days,
			EXTRACT(EPOCH FROM (actual_end - actual_start)) / 86400.0 AS actual_days
		FROM project_tasks
		WHERE project_id = $1
			AND status = 'Completed'
			AND actual_start IS NOT NULL
			AND actual_end IS NOT NULL
			AND calculated_duration_days > 0
			AND wbs_code IS NOT NULL
			AND wbs_code != ''
	`

	rows, err := s.db.Query(ctx, query, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to query completed tasks: %w", err)
	}
	defer rows.Close()

	var entries []CalibrationEntry
	for rows.Next() {
		var wbsCode string
		var predictedDays, actualDays float64
		if err := rows.Scan(&wbsCode, &predictedDays, &actualDays); err != nil {
			return 0, fmt.Errorf("scan task row: %w", err)
		}

		if predictedDays <= 0 || actualDays <= 0 {
			continue
		}

		ratio := actualDays / predictedDays
		entries = append(entries, CalibrationEntry{
			OrgID:         orgID,
			ProjectID:     projectID,
			WBSCode:       wbsCode,
			PredictedDays: predictedDays,
			ActualDays:    actualDays,
			Ratio:         ratio,
		})
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate completed tasks: %w", err)
	}

	if len(entries) == 0 {
		slog.Info("calibration: no eligible tasks for calibration",
			"project_id", projectID,
			"org_id", orgID,
		)
		return 0, nil
	}

	// Batch insert calibration entries
	insertQuery := `
		INSERT INTO calibration_entries (org_id, project_id, wbs_code, predicted_days, actual_days, ratio)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, e := range entries {
		_, err := tx.Exec(ctx, insertQuery, e.OrgID, e.ProjectID, e.WBSCode, e.PredictedDays, e.ActualDays, e.Ratio)
		if err != nil {
			return 0, fmt.Errorf("failed to insert calibration entry for %s: %w", e.WBSCode, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit calibration entries: %w", err)
	}

	slog.Info("calibration: stored entries",
		"project_id", projectID,
		"org_id", orgID,
		"count", len(entries),
	)

	return len(entries), nil
}

// GetOrgMultipliers returns blended duration multipliers for an org.
// Blending: 70% org-trained average ratio, 30% global default (1.0).
// Only returns multipliers for WBS codes with at least 2 data points.
func (s *CalibrationService) GetOrgMultipliers(ctx context.Context, orgID uuid.UUID) ([]OrgMultiplier, error) {
	query := `
		SELECT wbs_code, AVG(ratio) AS avg_ratio, COUNT(*) AS sample_size
		FROM calibration_entries
		WHERE org_id = $1 AND created_at > NOW() - INTERVAL '2 years'
		GROUP BY wbs_code
		HAVING COUNT(*) >= 2
		ORDER BY wbs_code
	`

	rows, err := s.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query org multipliers: %w", err)
	}
	defer rows.Close()

	const (
		orgWeight    = 0.7
		globalWeight = 0.3
		globalDefault = 1.0
	)

	var multipliers []OrgMultiplier
	for rows.Next() {
		var wbsCode string
		var avgRatio float64
		var sampleSize int
		if err := rows.Scan(&wbsCode, &avgRatio, &sampleSize); err != nil {
			return nil, fmt.Errorf("scan multiplier row: %w", err)
		}

		// Blend: 70% org-trained + 30% global default
		blended := (orgWeight * avgRatio) + (globalWeight * globalDefault)

		multipliers = append(multipliers, OrgMultiplier{
			WBSCode:    wbsCode,
			Multiplier: blended,
			SampleSize: sampleSize,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate multipliers: %w", err)
	}

	return multipliers, nil
}

// GetOrgMultiplierMap returns multipliers as a map[wbs_code]float64 for easy lookup.
func (s *CalibrationService) GetOrgMultiplierMap(ctx context.Context, orgID uuid.UUID) (map[string]float64, error) {
	multipliers, err := s.GetOrgMultipliers(ctx, orgID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]float64, len(multipliers))
	for _, m := range multipliers {
		result[m.WBSCode] = m.Multiplier
	}
	return result, nil
}
