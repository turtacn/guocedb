package authz

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/security/authn"
)

// Action represents a type of action a user can perform.
type Action string

const (
	SelectAction Action = "SELECT"
	InsertAction Action = "INSERT"
	UpdateAction Action = "UPDATE"
	DeleteAction Action = "DELETE"
	CreateAction Action = "CREATE"
	DropAction   Action = "DROP"
	AlterAction  Action = "ALTER"
)

// Resource represents something that can be acted upon.
type Resource struct {
	Type     string // e.g., "database", "table"
	Name     string
	ParentDB string // The database this resource belongs to, if applicable
}

// Authorizer is the interface for checking user permissions.
type Authorizer interface {
	// CheckPermission verifies if a user has permission to perform an action on a resource.
	CheckPermission(ctx *sql.Context, user *authn.User, action Action, resource Resource) (bool, error)
}

// PermissiveAuthorizer is a placeholder implementation that allows all actions.
// This is useful for development but should be replaced for production.
type PermissiveAuthorizer struct{}

// NewPermissiveAuthorizer creates a new authorizer that grants all permissions.
func NewPermissiveAuthorizer() Authorizer {
	return &PermissiveAuthorizer{}
}

// CheckPermission always returns true for the permissive authorizer.
func (a *PermissiveAuthorizer) CheckPermission(ctx *sql.Context, user *authn.User, action Action, resource Resource) (bool, error) {
	// The root user is always allowed.
	if user != nil && user.Name == "root" {
		return true, nil
	}
	// For development, allow all other actions as well.
	return true, nil
}

// TODO: Implement a full Role-Based Access Control (RBAC) system.
// This would involve:
// - Defining roles and privileges.
// - Storing user/role mappings and privilege grants in the storage engine.
// - Implementing logic to check the user's roles and privileges against the requested action/resource.
// TODO: Support column-level and row-level security policies.
