package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/data"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/prompts"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MaterialService handles material extraction, quantity estimation, and persistence.
// Follows the InvoiceService constructor pattern.
type MaterialService struct {
	db     *pgxpool.Pool
	client ai.Client
	cfg    *config.Config
}

// NewMaterialService creates a new material service.
func NewMaterialService(db *pgxpool.Pool, client ai.Client, cfg *config.Config) *MaterialService {
	return &MaterialService{db: db, client: client, cfg: cfg}
}

// aiMaterialExtraction is the raw JSON structure returned by the AI model.
type aiMaterialExtraction struct {
	Materials []struct {
		Name         string   `json:"name"`
		Category     string   `json:"category"`
		WBSPhaseCode string   `json:"wbs_phase_code"`
		Quantity     *float64 `json:"quantity"` // nullable
		Unit         string   `json:"unit"`
		Brand        *string  `json:"brand"`
		Model        *string  `json:"model"`
		Notes        string   `json:"notes"`
		Confidence   float64  `json:"confidence"`
	} `json:"materials"`
}

// ExtractMaterials runs AI material extraction on blueprint/document data.
// Returns extracted materials with quantity estimates. Uses Gemini Flash for vision.
func (s *MaterialService) ExtractMaterials(ctx context.Context, imageData []byte, mimeType string) ([]models.MaterialEstimate, error) {
	if s.client == nil {
		return nil, fmt.Errorf("AI client not configured")
	}

	req := ai.NewMultimodalRequest(
		ai.ModelTypeFlashPreview,
		prompts.MaterialExtractionPrompt(),
		imageData,
		mimeType,
	)
	req.Temperature = 0.1 // Low temperature for factual extraction
	req.ReturnLogprobs = true

	resp, err := s.client.GenerateContent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("material extraction AI call failed: %w", err)
	}

	// Clean markdown code fences from response
	text := strings.TrimSpace(resp.Text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var extraction aiMaterialExtraction
	if err := json.Unmarshal([]byte(text), &extraction); err != nil {
		return nil, fmt.Errorf("failed to parse material extraction: %w", err)
	}

	estimates := make([]models.MaterialEstimate, 0, len(extraction.Materials))
	for _, m := range extraction.Materials {
		if m.Name == "" {
			continue
		}

		qty := 0.0
		if m.Quantity != nil {
			qty = *m.Quantity
		}

		// Validate and default category
		category := normalizeCategory(m.Category)
		wbsPhase := m.WBSPhaseCode
		if wbsPhase == "" {
			wbsPhase = categoryToWBSPhase(category)
		}

		// Default unit
		unit := m.Unit
		if unit == "" {
			unit = "ea"
		}

		est := models.MaterialEstimate{
			Name:           m.Name,
			Category:       category,
			WBSPhaseCode:   wbsPhase,
			Quantity:       qty,
			Unit:           unit,
			UnitCostCents:  0, // AI doesn't estimate costs; filled by enrichment
			TotalCostCents: 0,
			Confidence:     m.Confidence,
			Source:         "ai",
		}

		estimates = append(estimates, est)
	}

	slog.Info("material_extraction_complete",
		"item_count", len(estimates),
		"ai_confidence", resp.Confidence,
	)

	return estimates, nil
}

// EstimateFromProjectAttributes generates material estimates using heuristic formulas
// when no blueprint is available. Uses cost indices for pricing.
func (s *MaterialService) EstimateFromProjectAttributes(
	_ context.Context,
	gsf float64, stories int, foundationType string,
	bedrooms int, bathrooms int, region string,
) ([]models.MaterialEstimate, error) {
	if gsf <= 0 {
		return nil, fmt.Errorf("invalid GSF: %.0f", gsf)
	}

	quantities := data.EstimateQuantities(gsf, stories, foundationType, bedrooms, bathrooms)

	// Apply regional multiplier to costs
	multiplier := 1.0
	if region != "" {
		if m, ok := data.RegionalMultipliers()[region]; ok {
			multiplier = m
		}
	}

	estimates := make([]models.MaterialEstimate, 0, len(quantities))
	for _, q := range quantities {
		unitCostCents := int64(math.Round(float64(q.UnitCostCents) * multiplier))
		totalCostCents := int64(math.Round(q.Quantity * float64(unitCostCents)))

		estimates = append(estimates, models.MaterialEstimate{
			Name:           q.MaterialName,
			Category:       q.Category,
			WBSPhaseCode:   q.WBSPhaseCode,
			Quantity:       q.Quantity,
			Unit:           q.Unit,
			UnitCostCents:  unitCostCents,
			TotalCostCents: totalCostCents,
			Confidence:     q.Confidence,
			Source:         "default",
		})
	}

	return estimates, nil
}

// SaveMaterials persists material estimates to the project_materials table.
// Uses atomic upsert (INSERT ON CONFLICT) to prevent duplicates.
func (s *MaterialService) SaveMaterials(ctx context.Context, projectID uuid.UUID, materials []models.MaterialEstimate) error {
	if len(materials) == 0 {
		return nil
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, m := range materials {
		totalCents := m.TotalCostCents
		if totalCents == 0 && m.Quantity > 0 && m.UnitCostCents > 0 {
			totalCents = int64(math.Round(m.Quantity * float64(m.UnitCostCents)))
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO project_materials (
				project_id, wbs_phase_code, name, category, quantity, unit,
				unit_cost_cents, total_cost_cents, source, confidence,
				brand, model, notes
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			ON CONFLICT (project_id, wbs_phase_code, name)
			DO UPDATE SET
				quantity = EXCLUDED.quantity,
				unit = EXCLUDED.unit,
				unit_cost_cents = EXCLUDED.unit_cost_cents,
				total_cost_cents = EXCLUDED.total_cost_cents,
				confidence = EXCLUDED.confidence,
				updated_at = NOW()
			WHERE project_materials.source != 'user'`,
			projectID, m.WBSPhaseCode, m.Name, m.Category, m.Quantity, m.Unit,
			m.UnitCostCents, totalCents, m.Source, m.Confidence,
			"", "", "", // brand, model, notes
		)
		if err != nil {
			return fmt.Errorf("failed to upsert material %q: %w", m.Name, err)
		}
	}

	return tx.Commit(ctx)
}

// ListMaterials returns all materials for a project with multi-tenancy enforcement.
func (s *MaterialService) ListMaterials(ctx context.Context, projectID, orgID uuid.UUID) ([]models.ProjectMaterial, error) {
	rows, err := s.db.Query(ctx, `
		SELECT pm.id, pm.project_id, pm.wbs_phase_code, pm.name, pm.category,
			pm.quantity, pm.unit, pm.unit_cost_cents, pm.total_cost_cents,
			pm.source, pm.confidence, pm.brand, pm.model, pm.sku, pm.notes,
			pm.created_at, pm.updated_at
		FROM project_materials pm
		JOIN projects p ON pm.project_id = p.id
		WHERE pm.project_id = $1 AND p.org_id = $2
		ORDER BY pm.wbs_phase_code, pm.name`, projectID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query materials: %w", err)
	}
	defer rows.Close()

	var materials []models.ProjectMaterial
	for rows.Next() {
		var m models.ProjectMaterial
		if err := rows.Scan(
			&m.ID, &m.ProjectID, &m.WBSPhaseCode, &m.Name, &m.Category,
			&m.Quantity, &m.Unit, &m.UnitCostCents, &m.TotalCostCents,
			&m.Source, &m.Confidence, &m.Brand, &m.Model, &m.SKU, &m.Notes,
			&m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan material: %w", err)
		}
		materials = append(materials, m)
	}

	return materials, rows.Err()
}

// UpdateMaterial modifies a single material entry (user edits).
// Only non-nil fields in the update request are applied.
// Marks the source as 'user' to prevent AI from overwriting.
func (s *MaterialService) UpdateMaterial(ctx context.Context, materialID, orgID uuid.UUID, updates models.MaterialUpdateRequest) (*models.ProjectMaterial, error) {
	// Build dynamic SET clause
	setClauses := []string{"source = 'user'", "updated_at = NOW()"}
	args := []interface{}{materialID, orgID}
	argIdx := 3

	if updates.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *updates.Name)
		argIdx++
	}
	if updates.Category != nil {
		setClauses = append(setClauses, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *updates.Category)
		argIdx++
	}
	if updates.WBSPhaseCode != nil {
		setClauses = append(setClauses, fmt.Sprintf("wbs_phase_code = $%d", argIdx))
		args = append(args, *updates.WBSPhaseCode)
		argIdx++
	}
	if updates.Quantity != nil {
		setClauses = append(setClauses, fmt.Sprintf("quantity = $%d", argIdx))
		args = append(args, *updates.Quantity)
		argIdx++
	}
	if updates.Unit != nil {
		setClauses = append(setClauses, fmt.Sprintf("unit = $%d", argIdx))
		args = append(args, *updates.Unit)
		argIdx++
	}
	if updates.UnitCostCents != nil {
		setClauses = append(setClauses, fmt.Sprintf("unit_cost_cents = $%d", argIdx))
		args = append(args, *updates.UnitCostCents)
		argIdx++

		// Recalculate total if quantity is known
		if updates.Quantity != nil {
			totalCents := int64(math.Round(*updates.Quantity * float64(*updates.UnitCostCents)))
			setClauses = append(setClauses, fmt.Sprintf("total_cost_cents = $%d", argIdx))
			args = append(args, totalCents)
			argIdx++
		}
	}
	if updates.Brand != nil {
		setClauses = append(setClauses, fmt.Sprintf("brand = $%d", argIdx))
		args = append(args, *updates.Brand)
		argIdx++
	}
	if updates.Model != nil {
		setClauses = append(setClauses, fmt.Sprintf("model = $%d", argIdx))
		args = append(args, *updates.Model)
		argIdx++
	}
	if updates.SKU != nil {
		setClauses = append(setClauses, fmt.Sprintf("sku = $%d", argIdx))
		args = append(args, *updates.SKU)
		argIdx++
	}
	if updates.Notes != nil {
		setClauses = append(setClauses, fmt.Sprintf("notes = $%d", argIdx))
		args = append(args, *updates.Notes)
		argIdx++
	}

	query := fmt.Sprintf(`
		UPDATE project_materials pm
		SET %s
		FROM projects p
		WHERE pm.id = $1
			AND pm.project_id = p.id
			AND p.org_id = $2
		RETURNING pm.id, pm.project_id, pm.wbs_phase_code, pm.name, pm.category,
			pm.quantity, pm.unit, pm.unit_cost_cents, pm.total_cost_cents,
			pm.source, pm.confidence, pm.brand, pm.model, pm.sku, pm.notes,
			pm.created_at, pm.updated_at`,
		strings.Join(setClauses, ", "))

	var m models.ProjectMaterial
	err := s.db.QueryRow(ctx, query, args...).Scan(
		&m.ID, &m.ProjectID, &m.WBSPhaseCode, &m.Name, &m.Category,
		&m.Quantity, &m.Unit, &m.UnitCostCents, &m.TotalCostCents,
		&m.Source, &m.Confidence, &m.Brand, &m.Model, &m.SKU, &m.Notes,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, types.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update material: %w", err)
	}

	return &m, nil
}

// DeleteMaterial removes a material entry with multi-tenancy enforcement.
func (s *MaterialService) DeleteMaterial(ctx context.Context, materialID, orgID uuid.UUID) error {
	result, err := s.db.Exec(ctx, `
		DELETE FROM project_materials pm
		USING projects p
		WHERE pm.id = $1
			AND pm.project_id = p.id
			AND p.org_id = $2`, materialID, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete material: %w", err)
	}
	if result.RowsAffected() == 0 {
		return types.ErrNotFound
	}
	return nil
}

// normalizeCategory ensures the category matches expected values.
func normalizeCategory(category string) string {
	valid := map[string]bool{
		"structural": true, "framing": true, "roofing": true, "siding": true,
		"insulation": true, "drywall": true, "flooring": true, "plumbing": true,
		"electrical": true, "hvac": true, "millwork": true, "finishes": true,
		"fixtures": true, "appliances": true,
	}
	lower := strings.ToLower(strings.TrimSpace(category))
	if valid[lower] {
		return lower
	}
	return "finishes" // safe default
}

// categoryToWBSPhase maps material categories to WBS phase codes.
func categoryToWBSPhase(category string) string {
	mapping := map[string]string{
		"structural": "8.x",
		"framing":    "9.x",
		"roofing":    "13.x",
		"siding":     "13.x",
		"insulation": "11.x",
		"drywall":    "11.x",
		"flooring":   "12.x",
		"plumbing":   "10.x",
		"electrical": "10.x",
		"hvac":       "10.x",
		"millwork":   "12.x",
		"finishes":   "12.x",
		"fixtures":   "12.x",
		"appliances": "12.x",
	}
	if code, ok := mapping[category]; ok {
		return code
	}
	return "12.x"
}
