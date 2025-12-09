package diagnostic

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDiagnosticCollect(t *testing.T) {
	diag := NewDiagnostics(nil, nil, nil, nil)

	info := diag.Collect()

	require.NotZero(t, info.Timestamp)
	require.NotEmpty(t, info.Runtime.GoVersion)
	require.Greater(t, info.Runtime.NumGoroutine, 0)
	require.Greater(t, info.Runtime.NumCPU, 0)
}

func TestDiagnosticEndpoint(t *testing.T) {
	diag := NewDiagnostics(nil, nil, nil, nil)

	srv := httptest.NewServer(diag.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/debug/diagnostic")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var info DiagnosticInfo
	json.NewDecoder(resp.Body).Decode(&info)
	require.NotEmpty(t, info.Runtime.GoVersion)
}

func TestMemoryEndpoint(t *testing.T) {
	diag := NewDiagnostics(nil, nil, nil, nil)

	srv := httptest.NewServer(diag.Handler())
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/debug/memory")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var info map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&info)
	require.Contains(t, info, "alloc_bytes")
	require.Contains(t, info, "heap_alloc_bytes")
}

func TestSlowQueryRecording(t *testing.T) {
	diag := NewDiagnostics(nil, nil, nil, nil)
	diag.SetSlowQueryThreshold(100 * time.Millisecond)

	// Record fast query (should not be recorded)
	diag.RecordSlowQuery("SELECT 1", 50*time.Millisecond, "user1", 1)

	// Record slow queries
	diag.RecordSlowQuery("SELECT * FROM large_table", 500*time.Millisecond, "user2", 10000)
	diag.RecordSlowQuery("UPDATE t SET x=1", 200*time.Millisecond, "user3", 100)

	queries := diag.GetSlowQueries()
	require.Len(t, queries, 2)
	require.Equal(t, "user2", queries[0].User)
}

func TestSlowQueryRingBuffer(t *testing.T) {
	diag := NewDiagnostics(nil, nil, nil, nil)
	diag.slowThreshold = 0
	diag.maxSlowQueries = 5

	// Record more than max
	for i := 0; i < 10; i++ {
		diag.RecordSlowQuery("SELECT 1", time.Second, "user", 1)
	}

	queries := diag.GetSlowQueries()
	require.Len(t, queries, 5)
}

func TestPprofEndpoints(t *testing.T) {
	diag := NewDiagnostics(nil, nil, nil, nil)

	srv := httptest.NewServer(diag.Handler())
	defer srv.Close()

	// Verify pprof endpoints are accessible
	endpoints := []string{
		"/debug/pprof/",
		"/debug/pprof/goroutine",
		"/debug/pprof/heap",
		"/debug/pprof/allocs",
	}

	for _, ep := range endpoints {
		resp, err := http.Get(srv.URL + ep)
		require.NoError(t, err, "endpoint: %s", ep)
		require.Equal(t, http.StatusOK, resp.StatusCode, "endpoint: %s", ep)
		resp.Body.Close()
	}
}

func TestGCEndpoint(t *testing.T) {
	diag := NewDiagnostics(nil, nil, nil, nil)
	srv := httptest.NewServer(diag.Handler())
	defer srv.Close()

	// GET returns GC stats
	resp, _ := http.Get(srv.URL + "/debug/gc")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var info map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&info)
	require.Contains(t, info, "num_gc")

	// POST triggers GC
	resp, _ = http.Post(srv.URL+"/debug/gc", "", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestTruncateQuery(t *testing.T) {
	shortQuery := "SELECT 1"
	require.Equal(t, shortQuery, truncateQuery(shortQuery, 100))

	longQuery := string(make([]byte, 200))
	truncated := truncateQuery(longQuery, 100)
	require.Equal(t, 103, len(truncated)) // 100 + "..."
	require.Contains(t, truncated, "...")
}

type mockConnManager struct{}

func (m *mockConnManager) GetActiveConnections() []ConnectionInfo {
	return []ConnectionInfo{
		{ID: 1, User: "user1", Host: "localhost", Database: "db1", Time: time.Now()},
		{ID: 2, User: "user2", Host: "localhost", Database: "db2", Time: time.Now()},
	}
}

func (m *mockConnManager) GetTotalConnections() int64 {
	return 100
}

func TestCollectConnectionsWithManager(t *testing.T) {
	diag := NewDiagnostics(&mockConnManager{}, nil, nil, nil)

	info := diag.collectConnections()
	require.Equal(t, 2, info.Active)
	require.Equal(t, int64(100), info.Total)
	require.Equal(t, 1, info.ByUser["user1"])
	require.Equal(t, 1, info.ByUser["user2"])
}
