// Package auth provides authentication services for GuoceDB.
package auth

import (
	"context"
	"sync"
	"time"
)

// attemptInfo tracks authentication failure attempts.
type attemptInfo struct {
	count    int
	lastFail time.Time
}

// Authenticator handles user authentication.
type Authenticator struct {
	userStore    UserStore
	maxAttempts  int
	lockDuration time.Duration
	attempts     sync.Map // username -> *attemptInfo
}

// NewAuthenticator creates a new authenticator with the given user store.
func NewAuthenticator(store UserStore) *Authenticator {
	return &Authenticator{
		userStore:    store,
		maxAttempts:  5,
		lockDuration: 15 * time.Minute,
	}
}

// Authenticate verifies user credentials and returns the user on success.
func (a *Authenticator) Authenticate(ctx context.Context, username, password string) (*User, error) {
	// Check if user is temporarily locked due to failed attempts
	if a.isLocked(username) {
		return nil, ErrAccountLocked
	}
	
	// Get user from store
	user, err := a.userStore.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		// Don't reveal whether user exists
		a.recordFailure(username)
		return nil, ErrAuthenticationFailed
	}
	
	// Check if user account is locked
	if user.Locked {
		return nil, ErrAccountLocked
	}
	
	// Verify password
	if !VerifyPassword(password, user.PasswordHash) {
		a.recordFailure(username)
		return nil, ErrAuthenticationFailed
	}
	
	// Check if password has expired
	if user.ExpireAt != nil && time.Now().After(*user.ExpireAt) {
		return nil, ErrPasswordExpired
	}
	
	// Clear failure record on successful authentication
	a.attempts.Delete(username)
	
	return user, nil
}

// isLocked checks if a user is temporarily locked due to failed attempts.
func (a *Authenticator) isLocked(username string) bool {
	v, ok := a.attempts.Load(username)
	if !ok {
		return false
	}
	
	info := v.(*attemptInfo)
	if info.count >= a.maxAttempts {
		// Check if lock duration has passed
		if time.Since(info.lastFail) < a.lockDuration {
			return true
		}
		// Lock expired, reset
		a.attempts.Delete(username)
	}
	
	return false
}

// recordFailure records a failed authentication attempt.
func (a *Authenticator) recordFailure(username string) {
	v, _ := a.attempts.LoadOrStore(username, &attemptInfo{})
	info := v.(*attemptInfo)
	info.count++
	info.lastFail = time.Now()
}

// SetMaxAttempts configures the maximum number of failed attempts before locking.
func (a *Authenticator) SetMaxAttempts(max int) {
	a.maxAttempts = max
}

// SetLockDuration configures how long a user is locked after max failed attempts.
func (a *Authenticator) SetLockDuration(duration time.Duration) {
	a.lockDuration = duration
}

// ResetFailures clears the failure count for a user.
func (a *Authenticator) ResetFailures(username string) {
	a.attempts.Delete(username)
}

var (
	ErrAccountLocked        = &AuthError{Code: "ACCOUNT_LOCKED", Message: "account is locked"}
	ErrAuthenticationFailed = &AuthError{Code: "AUTH_FAILED", Message: "authentication failed"}
	ErrPasswordExpired      = &AuthError{Code: "PASSWORD_EXPIRED", Message: "password has expired"}
)
