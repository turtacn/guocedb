package diagnostic

import "time"

// ConnectionManager interface for connection diagnostics
type ConnectionManager interface {
	GetActiveConnections() []ConnectionInfo
	GetTotalConnections() int64
}

// ConnectionInfo represents connection information
type ConnectionInfo struct {
	ID       uint32
	User     string
	Host     string
	Database string
	Time     time.Time
}

// QueryManager interface for query diagnostics
type QueryManager interface {
	GetActiveQueries() []QueryInfo
	KillQuery(id string) error
}

// QueryInfo represents active query information
type QueryInfo struct {
	ID        string
	User      string
	Database  string
	Query     string
	StartTime time.Time
	State     string
}

// TransactionManager interface for transaction diagnostics
type TransactionManager interface {
	GetStats() TransactionStats
	GetActiveTransactions() []TransactionInfo
}

// TransactionStats represents transaction statistics
type TransactionStats struct {
	Active     int
	Committed  int64
	Rolledback int64
	Conflicts  int64
}

// TransactionInfo represents active transaction information
type TransactionInfo struct {
	ID        string
	User      string
	StartTime time.Time
	State     string
}

// StorageEngine interface for storage diagnostics
type StorageEngine interface {
	Size() (lsm int64, vlog int64)
	KeyCount() int64
	TableCount() map[int]int
}
