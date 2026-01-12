package models

import "github.com/google/uuid"

// See BACKEND_SCOPE.md Section 4.2
type MultiplierSource string

const (
	MultiplierSourceDefault      MultiplierSource = "default"
	MultiplierSourceOrgTrained   MultiplierSource = "org_trained"
	MultiplierSourceGlobalTrained MultiplierSource = "global_trained"
)

// DurationMultiplier represents a weighted variable for the DHSM calculator.
// See BACKEND_SCOPE.md Section 4.2
type DurationMultiplier struct {
	ID                uuid.UUID         `json:"id" db:"id"`
	OrgID             *uuid.UUID        `json:"org_id,omitempty" db:"org_id"` // Nullable for global defaults
	WBSTaskCode       string            `json:"wbs_task_code" db:"wbs_task_code"` // "*" for global, or specific code
	VariableKey       string            `json:"variable_key" db:"variable_key"` // e.g., "square_footage"
	Weight            float64           `json:"weight" db:"weight"` // e.g., 0.15
	MultiplierFormula string            `json:"multiplier_formula" db:"multiplier_formula"` // e.g., "linear"
	MinValue          float64           `json:"min_value" db:"min_value"`
	MaxValue          float64           `json:"max_value" db:"max_value"`
	Source            MultiplierSource  `json:"source" db:"source"`
	Confidence        float64           `json:"confidence" db:"confidence"`
}
