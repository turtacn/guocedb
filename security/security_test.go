package security

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/turtacn/guocedb/security/audit"
	"github.com/turtacn/guocedb/security/authz"
)

func TestSecurityManagerFlow(t *testing.T) {
	tmpDir := t.TempDir()
	auditFile := filepath.Join(tmpDir, "audit.log")
	
	config := SecurityConfig{
		Enabled: true,
		AuditConfig: audit.AuditConfig{
			FilePath: auditFile,
			Async:    false,
		},
		MaxAuthFails: 5,
		LockDuration: 15 * time.Minute,
	}
	
	sm, err := NewSecurityManager(config)
	if err != nil {
		t.Fatalf("Failed to create SecurityManager: %v", err)
	}
	defer sm.Close()
	
	ctx := context.Background()
	
	// 1. Create user
	err = sm.CreateUser(ctx, "testuser", "password123", []string{"readonly"})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	// 2. Authenticate
	user, err := sm.Authenticate(ctx, "testuser", "password123", "127.0.0.1")
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}
	
	// 3. Check privilege - should have SELECT
	err = sm.CheckPrivilege(ctx, user, "testdb", "users", authz.PrivilegeSelect)
	if err != nil {
		t.Errorf("Should have SELECT privilege: %v", err)
	}
	
	// 4. Check privilege - should NOT have DELETE
	err = sm.CheckPrivilege(ctx, user, "testdb", "users", authz.PrivilegeDelete)
	if err != authz.ErrAccessDenied {
		t.Errorf("Should not have DELETE privilege, got: %v", err)
	}
	
	// 5. Verify audit log exists
	if _, err := os.Stat(auditFile); os.IsNotExist(err) {
		t.Error("Audit log file should exist")
	}
}

func TestSecurityManagerDisabled(t *testing.T) {
	config := SecurityConfig{
		Enabled: false,
	}
	
	sm, err := NewSecurityManager(config)
	if err != nil {
		t.Fatalf("Failed to create SecurityManager: %v", err)
	}
	
	if sm.IsEnabled() {
		t.Error("Security should be disabled")
	}
	
	ctx := context.Background()
	
	// Authentication should succeed with any credentials
	user, err := sm.Authenticate(ctx, "anyuser", "anypass", "127.0.0.1")
	if err != nil {
		t.Errorf("Authentication should succeed when disabled: %v", err)
	}
	if user == nil {
		t.Error("User should not be nil")
	}
	if !user.Privileges.Has(authz.PrivilegeAll) {
		t.Error("User should have all privileges when security is disabled")
	}
	
	// Authorization should always succeed
	err = sm.CheckPrivilege(ctx, user, "anydb", "anytable", authz.PrivilegeAll)
	if err != nil {
		t.Errorf("Authorization should succeed when disabled: %v", err)
	}
}

func TestSecurityManagerRoleGrant(t *testing.T) {
	config := SecurityConfig{
		Enabled: true,
		AuditConfig: audit.AuditConfig{
			FilePath: "stdout",
		},
	}
	
	sm, err := NewSecurityManager(config)
	if err != nil {
		t.Fatalf("Failed to create SecurityManager: %v", err)
	}
	defer sm.Close()
	
	ctx := context.Background()
	
	// Create user with no roles
	err = sm.CreateUser(ctx, "testuser", "password123", []string{})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	// User should not have DELETE privilege
	user, _ := sm.GetUser(ctx, "testuser")
	err = sm.CheckPrivilege(ctx, user, "testdb", "users", authz.PrivilegeDelete)
	if err != authz.ErrAccessDenied {
		t.Error("User should not have DELETE privilege initially")
	}
	
	// Grant readwrite role
	err = sm.GrantRole(ctx, "testuser", "readwrite")
	if err != nil {
		t.Fatalf("Failed to grant role: %v", err)
	}
	
	// Now user should have DELETE privilege
	user, _ = sm.GetUser(ctx, "testuser")
	err = sm.CheckPrivilege(ctx, user, "testdb", "users", authz.PrivilegeDelete)
	if err != nil {
		t.Errorf("User should have DELETE privilege after role grant: %v", err)
	}
}

func TestSecurityManagerRoleRevoke(t *testing.T) {
	config := SecurityConfig{
		Enabled: true,
		AuditConfig: audit.AuditConfig{
			FilePath: "stdout",
		},
	}
	
	sm, err := NewSecurityManager(config)
	if err != nil {
		t.Fatalf("Failed to create SecurityManager: %v", err)
	}
	defer sm.Close()
	
	ctx := context.Background()
	
	// Create user with readwrite role
	err = sm.CreateUser(ctx, "testuser", "password123", []string{"readwrite"})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	// User should have DELETE privilege
	user, _ := sm.GetUser(ctx, "testuser")
	err = sm.CheckPrivilege(ctx, user, "testdb", "users", authz.PrivilegeDelete)
	if err != nil {
		t.Error("User should have DELETE privilege initially")
	}
	
	// Revoke readwrite role
	err = sm.RevokeRole(ctx, "testuser", "readwrite")
	if err != nil {
		t.Fatalf("Failed to revoke role: %v", err)
	}
	
	// Now user should not have DELETE privilege
	user, _ = sm.GetUser(ctx, "testuser")
	err = sm.CheckPrivilege(ctx, user, "testdb", "users", authz.PrivilegeDelete)
	if err != authz.ErrAccessDenied {
		t.Error("User should not have DELETE privilege after role revoke")
	}
}

func TestSecurityManagerMultipleRoles(t *testing.T) {
	config := SecurityConfig{
		Enabled: true,
		AuditConfig: audit.AuditConfig{
			FilePath: "stdout",
		},
	}
	
	sm, err := NewSecurityManager(config)
	if err != nil {
		t.Fatalf("Failed to create SecurityManager: %v", err)
	}
	defer sm.Close()
	
	ctx := context.Background()
	
	// Create user with multiple roles
	err = sm.CreateUser(ctx, "testuser", "password123", []string{"readonly", "ddladmin"})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	user, _ := sm.GetUser(ctx, "testuser")
	
	// Should have SELECT from readonly
	err = sm.CheckPrivilege(ctx, user, "testdb", "users", authz.PrivilegeSelect)
	if err != nil {
		t.Errorf("Should have SELECT from readonly role: %v", err)
	}
	
	// Should have CREATE from ddladmin
	err = sm.CheckPrivilege(ctx, user, "testdb", "users", authz.PrivilegeCreate)
	if err != nil {
		t.Errorf("Should have CREATE from ddladmin role: %v", err)
	}
	
	// Should not have DELETE (not in either role)
	err = sm.CheckPrivilege(ctx, user, "testdb", "users", authz.PrivilegeDelete)
	if err != authz.ErrAccessDenied {
		t.Error("Should not have DELETE privilege")
	}
}

func TestSecurityManagerCustomRole(t *testing.T) {
	config := SecurityConfig{
		Enabled: true,
		AuditConfig: audit.AuditConfig{
			FilePath: "stdout",
		},
	}
	
	sm, err := NewSecurityManager(config)
	if err != nil {
		t.Fatalf("Failed to create SecurityManager: %v", err)
	}
	defer sm.Close()
	
	ctx := context.Background()
	
	// Create custom role
	err = sm.CreateRole(ctx, "customrole", authz.PrivilegeSelect|authz.PrivilegeUpdate)
	if err != nil {
		t.Fatalf("Failed to create custom role: %v", err)
	}
	
	// Create user with custom role
	err = sm.CreateUser(ctx, "testuser", "password123", []string{"customrole"})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	user, _ := sm.GetUser(ctx, "testuser")
	
	// Should have SELECT
	err = sm.CheckPrivilege(ctx, user, "testdb", "users", authz.PrivilegeSelect)
	if err != nil {
		t.Errorf("Should have SELECT: %v", err)
	}
	
	// Should have UPDATE
	err = sm.CheckPrivilege(ctx, user, "testdb", "users", authz.PrivilegeUpdate)
	if err != nil {
		t.Errorf("Should have UPDATE: %v", err)
	}
	
	// Should not have INSERT
	err = sm.CheckPrivilege(ctx, user, "testdb", "users", authz.PrivilegeInsert)
	if err != authz.ErrAccessDenied {
		t.Error("Should not have INSERT")
	}
}

func TestSecurityManagerDropUser(t *testing.T) {
	config := SecurityConfig{
		Enabled: true,
		AuditConfig: audit.AuditConfig{
			FilePath: "stdout",
		},
	}
	
	sm, err := NewSecurityManager(config)
	if err != nil {
		t.Fatalf("Failed to create SecurityManager: %v", err)
	}
	defer sm.Close()
	
	ctx := context.Background()
	
	// Create and then drop user
	err = sm.CreateUser(ctx, "tempuser", "password123", []string{"readonly"})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	user, _ := sm.GetUser(ctx, "tempuser")
	if user == nil {
		t.Error("User should exist")
	}
	
	err = sm.DropUser(ctx, "tempuser")
	if err != nil {
		t.Fatalf("Failed to drop user: %v", err)
	}
	
	user, _ = sm.GetUser(ctx, "tempuser")
	if user != nil {
		t.Error("User should no longer exist")
	}
}

func TestSecurityManagerListUsers(t *testing.T) {
	config := SecurityConfig{
		Enabled: true,
		AuditConfig: audit.AuditConfig{
			FilePath: "stdout",
		},
	}
	
	sm, err := NewSecurityManager(config)
	if err != nil {
		t.Fatalf("Failed to create SecurityManager: %v", err)
	}
	defer sm.Close()
	
	ctx := context.Background()
	
	// Should have at least root user
	users, err := sm.ListUsers(ctx)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}
	
	if len(users) < 1 {
		t.Error("Should have at least one user (root)")
	}
	
	// Create additional users
	sm.CreateUser(ctx, "user1", "pass1", []string{"readonly"})
	sm.CreateUser(ctx, "user2", "pass2", []string{"readwrite"})
	
	users, _ = sm.ListUsers(ctx)
	if len(users) < 3 {
		t.Errorf("Should have at least 3 users, got %d", len(users))
	}
}

func TestSecurityManagerAuditQuery(t *testing.T) {
	tmpDir := t.TempDir()
	auditFile := filepath.Join(tmpDir, "audit.log")
	
	config := SecurityConfig{
		Enabled: true,
		AuditConfig: audit.AuditConfig{
			FilePath: auditFile,
			Async:    false,
		},
	}
	
	sm, err := NewSecurityManager(config)
	if err != nil {
		t.Fatalf("Failed to create SecurityManager: %v", err)
	}
	defer sm.Close()
	
	ctx := context.Background()
	
	// Create and authenticate user
	sm.CreateUser(ctx, "testuser", "password123", []string{"readonly"})
	user, _ := sm.Authenticate(ctx, "testuser", "password123", "127.0.0.1")
	
	// Audit a query
	sm.AuditQuery(user, "127.0.0.1", "testdb", "SELECT * FROM users", 50*time.Millisecond, 10)
	
	// Verify audit log
	data, err := os.ReadFile(auditFile)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}
	
	content := string(data)
	if !containsString(content, "QUERY") {
		t.Error("Audit log should contain QUERY event")
	}
	if !containsString(content, "SELECT * FROM users") {
		t.Error("Audit log should contain the query")
	}
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
