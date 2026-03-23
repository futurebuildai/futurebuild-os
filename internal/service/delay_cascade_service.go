package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/physics"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DelayCascade represents the full impact analysis of a task slip.
type DelayCascade struct {
	TriggerTaskID     uuid.UUID      `json:"trigger_task_id"`
	TriggerTaskName   string         `json:"trigger_task_name"`
	SlipDays          int            `json:"slip_days"`
	AffectedTasks     []AffectedTask `json:"affected_tasks"`
	NewProjectedEnd   time.Time      `json:"new_projected_end"`
	OriginalEnd       time.Time      `json:"original_end"`
	CriticalPathShift int            `json:"critical_path_shift"`
}

// AffectedTask represents a single task affected by the delay cascade.
type AffectedTask struct {
	TaskID     uuid.UUID `json:"task_id"`
	TaskName   string    `json:"task_name"`
	WBSCode    string    `json:"wbs_code"`
	OldStart   time.Time `json:"old_start"`
	NewStart   time.Time `json:"new_start"`
	SlipDays   int       `json:"slip_days"`
	IsCritical bool      `json:"is_critical"`
}

// DelayCascadeService simulates the cascading impact of a task slip on the project schedule.
type DelayCascadeService struct {
	db         *pgxpool.Pool
	feedWriter FeedWriter
}

// FeedWriter is satisfied by *FeedService — allows writing feed cards.
type FeedWriter interface {
	WriteCard(ctx context.Context, card *models.FeedCard) error
}

// NewDelayCascadeService creates a new delay cascade analysis service.
func NewDelayCascadeService(db *pgxpool.Pool, feedWriter FeedWriter) *DelayCascadeService {
	return &DelayCascadeService{db: db, feedWriter: feedWriter}
}

// SimulateDelayCascade calculates the cascading impact of a task slipping by the given days.
// It loads the project schedule, clones it, applies the slip, re-runs CPM, and diffs results.
func (s *DelayCascadeService) SimulateDelayCascade(
	ctx context.Context,
	projectID, orgID, taskID uuid.UUID,
	slipDays int,
) (*DelayCascade, error) {
	// 1. Load project tasks and dependencies
	tasks, err := s.loadProjectTasks(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("load tasks: %w", err)
	}

	deps, err := s.loadTaskDependencies(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("load dependencies: %w", err)
	}

	// Find the start date
	startDate := time.Now()
	for _, t := range tasks {
		if t.PlannedStart != nil && t.PlannedStart.Before(startDate) {
			startDate = *t.PlannedStart
		}
	}

	// 2. Run original CPM
	cal := &physics.StandardCalendar{}
	origGraph := physics.BuildDependencyGraph(tasks, deps)
	origSchedule, err := physics.ForwardPass(origGraph, startDate, cal, nil)
	if err != nil {
		return nil, fmt.Errorf("original forward pass: %w", err)
	}
	origCriticalPath, err := physics.BackwardPass(origGraph, origSchedule, cal, nil)
	if err != nil {
		return nil, fmt.Errorf("original backward pass: %w", err)
	}

	// Find original project end
	var origEnd time.Time
	for _, sched := range origSchedule {
		if sched.EarlyFinish.After(origEnd) {
			origEnd = sched.EarlyFinish
		}
	}

	// 3. Clone tasks, add slip days to target task
	modifiedTasks := make([]models.ProjectTask, len(tasks))
	copy(modifiedTasks, tasks)

	var triggerName string
	for i, t := range modifiedTasks {
		if t.ID == taskID {
			triggerName = t.Name
			modifiedTasks[i].CalculatedDuration += float64(slipDays)
			break
		}
	}
	if triggerName == "" {
		return nil, fmt.Errorf("task %s not found in project", taskID)
	}

	// 4. Run modified CPM
	modGraph := physics.BuildDependencyGraph(modifiedTasks, deps)
	modSchedule, err := physics.ForwardPass(modGraph, startDate, cal, nil)
	if err != nil {
		return nil, fmt.Errorf("modified forward pass: %w", err)
	}
	modCriticalPath, err := physics.BackwardPass(modGraph, modSchedule, cal, nil)
	if err != nil {
		return nil, fmt.Errorf("modified backward pass: %w", err)
	}

	// Find modified project end
	var modEnd time.Time
	for _, sched := range modSchedule {
		if sched.EarlyFinish.After(modEnd) {
			modEnd = sched.EarlyFinish
		}
	}

	// 5. Diff: identify affected tasks
	critPathSet := make(map[string]bool, len(modCriticalPath))
	for _, cp := range modCriticalPath {
		critPathSet[cp] = true
	}

	var affected []AffectedTask
	for _, t := range tasks {
		origSched, origOK := origSchedule[t.ID]
		modSched, modOK := modSchedule[t.ID]
		if !origOK || !modOK {
			continue
		}

		slipAmount := int(modSched.EarlyStart.Sub(origSched.EarlyStart).Hours() / 24)
		if slipAmount > 0 {
			affected = append(affected, AffectedTask{
				TaskID:     t.ID,
				TaskName:   t.Name,
				WBSCode:    t.WBSCode,
				OldStart:   origSched.EarlyStart,
				NewStart:   modSched.EarlyStart,
				SlipDays:   slipAmount,
				IsCritical: critPathSet[t.WBSCode],
			})
		}
	}

	critShift := int(modEnd.Sub(origEnd).Hours() / 24)

	cascade := &DelayCascade{
		TriggerTaskID:     taskID,
		TriggerTaskName:   triggerName,
		SlipDays:          slipDays,
		AffectedTasks:     affected,
		NewProjectedEnd:   modEnd,
		OriginalEnd:       origEnd,
		CriticalPathShift: critShift,
	}

	// 6. Write feed card if there's significant impact
	if s.feedWriter != nil && (critShift > 0 || len(affected) > 3) {
		s.writeDelayCascadeCard(ctx, projectID, orgID, cascade, origCriticalPath)
	}

	return cascade, nil
}

// writeDelayCascadeCard creates a feed card showing the cascade impact.
func (s *DelayCascadeService) writeDelayCascadeCard(ctx context.Context, projectID, orgID uuid.UUID, cascade *DelayCascade, origCP []string) {
	agentSource := "DelayCascadeService"
	headline := fmt.Sprintf("%s slipped %dd — %d downstream tasks affected",
		cascade.TriggerTaskName, cascade.SlipDays, len(cascade.AffectedTasks))

	body := fmt.Sprintf("A %d-day delay on \"%s\" cascades to %d downstream tasks.",
		cascade.SlipDays, cascade.TriggerTaskName, len(cascade.AffectedTasks))

	if cascade.CriticalPathShift > 0 {
		body += fmt.Sprintf("\n\nProject end date shifts from %s to %s (+%d days).",
			cascade.OriginalEnd.Format("Jan 02"), cascade.NewProjectedEnd.Format("Jan 02"), cascade.CriticalPathShift)
	}

	// Show top 5 affected tasks
	limit := 5
	if len(cascade.AffectedTasks) < limit {
		limit = len(cascade.AffectedTasks)
	}
	for i := 0; i < limit; i++ {
		at := cascade.AffectedTasks[i]
		critical := ""
		if at.IsCritical {
			critical = " [CRITICAL]"
		}
		body += fmt.Sprintf("\n- %s%s: +%d days (new start: %s)",
			at.TaskName, critical, at.SlipDays, at.NewStart.Format("Jan 02"))
	}

	priority := models.FeedCardPriorityNormal
	if cascade.CriticalPathShift > 0 {
		priority = models.FeedCardPriorityUrgent
	}

	var consequence *string
	if cascade.CriticalPathShift > 0 {
		c := fmt.Sprintf("Project end date moves %d days later to %s",
			cascade.CriticalPathShift, cascade.NewProjectedEnd.Format("Jan 02, 2006"))
		consequence = &c
	}

	card := &models.FeedCard{
		OrgID:       orgID,
		ProjectID:   projectID,
		CardType:    models.FeedCardScheduleRecalc,
		Priority:    priority,
		Headline:    headline,
		Body:        body,
		Consequence: consequence,
		Horizon:     models.FeedCardHorizonThisWeek,
		AgentSource: &agentSource,
		Actions: []models.FeedCardAction{
			{ID: "view_schedule", Label: "View Schedule", Style: "primary"},
			{ID: "dismiss", Label: "Dismiss", Style: "secondary"},
		},
	}

	if err := s.feedWriter.WriteCard(ctx, card); err != nil {
		slog.Error("failed to write delay cascade feed card", "project_id", projectID, "error", err)
	}
}

// loadProjectTasks loads all tasks for a project.
func (s *DelayCascadeService) loadProjectTasks(ctx context.Context, projectID uuid.UUID) ([]models.ProjectTask, error) {
	query := `
		SELECT id, project_id, wbs_code, name, status, calculated_duration_days,
			planned_start, planned_end, is_on_critical_path, total_float_days, is_inspection
		FROM project_tasks
		WHERE project_id = $1
		ORDER BY wbs_code
	`
	rows, err := s.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.ProjectTask
	for rows.Next() {
		var t models.ProjectTask
		if err := rows.Scan(
			&t.ID, &t.ProjectID, &t.WBSCode, &t.Name, &t.Status,
			&t.CalculatedDuration, &t.PlannedStart, &t.PlannedEnd,
			&t.IsOnCriticalPath, &t.TotalFloatDays, &t.IsInspection,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// loadTaskDependencies loads all dependencies for a project.
func (s *DelayCascadeService) loadTaskDependencies(ctx context.Context, projectID uuid.UUID) ([]models.TaskDependency, error) {
	query := `
		SELECT id, project_id, predecessor_id, successor_id, dependency_type, lag_days
		FROM task_dependencies
		WHERE project_id = $1
	`
	rows, err := s.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []models.TaskDependency
	for rows.Next() {
		var d models.TaskDependency
		if err := rows.Scan(&d.ID, &d.ProjectID, &d.PredecessorID, &d.SuccessorID, &d.DependencyType, &d.LagDays); err != nil {
			return nil, err
		}
		deps = append(deps, d)
	}
	return deps, nil
}
