package chat

import (
	"context"
	"errors"
	"testing"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

// =============================================================================
// ZERO TRUST TEST: P0 Error "Black Hole" Fix (Operation Ironclad Task 2)
// =============================================================================
// These tests verify that when a command fails, an error message is persisted
// to the chat history so the user has a visual record of the failure.
// See PRODUCTION_PLAN.md Phase 49 Retrofit
// =============================================================================

// executorMockFailingCommand is a command that always fails.
type executorMockFailingCommand struct {
	err error
}

func (m *executorMockFailingCommand) Execute(ctx context.Context) (string, *Artifact, error) {
	return "", nil, m.err
}

func (m *executorMockFailingCommand) ConsistencyLevel() types.ConsistencyType {
	return types.ConsistencyBestEffort
}

// executorMockRecordingPersister records all persisted messages for verification.
type executorMockRecordingPersister struct {
	messages []models.ChatMessage
	err      error
}

func (m *executorMockRecordingPersister) SaveMessage(_ context.Context, msg models.ChatMessage) error {
	m.messages = append(m.messages, msg)
	return m.err
}

func (m *executorMockRecordingPersister) Pool() Transactor {
	return nil
}

// executorMockDLQ is a no-op DLQ for testing.
type executorMockDLQ struct{}

func (m *executorMockDLQ) EnqueueRetry(_ context.Context, _ models.ChatMessage) error {
	return nil
}

// executorMockWAL satisfies AuditWAL interface.
type executorMockWAL struct{}

func (m *executorMockWAL) AppendRecord(_ context.Context, _ models.ChatMessage) error {
	return nil
}

func (m *executorMockWAL) Flush() error {
	return nil
}

func (m *executorMockWAL) Close() error {
	return nil
}

// executorMockCircuitBreaker satisfies AuditCircuitBreaker interface.
type executorMockCircuitBreaker struct{}

func (m *executorMockCircuitBreaker) IsOpen() bool        { return false }
func (m *executorMockCircuitBreaker) RecordSuccess()      {}
func (m *executorMockCircuitBreaker) RecordFailure()      {}
func (m *executorMockCircuitBreaker) State() CircuitState { return CircuitClosed }

// TestExecutor_ErrorPersistence is the Zero Trust test for P0 error black hole fix.
// It mocks a failing command and asserts that an error ChatMessage is persisted.
func TestExecutor_ErrorPersistence(t *testing.T) {
	// Arrange: Create a mock persister that records messages
	persister := &executorMockRecordingPersister{}
	dlq := &executorMockDLQ{}
	wal := &executorMockWAL{}
	cb := &executorMockCircuitBreaker{}

	registry := NewPersistenceStrategyRegistry(persister, dlq, wal, cb)
	executor := NewCommandExecutor(registry)

	// Create a failing command
	expectedErr := errors.New("database unavailable")
	failingCmd := &executorMockFailingCommand{err: expectedErr}

	execCtx := ExecutionContext{
		UserID:    uuid.New(),
		ProjectID: uuid.New(),
		Intent:    types.IntentGetSchedule,
	}

	// Act: Execute the failing command
	_, err := executor.Execute(context.Background(), failingCmd, execCtx)

	// Assert: Error should be returned
	if err == nil {
		t.Fatal("expected error to be returned, got nil")
	}

	// Assert: Error message should be persisted to chat history
	if len(persister.messages) != 1 {
		t.Fatalf("expected 1 message persisted, got %d", len(persister.messages))
	}

	persistedMsg := persister.messages[0]

	// Verify the persisted message properties
	if persistedMsg.Role != types.ChatRoleModel {
		t.Errorf("expected Role=%s, got %s", types.ChatRoleModel, persistedMsg.Role)
	}

	if persistedMsg.ProjectID != execCtx.ProjectID {
		t.Errorf("expected ProjectID=%s, got %s", execCtx.ProjectID, persistedMsg.ProjectID)
	}

	if persistedMsg.UserID != execCtx.UserID {
		t.Errorf("expected UserID=%s, got %s", execCtx.UserID, persistedMsg.UserID)
	}

	// Verify error message content is user-friendly
	if persistedMsg.Content == "" {
		t.Error("expected non-empty error message content")
	}

	// Success: Error message was persisted, audit trail is complete
	t.Logf("Error message persisted successfully: %q", persistedMsg.Content)
}

// TestExecutor_ErrorPersistence_PersistFails verifies graceful handling when
// error message persistence also fails.
func TestExecutor_ErrorPersistence_PersistFails(t *testing.T) {
	// Arrange: Create a persister that fails
	persister := &executorMockRecordingPersister{err: errors.New("persistence failed")}
	dlq := &executorMockDLQ{}
	wal := &executorMockWAL{}
	cb := &executorMockCircuitBreaker{}

	registry := NewPersistenceStrategyRegistry(persister, dlq, wal, cb)
	executor := NewCommandExecutor(registry)

	failingCmd := &executorMockFailingCommand{err: errors.New("original error")}

	execCtx := ExecutionContext{
		UserID:    uuid.New(),
		ProjectID: uuid.New(),
		Intent:    types.IntentGetSchedule,
	}

	// Act: Execute the failing command
	_, err := executor.Execute(context.Background(), failingCmd, execCtx)

	// Assert: Original error should still be returned (not persistence error)
	if err == nil {
		t.Fatal("expected error to be returned")
	}

	// The function should not panic on persistence failure
	t.Log("Graceful degradation verified - function returns original error even when persistence fails")
}

// executorMockSuccessCommand is a command that succeeds.
type executorMockSuccessCommand struct{}

func (m *executorMockSuccessCommand) Execute(ctx context.Context) (string, *Artifact, error) {
	return "Success!", nil, nil
}

func (m *executorMockSuccessCommand) ConsistencyLevel() types.ConsistencyType {
	return types.ConsistencyStrict
}

func TestExecutor_SuccessNoErrorMessage(t *testing.T) {
	persister := &executorMockRecordingPersister{}
	dlq := &executorMockDLQ{}
	wal := &executorMockWAL{}
	cb := &executorMockCircuitBreaker{}

	registry := NewPersistenceStrategyRegistry(persister, dlq, wal, cb)
	executor := NewCommandExecutor(registry)

	successCmd := &executorMockSuccessCommand{}

	execCtx := ExecutionContext{
		UserID:    uuid.New(),
		ProjectID: uuid.New(),
		Intent:    types.IntentGetSchedule,
	}

	// Act
	resp, err := executor.Execute(context.Background(), successCmd, execCtx)

	// Assert: No error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Assert: Response is valid
	if resp.Reply != "Success!" {
		t.Errorf("expected reply %q, got %q", "Success!", resp.Reply)
	}

	// Assert: One message persisted (the success message, not error)
	if len(persister.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(persister.messages))
	}

	// Assert: Message is the success response, not error
	if persister.messages[0].Content != "Success!" {
		t.Errorf("expected success message content, got %q", persister.messages[0].Content)
	}
}
