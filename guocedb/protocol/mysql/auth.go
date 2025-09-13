package mysql

import (
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql/mysql_db"
	"github.com/turtacn/guocedb/security/authn"
)

// GuocedbLogin is a custom login method for the MySQL server that uses our authenticator.
func GuocedbLogin(ctx *server.Context, user, password string, salt []byte) (mysql_db.AuthedUser, error) {
	// In a real server, the authenticator would be retrieved from a dependency injection container
	// or a global registry. For now, we create a new one.
	authenticator := authn.NewDefaultAuthenticator()

	// The context passed to Authenticate should be the real request context.
	principal, err := authenticator.Authenticate(ctx, user, password, salt)
	if err != nil {
		return mysql_db.AuthedUser{}, err
	}

	// The AuthedUser object is from GMS and is used to populate the session.
	// It's also used by the mysql_db database for privilege checks if you use it.
	return mysql_db.AuthedUser{
		User: principal.Name,
		// Host would be determined from the connection context.
		Host:          ctx.Client().Address,
		PrivilegeSet:  mysql_db.NewPrivilegeSet(), // Start with no privileges.
		Authenticated: true,
	}, nil
}

// NewAuthProvider creates a new server.auth.Server configured for guocedb.
func NewAuthProvider() server.AuthProvider {
	// We use the NativePassword method, but provide our own login function.
	return server.NewNativePasswordServer(GuocedbLogin)
}

// TODO: Enhance the AuthedUser with roles and privileges from our authz system.
// TODO: Support TLS options for secure connections.
