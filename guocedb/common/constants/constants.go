package constants

// System-related constants
const (
	DatabaseVersion = "0.1.0-alpha"
	ProtocolVersion = 8
	ServerName      = "guocedb"
)

// Default configuration values
const (
	DefaultMySQLPort   = 3306
	DefaultGRPCPort    = 50051
	DefaultHTTPPort    = 8080
	DefaultTimeout     = "30s"
	DefaultCacheSizeMB = 256
)

// Storage engine types
const (
	StorageEngineBadger = "badger"
	StorageEngineKVD    = "kvd"
	StorageEngineMDD    = "mdd"
	StorageEngineMDI    = "mdi"
)

// Error code constants
// Using MySQL error codes for compatibility where possible.
const (
	ErrCodeSyntaxError    = 1064
	ErrCodeNotImplemented = 4000 // Custom error code
)

// Error message templates
const (
	MsgNotImplemented = "feature not implemented"
)
