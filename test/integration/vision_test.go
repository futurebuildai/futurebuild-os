package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/ai"
)

func TestVisionService_VerifyTask(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	cfg := config.LoadConfig()
	if cfg.VertexProjectID == "" {
		t.Skip("Skipping Vision test: VERTEX_PROJECT_ID not set")
	}

	ctx := context.Background()
	models := map[ai.ModelType]string{
		ai.ModelTypeFlash:     cfg.VertexModelFlashID,
		ai.ModelTypeEmbedding: cfg.VertexModelEmbeddingID,
	}

	vertexClient, err := ai.NewVertexClient(ctx, cfg.VertexProjectID, cfg.VertexLocation, models)
	if err != nil {
		t.Skipf("Skipping Vision test: failed to create Vertex client: %v", err)
		return
	}
	defer vertexClient.Close()

	visionService := service.NewVisionService(vertexClient)

	// Create a mock server to serve an image
	// We'll use a 1x1 transparent PNG as a dummy image
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

	t.Run("Verify Task Completion", func(t *testing.T) {
		// Since we are using a dummy tiny image, AI might say no or yes depending on the description.
		// For the purpose of the integration test, we want to ensure the plumbing works:
		// Download -> AI Call -> JSON Parse.

		isVerified, confidence, err := visionService.VerifyTask(ctx, ts.URL, "A single transparent pixel representing the start of a digital foundation.")

		require.NoError(t, err)
		t.Logf("Vision Result - Verified: %v, Confidence: %f", isVerified, confidence)

		// We don't strictly assert isVerified because it's LLM based on a dummy image,
		// but confidence should be returned (even if 0).
		assert.GreaterOrEqual(t, confidence, 0.0)
		assert.LessOrEqual(t, confidence, 1.0)
	})

	t.Run("Invalid Image URL", func(t *testing.T) {
		_, _, err := visionService.VerifyTask(ctx, "http://invalid-url-that-does-not-exist.com/img.png", "A test description")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to download image")
	})
}
