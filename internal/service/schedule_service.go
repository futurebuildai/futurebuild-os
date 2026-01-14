// Package service provides business logic for the FutureBuild application.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/physics"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
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

	// Step 4: Build dependency graph and run CPM
	// See BACKEND_SCOPE.md Section 6.3
	graph := physics.BuildDependencyGraph(tasks, deps)

	// Detect cycles before processing
	if err := physics.DetectCycle(graph); err != nil {
		return nil, fmt.Errorf("schedule contains circular dependencies: %w", err)
	}

	schedule, err := physics.ForwardPass(graph, projectStart)
	if err != nil {
		return nil, fmt.Errorf("forward pass failed: %w", err)
	}

	criticalPath, err := physics.BackwardPass(graph, schedule)
	if err != nil {
		return nil, fmt.Errorf("backward pass failed: %w", err)
	}

	// Step 5: Batch-update all task schedules in a single transaction
	// See PRODUCTION_PLAN.md Step 32 (Optimization: transaction-based batch update)
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var projectEnd time.Time
	for taskID, sched := range schedule {
		_, err := tx.Exec(ctx, `
			UPDATE project_tasks 
			SET early_start = $1, early_finish = $2, 
			    late_start = $3, late_finish = $4,
			    total_float_days = $5, is_on_critical_path = $6,
			    updated_at = NOW()
			WHERE id = $7 AND project_id = $8
		`, sched.EarlyStart, sched.EarlyFinish,
			sched.LateStart, sched.LateFinish,
			sched.TotalFloat, sched.IsCritical,
			taskID, projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to update task %s: %w", taskID, err)
		}

		// Track the latest finish date for project end
		if sched.EarlyFinish.After(projectEnd) {
			projectEnd = sched.EarlyFinish
		}
	}

	// Step 6: Update project's target_end_date
	_, err = tx.Exec(ctx, `
		UPDATE projects SET target_end_date = $1, updated_at = NOW() WHERE id = $2
	`, projectEnd, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to update project end date: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
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
func (s *ScheduleService) UpdateTaskStatus(ctx context.Context, taskID, projectID, orgID uuid.UUID, status types.TaskStatus) error {
	// Harden multi-tenancy by joining with projects
	_, err := s.db.Exec(ctx, `
		UPDATE project_tasks pt
		SET status = $1, updated_at = NOW()
		FROM projects p
		WHERE pt.project_id = p.id AND pt.id = $2 AND pt.project_id = $3 AND p.org_id = $4
	`, status, taskID, projectID, orgID)
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
