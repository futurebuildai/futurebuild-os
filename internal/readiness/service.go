package readiness

import (
	"context"
	"sync"
	"time"
)

// criticalServices are services that must be healthy in production/staging.
// A not_configured critical service in production triggers overall "failed".
// In staging it triggers "degraded".
var criticalServices = map[string]bool{
	"database":  true,
	"clerk":     true,
	"redis":     true,
	"resend":    true,
	"sendgrid":  true,
	"vertex_ai": true,
}

// Service runs all registered probes concurrently and aggregates the results.
type Service struct {
	timeout  time.Duration
	checkers []Checker
}

// NewService creates a readiness service with the given per-probe timeout and checkers.
func NewService(timeout time.Duration, checkers ...Checker) *Service {
	return &Service{
		timeout:  timeout,
		checkers: checkers,
	}
}

// Run executes all probes concurrently and returns an aggregated report.
func (s *Service) Run(ctx context.Context, env string) Report {
	start := time.Now()

	results := make([]CheckResult, len(s.checkers))
	var wg sync.WaitGroup

	for i, checker := range s.checkers {
		wg.Add(1)
		go func(idx int, c Checker) {
			defer wg.Done()
			probeCtx, cancel := context.WithTimeout(ctx, s.timeout)
			defer cancel()
			results[idx] = c.Check(probeCtx)
		}(i, checker)
	}

	wg.Wait()

	overall := s.computeOverallStatus(results, env)

	return Report{
		Status:      overall,
		Environment: env,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Duration:    time.Since(start).Milliseconds(),
		Checks:      results,
	}
}

// computeOverallStatus determines the aggregate status based on individual results
// and the environment context.
func (s *Service) computeOverallStatus(results []CheckResult, env string) Status {
	isProd := env == "production" || env == "prod"
	isStaging := env == "staging" || env == "stage"

	hasEmail := false
	overall := StatusHealthy

	for _, r := range results {
		switch r.Status {
		case StatusFailed:
			if isProd || isStaging {
				return StatusFailed
			}
			overall = StatusDegraded

		case StatusNotConfigured:
			isCritical := criticalServices[r.Name]

			// Email is special: either Resend OR SendGrid satisfies the requirement.
			if r.Name == "resend" || r.Name == "sendgrid" {
				// Defer email evaluation until we've seen both.
				continue
			}

			if isCritical {
				if isProd {
					return StatusFailed
				}
				if isStaging {
					overall = StatusDegraded
				}
			}

		case StatusDegraded:
			if overall == StatusHealthy {
				overall = StatusDegraded
			}
		}

		// Track whether at least one email provider is healthy.
		if (r.Name == "resend" || r.Name == "sendgrid") && r.Status == StatusHealthy {
			hasEmail = true
		}
	}

	// Evaluate email providers: at least one must be configured in prod/staging.
	if !hasEmail {
		emailConfigured := false
		for _, r := range results {
			if (r.Name == "resend" || r.Name == "sendgrid") && r.Status != StatusNotConfigured {
				emailConfigured = true
				break
			}
		}
		if !emailConfigured {
			if isProd {
				return StatusFailed
			}
			if isStaging && overall == StatusHealthy {
				overall = StatusDegraded
			}
		}
	}

	return overall
}
