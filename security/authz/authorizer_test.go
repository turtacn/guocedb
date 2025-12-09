package authz

import (
	"context"
	"testing"
)

// mockUser implements the User interface for testing
type mockUser struct {
	username   string
	roles      []string
	privileges Privilege
}

func (m *mockUser) GetUsername() string     { return m.username }
func (m *mockUser) GetRoles() []string      { return m.roles }
func (m *mockUser) GetPrivileges() Privilege { return m.privileges }

func TestCheckPrivilegeGranted(t *testing.T) {
	ctx := context.Background()
	roleStore := NewInMemoryRoleStore()
	authz := NewAuthorizer(roleStore)
	
	user := &mockUser{
		username: "reader",
		roles:    []string{"readonly"},
	}
	
	err := authz.CheckPrivilege(ctx, user, "testdb", "users", PrivilegeSelect)
	if err != nil {
		t.Errorf("Should have SELECT privilege: %v", err)
	}
}

func TestCheckPrivilegeDenied(t *testing.T) {
	ctx := context.Background()
	roleStore := NewInMemoryRoleStore()
	authz := NewAuthorizer(roleStore)
	
	user := &mockUser{
		username: "reader",
		roles:    []string{"readonly"},
	}
	
	err := authz.CheckPrivilege(ctx, user, "testdb", "users", PrivilegeDelete)
	if err != ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied, got %v", err)
	}
}

func TestAdminBypassCheck(t *testing.T) {
	ctx := context.Background()
	roleStore := NewInMemoryRoleStore()
	authz := NewAuthorizer(roleStore)
	
	user := &mockUser{
		username:   "admin",
		privileges: PrivilegeAdmin,
	}
	
	// Admin should pass all privilege checks
	err := authz.CheckPrivilege(ctx, user, "anydb", "anytable", PrivilegeAll)
	if err != nil {
		t.Errorf("Admin should bypass all checks: %v", err)
	}
}

func TestCheckPrivilegeWithDirectPrivilege(t *testing.T) {
	ctx := context.Background()
	roleStore := NewInMemoryRoleStore()
	authz := NewAuthorizer(roleStore)
	
	user := &mockUser{
		username:   "testuser",
		privileges: PrivilegeSelect | PrivilegeInsert,
		roles:      []string{},
	}
	
	// Should have SELECT via direct privilege
	err := authz.CheckPrivilege(ctx, user, "testdb", "users", PrivilegeSelect)
	if err != nil {
		t.Errorf("Should have SELECT via direct privilege: %v", err)
	}
	
	// Should have INSERT via direct privilege
	err = authz.CheckPrivilege(ctx, user, "testdb", "users", PrivilegeInsert)
	if err != nil {
		t.Errorf("Should have INSERT via direct privilege: %v", err)
	}
	
	// Should not have DELETE
	err = authz.CheckPrivilege(ctx, user, "testdb", "users", PrivilegeDelete)
	if err != ErrAccessDenied {
		t.Errorf("Should not have DELETE privilege: %v", err)
	}
}

func TestCheckPrivilegeReadWriteRole(t *testing.T) {
	ctx := context.Background()
	roleStore := NewInMemoryRoleStore()
	authz := NewAuthorizer(roleStore)
	
	user := &mockUser{
		username: "writer",
		roles:    []string{"readwrite"},
	}
	
	// Should have all DML privileges
	for _, priv := range []Privilege{PrivilegeSelect, PrivilegeInsert, PrivilegeUpdate, PrivilegeDelete} {
		err := authz.CheckPrivilege(ctx, user, "testdb", "users", priv)
		if err != nil {
			t.Errorf("Should have %v privilege: %v", priv, err)
		}
	}
	
	// Should not have DDL privileges
	err := authz.CheckPrivilege(ctx, user, "testdb", "users", PrivilegeCreate)
	if err != ErrAccessDenied {
		t.Errorf("Should not have CREATE privilege: %v", err)
	}
}

func TestCheckPrivilegesMultiple(t *testing.T) {
	ctx := context.Background()
	roleStore := NewInMemoryRoleStore()
	authz := NewAuthorizer(roleStore)
	
	user := &mockUser{
		username: "writer",
		roles:    []string{"readwrite"},
	}
	
	checks := []PrivilegeCheck{
		{Database: "testdb", Table: "users", Privilege: PrivilegeSelect},
		{Database: "testdb", Table: "orders", Privilege: PrivilegeInsert},
		{Database: "testdb", Table: "products", Privilege: PrivilegeUpdate},
	}
	
	err := authz.CheckPrivileges(ctx, user, checks)
	if err != nil {
		t.Errorf("All checks should pass: %v", err)
	}
}

func TestCheckPrivilegesMultipleFail(t *testing.T) {
	ctx := context.Background()
	roleStore := NewInMemoryRoleStore()
	authz := NewAuthorizer(roleStore)
	
	user := &mockUser{
		username: "reader",
		roles:    []string{"readonly"},
	}
	
	checks := []PrivilegeCheck{
		{Database: "testdb", Table: "users", Privilege: PrivilegeSelect},
		{Database: "testdb", Table: "orders", Privilege: PrivilegeDelete}, // Should fail
	}
	
	err := authz.CheckPrivileges(ctx, user, checks)
	if err != ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied, got %v", err)
	}
}

func TestCheckPrivilegeDatabaseLevel(t *testing.T) {
	ctx := context.Background()
	roleStore := NewInMemoryRoleStore()
	
	// Create custom role with database-level privileges
	role := &Role{
		Name:       "db_specific",
		Privileges: PrivilegeNone,
		DatabasePrivileges: map[string]Privilege{
			"testdb": PrivilegeSelect | PrivilegeInsert,
		},
	}
	roleStore.CreateRole(ctx, role)
	
	authz := NewAuthorizer(roleStore)
	user := &mockUser{
		username: "dbuser",
		roles:    []string{"db_specific"},
	}
	
	// Should have SELECT on testdb
	err := authz.CheckPrivilege(ctx, user, "testdb", "users", PrivilegeSelect)
	if err != nil {
		t.Errorf("Should have SELECT on testdb: %v", err)
	}
	
	// Should not have DELETE on testdb
	err = authz.CheckPrivilege(ctx, user, "testdb", "users", PrivilegeDelete)
	if err != ErrAccessDenied {
		t.Errorf("Should not have DELETE on testdb: %v", err)
	}
	
	// Should not have any privilege on otherdb
	err = authz.CheckPrivilege(ctx, user, "otherdb", "users", PrivilegeSelect)
	if err != ErrAccessDenied {
		t.Errorf("Should not have privileges on otherdb: %v", err)
	}
}

func TestCheckPrivilegeTableLevel(t *testing.T) {
	ctx := context.Background()
	roleStore := NewInMemoryRoleStore()
	
	// Create custom role with table-level privileges
	role := &Role{
		Name:       "table_specific",
		Privileges: PrivilegeNone,
		TablePrivileges: map[string]map[string]Privilege{
			"testdb": {
				"users": PrivilegeSelect,
			},
		},
	}
	roleStore.CreateRole(ctx, role)
	
	authz := NewAuthorizer(roleStore)
	user := &mockUser{
		username: "tableuser",
		roles:    []string{"table_specific"},
	}
	
	// Should have SELECT on testdb.users
	err := authz.CheckPrivilege(ctx, user, "testdb", "users", PrivilegeSelect)
	if err != nil {
		t.Errorf("Should have SELECT on testdb.users: %v", err)
	}
	
	// Should not have SELECT on testdb.orders
	err = authz.CheckPrivilege(ctx, user, "testdb", "orders", PrivilegeSelect)
	if err != ErrAccessDenied {
		t.Errorf("Should not have SELECT on testdb.orders: %v", err)
	}
}

func TestCheckPrivilegeMultipleRoles(t *testing.T) {
	ctx := context.Background()
	roleStore := NewInMemoryRoleStore()
	authz := NewAuthorizer(roleStore)
	
	user := &mockUser{
		username: "multiuser",
		roles:    []string{"readonly", "ddladmin"},
	}
	
	// Should have SELECT from readonly role
	err := authz.CheckPrivilege(ctx, user, "testdb", "users", PrivilegeSelect)
	if err != nil {
		t.Errorf("Should have SELECT from readonly role: %v", err)
	}
	
	// Should have CREATE from ddladmin role
	err = authz.CheckPrivilege(ctx, user, "testdb", "users", PrivilegeCreate)
	if err != nil {
		t.Errorf("Should have CREATE from ddladmin role: %v", err)
	}
	
	// Should not have DELETE (not in either role)
	err = authz.CheckPrivilege(ctx, user, "testdb", "users", PrivilegeDelete)
	if err != ErrAccessDenied {
		t.Errorf("Should not have DELETE: %v", err)
	}
}
