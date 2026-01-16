// Package physics implements the deterministic scheduling algorithms.
// See BACKEND_SCOPE.md Section 3.4 (Layer 3: Physics Engine)
package physics

import (
	"math"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
)

// CalculateSAF computes the Size Adjustment Factor.
// SAF = (GSF / StandardHouseSizeSF) ^ SizeAdjustmentExponent
// See CPM_RES_MODEL_SPEC.md Section 11.2.1
// Config decoupling: Magic numbers now come from PhysicsConfig.
func CalculateSAF(gsf float64, cfg config.PhysicsConfig) float64 {
	if gsf <= 0 {
		return 1.0 // Default for unset GSF
	}
	// Apply defaults for zero-value safety (FAANG Threshold #2)
	cfg = cfg.WithDefaults()
	return math.Pow(gsf/cfg.StandardHouseSizeSF, cfg.SizeAdjustmentExponent)
}

// CalculateTaskDuration computes the DHSM-adjusted task duration.
// Formula: Raw_Duration = (D_base * SAF * Multipliers)
// Final:   Duration_final = Ceiling(Raw_Duration * 2) / 2
// See CPM_RES_MODEL_SPEC.md Section 11.2.5
func CalculateTaskDuration(
	task models.WBSTask,
	gsf float64, // Gross Square Footage from Project
	context models.ProjectContext,
	multipliers []models.DurationMultiplier,
	forecast types.Forecast,
	cfg config.PhysicsConfig, // Config decoupling: accepts tunable physics config
) float64 {
	baseDuration := task.BaseDurationDays

	// Event Duration Locking: Inspections have SAF = 1.0
	// See CPM_RES_MODEL_SPEC.md Section 11.2.1
	saf := 1.0
	if !task.IsInspection {
		saf = CalculateSAF(gsf, cfg)
	}
	baseDuration *= saf

	for _, mult := range multipliers {
		// Match wildcard or specific WBS code
		if mult.WBSTaskCode != "*" && mult.WBSTaskCode != task.Code {
			continue
		}

		variableValue := getContextVariable(context, mult.VariableKey)
		if variableValue == 0 {
			continue // Variable not set, skip multiplier
		}

		adjustment := applyMultiplierFormula(variableValue, mult.Weight, mult.MultiplierFormula)
		baseDuration *= adjustment
	}

	// Apply SWIM Weather Overlay (Step 31 Correction)
	// Passing a placeholder ProjectTask to ApplyWeatherAdjustment
	tempTask := models.ProjectTask{
		WBSCode:            task.Code,
		CalculatedDuration: baseDuration,
	}
	baseDuration = ApplyWeatherAdjustment(tempTask, forecast)

	// Quantize to 0.5-day increments per CPM_RES_MODEL_SPEC.md Section 11.2.5
	// Formula: Duration_final = Ceiling(Raw_Duration * 2) / 2
	return math.Ceil(baseDuration*2) / 2
}

// getContextVariable extracts the variable value from ProjectContext.
func getContextVariable(ctx models.ProjectContext, key string) float64 {
	switch key {
	case "supply_chain_volatility":
		return float64(ctx.SupplyChainVolatility)
	case "rough_inspection_latency":
		return float64(ctx.RoughInspectionLatency)
	case "final_inspection_latency":
		return float64(ctx.FinalInspectionLatency)
	default:
		return 0
	}
}

// applyMultiplierFormula applies the multiplier formula to compute adjustment.
// Supports common formula types per BACKEND_SCOPE.md Section 6.1 examples.
func applyMultiplierFormula(value, weight float64, formula string) float64 {
	switch formula {
	case "linear":
		// Simple linear: 1 + (value * weight)
		return 1 + (value * weight)
	case "scaled":
		// Scaled formula: 1 + (value - baseline) / scale * weight
		// Uses sensible defaults for residential construction
		baseline := 2000.0 // Default baseline (e.g., 2000 sqft)
		scale := 10000.0   // Default scale factor
		return 1 + (value-baseline)/scale*weight
	case "step":
		// Step function: 1 + (value - 1) * weight
		// Used for discrete multipliers (e.g., floor count)
		return 1 + (value-1)*weight
	default:
		// Default to linear if formula not recognized
		return 1 + (value * weight)
	}
}

// DHSMResult represents the output of DHSM calculation for a task.
type DHSMResult struct {
	WBSCode            string  `json:"wbs_code"`
	BaseDuration       float64 `json:"base_duration"`
	CalculatedDuration float64 `json:"calculated_duration"`
}

// CalculateBatchDurations computes DHSM durations for multiple tasks.
func CalculateBatchDurations(
	tasks []models.WBSTask,
	gsf float64, // Gross Square Footage from Project
	context models.ProjectContext,
	multipliers []models.DurationMultiplier,
	forecast types.Forecast,
	cfg config.PhysicsConfig, // Config decoupling: accepts tunable physics config
) []DHSMResult {
	results := make([]DHSMResult, len(tasks))
	for i, task := range tasks {
		results[i] = DHSMResult{
			WBSCode:            task.Code,
			BaseDuration:       task.BaseDurationDays,
			CalculatedDuration: CalculateTaskDuration(task, gsf, context, multipliers, forecast, cfg),
		}
	}
	return results
}
