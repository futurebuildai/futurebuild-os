package models

import (
	"time"

	"github.com/google/uuid"
)

// See DATA_SPINE_SPEC.md Section 3.1
type ProjectStatus string

const (
	ProjectStatusPreconstruction ProjectStatus = "Preconstruction"
	ProjectStatusActive          ProjectStatus = "Active"
	ProjectStatusPaused          ProjectStatus = "Paused"
	ProjectStatusCompleted       ProjectStatus = "Completed"
)

// Project represents the high-level project container.
// See DATA_SPINE_SPEC.md Section 3.1
type Project struct {
	ID               uuid.UUID     `json:"id" db:"id"`
	OrgID            uuid.UUID     `json:"org_id" db:"org_id"`
	Name             string        `json:"name" db:"name"`
	Address          string        `json:"address" db:"address"`
	PermitIssuedDate *time.Time    `json:"permit_issued_date,omitempty" db:"permit_issued_date"`
	TargetEndDate    *time.Time    `json:"target_end_date,omitempty" db:"target_end_date"`
	GSF              float64       `json:"square_footage" db:"gsf"`
	Status           ProjectStatus `json:"status" db:"status"`
	Bedrooms         int           `json:"bedrooms" db:"bedrooms"`
	Bathrooms        int           `json:"bathrooms" db:"bathrooms"`
	Stories          int           `json:"stories" db:"stories"`
	LotSize          float64       `json:"lot_size" db:"lot_size"`
	FoundationType   string        `json:"foundation_type" db:"foundation_type"`
	Topography       string        `json:"topography" db:"topography"`
	SoilConditions   string        `json:"soil_conditions" db:"soil_conditions"`
}


// ProjectContext represents calibrated physics variables for a project.
// See DATA_SPINE_SPEC.md Section 3.2
type ProjectContext struct {
	ID                      uuid.UUID `json:"id" db:"id"`
	ProjectID               uuid.UUID `json:"project_id" db:"project_id"`
	SupplyChainVolatility   int       `json:"supply_chain_volatility" db:"supply_chain_volatility"`
	RoughInspectionLatency  int       `json:"rough_inspection_latency" db:"rough_inspection_latency"`
	FinalInspectionLatency  int       `json:"final_inspection_latency" db:"final_inspection_latency"`
	ZipCode                 string    `json:"zip_code" db:"zip_code"`
	ClimateZone             string    `json:"climate_zone" db:"climate_zone"`
}

