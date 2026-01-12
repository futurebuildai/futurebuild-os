package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/types"
)

func TestReviewGate_ConfidenceThreshold(t *testing.T) {
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

	// Setup Service (client isn't used for SaveExtraction)
	invoiceService := service.NewInvoiceService(db, nil)

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
			TotalAmount:      100.0,
			Confidence:       0.5, // Below 0.85 threshold
			SuggestedWBSCode: "1.1",
			Date:             "2026-01-01",
			InvoiceNumber:    "INV-LOW",
		}

		invoiceID, err := invoiceService.SaveExtraction(ctx, projectID, ext)
		require.NoError(t, err)

		var isRequired bool
		err = db.QueryRow(ctx, "SELECT is_human_review_required FROM invoices WHERE id = $1", invoiceID).Scan(&isRequired)
		assert.NoError(t, err)
		assert.True(t, isRequired, "Invoice with 0.5 confidence should require human review")
	})

	t.Run("High Confidence Bypasses Human Review", func(t *testing.T) {
		ext := &types.InvoiceExtraction{
			Vendor:           "High Confidence Vendor",
			TotalAmount:      200.0,
			Confidence:       0.95, // Above 0.85 threshold
			SuggestedWBSCode: "2.2",
			Date:             "2026-01-02",
			InvoiceNumber:    "INV-HIGH",
		}

		invoiceID, err := invoiceService.SaveExtraction(ctx, projectID, ext)
		require.NoError(t, err)

		var isRequired bool
		err = db.QueryRow(ctx, "SELECT is_human_review_required FROM invoices WHERE id = $1", invoiceID).Scan(&isRequired)
		assert.NoError(t, err)
		assert.False(t, isRequired, "Invoice with 0.95 confidence should NOT require human review")
	})
}
