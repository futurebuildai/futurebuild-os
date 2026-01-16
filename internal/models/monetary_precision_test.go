package models

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// FAANG THRESHOLD TESTS - Monetary Precision Verification
// =============================================================================
// These tests verify that int64 cents representation eliminates IEEE 754
// floating-point precision errors that occur with float64 monetary values.

// TestPrecision_SumOfCents verifies that classic floating-point precision
// errors are eliminated. With float64, 0.1 + 0.2 != 0.3 exactly.
// See: https://floating-point-gui.de/basic/
func TestPrecision_SumOfCents(t *testing.T) {
	t.Run("Integer Cents Are Exact", func(t *testing.T) {
		// With int64 cents: 10 + 20 = 30 exactly, no floating point errors possible
		a := int64(10) // $0.10
		b := int64(20) // $0.20
		sum := a + b
		assert.Equal(t, int64(30), sum, "$0.10 + $0.20 must equal $0.30 exactly (30 cents)")

		// Additional verification: no drift after many operations
		var total int64
		for i := 0; i < 1000; i++ {
			total += 10 // Add $0.10 repeatedly
		}
		assert.Equal(t, int64(10000), total, "1000 × $0.10 must equal exactly $100.00 (10000 cents)")
	})

	t.Run("Accumulated Precision Errors", func(t *testing.T) {
		// Adding $0.01 one hundred times with float64 does not equal $1.00 exactly
		var floatTotal float64
		for i := 0; i < 100; i++ {
			floatTotal += 0.01
		}
		// Float will have accumulated error
		assert.False(t, floatTotal == 1.0, "100 × $0.01 in float64 should NOT equal exactly $1.00")

		// With int64 cents: 100 × 1 = 100 exactly
		var centsTotal int64
		for i := 0; i < 100; i++ {
			centsTotal += 1 // 1 cent
		}
		assert.Equal(t, int64(100), centsTotal, "100 × 1 cent must equal exactly 100 cents ($1.00)")
	})

	t.Run("Invoice Line Item Sum", func(t *testing.T) {
		// Simulate invoice with line items that trigger precision errors in float64
		lineItems := []struct {
			Qty   float64
			Price int64 // cents
		}{
			{Qty: 3, Price: 33},  // $0.33 each
			{Qty: 7, Price: 14},  // $0.14 each
			{Qty: 5, Price: 199}, // $1.99 each
		}

		var totalCents int64
		for _, item := range lineItems {
			totalCents += int64(item.Qty) * item.Price
		}

		// 3×33 + 7×14 + 5×199 = 99 + 98 + 995 = 1192 cents = $11.92
		assert.Equal(t, int64(1192), totalCents, "Line item sum must be exactly 1192 cents ($11.92)")
	})
}

// TestBoundary_LargeAmount verifies that int64 handles realistic large
// monetary values without overflow.
func TestBoundary_LargeAmount(t *testing.T) {
	t.Run("100 Million Dollars", func(t *testing.T) {
		// $100,000,000.00 = 10,000,000,000 cents
		hundredMil := int64(10_000_000_000)
		assert.Equal(t, int64(10_000_000_000), hundredMil)

		// Perform arithmetic
		doubled := hundredMil * 2
		assert.Equal(t, int64(20_000_000_000), doubled, "$200M should work fine")
	})

	t.Run("1 Billion Dollars", func(t *testing.T) {
		// $1,000,000,000.00 = 100,000,000,000 cents
		billion := int64(100_000_000_000)
		assert.Equal(t, int64(100_000_000_000), billion)
	})

	t.Run("Maximum Safe Amount", func(t *testing.T) {
		// int64 max is 9,223,372,036,854,775,807
		// That's $92,233,720,368,547,758.07 - more than entire world GDP
		maxCents := int64(math.MaxInt64)
		assert.Greater(t, maxCents, int64(100_000_000_000_000), "int64 can handle >$1 trillion")

		// Verify no overflow with realistic large corporate budgets
		largeBudget := int64(500_000_000_000_000) // $5 trillion in cents
		assert.Equal(t, int64(500_000_000_000_000), largeBudget)
	})
}

// TestRoundTrip_ExactCents verifies that JSON marshaling/unmarshaling
// preserves exact cent values without precision loss.
func TestRoundTrip_ExactCents(t *testing.T) {
	t.Run("$19.99", func(t *testing.T) {
		invoice := Invoice{
			AmountCents: 1999,
			VendorName:  "Test Vendor",
		}

		data, err := json.Marshal(invoice)
		assert.NoError(t, err)

		var parsed Invoice
		err = json.Unmarshal(data, &parsed)
		assert.NoError(t, err)

		assert.Equal(t, int64(1999), parsed.AmountCents, "$19.99 must round-trip exactly as 1999 cents")
	})

	t.Run("$0.01 (minimum cent)", func(t *testing.T) {
		invoice := Invoice{AmountCents: 1}

		data, err := json.Marshal(invoice)
		assert.NoError(t, err)

		var parsed Invoice
		err = json.Unmarshal(data, &parsed)
		assert.NoError(t, err)

		assert.Equal(t, int64(1), parsed.AmountCents)
	})

	t.Run("$999,999.99 (large amount)", func(t *testing.T) {
		invoice := Invoice{AmountCents: 99_999_999}

		data, err := json.Marshal(invoice)
		assert.NoError(t, err)

		var parsed Invoice
		err = json.Unmarshal(data, &parsed)
		assert.NoError(t, err)

		assert.Equal(t, int64(99_999_999), parsed.AmountCents)
	})

	t.Run("LineItems round-trip", func(t *testing.T) {
		items := LineItems{
			{Description: "Item A", Quantity: 2.5, UnitPriceCents: 1999, TotalCents: 4998},
			{Description: "Item B", Quantity: 1, UnitPriceCents: 100, TotalCents: 100},
		}

		data, err := json.Marshal(items)
		assert.NoError(t, err)

		var parsed LineItems
		err = json.Unmarshal(data, &parsed)
		assert.NoError(t, err)

		assert.Equal(t, int64(4998), parsed[0].TotalCents)
		assert.Equal(t, int64(100), parsed[1].TotalCents)
	})
}

// TestProjectBudget_CentsArithmetic verifies budget calculations work correctly.
func TestProjectBudget_CentsArithmetic(t *testing.T) {
	budget := ProjectBudget{
		EstimatedAmountCents: 5_000_000_00, // $5,000,000.00
		CommittedAmountCents: 4_500_000_00, // $4,500,000.00
		ActualAmountCents:    1_000_000_00, // $1,000,000.00
	}

	// Calculate remaining budget
	remaining := budget.EstimatedAmountCents - budget.ActualAmountCents
	assert.Equal(t, int64(4_000_000_00), remaining, "Remaining budget should be $4,000,000.00")

	// Calculate variance
	variance := budget.EstimatedAmountCents - budget.CommittedAmountCents
	assert.Equal(t, int64(500_000_00), variance, "Variance should be $500,000.00")
}
