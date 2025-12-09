package testutil

import (
	"database/sql"
	"os"
	"strings"
	"testing"
)

// ColumnInfo represents column metadata
type ColumnInfo struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default sql.NullString
	Extra   string
}

// ScanStringSlice scans all rows into a string slice
func ScanStringSlice(rows *sql.Rows) []string {
	defer rows.Close()
	var result []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			continue
		}
		result = append(result, s)
	}
	return result
}

// ScanDescribeResult scans DESCRIBE table output
func ScanDescribeResult(rows *sql.Rows) []ColumnInfo {
	defer rows.Close()
	var result []ColumnInfo
	for rows.Next() {
		var c ColumnInfo
		if err := rows.Scan(&c.Field, &c.Type, &c.Null, &c.Key, &c.Default, &c.Extra); err != nil {
			continue
		}
		result = append(result, c)
	}
	return result
}

// SetupTestTable creates a standard test table with sample data
func SetupTestTable(client *TestClient) {
	client.MustExec(
		"CREATE DATABASE IF NOT EXISTS testdb",
		"USE testdb",
		"DROP TABLE IF EXISTS products",
		`CREATE TABLE products (
			id INT PRIMARY KEY,
			name VARCHAR(100),
			price DECIMAL(10,2)
		)`,
		"INSERT INTO products VALUES (1, 'Apple', 1.50)",
		"INSERT INTO products VALUES (2, 'Banana', 0.75)",
		"INSERT INTO products VALUES (3, 'Orange', 2.00)",
		"INSERT INTO products VALUES (4, 'Grape', 3.50)",
	)
}

// SetupBankAccounts creates a bank schema for testing transactions
func SetupBankAccounts(client *TestClient) {
	client.MustExec(
		"CREATE DATABASE IF NOT EXISTS bank",
		"USE bank",
		"DROP TABLE IF EXISTS accounts",
		"CREATE TABLE accounts (id INT PRIMARY KEY, balance DECIMAL(10,2))",
		"INSERT INTO accounts VALUES (1, 1000.00), (2, 500.00)",
	)
}

// SetupOrdersSchema creates a schema with users and orders for JOIN testing
func SetupOrdersSchema(client *TestClient) {
	client.MustExec(
		"CREATE DATABASE IF NOT EXISTS shop",
		"USE shop",
		"DROP TABLE IF EXISTS orders",
		"DROP TABLE IF EXISTS users",
		`CREATE TABLE users (
			id INT PRIMARY KEY,
			name VARCHAR(100)
		)`,
		`CREATE TABLE orders (
			id INT PRIMARY KEY,
			user_id INT,
			amount DECIMAL(10,2)
		)`,
		"INSERT INTO users VALUES (1, 'Alice'), (2, 'Bob')",
		"INSERT INTO orders VALUES (1, 1, 100.00), (2, 1, 200.00), (3, 2, 150.00)",
	)
}

// WriteTempConfig writes configuration to a temporary file
func WriteTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "guocedb-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp config: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

// ExecSQLFile executes SQL statements from a file
func ExecSQLFile(t *testing.T, client *TestClient, path string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read SQL file %s: %v", path, err)
	}

	statements := strings.Split(string(content), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" && !strings.HasPrefix(stmt, "--") {
			client.Exec(stmt)
		}
	}
}

// CollectRows collects all rows from a result set into a slice of maps
func CollectRows(t *testing.T, rows *sql.Rows) []map[string]interface{} {
	t.Helper()
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		t.Fatalf("failed to get columns: %v", err)
	}

	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}

		row := make(map[string]interface{})
		for i, col := range cols {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return results
}

// CountRows returns the number of rows in a result set
func CountRows(rows *sql.Rows) int {
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
	}
	return count
}
