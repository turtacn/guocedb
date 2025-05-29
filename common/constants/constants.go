// Package constants defines all global constants used throughout the Guocedb project.
// This file centralizes shared configuration and hardcoded values, reducing "magic numbers"
// and redundant definitions across the codebase.
package constants

import "time" // Import time for time-related constants.

// --- General Constants ---

// DefaultConfigPath is the default path to the Guocedb configuration file.
const DefaultConfigPath = "./configs/config.yaml"

// DefaultDataPath is the default directory where database data will be stored.
const DefaultDataPath = "./data"

// ProjectName is the name of the Guocedb project.
const ProjectName = "guocedb"

// Version is the current version of Guocedb.
const Version = "0.1.0-alpha"

// BufferSize defines a common default buffer size in bytes for I/O operations.
const BufferSize = 4096 // 4KB

// --- Network Constants ---

// DefaultMySQLPort is the default port for MySQL protocol connections.
const DefaultMySQLPort = 3306

// DefaultManagementGRPCPort is the default port for the gRPC management API.
const DefaultManagementGRPCPort = 8080

// DefaultManagementRESTPort is the default port for the RESTful management API (future extension).
const DefaultManagementRESTPort = 8081

// NetworkReadTimeout specifies the default timeout for network read operations.
const NetworkReadTimeout = 5 * time.Second

// NetworkWriteTimeout specifies the default timeout for network write operations.
const NetworkWriteTimeout = 5 * time.Second

// MaxConnections defines the maximum number of concurrent client connections allowed.
const MaxConnections = 1000

// --- Storage Constants ---

// DefaultBadgerPath is the default subdirectory for BadgerDB data within DefaultDataPath.
const DefaultBadgerPath = DefaultDataPath + "/badger"

// DefaultMetadataPath is the default subdirectory for persistent metadata within DefaultDataPath.
const DefaultMetadataPath = DefaultDataPath + "/metadata"

// BadgerValueLogFileSize is the maximum size of a BadgerDB value log file in bytes.
const BadgerValueLogFileSize = 128 << 20 // 128 MB

// BadgerSyncWrites enables or disables synchronous writes for BadgerDB (forces disk flush).
const BadgerSyncWrites = false // For performance, often set to false and rely on OS/application syncs

// --- Security Constants ---

// DefaultSaltLength is the default byte length for cryptographic salts.
const DefaultSaltLength = 16

// DefaultPBKDF2Iterations is the default number of iterations for PBKDF2 key derivation.
const DefaultPBKDF2Iterations = 10000

// DefaultUser is the default database user for initial setup.
const DefaultUser = "root"

// DefaultPassword is the default password for the default database user (should be changed immediately).
const DefaultPassword = "password" // WARNING: This is a placeholder. CHANGE IN PRODUCTION!

// --- Logging Constants ---

// DefaultLogLevel is the default severity level for logging.
const DefaultLogLevel = "INFO" // Corresponds to enum.LogLevelInfo

// DefaultLogFilePath is the default path for the Guocedb log file.
const DefaultLogFilePath = "./logs/guocedb.log"

// LogFileMaxSizeMB is the maximum size in MB before a log file is rotated.
const LogFileMaxSizeMB = 100

// LogFileMaxBackups is the maximum number of old log files to retain.
const LogFileMaxBackups = 5

// LogFileMaxAgeDays is the maximum number of days to retain old log files.
const LogFileMaxAgeDays = 7

// --- MySQL Protocol Constants ---

// MySQLServerVersion is the reported MySQL server version.
const MySQLServerVersion = "8.0.30-guocedb" // Mimics a common MySQL version string

// MySQLDefaultCharset is the default character set used for MySQL connections.
const MySQLDefaultCharset = 33 // utf8mb4_general_ci

// MySQLCapabilityFlags are the default capabilities advertised by the MySQL server.
// These flags determine what features the client and server can negotiate.
const MySQLCapabilityFlags = (1 << 0) | // CLIENT_LONG_PASSWORD
	(1 << 1) | // CLIENT_FOUND_ROWS
	(1 << 2) | // CLIENT_LONG_FLAG
	(1 << 3) | // CLIENT_CONNECT_WITH_DB
	(1 << 5) | // CLIENT_PROTOCOL_41 (important for modern clients)
	(1 << 6) | // CLIENT_TRANSACTIONS
	(1 << 11) | // CLIENT_DEPRECATE_EOF (indicates OK_Packet instead of EOF_Packet for result set end)
	(1 << 12) | // CLIENT_SECURE_CONNECTION
	(1 << 13) | // CLIENT_MULTI_STATEMENTS
	(1 << 14) | // CLIENT_MULTI_RESULTS
	(1 << 15) // CLIENT_PS_MULTI_RESULTS

// MySQLStatusFlags are the default status flags sent by the MySQL server.
// These flags indicate the server's current status (e.g., transaction active).
const MySQLStatusFlags = 0 // Initially no special status
//Personal.AI order the ending
