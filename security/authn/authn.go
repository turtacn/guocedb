// Package authn contains the authentication subsystem.
// It is responsible for verifying the identity of users connecting to the database.
//
// authn 包包含认证子系统。
// 它负责验证连接到数据库的用户的身份。
package authn

import (
	"context"
	"fmt"
	"github.com/turtacn/guocedb/common/log"

	"github.com/turtacn/guocedb/common/errors"
	// Consider importing relevant go-mysql-server types if integrating closely.
	// 考虑导入相关的 go-mysql-server 类型（如果紧密集成的话）。
	// e.g., "github.com/dolthub/go-mysql-server/server/mysql"
)

// Identity represents an authenticated user or service principal.
// It contains information about the authenticated entity.
//
// Identity 表示已认证的用户或服务 principal。
// 它包含有关已认证实体的信息。
type Identity interface {
	// GetUsername returns the username of the authenticated entity.
	// GetUsername 返回已认证实体的用户名。
	GetUsername() string

	// GetAttribute returns a specific attribute of the identity (e.g., source IP, roles).
	// GetAttribute 返回身份的特定属性（例如，源 IP、角色）。
	GetAttribute(name string) (interface{}, bool)

	// TODO: Add methods for UUID, roles, etc.
	// TODO: 添加用于 UUID、角色等的方法。
}

// SimpleIdentity is a basic implementation of the Identity interface.
// SimpleIdentity 是 Identity 接口的基本实现。
type SimpleIdentity struct {
	Username string
	Attributes map[string]interface{}
}

func (i *SimpleIdentity) GetUsername() string {
	return i.Username
}

func (i *SimpleIdentity) GetAttribute(name string) (interface{}, bool) {
	attr, ok := i.Attributes[name]
	return attr, ok
}


// Authenticator is the interface for verifying user credentials.
// Authenticator 是用于验证用户凭据的接口。
type Authenticator interface {
	// Authenticate attempts to authenticate the given credentials.
	// It returns an Identity if authentication is successful, otherwise an error.
	// Credentials can be a username/password pair, a token, etc.
	//
	// Authenticate 尝试认证给定的凭据。
	// 如果认证成功，则返回一个 Identity，否则返回错误。
	// 凭据可以是用户名/密码对、令牌等。
	Authenticate(ctx context.Context, credentials interface{}) (Identity, error)

	// TODO: Add methods for user management (create, drop, alter user).
	// TODO: 添加用于用户管理的方法（创建、删除、修改用户）。
	// Or user management could be a separate service/package.
	// 或者用户管理可以是一个独立的服务/包。
}

// TODO: Implement concrete Authenticator implementations here,
// potentially reading user credentials from configuration, a system table,
// or integrating with external authentication systems (LDAP, OAuth).
//
// TODO: 在此处实现具体的 Authenticator 实现，
// 可以从配置、系统表读取用户凭据，
// 或与外部认证系统（LDAP, OAuth）集成。

// PlaceholderAuthenticator is a dummy authenticator for basic testing.
// PlaceholderAuthenticator 是一个用于基本测试的虚拟认证器。
type PlaceholderAuthenticator struct {
	// Hardcoded users for demonstration
	// 用于演示的硬编码用户
	validUsers map[string]string // username -> password
}

// NewPlaceholderAuthenticator creates a new PlaceholderAuthenticator.
// NewPlaceholderAuthenticator 创建一个新的 PlaceholderAuthenticator。
func NewPlaceholderAuthenticator() Authenticator {
	// This should be replaced with a real implementation that loads users securely.
	// 这应该被替换为安全加载用户的真实实现。
	return &PlaceholderAuthenticator{
		validUsers: map[string]string{
			"root": "password", // Example
			"test": "test",
		},
	}
}

// Authenticate performs dummy authentication.
// It checks if the credentials match hardcoded users.
// Credentials are expected to be a map[string]string {"username": "...", "password": "..."}
//
// Authenticate 执行虚拟认证。
// 它检查凭据是否与硬编码用户匹配。
// 凭据应为 map[string]string {"username": "...", "password": "..."}
func (a *PlaceholderAuthenticator) Authenticate(ctx context.Context, credentials interface{}) (Identity, error) {
	log.Debug("PlaceholderAuthenticator Authenticate called.") // 调用 PlaceholderAuthenticator Authenticate。

	credsMap, ok := credentials.(map[string]string)
	if !ok {
		log.Warn("PlaceholderAuthenticator: Invalid credentials format.") // PlaceholderAuthenticator：凭据格式无效。
		return nil, errors.ErrAuthenticationFailed.New("invalid credentials format") // 认证失败：凭据格式无效。
	}

	username, userOk := credsMap["username"]
	password, passOk := credsMap["password"]
	if !userOk || !passOk {
		log.Warn("PlaceholderAuthenticator: Missing username or password in credentials.") // PlaceholderAuthenticator：凭据中缺少用户名或密码。
		return nil, errors.ErrAuthenticationFailed.New("missing username or password") // 认证失败：缺少用户名或密码。
	}

	expectedPassword, ok := a.validUsers[username]
	if !ok {
		log.Warn("PlaceholderAuthenticator: User '%s' not found.", username) // PlaceholderAuthenticator：用户 '%s' 未找到。
		return nil, errors.ErrAuthenticationFailed.New("user not found") // 认证失败：用户未找到。
	}

	// Insecure plaintext comparison for placeholder
	// 占位符的不安全明文比较
	if password != expectedPassword {
		log.Warn("PlaceholderAuthenticator: Invalid password for user '%s'.", username) // PlaceholderAuthenticator：用户 '%s' 密码无效。
		return nil, errors.ErrAuthenticationFailed.New("invalid password") // 认证失败：密码无效。
	}

	log.Info("PlaceholderAuthenticator: Authentication successful for user '%s'.", username) // PlaceholderAuthenticator：用户 '%s' 认证成功。
	// Return a SimpleIdentity for the authenticated user
	// 返回已认证用户的 SimpleIdentity
	return &SimpleIdentity{Username: username, Attributes: make(map[string]interface{})}, nil
}

// Note: The mysql.Authenticator in protocol/mysql wraps this internal Authenticator
// to integrate with the go-mysql-server server logic.
//
// 注意：protocol/mysql 中的 mysql.Authenticator 包装此内部 Authenticator，
// 以与 go-mysql-server 服务器逻辑集成。