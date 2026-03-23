package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/data"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BudgetService handles budget seeding, breakdown, and financial summaries.
type BudgetService struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

// NewBudgetService creates a new budget service.
func NewBudgetService(db *pgxpool.Pool, cfg *config.Config) *BudgetService {
	return &BudgetService{db: db, cfg: cfg}
}

// FinancialSummary is the response type for financial overview endpoints.
// Matches the frontend FinancialSummary TypeScript interface (Rosetta Stone).
// MONETARY PRECISION: All values in int64 cents.
type FinancialSummary struct {
	ProjectID   *uuid.UUID                 `json:"project_id,omitempty"`
	BudgetTotal int64                      `json:"budget_total"`
	SpendTotal  int64                      `json:"spend_total"`
	Variance    int64                      `json:"variance"`
	LastUpdated string                     `json:"last_updated"`
	Categories  []FinancialCategorySummary `json:"categories"`
}

// FinancialCategorySummary is a per-phase budget vs spend breakdown.
type FinancialCategorySummary struct {
	Name   string `json:"name"`
	Budget int64  `json:"budget"`
	Spend  int64  `json:"spend"`
	Status string `json:"status"` // on_track, at_risk, over_budget
}

// WBS phase names for budget display.
var wbsPhaseNames = map[string]string{
	"7.x":  "Site Prep",
	"8.x":  "Foundation",
	"9.x":  "Framing",
	"10.x": "Rough-Ins",
	"11.x": "Insulation/Drywall",
	"12.x": "Interior Finishes",
	"13.x": "Exterior",
	"14.x": "Commissioning & Closeout",
	"15.x": "Warranty",
}

// SeedBudget creates initial project_budgets rows from material estimates.
// Groups materials by WBS phase and sums their costs. Applies regional multiplier.
// Uses cost indices for phases not covered by specific materials.
func (s *BudgetService) SeedBudget(
	ctx context.Context,
	projectID uuid.UUID,
	materials []models.MaterialEstimate,
	gsf float64,
	foundationType string,
	stories int,
	regionalMultiplier float64,
) (*models.BudgetEstimate, error) {
	if regionalMultiplier <= 0 {
		regionalMultiplier = 1.0
	}

	// Group material costs by WBS phase
	phaseMaterialCosts := make(map[string]int64)
	phaseConfidences := make(map[string][]float64)
	for _, m := range materials {
		phaseMaterialCosts[m.WBSPhaseCode] += m.TotalCostCents
		phaseConfidences[m.WBSPhaseCode] = append(phaseConfidences[m.WBSPhaseCode], m.Confidence)
	}

	// Get cost indices for phases not covered by material estimates
	costIndices := data.NationalCostIndices()
	foundationAdj := data.FoundationCostAdjustment(foundationType)
	storiesAdj := data.StoriesCostAdjustment(stories)
	if gsf <= 0 {
		gsf = 2250 // Default GSF if not provided
	}

	var phaseEstimates []models.PhaseBudgetEstimate
	var totalEstimated int64

	for _, idx := range costIndices {
		materialsCents := phaseMaterialCosts[idx.WBSPhaseCode]

		// Calculate index-based estimate
		baseEstimate := int64(math.Round(float64(idx.CostPerSqFtCents) * gsf * regionalMultiplier))

		// Apply foundation adjustment to phase 8.x
		if idx.WBSPhaseCode == "8.x" {
			baseEstimate = int64(math.Round(float64(baseEstimate) * foundationAdj))
		}

		// Apply stories adjustment
		baseEstimate = int64(math.Round(float64(baseEstimate) * storiesAdj))

		// Use material costs if available and higher confidence,
		// otherwise use cost index estimate
		estimatedCents := baseEstimate
		confidence := 0.40 // Default confidence for index-based estimates
		source := models.MaterialSourceDefault

		if materialsCents > 0 {
			// Materials cover the material portion; add estimated labor
			laborCents := int64(math.Round(float64(baseEstimate) * idx.LaborSharePct))
			estimatedCents = materialsCents + laborCents
			confidence = avgFloat64(phaseConfidences[idx.WBSPhaseCode])
			source = models.MaterialSourceAI
		}

		materialsCentsForPhase := int64(math.Round(float64(estimatedCents) * (1 - idx.LaborSharePct)))
		laborCents := estimatedCents - materialsCentsForPhase

		phase := models.PhaseBudgetEstimate{
			WBSPhaseCode:   idx.WBSPhaseCode,
			PhaseName:      idx.PhaseName,
			EstimatedCents: estimatedCents,
			MaterialsCents: materialsCentsForPhase,
			LaborCents:     laborCents,
			Confidence:     confidence,
		}

		phaseEstimates = append(phaseEstimates, phase)
		totalEstimated += estimatedCents

		// Upsert to project_budgets
		_, err := s.db.Exec(ctx, `
			INSERT INTO project_budgets (
				id, project_id, wbs_phase_id, estimated_amount_cents,
				committed_amount_cents, actual_amount_cents,
				source, confidence
			) VALUES ($1, $2, $3, $4, 0, 0, $5, $6)
			ON CONFLICT (project_id, wbs_phase_id)
			DO UPDATE SET
				estimated_amount_cents = EXCLUDED.estimated_amount_cents,
				source = EXCLUDED.source,
				confidence = EXCLUDED.confidence
			WHERE NOT project_budgets.is_locked`,
			uuid.New(), projectID, idx.WBSPhaseCode,
			estimatedCents, source, confidence,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert budget for phase %s: %w", idx.WBSPhaseCode, err)
		}
	}

	overallConfidence := 0.0
	for _, p := range phaseEstimates {
		overallConfidence += p.Confidence
	}
	if len(phaseEstimates) > 0 {
		overallConfidence /= float64(len(phaseEstimates))
	}

	estimate := &models.BudgetEstimate{
		TotalEstimatedCents: totalEstimated,
		PhaseBreakdown:      phaseEstimates,
		RegionalMultiplier:  regionalMultiplier,
		ConfidenceOverall:   overallConfidence,
	}

	slog.Info("budget_seeded",
		"project_id", projectID,
		"total_cents", totalEstimated,
		"phases", len(phaseEstimates),
		"confidence", overallConfidence,
	)

	return estimate, nil
}

// ComputeBudgetEstimate generates a BudgetEstimate without persisting it.
// Used during onboarding to show estimates before project creation.
func (s *BudgetService) ComputeBudgetEstimate(
	materials []models.MaterialEstimate,
	gsf float64,
	foundationType string,
	stories int,
	regionalMultiplier float64,
) *models.BudgetEstimate {
	if regionalMultiplier <= 0 {
		regionalMultiplier = 1.0
	}
	if gsf <= 0 {
		gsf = 2250
	}

	phaseMaterialCosts := make(map[string]int64)
	phaseConfidences := make(map[string][]float64)
	for _, m := range materials {
		phaseMaterialCosts[m.WBSPhaseCode] += m.TotalCostCents
		phaseConfidences[m.WBSPhaseCode] = append(phaseConfidences[m.WBSPhaseCode], m.Confidence)
	}

	costIndices := data.NationalCostIndices()
	foundationAdj := data.FoundationCostAdjustment(foundationType)
	storiesAdj := data.StoriesCostAdjustment(stories)

	var phases []models.PhaseBudgetEstimate
	var total int64

	for _, idx := range costIndices {
		materialsCents := phaseMaterialCosts[idx.WBSPhaseCode]
		baseEstimate := int64(math.Round(float64(idx.CostPerSqFtCents) * gsf * regionalMultiplier))
		if idx.WBSPhaseCode == "8.x" {
			baseEstimate = int64(math.Round(float64(baseEstimate) * foundationAdj))
		}
		baseEstimate = int64(math.Round(float64(baseEstimate) * storiesAdj))

		estimatedCents := baseEstimate
		confidence := 0.40

		if materialsCents > 0 {
			laborCents := int64(math.Round(float64(baseEstimate) * idx.LaborSharePct))
			estimatedCents = materialsCents + laborCents
			confidence = avgFloat64(phaseConfidences[idx.WBSPhaseCode])
		}

		matPortion := int64(math.Round(float64(estimatedCents) * (1 - idx.LaborSharePct)))
		laborPortion := estimatedCents - matPortion

		phases = append(phases, models.PhaseBudgetEstimate{
			WBSPhaseCode:   idx.WBSPhaseCode,
			PhaseName:      idx.PhaseName,
			EstimatedCents: estimatedCents,
			MaterialsCents: matPortion,
			LaborCents:     laborPortion,
			Confidence:     confidence,
		})
		total += estimatedCents
	}

	overallConf := 0.0
	for _, p := range phases {
		overallConf += p.Confidence
	}
	if len(phases) > 0 {
		overallConf /= float64(len(phases))
	}

	return &models.BudgetEstimate{
		TotalEstimatedCents: total,
		PhaseBreakdown:      phases,
		RegionalMultiplier:  regionalMultiplier,
		ConfidenceOverall:   overallConf,
	}
}

// GetBudgetBreakdown returns all budget phase rows for a project.
func (s *BudgetService) GetBudgetBreakdown(ctx context.Context, projectID, orgID uuid.UUID) ([]models.ProjectBudget, error) {
	rows, err := s.db.Query(ctx, `
		SELECT pb.id, pb.project_id, pb.wbs_phase_id,
			pb.estimated_amount_cents, pb.committed_amount_cents, pb.actual_amount_cents,
			pb.source, pb.confidence, pb.is_locked
		FROM project_budgets pb
		JOIN projects p ON pb.project_id = p.id
		WHERE pb.project_id = $1 AND p.org_id = $2
		ORDER BY pb.wbs_phase_id`, projectID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query budgets: %w", err)
	}
	defer rows.Close()

	var budgets []models.ProjectBudget
	for rows.Next() {
		var b models.ProjectBudget
		if err := rows.Scan(
			&b.ID, &b.ProjectID, &b.WBSPhaseID,
			&b.EstimatedAmountCents, &b.CommittedAmountCents, &b.ActualAmountCents,
			&b.Source, &b.Confidence, &b.IsLocked,
		); err != nil {
			return nil, fmt.Errorf("failed to scan budget: %w", err)
		}
		budgets = append(budgets, b)
	}
	return budgets, rows.Err()
}

// UpdateBudgetPhase allows a user to override a phase's estimated amount.
// Marks the budget as user-sourced and locked to prevent AI overwrite.
func (s *BudgetService) UpdateBudgetPhase(ctx context.Context, budgetID, orgID uuid.UUID, estimatedCents int64) (*models.ProjectBudget, error) {
	var b models.ProjectBudget
	err := s.db.QueryRow(ctx, `
		UPDATE project_budgets pb
		SET estimated_amount_cents = $3,
			source = 'user',
			is_locked = TRUE
		FROM projects p
		WHERE pb.id = $1
			AND pb.project_id = p.id
			AND p.org_id = $2
		RETURNING pb.id, pb.project_id, pb.wbs_phase_id,
			pb.estimated_amount_cents, pb.committed_amount_cents, pb.actual_amount_cents,
			pb.source, pb.confidence, pb.is_locked`,
		budgetID, orgID, estimatedCents,
	).Scan(
		&b.ID, &b.ProjectID, &b.WBSPhaseID,
		&b.EstimatedAmountCents, &b.CommittedAmountCents, &b.ActualAmountCents,
		&b.Source, &b.Confidence, &b.IsLocked,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, types.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update budget phase: %w", err)
	}
	return &b, nil
}

// GetFinancialSummary computes budget vs spend for a single project.
// Budget from project_budgets, spend from SUM(approved invoices).
func (s *BudgetService) GetFinancialSummary(ctx context.Context, projectID, orgID uuid.UUID) (*FinancialSummary, error) {
	// Verify project access
	var exists bool
	err := s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND org_id = $2)`,
		projectID, orgID).Scan(&exists)
	if err != nil || !exists {
		return nil, types.ErrNotFound
	}

	// Get budget totals by phase
	budgetRows, err := s.db.Query(ctx, `
		SELECT wbs_phase_id, estimated_amount_cents
		FROM project_budgets
		WHERE project_id = $1
		ORDER BY wbs_phase_id`, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query budgets: %w", err)
	}
	defer budgetRows.Close()

	phaseBudgets := make(map[string]int64)
	var budgetTotal int64
	for budgetRows.Next() {
		var phaseID string
		var cents int64
		if err := budgetRows.Scan(&phaseID, &cents); err != nil {
			return nil, fmt.Errorf("failed to scan budget: %w", err)
		}
		phaseBudgets[phaseID] = cents
		budgetTotal += cents
	}
	if err := budgetRows.Err(); err != nil {
		return nil, err
	}

	// Get spend totals by WBS code from approved invoices
	spendRows, err := s.db.Query(ctx, `
		SELECT COALESCE(detected_wbs_code, 'unallocated'), SUM(amount_cents)
		FROM invoices
		WHERE project_id = $1 AND status = 'Approved'
		GROUP BY detected_wbs_code`, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query spend: %w", err)
	}
	defer spendRows.Close()

	phaseSpend := make(map[string]int64)
	var spendTotal int64
	for spendRows.Next() {
		var wbsCode string
		var cents int64
		if err := spendRows.Scan(&wbsCode, &cents); err != nil {
			return nil, fmt.Errorf("failed to scan spend: %w", err)
		}
		// Map specific WBS codes to phase codes
		phaseCode := wbsCodeToPhase(wbsCode)
		phaseSpend[phaseCode] += cents
		spendTotal += cents
	}
	if err := spendRows.Err(); err != nil {
		return nil, err
	}

	// Build category summaries
	var categories []FinancialCategorySummary
	for _, idx := range data.NationalCostIndices() {
		budget := phaseBudgets[idx.WBSPhaseCode]
		spend := phaseSpend[idx.WBSPhaseCode]
		status := "on_track"
		if budget > 0 {
			ratio := float64(spend) / float64(budget)
			if ratio > 1.0 {
				status = "over_budget"
			} else if ratio > 0.85 {
				status = "at_risk"
			}
		}
		categories = append(categories, FinancialCategorySummary{
			Name:   idx.PhaseName,
			Budget: budget,
			Spend:  spend,
			Status: status,
		})
	}

	return &FinancialSummary{
		ProjectID:   &projectID,
		BudgetTotal: budgetTotal,
		SpendTotal:  spendTotal,
		Variance:    budgetTotal - spendTotal,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
		Categories:  categories,
	}, nil
}

// GetGlobalFinancialSummary aggregates financials across all projects for an org.
func (s *BudgetService) GetGlobalFinancialSummary(ctx context.Context, orgID uuid.UUID) (*FinancialSummary, error) {
	var budgetTotal, spendTotal int64

	err := s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(pb.estimated_amount_cents), 0)
		FROM project_budgets pb
		JOIN projects p ON pb.project_id = p.id
		WHERE p.org_id = $1`, orgID).Scan(&budgetTotal)
	if err != nil {
		return nil, fmt.Errorf("failed to query total budget: %w", err)
	}

	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(i.amount_cents), 0)
		FROM invoices i
		JOIN projects p ON i.project_id = p.id
		WHERE p.org_id = $1 AND i.status = 'Approved'`, orgID).Scan(&spendTotal)
	if err != nil {
		return nil, fmt.Errorf("failed to query total spend: %w", err)
	}

	// Per-phase aggregation across all projects
	rows, err := s.db.Query(ctx, `
		SELECT pb.wbs_phase_id,
			SUM(pb.estimated_amount_cents) as budget,
			COALESCE((
				SELECT SUM(i.amount_cents) FROM invoices i
				JOIN projects p2 ON i.project_id = p2.id
				WHERE p2.org_id = $1 AND i.status = 'Approved'
					AND i.detected_wbs_code LIKE pb.wbs_phase_id || '%'
			), 0) as spend
		FROM project_budgets pb
		JOIN projects p ON pb.project_id = p.id
		WHERE p.org_id = $1
		GROUP BY pb.wbs_phase_id
		ORDER BY pb.wbs_phase_id`, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query phase summary: %w", err)
	}
	defer rows.Close()

	var categories []FinancialCategorySummary
	for rows.Next() {
		var phaseID string
		var budget, spend int64
		if err := rows.Scan(&phaseID, &budget, &spend); err != nil {
			return nil, fmt.Errorf("failed to scan phase summary: %w", err)
		}
		status := "on_track"
		if budget > 0 {
			ratio := float64(spend) / float64(budget)
			if ratio > 1.0 {
				status = "over_budget"
			} else if ratio > 0.85 {
				status = "at_risk"
			}
		}
		name := wbsPhaseNames[phaseID]
		if name == "" {
			name = phaseID
		}
		categories = append(categories, FinancialCategorySummary{
			Name:   name,
			Budget: budget,
			Spend:  spend,
			Status: status,
		})
	}

	return &FinancialSummary{
		BudgetTotal: budgetTotal,
		SpendTotal:  spendTotal,
		Variance:    budgetTotal - spendTotal,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
		Categories:  categories,
	}, nil
}

// wbsCodeToPhase maps a specific WBS code (e.g., "8.3") to its phase (e.g., "8.x").
func wbsCodeToPhase(code string) string {
	parts := fmt.Sprintf("%s", code)
	dotIdx := -1
	for i, c := range parts {
		if c == '.' {
			dotIdx = i
			break
		}
	}
	if dotIdx > 0 {
		return parts[:dotIdx] + ".x"
	}
	return code
}

// avgFloat64 computes the average of a slice of float64 values.
func avgFloat64(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}
