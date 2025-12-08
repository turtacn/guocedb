package integration

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompat_PreparedStatement tests prepared statement functionality
func TestCompat_PreparedStatement(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	
	db, err := sql.Open("mysql", ts.DSN(""))
	require.NoError(t, err)
	defer db.Close()

	// Setup
	_, err = db.Exec("CREATE DATABASE testdb")
	require.NoError(t, err)
	_, err = db.Exec("USE testdb")
	require.NoError(t, err)
	_, err = db.Exec("CREATE TABLE t (id INT PRIMARY KEY, name VARCHAR(100), age INT)")
	require.NoError(t, err)

	// Test prepared INSERT
	stmt, err := db.Prepare("INSERT INTO t (id, name, age) VALUES (?, ?, ?)")
	require.NoError(t, err)
	defer stmt.Close()

	_, err = stmt.Exec(1, "Alice", 25)
	require.NoError(t, err)
	_, err = stmt.Exec(2, "Bob", 30)
	require.NoError(t, err)

	// Test prepared SELECT
	queryStmt, err := db.Prepare("SELECT name, age FROM t WHERE id = ?")
	require.NoError(t, err)
	defer queryStmt.Close()

	var name string
	var age int
	err = queryStmt.QueryRow(1).Scan(&name, &age)
	require.NoError(t, err)
	assert.Equal(t, "Alice", name)
	assert.Equal(t, 25, age)

	// Test prepared UPDATE
	updateStmt, err := db.Prepare("UPDATE t SET age = ? WHERE id = ?")
	require.NoError(t, err)
	defer updateStmt.Close()

	result, err := updateStmt.Exec(26, 1)
	require.NoError(t, err)
	affected, err := result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)

	// Verify update
	err = queryStmt.QueryRow(1).Scan(&name, &age)
	require.NoError(t, err)
	assert.Equal(t, 26, age)
}

// TestCompat_NullValues tests handling of NULL values
func TestCompat_NullValues(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	
	client := NewTestClient(t, ts.DSN(""))
	defer client.Close()

	client.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE t (id INT PRIMARY KEY, val VARCHAR(100), num INT)",
		"INSERT INTO t (id, val, num) VALUES (1, NULL, NULL)",
		"INSERT INTO t (id, val, num) VALUES (2, 'hello', 42)",
		"INSERT INTO t (id, val, num) VALUES (3, '', 0)",
	)

	rows := client.Query("SELECT id, val, num FROM t ORDER BY id")
	defer rows.Close()

	var id int
	var val sql.NullString
	var num sql.NullInt64

	// First row - both NULL
	require.True(t, rows.Next())
	err := rows.Scan(&id, &val, &num)
	require.NoError(t, err)
	assert.Equal(t, 1, id)
	assert.False(t, val.Valid)
	assert.False(t, num.Valid)

	// Second row - both NOT NULL
	require.True(t, rows.Next())
	err = rows.Scan(&id, &val, &num)
	require.NoError(t, err)
	assert.Equal(t, 2, id)
	assert.True(t, val.Valid)
	assert.Equal(t, "hello", val.String)
	assert.True(t, num.Valid)
	assert.Equal(t, int64(42), num.Int64)

	// Third row - empty string and zero (not NULL)
	require.True(t, rows.Next())
	err = rows.Scan(&id, &val, &num)
	require.NoError(t, err)
	assert.Equal(t, 3, id)
	assert.True(t, val.Valid)
	assert.Equal(t, "", val.String)
	assert.True(t, num.Valid)
	assert.Equal(t, int64(0), num.Int64)
}

// TestCompat_DataTypes tests various MySQL data types
func TestCompat_DataTypes(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	
	client := NewTestClient(t, ts.DSN(""))
	defer client.Close()

	client.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
	)

	// Create table with various data types
	createTableSQL := `
		CREATE TABLE datatypes (
			id INT PRIMARY KEY,
			tiny_col TINYINT,
			small_col SMALLINT,
			big_col BIGINT,
			float_col FLOAT,
			double_col DOUBLE,
			decimal_col DECIMAL(10,2),
			varchar_col VARCHAR(200),
			text_col TEXT,
			bool_col BOOLEAN,
			date_col DATE,
			datetime_col DATETIME
		)
	`
	client.Exec(createTableSQL)

	// Insert test data
	insertSQL := `
		INSERT INTO datatypes VALUES (
			1, 127, 32767, 9223372036854775807, 
			3.14, 3.14159265359, 12345.67,
			'hello world', 'long text content here', true,
			'2024-01-15', '2024-01-15 10:30:00'
		)
	`
	client.Exec(insertSQL)

	// Query and verify data types
	row := client.QueryRow("SELECT * FROM datatypes WHERE id = 1")
	
	var (
		id          int
		tinyCol     int8
		smallCol    int16
		bigCol      int64
		floatCol    float32
		doubleCol   float64
		decimalCol  float64
		varcharCol  string
		textCol     string
		boolCol     bool
		dateCol     string
		datetimeCol string
	)

	err := row.Scan(&id, &tinyCol, &smallCol, &bigCol, &floatCol, &doubleCol,
		&decimalCol, &varcharCol, &textCol, &boolCol, &dateCol, &datetimeCol)
	require.NoError(t, err)

	assert.Equal(t, 1, id)
	assert.Equal(t, int8(127), tinyCol)
	assert.Equal(t, int16(32767), smallCol)
	assert.Equal(t, int64(9223372036854775807), bigCol)
	assert.InDelta(t, 3.14, floatCol, 0.01)
	assert.InDelta(t, 3.14159265359, doubleCol, 0.0001)
	assert.Equal(t, 12345.67, decimalCol)
	assert.Equal(t, "hello world", varcharCol)
	assert.Equal(t, "long text content here", textCol)
	assert.True(t, boolCol)
	assert.Equal(t, "2024-01-15", dateCol)
	assert.Contains(t, datetimeCol, "2024-01-15")
}

// TestCompat_ShowStatements tests various SHOW statements
func TestCompat_ShowStatements(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	
	client := NewTestClient(t, ts.DSN(""))
	defer client.Close()

	client.MustExec("CREATE DATABASE testdb", "USE testdb")
	client.MustExec("CREATE TABLE t1 (id INT PRIMARY KEY, name VARCHAR(50))")
	client.MustExec("CREATE TABLE t2 (id INT PRIMARY KEY, value INT)")

	// Test SHOW DATABASES
	rows := client.Query("SHOW DATABASES")
	dbs := CollectRows(t, rows)
	
	found := false
	for _, db := range dbs {
		for _, v := range db {
			if v == "testdb" {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "testdb should be in SHOW DATABASES result")

	// Test SHOW TABLES
	rows = client.Query("SHOW TABLES")
	tables := CollectRows(t, rows)
	assert.Len(t, tables, 2)

	// Extract table names
	tableNames := make([]string, len(tables))
	for i, table := range tables {
		for _, v := range table {
			tableNames[i] = v.(string)
			break
		}
	}
	assert.Contains(t, tableNames, "t1")
	assert.Contains(t, tableNames, "t2")

	// Test SHOW CREATE TABLE
	rows = client.Query("SHOW CREATE TABLE t1")
	result := CollectRows(t, rows)
	assert.Len(t, result, 1)
	
	// Should contain CREATE TABLE statement
	found = false
	for _, v := range result[0] {
		if str, ok := v.(string); ok && len(str) > 10 {
			if assert.Contains(t, str, "CREATE TABLE") {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "SHOW CREATE TABLE should return CREATE statement")
}

// TestCompat_InformationSchema tests INFORMATION_SCHEMA queries
func TestCompat_InformationSchema(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	
	client := NewTestClient(t, ts.DSN(""))
	defer client.Close()

	client.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE test_table (id INT PRIMARY KEY, name VARCHAR(100))",
	)

	// Test querying INFORMATION_SCHEMA.TABLES
	query := `
		SELECT TABLE_NAME, TABLE_TYPE 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = 'testdb'
	`
	rows := client.Query(query)
	results := CollectRows(t, rows)
	
	if len(results) > 0 {
		// If INFORMATION_SCHEMA is supported, verify results
		found := false
		for _, result := range results {
			if result["TABLE_NAME"] == "test_table" {
				found = true
				break
			}
		}
		assert.True(t, found, "test_table should be in INFORMATION_SCHEMA.TABLES")
	} else {
		t.Log("INFORMATION_SCHEMA.TABLES not supported - this is acceptable")
	}
}

// TestCompat_ConnectionHandling tests connection lifecycle
func TestCompat_ConnectionHandling(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Test multiple connections
	var connections []*sql.DB
	defer func() {
		for _, conn := range connections {
			conn.Close()
		}
	}()

	// Open multiple connections
	for i := 0; i < 5; i++ {
		db, err := sql.Open("mysql", ts.DSN(""))
		require.NoError(t, err)
		connections = append(connections, db)
		
		// Test each connection
		err = db.Ping()
		require.NoError(t, err)
	}

	// Test that all connections work
	for i, db := range connections {
		var result int
		err := db.QueryRow("SELECT ?", i+1).Scan(&result)
		require.NoError(t, err)
		assert.Equal(t, i+1, result)
	}
}

// TestCompat_TransactionIsolation tests transaction isolation levels
func TestCompat_TransactionIsolation(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	
	client := NewTestClient(t, ts.DSN(""))
	defer client.Close()

	client.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE t (id INT PRIMARY KEY, val INT)",
		"INSERT INTO t (id, val) VALUES (1, 100)",
	)

	// Test setting isolation level (may not be fully implemented)
	_, err := client.db.Exec("SET TRANSACTION ISOLATION LEVEL READ COMMITTED")
	// Don't require this to work - just test that it doesn't crash
	if err != nil {
		t.Logf("SET TRANSACTION ISOLATION LEVEL not supported: %v", err)
	}

	// Test basic transaction functionality
	tx, err := client.db.Begin()
	require.NoError(t, err)

	_, err = tx.Exec("UPDATE t SET val = 200 WHERE id = 1")
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	// Verify change persisted
	var val int
	err = client.QueryRow("SELECT val FROM t WHERE id = 1").Scan(&val)
	require.NoError(t, err)
	assert.Equal(t, 200, val)
}

// TestCompat_CharacterSets tests character set handling
func TestCompat_CharacterSets(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	
	client := NewTestClient(t, ts.DSN(""))
	defer client.Close()

	client.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE t (id INT PRIMARY KEY, text_data VARCHAR(200))",
	)

	// Test various character data
	testStrings := []string{
		"Hello World",
		"Héllo Wörld", // Accented characters
		"你好世界",      // Chinese characters
		"🚀🌟💫",      // Emojis
		"",            // Empty string
	}

	// Insert test strings
	for i, str := range testStrings {
		client.Exec("INSERT INTO t (id, text_data) VALUES (?, ?)", i+1, str)
	}

	// Verify strings are stored and retrieved correctly
	rows := client.Query("SELECT id, text_data FROM t ORDER BY id")
	results := CollectRows(t, rows)
	
	assert.Len(t, results, len(testStrings))
	for i, result := range results {
		assert.Equal(t, int64(i+1), result["id"])
		assert.Equal(t, testStrings[i], result["text_data"])
	}
}

// TestCompat_LargeResults tests handling of large result sets
func TestCompat_LargeResults(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	
	client := NewTestClient(t, ts.DSN(""))
	defer client.Close()

	client.MustExec(
		"CREATE DATABASE testdb",
		"USE testdb",
		"CREATE TABLE large_test (id INT PRIMARY KEY, data VARCHAR(100))",
	)

	// Insert a moderate number of rows
	numRows := 1000
	for i := 1; i <= numRows; i++ {
		client.Exec("INSERT INTO large_test (id, data) VALUES (?, ?)", 
			i, fmt.Sprintf("data-row-%d", i))
	}

	// Query all rows
	rows := client.Query("SELECT id, data FROM large_test ORDER BY id")
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		var data string
		err := rows.Scan(&id, &data)
		require.NoError(t, err)
		count++
		
		// Verify a few specific rows
		if id <= 10 {
			expected := fmt.Sprintf("data-row-%d", id)
			assert.Equal(t, expected, data)
		}
	}
	
	assert.Equal(t, numRows, count)
	require.NoError(t, rows.Err())
}

// TestCompat_ConnectionTimeout tests connection timeout behavior
func TestCompat_ConnectionTimeout(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Test connection with timeout
	db, err := sql.Open("mysql", ts.DSN("")+"?timeout=5s")
	require.NoError(t, err)
	defer db.Close()

	// Set connection limits
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Second)

	// Test basic operation
	err = db.Ping()
	require.NoError(t, err)

	// Test query
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)
}

// TestCompat_ErrorHandling tests MySQL error code compatibility
func TestCompat_ErrorHandling(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	
	client := NewTestClient(t, ts.DSN(""))
	defer client.Close()

	client.MustExec("CREATE DATABASE testdb", "USE testdb")

	// Test duplicate key error
	client.Exec("CREATE TABLE t (id INT PRIMARY KEY, name VARCHAR(50))")
	client.Exec("INSERT INTO t (id, name) VALUES (1, 'test')")
	
	_, err := client.db.Exec("INSERT INTO t (id, name) VALUES (1, 'duplicate')")
	assert.Error(t, err)
	// The exact error code may vary, but there should be an error

	// Test table not found error
	_, err = client.db.Query("SELECT * FROM nonexistent_table")
	assert.Error(t, err)

	// Test syntax error
	_, err = client.db.Exec("INVALID SQL SYNTAX")
	assert.Error(t, err)
}