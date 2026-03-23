package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/colton/futurebuild/internal/data"
	"github.com/colton/futurebuild/pkg/ai"
)

// RegisterMarketTools registers market conditions and seasonal cost tools.
func RegisterMarketTools(r *Registry) {
	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "get_market_conditions",
			Description: "Get seasonal construction cost forecast and labor availability for a region and start date. Shows optimal start windows and material cost trends.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"region": {
						"type": "string",
						"description": "Regional cost region (e.g., 'TX-Austin', 'CA-Bay Area')"
					},
					"start_date": {
						"type": "string",
						"description": "Project start date in YYYY-MM-DD format"
					}
				},
				"required": ["start_date"]
			}`),
		},
		Handler: func(ctx context.Context, input json.RawMessage) (string, error) {
			var params struct {
				Region    string `json:"region"`
				StartDate string `json:"start_date"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return "", fmt.Errorf("parse input: %w", err)
			}

			startDate, err := time.Parse("2006-01-02", params.StartDate)
			if err != nil {
				return "", fmt.Errorf("invalid start_date: %w", err)
			}
			startMonth := int(startDate.Month())

			// Calculate seasonal cost factor for starting month
			currentFactor := data.MonthlySeasonalCostFactor(startMonth)

			// Find optimal start month (lowest cost factor)
			bestMonth := 1
			bestFactor := 2.0
			monthlyFactors := make(map[int]float64)
			for m := 1; m <= 12; m++ {
				f := data.MonthlySeasonalCostFactor(m)
				monthlyFactors[m] = f
				if f < bestFactor {
					bestFactor = f
					bestMonth = m
				}
			}

			// Material cost breakdown for starting month
			materialCosts := map[string]float64{
				"lumber":   data.GetSeasonalMaterialIndex("lumber", startMonth),
				"concrete": data.GetSeasonalMaterialIndex("concrete", startMonth),
				"steel":    data.GetSeasonalMaterialIndex("steel", startMonth),
				"hvac":     data.GetSeasonalMaterialIndex("hvac", startMonth),
				"drywall":  data.GetSeasonalMaterialIndex("drywall", startMonth),
			}

			// Labor availability for key trades
			laborAvailability := map[string]float64{
				"general":     data.GetLaborAvailabilityIndex("general", startMonth),
				"electrician": data.GetLaborAvailabilityIndex("electrician", startMonth),
				"plumber":     data.GetLaborAvailabilityIndex("plumber", startMonth),
				"framer":      data.GetLaborAvailabilityIndex("framer", startMonth),
				"hvac":        data.GetLaborAvailabilityIndex("hvac", startMonth),
			}

			// Calculate savings vs best month
			savingsPercent := (currentFactor - bestFactor) / currentFactor * 100

			result := map[string]interface{}{
				"start_month":          startDate.Month().String(),
				"seasonal_cost_factor": fmt.Sprintf("%.3f", currentFactor),
				"material_indices":     materialCosts,
				"labor_availability":   laborAvailability,
				"optimal_start_month":  time.Month(bestMonth).String(),
				"optimal_cost_factor":  fmt.Sprintf("%.3f", bestFactor),
				"potential_savings_pct": fmt.Sprintf("%.1f%%", savingsPercent),
			}

			if params.Region != "" {
				if mult, ok := data.RegionalMultipliers()[params.Region]; ok {
					result["regional_multiplier"] = mult
				}
			}

			b, _ := json.Marshal(result)
			return string(b), nil
		},
	})
}
