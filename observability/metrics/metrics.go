package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespace = "guocedb"
)

var (
	// Connection metrics
	ConnectionsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "connections_active",
		Help:      "Number of active connections",
	})

	ConnectionsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "connections_total",
		Help:      "Total number of connections",
	})

	ConnectionsRejected = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "connections_rejected_total",
		Help:      "Total number of rejected connections",
	}, []string{"reason"})

	// Query metrics
	QueriesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "queries_total",
		Help:      "Total number of queries",
	}, []string{"type", "status"})

	QueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "query_duration_seconds",
		Help:      "Query duration in seconds",
		Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"type"})

	// Transaction metrics
	TransactionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "transactions_total",
		Help:      "Total number of transactions",
	}, []string{"status"}) // status: commit, rollback, conflict

	TransactionDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "transaction_duration_seconds",
		Help:      "Transaction duration in seconds",
		Buckets:   prometheus.DefBuckets,
	})

	// Row operation metrics
	RowsRead = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "rows_read_total",
		Help:      "Total number of rows read",
	})

	RowsWritten = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "rows_written_total",
		Help:      "Total number of rows written",
	})

	// Storage metrics
	StorageBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "storage_bytes",
		Help:      "Storage size in bytes",
	}, []string{"type"}) // type: data, index, wal

	StorageKeys = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "storage_keys",
		Help:      "Number of keys in storage",
	})

	// Error metrics
	ErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "errors_total",
		Help:      "Total number of errors",
	}, []string{"type"}) // type: parse, execution, transaction, storage, auth
)

// RecordQuery records query execution metrics
func RecordQuery(queryType string, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	QueriesTotal.WithLabelValues(queryType, status).Inc()
	QueryDuration.WithLabelValues(queryType).Observe(duration.Seconds())
}

// RecordTransaction records transaction execution metrics
func RecordTransaction(status string, duration time.Duration) {
	TransactionsTotal.WithLabelValues(status).Inc()
	TransactionDuration.Observe(duration.Seconds())
}

// RecordError records error metrics
func RecordError(errType string) {
	ErrorsTotal.WithLabelValues(errType).Inc()
}

// RecordConnectionRejected records rejected connection with reason
func RecordConnectionRejected(reason string) {
	ConnectionsRejected.WithLabelValues(reason).Inc()
}

// RecordRowsRead records number of rows read
func RecordRowsRead(count int64) {
	RowsRead.Add(float64(count))
}

// RecordRowsWritten records number of rows written
func RecordRowsWritten(count int64) {
	RowsWritten.Add(float64(count))
}

// UpdateStorageBytes updates storage size metrics
func UpdateStorageBytes(storageType string, bytes int64) {
	StorageBytes.WithLabelValues(storageType).Set(float64(bytes))
}

// UpdateStorageKeys updates storage keys count
func UpdateStorageKeys(count int64) {
	StorageKeys.Set(float64(count))
}
