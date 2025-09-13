package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// QueriesTotal is a counter for the total number of queries processed.
	QueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "guocedb_queries_total",
			Help: "Total number of queries processed by the server.",
		},
		[]string{"type"}, // e.g., "select", "insert"
	)

	// ActiveConnections is a gauge for the number of currently active connections.
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "guocedb_active_connections",
			Help: "Number of currently active client connections.",
		},
	)

	// QueryDuration is a histogram for query execution times.
	QueryDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "guocedb_query_duration_seconds",
			Help:    "Histogram of query execution times.",
			Buckets: prometheus.DefBuckets, // Default buckets
		},
	)
)

// NewHandler creates a new HTTP handler for exposing Prometheus metrics.
// This handler should be registered on an internal HTTP server.
func NewHandler() http.Handler {
	return promhttp.Handler()
}

// TODO: Define more granular metrics for other components (cache hit/miss, transaction rates, etc.).
// TODO: Use a custom registry for better isolation if running multiple servers in one process.
