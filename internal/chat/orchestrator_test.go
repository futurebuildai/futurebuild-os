package chat

import (
	"context"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock Implementations ---

// MockMessagePersister records calls to SaveMessage for verification.
type MockMessagePersister struct {
	Messages []models.ChatMessage
	Err      error // Error to return (for testing error paths)
}

func (m *MockMessagePersister) SaveMessage(_ context.Context, msg models.ChatMessage) error {
	if m.Err != nil {
		return m.Err
	}
	m.Messages = append(m.Messages, msg)
	return nil
}

// MockTaskService is a no-op mock for TaskService.
type MockTaskService struct{}

func (m *MockTaskService) UpdateTaskStatus(_ context.Context, _, _, _ uuid.UUID, _ types.TaskStatus) error {
	return nil
}

// MockScheduleService is a configurable mock for ScheduleService.
type MockScheduleService struct {
	Summary *service.ProjectScheduleSummary
	Err     error
}

func (m *MockScheduleService) GetTask(_ context.Context, _, _, _ uuid.UUID) (*models.ProjectTask, error) {
	return nil, nil
}

func (m *MockScheduleService) GetProjectSchedule(_ context.Context, _, _ uuid.UUID) (*service.ProjectScheduleSummary, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	// Return default summary if none provided (for tests that don't care about schedule data)
	if m.Summary == nil {
		return &service.ProjectScheduleSummary{
			ProjectEnd:        time.Now().AddDate(0, 6, 0),
			CriticalPathCount: 0,
			TotalTasks:        0,
			CompletedTasks:    0,
		}, nil
	}
	return m.Summary, nil
}

// MockInvoiceService is a no-op mock for InvoiceService.
type MockInvoiceService struct{}

func (m *MockInvoiceService) AnalyzeInvoice(_ context.Context, _ uuid.UUID, _ uuid.UUID) (uuid.UUID, *types.InvoiceExtraction, error) {
	return uuid.Nil, nil, nil
}

func (m *MockInvoiceService) SaveExtraction(_ context.Context, _ uuid.UUID, _ *types.InvoiceExtraction, _ *uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}

// --- Tests ---

func TestProcessRequest_PersistsUserAndModelMessages(t *testing.T) {
	// Arrange
	mockDB := &MockMessagePersister{}
	orchestrator := &Orchestrator{
		db:              mockDB,
		TaskService:     &MockTaskService{},
		ScheduleService: &MockScheduleService{},
		InvoiceService:  &MockInvoiceService{},
	}

	userID := uuid.New()
	projectID := uuid.New()
	req := ChatRequest{
		ProjectID: projectID,
		Message:   "Show me the project schedule",
	}

	// Act
	resp, err := orchestrator.ProcessRequest(context.Background(), userID, uuid.New(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify two messages were saved (User + Model)
	assert.Len(t, mockDB.Messages, 2, "Expected exactly 2 messages to be persisted")

	// Verify User message
	assert.Equal(t, types.ChatRoleUser, mockDB.Messages[0].Role)
	assert.Equal(t, req.Message, mockDB.Messages[0].Content)
	assert.Equal(t, userID, mockDB.Messages[0].UserID)
	assert.Equal(t, projectID, mockDB.Messages[0].ProjectID)

	// Verify Model message
	assert.Equal(t, types.ChatRoleModel, mockDB.Messages[1].Role)
	assert.Equal(t, resp.Reply, mockDB.Messages[1].Content)
}

func TestProcessRequest_ClassifiesIntentCorrectly(t *testing.T) {
	testCases := []struct {
		message        string
		expectedIntent types.Intent
	}{
		{"Show me the project schedule", types.IntentGetSchedule},
		{"I have an invoice to process", types.IntentProcessInvoice},
		{"Why is the project delayed?", types.IntentExplainDelay},
		{"Mark the framing task as complete", types.IntentUpdateTaskStatus},
		{"What's the weather like?", types.IntentUnknown},
	}

	for _, tc := range testCases {
		t.Run(tc.message, func(t *testing.T) {
			// Arrange
			mockDB := &MockMessagePersister{}
			orchestrator := &Orchestrator{
				db:              mockDB,
				TaskService:     &MockTaskService{},
				ScheduleService: &MockScheduleService{},
				InvoiceService:  &MockInvoiceService{},
			}

			req := ChatRequest{
				ProjectID: uuid.New(),
				Message:   tc.message,
			}

			// Act
			resp, err := orchestrator.ProcessRequest(context.Background(), uuid.New(), uuid.New(), req)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIntent, resp.Intent)
		})
	}
}

func TestProcessRequest_ReturnsErrorOnUserMessagePersistFailure(t *testing.T) {
	// Arrange
	mockDB := &MockMessagePersister{Err: assert.AnError}
	orchestrator := &Orchestrator{
		db:              mockDB,
		TaskService:     &MockTaskService{},
		ScheduleService: &MockScheduleService{},
		InvoiceService:  &MockInvoiceService{},
	}

	req := ChatRequest{
		ProjectID: uuid.New(),
		Message:   "Test",
	}

	// Act
	resp, err := orchestrator.ProcessRequest(context.Background(), uuid.New(), uuid.New(), req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to persist user message")
}

// Extending the test file to support the complexity needed for the 2nd call failure
type FailOnSecondSavePersister struct {
	CallCount int
}

func (m *FailOnSecondSavePersister) SaveMessage(_ context.Context, _ models.ChatMessage) error {
	m.CallCount++
	if m.CallCount == 2 {
		return assert.AnError
	}
	return nil
}

func TestProcessRequest_ModelPersistError(t *testing.T) {
	// Arrange
	mockDB := &FailOnSecondSavePersister{}
	orchestrator := &Orchestrator{
		db:              mockDB,
		TaskService:     &MockTaskService{},
		ScheduleService: &MockScheduleService{},
		InvoiceService:  &MockInvoiceService{},
	}

	req := ChatRequest{
		ProjectID: uuid.New(),
		Message:   "Test",
	}

	// Act
	resp, err := orchestrator.ProcessRequest(context.Background(), uuid.New(), uuid.New(), req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to persist model message")
}

// --- Command Pattern Tests ---

func TestGetScheduleCommand_FormatsDataCorrectly(t *testing.T) {
	// Arrange
	mockSchedule := &MockScheduleService{
		Summary: &service.ProjectScheduleSummary{
			ProjectEnd:        time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
			CriticalPathCount: 5,
			TotalTasks:        20,
			CompletedTasks:    8,
		},
	}

	cmd := &GetScheduleCommand{
		scheduleService: mockSchedule,
		projectID:       uuid.New(),
		orgID:           uuid.New(),
	}

	// Act
	result, err := cmd.Execute(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Contains(t, result, "Jun 15, 2026")
	assert.Contains(t, result, "Critical Path Tasks: 5")
	assert.Contains(t, result, "Total Tasks: 20")
	assert.Contains(t, result, "Completed: 8")
}

func TestGetScheduleCommand_ReturnsErrorOnServiceFailure(t *testing.T) {
	// Arrange
	mockSchedule := &MockScheduleService{Err: assert.AnError}

	cmd := &GetScheduleCommand{
		scheduleService: mockSchedule,
		projectID:       uuid.New(),
		orgID:           uuid.New(),
	}

	// Act
	result, err := cmd.Execute(context.Background())

	// Assert
	require.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "failed to get schedule")
}

func TestProcessRequest_CallsScheduleServiceForGetScheduleIntent(t *testing.T) {
	// Arrange
	mockDB := &MockMessagePersister{}
	mockSchedule := &MockScheduleService{
		Summary: &service.ProjectScheduleSummary{
			ProjectEnd:        time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC),
			CriticalPathCount: 3,
			TotalTasks:        15,
			CompletedTasks:    10,
		},
	}

	orchestrator := &Orchestrator{
		db:              mockDB,
		TaskService:     &MockTaskService{},
		ScheduleService: mockSchedule,
		InvoiceService:  &MockInvoiceService{},
	}

	req := ChatRequest{
		ProjectID: uuid.New(),
		Message:   "Show me the project schedule", // triggers IntentGetSchedule
	}

	// Act
	resp, err := orchestrator.ProcessRequest(context.Background(), uuid.New(), uuid.New(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, types.IntentGetSchedule, resp.Intent)
	// Verify the response contains real data from the mock service
	assert.Contains(t, resp.Reply, "Dec 01, 2026")
	assert.Contains(t, resp.Reply, "Critical Path Tasks: 3")
}
