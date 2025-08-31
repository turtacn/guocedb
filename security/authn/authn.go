// Package authn provides authentication services for guocedb.
package authn

import (
	"sync"
	"time"

	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/security/crypto"
)

// User represents an authenticated user.
type User struct {
	Name       string
	HashedPass string
	// Roles and other metadata would go here.
}

// Session represents a user's session.
type Session struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Authenticator is the interface for authentication.
type Authenticator interface {
	// Authenticate a user with given credentials.
	Authenticate(username, password string) (*User, error)
	// CreateSession creates a new session for a user.
	CreateSession(user *User) (*Session, error)
	// ValidateSession checks if a session is valid.
	ValidateSession(sessionID string) (*Session, error)
}

// MemoryAuthenticator is a simple in-memory authenticator.
// A real implementation would use the storage engine to persist users.
type MemoryAuthenticator struct {
	mu      sync.RWMutex
	users    map[string]*User
	sessions map[string]*Session
}

// NewMemoryAuthenticator creates a new in-memory authenticator.
func NewMemoryAuthenticator() *MemoryAuthenticator {
	// Add a default user for testing
	hashedPass, _ := crypto.HashPassword("password")
	users := map[string]*User{
		"root": {Name: "root", HashedPass: hashedPass},
	}

	return &MemoryAuthenticator{
		users:    users,
		sessions: make(map[string]*Session),
	}
}

// Authenticate checks user credentials.
func (a *MemoryAuthenticator) Authenticate(username, password string) (*User, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	user, ok := a.users[username]
	if !ok {
		return nil, errors.New(1, "user not found")
	}

	if !crypto.CheckPasswordHash(password, user.HashedPass) {
		return nil, errors.New(1, "invalid password")
	}

	return user, nil
}

// CreateSession creates a new session.
func (a *MemoryAuthenticator) CreateSession(user *User) (*Session, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	sessionIDBytes, err := crypto.GenerateRandomBytes(32)
	if err != nil {
		return nil, err
	}
	sessionID := crypto.SHA256(sessionIDBytes)

	session := &Session{
		ID:        sessionID,
		UserID:    user.Name,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	a.sessions[sessionID] = session

	return session, nil
}

// ValidateSession validates a session ID.
func (a *MemoryAuthenticator) ValidateSession(sessionID string) (*Session, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	session, ok := a.sessions[sessionID]
	if !ok {
		return nil, errors.New(1, "session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		// Clean up expired session
		go func() {
			a.mu.Lock()
			delete(a.sessions, sessionID)
			a.mu.Unlock()
		}()
		return nil, errors.New(1, "session expired")
	}

	return session, nil
}

// IntegrateWithExternal is a placeholder for external auth (e.g., LDAP, OAuth).
func (a *MemoryAuthenticator) IntegrateWithExternal() error {
	return errors.ErrNotImplemented
}

// Enforce interface compliance
var _ Authenticator = (*MemoryAuthenticator)(nil)
