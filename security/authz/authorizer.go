// Package authz provides authorization services for GuoceDB.
package authz

import (
	"context"
)

// Authorizer handles privilege checking for users.
type Authorizer struct {
	roleStore RoleStore
}

// NewAuthorizer creates a new authorizer with the given role store.
func NewAuthorizer(roleStore RoleStore) *Authorizer {
	return &Authorizer{
		roleStore: roleStore,
	}
}

// User interface for authorization - we only need username, roles, and privileges.
type User interface {
	GetUsername() string
	GetRoles() []string
	GetPrivileges() Privilege
}

// CheckPrivilege verifies if a user has the required privilege on a resource.
func (a *Authorizer) CheckPrivilege(ctx context.Context, user User, database, table string, required Privilege) error {
	// Super users bypass all checks
	if user.GetPrivileges().Has(PrivilegeAdmin) {
		return nil
	}
	
	// Check user's direct privileges
	if user.GetPrivileges().Has(required) {
		return nil
	}
	
	// Check role-based privileges
	for _, roleName := range user.GetRoles() {
		role, err := a.roleStore.GetRole(ctx, roleName)
		if err != nil || role == nil {
			continue
		}
		
		// Check global role privileges
		if role.Privileges.Has(required) {
			return nil
		}
		
		// Check database-level privileges
		if role.DatabasePrivileges != nil {
			if dbPriv, ok := role.DatabasePrivileges[database]; ok {
				if dbPriv.Has(required) {
					return nil
				}
			}
		}
		
		// Check table-level privileges
		if role.TablePrivileges != nil {
			if dbTables, ok := role.TablePrivileges[database]; ok {
				if tablePriv, ok := dbTables[table]; ok {
					if tablePriv.Has(required) {
						return nil
					}
				}
			}
		}
	}
	
	return ErrAccessDenied
}

// CheckPrivileges performs multiple privilege checks in batch.
func (a *Authorizer) CheckPrivileges(ctx context.Context, user User, checks []PrivilegeCheck) error {
	for _, check := range checks {
		if err := a.CheckPrivilege(ctx, user, check.Database, check.Table, check.Privilege); err != nil {
			return err
		}
	}
	return nil
}

// PrivilegeCheck represents a single privilege check request.
type PrivilegeCheck struct {
	Database  string
	Table     string
	Privilege Privilege
}

var (
	ErrAccessDenied = &AuthzError{Code: "ACCESS_DENIED", Message: "access denied"}
)
