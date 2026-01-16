// Package agents provides AI agent implementations for FutureBuild.
// This file implements ProjectRepository using PostgreSQL via ProjectService.
// P1 Scalability Fix: DailyFocusAgent O(N) Memory Elimination
package agents

import (
	"context"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
)

// PgProjectRepository implements ProjectRepository by delegating to ProjectService.
// This is the production implementation that wraps the service layer.
type PgProjectRepository struct {
	svc *service.ProjectService
}

// NewPgProjectRepository creates a new PostgreSQL-backed project repository.
func NewPgProjectRepository(svc *service.ProjectService) *PgProjectRepository {
	return &PgProjectRepository{svc: svc}
}

// StreamActiveProjects iterates through active projects via callback.
// O(1) memory - delegates to ProjectService.StreamActiveProjects.
// P1 Scalability Fix: Prevents OOM at 100K+ projects.
func (r *PgProjectRepository) StreamActiveProjects(ctx context.Context, process ProjectProcessor) error {
	// Adapt the callback type from agents.ProjectProcessor to service.ProjectProcessor
	return r.svc.StreamActiveProjects(ctx, func(p models.Project) error {
		return process(p)
	})
}
