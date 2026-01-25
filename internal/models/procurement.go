package models

import (
	"time"

	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

// ProcurementItem represents a long-lead item that must be monitored.
// This model aligns with the physical schema in 000045_create_procurement_items.
// See BACKEND_SCOPE.md Section 2.5
type ProcurementItem struct {
	ID                   uuid.UUID                    `json:"id" db:"id"`
	ProjectTaskID        uuid.UUID                    `json:"project_task_id" db:"project_task_id" validate:"required"`
	ProjectID            uuid.UUID                    `json:"project_id" db:"-"` // Derived via JOIN on project_tasks, not stored
	WBSCode              string                       `json:"wbs_code" db:"-"`   // Derived via JOIN on project_tasks, not stored
	ItemName             string                       `json:"item_name" db:"name" validate:"required"`
	LeadTimeWeeks        int                          `json:"lead_time_weeks" db:"lead_time_weeks"`
	Status               types.ProcurementAlertStatus `json:"status" db:"status"`
	CalculatedOrderDate  *time.Time                   `json:"calculated_order_date,omitempty" db:"calculated_order_date"`
	ExpectedDeliveryDate *time.Time                   `json:"expected_delivery_date,omitempty" db:"expected_delivery_date"`
	CreatedAt            time.Time                    `json:"created_at" db:"created_at"`
}
