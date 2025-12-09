// Package authz provides authorization services for GuoceDB.
package authz

import "strings"

// Privilege represents a database privilege using bitflags.
type Privilege uint32

const (
	// PrivilegeNone represents no privileges
	PrivilegeNone Privilege = 0
	
	// Basic DML privileges
	PrivilegeSelect Privilege = 1 << iota
	PrivilegeInsert
	PrivilegeUpdate
	PrivilegeDelete
	
	// DDL privileges
	PrivilegeCreate
	PrivilegeDrop
	PrivilegeAlter
	PrivilegeIndex
	
	// Administrative privileges
	PrivilegeGrant
	PrivilegeAdmin // Super user privilege
	
	// Composite privileges
	PrivilegeReadOnly  = PrivilegeSelect
	PrivilegeReadWrite = PrivilegeSelect | PrivilegeInsert | PrivilegeUpdate | PrivilegeDelete
	PrivilegeDDL       = PrivilegeCreate | PrivilegeDrop | PrivilegeAlter | PrivilegeIndex
	PrivilegeAll       = PrivilegeReadWrite | PrivilegeDDL | PrivilegeGrant | PrivilegeAdmin
)

// Has checks if the privilege set contains the specified privilege.
func (p Privilege) Has(check Privilege) bool {
	return p&check == check
}

// String returns a human-readable representation of the privilege.
func (p Privilege) String() string {
	if p == PrivilegeNone {
		return "NONE"
	}
	if p == PrivilegeAll {
		return "ALL"
	}
	
	var parts []string
	if p.Has(PrivilegeSelect) {
		parts = append(parts, "SELECT")
	}
	if p.Has(PrivilegeInsert) {
		parts = append(parts, "INSERT")
	}
	if p.Has(PrivilegeUpdate) {
		parts = append(parts, "UPDATE")
	}
	if p.Has(PrivilegeDelete) {
		parts = append(parts, "DELETE")
	}
	if p.Has(PrivilegeCreate) {
		parts = append(parts, "CREATE")
	}
	if p.Has(PrivilegeDrop) {
		parts = append(parts, "DROP")
	}
	if p.Has(PrivilegeAlter) {
		parts = append(parts, "ALTER")
	}
	if p.Has(PrivilegeIndex) {
		parts = append(parts, "INDEX")
	}
	if p.Has(PrivilegeGrant) {
		parts = append(parts, "GRANT")
	}
	if p.Has(PrivilegeAdmin) {
		parts = append(parts, "ADMIN")
	}
	
	if len(parts) == 0 {
		return "UNKNOWN"
	}
	return strings.Join(parts, ",")
}

// ParsePrivilege converts a string to a Privilege.
func ParsePrivilege(s string) Privilege {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "SELECT":
		return PrivilegeSelect
	case "INSERT":
		return PrivilegeInsert
	case "UPDATE":
		return PrivilegeUpdate
	case "DELETE":
		return PrivilegeDelete
	case "CREATE":
		return PrivilegeCreate
	case "DROP":
		return PrivilegeDrop
	case "ALTER":
		return PrivilegeAlter
	case "INDEX":
		return PrivilegeIndex
	case "GRANT":
		return PrivilegeGrant
	case "ADMIN":
		return PrivilegeAdmin
	case "ALL":
		return PrivilegeAll
	default:
		return PrivilegeNone
	}
}
