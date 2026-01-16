// Package mocks provides test doubles for time-travel simulation.
// See PRODUCTION_PLAN.md Step 49 Amendment 3
package mocks

import (
	"context"

	"github.com/colton/futurebuild/pkg/ai"
)

// MockAIClient returns static content to avoid real AI calls.
// This speeds up simulation from ~2s/iteration to <1ms.
// L7 Vendor Abstraction: Updated to use vendor-agnostic types.
type MockAIClient struct{}

// GenerateContent returns a static mock response.
func (m *MockAIClient) GenerateContent(ctx context.Context, req ai.GenerateRequest) (ai.GenerateResponse, error) {
	return ai.GenerateResponse{
		Text:       "Mock AI Response - Simulation Mode",
		TokensUsed: 100,
		Confidence: 0.95,
	}, nil
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
