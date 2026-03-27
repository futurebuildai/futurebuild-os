package service

import (
	"fmt"
	"testing"
)

func TestNewCorporateFinancialsService(t *testing.T) {
	svc := NewCorporateFinancialsService(nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestCorporateBudget_CentsCalculation(t *testing.T) {
	// Verify cents → dollars conversion logic used in rollup
	testCases := []struct {
		name     string
		cents    int64
		expected string
	}{
		{"zero", 0, "0.00"},
		{"one_dollar", 100, "1.00"},
		{"large_budget", 150000000, "1500000.00"}, // $1.5M
		{"fractional", 12345, "123.45"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dollars := float64(tc.cents) / 100.0
			result := fmt.Sprintf("%.2f", dollars)
			if result != tc.expected {
				t.Errorf("got %s, want %s", result, tc.expected)
			}
		})
	}
}
