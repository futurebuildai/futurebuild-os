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
type ProjectBudget struct {
	ID              uuid.UUID `json:"id" db:"id"`
	ProjectID       uuid.UUID `json:"project_id" db:"project_id"`
	WBSPhaseID      string    `json:"wbs_phase_id" db:"wbs_phase_id"`
	EstimatedAmount float64   `json:"estimated_amount" db:"estimated_amount"`
	CommittedAmount float64   `json:"committed_amount" db:"committed_amount"`
	ActualAmount    float64   `json:"actual_amount" db:"actual_amount"`
}

// LineItem represents a single entry in an invoice.
// See API_AND_TYPES_SPEC.md Section 3.1
type LineItem struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
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
type Invoice struct {
	ID                    uuid.UUID     `json:"id" db:"id"`
	ProjectID             uuid.UUID     `json:"project_id" db:"project_id"`
	VendorName            string        `json:"vendor_name" db:"vendor_name"`
	Amount                float64       `json:"amount" db:"amount"`
	LineItems             LineItems     `json:"line_items" db:"line_items"`
	DetectedWBSCode       string        `json:"detected_wbs_code" db:"detected_wbs_code"`
	Status                InvoiceStatus `json:"status" db:"status"`
	InvoiceDate           *time.Time    `json:"invoice_date" db:"invoice_date"`
	InvoiceNumber         *string       `json:"invoice_number" db:"invoice_number"`
	Confidence            float64       `json:"confidence" db:"confidence"`
	IsHumanReviewRequired bool          `json:"is_human_review_required" db:"is_human_review_required"`
}
