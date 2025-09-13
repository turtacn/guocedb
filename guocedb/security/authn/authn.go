package authn

import (
	"context"
	"errors"
	"fmt"

	"github.com/turtacn/guocedb/security/crypto"
	"golang.orgx/crypto/bcrypt"
)

// User represents an authenticated user principal.
type User struct {
	Name       string
	Attributes map[string]interface{}
}

// Authenticator is the interface for authenticating users.
type Authenticator interface {
	// Authenticate checks the user's credentials and returns a User principal on success.
	Authenticate(ctx context.Context, userName string, password string, salt []byte) (*User, error)
}

// DefaultAuthenticator is a simple username/password authenticator.
// In a real system, this would check credentials against a user store.
type DefaultAuthenticator struct {
	// This would typically be a user store, e.g., backed by the storage engine.
	// For this placeholder, we'll use a simple map.
	// The key is the username, and the value is the bcrypt-hashed password.
	userStore map[string][]byte
}

// NewDefaultAuthenticator creates a new DefaultAuthenticator.
func NewDefaultAuthenticator() *DefaultAuthenticator {
	// Create a dummy user for demonstration purposes.
	// The password is "password".
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	return &DefaultAuthenticator{
		userStore: map[string][]byte{
			"root": hashedPassword,
		},
	}
}

// Authenticate performs a simple password check.
func (a *DefaultAuthenticator) Authenticate(ctx context.Context, userName string, password string, salt []byte) (*User, error) {
	hashedPassword, ok := a.userStore[userName]
	if !ok {
		return nil, errors.New("user not found")
	}

	// Compare the provided password with the stored hash.
	// Note: MySQL's native password auth is more complex and uses the salt.
	// This is a simplified example.
	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	return &User{Name: userName}, nil
}


// TODO: Implement MySQL native authentication methods (e.g., mysql_native_password, caching_sha2_password).
// TODO: Integrate with a persistent user store instead of the in-memory map.
// TODO: Support other authentication mechanisms (e.g., TLS client certificates, external identity providers).
