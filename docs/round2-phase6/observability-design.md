# GuoceDB Observability Design

## Overview

This document describes the observability implementation for GuoceDB, including metrics collection, health checks, and diagnostic endpoints.

## Architecture

```
observability/
├── server.go              # Main observability HTTP server
├── metrics/
│   ├── metrics.go         # Prometheus metric definitions
│   ├── collector.go       # Custom storage collector
│   └── handler.go         # /metrics HTTP handler
├── health/
│   ├── health.go          # Health check logic
│   └── handler.go         # /health, /ready, /live handlers
└── diagnostic/
    ├── diagnostic.go      # Diagnostic data collection
    ├── handler.go         # /debug/* handlers
    └── interfaces.go      # Dependency interfaces
```

## Components

### 1. Metrics Package

Provides Prometheus-compatible metrics for monitoring GuoceDB performance.

**Key Metrics:**
- Connection metrics (active, total, rejected)
- Query metrics (count, duration, errors)
- Transaction metrics (commit, rollback, conflicts)
- Storage metrics (size, keys, tables)
- Error metrics (by type)

**Implementation:**
- Uses `prometheus/client_golang` for metric registration
- Provides helper functions for recording metrics
- Custom collector for BadgerDB storage statistics

### 2. Health Package

Implements health check endpoints for Kubernetes and service monitoring.

**Endpoints:**
- `/health` - Comprehensive health status with check details
- `/ready` - Kubernetes readiness probe
- `/live` - Kubernetes liveness probe

**Features:**
- Concurrent check execution
- Configurable timeout
- Extensible check registration
- Version reporting
- Uptime tracking

### 3. Diagnostic Package

Provides runtime diagnostic information and profiling endpoints.

**Endpoints:**
- `/debug/diagnostic` - Complete diagnostic snapshot
- `/debug/queries` - Active queries
- `/debug/connections` - Connection details
- `/debug/memory` - Memory statistics
- `/debug/gc` - Garbage collection stats and trigger
- `/debug/pprof/*` - Go profiling endpoints

**Features:**
- Runtime metrics (goroutines, memory, GC)
- Slow query recording
- Connection tracking
- Transaction statistics
- Storage statistics

### 4. Observability Server

Unified HTTP server managing all observability endpoints.

**Configuration:**
```go
type ServerConfig struct {
    Enabled     bool
    Address     string  // e.g., ":9090"
    MetricsPath string  // e.g., "/metrics"
    EnablePprof bool
}
```

**Features:**
- Single server for all observability endpoints
- Optional pprof endpoints
- Graceful shutdown
- Endpoint discovery via root path

## Usage

### Starting the Observability Server

```go
import (
    "github.com/turtacn/guocedb/observability"
    "github.com/turtacn/guocedb/observability/health"
    "github.com/turtacn/guocedb/observability/diagnostic"
)

// Create health checker
checker := health.NewChecker()
checker.SetVersion("1.0.0")
checker.AddCheck("storage", health.StorageHealthCheck(storage))

// Create diagnostics
diag := diagnostic.NewDiagnostics(connMgr, queryMgr, txnMgr, storage)

// Configure and start server
config := observability.DefaultConfig()
server := observability.NewServer(config, checker, diag)
server.Start()

// Graceful shutdown
defer server.Stop(context.Background())
```

### Recording Metrics

```go
import "github.com/turtacn/guocedb/observability/metrics"

// Connection metrics
metrics.ConnectionsActive.Inc()
metrics.ConnectionsTotal.Inc()
metrics.RecordConnectionRejected("max_connections")

// Query metrics
metrics.RecordQuery("select", duration, success)

// Transaction metrics
metrics.RecordTransaction("commit", duration)

// Error metrics
metrics.RecordError("execution")
```

### Adding Health Checks

```go
checker.AddCheck("database", func(ctx context.Context) error {
    // Test database connectivity
    return db.Ping()
})

checker.AddCheck("storage", func(ctx context.Context) error {
    // Test storage access
    _, err := storage.Get([]byte("test"))
    return err
})
```

## Prometheus Integration

### Scrape Configuration

```yaml
scrape_configs:
  - job_name: 'guocedb'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### Example Queries

```promql
# Query rate by type
rate(guocedb_queries_total[5m])

# 95th percentile query latency
histogram_quantile(0.95, rate(guocedb_query_duration_seconds_bucket[5m]))

# Active connections
guocedb_connections_active

# Transaction success rate
rate(guocedb_transactions_total{status="commit"}[5m]) / 
rate(guocedb_transactions_total[5m])

# Error rate
rate(guocedb_errors_total[5m])
```

## Kubernetes Integration

### Deployment Configuration

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: guocedb
spec:
  containers:
  - name: guocedb
    image: guocedb:latest
    ports:
    - name: mysql
      containerPort: 3306
    - name: observability
      containerPort: 9090
    livenessProbe:
      httpGet:
        path: /live
        port: 9090
      initialDelaySeconds: 30
      periodSeconds: 10
    readinessProbe:
      httpGet:
        path: /ready
        port: 9090
      initialDelaySeconds: 5
      periodSeconds: 5
```

## Performance Considerations

1. **Metric Collection Overhead:**
   - Minimal overhead using atomic operations
   - Histogram buckets pre-configured for typical latencies
   - Avoid excessive cardinality in labels

2. **Health Check Timeout:**
   - Default 5-second timeout
   - Checks run concurrently
   - Configure based on check complexity

3. **Diagnostic Data:**
   - Slow query buffer limited to 100 queries
   - Ring buffer prevents memory growth
   - Pprof endpoints should be protected in production

## Security Considerations

1. **Endpoint Protection:**
   - Observability server should be on separate port
   - Use firewall rules to restrict access
   - Consider authentication for sensitive endpoints

2. **Information Disclosure:**
   - Metrics may reveal system information
   - pprof endpoints expose internal details
   - Disable pprof in production if not needed

3. **Resource Usage:**
   - pprof profiling can impact performance
   - Limit concurrent profile requests
   - Monitor observability server resource usage

## Future Enhancements

1. **Tracing Integration:**
   - OpenTelemetry support
   - Distributed tracing
   - Request correlation

2. **Advanced Metrics:**
   - Query plan metrics
   - Cache hit rates
   - Lock contention

3. **Alerting:**
   - Built-in alert rules
   - Webhook notifications
   - Alert aggregation

4. **Dashboard:**
   - Embedded Grafana dashboards
   - Real-time query visualization
   - Performance recommendations
