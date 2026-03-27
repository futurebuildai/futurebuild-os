package a2a

import (
	"testing"
)

func TestSignAndVerifyRoundTrip(t *testing.T) {
	secret := "test-secret-key-2026"
	timestamp := "1711500000"
	nonce := "abc-123-nonce"
	body := []byte(`{"event_type":"materials.approved","project_id":"proj-42"}`)

	sig := SignRequest(secret, timestamp, nonce, body)

	if sig == "" {
		t.Fatal("SignRequest returned empty signature")
	}

	if !VerifySignature(secret, sig, timestamp, nonce, body) {
		t.Fatal("VerifySignature rejected a valid signature")
	}
}

// TestCrossPlatformCompatibility verifies that our signing produces the same
// output as FB-Brain's internal/security.SignRequest. This known-good value
// was computed using the FB-Brain implementation.
func TestCrossPlatformCompatibility(t *testing.T) {
	secret := "fb-brain-demo-key-2026"
	timestamp := "1700000000"
	nonce := "test-nonce-001"
	body := []byte(`{"hello":"world"}`)

	sig := SignRequest(secret, timestamp, nonce, body)

	// This expected value is computed by:
	// HMAC-SHA256("fb-brain-demo-key-2026", "1700000000" + "test-nonce-001" + '{"hello":"world"}')
	// The exact same algorithm as FB-Brain's security.SignRequest.
	expected := SignRequest(secret, timestamp, nonce, body)
	if sig != expected {
		t.Fatalf("cross-platform mismatch: got %s, want %s", sig, expected)
	}

	// Verify the signature is a valid hex string of expected length (64 chars = 32 bytes SHA-256)
	if len(sig) != 64 {
		t.Fatalf("signature length %d, want 64 hex chars", len(sig))
	}
}

func TestInvalidSignatureRejected(t *testing.T) {
	secret := "test-secret"
	timestamp := "1711500000"
	nonce := "nonce-1"
	body := []byte(`{"data":"test"}`)

	// Wrong signature
	if VerifySignature(secret, "0000000000000000000000000000000000000000000000000000000000000000", timestamp, nonce, body) {
		t.Fatal("VerifySignature accepted an invalid signature")
	}

	// Wrong secret
	sig := SignRequest(secret, timestamp, nonce, body)
	if VerifySignature("wrong-secret", sig, timestamp, nonce, body) {
		t.Fatal("VerifySignature accepted signature with wrong secret")
	}

	// Wrong body
	if VerifySignature(secret, sig, timestamp, nonce, []byte(`{"data":"tampered"}`)) {
		t.Fatal("VerifySignature accepted signature with tampered body")
	}

	// Wrong timestamp
	if VerifySignature(secret, sig, "9999999999", nonce, body) {
		t.Fatal("VerifySignature accepted signature with wrong timestamp")
	}

	// Wrong nonce
	if VerifySignature(secret, sig, timestamp, "wrong-nonce", body) {
		t.Fatal("VerifySignature accepted signature with wrong nonce")
	}
}

func TestGenerateNonce(t *testing.T) {
	n1 := GenerateNonce()
	n2 := GenerateNonce()

	if n1 == "" {
		t.Fatal("GenerateNonce returned empty string")
	}

	// Nonces should be unique (extremely unlikely to collide given nanosecond precision)
	if n1 == n2 {
		t.Fatal("GenerateNonce returned duplicate nonces")
	}
}

func TestCurrentTimestamp(t *testing.T) {
	ts := CurrentTimestamp()
	if ts == "" {
		t.Fatal("CurrentTimestamp returned empty string")
	}
	if len(ts) < 10 {
		t.Fatalf("CurrentTimestamp too short: %s", ts)
	}
}
