package models

import (
	"time"

	"github.com/google/uuid"
)

// ThreadType distinguishes portal threads from internal threads.
type ThreadType string

const (
	ThreadTypeGeneral ThreadType = "general"
	ThreadTypePortal  ThreadType = "portal"
	ThreadTypeTopic   ThreadType = "topic"
)

// Thread represents a conversation thread within a project.
// Projects can have multiple concurrent threads. One thread per project
// is marked as the "General" thread (is_general = true).
type Thread struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	ProjectID  uuid.UUID  `json:"project_id" db:"project_id" validate:"required"`
	Title      string     `json:"title" db:"title" validate:"required"`
	IsGeneral  bool       `json:"is_general" db:"is_general"`
	ThreadType ThreadType `json:"thread_type" db:"thread_type"`
	ArchivedAt *time.Time `json:"archived_at,omitempty" db:"archived_at"`
	CreatedBy  *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}
