package readiness

import (
	"context"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisProbe checks Redis connectivity with PING.
type RedisProbe struct {
	addr string
}

// NewRedisProbe creates a probe that connects to the given Redis address and pings.
// The addr can be a plain host:port or a redis:// URL.
func NewRedisProbe(addr string) *RedisProbe {
	return &RedisProbe{addr: addr}
}

func (p *RedisProbe) Name() string { return "redis" }

func (p *RedisProbe) Check(ctx context.Context) CheckResult {
	start := time.Now()

	if p.addr == "" {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusNotConfigured,
			Message:  "REDIS_URL not set",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	// Support both "host:port" and "redis://..." URL formats.
	var opts *redis.Options
	if strings.HasPrefix(p.addr, "redis://") || strings.HasPrefix(p.addr, "rediss://") {
		var err error
		opts, err = redis.ParseURL(p.addr)
		if err != nil {
			return CheckResult{
				Name:     p.Name(),
				Status:   StatusFailed,
				Message:  "invalid Redis URL: " + err.Error(),
				Duration: time.Since(start).Milliseconds(),
			}
		}
	} else {
		opts = &redis.Options{Addr: p.addr}
	}

	rdb := redis.NewClient(opts)
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusFailed,
			Message:  err.Error(),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return CheckResult{
		Name:     p.Name(),
		Status:   StatusHealthy,
		Duration: time.Since(start).Milliseconds(),
	}
}
