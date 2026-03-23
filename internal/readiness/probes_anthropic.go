package readiness

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// AnthropicProbe verifies Anthropic API connectivity by checking the API key format
// and making a lightweight HEAD request to the API endpoint.
type AnthropicProbe struct {
	apiKey string
}

// NewAnthropicProbe creates a probe that validates the Anthropic API key and endpoint.
func NewAnthropicProbe(apiKey string) *AnthropicProbe {
	return &AnthropicProbe{apiKey: apiKey}
}

func (p *AnthropicProbe) Name() string { return "anthropic" }

func (p *AnthropicProbe) Check(ctx context.Context) CheckResult {
	start := time.Now()

	if p.apiKey == "" {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusNotConfigured,
			Message:  "ANTHROPIC_API_KEY not set",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	// Validate key format (should start with sk-ant-)
	if !strings.HasPrefix(p.apiKey, "sk-ant-") {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusDegraded,
			Message:  "API key format unexpected (should start with sk-ant-)",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	// Lightweight connectivity check: POST to messages endpoint with empty body
	// This will return a 400 (bad request) if reachable, which proves connectivity.
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", nil)
	if err != nil {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  fmt.Sprintf("request creation failed: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  fmt.Sprintf("API unreachable: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	defer resp.Body.Close()

	// 400 = reachable but bad request (expected for empty body)
	// 401 = reachable but invalid key
	// 200/other = reachable
	if resp.StatusCode == 401 {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  "API key invalid (401)",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return CheckResult{
		Name:     p.Name(),
		Status:   StatusHealthy,
		Duration: time.Since(start).Milliseconds(),
	}
}
