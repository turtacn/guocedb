// Package enum defines global enumeration types used throughout the Guocedb project.
// These enumerations provide a set of named constants representing predefined choices.
package enum

// GuocedbErrorType represents the type of errors that can occur in Guocedb.
// It is used to categorize errors for consistent error handling and logging.
type GuocedbErrorType int

const (
	// ErrUnknown represents an unknown or uncategorized error.
	ErrUnknown GuocedbErrorType = iota
	// ErrConfiguration represents an error related to system configuration.
	ErrConfiguration
	// ErrNetwork represents an error in network communication.
	ErrNetwork
	// ErrStorage represents an error in the storage layer.
	ErrStorage
	// ErrCompute represents an error in the compute layer (e.g., query processing).
	ErrCompute
	// ErrSecurity represents an error in the security layer (e.g., authentication, authorization).
	ErrSecurity
	// ErrCatalog represents an error related to metadata catalog operations.
	ErrCatalog
	// ErrProtocol represents an error in handling database protocols (e.g., MySQL protocol).
	ErrProtocol
	// ErrTransaction represents an error during transaction processing.
	ErrTransaction
	// ErrInvalidArgument represents an error due to an invalid input argument.
	ErrInvalidArgument
	// ErrNotFound represents an error when a requested resource is not found.
	ErrNotFound
	// ErrAlreadyExists represents an error when a resource already exists.
	ErrAlreadyExists
	// ErrPermissionDenied represents an error due to insufficient permissions.
	ErrPermissionDenied
	// ErrNotSupported represents an error when a feature is not supported.
	ErrNotSupported
)

// SQLValueType represents the fundamental SQL data types supported by Guocedb.
// This helps in consistent type handling across parsing, analysis, execution, and storage.
type SQLValueType int

const (
	// SQLTypeUnknown represents an unknown SQL type.
	SQLTypeUnknown SQLValueType = iota
	// SQLTypeBoolean represents a BOOLEAN type.
	SQLTypeBoolean
	// SQLTypeTinyInt represents a TINYINT type.
	SQLTypeTinyInt
	// SQLTypeSmallInt represents a SMALLINT type.
	SQLTypeSmallInt
	// SQLTypeMediumInt represents a MEDIUMINT type.
	SQLTypeMediumInt
	// SQLTypeInt represents an INT type.
	SQLTypeInt
	// SQLTypeBigInt represents a BIGINT type.
	SQLTypeBigInt
	// SQLTypeFloat represents a FLOAT type.
	SQLTypeFloat
	// SQLTypeDouble represents a DOUBLE type.
	SQLTypeDouble
	// SQLTypeDecimal represents a DECIMAL type.
	SQLTypeDecimal
	// SQLTypeVarchar represents a VARCHAR type.
	SQLTypeVarchar
	// SQLTypeChar represents a CHAR type.
	SQLTypeChar
	// SQLTypeText represents a TEXT type.
	SQLTypeText
	// SQLTypeBinary represents a BINARY type.
	SQLTypeBinary
	// SQLTypeVarBinary represents a VARBINARY type.
	SQLTypeVarBinary
	// SQLTypeBlob represents a BLOB type.
	SQLTypeBlob
	// SQLTypeDate represents a DATE type.
	SQLTypeDate
	// SQLTypeTime represents a TIME type.
	SQLTypeTime
	// SQLTypeDatetime represents a DATETIME type.
	SQLTypeDatetime
	// SQLTypeTimestamp represents a TIMESTAMP type.
	SQLTypeTimestamp
	// SQLTypeYear represents a YEAR type.
	SQLTypeYear
	// SQLTypeJSON represents a JSON type.
	SQLTypeJSON
	// SQLTypeEnum represents an ENUM type.
	SQLTypeEnum
	// SQLTypeSet represents a SET type.
	SQLTypeSet
)

// LogLevel represents the severity level for log messages.
// It is used by the unified logging interface (common/log/logger.go) to filter and format logs.
type LogLevel int

const (
	// LogLevelDebug represents detailed debugging information.
	LogLevelDebug LogLevel = iota
	// LogLevelInfo represents important operational information.
	LogLevelInfo
	// LogLevelWarn represents potentially harmful situations.
	LogLevelWarn
	// LogLevelError represents error events that might still allow the application to continue.
	LogLevelError
	// LogLevelFatal represents very severe error events that will presumably lead the application to abort.
	LogLevelFatal
)

// StorageEngineType represents the type of storage engines supported by Guocedb.
// This enum allows for plugin-based storage engine integration in the storage layer.
type StorageEngineType int

const (
	// StorageEngineUnknown represents an unknown storage engine type.
	StorageEngineUnknown StorageEngineType = iota
	// StorageEngineBadger represents the Badger KV store engine.
	StorageEngineBadger
	// StorageEngineKVD represents a generic Key-Value Data engine placeholder.
	StorageEngineKVD
	// StorageEngineMDD represents a generic Multi-Dimensional Data engine placeholder.
	StorageEngineMDD
	// StorageEngineMDI represents a generic Multi-Dimensional Index engine placeholder.
	StorageEngineMDI
)

// TransactionState represents the current state of a database transaction.
// Used by the transaction manager (compute/transaction/manager.go) for lifecycle management.
type TransactionState int

const (
	// TxnStateActive represents an active transaction.
	TxnStateActive TransactionState = iota
	// TxnStateCommitting represents a transaction in the process of committing.
	TxnStateCommitting
	// TxnStateCommitted represents a successfully committed transaction.
	TxnStateCommitted
	// TxnStateAborting represents a transaction in the process of aborting.
	TxnStateAborting
	// TxnStateAborted represents an aborted transaction.
	TxnStateAborted
	// TxnStateRollbacking represents a transaction in the process of rolling back.
	TxnStateRollbacking
	// TxnStateRolledBack represents a successfully rolled back transaction.
	TxnStateRolledBack
)

// MySQLCommandType represents the type of MySQL commands.
// Used by the MySQL protocol handler (protocol/mysql/handler.go) to differentiate incoming client requests.
type MySQLCommandType int

const (
	// ComQuery represents a SQL query command.
	ComQuery MySQLCommandType = iota + 3 // Start from 3 as per MySQL protocol spec
	// ComInitDB represents a command to change the default database.
	ComInitDB
	// ComPing represents a ping command.
	ComPing
	// ComQuit represents a quit command.
	ComQuit
	// ComStatistics represents a statistics command.
	ComStatistics
	// ComProcessInfo represents a process info command.
	ComProcessInfo
	// ComConnectOut represents a connect out command.
	ComConnectOut
	// ComChangeUser represents a change user command.
	ComChangeUser
	// ComDaemon represents a daemon command.
	ComDaemon
	// ComResetConnection represents a reset connection command.
	ComResetConnection
)

// AuthMethodType represents the authentication method used for a user.
// Used by the authentication module (security/authn/authn.go) to determine how to verify credentials.
type AuthMethodType int

const (
	// AuthMethodNative represents MySQL native password authentication.
	AuthMethodNative AuthMethodType = iota
	// AuthMethodSHA256 represents SHA256 password authentication.
	AuthMethodSHA256
	// AuthMethodCleartext represents cleartext password authentication (not recommended for production).
	AuthMethodCleartext
	// AuthMethodLDAP represents LDAP authentication.
	AuthMethodLDAP
)

// PermissionType represents the type of permission granted to a user or role.
// Used by the authorization module (security/authz/authz.go) to manage access control.
type PermissionType int

const (
	// PermSelect represents SELECT privilege.
	PermSelect PermissionType = iota
	// PermInsert represents INSERT privilege.
	PermInsert
	// PermUpdate represents UPDATE privilege.
	PermUpdate
	// PermDelete represents DELETE privilege.
	PermDelete
	// PermCreate represents CREATE privilege.
	PermCreate
	// PermDrop represents DROP privilege.
	PermDrop
	// PermAlter represents ALTER privilege.
	PermAlter
	// PermGrant represents GRANT privilege.
	PermGrant
	// PermRevoke represents REVOKE privilege.
	PermRevoke
	// PermAll represents all privileges.
	PermAll
)

// SystemComponentType represents the various components within the Guocedb system.
// Used for logging, metrics, and diagnostics to identify the source of operations or issues.
type SystemComponentType int

const (
	// ComponentUnknown represents an unknown component.
	ComponentUnknown SystemComponentType = iota
	// ComponentServer represents the main database server.
	ComponentServer
	// ComponentClient represents the command-line client.
	ComponentClient
	// ComponentConfig represents the configuration module.
	ComponentConfig
	// ComponentErrors represents the error handling module.
	ComponentErrors
	// ComponentLogger represents the logging module.
	ComponentLogger
	// ComponentTypes represents the basic types module.
	ComponentTypes
	// ComponentParser represents the SQL parser.
	ComponentParser
	// ComponentAnalyzer represents the SQL analyzer.
	ComponentAnalyzer
	// ComponentOptimizer represents the query optimizer.
	ComponentOptimizer
	// ComponentExecutor represents the query execution engine.
	ComponentExecutor
	// ComponentScheduler represents the distributed scheduler.
	ComponentScheduler
	// ComponentTransaction represents the transaction manager.
	ComponentTransaction
	// ComponentCatalog represents the metadata catalog.
	ComponentCatalog
	// ComponentStorage represents the storage abstraction layer.
	ComponentStorage
	// ComponentBadgerEngine represents the Badger storage engine.
	ComponentBadgerEngine
	// ComponentNetwork represents the network layer.
	ComponentNetwork
	// ComponentMySQLProtocol represents the MySQL protocol handler.
	ComponentMySQLProtocol
	// ComponentAuthentication represents the authentication module.
	ComponentAuthentication
	// ComponentAuthorization represents the authorization module.
	ComponentAuthorization
	// ComponentEncryption represents the encryption module.
	ComponentEncryption
	// ComponentAudit represents the audit logging module.
	ComponentAudit
	// ComponentMetrics represents the performance metrics module.
	ComponentMetrics
	// ComponentStatus represents the status reporting module.
	ComponentStatus
	// ComponentDiagnostic represents the diagnostic tools module.
	ComponentDiagnostic
	// ComponentAPI represents the external API layer.
	ComponentAPI
)

//Personal.AI order the ending
