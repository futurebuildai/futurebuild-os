package readiness

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabaseProbe checks PostgreSQL connectivity via pool.Ping.
type DatabaseProbe struct {
	pool *pgxpool.Pool
}

// NewDatabaseProbe creates a probe that pings the given connection pool.
// Unlike other probes, it reuses the existing pool because Ping is non-destructive
// and creating a separate pool per check would be wasteful.
func NewDatabaseProbe(pool *pgxpool.Pool) *DatabaseProbe {
	return &DatabaseProbe{pool: pool}
}

func (p *DatabaseProbe) Name() string { return "database" }

func (p *DatabaseProbe) Check(ctx context.Context) CheckResult {
	start := time.Now()

	if p.pool == nil {
		return CheckResult{
			Name:     p.Name(),
			Status:   StatusNotConfigured,
			Message:  "database pool is nil",
			Duration: time.Since(start).Milliseconds(),
		}
	}

	if err := p.pool.Ping(ctx); err != nil {
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
