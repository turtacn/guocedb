package integration

import (
	"testing"
	"github.com/stretchr/testify/require"

	"github.com/turtacn/guocedb/integration/testutil"
)

// TestE2E_InsertSelect tests INSERT and SELECT operations
func TestE2E_InsertSelect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	testutil.SetupTestTable(client)
	client.Exec("USE testdb")

	// Query all rows
	rows := client.Query("SELECT * FROM products ORDER BY id")
	results := testutil.CollectRows(t, rows)
	require.GreaterOrEqual(t, len(results), 4)

	// Query with WHERE clause
	count := client.MustQueryInt("SELECT COUNT(*) FROM products WHERE price > 1.00")
	require.GreaterOrEqual(t, count, 2)

	// Query with specific columns
	name := client.MustQueryString("SELECT name FROM products WHERE id = 1")
	require.NotEmpty(t, name)
}

// TestE2E_Update tests UPDATE operations
func TestE2E_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	testutil.SetupTestTable(client)
	client.Exec("USE testdb")

	// Update single row
	result := client.Exec("UPDATE products SET price = 2.00 WHERE id = 1")
	affected, _ := result.RowsAffected()
	require.Equal(t, int64(1), affected)

	// Verify update
	price := client.MustQueryFloat("SELECT price FROM products WHERE id = 1")
	require.InDelta(t, 2.00, price, 0.01)

	// Update multiple rows
	client.Exec("UPDATE products SET price = price * 1.1 WHERE price < 3.00")

	// Update with no matching rows should succeed
	result = client.Exec("UPDATE products SET price = 999 WHERE id = 9999")
	affected, _ = result.RowsAffected()
	require.Equal(t, int64(0), affected)
}

// TestE2E_Delete tests DELETE operations
func TestE2E_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	testutil.SetupTestTable(client)
	client.Exec("USE testdb")

	// Count before delete
	countBefore := client.MustQueryInt("SELECT COUNT(*) FROM products")

	// Delete single row
	result := client.Exec("DELETE FROM products WHERE id = 1")
	affected, _ := result.RowsAffected()
	require.Equal(t, int64(1), affected)

	// Verify delete
	countAfter := client.MustQueryInt("SELECT COUNT(*) FROM products")
	require.Equal(t, countBefore-1, countAfter)

	// Delete with condition
	client.Exec("DELETE FROM products WHERE price < 1.00")

	// Delete non-existent row should succeed
	result = client.Exec("DELETE FROM products WHERE id = 9999")
	affected, _ = result.RowsAffected()
	require.Equal(t, int64(0), affected)
}

// TestE2E_BatchInsert tests batch INSERT operations
func TestE2E_BatchInsert(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	client.Exec("CREATE DATABASE testdb")
	client.Exec("USE testdb")
	client.Exec("CREATE TABLE items (id INT PRIMARY KEY, name VARCHAR(100))")

	// Batch insert
	client.Exec("INSERT INTO items VALUES (1, 'item1'), (2, 'item2'), (3, 'item3')")

	// Verify count
	count := client.MustQueryInt("SELECT COUNT(*) FROM items")
	require.Equal(t, 3, count)
}

// TestE2E_SelectJoin tests JOIN operations
func TestE2E_SelectJoin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	testutil.SetupOrdersSchema(client)
	client.Exec("USE shop")

	// Inner join
	rows := client.Query(`
		SELECT u.name, o.amount
		FROM users u
		INNER JOIN orders o ON u.id = o.user_id
		ORDER BY o.amount
	`)
	results := testutil.CollectRows(t, rows)
	require.GreaterOrEqual(t, len(results), 2)
}

// TestE2E_SelectAggregate tests aggregate functions
func TestE2E_SelectAggregate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	testutil.SetupTestTable(client)
	client.Exec("USE testdb")

	// COUNT
	count := client.MustQueryInt("SELECT COUNT(*) FROM products")
	require.GreaterOrEqual(t, count, 4)

	// SUM
	var total float64
	err := client.QueryRow("SELECT SUM(price) FROM products").Scan(&total)
	require.NoError(t, err)
	require.Greater(t, total, 0.0)

	// AVG
	var avg float64
	err = client.QueryRow("SELECT AVG(price) FROM products").Scan(&avg)
	require.NoError(t, err)
	require.Greater(t, avg, 0.0)

	// MIN/MAX
	var minPrice, maxPrice float64
	err = client.QueryRow("SELECT MIN(price), MAX(price) FROM products").Scan(&minPrice, &maxPrice)
	require.NoError(t, err)
	require.Less(t, minPrice, maxPrice)
}

// TestE2E_SelectOrderByLimit tests ORDER BY and LIMIT
func TestE2E_SelectOrderByLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ts := testutil.NewTestServer(t).Start()
	defer ts.Stop()

	client := testutil.NewTestClient(t, ts.DSN())
	defer client.Close()

	testutil.SetupTestTable(client)
	client.Exec("USE testdb")

	// ORDER BY ASC
	rows := client.Query("SELECT name FROM products ORDER BY price ASC LIMIT 2")
	results := testutil.CollectRows(t, rows)
	require.Len(t, results, 2)

	// ORDER BY DESC
	rows = client.Query("SELECT name FROM products ORDER BY price DESC LIMIT 1")
	results = testutil.CollectRows(t, rows)
	require.Len(t, results, 1)
}
