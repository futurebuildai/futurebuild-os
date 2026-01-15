// Package mocks provides test doubles for time-travel simulation.
// See PRODUCTION_PLAN.md Step 49 Amendment 3
package mocks

import "github.com/colton/futurebuild/pkg/types"

// MockWeatherService returns a static sunny forecast (0mm rain)
// to ensure ProcurementAgent does not add random weather buffer.
type MockWeatherService struct{}

// GetForecast returns a deterministic "Sunny" forecast.
// PrecipitationProbability is 0 to ensure no weather buffer is added.
func (m *MockWeatherService) GetForecast(lat, long float64) (types.Forecast, error) {
	return types.Forecast{
		HighTempC:                25.0,
		LowTempC:                 15.0,
		PrecipitationMM:          0.0,
		PrecipitationProbability: 0.0,
		Conditions:               "Sunny",
	}, nil
}
