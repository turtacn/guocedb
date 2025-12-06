package auth

import (
	"net"
	"time"

	"github.com/turtacn/guocedb/compute/sql"
	"github.com/dolthub/vitess/go/mysql"

	"github.com/sirupsen/logrus"
)

// AuditMethod is called to log the audit trail of actions.
type AuditMethod interface {
	// Authentication logs an authentication event.
	Authentication(user, address string, err error)
	// Authorization logs an authorization event.
	Authorization(ctx *sql.Context, p Permission, err error)
	// Query logs a query execution.
	Query(ctx *sql.Context, d time.Duration, err error)
}

// MysqlAudit wraps mysql.AuthServer to emit audit trails.
type MysqlAudit struct {
	mysql.AuthServer
	audit AuditMethod
}

// AuthMethods returns the wrapped auth methods.
func (m *MysqlAudit) AuthMethods() []mysql.AuthMethod {
	methods := m.AuthServer.AuthMethods()
	wrapped := make([]mysql.AuthMethod, len(methods))
	for i, method := range methods {
		wrapped[i] = &AuditAuthMethod{
			AuthMethod: method,
			audit:      m.audit,
		}
	}
	return wrapped
}

// DefaultAuthMethodDescription delegates to the underlying server.
func (m *MysqlAudit) DefaultAuthMethodDescription() mysql.AuthMethodDescription {
	return m.AuthServer.DefaultAuthMethodDescription()
}

// AuditAuthMethod wraps mysql.AuthMethod to intercept and log authentication.
type AuditAuthMethod struct {
	mysql.AuthMethod
	audit AuditMethod
}

// HandleAuthPluginData intercepts the authentication step to log the result.
func (am *AuditAuthMethod) HandleAuthPluginData(
	conn *mysql.Conn,
	user string,
	serverAuthPluginData []byte,
	clientAuthPluginData []byte,
	remoteAddr net.Addr,
) (mysql.Getter, error) {
	getter, err := am.AuthMethod.HandleAuthPluginData(conn, user, serverAuthPluginData, clientAuthPluginData, remoteAddr)

	// We use remoteAddr.String() for the address.
	// If remoteAddr is nil (unlikely in real connection), we handle it safely?
	// The interface implies it's a net.Addr, so String() should be safe if not nil.
	addrStr := ""
	if remoteAddr != nil {
		addrStr = remoteAddr.String()
	}

	am.audit.Authentication(user, addrStr, err)
	return getter, err
}

// NewAudit creates a wrapped Auth that sends audit trails to the specified
// method.
func NewAudit(auth Auth, method AuditMethod) Auth {
	return &Audit{
		auth:   auth,
		method: method,
	}
}

// Audit is an Auth method proxy that sends audit trails to the specified
// AuditMethod.
type Audit struct {
	auth   Auth
	method AuditMethod
}

// Mysql implements Auth interface.
func (a *Audit) Mysql() mysql.AuthServer {
	return &MysqlAudit{
		AuthServer: a.auth.Mysql(),
		audit:      a.method,
	}
}

// Allowed implements Auth interface.
func (a *Audit) Allowed(ctx *sql.Context, permission Permission) error {
	err := a.auth.Allowed(ctx, permission)
	a.method.Authorization(ctx, permission, err)

	return err
}

// Query implements AuditQuery interface.
func (a *Audit) Query(ctx *sql.Context, d time.Duration, err error) {
	if q, ok := a.auth.(*Audit); ok {
		q.Query(ctx, d, err)
	}

	a.method.Query(ctx, d, err)
}

// NewAuditLog creates a new AuditMethod that logs to a logrus.Logger.
func NewAuditLog(l *logrus.Logger) AuditMethod {
	la := l.WithField("system", "audit")

	return &AuditLog{
		log: la,
	}
}

const auditLogMessage = "audit trail"

// AuditLog logs audit trails to a logrus.Logger.
type AuditLog struct {
	log *logrus.Entry
}

// Authentication implements AuditMethod interface.
func (a *AuditLog) Authentication(user string, address string, err error) {
	fields := logrus.Fields{
		"action":  "authentication",
		"user":    user,
		"address": address,
		"success": true,
	}

	if err != nil {
		fields["success"] = false
		fields["err"] = err
	}

	a.log.WithFields(fields).Info(auditLogMessage)
}

func auditInfo(ctx *sql.Context, err error) logrus.Fields {
	fields := logrus.Fields{
		"user":          ctx.Client().User,
		"query":         ctx.Query(),
		"address":       ctx.Client().Address,
		"connection_id": ctx.Session.ID(),
		"pid":           ctx.Pid(),
		"success":       true,
	}

	if err != nil {
		fields["success"] = false
		fields["err"] = err
	}

	return fields
}

// Authorization implements AuditMethod interface.
func (a *AuditLog) Authorization(ctx *sql.Context, p Permission, err error) {
	fields := auditInfo(ctx, err)
	fields["action"] = "authorization"
	fields["permission"] = p.String()

	a.log.WithFields(fields).Info(auditLogMessage)
}

func (a *AuditLog) Query(ctx *sql.Context, d time.Duration, err error) {
	fields := auditInfo(ctx, err)
	fields["action"] = "query"
	fields["duration"] = d

	a.log.WithFields(fields).Info(auditLogMessage)
}
