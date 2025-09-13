package mysql

import (
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/turtacn/guocedb/compute/executor"
)

// Handler is the main handler for the MySQL server, implementing the server.Handler interface.
// It connects the protocol layer with the query execution engine.
type Handler struct {
	engine *executor.Engine
	// Add other dependencies like the connection manager if needed
}

var _ server.Handler = (*Handler)(nil)

// NewHandler creates a new server handler.
func NewHandler(engine *executor.Engine) *Handler {
	return &Handler{
		engine: engine,
	}
}

// NewSession is called when a new client connects. It uses our custom session builder.
func (h *Handler) NewSession(conn *server.Conn) (sql.Session, error) {
	return SessionBuilder(conn.RemoteAddr().String(), conn.User, conn.Password)
}

// CloseSession is called when a client disconnects.
func (h *Handler) CloseSession(conn *server.Conn) {
	// Here you could use a ConnectionManager to track the session's end.
}

// ComQuery is called when a client sends a query. This is the main entry point for query execution.
func (h *Handler) ComQuery(c *server.Conn, query string) (sql.Schema, sql.RowIter, error) {
	ctx, err := h.NewSession(c)
	if err != nil {
		return nil, nil, err
	}
	// The GMS engine has a complex Query method. We are simplifying here by just calling our engine.
	// In a real implementation, you would need to handle transaction state, etc., on the session.
	return h.engine.Query(ctx, query)
}

// Other methods of the server.Handler interface can be implemented as needed.
// For now, these stubs are sufficient for basic functionality.
func (h *Handler) ComPrepare(c *server.Conn, query string) (sql.Schema, []*sql.ColumnDefinition, error) {
	return nil, nil, nil
}
func (h *Handler) ComStatementExecute(c *server.Conn, statementId uint32, values []sql.Value) (sql.Schema, sql.RowIter, error) {
	return nil, nil, nil
}
func (h *Handler) ComResetConnection(c *server.Conn) {}
func (h *Handler) ComChangeUser(c *server.Conn, user string, password string, db string) (bool, error) {
	return true, nil
}
func (h *Handler) ComPing(c *server.Conn) error {
	return nil
}
func (h *Handler) ComInitDB(c *server.Conn, dbName string) error {
	c.Session.SetCurrentDatabase(dbName)
	return nil
}
func (h *Handler) ComCloseStatement(c *server.Conn, statementId uint32) error {
	return nil
}
func (h *Handler) ComQuit(c *server.Conn) {}
