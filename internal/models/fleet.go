package models

import (
	"time"

	"github.com/google/uuid"
)

// AssetStatus defines the lifecycle state of a fleet asset.
type AssetStatus string

const (
	AssetStatusAvailable   AssetStatus = "available"
	AssetStatusInUse       AssetStatus = "in_use"
	AssetStatusMaintenance AssetStatus = "maintenance"
	AssetStatusRetired     AssetStatus = "retired"
)

// AllocationStatus defines the state of an equipment allocation.
type AllocationStatus string

const (
	AllocationStatusPlanned   AllocationStatus = "planned"
	AllocationStatusActive    AllocationStatus = "active"
	AllocationStatusCompleted AllocationStatus = "completed"
	AllocationStatusCancelled AllocationStatus = "cancelled"
)

// FleetAsset represents heavy equipment, vehicles, or tools.
// See BACKEND_SCOPE.md Section 20.3
// MONETARY PRECISION: Cost/value stored as int64 cents.
type FleetAsset struct {
	ID                uuid.UUID   `json:"id" db:"id"`
	OrgID             uuid.UUID   `json:"org_id" db:"org_id"`
	AssetNumber       string      `json:"asset_number" db:"asset_number"`
	AssetType         string      `json:"asset_type" db:"asset_type"`
	Make              *string     `json:"make,omitempty" db:"make"`
	Model             *string     `json:"model,omitempty" db:"model"`
	Year              *int        `json:"year,omitempty" db:"year"`
	VIN               *string     `json:"vin,omitempty" db:"vin"`
	LicensePlate      *string     `json:"license_plate,omitempty" db:"license_plate"`
	PurchaseDate      *time.Time  `json:"purchase_date,omitempty" db:"purchase_date"`
	PurchaseCostCents *int64      `json:"purchase_cost_cents,omitempty" db:"purchase_cost_cents"`
	CurrentValueCents *int64      `json:"current_value_cents,omitempty" db:"current_value_cents"`
	Status            AssetStatus `json:"status" db:"status"`
	Location          *string     `json:"location,omitempty" db:"location"`
	Notes             *string     `json:"notes,omitempty" db:"notes"`
	VisibleToRoles    []string    `json:"visible_to_roles,omitempty" db:"visible_to_roles"` // Phase 20: per-asset role visibility
	CreatedAt         time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at" db:"updated_at"`
}

// EquipmentAllocation represents resource assignment to a project/task.
// Resource-constrained: one asset can only be at one project at a time.
// See BACKEND_SCOPE.md Section 20.3
type EquipmentAllocation struct {
	ID            uuid.UUID        `json:"id" db:"id"`
	AssetID       uuid.UUID        `json:"asset_id" db:"asset_id"`
	ProjectID     uuid.UUID        `json:"project_id" db:"project_id"`
	TaskID        *uuid.UUID       `json:"task_id,omitempty" db:"task_id"`
	AllocatedFrom time.Time        `json:"allocated_from" db:"allocated_from"`
	AllocatedTo   time.Time        `json:"allocated_to" db:"allocated_to"`
	Status        AllocationStatus `json:"status" db:"status"`
	Notes         *string          `json:"notes,omitempty" db:"notes"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at" db:"updated_at"`
}

// MaintenanceLog tracks EAM service activities for fleet assets.
// See BACKEND_SCOPE.md Section 20.3
type MaintenanceLog struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	AssetID         uuid.UUID  `json:"asset_id" db:"asset_id"`
	MaintenanceType string     `json:"maintenance_type" db:"maintenance_type"`
	Description     string     `json:"description" db:"description"`
	ScheduledDate   *time.Time `json:"scheduled_date,omitempty" db:"scheduled_date"`
	CompletedDate   *time.Time `json:"completed_date,omitempty" db:"completed_date"`
	CostCents       *int64     `json:"cost_cents,omitempty" db:"cost_cents"`
	VendorName      *string    `json:"vendor_name,omitempty" db:"vendor_name"`
	Notes           *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}
