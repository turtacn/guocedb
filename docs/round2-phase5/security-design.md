# GuoceDB Security Layer Design

## Overview

This document describes the security architecture implemented in Phase 5, providing a comprehensive security framework for GuoceDB including authentication, authorization, and audit logging.

## Architecture

The security layer consists of three main components unified through a SecurityManager facade:

```
┌─────────────────────────────────────────────────────────────┐
│                     SecurityManager                          │
│  (Unified facade for all security operations)               │
└────────────┬──────────────┬──────────────┬──────────────────┘
             │              │              │
     ┌───────▼──────┐ ┌────▼──────┐ ┌────▼──────────┐
     │Authenticator │ │Authorizer │ │ AuditLogger   │
     │  (authn)     │ │  (authz)  │ │  (audit)      │
     └──────┬───────┘ └─────┬─────┘ └───────┬───────┘
            │               │                │
     ┌──────▼───────┐ ┌─────▼──────┐ ┌──────▼───────┐
     │  UserStore   │ │ RoleStore  │ │ Event Logger │
     │ (in-memory)  │ │(in-memory) │ │  (file/json) │
     └──────────────┘ └────────────┘ └──────────────┘
```

## Components

### 1. Authentication (`security/auth`)

#### User Model (`user.go`)
Represents an authenticated user with:
- Username and password hash
- Assigned roles
- Direct privileges
- Account status (locked, expired)

```go
type User struct {
    ID           uint64
    Username     string
    PasswordHash string          // bcrypt hash
    Roles        []string        // Role names
    Privileges   authz.Privilege // Direct privileges
    CreatedAt    time.Time
    UpdatedAt    time.Time
    Locked       bool            // Account lock status
    ExpireAt     *time.Time      // Password expiration
}
```

#### Password Security (`password.go`)
- **bcrypt hashing**: Cost factor 12 for secure password storage
- **MySQL native password**: SHA1(SHA1(password)) for compatibility
- Password verification with timing-attack resistance

#### Authenticator (`authenticator.go`)
Handles user authentication with:
- Password verification
- Failed attempt tracking (configurable max attempts)
- Temporary account locking (configurable duration)
- Password expiration checking
- Account lock status verification

**Default Configuration:**
- Max failed attempts: 5
- Lock duration: 15 minutes

### 2. Authorization (`security/authz`)

#### Privilege Model (`privilege.go`)
Bitflag-based privilege system supporting:

**Basic Privileges:**
- `SELECT`, `INSERT`, `UPDATE`, `DELETE` (DML)
- `CREATE`, `DROP`, `ALTER`, `INDEX` (DDL)
- `GRANT`, `ADMIN` (Administrative)

**Composite Privileges:**
- `PrivilegeReadOnly`: SELECT only
- `PrivilegeReadWrite`: All DML operations
- `PrivilegeDDL`: All DDL operations
- `PrivilegeAll`: Complete access

#### Role Model (`role.go`)
Hierarchical privilege system:

```go
type Role struct {
    Name       string
    Privileges Privilege              // Global privileges
    
    // Fine-grained control
    DatabasePrivileges map[string]Privilege
    TablePrivileges    map[string]map[string]Privilege
}
```

**Predefined Roles:**
- `admin`: All privileges
- `readwrite`: DML operations
- `readonly`: SELECT only
- `ddladmin`: DDL operations

#### Authorizer (`authorizer.go`)
Privilege checking logic:
1. Check if user has ADMIN privilege (bypass all checks)
2. Check user's direct privileges
3. Check privileges from assigned roles
4. Check database-level privileges
5. Check table-level privileges

### 3. Audit Logging (`security/audit`)

#### Event Model (`event.go`)
Comprehensive audit trail with event types:
- `AUTHENTICATION`: Login attempts
- `AUTHORIZATION`: Privilege checks
- `QUERY`: Query execution
- `DDL`: Schema modifications
- `DML`: Data modifications
- `ADMIN`: Administrative operations
- `CONNECTION`: Connection events

Each event captures:
```go
type AuditEvent struct {
    Timestamp    time.Time
    EventType    EventType
    Result       EventResult  // SUCCESS/FAILURE/DENIED
    Username     string
    ClientIP     string
    Database     string
    Statement    string       // Truncated SQL
    Object       string       // Table/index name
    Privilege    string       // Required privilege
    ErrorMsg     string
    Duration     time.Duration
    RowsAffected int64
}
```

#### Audit Logger (`logger.go`)
Features:
- **Async logging**: Optional buffered channel for performance
- **JSON format**: Structured logs for parsing
- **IP filtering**: Exclude specific clients from audit
- **Automatic flushing**: Periodic and on-close
- **File rotation support**: Append-only mode

### 4. Security Manager (`security/security.go`)

Unified facade providing:

**Authentication Operations:**
- `Authenticate(username, password, clientIP)`: Verify credentials
- `CreateUser()`, `DropUser()`: User management
- `ListUsers()`: User enumeration

**Authorization Operations:**
- `CheckPrivilege()`: Single privilege check
- `CheckPrivileges()`: Batch privilege checks
- `GrantRole()`, `RevokeRole()`: Role assignment
- `CreateRole()`, `DropRole()`: Role management

**Audit Operations:**
- `AuditQuery()`: Log query execution
- `AuditConnection()`: Log connection attempts
- Automatic authentication/authorization logging

**Configuration:**
```go
type SecurityConfig struct {
    Enabled      bool              // Master switch
    AuditConfig  audit.AuditConfig
    MaxAuthFails int
    LockDuration time.Duration
}
```

## Security Features

### 1. Password Security
- **bcrypt hashing** with cost factor 12
- **Salt per password** (automatic with bcrypt)
- **Timing-attack resistance** in verification
- **MySQL native password** support for compatibility

### 2. Account Protection
- **Failed login tracking** per username
- **Automatic locking** after configurable attempts
- **Time-based unlocking** after configurable duration
- **Manual account locking** capability
- **Password expiration** support

### 3. Privilege Hierarchy
1. **ADMIN privilege**: Bypasses all checks
2. **Direct user privileges**: Highest priority
3. **Global role privileges**: Applied everywhere
4. **Database-level privileges**: Scoped to database
5. **Table-level privileges**: Most granular

### 4. Audit Trail
- **All authentication attempts** logged (success/failure)
- **Privilege denials** logged for security monitoring
- **Query execution** with duration and row counts
- **JSON format** for easy parsing and analysis
- **IP tracking** for all operations

## Usage Examples

### Basic Setup

```go
// Create security manager
config := security.SecurityConfig{
    Enabled: true,
    AuditConfig: audit.AuditConfig{
        FilePath:   "/var/log/guocedb/audit.log",
        Async:      true,
        BufferSize: 1000,
    },
    MaxAuthFails: 5,
    LockDuration: 15 * time.Minute,
}

sm, err := security.NewSecurityManager(config)
if err != nil {
    log.Fatal(err)
}
defer sm.Close()
```

### User Management

```go
// Create a new user
err := sm.CreateUser(ctx, "alice", "secret123", []string{"readwrite"})

// Grant additional role
err = sm.GrantRole(ctx, "alice", "ddladmin")

// Authenticate
user, err := sm.Authenticate(ctx, "alice", "secret123", "192.168.1.100")
if err != nil {
    // Handle authentication failure
}
```

### Authorization

```go
// Single privilege check
err := sm.CheckPrivilege(ctx, user, "mydb", "users", authz.PrivilegeSelect)
if err == authz.ErrAccessDenied {
    // Handle authorization failure
}

// Batch privilege checks
checks := []authz.PrivilegeCheck{
    {Database: "mydb", Table: "users", Privilege: authz.PrivilegeSelect},
    {Database: "mydb", Table: "orders", Privilege: authz.PrivilegeInsert},
}
err = sm.CheckPrivileges(ctx, user, checks)
```

### Custom Roles

```go
// Create a custom role with specific privileges
privileges := authz.PrivilegeSelect | authz.PrivilegeInsert | authz.PrivilegeUpdate
err := sm.CreateRole(ctx, "data_editor", privileges)

// Assign to user
err = sm.GrantRole(ctx, "bob", "data_editor")
```

## Testing

Comprehensive test coverage includes:

### Unit Tests
- **Password hashing**: bcrypt and MySQL native
- **Authenticator**: Success, failure, locking, expiry
- **Authorizer**: Direct, role-based, hierarchical checks
- **Audit logger**: Sync/async, filtering, JSON format

### Integration Tests
- **Full workflow**: User creation → authentication → authorization
- **Role management**: Grant, revoke, multiple roles
- **Custom roles**: Create, assign, verify
- **Audit trail**: Event logging and retrieval
- **Security disabled**: Bypass mode verification

### Test Coverage
```bash
# Run all security tests
go test ./security/... -v

# Run with coverage
go test ./security/... -cover
```

## Performance Considerations

### 1. Authentication
- bcrypt cost factor 12: ~100-200ms per hash
- Failed attempt tracking: O(1) lookups
- Recommendation: Cache authenticated sessions

### 2. Authorization
- Bitflag operations: O(1) privilege checks
- Role enumeration: O(n) where n = number of roles
- Recommendation: Keep role assignments minimal

### 3. Audit Logging
- **Async mode**: Non-blocking, buffered writes
- **Sync mode**: Guaranteed persistence, higher latency
- **Statement truncation**: 1000 chars max for performance
- Recommendation: Use async for high-throughput systems

## Security Best Practices

### 1. Deployment
- Enable security in production environments
- Use strong passwords (enforce externally)
- Configure appropriate lock durations
- Monitor audit logs for suspicious activity

### 2. Privilege Management
- Follow principle of least privilege
- Use predefined roles when possible
- Avoid granting ADMIN privilege unnecessarily
- Regularly audit user privileges

### 3. Audit Configuration
- Store audit logs on separate storage
- Implement log rotation
- Monitor for authentication failures
- Alert on privilege escalation attempts

### 4. Password Policy
- Enforce minimum password length (external)
- Implement password expiration where needed
- Use strong bcrypt cost factor (12+)
- Consider multi-factor authentication integration

## Future Enhancements

### Short Term
- Integration with MySQL handler connection lifecycle
- Query-level privilege extraction from SQL AST
- Real-time privilege checking in query execution

### Medium Term
- External authentication (LDAP, OAuth)
- Certificate-based authentication
- Session management with timeouts
- Privilege inheritance and delegation

### Long Term
- Row-level security policies
- Column-level access control
- Dynamic privilege evaluation
- Distributed audit log aggregation
- Integration with external SIEM systems

## Migration Notes

### From Old Security Layer
The new security layer replaces:
- `security/authn/authn.go` → `security/auth/*`
- `security/authz/authz.go` → `security/authz/*`
- `security/audit/audit.go` → `security/audit/*`

Key differences:
- Unified SecurityManager facade
- Bitflag-based privileges (vs. string permissions)
- Hierarchical role system
- Comprehensive audit events
- Failed attempt tracking

### Backward Compatibility
- Old user stores can be migrated via `GetUser()` + `CreateUser()`
- Privilege strings can be mapped to bitflags via `ParsePrivilege()`
- Existing roles can be recreated with new structure

## Conclusion

The Phase 5 security layer provides a production-ready foundation for GuoceDB security, featuring:
- ✅ Strong password hashing with bcrypt
- ✅ Role-based access control (RBAC)
- ✅ Comprehensive audit logging
- ✅ Account protection mechanisms
- ✅ Hierarchical privilege system
- ✅ Extensible architecture for future enhancements

All components are well-tested, documented, and ready for integration with the MySQL protocol handler.
