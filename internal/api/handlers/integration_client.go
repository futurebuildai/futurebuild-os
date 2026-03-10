package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

func (c *IntegrationClient) StartMaterialsFlow(orgID, projectID, cardID string) (*FlowResponse, error) {
	return c.postFlow("/api/flows/materials/start", map[string]string{
		"org_id":     orgID,
		"project_id": projectID,
		"card_id":    cardID,
	})
}

func (c *IntegrationClient) ApproveMaterialsQuote(rfqID string) (*FlowResponse, error) {
	return c.postFlow(fmt.Sprintf("/api/flows/materials/%s/approve", rfqID), nil)
}

func (c *IntegrationClient) StartLaborFlow(orgID, projectID, cardID string) (*FlowResponse, error) {
	return c.postFlow("/api/flows/labor/start", map[string]string{
		"org_id":     orgID,
		"project_id": projectID,
		"card_id":    cardID,
	})
}

func (c *IntegrationClient) ApproveLaborBid(rfqID string) (*FlowResponse, error) {
	return c.postFlow(fmt.Sprintf("/api/flows/labor/%s/approve", rfqID), nil)
}

func (c *IntegrationClient) ConfirmDelivery(rfqID string) (*FlowResponse, error) {
	return c.postFlow(fmt.Sprintf("/api/flows/delivery/%s/confirm", rfqID), nil)
}

func (c *IntegrationClient) postFlow(path string, body interface{}) (*FlowResponse, error) {
	var reqBody io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest("POST", c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Integration-Key", c.apiKey)

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
