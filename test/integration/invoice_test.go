package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/ai"
)

func TestInvoice_AnalyzeAndSave(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	cfg := getTestConfig()
	ctx := context.Background()

	// 1. Setup DB Connection
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Skipf("Skipping test: database not reachable: %v", err)
	}

	// 2. Setup Vertex Client
	var client ai.Client
	modelIDs := map[ai.ModelType]string{
		ai.ModelTypeFlash:     cfg.VertexModelFlashID,
		ai.ModelTypePro:       cfg.VertexModelProID,
		ai.ModelTypeEmbedding: cfg.VertexModelEmbeddingID,
	}
	if cfg.VertexProjectID != "" {
		client, err = ai.NewVertexClient(context.Background(), cfg.VertexProjectID, cfg.VertexLocation, modelIDs)
		require.NoError(t, err)
	} else {
		t.Log("Vertex generated ID missing, using Mock client")
		client = &MockVertexClient{}
	}
	defer client.Close()

	// 3. Setup Service
	invoiceService := service.NewInvoiceService(db, client)

	// 4. Test Data Setup
	orgID := uuid.New()
	randomSlug := fmt.Sprintf("invoice-test-org-%s", uuid.New().String())
	_, err = db.Exec(ctx, "INSERT INTO organizations (id, name, slug) VALUES ($1, 'Invoice Test Org', $2)", orgID, randomSlug)
	require.NoError(t, err)

	projectID := uuid.New()
	_, err = db.Exec(ctx, "INSERT INTO projects (id, org_id, name, status) VALUES ($1, $2, 'Invoice Test Project', 'Active')", projectID, orgID)
	require.NoError(t, err)

	userID := uuid.New()
	randomEmail := fmt.Sprintf("invoice-%s@test.com", uuid.New().String())
	_, err = db.Exec(ctx, "INSERT INTO users (id, org_id, email, name, role) VALUES ($1, $2, $3, 'Invoice Tester', 'Admin')", userID, orgID, randomEmail)
	require.NoError(t, err)

	docID := uuid.New()
	invoiceText := `
		INVOICE #INV-2026-001
		Vendor: ACME Concrete Supplies
		Date: 2026-01-15
		
		Description                     Qty     Price     Total
		3000 PSI Ready Mix Concrete     10      120.00    1200.00
		Delivery Fee                    1       150.00    150.00
		Environmental Surcharge         1       50.00     50.00
		
		Total Amount Due: $1400.00
		WBS Code: 6.2 (Foundations)
	`
	_, err = db.Exec(ctx, `
		INSERT INTO documents (id, project_id, type, filename, extracted_text, processing_status, uploaded_by) 
		VALUES ($1, $2, 'invoice', 'acme_concrete.pdf', $3, 'completed', $4)
	`, docID, projectID, invoiceText, userID)
	require.NoError(t, err)

	// 5. Run Analysis
	t.Logf("Analyzing invoice %s...", docID)
	pID, extraction, err := invoiceService.AnalyzeInvoice(ctx, orgID, docID)
	require.NoError(t, err)
	assert.Equal(t, projectID, pID)

	assert.Equal(t, "ACME Concrete Supplies", extraction.Vendor)
	assert.Equal(t, int64(140000), extraction.TotalAmountCents) // $1400.00 in cents
	assert.Greater(t, extraction.Confidence, 0.5)
	assert.Len(t, extraction.LineItems, 3)

	// 6. Save Extraction
	t.Logf("Saving extraction for project %s...", projectID)
	invoiceID, err := invoiceService.SaveExtraction(ctx, projectID, extraction, nil)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, invoiceID)

	// 7. Verify Persistence
	var vendorName string
	var amountCents int64
	var status models.InvoiceStatus
	var confidence float64
	var invoiceDate time.Time
	var invoiceNumber string
	err = db.QueryRow(ctx, "SELECT vendor_name, amount_cents, status, confidence, invoice_date, invoice_number FROM invoices WHERE id = $1", invoiceID).Scan(
		&vendorName, &amountCents, &status, &confidence, &invoiceDate, &invoiceNumber)
	assert.NoError(t, err)
	assert.Equal(t, "ACME Concrete Supplies", vendorName)
	assert.Equal(t, int64(140000), amountCents) // $1400.00 in cents
	assert.Equal(t, models.InvoiceStatusPending, status)
	assert.Greater(t, confidence, 0.5)
	assert.False(t, invoiceDate.IsZero())
	assert.NotEmpty(t, invoiceNumber)
}
