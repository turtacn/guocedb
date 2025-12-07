# Round 1 Phase 4: Storage Engine Interface (BadgerDB)

## Overview

This phase completes the implementation of the BadgerDB storage engine adapter, satisfying the `sql.Database` and `sql.Table` interfaces required by the `guocedb` compute engine.

## Design Decisions

### 1. Persistence Model

*   **Metadata:** Table schemas are persisted in BadgerDB using a metadata prefix (`MetaPrefix`).
    *   `Key`: `meta:<db_name>/table:<table_name>`
    *   `Value`: JSON serialization of the schema.
*   **Data:** Row data is persisted using a data prefix.
    *   `Key`: `data:<db_name>:<table_name>:<pk_bytes>`
    *   `Value`: GOB serialization of the `sql.Row`.
*   **Recovery:** On database initialization (`NewDatabase`), the engine scans the metadata prefix to reconstruct the in-memory `Tables` map.

### 2. Schema Serialization

Since `sql.Type` is an interface and cannot be directly marshaled to JSON, a DTO `SerializableColumn` is used.
*   It maps `sql.Type` to `int32` (representing `query.Type` from Vitess).
*   During unmarshaling, `sql.MysqlTypeToType` is used to restore the correct `sql.Type` implementation.

### 3. Interface Adaptation

The core `compute/sql` package in this repository version does **not** define `InsertableTable`, `UpdatableTable`, or `DeletableTable` interfaces, nor does it define `RowInserter`.

To support the requested CRUD functionality and transactional patterns (StatementBegin/Complete), we defined local interfaces in `storage/engines/badger/table.go` that mirror these concepts. The `Table` struct implements `sql.Inserter` (the core interface) by delegating to these local helpers.

*   `Table.Insert` (Core Interface): Creates a short-lived `rowEditor`, performs the insert, and commits. This provides implicit transaction safety per row.

### 4. Transaction Handling

Since the current execution plans (`InsertInto`) do not expose explicit `StatementBegin` / `StatementComplete` hooks to the table, atomic bulk inserts are not natively supported by the engine-storage contract yet.
*   **Current Implementation:** `Insert` wraps each row in a Badger transaction.
*   **Future Improvement:** If `sql.Context` or execution plans are updated to carry transaction state, `rowEditor` can be reused across rows.

## Limitations

*   **Primary Keys:** The current `sql.Column` definition does not track Primary Key status. The engine currently assumes the first column is the PK for storage key generation, or relies on the caller providing a unique key.
*   **Atomicity:** Bulk inserts are not atomic (row-by-row commit).

## Verified Functionality

*   Database Creation/persistence.
*   Table Creation/Dropping/Persistence.
*   Row Insertion, Update, Deletion.
*   Data persistence across restarts.
