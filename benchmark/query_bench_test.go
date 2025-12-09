package benchmark

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/turtacn/guocedb/integration/testutil"
)

var (
	benchServer *testutil.TestServer
	benchDB     *sql.DB
)

// setupBenchmark initializes the benchmark environment
func setupBenchmark() {
	if benchServer == nil {
		// Use a mock testing.T for setup
		t := &testing.T{}
		benchServer = testutil.NewTestServer(t).Start()

		var err error
		benchDB, err = sql.Open("mysql", benchServer.DSN())
		if err != nil {
			panic(fmt.Sprintf("failed to open benchmark DB: %v", err))
		}

		// Setup benchmark schema
		benchDB.Exec("CREATE DATABASE IF NOT EXISTS bench")
		benchDB.Exec("USE bench")
		benchDB.Exec(`CREATE TABLE IF NOT EXISTS users (
			id INT PRIMARY KEY,
			name VARCHAR(100),
			email VARCHAR(200),
			age INT,
			INDEX idx_age (age)
		)`)

		// Insert test data
		tx, _ := benchDB.Begin()
		stmt, _ := tx.Prepare("INSERT INTO users VALUES (?, ?, ?, ?)")
		for i := 0; i < 10000; i++ {
			stmt.Exec(i, fmt.Sprintf("user%d", i), fmt.Sprintf("user%d@test.com", i), i%100)
		}
		tx.Commit()
	}
}

// BenchmarkSelect_PointQuery benchmarks single-row point queries
func BenchmarkSelect_PointQuery(b *testing.B) {
	setupBenchmark()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var name string
		benchDB.QueryRow("SELECT name FROM bench.users WHERE id = ?", i%10000).Scan(&name)
	}
}

// BenchmarkSelect_RangeScan benchmarks range scan queries
func BenchmarkSelect_RangeScan(b *testing.B) {
	setupBenchmark()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, _ := benchDB.Query("SELECT * FROM bench.users WHERE age BETWEEN ? AND ?", 20, 30)
		for rows.Next() {
			// Consume rows
		}
		rows.Close()
	}
}

// BenchmarkSelect_Count benchmarks COUNT queries
func BenchmarkSelect_Count(b *testing.B) {
	setupBenchmark()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var count int
		benchDB.QueryRow("SELECT COUNT(*) FROM bench.users").Scan(&count)
	}
}

// BenchmarkSelect_OrderBy benchmarks ORDER BY queries
func BenchmarkSelect_OrderBy(b *testing.B) {
	setupBenchmark()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, _ := benchDB.Query("SELECT * FROM bench.users ORDER BY age LIMIT 100")
		for rows.Next() {
			// Consume rows
		}
		rows.Close()
	}
}

// BenchmarkSelect_Aggregate benchmarks aggregate queries
func BenchmarkSelect_Aggregate(b *testing.B) {
	setupBenchmark()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var avg float64
		benchDB.QueryRow("SELECT AVG(age) FROM bench.users").Scan(&avg)
	}
}

// BenchmarkInsert_Single benchmarks single row inserts
func BenchmarkInsert_Single(b *testing.B) {
	setupBenchmark()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchDB.Exec("INSERT INTO bench.users VALUES (?, ?, ?, ?)",
			100000+i, "benchuser", "bench@test.com", 25)
	}
}

// BenchmarkInsert_Batch benchmarks batch inserts
func BenchmarkInsert_Batch(b *testing.B) {
	setupBenchmark()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx, _ := benchDB.Begin()
		stmt, _ := tx.Prepare("INSERT INTO bench.users VALUES (?, ?, ?, ?)")
		for j := 0; j < 100; j++ {
			stmt.Exec(200000+i*100+j, "batchuser", "batch@test.com", 30)
		}
		tx.Commit()
	}
}

// BenchmarkUpdate_PointUpdate benchmarks single row updates
func BenchmarkUpdate_PointUpdate(b *testing.B) {
	setupBenchmark()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchDB.Exec("UPDATE bench.users SET age = ? WHERE id = ?",
			(i%100)+1, i%10000)
	}
}

// BenchmarkUpdate_RangeUpdate benchmarks range updates
func BenchmarkUpdate_RangeUpdate(b *testing.B) {
	setupBenchmark()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchDB.Exec("UPDATE bench.users SET age = age + 1 WHERE age BETWEEN 20 AND 30")
	}
}

// BenchmarkDelete_Single benchmarks single row deletes
func BenchmarkDelete_Single(b *testing.B) {
	setupBenchmark()

	// Pre-insert rows to delete
	for i := 0; i < b.N; i++ {
		benchDB.Exec("INSERT INTO bench.users VALUES (?, ?, ?, ?)",
			300000+i, "deluser", "del@test.com", 40)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchDB.Exec("DELETE FROM bench.users WHERE id = ?", 300000+i)
	}
}

// BenchmarkParallel_PointQuery benchmarks parallel point queries
func BenchmarkParallel_PointQuery(b *testing.B) {
	setupBenchmark()

	b.RunParallel(func(pb *testing.PB) {
		id := 0
		for pb.Next() {
			var name string
			benchDB.QueryRow("SELECT name FROM bench.users WHERE id = ?", id%10000).Scan(&name)
			id++
		}
	})
}

// BenchmarkParallel_MixedWorkload benchmarks parallel mixed read/write workload
func BenchmarkParallel_MixedWorkload(b *testing.B) {
	setupBenchmark()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%10 < 2 { // 20% writes
				benchDB.Exec("UPDATE bench.users SET age = age + 1 WHERE id = ?", i%1000)
			} else { // 80% reads
				var name string
				benchDB.QueryRow("SELECT name FROM bench.users WHERE id = ?", i%10000).Scan(&name)
			}
			i++
		}
	})
}

// TestMain handles benchmark setup and teardown
func TestMain(m *testing.M) {
	// Run benchmarks
	code := m.Run()

	// Cleanup
	if benchDB != nil {
		benchDB.Close()
	}
	if benchServer != nil {
		benchServer.Stop()
	}

	// Exit with the test result code
	// Note: We can't call os.Exit here as it would skip deferred cleanup
	// The test framework will handle the exit
	_ = code
}
