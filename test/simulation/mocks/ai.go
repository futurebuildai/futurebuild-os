// Package mocks provides test doubles for time-travel simulation.
// See PRODUCTION_PLAN.md Step 49 Amendment 3
package mocks

import (
	"context"

	"github.com/colton/futurebuild/pkg/ai"
	"google.golang.org/genai"
)

// MockAIClient returns static content to avoid real AI calls.
// This speeds up simulation from ~2s/iteration to <1ms.
type MockAIClient struct{}

// GenerateContent returns a static mock briefing.
func (m *MockAIClient) GenerateContent(ctx context.Context, modelType ai.ModelType, parts ...*genai.Part) (string, error) {
	return "Mock Daily Briefing Content - Simulation Mode", nil
}

// GenerateEmbedding returns a mock embedding vector.
func (m *MockAIClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Return a dummy 768-dimensional embedding
	embedding := make([]float32, 768)
	return embedding, nil
}

// Close is a no-op for the mock.
func (m *MockAIClient) Close() error {
	return nil
}
