package enum

// StorageEngineType defines the enumeration for storage engine types.
type StorageEngineType int

const (
	BadgerDB StorageEngineType = iota
	KVD
	MDD
	MDI
)

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

// TransactionIsolationLevel defines the enumeration for transaction isolation levels.
type TransactionIsolationLevel int

const (
	ReadUncommitted TransactionIsolationLevel = iota
	ReadCommitted
	RepeatableRead
	Serializable
)

func (l TransactionIsolationLevel) String() string {
	return [...]string{"ReadUncommitted", "ReadCommitted", "RepeatableRead", "Serializable"}[l]
}

// QueryType defines the enumeration for SQL query types.
type QueryType int

const (
	SELECT QueryType = iota
	INSERT
	UPDATE
	DELETE
	CREATE
	DROP
	ALTER
)

func (q QueryType) String() string {
	return [...]string{"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "DROP", "ALTER"}[q]
}

// SystemStatus defines the enumeration for system status.
type SystemStatus int

const (
	Starting SystemStatus = iota
	Running
	Stopped
	Error
)

func (s SystemStatus) String() string {
	return [...]string{"Starting", "Running", "Stopped", "Error"}[s]
}
