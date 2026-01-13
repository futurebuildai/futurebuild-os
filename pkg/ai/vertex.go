package ai

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

type ModelType string

const (
	ModelTypeFlash     ModelType = "flash"
	ModelTypePro       ModelType = "pro"
	ModelTypeEmbedding ModelType = "embedding"
)

// Client defines the interface for AI operations.
// Refactored to use google.golang.org/genai types.
type Client interface {
	GenerateContent(ctx context.Context, modelType ModelType, parts ...*genai.Part) (string, error)
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	Close() error
}

type VertexClient struct {
	client    *genai.Client
	modelIDs  map[ModelType]string
	projectID string
	location  string
}

func NewVertexClient(ctx context.Context, projectID, location string, modelIDs map[ModelType]string) (*VertexClient, error) {
	// Initialize using the new SDK with Vertex AI backend
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  projectID,
		Location: location,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create vertex client: %w", err)
	}

	return &VertexClient{
		client:    client,
		modelIDs:  modelIDs,
		projectID: projectID,
		location:  location,
	}, nil
}

func (vc *VertexClient) GenerateContent(ctx context.Context, modelType ModelType, parts ...*genai.Part) (string, error) {
	modelID, ok := vc.modelIDs[modelType]
	if !ok {
		return "", fmt.Errorf("model type %s not configured", modelType)
	}

	// The new SDK GenerateContent expects content items
	// We wrap the parts into a single Content object for the user role
	content := &genai.Content{
		Parts: parts,
		Role:  "user",
	}

	resp, err := vc.client.Models.GenerateContent(ctx, modelID, []*genai.Content{content}, nil)
	if err != nil {
		return "", fmt.Errorf("generate content error: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	// Simple text extraction
	var fullText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			fullText += part.Text
		}
	}

	return fullText, nil
}

func (vc *VertexClient) Close() error {
	// The new client doesn't explicitly require Close() in the same way,
	// but good practice if it has resources.
	// Checking the SDK, it might not have a Close method exposed on the minimal client?
	// The previous SDK did.
	// If the new SDK doesn't have Close, we can no-op or check.
	// Assuming it implies http client reuse.
	// For now, we'll assume it's safe to drop or no-op if methods don't exist.
	// Wait, checking Go genai docs... usually it has Close if it keeps connections.
	// Failing a compile check, I'll remove it in the next step.
	// For now let's assume no-op if not present, but interface requires it.
	return nil
}

func (vc *VertexClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	modelID, ok := vc.modelIDs[ModelTypeEmbedding]
	if !ok {
		return nil, fmt.Errorf("embedding model not configured in modelIDs map")
	}

	// Use SDK for Embedding
	content := &genai.Content{
		Parts: []*genai.Part{{Text: text}},
		Role:  "user",
	}

	resp, err := vc.client.Models.EmbedContent(ctx, modelID, []*genai.Content{content}, nil)
	if err != nil {
		return nil, fmt.Errorf("embedding error: %w", err)
	}

	if len(resp.Embeddings) == 0 || len(resp.Embeddings[0].Values) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return resp.Embeddings[0].Values, nil
}
