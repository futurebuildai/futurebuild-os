package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/ai"
)

// RegisterProjectTools registers project and procurement related tools.
func RegisterProjectTools(r *Registry, projectSvc service.ProjectServicer, weatherSvc service.WeatherServicer) {
	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "get_project",
			Description: "Get project details including name, address, status, budget, and key dates.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{},"required":[]}`),
		},
		Handler: func(ctx context.Context, _ json.RawMessage) (string, error) {
			scope := MustGetScope(ctx)
			project, err := projectSvc.GetProject(ctx, scope.ProjectID, scope.OrgID)
			if err != nil {
				return "", fmt.Errorf("get project: %w", err)
			}
			b, _ := json.Marshal(project)
			return string(b), nil
		},
	})

	r.Register(Tool{
		Definition: ai.ToolDefinition{
			Name:        "list_procurement_items",
			Description: "List all procurement (long-lead) items for the project with their alert status, lead times, and calculated order dates. Useful for understanding supply chain risks.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{},"required":[]}`),
		},
		Handler: func(ctx context.Context, _ json.RawMessage) (string, error) {
			scope := MustGetScope(ctx)
			items, err := projectSvc.ListProcurementItems(ctx, scope.ProjectID, scope.OrgID)
			if err != nil {
				return "", fmt.Errorf("list procurement items: %w", err)
			}
			b, _ := json.Marshal(items)
			return string(b), nil
		},
	})

	if weatherSvc != nil {
		r.Register(Tool{
			Definition: ai.ToolDefinition{
				Name:        "get_weather_forecast",
				Description: "Get the weather forecast for the project location. Returns high/low temp, precipitation probability, and conditions. Critical for scheduling exterior work.",
				InputSchema: json.RawMessage(`{"type":"object","properties":{"latitude":{"type":"number","description":"Latitude of the project site"},"longitude":{"type":"number","description":"Longitude of the project site"}},"required":["latitude","longitude"]}`),
			},
			Handler: func(ctx context.Context, input json.RawMessage) (string, error) {
				var params struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				}
				if err := json.Unmarshal(input, &params); err != nil {
					return "", fmt.Errorf("parse input: %w", err)
				}
				forecast, err := weatherSvc.GetForecast(params.Latitude, params.Longitude)
				if err != nil {
					return "", fmt.Errorf("get forecast: %w", err)
				}
				b, _ := json.Marshal(forecast)
				return string(b), nil
			},
		})
	}
}
