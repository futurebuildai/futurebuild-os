package models

import (
	"time"

	"github.com/google/uuid"
)

// CorporateBudget aggregates all project budgets for org-wide financial health.
// See BACKEND_SCOPE.md Section 20.1
// MONETARY PRECISION: All amounts stored as int64 cents to prevent IEEE 754 float drift.
type CorporateBudget struct {
	ID                  uuid.UUID `json:"id" db:"id"`
	OrgID               uuid.UUID `json:"org_id" db:"org_id"`
	FiscalYear          int       `json:"fiscal_year" db:"fiscal_year"`
	Quarter             int       `json:"quarter" db:"quarter"`
	TotalEstimatedCents int64     `json:"total_estimated_cents" db:"total_estimated_cents"`
	TotalCommittedCents int64     `json:"total_committed_cents" db:"total_committed_cents"`
	TotalActualCents    int64     `json:"total_actual_cents" db:"total_actual_cents"`
	ProjectCount        int       `json:"project_count" db:"project_count"`
	LastRollupAt        time.Time `json:"last_rollup_at" db:"last_rollup_at"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// GLSyncStatus defines the state of a GL sync operation.
type GLSyncStatus string

const (
	GLSyncStatusPending   GLSyncStatus = "pending"
	GLSyncStatusCompleted GLSyncStatus = "completed"
	GLSyncStatusFailed    GLSyncStatus = "failed"
)

// GLSyncLog tracks export to external ERP systems for audit trail.
// See BACKEND_SCOPE.md Section 20.1
type GLSyncLog struct {
	ID            uuid.UUID    `json:"id" db:"id"`
	OrgID         uuid.UUID    `json:"org_id" db:"org_id"`
	SyncType      string       `json:"sync_type" db:"sync_type"`
	Status        GLSyncStatus `json:"status" db:"status"`
	RecordsSynced int          `json:"records_synced" db:"records_synced"`
	ErrorMessage  *string      `json:"error_message,omitempty" db:"error_message"`
	SyncedAt      *time.Time   `json:"synced_at,omitempty" db:"synced_at"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
}

// ARAgingSnapshot provides cash flow visibility by standard aging buckets.
// See BACKEND_SCOPE.md Section 20.1
// MONETARY PRECISION: All amounts stored as int64 cents.
type ARAgingSnapshot struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	OrgID                uuid.UUID `json:"org_id" db:"org_id"`
	SnapshotDate         time.Time `json:"snapshot_date" db:"snapshot_date"`
	CurrentCents         int64     `json:"current_cents" db:"current_cents"`
	Days30Cents          int64     `json:"days_30_cents" db:"days_30_cents"`
	Days60Cents          int64     `json:"days_60_cents" db:"days_60_cents"`
	Days90PlusCents      int64     `json:"days_90_plus_cents" db:"days_90_plus_cents"`
	TotalReceivableCents int64     `json:"total_receivable_cents" db:"total_receivable_cents"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
}
