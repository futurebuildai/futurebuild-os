package models

import (
	"time"

	"github.com/google/uuid"
)

// WBSTemplate represents a master project template (CPM-res1.0).
// See BACKEND_SCOPE.md Section 4.2
type WBSTemplate struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Name          string    `json:"name" db:"name" validate:"required"`
	Version       string    `json:"version" db:"version"`
	IsDefault     bool      `json:"is_default" db:"is_default"`
	EntryPointWBS string    `json:"entry_point_wbs" db:"entry_point_wbs"` // Default "5.2"
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// WBSPhase represents a phase within a WBS template.
// See BACKEND_SCOPE.md Section 4.2
type WBSPhase struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	TemplateID         uuid.UUID `json:"template_id" db:"template_id" validate:"required"`
	Code               string    `json:"code" db:"code" validate:"required"`
	Name               string    `json:"name" db:"name" validate:"required"`
	IsWeatherSensitive bool      `json:"is_weather_sensitive" db:"is_weather_sensitive"`
	SortOrder          int       `json:"sort_order" db:"sort_order"`
}

// WBSTask represents a master task definition in the WBS library.
// See BACKEND_SCOPE.md Section 4.2
type WBSTask struct {
	ID                uuid.UUID `json:"id" db:"id"`
	PhaseID           uuid.UUID `json:"phase_id" db:"phase_id" validate:"required"`
	Code              string    `json:"code" db:"code" validate:"required"`
	Name              string    `json:"name" db:"name" validate:"required"`
	BaseDurationDays  float64   `json:"base_duration_days" db:"base_duration_days"`
	ResponsibleParty  string    `json:"responsible_party" db:"responsible_party"`
	Deliverable       string    `json:"deliverable" db:"deliverable"`
	Notes             string    `json:"notes" db:"notes"`
	IsInspection      bool      `json:"is_inspection" db:"is_inspection"`
	IsMilestone       bool      `json:"is_milestone" db:"is_milestone"`
	IsLongLead        bool      `json:"is_long_lead" db:"is_long_lead"`
	LeadTimeWeeksMin  int       `json:"lead_time_weeks_min" db:"lead_time_weeks_min"`
	LeadTimeWeeksMax  int       `json:"lead_time_weeks_max" db:"lead_time_weeks_max"`
	PredecessorCodes  []string  `json:"predecessor_codes,omitempty" db:"predecessor_codes"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}
