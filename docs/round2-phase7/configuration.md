# GuoceDB Configuration Reference

## Overview

GuoceDB supports flexible configuration through multiple sources with a clear priority order.

## Configuration Sources (Priority Order)

1. **Command-line flags** (highest priority)
2. **Environment variables** (prefix: `GUOCEDB_`)
3. **Configuration file** (YAML format)
4. **Default values** (lowest priority)

## Environment Variable Naming

Environment variables use uppercase with underscore separators:

- `server.port` → `GUOCEDB_SERVER_PORT`
- `storage.data_dir` → `GUOCEDB_STORAGE_DATA_DIR`
- `security.enabled` → `GUOCEDB_SECURITY_ENABLED`
- `logging.level` → `GUOCEDB_LOGGING_LEVEL`

## Command-line Flags

### Global Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config, -c` | string | - | Configuration file path |

### Server Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--host` | string | 0.0.0.0 | Server listen address |
| `--port, -p` | int | 3306 | MySQL protocol port |
| `--data-dir, -d` | string | ./data | Data directory path |
| `--log-level` | string | info | Log level (debug, info, warn, error) |
| `--auth` | bool | false | Enable authentication |
| `--metrics` | bool | true | Enable metrics endpoint |

### Commands

| Command | Description |
|---------|-------------|
| `guocedb` | Start the database server |
| `guocedb version` | Print version information |
| `guocedb check` | Validate configuration |
| `guocedb help` | Show help |

## Configuration File Format

GuoceDB uses YAML for configuration files.

### Example Configuration

```yaml
# GuoceDB Configuration

server:
  host: "0.0.0.0"
  port: 3306
  max_connections: 1000
  connect_timeout: 10s
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 8h
  shutdown_timeout: 30s

storage:
  data_dir: "./data"
  wal_dir: ""  # Empty means use data_dir
  max_memtable_size: 67108864  # 64MB
  num_compactors: 4
  sync_writes: false
  valuelog_gc: true

security:
  enabled: false
  root_password: ""
  auth_plugin: "mysql_native_password"
  max_auth_attempts: 5
  lock_duration: 15m
  audit_log:
    enabled: false
    file_path: "./audit.log"
    async: true

observability:
  enabled: true
  address: ":9090"
  metrics_path: "/metrics"
  enable_pprof: true

logging:
  level: "info"     # debug, info, warn, error
  format: "json"    # json, text
  output: "stdout"  # stdout or file path
  max_size: 100     # MB
  max_backups: 3
  max_age: 7        # days
```

## Configuration Sections

### Server Configuration

Controls the MySQL protocol server behavior.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `host` | string | 0.0.0.0 | Network interface to bind to |
| `port` | int | 3306 | TCP port for MySQL protocol |
| `max_connections` | int | 1000 | Maximum concurrent client connections |
| `connect_timeout` | duration | 10s | Timeout for initial connection |
| `read_timeout` | duration | 30s | Timeout for reading from client |
| `write_timeout` | duration | 30s | Timeout for writing to client |
| `idle_timeout` | duration | 8h | Idle connection timeout |
| `shutdown_timeout` | duration | 30s | Graceful shutdown timeout |

### Storage Configuration

Controls the underlying storage engine (BadgerDB).

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `data_dir` | string | ./data | Primary data directory (required) |
| `wal_dir` | string | "" | Write-ahead log directory (uses data_dir if empty) |
| `max_memtable_size` | int64 | 67108864 | Max MemTable size in bytes (min: 1MB) |
| `num_compactors` | int | 4 | Number of compaction goroutines |
| `sync_writes` | bool | false | Sync writes to disk (slower but safer) |
| `valuelog_gc` | bool | true | Enable value log garbage collection |

### Security Configuration

Controls authentication and authorization.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | false | Enable security features |
| `root_password` | string | "" | Initial root password (if empty, set on first start) |
| `auth_plugin` | string | mysql_native_password | Authentication plugin (mysql_native_password, caching_sha2_password) |
| `max_auth_attempts` | int | 5 | Maximum failed login attempts before lockout |
| `lock_duration` | duration | 15m | Lockout duration after max failed attempts |

#### Audit Log Configuration

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `audit_log.enabled` | bool | false | Enable audit logging |
| `audit_log.file_path` | string | ./audit.log | Audit log file path |
| `audit_log.async` | bool | true | Use async logging for performance |

### Observability Configuration

Controls metrics, health checks, and profiling endpoints.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | true | Enable observability HTTP server |
| `address` | string | :9090 | HTTP server listen address |
| `metrics_path` | string | /metrics | Prometheus metrics endpoint path |
| `enable_pprof` | bool | true | Enable pprof profiling endpoints |

#### Available Endpoints

When observability is enabled:

- `GET /metrics` - Prometheus metrics
- `GET /ready` - Readiness probe (returns 200 when ready)
- `GET /health` - Health check (returns 200 when healthy)
- `GET /debug/pprof/*` - pprof profiling endpoints (if enabled)

### Logging Configuration

Controls application logging behavior.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `level` | string | info | Log level: debug, info, warn, error |
| `format` | string | json | Output format: json, text |
| `output` | string | stdout | Output destination: stdout or file path |
| `max_size` | int | 100 | Max log file size in MB (file output only) |
| `max_backups` | int | 3 | Max number of old log files to retain |
| `max_age` | int | 7 | Max days to retain old log files |

## Configuration Examples

### Development Setup

```yaml
server:
  port: 3306
storage:
  data_dir: "./dev-data"
  sync_writes: false
security:
  enabled: false
logging:
  level: "debug"
  format: "text"
```

### Production Setup

```yaml
server:
  host: "0.0.0.0"
  port: 3306
  max_connections: 10000
  shutdown_timeout: 60s
storage:
  data_dir: "/var/lib/guocedb/data"
  wal_dir: "/var/lib/guocedb/wal"
  sync_writes: true
  num_compactors: 8
security:
  enabled: true
  root_password: "${ROOT_PASSWORD}"  # From environment
  max_auth_attempts: 3
  audit_log:
    enabled: true
    file_path: "/var/log/guocedb/audit.log"
observability:
  enabled: true
  address: ":9090"
logging:
  level: "info"
  format: "json"
  output: "/var/log/guocedb/server.log"
  max_size: 500
  max_backups: 10
  max_age: 30
```

### Environment Variable Override

```bash
# Start with config file but override port
export GUOCEDB_SERVER_PORT=13306
export GUOCEDB_SECURITY_ENABLED=true
./guocedb --config production.yaml
```

### Command-line Override

```bash
# Override specific settings via CLI
./guocedb \
  --config production.yaml \
  --port 13306 \
  --log-level debug \
  --auth
```

## Configuration Validation

GuoceDB validates configuration on startup and provides detailed error messages.

### Check Configuration

Validate your configuration without starting the server:

```bash
./guocedb check --config path/to/config.yaml
```

### Common Validation Errors

- **Port out of range**: Port must be between 1-65535
- **Empty data directory**: `storage.data_dir` is required
- **Invalid log level**: Must be one of: debug, info, warn, error
- **Invalid log format**: Must be one of: json, text
- **Small memtable size**: Must be at least 1MB
- **Invalid auth plugin**: Must be supported plugin name

## Best Practices

1. **Use configuration files** for base settings, override with env vars for secrets
2. **Enable security** in production environments
3. **Enable sync_writes** for data durability (with performance tradeoff)
4. **Tune max_connections** based on expected workload
5. **Use structured logging** (JSON format) in production
6. **Enable observability** for monitoring and debugging
7. **Set appropriate timeouts** for your workload patterns
8. **Use separate wal_dir** on fast storage (SSD) for better write performance

## Troubleshooting

### Server Won't Start

1. Check configuration with `guocedb check`
2. Verify data directory exists and is writable
3. Ensure port is not already in use
4. Check log output for specific errors

### Performance Issues

1. Increase `max_connections` if hitting limits
2. Tune `num_compactors` based on CPU cores
3. Enable `sync_writes` only if data durability is critical
4. Monitor metrics at `:9090/metrics`
5. Use pprof for CPU/memory profiling

### Security Issues

1. Ensure `security.enabled = true`
2. Set strong root password
3. Enable audit logging
4. Review failed authentication attempts
5. Adjust `max_auth_attempts` and `lock_duration` as needed
