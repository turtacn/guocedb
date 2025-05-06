// Package mysql provides the MySQL server protocol implementation.
// auth.go handles client authentication logic.
//
// mysql 包提供了 MySQL 服务器协议实现。
// auth.go 处理客户端认证逻辑。
package mysql

import (
	"context"
	"fmt"
	"strings" // Needed for string comparisons

	"github.com/dolthub/go-mysql-server/server/mysql" // Import GMS MySQL server types
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
)

// Authenticator is an implementation of the mysql.Authenticator interface for Guocedb.
// It handles the authentication process for incoming MySQL connections.
//
// Authenticator 是 Guocedb 的 mysql.Authenticator 接口的实现。
// 它处理传入 MySQL 连接的认证过程。
type Authenticator struct {
	// TODO: Add a reference to a user/permission manager if complex authentication is needed.
	// TODO: 如果需要复杂的认证，添加对用户/权限管理器的引用。
	// For now, use hardcoded credentials.
	// 目前使用硬编码凭据。
	validUsers map[string]string // Map of username to password
}

// NewAuthenticator creates a new Authenticator instance.
// It is initialized with a simple hardcoded user for demonstration.
//
// NewAuthenticator 创建一个新的 Authenticator 实例。
// 它使用一个简单的硬编码用户进行演示初始化。
func NewAuthenticator() mysql.Authenticator {
	log.Info("Initializing MySQL authenticator.") // 初始化 MySQL 认证器。
	// In a real system, this would load users from config or a system table.
	// 在实际系统中，这将从配置或系统表中加载用户。
	validUsers := map[string]string{
		"root": "password", // Example hardcoded user/password
		"test": "test",
	}
	log.Warn("Using hardcoded user credentials ('%s', 'password') for authentication. Replace in production!", "root") // 使用硬编码用户凭据进行认证。在生产环境中替换！

	return &Authenticator{
		validUsers: validUsers,
	}
}

// Authenticate performs the authentication check against the provided principal.
// It compares the principal's username and password against the valid users.
//
// Authenticate 对提供的 principal 执行认证检查。
// 它将 principal 的用户名和密码与有效用户进行比较。
func (a *Authenticator) Authenticate(ctx context.Context, principal mysql.Principal) error {
	log.Debug("Authenticating user '%s' from %s (Conn ID: %d)", principal.Username, principal.RemoteAddress, principal.ConnectionID) // 认证用户。

	providedUsername := principal.Username
	providedPassword := principal.Password // This is the plaintext password received during authentication

	// Check against hardcoded valid users
	// 对硬编码的有效用户进行检查
	expectedPassword, ok := a.validUsers[providedUsername]
	if !ok {
		log.Warn("Authentication failed for user '%s': user not found.", providedUsername) // 认证失败：用户未找到。
		return mysql.ErrAccessDenied.New(principal.Username, principal.RemoteAddress) // Return GMS MySQL access denied error
	}

	// Basic password comparison (in a real system, hash comparison is required)
	// 基本密码比较（在实际系统中，需要进行哈希比较）
	if providedPassword != expectedPassword {
		log.Warn("Authentication failed for user '%s': invalid password.", providedUsername) // 认证失败：密码无效。
		return mysql.ErrAccessDenied.New(principal.Username, principal.RemoteAddress) // Return GMS MySQL access denied error
	}

	// Authentication successful
	// 认证成功
	log.Info("Authentication successful for user '%s' from %s", principal.Username, principal.RemoteAddress) // 用户 '%s' 从 %s 认证成功。
	return nil
}

// AllowedCleartextPasswords returns whether cleartext passwords are allowed.
// In a real system, this should be false unless using SSL/TLS.
// AllowedCleartextPasswords 返回是否允许明文密码。
// 在实际系统中，除非使用 SSL/TLS，否则应为 false。
func (a *Authenticator) AllowedCleartextPasswords() bool {
	// WARNING: Allowing cleartext passwords is insecure without TLS.
	// This is enabled for basic testing.
	//
	// 警告：在没有 TLS 的情况下允许明文密码是不安全的。
	// 此功能用于基本测试。
	log.Warn("Authentication allows cleartext passwords (insecure).") // 认证允许明文密码（不安全）。
	return true // Allow cleartext for simple testing
}