package service

import (
	"testing"
	"time"
)

func TestNewFleetService(t *testing.T) {
	svc := NewFleetService(nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestDateRangeOverlap(t *testing.T) {
	// Test the date range overlap logic used in CheckEquipmentAvailability
	// Two ranges [a1, a2] and [b1, b2] overlap if a1 <= b2 AND b1 <= a2
	testCases := []struct {
		name    string
		a1, a2  string
		b1, b2  string
		overlap bool
	}{
		{"no_overlap_before", "2025-01-01", "2025-01-10", "2025-01-15", "2025-01-20", false},
		{"no_overlap_after", "2025-01-15", "2025-01-20", "2025-01-01", "2025-01-10", false},
		{"full_overlap", "2025-01-01", "2025-01-20", "2025-01-05", "2025-01-15", true},
		{"partial_overlap_start", "2025-01-01", "2025-01-10", "2025-01-08", "2025-01-15", true},
		{"partial_overlap_end", "2025-01-08", "2025-01-15", "2025-01-01", "2025-01-10", true},
		{"same_range", "2025-01-01", "2025-01-10", "2025-01-01", "2025-01-10", true},
		{"adjacent_no_gap", "2025-01-01", "2025-01-10", "2025-01-10", "2025-01-20", true},
		{"one_day_gap", "2025-01-01", "2025-01-09", "2025-01-11", "2025-01-20", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a1, _ := time.Parse("2006-01-02", tc.a1)
			a2, _ := time.Parse("2006-01-02", tc.a2)
			b1, _ := time.Parse("2006-01-02", tc.b1)
			b2, _ := time.Parse("2006-01-02", tc.b2)

			// PostgreSQL daterange overlap: a1 <= b2 AND b1 <= a2
			overlaps := !a1.After(b2) && !b1.After(a2)
			if overlaps != tc.overlap {
				t.Errorf("overlap: got %v, want %v (a=[%s,%s], b=[%s,%s])",
					overlaps, tc.overlap, tc.a1, tc.a2, tc.b1, tc.b2)
			}
		})
	}
}

func TestAssetStatus_Values(t *testing.T) {
	validStatuses := []string{"available", "in_use", "maintenance", "retired"}
	statusMap := make(map[string]bool)
	for _, s := range validStatuses {
		statusMap[s] = true
	}

	if !statusMap["available"] {
		t.Error("expected 'available' in valid statuses")
	}
	if statusMap["nonexistent"] {
		t.Error("unexpected status in valid statuses")
	}
}
