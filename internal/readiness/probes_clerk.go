package readiness

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ClerkProbe verifies Clerk OIDC by fetching the JWKS endpoint.
type ClerkProbe struct {
	issuerURL string
}

// NewClerkProbe creates a probe that fetches {issuerURL}/.well-known/jwks.json.
func NewClerkProbe(issuerURL string) *ClerkProbe {
	return &ClerkProbe{issuerURL: issuerURL}
}

func (p *ClerkProbe) Name() string { return "clerk" }

func (p *ClerkProbe) Check(ctx context.Context) CheckResult {
	start := time.Now()

	if p.issuerURL == "" {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusNotConfigured,
			Message:  "CLERK_ISSUER_URL not set",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	url := p.issuerURL + "/.well-known/jwks.json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  fmt.Sprintf("failed to build request: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  fmt.Sprintf("JWKS fetch failed: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  fmt.Sprintf("JWKS returned HTTP %d", resp.StatusCode),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	// Verify response is valid JSON with a "keys" array.
	var jwks struct {
		Keys []json.RawMessage `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  fmt.Sprintf("invalid JWKS JSON: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return CheckResult{
		Name:     p.Name(),
		Status:   StatusHealthy,
		Message:  fmt.Sprintf("%d keys loaded", len(jwks.Keys)),
		Duration: time.Since(start).Milliseconds(),
	}
}
