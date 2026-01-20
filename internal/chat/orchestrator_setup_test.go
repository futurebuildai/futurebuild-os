package chat

import (
	"context"
	"testing"

	"github.com/colton/futurebuild/internal/models"
	"github.com/stretchr/testify/assert"
)

// mockDLQ is a no-op DLQPersister for testing.
type mockDLQ struct{}

func (m *mockDLQ) EnqueueRetry(_ context.Context, _ models.ChatMessage) error {
	return nil
}

func TestNewOrchestrator_InitializesCorrectly(t *testing.T) {
	// Arrange
	// Create a MessagePersister (can use nil pool since NewPgxMessageStore just wraps it)
	// Operation Ironclad Task 2: NewOrchestrator now takes MessagePersister interface
	mockPersister := NewPgxMessageStore(nil) // nil pool is fine for structure test

	mockTask := &MockTaskService{}
	mockSchedule := &MockScheduleService{}
	mockInvoice := &MockInvoiceService{}
	mockDlq := &mockDLQ{}

	// Act
	// P0 FIX: DLQ is now mandatory
	orch, err := NewOrchestrator(mockPersister, mockTask, mockSchedule, mockInvoice, mockDlq, nil, nil)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, orch)
	assert.NotNil(t, orch.db, "MessagePersister should be initialized")
	assert.Equal(t, mockPersister, orch.db, "Should use injected MessagePersister")

	assert.Equal(t, mockTask, orch.TaskService)
	assert.Equal(t, mockSchedule, orch.ScheduleService)
	assert.Equal(t, mockInvoice, orch.InvoiceService)
}
