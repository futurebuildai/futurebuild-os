package readiness

import (
	"context"
	"fmt"
	"time"

	"github.com/resend/resend-go/v2"
)

// ResendProbe verifies Resend API key by listing API keys.
type ResendProbe struct {
	apiKey string
}

// NewResendProbe creates a probe that calls Resend's ApiKeys.List endpoint.
func NewResendProbe(apiKey string) *ResendProbe {
	return &ResendProbe{apiKey: apiKey}
}

func (p *ResendProbe) Name() string { return "resend" }

func (p *ResendProbe) Check(ctx context.Context) CheckResult {
	start := time.Now()

	if p.apiKey == "" {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusNotConfigured,
			Message:  "RESEND_API_KEY not set",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	client := resend.NewClient(p.apiKey)
	_, err := client.ApiKeys.List()
	if err != nil {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  fmt.Sprintf("API key validation failed: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return CheckResult{
		Name:     p.Name(),
		Status:   StatusHealthy,
		Duration: time.Since(start).Milliseconds(),
	}
}
