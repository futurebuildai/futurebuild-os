# cpm

## Intent
*   **High Level:** Critical Path Method schedule engine — computes forward/backward passes, float, and the critical path for a project task graph.
*   **Business Value:** Produces the schedule that drives the Gantt view and daily priority calculations.

## Responsibility
*   Accept a set of tasks with durations and dependency edges.
*   Compute Early Start, Early Finish, Late Start, Late Finish, and Total Float for each task.
*   Identify the critical path (tasks where Total Float = 0).
*   Return the sorted task list, critical path IDs, and overall project duration.

## Key Logic
*   **Topological Sort (Kahn's Algorithm):** Validates the DAG and produces execution order.
*   **Forward Pass:** Early Start = max(predecessor Early Finish); Early Finish = Early Start + Duration.
*   **Backward Pass:** Late Finish = min(successor Late Start); Late Start = Late Finish − Duration.
*   **Float Calculation:** Total Float = Late Start − Early Start; isCritical = (Float === 0).

## Reference Implementation
A TypeScript implementation exists at `frontend/src/services/cpm-engine.ts` and serves as the source of truth for algorithm correctness. The Go implementation should produce identical results.

## API Surface
```go
// POST /api/v1/projects/:id/schedule/generate
// Accepts WBS from Interrogator extraction, calls cpm.Calculate(), stores result.
// Returns GanttData with critical path highlighted.

type Task struct {
    ID           string
    Name         string
    Duration     int      // days
    Dependencies []string
    EarlyStart   int      // computed
    EarlyFinish  int      // computed
    LateStart    int      // computed
    LateFinish   int      // computed
    TotalFloat   int      // computed
    IsCritical   bool     // computed
}

type Schedule struct {
    Tasks        []Task
    CriticalPath []string
    Duration     int
}

func Calculate(tasks []Task) (*Schedule, error)
```

## Dependencies
*   **Upstream:** ScheduleHandler (HTTP endpoint), InterrogatorService (provides extracted WBS)
*   **Downstream:** None (leaf computation)
