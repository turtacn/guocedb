// Package authz provides authorization services for GuoceDB.
package authz

import (
	"context"
	"strings"
	"sync"
)

// Role represents a set of privileges.
type Role struct {
	Name       string
	Privileges Privilege
	
	// Fine-grained privileges (optional)
	DatabasePrivileges map[string]Privilege              // db -> privilege
	TablePrivileges    map[string]map[string]Privilege // db -> table -> privilege
}

// RoleStore is the interface for role persistence.
type RoleStore interface {
	GetRole(ctx context.Context, name string) (*Role, error)
	CreateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, name string) error
	ListRoles(ctx context.Context) ([]*Role, error)
}

// PredefinedRoles contains commonly used roles.
var PredefinedRoles = map[string]*Role{
	"admin": {
		Name:       "admin",
		Privileges: PrivilegeAll,
	},
	"readwrite": {
		Name:       "readwrite",
		Privileges: PrivilegeReadWrite,
	},
	"readonly": {
		Name:       "readonly",
		Privileges: PrivilegeReadOnly,
	},
	"ddladmin": {
		Name:       "ddladmin",
		Privileges: PrivilegeDDL,
	},
}

// InMemoryRoleStore is an in-memory implementation of RoleStore.
type InMemoryRoleStore struct {
	mu    sync.RWMutex
	roles map[string]*Role
}

// NewInMemoryRoleStore creates a new in-memory role store with predefined roles.
func NewInMemoryRoleStore() *InMemoryRoleStore {
	store := &InMemoryRoleStore{
		roles: make(map[string]*Role),
	}
	
	// Load predefined roles
	for name, role := range PredefinedRoles {
		// Create a copy to avoid sharing references
		roleCopy := &Role{
			Name:       role.Name,
			Privileges: role.Privileges,
		}
		if role.DatabasePrivileges != nil {
			roleCopy.DatabasePrivileges = make(map[string]Privilege)
			for db, priv := range role.DatabasePrivileges {
				roleCopy.DatabasePrivileges[db] = priv
			}
		}
		if role.TablePrivileges != nil {
			roleCopy.TablePrivileges = make(map[string]map[string]Privilege)
			for db, tables := range role.TablePrivileges {
				roleCopy.TablePrivileges[db] = make(map[string]Privilege)
				for table, priv := range tables {
					roleCopy.TablePrivileges[db][table] = priv
				}
			}
		}
		store.roles[name] = roleCopy
	}
	
	return store
}

// GetRole retrieves a role by name (case-insensitive).
func (s *InMemoryRoleStore) GetRole(ctx context.Context, name string) (*Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	role, ok := s.roles[strings.ToLower(name)]
	if !ok {
		return nil, nil
	}
	
	// Return a copy to prevent external modification
	roleCopy := &Role{
		Name:       role.Name,
		Privileges: role.Privileges,
	}
	
	if role.DatabasePrivileges != nil {
		roleCopy.DatabasePrivileges = make(map[string]Privilege)
		for db, priv := range role.DatabasePrivileges {
			roleCopy.DatabasePrivileges[db] = priv
		}
	}
	
	if role.TablePrivileges != nil {
		roleCopy.TablePrivileges = make(map[string]map[string]Privilege)
		for db, tables := range role.TablePrivileges {
			roleCopy.TablePrivileges[db] = make(map[string]Privilege)
			for table, priv := range tables {
				roleCopy.TablePrivileges[db][table] = priv
			}
		}
	}
	
	return roleCopy, nil
}

// CreateRole adds a new role to the store.
func (s *InMemoryRoleStore) CreateRole(ctx context.Context, role *Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	name := strings.ToLower(role.Name)
	if _, exists := s.roles[name]; exists {
		return ErrRoleAlreadyExists
	}
	
	s.roles[name] = role
	return nil
}

// DeleteRole removes a role from the store.
func (s *InMemoryRoleStore) DeleteRole(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.roles, strings.ToLower(name))
	return nil
}

// ListRoles returns all roles in the store.
func (s *InMemoryRoleStore) ListRoles(ctx context.Context) ([]*Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	roles := make([]*Role, 0, len(s.roles))
	for _, role := range s.roles {
		roleCopy := &Role{
			Name:       role.Name,
			Privileges: role.Privileges,
		}
		if role.DatabasePrivileges != nil {
			roleCopy.DatabasePrivileges = make(map[string]Privilege)
			for db, priv := range role.DatabasePrivileges {
				roleCopy.DatabasePrivileges[db] = priv
			}
		}
		if role.TablePrivileges != nil {
			roleCopy.TablePrivileges = make(map[string]map[string]Privilege)
			for db, tables := range role.TablePrivileges {
				roleCopy.TablePrivileges[db] = make(map[string]Privilege)
				for table, priv := range tables {
					roleCopy.TablePrivileges[db][table] = priv
				}
			}
		}
		roles = append(roles, roleCopy)
	}
	
	return roles, nil
}

var (
	ErrRoleAlreadyExists = &AuthzError{Code: "ROLE_EXISTS", Message: "role already exists"}
	ErrRoleNotFound      = &AuthzError{Code: "ROLE_NOT_FOUND", Message: "role not found"}
)

// AuthzError represents an authorization error.
type AuthzError struct {
	Code    string
	Message string
}

func (e *AuthzError) Error() string {
	return e.Code + ": " + e.Message
}
