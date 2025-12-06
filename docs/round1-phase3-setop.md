# Round 1 Phase 3: SQL Parser SetOp Adaptation & Vitess AST Compatibility

## Overview
This phase focused on adapting the SQL parser to handle `SetOp` structures from the Vitess AST, enabling support for `UNION`, `INTERSECT`, and `EXCEPT` operations, including their `ALL` variants.

## Changes

### 1. SQL Parser (`compute/sql/parse/parse.go`)
- Implemented `convertSetOp` function to handle `*sqlparser.SetOp`.
- Mapped `sqlparser.SetOp` types to `plan.Union`, `plan.Intersect`, and `plan.Except`.
- Handled `UNION ALL`, `INTERSECT ALL`, and `EXCEPT ALL` by passing `false` to the `distinct` parameter.
- Integrated `SetOp` handling into the main `convert` switch case.
- Supported `OrderBy` and `Limit` clauses on `SetOp` nodes.

### 2. Plan Nodes (`compute/sql/plan/`)
- **`union.go`**: Implemented `Union` node with `RowIter` supporting both distinct and all modes. `Union DISTINCT` uses a hash map to dedup rows.
- **`intersect.go`**: Implemented `Intersect` node. `RowIter` loads the right side into a map and streams the left side, emitting rows present in both. Supports distinct logic.
- **`except.go`**: Implemented `Except` node. `RowIter` loads the right side into a map and streams the left side, emitting rows NOT present in the right side. Supports distinct logic.
- **`setop.go`**: Added shared helper `hashRow` using `mitchellh/hashstructure` for row hashing.

### 3. Error Handling (`compute/sql/core.go`)
- Added `ErrInvalidChildrenNumber` to `compute/sql/core.go` for validating child node counts in `WithChildren`.

### 4. Tests (`compute/sql/parse/setop_test.go`)
- Added comprehensive unit tests for:
    - `UNION` and `UNION ALL`
    - `INTERSECT` and `INTERSECT ALL`
    - `EXCEPT` and `EXCEPT ALL`
- Verified correct plan structure generation.

## Verification
- Ran `go test ./compute/sql/parse/...` to ensure parser correctness.
- Verified that existing tests pass and new SetOp tests pass.
- Ensured no CGO dependencies were introduced.

## Notes
- The Vitess AST `SetOp` struct in the pinned version uses `Type` string to distinguish between `UNION` and `UNION ALL` (implicitly via `UNION` vs `UNION ALL` string checks or absence of `Distinct` field). We observed `Type` values like `union`, `union all`, `intersect`, `except`. The implementation handles these string values to set the `Distinct` flag correctly.
