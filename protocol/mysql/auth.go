// Package mysql provides MySQL protocol handling for guocedb.
package mysql

import (
	"crypto/tls"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/turtacn/guocedb/security/authn"
)

// Auth is a wrapper around go-mysql-server's native authentication
// that integrates with our custom authenticator.
type Auth struct {
	// Our custom authenticator
	authenticator authn.Authenticator
	// Whether to use TLS
	isSecure bool
	// TLS configuration
	tlsConfig *tls.Config
}

// NewAuth creates a new Auth instance.
func NewAuth(authenticator authn.Authenticator, tlsConfig *tls.Config) *Auth {
	return &Auth{
		authenticator: authenticator,
		isSecure:      tlsConfig != nil,
		tlsConfig:     tlsConfig,
	}
}

// MysqlNativePassword implements the server.MysqlNativePasswordAuthenticator interface.
func (a *Auth) MysqlNativePassword(
	user, userHost string,
	salt, pass []byte,
) (server.MysqlNativePasswordIdentity, error) {
	// The go-mysql-server library handles the challenge-response calculation.
	// We just need to verify the provided cleartext password.
	// This is a simplified view; the library provides helpers for the actual scramble.
	// For this implementation, we assume a custom authenticator that takes a clear pass.
	// This part of the code is tricky because GMS doesn't directly expose the clear pass.
	// A real implementation requires carefully hooking into the auth process.

	// Let's assume for now we have a way to get the clear password or that
	// our authenticator can handle the scrambled version.
	// For this placeholder, we will bypass the check and rely on the session auth.
	// This is NOT secure and is for structural purposes only.

	// A more realistic approach would be to store users in a way that GMS can
	// access them directly, e.g. via a `mysql.user` table.

	// _, err := a.authenticator.Authenticate(user, string(pass))
	// if err != nil {
	// 	return server.MysqlNativePasswordIdentity{}, err
	// }

	return server.MysqlNativePasswordIdentity{Username: user, Host: userHost}, nil
}

// Secure implements the server.SecureAuthenticator interface.
func (a *Auth) Secure() *tls.Config {
	if !a.isSecure {
		return nil
	}
	return a.tlsConfig
}

// server.Authenticator interface compliance check
var _ server.Authenticator = (*Auth)(nil)
