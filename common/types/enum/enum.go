// Package enum defines global enumeration types used throughout guocedb.
package enum

// StorageEngineType represents the type of storage engine.
type StorageEngineType int

const (
	// BadgerDB is a key-value store written in Go.
	BadgerDB StorageEngineType = iota
	// KVD is a placeholder for the KVD storage engine.
	KVD
	// MDD is a placeholder for the MDD storage engine.
	MDD
	// MDI is a placeholder for the MDI storage engine.
	MDI
)

// String returns the string representation of a StorageEngineType.
func (s StorageEngineType) String() string {
	switch s {
	case BadgerDB:
		return "BadgerDB"
	case KVD:
		return "KVD"
	case MDD:
		return "MDD"
	case MDI:
		return "MDI"
	default:
		return "Unknown"
	}
}

// TransactionIsolationLevel represents the four standard transaction isolation levels.
type TransactionIsolationLevel int

const (
	// ReadUncommitted allows dirty reads.
	ReadUncommitted TransactionIsolationLevel = iota
	// ReadCommitted prevents dirty reads.
	ReadCommitted
	// RepeatableRead prevents non-repeatable reads.
	RepeatableRead
	// Serializable prevents phantom reads.
	Serializable
)

// String returns the string representation of a TransactionIsolationLevel.
func (t TransactionIsolationLevel) String() string {
	switch t {
	case ReadUncommitted:
		return "READ UNCOMMITTED"
	case ReadCommitted:
		return "READ COMMITTED"
	case RepeatableRead:
		return "REPEATABLE READ"
	case Serializable:
		return "SERIALIZABLE"
	default:
		return "Unknown"
	}
}

// QueryType represents the type of a SQL query.
type QueryType int

const (
	// SELECT represents a SELECT query.
	SELECT QueryType = iota
	// INSERT represents an INSERT query.
	INSERT
	// UPDATE represents an UPDATE query.
	UPDATE
	// DELETE represents a DELETE query.
	DELETE
	// DDL represents a Data Definition Language query (e.g., CREATE, ALTER, DROP).
	DDL
	// DML represents a Data Manipulation Language query (e.g., INSERT, UPDATE, DELETE).
	DML
)

// String returns the string representation of a QueryType.
func (q QueryType) String() string {
	switch q {
	case SELECT:
		return "SELECT"
	case INSERT:
		return "INSERT"
	case UPDATE:
		return "UPDATE"
	case DELETE:
		return "DELETE"
	case DDL:
		return "DDL"
	case DML:
		return "DML"
	default:
		return "Unknown"
	}
}

// SystemStatus represents the operational status of the guocedb system.
type SystemStatus int

const (
	// Running indicates the system is fully operational.
	Running SystemStatus = iota
	// Stopped indicates the system is not running.
	Stopped
	// Error indicates the system is in an error state.
	Error
	// Starting indicates the system is in the process of starting up.
	Starting
	// Stopping indicates the system is in the process of shutting down.
	Stopping
)

// String returns the string representation of a SystemStatus.
func (s SystemStatus) String() string {
	switch s {
	case Running:
		return "Running"
	case Stopped:
		return "Stopped"
	case Error:
		return "Error"
	case Starting:
		return "Starting"
	case Stopping:
		return "Stopping"
	default:
		return "Unknown"
	}
}
