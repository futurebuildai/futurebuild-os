// Package agents provides AI agent implementations for FutureBuild.
// This file contains tests for N+1 query pattern detection and remediation.
// See L7 Technical Debt Remediation: Task 1
package agents

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/pkg/clock"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

// =============================================================================
// RED TEST: Task 1 - Prove N+1 Query Pattern Exists in flushBatch
// =============================================================================
// This test proves that the current implementation calls ShouldSendNotification
// once per item in the batch, resulting in O(N) database queries.
//
// EXPECTED BEHAVIOR (BEFORE FIX): This test FAILS because queriesExecuted > 1
// EXPECTED BEHAVIOR (AFTER FIX): This test PASSES because queriesExecuted = 1
// =============================================================================

// queryCountingRepository wraps a mock repository and counts DB calls.
// This is the "observation probe" for detecting N+1 patterns.
type queryCountingRepository struct {
	ProcurementRepository // Embed interface for default behavior
	shouldSendCalls       atomic.Int32
	batchHistoryCalls     atomic.Int32
}

func (r *queryCountingRepository) StreamItems(ctx context.Context, process ItemProcessor) error {
	// Simulate 10 items that all need notifications
	for i := 0; i < 10; i++ {
		earlyStart := time.Now().Add(24 * time.Hour) // Tomorrow
		zipCode := "78701"
		item := procurementRow{
			ID:            uuid.New(),
			Name:          fmt.Sprintf("Test Item %d", i),
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

func (r *queryCountingRepository) UpdateBatch(ctx context.Context, now time.Time, batch []alertResult) error {
	// No-op: we're testing notification queries, not updates
	return nil
}

func (r *queryCountingRepository) HydrateProject(ctx context.Context, projectID uuid.UUID) error {
	return nil
}

// ShouldSendNotification - THIS IS THE N+1 PROBLEM
// Each call here = 1 database query in production
func (r *queryCountingRepository) ShouldSendNotification(ctx context.Context, itemID uuid.UUID, now time.Time) (bool, error) {
	r.shouldSendCalls.Add(1) // Count the query
	return true, nil         // Always say "yes, send notification" to trigger the code path
}

func (r *queryCountingRepository) LogNotification(ctx context.Context, result alertResult, now time.Time) error {
	return nil
}

// GetNotificationHistoryForBatch - THE FIX: Batch query
// This method should be called ONCE for all items in the batch
func (r *queryCountingRepository) GetNotificationHistoryForBatch(ctx context.Context, itemIDs []uuid.UUID, now time.Time) (map[uuid.UUID]bool, error) {
	r.batchHistoryCalls.Add(1) // Count the batch query
	// Return empty map = no dampening = all notifications should send
	return make(map[uuid.UUID]bool), nil
}

// TestFlushBatch_QueryCount_RedTest proves N+1 pattern exists.
//
// This test will FAIL before the fix is applied because:
// - Current flushBatch calls shouldSendNotification per item
// - With 10 items, we expect 10 calls to ShouldSendNotification
//
// After the fix:
// - flushBatch should call GetNotificationHistoryForBatch ONCE
// - shouldSendCalls should be 0
// - batchHistoryCalls should be 1
func TestFlushBatch_QueryCount_RedTest(t *testing.T) {
	repo := &queryCountingRepository{}

	// Create agent with config that triggers notifications
	cfg := config.ProcurementConfig{
		StagingBufferDays:        2,
		LeadTimeWarningThreshold: 3,
		DefaultWeatherBufferDays: 3,
	}

	agent := &ProcurementAgent{
		repo:      repo,
		weather:   &mockWeatherService{forecast: types.Forecast{PrecipitationProbability: 0.1}},
		clock:     clock.RealClock{},
		batchSize: 100,
		config:    cfg.WithDefaults(),
	}

	// Execute the agent - this will stream items and flush batches
	ctx := context.Background()
	err := agent.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// =============================================================================
	// ASSERTION: The N+1 pattern proof
	// =============================================================================
	// BEFORE FIX: shouldSendCalls == 10 (one per item) - TEST FAILS
	// AFTER FIX:  shouldSendCalls == 0, batchHistoryCalls == 1 - TEST PASSES
	// =============================================================================

	shouldSendCalls := repo.shouldSendCalls.Load()
	batchHistoryCalls := repo.batchHistoryCalls.Load()

	t.Logf("Query counts - ShouldSendNotification: %d, GetNotificationHistoryForBatch: %d",
		shouldSendCalls, batchHistoryCalls)

	// The fix should eliminate per-item calls
	if shouldSendCalls > 0 {
		t.Errorf("N+1 PATTERN DETECTED: ShouldSendNotification was called %d times (expected 0 after fix)",
			shouldSendCalls)
	}

	// The fix should use exactly 1 batch query (or 0 if no items need notification)
	if batchHistoryCalls == 0 && shouldSendCalls == 0 {
		// Edge case: no notifications needed at all (OK)
		t.Log("No notifications needed - batch query not required")
	} else if batchHistoryCalls != 1 {
		t.Errorf("Expected exactly 1 batch query, got %d", batchHistoryCalls)
	}
}

// TestFlushBatch_EmptyBatch_NoQueries ensures no DB calls for empty batches.
func TestFlushBatch_EmptyBatch_NoQueries(t *testing.T) {
	repo := &queryCountingRepository{}

	cfg := config.DefaultProcurementConfig()
	agent := &ProcurementAgent{
		repo:      repo,
		clock:     clock.RealClock{},
		batchSize: 100,
		config:    cfg,
	}

	// Flush an empty batch
	ctx := context.Background()
	err := agent.flushBatch(ctx, []alertResult{})
	if err != nil {
		t.Fatalf("flushBatch failed: %v", err)
	}

	// Empty batch should make ZERO database calls
	if calls := repo.shouldSendCalls.Load(); calls != 0 {
		t.Errorf("Empty batch should not call ShouldSendNotification, got %d calls", calls)
	}
	if calls := repo.batchHistoryCalls.Load(); calls != 0 {
		t.Errorf("Empty batch should not call GetNotificationHistoryForBatch, got %d calls", calls)
	}
}

// TestFlushBatch_BatchOf500_SingleQuery proves O(1) complexity at scale.
func TestFlushBatch_BatchOf500_SingleQuery(t *testing.T) {
	repo := &queryCountingRepository{}

	cfg := config.DefaultProcurementConfig()
	agent := &ProcurementAgent{
		repo:      repo,
		clock:     clock.RealClock{},
		batchSize: 500,
		config:    cfg,
	}

	// Create 500 items that all need notifications
	batch := make([]alertResult, 500)
	for i := range batch {
		batch[i] = alertResult{
			ID:           uuid.New(),
			NewStatus:    types.ProcurementAlertCritical,
			ShouldNotify: true,
			Message:      "Test notification",
		}
	}

	ctx := context.Background()
	err := agent.flushBatch(ctx, batch)
	if err != nil {
		t.Fatalf("flushBatch failed: %v", err)
	}

	shouldSendCalls := repo.shouldSendCalls.Load()
	batchHistoryCalls := repo.batchHistoryCalls.Load()

	t.Logf("With 500 items - ShouldSendNotification: %d, GetNotificationHistoryForBatch: %d",
		shouldSendCalls, batchHistoryCalls)

	// =============================================================================
	// CRITICAL ASSERTION: O(1) Database Complexity
	// =============================================================================
	// BEFORE FIX: shouldSendCalls == 500 (N+1 pattern) - TEST FAILS
	// AFTER FIX:  batchHistoryCalls == 1 (O(1) complexity) - TEST PASSES
	// =============================================================================

	if shouldSendCalls > 0 {
		t.Errorf("P0 PERFORMANCE BUG: %d DB queries executed (expected 1 batch query)",
			shouldSendCalls)
		t.Errorf("This would cause 500 round-trips to the database per batch!")
	}

	if batchHistoryCalls != 1 {
		t.Errorf("Expected exactly 1 batch query for 500 items, got %d", batchHistoryCalls)
	}
}
