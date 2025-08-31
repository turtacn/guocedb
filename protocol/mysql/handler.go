// Package mysql provides MySQL protocol handling for guocedb.
package mysql

import (
	"time"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/compute/executor"
	"github.com/turtacn/guocedb/maintenance/metrics"
)

// Handler is the custom implementation of the go-mysql-server Handler interface.
// It connects the GMS server with our custom query execution engine.
type Handler struct {
	engine *executor.Engine
}

// NewHandler creates a new protocol handler.
func NewHandler(engine *executor.Engine) *Handler {
	return &Handler{engine: engine}
}

// NewSession is called when a new client connection is established.
func (h *Handler) NewSession(conn *server.Conn) (sql.Session, error) {
	// Create a new GMS session.
	// We can customize the session here, for example by setting a default database.
	return sql.NewBaseSessionWithClientServer(
		"guocedb",
		conn.ConnectionID,
		conn.Peer,
		conn.User,
		conn.UserHost,
		sql.SystemVariables,
		sql.NewInMemCollationPersister(),
	), nil
}

// CloseSession is called when a client connection is closed.
func (h *Handler) CloseSession(conn *server.Conn, sess sql.Session) {
	// Perform any cleanup needed for the session.
}

// ComQuery is called when a client sends a query.
func (h *Handler) ComQuery(
	conn *server.Conn,
	sess sql.Session,
	query string,
	callback func(res *sql.Result, err error),
) {
	// Create a new SQL context for the query.
	ctx := sql.NewContext(sess)
	ctx.SetQuery(query)

	start := time.Now()
	schema, rowIter, err := h.engine.Query(ctx, query)
	metrics.RecordQuery(start, err)

	// The callback sends the result back to the client.
	if err != nil {
		callback(nil, err)
		return
	}

	result := &sql.Result{
		Schema:   schema,
		Rows:     rowIter,
		Affected: 0, // This would be set for DML queries.
	}
	callback(result, nil)
}

// The following methods are part of the Handler interface but are less critical
// for a basic implementation. We'll provide placeholder implementations.

func (h *Handler) ComPrepare(conn *server.Conn, sess sql.Session, query string, callback func(res *sql.Result, err error)) {
	// Placeholder for prepared statements
	callback(nil, nil)
}

func (h *Handler) ComStatementExecute(
	conn *server.Conn,
	sess sql.Session,
	statementId uint32,
	params []sql.Expression,
	callback func(res *sql.Result, err error),
) {
	// Placeholder for prepared statements
	callback(nil, nil)
}

func (h *Handler) ComResetConnection(conn *server.Conn, sess sql.Session) {
	// Placeholder
}

func (h *Handler) ComChangeUser(conn *server.Conn, sess sql.Session, user, host, db string) error {
	// Placeholder
	return nil
}

func (h *Handler) ComInitDB(conn *server.Conn, sess sql.Session, dbName string) error {
	sess.SetCurrentDatabase(dbName)
	return nil
}

// server.Handler interface compliance check
var _ server.Handler = (*Handler)(nil)
