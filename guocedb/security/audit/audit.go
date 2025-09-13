package audit

import (
	"time"

	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/security/authn"
)

// Event represents a security-relevant event that should be audited.
type Event struct {
	Timestamp time.Time
	User      *authn.User
	ClientIP  string
	Action    string
	Resource  string
	Success   bool
	Details   map[string]interface{}
}

// Auditor is the interface for logging audit events.
type Auditor interface {
	// Log records an audit event.
	Log(event Event) error
}

// LogAuditor is an implementation of Auditor that writes events to the standard logger.
type LogAuditor struct {
	logger log.Logger
}

// NewLogAuditor creates a new auditor that logs to the provided logger.
func NewLogAuditor(logger log.Logger) Auditor {
	// We can create a dedicated logger for audit events with a specific field.
	return &LogAuditor{
		logger: logger.WithField("component", "audit"),
	}
}

// Log formats the audit event and writes it to the log.
func (a *LogAuditor) Log(event Event) error {
	fields := map[string]interface{}{
		"client_ip": event.ClientIP,
		"action":    event.Action,
		"resource":  event.Resource,
		"success":   event.Success,
	}

	if event.User != nil {
		fields["user"] = event.User.Name
	} else {
		fields["user"] = "anonymous"
	}

	for k, v := range event.Details {
		fields[k] = v
	}

	a.logger.WithFields(fields).Infof("Audit Event: %s", event.Action)
	return nil
}

// TODO: Implement a more robust auditor that writes to a dedicated, tamper-proof audit log file or a remote sink.
// TODO: Add support for filtering and querying audit logs.
// TODO: Ensure audit logs meet compliance requirements (e.g., GDPR, HIPAA).
