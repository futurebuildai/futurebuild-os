package readiness

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// SendGridProbe verifies SendGrid API key by fetching /v3/scopes.
type SendGridProbe struct {
	apiKey string
}

// NewSendGridProbe creates a probe that calls SendGrid's scopes endpoint.
func NewSendGridProbe(apiKey string) *SendGridProbe {
	return &SendGridProbe{apiKey: apiKey}
}

func (p *SendGridProbe) Name() string { return "sendgrid" }

func (p *SendGridProbe) Check(ctx context.Context) CheckResult {
	start := time.Now()

	if p.apiKey == "" {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusNotConfigured,
			Message:  "SENDGRID_API_KEY not set",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.sendgrid.com/v3/scopes", nil)
	if err != nil {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  fmt.Sprintf("failed to build request: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  fmt.Sprintf("API request failed: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  fmt.Sprintf("authentication failed: HTTP %d", resp.StatusCode),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	if resp.StatusCode >= 400 {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  fmt.Sprintf("unexpected HTTP %d", resp.StatusCode),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return CheckResult{
		Name:     p.Name(),
		Status:   StatusHealthy,
		Duration: time.Since(start).Milliseconds(),
	}
}
