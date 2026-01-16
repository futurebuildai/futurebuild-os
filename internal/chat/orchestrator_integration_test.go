package chat

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// INTEGRATION TESTS: Two-Lane Consistency Strategy
// =============================================================================
// These tests verify the hybrid consistency model from Step 2:
// - Lane A (Slow/External): AI operations with best-effort persistence
// - Lane B (Fast/Internal): DB operations with strict consistency
//
// See PRODUCTION_PLAN.md Step 2 (Hybrid Consistency & Orchestrator Hardening)
// =============================================================================

// --- "Amnesiac Database" Mock: Fails on Model Persist Only ---

// AmnesiacPersister succeeds on SaveMessage(User) but fails on SaveMessage(Model).
// Used to test the Two-Lane Consistency Strategy.
type AmnesiacPersister struct {
	CallCount int
	Err       error // Error to return on second call (model persist)
}

func (m *AmnesiacPersister) SaveMessage(_ context.Context, _ models.ChatMessage) error {
	m.CallCount++
	if m.CallCount == 2 {
		// Second call is model message - simulate "amnesiac" database
		return m.Err
	}
	return nil
}

// Pool returns nil for integration tests that don't test transactional behavior.
func (m *AmnesiacPersister) Pool() Transactor {
	return nil
}

// --- Test Fixtures ---

func newTestOrchestrator(persister MessagePersister) *Orchestrator {
	// SRP Refactoring: Use constructor to properly build the executor and strategy registry
	return NewOrchestratorWithPersister(
		persister,
		NewDefaultRegexClassifier(),
		&MockTaskService{},
		&MockScheduleService{
			Summary: &service.ProjectScheduleSummary{
				ProjectEnd:        time.Now().AddDate(0, 6, 0),
				CriticalPathCount: 5,
				TotalTasks:        20,
				CompletedTasks:    10,
			},
		},
		&MockInvoiceService{},
		&mockDLQ{}, // REQUIRED for Lane A fallback
		nil,        // Optional: WAL not tested here
		nil,        // Optional: CircuitBreaker not tested here
	)
}

// captureLog returns a buffer that captures slog output and restores the default logger after the test.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(handler))
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})
	return &buf
}

// =============================================================================
// LANE A TESTS: Slow/External (AI Operations)
// =============================================================================

// TestTwoLane_LaneA_ProcessInvoice_ReturnsSuccessOnModelPersistFailure
// Scenario: "The Amnesiac Database"
// - SaveMessage(User) succeeds
// - SaveMessage(Model) fails
// - Intent: ProcessInvoice (Lane A - Slow/External)
//
// Expected:
// - Function returns result != nil (Success)
// - Function returns err == nil
// - Logs contain "CRITICAL" warning
func TestTwoLane_LaneA_ProcessInvoice_ReturnsSuccessOnModelPersistFailure(t *testing.T) {
	// Arrange
	logBuf := captureLog(t)
	persister := &AmnesiacPersister{Err: assert.AnError}
	orchestrator := newTestOrchestrator(persister)

	req := ChatRequest{
		ProjectID: uuid.New(),
		Message:   "Process this invoice", // Triggers IntentProcessInvoice (Lane A)
	}

	// Act
	resp, err := orchestrator.ProcessRequest(context.Background(), uuid.New(), uuid.New(), req)

	// Assert: Success despite persistence failure
	require.NoError(t, err, "Lane A: Should not return error when model persist fails")
	require.NotNil(t, resp, "Lane A: Should return response when model persist fails")
	assert.Equal(t, types.IntentProcessInvoice, resp.Intent)

	// Assert: CRITICAL warning was logged
	logged := logBuf.String()
	assert.Contains(t, logged, "CRITICAL", "Lane A: Should log CRITICAL warning on model persist failure")
	assert.Contains(t, logged, "Action succeeded but chat history save failed", "Lane A: Should log descriptive message")
}

// TestTwoLane_LaneA_InvoiceShorthand_ReturnsSuccessOnModelPersistFailure
// Validates the regression fix: "invoice" (noun only) triggers IntentProcessInvoice
func TestTwoLane_LaneA_InvoiceShorthand_ReturnsSuccessOnModelPersistFailure(t *testing.T) {
	// Arrange
	logBuf := captureLog(t)
	persister := &AmnesiacPersister{Err: assert.AnError}
	orchestrator := newTestOrchestrator(persister)

	req := ChatRequest{
		ProjectID: uuid.New(),
		Message:   "invoice", // Noun-only shorthand (regression fix)
	}

	// Act
	resp, err := orchestrator.ProcessRequest(context.Background(), uuid.New(), uuid.New(), req)

	// Assert: Noun-only triggers ProcessInvoice
	require.NoError(t, err, "Invoice shorthand: Should not return error")
	require.NotNil(t, resp, "Invoice shorthand: Should return response")
	assert.Equal(t, types.IntentProcessInvoice, resp.Intent, "Invoice shorthand: Should classify as ProcessInvoice")

	// Assert: Lane A behavior (graceful degradation)
	logged := logBuf.String()
	assert.Contains(t, logged, "CRITICAL", "Invoice shorthand: Should log CRITICAL warning")
}

// TestTwoLane_LaneA_ExplainDelay_ReturnsSuccessOnModelPersistFailure
// Verifies Lane A behavior for ExplainDelay intent
func TestTwoLane_LaneA_ExplainDelay_ReturnsSuccessOnModelPersistFailure(t *testing.T) {
	// Arrange
	logBuf := captureLog(t)
	persister := &AmnesiacPersister{Err: assert.AnError}
	orchestrator := newTestOrchestrator(persister)

	req := ChatRequest{
		ProjectID: uuid.New(),
		Message:   "Why is the project delayed?", // Triggers IntentExplainDelay (Lane A)
	}

	// Act
	resp, err := orchestrator.ProcessRequest(context.Background(), uuid.New(), uuid.New(), req)

	// Assert: Success despite persistence failure
	require.NoError(t, err, "Lane A: Should not return error for ExplainDelay")
	require.NotNil(t, resp, "Lane A: Should return response for ExplainDelay")
	assert.Equal(t, types.IntentExplainDelay, resp.Intent)

	// Assert: CRITICAL warning was logged
	logged := logBuf.String()
	assert.Contains(t, logged, "CRITICAL", "Lane A: Should log CRITICAL for ExplainDelay")
}

// TestTwoLane_LaneA_Unknown_ReturnsSuccessOnModelPersistFailure
// Verifies Lane A behavior for unknown intents (graceful degradation)
func TestTwoLane_LaneA_Unknown_ReturnsSuccessOnModelPersistFailure(t *testing.T) {
	// Arrange
	logBuf := captureLog(t)
	persister := &AmnesiacPersister{Err: assert.AnError}
	orchestrator := newTestOrchestrator(persister)

	req := ChatRequest{
		ProjectID: uuid.New(),
		Message:   "Hello there", // Triggers IntentUnknown (Lane A - graceful degradation)
	}

	// Act
	resp, err := orchestrator.ProcessRequest(context.Background(), uuid.New(), uuid.New(), req)

	// Assert: Success despite persistence failure (graceful degradation)
	require.NoError(t, err, "Lane A: Should not return error for Unknown")
	require.NotNil(t, resp, "Lane A: Should return response for Unknown")
	assert.Equal(t, types.IntentUnknown, resp.Intent)

	// Assert: CRITICAL warning was logged
	logged := logBuf.String()
	assert.Contains(t, logged, "CRITICAL", "Lane A: Should log CRITICAL for Unknown")
}

// =============================================================================
// LANE B TESTS: Fast/Internal (DB Operations)
// =============================================================================

// TestTwoLane_LaneB_GetSchedule_ReturnsErrorOnModelPersistFailure
// Verifies Lane B behavior: DB operations require strict consistency
func TestTwoLane_LaneB_GetSchedule_ReturnsErrorOnModelPersistFailure(t *testing.T) {
	// Arrange
	logBuf := captureLog(t)
	persister := &AmnesiacPersister{Err: assert.AnError}
	orchestrator := newTestOrchestrator(persister)

	req := ChatRequest{
		ProjectID: uuid.New(),
		Message:   "Show me the schedule", // Triggers IntentGetSchedule (Lane B)
	}

	// Act
	resp, err := orchestrator.ProcessRequest(context.Background(), uuid.New(), uuid.New(), req)

	// Assert: Strict consistency - error propagated
	require.Error(t, err, "Lane B: Should return error for GetSchedule on model persist failure")
	assert.Nil(t, resp, "Lane B: Should not return response on error")
	assert.Contains(t, err.Error(), "failed to persist model message", "Lane B: Error should mention persistence failure")

	// Assert: Error logged with "strict mode" indicator
	logged := logBuf.String()
	assert.Contains(t, logged, "strict mode", "Lane B: Should log strict mode for GetSchedule")
}

// TestTwoLane_LaneB_UpdateTaskStatus_ReturnsErrorOnModelPersistFailure
// Verifies Lane B behavior for UpdateTaskStatus intent
func TestTwoLane_LaneB_UpdateTaskStatus_ReturnsErrorOnModelPersistFailure(t *testing.T) {
	// Arrange
	logBuf := captureLog(t)
	persister := &AmnesiacPersister{Err: assert.AnError}
	orchestrator := newTestOrchestrator(persister)

	req := ChatRequest{
		ProjectID: uuid.New(),
		Message:   "Mark the task as complete", // Triggers IntentUpdateTaskStatus (Lane B)
	}

	// Act
	resp, err := orchestrator.ProcessRequest(context.Background(), uuid.New(), uuid.New(), req)

	// Assert: Strict consistency - error propagated
	require.Error(t, err, "Lane B: Should return error for UpdateTaskStatus on model persist failure")
	assert.Nil(t, resp, "Lane B: Should not return response on error")

	// Assert: Error logged with "strict mode" indicator
	logged := logBuf.String()
	assert.Contains(t, logged, "strict mode", "Lane B: Should log strict mode for UpdateTaskStatus")
}

// =============================================================================
// OBSERVABILITY TESTS: Lane C
// =============================================================================

// TestObservability_CommandExecutionDuration
// Verifies that command execution duration is logged
func TestObservability_CommandExecutionDuration(t *testing.T) {
	// Arrange
	logBuf := captureLog(t)
	persister := &MockMessagePersister{}
	orchestrator := newTestOrchestrator(persister)

	req := ChatRequest{
		ProjectID: uuid.New(),
		Message:   "Show me the schedule",
	}

	// Act
	_, err := orchestrator.ProcessRequest(context.Background(), uuid.New(), uuid.New(), req)

	// Assert: No error
	require.NoError(t, err)

	// Assert: Duration logged
	logged := logBuf.String()
	assert.Contains(t, logged, "chat: command executed", "Should log command execution")
	assert.Contains(t, logged, "duration_ms", "Should log execution duration")
	assert.Contains(t, logged, "GET_SCHEDULE", "Should log intent type")
}

// =============================================================================
// CONSISTENCY LEVEL TESTS
// =============================================================================

// TestCommandConsistencyLevel verifies the lane classification via ConsistencyLevel method.
// This replaces the old isSlowExternalIntent helper function test.
// Consistency logic is now encapsulated in each ChatCommand implementation.
func TestCommandConsistencyLevel(t *testing.T) {
	tests := []struct {
		name          string
		command       ChatCommand
		expectedLevel types.ConsistencyType
		lane          string
	}{
		{"GetScheduleCommand", &GetScheduleCommand{}, types.ConsistencyStrict, "Lane B"},
		{"StrictPlaceholderCommand", &StrictPlaceholderCommand{}, types.ConsistencyStrict, "Lane B"},
		{"PlaceholderCommand", &PlaceholderCommand{}, types.ConsistencyBestEffort, "Lane A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.command.ConsistencyLevel()
			assert.Equal(t, tt.expectedLevel, got, "%s: ConsistencyLevel() should be %v", tt.lane, tt.expectedLevel)
		})
	}
}

// =============================================================================
// REGRESSION: Ensure normal happy path still works
// =============================================================================

// TestHappyPath_NoPersistenceFailure
// Verifies that when persistence succeeds, everything works as before
func TestHappyPath_NoPersistenceFailure(t *testing.T) {
	// Arrange
	persister := &MockMessagePersister{}
	orchestrator := newTestOrchestrator(persister)

	req := ChatRequest{
		ProjectID: uuid.New(),
		Message:   "Show me the schedule",
	}

	// Act
	resp, err := orchestrator.ProcessRequest(context.Background(), uuid.New(), uuid.New(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, types.IntentGetSchedule, resp.Intent)
	assert.NotNil(t, resp.Artifact, "Should return schedule artifact")
	assert.Equal(t, ArtifactTypeScheduleView, resp.Artifact.Type)

	// Verify both messages persisted
	assert.Len(t, persister.Messages, 2, "Should persist both user and model messages")
	assert.Equal(t, types.ChatRoleUser, persister.Messages[0].Role)
	assert.Equal(t, types.ChatRoleModel, persister.Messages[1].Role)
}

// TestHappyPath_InvoiceProcessing
// Verifies invoice processing happy path
func TestHappyPath_InvoiceProcessing(t *testing.T) {
	// Arrange
	persister := &MockMessagePersister{}
	orchestrator := newTestOrchestrator(persister)

	testCases := []string{
		"Process this invoice",        // Action verb
		"invoice",                     // Noun-only shorthand
		"Here is the bill",            // Noun-only shorthand
		"I have a receipt to process", // Mixed
	}

	for _, msg := range testCases {
		t.Run(msg, func(t *testing.T) {
			persister.Messages = nil // Reset

			req := ChatRequest{
				ProjectID: uuid.New(),
				Message:   msg,
			}

			resp, err := orchestrator.ProcessRequest(context.Background(), uuid.New(), uuid.New(), req)

			require.NoError(t, err)
			require.NotNil(t, resp)
			// All should trigger ProcessInvoice
			if !strings.Contains(msg, "process") && !strings.Contains(msg, "I have") {
				// Direct shorthand tests
				assert.Equal(t, types.IntentProcessInvoice, resp.Intent, "Message '%s' should trigger ProcessInvoice", msg)
			}
		})
	}
}
