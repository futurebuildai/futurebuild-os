// Package readiness provides integration health probes for third-party services.
// Each probe makes a single lightweight, non-destructive API call to verify connectivity.
// See PRODUCT_VISION.md: Integration Readiness Check System.
package readiness

import "context"

// Status represents the health state of an integration probe.
type Status string

const (
	StatusHealthy       Status = "healthy"
	StatusDegraded      Status = "degraded"
	StatusFailed        Status = "failed"
	StatusNotConfigured Status = "not_configured"
)

// CheckResult is the outcome of a single probe execution.
type CheckResult struct {
	Name     string `json:"name"`
	Status   Status `json:"status"`
	Message  string `json:"message,omitempty"`
	Duration int64  `json:"duration_ms"`
}

// Checker is the interface that all probes implement.
type Checker interface {
	Name() string
	Check(ctx context.Context) CheckResult
}

// Report is the aggregated output of all probe results.
type Report struct {
	Status      Status        `json:"status"`
	Environment string        `json:"environment"`
	Timestamp   string        `json:"timestamp"`
	Duration    int64         `json:"total_duration_ms"`
	Checks      []CheckResult `json:"checks"`
}
