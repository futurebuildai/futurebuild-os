package models

import (
	"time"

	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

// See DATA_SPINE_SPEC.md Section 5.1
// See DATA_SPINE_SPEC.md Section 5.1
// CommunicationDirection is imported from pkg/types

// CommunicationChannel is imported from pkg/types

// CommunicationLog represents the history of interaction.
// See DATA_SPINE_SPEC.md Section 5.1 (Revised with Logical fix for User logging)
type CommunicationLog struct {
	ID        uuid.UUID                    `json:"id" db:"id"`
	ProjectID uuid.UUID                    `json:"project_id" db:"project_id" validate:"required"`
	ContactID *uuid.UUID                   `json:"contact_id,omitempty" db:"contact_id"`
	Direction types.CommunicationDirection `json:"direction" db:"direction" validate:"required"`
	Content   string                       `json:"content" db:"content" validate:"required"`
	Channel   types.CommunicationChannel   `json:"channel" db:"channel" validate:"required"`
	Timestamp time.Time                    `json:"timestamp" db:"timestamp"`
}

// See DATA_SPINE_SPEC.md Section 5.2
// See DATA_SPINE_SPEC.md Section 5.2
// NotificationType is imported from pkg/types

// NotificationStatus is imported from pkg/types

// Notification represents a system alert.
// See DATA_SPINE_SPEC.md Section 5.2 (Strict Parity: No timestamps)
type Notification struct {
	ID       uuid.UUID                `json:"id" db:"id"`
	UserID   uuid.UUID                `json:"user_id" db:"user_id" validate:"required"`
	Type     types.NotificationType   `json:"type" db:"type" validate:"required"`
	Priority int                      `json:"priority" db:"priority"`
	Status   types.NotificationStatus `json:"status" db:"status"`
}

// ChatMessage represents a single message in an agent session.
// See DATA_SPINE_SPEC.md Section 5.3
// Portal messages use ContactID (UserID is nil); internal messages use UserID (ContactID is nil).
type ChatMessage struct {
	ID        uuid.UUID        `json:"id" db:"id"`
	ProjectID uuid.UUID        `json:"project_id" db:"project_id" validate:"required"`
	ThreadID  uuid.UUID        `json:"thread_id" db:"thread_id" validate:"required"`
	UserID    *uuid.UUID       `json:"user_id,omitempty" db:"user_id"`
	ContactID *uuid.UUID       `json:"contact_id,omitempty" db:"contact_id"`
	Role      types.ChatRole   `json:"role" db:"role" validate:"required"`
	Content   string           `json:"content" db:"content" validate:"required"`
	ToolCalls []types.ToolCall `json:"tool_calls,omitempty" db:"tool_calls"` // CTO-002: Typed struct
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
}
