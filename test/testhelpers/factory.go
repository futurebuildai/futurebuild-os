package testhelpers

import (
	"context"
	"testing"

	"github.com/colton/futurebuild/internal/api/handlers"
	"github.com/colton/futurebuild/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IntegrationStack holds the wired application components
type IntegrationStack struct {
	DB             *pgxpool.Pool
	Router         *chi.Mux
	ProjectService *service.ProjectService
}

// NewIntegrationStack spins up the container, migrates DB, and wires the app.
// It registers cleanup on the testing.T
func NewIntegrationStack(t *testing.T) *IntegrationStack {
	// 1. Start Container (Handles Migrations & Pool creation)
	dbPool, cleanup := StartPostgresContainer(t)
	t.Cleanup(cleanup)

	// 2. Wire Services and Handlers
	projectService := service.NewProjectService(dbPool)
	projectHandler := handlers.NewProjectHandler(projectService)

	r := chi.NewRouter()
	r.Post("/projects", projectHandler.CreateProject)
	r.Get("/projects/{id}", projectHandler.GetProject)

	return &IntegrationStack{
		DB:             dbPool,
		Router:         r,
		ProjectService: projectService,
	}
}

// TruncateAll cleans the database between tests
// Uses dynamic discovery to only truncate tables that exist
func (s *IntegrationStack) TruncateAll(ctx context.Context) error {
	// Get list of user tables (exclude system tables)
	query := `
		SELECT tablename FROM pg_tables 
		WHERE schemaname = 'public' 
		  AND tablename NOT LIKE 'pg_%'
		  AND tablename != 'schema_migrations'
	`
	rows, err := s.DB.Query(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		tables = append(tables, name)
	}

	if len(tables) == 0 {
		return nil
	}

	// Truncate all discovered tables with CASCADE
	truncateSQL := "TRUNCATE TABLE " + tables[0]
	for _, t := range tables[1:] {
		truncateSQL += ", " + t
	}
	truncateSQL += " CASCADE"

	_, err = s.DB.Exec(ctx, truncateSQL)
	return err
}
