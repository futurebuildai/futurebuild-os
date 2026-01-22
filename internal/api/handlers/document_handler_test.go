package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloud.google.com/go/vertexai/genai"
	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/server"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAIClient is a mock of the ai.Client interface.
type MockAIClient struct {
	mock.Mock
}

func (m *MockAIClient) GenerateContent(ctx context.Context, modelType ai.ModelType, parts ...genai.Part) (string, error) {
	args := m.Called(ctx, modelType, parts)
	return args.String(0), args.Error(1)
}

func (m *MockAIClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	args := m.Called(ctx, text)
	return args.Get(0).([]float32), args.Error(1)
}

func (m *MockAIClient) Close() error {
	return nil
}

func TestAnalyzeDocument(t *testing.T) {
	cfg := &config.Config{AppPort: 8080}
	orgID := uuid.New()
	docID := uuid.New()

	t.Run("Missing Auth", func(t *testing.T) {
		s := server.NewServer(nil, cfg, nil)
		ts := httptest.NewServer(s.Router)
		defer ts.Close()

		reqBody := map[string]uuid.UUID{"document_id": docID}
		body, _ := json.Marshal(reqBody)
		resp, err := http.Post(ts.URL+"/api/v1/documents/analyze", "application/json", bytes.NewBuffer(body))
		assert.NoError(t, err)
		// L7 Security Hardening: JWT auth required before body parsing
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Invalid Body after Auth fails", func(t *testing.T) {
		s := server.NewServer(nil, cfg, nil)
		ts := httptest.NewServer(s.Router)
		defer ts.Close()

		req, _ := http.NewRequest("POST", ts.URL+"/api/v1/documents/analyze", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("X-Org-ID", orgID.String()) // Header is now ignored; JWT required
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		// L7 Security Hardening: JWT auth required before body parsing
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	// Note: Fully testing Success path requires a DB mock or test DB setup
	// because AnalyzeInvoice queries the DB for doc text.
	// For now, these basic error path tests verify the handler is registered and doing initial validation.
}
