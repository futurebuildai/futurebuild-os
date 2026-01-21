//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/types"
)

func TestDirectory_GetContactForPhase(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	cfg := getTestConfig()
	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Skipf("Skipping test: database not reachable: %v", err)
	}

	directoryService := service.NewDirectoryService(db)

	// 1. Setup Data for Org A
	orgID := uuid.New()
	slugA := fmt.Sprintf("org-a-%s", uuid.New().String())
	_, err = db.Exec(ctx, "INSERT INTO organizations (id, name, slug) VALUES ($1, 'Org A', $2)", orgID, slugA)
	require.NoError(t, err)

	projectID := uuid.New()
	_, err = db.Exec(ctx, "INSERT INTO projects (id, org_id, name, status) VALUES ($1, $2, 'Project A', 'Active')", projectID, orgID)
	require.NoError(t, err)

	phaseCode := "9.x"

	contactID := uuid.New()
	_, err = db.Exec(ctx, "INSERT INTO contacts (id, org_id, name, company, role, phone, email, contact_preference) VALUES ($1, $2, 'John Plumber', 'JP Plumbing', 'Subcontractor', '555-0123', 'john@jpplumbing.com', 'SMS')", contactID, orgID)
	require.NoError(t, err)

	_, err = db.Exec(ctx, "INSERT INTO project_assignments (project_id, contact_id, wbs_phase_id) VALUES ($1, $2, $3)", projectID, contactID, phaseCode)
	require.NoError(t, err)

	// 2. Setup Data for Org B (Isolation Check)
	orgBID := uuid.New()
	slugB := fmt.Sprintf("org-b-%s", uuid.New().String())
	_, err = db.Exec(ctx, "INSERT INTO organizations (id, name, slug) VALUES ($1, 'Org B', $2)", orgBID, slugB)
	require.NoError(t, err)

	// Test Case 1: Success
	t.Run("Success_ValidPhase", func(t *testing.T) {
		contact, err := directoryService.GetContactForPhase(ctx, projectID, orgID, phaseCode)
		require.NoError(t, err)
		assert.Equal(t, contactID, contact.ID)
		assert.Equal(t, "John Plumber", contact.Name)
		assert.Equal(t, "JP Plumbing", contact.Company)
		assert.Equal(t, "Subcontractor", string(contact.Role))
		assert.Equal(t, types.ContactPreferenceSMS, contact.ContactPreference)
	})

	// Test Case 2: Multi-Tenancy Failure (Wrong OrgID)
	t.Run("Failure_WrongOrgID", func(t *testing.T) {
		contact, err := directoryService.GetContactForPhase(ctx, projectID, orgBID, phaseCode)
		assert.Error(t, err)
		assert.Nil(t, contact)
		assert.Contains(t, err.Error(), "not found")
	})

	// Test Case 3: Wrong Phase Code
	t.Run("Failure_InvalidPhaseCode", func(t *testing.T) {
		contact, err := directoryService.GetContactForPhase(ctx, projectID, orgID, "99.x")
		assert.Error(t, err)
		assert.Nil(t, contact)
	})

	// Test Case 4: Assignment Deleted
	t.Run("Failure_AssignmentRemoved", func(t *testing.T) {
		_, err := db.Exec(ctx, "DELETE FROM project_assignments WHERE project_id = $1 AND wbs_phase_id = $2", projectID, phaseCode)
		require.NoError(t, err)

		contact, err := directoryService.GetContactForPhase(ctx, projectID, orgID, phaseCode)
		assert.Error(t, err)
		assert.Nil(t, contact)
	})
}
