package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/physics"
	"github.com/colton/futurebuild/pkg/types"
)

// TestRecalculateSchedule_CriticalPathChange verifies that changing a critical path task's
// duration shifts the project end date.
// See PRODUCTION_PLAN.md Step 32.6
func TestRecalculateSchedule_CriticalPathChange(t *testing.T) {
	projectID := uuid.New()
	// Feb 2, 2026 is a Monday
	projectStart := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)

	// Create a linear chain: A (2 days) → B (3 days) → C (1 day)
	// All tasks are on the critical path
	taskA := models.ProjectTask{
		ID:                 uuid.New(),
		ProjectID:          projectID,
		WBSCode:            "7.1",
		Name:               "Site Prep",
		CalculatedDuration: 2.0,
	}
	taskB := models.ProjectTask{
		ID:                 uuid.New(),
		ProjectID:          projectID,
		WBSCode:            "7.2",
		Name:               "Clearing",
		CalculatedDuration: 3.0,
	}
	taskC := models.ProjectTask{
		ID:                 uuid.New(),
		ProjectID:          projectID,
		WBSCode:            "7.3",
		Name:               "Layout",
		CalculatedDuration: 1.0,
	}

	tasks := []models.ProjectTask{taskA, taskB, taskC}
	deps := []models.TaskDependency{
		{ID: uuid.New(), ProjectID: projectID, PredecessorID: taskA.ID, SuccessorID: taskB.ID, DependencyType: types.DependencyTypeFS},
		{ID: uuid.New(), ProjectID: projectID, PredecessorID: taskB.ID, SuccessorID: taskC.ID, DependencyType: types.DependencyTypeFS},
	}

	// Initial CPM: Project duration = 2 + 3 + 1 = 6 working days
	cal := &physics.StandardCalendar{}
	graph := physics.BuildDependencyGraph(tasks, deps)
	schedule, err := physics.ForwardPass(graph, projectStart, cal)
	require.NoError(t, err)

	criticalPath, err := physics.BackwardPass(graph, schedule, cal)
	require.NoError(t, err)

	// Verify all tasks are critical
	assert.Len(t, criticalPath, 3, "All tasks should be on critical path")

	// Initial project end: projectStart + 6 working days
	initialEnd := cal.AddWorkingDays(projectStart, 6)
	taskCSchedule := schedule[taskC.ID]
	assert.Equal(t, initialEnd, taskCSchedule.EarlyFinish, "Initial project end should be 6 working days from start")

	// Now simulate overriding Task B's duration from 3 to 7 days
	override := 7.0
	taskB.ManualOverrideDays = &override

	// Rebuild graph with override
	tasks[1] = taskB
	graph = physics.BuildDependencyGraph(tasks, deps)
	schedule, err = physics.ForwardPass(graph, projectStart, cal)
	require.NoError(t, err)

	_, err = physics.BackwardPass(graph, schedule, cal)
	require.NoError(t, err)

	// New project end: 2 + 7 + 1 = 10 working days
	newEnd := cal.AddWorkingDays(projectStart, 10)
	taskCSchedule = schedule[taskC.ID]
	assert.Equal(t, newEnd, taskCSchedule.EarlyFinish, "Project end should shift to 10 working days")

	// Verify the shift is 4 working days (7-3)
	expectedShift := cal.AddWorkingDays(initialEnd, 4)
	assert.Equal(t, expectedShift, taskCSchedule.EarlyFinish, "Project end should shift by 4 working days (7-3)")
}

// TestRecalculateSchedule_NonCriticalChange verifies that changing a non-critical task's
// duration consumes float but does not shift the project end date (if within float).
// See PRODUCTION_PLAN.md Step 32.6
func TestRecalculateSchedule_NonCriticalChange(t *testing.T) {
	projectID := uuid.New()
	// Feb 2, 2026 is a Monday
	projectStart := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)

	// Create a diamond pattern:
	// A (1 day) → B (5 days, critical)
	// A (1 day) → C (2 days, non-critical with float)
	// B → D (1 day)
	// C → D (1 day)
	taskA := models.ProjectTask{
		ID:                 uuid.New(),
		ProjectID:          projectID,
		WBSCode:            "8.1",
		Name:               "Excavation",
		CalculatedDuration: 1.0,
	}
	taskB := models.ProjectTask{
		ID:                 uuid.New(),
		ProjectID:          projectID,
		WBSCode:            "8.2",
		Name:               "Foundation (Critical)",
		CalculatedDuration: 5.0,
	}
	taskC := models.ProjectTask{
		ID:                 uuid.New(),
		ProjectID:          projectID,
		WBSCode:            "8.3",
		Name:               "Utilities (Non-Critical)",
		CalculatedDuration: 2.0,
	}
	taskD := models.ProjectTask{
		ID:                 uuid.New(),
		ProjectID:          projectID,
		WBSCode:            "8.4",
		Name:               "Backfill",
		CalculatedDuration: 1.0,
	}

	tasks := []models.ProjectTask{taskA, taskB, taskC, taskD}
	deps := []models.TaskDependency{
		{ID: uuid.New(), ProjectID: projectID, PredecessorID: taskA.ID, SuccessorID: taskB.ID, DependencyType: types.DependencyTypeFS},
		{ID: uuid.New(), ProjectID: projectID, PredecessorID: taskA.ID, SuccessorID: taskC.ID, DependencyType: types.DependencyTypeFS},
		{ID: uuid.New(), ProjectID: projectID, PredecessorID: taskB.ID, SuccessorID: taskD.ID, DependencyType: types.DependencyTypeFS},
		{ID: uuid.New(), ProjectID: projectID, PredecessorID: taskC.ID, SuccessorID: taskD.ID, DependencyType: types.DependencyTypeFS},
	}

	// Initial CPM
	cal := &physics.StandardCalendar{}
	graph := physics.BuildDependencyGraph(tasks, deps)
	schedule, err := physics.ForwardPass(graph, projectStart, cal)
	require.NoError(t, err)

	_, err = physics.BackwardPass(graph, schedule, cal)
	require.NoError(t, err)

	// Initial project end: 1 + 5 + 1 = 7 working days (critical path A→B→D)
	initialEnd := cal.AddWorkingDays(projectStart, 7)
	taskDSchedule := schedule[taskD.ID]
	assert.Equal(t, initialEnd, taskDSchedule.EarlyFinish, "Initial project end should be 7 working days from start")

	// Task C has positive float
	taskCSchedule := schedule[taskC.ID]
	assert.Greater(t, taskCSchedule.TotalFloat, 0.0, "Task C should have positive float")
	assert.False(t, taskCSchedule.IsCritical, "Task C should NOT be critical")

	// Override Task C duration from 2 to 4 days (still within float)
	override := 4.0
	taskC.ManualOverrideDays = &override

	// Rebuild graph with override
	tasks[2] = taskC
	graph = physics.BuildDependencyGraph(tasks, deps)
	schedule, err = physics.ForwardPass(graph, projectStart, cal)
	require.NoError(t, err)

	_, err = physics.BackwardPass(graph, schedule, cal)
	require.NoError(t, err)

	// Project end should NOT change since C is still non-critical
	taskDSchedule = schedule[taskD.ID]
	assert.Equal(t, initialEnd, taskDSchedule.EarlyFinish, "Project end should NOT shift")

	// Task C should still have some float
	taskCSchedule = schedule[taskC.ID]
	assert.Greater(t, taskCSchedule.TotalFloat, 0.0, "Task C should still have positive float")
}

// TestRecalculateSchedule_NonCriticalExceedsFloat verifies that if a non-critical task
// exceeds its float, it becomes critical and shifts the project end date.
func TestRecalculateSchedule_NonCriticalExceedsFloat(t *testing.T) {
	projectID := uuid.New()
	// Feb 2, 2026 is a Monday
	projectStart := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)

	// Same diamond pattern as above
	taskA := models.ProjectTask{
		ID:                 uuid.New(),
		ProjectID:          projectID,
		WBSCode:            "8.1",
		Name:               "Excavation",
		CalculatedDuration: 1.0,
	}
	taskB := models.ProjectTask{
		ID:                 uuid.New(),
		ProjectID:          projectID,
		WBSCode:            "8.2",
		Name:               "Foundation (Critical)",
		CalculatedDuration: 5.0,
	}
	taskC := models.ProjectTask{
		ID:                 uuid.New(),
		ProjectID:          projectID,
		WBSCode:            "8.3",
		Name:               "Utilities (Non-Critical)",
		CalculatedDuration: 2.0, // Has float
	}
	taskD := models.ProjectTask{
		ID:                 uuid.New(),
		ProjectID:          projectID,
		WBSCode:            "8.4",
		Name:               "Backfill",
		CalculatedDuration: 1.0,
	}

	// Override C to exceed float: 2 + 5 = 7 days (exceeds B's 5 days by 2)
	override := 7.0
	taskC.ManualOverrideDays = &override

	tasks := []models.ProjectTask{taskA, taskB, taskC, taskD}
	deps := []models.TaskDependency{
		{ID: uuid.New(), ProjectID: projectID, PredecessorID: taskA.ID, SuccessorID: taskB.ID, DependencyType: types.DependencyTypeFS},
		{ID: uuid.New(), ProjectID: projectID, PredecessorID: taskA.ID, SuccessorID: taskC.ID, DependencyType: types.DependencyTypeFS},
		{ID: uuid.New(), ProjectID: projectID, PredecessorID: taskB.ID, SuccessorID: taskD.ID, DependencyType: types.DependencyTypeFS},
		{ID: uuid.New(), ProjectID: projectID, PredecessorID: taskC.ID, SuccessorID: taskD.ID, DependencyType: types.DependencyTypeFS},
	}

	cal := &physics.StandardCalendar{}
	graph := physics.BuildDependencyGraph(tasks, deps)
	schedule, err := physics.ForwardPass(graph, projectStart, cal)
	require.NoError(t, err)

	criticalPath, err := physics.BackwardPass(graph, schedule, cal)
	require.NoError(t, err)

	// Task C is now critical (longer than B)
	taskCSchedule := schedule[taskC.ID]
	assert.True(t, taskCSchedule.IsCritical, "Task C should now be critical")
	assert.InDelta(t, 0.0, taskCSchedule.TotalFloat, 0.001, "Task C should have zero float")

	// C is now on critical path
	assert.Contains(t, criticalPath, "8.3", "Task C should be on critical path")

	// New project end: 1 + 7 + 1 = 9 working days
	expectedEnd := cal.AddWorkingDays(projectStart, 9)
	taskDSchedule := schedule[taskD.ID]
	assert.Equal(t, expectedEnd, taskDSchedule.EarlyFinish, "Project end should be 9 working days from start")

	// Task B now has float
	taskBSchedule := schedule[taskB.ID]
	assert.False(t, taskBSchedule.IsCritical, "Task B should now be NON-critical")
	assert.Greater(t, taskBSchedule.TotalFloat, 0.0, "Task B should have positive float")
}

// TestRecalculateSchedule_EmptyProject verifies handling of projects with no tasks.
func TestRecalculateSchedule_EmptyProject(t *testing.T) {
	// Feb 2, 2026 is a Monday
	projectStart := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)

	graph := physics.BuildDependencyGraph(nil, nil)
	schedule, err := physics.ForwardPass(graph, projectStart, &physics.StandardCalendar{})
	require.NoError(t, err)
	assert.Len(t, schedule, 0, "Empty project should have no scheduled tasks")

	criticalPath, err := physics.BackwardPass(graph, schedule, &physics.StandardCalendar{})
	require.NoError(t, err)
	assert.Nil(t, criticalPath, "Empty project should have no critical path")
}
