package diagnostic

import (
	"runtime"
	"sync"
	"time"
)

// DiagnosticInfo represents comprehensive diagnostic information
type DiagnosticInfo struct {
	Timestamp    time.Time          `json:"timestamp"`
	Runtime      RuntimeInfo        `json:"runtime"`
	Connections  ConnectionsInfo    `json:"connections"`
	Queries      QueriesInfo        `json:"queries"`
	Transactions TransactionsInfo   `json:"transactions"`
	Storage      StorageInfo        `json:"storage"`
}

// RuntimeInfo represents Go runtime information
type RuntimeInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
	GOMAXPROCS   int    `json:"gomaxprocs"`
	MemAlloc     uint64 `json:"mem_alloc_bytes"`
	MemTotal     uint64 `json:"mem_total_bytes"`
	MemSys       uint64 `json:"mem_sys_bytes"`
	NumGC        uint32 `json:"num_gc"`
}

// ConnectionsInfo represents connection information
type ConnectionsInfo struct {
	Active int              `json:"active"`
	Total  int64            `json:"total"`
	ByUser map[string]int   `json:"by_user"`
	ByHost map[string]int   `json:"by_host"`
}

// QueriesInfo represents query information
type QueriesInfo struct {
	ActiveQueries []ActiveQuery `json:"active_queries"`
	SlowQueries   []SlowQuery   `json:"slow_queries,omitempty"`
}

// ActiveQuery represents an active query
type ActiveQuery struct {
	ID        string        `json:"id"`
	User      string        `json:"user"`
	Database  string        `json:"database"`
	Query     string        `json:"query"`
	StartTime time.Time     `json:"start_time"`
	Duration  time.Duration `json:"duration_ms"`
	State     string        `json:"state"`
}

// SlowQuery represents a slow query
type SlowQuery struct {
	Query     string        `json:"query"`
	Duration  time.Duration `json:"duration_ms"`
	Timestamp time.Time     `json:"timestamp"`
	User      string        `json:"user"`
	RowsRead  int64         `json:"rows_read"`
}

// TransactionsInfo represents transaction information
type TransactionsInfo struct {
	Active     int   `json:"active"`
	Committed  int64 `json:"committed"`
	Rolledback int64 `json:"rolledback"`
	Conflicts  int64 `json:"conflicts"`
}

// StorageInfo represents storage information
type StorageInfo struct {
	DataSize  int64 `json:"data_size_bytes"`
	IndexSize int64 `json:"index_size_bytes"`
	WalSize   int64 `json:"wal_size_bytes"`
	NumKeys   int64 `json:"num_keys"`
	NumTables int   `json:"num_tables"`
}

// Diagnostics manages diagnostic data collection
type Diagnostics struct {
	connMgr    ConnectionManager
	queryMgr   QueryManager
	txnMgr     TransactionManager
	storage    StorageEngine

	slowQueries    []SlowQuery
	slowQueryMu    sync.RWMutex
	slowThreshold  time.Duration
	maxSlowQueries int
}

// NewDiagnostics creates a new Diagnostics instance
func NewDiagnostics(connMgr ConnectionManager, queryMgr QueryManager, txnMgr TransactionManager, storage StorageEngine) *Diagnostics {
	return &Diagnostics{
		connMgr:        connMgr,
		queryMgr:       queryMgr,
		txnMgr:         txnMgr,
		storage:        storage,
		slowQueries:    make([]SlowQuery, 0, 100),
		slowThreshold:  time.Second,
		maxSlowQueries: 100,
	}
}

// SetSlowQueryThreshold sets the threshold for slow queries
func (d *Diagnostics) SetSlowQueryThreshold(threshold time.Duration) {
	d.slowThreshold = threshold
}

// Collect collects all diagnostic information
func (d *Diagnostics) Collect() *DiagnosticInfo {
	return &DiagnosticInfo{
		Timestamp:    time.Now(),
		Runtime:      d.collectRuntime(),
		Connections:  d.collectConnections(),
		Queries:      d.collectQueries(),
		Transactions: d.collectTransactions(),
		Storage:      d.collectStorage(),
	}
}

func (d *Diagnostics) collectRuntime() RuntimeInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return RuntimeInfo{
		GoVersion:    runtime.Version(),
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
		GOMAXPROCS:   runtime.GOMAXPROCS(0),
		MemAlloc:     m.Alloc,
		MemTotal:     m.TotalAlloc,
		MemSys:       m.Sys,
		NumGC:        m.NumGC,
	}
}

func (d *Diagnostics) collectConnections() ConnectionsInfo {
	info := ConnectionsInfo{
		ByUser: make(map[string]int),
		ByHost: make(map[string]int),
	}

	if d.connMgr != nil {
		conns := d.connMgr.GetActiveConnections()
		info.Active = len(conns)
		info.Total = d.connMgr.GetTotalConnections()

		for _, conn := range conns {
			info.ByUser[conn.User]++
			info.ByHost[conn.Host]++
		}
	}

	return info
}

func (d *Diagnostics) collectQueries() QueriesInfo {
	info := QueriesInfo{
		ActiveQueries: make([]ActiveQuery, 0),
	}

	if d.queryMgr != nil {
		for _, q := range d.queryMgr.GetActiveQueries() {
			info.ActiveQueries = append(info.ActiveQueries, ActiveQuery{
				ID:        q.ID,
				User:      q.User,
				Database:  q.Database,
				Query:     truncateQuery(q.Query, 200),
				StartTime: q.StartTime,
				Duration:  time.Since(q.StartTime),
				State:     q.State,
			})
		}
	}

	d.slowQueryMu.RLock()
	info.SlowQueries = make([]SlowQuery, len(d.slowQueries))
	copy(info.SlowQueries, d.slowQueries)
	d.slowQueryMu.RUnlock()

	return info
}

func (d *Diagnostics) collectTransactions() TransactionsInfo {
	info := TransactionsInfo{}

	if d.txnMgr != nil {
		stats := d.txnMgr.GetStats()
		info.Active = stats.Active
		info.Committed = stats.Committed
		info.Rolledback = stats.Rolledback
		info.Conflicts = stats.Conflicts
	}

	return info
}

func (d *Diagnostics) collectStorage() StorageInfo {
	info := StorageInfo{}

	if d.storage != nil {
		lsm, vlog := d.storage.Size()
		info.DataSize = lsm
		info.IndexSize = 0 // BadgerDB doesn't separate index
		info.WalSize = vlog
		info.NumKeys = d.storage.KeyCount()
		
		tables := d.storage.TableCount()
		for _, count := range tables {
			info.NumTables += count
		}
	}

	return info
}

// RecordSlowQuery records a slow query
func (d *Diagnostics) RecordSlowQuery(query string, duration time.Duration, user string, rowsRead int64) {
	if duration < d.slowThreshold {
		return
	}

	d.slowQueryMu.Lock()
	defer d.slowQueryMu.Unlock()

	sq := SlowQuery{
		Query:     truncateQuery(query, 500),
		Duration:  duration,
		Timestamp: time.Now(),
		User:      user,
		RowsRead:  rowsRead,
	}

	// Ring buffer
	if len(d.slowQueries) >= d.maxSlowQueries {
		d.slowQueries = d.slowQueries[1:]
	}
	d.slowQueries = append(d.slowQueries, sq)
}

// GetSlowQueries returns recorded slow queries
func (d *Diagnostics) GetSlowQueries() []SlowQuery {
	d.slowQueryMu.RLock()
	defer d.slowQueryMu.RUnlock()

	result := make([]SlowQuery, len(d.slowQueries))
	copy(result, d.slowQueries)
	return result
}

func truncateQuery(query string, maxLen int) string {
	if len(query) <= maxLen {
		return query
	}
	return query[:maxLen] + "..."
}
