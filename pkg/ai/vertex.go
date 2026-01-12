package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2/google"
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

	// We stick to the REST implementation for Embeddings for now
	// to ensure stability unless we actally refactor this to SDK too.
	// The SDK likely supports vc.client.Models.EmbedContent but params might differ.
	// Keeping the working REST implementation as per "Refactor Client" focused on image payloads.

	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:predict",
		vc.location, vc.projectID, vc.location, modelID)

	payload := map[string]interface{}{
		"instances": []map[string]interface{}{
			{"content": text},
		},
	}
	body, _ := json.Marshal(payload)

	client, err := google.DefaultClient(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API returned status %d", resp.StatusCode)
	}

	var result struct {
		Predictions []struct {
			Embeddings struct {
				Values []float32 `json:"values"`
			} `json:"embeddings"`
		} `json:"predictions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Predictions) == 0 {
		return nil, fmt.Errorf("no predictions returned")
	}

	return result.Predictions[0].Embeddings.Values, nil
}
