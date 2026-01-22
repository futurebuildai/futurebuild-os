package agents

import (
	"context"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/service/mocks"
	"github.com/colton/futurebuild/pkg/clock"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProcurementAgent_WeatherBuffer_Storm verifies that rain forecasts add +2 days buffer.
// This is the "Proof of Value" test demonstrating service mocking.
func TestProcurementAgent_WeatherBuffer_Storm(t *testing.T) {
	// 1. Setup
	repo := NewMockProcurementRepository()
	weather := &mocks.MockWeatherService{}
	clk := clock.NewMockClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	cfg := config.ProcurementConfig{
		DefaultWeatherBufferDays: 3,
		StagingBufferDays:        2,
		LeadTimeWarningThreshold: 7,
	}

	agent := NewProcurementAgent(repo, weather, clk, cfg)

	// 2. Mock Data
	itemID := uuid.New()
	zipCode := "90210"
	earlyStart := clk.Now().AddDate(0, 0, 20) // Start in 20 days (Jan 21)
	item := procurementRow{
		ID:            itemID,
		Name:          "Steel Beams",
		LeadTimeWeeks: 1, // 7 days
		Status:        types.ProcurementAlertOK,
		EarlyStart:    &earlyStart,
		ZipCode:       &zipCode,
	}
	repo.Items = []procurementRow{item}

	// 3. Inject Storm (High precip probability)
	weather.ForecastResp = types.Forecast{
		PrecipitationProbability: 0.8,
		Conditions:               "Heavy Rain",
	}

	// 4. Calculation:
	// NeedBy = EarlyStart(Jan 21) - Staging(2) = Jan 19
	// Base Order = NeedBy(Jan 19) - Lead(7) - Buffer(3) = Jan 9
	// Rain adds +2: Order = Jan 9 - 2 = Jan 7
	expectedDate := earlyStart.AddDate(0, 0, -(2 + 7 + 3 + 2)) // Jan 7

	// 5. Execute
	err := agent.Execute(context.Background())
	require.NoError(t, err)

	// 6. Assertions
	updatedResults := repo.GetAllUpdatedResults()
	require.Len(t, updatedResults, 1, "Should have one updated result")

	result := updatedResults[0]
	assert.Equal(t, itemID, result.ID)
	assert.Equal(t, expectedDate.Format("2006-01-02"), result.CalculatedOrderDate.Format("2006-01-02"),
		"Order date should account for +2 rain buffer")

	// 7. Verify weather service was called (L7 assertion)
	assert.Len(t, weather.Calls, 1, "Weather service should be called once")
}

// TestProcurementAgent_WeatherBuffer_Baseline verifies the baseline calculation WITHOUT rain.
// This ensures the test above isn't just "lucky" and the delta is real.
func TestProcurementAgent_WeatherBuffer_Baseline(t *testing.T) {
	// 1. Setup
	repo := NewMockProcurementRepository()
	weather := &mocks.MockWeatherService{}
	clk := clock.NewMockClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	cfg := config.ProcurementConfig{
		DefaultWeatherBufferDays: 3,
		StagingBufferDays:        2,
		LeadTimeWarningThreshold: 7,
	}

	agent := NewProcurementAgent(repo, weather, clk, cfg)

	// 2. Mock Data
	itemID := uuid.New()
	zipCode := "90210"
	earlyStart := clk.Now().AddDate(0, 0, 20) // Start in 20 days (Jan 21)
	item := procurementRow{
		ID:            itemID,
		Name:          "Steel Beams",
		LeadTimeWeeks: 1, // 7 days
		Status:        types.ProcurementAlertOK,
		EarlyStart:    &earlyStart,
		ZipCode:       &zipCode,
	}
	repo.Items = []procurementRow{item}

	// 3. Inject Clear Weather (NO rain bonus)
	weather.ForecastResp = types.Forecast{
		PrecipitationProbability: 0.1, // Low probability
		Conditions:               "Sunny",
	}

	// 4. Calculation (NO rain buffer):
	// NeedBy = EarlyStart(Jan 21) - Staging(2) = Jan 19
	// Order = NeedBy(Jan 19) - Lead(7) - Buffer(3) = Jan 9
	expectedDate := earlyStart.AddDate(0, 0, -(2 + 7 + 3)) // Jan 9 (NOT Jan 7)

	// 5. Execute
	err := agent.Execute(context.Background())
	require.NoError(t, err)

	// 6. Assertions
	updatedResults := repo.GetAllUpdatedResults()
	require.Len(t, updatedResults, 1, "Should have one updated result")

	result := updatedResults[0]
	assert.Equal(t, itemID, result.ID)
	assert.Equal(t, expectedDate.Format("2006-01-02"), result.CalculatedOrderDate.Format("2006-01-02"),
		"Order date should NOT include rain buffer for clear weather")
}

// TestProcurementAgent_MissingZipCode verifies ConfigError status for missing location data.
// L7 Zero-Trust: Fail loudly on missing data, don't use defaults for location-sensitive calculations.
func TestProcurementAgent_MissingZipCode(t *testing.T) {
	// 1. Setup
	repo := NewMockProcurementRepository()
	weather := &mocks.MockWeatherService{}
	clk := clock.NewMockClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	cfg := config.ProcurementConfig{
		DefaultWeatherBufferDays: 3,
		StagingBufferDays:        2,
		LeadTimeWarningThreshold: 7,
	}

	agent := NewProcurementAgent(repo, weather, clk, cfg)

	// 2. Mock Data with MISSING ZipCode
	itemID := uuid.New()
	earlyStart := clk.Now().AddDate(0, 0, 20)
	item := procurementRow{
		ID:            itemID,
		Name:          "HVAC Unit",
		LeadTimeWeeks: 2,
		Status:        types.ProcurementAlertOK,
		EarlyStart:    &earlyStart,
		ZipCode:       nil, // MISSING!
	}
	repo.Items = []procurementRow{item}

	// 3. Execute
	err := agent.Execute(context.Background())
	require.NoError(t, err)

	// 4. Assertions - should flag as ConfigError
	updatedResults := repo.GetAllUpdatedResults()
	require.Len(t, updatedResults, 1)

	result := updatedResults[0]
	assert.Equal(t, types.ProcurementAlertConfigError, result.NewStatus,
		"Missing ZipCode should trigger ConfigError status")
	assert.True(t, result.ShouldNotify, "ConfigError should trigger notification")

	// 5. Weather service should NOT be called (L7 short-circuit)
	assert.Len(t, weather.Calls, 0, "Weather should not be called for items with missing ZipCode")
}
