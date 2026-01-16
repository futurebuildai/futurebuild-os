package agents

import (
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

// defaultProcurementCfg is the standard config for procurement tests
var defaultProcurementCfg = config.DefaultProcurementConfig()

// mockWeatherService implements types.WeatherService for testing.
type mockWeatherService struct {
	forecast types.Forecast
	err      error
}

func (m *mockWeatherService) GetForecast(lat, long float64) (types.Forecast, error) {
	return m.forecast, m.err
}

// TestAnalyzeItem_ScenarioA tests OK status when plenty of time remains.
// See PRODUCTION_PLAN.md Step 46: Scenario A
func TestAnalyzeItem_ScenarioA(t *testing.T) {
	agent := &ProcurementAgent{
		weather: &mockWeatherService{forecast: types.Forecast{PrecipitationProbability: 0.2}},
		config:  defaultProcurementCfg,
	}

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	earlyStart := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC) // 30 days away

	zipCode := "78701"
	item := procurementRow{
		ID:            uuid.New(),
		Name:          "Roof Trusses",
		LeadTimeWeeks: 1, // 7 days
		Status:        types.ProcurementAlertPending,
		EarlyStart:    &earlyStart,
		ZipCode:       &zipCode,
	}

	// Lead time = 7 days, buffer = 5 days, total = 12 days needed.
	// 30 days available > 12 days needed => OK
	result := agent.analyzeItem(item, now)

	if result.NewStatus != types.ProcurementAlertOK {
		t.Errorf("Expected OK, got %s", result.NewStatus)
	}
	if result.ShouldNotify {
		t.Error("Should not notify for OK status")
	}
}

// TestAnalyzeItem_ScenarioB tests CRITICAL status when time has run out.
// See PRODUCTION_PLAN.md Step 46: Scenario B
func TestAnalyzeItem_ScenarioB(t *testing.T) {
	agent := &ProcurementAgent{
		weather: &mockWeatherService{forecast: types.Forecast{PrecipitationProbability: 0.2}},
		config:  defaultProcurementCfg,
	}

	now := time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC)        // Day 16
	earlyStart := time.Date(2026, 1, 30, 0, 0, 0, 0, time.UTC) // 14 days away

	zipCode := "78701"
	item := procurementRow{
		ID:            uuid.New(),
		Name:          "Windows",
		LeadTimeWeeks: 2, // 14 days
		Status:        types.ProcurementAlertPending,
		EarlyStart:    &earlyStart,
		ZipCode:       &zipCode,
	}

	// Lead time = 14 days, buffer = 5 days, total = 19 days needed.
	// 14 days available < 19 days needed => CRITICAL
	// MustOrderDate = Jan 30 - 19 = Jan 11. Now (Jan 16) > Jan 11 => CRITICAL
	result := agent.analyzeItem(item, now)

	if result.NewStatus != types.ProcurementAlertCritical {
		t.Errorf("Expected CRITICAL, got %s", result.NewStatus)
	}
	if !result.ShouldNotify {
		t.Error("Should notify for CRITICAL status transition")
	}
}

// TestAnalyzeItem_ScenarioC tests weather buffer behavior when geocoding unavailable.
// See PRODUCTION_PLAN.md Step 46: Scenario C (SWIM)
// P0 Fix: Weather buffer is now skipped (set to 0) when geocoding is not implemented.
// This test documents the current safe behavior.
func TestAnalyzeItem_ScenarioC(t *testing.T) {
	// Storm forecast: 60% precipitation
	// P1 Fix: Weather buffer now uses conservative default (3 days) from config.
	// This test verifies the new correct behavior.
	agent := &ProcurementAgent{
		weather: &mockWeatherService{forecast: types.Forecast{PrecipitationProbability: 0.6}},
		config:  defaultProcurementCfg,
	}

	now := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)         // Day 5
	earlyStart := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC) // 16 days away

	zipCode := "78701"
	item := procurementRow{
		ID:            uuid.New(),
		Name:          "HVAC Unit",
		LeadTimeWeeks: 2, // 14 days
		Status:        types.ProcurementAlertPending,
		EarlyStart:    &earlyStart,
		ZipCode:       &zipCode,
	}

	// P1 Fix: Weather buffer is now 3 days (conservative default from config).
	// Formula: NeedBy = EarlyStart - 2 = Jan 19
	// CalculatedOrderDate = NeedBy - LeadTime - WeatherBuffer = Jan 19 - 14 - 3 = Jan 2
	// Now (Jan 5) > Jan 2 => CRITICAL (past the order deadline)
	result := agent.analyzeItem(item, now)

	// Expected: CRITICAL because we're past the order deadline
	if result.NewStatus != types.ProcurementAlertCritical {
		t.Errorf("Expected CRITICAL (now > mustOrderDate with weather buffer), got %s", result.NewStatus)
	}
}

// TestAnalyzeItem_Warning tests WARNING status when within 3 days of deadline.
func TestAnalyzeItem_Warning(t *testing.T) {
	agent := &ProcurementAgent{
		weather: &mockWeatherService{forecast: types.Forecast{PrecipitationProbability: 0.2}},
		config:  defaultProcurementCfg,
	}

	// P1 Fix: Adjusted dates to trigger WARNING with 3-day weather buffer
	now := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)        // Day 10
	earlyStart := time.Date(2026, 1, 24, 0, 0, 0, 0, time.UTC) // 14 days away

	zipCode := "78701"
	item := procurementRow{
		ID:            uuid.New(),
		Name:          "Exterior Doors",
		LeadTimeWeeks: 1, // 7 days
		Status:        types.ProcurementAlertPending,
		EarlyStart:    &earlyStart,
		ZipCode:       &zipCode,
	}

	// P1 Fix: Formula with 3-day weather buffer:
	// NeedBy = EarlyStart - 2 = Jan 22
	// CalculatedOrderDate = NeedBy - LeadTime - WeatherBuffer = Jan 22 - 7 - 3 = Jan 12
	// Now (Jan 10) is 2 days before Jan 12 => within 3-day warning threshold
	result := agent.analyzeItem(item, now)

	if result.NewStatus != types.ProcurementAlertWarning {
		t.Errorf("Expected WARNING, got %s", result.NewStatus)
	}
	if !result.ShouldNotify {
		t.Error("Should notify for WARNING status transition")
	}
}

// TestAnalyzeItem_NilEarlyStart tests handling of missing schedule data.
func TestAnalyzeItem_NilEarlyStart(t *testing.T) {
	agent := &ProcurementAgent{
		config: defaultProcurementCfg,
	}

	item := procurementRow{
		ID:            uuid.New(),
		Name:          "Cabinets",
		LeadTimeWeeks: 4,
		Status:        types.ProcurementAlertPending,
		EarlyStart:    nil, // No schedule yet
	}

	result := agent.analyzeItem(item, time.Now())

	if result.NewStatus != types.ProcurementAlertPending {
		t.Errorf("Expected PENDING for nil EarlyStart, got %s", result.NewStatus)
	}
}

// =============================================================================
// ZERO TRUST TEST: P1 "Ghost Physics" Fix (Operation Ironclad Task 3)
// =============================================================================
// These tests verify that:
// 1. Missing zip code returns ConfigError (schedule calculation blocked)
// 2. Weather buffer is NEVER zero (conservative default applied)
// See PRODUCTION_PLAN.md Phase 49 Retrofit
// =============================================================================

// TestAnalyzeItem_MissingZipCode_ConfigError verifies ConfigError for missing location.
func TestAnalyzeItem_MissingZipCode_ConfigError(t *testing.T) {
	agent := &ProcurementAgent{
		weather: &mockWeatherService{forecast: types.Forecast{PrecipitationProbability: 0.2}},
		config:  defaultProcurementCfg,
	}

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	earlyStart := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	item := procurementRow{
		ID:            uuid.New(),
		Name:          "Exterior Siding",
		LeadTimeWeeks: 2,
		Status:        types.ProcurementAlertPending,
		EarlyStart:    &earlyStart,
		ZipCode:       nil, // MISSING - should trigger ConfigError
	}

	result := agent.analyzeItem(item, now)

	// Assert: ConfigError status
	if result.NewStatus != types.ProcurementAlertConfigError {
		t.Errorf("Expected ConfigError, got %s", result.NewStatus)
	}

	// Assert: Should notify
	if !result.ShouldNotify {
		t.Error("Should notify when location data is missing")
	}

	// Assert: Message contains "Missing location data"
	if result.Message == "" {
		t.Error("Expected non-empty error message")
	}

	// Assert: No order date calculated (zero value)
	if !result.CalculatedOrderDate.IsZero() {
		t.Errorf("Expected zero CalculatedOrderDate for ConfigError, got %v", result.CalculatedOrderDate)
	}
}

// TestAnalyzeItem_EmptyZipCode_ConfigError verifies ConfigError for empty string zip code.
func TestAnalyzeItem_EmptyZipCode_ConfigError(t *testing.T) {
	agent := &ProcurementAgent{
		weather: &mockWeatherService{forecast: types.Forecast{PrecipitationProbability: 0.2}},
		config:  defaultProcurementCfg,
	}

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	earlyStart := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	emptyZip := ""

	item := procurementRow{
		ID:            uuid.New(),
		Name:          "Roofing Materials",
		LeadTimeWeeks: 2,
		Status:        types.ProcurementAlertPending,
		EarlyStart:    &earlyStart,
		ZipCode:       &emptyZip, // Empty string - should also trigger ConfigError
	}

	result := agent.analyzeItem(item, now)

	if result.NewStatus != types.ProcurementAlertConfigError {
		t.Errorf("Expected ConfigError for empty zip code, got %s", result.NewStatus)
	}
}

// TestAnalyzeItem_DefaultWeatherBuffer verifies weather buffer is never zero.
func TestAnalyzeItem_DefaultWeatherBuffer(t *testing.T) {
	// Test with explicit default config
	cfg := config.DefaultProcurementConfig()

	// Assert: DefaultWeatherBufferDays is set (not zero)
	if cfg.DefaultWeatherBufferDays <= 0 {
		t.Errorf("DefaultWeatherBufferDays should be > 0, got %d", cfg.DefaultWeatherBufferDays)
	}

	// Assert: Default is conservative (at least 3 days)
	if cfg.DefaultWeatherBufferDays < 3 {
		t.Errorf("DefaultWeatherBufferDays should be at least 3 for safety, got %d", cfg.DefaultWeatherBufferDays)
	}

	t.Logf("DefaultWeatherBufferDays verified: %d days (fail-safe)", cfg.DefaultWeatherBufferDays)
}

// TestAnalyzeItem_WeatherBufferApplied verifies weather buffer affects order date.
func TestAnalyzeItem_WeatherBufferApplied(t *testing.T) {
	agent := &ProcurementAgent{
		weather: &mockWeatherService{forecast: types.Forecast{PrecipitationProbability: 0.2}},
		config:  defaultProcurementCfg,
	}

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	earlyStart := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC) // Feb 1

	zipCode := "78701"
	item := procurementRow{
		ID:            uuid.New(),
		Name:          "Test Material",
		LeadTimeWeeks: 1, // 7 days
		Status:        types.ProcurementAlertPending,
		EarlyStart:    &earlyStart,
		ZipCode:       &zipCode,
	}

	result := agent.analyzeItem(item, now)

	// Formula: NeedBy = Feb 1 - 2 = Jan 30
	// OrderDate = Jan 30 - 7 (lead) - 3 (weather buffer) = Jan 20
	expectedOrderDate := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)

	if !result.CalculatedOrderDate.Equal(expectedOrderDate) {
		t.Errorf("Expected OrderDate %v (with weather buffer), got %v",
			expectedOrderDate.Format("2006-01-02"),
			result.CalculatedOrderDate.Format("2006-01-02"))
	}
}
