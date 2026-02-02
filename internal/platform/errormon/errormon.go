// Package errormon provides structured error reporting for the FutureBuild API.
// Currently uses slog for structured logging. Swap the Reporter implementation
// to integrate Sentry, Datadog, or another provider in production.
// See L7 Gate Audit Item 15: Error Monitoring.
package errormon

import (
	"context"
	"log/slog"
	"sync"
)

// Reporter defines the error reporting interface.
// Implementations must be safe for concurrent use.
type Reporter interface {
	// CaptureError reports a non-fatal error with optional key-value context.
	CaptureError(ctx context.Context, err error, kv ...any)
	// CaptureCritical reports a critical error that requires immediate attention.
	CaptureCritical(ctx context.Context, err error, kv ...any)
	// Flush ensures all pending reports are sent. Call before shutdown.
	Flush()
}

// slogReporter is the default Reporter that emits structured slog output.
// Replace with Sentry/Datadog SDK when external monitoring is provisioned.
type slogReporter struct {
	logger *slog.Logger
}

func (r *slogReporter) CaptureError(ctx context.Context, err error, kv ...any) {
	attrs := append([]any{"error", err.Error()}, kv...)
	r.logger.ErrorContext(ctx, "error_captured", attrs...)
}

func (r *slogReporter) CaptureCritical(ctx context.Context, err error, kv ...any) {
	attrs := append([]any{"error", err.Error(), "severity", "CRITICAL"}, kv...)
	r.logger.ErrorContext(ctx, "critical_error_captured", attrs...)
}

func (r *slogReporter) Flush() {
	// slog is synchronous — no-op.
}

// global is the package-level reporter, initialized once.
var (
	global     Reporter = &slogReporter{logger: slog.Default()}
	globalOnce sync.Once
)

// Init sets the global reporter. Must be called once at startup.
// Subsequent calls are no-ops (fail-safe).
func Init(r Reporter) {
	globalOnce.Do(func() {
		global = r
	})
}

// Get returns the global Reporter instance.
func Get() Reporter {
	return global
}

// NewSlogReporter creates a Reporter backed by the given slog.Logger.
func NewSlogReporter(logger *slog.Logger) Reporter {
	if logger == nil {
		logger = slog.Default()
	}
	return &slogReporter{logger: logger}
}
