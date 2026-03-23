package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
)

// RegisterScheduleTools registers all schedule-related tools.
func RegisterScheduleTools(r *Registry, svc service.ScheduleServicer) {
	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "get_project_schedule",
			Description: "Get a summary of the project schedule including end date, critical path count, total tasks, and completed tasks.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{},"required":[]}`),
		},
		Handler: func(ctx context.Context, _ json.RawMessage) (string, error) {
			scope := MustGetScope(ctx)
			summary, err := svc.GetProjectSchedule(ctx, scope.ProjectID, scope.OrgID)
			if err != nil {
				return "", fmt.Errorf("get project schedule: %w", err)
			}
			b, _ := json.Marshal(summary)
			return string(b), nil
		},
	})

	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "get_agent_focus_tasks",
			Description: "Get today's priority tasks for the project, including critical path tasks starting soon, tasks needing attention, and overdue items.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{},"required":[]}`),
		},
		Handler: func(ctx context.Context, _ json.RawMessage) (string, error) {
			scope := MustGetScope(ctx)
			tasks, err := svc.GetAgentFocusTasks(ctx, scope.ProjectID)
			if err != nil {
				return "", fmt.Errorf("get focus tasks: %w", err)
			}
			b, _ := json.Marshal(tasks)
			return string(b), nil
		},
	})

	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "get_task",
			Description: "Get detailed information about a specific task by its ID.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"task_id":{"type":"string","description":"UUID of the task"}},"required":["task_id"]}`),
		},
		Handler: func(ctx context.Context, input json.RawMessage) (string, error) {
			scope := MustGetScope(ctx)
			var params struct {
				TaskID string `json:"task_id"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return "", fmt.Errorf("parse input: %w", err)
			}
			taskID, err := uuid.Parse(params.TaskID)
			if err != nil {
				return "", fmt.Errorf("invalid task_id: %w", err)
			}
			task, err := svc.GetTask(ctx, taskID, scope.ProjectID, scope.OrgID)
			if err != nil {
				return "", fmt.Errorf("get task: %w", err)
			}
			b, _ := json.Marshal(task)
			return string(b), nil
		},
	})

	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "update_task_status",
			Description: "Update the status of a project task. Valid statuses: NotStarted, InProgress, Completed, OnHold. This modifies project state and should be used with approval.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"task_id":{"type":"string","description":"UUID of the task"},"status":{"type":"string","enum":["NotStarted","InProgress","Completed","OnHold"],"description":"New status"}},"required":["task_id","status"]}`),
		},
		Handler: func(ctx context.Context, input json.RawMessage) (string, error) {
			scope := MustGetScope(ctx)
			var params struct {
				TaskID string `json:"task_id"`
				Status string `json:"status"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return "", fmt.Errorf("parse input: %w", err)
			}
			taskID, err := uuid.Parse(params.TaskID)
			if err != nil {
				return "", fmt.Errorf("invalid task_id: %w", err)
			}
			if err := svc.UpdateTaskStatus(ctx, taskID, scope.ProjectID, scope.OrgID, types.TaskStatus(params.Status)); err != nil {
				return "", fmt.Errorf("update task status: %w", err)
			}
			return fmt.Sprintf(`{"success":true,"task_id":"%s","new_status":"%s"}`, taskID, params.Status), nil
		},
	})

	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "create_task_progress",
			Description: "Log a progress update for a task with a percentage complete and optional notes.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"task_id":{"type":"string","description":"UUID of the task"},"percent_complete":{"type":"integer","minimum":0,"maximum":100,"description":"Completion percentage"},"notes":{"type":"string","description":"Optional progress notes"}},"required":["task_id","percent_complete"]}`),
		},
		Handler: func(ctx context.Context, input json.RawMessage) (string, error) {
			scope := MustGetScope(ctx)
			var params struct {
				TaskID          string `json:"task_id"`
				PercentComplete int    `json:"percent_complete"`
				Notes           string `json:"notes"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return "", fmt.Errorf("parse input: %w", err)
			}
			taskID, err := uuid.Parse(params.TaskID)
			if err != nil {
				return "", fmt.Errorf("invalid task_id: %w", err)
			}
			if err := svc.CreateTaskProgress(ctx, scope.ProjectID, taskID, scope.UserID, params.PercentComplete, params.Notes); err != nil {
				return "", fmt.Errorf("create task progress: %w", err)
			}
			return fmt.Sprintf(`{"success":true,"task_id":"%s","percent_complete":%d}`, taskID, params.PercentComplete), nil
		},
	})

	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "recalculate_schedule",
			Description: "Trigger a full CPM (Critical Path Method) recalculation of the project schedule. Use after task status changes or duration overrides to see the updated critical path and end date.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{},"required":[]}`),
		},
		Handler: func(ctx context.Context, _ json.RawMessage) (string, error) {
			scope := MustGetScope(ctx)
			result, err := svc.RecalculateSchedule(ctx, scope.ProjectID, scope.OrgID)
			if err != nil {
				return "", fmt.Errorf("recalculate schedule: %w", err)
			}
			b, _ := json.Marshal(result)
			return string(b), nil
		},
	})
}
