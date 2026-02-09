package readiness

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// BirdProbe verifies Bird (MessageBird) credentials by fetching the account balance.
type BirdProbe struct {
	accessKey string
}

// NewBirdProbe creates a probe that fetches GET /balance to verify credentials.
func NewBirdProbe(accessKey string) *BirdProbe {
	return &BirdProbe{accessKey: accessKey}
}

func (p *BirdProbe) Name() string { return "bird" }

func (p *BirdProbe) Check(ctx context.Context) CheckResult {
	start := time.Now()

	if p.accessKey == "" {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusNotConfigured,
			Message:  "BIRD_ACCESS_KEY not set",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	url := "https://rest.messagebird.com/balance"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  fmt.Sprintf("failed to build request: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	req.Header.Set("Authorization", "AccessKey "+p.accessKey)

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

	if resp.StatusCode == http.StatusUnauthorized {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  "authentication failed: invalid access key",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	if resp.StatusCode != http.StatusOK {
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
