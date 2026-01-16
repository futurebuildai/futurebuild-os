// Package agents provides AI agent implementations for FutureBuild.
// This file provides a mock implementation of ProjectRepository for testing.
// P1 Scalability Fix: DailyFocusAgent O(N) Memory Elimination
package agents

import (
	"context"

	"github.com/colton/futurebuild/internal/models"
)

// MockProjectRepository implements ProjectRepository for unit testing.
// Enables testing of DailyFocusAgent business logic without a database.
type MockProjectRepository struct {
	// Projects to return from StreamActiveProjects
	Projects []models.Project

	// Error injection for testing error handling
	StreamErr error

	// Track which projects were processed (for assertions)
	ProcessedProjects []models.Project

	// FailAtIndex causes StreamActiveProjects to fail after processing N projects
	// Set to -1 to disable (default)
	FailAtIndex int
	FailError   error
}

// NewMockProjectRepository creates a mock repository with sensible defaults.
func NewMockProjectRepository() *MockProjectRepository {
	return &MockProjectRepository{
		Projects:          []models.Project{},
		ProcessedProjects: []models.Project{},
		FailAtIndex:       -1,
	}
}

// StreamActiveProjects iterates through mock projects via callback.
func (m *MockProjectRepository) StreamActiveProjects(ctx context.Context, process ProjectProcessor) error {
	if m.StreamErr != nil {
		return m.StreamErr
	}

	for i, p := range m.Projects {
		// Check for simulated failure at specific index
		if m.FailAtIndex >= 0 && i >= m.FailAtIndex {
			return m.FailError
		}

		if err := process(p); err != nil {
			return err
		}
		m.ProcessedProjects = append(m.ProcessedProjects, p)
	}
	return nil
}

// --- Test Helpers ---

// WithProjects sets the projects to stream.
func (m *MockProjectRepository) WithProjects(projects ...models.Project) *MockProjectRepository {
	m.Projects = projects
	return m
}

// WithStreamError sets an error to return immediately from StreamActiveProjects.
func (m *MockProjectRepository) WithStreamError(err error) *MockProjectRepository {
	m.StreamErr = err
	return m
}

// WithFailureAt configures the mock to fail after processing N projects.
// Useful for testing error handling mid-stream.
func (m *MockProjectRepository) WithFailureAt(index int, err error) *MockProjectRepository {
	m.FailAtIndex = index
	m.FailError = err
	return m
}
