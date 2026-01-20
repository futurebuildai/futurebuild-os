// Package agents provides AI agent implementations for FutureBuild.
// This file provides a mock implementation of ProcurementRepository for testing.
// See PRODUCTION_PLAN.md: Testing Strategy & CI Reliability Remediation
package agents

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MockProcurementRepository implements ProcurementRepository for unit testing.
// Enables testing of ProcurementAgent business logic without a database.
type MockProcurementRepository struct {
	// Items to return from StreamItems
	Items []procurementRow

	// Recorded calls for assertions
	UpdatedBatches   [][]alertResult
	LoggedNotifs     []alertResult
	HydratedProjects []uuid.UUID

	// Error injection for testing error handling
	StreamItemsErr     error
	UpdateBatchErr     error
	HydrateProjectErr  error
	ShouldSendNotifErr error
	LogNotificationErr error

	// ShouldSendNotification return value (default true)
	ShouldSendNotifResult bool
}

// NewMockProcurementRepository creates a mock repository with sensible defaults.
func NewMockProcurementRepository() *MockProcurementRepository {
	return &MockProcurementRepository{
		Items:                 []procurementRow{},
		UpdatedBatches:        [][]alertResult{},
		LoggedNotifs:          []alertResult{},
		HydratedProjects:      []uuid.UUID{},
		ShouldSendNotifResult: true, // Default: allow notifications
	}
}

// StreamItems iterates through mock items via callback.
func (m *MockProcurementRepository) StreamItems(ctx context.Context, process ItemProcessor) error {
	if m.StreamItemsErr != nil {
		return m.StreamItemsErr
	}
	for _, item := range m.Items {
		if err := process(item); err != nil {
			return err
		}
	}
	return nil
}

// UpdateBatch records the batch for later assertions.
func (m *MockProcurementRepository) UpdateBatch(ctx context.Context, now time.Time, batch []alertResult) error {
	if m.UpdateBatchErr != nil {
		return m.UpdateBatchErr
	}
	// Make a copy to avoid mutation issues
	batchCopy := make([]alertResult, len(batch))
	copy(batchCopy, batch)
	m.UpdatedBatches = append(m.UpdatedBatches, batchCopy)
	return nil
}

// HydrateProject records the project ID for later assertions.
func (m *MockProcurementRepository) HydrateProject(ctx context.Context, projectID uuid.UUID) error {
	if m.HydrateProjectErr != nil {
		return m.HydrateProjectErr
	}
	m.HydratedProjects = append(m.HydratedProjects, projectID)
	return nil
}

// ShouldSendNotification returns the configured result.
func (m *MockProcurementRepository) ShouldSendNotification(ctx context.Context, itemID uuid.UUID, now time.Time) (bool, error) {
	if m.ShouldSendNotifErr != nil {
		return false, m.ShouldSendNotifErr
	}
	return m.ShouldSendNotifResult, nil
}

// LogNotification records the notification for later assertions.
func (m *MockProcurementRepository) LogNotification(ctx context.Context, result alertResult, now time.Time) error {
	if m.LogNotificationErr != nil {
		return m.LogNotificationErr
	}
	m.LoggedNotifs = append(m.LoggedNotifs, result)
	return nil
}

// --- Test Helpers ---

// GetAllUpdatedResults returns all results across all batches.
func (m *MockProcurementRepository) GetAllUpdatedResults() []alertResult {
	var all []alertResult
	for _, batch := range m.UpdatedBatches {
		all = append(all, batch...)
	}
	return all
}

// CountStatusUpdates counts how many items were updated to a specific status.
func (m *MockProcurementRepository) CountStatusUpdates(status string) int {
	count := 0
	for _, result := range m.GetAllUpdatedResults() {
		if string(result.NewStatus) == status {
			count++
		}
	}
	return count
}

// LogNotificationsBatch records the batch for later assertions.
func (m *MockProcurementRepository) LogNotificationsBatch(ctx context.Context, results []alertResult, now time.Time) error {
	if m.LogNotificationErr != nil {
		return m.LogNotificationErr
	}
	m.LoggedNotifs = append(m.LoggedNotifs, results...)
	return nil
}
