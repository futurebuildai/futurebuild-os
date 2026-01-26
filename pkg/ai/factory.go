package ai

import (
	"context"
	"fmt"
	"os"
)

// Factory creates AI clients.
type Factory struct {
	vertexProjectID string
	vertexLocation  string
	anthropicKey    string
}

// NewFactory returns a new AI Client Factory.
func NewFactory(vertexProjectID, vertexLocation, anthropicKey string) *Factory {
	return &Factory{
		vertexProjectID: vertexProjectID,
		vertexLocation:  vertexLocation,
		anthropicKey:    anthropicKey,
	}
}

// NewClient creates a specific client for a provider.
func (f *Factory) NewClient(ctx context.Context, provider Provider) (Client, error) {
	switch provider {
	case ProviderVertex:
		// Default Vertex Model Map
		// See IMPLEMENTATION_PLAN.md for model IDs
		modelMap := map[ModelType]string{
			ModelTypeFlashPreview: "gemini-3.0-flash-preview", // Coordinator
			ModelTypeCodeAssist:   "gemini-1.5-pro",           // Historian (Proxy for Code Assist API)
			ModelTypeEmbedding:    "text-embedding-004",
		}
		// Allow override via env vars if needed
		return NewVertexClient(ctx, f.vertexProjectID, f.vertexLocation, modelMap)

	case ProviderAnthropic:
		if f.anthropicKey == "" {
			return nil, fmt.Errorf("anthropic api key not configured")
		}
		modelMap := map[ModelType]string{
			ModelTypeOpus: "claude-3-opus-20240229", // Fallback if 4.5 not available
		}

		// P2 Fix: Prioritize 4.5 if env var is set, or if we want to default to a placeholder
		// Ideally we use a real ID once announced.
		if v := os.Getenv("CLAUDE_OPUS_MODEL_ID"); v != "" {
			modelMap[ModelTypeOpus] = v
		}

		return NewAnthropicClient(f.anthropicKey, modelMap), nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
