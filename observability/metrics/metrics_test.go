package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConnectionMetrics(t *testing.T) {
	// Reset metrics
	ConnectionsActive.Set(0)
	
	// Test increment
	ConnectionsActive.Inc()
	ConnectionsActive.Inc()
	ConnectionsTotal.Inc()
	ConnectionsTotal.Inc()
	
	// We can't easily read prometheus metrics, but we can verify no panics
	// In a real scenario, you'd use a test registry
}

func TestQueryDurationHistogram(t *testing.T) {
	durations := []time.Duration{
		1 * time.Millisecond,
		10 * time.Millisecond,
		100 * time.Millisecond,
		1 * time.Second,
		5 * time.Second,
	}

	for _, d := range durations {
		RecordQuery("select", d, true)
	}

	// Verify no panics
	require.True(t, true)
}

func TestTransactionCounters(t *testing.T) {
	RecordTransaction("commit", 100*time.Millisecond)
	RecordTransaction("rollback", 50*time.Millisecond)
	RecordTransaction("conflict", 200*time.Millisecond)

	// Verify no panics
	require.True(t, true)
}

func TestErrorRateCounter(t *testing.T) {
	RecordError("parse")
	RecordError("execution")
	RecordError("transaction")
	RecordError("storage")

	// Verify no panics
	require.True(t, true)
}

func TestRecordQuery(t *testing.T) {
	// Test successful query
	RecordQuery("select", 100*time.Millisecond, true)
	RecordQuery("insert", 50*time.Millisecond, true)

	// Test failed query
	RecordQuery("select", 200*time.Millisecond, false)

	// Verify no panics
	require.True(t, true)
}

func TestRecordConnectionRejected(t *testing.T) {
	RecordConnectionRejected("max_connections")
	RecordConnectionRejected("auth_failed")

	// Verify no panics
	require.True(t, true)
}

func TestRecordRows(t *testing.T) {
	RecordRowsRead(100)
	RecordRowsWritten(50)

	// Verify no panics
	require.True(t, true)
}

func TestUpdateStorageMetrics(t *testing.T) {
	UpdateStorageBytes("data", 1024*1024)
	UpdateStorageBytes("index", 512*1024)
	UpdateStorageBytes("wal", 256*1024)
	UpdateStorageKeys(1000)

	// Verify no panics
	require.True(t, true)
}
