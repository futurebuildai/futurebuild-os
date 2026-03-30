package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/colton/futurebuild/internal/worker"
	"github.com/colton/futurebuild/pkg/a2a"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel/trace"
)

// IntegrationClient is a lightweight HTTP client for communicating with FB-Brain.
type IntegrationClient struct {
	baseURL     string
	apiKey      string
	httpClient  *http.Client
	asynqClient *asynq.Client // Phase 20: queue failed requests for retry
}

// NewIntegrationClient creates an IntegrationClient. If redisAddr is non-empty,
// failed requests are queued for async retry via asynq instead of being dropped.
func NewIntegrationClient(redisAddr string) *IntegrationClient {
	baseURL := os.Getenv("FB_BRAIN_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8082"
	}
	apiKey := os.Getenv("INTEGRATION_API_KEY")
	if apiKey == "" {
		apiKey = "fb-brain-demo-key-2026"
	}
	c := &IntegrationClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
	if redisAddr != "" {
		c.asynqClient = asynq.NewClient(worker.ParseRedisOpt(redisAddr))
	}
	return c
}

type FlowResponse struct {
	RFQID   string `json:"rfq_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (c *IntegrationClient) StartMaterialsFlow(ctx context.Context, orgID, projectID, cardID string) (*FlowResponse, error) {
	return c.postFlow(ctx, "/api/flows/materials/start", map[string]string{
		"org_id":     orgID,
		"project_id": projectID,
		"card_id":    cardID,
	})
}

func (c *IntegrationClient) ApproveMaterialsQuote(ctx context.Context, rfqID string) (*FlowResponse, error) {
	return c.postFlow(ctx, fmt.Sprintf("/api/flows/materials/%s/approve", rfqID), nil)
}

func (c *IntegrationClient) StartLaborFlow(ctx context.Context, orgID, projectID, cardID string) (*FlowResponse, error) {
	return c.postFlow(ctx, "/api/flows/labor/start", map[string]string{
		"org_id":     orgID,
		"project_id": projectID,
		"card_id":    cardID,
	})
}

func (c *IntegrationClient) ApproveLaborBid(ctx context.Context, rfqID string) (*FlowResponse, error) {
	return c.postFlow(ctx, fmt.Sprintf("/api/flows/labor/%s/approve", rfqID), nil)
}

func (c *IntegrationClient) ConfirmDelivery(ctx context.Context, rfqID string) (*FlowResponse, error) {
	return c.postFlow(ctx, fmt.Sprintf("/api/flows/delivery/%s/confirm", rfqID), nil)
}

func (c *IntegrationClient) postFlow(ctx context.Context, path string, body interface{}) (*FlowResponse, error) {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Legacy header (kept for backward compat during rollout)
	req.Header.Set("X-Integration-Key", c.apiKey)

	// HMAC signature headers
	ts := a2a.CurrentTimestamp()
	nonce := a2a.GenerateNonce()
	sig := a2a.SignRequest(c.apiKey, ts, nonce, bodyBytes)
	req.Header.Set("X-Signature", sig)
	req.Header.Set("X-Timestamp", ts)
	req.Header.Set("X-Nonce", nonce)

	// A2A distributed tracing: prefer W3C traceparent if OTel span active; fallback UUIDv4
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		sc := span.SpanContext()
		req.Header.Set("X-Trace-ID", fmt.Sprintf("00-%s-%s-01", sc.TraceID(), sc.SpanID()))
	} else {
		req.Header.Set("X-Trace-ID", uuid.NewString())
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Phase 20: Queue for async retry via asynq instead of dropping
		if c.asynqClient != nil {
			traceID := req.Header.Get("X-Trace-ID")
			task, qErr := worker.NewA2AWebhookDispatchTask(worker.A2AWebhookDispatchPayload{
				Path:    path,
				Body:    bodyBytes,
				APIKey:  c.apiKey,
				BaseURL: c.baseURL,
				TraceID: traceID,
			})
			if qErr == nil {
				_, qErr = c.asynqClient.Enqueue(task)
			}
			if qErr != nil {
				slog.Error("a2a/integration: failed to queue retry", "error", qErr, "path", path)
			} else {
				slog.Info("a2a/integration: queued for async retry", "path", path, "trace_id", traceID)
			}
		} else {
			slog.Warn("a2a/integration: no queue available, event dropped", "error", err, "path", path)
		}
		return &FlowResponse{
			Status:  "queued",
			Message: "A2A connection unavailable, event queued for retry",
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		slog.Warn("a2a/integration: fb-brain returned error", "status", resp.StatusCode, "body", string(respBody))
		return &FlowResponse{
			Status:  "error",
			Message: fmt.Sprintf("FB-Brain failed with status %d", resp.StatusCode),
		}, nil
	}

	var flowResp FlowResponse
	if err := json.NewDecoder(resp.Body).Decode(&flowResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &flowResp, nil
}
