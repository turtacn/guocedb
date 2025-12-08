package transaction

import (
	"fmt"
	"strings"
)

// IsolationLevel represents the transaction isolation level
type IsolationLevel int

const (
	// LevelReadUncommitted allows dirty reads, non-repeatable reads, and phantom reads
	LevelReadUncommitted IsolationLevel = iota
	// LevelReadCommitted prevents dirty reads but allows non-repeatable reads and phantom reads
	LevelReadCommitted
	// LevelRepeatableRead prevents dirty reads and non-repeatable reads but allows phantom reads
	LevelRepeatableRead
	// LevelSerializable prevents all phenomena (dirty reads, non-repeatable reads, phantom reads)
	LevelSerializable
)

// String returns the string representation of the isolation level
func (l IsolationLevel) String() string {
	switch l {
	case LevelReadUncommitted:
		return "READ UNCOMMITTED"
	case LevelReadCommitted:
		return "READ COMMITTED"
	case LevelRepeatableRead:
		return "REPEATABLE READ"
	case LevelSerializable:
		return "SERIALIZABLE"
	default:
		return "UNKNOWN"
	}
}

// ParseIsolationLevel parses a string into an IsolationLevel
func ParseIsolationLevel(s string) (IsolationLevel, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "READ UNCOMMITTED":
		return LevelReadUncommitted, nil
	case "READ COMMITTED":
		return LevelReadCommitted, nil
	case "REPEATABLE READ":
		return LevelRepeatableRead, nil
	case "SERIALIZABLE":
		return LevelSerializable, nil
	default:
		return 0, fmt.Errorf("unknown isolation level: %s", s)
	}
}