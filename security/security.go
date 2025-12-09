// Package security provides unified security management for GuoceDB.
package security

import (
	"context"
	"time"

	"github.com/turtacn/guocedb/security/audit"
	"github.com/turtacn/guocedb/security/auth"
	"github.com/turtacn/guocedb/security/authz"
)

// SecurityManager is the unified facade for all security operations.
type SecurityManager struct {
	authenticator *auth.Authenticator
	authorizer    *authz.Authorizer
	auditLogger   *audit.AuditLogger
	userStore     auth.UserStore
	roleStore     authz.RoleStore
	enabled       bool
}

// SecurityConfig configures the security manager.
type SecurityConfig struct {
	Enabled      bool
	AuditConfig  audit.AuditConfig
	MaxAuthFails int
	LockDuration time.Duration
}

// NewSecurityManager creates a new security manager with the given configuration.
func NewSecurityManager(config SecurityConfig) (*SecurityManager, error) {
	if !config.Enabled {
		return &SecurityManager{enabled: false}, nil
	}
	
	// Initialize stores
	userStore := auth.NewInMemoryUserStore()
	roleStore := authz.NewInMemoryRoleStore()
	
	// Initialize audit logger
	auditLogger, err := audit.NewAuditLogger(config.AuditConfig)
	if err != nil {
		return nil, err
	}
	
	// Initialize authenticator
	authenticator := auth.NewAuthenticator(userStore)
	if config.MaxAuthFails > 0 {
		authenticator.SetMaxAttempts(config.MaxAuthFails)
	}
	if config.LockDuration > 0 {
		authenticator.SetLockDuration(config.LockDuration)
	}
	
	// Initialize authorizer
	authorizer := authz.NewAuthorizer(roleStore)
	
	return &SecurityManager{
		authenticator: authenticator,
		authorizer:    authorizer,
		auditLogger:   auditLogger,
		userStore:     userStore,
		roleStore:     roleStore,
		enabled:       true,
	}, nil
}

// Authenticate verifies user credentials and returns the authenticated user.
func (sm *SecurityManager) Authenticate(ctx context.Context, username, password, clientIP string) (*auth.User, error) {
	if !sm.enabled {
		// Security disabled, return a super user
		return &auth.User{
			Username:   username,
			Privileges: authz.PrivilegeAll,
		}, nil
	}
	
	user, err := sm.authenticator.Authenticate(ctx, username, password)
	
	// Audit the authentication attempt
	sm.auditLogger.Log(audit.NewAuthenticationEvent(username, clientIP, err == nil))
	
	return user, err
}

// CheckPrivilege verifies if a user has the required privilege on a resource.
func (sm *SecurityManager) CheckPrivilege(ctx context.Context, user *auth.User, database, table string, privilege authz.Privilege) error {
	if !sm.enabled {
		return nil
	}
	
	err := sm.authorizer.CheckPrivilege(ctx, user, database, table, privilege)
	
	// Audit only denials to reduce log volume
	if err != nil {
		event := audit.NewAuthorizationEvent(
			user.Username,
			"", // ClientIP not available here
			database,
			table,
			privilege.String(),
			true,
		)
		sm.auditLogger.Log(event)
	}
	
	return err
}

// CheckPrivileges performs multiple privilege checks in batch.
func (sm *SecurityManager) CheckPrivileges(ctx context.Context, user *auth.User, checks []authz.PrivilegeCheck) error {
	if !sm.enabled {
		return nil
	}
	
	return sm.authorizer.CheckPrivileges(ctx, user, checks)
}

// AuditQuery records a query execution in the audit log.
func (sm *SecurityManager) AuditQuery(user *auth.User, clientIP, database, statement string, duration time.Duration, rowsAffected int64) {
	if !sm.enabled {
		return
	}
	
	event := audit.NewQueryEvent(user.Username, clientIP, database, statement, duration, rowsAffected)
	sm.auditLogger.Log(event)
}

// AuditConnection records a connection attempt in the audit log.
func (sm *SecurityManager) AuditConnection(username, clientIP string, success bool) {
	if !sm.enabled {
		return
	}
	
	event := audit.NewConnectionEvent(username, clientIP, success)
	sm.auditLogger.Log(event)
}

// CreateUser creates a new user with the specified credentials and roles.
func (sm *SecurityManager) CreateUser(ctx context.Context, username, password string, roles []string) error {
	if !sm.enabled {
		return nil
	}
	
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	
	user := &auth.User{
		Username:     username,
		PasswordHash: hash,
		Roles:        roles,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	return sm.userStore.CreateUser(ctx, user)
}

// DropUser removes a user from the system.
func (sm *SecurityManager) DropUser(ctx context.Context, username string) error {
	if !sm.enabled {
		return nil
	}
	
	return sm.userStore.DeleteUser(ctx, username)
}

// GetUser retrieves a user by username.
func (sm *SecurityManager) GetUser(ctx context.Context, username string) (*auth.User, error) {
	if !sm.enabled {
		return nil, nil
	}
	
	return sm.userStore.GetUser(ctx, username)
}

// ListUsers returns all users in the system.
func (sm *SecurityManager) ListUsers(ctx context.Context) ([]*auth.User, error) {
	if !sm.enabled {
		return nil, nil
	}
	
	return sm.userStore.ListUsers(ctx)
}

// GrantRole assigns a role to a user.
func (sm *SecurityManager) GrantRole(ctx context.Context, username, roleName string) error {
	if !sm.enabled {
		return nil
	}
	
	user, err := sm.userStore.GetUser(ctx, username)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	
	// Verify role exists
	role, err := sm.roleStore.GetRole(ctx, roleName)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}
	
	// Check if user already has the role
	for _, r := range user.Roles {
		if r == roleName {
			return nil // Already has the role
		}
	}
	
	user.Roles = append(user.Roles, roleName)
	user.UpdatedAt = time.Now()
	
	return sm.userStore.UpdateUser(ctx, user)
}

// RevokeRole removes a role from a user.
func (sm *SecurityManager) RevokeRole(ctx context.Context, username, roleName string) error {
	if !sm.enabled {
		return nil
	}
	
	user, err := sm.userStore.GetUser(ctx, username)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	
	// Remove the role
	newRoles := make([]string, 0, len(user.Roles))
	for _, r := range user.Roles {
		if r != roleName {
			newRoles = append(newRoles, r)
		}
	}
	
	user.Roles = newRoles
	user.UpdatedAt = time.Now()
	
	return sm.userStore.UpdateUser(ctx, user)
}

// CreateRole creates a new role with the specified privileges.
func (sm *SecurityManager) CreateRole(ctx context.Context, roleName string, privileges authz.Privilege) error {
	if !sm.enabled {
		return nil
	}
	
	role := &authz.Role{
		Name:       roleName,
		Privileges: privileges,
	}
	
	return sm.roleStore.CreateRole(ctx, role)
}

// DropRole removes a role from the system.
func (sm *SecurityManager) DropRole(ctx context.Context, roleName string) error {
	if !sm.enabled {
		return nil
	}
	
	return sm.roleStore.DeleteRole(ctx, roleName)
}

// GetRole retrieves a role by name.
func (sm *SecurityManager) GetRole(ctx context.Context, roleName string) (*authz.Role, error) {
	if !sm.enabled {
		return nil, nil
	}
	
	return sm.roleStore.GetRole(ctx, roleName)
}

// ListRoles returns all roles in the system.
func (sm *SecurityManager) ListRoles(ctx context.Context) ([]*authz.Role, error) {
	if !sm.enabled {
		return nil, nil
	}
	
	return sm.roleStore.ListRoles(ctx)
}

// Close shuts down the security manager and flushes any pending audit logs.
func (sm *SecurityManager) Close() error {
	if sm.auditLogger != nil {
		return sm.auditLogger.Close()
	}
	return nil
}

// IsEnabled returns whether security is enabled.
func (sm *SecurityManager) IsEnabled() bool {
	return sm.enabled
}
