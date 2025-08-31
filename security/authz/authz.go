// Package authz provides authorization services for guocedb.
package authz

import (
	"fmt"
	"sync"

	"github.com/turtacn/guocedb/security/authn"
)

// Permission is a string representing an action (e.g., "SELECT", "INSERT").
type Permission string

const (
	PermSelect Permission = "SELECT"
	PermInsert Permission = "INSERT"
	PermUpdate Permission = "UPDATE"
	PermDelete Permission = "DELETE"
	PermCreate Permission = "CREATE"
	PermDrop   Permission = "DROP"
	PermAll    Permission = "ALL"
)

// Role defines a set of permissions.
type Role struct {
	Name        string
	Permissions map[string][]Permission // resource -> permissions
}

// Authorizer is the interface for authorization checks.
type Authorizer interface {
	// Can checks if a user has a specific permission on a resource.
	Can(user *authn.User, perm Permission, resource string) (bool, error)
	// Grant gives a permission to a role on a resource.
	Grant(roleName string, perm Permission, resource string) error
	// Revoke removes a permission from a role on a resource.
	Revoke(roleName string, perm Permission, resource string) error
}

// MemoryAuthorizer is a simple in-memory authorizer.
// A real implementation would persist roles and permissions.
type MemoryAuthorizer struct {
	mu    sync.RWMutex
	roles map[string]*Role
	// This maps a user to a set of role names.
	userRoles map[string][]string
}

// NewMemoryAuthorizer creates a new in-memory authorizer.
func NewMemoryAuthorizer() *MemoryAuthorizer {
	// Create a default admin role with all permissions on all resources
	adminRole := &Role{
		Name: "admin",
		Permissions: map[string][]Permission{
			"*": {PermAll}, // "*" is a wildcard for all resources
		},
	}
	roles := map[string]*Role{"admin": adminRole}

	// Assign the "root" user to the "admin" role
	userRoles := map[string][]string{
		"root": {"admin"},
	}

	return &MemoryAuthorizer{
		roles:     roles,
		userRoles: userRoles,
	}
}

// Can checks if a user has the required permission.
func (a *MemoryAuthorizer) Can(user *authn.User, perm Permission, resource string) (bool, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	userRoleNames, ok := a.userRoles[user.Name]
	if !ok {
		return false, nil // User has no roles
	}

	for _, roleName := range userRoleNames {
		role, ok := a.roles[roleName]
		if !ok {
			continue
		}

		// Check for specific resource permissions
		if perms, ok := role.Permissions[resource]; ok {
			for _, p := range perms {
				if p == perm || p == PermAll {
					return true, nil
				}
			}
		}

		// Check for wildcard resource permissions
		if perms, ok := role.Permissions["*"]; ok {
			for _, p := range perms {
				if p == perm || p == PermAll {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func (a *MemoryAuthorizer) Grant(roleName string, perm Permission, resource string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Implementation to grant permission
	return fmt.Errorf("not implemented")
}

func (a *MemoryAuthorizer) Revoke(roleName string, perm Permission, resource string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Implementation to revoke permission
	return fmt.Errorf("not implemented")
}

// Enforce interface compliance
var _ Authorizer = (*MemoryAuthorizer)(nil)
