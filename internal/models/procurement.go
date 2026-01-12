package models

import (
	"time"

	"github.com/google/uuid"
)

// ProcurementItem represents a long-lead item that must be ordered.
// See BACKEND_SCOPE.md Section 4.2
type ProcurementItem struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	ProjectID            uuid.UUID  `json:"project_id" db:"project_id" validate:"required"`
	WBSTaskID            uuid.UUID  `json:"wbs_task_id" db:"wbs_task_id" validate:"required"`
	ItemName             string     `json:"item_name" db:"item_name" validate:"required"`
	VendorID             *uuid.UUID `json:"vendor_id,omitempty" db:"vendor_id"`
	LeadTimeWeeks        int        `json:"lead_time_weeks" db:"lead_time_weeks"`
	NeedByDate           *time.Time `json:"need_by_date,omitempty" db:"need_by_date"`
	CalculatedOrderDate  *time.Time `json:"calculated_order_date,omitempty" db:"calculated_order_date"`
	ActualOrderDate      *time.Time `json:"actual_order_date,omitempty" db:"actual_order_date"`
	ExpectedDeliveryDate *time.Time `json:"expected_delivery_date,omitempty" db:"expected_delivery_date"`
	ActualDeliveryDate   *time.Time `json:"actual_delivery_date,omitempty" db:"actual_delivery_date"`
	Status               string     `json:"status" db:"status"` // Enum: not_ordered, ordered, in_transit, delivered, installed
	PONumber             string     `json:"po_number,omitempty" db:"po_number"`
	Notes                string     `json:"notes,omitempty" db:"notes"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
}
