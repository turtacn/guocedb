// test/integration/integration_test.go

// Package integration contains integration tests for Guocedb.
// Integration tests verify the interaction between different components
// and typically require a running Guocedb server instance.
//
// integration 包包含 Guocedb 的集成测试。
// 集成测试验证不同组件之间的交互，通常需要运行中的 Guocedb 服务器实例。
package integration

import (
	"context"
	"database/sql" // Standard Go SQL package for interacting with the running server
	"fmt"
	"net/http"
	"os"
	"testing" // Standard Go testing library
	"time"

	// Import the MySQL driver to interact with the server via MySQL protocol.
	// 导入 MySQL 驱动程序，通过 MySQL 协议与服务器交互。
	_ "github.com/go-sql-driver/mysql"

	// Import necessary Guocedb packages for setting up the test environment (e.g., config, server, engine).
	// 导入必要的 Guocedb 包以设置测试环境（例如，config, server, engine）。
	"github.com/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/engine"
	"github.com/turtacn/guocedb/network/server"
	"github.com/turtacn/guocedb/common/types/enum" // For config keys
)

// TestMain is a special function that runs before any tests in the package.
// We can use it to set up and tear down the database server for integration tests.
//
// TestMain 是一个特殊函数，在包中的任何测试运行之前运行。
// 我们可以使用它来为集成测试设置和拆除数据库服务器。
func TestMain(m *testing.M) {
	// Set up the database server for integration tests.
	// This involves:
	// 1. Loading test configuration.
	// 2. Initializing the database engine.
	// 3. Initializing and starting the network server (e.g., MySQL).
	//
	// 为集成测试设置数据库服务器。
	// 这包括：
	// 1. 加载测试配置。
	// 2. 初始化数据库引擎。
	// 3. 初始化并启动网络服务器（例如，MySQL）。

	log.InitLogger(map[string]string{"log_level": "info"}) // Basic logging for setup/teardown / 为设置/拆除进行基本日志记录
	log.Info("Setting up Guocedb server for integration tests...") // 正在为集成测试设置 Guocedb 服务器...

	// Load a test configuration (might be a modified version of config.yaml.example)
	// 加载测试配置（可能是 config.yaml.example 的修改版本）
	// Use a temporary directory for storage paths to avoid interfering with development data.
	// 使用临时目录作为存储路径，避免干扰开发数据。
	testConfigPath := "./test_config.yaml" // Create a temporary test config if needed
	// For now, let's create a simple config struct directly for testing.
	// 目前，我们直接创建一个简单的 config 结构体用于测试。
	cfg := config.NewConfig()
	// Set test-specific config values
	// 设置特定于测试的配置值
	cfg.Set(enum.ConfigKeyMySQLBindHost, "127.0.0.1")
	cfg.Set(enum.ConfigKeyMySQLBindPort, 13306) // Use a different port than default MySQL
	cfg.Set(enum.ConfigKeyStorageEngine, "badger")
	cfg.Set(enum.ConfigKeyStorageBadgerDataPath, "./testdata/badger_data") // Test-specific data path
	cfg.Set(enum.ConfigKeyStorageBadgerWalPath, "./testdata/badger_wal")   // Test-specific WAL path
	cfg.Set(enum.ConfigKeyStorageBadgerSyncWrites, false)                  // Faster tests, less durability
	cfg.Set(enum.ConfigKeyLoggingLevel, "debug")                           // Verbose logging during tests

	// Ensure test data directories are clean
	// 确保测试数据目录是干净的
	os.RemoveAll(cfg.GetString(enum.ConfigKeyStorageBadgerDataPath))
	os.RemoveAll(cfg.GetString(enum.ConfigKeyStorageBadgerWalPath))


	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize the database engine
	// 初始化数据库引擎
	eng, err := engine.NewEngine(ctx, cfg)
	if err != nil {
		log.Fatal("Failed to initialize database engine for integration tests: %v", err) // 初始化数据库引擎失败。
	}
	// Defer stopping the engine until TestMain exits.
	// 延迟停止引擎直到 TestMain 退出。
	defer func() {
		log.Info("Stopping database engine after integration tests.") // 集成测试后停止数据库引擎。
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if stopErr := eng.Stop(shutdownCtx); stopErr != nil {
			log.Error("Failed to stop engine after integration tests: %v", stopErr) // 集成测试后停止引擎失败。
		}
		// Clean up test data directories
		// 清理测试数据目录
		os.RemoveAll(cfg.GetString(enum.ConfigKeyStorageBadgerDataPath))
		os.RemoveAll(cfg.GetString(enum.ConfigKeyStorageBadgerWalPath))
	}()

	// Initialize and start the network server
	// 初始化并启动网络服务器
	netServer, err := server.NewServer(ctx, cfg, eng) // Pass config and engine
	if err != nil {
		log.Fatal("Failed to initialize network server for integration tests: %v", err) // 初始化网络服务器失败。
	}
	// Start the network server in a goroutine. It will block until ctx is cancelled or error.
	// 在 goroutine 中启动网络服务器。它将阻塞直到 ctx 被取消或出错。
	go func() {
		if startErr := netServer.Start(); startErr != nil && startErr != http.ErrServerClosed {
			log.Error("Network server failed during integration tests: %v", startErr) // 网络服务器失败。
			// Consider signaling a fatal error to the main TestMain goroutine if necessary.
			// 如果需要，考虑向 TestMain 主 goroutine 发送致命错误信号。
		}
	}()
	log.Info("Guocedb server started in background for integration tests.") // 集成测试 Guocedb 服务器已在后台启动。

	// Wait a moment for the server to start listening.
	// 等待服务器开始监听片刻。
	time.Sleep(100 * time.Millisecond) // Adjust as needed / 根据需要调整


	// Run the tests
	// 运行测试
	exitCode := m.Run() // This runs all the Test* functions in the package. / 这运行包中的所有 Test* 函数。


	// Tear down the database server.
	// This involves stopping the network server and the database engine.
	//
	// 拆除数据库服务器。
	// 这包括停止网络服务器和数据库引擎。
	log.Info("Tearing down Guocedb server after integration tests.") // 集成测试后拆除 Guocedb 服务器。

	// Stop the network server.
	// 停止网络服务器。
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if stopErr := netServer.Stop(shutdownCtx); stopErr != nil { // Pass context to Stop
		log.Error("Failed to stop network server after integration tests: %v", stopErr) // 集成测试后停止网络服务器失败。
		// Continue with engine stop even if network server stop failed.
		// 即使网络服务器停止失败，也继续进行引擎停止。
	} else {
		log.Info("Network server stopped after integration tests.") // 集成测试后网络服务器已停止。
	}

	// Engine cleanup is handled by the deferred eng.Stop() call above.
	// 引擎清理由上面延迟的 eng.Stop() 调用处理。

	// Exit with the code from running the tests.
	// 使用运行测试的退出码退出。
	os.Exit(exitCode)
}

// getTestDBConnection establishes a connection to the running test server.
// getTestDBConnection 与运行中的测试服务器建立连接。
func getTestDBConnection(t *testing.T) *sql.DB {
	// Get the test server address from the test config (same config used in TestMain).
	// 从测试配置获取测试服务器地址（与 TestMain 中使用的配置相同）。
	// Access the config object used in TestMain? Or read it again?
	// Accessing the config directly from TestMain is cleaner but requires passing it or making it global.
	// Let's assume the test config is globally accessible or can be re-read.
	// For simplicity, hardcode the test port used in TestMain config for now.
	//
	// 访问 TestMain 中直接使用的 config 对象？还是重新读取？
	// 直接从 TestMain 访问 config 更清晰，但需要传递它或使其成为全局。
	// 为了简单，暂时硬编码 TestMain config 中使用的测试端口。
	testHost := "127.0.0.1"
	testPort := 13306 // Must match the port set in TestMain config.

	// Construct the DSN for the test connection.
	// 构造测试连接的 DSN。
	// Use dummy user/password defined in the PlaceholderAuthenticator/test config.
	// 使用在 PlaceholderAuthenticator/测试配置中定义的虚拟用户/密码。
	dsn := fmt.Sprintf("root:password@tcp(%s:%d)/", testHost, testPort) // Connect without selecting a database initially.

	log.Debug("Connecting to test database with DSN: tcp(%s:%d)", testHost, testPort) // 连接到测试数据库。

	// Open database connection
	// 开启数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("Failed to open test database connection: %v", err) // 开启测试数据库连接失败。
	}

	// Ping the database to verify the connection
	// ping 数据库以验证连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		db.Close() // Close connection if ping fails
		t.Fatalf("Failed to ping test database: %v", err) // ping 测试数据库失败。
	}

	log.Debug("Test database connection verified.") // 测试数据库连接已验证。
	return db // Return the opened database connection
}

// TestExampleIntegration is a placeholder for an integration test.
// It connects to the running server and performs a simple operation.
//
// TestExampleIntegration 是集成测试的占位符。
// 它连接到运行中的服务器并执行简单操作。
func TestExampleIntegration(t *testing.T) {
	t.Skip("Placeholder integration test - implement actual tests.") // Skip this placeholder test by default. # 占位符集成测试 - 实现实际测试。

	// Get a connection to the running test server.
	// 获取与运行中的测试服务器的连接。
	db := getTestDBConnection(t) // t.Fatalf if connection fails
	defer db.Close() // Ensure connection is closed after the test.


	// Example Integration Test: Create a database, use it, create a table, insert data, select data, drop table, drop database.
	// 示例集成测试：创建一个数据库，使用它，创建一个表，插入数据，选择数据，删除表，删除数据库。

	testDBName := "integration_test_db"
	testTableName := "test_table"

	ctx := context.Background() // Use background context for test operations

	// 1. Drop database if it exists from previous runs
	// 1. 如果数据库存在，从之前的运行中删除它
	log.Info("Integration Test: Dropping database '%s' if it exists.", testDBName) // 集成测试：如果数据库存在，删除它。
	_, err := db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	if err != nil {
		t.Fatalf("Failed to drop database '%s': %v", testDBName, err) // 删除数据库失败。
	}
	log.Info("Integration Test: Database '%s' dropped (if existed).", testDBName) // 集成测试：数据库 '%s' 已删除（如果存在）。

	// 2. Create database
	// 2. 创建数据库
	log.Info("Integration Test: Creating database '%s'.", testDBName) // 集成测试：正在创建数据库 '%s'。
	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", testDBName))
	if err != nil {
		t.Fatalf("Failed to create database '%s': %v", testDBName, err) // 创建数据库失败。
	}
	log.Info("Integration Test: Database '%s' created.", testDBName) // 集成测试：数据库 '%s' 已创建。

	// 3. Use the created database
	// 3. 使用创建的数据库
	log.Info("Integration Test: Using database '%s'.", testDBName) // 集成测试：正在使用数据库 '%s'。
	_, err = db.ExecContext(ctx, fmt.Sprintf("USE %s", testDBName))
	if err != nil {
		t.Fatalf("Failed to use database '%s': %v", testDBName, err) // 使用数据库失败。
	}
	log.Info("Integration Test: Using database '%s' successful.", testDBName) // 集成测试：使用数据库 '%s' 成功。


	// 4. Create a table
	// 4. 创建一个表
	log.Info("Integration Test: Creating table '%s'.", testTableName) // 集成测试：正在创建表 '%s'。
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE %s (
			id INT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			age INT
		)`, testTableName)
	_, err = db.ExecContext(ctx, createTableSQL)
	if err != nil {
		t.Fatalf("Failed to create table '%s': %v", testTableName, err) // 创建表失败。
	}
	log.Info("Integration Test: Table '%s' created.", testTableName) // 集成测试：表 '%s' 已创建。


	// 5. Insert data
	// 5. 插入数据
	log.Info("Integration Test: Inserting data into '%s'.", testTableName) // 集成测试：正在插入数据到 '%s'。
	insertDataSQL := fmt.Sprintf("INSERT INTO %s (id, name, age) VALUES (?, ?, ?)", testTableName)
	result, err := db.ExecContext(ctx, insertDataSQL, 1, "Alice", 30)
	if err != nil {
		t.Fatalf("Failed to insert data into '%s': %v", testTableName, err) // 插入数据失败。
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected != 1 {
		t.Errorf("Expected 1 row affected by insert, got %d", rowsAffected) // 插入操作应影响 1 行，实际影响 %d。
	}
	log.Info("Integration Test: Data inserted into '%s'.", testTableName) // 集成测试：数据已插入到 '%s'。


	// 6. Select data
	// 6. 选择数据
	log.Info("Integration Test: Selecting data from '%s'.", testTableName) // 集成测试：正在从 '%s' 选择数据。
	selectDataSQL := fmt.Sprintf("SELECT id, name, age FROM %s WHERE id = ?", testTableName)
	rows, err := db.QueryContext(ctx, selectDataSQL, 1)
	if err != nil {
		t.Fatalf("Failed to select data from '%s': %v", testTableName, err) // 选择数据失败。
	}
	defer rows.Close()

	// Verify selected data
	// 验证选择的数据
	if !rows.Next() {
		t.Error("Expected to find data for id 1, but found none.") // 期望找到 id 为 1 的数据，但未找到。
	}
	var id, age int
	var name string
	err = rows.Scan(&id, &name, &age)
	if err != nil {
		t.Fatalf("Failed to scan selected row: %v", err) // 扫描选择的行失败。
	}
	if id != 1 || name != "Alice" || age != 30 {
		t.Errorf("Selected data mismatch: got id=%d, name=%s, age=%d, expected id=1, name=Alice, age=30", id, name, age) // 选择的数据不匹配。
	}
	if rows.Next() {
		t.Error("Expected only one row for id 1, found more.") // 期望 id 为 1 的只有一行，找到更多。
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("Error during row iteration: %v", err) // 行迭代期间出错。
	}
	log.Info("Integration Test: Data selected and verified from '%s'.", testTableName) // 集成测试：数据已选择和验证从 '%s'。


	// 7. Drop table
	// 7. 删除表
	log.Info("Integration Test: Dropping table '%s'.", testTableName) // 集成测试：正在删除表 '%s'。
	_, err = db.ExecContext(ctx, fmt.Sprintf("DROP TABLE %s", testTableName))
	if err != nil {
		t.Fatalf("Failed to drop table '%s': %v", testTableName, err) // 删除表失败。
	}
	log.Info("Integration Test: Table '%s' dropped.", testTableName) // 集成测试：表 '%s' 已删除。


	// 8. Drop database
	// 8. 删除数据库
	log.Info("Integration Test: Dropping database '%s'.", testDBName) // 集成测试：正在删除数据库 '%s'。
	_, err = db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE %s", testDBName))
	if err != nil {
		t.Fatalf("Failed to drop database '%s': %v", testDBName, err) // 删除数据库失败。
	}
	log.Info("Integration Test: Database '%s' dropped.", testDBName) // 集成测试：数据库 '%s' 已删除。


	log.Info("Integration Test: Basic DDL/DML/SELECT test completed successfully.") // 集成测试：基本 DDL/DML/SELECT 测试成功完成。

	// TODO: Add more integration tests covering different SQL features, error cases, transactions, etc.
	// Remember to clean up any created resources (databases, tables) in each test or TestMain teardown.
	// TODO: 添加更多集成测试，覆盖不同的 SQL 特性、错误情况、事务等。
	// 记住在每个测试或 TestMain 拆除中清理任何创建的资源（数据库、表）。
}

// Add more Test* functions here for different integration scenarios.
// 在此处为不同的集成场景添加更多 Test* 函数。