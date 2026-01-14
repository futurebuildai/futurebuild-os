package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiting(t *testing.T) {
	cfg := &config.Config{AppPort: 8080, TrustedProxies: []string{"127.0.0.1"}}

	s := server.NewServer(nil, cfg, nil)
	ts := httptest.NewServer(s.Router)
	defer ts.Close()

	client := &http.Client{}

	t.Run("Exhaust rate limit for single IP", func(t *testing.T) {
		ip := "1.2.3.4"

		// Burst is 2, Rate is 5/min (1 every 12s).
		// We should be able to send 2 immediately (burst), then it depends on timing.
		// For the test, we'll send 6 rapidly and expect 429 eventually.

		for i := 1; i <= 6; i++ {
			req, _ := http.NewRequest("POST", ts.URL+"/api/v1/auth/login", nil)
			req.Header.Set("X-Forwarded-For", ip)

			resp, err := client.Do(req)
			assert.NoError(t, err)

			if i <= 2 {
				// Burst allows first 2. We accept 400 as "passed middleware"
				// since we're not sending a valid body.
				assert.True(t, resp.StatusCode < 429, "Request %d should pass middleware (got %d)", i, resp.StatusCode)
			} else if i > 2 && i < 6 {
				// Between 3 and 5, it should be 429
				assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
			}

			if i == 6 {
				assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode, "6th request should be rate limited")

				var body map[string]string
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
				assert.Equal(t, "Too Many Requests", body["error"])
				assert.Equal(t, "12s", body["retry_after"])
				assert.Equal(t, "12", resp.Header.Get("Retry-After"))
			}
			resp.Body.Close()
		}
	})

	t.Run("X-Forwarded-For ignored from untrusted proxy", func(t *testing.T) {
		// New server with NO trusted proxies (nil/empty)
		s2 := server.NewServer(nil, &config.Config{AppPort: 8081}, nil)
		ts2 := httptest.NewServer(s2.Router)
		defer ts2.Close()

		ipSpook := "9.9.9.9"

		// 1st request with spoofed header
		req, _ := http.NewRequest("POST", ts2.URL+"/api/v1/auth/login", nil)
		req.Header.Set("X-Forwarded-For", ipSpook)
		// We can't easily spoof RemoteAddr in httptest.NewServer easily without more setup,
		// but the middleware now defaults to RemoteAddr (usually 127.0.0.1 in tests)
		// if NOT in the proxies map.

		// Send 10 requests. If it uses RemoteAddr (127.0.0.1), all 10 hits same bucket.
		// If it uses X-Forwarded-For (spoofed), it would hits different buckets if we changed it.
		// Let's just verify it returns 429 using the SAME bucket for different XFF headers if proxy is untrusted.

		for i := 0; i < 3; i++ {
			req, _ := http.NewRequest("POST", ts2.URL+"/api/v1/auth/login", nil)
			req.Header.Set("X-Forwarded-For", fmt.Sprintf("1.1.1.%d", i))
			resp, _ := client.Do(req)
			if i < 2 {
				assert.True(t, resp.StatusCode < 429)
			} else {
				assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode, "Should be rate limited because untrusted XFF is ignored and everyone shares RemoteAddr (127.0.0.1)")
			}
			resp.Body.Close()
		}
	})

	t.Run("X-Forwarded-For honored from trusted proxy", func(t *testing.T) {
		// New server with 127.0.0.1 (httptest default) as trusted proxy
		s3 := server.NewServer(nil, &config.Config{AppPort: 8082, TrustedProxies: []string{"127.0.0.1"}}, nil)
		ts3 := httptest.NewServer(s3.Router)
		defer ts3.Close()

		// Send 2 requests each for 2 different "forwarded" IPs.
		// Both should pass because they are tracked independently.
		ips := []string{"10.10.10.1", "10.10.10.2"}
		for _, ip := range ips {
			for i := 0; i < 2; i++ {
				req, _ := http.NewRequest("POST", ts3.URL+"/api/v1/auth/login", nil)
				req.Header.Set("X-Forwarded-For", ip)
				resp, _ := client.Do(req)
				assert.True(t, resp.StatusCode < 429)
				resp.Body.Close()
			}
		}
	})
}
