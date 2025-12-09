# Phase 4: Transaction Management - ACID Implementation

## Overview

This document describes the complete ACID transaction implementation for GuoceDB, built on top of BadgerDB's transactional capabilities.

## Architecture

### Components

1. **Transaction Interface** (`compute/transaction/transaction.go`)
   - Core transaction abstraction
   - Wraps BadgerDB transactions
   - Provides ACID guarantees

2. **Transaction Manager** (`compute/transaction/manager.go`)
   - Manages transaction lifecycle
   - Tracks active transactions
   - Handles conflict detection

3. **Plan Nodes** (`compute/sql/plan/{begin,commit,rollback}.go`)
   - BEGIN: Start new transaction
   - COMMIT: Finalize transaction changes
   - ROLLBACK: Discard transaction changes

4. **Handler Integration** (`compute/server/handler.go`)
   - Intercepts transaction SQL statements
   - Manages per-connection transaction state
   - Coordinates with TransactionManager

## ACID Properties

### Atomicity
- All operations within a transaction either complete successfully or are completely rolled back
- Implemented via BadgerDB's transaction commit/discard mechanisms
- Tested in: `TestManager_RollbackDiscards`, `TestTxn_RollbackMultipleOperations`

### Consistency
- Database constraints are maintained across transactions
- Transaction conflicts are detected during commit
- Tested in: `TestConcurrentWriteConflict`, `TestWriteAfterReadConflict`

### Isolation
- **Snapshot Isolation**: Each transaction sees a consistent snapshot of the database at transaction start
- Uncommitted changes from other transactions are not visible (prevents dirty reads)
- Multiple reads within a transaction return consistent results (repeatable read)
- Implemented via BadgerDB's MVCC (Multi-Version Concurrency Control)
- Tested in: `TestSnapshotIsolation`, `TestDirtyReadPrevention`, `TestRepeatableRead`

### Durability
- Committed transactions are persisted to disk
- Data survives process restarts
- Implemented via BadgerDB's WAL (Write-Ahead Logging)
- Tested in: `TestManager_CommitPersists`

## Transaction Lifecycle

```
┌─────────┐
│ Client  │
└────┬────┘
     │
     ├─── BEGIN ────────────────────────┐
     │                                   │
     │    ┌──────────────────────┐     │
     │    │ TransactionManager   │     │
     │    │  - Create Transaction│◄────┘
     │    │  - Register in Map   │
     │    └──────────────────────┘
     │
     ├─── SQL Operations ───────────────┐
     │                                   │
     │    ┌──────────────────────┐     │
     │    │ Transaction          │     │
     │    │  - Get/Set/Delete    │◄────┘
     │    │  - Snapshot Reads    │
     │    └──────────────────────┘
     │
     ├─── COMMIT/ROLLBACK ──────────────┐
     │                                   │
     │    ┌──────────────────────┐     │
     │    │ TransactionManager   │     │
     │    │  - Finalize Txn      │◄────┘
     │    │  - Remove from Map   │
     │    └──────────────────────┘
     │
     ▼
```

## Concurrency Control

### Write Conflicts
- BadgerDB detects write-write conflicts at commit time
- Returns `ErrConflict` which is mapped to `ErrTransactionConflict`
- Clients must retry failed transactions

### Read-Write Conflicts
- Snapshot isolation prevents most read-write conflicts
- Transactions operate on snapshot at transaction start
- Writes from concurrent transactions don't affect snapshot

### Concurrent Readers
- Multiple read-only transactions can run concurrently
- No locking required for readers
- Excellent read scalability

## API

### Transaction Manager

```go
// Create manager
mgr := transaction.NewManagerWithDB(badgerDB)

// Begin transaction
txn, err := mgr.Begin(&TransactionOptions{
    IsolationLevel: LevelReadCommitted,
    ReadOnly:       false,
})

// Commit or rollback
err = mgr.Commit(txn)
err = mgr.Rollback(txn)
```

### Transaction Operations

```go
// Read
value, err := txn.Get(key)

// Write
err = txn.Set(key, value)

// Delete
err = txn.Delete(key)

// Iterator
iter, err := txn.Iterator(prefix)
```

### SQL Interface

```sql
-- Start transaction
BEGIN;

-- Execute operations
INSERT INTO users (id, name) VALUES (1, 'Alice');
UPDATE accounts SET balance = balance - 100 WHERE id = 1;

-- Finalize
COMMIT;  -- or ROLLBACK;
```

## Error Handling

### Transaction Errors

- `ErrTransactionClosed`: Operation on closed transaction
- `ErrTransactionConflict`: Write conflict detected
- `ErrTransactionNotFound`: Invalid transaction ID
- `ErrReadOnlyTransaction`: Write operation on read-only transaction
- `ErrKeyNotFound`: Key does not exist
- `ErrNoActiveTransaction`: No active transaction in context
- `ErrNestedTransaction`: Nested transactions not supported

### Error Handling Strategy

1. **Conflict Detection**: Client receives error and should retry
2. **Closed Transaction**: Programming error, should not retry
3. **Read-Only Violation**: Programming error, fix code
4. **Key Not Found**: Normal condition, handle in application logic

## Testing Strategy

### Unit Tests

1. **Manager Tests** (`manager_test.go`)
   - Transaction lifecycle
   - Multiple concurrent transactions
   - Manager cleanup

2. **Isolation Tests** (`isolation_test.go`)
   - Snapshot isolation
   - Dirty read prevention
   - Repeatable read
   - Read-only transactions

3. **Concurrent Tests** (`concurrent_test.go`)
   - Write conflicts
   - Concurrent readers
   - High concurrency stress tests
   - Manager thread-safety

### Integration Tests

1. **SQL Integration** (`integration/transaction_test.go`)
   - BEGIN/COMMIT/ROLLBACK statements
   - Multi-client scenarios
   - Isolation between clients
   - Long-running transactions

## Performance Characteristics

### Read Performance
- **Single Transaction**: O(log n) via BadgerDB LSM tree
- **Concurrent Reads**: Linear scaling with number of cores
- **No Lock Contention**: MVCC enables lock-free reads

### Write Performance
- **Sequential Writes**: ~100k ops/sec (depends on hardware)
- **Concurrent Writes**: Limited by LSM tree compaction
- **Conflict Rate**: Depends on workload hot spots

### Memory Usage
- **Per Transaction**: ~1KB overhead + snapshot memory
- **Active Transactions**: Manager tracks in map
- **MVCC Versions**: Cleaned up by BadgerDB garbage collection

## Limitations

1. **No Distributed Transactions**: Single-node only
2. **No Savepoints**: Cannot partially rollback
3. **No Nested Transactions**: One transaction per connection
4. **Conflict Detection**: Only at commit time (optimistic)
5. **Isolation Level**: Fixed at Snapshot Isolation (no true Serializable)

## Future Enhancements

1. **Transaction Timeouts**: Auto-rollback long-running transactions
2. **Deadlock Detection**: Detect and resolve deadlocks
3. **Transaction Metrics**: Latency, conflict rate, active count
4. **Distributed Transactions**: Two-phase commit for multi-node
5. **Read-Your-Writes**: Guarantee within autocommit mode
6. **Pessimistic Locking**: SELECT FOR UPDATE support

## References

- BadgerDB Documentation: https://dgraph.io/docs/badger/
- MVCC Explanation: https://en.wikipedia.org/wiki/Multiversion_concurrency_control
- Snapshot Isolation: https://en.wikipedia.org/wiki/Snapshot_isolation
