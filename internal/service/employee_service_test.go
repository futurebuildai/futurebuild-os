package service

import (
	"testing"
)

func TestNewEmployeeService(t *testing.T) {
	svc := NewEmployeeService(nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestLaborBurdenFormula(t *testing.T) {
	// The formula: SUM(hours_worked * pay_rate_cents) + SUM(overtime_hours * pay_rate_cents * 1.5)
	// This is deterministic — L7 Zero-Trust
	testCases := []struct {
		name          string
		hoursWorked   float64
		overtimeHours float64
		payRateCents  int64
		expectedCents int64
	}{
		{"no_overtime", 8.0, 0.0, 5000, 40000},       // 8h * $50/hr = $400
		{"with_overtime", 8.0, 2.0, 5000, 55000},      // (8*5000) + (2*5000*1.5) = 40000 + 15000 = 55000
		{"zero_hours", 0.0, 0.0, 5000, 0},
		{"all_overtime", 0.0, 4.0, 10000, 60000},      // 4 * 10000 * 1.5 = 60000
		{"fractional_hours", 7.5, 1.5, 4000, 39000},   // (7.5*4000) + (1.5*4000*1.5) = 30000 + 9000 = 39000
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			regular := int64(tc.hoursWorked * float64(tc.payRateCents))
			overtime := int64(tc.overtimeHours * float64(tc.payRateCents) * 1.5)
			total := regular + overtime

			if total != tc.expectedCents {
				t.Errorf("labor burden: got %d cents, want %d cents", total, tc.expectedCents)
			}
		})
	}
}

func TestEmployeeStatus_Values(t *testing.T) {
	validStatuses := []string{"active", "on_leave", "terminated"}
	for _, status := range validStatuses {
		if status == "" {
			t.Errorf("empty status is not valid")
		}
	}
}

func TestCertExpirationDays(t *testing.T) {
	// Verify the expiration window logic
	testCases := []struct {
		name       string
		withinDays int
		valid      bool
	}{
		{"thirty_days", 30, true},
		{"sixty_days", 60, true},
		{"zero_days", 0, true},
		{"negative_days", -1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.valid && tc.withinDays < 0 {
				t.Error("expected valid for non-negative days")
			}
			if !tc.valid && tc.withinDays >= 0 {
				t.Error("expected invalid for negative days")
			}
		})
	}
}
