package clock

import (
	"testing"
	"time"
)

// TestRealClock_Now verifies RealClock returns current time.
func TestRealClock_Now(t *testing.T) {
	clock := RealClock{}
	before := time.Now()
	got := clock.Now()
	after := time.Now()

	if got.Before(before) || got.After(after) {
		t.Errorf("RealClock.Now() = %v, want between %v and %v", got, before, after)
	}
}

// TestMockClock_Now verifies MockClock returns the set time.
func TestMockClock_Now(t *testing.T) {
	start := time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC)
	clock := NewMockClock(start)

	got := clock.Now()
	if !got.Equal(start) {
		t.Errorf("MockClock.Now() = %v, want %v", got, start)
	}
}

// TestMockClock_Advance verifies Advance moves time forward.
func TestMockClock_Advance(t *testing.T) {
	start := time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC)
	clock := NewMockClock(start)

	clock.Advance(24 * time.Hour)
	got := clock.Now()
	want := time.Date(2026, 1, 2, 8, 0, 0, 0, time.UTC)

	if !got.Equal(want) {
		t.Errorf("After Advance(24h), Now() = %v, want %v", got, want)
	}
}

// TestMockClock_Set verifies Set jumps to specific time.
func TestMockClock_Set(t *testing.T) {
	start := time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC)
	clock := NewMockClock(start)

	target := time.Date(2026, 2, 15, 12, 0, 0, 0, time.UTC)
	clock.Set(target)

	got := clock.Now()
	if !got.Equal(target) {
		t.Errorf("After Set(), Now() = %v, want %v", got, target)
	}
}

// TestMockClock_SimulationLoop verifies clock works in a 30-day loop.
// See PRODUCTION_PLAN.md Step 49
func TestMockClock_SimulationLoop(t *testing.T) {
	start := time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC)
	clock := NewMockClock(start)

	for day := 0; day <= 30; day++ {
		expected := start.Add(time.Duration(day) * 24 * time.Hour)
		if got := clock.Now(); !got.Equal(expected) {
			t.Errorf("Day %d: Now() = %v, want %v", day, got, expected)
		}
		clock.Advance(24 * time.Hour)
	}
}
