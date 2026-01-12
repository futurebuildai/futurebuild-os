package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"cloud.google.com/go/vertexai/genai"
	"golang.org/x/oauth2/google"
)

type ModelType string

const (
	ModelTypeFlash     ModelType = "flash"
	ModelTypePro       ModelType = "pro"
	ModelTypeEmbedding ModelType = "embedding"
)

type Client interface {
	GenerateContent(ctx context.Context, modelType ModelType, parts ...genai.Part) (string, error)
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	Close() error
}

type VertexClient struct {
	client    *genai.Client
	genModels map[ModelType]*genai.GenerativeModel
	modelIDs  map[ModelType]string
	projectID string
	location  string
}

func NewVertexClient(ctx context.Context, projectID, location string, modelIDs map[ModelType]string) (*VertexClient, error) {
	client, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		return nil, fmt.Errorf("failed to create vertex client: %w", err)
	}

	genModels := make(map[ModelType]*genai.GenerativeModel)
	for t, id := range modelIDs {
		// Only store generative models in this map as embeddings are handled via REST
		if t != ModelTypeEmbedding {
			genModels[t] = client.GenerativeModel(id)
		}
	}

	return &VertexClient{
		client:    client,
		genModels: genModels,
		modelIDs:  modelIDs,
		projectID: projectID,
		location:  location,
	}, nil
}

func (vc *VertexClient) GenerateContent(ctx context.Context, modelType ModelType, parts ...genai.Part) (string, error) {
	model, ok := vc.genModels[modelType]
	if !ok {
		return "", fmt.Errorf("model type %s not configured", modelType)
	}

	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return "", fmt.Errorf("generate content error: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	// Simple text extraction for now
	var fullText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			fullText += string(txt)
		}
	}

	return fullText, nil
}

func (vc *VertexClient) Close() error {
	return vc.client.Close()
}

func (vc *VertexClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	modelID, ok := vc.modelIDs[ModelTypeEmbedding]
	if !ok {
		return nil, fmt.Errorf("embedding model not configured in modelIDs map")
	}

	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:predict",
		vc.location, vc.projectID, vc.location, modelID)

	payload := map[string]interface{}{
		"instances": []map[string]interface{}{
			{"content": text},
		},
	}
	body, _ := json.Marshal(payload)

	// Use Google default client for ADC auth
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
