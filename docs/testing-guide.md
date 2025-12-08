# GuoceDB Testing Guide

This guide provides comprehensive information about testing GuoceDB, including unit tests, integration tests, and performance testing.

## Table of Contents

- [Test Categories](#test-categories)
- [Running Tests](#running-tests)
- [Integration Test Framework](#integration-test-framework)
- [Writing New Tests](#writing-new-tests)
- [Test Data Management](#test-data-management)
- [Performance Testing](#performance-testing)
- [CI/CD Integration](#cicd-integration)
- [Troubleshooting](#troubleshooting)

## Test Categories

GuoceDB uses a multi-layered testing approach:

| Test Type | Location | Command | Purpose |
|-----------|----------|---------|---------|
| Unit Tests | `*_test.go` files | `go test ./...` | Test individual functions and components |
| Integration Tests | `integration/` | `go test ./integration/...` | Test end-to-end functionality |
| Compatibility Tests | `integration/compatibility_test.go` | `go test -run Compat` | Test MySQL protocol compatibility |
| Concurrent Tests | `integration/concurrent_test.go` | `go test -run Concurrent` | Test concurrent access and performance |
| Transaction Tests | `integration/transaction_test.go` | `go test -run Txn` | Test transaction isolation and ACID properties |

## Running Tests

### All Tests
```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run with race detection
go test ./... -race

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Tests Only
```bash
# Run all integration tests
go test ./integration/... -v

# Run specific test categories
go test -run TestSQL_ ./integration/...          # SQL functionality tests
go test -run TestTxn_ ./integration/...          # Transaction tests
go test -run TestConcurrent_ ./integration/...   # Concurrent tests
go test -run TestCompat_ ./integration/...       # Compatibility tests
```

### Performance and Stress Tests
```bash
# Run concurrent/stress tests (may take longer)
go test -run TestConcurrent_StressTest ./integration/... -v

# Skip stress tests in short mode
go test ./integration/... -short
```

### Test with Timeout
```bash
# Set timeout for long-running tests
go test ./integration/... -timeout 10m
```

## Integration Test Framework

The integration test framework provides a complete testing environment with automatic server lifecycle management.

### Basic Usage

```go
package integration

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMyFeature(t *testing.T) {
    // Create test server
    ts := NewTestServer(t)
    defer ts.Close()
    
    // Create test client
    client := NewTestClient(t, ts.DSN(""))
    defer client.Close()
    
    // Run your tests
    client.MustExec("CREATE DATABASE mydb")
    client.MustExec("USE mydb")
    client.MustExec("CREATE TABLE t (id INT PRIMARY KEY, name VARCHAR(50))")
    
    // Verify results
    AssertTableExists(t, client, "t")
}
```

### Framework Components

#### TestServer
- **Purpose**: Manages GuoceDB server lifecycle for testing
- **Features**: 
  - Automatic port allocation
  - Temporary data directories
  - Graceful startup/shutdown
  - Connection readiness detection

#### TestClient
- **Purpose**: Simplified database client for tests
- **Features**:
  - Automatic error handling with `require.NoError`
  - Batch query execution with `MustExec`
  - Connection pool management

#### Helper Functions
- `CollectRows(t, rows)` - Convert sql.Rows to []map[string]interface{}
- `AssertTableExists(t, client, tableName)` - Verify table existence
- `AssertDatabaseExists(t, client, dbName)` - Verify database existence
- `GetRowCount(t, client, tableName)` - Count rows in table
- `ExecSQLFile(t, client, path)` - Execute SQL script from file

## Writing New Tests

### Test Structure

Follow this structure for new integration tests:

```go
func TestFeature_SpecificBehavior(t *testing.T) {
    // 1. Setup
    ts := NewTestServer(t)
    defer ts.Close()
    
    client := NewTestClient(t, ts.DSN(""))
    defer client.Close()
    
    // 2. Prepare test data (if needed)
    ExecSQLFile(t, client, "testdata/schema.sql")
    ExecSQLFile(t, client, "testdata/seed.sql")
    
    // 3. Execute test operations
    // ... your test logic here ...
    
    // 4. Verify results
    // Use assert/require for validation
}
```

### Naming Conventions

- Test functions: `TestCategory_SpecificFeature`
- Categories: `SQL`, `Txn`, `Concurrent`, `Compat`
- Examples:
  - `TestSQL_CreateTable`
  - `TestTxn_ReadCommitted`
  - `TestConcurrent_MultiClient`
  - `TestCompat_PreparedStatement`

### Best Practices

1. **Use descriptive test names** that explain what is being tested
2. **Clean up resources** with `defer` statements
3. **Use helper functions** to reduce code duplication
4. **Test both success and failure cases**
5. **Include edge cases** and boundary conditions
6. **Use table-driven tests** for multiple similar scenarios

### Example: Table-Driven Test

```go
func TestSQL_DataTypes(t *testing.T) {
    ts := NewTestServer(t)
    defer ts.Close()
    
    client := NewTestClient(t, ts.DSN(""))
    defer client.Close()
    
    client.MustExec("CREATE DATABASE testdb", "USE testdb")
    
    tests := []struct {
        name     string
        colType  string
        value    interface{}
        expected interface{}
    }{
        {"Integer", "INT", 42, int64(42)},
        {"String", "VARCHAR(50)", "hello", "hello"},
        {"Boolean", "BOOLEAN", true, true},
        {"Decimal", "DECIMAL(10,2)", 123.45, 123.45},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tableName := fmt.Sprintf("test_%s", strings.ToLower(tt.name))
            client.Exec(fmt.Sprintf("CREATE TABLE %s (id INT PRIMARY KEY, val %s)", tableName, tt.colType))
            client.Exec(fmt.Sprintf("INSERT INTO %s (id, val) VALUES (1, ?)", tableName), tt.value)
            
            var result interface{}
            err := client.QueryRow(fmt.Sprintf("SELECT val FROM %s WHERE id = 1", tableName)).Scan(&result)
            require.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## Test Data Management

### Standard Test Schema

The integration tests use standardized test data located in `integration/testdata/`:

- **`schema.sql`**: Standard table definitions
  - `users` - User information
  - `orders` - Order data
  - `products` - Product catalog
  - `order_items` - Order line items
  - `categories` - Product categories

- **`seed.sql`**: Sample data for testing
  - 5 users with varying ages
  - 5 products with different prices
  - 5 orders with different statuses
  - Order items linking orders to products

### Using Test Data

```go
// Load standard schema and data
ExecSQLFile(t, client, "testdata/schema.sql")
ExecSQLFile(t, client, "testdata/seed.sql")

// Now you can test with realistic data
rows := client.Query("SELECT u.name, COUNT(o.id) FROM users u LEFT JOIN orders o ON u.id = o.user_id GROUP BY u.id")
results := CollectRows(t, rows)
// ... verify results ...
```

### Custom Test Data

For tests requiring specific data:

```go
func TestCustomScenario(t *testing.T) {
    ts := NewTestServer(t)
    defer ts.Close()
    
    client := NewTestClient(t, ts.DSN(""))
    defer client.Close()
    
    // Create custom schema for this test
    client.MustExec(
        "CREATE DATABASE testdb",
        "USE testdb",
        "CREATE TABLE custom_table (id INT PRIMARY KEY, data VARCHAR(100))",
    )
    
    // Insert specific test data
    for i := 1; i <= 100; i++ {
        client.Exec("INSERT INTO custom_table (id, data) VALUES (?, ?)", i, fmt.Sprintf("data-%d", i))
    }
    
    // Run your test...
}
```

## Performance Testing

### Concurrent Tests

Test concurrent access patterns:

```go
func TestConcurrent_MyFeature(t *testing.T) {
    ts := NewTestServer(t)
    defer ts.Close()
    
    // Setup
    setupClient := NewTestClient(t, ts.DSN(""))
    // ... setup code ...
    setupClient.Close()
    
    var wg sync.WaitGroup
    numClients := 10
    
    for i := 0; i < numClients; i++ {
        wg.Add(1)
        go func(clientID int) {
            defer wg.Done()
            
            client := NewTestClient(t, ts.DSN("testdb"))
            defer client.Close()
            
            // Perform concurrent operations
            // ...
        }(i)
    }
    
    wg.Wait()
    
    // Verify results
    // ...
}
```

### Stress Testing

For stress tests, use the `testing.Short()` flag:

```go
func TestStress_HighLoad(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }
    
    // ... stress test implementation ...
}
```

Run stress tests with:
```bash
# Run all tests including stress tests
go test ./integration/...

# Skip stress tests
go test ./integration/... -short
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Run Unit Tests
      run: go test ./... -short -race -coverprofile=coverage.out
    
    - name: Run Integration Tests
      run: go test ./integration/... -v
    
    - name: Upload Coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

### Test Requirements

For CI environments, ensure:

1. **Go version compatibility** (1.21+)
2. **Sufficient memory** for concurrent tests (2GB+ recommended)
3. **Timeout settings** for long-running tests
4. **Race detection** enabled
5. **Coverage reporting** configured

### Test Timeouts

Set appropriate timeouts for different test categories:

```bash
# Unit tests (fast)
go test ./compute/... -timeout 2m

# Integration tests (moderate)
go test ./integration/... -timeout 10m

# Stress tests (slow)
go test -run Stress ./integration/... -timeout 30m
```

## Troubleshooting

### Common Issues

#### 1. Port Conflicts
**Problem**: Test server fails to start due to port conflicts
**Solution**: The framework uses random ports, but if you see conflicts:
```go
// Check if server started properly
ts := NewTestServer(t)
defer ts.Close()
t.Logf("Test server running on: %s", ts.Addr())
```

#### 2. Connection Timeouts
**Problem**: Tests fail with connection timeouts
**Solution**: Increase timeout or check server readiness:
```go
// Add custom wait logic if needed
WaitForCondition(t, func() bool {
    conn, err := net.Dial("tcp", ts.Addr())
    if err == nil {
        conn.Close()
        return true
    }
    return false
}, 10*time.Second, "Server should be ready")
```

#### 3. Data Race Warnings
**Problem**: Race detector reports data races
**Solution**: Use proper synchronization:
```go
var mu sync.Mutex
var counter int64

// In goroutines:
mu.Lock()
counter++
mu.Unlock()

// Or use atomic operations:
atomic.AddInt64(&counter, 1)
```

#### 4. Test Data Conflicts
**Problem**: Tests interfere with each other's data
**Solution**: Use unique database names:
```go
func TestMyFeature(t *testing.T) {
    ts := NewTestServer(t)
    defer ts.Close()
    
    // Use test-specific database name
    dbName := fmt.Sprintf("test_%s_%d", t.Name(), time.Now().UnixNano())
    client := NewTestClient(t, ts.DSN(""))
    client.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
    client.Exec(fmt.Sprintf("USE %s", dbName))
    // ...
}
```

### Debug Mode

Enable debug logging for troubleshooting:

```bash
# Run with verbose output
go test ./integration/... -v

# Run specific test with debug info
go test -run TestSQL_CreateTable ./integration/... -v

# Enable race detection
go test ./integration/... -race
```

### Performance Debugging

Profile test performance:

```bash
# CPU profiling
go test ./integration/... -cpuprofile=cpu.prof

# Memory profiling  
go test ./integration/... -memprofile=mem.prof

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

## Test Coverage

### Generating Coverage Reports

```bash
# Generate coverage for all packages
go test ./... -coverprofile=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# View coverage by package
go tool cover -func=coverage.out
```

### Coverage Goals

- **Unit Tests**: >80% line coverage
- **Integration Tests**: >70% feature coverage
- **Critical Paths**: >95% coverage (transactions, data integrity)

### Excluding Files from Coverage

Use build tags to exclude test utilities:

```go
//go:build !coverage
// +build !coverage

package testutil

// Test utility functions that shouldn't count toward coverage
```

## Contributing Test Guidelines

When contributing new tests:

1. **Follow naming conventions** described above
2. **Add tests for new features** - every new feature should have corresponding tests
3. **Update existing tests** when modifying behavior
4. **Include both positive and negative test cases**
5. **Document complex test scenarios** with comments
6. **Ensure tests are deterministic** - avoid time-dependent or random behavior
7. **Keep tests focused** - one test should verify one specific behavior
8. **Use appropriate test categories** - unit vs integration vs compatibility

### Test Review Checklist

Before submitting tests:

- [ ] Tests have descriptive names
- [ ] Tests clean up resources properly
- [ ] Tests handle errors appropriately
- [ ] Tests are deterministic and repeatable
- [ ] Tests cover both success and failure cases
- [ ] Tests follow the project's coding standards
- [ ] Tests run successfully in CI environment
- [ ] Tests don't have unnecessary dependencies
- [ ] Tests are properly categorized
- [ ] Tests include appropriate documentation

---

For more information about GuoceDB testing, see:
- [Development Guide](development-guide.md)
- [Architecture Overview](architecture.md)
- [API Documentation](api-reference.md)