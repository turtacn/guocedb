// Package audit provides audit logging for GuoceDB.
package audit

import (
	"strings"
	"time"
)

// EventType represents the type of audit event.
type EventType string

const (
	EventTypeAuthentication EventType = "AUTHENTICATION"
	EventTypeAuthorization  EventType = "AUTHORIZATION"
	EventTypeQuery          EventType = "QUERY"
	EventTypeDDL            EventType = "DDL"
	EventTypeDML            EventType = "DML"
	EventTypeAdmin          EventType = "ADMIN"
	EventTypeConnection     EventType = "CONNECTION"
)

// EventResult represents the outcome of an event.
type EventResult string

const (
	ResultSuccess EventResult = "SUCCESS"
	ResultFailure EventResult = "FAILURE"
	ResultDenied  EventResult = "DENIED"
)

// AuditEvent represents a single audit log entry.
type AuditEvent struct {
	Timestamp    time.Time              `json:"timestamp"`
	EventType    EventType              `json:"event_type"`
	Result       EventResult            `json:"result"`
	Username     string                 `json:"username"`
	ClientIP     string                 `json:"client_ip"`
	Database     string                 `json:"database,omitempty"`
	Statement    string                 `json:"statement,omitempty"`
	Object       string                 `json:"object,omitempty"`
	Privilege    string                 `json:"privilege,omitempty"`
	ErrorMsg     string                 `json:"error_msg,omitempty"`
	Duration     time.Duration          `json:"duration_ms,omitempty"`
	RowsAffected int64                  `json:"rows_affected,omitempty"`
	Extra        map[string]interface{} `json:"extra,omitempty"`
}

// NewAuthenticationEvent creates an audit event for authentication attempts.
func NewAuthenticationEvent(username, clientIP string, success bool) *AuditEvent {
	result := ResultSuccess
	if !success {
		result = ResultFailure
	}
	
	return &AuditEvent{
		Timestamp: time.Now(),
		EventType: EventTypeAuthentication,
		Result:    result,
		Username:  username,
		ClientIP:  clientIP,
	}
}

// NewAuthorizationEvent creates an audit event for authorization checks.
func NewAuthorizationEvent(username, clientIP, database, object, privilege string, denied bool) *AuditEvent {
	result := ResultSuccess
	if denied {
		result = ResultDenied
	}
	
	return &AuditEvent{
		Timestamp: time.Now(),
		EventType: EventTypeAuthorization,
		Result:    result,
		Username:  username,
		ClientIP:  clientIP,
		Database:  database,
		Object:    object,
		Privilege: privilege,
	}
}

// NewQueryEvent creates an audit event for query execution.
func NewQueryEvent(username, clientIP, database, statement string, duration time.Duration, rowsAffected int64) *AuditEvent {
	return &AuditEvent{
		Timestamp:    time.Now(),
		EventType:    EventTypeQuery,
		Result:       ResultSuccess,
		Username:     username,
		ClientIP:     clientIP,
		Database:     database,
		Statement:    truncateStatement(statement, 1000),
		Duration:     duration,
		RowsAffected: rowsAffected,
	}
}

// NewConnectionEvent creates an audit event for connection attempts.
func NewConnectionEvent(username, clientIP string, success bool) *AuditEvent {
	result := ResultSuccess
	if !success {
		result = ResultFailure
	}
	
	return &AuditEvent{
		Timestamp: time.Now(),
		EventType: EventTypeConnection,
		Result:    result,
		Username:  username,
		ClientIP:  clientIP,
	}
}

// truncateStatement truncates a SQL statement to a maximum length.
func truncateStatement(statement string, maxLen int) string {
	// Remove excessive whitespace
	statement = strings.TrimSpace(statement)
	statement = strings.Join(strings.Fields(statement), " ")
	
	if len(statement) <= maxLen {
		return statement
	}
	
	return statement[:maxLen] + "..."
}
