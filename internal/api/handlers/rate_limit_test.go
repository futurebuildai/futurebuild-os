package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

// TestRateLimiting tests the IPRateLimiter middleware directly using a minimal router.
// Phase 12: Tests rewritten to use a standalone router instead of the full server,
// since the old /auth/login endpoint was removed (Clerk handles main app auth).
func TestRateLimiting(t *testing.T) {
	t.Run("Exhaust rate limit for single IP", func(t *testing.T) {
		limiter := middleware.NewIPRateLimiter(rate.Every(12*time.Second), 2, []string{"127.0.0.1"})

		r := chi.NewRouter()
		r.Route("/api/v1/portal/auth", func(r chi.Router) {
			r.Use(middleware.RateLimit(limiter))
			r.Post("/login", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		})
		ts := httptest.NewServer(r)
		defer ts.Close()

		client := &http.Client{}
		ip := "1.2.3.4"

		// Burst is 2. First 2 requests pass, subsequent requests get 429.
		for i := 1; i <= 6; i++ {
			req, _ := http.NewRequest("POST", ts.URL+"/api/v1/portal/auth/login", nil)
			req.Header.Set("X-Forwarded-For", ip)
			resp, err := client.Do(req)
			assert.NoError(t, err)

			if i <= 2 {
				assert.True(t, resp.StatusCode < 429, "Request %d should pass (got %d)", i, resp.StatusCode)
			} else {
				assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode, "Request %d should be rate limited", i)
			}
			resp.Body.Close()
		}
	})

	t.Run("X-Forwarded-For ignored from untrusted proxy", func(t *testing.T) {
		// No trusted proxies — XFF should be ignored, all requests share RemoteAddr
		limiter := middleware.NewIPRateLimiter(rate.Every(12*time.Second), 2, nil)

		r := chi.NewRouter()
		r.Route("/api/v1/portal/auth", func(r chi.Router) {
			r.Use(middleware.RateLimit(limiter))
			r.Post("/login", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		})
		ts := httptest.NewServer(r)
		defer ts.Close()

		client := &http.Client{}

		// Send 3 requests with different XFF headers.
		// If XFF is ignored, all share the same RemoteAddr bucket.
		for i := 0; i < 3; i++ {
			req, _ := http.NewRequest("POST", ts.URL+"/api/v1/portal/auth/login", nil)
			req.Header.Set("X-Forwarded-For", "9.9.9.9")
			resp, _ := client.Do(req)
			if i < 2 {
				assert.True(t, resp.StatusCode < 429, "Request %d should pass", i+1)
			} else {
				assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode,
					"Should be rate limited because untrusted XFF is ignored")
			}
			resp.Body.Close()
		}
	})

	t.Run("X-Forwarded-For honored from trusted proxy", func(t *testing.T) {
		// 127.0.0.1 is httptest default RemoteAddr — trust it
		limiter := middleware.NewIPRateLimiter(rate.Every(12*time.Second), 2, []string{"127.0.0.1"})

		r := chi.NewRouter()
		r.Route("/api/v1/portal/auth", func(r chi.Router) {
			r.Use(middleware.RateLimit(limiter))
			r.Post("/login", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		})
		ts := httptest.NewServer(r)
		defer ts.Close()

		client := &http.Client{}

		// Send 2 requests each for 2 different forwarded IPs.
		// Both should pass because they are tracked independently.
		ips := []string{"10.10.10.1", "10.10.10.2"}
		for _, ip := range ips {
			for i := 0; i < 2; i++ {
				req, _ := http.NewRequest("POST", ts.URL+"/api/v1/portal/auth/login", nil)
				req.Header.Set("X-Forwarded-For", ip)
				resp, _ := client.Do(req)
				assert.True(t, resp.StatusCode < 429,
					"IP %s request %d should pass (got %d)", ip, i+1, resp.StatusCode)
				resp.Body.Close()
			}
		}
	})
}
