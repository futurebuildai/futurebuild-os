// Package clock provides time abstraction for deterministic testing.
// See PRODUCTION_PLAN.md Step 49 (Time-Travel Simulation)
package clock

import (
	"sync"
	"time"
)

// Clock abstracts time operations for testability.
// This enables deterministic simulation by injecting controlled time.
type Clock interface {
	Now() time.Time
}

// RealClock is the production implementation using system time.
type RealClock struct{}

// Now returns the current system time.
func (RealClock) Now() time.Time {
	return time.Now()
}

// MockClock is a test harness for time-travel simulation.
// It is safe for concurrent use.
// See PRODUCTION_PLAN.md Step 49 Amendment 3
type MockClock struct {
	mu      sync.RWMutex
	current time.Time
}

// NewMockClock creates a MockClock starting at the given time.
// The canonical start for simulations is 2026-01-01 08:00:00 UTC.
func NewMockClock(start time.Time) *MockClock {
	return &MockClock{current: start}
}

// Now returns the mocked current time.
func (m *MockClock) Now() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

// Advance moves the clock forward by the given duration.
// This is used in simulation loops to fast-forward time.
func (m *MockClock) Advance(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.current = m.current.Add(d)
}

// Set sets the clock to a specific time.
// Useful for jumping to specific scenario checkpoints.
func (m *MockClock) Set(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.current = t
}
