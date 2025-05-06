// Package authz contains the authorization subsystem.
// It is responsible for checking if an authenticated user has permission to perform an action on a resource.
//
// authz 包包含授权子系统。
// 它负责检查已认证的用户是否具有对资源执行操作的权限。
package authz

import (
	"context"
	"fmt"

	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/security/authn" // Import authentication identity
)

// Action represents a specific database operation or privilege.
// Examples: SELECT, INSERT, CREATE TABLE, DROP DATABASE, ALTER USER.
//
// Action 表示特定的数据库操作或权限。
// 示例：SELECT, INSERT, CREATE TABLE, DROP DATABASE, ALTER USER。
type Action string

// Define common database actions
// 定义常见的数据库操作
const (
	ActionSelect       Action = "SELECT"
	ActionInsert       Action = "INSERT"
	ActionUpdate       Action = "UPDATE"
	ActionDelete       Action = "DELETE"
	ActionCreateDatabase Action = "CREATE DATABASE"
	ActionDropDatabase Action = "DROP DATABASE"
	ActionCreateTable  Action = "CREATE TABLE"
	ActionDropTable    Action = "DROP TABLE"
	ActionAlterTable   Action = "ALTER TABLE"
	// TODO: Add more actions
	// TODO: 添加更多操作
)


// Resource represents the target of an action (e.g., a specific table, a database, a user).
// Resources can be hierarchical (e.g., a table within a database).
//
// Resource 表示操作的目标（例如，特定的表、数据库、用户）。
// 资源可以是分层的（例如，数据库中的表）。
type Resource interface {
	// GetType returns the type of the resource (e.g., "database", "table", "user").
	// GetType 返回资源的类型（例如，“database”、“table”、“user”）。
	GetType() string

	// GetName returns the name of the resource.
	// GetName 返回资源的名称。
	GetName() string

	// GetParent returns the parent resource, or nil if it's a top-level resource.
	// GetParent 返回父资源，如果是顶级资源则返回 nil。
	GetParent() Resource

	// String returns a string representation of the resource (e.g., "database:mydb", "table:mydb.mytable").
	// String 返回资源的字符串表示（例如，“database:mydb”、“table:mydb.mytable”）。
	String() string

	// TODO: Add method for checking if this resource matches another resource pattern.
	// TODO: 添加方法检查此资源是否匹配另一个资源模式。
}

// DatabaseResource represents a database as a resource.
// DatabaseResource 表示数据库作为资源。
type DatabaseResource struct {
	Name string
}
func (r *DatabaseResource) GetType() string { return "database" }
func (r *DatabaseResource) GetName() string { return r.Name }
func (r *DatabaseResource) GetParent() Resource { return nil } // Top-level
func (r *DatabaseResource) String() string { return fmt.Sprintf("database:%s", r.Name) }


// TableResource represents a table within a database as a resource.
// TableResource 表示数据库中的表作为资源。
type TableResource struct {
	Database string // Parent database name
	Name string
	Parent Resource // Reference to parent DatabaseResource
}
func (r *TableResource) GetType() string { return "table" }
func (r *TableResource) GetName() string { return r.Name }
func (r *TableResource) GetParent() Resource {
	// Lazily create parent if not set? Or require it in constructor?
	// For simplicity, return a new DatabaseResource based on the name.
	//
	// 如果未设置，延迟创建父级？还是在构造函数中要求？
	// 为了简化，根据名称返回一个新的 DatabaseResource。
	return &DatabaseResource{Name: r.Database}
}
func (r *TableResource) String() string { return fmt.Sprintf("table:%s.%s", r.Database, r.Name) }


// Authorizer is the interface for checking permissions.
// Authorizer 是用于检查权限的接口。
// The compute layer (analyzer, executor) will use this interface.
// 计算层（分析器、执行器）将使用此接口。
type Authorizer interface {
	// CheckPermission checks if the given identity has permission to perform the action on the resource.
	// It returns an error if permission is denied, otherwise nil.
	//
	// CheckPermission 检查给定的身份是否具有对资源执行操作的权限。
	// 如果权限被拒绝，则返回错误，否则返回 nil。
	CheckPermission(ctx context.Context, identity authn.Identity, action Action, resource Resource) error

	// TODO: Add methods for grant/revoke permissions, role management.
	// TODO: 添加用于授予/撤销权限、角色管理的方法。
	// Or permission management could be a separate service/package.
	// 或者权限管理可以是一个独立的服务/包。
}

// TODO: Implement concrete Authorizer implementations here,
// potentially reading permissions from configuration or system tables,
// and implementing logic for checking grants/roles.
//
// TODO: 在此处实现具体的 Authorizer 实现，
// 可以从配置或系统表读取权限，
// 并实现检查授予/角色的逻辑。

// PlaceholderAuthorizer is a dummy authorizer for basic testing.
// PlaceholderAuthorizer 是一个用于基本测试的虚拟授权器。
// It grants all permissions to the "root" user and denies all others.
// 它授予“root”用户所有权限，拒绝其他所有用户。
type PlaceholderAuthorizer struct {
	// For now, stateless. Could hold permission data.
	// 目前是无状态的。可以保存权限数据。
}

// NewPlaceholderAuthorizer creates a new PlaceholderAuthorizer.
// NewPlaceholderAuthorizer 创建一个新的 PlaceholderAuthorizer。
func NewPlaceholderAuthorizer() Authorizer {
	log.Info("Initializing placeholder authorizer.") // 初始化占位符授权器。
	// Replace with a real implementation that manages permissions securely.
	// 替换为安全管理权限的真实实现。
	return &PlaceholderAuthorizer{}
}

// CheckPermission performs dummy authorization.
// It grants all permissions to "root" and denies for any other user.
//
// CheckPermission 执行虚拟授权。
// 它授予“root”所有权限，拒绝其他任何用户。
func (a *PlaceholderAuthorizer) CheckPermission(ctx context.Context, identity authn.Identity, action Action, resource Resource) error {
	log.Debug("PlaceholderAuthorizer CheckPermission called for user '%s', action '%s', resource '%s'", identity.GetUsername(), action, resource.String()) // 调用 PlaceholderAuthorizer CheckPermission。

	if identity.GetUsername() == "root" {
		log.Debug("PlaceholderAuthorizer: Granting permission to root user.") // PlaceholderAuthorizer：授予 root 用户权限。
		return nil // Root user has all permissions
	}

	log.Warn("PlaceholderAuthorizer: Permission denied for user '%s', action '%s', resource '%s'.", identity.GetUsername(), action, resource.String()) // PlaceholderAuthorizer：用户 '%s' 权限被拒绝。
	// Return a standard GMS access denied error if possible, or our own error.
	// If using GMS analyzer/executor, returning GMS error might be better.
	//
	// 如果可能，返回标准 GMS 权限拒绝错误，或返回我们自己的错误。
	// 如果使用 GMS analyzer/executor，返回 GMS 错误可能更好。
	// GMS uses sql.ErrPermissionDenied.New.
	// GMS 使用 sql.ErrPermissionDenied.New。
	return errors.ErrPermissionDenied.New(fmt.Sprintf("user '%s' cannot perform '%s' on '%s'", identity.GetUsername(), action, resource.String())) // Use our error type
}