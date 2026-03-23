package models

import (
	"time"

	"github.com/google/uuid"
)

// MaterialSource tracks the origin of a material entry.
type MaterialSource string

const (
	MaterialSourceAI      MaterialSource = "ai"
	MaterialSourceUser    MaterialSource = "user"
	MaterialSourceDefault MaterialSource = "default"
)

// ProjectMaterial represents a material line item in a project's bill of materials.
// See migrations/000077_create_project_materials.up.sql
// MONETARY PRECISION: All costs stored as int64 cents to prevent IEEE 754 float drift.
type ProjectMaterial struct {
	ID             uuid.UUID      `json:"id" db:"id"`
	ProjectID      uuid.UUID      `json:"project_id" db:"project_id"`
	WBSPhaseCode   string         `json:"wbs_phase_code" db:"wbs_phase_code"`
	Name           string         `json:"name" db:"name"`
	Category       string         `json:"category" db:"category"`
	Quantity       float64        `json:"quantity" db:"quantity"`
	Unit           string         `json:"unit" db:"unit"`
	UnitCostCents  int64          `json:"unit_cost_cents" db:"unit_cost_cents"`
	TotalCostCents int64          `json:"total_cost_cents" db:"total_cost_cents"`
	Source         MaterialSource `json:"source" db:"source"`
	Confidence     float64        `json:"confidence" db:"confidence"`
	Brand          string         `json:"brand,omitempty" db:"brand"`
	Model          string         `json:"model,omitempty" db:"model"`
	SKU            string         `json:"sku,omitempty" db:"sku"`
	Notes          string         `json:"notes,omitempty" db:"notes"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
}

// MaterialEstimate is a lightweight representation used in onboarding responses
// and budget seeding. Not persisted directly — converted to ProjectMaterial on save.
type MaterialEstimate struct {
	Name           string  `json:"name"`
	Category       string  `json:"category"`
	WBSPhaseCode   string  `json:"wbs_phase_code"`
	Quantity       float64 `json:"quantity"`
	Unit           string  `json:"unit"`
	UnitCostCents  int64   `json:"unit_cost_cents"`
	TotalCostCents int64   `json:"total_cost_cents"`
	Confidence     float64 `json:"confidence"`
	Source         string  `json:"source"` // "ai" or "default"
}

// BudgetEstimate is the aggregate budget projection returned during onboarding.
// Groups material costs by WBS phase for the budget review step.
type BudgetEstimate struct {
	TotalEstimatedCents int64                 `json:"total_estimated_cents"`
	PhaseBreakdown      []PhaseBudgetEstimate `json:"phase_breakdown"`
	RegionalMultiplier  float64               `json:"regional_multiplier"`
	ConfidenceOverall   float64               `json:"confidence_overall"`
}

// PhaseBudgetEstimate is a single phase's cost breakdown within a BudgetEstimate.
type PhaseBudgetEstimate struct {
	WBSPhaseCode  string  `json:"wbs_phase_code"`
	PhaseName     string  `json:"phase_name"`
	EstimatedCents int64  `json:"estimated_cents"`
	MaterialsCents int64  `json:"materials_cents"`
	LaborCents     int64  `json:"labor_cents"`
	Confidence     float64 `json:"confidence"`
}

// MaterialUpdateRequest holds the fields a user can modify on a ProjectMaterial.
type MaterialUpdateRequest struct {
	Name           *string  `json:"name,omitempty"`
	Category       *string  `json:"category,omitempty"`
	WBSPhaseCode   *string  `json:"wbs_phase_code,omitempty"`
	Quantity       *float64 `json:"quantity,omitempty"`
	Unit           *string  `json:"unit,omitempty"`
	UnitCostCents  *int64   `json:"unit_cost_cents,omitempty"`
	Brand          *string  `json:"brand,omitempty"`
	Model          *string  `json:"model,omitempty"`
	SKU            *string  `json:"sku,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
}

// CreateMaterialRequest is the request body for manually adding a material.
type CreateMaterialRequest struct {
	WBSPhaseCode  string  `json:"wbs_phase_code"`
	Name          string  `json:"name"`
	Category      string  `json:"category"`
	Quantity      float64 `json:"quantity"`
	Unit          string  `json:"unit"`
	UnitCostCents int64   `json:"unit_cost_cents"`
	Brand         string  `json:"brand,omitempty"`
	Model         string  `json:"model,omitempty"`
	SKU           string  `json:"sku,omitempty"`
	Notes         string  `json:"notes,omitempty"`
}

// BudgetSeedRequest is the request body for seeding a project budget from materials.
type BudgetSeedRequest struct {
	Materials          []MaterialEstimate `json:"materials"`
	RegionalMultiplier float64            `json:"regional_multiplier"`
}
