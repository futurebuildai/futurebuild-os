package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitHubWebhookHandler_VerifySignature(t *testing.T) {
	secret := "test_secret"
	payload := []byte(`{"action":"opened"}`)
	
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validSig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	handler := NewGitHubWebhookHandler(secret, "localhost:6379")
	defer handler.Close()

	tests := []struct {
		name      string
		signature string
		payload   []byte
		want      bool
	}{
		{"Valid Signature", validSig, payload, true},
		{"Invalid Signature", "sha256=invalid", payload, false},
		{"Missing Prefix", hex.EncodeToString(mac.Sum(nil)), payload, false},
		{"Wrong Secret", "sha256=wrong", payload, false},
		{"Empty Signature", "", payload, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.verifySignature(tt.payload, tt.signature)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGitHubWebhookHandler_HandleGitHubWebhook_AuthFail(t *testing.T) {
	// Test fail-closed when secret is missing
	handler := NewGitHubWebhookHandler("", "localhost:6379")
	defer handler.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/github", bytes.NewReader([]byte("{}")))
	rr := httptest.NewRecorder()

	handler.HandleGitHubWebhook(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}
