// Package agents provides AI agent implementations for FutureBuild.
// This file defines the ProcurementRepository interface for testability.
// See PRODUCTION_PLAN.md: Testing Strategy & CI Reliability Remediation
package agents

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ProcurementRepository abstracts database operations for procurement.
// Enables unit testing of business logic without a database connection.
//
// FAANG Standard: Depend on abstractions, not concretions.
// This allows ProcurementAgent logic to be tested with mocks.
type ProcurementRepository interface {
	// StreamItems iterates through active procurement items via callback.
	// Uses cursor-based iteration to prevent OOM at scale.
	StreamItems(ctx context.Context, process ItemProcessor) error

	// UpdateBatch persists status updates in a single transaction.
	// Reduces N database round-trips to 1 per batch.
	UpdateBatch(ctx context.Context, now time.Time, batch []alertResult) error

	// HydrateProject populates procurement_items for a new project.
	// Called via event-driven task on project creation.
	HydrateProject(ctx context.Context, projectID uuid.UUID) error

	// ShouldSendNotification checks 72-hour dampening window.
	// Returns true if no notification was sent in the last 72 hours.
	ShouldSendNotification(ctx context.Context, itemID uuid.UUID, now time.Time) (bool, error)

	// LogNotification persists alert to communication_logs.
	LogNotification(ctx context.Context, result alertResult, now time.Time) error

	// GetNotificationHistoryForBatch retrieves dampening status for multiple items in one query.
	// Returns a map where true = notification was sent in last 72 hours (should dampen).
	// P0 Performance Fix: Reduces N database round-trips to 1.
	GetNotificationHistoryForBatch(ctx context.Context, itemIDs []uuid.UUID, now time.Time) (map[uuid.UUID]bool, error)

	// LogNotificationsBatch persists multiple alerts to communication_logs in a single operation.
	// P1 Performance Fix: Reduces N database round-trips to 1 per batch.
	LogNotificationsBatch(ctx context.Context, results []alertResult, now time.Time) error
}
