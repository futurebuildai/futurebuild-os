package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/chat"
	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/server"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noOpClient implements ai.Client for testing without real AI calls.
// L7 Vendor Abstraction: Updated to use vendor-agnostic types.
type noOpClient struct{}

func (m *noOpClient) GenerateContent(ctx context.Context, req ai.GenerateRequest) (ai.GenerateResponse, error) {
	// Return a mock response for chat tests
	return ai.GenerateResponse{
		Text: `{"reply": "Mock reply for testing", "intent": "process_invoice"}`,
	}, nil
}
func (m *noOpClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, nil
}
func (m *noOpClient) Close() error { return nil }

// getTestConfig creates configuration suitable for integration tests.
// Uses defaults if environment variables are not set.
func getTestConfig() *config.Config {
	cfg, err := config.LoadConfig()
	if err == nil {
		return cfg
	}

	// Create a test config with sensible defaults
	return &config.Config{
		DatabaseURL:   "postgres://fb_user:fb_pass@localhost:5433/futurebuild?sslmode=disable",
		RedisURL:      "localhost:6379",
		AppPort:       8080,
		JWTSecret:     "test-secret-for-integration-tests",
		JWTExpiry:     24 * time.Hour,
		WebhookSecret: "test-webhook-secret",
	}
}

func TestChat_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 1. Setup DB Connection
	cfg := getTestConfig()
	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	// Verify database is reachable
	if err := db.Ping(ctx); err != nil {
		t.Skipf("Skipping test: database not reachable: %v", err)
	}

	// 2. Create Test Fixtures (Org, User, Project)
	orgID := uuid.New()
	orgSlug := fmt.Sprintf("chat-test-org-%s", uuid.New().String()[:8])
	_, err = db.Exec(ctx, "INSERT INTO organizations (id, name, slug) VALUES ($1, 'Chat Test Org', $2)", orgID, orgSlug)
	require.NoError(t, err)
	defer func() {
		_, _ = db.Exec(ctx, "DELETE FROM organizations WHERE id = $1", orgID)
	}()

	userID := uuid.New()
	_, err = db.Exec(ctx, "INSERT INTO users (id, org_id, email, name, role) VALUES ($1, $2, $3, 'Test User', 'Builder')",
		userID, orgID, fmt.Sprintf("test-%s@example.com", userID.String()[:8]))
	require.NoError(t, err)

	projectID := uuid.New()
	_, err = db.Exec(ctx, "INSERT INTO projects (id, org_id, name, status) VALUES ($1, $2, 'Chat Test Project', 'Active')", projectID, orgID)
	require.NoError(t, err)

	// 3. Generate JWT
	claims := &types.Claims{
		UserID: userID.String(),
		OrgID:  orgID.String(),
		Role:   types.UserRoleBuilder,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(cfg.JWTSecret))
	require.NoError(t, err)

	// 4. Start Test Server
	// We use a real Server instance to test the full router and auth middleware.
	// Since we aren't calling Gemini, a no-op client is fine for the server setup.
	s := server.NewServer(db, cfg, &noOpClient{})
	ts := httptest.NewServer(s.Router)
	defer ts.Close()

	// 5. Execute Request
	chatReq := chat.ChatRequest{
		ProjectID: projectID,
		Message:   "Analyze this invoice",
	}
	body, err := json.Marshal(chatReq)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/chat", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+signedToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// 6. Assertions: HTTP Status & Body
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var chatResp chat.ChatResponse
	err = json.NewDecoder(resp.Body).Decode(&chatResp)
	require.NoError(t, err)
	assert.NotEmpty(t, chatResp.Reply)
	assert.Equal(t, types.IntentProcessInvoice, chatResp.Intent)

	// 7. Assertions: Database State
	// Verify that both the User message and Model reply were saved.
	var count int
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM chat_messages WHERE project_id = $1", projectID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count, "Should have 2 messages (User + Model)")

	// Check User message
	var userContent string
	var userRole types.ChatRole
	err = db.QueryRow(ctx, "SELECT content, role FROM chat_messages WHERE project_id = $1 AND role = 'user' LIMIT 1", projectID).Scan(&userContent, &userRole)
	require.NoError(t, err)
	assert.Equal(t, "Analyze this invoice", userContent)
	assert.Equal(t, types.ChatRoleUser, userRole)

	// Check Model message
	var modelContent string
	var modelRole types.ChatRole
	err = db.QueryRow(ctx, "SELECT content, role FROM chat_messages WHERE project_id = $1 AND role = 'model' LIMIT 1", projectID).Scan(&modelContent, &modelRole)
	require.NoError(t, err)
	assert.Equal(t, chatResp.Reply, modelContent)
	assert.Equal(t, types.ChatRoleModel, modelRole)
}

// CTO-003 Remediation: Negative Auth Test - No Token
func TestChat_NoToken_Unauthorized(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	cfg := getTestConfig()
	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Skipf("Skipping test: database not reachable: %v", err)
	}

	s := server.NewServer(db, cfg, &noOpClient{})
	ts := httptest.NewServer(s.Router)
	defer ts.Close()

	// Request WITHOUT Authorization header
	chatReq := chat.ChatRequest{
		ProjectID: uuid.New(),
		Message:   "Test message",
	}
	body, err := json.Marshal(chatReq)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/chat", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	// Deliberately NOT setting Authorization header

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// CTO-003 Remediation: Negative Auth Test - Invalid Token
func TestChat_InvalidToken_Unauthorized(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	cfg := getTestConfig()
	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Skipf("Skipping test: database not reachable: %v", err)
	}

	s := server.NewServer(db, cfg, &noOpClient{})
	ts := httptest.NewServer(s.Router)
	defer ts.Close()

	// Request with GARBAGE token
	chatReq := chat.ChatRequest{
		ProjectID: uuid.New(),
		Message:   "Test message",
	}
	body, err := json.Marshal(chatReq)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/chat", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer totally-invalid-token-garbage")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// CTO-003 Remediation: Negative Auth Test - Expired Token
func TestChat_ExpiredToken_Unauthorized(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	cfg := getTestConfig()
	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Skipf("Skipping test: database not reachable: %v", err)
	}

	s := server.NewServer(db, cfg, &noOpClient{})
	ts := httptest.NewServer(s.Router)
	defer ts.Close()

	// Generate EXPIRED JWT
	claims := &types.Claims{
		UserID: uuid.New().String(),
		OrgID:  uuid.New().String(),
		Role:   types.UserRoleBuilder,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-24 * time.Hour)), // EXPIRED
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(cfg.JWTSecret))
	require.NoError(t, err)

	// Request with expired token
	chatReq := chat.ChatRequest{
		ProjectID: uuid.New(),
		Message:   "Test message",
	}
	body, err := json.Marshal(chatReq)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/chat", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+signedToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
