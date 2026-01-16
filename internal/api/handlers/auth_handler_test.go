package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type SpyNotificationService struct {
	LastEmailBody string
}

func (s *SpyNotificationService) SendSMS(contactID string, message string) error { return nil }
func (s *SpyNotificationService) SendEmail(to string, subject string, body string) error {
	s.LastEmailBody = body
	return nil
}

func TestAuthFlow_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://fb_user:fb_pass@localhost:5433/futurebuild?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer pool.Close()

	// Verify database is reachable
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("Skipping test: database not reachable: %v", err)
	}

	// Setup Test Data
	orgID := uuid.New()
	_, err = pool.Exec(ctx, "INSERT INTO organizations (id, name, slug) VALUES ($1, $2, $3)", orgID, "Test Org", "test-org-"+orgID.String()[:8])
	if err != nil {
		t.Fatalf("Failed to create test org: %v", err)
	}
	defer func() { _, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", orgID) }()

	userID := uuid.New()
	email := "test-" + userID.String()[:8] + "@example.com"
	_, err = pool.Exec(ctx, "INSERT INTO users (id, org_id, email, name, role) VALUES ($1, $2, $3, $4, $5)", userID, orgID, email, "Test User", "Builder")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Setup Handler
	cfg := &config.Config{
		JWTSecret: "test-secret",
		JWTExpiry: 1 * time.Hour,
	}
	authService := service.NewAuthService(pool, cfg)
	spyNotify := &SpyNotificationService{}
	handler := NewAuthHandler(authService, spyNotify, "http://localhost:8080")

	// 1. Request Login
	loginPayload, _ := json.Marshal(types.AuthRequest{Email: email})
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginPayload))
	rr := httptest.NewRecorder()
	handler.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Login failed: got status %d", rr.Code)
	}

	var loginResp types.AuthResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &loginResp))
	if loginResp.Message != "If this user exists, a login link has been sent." {
		t.Errorf("Unexpected login response message: %s", loginResp.Message)
	}

	// 2. Extract Token from "Email"
	if spyNotify.LastEmailBody == "" {
		t.Fatal("NotificationService did not receive email")
	}
	tokenParts := strings.Split(spyNotify.LastEmailBody, "token=")
	if len(tokenParts) < 2 {
		t.Fatal("Token not found in email body")
	}
	token := tokenParts[1]

	// 3. Verify Token
	req, _ = http.NewRequest("GET", "/api/v1/auth/verify?token="+token, nil)
	rr = httptest.NewRecorder()
	handler.Verify(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Verify failed: got status %d, body: %s", rr.Code, rr.Body.String())
	}

	var tokenResp types.TokenResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &tokenResp))
	if tokenResp.AccessToken == "" {
		t.Error("Verified response missing AccessToken")
	}
	if tokenResp.Principal.Email != email {
		t.Errorf("Verified principal email mismatch: expected %s, got %s", email, tokenResp.Principal.Email)
	}
	if tokenResp.TokenType != "Bearer" {
		t.Errorf("Expected TokenType Bearer, got %s", tokenResp.TokenType)
	}
	if tokenResp.Principal.CreatedAt == "" {
		t.Error("Verified response missing CreatedAt")
	}

	// 4. Repeat Verify (Should Fail - One-time Use)
	rr = httptest.NewRecorder()
	handler.Verify(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Verify should have failed for used token: got status %d", rr.Code)
	}
}
