package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/ai"
)

// RegisterOnboardingTools adds onboarding-specific tools to the registry.
// These tools enable Claude to orchestrate the onboarding flow:
// - generate_schedule_preview: instant CPM schedule from project attributes
// - compare_scenarios: what-if analysis across multiple configurations
// - set_project_progress: mark phases as completed for in-progress projects
func RegisterOnboardingTools(
	r *Registry,
	previewSvc *service.SchedulePreviewService,
) {
	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "generate_schedule_preview",
			Description: "Generate an instant construction schedule preview from project attributes. Returns projected end date, critical path, phase timeline, and Gantt data. Use this after collecting square footage, foundation type, and start date.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"square_footage": {"type": "number", "description": "Gross square footage of the building"},
					"foundation_type": {"type": "string", "enum": ["slab", "crawlspace", "basement"]},
					"start_date": {"type": "string", "description": "Project start date in YYYY-MM-DD format"},
					"stories": {"type": "integer", "description": "Number of stories (1, 2, or 3+)"},
					"address": {"type": "string", "description": "Project street address"},
					"topography": {"type": "string", "enum": ["flat", "sloped", "hillside"]},
					"soil_conditions": {"type": "string", "enum": ["normal", "rocky", "clay", "sandy"]},
					"bedrooms": {"type": "integer"},
					"bathrooms": {"type": "integer"},
					"is_in_progress": {"type": "boolean", "description": "True if the project is already under construction"},
					"completed_phases": {
						"type": "array",
						"items": {
							"type": "object",
							"properties": {
								"wbs_code": {"type": "string"},
								"actual_end": {"type": "string"},
								"status": {"type": "string"}
							}
						}
					},
					"current_date": {"type": "string", "description": "Current date for in-progress projects (YYYY-MM-DD)"}
				},
				"required": ["square_footage", "foundation_type", "start_date", "stories"]
			}`),
		},
		Handler: func(ctx context.Context, input json.RawMessage) (string, error) {
			var req service.SchedulePreviewRequest
			if err := json.Unmarshal(input, &req); err != nil {
				return "", fmt.Errorf("invalid input: %w", err)
			}
			preview, err := previewSvc.GeneratePreview(req)
			if err != nil {
				return "", err
			}
			result, err := json.Marshal(preview)
			if err != nil {
				return "", err
			}
			return string(result), nil
		},
	})

	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "compare_scenarios",
			Description: "Compare multiple schedule scenarios side by side (what-if analysis). Provide a base scenario and up to 3 alternatives. Returns projected end dates and delta days for each.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"base": {
						"type": "object",
						"description": "The base scenario to compare against",
						"properties": {
							"square_footage": {"type": "number"},
							"foundation_type": {"type": "string"},
							"start_date": {"type": "string"},
							"stories": {"type": "integer"}
						},
						"required": ["square_footage", "foundation_type", "start_date", "stories"]
					},
					"alternatives": {
						"type": "array",
						"maxItems": 3,
						"items": {
							"type": "object",
							"properties": {
								"square_footage": {"type": "number"},
								"foundation_type": {"type": "string"},
								"start_date": {"type": "string"},
								"stories": {"type": "integer"}
							},
							"required": ["square_footage", "foundation_type", "start_date", "stories"]
						}
					}
				},
				"required": ["base", "alternatives"]
			}`),
		},
		Handler: func(ctx context.Context, input json.RawMessage) (string, error) {
			var req service.ScenarioComparisonRequest
			if err := json.Unmarshal(input, &req); err != nil {
				return "", fmt.Errorf("invalid input: %w", err)
			}
			result, err := previewSvc.CompareScenarios(req)
			if err != nil {
				return "", err
			}
			out, err := json.Marshal(result)
			if err != nil {
				return "", err
			}
			return string(out), nil
		},
	})

	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "set_project_progress",
			Description: "For projects already under construction: mark specific phases or tasks as completed with their actual completion dates. Call this when the user says something like 'foundation is done' or 'we finished framing last week'. Use WBS phase codes like '7.x' for Site Prep, '8.x' for Foundation, '9.x' for Framing, '10.x' for Rough-Ins.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"completed_phases": {
						"type": "array",
						"items": {
							"type": "object",
							"properties": {
								"wbs_code": {"type": "string", "description": "WBS code (e.g. '8.x' for all foundation tasks, or '8.0' for a specific task)"},
								"actual_end": {"type": "string", "description": "Actual completion date in YYYY-MM-DD format"},
								"status": {"type": "string", "enum": ["completed", "in_progress"], "description": "Whether the phase is fully completed or still in progress"}
							},
							"required": ["wbs_code", "actual_end", "status"]
						}
					}
				},
				"required": ["completed_phases"]
			}`),
		},
		Handler: func(ctx context.Context, input json.RawMessage) (string, error) {
			var req struct {
				CompletedPhases []service.CompletedPhaseInput `json:"completed_phases"`
			}
			if err := json.Unmarshal(input, &req); err != nil {
				return "", fmt.Errorf("invalid input: %w", err)
			}
			result, err := json.Marshal(map[string]interface{}{
				"status":           "progress_recorded",
				"phases_count":     len(req.CompletedPhases),
				"completed_phases": req.CompletedPhases,
				"message":          fmt.Sprintf("Recorded %d completed phases. Call generate_schedule_preview with is_in_progress=true and these completed_phases to see the remaining schedule.", len(req.CompletedPhases)),
			})
			if err != nil {
				return "", err
			}
			return string(result), nil
		},
	})
}
