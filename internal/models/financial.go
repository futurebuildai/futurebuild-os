package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// InvoiceStatus defines the lifecycle of an Invoice.
// See API_AND_TYPES_SPEC.md Section 1.3
type InvoiceStatus string

const (
	InvoiceStatusDraft    InvoiceStatus = "Draft"
	InvoiceStatusPending  InvoiceStatus = "Pending"
	InvoiceStatusApproved InvoiceStatus = "Approved"
	InvoiceStatusRejected InvoiceStatus = "Rejected"
	InvoiceStatusExported InvoiceStatus = "Exported"
)

// IsEditable returns true if the invoice can be modified by a user.
// Only Draft invoices are editable. All other states are locked.
func (s InvoiceStatus) IsEditable() bool {
	return s == InvoiceStatusDraft
}

// ProjectBudget represents the financial tracking for a specific WBS Phase.
// See DATA_SPINE_SPEC.md Section 4.1
// MONETARY PRECISION: All amounts stored as int64 cents to prevent IEEE 754 float drift.
type ProjectBudget struct {
	ID                   uuid.UUID      `json:"id" db:"id"`
	ProjectID            uuid.UUID      `json:"project_id" db:"project_id"`
	WBSPhaseID           string         `json:"wbs_phase_id" db:"wbs_phase_id"`
	EstimatedAmountCents int64          `json:"estimated_amount_cents" db:"estimated_amount_cents"`
	CommittedAmountCents int64          `json:"committed_amount_cents" db:"committed_amount_cents"`
	ActualAmountCents    int64          `json:"actual_amount_cents" db:"actual_amount_cents"`
	Source               MaterialSource `json:"source,omitempty" db:"source"`
	Confidence           float64        `json:"confidence,omitempty" db:"confidence"`
	IsLocked             bool           `json:"is_locked" db:"is_locked"`
}

// LineItem represents a single entry in an invoice.
// See API_AND_TYPES_SPEC.md Section 3.1
// MONETARY PRECISION: UnitPrice and Total stored as int64 cents.
type LineItem struct {
	Description    string  `json:"description"`
	Quantity       float64 `json:"quantity"` // Kept as float - quantities can be fractional (e.g., 2.5 hours)
	UnitPriceCents int64   `json:"unit_price_cents"`
	TotalCents     int64   `json:"total_cents"`
}

// LineItems is a slice of LineItem that implements Scanner/Valuer for JSONB.
type LineItems []LineItem

// Value implements the driver.Valuer interface.
func (l LineItems) Value() (driver.Value, error) {
	return json.Marshal(l)
}

// Scan implements the sql.Scanner interface.
func (l *LineItems) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, l)
}

// Invoice represents a parsed financial artifact from the Action Engine.
// See DATA_SPINE_SPEC.md Section 4.2
// MONETARY PRECISION: Amount stored as int64 cents to prevent IEEE 754 float drift.
type Invoice struct {
	ID                    uuid.UUID     `json:"id" db:"id"`
	ProjectID             uuid.UUID     `json:"project_id" db:"project_id"`
	VendorName            string        `json:"vendor_name" db:"vendor_name"`
	AmountCents           int64         `json:"amount_cents" db:"amount_cents"`
	LineItems             LineItems     `json:"line_items" db:"line_items"`
	DetectedWBSCode       string        `json:"detected_wbs_code" db:"detected_wbs_code"`
	Status                InvoiceStatus `json:"status" db:"status"`
	InvoiceDate           *time.Time    `json:"invoice_date" db:"invoice_date"`
	InvoiceNumber         *string       `json:"invoice_number" db:"invoice_number"`
	Confidence            float64       `json:"confidence" db:"confidence"`
	IsHumanReviewRequired bool          `json:"is_human_review_required" db:"is_human_review_required"`
	SourceDocumentID      *uuid.UUID    `json:"source_document_id,omitempty" db:"source_document_id"`
	// Step 83: Approval/Rejection metadata
	ApprovedByID    *string    `json:"approved_by_id,omitempty" db:"approved_by_id"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty" db:"approved_at"`
	RejectedByID    *string    `json:"rejected_by_id,omitempty" db:"rejected_by_id"`
	RejectedAt      *time.Time `json:"rejected_at,omitempty" db:"rejected_at"`
	RejectionReason *string    `json:"rejection_reason,omitempty" db:"rejection_reason"`
}

// IsApprovable returns true if the invoice can be approved.
// Only Draft invoices can be approved.
func (s InvoiceStatus) IsApprovable() bool {
	return s == InvoiceStatusDraft
}

// IsRejectable returns true if the invoice can be rejected.
// Only Draft invoices can be rejected.
func (s InvoiceStatus) IsRejectable() bool {
	return s == InvoiceStatusDraft
}
