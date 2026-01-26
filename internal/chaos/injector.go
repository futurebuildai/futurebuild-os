// Package chaos provides fault injection for testing self-healing capabilities.
// Used by the Tree Planting integration test to validate FutureShade's
// autonomous diagnosis and remediation.
package chaos

import (
	"context"
	"fmt"
	"sync"
)

// FaultType defines the category of injected fault.
type FaultType string

const (
	// FaultNone indicates no fault is active.
	FaultNone FaultType = ""

	// FaultServiceError simulates a service-level error (e.g., external API failure).
	FaultServiceError FaultType = "SERVICE_ERROR"

	// FaultConfigDrift simulates configuration drift (e.g., missing/invalid setting).
	FaultConfigDrift FaultType = "CONFIG_DRIFT"

	// FaultDBExhausted simulates database connection pool exhaustion.
	FaultDBExhausted FaultType = "DB_EXHAUSTED"
)

// ChaosConfig defines the parameters for a fault injection.
type ChaosConfig struct {
	// TargetMethod is the method name to inject the fault into (e.g., "CreateProject").
	TargetMethod string `json:"target_method"`

	// ActiveFault is the type of fault to inject.
	ActiveFault FaultType `json:"active_fault"`

	// ErrorMessage is the custom error message to return when the fault triggers.
	ErrorMessage string `json:"error_message"`
}

// Injector defines the interface for chaos injection.
// Implementations must be thread-safe.
type Injector interface {
	// ShouldFail checks if the specified method should fail.
	// Returns true and an error if a fault is active for this method.
	ShouldFail(ctx context.Context, method string) (bool, error)

	// RegisterFault registers a fault configuration.
	RegisterFault(config ChaosConfig)

	// ClearFaults removes all registered faults.
	ClearFaults()

	// GetActiveFaults returns all currently registered fault configurations.
	GetActiveFaults() []ChaosConfig
}

// ChaosError is a custom error type that carries fault information.
// Allows downstream handlers to inspect the fault type.
type ChaosError struct {
	Fault   FaultType
	Message string
}

// Error implements the error interface.
func (e *ChaosError) Error() string {
	return fmt.Sprintf("[CHAOS:%s] %s", e.Fault, e.Message)
}

// MemoryInjector is a thread-safe in-memory implementation of Injector.
type MemoryInjector struct {
	mu     sync.RWMutex
	faults map[string]ChaosConfig // keyed by TargetMethod
}

// NewMemoryInjector creates a new MemoryInjector instance.
func NewMemoryInjector() *MemoryInjector {
	return &MemoryInjector{
		faults: make(map[string]ChaosConfig),
	}
}

// ShouldFail checks if the specified method should fail.
// Returns true and a ChaosError if a fault is active for this method.
func (m *MemoryInjector) ShouldFail(ctx context.Context, method string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, exists := m.faults[method]
	if !exists || config.ActiveFault == FaultNone {
		return false, nil
	}

	return true, &ChaosError{
		Fault:   config.ActiveFault,
		Message: config.ErrorMessage,
	}
}

// RegisterFault registers a fault configuration.
// If a fault already exists for the target method, it is replaced.
func (m *MemoryInjector) RegisterFault(config ChaosConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.faults[config.TargetMethod] = config
}

// ClearFaults removes all registered faults.
func (m *MemoryInjector) ClearFaults() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.faults = make(map[string]ChaosConfig)
}

// GetActiveFaults returns all currently registered fault configurations.
func (m *MemoryInjector) GetActiveFaults() []ChaosConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	configs := make([]ChaosConfig, 0, len(m.faults))
	for _, config := range m.faults {
		configs = append(configs, config)
	}
	return configs
}
