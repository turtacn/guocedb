package benchmark

import (
	"testing"
)

// BenchmarkTransaction_SimpleCommit benchmarks simple transaction commit
func BenchmarkTransaction_SimpleCommit(b *testing.B) {
	setupBenchmark()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx, _ := benchDB.Begin()
		tx.Exec("SELECT 1")
		tx.Commit()
	}
}

// BenchmarkTransaction_MultiStatement benchmarks multi-statement transactions
func BenchmarkTransaction_MultiStatement(b *testing.B) {
	setupBenchmark()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx, _ := benchDB.Begin()
		tx.Exec("INSERT INTO bench.users VALUES (?, ?, ?, ?)",
			400000+i, "txnuser", "txn@test.com", 40)
		tx.Exec("UPDATE bench.users SET age = 41 WHERE id = ?", 400000+i)
		tx.Exec("SELECT * FROM bench.users WHERE id = ?", 400000+i)
		tx.Commit()
	}
}

// BenchmarkTransaction_Rollback benchmarks transaction rollback
func BenchmarkTransaction_Rollback(b *testing.B) {
	setupBenchmark()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx, _ := benchDB.Begin()
		tx.Exec("INSERT INTO bench.users VALUES (?, ?, ?, ?)",
			500000+i, "rollbackuser", "rb@test.com", 50)
		tx.Rollback()
	}
}

// BenchmarkTransaction_ReadOnly benchmarks read-only transactions
func BenchmarkTransaction_ReadOnly(b *testing.B) {
	setupBenchmark()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx, _ := benchDB.Begin()
		tx.Query("SELECT * FROM bench.users WHERE id = ?", i%1000)
		tx.Query("SELECT COUNT(*) FROM bench.users WHERE age > ?", i%100)
		tx.Commit()
	}
}

// BenchmarkTransaction_UpdateCommit benchmarks update transactions
func BenchmarkTransaction_UpdateCommit(b *testing.B) {
	setupBenchmark()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx, _ := benchDB.Begin()
		tx.Exec("UPDATE bench.users SET age = ? WHERE id = ?", (i%100)+1, i%1000)
		tx.Commit()
	}
}

// BenchmarkTransaction_ParallelCommit benchmarks parallel transaction commits
func BenchmarkTransaction_ParallelCommit(b *testing.B) {
	setupBenchmark()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			tx, _ := benchDB.Begin()
			tx.Exec("SELECT * FROM bench.users WHERE id = ?", i%1000)
			tx.Exec("UPDATE bench.users SET age = age + 1 WHERE id = ?", i%1000)
			tx.Commit()
			i++
		}
	})
}

// BenchmarkTransaction_LongRunning benchmarks long-running transactions
func BenchmarkTransaction_LongRunning(b *testing.B) {
	setupBenchmark()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx, _ := benchDB.Begin()
		for j := 0; j < 50; j++ {
			tx.Query("SELECT * FROM bench.users WHERE id = ?", (i*50+j)%10000)
		}
		tx.Commit()
	}
}

// BenchmarkTransaction_SmallBatch benchmarks small batch transactions
func BenchmarkTransaction_SmallBatch(b *testing.B) {
	setupBenchmark()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx, _ := benchDB.Begin()
		stmt, _ := tx.Prepare("INSERT INTO bench.users VALUES (?, ?, ?, ?)")
		for j := 0; j < 10; j++ {
			stmt.Exec(600000+i*10+j, "batchuser", "batch@test.com", 60)
		}
		tx.Commit()
	}
}

// BenchmarkTransaction_MediumBatch benchmarks medium batch transactions
func BenchmarkTransaction_MediumBatch(b *testing.B) {
	setupBenchmark()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx, _ := benchDB.Begin()
		stmt, _ := tx.Prepare("INSERT INTO bench.users VALUES (?, ?, ?, ?)")
		for j := 0; j < 100; j++ {
			stmt.Exec(700000+i*100+j, "meduser", "med@test.com", 70)
		}
		tx.Commit()
	}
}

// BenchmarkTransaction_ConflictRetry benchmarks transaction with potential conflicts
func BenchmarkTransaction_ConflictRetry(b *testing.B) {
	setupBenchmark()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Try to update the same row from multiple goroutines
			tx, _ := benchDB.Begin()
			tx.Exec("UPDATE bench.users SET age = age + 1 WHERE id = 1")
			tx.Commit() // May fail due to conflict
		}
	})
}
