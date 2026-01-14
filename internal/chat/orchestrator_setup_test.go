package chat

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func TestNewOrchestrator_InitializesCorrectly(t *testing.T) {
	// Arrange
	// We can pass nil for the pool since NewOrchestrator wraps it.
	// We just want to check if the struct is populated.
	// NOTE: If we pass nil, NewPgxMessageStore gets nil db, which is fine for this structure test.
	var pool *pgxpool.Pool = nil

	mockTask := &MockTaskService{}
	mockSchedule := &MockScheduleService{}
	mockInvoice := &MockInvoiceService{}

	// Act
	orch := NewOrchestrator(pool, mockTask, mockSchedule, mockInvoice)

	// Assert
	assert.NotNil(t, orch)
	assert.NotNil(t, orch.db, "Default MessagePersister should be initialized")
	assert.IsType(t, &PgxMessageStore{}, orch.db, "Should use PgxMessageStore by default")

	assert.Equal(t, mockTask, orch.TaskService)
	assert.Equal(t, mockSchedule, orch.ScheduleService)
	assert.Equal(t, mockInvoice, orch.InvoiceService)
}
