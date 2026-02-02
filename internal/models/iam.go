package models

import (
	"time"

	"github.com/google/uuid"
)

// Identity represents a polymorphic authenticated entity (User or Contact).
type Identity interface {
	GetID() uuid.UUID
	GetOrgID() uuid.UUID
	GetEmail() string
	GetName() string
	GetRole() UserRole
	GetCreatedAt() time.Time
	IsInternal() bool
}

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

func (u *User) GetID() uuid.UUID    { return u.ID }
func (u *User) GetOrgID() uuid.UUID { return u.OrgID }
func (u *User) GetEmail() string    { return u.Email }
func (u *User) GetName() string     { return u.Name }
func (u *User) GetRole() UserRole   { return u.Role }
func (u *User) GetCreatedAt() time.Time { return u.CreatedAt }
func (u *User) IsInternal() bool    { return true }


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
	Role              UserRole          `json:"role" db:"role" validate:"required"`
	ContactPreference ContactPreference `json:"contact_preference" db:"contact_preference"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`
}

func (c *Contact) GetID() uuid.UUID    { return c.ID }
func (c *Contact) GetOrgID() uuid.UUID { return c.OrgID }
func (c *Contact) GetEmail() string    {
	if c.Email == nil {
		return ""
	}
	return *c.Email
}
func (c *Contact) GetName() string   { return c.Name }
func (c *Contact) GetRole() UserRole { return c.Role }
func (c *Contact) GetCreatedAt() time.Time { return c.CreatedAt }
func (c *Contact) IsInternal() bool  { return false }
// PortalToken represents a stateful magic link token for portal contacts.
// Main app auth uses Clerk (Phase 12). Only portal contacts use magic links.
type PortalToken struct {
	TokenHash string    `db:"token_hash"`
	ContactID uuid.UUID `db:"contact_id"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
	Used      bool      `db:"used"`
}
