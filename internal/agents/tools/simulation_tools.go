package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterSimulationTools registers schedule simulation tools for what-if analysis.
func RegisterSimulationTools(r *Registry, db *pgxpool.Pool, cascadeSvc service.DelayCascadeServicer) {
	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "simulate_schedule_change",
			Description: "Simulate the impact of a schedule change on the project. Shows cascading effects on downstream tasks, critical path, and project end date. Use this when users ask 'what if' questions about delays or schedule changes.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"task_name_or_wbs": {
						"type": "string",
						"description": "Task name (fuzzy match) or exact WBS code (e.g. '9.1' for Framing)"
					},
					"days": {
						"type": "integer",
						"description": "Number of days to add (positive = delay) or remove (negative = accelerate)"
					}
				},
				"required": ["task_name_or_wbs", "days"]
			}`),
		},
		Handler: func(ctx context.Context, input json.RawMessage) (string, error) {
			scope := MustGetScope(ctx)

			var params struct {
				TaskNameOrWBS string `json:"task_name_or_wbs"`
				Days          int    `json:"days"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return "", fmt.Errorf("parse input: %w", err)
			}
			if params.Days == 0 {
				return `{"error":"days must be non-zero"}`, nil
			}

			// Resolve task by name or WBS code
			taskID, taskName, err := resolveTaskByNameOrWBS(ctx, db, scope.ProjectID, params.TaskNameOrWBS)
			if err != nil {
				return "", fmt.Errorf("resolve task: %w", err)
			}

			// Run cascade simulation
			cascade, err := cascadeSvc.SimulateDelayCascade(ctx, scope.ProjectID, scope.OrgID, taskID, params.Days)
			if err != nil {
				return "", fmt.Errorf("simulate delay cascade: %w", err)
			}

			// Format result for Claude to present
			result := map[string]interface{}{
				"trigger_task":       taskName,
				"change_days":        params.Days,
				"affected_tasks":     len(cascade.AffectedTasks),
				"critical_path_shift": cascade.CriticalPathShift,
				"original_end":       cascade.OriginalEnd.Format("2006-01-02"),
				"new_projected_end":  cascade.NewProjectedEnd.Format("2006-01-02"),
			}

			// Include top affected tasks
			type affectedSummary struct {
				Name       string `json:"name"`
				WBSCode    string `json:"wbs_code"`
				SlipDays   int    `json:"slip_days"`
				NewStart   string `json:"new_start"`
				IsCritical bool   `json:"is_critical"`
			}
			limit := 10
			if len(cascade.AffectedTasks) < limit {
				limit = len(cascade.AffectedTasks)
			}
			affected := make([]affectedSummary, limit)
			for i := 0; i < limit; i++ {
				at := cascade.AffectedTasks[i]
				affected[i] = affectedSummary{
					Name:       at.TaskName,
					WBSCode:    at.WBSCode,
					SlipDays:   at.SlipDays,
					NewStart:   at.NewStart.Format("2006-01-02"),
					IsCritical: at.IsCritical,
				}
			}
			result["top_affected_tasks"] = affected

			b, _ := json.Marshal(result)
			return string(b), nil
		},
	})
}

// resolveTaskByNameOrWBS looks up a task by exact WBS code or fuzzy name match.
func resolveTaskByNameOrWBS(ctx context.Context, db *pgxpool.Pool, projectID uuid.UUID, query string) (uuid.UUID, string, error) {
	query = strings.TrimSpace(query)

	// Try exact WBS code match first
	var id uuid.UUID
	var name string
	err := db.QueryRow(ctx,
		`SELECT id, name FROM project_tasks WHERE project_id = $1 AND wbs_code = $2 LIMIT 1`,
		projectID, query,
	).Scan(&id, &name)
	if err == nil {
		return id, name, nil
	}

	// Fuzzy name match using ILIKE
	err = db.QueryRow(ctx,
		`SELECT id, name FROM project_tasks WHERE project_id = $1 AND name ILIKE '%' || $2 || '%' ORDER BY length(name) ASC LIMIT 1`,
		projectID, query,
	).Scan(&id, &name)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("no task found matching %q", query)
	}
	return id, name, nil
}
