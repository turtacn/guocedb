package integration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/turtacn/guocedb/observability"
	"github.com/turtacn/guocedb/observability/diagnostic"
	"github.com/turtacn/guocedb/observability/health"
	"github.com/turtacn/guocedb/observability/metrics"
)

func TestObservabilityServer(t *testing.T) {
	// Create health checker
	checker := health.NewChecker()
	checker.SetVersion("1.0.0-test")
	checker.AddCheck("test", health.AlwaysHealthyCheck())

	// Create diagnostics
	diag := diagnostic.NewDiagnostics(nil, nil, nil, nil)

	// Create and start observability server
	config := observability.ServerConfig{
		Enabled:     true,
		Address:     ":19090",
		MetricsPath: "/metrics",
		EnablePprof: true,
	}

	server := observability.NewServer(config, checker, diag)
	err := server.Start()
	require.NoError(t, err)

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	baseURL := "http://localhost:19090"

	// Test /metrics endpoint
	t.Run("MetricsEndpoint", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/metrics")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		// Should contain Prometheus metrics
		require.Contains(t, bodyStr, "go_goroutines")
	})

	// Test /health endpoint
	t.Run("HealthEndpoint", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result health.HealthResponse
		json.NewDecoder(resp.Body).Decode(&result)
		require.Equal(t, health.StatusHealthy, result.Status)
		require.Equal(t, "1.0.0-test", result.Version)
	})

	// Test /ready endpoint
	t.Run("ReadyEndpoint", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/ready")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test /live endpoint
	t.Run("LiveEndpoint", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/live")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		require.Equal(t, "alive", result["status"])
	})

	// Test /debug/diagnostic endpoint
	t.Run("DiagnosticEndpoint", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/debug/diagnostic")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result diagnostic.DiagnosticInfo
		json.NewDecoder(resp.Body).Decode(&result)
		require.NotEmpty(t, result.Runtime.GoVersion)
	})

	// Test / (index) endpoint
	t.Run("IndexEndpoint", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		require.Equal(t, "guocedb", result["service"])
		require.NotNil(t, result["endpoints"])
	})
}

func TestMetricsRecording(t *testing.T) {
	// Record some metrics
	metrics.ConnectionsActive.Inc()
	metrics.ConnectionsTotal.Inc()
	metrics.RecordQuery("select", 100*time.Millisecond, true)
	metrics.RecordQuery("insert", 50*time.Millisecond, true)
	metrics.RecordTransaction("commit", 150*time.Millisecond)
	metrics.RecordError("execution")

	// Create server
	config := observability.ServerConfig{
		Enabled:     true,
		Address:     ":19091",
		MetricsPath: "/metrics",
		EnablePprof: false,
	}

	server := observability.NewServer(config, nil, nil)
	err := server.Start()
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	// Verify metrics are exposed
	resp, err := http.Get("http://localhost:19091/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	require.Contains(t, bodyStr, "guocedb_connections_active")
	require.Contains(t, bodyStr, "guocedb_queries_total")
	require.Contains(t, bodyStr, "guocedb_transactions_total")
	require.Contains(t, bodyStr, "guocedb_errors_total")
}

func TestHealthCheckFailure(t *testing.T) {
	checker := health.NewChecker()
	checker.AddCheck("failing", func(ctx context.Context) error {
		return ctx.Err()
	})

	config := observability.ServerConfig{
		Enabled:     true,
		Address:     ":19092",
		MetricsPath: "/metrics",
		EnablePprof: false,
	}

	server := observability.NewServer(config, checker, nil)
	err := server.Start()
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	// Health should fail
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:19092/health", nil)
	resp, err := http.DefaultClient.Do(req)
	
	// Either timeout or service unavailable
	if err == nil {
		require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
		resp.Body.Close()
	}
}

func TestServerDisabled(t *testing.T) {
	config := observability.ServerConfig{
		Enabled: false,
	}

	server := observability.NewServer(config, nil, nil)
	err := server.Start()
	require.NoError(t, err)

	// Server should not be running
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = server.Stop(ctx)
	require.NoError(t, err)
}

func parsePrometheusMetrics(body io.Reader) map[string]float64 {
	metrics := make(map[string]float64)
	data, _ := io.ReadAll(body)
	lines := strings.Split(string(data), "\n")
	
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			// Simplified parsing - real implementation would be more robust
			metrics[parts[0]] = 0
		}
	}
	
	return metrics
}
