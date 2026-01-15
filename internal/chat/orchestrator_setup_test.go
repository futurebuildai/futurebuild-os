package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOrchestrator_InitializesCorrectly(t *testing.T) {
	// Arrange
	// Create a MessagePersister (can use nil pool since NewPgxMessageStore just wraps it)
	// Operation Ironclad Task 2: NewOrchestrator now takes MessagePersister interface
	mockPersister := NewPgxMessageStore(nil) // nil pool is fine for structure test

	mockTask := &MockTaskService{}
	mockSchedule := &MockScheduleService{}
	mockInvoice := &MockInvoiceService{}

	// Act
	orch := NewOrchestrator(mockPersister, mockTask, mockSchedule, mockInvoice)

	// Assert
	assert.NotNil(t, orch)
	assert.NotNil(t, orch.db, "MessagePersister should be initialized")
	assert.Equal(t, mockPersister, orch.db, "Should use injected MessagePersister")

	assert.Equal(t, mockTask, orch.TaskService)
	assert.Equal(t, mockSchedule, orch.ScheduleService)
	assert.Equal(t, mockInvoice, orch.InvoiceService)
}
