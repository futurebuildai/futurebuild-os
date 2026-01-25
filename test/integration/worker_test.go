package integration

import (
	"context"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/worker"
	"github.com/colton/futurebuild/test/testhelpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestWorker_ProjectHydration verifies the async "Nerve System".
// Validates: API (Service) -> Redis (Asynq) -> Worker -> DB (Procurement Items)
// See PRODUCTION_PLAN.md Step 62.3
//
// L7 Quality: Full end-to-end integration test with idempotency verification.
func TestWorker_ProjectHydration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 1. Setup Stack (Redis + Postgres + Asynq + Agents)
	stack := testhelpers.NewIntegrationStack(t)

	// 2. Ensure clean state
	ctx := context.Background()
	require.NoError(t, stack.TruncateAll(ctx))

	// 3. Create Organization (Required for Foreign Key)
	orgID := uuid.New()
	_, err := stack.DB.Exec(ctx, `INSERT INTO organizations (id, name, slug) VALUES ($1, $2, $3)`, orgID, "Test Org", "test-org")
	require.NoError(t, err)

	// 4. Seed WBS Data (System Catalog)
	tmplID := uuid.New()
	phaseID := uuid.New()
	wbsTaskID := uuid.New()
	_, err = stack.DB.Exec(ctx, `INSERT INTO wbs_templates (id, name, version, is_default, created_at) VALUES ($1, 'Test Template', 'v1', true, NOW())`, tmplID)
	require.NoError(t, err)

	_, err = stack.DB.Exec(ctx, `INSERT INTO wbs_phases (id, template_id, code, name, sort_order) VALUES ($1, $2, '6.0', 'Procurement', 1)`, phaseID, tmplID)
	require.NoError(t, err)

	_, err = stack.DB.Exec(ctx, `INSERT INTO wbs_tasks (id, phase_id, code, name, is_long_lead, lead_time_weeks_min, lead_time_weeks_max, created_at) 
		VALUES ($1, $2, '6.1', 'Windows', true, 8, 20, NOW())`, wbsTaskID, phaseID)
	require.NoError(t, err)

	// 5. Create Project
	projectID := uuid.New()
	project := &models.Project{
		ID:               projectID,
		OrgID:            orgID,
		Name:             "Async Nerve Test Project",
		Address:          "123 Test Lane",
		Status:           "Preconstruction",
		PermitIssuedDate: nil,
	}
	err = stack.ProjectService.CreateProject(ctx, project)
	require.NoError(t, err)

	// 6. Seed Project Tasks (required for hydration to find tasks)
	taskID := uuid.New()
	_, err = stack.DB.Exec(ctx, `
		INSERT INTO project_tasks (id, project_id, wbs_code, name, status)
		VALUES ($1, $2, '6.1', 'Windows', 'Pending');
	`, taskID, projectID)
	require.NoError(t, err)

	// 7. Re-Enqueue Hydration (Async Full Loop)
	// Ensures hydration runs AFTER tasks exist
	hTask, err := worker.NewHydrateProjectTask(projectID)
	require.NoError(t, err)
	_, err = stack.AsynqClient.Enqueue(hTask)
	require.NoError(t, err)

	// 8. Await Side Effects (Procurement Item Creation)
	require.Eventually(t, func() bool {
		var count int
		err := stack.DB.QueryRow(ctx, `
			SELECT COUNT(*) 
			FROM procurement_items pi
			JOIN project_tasks pt ON pi.project_task_id = pt.id
			WHERE pt.project_id = $1
		`, projectID).Scan(&count)
		if err != nil {
			t.Logf("Query error: %v", err)
			return false
		}

		if count > 0 {
			return true
		}
		return false
	}, 10*time.Second, 200*time.Millisecond, "Procurement items should be created via async worker")

	// 9. Capture count for idempotency check
	var countBefore int
	err = stack.DB.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM procurement_items pi
		JOIN project_tasks pt ON pi.project_task_id = pt.id
		WHERE pt.project_id = $1
	`, projectID).Scan(&countBefore)
	require.NoError(t, err)
	require.Greater(t, countBefore, 0, "Should have at least one procurement item")

	// 10. Idempotency Check: Re-run hydration
	err = stack.ProcurementAgent.HydrateProject(ctx, projectID)
	require.NoError(t, err)

	// 11. Verify count is stable (idempotent operation)
	var countAfter int
	err = stack.DB.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM procurement_items pi
		JOIN project_tasks pt ON pi.project_task_id = pt.id
		WHERE pt.project_id = $1
	`, projectID).Scan(&countAfter)
	require.NoError(t, err)

	// L7 Quality: Assert idempotency - count should not change
	require.Equal(t, countBefore, countAfter, "HydrateProject should be idempotent")
}
