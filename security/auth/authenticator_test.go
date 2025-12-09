package auth

import (
	"context"
	"testing"
	"time"

	"github.com/turtacn/guocedb/security/authz"
)

func TestAuthenticateSuccess(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryUserStore()
	
	// Create a test user
	hash, _ := HashPassword("secret123")
	err := store.CreateUser(ctx, &User{
		Username:     "testuser",
		PasswordHash: hash,
		Privileges:   authz.PrivilegeReadOnly,
	})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	authn := NewAuthenticator(store)
	user, err := authn.Authenticate(ctx, "testuser", "secret123")
	
	if err != nil {
		t.Fatalf("Authentication should succeed: %v", err)
	}
	
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}
}

func TestAuthenticateWrongPassword(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryUserStore()
	
	hash, _ := HashPassword("secret123")
	store.CreateUser(ctx, &User{
		Username:     "testuser",
		PasswordHash: hash,
	})
	
	authn := NewAuthenticator(store)
	_, err := authn.Authenticate(ctx, "testuser", "wrongpass")
	
	if err == nil {
		t.Error("Authentication should fail with wrong password")
	}
	
	if err != ErrAuthenticationFailed {
		t.Errorf("Expected ErrAuthenticationFailed, got %v", err)
	}
}

func TestAuthenticateUserNotFound(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryUserStore()
	
	authn := NewAuthenticator(store)
	_, err := authn.Authenticate(ctx, "nonexistent", "password")
	
	if err == nil {
		t.Error("Authentication should fail for non-existent user")
	}
	
	if err != ErrAuthenticationFailed {
		t.Errorf("Expected ErrAuthenticationFailed, got %v", err)
	}
}

func TestAuthenticateLockAfterFailures(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryUserStore()
	
	hash, _ := HashPassword("secret123")
	store.CreateUser(ctx, &User{
		Username:     "testuser",
		PasswordHash: hash,
	})
	
	authn := NewAuthenticator(store)
	authn.SetMaxAttempts(3)
	
	// Attempt 3 failures
	for i := 0; i < 3; i++ {
		_, err := authn.Authenticate(ctx, "testuser", "wrong")
		if err == nil {
			t.Error("Authentication should fail")
		}
	}
	
	// Fourth attempt should be locked, even with correct password
	_, err := authn.Authenticate(ctx, "testuser", "secret123")
	if err != ErrAccountLocked {
		t.Errorf("Expected ErrAccountLocked, got %v", err)
	}
}

func TestAuthenticateLockExpiry(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryUserStore()
	
	hash, _ := HashPassword("secret123")
	store.CreateUser(ctx, &User{
		Username:     "testuser",
		PasswordHash: hash,
	})
	
	authn := NewAuthenticator(store)
	authn.SetMaxAttempts(2)
	authn.SetLockDuration(100 * time.Millisecond)
	
	// Trigger lock
	authn.Authenticate(ctx, "testuser", "wrong")
	authn.Authenticate(ctx, "testuser", "wrong")
	
	// Should be locked
	_, err := authn.Authenticate(ctx, "testuser", "secret123")
	if err != ErrAccountLocked {
		t.Error("Should be locked")
	}
	
	// Wait for lock to expire
	time.Sleep(150 * time.Millisecond)
	
	// Should succeed now
	user, err := authn.Authenticate(ctx, "testuser", "secret123")
	if err != nil {
		t.Errorf("Should authenticate after lock expiry: %v", err)
	}
	if user == nil {
		t.Error("User should not be nil")
	}
}

func TestAuthenticateLockedAccount(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryUserStore()
	
	hash, _ := HashPassword("secret123")
	store.CreateUser(ctx, &User{
		Username:     "testuser",
		PasswordHash: hash,
		Locked:       true,
	})
	
	authn := NewAuthenticator(store)
	_, err := authn.Authenticate(ctx, "testuser", "secret123")
	
	if err != ErrAccountLocked {
		t.Errorf("Expected ErrAccountLocked for locked account, got %v", err)
	}
}

func TestAuthenticatePasswordExpired(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryUserStore()
	
	hash, _ := HashPassword("secret123")
	expireTime := time.Now().Add(-1 * time.Hour)
	store.CreateUser(ctx, &User{
		Username:     "testuser",
		PasswordHash: hash,
		ExpireAt:     &expireTime,
	})
	
	authn := NewAuthenticator(store)
	_, err := authn.Authenticate(ctx, "testuser", "secret123")
	
	if err != ErrPasswordExpired {
		t.Errorf("Expected ErrPasswordExpired, got %v", err)
	}
}

func TestResetFailures(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryUserStore()
	
	hash, _ := HashPassword("secret123")
	store.CreateUser(ctx, &User{
		Username:     "testuser",
		PasswordHash: hash,
	})
	
	authn := NewAuthenticator(store)
	authn.SetMaxAttempts(2)
	
	// Trigger failures
	authn.Authenticate(ctx, "testuser", "wrong")
	authn.Authenticate(ctx, "testuser", "wrong")
	
	// Reset failures
	authn.ResetFailures("testuser")
	
	// Should be able to authenticate now
	user, err := authn.Authenticate(ctx, "testuser", "secret123")
	if err != nil {
		t.Errorf("Should authenticate after reset: %v", err)
	}
	if user == nil {
		t.Error("User should not be nil")
	}
}
