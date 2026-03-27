package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// A2AExecutionLog tracks cross-system agent execution for Why-Trail visualization.
// See FRONTEND_SCOPE.md Section 15.1
type A2AExecutionLog struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	OrgID        uuid.UUID       `json:"org_id" db:"org_id"`
	WorkflowID   *uuid.UUID      `json:"workflow_id,omitempty" db:"workflow_id"`
	SourceSystem string          `json:"source_system" db:"source_system"`
	TargetSystem string          `json:"target_system" db:"target_system"`
	ActionType   string          `json:"action_type" db:"action_type"`
	Payload      json.RawMessage `json:"payload,omitempty" db:"payload"`
	Status       string          `json:"status" db:"status"`
	ErrorMessage *string         `json:"error_message,omitempty" db:"error_message"`
	DurationMs   *int            `json:"duration_ms,omitempty" db:"duration_ms"`
	ExecutedAt   time.Time       `json:"executed_at" db:"executed_at"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
}

// ActiveAgentConnection tracks agent health and pause state for the OS-to-Brain bridge.
// See FRONTEND_SCOPE.md Section 15.1
type ActiveAgentConnection struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	OrgID           uuid.UUID  `json:"org_id" db:"org_id"`
	AgentName       string     `json:"agent_name" db:"agent_name"`
	AgentType       string     `json:"agent_type" db:"agent_type"`
	BrainWorkflowID *uuid.UUID `json:"brain_workflow_id,omitempty" db:"brain_workflow_id"`
	Status          string     `json:"status" db:"status"`
	LastExecutionAt *time.Time `json:"last_execution_at,omitempty" db:"last_execution_at"`
	ExecutionCount  int        `json:"execution_count" db:"execution_count"`
	ErrorCount      int        `json:"error_count" db:"error_count"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}
