// Package constants defines global constants for the guocedb project.
package constants

const (
	// DatabaseVersion is the current version of the database.
	DatabaseVersion = "0.1.0-alpha"
	// ProtocolVersion is the MySQL protocol version supported.
	ProtocolVersion = "8.0.23"
)

const (
	// DefaultPort is the default port for the MySQL protocol server.
	DefaultPort = 3306
	// DefaultGRPCPort is the default port for the gRPC management server.
	DefaultGRPCPort = 50051
	// DefaultMetricsPort is the default port for the Prometheus metrics endpoint.
	DefaultMetricsPort = 9100
	// DefaultTimeout is the default query timeout in seconds.
	DefaultTimeout = 30
	// DefaultMaxConnections is the default maximum number of client connections.
	DefaultMaxConnections = 1024
	// DefaultCacheSize is the default cache size in MB.
	DefaultCacheSize = 256
)

// Storage engine type constants
const (
	StorageEngineBadger = "badger"
	StorageEngineKVD    = "kvd"
	StorageEngineMDD    = "mdd"
	StorageEngineMDI    = "mdi"
)

// Error message templates
const (
	ErrFmtParseError      = "SQL parse error: %v"
	ErrFmtAnalyseError    = "SQL analyse error: %v"
	ErrFmtOptimizeError   = "SQL optimize error: %v"
	ErrFmtExecuteError    = "SQL execute error: %v"
	ErrFmtTxnError        = "Transaction error: %v"
	ErrFmtStorageError    = "Storage error: %v"
	ErrFmtConfigLoadError = "Failed to load configuration: %v"
)

// Error codes (example placeholders)
const (
	ErrCodeUnknown = 1000
	ErrCodeSyntax  = 1001
	ErrCodeRuntime = 1002
	ErrCodeSystem  = 1003
)
