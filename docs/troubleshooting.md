# GuoceDB Troubleshooting Guide

## Connection Issues

### Cannot Connect to Server

**Symptoms:**
```
ERROR 2003 (HY000): Can't connect to MySQL server on '127.0.0.1:3306'
```

**Solutions:**

1. Check if server is running:
```bash
ps aux | grep guocedb
systemctl status guocedb
```

2. Check listening port:
```bash
netstat -tlnp | grep 3306
ss -tlnp | grep 3306
```

3. Check firewall:
```bash
sudo ufw status
sudo iptables -L -n | grep 3306
```

4. Check logs:
```bash
tail -f /var/log/guocedb/guocedb.log
journalctl -u guocedb -f
```

### Access Denied

**Symptoms:**
```
ERROR 1045 (28000): Access denied for user 'root'@'localhost'
```

**Solutions:**

1. Verify credentials
2. Check if security is enabled in config
3. Reset root password if needed
4. Check authentication logs

### Too Many Connections

**Symptoms:**
```
ERROR 1040 (08004): Too many connections
```

**Solutions:**

1. Increase `max_connections`:
```yaml
server:
  max_connections: 2000
```

2. Check for connection leaks in application code
3. Monitor active connections:
```sql
SHOW PROCESSLIST;
```

## Query Performance Issues

### Slow Queries

**Diagnostics:**

1. Check query execution time
2. Look for full table scans
3. Review indexes

**Solutions:**

1. Add appropriate indexes:
```sql
CREATE INDEX idx_column ON table_name(column_name);
```

2. Use LIMIT to reduce result sets:
```sql
SELECT * FROM large_table LIMIT 100;
```

3. Optimize query structure
4. Review data types (use appropriate sizes)

### Query Timeout

**Symptoms:**
- Query hangs indefinitely
- Timeout errors

**Solutions:**

1. Check for long-running transactions:
```sql
SHOW PROCESSLIST;
```

2. Add indexes to improve performance
3. Break large queries into smaller chunks
4. Increase timeout if needed

## Transaction Issues

### Transaction Conflicts

**Symptoms:**
```
ERROR: transaction conflict, please retry
```

**Solutions:**

1. This is normal under high concurrency
2. Implement retry logic:
```go
for retries := 0; retries < 3; retries++ {
    err := performTransaction()
    if err == nil || !isConflictError(err) {
        break
    }
    time.Sleep(backoff(retries))
}
```

3. Reduce transaction scope
4. Keep transactions short

### Deadlocks

**Symptoms:**
- Transactions stuck waiting
- Deadlock detected errors

**Solutions:**

1. Access tables in consistent order
2. Keep transactions small
3. Use appropriate indexes
4. Consider reducing isolation level (if supported)

## Storage Issues

### Disk Full

**Symptoms:**
```
ERROR: no space left on device
```

**Solutions:**

1. Check disk usage:
```bash
df -h /var/lib/guocedb
du -sh /var/lib/guocedb/*
```

2. Clean up old backups
3. Archive old data
4. Add more disk space
5. Enable compression (if available)

### Slow Startup

**Symptoms:**
- Server takes long time to start
- WAL replay messages in logs

**Solutions:**

1. This may be normal after crash (WAL replay)
2. Check logs for progress
3. Ensure SSD storage
4. Monitor with:
```bash
journalctl -u guocedb -f
```

### Data Corruption

**Symptoms:**
- Unexpected query results
- Checksum errors in logs
- Startup failures

**Solutions:**

1. **DO NOT** delete data files
2. Stop server immediately
3. Check for hardware issues
4. Restore from backup
5. Contact support with logs

## Performance Issues

### High CPU Usage

**Diagnostics:**

```bash
# Check CPU usage
top -p $(pgrep guocedb)

# Profile CPU
curl http://localhost:9090/debug/pprof/profile?seconds=30 > cpu.pprof
go tool pprof cpu.pprof
```

**Solutions:**

1. Check for full table scans - add indexes
2. Reduce query complexity
3. Scale up CPU resources
4. Enable query caching (if available)

### High Memory Usage

**Diagnostics:**

```bash
# Check memory usage
ps aux | grep guocedb

# Profile memory
curl http://localhost:9090/debug/pprof/heap > heap.pprof
go tool pprof heap.pprof
```

**Solutions:**

1. Limit connection pool size
2. Reduce result set sizes
3. Check for memory leaks
4. Scale up memory resources

### High Latency

**Diagnostics:**

1. Check metrics:
```bash
curl http://localhost:9090/metrics | grep duration
```

2. Check I/O:
```bash
iostat -x 1
```

**Solutions:**

1. Use SSD storage
2. Add indexes
3. Optimize queries
4. Review storage configuration
5. Check network latency

## Logging Issues

### Log Files Too Large

**Solutions:**

1. Configure log rotation:
```yaml
logging:
  max_size: 100  # MB
  max_backups: 5
  max_age: 30    # days
```

2. Reduce log level:
```yaml
logging:
  level: info  # or warn
```

### Cannot Find Logs

**Check locations:**

```bash
# Default locations
/var/log/guocedb/guocedb.log
journalctl -u guocedb

# Configuration
cat /etc/guocedb/config.yaml | grep output
```

## Monitoring & Alerts

### Recommended Alerts

| Metric                               | Threshold  | Severity |
| ------------------------------------ | ---------- | -------- |
| Server Down                          | up == 0    | Critical |
| High Connection Usage                | > 80%      | Warning  |
| Query Error Rate                     | > 10/s     | Warning  |
| Transaction Conflicts                | > 100/s    | Info     |
| Disk Usage                           | > 85%      | Warning  |
| CPU Usage                            | > 90%      | Warning  |
| Memory Usage                         | > 90%      | Warning  |

### Health Check Endpoints

```bash
# Liveness
curl http://localhost:9090/live

# Readiness
curl http://localhost:9090/ready

# Detailed health
curl http://localhost:9090/health
```

## Common Error Codes

| Code | Message                    | Solution                          |
| ---- | -------------------------- | --------------------------------- |
| 1040 | Too many connections       | Increase max_connections          |
| 1045 | Access denied              | Check credentials                 |
| 1062 | Duplicate entry            | Check unique constraints          |
| 1146 | Table doesn't exist        | Verify table name and database    |
| 2003 | Can't connect              | Check server is running           |
| 2006 | Server has gone away       | Check connection timeout settings |

## Backup & Recovery Issues

### Backup Fails

**Solutions:**

1. Check disk space
2. Verify permissions
3. Check if tables are locked
4. Use consistent snapshot:
```bash
mysqldump --single-transaction ...
```

### Restore Fails

**Solutions:**

1. Check backup file integrity
2. Verify database exists
3. Check user permissions
4. Review error messages
5. Restore to empty database first

## Upgrade Issues

### Version Incompatibility

**Solutions:**

1. Review release notes
2. Check migration guide
3. Test on staging first
4. Keep backup before upgrade

### Data Migration Problems

**Solutions:**

1. Verify data format compatibility
2. Use import/export tools
3. Check for schema changes
4. Validate data after migration

## Diagnostic Commands

### Server Status

```sql
SHOW STATUS;
SHOW VARIABLES;
SHOW PROCESSLIST;
```

### Connection Info

```sql
SELECT USER(), DATABASE(), VERSION();
SHOW FULL PROCESSLIST;
```

### Storage Info

```bash
# Data directory size
du -sh /var/lib/guocedb

# File count
find /var/lib/guocedb -type f | wc -l

# Disk I/O
iostat -x 1 5
```

### Network Diagnostics

```bash
# Test connectivity
telnet localhost 3306

# Check port usage
lsof -i :3306

# Network stats
netstat -s
```

## Getting Help

### Collect Diagnostics

```bash
# Server version
guocedb --version

# Configuration
cat /etc/guocedb/config.yaml

# Logs
tail -100 /var/log/guocedb/guocedb.log

# System info
uname -a
cat /etc/os-release

# Metrics snapshot
curl http://localhost:9090/metrics > metrics.txt
```

### Enable Debug Logging

```yaml
logging:
  level: debug
```

Restart server and reproduce issue.

### Report Issue

When reporting issues, include:

1. GuoceDB version
2. Operating system
3. Configuration (sanitized)
4. Error messages
5. Steps to reproduce
6. Relevant logs

Submit to: https://github.com/turtacn/guocedb/issues

## See Also

- [Deployment Guide](deployment.md)
- [Architecture Documentation](architecture.md)
- [SQL Reference](sql-reference.md)
