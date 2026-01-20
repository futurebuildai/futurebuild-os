package e2e

import (
	"context"
	"os"
	"testing"

	"github.com/colton/futurebuild/pkg/ai"
	"github.com/stretchr/testify/require"
)

// TestLiveAI_VertexConnection verifies that the AI client can actually talk to Google Vertex AI.
// This is a "Contract Test" that ensures our integration code matches the live API reality.
//
// Skipped unless FORCE_LIVE_TESTS=1 or VERTEX_PROJECT_ID is explicitly set.
func TestLiveAI_VertexConnection(t *testing.T) {
	projectID := os.Getenv("VERTEX_PROJECT_ID")
	location := os.Getenv("VERTEX_LOCATION")

	if projectID == "" {
		if os.Getenv("GOOGLE_CLOUD_PROJECT") != "" {
			projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
		} else {
			t.Skip("Skipping live AI test: VERTEX_PROJECT_ID not set")
		}
	}
	if location == "" {
		location = "us-central1" // Default
	}

	// Double check we want to run this (costs money/quota)
	if os.Getenv("FORCE_LIVE_TESTS") != "1" {
		t.Skip("Skipping live AI test: FORCE_LIVE_TESTS != 1")
	}

	ctx := context.Background()

	// Configure models
	models := map[ai.ModelType]string{
		ai.ModelTypeFlash: "gemini-1.5-flash-001",
	}

	// 1. Initialize Client
	client, err := ai.NewVertexClient(ctx, projectID, location, models)
	require.NoError(t, err, "Failed to create Vertex client")
	defer client.Close()

	// 2. Generate Content
	req := ai.GenerateRequest{
		Model: ai.ModelTypeFlash,
		Parts: []ai.ContentPart{
			{Text: "Reply with the exact word: 'PONG'"},
		},
	}

	t.Logf("Sending live request to Vertex AI (Project: %s, Location: %s)...", projectID, location)
	resp, err := client.GenerateContent(ctx, req)
	require.NoError(t, err, "Live generation request failed")

	t.Logf("Received response: %s", resp.Text)

	// 3. functional Assertion
	require.NotEmpty(t, resp.Text, "Response should not be empty")
	require.Contains(t, resp.Text, "PONG", "Response should contain expected text")
}
