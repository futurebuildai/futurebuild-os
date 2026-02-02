package models

import (
	"time"

	"github.com/google/uuid"
)

// BusinessConfig holds per-organization physics tuning parameters.
// See STEP_87_CONFIG_PERSISTENCE.md Section 1
type BusinessConfig struct {
	ID              uuid.UUID `json:"id" db:"id"`
	OrgID           uuid.UUID `json:"org_id" db:"org_id"`
	SpeedMultiplier float64   `json:"speed_multiplier" db:"speed_multiplier"`
	WorkDays        []int     `json:"work_days" db:"work_days"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// DefaultBusinessConfig returns the default physics configuration.
func DefaultBusinessConfig(orgID uuid.UUID) *BusinessConfig {
	return &BusinessConfig{
		OrgID:           orgID,
		SpeedMultiplier: 1.0,
		WorkDays:        []int{1, 2, 3, 4, 5},
	}
}
