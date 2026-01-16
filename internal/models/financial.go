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
	InvoiceStatusPending  InvoiceStatus = "Pending"
	InvoiceStatusApproved InvoiceStatus = "Approved"
	InvoiceStatusExported InvoiceStatus = "Exported"
)

// ProjectBudget represents the financial tracking for a specific WBS Phase.
// See DATA_SPINE_SPEC.md Section 4.1
// MONETARY PRECISION: All amounts stored as int64 cents to prevent IEEE 754 float drift.
type ProjectBudget struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	ProjectID            uuid.UUID `json:"project_id" db:"project_id"`
	WBSPhaseID           string    `json:"wbs_phase_id" db:"wbs_phase_id"`
	EstimatedAmountCents int64     `json:"estimated_amount_cents" db:"estimated_amount_cents"`
	CommittedAmountCents int64     `json:"committed_amount_cents" db:"committed_amount_cents"`
	ActualAmountCents    int64     `json:"actual_amount_cents" db:"actual_amount_cents"`
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
	Confidence            float64       `json:"confidence" db:"confidence"` // Confidence remains float (0.0-1.0)
	IsHumanReviewRequired bool          `json:"is_human_review_required" db:"is_human_review_required"`
	SourceDocumentID      *uuid.UUID    `json:"source_document_id,omitempty" db:"source_document_id"` // See PRODUCTION_PLAN.md Step 41
}
