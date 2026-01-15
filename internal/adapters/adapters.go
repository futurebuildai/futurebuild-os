// Package adapters provides adapter types for interface satisfaction.
// These adapters bridge service layer implementations to the interfaces
// required by agents and handlers.
// See PRODUCTION_PLAN.md Technical Debt Remediation (P2) Section C.
package adapters

import (
	"context"

	"github.com/colton/futurebuild/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- ScheduleServiceAdapter ---

// ScheduleServiceAdapter adapts ScheduleService to InboundProgressUpdater interface.
// See PRODUCTION_PLAN.md Step 48 (Separation of Concerns)
type ScheduleServiceAdapter struct {
	ss *service.ScheduleService
	db *pgxpool.Pool
}

// NewScheduleServiceAdapter creates a new adapter wrapping a ScheduleService.
func NewScheduleServiceAdapter(ss *service.ScheduleService, db *pgxpool.Pool) *ScheduleServiceAdapter {
	return &ScheduleServiceAdapter{ss: ss, db: db}
}

// UpdateTaskProgress implements agents.InboundProgressUpdater.
func (a *ScheduleServiceAdapter) UpdateTaskProgress(ctx context.Context, taskID uuid.UUID, percent int) error {
	// Fetch projectID from task for proper service call
	var projectID uuid.UUID
	err := a.db.QueryRow(ctx, `SELECT project_id FROM project_tasks WHERE id = $1`, taskID).Scan(&projectID)
	if err != nil {
		return err
	}
	// Use nil userID for automated updates
	return a.ss.CreateTaskProgress(ctx, projectID, taskID, uuid.Nil, percent, "Updated via inbound webhook")
}

// RecalculateSchedule implements agents.InboundProgressUpdater.
func (a *ScheduleServiceAdapter) RecalculateSchedule(ctx context.Context, projectID, orgID uuid.UUID) error {
	_, err := a.ss.RecalculateSchedule(ctx, projectID, orgID)
	return err
}

// --- VisionServiceAdapter ---

// VisionServiceAdapter adapts VisionService to InboundVisionVerifier interface.
// See PRODUCTION_PLAN.md Step 48 (Separation of Concerns)
type VisionServiceAdapter struct {
	vs *service.VisionService
}

// NewVisionServiceAdapter creates a new adapter wrapping a VisionService.
func NewVisionServiceAdapter(vs *service.VisionService) *VisionServiceAdapter {
	return &VisionServiceAdapter{vs: vs}
}

// VerifyTask implements agents.InboundVisionVerifier.
func (a *VisionServiceAdapter) VerifyTask(ctx context.Context, imageURL, taskDescription string) (bool, float64, error) {
	return a.vs.VerifyTask(ctx, imageURL, taskDescription)
}
