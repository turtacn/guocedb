// cmd/guocedb-cli/main.go

// The main package for the Guocedb command-line client executable.
// It connects to a running Guocedb server and executes SQL queries.
//
// Guocedb 命令行客户端可执行文件的主包。
// 它连接到运行中的 Guocedb 服务器并执行 SQL 查询。
package main

import (
	"context"
	"database/sql" // Standard Go SQL package
	"flag"
	"fmt"
	"os"
	"strings" // For joining query parts
	"time"

	// Import the MySQL driver. The underscore means we import it for its side effects
	// (registering itself with database/sql), but don't use any of its exported names directly.
	//
	// 导入 MySQL 驱动程序。下划线表示我们导入它是为了其副作用
	// （向 database/sql 注册自己），但不直接使用其任何导出的名称。
	_ "github.com/go-sql-driver/mysql"

	"github.com/turtacn/guocedb/common/log" // Use the common logging system
)

var (
	// Connection flags
	// 连接标志
	host     = flag.String("h", "127.0.0.1", "Database host") // 数据库主机
	port     = flag.Int("P", 3306, "Database port")         // 数据库端口 (default MySQL port)
	user     = flag.String("u", "root", "Database user")      // 数据库用户
	password = flag.String("p", "password", "Database password") // 数据库密码
	database = flag.String("D", "", "Database name to use") // 要使用的数据库名

	// Query flag
	// 查询标志
	query = flag.String("e", "", "Execute a single query and exit") // 执行单个查询并退出

	// Other flags
	// 其他标志
	verbose = flag.Bool("v", false, "Enable verbose logging") // 启用详细日志记录
)

func main() {
	flag.Parse() // Parse command-line flags / 解析命令行标志

	// Initialize basic logging (can be configured further if needed)
	// 初始化基本日志记录（如果需要可以进一步配置）
	// For a simple CLI, basic console logging is usually sufficient.
	// 对于简单的 CLI，基本控制台日志通常就足够了。
	// Use the common log package, maybe configure level based on -v flag.
	// 使用通用 log 包，可能根据 -v 标志配置级别。
	if *verbose {
		log.InitLogger(map[string]string{"log_level": "debug"}) // Set log level to debug if verbose
	} else {
		log.InitLogger(map[string]string{"log_level": "info"}) // Default to info level
	}
	log.Info("Guocedb CLI starting...") // Guocedb CLI 正在启动...


	// Construct the DSN (Data Source Name) for the MySQL connection.
	// Refer to go-sql-driver/mysql documentation for DSN format.
	//
	// 构造 MySQL 连接的 DSN (数据源名称)。
	// 参考 go-sql-driver/mysql 文档了解 DSN 格式。
	// Format: [user[:password]@][net[(addr)]]/dbname[?param1=value1&paramN=valueN]
	// 格式：[用户[:密码]@][网络[(地址)]]/数据库名[?参数1=值1&参数N=值N]
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		*user,
		*password,
		*host,
		*port,
		*database, // Use the database flag value
	)

	log.Debug("Connecting to database with DSN: tcp(%s:%d)/%s (user: %s)", *host, *port, *database, *user) // 使用 DSN 连接到数据库。

	// Open database connection
	// 开启数据库连接
	// The "mysql" driver is registered by the underscore import above.
	// "mysql" 驱动程序由上面的下划线导入注册。
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Error("Failed to open database connection: %v", err) // 开启数据库连接失败。
		fmt.Fprintf(os.Stderr, "Error: Failed to open database connection: %v\n", err) // 输出错误到 stderr
		os.Exit(1) // Exit on connection error / 连接错误时退出
	}
	// Ensure the connection is closed when main exits.
	// 确保在 main 退出时关闭连接。
	defer db.Close()
	log.Info("Database connection opened.") // 数据库连接已开启。


	// Ping the database to verify the connection
	// ping 数据库以验证连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Use a context with timeout for ping
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		log.Error("Failed to ping database: %v", err) // ping 数据库失败。
		fmt.Fprintf(os.Stderr, "Error: Failed to ping database: %v\n", err) // 输出错误到 stderr
		os.Exit(1) // Exit on ping failure / ping 失败时退出
	}
	log.Info("Database connection verified.") // 数据库连接已验证。


	// Execute the query if the -e flag is provided
	// 如果提供了 -e 标志，执行查询
	if *query != "" {
		log.Debug("Executing query: %s", *query) // 正在执行查询。
		err = executeQuery(db, *query) // Call a helper function to execute and display results
		if err != nil {
			log.Error("Query execution failed: %v", err) // 查询执行失败。
			fmt.Fprintf(os.Stderr, "Error: Query execution failed: %v\n", err) // 输出错误到 stderr
			os.Exit(1) // Exit on query execution failure / 查询执行失败时退出
		}
		log.Debug("Query execution finished.") // 查询执行完成。

	} else {
		// TODO: Implement interactive mode if no -e flag is provided.
		// This would involve reading commands from standard input.
		// For now, just exit if no query is provided.
		//
		// TODO: 如果没有提供 -e 标志，实现交互模式。
		// 这将涉及从标准输入读取命令。
		// 目前，如果未提供查询，则直接退出。
		log.Info("No query provided. Use -e to execute a query.") // 未提供查询。使用 -e 执行查询。
		// Implement interactive mode here in a real CLI
		// 在真实的 CLI 中在此处实现交互模式
		// fmt.Println("Interactive mode not implemented. Exiting.")
	}


	log.Info("Guocedb CLI finished.") // Guocedb CLI 已完成。
	os.Exit(0) // Exit successfully / 成功退出
}

// executeQuery executes a single SQL query and prints the results.
// executeQuery 执行单个 SQL 查询并打印结果。
func executeQuery(db *sql.DB, query string) error {
	// Execute the query. Use QueryContext for SELECT and ExecContext for INSERT/UPDATE/DELETE/DDL.
	// Since we don't know the query type, use a method that handles both, or try QueryContext first.
	//
	// 执行查询。SELECT 使用 QueryContext，INSERT/UPDATE/DELETE/DDL 使用 ExecContext。
	// 由于我们不知道查询类型，使用一个既能处理两者的方法，或者先尝试 QueryContext。

	// For simplicity, let's try QueryContext first, assuming most CLI usage is SELECT.
	// If it's not a SELECT, we might get an error or empty result set.
	// A real CLI parses the query type or uses ExecContext for non-SELECT.
	//
	// 为了简化，先尝试 QueryContext，假设大多数 CLI 用法是 SELECT。
	// 如果它不是 SELECT，我们可能会得到错误或空结果集。
	// 真实的 CLI 会解析查询类型或对非 SELECT 使用 ExecContext。

	log.Debug("Attempting to execute query using db.QueryContext...") // 尝试使用 db.QueryContext 执行查询。
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Use a context with timeout for query execution
	defer cancel()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		// Check if it's a non-SELECT query and try ExecContext
		// 检查是否是非 SELECT 查询并尝试 ExecContext
		// Simple check: Does the error message indicate a non-SELECT? Or check query prefix?
		// 简单的检查：错误消息是否指示非 SELECT？还是检查查询前缀？
		// A robust CLI would parse the query. For now, a basic check or rely on error messages.
		// 健壮的 CLI 会解析查询。目前，进行基本检查或依赖错误消息。
		upperQuery := strings.ToUpper(strings.TrimSpace(query))
		if strings.HasPrefix(upperQuery, "INSERT") ||
			strings.HasPrefix(upperQuery, "UPDATE") ||
			strings.HasPrefix(upperQuery, "DELETE") ||
			strings.HasPrefix(upperQuery, "CREATE") ||
			strings.HasPrefix(upperQuery, "DROP") ||
			strings.HasPrefix(upperQuery, "ALTER") ||
			strings.HasPrefix(upperQuery, "USE") || // Handle USE separately if needed
			strings.HasPrefix(upperQuery, "SET") { // Handle SET separately if needed

			log.Debug("Query seems to be non-SELECT. Trying db.ExecContext...") // 查询似乎是非 SELECT。正在尝试 db.ExecContext。
			// Non-SELECT query, use ExecContext
			// 非 SELECT 查询，使用 ExecContext
			result, execErr := db.ExecContext(ctx, query)
			if execErr != nil {
				log.Error("db.ExecContext failed: %v", execErr) // db.ExecContext 失败。
				return fmt.Errorf("query execution failed: %w", execErr)
			}
			// For DML, print affected rows/last insert ID
			// 对于 DML，打印受影响的行数/最后插入的 ID
			affected, _ := result.RowsAffected() // Ignore error for simplicity
			lastInsertID, _ := result.LastInsertId() // Ignore error
			fmt.Printf("Query executed successfully. Rows affected: %d, Last insert ID: %d\n", affected, lastInsertID) // 查询执行成功。
			return nil // Successfully executed non-SELECT query
		}

		// If error still exists and it was not handled as a non-SELECT, return the original error.
		// 如果错误仍然存在且未作为非 SELECT 处理，则返回原始错误。
		log.Error("db.QueryContext failed and not identified as non-SELECT: %v", err) // db.QueryContext 失败且未识别为非 SELECT。
		return fmt.Errorf("query execution failed: %w", err)
	}
	// Ensure rows are closed when the function exits.
	// 确保在函数退出时关闭行。
	defer rows.Close()
	log.Debug("db.QueryContext returned rows.") // db.QueryContext 返回行。


	// Get column names
	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		log.Error("Failed to get column names: %v", err) // 获取列名失败。
		return fmt.Errorf("failed to get column names: %w", err)
	}

	// Print column names (basic formatting)
	// 打印列名（基本格式）
	fmt.Println(strings.Join(columns, "\t")) // Use tab as separator

	// Print rows
	// 打印行
	// Create a slice of interface{} to hold values for scanning.
	// 创建一个 interface{} 切片以保存用于扫描的值。
	values := make([]interface{}, len(columns))
	// Create a slice of sql.RawBytes to scan into. RawBytes avoids type conversion issues initially.
	// 创建一个 sql.RawBytes 切片用于扫描。RawBytes 最初避免类型转换问题。
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i] // Scan into pointers to values
	}

	for rows.Next() {
		// Scan the row into the values slice
		// 将行扫描到 values 切片中
		err = rows.Scan(scanArgs...)
		if err != nil {
			log.Error("Failed to scan row: %v", err) // 扫描行失败。
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Print values in the row
		// 打印行中的值
		rowValues := make([]string, len(columns))
		for i, col := range values {
			// Handle NULL values
			// 处理 NULL 值
			if col == nil {
				rowValues[i] = "NULL"
			} else {
				// Format other value types as strings
				// 将其他值类型格式化为字符串
				switch v := col.(type) {
				case []byte:
					rowValues[i] = string(v) // Convert bytes to string (assuming UTF-8)
				case time.Time:
					rowValues[i] = v.Format("2006-01-02 15:04:05") // Format time
				default:
					rowValues[i] = fmt.Sprintf("%v", v) // Default formatting
				}
			}
		}
		fmt.Println(strings.Join(rowValues, "\t")) // Print row values
	}

	// Check for errors after iterating through rows
	// 遍历行后检查错误
	if err = rows.Err(); err != nil {
		log.Error("Error during row iteration: %v", err) // 行迭代期间出错。
		return fmt.Errorf("error during row iteration: %w", err)
	}

	return nil // Query executed and results printed successfully
}

// TODO: Implement interactive mode if no -e flag is provided.
// This would involve reading lines from os.Stdin, handling multi-line statements,
// and looping until the user types 'exit' or 'quit'.
//
// TODO: 如果未提供 -e 标志，实现交互模式。
// 这将涉及从 os.Stdin 读取行、处理多行语句，
// 并循环直到用户输入 'exit' 或 'quit'。
// Example structure for interactive mode:
//
// func runInteractive(db *sql.DB) {
//     reader := bufio.NewReader(os.Stdin)
//     fmt.Print("guocedb> ")
//     for {
//         input, _ := reader.ReadString('\n')
//         input = strings.TrimSpace(input)
//
//         if input == "" {
//             fmt.Print("guocedb> ")
//             continue
//         }
//
//         if strings.EqualFold(input, "exit") || strings.EqualFold(input, "quit") {
//             break // Exit loop
//         }
//
//         // Handle multi-line statements (more complex)
//         // 处理多行语句（更复杂）
//         query := input // For single line simplicity
//
//         err := executeQuery(db, query)
//         if err != nil {
//             fmt.Fprintf(os.Stderr, "Error: %v\n", err)
//         }
//         fmt.Print("guocedb> ")
//     }
// }