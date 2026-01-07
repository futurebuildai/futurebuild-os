package models

import (
	"time"

	"github.com/google/uuid"
)

// Organization represents the multi-tenant container.
// See DATA_SPINE_SPEC.md Section 2.1
type Organization struct {
	ID           uuid.UUID      `json:"id" db:"id"`
	Name         string         `json:"name" db:"name" validate:"required"`
	Slug         string         `json:"slug" db:"slug" validate:"required"`
	Settings     map[string]any `json:"settings" db:"settings"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	ProjectLimit int            `json:"project_limit" db:"project_limit"`
}

// See DATA_SPINE_SPEC.md Section 2.2 & API_AND_TYPES_SPEC.md Section 1.2
type UserRole string

const (
	UserRoleAdmin         UserRole = "Admin"
	UserRoleBuilder       UserRole = "Builder"
	UserRoleClient        UserRole = "Client"
	UserRoleSubcontractor UserRole = "Subcontractor"
)

// User represents internal users managed via Magic Link auth.
// See DATA_SPINE_SPEC.md Section 2.2
type User struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	OrgID     uuid.UUID  `json:"org_id" db:"org_id" validate:"required"`
	Email     string     `json:"email" db:"email" validate:"required,email"`
	Name      string     `json:"name" db:"name" validate:"required"`
	Role      UserRole   `json:"role" db:"role" validate:"required"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// See DATA_SPINE_SPEC.md Section 2.3
type ContactRole string

const (
	ContactRoleClient        ContactRole = "Client"
	ContactRoleSubcontractor ContactRole = "Subcontractor"
)

type ContactPreference string

const (
	ContactPreferenceSMS   ContactPreference = "SMS"
	ContactPreferenceEmail ContactPreference = "Email"
	ContactPreferenceBoth  ContactPreference = "Both"
)

// Contact represents the Global Address Book entry.
// See DATA_SPINE_SPEC.md Section 2.3
type Contact struct {
	ID                uuid.UUID         `json:"id" db:"id"`
	OrgID             uuid.UUID         `json:"org_id" db:"org_id" validate:"required"`
	Name              string            `json:"name" db:"name" validate:"required"`
	Company           *string           `json:"company,omitempty" db:"company"`
	Phone             *string           `json:"phone,omitempty" db:"phone"`
	Email             *string           `json:"email,omitempty" db:"email"`
	GlobalRole        ContactRole       `json:"global_role" db:"global_role" validate:"required"`
	ContactPreference ContactPreference `json:"contact_preference" db:"contact_preference"`
}
