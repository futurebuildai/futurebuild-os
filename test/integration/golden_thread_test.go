package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/colton/futurebuild/test/testhelpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestGoldenThread_LifeOfAProject verifies the "Golden Thread" E2E flow.
// Path: API (Create) -> DB (Persist) -> Redis (Enqueue) -> Worker (Process) -> DB (Hydrate) -> API (Observe)
// See PRODUCTION_PLAN.md Step 62.4
//
// L7 Quality: Fully decoupled integration test using the real HTTP handler stack.
func TestGoldenThread_LifeOfAProject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 1. Setup Stack
	stack := testhelpers.NewIntegrationStack(t)
	ctx := context.Background()
	require.NoError(t, stack.TruncateAll(ctx))

	// 2. Setup Identities
	orgID := uuid.New()
	userID := uuid.New()
	_, err := stack.DB.Exec(ctx, `INSERT INTO organizations (id, name, slug) VALUES ($1, $2, $3)`, orgID, "Golden Thread Co", "golden-thread")
	require.NoError(t, err)

	claims := &types.Claims{
		OrgID:  orgID.String(),
		UserID: userID.String(),
		Role:   types.UserRoleBuilder,
	}

	// 3. Seed WBS Data (Systems Catalog)
	// This simulates a pre-existing construction knowledge base.
	// Define constants to prevent test/hydration logic drift.
	const (
		testWBSTaskCode   = "6.1"
		testWBSTaskName   = "Custom Windows"
		testLeadTimeWeeks = 12
	)

	tmplID := uuid.New()
	phaseID := uuid.New()
	wbsTaskID := uuid.New()
	_, err = stack.DB.Exec(ctx, `INSERT INTO wbs_templates (id, name, version, is_default) VALUES ($1, 'Master Build', 'v1', true)`, tmplID)
	require.NoError(t, err)
	_, err = stack.DB.Exec(ctx, `INSERT INTO wbs_phases (id, template_id, code, name, sort_order) VALUES ($1, $2, '6.0', 'Procurement', 1)`, phaseID, tmplID)
	require.NoError(t, err)
	_, err = stack.DB.Exec(ctx, `INSERT INTO wbs_tasks (id, phase_id, code, name, is_long_lead, lead_time_weeks_min) VALUES ($1, $2, $3, $4, true, $5)`, wbsTaskID, phaseID, testWBSTaskCode, testWBSTaskName, testLeadTimeWeeks)
	require.NoError(t, err)

	// 4. API CALL: Create Project
	// Trigger: POST /projects
	projectPayload := models.Project{
		Name:    "E2E Golden Thread Mansion",
		Address: "1 Wealth Road, FutureCity",
		GSF:     5500.0,
	}
	body, _ := json.Marshal(projectPayload)
	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer(body))
	req = req.WithContext(middleware.WithClaims(req.Context(), claims))

	w := httptest.NewRecorder()
	stack.Router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code, "API should return 201 Created")

	var createdProject models.Project
	require.NoError(t, json.NewDecoder(w.Body).Decode(&createdProject))
	require.NotEqual(t, uuid.Nil, createdProject.ID)
	require.Equal(t, projectPayload.Name, createdProject.Name)

	// 5. SIMULATION: Ingest/Hydrate Project Tasks
	// ⚠️ KNOWN LIMITATION (P1-1): In production, a "Blueprint Analyzer" or WBS ingestion
	// flow would create project_tasks when a project is created. This test manually seeds
	// the task to focus on verifying the Procurement Hydration flow specifically.
	// TODO: When blueprint ingestion is implemented, update this test to verify the full flow.
	taskID := uuid.New()
	_, err = stack.DB.Exec(ctx, `
		INSERT INTO project_tasks (id, project_id, wbs_code, name, status)
		VALUES ($1, $2, $3, $4, 'Pending');
	`, taskID, createdProject.ID, testWBSTaskCode, testWBSTaskName)
	require.NoError(t, err)

	// 6. OBSERVE: Verification of Enqueued Task
	// We don't need to manually enqueue; CreateProject already did it.
	// We just need to wait for the worker to process it.
	// Note: The worker runs in the background as part of the IntegrationStack.

	// 7. AWAIT & API CALL: Observe Side Effects
	// Trigger: GET /projects/{id}/procurement
	var observedItems []models.ProcurementItem
	require.Eventually(t, func() bool {
		obsReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/projects/%s/procurement", createdProject.ID), nil)
		obsReq = obsReq.WithContext(middleware.WithClaims(obsReq.Context(), claims))

		obsW := httptest.NewRecorder()
		stack.Router.ServeHTTP(obsW, obsReq)

		if obsW.Code != http.StatusOK {
			t.Logf("GET procurement items failed with status %d", obsW.Code)
			return false
		}

		if err := json.NewDecoder(obsW.Body).Decode(&observedItems); err != nil {
			return false
		}

		// Success Condition: The worker has created the procurement item
		return len(observedItems) > 0
	}, 15*time.Second, 500*time.Millisecond, "Worker should hydrate project and API should expose items")

	// Assertions moved outside Eventually to avoid panic on early iterations (P1-2 fix)
	require.Equal(t, testWBSTaskName, observedItems[0].ItemName)
	require.Equal(t, testLeadTimeWeeks, observedItems[0].LeadTimeWeeks)

	// 8. FINAL AUDIT: Verify persistent state correctness
	var count int
	err = stack.DB.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM procurement_items pi
		JOIN project_tasks pt ON pi.project_task_id = pt.id
		WHERE pt.project_id = $1
	`, createdProject.ID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "Should have exactly 1 procurement item")
}
