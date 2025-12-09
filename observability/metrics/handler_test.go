package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMetricsEndpoint(t *testing.T) {
	srv := httptest.NewServer(Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Verify contains Go runtime metrics
	require.Contains(t, bodyStr, "go_goroutines")
	require.Contains(t, bodyStr, "go_memstats")
}

func TestCustomMetricsRecorded(t *testing.T) {
	// Simulate some operations
	ConnectionsActive.Inc()
	ConnectionsTotal.Inc()
	RecordQuery("select", 100*time.Millisecond, true)
	RecordQuery("insert", 50*time.Millisecond, true)
	RecordQuery("select", 200*time.Millisecond, false)
	RecordTransaction("commit", 150*time.Millisecond)

	srv := httptest.NewServer(Handler())
	defer srv.Close()

	resp, _ := http.Get(srv.URL)
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Verify custom metrics
	require.Contains(t, bodyStr, "guocedb_connections_active")
	require.Contains(t, bodyStr, "guocedb_queries_total")
	require.Contains(t, bodyStr, "guocedb_query_duration_seconds")
	require.Contains(t, bodyStr, "guocedb_transactions_total")

	// Verify labels
	require.Contains(t, bodyStr, `type="select"`)
}

func TestMetricsContentType(t *testing.T) {
	srv := httptest.NewServer(Handler())
	defer srv.Close()

	resp, _ := http.Get(srv.URL)
	contentType := resp.Header.Get("Content-Type")

	// Prometheus format
	require.True(t,
		strings.Contains(contentType, "text/plain") ||
			strings.Contains(contentType, "application/openmetrics-text"),
	)
}

func TestQueryDurationHistogramExposed(t *testing.T) {
	// Record different latencies
	durations := []time.Duration{
		1 * time.Millisecond,
		10 * time.Millisecond,
		100 * time.Millisecond,
		1 * time.Second,
		5 * time.Second,
	}

	for _, d := range durations {
		RecordQuery("test", d, true)
	}

	srv := httptest.NewServer(Handler())
	defer srv.Close()

	resp, _ := http.Get(srv.URL)
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Verify histogram buckets
	require.Contains(t, bodyStr, "guocedb_query_duration_seconds_bucket")
	require.Contains(t, bodyStr, "guocedb_query_duration_seconds_count")
	require.Contains(t, bodyStr, "guocedb_query_duration_seconds_sum")
}

func TestHandlerWithAuth(t *testing.T) {
	handler := HandlerWithAuth("admin", "secret")
	srv := httptest.NewServer(handler)
	defer srv.Close()

	// Test without auth
	resp, _ := http.Get(srv.URL)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Test with wrong credentials
	req, _ := http.NewRequest("GET", srv.URL, nil)
	req.SetBasicAuth("admin", "wrong")
	resp, _ = http.DefaultClient.Do(req)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Test with correct credentials
	req, _ = http.NewRequest("GET", srv.URL, nil)
	req.SetBasicAuth("admin", "secret")
	resp, _ = http.DefaultClient.Do(req)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
