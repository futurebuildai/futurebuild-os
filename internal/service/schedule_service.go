// Package service provides business logic for the FutureBuild application.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/physics"
	"github.com/colton/futurebuild/internal/platform/db"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ScheduleService handles schedule recalculation via CPM.
// See BACKEND_SCOPE.md Section 3.4 (Schedule Recalculator)
type ScheduleService struct {
	db *pgxpool.Pool
}

// NewScheduleService creates a new ScheduleService instance.
func NewScheduleService(db *pgxpool.Pool) *ScheduleService {
	return &ScheduleService{db: db}
}

// ProjectScheduleSummary represents a high-level schedule overview.
// Used by the Chat Orchestrator to provide schedule status responses.
type ProjectScheduleSummary struct {
	ProjectEnd        time.Time
	CriticalPathCount int
	TotalTasks        int
	CompletedTasks    int
}

// GetProjectSchedule returns a summary of the project's schedule status.
// See PRODUCTION_PLAN.md Step 43 (Chat Orchestrator Command Pattern)
func (s *ScheduleService) GetProjectSchedule(ctx context.Context, projectID, orgID uuid.UUID) (*ProjectScheduleSummary, error) {
	// Verify project ownership and fetch target_end_date
	var targetEndDate *time.Time
	err := s.db.QueryRow(ctx, `
		SELECT target_end_date FROM projects 
		WHERE id = $1 AND org_id = $2
	`, projectID, orgID).Scan(&targetEndDate)
	if err != nil {
		return nil, fmt.Errorf("project not found or access denied: %w", err)
	}

	// Aggregate task statistics
	var totalTasks, completedTasks, criticalPathCount int
	err = s.db.QueryRow(ctx, `
		SELECT 
			COUNT(*) AS total_tasks,
			COUNT(*) FILTER (WHERE status = 'complete') AS completed_tasks,
			COUNT(*) FILTER (WHERE is_on_critical_path = true) AS critical_path_count
		FROM project_tasks WHERE project_id = $1
	`, projectID).Scan(&totalTasks, &completedTasks, &criticalPathCount)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate task stats: %w", err)
	}

	summary := &ProjectScheduleSummary{
		TotalTasks:        totalTasks,
		CompletedTasks:    completedTasks,
		CriticalPathCount: criticalPathCount,
	}
	if targetEndDate != nil {
		summary.ProjectEnd = *targetEndDate
	}

	return summary, nil
}

// GetAgentFocusTasks fetches tasks relevant for the Daily Focus briefing.
// Returns: In Progress, Starting Soon (7 days), or Critical Path tasks.
// See PRODUCTION_PLAN.md Step 49 (Service Layer Pattern)
func (s *ScheduleService) GetAgentFocusTasks(ctx context.Context, projectID uuid.UUID) ([]models.ProjectTask, error) {
	query := `
		SELECT id, name, status, planned_start, planned_end, is_on_critical_path
		FROM project_tasks
		WHERE project_id = $1
		  AND (
			  status = 'in_progress'
			  OR (status = 'pending' AND planned_start <= CURRENT_DATE + INTERVAL '7 days')
			  OR (is_on_critical_path = true AND status != 'completed')
		  )
		ORDER BY planned_start ASC
		LIMIT 20
	`
	rows, err := s.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query agent focus tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.ProjectTask
	for rows.Next() {
		var t models.ProjectTask
		err := rows.Scan(&t.ID, &t.Name, &t.Status, &t.PlannedStart, &t.PlannedEnd, &t.IsOnCriticalPath)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// RecalculateSchedule runs the full CPM (ForwardPass + BackwardPass) for a project.
// Updates all task ES/EF/LS/LF values and the project's target_end_date.
// See PRODUCTION_PLAN.md Step 32
func (s *ScheduleService) RecalculateSchedule(ctx context.Context, projectID, orgID uuid.UUID) (*physics.CPMResult, error) {
	// Step 1: Verify project ownership (multi-tenancy)
	var permitDate *time.Time
	err := s.db.QueryRow(ctx, `
		SELECT permit_issued_date FROM projects 
		WHERE id = $1 AND org_id = $2
	`, projectID, orgID).Scan(&permitDate)
	if err != nil {
		return nil, fmt.Errorf("project not found or access denied: %w", err)
	}

	// Logic Fix: Fallback to current time if permit not yet issued
	// See /CTO Audit Correction #2
	projectStart := time.Now().UTC()
	if permitDate != nil {
		projectStart = *permitDate
	}

	// Step 2: Fetch all tasks for the project
	// See DATA_SPINE_SPEC.md Section 3.3
	tasks, err := s.getProjectTasks(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks: %w", err)
	}

	if len(tasks) == 0 {
		return &physics.CPMResult{
			Tasks:        []physics.TaskSchedule{},
			CriticalPath: []string{},
			ProjectEnd:   projectStart,
		}, nil
	}

	// Step 3: Fetch all dependencies for the project
	// See DATA_SPINE_SPEC.md Section 3.4
	deps, err := s.getProjectDependencies(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dependencies: %w", err)
	}

	// Step 3.5: Fetch material constraints from procurement_items
	// See PRODUCTION_PLAN.md Step 46 (MRP Feedback Loop)
	materialConstraints, err := s.getMaterialConstraints(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch material constraints: %w", err)
	}

	// Step 4: Build dependency graph and run CPM
	// See BACKEND_SCOPE.md Section 6.3
	graph := physics.BuildDependencyGraph(tasks, deps)

	// Detect cycles before processing
	if err := physics.DetectCycle(graph); err != nil {
		return nil, fmt.Errorf("schedule contains circular dependencies: %w", err)
	}

	schedule, err := physics.ForwardPass(graph, projectStart, &physics.StandardCalendar{}, materialConstraints)
	if err != nil {
		return nil, fmt.Errorf("forward pass failed: %w", err)
	}

	criticalPath, err := physics.BackwardPass(graph, schedule, &physics.StandardCalendar{}, physics.DefaultSchedulingConfig())
	if err != nil {
		return nil, fmt.Errorf("backward pass failed: %w", err)
	}

	// Step 5: Batch-update all task schedules in a single transaction
	// FAANG STANDARD: Replaced N+1 Write Anti-Pattern with Bulk Update via UNNEST.
	// See PRODUCTION_PLAN.md Step 32 (Optimization: transaction-based batch update)
	//
	// Distributed Transaction Support (Step 45 Zombie Write Fix):
	// - If called within a Lane B flow with injected Tx, use that transaction (caller owns lifecycle)
	// - Otherwise, start our own transaction (service owns lifecycle)
	var tx pgx.Tx
	var txErr error
	localTx := false // Flag: did we start this transaction?

	if injectedTx, ok := db.ExtractTx(ctx); ok {
		tx = injectedTx
	} else {
		tx, txErr = s.db.Begin(ctx)
		if txErr != nil {
			return nil, fmt.Errorf("failed to start transaction: %w", txErr)
		}
		localTx = true
		defer func() {
			if localTx {
				_ = tx.Rollback(ctx)
			}
		}()
	}

	// 5a. Prepare Columnar Slices (In-Memory)
	// Performance: O(N) memory, O(1) network round-trips
	count := len(schedule)
	ids := make([]uuid.UUID, 0, count)
	earlyStarts := make([]time.Time, 0, count)
	earlyFinishes := make([]time.Time, 0, count)
	lateStarts := make([]time.Time, 0, count)
	lateFinishes := make([]time.Time, 0, count)
	totalFloats := make([]float64, 0, count)
	isCriticals := make([]bool, 0, count)

	var projectEnd time.Time

	for _, sched := range schedule {
		ids = append(ids, sched.TaskID)
		earlyStarts = append(earlyStarts, sched.EarlyStart)
		earlyFinishes = append(earlyFinishes, sched.EarlyFinish)
		lateStarts = append(lateStarts, sched.LateStart)
		lateFinishes = append(lateFinishes, sched.LateFinish)
		totalFloats = append(totalFloats, sched.TotalFloat)
		isCriticals = append(isCriticals, sched.IsCritical)

		// Track max project end date
		if sched.EarlyFinish.After(projectEnd) {
			projectEnd = sched.EarlyFinish
		}
	}

	// 5b. Execute Single Bulk Update Query
	// Security: $8 (projectID) enforces multi-tenancy scope
	if count > 0 {
		query := `
			UPDATE project_tasks pt
			SET 
				early_start = data.early_start,
				early_finish = data.early_finish,
				late_start = data.late_start,
				late_finish = data.late_finish,
				total_float_days = data.total_float,
				is_on_critical_path = data.is_critical,
				updated_at = NOW()
			FROM (
				SELECT * FROM UNNEST(
					$1::uuid[], $2::timestamptz[], $3::timestamptz[], 
					$4::timestamptz[], $5::timestamptz[], $6::float8[], $7::bool[]
				) AS t(id, early_start, early_finish, late_start, late_finish, total_float, is_critical)
			) AS data
			WHERE pt.id = data.id AND pt.project_id = $8
		`
		_, err = tx.Exec(ctx, query,
			ids, earlyStarts, earlyFinishes, lateStarts, lateFinishes, totalFloats, isCriticals,
			projectID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to bulk update task schedules: %w", err)
		}
	}

	// Step 6: Update project's target_end_date
	_, err = tx.Exec(ctx, `
		UPDATE projects SET target_end_date = $1, updated_at = NOW() WHERE id = $2
	`, projectEnd, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to update project end date: %w", err)
	}

	// Only commit if we started the transaction (localTx)
	if localTx {
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	// Build result
	result := &physics.CPMResult{
		CriticalPath: criticalPath,
		ProjectEnd:   projectEnd,
		Tasks:        make([]physics.TaskSchedule, 0, len(schedule)),
	}
	for _, sched := range schedule {
		result.Tasks = append(result.Tasks, sched)
	}

	return result, nil
}

// getProjectTasks fetches all tasks for a project.
func (s *ScheduleService) getProjectTasks(ctx context.Context, projectID uuid.UUID) ([]models.ProjectTask, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, project_id, wbs_code, name, is_inspection,
		       calculated_duration, weather_adjusted_duration, manual_override_days,
		       status
		FROM project_tasks WHERE project_id = $1
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.ProjectTask
	for rows.Next() {
		var t models.ProjectTask
		err := rows.Scan(&t.ID, &t.ProjectID, &t.WBSCode, &t.Name, &t.IsInspection,
			&t.CalculatedDuration, &t.WeatherAdjustedDuration, &t.ManualOverrideDays,
			&t.Status)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// getProjectDependencies fetches all task dependencies for a project.
func (s *ScheduleService) getProjectDependencies(ctx context.Context, projectID uuid.UUID) ([]models.TaskDependency, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, project_id, predecessor_id, successor_id, dependency_type, lag_days
		FROM task_dependencies WHERE project_id = $1
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []models.TaskDependency
	for rows.Next() {
		var d models.TaskDependency
		err := rows.Scan(&d.ID, &d.ProjectID, &d.PredecessorID, &d.SuccessorID,
			&d.DependencyType, &d.LagDays)
		if err != nil {
			return nil, err
		}
		deps = append(deps, d)
	}
	return deps, rows.Err()
}

// GetTask fetches a single task by ID with multi-tenancy check.
// See DATA_SPINE_SPEC.md Section 3.3
func (s *ScheduleService) GetTask(ctx context.Context, taskID, projectID, orgID uuid.UUID) (*models.ProjectTask, error) {
	var t models.ProjectTask
	err := s.db.QueryRow(ctx, `
		SELECT pt.id, pt.project_id, pt.wbs_code, pt.name, pt.is_inspection,
		       pt.calculated_duration, pt.weather_adjusted_duration, pt.manual_override_days,
		       pt.status, pt.early_start, pt.early_finish
		FROM project_tasks pt
		JOIN projects p ON pt.project_id = p.id
		WHERE pt.id = $1 AND pt.project_id = $2 AND p.org_id = $3
	`, taskID, projectID, orgID).Scan(
		&t.ID, &t.ProjectID, &t.WBSCode, &t.Name, &t.IsInspection,
		&t.CalculatedDuration, &t.WeatherAdjustedDuration, &t.ManualOverrideDays,
		&t.Status, &t.EarlyStart, &t.EarlyFinish)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}
	return &t, nil
}

// UpdateTaskDuration updates a task's manual override duration.
// See PRODUCTION_PLAN.md Step 32 (Trigger Point 1)
func (s *ScheduleService) UpdateTaskDuration(ctx context.Context, taskID, projectID, orgID uuid.UUID, overrideDays float64, reason string) error {
	// Harden multi-tenancy by joining with projects
	_, err := s.db.Exec(ctx, `
		UPDATE project_tasks pt
		SET manual_override_days = $1, override_reason = $2, updated_at = NOW()
		FROM projects p
		WHERE pt.project_id = p.id AND pt.id = $3 AND pt.project_id = $4 AND p.org_id = $5
	`, overrideDays, reason, taskID, projectID, orgID)
	return err
}

// UpdateTaskStatus updates a task's status.
// See CPM_RES_MODEL_SPEC.md Section 20.2 (Task Status Transitions)
//
// Distributed Transaction Support (Step 45 Zombie Write Fix):
// - If a transaction is injected via context (db.InjectTx), uses that transaction.
// - Otherwise, uses the connection pool directly (legacy behavior).
func (s *ScheduleService) UpdateTaskStatus(ctx context.Context, taskID, projectID, orgID uuid.UUID, status types.TaskStatus) error {
	updateSQL := `
		UPDATE project_tasks pt
		SET status = $1, updated_at = NOW()
		FROM projects p
		WHERE pt.project_id = p.id AND pt.id = $2 AND pt.project_id = $3 AND p.org_id = $4
	`

	// Check for context-propagated transaction (caller owns Tx lifecycle)
	if tx, ok := db.ExtractTx(ctx); ok {
		_, err := tx.Exec(ctx, updateSQL, status, taskID, projectID, orgID)
		return err
	}

	// Legacy behavior: use pool directly
	_, err := s.db.Exec(ctx, updateSQL, status, taskID, projectID, orgID)
	return err
}

// CreateTaskProgress records a progress update in the task_progress table.
// See BACKEND_SCOPE.md Section 3.4 (Live Project Graph)
func (s *ScheduleService) CreateTaskProgress(ctx context.Context, projectID, taskID, userID uuid.UUID, percentComplete int, notes string) error {
	// Harden security: ensure task belongs to project
	_, err := s.db.Exec(ctx, `
		INSERT INTO task_progress (id, task_id, reported_by, reported_at, percent_complete, notes)
		SELECT $1, $2, $3, NOW(), $4, $5
		FROM project_tasks WHERE id = $2 AND project_id = $6
	`, uuid.New(), taskID, userID, percentComplete, notes, projectID)
	return err
}

// CreateInspectionRecord records an inspection result in the inspection_records table.
// See BACKEND_SCOPE.md Section 3.4 (Live Project Graph)
func (s *ScheduleService) CreateInspectionRecord(ctx context.Context, projectID, taskID uuid.UUID, inspectorName, result, notes string, inspectionDate time.Time) error {
	// Harden security: ensure task belongs to project
	_, err := s.db.Exec(ctx, `
		INSERT INTO inspection_records (id, task_id, inspector_name, inspection_date, result, notes, recorded_at)
		SELECT $1, $2, $3, $4, $5, $6, NOW()
		FROM project_tasks WHERE id = $2 AND project_id = $7
	`, uuid.New(), taskID, inspectorName, inspectionDate, result, notes, projectID)
	return err
}

// GetGanttData builds the full GanttData structure for a project, including tasks and dependencies.
// Used by the ScheduleHandler to serve GET /projects/{id}/schedule.
// See STEP_89_DEPENDENCY_ARROWS.md (Phase 14: Gantt dependency arrows)
func (s *ScheduleService) GetGanttData(ctx context.Context, projectID, orgID uuid.UUID) (*types.GanttData, error) {
	// Step 1: Verify project ownership and fetch metadata
	var targetEndDate *time.Time
	err := s.db.QueryRow(ctx, `
		SELECT target_end_date FROM projects
		WHERE id = $1 AND org_id = $2
	`, projectID, orgID).Scan(&targetEndDate)
	if err != nil {
		return nil, fmt.Errorf("project not found or access denied: %w", err)
	}

	// Step 2: Fetch all tasks with CPM schedule data
	rows, err := s.db.Query(ctx, `
		SELECT wbs_code, name, status, early_start, early_finish,
		       COALESCE(manual_override_days, weather_adjusted_duration, calculated_duration, 0) AS duration_days,
		       is_on_critical_path
		FROM project_tasks
		WHERE project_id = $1
		ORDER BY wbs_code ASC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks: %w", err)
	}
	defer rows.Close()

	var tasks []types.GanttTask
	var criticalPath []string

	for rows.Next() {
		var t types.GanttTask
		var earlyStart, earlyFinish *time.Time

		err := rows.Scan(&t.WBSCode, &t.Name, &t.Status, &earlyStart, &earlyFinish,
			&t.DurationDays, &t.IsCritical)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		if earlyStart != nil {
			t.EarlyStart = earlyStart.Format("2006-01-02")
		}
		if earlyFinish != nil {
			t.EarlyFinish = earlyFinish.Format("2006-01-02")
		}
		if t.IsCritical {
			criticalPath = append(criticalPath, t.WBSCode)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tasks: %w", err)
	}

	// Step 3: Fetch dependencies and map task UUIDs to WBS codes
	// We need a join to resolve predecessor/successor UUIDs to WBS codes
	depRows, err := s.db.Query(ctx, `
		SELECT pred.wbs_code, succ.wbs_code
		FROM task_dependencies td
		JOIN project_tasks pred ON td.predecessor_id = pred.id
		JOIN project_tasks succ ON td.successor_id = succ.id
		WHERE td.project_id = $1
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dependencies: %w", err)
	}
	defer depRows.Close()

	var dependencies []types.GanttDependency
	for depRows.Next() {
		var dep types.GanttDependency
		if err := depRows.Scan(&dep.From, &dep.To); err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		dependencies = append(dependencies, dep)
	}
	if err := depRows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate dependencies: %w", err)
	}

	// Step 4: Build GanttData
	projectedEnd := ""
	if targetEndDate != nil {
		projectedEnd = targetEndDate.Format("2006-01-02")
	}

	ganttData := &types.GanttData{
		ProjectID:        projectID,
		CalculatedAt:     time.Now().UTC().Format(time.RFC3339),
		ProjectedEndDate: projectedEnd,
		CriticalPath:     criticalPath,
		Tasks:            tasks,
		Dependencies:     dependencies,
	}

	// Ensure non-nil slices for clean JSON serialization ([] not null)
	if ganttData.Tasks == nil {
		ganttData.Tasks = []types.GanttTask{}
	}
	if ganttData.CriticalPath == nil {
		ganttData.CriticalPath = []string{}
	}
	if ganttData.Dependencies == nil {
		ganttData.Dependencies = []types.GanttDependency{}
	}

	return ganttData, nil
}

// getMaterialConstraints fetches expected delivery dates for procurement items.
// Returns map of TaskID -> EarliestAvailableDate for material constraint enforcement.
// See PRODUCTION_PLAN.md Step 46 (MRP Feedback Loop)
func (s *ScheduleService) getMaterialConstraints(ctx context.Context, projectID uuid.UUID) (map[uuid.UUID]time.Time, error) {
	// Query procurement_items for tasks with known delivery dates
	// Multi-tenancy enforced via project_tasks.project_id join
	query := `
		SELECT pi.project_task_id, pi.expected_delivery_date
		FROM procurement_items pi
		JOIN project_tasks pt ON pi.project_task_id = pt.id
		WHERE pt.project_id = $1 AND pi.expected_delivery_date IS NOT NULL
	`
	rows, err := s.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	constraints := make(map[uuid.UUID]time.Time)
	for rows.Next() {
		var taskID uuid.UUID
		var deliveryDate time.Time
		if err := rows.Scan(&taskID, &deliveryDate); err != nil {
			return nil, err
		}
		constraints[taskID] = deliveryDate
	}
	return constraints, rows.Err()
}
