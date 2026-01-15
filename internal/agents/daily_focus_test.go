package agents

import (
	"context"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/genai"
)

// Mocks
type MockWeatherService struct{ mock.Mock }

func (m *MockWeatherService) GetForecast(lat, long float64) (types.Forecast, error) {
	args := m.Called(lat, long)
	return args.Get(0).(types.Forecast), args.Error(1)
}

type MockNotificationService struct{ mock.Mock }

func (m *MockNotificationService) SendSMS(contactID string, message string) error {
	args := m.Called(contactID, message)
	return args.Error(0)
}
func (m *MockNotificationService) SendEmail(to string, subject string, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

type MockAIClient struct{ mock.Mock }

func (m *MockAIClient) GenerateContent(ctx context.Context, modelType ai.ModelType, parts ...*genai.Part) (string, error) {
	args := m.Called(ctx, modelType, parts)
	return args.String(0), args.Error(1)
}
func (m *MockAIClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	args := m.Called(ctx, text)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]float32), args.Error(1)
}
func (m *MockAIClient) Close() error { return nil }

func TestDailyFocusAgent_BuildPrompt(t *testing.T) {
	// Setup
	p := models.Project{
		Name:    "Test Project",
		Address: "123 Test Lane",
	}
	w := types.Forecast{
		HighTempC: 30, LowTempC: 20, PrecipitationMM: 5, Conditions: "Rainy",
	}
	now := time.Now()
	later := now.Add(24 * time.Hour)
	tasks := []models.ProjectTask{
		{Name: "Pour Concrete", Status: "in_progress", IsOnCriticalPath: true, PlannedStart: &now},
		{Name: "Install Windows", Status: "pending", IsOnCriticalPath: false, PlannedStart: &later},
	}

	agent := &DailyFocusAgent{} // No deps needed for pure function check if we export/expose prompt builder
	// But BuildPrompt is private. We can test via ProcessProject if we mock DB?
	// Or we can just use reflection or internal test.
	// Since we are in `agents` package, we can access private methods if test is in `agents` package (not `agents_test`).

	prompt := agent.buildPrompt(p, w, tasks)

	assert.Contains(t, prompt, "Test Project")
	assert.Contains(t, prompt, "Rainy")
	assert.Contains(t, prompt, "[CRITICAL PATH] Pour Concrete")
	assert.Contains(t, prompt, "Install Windows")
}
