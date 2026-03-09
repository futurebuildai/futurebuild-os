//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/colton/futurebuild/internal/api/handlers"
	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/colton/futurebuild/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAIClient for integration tests
type MockAIClient struct {
	GenerateContentFunc func(ctx context.Context, req ai.GenerateRequest) (ai.GenerateResponse, error)
}

func (m *MockAIClient) GenerateContent(ctx context.Context, req ai.GenerateRequest) (ai.GenerateResponse, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(ctx, req)
	}
	// Default: return empty extraction in parseUserMessage format
	return ai.GenerateResponse{
		Text: `{
			"values": {},
			"confidence": {}
		}`,
	}, nil
}

func (m *MockAIClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, nil
}

func (m *MockAIClient) Close() error {
	return nil
}

// Test Suite 1: API Contract

func TestOnboardEndpoint_ReturnsValidJSON(t *testing.T) {
	mockClient := &MockAIClient{
		GenerateContentFunc: func(ctx context.Context, req ai.GenerateRequest) (ai.GenerateResponse, error) {
			return ai.GenerateResponse{
				Text: `{
					"name": "Smith Residence",
					"address": "123 Main St, Austin, TX",
					"gsf": 3200,
					"foundation_type": "slab",
					"stories": 2,
					"bedrooms": 4,
					"bathrooms": 3,
					"confidence": {
						"name": 0.95,
						"address": 0.90,
						"gsf": 0.85
					}
				}`,
			}, nil
		},
	}

	interrogatorService := service.NewInterrogatorService(mockClient, nil)
	handler := handlers.NewOnboardingHandler(interrogatorService)

	reqBody := models.OnboardRequest{
		SessionID:    "test_session_1",
		Message:      "3200 sqft home in Austin",
		CurrentState: map[string]any{},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/onboard", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Simulate auth middleware setting context values
	claims := &types.Claims{
		UserID: "user_123",
		OrgID:  "tenant_456",
		Role:   types.UserRoleBuilder,
	}
	req = req.WithContext(middleware.WithClaims(req.Context(), claims))

	rr := httptest.NewRecorder()
	handler.HandleOnboard(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var resp models.OnboardResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "test_session_1", resp.SessionID)
	assert.NotEmpty(t, resp.Reply)
	assert.NotNil(t, resp.ExtractedValues)
	assert.NotNil(t, resp.ConfidenceScores)
}

func TestOnboardEndpoint_RequiresAuth(t *testing.T) {
	mockClient := &MockAIClient{}
	interrogatorService := service.NewInterrogatorService(mockClient, nil)
	handler := handlers.NewOnboardingHandler(interrogatorService)

	reqBody := models.OnboardRequest{
		SessionID:    "test_session_1",
		Message:      "test",
		CurrentState: map[string]any{},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// No auth context values
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/onboard", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.HandleOnboard(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestOnboardEndpoint_ExtractsFromMessage(t *testing.T) {
	mockClient := &MockAIClient{
		GenerateContentFunc: func(ctx context.Context, req ai.GenerateRequest) (ai.GenerateResponse, error) {
			// Simulate extraction from user message - parseUserMessage format
			return ai.GenerateResponse{
				Text: `{
					"values": {
						"address": "Austin, TX",
						"gsf": 3200,
						"foundation_type": "slab"
					},
					"confidence": {
						"address": 0.90,
						"gsf": 0.85,
						"foundation_type": 0.80
					}
				}`,
			}, nil
		},
	}

	interrogatorService := service.NewInterrogatorService(mockClient, nil)
	handler := handlers.NewOnboardingHandler(interrogatorService)

	reqBody := models.OnboardRequest{
		SessionID:    "test_session_2",
		Message:      "3200 sqft home in Austin with slab foundation",
		CurrentState: map[string]any{},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/onboard", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	claims := &types.Claims{UserID: "user_123", OrgID: "tenant_456", Role: types.UserRoleBuilder}
	req = req.WithContext(middleware.WithClaims(req.Context(), claims))

	rr := httptest.NewRecorder()
	handler.HandleOnboard(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.OnboardResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	// Verify extraction happened
	assert.Contains(t, resp.ExtractedValues, "address")
	assert.Contains(t, resp.ExtractedValues, "gsf")
	assert.Contains(t, resp.ExtractedValues, "foundation_type")

	// Name missing, so not ready to create
	assert.False(t, resp.ReadyToCreate)
}

func TestOnboardEndpoint_ProgressesToReady(t *testing.T) {
	// Turn 1: User provides partial info (missing name)
	mockClient1 := &MockAIClient{
		GenerateContentFunc: func(ctx context.Context, req ai.GenerateRequest) (ai.GenerateResponse, error) {
			return ai.GenerateResponse{
				Text: `{
					"values": {
						"address": "Austin, TX",
						"gsf": 3200,
						"foundation_type": "slab"
					},
					"confidence": {
						"address": 0.90,
						"gsf": 0.85
					}
				}`,
			}, nil
		},
	}

	interrogatorService1 := service.NewInterrogatorService(mockClient1, nil)
	handler1 := handlers.NewOnboardingHandler(interrogatorService1)

	reqBody1 := models.OnboardRequest{
		SessionID:    "test_session_3",
		Message:      "3200 sqft home in Austin",
		CurrentState: map[string]any{},
	}
	bodyBytes1, _ := json.Marshal(reqBody1)

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/agent/onboard", bytes.NewReader(bodyBytes1))
	req1.Header.Set("Content-Type", "application/json")
	claims := &types.Claims{UserID: "user_123", OrgID: "tenant_456", Role: types.UserRoleBuilder}
	req1 = req1.WithContext(middleware.WithClaims(req1.Context(), claims))

	rr1 := httptest.NewRecorder()
	handler1.HandleOnboard(rr1, req1)

	var resp1 models.OnboardResponse
	json.NewDecoder(rr1.Body).Decode(&resp1)

	assert.False(t, resp1.ReadyToCreate, "Should not be ready without name")
	assert.Equal(t, "name", resp1.NextPriorityField)

	// Turn 2: User provides name
	mockClient2 := &MockAIClient{
		GenerateContentFunc: func(ctx context.Context, req ai.GenerateRequest) (ai.GenerateResponse, error) {
			return ai.GenerateResponse{
				Text: `{
					"values": {
						"name": "Smith Residence"
					},
					"confidence": {
						"name": 0.95
					}
				}`,
			}, nil
		},
	}

	interrogatorService2 := service.NewInterrogatorService(mockClient2, nil)
	handler2 := handlers.NewOnboardingHandler(interrogatorService2)

	// Current state includes extracted values from turn 1
	reqBody2 := models.OnboardRequest{
		SessionID: "test_session_3",
		Message:   "Smith Residence",
		CurrentState: map[string]any{
			"address":         "Austin, TX",
			"gsf":             3200.0,
			"foundation_type": "slab",
		},
	}
	bodyBytes2, _ := json.Marshal(reqBody2)

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/agent/onboard", bytes.NewReader(bodyBytes2))
	req2.Header.Set("Content-Type", "application/json")
	claims2 := &types.Claims{UserID: "user_123", OrgID: "tenant_456", Role: types.UserRoleBuilder}
	req2 = req2.WithContext(middleware.WithClaims(req2.Context(), claims2))

	rr2 := httptest.NewRecorder()
	handler2.HandleOnboard(rr2, req2)

	var resp2 models.OnboardResponse
	json.NewDecoder(rr2.Body).Decode(&resp2)

	assert.True(t, resp2.ReadyToCreate, "Should be ready with name + address")
}

// Test Suite 2: Security Integration

func TestOnboardEndpoint_RejectsOversizedBody(t *testing.T) {
	mockClient := &MockAIClient{}
	interrogatorService := service.NewInterrogatorService(mockClient, nil)
	handler := handlers.NewOnboardingHandler(interrogatorService)

	// Create a 2MB request body (exceeds 1MB limit)
	largeMessage := strings.Repeat("a", 2*1024*1024)
	reqBody := models.OnboardRequest{
		SessionID:    "test",
		Message:      largeMessage,
		CurrentState: map[string]any{},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/onboard", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	claims := &types.Claims{UserID: "user_123", OrgID: "tenant_456", Role: types.UserRoleBuilder}
	req = req.WithContext(middleware.WithClaims(req.Context(), claims))

	rr := httptest.NewRecorder()
	handler.HandleOnboard(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOnboardEndpoint_RejectsInvalidSessionID(t *testing.T) {
	mockClient := &MockAIClient{}
	interrogatorService := service.NewInterrogatorService(mockClient, nil)
	handler := handlers.NewOnboardingHandler(interrogatorService)

	testCases := []struct {
		name      string
		sessionID string
		wantError bool
	}{
		{"empty session_id", "", true},
		{"too long session_id", strings.Repeat("a", 101), true},
		{"valid session_id", "valid_session_123", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := models.OnboardRequest{
				SessionID:    tc.sessionID,
				Message:      "test",
				CurrentState: map[string]any{},
			}
			bodyBytes, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/onboard", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			claims := &types.Claims{UserID: "user_123", OrgID: "tenant_456", Role: types.UserRoleBuilder}
			req = req.WithContext(middleware.WithClaims(req.Context(), claims))

			rr := httptest.NewRecorder()
			handler.HandleOnboard(rr, req)

			if tc.wantError {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
			} else {
				assert.Equal(t, http.StatusOK, rr.Code)
			}
		})
	}
}

func TestOnboardEndpoint_RejectsInvalidDocumentURL(t *testing.T) {
	mockClient := &MockAIClient{}
	interrogatorService := service.NewInterrogatorService(mockClient, nil)
	handler := handlers.NewOnboardingHandler(interrogatorService)

	testCases := []struct {
		name              string
		documentURL       string
		wantValidationErr bool // Caught by input validation (400)
		wantSSRFBlock     bool // Caught by SSRF protection (200 with error reply)
	}{
		{"valid http URL", "http://example.com/blueprint.jpg", false, false},
		{"valid https URL", "https://example.com/blueprint.jpg", false, false},
		{"invalid scheme", "file:///etc/passwd", true, false},                          // Input validation blocks file://
		{"too long URL", "http://" + strings.Repeat("a", 2000) + ".com", true, false},  // Input validation catches length
		{"malformed URL", "not a url", true, false},                                    // Input validation catches
		{"aws metadata ssrf", "http://169.254.169.254/latest/meta-data/", false, true}, // SSRF protection blocks local IP
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := models.OnboardRequest{
				SessionID:    "test",
				DocumentURL:  tc.documentURL,
				CurrentState: map[string]any{},
			}
			bodyBytes, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/onboard", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			claims := &types.Claims{UserID: "user_123", OrgID: "tenant_456", Role: types.UserRoleBuilder}
			req = req.WithContext(middleware.WithClaims(req.Context(), claims))

			rr := httptest.NewRecorder()
			handler.HandleOnboard(rr, req)

			if tc.wantValidationErr {
				// Input validation layer returns 400
				assert.Equal(t, http.StatusBadRequest, rr.Code)
			} else if tc.wantSSRFBlock {
				// SSRF protection gracefully handles error, returns 200 with user-friendly message
				assert.Equal(t, http.StatusOK, rr.Code)

				var resp models.OnboardResponse
				err := json.NewDecoder(rr.Body).Decode(&resp)
				require.NoError(t, err)

				// Should contain friendly error message
				// Should contain friendly error message or graceful degradation mode
				assert.Condition(t, func() bool {
					return strings.Contains(resp.Reply, "couldn't read") || strings.Contains(resp.Reply, "manual_mode")
				}, "Reply should indicate error or fallback mode, got: "+resp.Reply)
			} else {
				// Valid URL (may fail to fetch, but that's OK for this test)
				assert.Equal(t, http.StatusOK, rr.Code)
			}
		})
	}
}

func TestOnboardEndpoint_RejectsTooManyStateFields(t *testing.T) {
	mockClient := &MockAIClient{}
	interrogatorService := service.NewInterrogatorService(mockClient, nil)
	handler := handlers.NewOnboardingHandler(interrogatorService)

	// Create state with 51 fields (exceeds 50 limit)
	largeState := make(map[string]any)
	for i := 0; i < 51; i++ {
		largeState[string(rune('a'+i))] = "value"
	}

	reqBody := models.OnboardRequest{
		SessionID:    "test",
		Message:      "test",
		CurrentState: largeState,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/onboard", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	claims := &types.Claims{UserID: "user_123", OrgID: "tenant_456", Role: types.UserRoleBuilder}
	req = req.WithContext(middleware.WithClaims(req.Context(), claims))

	rr := httptest.NewRecorder()
	handler.HandleOnboard(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOnboardEndpoint_RejectsNilCurrentState(t *testing.T) {
	mockClient := &MockAIClient{}
	interrogatorService := service.NewInterrogatorService(mockClient, nil)
	handler := handlers.NewOnboardingHandler(interrogatorService)

	reqBody := map[string]any{
		"session_id":    "test",
		"message":       "test",
		"current_state": nil,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/onboard", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	claims := &types.Claims{UserID: "user_123", OrgID: "tenant_456", Role: types.UserRoleBuilder}
	req = req.WithContext(middleware.WithClaims(req.Context(), claims))

	rr := httptest.NewRecorder()
	handler.HandleOnboard(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// Test Suite 3: End-to-End Flow

func TestOnboardingFlow_CompleteWizard(t *testing.T) {
	// This test simulates a complete onboarding flow from start to finish

	// Step 1: Initial message with partial info
	t.Log("Step 1: User provides initial project details")

	mockClient := &MockAIClient{
		GenerateContentFunc: func(ctx context.Context, req ai.GenerateRequest) (ai.GenerateResponse, error) {
			// Extract address and GSF from message
			return ai.GenerateResponse{
				Text: `{
					"values": {
						"address": "123 Main St, Austin, TX",
						"gsf": 3200
					},
					"confidence": {
						"address": 0.92,
						"gsf": 0.88
					}
				}`,
			}, nil
		},
	}

	interrogatorService := service.NewInterrogatorService(mockClient, nil)
	handler := handlers.NewOnboardingHandler(interrogatorService)

	reqBody := models.OnboardRequest{
		SessionID:    "wizard_test",
		Message:      "I'm building a 3200 sqft home at 123 Main St in Austin",
		CurrentState: map[string]any{},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/onboard", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	claims := &types.Claims{UserID: "user_123", OrgID: "tenant_456", Role: types.UserRoleBuilder}
	req = req.WithContext(middleware.WithClaims(req.Context(), claims))

	rr := httptest.NewRecorder()
	handler.HandleOnboard(rr, req)

	var resp1 models.OnboardResponse
	json.NewDecoder(rr.Body).Decode(&resp1)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.False(t, resp1.ReadyToCreate, "Missing name, should not be ready")
	assert.Equal(t, "name", resp1.NextPriorityField)
	assert.Contains(t, resp1.ExtractedValues, "address")
	assert.Contains(t, resp1.ExtractedValues, "gsf")

	t.Log("Step 2: User provides project name")

	// Step 2: User provides name
	mockClient2 := &MockAIClient{
		GenerateContentFunc: func(ctx context.Context, req ai.GenerateRequest) (ai.GenerateResponse, error) {
			return ai.GenerateResponse{
				Text: `{
					"values": {
						"name": "Smith Residence"
					},
					"confidence": {
						"name": 0.98
					}
				}`,
			}, nil
		},
	}

	interrogatorService2 := service.NewInterrogatorService(mockClient2, nil)
	handler2 := handlers.NewOnboardingHandler(interrogatorService2)

	reqBody2 := models.OnboardRequest{
		SessionID: "wizard_test",
		Message:   "Smith Residence",
		CurrentState: map[string]any{
			"address": "123 Main St, Austin, TX",
			"gsf":     3200.0,
		},
	}
	bodyBytes2, _ := json.Marshal(reqBody2)

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/agent/onboard", bytes.NewReader(bodyBytes2))
	req2.Header.Set("Content-Type", "application/json")
	claims2 := &types.Claims{UserID: "user_123", OrgID: "tenant_456", Role: types.UserRoleBuilder}
	req2 = req2.WithContext(middleware.WithClaims(req2.Context(), claims2))

	rr2 := httptest.NewRecorder()
	handler2.HandleOnboard(rr2, req2)

	var resp2 models.OnboardResponse
	json.NewDecoder(rr2.Body).Decode(&resp2)

	assert.Equal(t, http.StatusOK, rr2.Code)
	assert.True(t, resp2.ReadyToCreate, "Has name + address, should be ready")

	// Verify all P0 fields are populated
	finalState := make(map[string]any)
	for k, v := range reqBody2.CurrentState {
		finalState[k] = v
	}
	for k, v := range resp2.ExtractedValues {
		finalState[k] = v
	}

	assert.Contains(t, finalState, "name")
	assert.Contains(t, finalState, "address")

	t.Log("Onboarding wizard completed successfully")
}
