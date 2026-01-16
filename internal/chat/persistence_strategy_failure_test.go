package chat

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/colton/futurebuild/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// STRICT MOCKS FOR FAILURE MODE TESTING
// =============================================================================
// These mocks track call counts and allow configurable success/error behavior.
// Used to verify side effects occur exactly once per failure scenario.
//
// FAANG Quality Bar:
// - Mutation Testing: If dlq.EnqueueRetry call is removed, tests MUST fail.
// - Determinism: No race conditions, thread-safe call counters.
// - Independence: No Docker/DB/Redis dependencies - pure unit tests.
// =============================================================================

// --- StrictMockMessagePersister ---

// StrictMockMessagePersister is a strict mock that tracks SaveMessage calls.
type StrictMockMessagePersister struct {
	mu        sync.Mutex
	callCount int
	err       error // Error to return on SaveMessage
}

// NewStrictMockMessagePersister creates a mock that always fails with the given error.
func NewStrictMockMessagePersister(err error) *StrictMockMessagePersister {
	return &StrictMockMessagePersister{err: err}
}

func (m *StrictMockMessagePersister) SaveMessage(_ context.Context, _ models.ChatMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	return m.err
}

func (m *StrictMockMessagePersister) Pool() Transactor {
	return nil
}

// CallCount returns the number of times SaveMessage was called.
func (m *StrictMockMessagePersister) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

// --- StrictMockDLQPersister ---

// StrictMockDLQPersister is a strict mock that tracks EnqueueRetry calls.
type StrictMockDLQPersister struct {
	mu        sync.Mutex
	callCount int
	err       error // Error to return on EnqueueRetry (nil = success)
}

// NewStrictMockDLQPersister creates a mock with configurable success/error.
func NewStrictMockDLQPersister(err error) *StrictMockDLQPersister {
	return &StrictMockDLQPersister{err: err}
}

func (m *StrictMockDLQPersister) EnqueueRetry(_ context.Context, _ models.ChatMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	return m.err
}

// CallCount returns the number of times EnqueueRetry was called.
func (m *StrictMockDLQPersister) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

// --- StrictMockAuditWAL ---

// StrictMockAuditWAL is a strict mock that tracks AppendRecord calls.
type StrictMockAuditWAL struct {
	mu        sync.Mutex
	callCount int
	err       error // Error to return on AppendRecord (nil = success)
}

// NewStrictMockAuditWAL creates a mock with configurable success/error.
func NewStrictMockAuditWAL(err error) *StrictMockAuditWAL {
	return &StrictMockAuditWAL{err: err}
}

func (m *StrictMockAuditWAL) AppendRecord(_ context.Context, _ models.ChatMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	return m.err
}

func (m *StrictMockAuditWAL) Flush() error {
	return nil
}

func (m *StrictMockAuditWAL) Close() error {
	return nil
}

// CallCount returns the number of times AppendRecord was called.
func (m *StrictMockAuditWAL) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

// --- StrictMockCircuitBreaker ---

// StrictMockCircuitBreaker is a strict mock with configurable IsOpen() and call tracking.
type StrictMockCircuitBreaker struct {
	mu                 sync.Mutex
	isOpen             bool
	recordSuccessCount int
	recordFailureCount int
}

// NewStrictMockCircuitBreaker creates a mock with configurable IsOpen() return value.
func NewStrictMockCircuitBreaker(isOpen bool) *StrictMockCircuitBreaker {
	return &StrictMockCircuitBreaker{isOpen: isOpen}
}

func (m *StrictMockCircuitBreaker) RecordSuccess() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordSuccessCount++
}

func (m *StrictMockCircuitBreaker) RecordFailure() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordFailureCount++
}

func (m *StrictMockCircuitBreaker) IsOpen() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isOpen
}

func (m *StrictMockCircuitBreaker) State() CircuitState {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.isOpen {
		return CircuitOpen
	}
	return CircuitClosed
}

// RecordSuccessCount returns the number of times RecordSuccess was called.
func (m *StrictMockCircuitBreaker) RecordSuccessCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.recordSuccessCount
}

// RecordFailureCount returns the number of times RecordFailure was called.
func (m *StrictMockCircuitBreaker) RecordFailureCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.recordFailureCount
}

// =============================================================================
// FAILURE MODE TESTS (Table-Driven)
// =============================================================================
// Test Scenarios:
// 1. DB Down, DLQ Up → Success response, Enqueue called once
// 2. DB Down, DLQ Down, WAL Up → Success response, AppendRecord called once
// 3. Total System Failure → Error to user, circuit breaker trips
// 4. Circuit Breaker Open → Immediate failure without attempting DB
// =============================================================================

func TestBestEffortPersistence_FailureModes(t *testing.T) {
	// Common test fixtures
	dbErr := errors.New("database connection failed")
	dlqErr := errors.New("redis connection failed")
	walErr := errors.New("disk write failed")

	testMsg := models.ChatMessage{
		ID:        uuid.New(),
		ProjectID: uuid.New(),
		Content:   "test message",
	}

	tests := []struct {
		name string
		// Mock configuration
		dbError  error // Error from MessagePersister.SaveMessage
		dlqError error // Error from DLQPersister.EnqueueRetry (nil = success)
		walError error // Error from AuditWAL.AppendRecord (nil = success)
		cbOpen   bool  // Circuit breaker IsOpen() return value
		// Expected outcomes
		wantErr bool
		// Expected side effects (call counts)
		wantDBCalls        int
		wantDLQCalls       int
		wantWALCalls       int
		wantCBSuccessCalls int
		wantCBFailureCalls int
		wantErrContains    string
	}{
		{
			name:               "Scenario 1: DB Down, DLQ Up - Success with DLQ fallback",
			dbError:            dbErr,
			dlqError:           nil, // DLQ succeeds
			walError:           nil,
			cbOpen:             false,
			wantErr:            false,
			wantDBCalls:        1,
			wantDLQCalls:       1, // CRITICAL: DLQ must be called exactly once
			wantWALCalls:       0, // WAL not needed when DLQ succeeds
			wantCBSuccessCalls: 1, // Success recorded after DLQ
			wantCBFailureCalls: 0,
		},
		{
			name:               "Scenario 2: DB Down, DLQ Down, WAL Up - Success with WAL fallback",
			dbError:            dbErr,
			dlqError:           dlqErr, // DLQ fails
			walError:           nil,    // WAL succeeds
			cbOpen:             false,
			wantErr:            false,
			wantDBCalls:        1,
			wantDLQCalls:       1, // DLQ attempted
			wantWALCalls:       1, // CRITICAL: WAL must be called exactly once
			wantCBSuccessCalls: 1, // Success recorded after WAL
			wantCBFailureCalls: 0,
		},
		{
			name:               "Scenario 3: Total System Failure - Error with circuit breaker trip",
			dbError:            dbErr,
			dlqError:           dlqErr,
			walError:           walErr, // WAL also fails
			cbOpen:             false,
			wantErr:            true,
			wantDBCalls:        1,
			wantDLQCalls:       1, // DLQ attempted
			wantWALCalls:       1, // WAL attempted
			wantCBSuccessCalls: 0,
			wantCBFailureCalls: 1, // CRITICAL: Failure must be recorded
			wantErrContains:    "audit system unavailable",
		},
		{
			name:               "Scenario 4: Circuit Breaker Open - Immediate failure",
			dbError:            dbErr,
			dlqError:           nil,
			walError:           nil,
			cbOpen:             true, // Circuit breaker is OPEN
			wantErr:            true,
			wantDBCalls:        1, // DB is still attempted first
			wantDLQCalls:       0, // CRITICAL: DLQ not attempted when CB open
			wantWALCalls:       0, // WAL not attempted when CB open
			wantCBSuccessCalls: 0,
			wantCBFailureCalls: 0, // No additional failure recorded (already open)
			wantErrContains:    "system temporarily unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: Create strict mocks
			mockDB := NewStrictMockMessagePersister(tt.dbError)
			mockDLQ := NewStrictMockDLQPersister(tt.dlqError)
			mockWAL := NewStrictMockAuditWAL(tt.walError)
			mockCB := NewStrictMockCircuitBreaker(tt.cbOpen)

			// Create BestEffortPersistence with all mocks
			strategy := NewBestEffortPersistence(mockDB, mockDLQ, mockWAL, mockCB)

			// Act
			err := strategy.Persist(context.Background(), testMsg)

			// Assert: Return value
			if tt.wantErr {
				require.Error(t, err, "Expected error but got nil")
				assert.Contains(t, err.Error(), tt.wantErrContains, "Error message mismatch")
			} else {
				require.NoError(t, err, "Expected success but got error: %v", err)
			}

			// Assert: Side effects (call counts) - CRITICAL for mutation testing
			assert.Equal(t, tt.wantDBCalls, mockDB.CallCount(),
				"DB SaveMessage call count mismatch")
			assert.Equal(t, tt.wantDLQCalls, mockDLQ.CallCount(),
				"DLQ EnqueueRetry call count mismatch - if this fails, check dlq.EnqueueRetry call exists")
			assert.Equal(t, tt.wantWALCalls, mockWAL.CallCount(),
				"WAL AppendRecord call count mismatch")
			assert.Equal(t, tt.wantCBSuccessCalls, mockCB.RecordSuccessCount(),
				"CircuitBreaker RecordSuccess call count mismatch")
			assert.Equal(t, tt.wantCBFailureCalls, mockCB.RecordFailureCount(),
				"CircuitBreaker RecordFailure call count mismatch")
		})
	}
}

// TestBestEffortPersistence_HappyPath verifies the happy path still works.
func TestBestEffortPersistence_HappyPath(t *testing.T) {
	// Arrange: All systems up
	mockDB := NewStrictMockMessagePersister(nil) // DB succeeds
	mockDLQ := NewStrictMockDLQPersister(nil)
	mockWAL := NewStrictMockAuditWAL(nil)
	mockCB := NewStrictMockCircuitBreaker(false)

	strategy := NewBestEffortPersistence(mockDB, mockDLQ, mockWAL, mockCB)
	testMsg := models.ChatMessage{ID: uuid.New(), Content: "test"}

	// Act
	err := strategy.Persist(context.Background(), testMsg)

	// Assert: Success with no fallbacks
	require.NoError(t, err)
	assert.Equal(t, 1, mockDB.CallCount(), "DB should be called once")
	assert.Equal(t, 0, mockDLQ.CallCount(), "DLQ should not be called on success")
	assert.Equal(t, 0, mockWAL.CallCount(), "WAL should not be called on success")
	assert.Equal(t, 1, mockCB.RecordSuccessCount(), "Success should be recorded")
}

// TestStrictPersistence_FailsOnDBError verifies StrictPersistence behavior.
func TestStrictPersistence_FailsOnDBError(t *testing.T) {
	// Arrange
	dbErr := errors.New("database connection failed")
	mockDB := NewStrictMockMessagePersister(dbErr)
	strategy := NewStrictPersistence(mockDB)
	testMsg := models.ChatMessage{ID: uuid.New(), Content: "test"}

	// Act
	err := strategy.Persist(context.Background(), testMsg)

	// Assert: Error propagated
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to persist model message")
	assert.Equal(t, 1, mockDB.CallCount())
}

// TestStrictPersistence_SucceedsOnDBSuccess verifies StrictPersistence happy path.
func TestStrictPersistence_SucceedsOnDBSuccess(t *testing.T) {
	// Arrange
	mockDB := NewStrictMockMessagePersister(nil) // DB succeeds
	strategy := NewStrictPersistence(mockDB)
	testMsg := models.ChatMessage{ID: uuid.New(), Content: "test"}

	// Act
	err := strategy.Persist(context.Background(), testMsg)

	// Assert: Success
	require.NoError(t, err)
	assert.Equal(t, 1, mockDB.CallCount())
}

// =============================================================================
// MUTATION TESTING VERIFICATION
// =============================================================================
// These tests are designed to fail if critical code paths are removed.
// Run: go test -v -run TestMutation ./internal/chat/...
// =============================================================================

// TestMutation_DLQEnqueueMustBeCalled verifies tests fail if DLQ call is removed.
// HOWTO: Comment out `s.dlq.EnqueueRetry(ctx, msg)` in persistence_strategy.go.
// This test MUST fail if that line is commented out.
func TestMutation_DLQEnqueueMustBeCalled(t *testing.T) {
	// Arrange: DB fails, DLQ would succeed
	mockDB := NewStrictMockMessagePersister(errors.New("db error"))
	mockDLQ := NewStrictMockDLQPersister(nil) // DLQ succeeds
	mockWAL := NewStrictMockAuditWAL(nil)
	mockCB := NewStrictMockCircuitBreaker(false)

	strategy := NewBestEffortPersistence(mockDB, mockDLQ, mockWAL, mockCB)
	testMsg := models.ChatMessage{ID: uuid.New()}

	// Act
	err := strategy.Persist(context.Background(), testMsg)

	// Assert: DLQ MUST be called for this test to pass
	// If EnqueueRetry call is removed, this assertion will fail
	require.NoError(t, err, "Should succeed via DLQ fallback")
	require.Equal(t, 1, mockDLQ.CallCount(),
		"MUTATION TEST FAILED: DLQ.EnqueueRetry was not called. "+
			"Verify the dlq.EnqueueRetry(ctx, msg) call exists in BestEffortPersistence.Persist()")
}

// TestMutation_WALAppendMustBeCalled verifies tests fail if WAL call is removed.
// HOWTO: Comment out `s.wal.AppendRecord(ctx, msg)` in persistence_strategy.go.
// This test MUST fail if that line is commented out.
func TestMutation_WALAppendMustBeCalled(t *testing.T) {
	// Arrange: DB fails, DLQ fails, WAL would succeed
	mockDB := NewStrictMockMessagePersister(errors.New("db error"))
	mockDLQ := NewStrictMockDLQPersister(errors.New("dlq error"))
	mockWAL := NewStrictMockAuditWAL(nil) // WAL succeeds
	mockCB := NewStrictMockCircuitBreaker(false)

	strategy := NewBestEffortPersistence(mockDB, mockDLQ, mockWAL, mockCB)
	testMsg := models.ChatMessage{ID: uuid.New()}

	// Act
	err := strategy.Persist(context.Background(), testMsg)

	// Assert: WAL MUST be called for this test to pass
	require.NoError(t, err, "Should succeed via WAL fallback")
	require.Equal(t, 1, mockWAL.CallCount(),
		"MUTATION TEST FAILED: WAL.AppendRecord was not called. "+
			"Verify the wal.AppendRecord(ctx, msg) call exists in BestEffortPersistence.Persist()")
}

// TestMutation_CircuitBreakerTripsMustBeCalled verifies CB failure recording.
// HOWTO: Comment out `s.circuitBreaker.RecordFailure()` in persistence_strategy.go.
// This test MUST fail if that line is commented out.
func TestMutation_CircuitBreakerTripsMustBeCalled(t *testing.T) {
	// Arrange: All systems fail
	mockDB := NewStrictMockMessagePersister(errors.New("db error"))
	mockDLQ := NewStrictMockDLQPersister(errors.New("dlq error"))
	mockWAL := NewStrictMockAuditWAL(errors.New("wal error"))
	mockCB := NewStrictMockCircuitBreaker(false)

	strategy := NewBestEffortPersistence(mockDB, mockDLQ, mockWAL, mockCB)
	testMsg := models.ChatMessage{ID: uuid.New()}

	// Act
	err := strategy.Persist(context.Background(), testMsg)

	// Assert: CB RecordFailure MUST be called
	require.Error(t, err)
	require.Equal(t, 1, mockCB.RecordFailureCount(),
		"MUTATION TEST FAILED: CircuitBreaker.RecordFailure was not called. "+
			"Verify the circuitBreaker.RecordFailure() call exists in BestEffortPersistence.Persist()")
}
