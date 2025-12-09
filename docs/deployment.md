# GuoceDB Deployment Guide

## Quick Start

### Binary Installation

```bash
# Download latest release
wget https://github.com/turtacn/guocedb/releases/latest/download/guocedb-linux-amd64.tar.gz
tar -xzf guocedb-linux-amd64.tar.gz
sudo mv guocedb /usr/local/bin/

# Create data directory
sudo mkdir -p /var/lib/guocedb
sudo chown $USER:$USER /var/lib/guocedb

# Start server
guocedb --data-dir /var/lib/guocedb --port 3306
```

### Docker Installation

```bash
docker run -d \
  --name guocedb \
  -p 3306:3306 \
  -p 9090:9090 \
  -v /data/guocedb:/data \
  turtacn/guocedb:latest
```

### Building from Source

```bash
git clone https://github.com/turtacn/guocedb.git
cd guocedb
make build
./bin/guocedb --help
```

## Configuration

### Configuration File

Create `/etc/guocedb/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 3306
  max_connections: 1000
  shutdown_timeout: 30s

storage:
  data_dir: /var/lib/guocedb/data
  sync_writes: true

security:
  enabled: true
  root_password: "${GUOCEDB_ROOT_PASSWORD}"

observability:
  enabled: true
  address: ":9090"

logging:
  level: info
  format: json
  output: /var/log/guocedb/guocedb.log
```

### Environment Variables

- `GUOCEDB_ROOT_PASSWORD` - Root password for authentication
- `GUOCEDB_DATA_DIR` - Data directory path
- `GUOCEDB_PORT` - MySQL protocol port
- `GUOCEDB_LOG_LEVEL` - Logging level (debug, info, warn, error)

### Command-Line Options

```bash
guocedb \
  --config /etc/guocedb/config.yaml \
  --data-dir /var/lib/guocedb \
  --port 3306 \
  --log-level info
```

## Production Deployment

### System Requirements

**Minimum:**
- CPU: 2 cores
- RAM: 4 GB
- Disk: 50 GB SSD
- OS: Linux (Ubuntu 20.04+, CentOS 7+)

**Recommended:**
- CPU: 8+ cores
- RAM: 16+ GB
- Disk: 500+ GB NVMe SSD
- OS: Linux (Ubuntu 22.04, CentOS 8)

### Systemd Service

Create `/etc/systemd/system/guocedb.service`:

```ini
[Unit]
Description=GuoceDB Database Server
After=network.target

[Service]
Type=simple
User=guocedb
Group=guocedb
ExecStart=/usr/local/bin/guocedb --config /etc/guocedb/config.yaml
Restart=on-failure
RestartSec=5s
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable guocedb
sudo systemctl start guocedb
sudo systemctl status guocedb
```

### Kubernetes Deployment

StatefulSet manifest:

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: guocedb
  namespace: database
spec:
  serviceName: guocedb
  replicas: 1
  selector:
    matchLabels:
      app: guocedb
  template:
    metadata:
      labels:
        app: guocedb
    spec:
      containers:
      - name: guocedb
        image: turtacn/guocedb:latest
        ports:
        - containerPort: 3306
          name: mysql
        - containerPort: 9090
          name: metrics
        env:
        - name: GUOCEDB_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              name: guocedb-secret
              key: root-password
        volumeMounts:
        - name: data
          mountPath: /data
        - name: config
          mountPath: /etc/guocedb
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
          initialDelaySeconds: 10
          periodSeconds: 5
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
          limits:
            memory: "8Gi"
            cpu: "4000m"
      volumes:
      - name: config
        configMap:
          name: guocedb-config
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: fast-ssd
      resources:
        requests:
          storage: 100Gi
```

Service manifest:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: guocedb
  namespace: database
spec:
  selector:
    app: guocedb
  ports:
  - port: 3306
    targetPort: 3306
    name: mysql
  - port: 9090
    targetPort: 9090
    name: metrics
  type: ClusterIP
```

## Monitoring

### Prometheus Integration

Add to Prometheus scrape config:

```yaml
scrape_configs:
  - job_name: 'guocedb'
    static_configs:
      - targets: ['guocedb:9090']
```

### Key Metrics

- `guocedb_connections_active` - Active connections
- `guocedb_queries_total` - Total queries processed
- `guocedb_query_duration_seconds` - Query latency
- `guocedb_transactions_total` - Transaction count
- `guocedb_storage_size_bytes` - Storage size

### Health Endpoints

```bash
# Liveness (is process alive?)
curl http://localhost:9090/live

# Readiness (ready for traffic?)
curl http://localhost:9090/ready

# Detailed health
curl http://localhost:9090/health
```

## Backup & Recovery

### Logical Backup (mysqldump)

```bash
mysqldump -h 127.0.0.1 -P 3306 -u root -p \
  --all-databases \
  --single-transaction \
  > backup-$(date +%Y%m%d-%H%M%S).sql
```

### Physical Backup

```bash
# Stop server
systemctl stop guocedb

# Copy data directory
tar -czf backup-$(date +%Y%m%d).tar.gz /var/lib/guocedb

# Start server
systemctl start guocedb
```

### Restore

```bash
# From logical backup
mysql -h 127.0.0.1 -P 3306 -u root -p < backup.sql

# From physical backup
systemctl stop guocedb
tar -xzf backup.tar.gz -C /
systemctl start guocedb
```

## Security

### TLS/SSL Configuration

```yaml
security:
  tls:
    enabled: true
    cert_file: /etc/guocedb/certs/server.crt
    key_file: /etc/guocedb/certs/server.key
    ca_file: /etc/guocedb/certs/ca.crt
```

### Firewall Configuration

```bash
# Allow MySQL port
sudo ufw allow 3306/tcp

# Allow metrics port (restrict to monitoring network)
sudo ufw allow from 10.0.0.0/8 to any port 9090
```

### User Management

```sql
-- Create application user
CREATE USER 'app'@'%' IDENTIFIED BY 'secure_password';

-- Grant permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON appdb.* TO 'app'@'%';

-- Verify grants
SHOW GRANTS FOR 'app'@'%';
```

## Performance Tuning

### Storage Configuration

```yaml
storage:
  sync_writes: false  # Faster but less durable
  max_table_size: 1073741824  # 1GB
  num_compactors: 4  # Parallel compaction
```

### Connection Pooling

Application-side configuration:

```go
db.SetMaxOpenConns(100)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)
```

### Query Optimization

1. Use appropriate indexes
2. Limit result set sizes
3. Use prepared statements
4. Batch INSERT operations in transactions

## Troubleshooting

See [Troubleshooting Guide](troubleshooting.md) for common issues and solutions.

## Upgrading

### Minor Version Upgrade

```bash
# Stop server
systemctl stop guocedb

# Backup data
tar -czf backup-pre-upgrade.tar.gz /var/lib/guocedb

# Replace binary
sudo cp guocedb-new /usr/local/bin/guocedb

# Start server
systemctl start guocedb

# Verify
mysql -h 127.0.0.1 -P 3306 -u root -p -e "SELECT VERSION()"
```

### Major Version Upgrade

Follow release notes for major version upgrades. May require schema migrations.

## Additional Resources

- [Architecture Documentation](architecture.md)
- [SQL Reference](sql-reference.md)
- [Troubleshooting Guide](troubleshooting.md)
- [GitHub Repository](https://github.com/turtacn/guocedb)
