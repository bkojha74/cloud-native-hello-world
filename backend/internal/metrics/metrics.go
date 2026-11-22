package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed by endpoint, method, and status.",
		},
		[]string{"method", "status", "endpoint"},
	)

	HTTPRequestDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Latency of HTTP requests in seconds.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5},
		},
		[]string{"method", "endpoint"},
	)

	CacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of successful cache hits.",
		},
	)

	CacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses.",
		},
	)

	CacheErrorsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_errors_total",
			Help: "Total number of cache operations resulting in an error.",
		},
	)

	DBReadsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_reads_total",
			Help: "Total number of read queries executed against database engines.",
		},
		[]string{"database"},
	)

	DBErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_errors_total",
			Help: "Total number of errors encountered during database operations.",
		},
		[]string{"database"},
	)
)
