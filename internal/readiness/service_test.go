package readiness

import (
	"context"
	"testing"
	"time"
)

// mockChecker is a test double that returns a pre-configured CheckResult.
type mockChecker struct {
	name   string
	result CheckResult
}

func (m *mockChecker) Name() string                        { return m.name }
func (m *mockChecker) Check(_ context.Context) CheckResult { return m.result }

func newMock(name string, status Status) *mockChecker {
	return &mockChecker{
		name: name,
		result: CheckResult{
			Name:     name,
			Status:   status,
			Duration: 1,
		},
	}
}

func newMockWithMsg(name string, status Status, msg string) *mockChecker {
	return &mockChecker{
		name: name,
		result: CheckResult{
			Name:     name,
			Status:   status,
			Message:  msg,
			Duration: 1,
		},
	}
}

func TestService_AllHealthy(t *testing.T) {
	svc := NewService(5*time.Second,
		newMock("database", StatusHealthy),
		newMock("clerk", StatusHealthy),
		newMock("redis", StatusHealthy),
		newMock("resend", StatusHealthy),
		newMock("twilio", StatusHealthy),
		newMock("vertex_ai", StatusHealthy),
		newMock("s3", StatusHealthy),
	)

	report := svc.Run(context.Background(), "production")

	if report.Status != StatusHealthy {
		t.Errorf("expected healthy, got %s", report.Status)
	}
	if len(report.Checks) != 7 {
		t.Errorf("expected 7 checks, got %d", len(report.Checks))
	}
	if report.Environment != "production" {
		t.Errorf("expected production, got %s", report.Environment)
	}
}

func TestService_FailedInProduction(t *testing.T) {
	svc := NewService(5*time.Second,
		newMock("database", StatusHealthy),
		newMock("clerk", StatusHealthy),
		newMock("redis", StatusFailed),
		newMock("resend", StatusHealthy),
	)

	report := svc.Run(context.Background(), "production")

	if report.Status != StatusFailed {
		t.Errorf("expected failed, got %s", report.Status)
	}
}

func TestService_NotConfiguredCriticalInProd(t *testing.T) {
	svc := NewService(5*time.Second,
		newMock("database", StatusHealthy),
		newMock("clerk", StatusNotConfigured),
		newMock("redis", StatusHealthy),
		newMock("resend", StatusHealthy),
	)

	report := svc.Run(context.Background(), "production")

	if report.Status != StatusFailed {
		t.Errorf("expected failed (critical not_configured in prod), got %s", report.Status)
	}
}

func TestService_NotConfiguredCriticalInStaging(t *testing.T) {
	svc := NewService(5*time.Second,
		newMock("database", StatusHealthy),
		newMock("clerk", StatusNotConfigured),
		newMock("redis", StatusHealthy),
		newMock("resend", StatusHealthy),
	)

	report := svc.Run(context.Background(), "staging")

	if report.Status != StatusDegraded {
		t.Errorf("expected degraded (critical not_configured in staging), got %s", report.Status)
	}
}

func TestService_NotConfiguredInDev(t *testing.T) {
	svc := NewService(5*time.Second,
		newMock("database", StatusHealthy),
		newMock("clerk", StatusNotConfigured),
		newMock("redis", StatusNotConfigured),
		newMock("resend", StatusNotConfigured),
		newMock("sendgrid", StatusNotConfigured),
		newMock("twilio", StatusNotConfigured),
		newMock("vertex_ai", StatusNotConfigured),
		newMock("s3", StatusNotConfigured),
	)

	report := svc.Run(context.Background(), "development")

	if report.Status != StatusHealthy {
		t.Errorf("expected healthy (not_configured ok in dev), got %s", report.Status)
	}
}

func TestService_DegradableServiceFailed(t *testing.T) {
	// S3 and Twilio are degradable, not critical.
	svc := NewService(5*time.Second,
		newMock("database", StatusHealthy),
		newMock("clerk", StatusHealthy),
		newMock("redis", StatusHealthy),
		newMock("resend", StatusHealthy),
		newMock("vertex_ai", StatusHealthy),
		newMock("s3", StatusFailed),
		newMock("twilio", StatusFailed),
	)

	report := svc.Run(context.Background(), "production")

	if report.Status != StatusFailed {
		t.Errorf("expected failed (any failed in prod = overall failed), got %s", report.Status)
	}
}

func TestService_EmailProvidersFallback(t *testing.T) {
	// Resend not configured but SendGrid healthy — email requirement met.
	svc := NewService(5*time.Second,
		newMock("database", StatusHealthy),
		newMock("clerk", StatusHealthy),
		newMock("redis", StatusHealthy),
		newMock("resend", StatusNotConfigured),
		newMock("sendgrid", StatusHealthy),
		newMock("vertex_ai", StatusHealthy),
	)

	report := svc.Run(context.Background(), "production")

	if report.Status != StatusHealthy {
		t.Errorf("expected healthy (sendgrid covers email), got %s", report.Status)
	}
}

func TestService_NoEmailProviderInProd(t *testing.T) {
	// Neither email provider configured in production.
	svc := NewService(5*time.Second,
		newMock("database", StatusHealthy),
		newMock("clerk", StatusHealthy),
		newMock("redis", StatusHealthy),
		newMock("resend", StatusNotConfigured),
		newMock("sendgrid", StatusNotConfigured),
		newMock("vertex_ai", StatusHealthy),
	)

	report := svc.Run(context.Background(), "production")

	if report.Status != StatusFailed {
		t.Errorf("expected failed (no email in prod), got %s", report.Status)
	}
}

func TestService_FailedInDev(t *testing.T) {
	// In dev, a failed probe should degrade, not fail overall.
	svc := NewService(5*time.Second,
		newMock("database", StatusHealthy),
		newMock("redis", StatusFailed),
	)

	report := svc.Run(context.Background(), "development")

	if report.Status != StatusDegraded {
		t.Errorf("expected degraded (failed in dev = degraded), got %s", report.Status)
	}
}

func TestService_ReportTimestamp(t *testing.T) {
	svc := NewService(5*time.Second, newMock("database", StatusHealthy))
	report := svc.Run(context.Background(), "development")

	if report.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
	// Verify it parses as RFC3339.
	_, err := time.Parse(time.RFC3339, report.Timestamp)
	if err != nil {
		t.Errorf("expected RFC3339 timestamp, got %q: %v", report.Timestamp, err)
	}
}

func TestService_ConcurrentExecution(t *testing.T) {
	// Verify all probes run (none are silently dropped).
	svc := NewService(5*time.Second,
		newMockWithMsg("a", StatusHealthy, "probe-a"),
		newMockWithMsg("b", StatusHealthy, "probe-b"),
		newMockWithMsg("c", StatusHealthy, "probe-c"),
	)

	report := svc.Run(context.Background(), "development")

	if len(report.Checks) != 3 {
		t.Fatalf("expected 3 checks, got %d", len(report.Checks))
	}

	names := make(map[string]bool)
	for _, c := range report.Checks {
		names[c.Name] = true
	}
	for _, expected := range []string{"a", "b", "c"} {
		if !names[expected] {
			t.Errorf("missing check result for %q", expected)
		}
	}
}
