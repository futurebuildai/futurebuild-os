package models

import (
	"time"

	"github.com/google/uuid"
)

// EmployeeStatus defines the lifecycle state of an employee.
type EmployeeStatus string

const (
	EmployeeStatusActive     EmployeeStatus = "active"
	EmployeeStatusOnLeave    EmployeeStatus = "on_leave"
	EmployeeStatusTerminated EmployeeStatus = "terminated"
)

// PayType defines the compensation model.
type PayType string

const (
	PayTypeHourly PayType = "hourly"
	PayTypeSalary PayType = "salary"
)

// Employee represents an internal workforce member.
// See BACKEND_SCOPE.md Section 20.2
type Employee struct {
	ID              uuid.UUID      `json:"id" db:"id"`
	OrgID           uuid.UUID      `json:"org_id" db:"org_id"`
	ContactID       *uuid.UUID     `json:"contact_id,omitempty" db:"contact_id"`
	FirstName       string         `json:"first_name" db:"first_name"`
	LastName        string         `json:"last_name" db:"last_name"`
	EmployeeNumber  *string        `json:"employee_number,omitempty" db:"employee_number"`
	Email           *string        `json:"email,omitempty" db:"email"`
	Phone           *string        `json:"phone,omitempty" db:"phone"`
	HireDate        *time.Time     `json:"hire_date,omitempty" db:"hire_date"`
	TerminationDate *time.Time     `json:"termination_date,omitempty" db:"termination_date"`
	Status          EmployeeStatus `json:"status" db:"status"`
	PayRateCents    *int           `json:"pay_rate_cents,omitempty" db:"pay_rate_cents"`
	PayType         *PayType       `json:"pay_type,omitempty" db:"pay_type"`
	Classification  *string        `json:"classification,omitempty" db:"classification"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
}

// TimeLog tracks labor hours for deterministic project costing.
// See BACKEND_SCOPE.md Section 20.2
type TimeLog struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	EmployeeID    uuid.UUID  `json:"employee_id" db:"employee_id"`
	ProjectID     *uuid.UUID `json:"project_id,omitempty" db:"project_id"`
	TaskID        *uuid.UUID `json:"task_id,omitempty" db:"task_id"`
	LogDate       time.Time  `json:"log_date" db:"log_date"`
	HoursWorked   float64    `json:"hours_worked" db:"hours_worked"`
	OvertimeHours float64    `json:"overtime_hours" db:"overtime_hours"`
	Notes         *string    `json:"notes,omitempty" db:"notes"`
	Approved      bool       `json:"approved" db:"approved"`
	ApprovedBy    *uuid.UUID `json:"approved_by,omitempty" db:"approved_by"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty" db:"approved_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// CertStatus defines the state of a certification.
type CertStatus string

const (
	CertStatusValid        CertStatus = "valid"
	CertStatusExpiringSoon CertStatus = "expiring_soon"
	CertStatusExpired      CertStatus = "expired"
)

// Certification tracks employee credentials and compliance expiration.
// See BACKEND_SCOPE.md Section 20.2
type Certification struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	EmployeeID       uuid.UUID  `json:"employee_id" db:"employee_id"`
	CertType         string     `json:"cert_type" db:"cert_type"`
	CertNumber       *string    `json:"cert_number,omitempty" db:"cert_number"`
	IssueDate        *time.Time `json:"issue_date,omitempty" db:"issue_date"`
	ExpirationDate   time.Time  `json:"expiration_date" db:"expiration_date"`
	IssuingAuthority *string    `json:"issuing_authority,omitempty" db:"issuing_authority"`
	DocumentURL      *string    `json:"document_url,omitempty" db:"document_url"`
	Status           CertStatus `json:"status" db:"status"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// PrevailingWageRate defines government-mandated pay rates by region.
// See BACKEND_SCOPE.md Section 20.2
type PrevailingWageRate struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	OrgID              uuid.UUID `json:"org_id" db:"org_id"`
	Region             string    `json:"region" db:"region"`
	Classification     string    `json:"classification" db:"classification"`
	EffectiveDate      time.Time `json:"effective_date" db:"effective_date"`
	HourlyRateCents    int       `json:"hourly_rate_cents" db:"hourly_rate_cents"`
	FringeBenefitCents int       `json:"fringe_benefit_cents" db:"fringe_benefit_cents"`
	Source             *string   `json:"source,omitempty" db:"source"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}
