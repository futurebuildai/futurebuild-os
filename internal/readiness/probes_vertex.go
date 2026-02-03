package readiness

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/genai"
)

// VertexAIProbe verifies Vertex AI connectivity by listing models (page_size=1).
type VertexAIProbe struct {
	projectID string
	location  string
}

// NewVertexAIProbe creates a probe that creates a short-lived genai client
// and lists one model to verify credentials and project access.
func NewVertexAIProbe(projectID, location string) *VertexAIProbe {
	return &VertexAIProbe{projectID: projectID, location: location}
}

func (p *VertexAIProbe) Name() string { return "vertex_ai" }

func (p *VertexAIProbe) Check(ctx context.Context) CheckResult {
	start := time.Now()

	if p.projectID == "" || p.location == "" {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusNotConfigured,
			Message:  "VERTEX_PROJECT_ID or VERTEX_LOCATION not set",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  p.projectID,
		Location: p.location,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  fmt.Sprintf("client creation failed: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	// List models with page_size=1 as a minimal connectivity check.
	page, err := client.Models.List(ctx, &genai.ListModelsConfig{PageSize: 1})
	if err != nil {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  fmt.Sprintf("models.list failed: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	_ = page

	return CheckResult{
		Name:     p.Name(),
		Status:   StatusHealthy,
		Duration: time.Since(start).Milliseconds(),
	}
}
