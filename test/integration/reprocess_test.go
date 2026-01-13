package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocument_Reprocess(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	cfg := config.LoadConfig()
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = "postgres://fb_user:fb_pass@localhost:5433/futurebuild?sslmode=disable"
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	require.NoError(t, err)
	defer db.Close()

	// 1. Setup Services (Mock Client for consistent AI response)
	mockClient := &MockVertexClient{}
	docService := service.NewDocumentService(db, mockClient)
	invoiceService := service.NewInvoiceService(db, mockClient)

	// 2. Setup Test Data
	orgID := uuid.New()
	randomSlug := fmt.Sprintf("reprocess-org-%s", uuid.New().String())
	_, err = db.Exec(ctx, "INSERT INTO organizations (id, name, slug) VALUES ($1, 'Reprocess Org', $2)", orgID, randomSlug)
	require.NoError(t, err)

	projectID := uuid.New()
	_, err = db.Exec(ctx, "INSERT INTO projects (id, org_id, name, status) VALUES ($1, $2, 'Reprocess Project', 'Active')", projectID, orgID)
	require.NoError(t, err)

	userID := uuid.New()
	randomEmail := fmt.Sprintf("reprocess-%s@test.com", uuid.New().String())
	_, err = db.Exec(ctx, "INSERT INTO users (id, org_id, email, name, role) VALUES ($1, $2, $3, 'Reprocess User', 'Admin')", userID, orgID, randomEmail)
	require.NoError(t, err)

	docID := uuid.New()
	extractedText := "Invoice 123 from Vendor ABC for $500.00"
	_, err = db.Exec(ctx, `
		INSERT INTO documents (
			id, project_id, type, filename, storage_path, 
			mime_type, file_size_bytes, processing_status, extracted_text, uploaded_by
		) VALUES ($1, $2, 'invoice', 'invoice.pdf', '/auth/invoice.pdf', 
			'application/pdf', 1024, 'pending', $3, $4
		)
	`, docID, projectID, extractedText, userID)
	require.NoError(t, err)

	// 3. Initial Analysis (Simulate typical flow)
	_, initialExtraction, err := invoiceService.AnalyzeInvoice(ctx, orgID, docID)
	require.NoError(t, err)
	initialInvoiceID, err := invoiceService.SaveExtraction(ctx, projectID, initialExtraction, &docID)
	require.NoError(t, err)

	t.Logf("Initial Invoice ID: %s", initialInvoiceID)

	// 4. Update Document Text (simulate user uploading a correction)
	newText := "Invoice 123 from Vendor DEF for $600.00" // Vendor changed ABC -> DEF, Amount $500 -> $600
	_, err = db.Exec(ctx, "UPDATE documents SET extracted_text = $1 WHERE id = $2", newText, docID)
	require.NoError(t, err)

	// 5. Trigger Reprocess (Simulating Handler Call)
	// We call service directly first to check logic, then verify counts
	err = docService.ReprocessDocument(ctx, orgID, docID) // RAG reset
	require.NoError(t, err)

	// Simulate Handler Chain (Re-extract + Save)
	// Note: In a real HTTP test we'd hit the endpoint, but here we verify the Service Logic chain
	// as implemented in the Handler.
	_, reExtraction, err := invoiceService.AnalyzeInvoice(ctx, orgID, docID)
	require.NoError(t, err)

	reProcessedInvoiceID, err := invoiceService.SaveExtraction(ctx, projectID, reExtraction, &docID)
	require.NoError(t, err)

	// 6. Verify Results
	assert.Equal(t, initialInvoiceID, reProcessedInvoiceID, "Invoice ID should remain the same (UPSERT)")

	var vendorName string
	var amount float64
	err = db.QueryRow(ctx, "SELECT vendor_name, amount FROM invoices WHERE id = $1", initialInvoiceID).Scan(&vendorName, &amount)
	require.NoError(t, err)

	// Mock Client returns static "Mock Vendor", so we can't assert on vendor change unless we configure the mock per call.
	// However, we CAN assert that a NEW invoice record wasn't created.

	// Verify Document Reprocessed Count
	var reprocessedCount int
	var status string
	status, reprocessedCount, err = docService.GetDocumentStatus(ctx, docID)
	require.NoError(t, err)

	assert.Equal(t, 1, reprocessedCount, "Reprocessed count should be 1")
	assert.Equal(t, string(models.ProcessingStatusCompleted), status, "Status should be completed")
}
