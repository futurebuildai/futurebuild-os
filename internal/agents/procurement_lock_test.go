// Package agents provides AI agent implementations for FutureBuild.
// This file contains tests for distributed locking in concurrent agent execution.
// See L7 Technical Debt Remediation: Task 2
package agents

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/pkg/clock"
	pkgsync "github.com/colton/futurebuild/pkg/sync"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

// =============================================================================
// RED TEST: Task 2 - Prove Race Condition Vulnerability in Execute
// =============================================================================
// This test proves that without distributed locking, multiple concurrent
// Execute calls will all process the same items (race condition).
//
// EXPECTED BEHAVIOR (BEFORE FIX): This test FAILS because executionCount > 1
// EXPECTED BEHAVIOR (AFTER FIX): This test PASSES because exactly 1 execution runs
// =============================================================================

// concurrencyTrackingRepository tracks how many times Execute processes items.
type concurrencyTrackingRepository struct {
	ProcurementRepository
	executionCount atomic.Int32
	processedIDs   sync.Map // Thread-safe map to track processed item IDs
}

func (r *concurrencyTrackingRepository) StreamItems(ctx context.Context, process ItemProcessor) error {
	r.executionCount.Add(1)

	// Simulate processing 5 items
	for i := 0; i < 5; i++ {
		earlyStart := time.Now().Add(24 * time.Hour)
		zipCode := "78701"
		id := uuid.New()

		// Track that this ID was processed (for detecting duplicates)
		r.processedIDs.Store(id.String(), true)

		item := procurementRow{
			ID:            id,
			Name:          "Test Item",
			LeadTimeWeeks: 1,
			Status:        types.ProcurementAlertPending,
			EarlyStart:    &earlyStart,
			ZipCode:       &zipCode,
		}
		if err := process(item); err != nil {
			return err
		}
	}
	return nil
}

func (r *concurrencyTrackingRepository) UpdateBatch(ctx context.Context, now time.Time, batch []alertResult) error {
	return nil
}

func (r *concurrencyTrackingRepository) HydrateProject(ctx context.Context, projectID uuid.UUID) error {
	return nil
}

func (r *concurrencyTrackingRepository) ShouldSendNotification(ctx context.Context, itemID uuid.UUID, now time.Time) (bool, error) {
	return false, nil
}

func (r *concurrencyTrackingRepository) LogNotification(ctx context.Context, result alertResult, now time.Time) error {
	return nil
}

func (r *concurrencyTrackingRepository) GetNotificationHistoryForBatch(ctx context.Context, itemIDs []uuid.UUID, now time.Time) (map[uuid.UUID]bool, error) {
	return make(map[uuid.UUID]bool), nil
}

// mockBlockingMutex simulates a distributed lock that blocks concurrent execution.
// Uses a real sync.Mutex internally to ensure atomic lock acquisition.
// IMPORTANT: The lock is NEVER released during test to simulate a long-running process.
// This ensures the test is fully deterministic.
type mockBlockingMutex struct {
	mu           sync.Mutex
	lockHeld     bool
	lockAttempts atomic.Int32
}

func (m *mockBlockingMutex) TryLock(ctx context.Context, key string, ttl time.Duration) (pkgsync.UnlockFunc, error) {
	m.lockAttempts.Add(1)

	m.mu.Lock()
	if m.lockHeld {
		m.mu.Unlock()
		return nil, pkgsync.ErrLockHeld
	}

	m.lockHeld = true
	m.mu.Unlock()

	// Return a no-op unlock - lock stays held for test duration
	// This simulates a long-running process that holds the lock
	return func() error {
		// Lock stays held - this simulates the "winner" still running
		// while others try to acquire
		return nil
	}, nil
}

// TestExecute_ConcurrentInstances_RedTest proves race condition vulnerability.
//
// This test simulates a Blue/Green deployment where both instances call Execute.
// BEFORE FIX: Both instances process all items (race condition)
// AFTER FIX: Only one instance processes items, the other gracefully skips
func TestExecute_ConcurrentInstances_RedTest(t *testing.T) {
	repo := &concurrencyTrackingRepository{}
	mutex := &mockBlockingMutex{}

	cfg := config.DefaultProcurementConfig()

	// Create agent WITH distributed mutex
	agent := &ProcurementAgent{
		repo:      repo,
		weather:   &mockWeatherService{forecast: types.Forecast{PrecipitationProbability: 0.1}},
		clock:     clock.RealClock{},
		batchSize: 100,
		config:    cfg.WithDefaults(),
		mutex:     mutex,
	}

	// Use a barrier to ensure all goroutines start simultaneously
	const numGoroutines = 3
	var wg sync.WaitGroup
	startBarrier := make(chan struct{})
	executionResults := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-startBarrier // Wait for start signal
			executionResults[idx] = agent.Execute(context.Background())
		}(i)
	}

	// Let goroutines settle, then release them all at once
	time.Sleep(10 * time.Millisecond)
	close(startBarrier)

	wg.Wait()

	// All calls should succeed (lock failure is graceful, not an error)
	for i, err := range executionResults {
		if err != nil {
			t.Errorf("Execute %d failed unexpectedly: %v", i, err)
		}
	}

	executionCount := repo.executionCount.Load()
	lockAttempts := mutex.lockAttempts.Load()

	t.Logf("Lock attempts: %d, Actual executions: %d", lockAttempts, executionCount)

	// =============================================================================
	// CRITICAL ASSERTION: Only ONE instance should process items
	// =============================================================================
	// BEFORE FIX: executionCount == 3 (all instances race) - TEST FAILS
	// AFTER FIX:  executionCount == 1 (only winner processes) - TEST PASSES
	// =============================================================================

	if executionCount != 1 {
		t.Errorf("RACE CONDITION DETECTED: %d concurrent executions occurred (expected exactly 1)", executionCount)
		t.Error("This would cause duplicate processing and 'double-spend' notification spam in production!")
	}

	// All should have attempted to acquire lock
	if lockAttempts != int32(numGoroutines) {
		t.Errorf("Expected %d lock attempts, got %d", numGoroutines, lockAttempts)
	}
}

// TestExecute_NoMutex_SingleInstance tests backward compatibility when mutex is nil.
func TestExecute_NoMutex_SingleInstance(t *testing.T) {
	repo := &concurrencyTrackingRepository{}
	cfg := config.DefaultProcurementConfig()

	// Create agent WITHOUT distributed mutex (single-instance mode)
	agent := &ProcurementAgent{
		repo:      repo,
		weather:   &mockWeatherService{forecast: types.Forecast{PrecipitationProbability: 0.1}},
		clock:     clock.RealClock{},
		batchSize: 100,
		config:    cfg.WithDefaults(),
		mutex:     nil, // No mutex = single-instance mode
	}

	err := agent.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should execute normally without mutex
	if count := repo.executionCount.Load(); count != 1 {
		t.Errorf("Expected 1 execution, got %d", count)
	}
}

// TestExecute_LockFailure_GracefulExit tests that lock contention is NOT an error.
func TestExecute_LockFailure_GracefulExit(t *testing.T) {
	repo := &concurrencyTrackingRepository{}
	mutex := &mockBlockingMutex{}

	// Pre-acquire the lock to simulate another instance holding it
	mutex.mu.Lock()
	mutex.lockHeld = true
	mutex.mu.Unlock()

	cfg := config.DefaultProcurementConfig()
	agent := &ProcurementAgent{
		repo:      repo,
		weather:   &mockWeatherService{forecast: types.Forecast{PrecipitationProbability: 0.1}},
		clock:     clock.RealClock{},
		batchSize: 100,
		config:    cfg.WithDefaults(),
		mutex:     mutex,
	}

	// Execute should return nil (graceful exit), NOT an error
	err := agent.Execute(context.Background())
	if err != nil {
		t.Errorf("Execute returned error when lock was held: %v (expected graceful nil return)", err)
	}

	// Should NOT have processed any items
	if count := repo.executionCount.Load(); count != 0 {
		t.Errorf("Expected 0 executions when lock held, got %d", count)
	}
}
