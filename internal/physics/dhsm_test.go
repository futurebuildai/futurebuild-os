package physics

import (
	"math"
	"testing"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

// defaultGSF is the baseline GSF where SAF = 1.0
const defaultGSF = 2250.0

// === SAF CALCULATION TESTS ===

func TestCalculateSAF_Baseline(t *testing.T) {
	// At baseline 2250 GSF, SAF should be 1.0
	saf := CalculateSAF(2250.0)
	if saf != 1.0 {
		t.Errorf("expected SAF=1.0 at baseline, got %v", saf)
	}
}

func TestCalculateSAF_LargerHome(t *testing.T) {
	// At 4500 GSF (2x baseline), SAF should be ~1.68 (2^0.75)
	saf := CalculateSAF(4500.0)
	expected := math.Pow(2.0, 0.75) // ~1.6818
	if math.Abs(saf-expected) > 0.01 {
		t.Errorf("expected SAF≈%.2f for 4500 GSF, got %.4f", expected, saf)
	}
}

func TestCalculateSAF_SmallerHome(t *testing.T) {
	// At 1500 GSF (0.67x baseline), SAF should be < 1.0
	saf := CalculateSAF(1500.0)
	expected := math.Pow(1500.0/2250.0, 0.75) // ~0.742
	if math.Abs(saf-expected) > 0.01 {
		t.Errorf("expected SAF≈%.2f for 1500 GSF, got %.4f", expected, saf)
	}
	if saf >= 1.0 {
		t.Errorf("SAF should be < 1.0 for smaller homes, got %v", saf)
	}
}

func TestCalculateSAF_ZeroGSF(t *testing.T) {
	// Zero or negative GSF should return 1.0 (safe default)
	saf := CalculateSAF(0)
	if saf != 1.0 {
		t.Errorf("expected SAF=1.0 for zero GSF, got %v", saf)
	}
}

// === EVENT DURATION LOCKING TESTS ===

func TestEventDurationLocking_InspectionBypassesSAF(t *testing.T) {
	// Inspection tasks should NOT be scaled by SAF
	task := models.WBSTask{
		ID:               uuid.New(),
		Code:             "8.2",
		Name:             "Footings Inspection",
		BaseDurationDays: 1.0,
		IsInspection:     true, // KEY: This is an inspection
	}
	context := models.ProjectContext{}

	// Use very large GSF - should NOT affect inspection duration
	result := CalculateTaskDuration(task, 10000.0, context, nil, types.Forecast{})

	// Expected: 1.0 (no SAF applied) → quantized to 1.0
	if result != 1.0 {
		t.Errorf("inspection should bypass SAF, expected 1.0, got %v", result)
	}
}

func TestEventDurationLocking_NonInspectionScalesBySAF(t *testing.T) {
	// Regular tasks SHOULD be scaled by SAF
	task := models.WBSTask{
		ID:               uuid.New(),
		Code:             "9.1",
		Name:             "Floor Framing",
		BaseDurationDays: 5.0,
		IsInspection:     false,
	}
	context := models.ProjectContext{}

	// Use 4500 GSF (2x baseline) → SAF ≈ 1.68
	result := CalculateTaskDuration(task, 4500.0, context, nil, types.Forecast{})

	// Expected: 5.0 * 1.68 = 8.4 → quantized to 8.5
	expected := 8.5
	if result != expected {
		t.Errorf("expected %.1f with SAF scaling, got %v", expected, result)
	}
}

// === QUANTIZATION TESTS ===

func TestQuantization_SnapsToHalfDay(t *testing.T) {
	testCases := []struct {
		base     float64
		expected float64
	}{
		{4.2, 4.5},  // 4.2 → ceil(8.4) / 2 = 4.5
		{5.0, 5.0},  // 5.0 → ceil(10) / 2 = 5.0
		{5.1, 5.5},  // 5.1 → ceil(10.2) / 2 = 5.5
		{3.75, 4.0}, // 3.75 → ceil(7.5) / 2 = 4.0
		{1.0, 1.0},  // 1.0 → ceil(2) / 2 = 1.0
		{0.3, 0.5},  // 0.3 → ceil(0.6) / 2 = 0.5
	}

	for _, tc := range testCases {
		task := models.WBSTask{
			ID:               uuid.New(),
			Code:             "test",
			Name:             "Test Task",
			BaseDurationDays: tc.base,
		}

		// Use baseline GSF so SAF = 1.0, no multipliers
		result := CalculateTaskDuration(task, defaultGSF, models.ProjectContext{}, nil, types.Forecast{})

		if result != tc.expected {
			t.Errorf("base=%.2f: expected %.1f, got %.1f", tc.base, tc.expected, result)
		}
	}
}

// === MULTIPLIER TESTS (UPDATED WITH NEW SIGNATURE) ===

func TestCalculateTaskDuration_NoMultipliers(t *testing.T) {
	task := models.WBSTask{
		ID:               uuid.New(),
		Code:             "9.1",
		Name:             "Floor Framing",
		BaseDurationDays: 5.0,
	}
	context := models.ProjectContext{}

	// Use baseline GSF so SAF = 1.0
	result := CalculateTaskDuration(task, defaultGSF, context, nil, types.Forecast{})

	// 5.0 → quantized to 5.0
	if result != 5.0 {
		t.Errorf("expected 5.0, got %v", result)
	}
}

func TestCalculateTaskDuration_SingleLinearMultiplier(t *testing.T) {
	task := models.WBSTask{
		ID:               uuid.New(),
		Code:             "9.1",
		Name:             "Floor Framing",
		BaseDurationDays: 5.0,
	}
	context := models.ProjectContext{
		SupplyChainVolatility: 2, // Value = 2
	}
	multipliers := []models.DurationMultiplier{
		{
			ID:                uuid.New(),
			WBSTaskCode:       "9.1",
			VariableKey:       "supply_chain_volatility",
			Weight:            0.1, // 10% per unit
			MultiplierFormula: "linear",
		},
	}

	// Expected: 5.0 * 1.0(SAF) * (1 + (2 * 0.1)) = 5.0 * 1.2 = 6.0
	result := CalculateTaskDuration(task, defaultGSF, context, multipliers, types.Forecast{})

	if result != 6.0 {
		t.Errorf("expected 6.0, got %v", result)
	}
}

func TestCalculateTaskDuration_WildcardMultiplier(t *testing.T) {
	task := models.WBSTask{
		ID:               uuid.New(),
		Code:             "12.4",
		Name:             "Tile Work",
		BaseDurationDays: 3.0,
	}
	context := models.ProjectContext{
		SupplyChainVolatility: 3,
	}
	multipliers := []models.DurationMultiplier{
		{
			ID:                uuid.New(),
			WBSTaskCode:       "*", // Wildcard - applies to all tasks
			VariableKey:       "supply_chain_volatility",
			Weight:            0.2,
			MultiplierFormula: "linear",
		},
	}

	// Expected: 3.0 * (1 + (3 * 0.2)) = 3.0 * 1.6 = 4.8 → quantized to 5.0
	result := CalculateTaskDuration(task, defaultGSF, context, multipliers, types.Forecast{})

	if result != 5.0 {
		t.Errorf("expected 5.0, got %v", result)
	}
}

func TestCalculateTaskDuration_MultipleMultipliers(t *testing.T) {
	task := models.WBSTask{
		ID:               uuid.New(),
		Code:             "9.1",
		Name:             "Floor Framing",
		BaseDurationDays: 5.0,
	}
	context := models.ProjectContext{
		SupplyChainVolatility:  2,
		RoughInspectionLatency: 3,
	}
	multipliers := []models.DurationMultiplier{
		{
			ID:                uuid.New(),
			WBSTaskCode:       "9.1",
			VariableKey:       "supply_chain_volatility",
			Weight:            0.1,
			MultiplierFormula: "linear",
		},
		{
			ID:                uuid.New(),
			WBSTaskCode:       "*",
			VariableKey:       "rough_inspection_latency",
			Weight:            0.05,
			MultiplierFormula: "linear",
		},
	}

	// Expected: 5.0 * (1 + (2 * 0.1)) * (1 + (3 * 0.05))
	//         = 5.0 * 1.2 * 1.15 = 6.9 → quantized to 7.0
	result := CalculateTaskDuration(task, defaultGSF, context, multipliers, types.Forecast{})

	if result != 7.0 {
		t.Errorf("expected 7.0, got %v", result)
	}
}

func TestCalculateTaskDuration_NonMatchingCode(t *testing.T) {
	task := models.WBSTask{
		ID:               uuid.New(),
		Code:             "9.1",
		Name:             "Floor Framing",
		BaseDurationDays: 5.0,
	}
	context := models.ProjectContext{
		SupplyChainVolatility: 2,
	}
	multipliers := []models.DurationMultiplier{
		{
			ID:                uuid.New(),
			WBSTaskCode:       "10.0", // Does not match 9.1
			VariableKey:       "supply_chain_volatility",
			Weight:            0.5,
			MultiplierFormula: "linear",
		},
	}

	// Should return base duration since multiplier doesn't match
	result := CalculateTaskDuration(task, defaultGSF, context, multipliers, types.Forecast{})

	if result != 5.0 {
		t.Errorf("expected 5.0 (no multiplier applied), got %v", result)
	}
}

func TestCalculateTaskDuration_StepFormula(t *testing.T) {
	task := models.WBSTask{
		ID:               uuid.New(),
		Code:             "9.2",
		Name:             "Wall Framing",
		BaseDurationDays: 4.0,
	}
	context := models.ProjectContext{
		SupplyChainVolatility: 3, // 3 floors
	}
	multipliers := []models.DurationMultiplier{
		{
			ID:                uuid.New(),
			WBSTaskCode:       "9.2",
			VariableKey:       "supply_chain_volatility",
			Weight:            0.3,
			MultiplierFormula: "step",
		},
	}

	// Step formula: 1 + (value - 1) * weight = 1 + (3 - 1) * 0.3 = 1.6
	// Expected: 4.0 * 1.6 = 6.4 → quantized to 6.5
	result := CalculateTaskDuration(task, defaultGSF, context, multipliers, types.Forecast{})

	if result != 6.5 {
		t.Errorf("expected 6.5, got %v", result)
	}
}

func TestCalculateBatchDurations(t *testing.T) {
	tasks := []models.WBSTask{
		{ID: uuid.New(), Code: "9.1", Name: "Floor Framing", BaseDurationDays: 5.0},
		{ID: uuid.New(), Code: "9.2", Name: "Wall Framing", BaseDurationDays: 4.0},
	}
	context := models.ProjectContext{
		SupplyChainVolatility: 2,
	}
	multipliers := []models.DurationMultiplier{
		{
			ID:                uuid.New(),
			WBSTaskCode:       "*",
			VariableKey:       "supply_chain_volatility",
			Weight:            0.1,
			MultiplierFormula: "linear",
		},
	}

	results := CalculateBatchDurations(tasks, defaultGSF, context, multipliers, types.Forecast{})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Both should have multiplier: * (1 + 2 * 0.1) = * 1.2
	// 5.0 * 1.2 = 6.0, 4.0 * 1.2 = 4.8 → quantized to 5.0
	if results[0].CalculatedDuration != 6.0 {
		t.Errorf("expected 6.0 for task 0, got %v", results[0].CalculatedDuration)
	}
	if results[1].CalculatedDuration != 5.0 {
		t.Errorf("expected 5.0 for task 1, got %v", results[1].CalculatedDuration)
	}
}

// === COMBINED SAF + MULTIPLIER TEST ===

func TestCalculateTaskDuration_SAFWithMultiplier(t *testing.T) {
	task := models.WBSTask{
		ID:               uuid.New(),
		Code:             "9.1",
		Name:             "Floor Framing",
		BaseDurationDays: 5.0,
		IsInspection:     false,
	}
	context := models.ProjectContext{
		SupplyChainVolatility: 2,
	}
	multipliers := []models.DurationMultiplier{
		{
			ID:                uuid.New(),
			WBSTaskCode:       "9.1",
			VariableKey:       "supply_chain_volatility",
			Weight:            0.1,
			MultiplierFormula: "linear",
		},
	}

	// GSF = 4500 → SAF = (4500/2250)^0.75 ≈ 1.68
	// Expected: 5.0 * 1.68 * 1.2 = 10.08 → quantized to 10.5
	result := CalculateTaskDuration(task, 4500.0, context, multipliers, types.Forecast{})

	if result != 10.5 {
		t.Errorf("expected 10.5 with SAF + multiplier, got %v", result)
	}
}

// === SWIM INTEGRATION TEST ===

func TestCalculateTaskDuration_WithWeather(t *testing.T) {
	// Task: Site Prep (Sensitive)
	task := models.WBSTask{
		ID:               uuid.New(),
		Code:             "7.2",
		Name:             "Site Clearing",
		BaseDurationDays: 4.0,
		IsInspection:     false,
	}
	context := models.ProjectContext{} // No multipliers

	// Forecast: Heavy Rain (1.15 multiplier)
	forecast := types.Forecast{
		PrecipitationMM: 15.0,
	}

	// Calculation: 4.0 (Base) * 1.0 (SAF) * 1.15 (Weather) = 4.6
	// Quantization: Ceiling(4.6 * 2) / 2 = Ceiling(9.2) / 2 = 10 / 2 = 5.0
	result := CalculateTaskDuration(task, defaultGSF, context, nil, forecast)

	expected := 5.0
	if result != expected {
		t.Errorf("expected %.1f with rain impact and quantization, got %v", expected, result)
	}

	// Forecast: Combined impact (Rain 1.15 * Freeze 1.25 = 1.4375)
	forecastCombined := types.Forecast{
		PrecipitationMM: 15.0,
		LowTempC:        -5.0,
	}

	// Calculation: 4.0 * 1.4375 = 5.75
	// Quantization: Ceiling(5.75 * 2) / 2 = Ceiling(11.5) / 2 = 6.0
	resultCombined := CalculateTaskDuration(task, defaultGSF, context, nil, forecastCombined)

	expectedCombined := 6.0
	if resultCombined != expectedCombined {
		t.Errorf("expected %.1f with rain/freeze impact and quantization, got %v", expectedCombined, resultCombined)
	}
}
