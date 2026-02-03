package readiness

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// TwilioProbe verifies Twilio credentials by fetching the account resource.
type TwilioProbe struct {
	accountSID string
	authToken  string
}

// NewTwilioProbe creates a probe that fetches /2010-04-01/Accounts/{sid}.json.
func NewTwilioProbe(accountSID, authToken string) *TwilioProbe {
	return &TwilioProbe{accountSID: accountSID, authToken: authToken}
}

func (p *TwilioProbe) Name() string { return "twilio" }

func (p *TwilioProbe) Check(ctx context.Context) CheckResult {
	start := time.Now()

	if p.accountSID == "" || p.authToken == "" {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusNotConfigured,
			Message:  "TWILIO_ACCOUNT_SID or TWILIO_AUTH_TOKEN not set",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	url := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s.json", p.accountSID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  fmt.Sprintf("failed to build request: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}
	req.SetBasicAuth(p.accountSID, p.authToken)

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
			Message:  "authentication failed: invalid credentials",
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
