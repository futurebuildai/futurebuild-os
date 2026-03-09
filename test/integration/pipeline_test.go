//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/colton/futurebuild/test/testhelpers"
)

// noOpClient is defined in chat_test.go (same package)

func TestPipeline_MockIngestion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	// 1. Setup DB Connection
	cfg := getTestConfig("")
	ctx := context.Background()
	db, cleanup := testhelpers.StartPostgresContainer(t)
	defer cleanup()
	var err error

	// 2. Setup Service (with no-op client as we won't call AI)
	invoiceService := service.NewInvoiceService(db, &noOpClient{}, cfg)

	// 3. Prepare Database State (Org & Project)
	orgID := uuid.New()
	randomSlug := fmt.Sprintf("pipeline-test-org-%s", uuid.New().String())
	_, err = db.Exec(ctx, "INSERT INTO organizations (id, name, slug) VALUES ($1, 'Pipeline Test Org', $2)", orgID, randomSlug)
	require.NoError(t, err)

	projectID := uuid.New()
	_, err = db.Exec(ctx, "INSERT INTO projects (id, org_id, name, status) VALUES ($1, $2, 'Pipeline Test Project', 'Active')", projectID, orgID)
	require.NoError(t, err)

	// 4. Load Fixture
	wd, err := os.Getwd() // Should be /.../repo
	require.NoError(t, err)

	// Adjust path depending on where test is run from.
	// Assuming `go test ./test/integration/...` is run from repo root.
	// But `os.Getwd` might need handling if test runs inside directory.
	// A safer bet is finding the fixtures relative to the test file if compiling,
	// but simplest is to try relative path from git root.

	// If run from test/integration, we go up one level.
	// If run from root, we go down.
	// Let's assume root execution context or try to find it.
	fixturePath := filepath.Join(wd, "../../test/fixtures/perfect_invoice.json")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		// Fallback: maybe we are at root
		fixturePath = "test/fixtures/perfect_invoice.json"
	}

	b, err := os.ReadFile(fixturePath)
	require.NoError(t, err, "Could not find perfect_invoice.json")

	var extraction types.InvoiceExtraction
	err = json.Unmarshal(b, &extraction)
	require.NoError(t, err, "Failed to unmarshal fixture")

	// 5. Execute Action: SaveExtraction
	t.Logf("Injecting invoice from fixture into project %s...", projectID)
	invoiceID, err := invoiceService.SaveExtraction(ctx, projectID, &extraction, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, invoiceID)

	// 6. Verify Assertions
	var (
		vendorName            string
		amountCents           int64
		status                models.InvoiceStatus
		confidence            float64
		invoiceNumber         string
		invoiceDate           time.Time
		detectedWBSCode       string
		isHumanReviewRequired bool
		lineItemsJSON         []byte
	)

	err = db.QueryRow(ctx, `
		SELECT 
			vendor_name, 
			amount_cents, 
			status, 
			confidence, 
			invoice_number, 
			invoice_date,
			detected_wbs_code,
			is_human_review_required,
			line_items
		FROM invoices 
		WHERE id = $1`, invoiceID).Scan(
		&vendorName,
		&amountCents,
		&status,
		&confidence,
		&invoiceNumber,
		&invoiceDate,
		&detectedWBSCode,
		&isHumanReviewRequired,
		&lineItemsJSON,
	)
	require.NoError(t, err)

	// Assertions based on "perfect_invoice.json"
	assert.Equal(t, "Acme Supply Co.", vendorName)
	assert.Equal(t, int64(150000), amountCents) // $1500.00 in cents
	assert.Equal(t, "INV-1001", invoiceNumber)
	assert.Equal(t, "2026-01-12", invoiceDate.Format("2006-01-02"))
	assert.Equal(t, "6.1.2", detectedWBSCode)
	assert.Equal(t, 0.95, confidence)
	assert.False(t, isHumanReviewRequired, "Should not require review for high confidence")
	assert.Equal(t, models.InvoiceStatusDraft, status) // Default status

	// Deep check line items
	var lineItems models.LineItems
	err = json.Unmarshal(lineItemsJSON, &lineItems)
	require.NoError(t, err)
	assert.Len(t, lineItems, 2)
	assert.Equal(t, "Lumber 2x4x10", lineItems[0].Description)
	assert.Equal(t, int64(55000), lineItems[0].TotalCents) // $550.00 in cents
}
