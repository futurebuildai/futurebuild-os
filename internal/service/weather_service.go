package service

import (
	"time"

	"github.com/colton/futurebuild/pkg/types"
)

// MockWeatherService implements types.WeatherService
type MockWeatherService struct{}

func NewMockWeatherService() *MockWeatherService {
	return &MockWeatherService{}
}

// GetForecast matches the interface: GetForecast(lat, long float64) (Forecast, error)
func (s *MockWeatherService) GetForecast(lat, long float64) (types.Forecast, error) {
	// Return a stub forecast
	return types.Forecast{
		Date:                     time.Now().Format("2006-01-02"),
		HighTempC:                25.0,
		LowTempC:                 15.0,
		PrecipitationMM:          0.0,
		PrecipitationProbability: 0.1,
		Conditions:               "Sunny",
	}, nil
}
