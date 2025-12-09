package health

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHealthyStatus(t *testing.T) {
	checker := NewChecker()
	checker.AddCheck("storage", func(ctx context.Context) error { return nil })
	checker.AddCheck("network", func(ctx context.Context) error { return nil })

	result := checker.Check(context.Background())

	require.Equal(t, StatusHealthy, result.Status)
	require.Len(t, result.Checks, 2)
	for _, c := range result.Checks {
		require.Equal(t, StatusHealthy, c.Status)
	}
}

func TestUnhealthyStatus(t *testing.T) {
	checker := NewChecker()
	checker.AddCheck("storage", func(ctx context.Context) error {
		return errors.New("disk full")
	})
	checker.AddCheck("network", func(ctx context.Context) error { return nil })

	result := checker.Check(context.Background())

	require.Equal(t, StatusUnhealthy, result.Status)
}

func TestHealthEndpointHTTP(t *testing.T) {
	checker := NewChecker()
	checker.AddCheck("db", func(ctx context.Context) error { return nil })

	srv := httptest.NewServer(checker.Handler())
	defer srv.Close()

	// Test /health
	resp, err := http.Get(srv.URL + "/health")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result HealthResponse
	json.NewDecoder(resp.Body).Decode(&result)
	require.Equal(t, StatusHealthy, result.Status)
}

func TestHealthEndpointUnhealthy(t *testing.T) {
	checker := NewChecker()
	checker.AddCheck("storage", func(ctx context.Context) error {
		return errors.New("storage unavailable")
	})

	srv := httptest.NewServer(checker.Handler())
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/health")
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	var result HealthResponse
	json.NewDecoder(resp.Body).Decode(&result)
	require.Equal(t, StatusUnhealthy, result.Status)
}

func TestReadyEndpoint(t *testing.T) {
	checker := NewChecker()
	checker.AddCheck("init", func(ctx context.Context) error { return nil })

	srv := httptest.NewServer(checker.Handler())
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/ready")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	require.Equal(t, "ready", result["status"])
}

func TestReadyEndpointNotReady(t *testing.T) {
	checker := NewChecker()
	checker.AddCheck("init", func(ctx context.Context) error {
		return errors.New("not ready yet")
	})

	srv := httptest.NewServer(checker.Handler())
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/ready")
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

func TestLiveEndpoint(t *testing.T) {
	checker := NewChecker()

	srv := httptest.NewServer(checker.Handler())
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/live")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	require.Equal(t, "alive", result["status"])
	require.NotEmpty(t, result["uptime"])
}

func TestHealthCheckTimeout(t *testing.T) {
	checker := NewChecker()
	checker.SetTimeout(100 * time.Millisecond)

	checker.AddCheck("slow", func(ctx context.Context) error {
		select {
		case <-time.After(1 * time.Second):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	result := checker.Check(context.Background())

	// Should timeout and fail
	require.Equal(t, StatusUnhealthy, result.Status)
}

func TestHealthCheckConcurrency(t *testing.T) {
	checker := NewChecker()

	// Add multiple checks
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("check-%d", i)
		checker.AddCheck(name, func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			return nil
		})
	}

	start := time.Now()
	result := checker.Check(context.Background())
	duration := time.Since(start)

	// Should execute concurrently, total time much less than 500ms
	require.Less(t, duration, 200*time.Millisecond)
	require.Equal(t, StatusHealthy, result.Status)
}

func TestSetVersion(t *testing.T) {
	checker := NewChecker()
	checker.SetVersion("1.0.0")
	checker.AddCheck("test", AlwaysHealthyCheck())

	result := checker.Check(context.Background())
	require.Equal(t, "1.0.0", result.Version)
}

func TestIsHealthy(t *testing.T) {
	checker := NewChecker()
	checker.AddCheck("test", AlwaysHealthyCheck())

	require.True(t, checker.IsHealthy(context.Background()))

	checker.AddCheck("fail", func(ctx context.Context) error {
		return errors.New("failure")
	})

	require.False(t, checker.IsHealthy(context.Background()))
}

func TestRemoveCheck(t *testing.T) {
	checker := NewChecker()
	checker.AddCheck("test1", AlwaysHealthyCheck())
	checker.AddCheck("test2", AlwaysHealthyCheck())

	result := checker.Check(context.Background())
	require.Len(t, result.Checks, 2)

	checker.RemoveCheck("test1")
	result = checker.Check(context.Background())
	require.Len(t, result.Checks, 1)
}
