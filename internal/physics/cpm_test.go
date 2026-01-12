package physics

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
)

// Helper to create a ProjectTask with minimal fields
func makeTask(wbsCode string) models.ProjectTask {
	return models.ProjectTask{
		ID:      uuid.New(),
		WBSCode: wbsCode,
		Name:    "Task " + wbsCode,
	}
}

// Helper to create a TaskDependency
func makeDep(projectID, predID, succID uuid.UUID, depType types.DependencyType, lag int) models.TaskDependency {
	return models.TaskDependency{
		ID:             uuid.New(),
		ProjectID:      projectID,
		PredecessorID:  predID,
		SuccessorID:    succID,
		DependencyType: depType,
		LagDays:        lag,
	}
}

// TestTopologicalSort_LinearDAG verifies processing order for A→B→C chain.
func TestTopologicalSort_LinearDAG(t *testing.T) {
	projectID := uuid.New()

	// Create tasks: A → B → C
	taskA := makeTask("1.1")
	taskB := makeTask("1.2")
	taskC := makeTask("1.3")

	tasks := []models.ProjectTask{taskA, taskB, taskC}
	deps := []models.TaskDependency{
		makeDep(projectID, taskA.ID, taskB.ID, types.DependencyTypeFS, 0),
		makeDep(projectID, taskB.ID, taskC.ID, types.DependencyTypeFS, 0),
	}

	g := BuildDependencyGraph(tasks, deps)
	sorted, err := TopologicalSort(g)

	require.NoError(t, err)
	require.Len(t, sorted, 3)

	// A must come before B, B must come before C
	indexA := indexOf(sorted, taskA.ID)
	indexB := indexOf(sorted, taskB.ID)
	indexC := indexOf(sorted, taskC.ID)

	assert.Less(t, indexA, indexB, "A should come before B")
	assert.Less(t, indexB, indexC, "B should come before C")
}

// TestTopologicalSort_BranchingDAG verifies parallel paths: A→B, A→C, B→D, C→D.
func TestTopologicalSort_BranchingDAG(t *testing.T) {
	projectID := uuid.New()

	taskA := makeTask("2.1")
	taskB := makeTask("2.2")
	taskC := makeTask("2.3")
	taskD := makeTask("2.4")

	tasks := []models.ProjectTask{taskA, taskB, taskC, taskD}
	deps := []models.TaskDependency{
		makeDep(projectID, taskA.ID, taskB.ID, types.DependencyTypeFS, 0),
		makeDep(projectID, taskA.ID, taskC.ID, types.DependencyTypeFS, 0),
		makeDep(projectID, taskB.ID, taskD.ID, types.DependencyTypeFS, 0),
		makeDep(projectID, taskC.ID, taskD.ID, types.DependencyTypeFS, 0),
	}

	g := BuildDependencyGraph(tasks, deps)
	sorted, err := TopologicalSort(g)

	require.NoError(t, err)
	require.Len(t, sorted, 4)

	indexA := indexOf(sorted, taskA.ID)
	indexB := indexOf(sorted, taskB.ID)
	indexC := indexOf(sorted, taskC.ID)
	indexD := indexOf(sorted, taskD.ID)

	assert.Less(t, indexA, indexB, "A should come before B")
	assert.Less(t, indexA, indexC, "A should come before C")
	assert.Less(t, indexB, indexD, "B should come before D")
	assert.Less(t, indexC, indexD, "C should come before D")
}

// TestDetectCycle_RejectsCircular verifies cycle detection per Step 34.
func TestDetectCycle_RejectsCircular(t *testing.T) {
	projectID := uuid.New()

	// Create circular dependency: A → B → C → A
	taskA := makeTask("3.1")
	taskB := makeTask("3.2")
	taskC := makeTask("3.3")

	tasks := []models.ProjectTask{taskA, taskB, taskC}
	deps := []models.TaskDependency{
		makeDep(projectID, taskA.ID, taskB.ID, types.DependencyTypeFS, 0),
		makeDep(projectID, taskB.ID, taskC.ID, types.DependencyTypeFS, 0),
		makeDep(projectID, taskC.ID, taskA.ID, types.DependencyTypeFS, 0), // Creates cycle
	}

	g := BuildDependencyGraph(tasks, deps)
	err := DetectCycle(g)

	require.Error(t, err, "Cycle should be detected")
	assert.Contains(t, err.Error(), "cycle detected", "Error should mention cycle")
	// Per Step 34: error should contain WBS codes
	assert.Contains(t, err.Error(), "3.1", "Error should contain WBS code 3.1")
}

// TestDependencyGraph_MetadataIntegrity verifies lag_days is stored correctly.
func TestDependencyGraph_MetadataIntegrity(t *testing.T) {
	projectID := uuid.New()

	taskA := makeTask("4.1")
	taskB := makeTask("4.2")

	tasks := []models.ProjectTask{taskA, taskB}
	deps := []models.TaskDependency{
		makeDep(projectID, taskA.ID, taskB.ID, types.DependencyTypeFS, 5), // 5-day lag
	}

	g := BuildDependencyGraph(tasks, deps)

	// Verify edge metadata via Deps map
	dep, exists := g.GetDependency(taskA.ID, taskB.ID)

	require.True(t, exists, "Dependency should exist in graph")
	assert.Equal(t, 5, dep.LagDays, "Lag should be 5 days")
	assert.Equal(t, types.DependencyTypeFS, dep.DependencyType, "Type should be FS")
}

// TestBuildDependencyGraph_EmptyTasks handles edge case of empty input.
func TestBuildDependencyGraph_EmptyTasks(t *testing.T) {
	g := BuildDependencyGraph(nil, nil)

	assert.NotNil(t, g, "Graph should not be nil")
	assert.Equal(t, 0, g.Graph.Nodes().Len(), "Graph should have no nodes")

	sorted, err := TopologicalSort(g)
	require.NoError(t, err)
	assert.Len(t, sorted, 0, "Sorted list should be empty")
}

// indexOf returns the index of id in slice, or -1 if not found.
func indexOf(slice []uuid.UUID, id uuid.UUID) int {
	for i, v := range slice {
		if v == id {
			return i
		}
	}
	return -1
}

// Helper to create a ProjectTask with duration
func makeTaskWithDuration(wbsCode string, calculatedDuration float64) models.ProjectTask {
	return models.ProjectTask{
		ID:                 uuid.New(),
		WBSCode:            wbsCode,
		Name:               "Task " + wbsCode,
		CalculatedDuration: calculatedDuration,
	}
}

// TestForwardPass_LinearDAG verifies ES/EF calculation for A→B→C chain.
func TestForwardPass_LinearDAG(t *testing.T) {
	projectID := uuid.New()
	projectStart := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create tasks: A (2 days) → B (3 days) → C (1 day)
	taskA := makeTaskWithDuration("7.1", 2.0)
	taskB := makeTaskWithDuration("7.2", 3.0)
	taskC := makeTaskWithDuration("7.3", 1.0)

	tasks := []models.ProjectTask{taskA, taskB, taskC}
	deps := []models.TaskDependency{
		makeDep(projectID, taskA.ID, taskB.ID, types.DependencyTypeFS, 0),
		makeDep(projectID, taskB.ID, taskC.ID, types.DependencyTypeFS, 0),
	}

	g := BuildDependencyGraph(tasks, deps)
	schedule, err := ForwardPass(g, projectStart)

	require.NoError(t, err)
	require.Len(t, schedule, 3)

	// Task A: ES = project start, EF = ES + 2 days
	schedA := schedule[taskA.ID]
	assert.Equal(t, projectStart, schedA.EarlyStart, "A should start at project start")
	assert.Equal(t, projectStart.AddDate(0, 0, 2), schedA.EarlyFinish, "A should finish 2 days later")

	// Task B: ES = A.EF, EF = ES + 3 days
	schedB := schedule[taskB.ID]
	assert.Equal(t, schedA.EarlyFinish, schedB.EarlyStart, "B should start when A finishes")
	expectedBFinish := projectStart.AddDate(0, 0, 5) // Day 2 + 3 = Day 5
	assert.Equal(t, expectedBFinish, schedB.EarlyFinish, "B should finish at day 5")

	// Task C: ES = B.EF, EF = ES + 1 day
	schedC := schedule[taskC.ID]
	assert.Equal(t, schedB.EarlyFinish, schedC.EarlyStart, "C should start when B finishes")
	expectedCFinish := projectStart.AddDate(0, 0, 6) // Day 5 + 1 = Day 6
	assert.Equal(t, expectedCFinish, schedC.EarlyFinish, "C should finish at day 6")
}

// TestForwardPass_BranchingDAG verifies max() logic with multiple predecessors.
func TestForwardPass_BranchingDAG(t *testing.T) {
	projectID := uuid.New()
	projectStart := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// A (2 days) → D
	// B (5 days) → D (takes max)
	// C (3 days) → D
	taskA := makeTaskWithDuration("8.1", 2.0)
	taskB := makeTaskWithDuration("8.2", 5.0) // Longest path
	taskC := makeTaskWithDuration("8.3", 3.0)
	taskD := makeTaskWithDuration("8.4", 1.0)

	tasks := []models.ProjectTask{taskA, taskB, taskC, taskD}
	deps := []models.TaskDependency{
		makeDep(projectID, taskA.ID, taskD.ID, types.DependencyTypeFS, 0),
		makeDep(projectID, taskB.ID, taskD.ID, types.DependencyTypeFS, 0),
		makeDep(projectID, taskC.ID, taskD.ID, types.DependencyTypeFS, 0),
	}

	g := BuildDependencyGraph(tasks, deps)
	schedule, err := ForwardPass(g, projectStart)

	require.NoError(t, err)

	// D's ES should be max(A.EF, B.EF, C.EF) = B.EF (day 5)
	schedB := schedule[taskB.ID]
	schedD := schedule[taskD.ID]

	assert.Equal(t, schedB.EarlyFinish, schedD.EarlyStart, "D should start at max predecessor finish (B)")
	expectedDFinish := projectStart.AddDate(0, 0, 6) // Day 5 + 1 = Day 6
	assert.Equal(t, expectedDFinish, schedD.EarlyFinish, "D should finish at day 6")
}

// TestForwardPass_LagDays verifies lag is added to constraint date.
func TestForwardPass_LagDays(t *testing.T) {
	projectID := uuid.New()
	projectStart := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// A (2 days) --[5 day lag]--> B (1 day)
	taskA := makeTaskWithDuration("9.1", 2.0)
	taskB := makeTaskWithDuration("9.2", 1.0)

	tasks := []models.ProjectTask{taskA, taskB}
	deps := []models.TaskDependency{
		makeDep(projectID, taskA.ID, taskB.ID, types.DependencyTypeFS, 5), // 5-day lag
	}

	g := BuildDependencyGraph(tasks, deps)
	schedule, err := ForwardPass(g, projectStart)

	require.NoError(t, err)

	schedA := schedule[taskA.ID]
	schedB := schedule[taskB.ID]

	// B.ES = A.EF + 5 days lag
	expectedBStart := schedA.EarlyFinish.AddDate(0, 0, 5)
	assert.Equal(t, expectedBStart, schedB.EarlyStart, "B should start 5 days after A finishes")
}

// TestForwardPass_SSType verifies Start-to-Start dependency.
func TestForwardPass_SSType(t *testing.T) {
	projectID := uuid.New()
	projectStart := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// A (5 days) --[SS+2]--> B (3 days)
	// B starts 2 days after A starts
	taskA := makeTaskWithDuration("10.1", 5.0)
	taskB := makeTaskWithDuration("10.2", 3.0)

	tasks := []models.ProjectTask{taskA, taskB}
	deps := []models.TaskDependency{
		makeDep(projectID, taskA.ID, taskB.ID, types.DependencyTypeSS, 2), // SS with 2-day lag
	}

	g := BuildDependencyGraph(tasks, deps)
	schedule, err := ForwardPass(g, projectStart)

	require.NoError(t, err)

	schedA := schedule[taskA.ID]
	schedB := schedule[taskB.ID]

	// B.ES = A.ES + 2 days
	expectedBStart := schedA.EarlyStart.AddDate(0, 0, 2)
	assert.Equal(t, expectedBStart, schedB.EarlyStart, "B should start 2 days after A starts (SS)")
}

// TestForwardPass_FFType verifies Finish-to-Finish dependency.
func TestForwardPass_FFType(t *testing.T) {
	projectID := uuid.New()
	projectStart := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// A (5 days) --[FF+0]--> B (3 days)
	// B finishes when A finishes, so B.ES = A.EF - B.duration
	taskA := makeTaskWithDuration("11.1", 5.0)
	taskB := makeTaskWithDuration("11.2", 3.0)

	tasks := []models.ProjectTask{taskA, taskB}
	deps := []models.TaskDependency{
		makeDep(projectID, taskA.ID, taskB.ID, types.DependencyTypeFF, 0),
	}

	g := BuildDependencyGraph(tasks, deps)
	schedule, err := ForwardPass(g, projectStart)

	require.NoError(t, err)

	schedA := schedule[taskA.ID]
	schedB := schedule[taskB.ID]

	// B.EF = A.EF, so B.ES = A.EF - 3 days
	expectedBStart := schedA.EarlyFinish.AddDate(0, 0, -3)
	assert.Equal(t, expectedBStart, schedB.EarlyStart, "B should start 3 days before A finishes (FF)")
	assert.Equal(t, schedA.EarlyFinish, schedB.EarlyFinish, "B should finish when A finishes (FF)")
}

// TestForwardPass_SFType verifies Start-to-Finish dependency.
func TestForwardPass_SFType(t *testing.T) {
	projectID := uuid.New()
	projectStart := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// A (5 days) --[SF+0]--> B (3 days)
	// SF: Successor finishes after predecessor starts
	// EF(B) = ES(A) + lag, so ES(B) = ES(A) + lag - duration(B)
	taskA := makeTaskWithDuration("21.1", 5.0)
	taskB := makeTaskWithDuration("21.2", 3.0)

	tasks := []models.ProjectTask{taskA, taskB}
	deps := []models.TaskDependency{
		makeDep(projectID, taskA.ID, taskB.ID, types.DependencyTypeSF, 0),
	}

	g := BuildDependencyGraph(tasks, deps)
	schedule, err := ForwardPass(g, projectStart)

	require.NoError(t, err)

	schedA := schedule[taskA.ID]
	schedB := schedule[taskB.ID]

	// B.EF = A.ES + lag = Jan 15
	// B.ES = B.EF - 3 = Jan 12
	expectedBStart := schedA.EarlyStart.AddDate(0, 0, -3)
	assert.Equal(t, expectedBStart, schedB.EarlyStart, "B.ES should be A.ES minus B duration (SF)")
}

// TestForwardPass_RootTask verifies root tasks start at project start.
func TestForwardPass_RootTask(t *testing.T) {
	projectStart := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	// Single task with no predecessors
	taskA := makeTaskWithDuration("5.3", 3.0)

	tasks := []models.ProjectTask{taskA}
	deps := []models.TaskDependency{} // No dependencies

	g := BuildDependencyGraph(tasks, deps)
	schedule, err := ForwardPass(g, projectStart)

	require.NoError(t, err)

	schedA := schedule[taskA.ID]
	assert.Equal(t, projectStart, schedA.EarlyStart, "Root task should start at project start")
	expectedFinish := projectStart.AddDate(0, 0, 3)
	assert.Equal(t, expectedFinish, schedA.EarlyFinish, "Root task should finish after duration")
}

// TestForwardPass_DurationPrecedence verifies ManualOverride > WeatherAdjusted > Calculated.
func TestForwardPass_DurationPrecedence(t *testing.T) {
	projectStart := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// Task with all three durations set
	override := 10.0
	taskA := models.ProjectTask{
		ID:                      uuid.New(),
		WBSCode:                 "12.1",
		Name:                    "Override Test",
		CalculatedDuration:      2.0,
		WeatherAdjustedDuration: 5.0,
		ManualOverrideDays:      &override,
	}

	tasks := []models.ProjectTask{taskA}
	deps := []models.TaskDependency{}

	g := BuildDependencyGraph(tasks, deps)
	schedule, err := ForwardPass(g, projectStart)

	require.NoError(t, err)

	schedA := schedule[taskA.ID]
	expectedFinish := projectStart.AddDate(0, 0, 10) // Should use override (10 days)
	assert.Equal(t, expectedFinish, schedA.EarlyFinish, "Should use ManualOverrideDays (10) over others")
}

// TestForwardPass_EmptyGraph handles edge case of empty input.
func TestForwardPass_EmptyGraph(t *testing.T) {
	projectStart := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	g := BuildDependencyGraph(nil, nil)
	schedule, err := ForwardPass(g, projectStart)

	require.NoError(t, err)
	assert.Len(t, schedule, 0, "Empty graph should produce empty schedule")
}

// ============================================================================
// Backward Pass Tests
// ============================================================================

// TestBackwardPass_LinearDAG verifies LS/LF calculation for A→B→C chain.
func TestBackwardPass_LinearDAG(t *testing.T) {
	projectID := uuid.New()
	projectStart := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create tasks: A (2 days) → B (3 days) → C (1 day)
	taskA := makeTaskWithDuration("13.1", 2.0)
	taskB := makeTaskWithDuration("13.2", 3.0)
	taskC := makeTaskWithDuration("13.3", 1.0)

	tasks := []models.ProjectTask{taskA, taskB, taskC}
	deps := []models.TaskDependency{
		makeDep(projectID, taskA.ID, taskB.ID, types.DependencyTypeFS, 0),
		makeDep(projectID, taskB.ID, taskC.ID, types.DependencyTypeFS, 0),
	}

	g := BuildDependencyGraph(tasks, deps)
	schedule, err := ForwardPass(g, projectStart)
	require.NoError(t, err)

	criticalPath, err := BackwardPass(g, schedule)
	require.NoError(t, err)

	// All tasks should be critical in a linear chain
	assert.Len(t, criticalPath, 3, "All tasks should be on critical path")

	// Task C (terminal): LF = project end (day 6), LS = day 5
	schedC := schedule[taskC.ID]
	expectedCLF := projectStart.AddDate(0, 0, 6) // Project end
	assert.Equal(t, expectedCLF, schedC.LateFinish, "C.LF should be project end")
	assert.Equal(t, schedC.EarlyStart, schedC.LateStart, "C should have zero float")
	assert.True(t, schedC.IsCritical, "C should be critical")

	// Task B: LF = C.LS, LS = LF - 3
	schedB := schedule[taskB.ID]
	assert.Equal(t, schedC.LateStart, schedB.LateFinish, "B.LF should equal C.LS")
	assert.True(t, schedB.IsCritical, "B should be critical")

	// Task A: LF = B.LS, LS = LF - 2
	schedA := schedule[taskA.ID]
	assert.Equal(t, schedB.LateStart, schedA.LateFinish, "A.LF should equal B.LS")
	assert.True(t, schedA.IsCritical, "A should be critical")
}

// TestBackwardPass_BranchingDAG verifies min() logic with multiple successors.
func TestBackwardPass_BranchingDAG(t *testing.T) {
	projectID := uuid.New()
	projectStart := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// A (2 days) → B (3 days),  A → C (1 day)
	// B and C are parallel, but B is longer - determines project end
	taskA := makeTaskWithDuration("14.1", 2.0)
	taskB := makeTaskWithDuration("14.2", 3.0) // Longer path
	taskC := makeTaskWithDuration("14.3", 1.0) // Shorter path

	tasks := []models.ProjectTask{taskA, taskB, taskC}
	deps := []models.TaskDependency{
		makeDep(projectID, taskA.ID, taskB.ID, types.DependencyTypeFS, 0),
		makeDep(projectID, taskA.ID, taskC.ID, types.DependencyTypeFS, 0),
	}

	g := BuildDependencyGraph(tasks, deps)
	schedule, err := ForwardPass(g, projectStart)
	require.NoError(t, err)

	criticalPath, err := BackwardPass(g, schedule)
	require.NoError(t, err)

	// A and B should be critical; C has float
	assert.Contains(t, criticalPath, "14.1", "A should be on critical path")
	assert.Contains(t, criticalPath, "14.2", "B should be on critical path")
	assert.NotContains(t, criticalPath, "14.3", "C should NOT be on critical path")

	// A's LF is constrained by min(B.LS, C.LS) = B.LS (tighter constraint)
	schedA := schedule[taskA.ID]
	schedB := schedule[taskB.ID]
	assert.Equal(t, schedB.LateStart, schedA.LateFinish, "A.LF should be constrained by B.LS")

	// C should have positive float
	schedC := schedule[taskC.ID]
	assert.Greater(t, schedC.TotalFloat, 0.0, "C should have positive float")
}

// TestBackwardPass_LagDays verifies lag subtraction in backward pass.
func TestBackwardPass_LagDays(t *testing.T) {
	projectID := uuid.New()
	projectStart := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// A (2 days) --[5 day lag]--> B (1 day)
	taskA := makeTaskWithDuration("15.1", 2.0)
	taskB := makeTaskWithDuration("15.2", 1.0)

	tasks := []models.ProjectTask{taskA, taskB}
	deps := []models.TaskDependency{
		makeDep(projectID, taskA.ID, taskB.ID, types.DependencyTypeFS, 5),
	}

	g := BuildDependencyGraph(tasks, deps)
	schedule, err := ForwardPass(g, projectStart)
	require.NoError(t, err)

	_, err = BackwardPass(g, schedule)
	require.NoError(t, err)

	schedA := schedule[taskA.ID]
	schedB := schedule[taskB.ID]

	// A.LF = B.LS - 5 lag days
	expectedALF := schedB.LateStart.AddDate(0, 0, -5)
	assert.Equal(t, expectedALF, schedA.LateFinish, "A.LF should account for 5-day lag")
}

// TestBackwardPass_CriticalPath verifies tasks with Float=0 are correctly identified.
func TestBackwardPass_CriticalPath(t *testing.T) {
	projectID := uuid.New()
	projectStart := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create diamond pattern: A → B, A → C, B → D, C → D
	// With B longer than C, A-B-D is critical
	taskA := makeTaskWithDuration("16.1", 1.0)
	taskB := makeTaskWithDuration("16.2", 5.0) // Longer
	taskC := makeTaskWithDuration("16.3", 2.0) // Shorter
	taskD := makeTaskWithDuration("16.4", 1.0)

	tasks := []models.ProjectTask{taskA, taskB, taskC, taskD}
	deps := []models.TaskDependency{
		makeDep(projectID, taskA.ID, taskB.ID, types.DependencyTypeFS, 0),
		makeDep(projectID, taskA.ID, taskC.ID, types.DependencyTypeFS, 0),
		makeDep(projectID, taskB.ID, taskD.ID, types.DependencyTypeFS, 0),
		makeDep(projectID, taskC.ID, taskD.ID, types.DependencyTypeFS, 0),
	}

	g := BuildDependencyGraph(tasks, deps)
	schedule, err := ForwardPass(g, projectStart)
	require.NoError(t, err)

	criticalPath, err := BackwardPass(g, schedule)
	require.NoError(t, err)

	// Critical path should be A → B → D
	assert.Len(t, criticalPath, 3, "Critical path should have 3 tasks")
	assert.Contains(t, criticalPath, "16.1", "A is critical")
	assert.Contains(t, criticalPath, "16.2", "B is critical")
	assert.Contains(t, criticalPath, "16.4", "D is critical")
	assert.NotContains(t, criticalPath, "16.3", "C is NOT critical")

	// Verify C has positive float
	schedC := schedule[taskC.ID]
	assert.Equal(t, 3.0, schedC.TotalFloat, "C should have 3 days float (5-2)")
}

// TestBackwardPass_SSType verifies Start-to-Start backward constraint.
func TestBackwardPass_SSType(t *testing.T) {
	projectID := uuid.New()
	projectStart := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// A (5 days) --[SS+2]--> B (3 days)
	taskA := makeTaskWithDuration("17.1", 5.0)
	taskB := makeTaskWithDuration("17.2", 3.0)

	tasks := []models.ProjectTask{taskA, taskB}
	deps := []models.TaskDependency{
		makeDep(projectID, taskA.ID, taskB.ID, types.DependencyTypeSS, 2),
	}

	g := BuildDependencyGraph(tasks, deps)
	schedule, err := ForwardPass(g, projectStart)
	require.NoError(t, err)

	_, err = BackwardPass(g, schedule)
	require.NoError(t, err)

	schedA := schedule[taskA.ID]
	schedB := schedule[taskB.ID]

	// With SS: LS(A) = LS(B) - lag = LS(B) - 2
	expectedALS := schedB.LateStart.AddDate(0, 0, -2)
	assert.Equal(t, expectedALS, schedA.LateStart, "A.LS should be B.LS minus 2 days (SS)")
}

// TestBackwardPass_FFType verifies Finish-to-Finish backward constraint.
func TestBackwardPass_FFType(t *testing.T) {
	projectID := uuid.New()
	projectStart := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// A (5 days) --[FF+0]--> B (3 days)
	taskA := makeTaskWithDuration("18.1", 5.0)
	taskB := makeTaskWithDuration("18.2", 3.0)

	tasks := []models.ProjectTask{taskA, taskB}
	deps := []models.TaskDependency{
		makeDep(projectID, taskA.ID, taskB.ID, types.DependencyTypeFF, 0),
	}

	g := BuildDependencyGraph(tasks, deps)
	schedule, err := ForwardPass(g, projectStart)
	require.NoError(t, err)

	_, err = BackwardPass(g, schedule)
	require.NoError(t, err)

	schedA := schedule[taskA.ID]
	schedB := schedule[taskB.ID]

	// With FF: LF(A) = LF(B) - lag = LF(B)
	assert.Equal(t, schedB.LateFinish, schedA.LateFinish, "A.LF should equal B.LF (FF)")
}

// TestBackwardPass_SFType verifies Start-to-Finish backward constraint.
func TestBackwardPass_SFType(t *testing.T) {
	projectID := uuid.New()
	projectStart := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// A (5 days) --[SF+0]--> B (3 days)
	// SF backward: LS(pred) = LF(succ) - lag
	taskA := makeTaskWithDuration("22.1", 5.0)
	taskB := makeTaskWithDuration("22.2", 3.0)

	tasks := []models.ProjectTask{taskA, taskB}
	deps := []models.TaskDependency{
		makeDep(projectID, taskA.ID, taskB.ID, types.DependencyTypeSF, 0),
	}

	g := BuildDependencyGraph(tasks, deps)
	schedule, err := ForwardPass(g, projectStart)
	require.NoError(t, err)

	_, err = BackwardPass(g, schedule)
	require.NoError(t, err)

	schedA := schedule[taskA.ID]
	schedB := schedule[taskB.ID]

	// With SF: LS(A) = LF(B) - lag
	expectedALS := schedB.LateFinish.AddDate(0, 0, 0) // lag = 0
	assert.Equal(t, expectedALS, schedA.LateStart, "A.LS should equal B.LF (SF)")
}

// TestBackwardPass_TerminalTask verifies terminal tasks have LF = project end.
func TestBackwardPass_TerminalTask(t *testing.T) {
	projectStart := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// Single task with no successors
	taskA := makeTaskWithDuration("19.1", 5.0)

	tasks := []models.ProjectTask{taskA}
	deps := []models.TaskDependency{}

	g := BuildDependencyGraph(tasks, deps)
	schedule, err := ForwardPass(g, projectStart)
	require.NoError(t, err)

	_, err = BackwardPass(g, schedule)
	require.NoError(t, err)

	schedA := schedule[taskA.ID]

	// Terminal task: LF = EF (project end)
	assert.Equal(t, schedA.EarlyFinish, schedA.LateFinish, "Terminal task LF should equal EF")
	assert.Equal(t, schedA.EarlyStart, schedA.LateStart, "Terminal task LS should equal ES")
	assert.Equal(t, 0.0, schedA.TotalFloat, "Terminal task should have zero float")
	assert.True(t, schedA.IsCritical, "Terminal task should be critical")
}

// TestBackwardPass_FloatCalculation verifies non-critical tasks show positive float.
func TestBackwardPass_FloatCalculation(t *testing.T) {
	projectID := uuid.New()
	projectStart := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// A (1 day) has two successors:
	// A → B (10 days) - determines project end
	// A → C (2 days)  - has 8 days of float
	taskA := makeTaskWithDuration("20.1", 1.0)
	taskB := makeTaskWithDuration("20.2", 10.0)
	taskC := makeTaskWithDuration("20.3", 2.0)

	tasks := []models.ProjectTask{taskA, taskB, taskC}
	deps := []models.TaskDependency{
		makeDep(projectID, taskA.ID, taskB.ID, types.DependencyTypeFS, 0),
		makeDep(projectID, taskA.ID, taskC.ID, types.DependencyTypeFS, 0),
	}

	g := BuildDependencyGraph(tasks, deps)
	schedule, err := ForwardPass(g, projectStart)
	require.NoError(t, err)

	_, err = BackwardPass(g, schedule)
	require.NoError(t, err)

	// C should have 8 days of float (10 - 2 = 8)
	schedC := schedule[taskC.ID]
	assert.Equal(t, 8.0, schedC.TotalFloat, "C should have 8 days of float")
	assert.False(t, schedC.IsCritical, "C should NOT be critical")

	// B should have zero float
	schedB := schedule[taskB.ID]
	assert.Equal(t, 0.0, schedB.TotalFloat, "B should have zero float")
	assert.True(t, schedB.IsCritical, "B should be critical")
}

// TestBackwardPass_EmptySchedule handles edge case of empty schedule.
func TestBackwardPass_EmptySchedule(t *testing.T) {
	g := BuildDependencyGraph(nil, nil)
	schedule := make(map[uuid.UUID]TaskSchedule)

	criticalPath, err := BackwardPass(g, schedule)

	require.NoError(t, err)
	assert.Nil(t, criticalPath, "Empty schedule should produce nil critical path")
}
