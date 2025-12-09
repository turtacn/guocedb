# Phase 7: Configuration Management & Service Bootstrap

## Quick Start

### Building

```bash
make build
```

### Running

```bash
# Start with default configuration
./bin/guocedb

# Start with config file
./bin/guocedb --config configs/guocedb.yaml

# Start with specific port
./bin/guocedb --port 13306

# Start with debug logging
./bin/guocedb --log-level debug
```

### Configuration Validation

```bash
# Check configuration
./bin/guocedb check --config configs/guocedb.yaml

# View help
./bin/guocedb --help

# View version
./bin/guocedb version
```

### Environment Variables

```bash
# Override port via environment
export GUOCEDB_SERVER_PORT=13306
./bin/guocedb

# Override multiple settings
export GUOCEDB_SERVER_PORT=13306
export GUOCEDB_LOGGING_LEVEL=debug
export GUOCEDB_SECURITY_ENABLED=true
./bin/guocedb --config configs/guocedb.yaml
```

## Testing

```bash
# Run all config tests
make test-config

# Run specific package tests
go test -v ./config/...
go test -v ./server/...

# Run all tests
make test
```

## Configuration Examples

### Development

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

### Production

```yaml
server:
  host: "0.0.0.0"
  port: 3306
  max_connections: 10000
storage:
  data_dir: "/var/lib/guocedb/data"
  sync_writes: true
security:
  enabled: true
  root_password: "${ROOT_PASSWORD}"
observability:
  enabled: true
  address: ":9090"
logging:
  level: "info"
  format: "json"
  output: "/var/log/guocedb/server.log"
```

## Architecture

### Configuration Priority

```
CLI Flags > Environment Variables > Config File > Defaults
```

### Server Lifecycle

```
New → Starting → Running → Stopping → Stopped
```

### Graceful Shutdown

1. Stop accepting new connections
2. Drain active connections (with timeout)
3. Close observability services
4. Close transaction manager
5. Close storage engine

## Documentation

- [Configuration Reference](configuration.md) - Complete configuration documentation
- [Phase 7 Summary](PHASE7_SUMMARY.md) - Detailed implementation summary
- [Architecture](../architecture.md) - Updated architecture documentation

## Key Features

✅ Multi-source configuration (YAML, env vars, CLI)
✅ Configuration validation with detailed errors
✅ Graceful shutdown with connection draining
✅ Lifecycle hooks for extensibility
✅ Signal handling (SIGTERM/SIGINT)
✅ Health checks and metrics
✅ Structured logging with rotation

## Status

**Branch**: `feat/round2-phase7-config-bootstrap`
**Tests**: 21/21 passing
**Build**: ✅ Success
**Status**: ✅ Ready for integration

## Next Steps

1. Integrate with actual storage implementation
2. Implement server lifecycle tests
3. Add hot configuration reload (SIGHUP)
4. Implement data directory initialization
5. Add configuration file generation tool
