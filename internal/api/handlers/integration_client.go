package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/colton/futurebuild/pkg/a2a"
)

// IntegrationClient is a lightweight HTTP client for communicating with FB-Brain.
type IntegrationClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewIntegrationClient() *IntegrationClient {
	baseURL := os.Getenv("FB_BRAIN_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8082"
	}
	apiKey := os.Getenv("INTEGRATION_API_KEY")
	if apiKey == "" {
		apiKey = "fb-brain-demo-key-2026"
	}
	return &IntegrationClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fb-brain request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fb-brain error %d: %s", resp.StatusCode, string(respBody))
	}

	var flowResp FlowResponse
	if err := json.NewDecoder(resp.Body).Decode(&flowResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &flowResp, nil
}
