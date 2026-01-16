// Package chat provides Circuit Breaker pattern for audit system availability.
// See PRODUCTION_PLAN.md Audit Trail Durability Remediation.
package chat

import (
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	// CircuitClosed is normal operation - requests flow through.
	CircuitClosed CircuitState = iota
	// CircuitOpen blocks requests - audit system unavailable.
	CircuitOpen
	// CircuitHalfOpen allows one probe request to test recovery.
	CircuitHalfOpen
)

// String returns the string representation of the circuit state.
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "CLOSED"
	case CircuitOpen:
		return "OPEN"
	case CircuitHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// AuditCircuitBreaker defines the interface for audit system circuit breaker.
// Used to implement read-only mode when audit infrastructure is unavailable.
type AuditCircuitBreaker interface {
	// RecordSuccess records a successful audit operation.
	// May transition from Open → HalfOpen → Closed.
	RecordSuccess()

	// RecordFailure records a failed audit operation.
	// May transition from Closed → Open.
	RecordFailure()

	// IsOpen returns true if the circuit is open (audit system unavailable).
	// When open, Lane A writes should be blocked (read-only mode).
	IsOpen() bool

	// State returns the current circuit state for observability.
	State() CircuitState
}

// CircuitBreakerConfig holds configuration for SimpleCircuitBreaker.
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of consecutive failures before opening.
	FailureThreshold int

	// OpenTimeout is how long to stay open before trying half-open.
	OpenTimeout time.Duration

	// HalfOpenSuccessThreshold is successes needed in half-open to close.
	HalfOpenSuccessThreshold int
}

// DefaultCircuitBreakerConfig returns sensible defaults.
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold:         3,                // Open after 3 consecutive failures
		OpenTimeout:              30 * time.Second, // Try half-open after 30s
		HalfOpenSuccessThreshold: 2,                // Close after 2 successes in half-open
	}
}

// SimpleCircuitBreaker implements AuditCircuitBreaker with threshold-based state machine.
// Thread-safe for concurrent access.
type SimpleCircuitBreaker struct {
	mu     sync.RWMutex
	config CircuitBreakerConfig

	state               CircuitState
	consecutiveFailures int
	halfOpenSuccesses   int
	openedAt            time.Time
}

// NewSimpleCircuitBreaker creates a new circuit breaker with the given config.
func NewSimpleCircuitBreaker(config CircuitBreakerConfig) *SimpleCircuitBreaker {
	return &SimpleCircuitBreaker{
		config: config,
		state:  CircuitClosed,
	}
}

// RecordSuccess records a successful audit operation.
func (cb *SimpleCircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		// Reset failures on success
		cb.consecutiveFailures = 0

	case CircuitHalfOpen:
		cb.halfOpenSuccesses++
		if cb.halfOpenSuccesses >= cb.config.HalfOpenSuccessThreshold {
			// Transition to closed
			cb.state = CircuitClosed
			cb.consecutiveFailures = 0
			cb.halfOpenSuccesses = 0
		}

	case CircuitOpen:
		// Shouldn't get success in open state, but reset if we do
		cb.state = CircuitHalfOpen
		cb.halfOpenSuccesses = 1
	}
}

// RecordFailure records a failed audit operation.
func (cb *SimpleCircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		cb.consecutiveFailures++
		if cb.consecutiveFailures >= cb.config.FailureThreshold {
			cb.state = CircuitOpen
			cb.openedAt = time.Now()
		}

	case CircuitHalfOpen:
		// Any failure in half-open reopens the circuit
		cb.state = CircuitOpen
		cb.openedAt = time.Now()
		cb.halfOpenSuccesses = 0

	case CircuitOpen:
		// Already open, reset timer
		cb.openedAt = time.Now()
	}
}

// IsOpen returns true if the circuit is open or needs to stay open.
// Handles automatic transition to half-open after timeout.
func (cb *SimpleCircuitBreaker) IsOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == CircuitOpen {
		// Check if we should transition to half-open
		if time.Since(cb.openedAt) >= cb.config.OpenTimeout {
			cb.state = CircuitHalfOpen
			cb.halfOpenSuccesses = 0
			return false // Allow probe request
		}
		return true
	}

	return false
}

// State returns the current circuit state.
func (cb *SimpleCircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// NoOpCircuitBreaker is a no-op implementation that never blocks.
// Used for testing or when circuit breaker is disabled.
type NoOpCircuitBreaker struct{}

// RecordSuccess does nothing.
func (cb *NoOpCircuitBreaker) RecordSuccess() {}

// RecordFailure does nothing.
func (cb *NoOpCircuitBreaker) RecordFailure() {}

// IsOpen always returns false.
func (cb *NoOpCircuitBreaker) IsOpen() bool { return false }

// State always returns CircuitClosed.
func (cb *NoOpCircuitBreaker) State() CircuitState { return CircuitClosed }
