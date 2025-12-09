# GuoceDB Metrics Reference

## Overview

This document provides a comprehensive reference for all metrics exposed by GuoceDB.

## Metric Types

- **Counter**: Monotonically increasing value (e.g., total queries)
- **Gauge**: Value that can go up or down (e.g., active connections)
- **Histogram**: Distribution of values (e.g., query latency)

## Connection Metrics

### guocedb_connections_active

- **Type**: Gauge
- **Description**: Number of currently active connections
- **Labels**: None

```promql
# Example: Current active connections
guocedb_connections_active
```

### guocedb_connections_total

- **Type**: Counter
- **Description**: Total number of connections since startup (cumulative)
- **Labels**: None

```promql
# Example: Connection rate over 5 minutes
rate(guocedb_connections_total[5m])
```

### guocedb_connections_rejected_total

- **Type**: Counter
- **Description**: Total number of rejected connections
- **Labels**: 
  - `reason`: Rejection reason

**Reason Values:**
- `max_connections` - Maximum connection limit reached
- `auth_failed` - Authentication failed
- `too_many_user_connections` - Per-user connection limit reached

```promql
# Example: Rejection rate by reason
rate(guocedb_connections_rejected_total[5m])
```

## Query Metrics

### guocedb_queries_total

- **Type**: Counter
- **Description**: Total number of queries executed
- **Labels**:
  - `type`: Query type
  - `status`: Execution status

**Type Values:**
- `select` - SELECT queries
- `insert` - INSERT queries
- `update` - UPDATE queries
- `delete` - DELETE queries
- `create` - CREATE TABLE/DATABASE
- `drop` - DROP TABLE/DATABASE
- `alter` - ALTER TABLE
- `begin` - BEGIN TRANSACTION
- `commit` - COMMIT
- `rollback` - ROLLBACK
- `other` - Other query types

**Status Values:**
- `success` - Query succeeded
- `error` - Query failed

```promql
# Example: Query rate by type
rate(guocedb_queries_total[5m])

# Example: Error rate
rate(guocedb_queries_total{status="error"}[5m]) / 
rate(guocedb_queries_total[5m])
```

### guocedb_query_duration_seconds

- **Type**: Histogram
- **Description**: Query execution duration in seconds
- **Labels**:
  - `type`: Query type

**Buckets**: 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10 seconds

```promql
# Example: 95th percentile latency
histogram_quantile(0.95, 
  rate(guocedb_query_duration_seconds_bucket[5m]))

# Example: Average latency
rate(guocedb_query_duration_seconds_sum[5m]) / 
rate(guocedb_query_duration_seconds_count[5m])
```

### guocedb_rows_read_total

- **Type**: Counter
- **Description**: Total number of rows read
- **Labels**: None

```promql
# Example: Rows read per second
rate(guocedb_rows_read_total[5m])
```

### guocedb_rows_written_total

- **Type**: Counter
- **Description**: Total number of rows written (INSERT/UPDATE/DELETE)
- **Labels**: None

```promql
# Example: Write throughput
rate(guocedb_rows_written_total[5m])
```

## Transaction Metrics

### guocedb_transactions_total

- **Type**: Counter
- **Description**: Total number of transactions
- **Labels**:
  - `status`: Transaction outcome

**Status Values:**
- `commit` - Successfully committed
- `rollback` - Rolled back
- `conflict` - Failed due to conflict
- `error` - Failed due to other error

```promql
# Example: Transaction success rate
rate(guocedb_transactions_total{status="commit"}[5m]) / 
rate(guocedb_transactions_total[5m])

# Example: Conflict rate
rate(guocedb_transactions_total{status="conflict"}[5m])
```

### guocedb_transaction_duration_seconds

- **Type**: Histogram
- **Description**: Transaction duration from BEGIN to COMMIT/ROLLBACK
- **Labels**: None

**Buckets**: Default Prometheus buckets (0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10)

```promql
# Example: 99th percentile transaction duration
histogram_quantile(0.99, 
  rate(guocedb_transaction_duration_seconds_bucket[5m]))
```

## Storage Metrics

### guocedb_storage_lsm_bytes

- **Type**: Gauge
- **Description**: LSM tree size in bytes
- **Labels**: None

```promql
# Example: LSM tree size in GB
guocedb_storage_lsm_bytes / 1024 / 1024 / 1024
```

### guocedb_storage_vlog_bytes

- **Type**: Gauge
- **Description**: Value log size in bytes
- **Labels**: None

```promql
# Example: Total storage size
guocedb_storage_lsm_bytes + guocedb_storage_vlog_bytes
```

### guocedb_storage_keys_total

- **Type**: Gauge
- **Description**: Total number of keys in storage
- **Labels**: None

```promql
# Example: Key count growth
rate(guocedb_storage_keys_total[1h])
```

### guocedb_storage_tables_total

- **Type**: Gauge
- **Description**: Number of SST tables in BadgerDB
- **Labels**:
  - `level`: LSM level (0, 1, 2, ...)

```promql
# Example: Tables at each level
guocedb_storage_tables_total

# Example: Total tables
sum(guocedb_storage_tables_total)
```

## Error Metrics

### guocedb_errors_total

- **Type**: Counter
- **Description**: Total number of errors
- **Labels**:
  - `type`: Error type

**Type Values:**
- `parse` - SQL parsing errors
- `execution` - Query execution errors
- `transaction` - Transaction errors
- `storage` - Storage layer errors
- `auth` - Authentication/authorization errors

```promql
# Example: Error rate by type
rate(guocedb_errors_total[5m])

# Example: Total error rate
sum(rate(guocedb_errors_total[5m]))
```

## HTTP Endpoints

### Metrics Endpoint

**URL**: `/metrics`  
**Method**: GET  
**Content-Type**: text/plain

Returns all metrics in Prometheus exposition format.

### Health Check Endpoints

#### /health

**URL**: `/health`  
**Method**: GET  
**Content-Type**: application/json

Returns detailed health status including all registered checks.

**Response Example:**
```json
{
  "status": "healthy",
  "timestamp": "2024-12-09T10:00:00Z",
  "checks": [
    {
      "name": "storage",
      "status": "healthy",
      "duration_ms": 5
    }
  ],
  "version": "1.0.0",
  "uptime_ms": 3600000
}
```

#### /ready

**URL**: `/ready`  
**Method**: GET  
**Content-Type**: application/json

Kubernetes readiness probe endpoint.

**Response Example (Ready):**
```json
{
  "status": "ready"
}
```

**Response Example (Not Ready):**
```json
{
  "status": "not ready",
  "reason": "storage unavailable"
}
```

#### /live

**URL**: `/live`  
**Method**: GET  
**Content-Type**: application/json

Kubernetes liveness probe endpoint.

**Response Example:**
```json
{
  "status": "alive",
  "uptime": "1h30m45s"
}
```

### Diagnostic Endpoints

#### /debug/diagnostic

**URL**: `/debug/diagnostic`  
**Method**: GET  
**Content-Type**: application/json

Returns comprehensive diagnostic information.

**Response Example:**
```json
{
  "timestamp": "2024-12-09T10:00:00Z",
  "runtime": {
    "go_version": "go1.25.0",
    "num_goroutine": 42,
    "num_cpu": 8,
    "mem_alloc_bytes": 10485760
  },
  "connections": {
    "active": 5,
    "total": 1000
  },
  "storage": {
    "data_size_bytes": 1073741824,
    "num_keys": 10000
  }
}
```

#### /debug/memory

**URL**: `/debug/memory`  
**Method**: GET  
**Content-Type**: application/json

Returns detailed memory statistics.

#### /debug/pprof/*

Standard Go pprof endpoints for profiling:
- `/debug/pprof/` - Index
- `/debug/pprof/goroutine` - Goroutine dump
- `/debug/pprof/heap` - Heap profile
- `/debug/pprof/profile` - CPU profile
- `/debug/pprof/trace` - Execution trace

## Grafana Dashboard Examples

### Query Performance Dashboard

```promql
# QPS by query type
sum(rate(guocedb_queries_total[5m])) by (type)

# P95 latency
histogram_quantile(0.95, sum(rate(guocedb_query_duration_seconds_bucket[5m])) by (le, type))

# Error rate
sum(rate(guocedb_queries_total{status="error"}[5m])) / 
sum(rate(guocedb_queries_total[5m]))
```

### Connection Dashboard

```promql
# Active connections
guocedb_connections_active

# Connection rate
rate(guocedb_connections_total[5m])

# Rejection rate
rate(guocedb_connections_rejected_total[5m])
```

### Storage Dashboard

```promql
# Total storage size
(guocedb_storage_lsm_bytes + guocedb_storage_vlog_bytes) / 1024 / 1024 / 1024

# Key count
guocedb_storage_keys_total

# LSM levels
guocedb_storage_tables_total
```

## Alert Rules

### High Error Rate

```yaml
- alert: HighErrorRate
  expr: |
    sum(rate(guocedb_errors_total[5m])) > 10
  for: 5m
  annotations:
    summary: "High error rate detected"
    description: "Error rate is {{ $value }} errors/sec"
```

### Slow Queries

```yaml
- alert: SlowQueries
  expr: |
    histogram_quantile(0.99, 
      rate(guocedb_query_duration_seconds_bucket[5m])) > 5
  for: 5m
  annotations:
    summary: "Queries are slow"
    description: "P99 latency is {{ $value }} seconds"
```

### High Connection Usage

```yaml
- alert: HighConnectionUsage
  expr: |
    guocedb_connections_active > 900
  for: 5m
  annotations:
    summary: "Connection limit approaching"
    description: "{{ $value }} active connections"
```
