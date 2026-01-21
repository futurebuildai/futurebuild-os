//go:build integration

package integration

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/colton/futurebuild/pkg/ai"
	"github.com/colton/futurebuild/pkg/storage"
)

// TestInfra_S3_Vertex verifies the connection to external infrastructure.
// It requires specific environment variables to be set.
func TestInfra_S3_Vertex(t *testing.T) {
	cfg := getTestConfig()

	t.Run("S3_Upload_And_Sign", func(t *testing.T) {
		if cfg.S3Endpoint == "" || cfg.S3AccessKey == "" {
			t.Skip("Skipping S3 test: S3_ENDPOINT or S3_ACCESS_KEY not set")
		}

		// Use 'minioadmin' defaults if testing against local docker compose from default config
		// But here we rely on what config loaded.

		// 1. Initialize Client
		// For local MinIO (standard port 9000), useSSL=false
		useSSL := !strings.Contains(cfg.S3Endpoint, "localhost") && !strings.Contains(cfg.S3Endpoint, "127.0.0.1")

		s3Client, err := storage.NewS3Client(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket, useSSL)
		require.NoError(t, err)

		// 2. Upload File
		key := "integration-test/hello.txt"
		content := "Hello FutureBuild Storage"
		reader := strings.NewReader(content)

		ctx := context.Background()
		err = s3Client.Upload(ctx, key, reader, int64(len(content)), "text/plain")
		require.NoError(t, err, "Failed to upload file to S3")

		// 3. Generate Signed URL
		url, err := s3Client.GetSignedURL(ctx, key, 1*time.Hour)
		require.NoError(t, err, "Failed to get signed URL")
		assert.NotEmpty(t, url)
		t.Logf("Generated S3 URL: %s", url)
	})

	t.Run("Vertex_GenAI_Check", func(t *testing.T) {
		if cfg.VertexProjectID == "" {
			t.Skip("Skipping Vertex test: VERTEX_PROJECT_ID not set")
		}

		// Check for credentials file
		credsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if credsFile == "" {
			// If not set, maybe default login works, but let's warn
			t.Log("GOOGLE_APPLICATION_CREDENTIALS not set, relying on Application Default Credentials")
		}

		ctx := context.Background()
		models := map[ai.ModelType]string{
			ai.ModelTypeFlash:     cfg.VertexModelFlashID,
			ai.ModelTypeEmbedding: cfg.VertexModelEmbeddingID,
		}

		client, err := ai.NewVertexClient(ctx, cfg.VertexProjectID, cfg.VertexLocation, models)
		if err != nil {
			t.Skipf("Skipping Vertex test: failed to create client (auth likely missing): %v", err)
			return
		}
		defer client.Close()

		// Simple prompt - L7 Vendor Abstraction: Use vendor-agnostic types
		req := ai.NewTextRequest(ai.ModelTypeFlash, "Say 'Hello Integration Test'")
		resp, err := client.GenerateContent(ctx, req)
		if err != nil {
			// If quota exceeded or auth fails, we might fail here.
			// skipping as this depends on external cloud configuration
			t.Skipf("Vertex generation failed (likely env/auth): %v", err)
			return
		}

		assert.NotEmpty(t, resp.Text)
		t.Logf("Vertex Response: %s", resp.Text)
	})
}
