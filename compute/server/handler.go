package server

import (
	"context"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	errors "gopkg.in/src-d/go-errors.v1"
	"github.com/turtacn/guocedb/compute/executor"
	"github.com/turtacn/guocedb/compute/auth"
	"github.com/turtacn/guocedb/compute/sql"
	"github.com/turtacn/guocedb/compute/transaction"

	"github.com/sirupsen/logrus"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/dolthub/vitess/go/vt/sqlparser"
)

var regKillCmd = regexp.MustCompile(`^kill (?:(query|connection) )?(\d+)$`)

var errConnectionNotFound = errors.NewKind("Connection not found: %c")

// TODO parametrize
const rowsBatch = 100

// Handler is a connection handler for a SQLe engine.
type Handler struct {
	mu              sync.Mutex
	e               *executor.Engine
	sm              *SessionManager
	sessionMgr      *EnhancedSessionManager // New session manager for enhanced functionality
	txnManager      *transaction.Manager    // Transaction manager
	c               map[uint32]*mysql.Conn
	disableMultiStmts bool
}

// NewHandler creates a new Handler given a SQLe engine.
func NewHandler(e *executor.Engine, sm *SessionManager) *Handler {
	return &Handler{
		e:          e,
		sm:         sm,
		sessionMgr: NewEnhancedSessionManager(),
		txnManager: transaction.NewManager(nil), // Will be updated when storage is available
		c:          make(map[uint32]*mysql.Conn),
	}
}

// NewHandlerWithTxnManager creates a new Handler with a specific transaction manager.
func NewHandlerWithTxnManager(e *executor.Engine, sm *SessionManager, txnMgr *transaction.Manager) *Handler {
	return &Handler{
		e:          e,
		sm:         sm,
		sessionMgr: NewEnhancedSessionManager(),
		txnManager: txnMgr,
		c:          make(map[uint32]*mysql.Conn),
	}
}

// NewConnection reports that a new connection has been established.
func (h *Handler) NewConnection(c *mysql.Conn) {
	h.mu.Lock()
	if _, ok := h.c[c.ConnectionID]; !ok {
		h.c[c.ConnectionID] = c
	}
	h.mu.Unlock()

	// Create a new session for this connection
	user := c.User
	client := "unknown"
	// Try to get remote address safely
	func() {
		defer func() {
			if r := recover(); r != nil {
				// If RemoteAddr() panics, just use "unknown"
				client = "unknown"
			}
		}()
		if addr := c.RemoteAddr(); addr != nil {
			client = addr.String()
		}
	}()
	sess := h.sessionMgr.NewSession(user, client)
	c.ConnectionID = sess.ID()

	logrus.Infof("NewConnection: client %v, user %s", c.ConnectionID, user)
}

// ConnectionClosed reports that a connection has been closed.
func (h *Handler) ConnectionClosed(c *mysql.Conn) {
	h.sm.CloseConn(c)
	h.sessionMgr.RemoveSession(c.ConnectionID)

	h.mu.Lock()
	delete(h.c, c.ConnectionID)
	h.mu.Unlock()

	if err := h.e.Catalog.UnlockTables(nil, c.ConnectionID); err != nil {
		logrus.Errorf("unable to unlock tables on session close: %s", err)
	}

	logrus.Infof("ConnectionClosed: client %v", c.ConnectionID)
}

// ComQuery executes a SQL query on the SQLe engine.
func (h *Handler) ComQuery(
	ctx context.Context,
	c *mysql.Conn,
	query string,
	callback mysql.ResultSpoolFn,
) (err error) {
	// Get the session and create context with current database
	sess := h.sessionMgr.GetSession(c.ConnectionID)
	var sqlCtx *sql.Context
	if sess != nil {
		sqlCtx = sess.Context(ctx)
		sqlCtx = sql.NewContext(ctx, sql.WithSession(sqlCtx.Session), sql.WithQuery(query))
	} else {
		// Fall back to old session manager
		sqlCtx = h.sm.NewContextWithQuery(c, query)
	}

	handled, err := h.handleKill(c, query)
	if err != nil {
		return err
	}

	if handled {
		return nil
	}

	// Handle transaction statements
	handled, err = h.handleTransactionStatements(sess, query, callback)
	if err != nil {
		return err
	}

	if handled {
		return nil
	}

	start := time.Now()
	schema, rows, err := h.e.Query(sqlCtx, query)
	defer func() {
		if q, ok := h.e.Auth.(*auth.Audit); ok {
			q.Query(sqlCtx, time.Since(start), err)
		}
	}()

	if err != nil {
		return err
	}

	var r *sqltypes.Result
	var proccesedAtLeastOneBatch bool
	for {
		if r == nil {
			r = &sqltypes.Result{Fields: schemaToFields(schema)}
		}

		if r.RowsAffected == rowsBatch {
			if err := callback(r, true); err != nil {
				return err
			}

			r = nil
			proccesedAtLeastOneBatch = true

			continue
		}

		row, err := rows.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		r.Rows = append(r.Rows, rowToSQL(schema, row))
		r.RowsAffected++
	}

	if err := rows.Close(); err != nil {
		return err
	}

	// Even if r.RowsAffected = 0, the callback must be
	// called to update the state in the go-vitess' listener
	// and avoid returning errors when the query doesn't
	// produce any results.
	if r != nil && (r.RowsAffected == 0 && proccesedAtLeastOneBatch) {
		return nil
	}

	return callback(r, false)
}

// ComInitDB changes the database for the current connection.
func (h *Handler) ComInitDB(c *mysql.Conn, schemaName string) error {
	// Get the session for this connection
	sess := h.sessionMgr.GetSession(c.ConnectionID)
	if sess == nil {
		return mysql.NewSQLError(mysql.ERUnknownComError, mysql.SSUnknownSQLState, "session not found")
	}

	// Check if the database exists
	_, err := h.e.Catalog.Database(schemaName)
	if err != nil {
		// Check if it's a "database not found" error
		if sql.ErrDatabaseNotFound.Is(err) {
			return mysql.NewSQLError(mysql.ERBadDb, mysql.SSClientError, "Unknown database '%s'", schemaName)
		}
		// Other errors
		return mysql.NewSQLError(mysql.ERUnknownComError, mysql.SSUnknownSQLState, "Error accessing database: %s", err.Error())
	}

	// Set the current database in the session
	sess.SetCurrentDB(schemaName)
	
	// Also set it in the old session manager for compatibility
	// But handle potential panics from RemoteAddr()
	func() {
		defer func() {
			if r := recover(); r != nil {
				// If SetDB panics, just ignore it
			}
		}()
		h.sm.SetDB(c, schemaName)
	}()
	
	return nil
}

// ComMultiQuery executes multiple SQL queries on the SQLe engine.
func (h *Handler) ComMultiQuery(
	ctx context.Context,
	c *mysql.Conn,
	query string,
	callback mysql.ResultSpoolFn,
) (string, error) {
	if h.disableMultiStmts {
		// Fall back to single query
		err := h.ComQuery(ctx, c, query, callback)
		return "", err
	}

	// Split the query into individual statements
	statements := h.splitStatements(query)
	if len(statements) == 0 {
		return "", nil
	}

	// Execute the first statement
	firstStmt := statements[0]
	err := h.ComQuery(ctx, c, firstStmt, callback)
	if err != nil {
		return "", err
	}

	// Return the remainder of the query (all statements after the first)
	if len(statements) > 1 {
		remainder := strings.Join(statements[1:], ";")
		return remainder, nil
	}

	return "", nil
}

func (h *Handler) ComPrepare(ctx context.Context, c *mysql.Conn, query string, prepare *mysql.PrepareData) ([]*query.Field, error) {
	// Not implemented
	return nil, nil
}

func (h *Handler) ComStmtExecute(ctx context.Context, c *mysql.Conn, prepare *mysql.PrepareData, callback func(*sqltypes.Result) error) error {
	// Not implemented
	return nil
}

func (h *Handler) ConnectionAborted(c *mysql.Conn, reason string) error {
	logrus.Infof("ConnectionAborted: client %v, reason: %s", c.ConnectionID, reason)
	return nil
}

func (h *Handler) ComResetConnection(c *mysql.Conn) error {
	// Not implemented
	return nil
}

func (h *Handler) ParserOptionsForConnection(c *mysql.Conn) (sqlparser.ParserOptions, error) {
	return sqlparser.ParserOptions{}, nil
}

// WarningCount is called at the end of each query to obtain
// the value to be returned to the client in the EOF packet.
// Note that this will be called either in the context of the
// ComQuery callback if the result does not contain any fields,
// or after the last ComQuery call completes.
func (h *Handler) WarningCount(c *mysql.Conn) uint16 {
	sess, ok := h.sm.sessions[c.ConnectionID]
	if !ok {
		return 0
	}

	return sess.WarningCount()
}

func (h *Handler) handleKill(conn *mysql.Conn, query string) (bool, error) {
	q := strings.ToLower(query)
	s := regKillCmd.FindStringSubmatch(q)
	if s == nil {
		return false, nil
	}

	id, err := strconv.ParseUint(s[2], 10, 64)
	if err != nil {
		return false, err
	}

	// KILL CONNECTION and KILL should close the connection. KILL QUERY only
	// cancels the query.
	//
	// https://dev.mysql.com/doc/refman/5.7/en/kill.html

	if s[1] == "query" {
		logrus.Infof("kill query: id %v", id)
		h.e.Catalog.Kill(id)
	} else {
		logrus.Infof("kill connection: id %v, pid: %v", conn.ConnectionID, id)
		h.mu.Lock()
		c, ok := h.c[conn.ConnectionID]
		delete(h.c, conn.ConnectionID)
		h.mu.Unlock()

		if !ok {
			return false, errConnectionNotFound.New(conn.ConnectionID)
		}

		h.e.Catalog.KillConnection(id)
		h.sm.CloseConn(c)
		c.Close()
	}

	return true, nil
}

func rowToSQL(s sql.Schema, row sql.Row) []sqltypes.Value {
	o := make([]sqltypes.Value, len(row))
	for i, v := range row {
		o[i] = s[i].Type.SQL(v)
	}

	return o
}

func schemaToFields(s sql.Schema) []*query.Field {
	fields := make([]*query.Field, len(s))
	for i, c := range s {
		fields[i] = &query.Field{
			Name: c.Name,
			Type: c.Type.Type(),
		}
	}

	return fields
}

// splitStatements splits a multi-statement query into individual statements
func (h *Handler) splitStatements(query string) []string {
	// Use vitess sqlparser to split statements properly
	statements, err := sqlparser.SplitStatementToPieces(query)
	if err != nil {
		// Fall back to simple semicolon splitting if parsing fails
		return h.simpleSplitStatements(query)
	}

	result := make([]string, 0)
	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// simpleSplitStatements provides a fallback for simple semicolon splitting
func (h *Handler) simpleSplitStatements(query string) []string {
	parts := strings.Split(query, ";")
	result := make([]string, 0)
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// handleTransactionStatements handles BEGIN, COMMIT, and ROLLBACK statements
func (h *Handler) handleTransactionStatements(sess *Session, query string, callback mysql.ResultSpoolFn) (bool, error) {
	if sess == nil {
		return false, nil
	}

	// Parse the SQL statement
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return false, nil // Not a transaction statement, let normal processing handle it
	}

	switch stmt.(type) {
	case *sqlparser.Begin:
		return true, h.handleBegin(sess, callback)
	case *sqlparser.Commit:
		return true, h.handleCommit(sess, callback)
	case *sqlparser.Rollback:
		return true, h.handleRollback(sess, callback)
	default:
		return false, nil // Not a transaction statement
	}
}

// handleBegin handles BEGIN statements
func (h *Handler) handleBegin(sess *Session, callback mysql.ResultSpoolFn) error {
	if sess.GetTransaction() != nil {
		return mysql.NewSQLError(1400, "HY000", "Transaction already started")
	}

	txn, err := h.txnManager.Begin(nil)
	if err != nil {
		return h.convertError(err)
	}

	sess.SetTransaction(txn)
	
	result := &sqltypes.Result{}
	return callback(result, false)
}

// handleCommit handles COMMIT statements
func (h *Handler) handleCommit(sess *Session, callback mysql.ResultSpoolFn) error {
	txn := sess.GetTransaction()
	if txn == nil {
		// No active transaction, silently succeed
		result := &sqltypes.Result{}
		return callback(result, false)
	}

	if t, ok := txn.(*transaction.Transaction); ok {
		err := h.txnManager.Commit(t)
		sess.SetTransaction(nil)
		if err != nil {
			return h.convertError(err)
		}
	}

	result := &sqltypes.Result{}
	return callback(result, false)
}

// handleRollback handles ROLLBACK statements
func (h *Handler) handleRollback(sess *Session, callback mysql.ResultSpoolFn) error {
	txn := sess.GetTransaction()
	if txn == nil {
		// No active transaction, silently succeed
		result := &sqltypes.Result{}
		return callback(result, false)
	}

	if t, ok := txn.(*transaction.Transaction); ok {
		err := h.txnManager.Rollback(t)
		sess.SetTransaction(nil)
		if err != nil {
			return h.convertError(err)
		}
	}

	result := &sqltypes.Result{}
	return callback(result, false)
}

// convertError converts internal errors to MySQL errors
func (h *Handler) convertError(err error) error {
	if err == nil {
		return nil
	}
	// For now, convert all errors to generic MySQL errors
	return mysql.NewSQLError(1105, "HY000", err.Error())
}
