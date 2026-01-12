package middleware

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// visitor tracks a rate.Limiter and the last time it was seen.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter manages rate limiters for individual IP addresses.
type IPRateLimiter struct {
	visitors map[string]*visitor
	mu       *sync.RWMutex
	r        rate.Limit
	b        int
	proxies  map[string]struct{}
}

// NewIPRateLimiter creates a new IPRateLimiter with the specified rate, burst, and trusted proxies.
func NewIPRateLimiter(r rate.Limit, b int, trustedProxies []string) *IPRateLimiter {
	proxies := make(map[string]struct{})
	for _, p := range trustedProxies {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			proxies[trimmed] = struct{}{}
		}
	}

	limiter := &IPRateLimiter{
		visitors: make(map[string]*visitor),
		mu:       &sync.RWMutex{},
		r:        r,
		b:        b,
		proxies:  proxies,
	}

	// Memory Hygiene: Granular cleanup routine.
	go limiter.cleanupRoutine()

	return limiter
}

// GetLimiter returns a rate.Limiter for the given IP address, creating one if it doesn't exist.
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	v, exists := i.visitors[ip]
	if !exists {
		l := rate.NewLimiter(i.r, i.b)
		i.visitors[ip] = &visitor{
			limiter:  l,
			lastSeen: time.Now(),
		}
		return l
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// ExtractIP pulls the client IP, only trusting headers from defined proxies.
func (i *IPRateLimiter) ExtractIP(r *http.Request) string {
	remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)

	// Only trust headers if the connection comes from a trusted proxy
	if _, trusted := i.proxies[remoteIP]; !trusted {
		return remoteIP
	}

	// 1. Check X-Forwarded-For
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 2. Check X-Real-IP
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	return remoteIP
}

// cleanupRoutine granularly removes inactive visitors every hour.
func (i *IPRateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		i.mu.Lock()
		for ip, v := range i.visitors {
			if time.Since(v.lastSeen) > 1*time.Hour {
				delete(i.visitors, ip)
			}
		}
		i.mu.Unlock()
	}
}

// RateLimit returns a middleware that limits requests by IP.
func RateLimit(limiter *IPRateLimiter) func(http.Handler) http.Handler {
	// Calculate dynamic retry seconds based on rate.
	// rate.Limit is req/sec. 1/rate = seconds per request.
	retrySeconds := int(math.Ceil(1.0 / float64(limiter.r)))
	retryAfterStr := fmt.Sprintf("%d", retrySeconds)
	retryJSONStr := fmt.Sprintf("%ds", retrySeconds)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := limiter.ExtractIP(r)
			l := limiter.GetLimiter(ip)

			if !l.Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", retryAfterStr)
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"error":       "Too Many Requests",
					"retry_after": retryJSONStr,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
