//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/colton/futurebuild/test/testhelpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReviewGate_ConfidenceThreshold(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	cfg := getTestConfig("")
	ctx := context.Background()
	db, cleanup := testhelpers.StartPostgresContainer(t)
	defer cleanup()
	var err error

	// Setup Service (client isn't used for SaveExtraction)
	invoiceService := service.NewInvoiceService(db, nil, cfg)

	// Setup Test Data
	orgID := uuid.New()
	randomSlug := fmt.Sprintf("review-test-org-%s", uuid.New().String())
	_, err = db.Exec(ctx, "INSERT INTO organizations (id, name, slug) VALUES ($1, 'Review Test Org', $2)", orgID, randomSlug)
	require.NoError(t, err)

	projectID := uuid.New()
	_, err = db.Exec(ctx, "INSERT INTO projects (id, org_id, name, status) VALUES ($1, $2, 'Review Test Project', 'Active')", projectID, orgID)
	require.NoError(t, err)

	t.Run("Low Confidence Flags Human Review", func(t *testing.T) {
		ext := &types.InvoiceExtraction{
			Vendor:           "Low Confidence Vendor",
			TotalAmountCents: 10000, // $100.00
			Confidence:       0.5,   // Below 0.85 threshold
			SuggestedWBSCode: "1.1",
			Date:             "2026-01-01",
			InvoiceNumber:    "INV-LOW",
		}

		invoiceID, err := invoiceService.SaveExtraction(ctx, projectID, ext, nil)
		require.NoError(t, err)

		var isRequired bool
		err = db.QueryRow(ctx, "SELECT is_human_review_required FROM invoices WHERE id = $1", invoiceID).Scan(&isRequired)
		assert.NoError(t, err)
		assert.True(t, isRequired, "Invoice with 0.5 confidence should require human review")
	})

	t.Run("High Confidence Bypasses Human Review", func(t *testing.T) {
		ext := &types.InvoiceExtraction{
			Vendor:           "High Confidence Vendor",
			TotalAmountCents: 20000, // $200.00
			Confidence:       0.95,  // Above 0.85 threshold
			SuggestedWBSCode: "2.2",
			Date:             "2026-01-02",
			InvoiceNumber:    "INV-HIGH",
		}

		invoiceID, err := invoiceService.SaveExtraction(ctx, projectID, ext, nil)
		require.NoError(t, err)

		var isRequired bool
		err = db.QueryRow(ctx, "SELECT is_human_review_required FROM invoices WHERE id = $1", invoiceID).Scan(&isRequired)
		assert.NoError(t, err)
		assert.False(t, isRequired, "Invoice with 0.95 confidence should NOT require human review")
	})
}
