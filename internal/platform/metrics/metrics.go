// Package metrics provides Prometheus instrumentation for the FutureBuild API.
// L7 Code Review: Observability Gap fix.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// RequestDuration tracks HTTP request durations by handler.
var RequestDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "futurebuild",
		Subsystem: "api",
		Name:      "request_duration_seconds",
		Help:      "HTTP request duration in seconds.",
		Buckets:   prometheus.DefBuckets,
	},
	[]string{"handler"},
)

// RequestsTotal tracks total HTTP requests by handler and status code.
var RequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "futurebuild",
		Subsystem: "api",
		Name:      "requests_total",
		Help:      "Total number of HTTP requests.",
	},
	[]string{"handler", "status_code"},
)
