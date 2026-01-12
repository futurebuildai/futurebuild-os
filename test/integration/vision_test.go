package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloud.google.com/go/vertexai/genai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/ai"
)

// MockVertexClient simulates Vertex AI interactions
type MockVertexClient struct {
	ShouldFail bool
}

func (m *MockVertexClient) GenerateContent(ctx context.Context, modelType ai.ModelType, parts ...genai.Part) (string, error) {
	if m.ShouldFail {
		return "", fmt.Errorf("mock ai failure")
	}
	// Return a valid JSON response simulating Gemini
	return `{"is_verified": true, "confidence": 0.95, "reasoning": "Mock verification passed"}`, nil
}

func (m *MockVertexClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}

func (m *MockVertexClient) Close() error {
	return nil
}

func TestVisionService_VerifyTask(t *testing.T) {
	// 1. Setup Client (Real or Mock)
	var client ai.Client
	cfg := config.LoadConfig()

	if cfg.VertexProjectID != "" {
		// Try to use Real Client
		ctx := context.Background()
		models := map[ai.ModelType]string{
			ai.ModelTypeFlash:     cfg.VertexModelFlashID,
			ai.ModelTypeEmbedding: cfg.VertexModelEmbeddingID,
		}
		var err error
		client, err = ai.NewVertexClient(ctx, cfg.VertexProjectID, cfg.VertexLocation, models)
		if err != nil {
			t.Logf("Failed to create real Vertex client: %v. Falling back to Mock.", err)
			client = &MockVertexClient{}
		} else {
			t.Log("Using Real Vertex AI Client")
		}
	} else {
		t.Log("VERTEX_PROJECT_ID not set. Using Mock Vertex AI Client.")
		client = &MockVertexClient{}
	}
	defer client.Close()

	visionService := service.NewVisionService(client)
	ctx := context.Background()

	// Create a mock server to serve an image
	dummyPNG := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
		0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41, 0x54, 0x08, 0xD7, 0x63, 0x60, 0x00, 0x02, 0x00,
		0x00, 0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44,
		0xAE, 0x42, 0x60, 0x82,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(dummyPNG)
	}))
	defer ts.Close()

	t.Run("Verify_Task_Completion", func(t *testing.T) {
		isVerified, confidence, err := visionService.VerifyTask(ctx, ts.URL, "A test description")
		require.NoError(t, err)
		t.Logf("Vision Result - Verified: %v, Confidence: %f", isVerified, confidence)

		// Assertions (relaxed for Real AI, strict for Mock)
		if _, ok := client.(*MockVertexClient); ok {
			assert.True(t, isVerified)
			assert.Equal(t, 0.95, confidence)
		} else {
			assert.GreaterOrEqual(t, confidence, 0.0)
			assert.LessOrEqual(t, confidence, 1.0)
		}
	})

	t.Run("Invalid_Image_URL", func(t *testing.T) {
		// Use a local port that is guaranteed to be closed to force connection error
		// Port 0 is not valid, but http client might just fail to dial
		// "http://127.0.0.1:0" is a good cross-platform way to trigger dial error
		_, _, err := visionService.VerifyTask(ctx, "http://127.0.0.1:0/img.png", "A test description")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to download image")
	})
}
