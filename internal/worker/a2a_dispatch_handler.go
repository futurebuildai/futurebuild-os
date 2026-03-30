package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/colton/futurebuild/pkg/a2a"
	"github.com/hibiken/asynq"
)

// a2aDispatchClient is a dedicated HTTP client for A2A webhook retries.
// Strict timeout prevents worker thread-pool exhaustion if FB-Brain hangs.
var a2aDispatchClient = &http.Client{Timeout: 10 * time.Second}

// HandleA2AWebhookDispatch retries a failed outbound A2A HTTP request.
// Asynq auto-retries on error per MaxRetry policy (5 retries).
// See Phase 20: A2A Full Integration Loop
func HandleA2AWebhookDispatch(ctx context.Context, t *asynq.Task) error {
	var p A2AWebhookDispatchPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL+p.Path, bytes.NewReader(p.Body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Re-sign with fresh HMAC headers
	ts := a2a.CurrentTimestamp()
	nonce := a2a.GenerateNonce()
	sig := a2a.SignRequest(p.APIKey, ts, nonce, p.Body)
	req.Header.Set("X-Signature", sig)
	req.Header.Set("X-Timestamp", ts)
	req.Header.Set("X-Nonce", nonce)
	req.Header.Set("X-Trace-ID", p.TraceID)

	resp, err := a2aDispatchClient.Do(req)
	if err != nil {
		slog.Warn("a2a/dispatch: retry failed", "path", p.Path, "trace_id", p.TraceID, "error", err)
		return err // asynq will retry
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("fb-brain 5xx: %d %s", resp.StatusCode, string(body))
	}

	slog.Info("a2a/dispatch: retry succeeded", "path", p.Path, "trace_id", p.TraceID, "status", resp.StatusCode)
	return nil
}
