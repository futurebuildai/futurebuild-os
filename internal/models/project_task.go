package models

import (
	"time"

	"github.com/google/uuid"
)

// See DATA_SPINE_SPEC.md Section 3.3
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "Pending"
	TaskStatusReady      TaskStatus = "Ready"
	TaskStatusInProgress TaskStatus = "In_Progress"
	TaskStatusCompleted  TaskStatus = "Completed"
	TaskStatusBlocked    TaskStatus = "Blocked"
	TaskStatusDelayed    TaskStatus = "Delayed"
)

// ProjectTask represents a specific instance of a task for a live project.
// See DATA_SPINE_SPEC.md Section 3.3
type ProjectTask struct {
	ID                       uuid.UUID  `json:"id" db:"id"`
	ProjectID                uuid.UUID  `json:"project_id" db:"project_id" validate:"required"`
	WBSCode                  string     `json:"wbs_code" db:"wbs_code" validate:"required"`
	Name                     string     `json:"name" db:"name" validate:"required"`
	EarlyStart               *time.Time `json:"early_start,omitempty" db:"early_start"`
	EarlyFinish              *time.Time `json:"early_finish,omitempty" db:"early_finish"`
	CalculatedDuration       float64    `json:"calculated_duration" db:"calculated_duration"`
	WeatherAdjustedDuration  float64    `json:"weather_adjusted_duration" db:"weather_adjusted_duration"`
	ManualOverrideDays       *float64   `json:"manual_override_days,omitempty" db:"manual_override_days"`
	Status                   TaskStatus `json:"status" db:"status"`
	VerifiedByVision         bool       `json:"verified_by_vision" db:"verified_by_vision"`
	VerificationConfidence  float64    `json:"verification_confidence" db:"verification_confidence"`
}

// See DATA_SPINE_SPEC.md Section 3.4
type DependencyType string

const (
	DependencyTypeFS DependencyType = "FS"
	DependencyTypeSS DependencyType = "SS"
	DependencyTypeFF DependencyType = "FF"
)

// TaskDependency represents an edge in the Project Task DAG.
// See DATA_SPINE_SPEC.md Section 3.4
type TaskDependency struct {
	ID             uuid.UUID      `json:"id" db:"id"`
	ProjectID      uuid.UUID      `json:"project_id" db:"project_id" validate:"required"`
	PredecessorID  uuid.UUID      `json:"predecessor_id" db:"predecessor_id" validate:"required"`
	SuccessorID    uuid.UUID      `json:"successor_id" db:"successor_id" validate:"required"`
	DependencyType DependencyType `json:"dependency_type" db:"dependency_type"`
	LagDays        int            `json:"lag_days" db:"lag_days"`
}

// ProjectAssignment links contacts to project phases.
// See DATA_SPINE_SPEC.md Section 3.5
type ProjectAssignment struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ProjectID  uuid.UUID `json:"project_id" db:"project_id" validate:"required"`
	ContactID  uuid.UUID `json:"contact_id" db:"contact_id" validate:"required"`
	WBSPhaseID string    `json:"wbs_phase_id" db:"wbs_phase_id" validate:"required"`
}
