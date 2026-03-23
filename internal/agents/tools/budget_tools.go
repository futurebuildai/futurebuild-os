package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/colton/futurebuild/internal/data"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/ai"
)

// RegisterBudgetTools registers budget and cost estimation tools.
func RegisterBudgetTools(r *Registry, budgetSvc service.BudgetServicer) {
	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "get_budget_summary",
			Description: "Get a financial summary of the project including total budget, amount spent, remaining budget, and per-category breakdown.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{},"required":[]}`),
		},
		Handler: func(ctx context.Context, _ json.RawMessage) (string, error) {
			scope := MustGetScope(ctx)
			summary, err := budgetSvc.GetFinancialSummary(ctx, scope.ProjectID, scope.OrgID)
			if err != nil {
				return "", fmt.Errorf("get financial summary: %w", err)
			}
			b, _ := json.Marshal(summary)
			return string(b), nil
		},
	})

	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "estimate_cost_impact",
			Description: "Estimate the cost impact of a project change (e.g., adding square footage, changing scope). Returns estimated cost delta and budget impact.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"description": {
						"type": "string",
						"description": "Description of the change (e.g., 'add 500sqft to second floor')"
					},
					"square_footage_delta": {
						"type": "number",
						"description": "Change in square footage (positive = addition, negative = reduction)"
					},
					"wbs_categories": {
						"type": "array",
						"items": {"type": "string"},
						"description": "Affected WBS phase codes (e.g., ['9.x', '10.x']). Empty = all phases."
					},
					"region": {
						"type": "string",
						"description": "Regional cost region (e.g., 'TX-Austin', 'CA-Bay Area'). Optional."
					}
				},
				"required": ["description", "square_footage_delta"]
			}`),
		},
		Handler: func(ctx context.Context, input json.RawMessage) (string, error) {
			scope := MustGetScope(ctx)

			var params struct {
				Description     string   `json:"description"`
				SqftDelta       float64  `json:"square_footage_delta"`
				WBSCategories   []string `json:"wbs_categories"`
				Region          string   `json:"region"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return "", fmt.Errorf("parse input: %w", err)
			}

			// Calculate cost delta
			deltaCents := data.EstimateCostDelta(params.SqftDelta, params.WBSCategories, params.Region)

			// Get current budget for comparison
			var budgetRemaining int64
			var overBudget bool
			summary, err := budgetSvc.GetFinancialSummary(ctx, scope.ProjectID, scope.OrgID)
			if err == nil {
				budgetRemaining = summary.BudgetTotal - summary.SpendTotal
				overBudget = (budgetRemaining - deltaCents) < 0
			}

			result := map[string]interface{}{
				"description":          params.Description,
				"estimated_cost_cents": deltaCents,
				"estimated_cost_formatted": fmt.Sprintf("$%s", formatCents(deltaCents)),
				"budget_remaining_cents": budgetRemaining,
				"budget_remaining_after": budgetRemaining - deltaCents,
				"over_budget":           overBudget,
			}

			if len(params.WBSCategories) > 0 {
				result["affected_phases"] = params.WBSCategories
			}

			b, _ := json.Marshal(result)
			return string(b), nil
		},
	})
}

// formatCents formats cents as a dollar string (e.g., 123456 → "1,234.56").
func formatCents(cents int64) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}
	dollars := cents / 100
	remainder := cents % 100

	// Format with comma separators
	s := fmt.Sprintf("%d", dollars)
	if len(s) > 3 {
		var result []byte
		for i, c := range s {
			if i > 0 && (len(s)-i)%3 == 0 {
				result = append(result, ',')
			}
			result = append(result, byte(c))
		}
		s = string(result)
	}

	if negative {
		return fmt.Sprintf("-%s.%02d", s, remainder)
	}
	return fmt.Sprintf("%s.%02d", s, remainder)
}
