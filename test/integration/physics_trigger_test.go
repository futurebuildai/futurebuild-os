package integration

import (
	"context"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/test/testhelpers"
	"github.com/google/uuid"
)

// TestPhysicsTrigger_TaskDurationUpdate verifies that updating a predecessor task's
// duration causes dependent (successor) task start dates to shift automatically.
// This is the "Atomic Scheduling" behavior defined by CPM recalculation.
//
// See PRODUCTION_PLAN.md Step 62.2.5 (L7 Verification)
func TestPhysicsTrigger_TaskDurationUpdate(t *testing.T) {
	// 1. Setup Integration Stack
	stack := testhelpers.NewIntegrationStack(t)
	ctx := context.Background()

	if err := stack.TruncateAll(ctx); err != nil {
		t.Fatalf("failed to truncate db: %v", err)
	}

	// 2. Create Test Data: Org + Project
	orgID := uuid.New()
	projectID := uuid.New()

	// Insert organization (FK constraint)
	_, err := stack.DB.Exec(ctx,
		"INSERT INTO organizations (id, name, slug) VALUES ($1, $2, $3)",
		orgID, "Physics Test Org", "physics-test-org")
	if err != nil {
		t.Fatalf("failed to insert org: %v", err)
	}

	// Insert project with permit date (CPM start anchor)
	permitDate := time.Now().Truncate(24 * time.Hour)
	_, err = stack.DB.Exec(ctx,
		`INSERT INTO projects (id, org_id, name, address, permit_issued_date, status)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		projectID, orgID, "Physics Test Project", "123 Test St", permitDate, models.ProjectStatusActive)
	if err != nil {
		t.Fatalf("failed to insert project: %v", err)
	}

	// 3. Create Task A (predecessor) - 5 day duration initially
	taskAID := uuid.New()
	_, err = stack.DB.Exec(ctx,
		`INSERT INTO project_tasks (id, project_id, wbs_code, name, calculated_duration, status)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		taskAID, projectID, "1.0", "Task A - Foundation", 5.0, "Pending")
	if err != nil {
		t.Fatalf("failed to insert Task A: %v", err)
	}

	// 4. Create Task B (successor) - depends on Task A
	taskBID := uuid.New()
	_, err = stack.DB.Exec(ctx,
		`INSERT INTO project_tasks (id, project_id, wbs_code, name, calculated_duration, status)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		taskBID, projectID, "2.0", "Task B - Framing", 3.0, "Pending")
	if err != nil {
		t.Fatalf("failed to insert Task B: %v", err)
	}

	// 5. Create Finish-to-Start dependency: A -> B
	depID := uuid.New()
	_, err = stack.DB.Exec(ctx,
		`INSERT INTO task_dependencies (id, project_id, predecessor_id, successor_id, dependency_type, lag_days)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		depID, projectID, taskAID, taskBID, "FS", 0)
	if err != nil {
		t.Fatalf("failed to insert dependency: %v", err)
	}

	// 6. Wire up ScheduleService and run initial calculation
	scheduleService := service.NewScheduleService(stack.DB)

	_, err = scheduleService.RecalculateSchedule(ctx, projectID, orgID)
	if err != nil {
		t.Fatalf("initial schedule calculation failed: %v", err)
	}

	// 7. Read Task B's initial early_start
	var initialTaskBStart time.Time
	err = stack.DB.QueryRow(ctx,
		"SELECT early_start FROM project_tasks WHERE id = $1", taskBID).Scan(&initialTaskBStart)
	if err != nil {
		t.Fatalf("failed to read Task B initial start: %v", err)
	}

	t.Logf("Initial Task B start: %v", initialTaskBStart)

	// 8. Update Task A duration from 5 days to 10 days (add 5 days)
	err = scheduleService.UpdateTaskDuration(ctx, taskAID, projectID, orgID, 10.0, "Extended duration for testing")
	if err != nil {
		t.Fatalf("failed to update Task A duration: %v", err)
	}

	// 9. Recalculate schedule (this is the "trigger" being tested)
	_, err = scheduleService.RecalculateSchedule(ctx, projectID, orgID)
	if err != nil {
		t.Fatalf("schedule recalculation after update failed: %v", err)
	}

	// 10. Read Task B's updated early_start
	var updatedTaskBStart time.Time
	err = stack.DB.QueryRow(ctx,
		"SELECT early_start FROM project_tasks WHERE id = $1", taskBID).Scan(&updatedTaskBStart)
	if err != nil {
		t.Fatalf("failed to read Task B updated start: %v", err)
	}

	t.Logf("Updated Task B start: %v", updatedTaskBStart)

	// 11. ASSERTION: Task B should have shifted forward
	// With 5 additional days on Task A, Task B's start should be ~5 working days later
	shift := updatedTaskBStart.Sub(initialTaskBStart)
	minExpectedShift := 4 * 24 * time.Hour // At least 4 days (accounting for weekends)
	maxExpectedShift := 8 * 24 * time.Hour // At most 8 days (5 working days + 2 weekend days buffer)

	if shift < minExpectedShift || shift > maxExpectedShift {
		t.Errorf("Task B start date shift out of expected range: got %v, want between %v and %v",
			shift, minExpectedShift, maxExpectedShift)
	}

	// Verify basic forward shift happened
	if !updatedTaskBStart.After(initialTaskBStart) {
		t.Errorf("Task B start should have shifted forward: initial=%v, updated=%v",
			initialTaskBStart, updatedTaskBStart)
	}

	t.Logf("✅ Physics trigger verified: Task B shifted by %v when Task A duration increased", shift)
}
