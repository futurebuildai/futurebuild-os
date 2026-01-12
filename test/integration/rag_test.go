package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/ai"
)

// TestRag_IngestAndSearch verifies the RAG pipeline end-to-end.
// Prerequisites:
// 1. Docker DB running with pgvector.
// 2. Valid Vertex AI credentials.
func TestRag_IngestAndSearch(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	cfg := config.LoadConfig()
	// Override DB URL if needed for local test run (e.g. port 5433)
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = "postgres://fb_user:fb_pass@localhost:5433/futurebuild?sslmode=disable"
	}

	// 1. Setup DB Connection
	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	require.NoError(t, err)
	defer db.Close()

	// ping to ensure connected
	err = db.Ping(context.Background())
	require.NoError(t, err, "Database connection failed")

	// 2. Setup Vertex Client
	// Build map for model IDs
	modelIDs := map[ai.ModelType]string{
		ai.ModelTypeFlash:     cfg.VertexModelFlashID,
		ai.ModelTypePro:       cfg.VertexModelProID,
		ai.ModelTypeEmbedding: cfg.VertexModelEmbeddingID,
	}

	client, err := ai.NewVertexClient(context.Background(), cfg.VertexProjectID, cfg.VertexLocation, modelIDs)
	require.NoError(t, err)
	defer client.Close()

	// 3. Setup Service
	docService := service.NewDocumentService(db, client)

	// 4. Create a Dummy Document
	docID := uuid.New()
	ctx := context.Background()

	// Use random slug to avoid collisions
	randomSlug := fmt.Sprintf("rag-test-org-%s", uuid.New().String())
	orgID := uuid.New()
	_, err = db.Exec(ctx, "INSERT INTO organizations (id, name, slug) VALUES ($1, 'RAG Test Org', $2)", orgID, randomSlug)
	if err != nil {
		t.Fatalf("Failed to create org: %v", err)
	}

	userID := uuid.New()
	randomEmail := fmt.Sprintf("rag-%s@test.com", uuid.New().String())
	_, err = db.Exec(ctx, "INSERT INTO users (id, org_id, email, name, role) VALUES ($1, $2, $3, 'RAG Tester', 'Admin')", userID, orgID, randomEmail)
	require.NoError(t, err)

	projectID := uuid.New()
	_, err = db.Exec(ctx, "INSERT INTO projects (id, org_id, name, status) VALUES ($1, $2, 'RAG Project', 'Active')", projectID, orgID)
	require.NoError(t, err)

	// Insert Document
	docText := `
		The concrete foundation for Lot 42 must be 3000 PSI. 
		Pour is scheduled for next Tuesday at 8 AM.
		Ensure rebar inspection passes before the truck arrives.
		Weather forecast predicts rain in the afternoon.
	`
	_, err = db.Exec(ctx, `
		INSERT INTO documents (id, project_id, type, filename, extracted_text, processing_status, uploaded_by) 
		VALUES ($1, $2, 'blueprint', 'test_foundations.pdf', $3, 'pending', $4)
	`, docID, projectID, docText, userID)
	require.NoError(t, err)

	// 5. Run Ingestion
	t.Logf("Ingesting document %s...", docID)
	err = docService.IngestDocument(ctx, docID)
	assert.NoError(t, err)

	// 6. Verify Chunks Created
	var count int
	err = db.QueryRow(ctx, "SELECT count(*) FROM document_chunks WHERE document_id = $1", docID).Scan(&count)
	assert.NoError(t, err)
	assert.Greater(t, count, 0, "Should have created chunks")

	// 7. Verify Vector Search
	queryText := "What is the concrete strength?"
	queryEmb, err := client.GenerateEmbedding(ctx, queryText)
	require.NoError(t, err)

	// Perform vector similarity search
	rows, err := db.Query(ctx, `
		SELECT chunk_content, embedding <=> $1 as distance 
		FROM document_chunks 
		WHERE document_id = $2 
		ORDER BY distance ASC 
		LIMIT 1
	`, pgvector.NewVector(queryEmb), docID)
	require.NoError(t, err)
	defer rows.Close()

	if rows.Next() {
		var content string
		var distance float64
		err = rows.Scan(&content, &distance)
		assert.NoError(t, err)
		t.Logf("Top Match: %s (Distance: %f)", content, distance)
		assert.Contains(t, content, "3000 PSI")
	} else {
		t.Errorf("No results found for vector search")
	}
}
