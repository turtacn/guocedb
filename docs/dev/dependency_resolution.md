# Dependency Resolution & Build Fixes (Phase 2)

## Overview

This document details the strategies employed to resolve dependency conflicts and build failures encountered during the `Phase 2` dependency hygiene pass.

## Key Issues & Solutions

### 1. `go-mysql-server` & `go-icu-regex` (CGO)

**Problem:**
The project had a dependency on `github.com/dolthub/go-mysql-server`. This library transitively pulls in `github.com/dolthub/go-icu-regex`, which requires CGO and the `libicu-dev` system library (specifically `unicode/uregex.h`). This caused persistent build failures in environments where `pkg-config` or `libicu-dev` were missing or misconfigured.

**Solution: The "Purge" Strategy**
Instead of fighting the environment configuration, we removed the dependency entirely.
- We replaced all imports of `github.com/dolthub/go-mysql-server/sql` with our internal fork `github.com/turtacn/guocedb/compute/sql`.
- We removed the `protocol/mysql` package which heavily relied on `go-mysql-server/server`.
- We refactored `cmd/guocedb-server` to use our internal `compute/server` implementation directly.

This allowed us to build with `CGO_ENABLED=0`, significantly simplifying the build process and portability.

### 2. Interface Mismatches

**Problem:**
There were several mismatches between our internal `compute/sql` interfaces and the implementations in `compute/catalog`, `compute/server`, etc.
- `sql.Database` required a `Tables()` method which was missing in `memory_catalog`.
- `mysql.Handler` (from Vitess) required `ComInitDB` and `ComMultiQuery` methods which were missing in our handler.
- `network.ManagedServer` required `Start()` (no return), while our server implemented `Start() error`.

**Solution:**
- Implemented `Tables()` in `memory_catalog` and `persistent_catalog`.
- Implemented `ComInitDB` and `ComMultiQuery` in `compute/server/handler.go`.
- Changed `compute/server/server.go` to match the `Start()` signature (running accept loop in a goroutine).

### 3. Parser AST Changes

**Problem:**
The project uses `vitess` for SQL parsing. The AST structure in modern `vitess` versions changed significantly compared to what the code expected.
- `Union` struct was replaced by `SetOp`.
- `Select.Distinct` moved to `Select.QueryOpts.Distinct`.
- `Scope` fields were moved or changed types.

**Solution:**
- Rewrote `compute/sql/parse/parse.go` to adapt to the new AST structure.

## Build Instructions

The project can now be built without CGO:

```bash
CGO_ENABLED=0 go build ./...
```
