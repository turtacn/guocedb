package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/turtacn/guocedb/compute/sql"
	"github.com/turtacn/guocedb/compute/executor"
	"github.com/turtacn/guocedb/compute/analyzer"
	"github.com/turtacn/guocedb/compute/optimizer"

)

// mockCatalog implements a simple catalog for testing
type mockCatalog struct {
	databases map[string]sql.Database
}

func newMockCatalog() *mockCatalog {
	return &mockCatalog{
		databases: make(map[string]sql.Database),
	}
}

func (m *mockCatalog) Database(name string) (sql.Database, error) {
	db, ok := m.databases[name]
	if !ok {
		return nil, sql.ErrDatabaseNotFound.New(name)
	}
	return db, nil
}

func (m *mockCatalog) AllDatabases() []sql.Database {
	var dbs []sql.Database
	for _, db := range m.databases {
		dbs = append(dbs, db)
	}
	return dbs
}

func (m *mockCatalog) AddDatabase(db sql.Database) {
	m.databases[db.Name()] = db
}



// mockEngine implements a simple engine for testing
type mockEngine struct {
	catalog *mockCatalog
}

func newMockEngine() *mockEngine {
	catalog := newMockCatalog()
	return &mockEngine{catalog: catalog}
}

func (m *mockEngine) Query(ctx *sql.Context, query string) (sql.Schema, sql.RowIter, error) {
	// Simple mock implementation for testing
	if query == "SELECT 1+1 AS result" {
		schema := sql.Schema{
			{Name: "result", Type: sql.Int32},
		}
		rows := &mockRowIter{
			rows: []sql.Row{{int32(2)}},
		}
		return schema, rows, nil
	}
	
	// Return empty result for other queries
	return sql.Schema{}, &mockRowIter{}, nil
}



// setupTestHandler creates a test handler with mock dependencies
func setupTestHandler() (*Handler, *mysql.Conn) {
	// Create mock catalog and add test database
	catalog := sql.NewCatalog()
	testDB := newMockDatabase("testdb")
	catalog.AddDatabase(testDB)

	// Create mock engine with required components
	analyzer := analyzer.NewAnalyzer(catalog)
	optimizer := optimizer.NewOptimizer()
	engine := executor.NewEngine(analyzer, optimizer, catalog)

	// Create session manager
	sm := NewSessionManager(DefaultSessionBuilder, nil, "localhost:3306")

	// Create handler
	handler := NewHandler(engine, sm)

	// Create mock connection
	conn := &mysql.Conn{
		ConnectionID: 1,
		User:         "testuser",
	}

	// Note: Cannot set remote address on vitess Conn directly
	// This would be set by the actual network connection

	// Initialize connection
	handler.NewConnection(conn)

	return handler, conn
}

func TestHandler_ComInitDB_Success(t *testing.T) {
	h, conn := setupTestHandler()

	err := h.ComInitDB(conn, "testdb")
	assert.NoError(t, err)

	sess := h.sessionMgr.GetSession(conn.ConnectionID)
	require.NotNil(t, sess)
	assert.Equal(t, "testdb", sess.GetCurrentDB())
}

func TestHandler_ComInitDB_NotFound(t *testing.T) {
	h, conn := setupTestHandler()

	err := h.ComInitDB(conn, "nonexistent")
	assert.Error(t, err)

	sqlErr, ok := err.(*mysql.SQLError)
	require.True(t, ok)
	assert.Equal(t, mysql.ERBadDb, sqlErr.Num)
	assert.Contains(t, sqlErr.Message, "Unknown database 'nonexistent'")
}

func TestHandler_ComQuery_Select(t *testing.T) {
	h, conn := setupTestHandler()

	var result *sqltypes.Result
	err := h.ComQuery(context.Background(), conn, "SELECT 1+1 AS result", func(r *sqltypes.Result, more bool) error {
		result = r
		return nil
	})

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Fields, 1)
	assert.Equal(t, "result", result.Fields[0].Name)
}

func TestHandler_ComMultiQuery_TwoStatements(t *testing.T) {
	h, conn := setupTestHandler()

	results := []*sqltypes.Result{}
	remainder, err := h.ComMultiQuery(context.Background(), conn, "SELECT 1; SELECT 2", func(r *sqltypes.Result, more bool) error {
		results = append(results, r)
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "SELECT 2", remainder)
	assert.Len(t, results, 1) // Only first statement executed
}

func TestHandler_SplitStatements(t *testing.T) {
	h, _ := setupTestHandler()

	tests := []struct {
		name     string
		query    string
		expected []string
	}{
		{
			name:     "single statement",
			query:    "SELECT 1",
			expected: []string{"SELECT 1"},
		},
		{
			name:     "two statements",
			query:    "SELECT 1; SELECT 2",
			expected: []string{"SELECT 1", "SELECT 2"},
		},
		{
			name:     "statements with whitespace",
			query:    "  SELECT 1  ;  SELECT 2  ; ",
			expected: []string{"SELECT 1", "SELECT 2"},
		},
		{
			name:     "empty query",
			query:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.splitStatements(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHandler_NewConnection(t *testing.T) {
	h, _ := setupTestHandler()

	// Create a new connection
	conn := &mysql.Conn{
		ConnectionID: 2,
		User:         "newuser",
	}
	// Note: Cannot set remote address on vitess Conn directly

	h.NewConnection(conn)

	// Check that session was created
	sess := h.sessionMgr.GetSession(conn.ConnectionID)
	require.NotNil(t, sess)
	assert.Equal(t, "newuser", sess.User())
	assert.Equal(t, "unknown", sess.Client())
}

func TestHandler_ConnectionClosed(t *testing.T) {
	h, conn := setupTestHandler()

	// Verify session exists
	sess := h.sessionMgr.GetSession(conn.ConnectionID)
	require.NotNil(t, sess)

	// Close connection
	h.ConnectionClosed(conn)

	// Verify session was removed
	sess = h.sessionMgr.GetSession(conn.ConnectionID)
	assert.Nil(t, sess)
}

func TestSessionManager_NewSession(t *testing.T) {
	sm := NewEnhancedSessionManager()

	sess := sm.NewSession("testuser", "127.0.0.1:12345")
	require.NotNil(t, sess)
	assert.Equal(t, "testuser", sess.User())
	assert.Equal(t, "127.0.0.1:12345", sess.Client())
	assert.Equal(t, uint32(1), sess.ID())

	// Second session should have incremented ID
	sess2 := sm.NewSession("user2", "127.0.0.1:54321")
	assert.Equal(t, uint32(2), sess2.ID())
}

func TestSession_CurrentDB(t *testing.T) {
	sess := NewSession(1, "testuser", "127.0.0.1:12345")

	// Initially no current database
	assert.Equal(t, "", sess.GetCurrentDB())

	// Set current database
	sess.SetCurrentDB("testdb")
	assert.Equal(t, "testdb", sess.GetCurrentDB())

	// Change current database
	sess.SetCurrentDB("otherdb")
	assert.Equal(t, "otherdb", sess.GetCurrentDB())
}

func TestSession_Variables(t *testing.T) {
	sess := NewSession(1, "testuser", "127.0.0.1:12345")

	// Initially no variables
	assert.Nil(t, sess.GetVar("test_var"))

	// Set variable
	sess.SetVar("test_var", "test_value")
	assert.Equal(t, "test_value", sess.GetVar("test_var"))

	// Set different types
	sess.SetVar("int_var", 42)
	sess.SetVar("bool_var", true)

	assert.Equal(t, 42, sess.GetVar("int_var"))
	assert.Equal(t, true, sess.GetVar("bool_var"))
}
