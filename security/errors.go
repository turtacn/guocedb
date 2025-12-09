// Package security provides unified security management for GuoceDB.
package security

import "errors"

// Authentication errors
var (
	ErrUserNotFound         = errors.New("user not found")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrAccountLocked        = errors.New("account is locked")
	ErrPasswordExpired      = errors.New("password has expired")
	ErrInvalidCredentials   = errors.New("invalid credentials")
)

// Authorization errors
var (
	ErrAccessDenied     = errors.New("access denied")
	ErrInsufficientPriv = errors.New("insufficient privileges")
	ErrRoleNotFound     = errors.New("role not found")
)

// Audit errors
var (
	ErrAuditLogFailed = errors.New("failed to write audit log")
)
