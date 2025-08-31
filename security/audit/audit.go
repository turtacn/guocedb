// Package audit provides audit logging for guocedb.
package audit

import (
	"time"

	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/security/authn"
)

// Event represents an audit log entry.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	User      string    `json:"user"`
	ClientIP  string    `json:"client_ip"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Query     string    `json:"query"`
	Status    string    `json:"status"` // e.g., "SUCCESS", "FAILURE"
	Error     string    `json:"error,omitempty"`
}

// Auditor is the interface for the audit logging service.
type Auditor interface {
	// Log records an audit event.
	Log(event Event) error
}

// LoggerAuditor is an implementation of Auditor that uses the system logger.
type LoggerAuditor struct {
	logger log.Logger
}

// NewLoggerAuditor creates a new auditor that writes to the system log.
func NewLoggerAuditor() *LoggerAuditor {
	// Create a dedicated logger for audit events.
	// This could be configured to write to a separate file.
	auditLogger := log.NewLogger("info", "json")
	return &LoggerAuditor{logger: auditLogger}
}

// Log records an audit event by writing it as a structured log message.
func (a *LoggerAuditor) Log(event Event) error {
	a.logger.WithFields(map[string]interface{}{
		"audit_event": true, // A flag to easily filter for audit logs
		"user":        event.User,
		"client_ip":   event.ClientIP,
		"action":      event.Action,
		"resource":    event.Resource,
		"query":       event.Query,
		"status":      event.Status,
		"error":       event.Error,
	}).Info("Audit Event")
	return nil
}

// LogQuerySuccess is a helper to log a successful query.
func (a *LoggerAuditor) LogQuerySuccess(user *authn.User, clientIP, query string) {
	event := Event{
		Timestamp: time.Now(),
		User:      user.Name,
		ClientIP:  clientIP,
		Action:    "QUERY",
		Query:     query,
		Status:    "SUCCESS",
	}
	_ = a.Log(event)
}

// LogQueryFailure is a helper to log a failed query.
func (a *LoggerAuditor) LogQueryFailure(user *authn.User, clientIP, query string, err error) {
	event := Event{
		Timestamp: time.Now(),
		User:      user.Name,
		ClientIP:  clientIP,
		Action:    "QUERY",
		Query:     query,
		Status:    "FAILURE",
		Error:     err.Error(),
	}
	_ = a.Log(event)
}

// Enforce interface compliance
var _ Auditor = (*LoggerAuditor)(nil)
