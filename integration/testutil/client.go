package testutil

import (
	"database/sql"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

// TestClient wraps a database connection for testing
type TestClient struct {
	t   *testing.T
	db  *sql.DB
	dsn string
}

// NewTestClient creates a new test client
func NewTestClient(t *testing.T, dsn string) *TestClient {
	t.Helper()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("failed to open connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping server: %v", err)
	}

	return &TestClient{
		t:   t,
		db:  db,
		dsn: dsn,
	}
}

// Close closes the database connection
func (tc *TestClient) Close() {
	if tc.db != nil {
		tc.db.Close()
	}
}

// Exec executes a query and fails the test on error
func (tc *TestClient) Exec(query string, args ...interface{}) sql.Result {
	result, err := tc.db.Exec(query, args...)
	if err != nil {
		tc.t.Fatalf("exec %q failed: %v", query, err)
	}
	return result
}

// Query executes a query and returns rows
func (tc *TestClient) Query(query string, args ...interface{}) *sql.Rows {
	rows, err := tc.db.Query(query, args...)
	if err != nil {
		tc.t.Fatalf("query %q failed: %v", query, err)
	}
	return rows
}

// QueryRow executes a query and returns a single row
func (tc *TestClient) QueryRow(query string, args ...interface{}) *sql.Row {
	return tc.db.QueryRow(query, args...)
}

// MustQueryInt executes a query and returns an integer result
func (tc *TestClient) MustQueryInt(query string) int {
	var val int
	if err := tc.db.QueryRow(query).Scan(&val); err != nil {
		tc.t.Fatalf("query int %q failed: %v", query, err)
	}
	return val
}

// MustQueryString executes a query and returns a string result
func (tc *TestClient) MustQueryString(query string) string {
	var val string
	if err := tc.db.QueryRow(query).Scan(&val); err != nil {
		tc.t.Fatalf("query string %q failed: %v", query, err)
	}
	return val
}

// MustQueryFloat executes a query and returns a float result
func (tc *TestClient) MustQueryFloat(query string) float64 {
	var val float64
	if err := tc.db.QueryRow(query).Scan(&val); err != nil {
		tc.t.Fatalf("query float %q failed: %v", query, err)
	}
	return val
}

// ExpectError executes a query expecting it to fail with a specific error message
func (tc *TestClient) ExpectError(query string, msgContains string) {
	_, err := tc.db.Exec(query)
	if err == nil {
		tc.t.Fatalf("expected error for %q, got nil", query)
	}
	if !strings.Contains(err.Error(), msgContains) {
		tc.t.Fatalf("expected error containing %q, got %v", msgContains, err)
	}
}

// BeginTx starts a transaction
func (tc *TestClient) BeginTx() *TestTx {
	tx, err := tc.db.Begin()
	if err != nil {
		tc.t.Fatalf("begin tx failed: %v", err)
	}
	return &TestTx{t: tc.t, tx: tx}
}

// MustExec executes multiple queries in sequence
func (tc *TestClient) MustExec(queries ...string) {
	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q != "" && !strings.HasPrefix(q, "--") {
			tc.Exec(q)
		}
	}
}

// TestTx wraps a transaction for testing
type TestTx struct {
	t  *testing.T
	tx *sql.Tx
}

// Exec executes a query within the transaction
func (tt *TestTx) Exec(query string, args ...interface{}) sql.Result {
	result, err := tt.tx.Exec(query, args...)
	if err != nil {
		tt.t.Fatalf("tx exec %q failed: %v", query, err)
	}
	return result
}

// Query executes a query within the transaction
func (tt *TestTx) Query(query string, args ...interface{}) *sql.Rows {
	rows, err := tt.tx.Query(query, args...)
	if err != nil {
		tt.t.Fatalf("tx query %q failed: %v", query, err)
	}
	return rows
}

// Commit commits the transaction
func (tt *TestTx) Commit() {
	if err := tt.tx.Commit(); err != nil {
		tt.t.Fatalf("commit failed: %v", err)
	}
}

// Rollback rolls back the transaction
func (tt *TestTx) Rollback() {
	if err := tt.tx.Rollback(); err != nil {
		tt.t.Fatalf("rollback failed: %v", err)
	}
}

// TryCommit attempts to commit and returns any error
func (tt *TestTx) TryCommit() error {
	return tt.tx.Commit()
}

// TryRollback attempts to rollback and returns any error
func (tt *TestTx) TryRollback() error {
	return tt.tx.Rollback()
}
