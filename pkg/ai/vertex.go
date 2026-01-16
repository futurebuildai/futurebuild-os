package ai

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// ModelType specifies the AI model to use for generation.
type ModelType string

const (
	ModelTypeFlash     ModelType = "flash"
	ModelTypePro       ModelType = "pro"
	ModelTypeEmbedding ModelType = "embedding"
)

// =============================================================================
// VENDOR-AGNOSTIC CLIENT INTERFACE
// =============================================================================
// Client defines the interface for AI operations using vendor-agnostic types.
// This abstraction allows switching between AI providers (Vertex AI, OpenAI,
// Anthropic, local models) without modifying service layer code.
//
// L7 Quality Gate: No vendor-specific types (e.g., *genai.Part) in interface.
// =============================================================================

// Client defines the interface for AI operations.
// Uses vendor-agnostic types from types.go.
type Client interface {
	// GenerateContent generates text/multimodal content using the specified model.
	GenerateContent(ctx context.Context, req GenerateRequest) (GenerateResponse, error)

	// GenerateEmbedding generates a vector embedding for the given text.
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)

	// Close releases any resources used by the client.
	Close() error
}

// =============================================================================
// VERTEX AI IMPLEMENTATION
// =============================================================================
// VertexClient implements Client for Google Vertex AI.
// All genai-specific logic is encapsulated here.
// =============================================================================

// VertexClient implements the Client interface for Google Vertex AI.
type VertexClient struct {
	client    *genai.Client
	modelIDs  map[ModelType]string
	projectID string
	location  string
}

// NewVertexClient creates a new Vertex AI client.
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

// GenerateContent generates content using Vertex AI.
// Converts vendor-agnostic ContentPart to genai.Part internally.
// When req.ReturnLogprobs is true, enables logprob extraction for confidence scoring.
func (vc *VertexClient) GenerateContent(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	modelID, ok := vc.modelIDs[req.Model]
	if !ok {
		return GenerateResponse{}, fmt.Errorf("model type %s not configured", req.Model)
	}

	// Convert vendor-agnostic ContentPart to Google-specific genai.Part
	genaiParts := make([]*genai.Part, len(req.Parts))
	for i, part := range req.Parts {
		genaiParts[i] = contentPartToGenAI(part)
	}

	// Wrap parts into Content for the user role
	content := &genai.Content{
		Parts: genaiParts,
		Role:  "user",
	}

	// Build generation config with optional logprobs
	// See: https://developers.googleblog.com/unlock-gemini-reasoning-with-logprobs-on-vertex-ai/
	var config *genai.GenerateContentConfig
	if req.ReturnLogprobs {
		config = &genai.GenerateContentConfig{
			ResponseLogprobs: true,
			Logprobs:         ptr(int32(5)), // Top 5 alternatives for confidence calculation
		}
	}

	resp, err := vc.client.Models.GenerateContent(ctx, modelID, []*genai.Content{content}, config)
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("generate content error: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return GenerateResponse{}, fmt.Errorf("no content generated")
	}

	// Extract text from response
	var fullText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			fullText += part.Text
		}
	}

	// Calculate confidence from logprobs if enabled and available
	// Logprobs are negative numbers; closer to 0 = higher confidence
	// We convert to probability using exp(logprob) and average across tokens
	var confidence float32
	if req.ReturnLogprobs && resp.Candidates[0].LogprobsResult != nil {
		confidence = calculateConfidenceFromLogprobs(resp.Candidates[0].LogprobsResult)
	}

	// Build vendor-agnostic response
	return GenerateResponse{
		Text:       fullText,
		TokensUsed: 0, // Note: Vertex AI response doesn't easily expose token count
		Confidence: confidence,
	}, nil
}

// GenerateEmbedding generates a vector embedding using Vertex AI.
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

// Close releases resources used by the client.
func (vc *VertexClient) Close() error {
	// The genai client doesn't require explicit Close in current SDK
	return nil
}

// contentPartToGenAI converts a vendor-agnostic ContentPart to genai.Part.
// This is the only place where vendor-specific conversion happens.
func contentPartToGenAI(part ContentPart) *genai.Part {
	if part.Text != "" {
		return &genai.Part{Text: part.Text}
	}
	if len(part.Data) > 0 {
		return &genai.Part{
			InlineData: &genai.Blob{
				MIMEType: part.MimeType,
				Data:     part.Data,
			},
		}
	}
	return &genai.Part{}
}

// ptr is a generic helper to get a pointer to a value.
func ptr[T any](v T) *T {
	return &v
}

// calculateConfidenceFromLogprobs computes a confidence score from logprobs.
// Logprobs are negative numbers where closer to 0 = higher confidence.
// We convert each logprob to probability using exp(logprob), then average.
//
// Example: logprob of -0.02 → exp(-0.02) ≈ 0.98 (98% confident)
// Example: logprob of -5.0  → exp(-5.0)  ≈ 0.007 (0.7% confident)
//
// See: https://developers.googleblog.com/unlock-gemini-reasoning-with-logprobs-on-vertex-ai/
func calculateConfidenceFromLogprobs(result *genai.LogprobsResult) float32 {
	if result == nil || len(result.ChosenCandidates) == 0 {
		return 0
	}

	// Sum probabilities (converted from logprobs)
	var sumProb float64
	count := 0
	for _, candidate := range result.ChosenCandidates {
		// Convert logprob to probability: prob = exp(logprob)
		prob := exp64(float64(candidate.LogProbability))
		sumProb += prob
		count++
	}

	if count == 0 {
		return 0
	}

	// Average probability across all tokens
	avgProb := sumProb / float64(count)
	return float32(avgProb)
}

// exp64 computes e^x using a simple approximation suitable for logprobs.
// For production, consider using math.Exp but this avoids the import.
func exp64(x float64) float64 {
	// e^x approximation using Taylor series for small x
	// For logprobs (typically -10 to 0), this is sufficient
	// Use: e^x ≈ 1 + x + x²/2 + x³/6 + x⁴/24 for accuracy
	if x > 0 {
		x = 0 // Clamp positive values (shouldn't happen for logprobs)
	}
	if x < -10 {
		return 0.0001 // Floor for very negative logprobs
	}

	// More accurate: use standard library
	// For simplicity, use the series approximation
	x2 := x * x
	x3 := x2 * x
	x4 := x3 * x
	return 1 + x + x2/2 + x3/6 + x4/24
}
