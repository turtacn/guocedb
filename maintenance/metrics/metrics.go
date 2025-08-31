// Package metrics provides performance metrics collection for guocedb.
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/common/log"
)

var (
	// QueryCounter counts the number of queries processed.
	QueryCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "guocedb_queries_total",
			Help: "Total number of queries processed, partitioned by status.",
		},
		[]string{"status"}, // "success" or "failure"
	)

	// QueryLatency measures the latency of queries.
	QueryLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "guocedb_query_latency_seconds",
			Help:    "Latency of queries in seconds.",
			Buckets: prometheus.DefBuckets,
		},
	)

	// ActiveConnections gauges the number of currently active connections.
	ActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "guocedb_active_connections",
			Help: "Number of currently active client connections.",
		},
	)
)

// Registry is a convenience wrapper around a Prometheus registry.
type Registry struct {
	*prometheus.Registry
}

// NewRegistry creates a new metrics registry and registers the default metrics.
func NewRegistry() *Registry {
	r := prometheus.NewRegistry()
	r.MustRegister(QueryCounter)
	r.MustRegister(QueryLatency)
	r.MustRegister(ActiveConnections)
	return &Registry{r}
}

// Serve starts an HTTP server to expose the metrics.
func (r *Registry) Serve(cfg *config.MetricsConfig) {
	if !cfg.Enable {
		log.GetLogger().Info("Metrics server is disabled.")
		return
	}

	addr := ":" + strconv.Itoa(cfg.Port)
	log.GetLogger().Infof("Starting metrics server on %s", addr)

	http.Handle("/metrics", promhttp.HandlerFor(r.Registry, promhttp.HandlerOpts{}))
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.GetLogger().Errorf("Metrics server failed: %v", err)
		}
	}()
}

// RecordQuery is a helper to record the outcome of a query.
func RecordQuery(start time.Time, err error) {
	latency := time.Since(start).Seconds()
	QueryLatency.Observe(latency)

	if err != nil {
		QueryCounter.WithLabelValues("failure").Inc()
	} else {
		QueryCounter.WithLabelValues("success").Inc()
	}
}
