package a2a

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// SignRequest computes an HMAC-SHA256 signature over the concatenation of
// timestamp + nonce + body using the provided secret.
// This mirrors FB-Brain's internal/security.SignRequest exactly so signatures
// are interoperable between the two systems.
func SignRequest(secret, timestamp, nonce string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte(nonce))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature checks that the provided signature matches the expected
// HMAC-SHA256 over timestamp + nonce + body. Uses constant-time comparison
// to prevent timing attacks.
func VerifySignature(secret, signature, timestamp, nonce string, body []byte) bool {
	expected := SignRequest(secret, timestamp, nonce, body)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// GenerateNonce returns a unique nonce based on the current time in nanoseconds.
// For production use this provides sufficient uniqueness; UUIDs are heavier.
func GenerateNonce() string {
	return fmt.Sprintf("%d-%x", time.Now().UnixNano(), time.Now().UnixNano()&0xFFFF)
}

// CurrentTimestamp returns the current Unix epoch time as a string.
func CurrentTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}
