// Package agents provides AI agent implementations for FutureBuild.
// This file defines the ProjectRepository interface for testability.
// P1 Scalability Fix: DailyFocusAgent O(N) Memory Elimination
package agents

import (
	"context"

	"github.com/colton/futurebuild/internal/models"
)

// ProjectProcessor is a callback function for processing projects one-by-one.
// Uses streaming to prevent OOM at scale.
type ProjectProcessor func(p models.Project) error

// ProjectRepository abstracts project streaming for agents.
// Enables unit testing of DailyFocusAgent without database.
//
// FAANG Standard: Depend on abstractions, not concretions.
// This allows DailyFocusAgent logic to be tested with mocks.
type ProjectRepository interface {
	// StreamActiveProjects iterates through active projects via callback.
	// O(1) memory - only one project loaded at a time.
	// P1 Scalability Fix: Prevents OOM at 100K+ projects.
	StreamActiveProjects(ctx context.Context, process ProjectProcessor) error
}
