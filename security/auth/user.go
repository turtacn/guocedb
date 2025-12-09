// Package auth provides authentication services for GuoceDB.
package auth

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/turtacn/guocedb/security/authz"
)

// User represents an authenticated user in the system.
type User struct {
	ID           uint64
	Username     string
	PasswordHash string
	Roles        []string
	Privileges   authz.Privilege
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Locked       bool
	ExpireAt     *time.Time
}

// GetUsername returns the username (implements authz.User interface).
func (u *User) GetUsername() string {
	return u.Username
}

// GetRoles returns the user's roles (implements authz.User interface).
func (u *User) GetRoles() []string {
	return u.Roles
}

// GetPrivileges returns the user's privileges (implements authz.User interface).
func (u *User) GetPrivileges() authz.Privilege {
	return u.Privileges
}

// UserStore is the interface for user persistence.
type UserStore interface {
	GetUser(ctx context.Context, username string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, username string) error
	ListUsers(ctx context.Context) ([]*User, error)
}

// InMemoryUserStore is an in-memory implementation of UserStore.
type InMemoryUserStore struct {
	mu    sync.RWMutex
	users map[string]*User
	nextID uint64
}

// NewInMemoryUserStore creates a new in-memory user store with a default root user.
func NewInMemoryUserStore() *InMemoryUserStore {
	store := &InMemoryUserStore{
		users:  make(map[string]*User),
		nextID: 1,
	}
	
	// Create default root user with empty password
	rootHash, _ := HashPassword("")
	_ = store.CreateUser(context.Background(), &User{
		Username:     "root",
		PasswordHash: rootHash,
		Privileges:   authz.PrivilegeAll,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})
	
	return store
}

// GetUser retrieves a user by username (case-insensitive).
func (s *InMemoryUserStore) GetUser(ctx context.Context, username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	user, ok := s.users[strings.ToLower(username)]
	if !ok {
		return nil, nil
	}
	
	// Return a copy to prevent external modification
	userCopy := *user
	userCopy.Roles = append([]string(nil), user.Roles...)
	return &userCopy, nil
}

// CreateUser adds a new user to the store.
func (s *InMemoryUserStore) CreateUser(ctx context.Context, user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	username := strings.ToLower(user.Username)
	if _, exists := s.users[username]; exists {
		return ErrUserAlreadyExists
	}
	
	user.ID = s.nextID
	s.nextID++
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	
	s.users[username] = user
	return nil
}

// UpdateUser updates an existing user.
func (s *InMemoryUserStore) UpdateUser(ctx context.Context, user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	username := strings.ToLower(user.Username)
	if _, exists := s.users[username]; !exists {
		return nil // User not found error handled elsewhere
	}
	
	user.UpdatedAt = time.Now()
	s.users[username] = user
	return nil
}

// DeleteUser removes a user from the store.
func (s *InMemoryUserStore) DeleteUser(ctx context.Context, username string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.users, strings.ToLower(username))
	return nil
}

// ListUsers returns all users in the store.
func (s *InMemoryUserStore) ListUsers(ctx context.Context) ([]*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		userCopy := *user
		userCopy.Roles = append([]string(nil), user.Roles...)
		users = append(users, &userCopy)
	}
	
	return users, nil
}

var (
	ErrUserAlreadyExists = &AuthError{Code: "USER_EXISTS", Message: "user already exists"}
)

// AuthError represents an authentication error.
type AuthError struct {
	Code    string
	Message string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
