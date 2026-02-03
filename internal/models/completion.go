package models

import (
	"time"

	"github.com/google/uuid"
)

// CompletionReport represents the final report generated when a project is marked complete.
// See Project Completion Flow Implementation Plan.
type CompletionReport struct {
	ID                   uuid.UUID              `json:"id" db:"id"`
	ProjectID            uuid.UUID              `json:"project_id" db:"project_id"`
	GeneratedBy          *uuid.UUID             `json:"generated_by,omitempty" db:"generated_by"`
	ScheduleSummary      ScheduleSummary        `json:"schedule_summary"`
	BudgetSummary        BudgetSummary          `json:"budget_summary"`
	WeatherImpactSummary *WeatherImpactSummary  `json:"weather_impact_summary,omitempty"`
	ProcurementSummary   *ProcurementSummary    `json:"procurement_summary,omitempty"`
	Notes                string                 `json:"notes,omitempty" db:"notes"`
	CreatedAt            time.Time              `json:"created_at" db:"created_at"`
}

// ScheduleSummary aggregates schedule metrics at project completion.
type ScheduleSummary struct {
	TotalTasks        int     `json:"total_tasks"`
	CompletedTasks    int     `json:"completed_tasks"`
	OnTimePercent     float64 `json:"on_time_percent"`
	TotalDurationDays int     `json:"total_duration_days"`
	ActualDurationDays int    `json:"actual_duration_days"`
}

// BudgetSummary aggregates financial metrics at project completion.
// All monetary values in cents (int64) to prevent IEEE 754 float drift.
type BudgetSummary struct {
	EstimatedCents int64 `json:"estimated_cents"`
	CommittedCents int64 `json:"committed_cents"`
	ActualCents    int64 `json:"actual_cents"`
	VarianceCents  int64 `json:"variance_cents"` // actual - estimated
}

// WeatherImpactSummary aggregates weather-related delays at project completion.
type WeatherImpactSummary struct {
	TotalDelayDays int `json:"total_delay_days"`
	PhasesAffected int `json:"phases_affected"`
}

// ProcurementSummary aggregates procurement metrics at project completion.
type ProcurementSummary struct {
	TotalItems      int   `json:"total_items"`
	TotalSpendCents int64 `json:"total_spend_cents"`
	VendorCount     int   `json:"vendor_count"`
}
