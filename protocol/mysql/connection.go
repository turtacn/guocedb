// Package mysql provides the MySQL server protocol implementation.
// connection.go manages a single client connection lifecycle.
//
// mysql 包提供了 MySQL 服务器协议实现。
// connection.go 管理单个客户端连接生命周期。
package mysql

import (
	"context"
	"fmt"
	"io" // For EOF
	"sync"

	"github.com/dolthub/go-mysql-server/server/mysql" // Import GMS MySQL server types
	"github.com/dolthub/go-mysql-server/sql"          // Import GMS SQL types
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	compute_catalog "github.com/turtacn/guocedb/compute/catalog"       // Import compute catalog
	compute_executor "github.com/turtacn/guocedb/compute/executor"     // Import compute executor
	compute_parser "github.com/turtacn/guocedb/compute/parser"         // Import compute parser
	compute_analyzer "github.com/turtacub/guocedb/compute/analyzer" // Import compute analyzer
	compute_optimizer "github.com/turtacub/guocedb/compute/optimizer" // Import compute optimizer
	compute_transaction "github.com/turtacn/guocedb/compute/transaction" // Import compute transaction manager
	"github.com/turtacn/guocedb/interfaces" // Import storage transaction interface
)

// Connection represents a single client connection to the MySQL server.
// It handles reading commands, processing queries, and sending responses.
//
// Connection 表示与 MySQL 服务器的单个客户端连接。
// 它处理读取命令、处理查询和发送响应。
type Connection struct {
	// conn is the underlying go-mysql-server connection object.
	// conn 是底层的 go-mysql-server 连接对象。
	conn *mysql.Conn

	// components provides access to the compute layer components.
	// components 提供对计算层组件的访问。
	components *ComputeComponents

	// session is the go-mysql-server sql.Session associated with this connection.
	// session 是与此连接关联的 go-mysql-server sql.Session。
	session sql.Session

	// transaction is the current transaction for this connection.
	// transaction 是此连接的当前事务。
	transaction interfaces.Transaction

	mu sync.Mutex // Protects access to mutable state like transaction
}

// ComputeComponents holds references to the necessary compute layer components.
// ComputeComponents 保存对必要计算层组件的引用。
type ComputeComponents struct {
	Parser compute_parser.Parser
	Analyzer compute_analyzer.Analyzer
	Optimizer compute_optimizer.Optimizer
	Executor compute_executor.Executor
	Catalog compute_catalog.Catalog
	TxManager compute_transaction.TransactionManager
}


// NewConnection creates a new Connection instance.
// It wraps the go-mysql-server connection and initializes the session and transaction.
//
// NewConnection 创建一个新的 Connection 实例。
// 它包装了 go-mysql-server 连接并初始化会话和事务。
func NewConnection(gmsConn *mysql.Conn, components *ComputeComponents) (*Connection, error) {
	log.Info("New connection established (ID: %d)", gmsConn.ConnectionID) // 建立新连接。

	// Create a new GMS sql.Session for this connection.
	// The session holds connection-specific state (e.g., current database, variables).
	// We need a custom sql.Session implementation that uses our compute components.
	//
	// 为此连接创建一个新的 GMS sql.Session。
	// 会话保存连接特定状态（例如，当前数据库、变量）。
	// 我们需要一个使用我们的计算组件的自定义 sql.Session 实现。
	session, err := NewGuocedbSession(gmsConn.ConnectionID, components.Catalog, components.TxManager) // Pass relevant components
	if err != nil {
		log.Error("Failed to create Guocedb session for connection ID %d: %v", gmsConn.ConnectionID, err) // 创建 Guocedb 会话失败。
		return nil, fmt.Errorf("failed to create session: %w", err)
	}


	// Begin an initial transaction for the connection.
	// MySQL typically operates within a transaction context.
	// 开启连接的初始事务。
	// MySQL 通常在事务 context 中运行。
	// Use context.Background() for now, as the request context starts later.
	// For a real system, consider how context flows.
	//
	// 目前使用 context.Background()，因为请求 context 稍后才开始。
	// 对于实际系统，考虑 context 如何流动。
	initialTx, err := components.TxManager.Begin(context.Background())
	if err != nil {
		log.Error("Failed to begin initial transaction for connection ID %d: %v", gmsConn.ConnectionID, err) // 开启连接的初始事务失败。
		// Clean up session?
		return nil, fmt.Errorf("failed to begin initial transaction: %w", err)
	}
	log.Debug("Initial transaction begun for connection ID %d", gmsConn.ConnectionID) // 连接的初始事务已开启。


	conn := &Connection{
		conn:       gmsConn,
		components: components,
		session:    session,
		transaction: initialTx,
	}

	// The GMS sql.Session needs a reference back to the connection's transaction.
	// We should set the transaction in the session after creating it.
	//
	// GMS sql.Session 需要引用回连接的事务。
	// 我们应该在创建会话后在会话中设置事务。
	// Our GuocedbSession implementation should handle this.
	// 我们的 GuocedbSession 实现应该处理这个问题。
	// Let's assume GuocedbSession has a method SetTransaction or it's set during creation.
	// 假设 GuocedbSession 有一个 SetTransaction 方法，或在创建期间设置。
	// If NewGuocedbSession takes TxManager, it can begin the transaction itself? No, TxManager belongs to ComputeComponents.
	// The Session interface has a GetTransaction method. The Handler will set the transaction in the sql.Context.
	// Let's rely on the Handler to set the correct transaction in the sql.Context before analysis/execution.
	// The `conn.transaction` field here is just for our internal tracking and commitment/rollback.
	//
	// 如果 NewGuocedbSession 接受 TxManager，它能否自己开始事务？不，TxManager 属于 ComputeComponents。
	// Session 接口有一个 GetTransaction 方法。Handler 将在 sql.Context 中设置事务。
	// 在分析/执行之前，让我们依靠 Handler 在 sql.Context 中设置正确的事务。
	// 此处的 `conn.transaction` 字段仅用于我们的内部跟踪和提交/回滚。


	// Start a goroutine to handle commands for this connection
	// 启动一个 goroutine 处理此连接的命令
	go conn.handleConnection()

	return conn, nil
}

// handleConnection is the main loop for processing commands from a client connection.
// handleConnection 是处理来自客户端连接命令的主循环。
func (c *Connection) handleConnection() {
	ctx := context.Background() // Use background context for the connection lifetime

	defer func() {
		// Ensure the connection is closed and resources are cleaned up
		// 确保连接关闭并清理资源
		log.Info("Closing connection (ID: %d)", c.conn.ConnectionID) // 关闭连接。

		// Attempt to roll back any active transaction
		// 尝试回滚任何活跃的事务
		c.mu.Lock()
		currentTx := c.transaction
		c.transaction = nil // Mark as closed
		c.mu.Unlock()

		if currentTx != nil {
			log.Debug("Rolling back active transaction for connection ID %d", c.conn.ConnectionID) // 回滚连接的活跃事务。
			if err := currentTx.Rollback(context.Background()); err != nil { // Use background context for rollback
				log.Error("Failed to rollback transaction for connection ID %d: %v", c.conn.ConnectionID, err) // 回滚事务失败。
			} else {
				log.Debug("Transaction rolled back successfully for connection ID %d", c.conn.ConnectionID) // 事务回滚成功。
			}
		}


		if err := c.conn.Close(); err != nil {
			log.Error("Failed to close connection ID %d: %v", c.conn.ConnectionID, err) // 关闭连接失败。
		}
		log.Info("Connection ID %d closed.", c.conn.ConnectionID) // 连接 ID %d 已关闭。
	}()

	for {
		// Read the next command from the client
		// 从客户端读取下一个命令
		packet, err := c.conn.ReadPacket()
		if err == io.EOF {
			log.Debug("Client closed connection (ID: %d)", c.conn.ConnectionID) // 客户端关闭连接。
			return // Exit loop on EOF
		}
		if err != nil {
			log.Error("Failed to read packet from connection ID %d: %v", c.conn.ConnectionID, err) // 从连接读取数据包失败。
			// Decide how to handle read errors (close connection?)
			// 决定如何处理读取错误（关闭连接？）
			return // Exit loop on other errors
		}

		// Process the command
		// 处理命令
		// The GMS server library handles initial command parsing (e.g., COM_QUERY, COM_PING).
		// The `mysql.Conn` object provides a method to handle the command.
		// It usually delegates to the sql.Handler.
		//
		// GMS 服务器库处理初始命令解析（例如 COM_QUERY, COM_PING）。
		// `mysql.Conn` 对象提供了一个处理命令的方法。
		// 它通常委托给 sql.Handler。

		// Let the GMS connection handle the packet, which will call our sql.Handler
		// 允许 GMS 连接处理数据包，这将调用我们的 sql.Handler
		if err := c.conn.HandlePacket(packet, c.session); err != nil { // Pass the session
			log.Error("Failed to handle packet for connection ID %d: %v", c.conn.ConnectionID, err) // 处理连接数据包失败。
			// Send error back to client if possible before closing
			// 如果可能，在关闭前将错误发送回客户端
			if writeErr := c.conn.WriteError(mysql.NewError(mysql.ErrUnknown, err.Error())); writeErr != nil {
				log.Error("Failed to write error back to client ID %d: %v", c.conn.ConnectionID, writeErr) // 将错误写回客户端失败。
			}
			// Decide how to handle packet handling errors (close connection?)
			// 决定如何处理数据包处理错误（关闭连接？）
			return // Exit loop on error
		}

		// Command processed, loop back to read next packet.
		// 命令已处理，循环回读下一个数据包。
	}
}


// TODO: Implement GuocedbSession struct implementing sql.Session.
// This struct will hold session-specific state and delegate to the compute components.
// It will be created by the sql.Handler's NewSession method.
//
// TODO: 实现实现 sql.Session 的 GuocedbSession 结构体。
// 此结构体将保存会话特定状态，并委托给计算组件。
// 它将由 sql.Handler 的 NewSession 方法创建。

// GuocedbSession is a custom implementation of the go-mysql-server sql.Session interface.
// It holds session state and provides access to the Catalog and TransactionManager.
//
// GuocedbSession 是 go-mysql-server sql.Session 接口的自定义实现。
// 它保存会话状态，并提供对 Catalog 和 TransactionManager 的访问。
type GuocedbSession struct {
	// connID is the connection ID associated with this session.
	// connID 是与此会话关联的连接 ID。
	connID uint32
	// catalog is the compute catalog available to this session.
	// catalog 是此会话可用的计算 catalog。
	catalog compute_catalog.Catalog
	// txManager is the transaction manager for this session.
	// txManager 是此会话的事务管理器。
	txManager compute_transaction.TransactionManager
	// currentDB is the currently selected database for this session.
	// currentDB 是此会话当前选择的数据库。
	currentDB string
	// sessionVars holds session variables.
	// sessionVars 保存会话变量。
	sessionVars *sql.SessionVariables
	// clientConn is the underlying mysql.Conn for the client.
	// clientConn 是客户端的底层 mysql.Conn。
	clientConn *mysql.Conn
	// client is the GMS ClientPrincipal for the connected user.
	// client 是连接用户的 GMS ClientPrincipal。
	client sql.ClientPrincipal

	// Note: GMS sql.Context will hold the *active* sql.Transaction.
	// The Session provides the TransactionManager to the GMS Context factory.
	//
	// 注意：GMS sql.Context 将保存*活跃*的 sql.Transaction。
	// Session 将 TransactionManager 提供给 GMS Context 工厂。
}

// NewGuocedbSession creates a new GuocedbSession.
// NewGuocedbSession 创建一个新的 GuocedbSession。
func NewGuocedbSession(connID uint32, cat compute_catalog.Catalog, txManager compute_transaction.TransactionManager) (sql.Session, error) {
	log.Debug("Creating GuocedbSession for connection ID %d", connID) // 创建 GuocedbSession。

	// Initialize session variables with defaults
	// 使用默认值初始化会话变量
	sessVars := sql.NewSessionVariables() // GMS provides default session variables

	// TODO: Get client principal information from the mysql.Conn and set it.
	// This is needed for privilege checks. The Handler's NewSession will receive the mysql.Conn.
	//
	// TODO: 从 mysql.Conn 获取客户端 principal 信息并设置。
	// 这是权限检查所需的。Handler 的 NewSession 将接收 mysql.Conn。
	// For now, client principal is nil.
	// 目前，客户端 principal 为 nil。

	session := &GuocedbSession{
		connID: connID,
		catalog: cat,
		txManager: txManager,
		currentDB: "", // Default to no database selected
		sessionVars: sessVars,
		clientConn: nil, // Will be set by Handler's NewSession
		client: nil, // Will be set by Handler's NewSession
	}

	return session, nil
}

// Client returns the ClientPrincipal for the session.
// Client 返回会话的 ClientPrincipal。
func (s *GuocedbSession) Client() sql.ClientPrincipal {
	return s.client // Return the stored client principal
}

// SetClient sets the ClientPrincipal for the session.
// SetClient 设置会话的 ClientPrincipal。
func (s *GuocedbSession) SetClient(client sql.ClientPrincipal) {
	s.client = client
}


// ID returns the connection ID for the session.
// ID 返回会话的连接 ID。
func (s *GuocedbSession) ID() uint32 {
	return s.connID
}

// Convert implements sql.Session.Convert, needed for type conversions.
// Convert 实现 sql.Session.Convert，类型转换所需。
func (s *GuocedbSession) Convert(ctx *sql.Context, v interface{}) (interface{}, error) {
	// Delegate to GMS internal conversion if possible, or implement custom logic.
	// This often involves using sql.Convert methods or schema type methods.
	//
	// 如果可能，委托给 GMS 内部转换，或实现自定义逻辑。
	// 这通常涉及使用 sql.Convert 方法或模式类型方法。
	log.Warn("GuocedbSession Convert called (Placeholder)") // 调用 GuocedbSession Convert（占位符）。
	// A simple placeholder: attempt basic string conversion.
	// 一个简单的占位符：尝试基本字符串转换。
	switch v := v.(type) {
	case string:
		return v, nil // No conversion needed
	case []byte:
		return string(v), nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		return v, nil // Return as is for now
	}
}


// Warn implements sql.Session.Warn, for logging warnings.
// Warn 实现 sql.Session.Warn，用于记录警告。
func (s *GuocedbSession) Warn(warn *sql.Warning) {
	// Log the warning using our logging system
	// 使用我们的日志系统记录警告
	log.Warn("Session Warning (Conn ID %d): Code=%d, State=%s, Message=%s", s.connID, warn.Code, warn.State, warn.Message) // 会话警告。
}

// Warnings implements sql.Session.Warnings, returns accumulated warnings.
// Warnings 实现 sql.Session.Warnings，返回累积的警告。
func (s *GuocedbSession) Warnings() []*sql.Warning {
	// TODO: Implement warning accumulation.
	// TODO: 实现警告累积。
	log.Warn("GuocedbSession Warnings called (Placeholder)") // 调用 GuocedbSession Warnings（占位符）。
	return []*sql.Warning{} // Return empty list for now
}

// SetWarnings implements sql.Session.SetWarnings.
// SetWarnings 实现 sql.Session.SetWarnings。
func (s *GuocedbSession) SetWarnings(warns []*sql.Warning) {
	// TODO: Implement setting warnings.
	// TODO: 实现设置警告。
	log.Warn("GuocedbSession SetWarnings called (Placeholder)") // 调用 GuocedbSession SetWarnings（占位符）。
}

// AddWarning implements sql.Session.AddWarning.
// AddWarning 实现 sql.Session.AddWarning。
func (s *GuocedbSession) AddWarning(code uint16, state, msg string) {
	// TODO: Implement adding a single warning.
	// TODO: 实现添加单个警告。
	log.Warn("GuocedbSession AddWarning called (Placeholder): Code=%d, State=%s, Message=%s", code, state, msg) // 调用 GuocedbSession AddWarning（占位符）。
}

// SetCurrentDatabase implements sql.Session.SetCurrentDatabase.
// SetCurrentDatabase 实现 sql.Session.SetCurrentDatabase。
func (s *GuocedbSession) SetCurrentDatabase(dbName string) {
	log.Debug("Setting current database for session ID %d to '%s'", s.connID, dbName) // 设置会话的当前数据库。
	s.currentDB = dbName
}

// GetCurrentDatabase implements sql.Session.GetCurrentDatabase.
// GetCurrentDatabase 实现 sql.Session.GetCurrentDatabase。
func (s *GuocedbSession) GetCurrentDatabase() string {
	return s.currentDB
}

// SetClientConn sets the underlying go-mysql-server client connection.
// SetClientConn 设置底层的 go-mysql-server 客户端连接。
func (s *GuocedbSession) SetClientConn(conn *mysql.Conn) {
	s.clientConn = conn
}

// GetClientConn returns the underlying go-mysql-server client connection.
// GetClientConn 返回底层的 go-mysql-server 客户端连接。
func (s *GuocedbSession) GetClientConn() *mysql.Conn {
	return s.clientConn
}


// GetTransaction returns the active sql.Transaction for the session context.
// GMS sql.Context will call this to get the transaction.
// GetTransaction 返回会话 context 的活跃 sql.Transaction。
// GMS sql.Context 将调用此方法获取事务。
func (s *GuocedbSession) GetTransaction() sql.Transaction {
	log.Debug("GuocedbSession GetTransaction called for session ID %d", s.connID) // 调用 GuocedbSession GetTransaction。
	// The session manages the transaction lifecycle (Begin, Commit, Rollback).
	// The sql.Context needs the *current* transaction.
	//
	// 会话管理事务生命周期（Begin, Commit, Rollback）。
	// sql.Context 需要*当前*事务。
	// We need to expose our interfaces.Transaction as a sql.Transaction.
	// This requires our interfaces.Transaction to implement sql.Transaction, or wrap it.
	// Let's assume our interfaces.Transaction implements sql.Transaction or can be cast.
	//
	// 我们需要将我们的 interfaces.Transaction 暴露为 sql.Transaction。
	// 这要求我们的 interfaces.Transaction 实现 sql.Transaction，或对其进行包装。
	// 假设我们的 interfaces.Transaction 实现 sql.Transaction 或可以进行类型转换。

	// Return the underlying sql.Transaction from our interfaces.Transaction
	// 从我们的 interfaces.Transaction 返回底层的 sql.Transaction
	c.mu.Lock()
	currentTx := c.transaction // Access the transaction field from the Connection struct
	c.mu.Unlock()

	if currentTx == nil {
		log.Debug("GuocedbSession GetTransaction: No active transaction.") // GuocedbSession GetTransaction：没有活跃事务。
		return nil // No active transaction
	}

	// Check if the underlying transaction implements sql.Transaction
	// 检查底层事务是否实现 sql.Transaction
	gmsTx, ok := currentTx.UnderlyingTx().(sql.Transaction)
	if !ok {
		log.Error("Underlying interfaces.Transaction does not provide a GMS sql.Transaction interface.") // 底层 interfaces.Transaction 未提供 GMS sql.Transaction 接口。
		// This indicates an integration issue.
		// 这表明存在集成问题。
		return nil // Return nil transaction, which might lead to errors later in GMS
	}

	return gmsTx // Return the underlying GMS sql.Transaction
}

// SetTransaction sets the active sql.Transaction for the session context.
// This might be called by GMS after BEGIN statements.
//
// SetTransaction 设置会话 context 的活跃 sql.Transaction。
// GMS 在执行 BEGIN 语句后可能会调用此方法。
// Note: This method might be less commonly used if the Session's TransactionManager
// is used by GMS Context factory to create transactions.
//
// 注意：如果 GMS Context 工厂使用 Session 的 TransactionManager 创建事务，
// 此方法可能不常用。
func (s *GuocedbSession) SetTransaction(tx sql.Transaction) {
	log.Debug("GuocedbSession SetTransaction called for session ID %d", s.connID) // 调用 GuocedbSession SetTransaction。
	// We need to wrap the GMS sql.Transaction into our interfaces.Transaction.
	// Need a wrapper for sql.Transaction that implements interfaces.Transaction.
	// Let's assume interfaces.Transaction can wrap sql.Transaction.
	//
	// 我们需要将 GMS sql.Transaction 包装到我们的 interfaces.Transaction 中。
	// 需要一个包装 sql.Transaction 并实现 interfaces.Transaction 的包装器。
	// 假设 interfaces.Transaction 可以包装 sql.Transaction。

	// Create a wrapper for the GMS sql.Transaction
	// 为 GMS sql.Transaction 创建一个包装器
	wrappedTx := NewGmsTransactionWrapper(tx) // Need to create this wrapper

	// Update the transaction field in the associated Connection struct.
	// Need a way for the Session to access the Connection's transaction field.
	// The Session could hold a pointer to the Connection, or the Connection
	// could call a method on the Session to update its internal state.
	//
	// 更新关联 Connection 结构体中的事务字段。
	// 需要一种方法让 Session 访问 Connection 的事务字段。
	// Session 可以持有 Connection 的指针，或者 Connection 可以调用 Session 上的方法
	// 更新其内部状态。

	// Option: The Connection struct sets the transaction directly in the Session?
	// Option: The Session holds a pointer to the Connection and updates it?
	// Let's add a SetActualTransaction method to GuocedbSession.

	// Update the connection's transaction field.
	// The session needs to tell the connection about the new transaction.
	// This requires a back-reference or a channel.
	//
	// 更新连接的事务字段。
	// 会话需要告知连接新的事务。
	// 这需要反向引用或 channel。

	// Let's add a SetConnectionTransaction method to GuocedbSession.
	// Let's assume this method exists for now.
	//
	// 添加 SetConnectionTransaction 方法到 GuocedbSession。
	// 假设此方法暂时存在。
	// s.SetConnectionTransaction(wrappedTx) // This is complex due to the Connection/Session relationship

	// Alternative: The GMS Context holds the transaction. The Session provides the Factory for it.
	// When GMS needs a transaction, it calls TransactionManager.Begin (via the Session's GetTransactionManager).
	// The GMS Context then manages the sql.Transaction obtained from Begin.
	// Our Session doesn't need to store the active GMS sql.Transaction directly.
	// It needs to provide the TransactionManager.
	//
	// 替代方案：GMS Context 持有事务。Session 提供工厂。
	// 当 GMS 需要事务时，它调用 TransactionManager.Begin（通过 Session 的 GetTransactionManager）。
	// GMS Context 然后管理从 Begin 获取的 sql.Transaction。
	// 我们的 Session 不需要直接存储活跃的 GMS sql.Transaction。
	// 它需要提供 TransactionManager。

	// Let's rethink. The sql.Handler's HandleQuery gets the query string.
	// It creates or gets a *sql.Context for the execution.
	// This *sql.Context needs the sql.Catalog and the sql.Transaction.
	// The sql.Context factory might call Session methods like GetTransactionManager to get resources.
	//
	// 重新思考。sql.Handler 的 HandleQuery 获取查询字符串。
	// 它为执行创建或获取一个 *sql.Context。
	// 此 *sql.Context 需要 sql.Catalog 和 sql.Transaction。
	// sql.Context 工厂可能会调用 Session 方法，如 GetTransactionManager 获取资源。

	// The `conn.transaction` field is for managing the *storage layer* transaction lifecycle (commit/rollback)
	// driven by COMMIT/ROLLBACK commands processed by our Handler/Executor.
	// The `sql.Context` manages the *GMS layer* transaction pointer used during plan execution.
	// These need to be linked.
	//
	// `conn.transaction` 字段用于管理由我们的 Handler/Executor 处理的 COMMIT/ROLLBACK 命令驱动的*存储层*事务生命周期。
	// `sql.Context` 管理在计划执行期间使用的*GMS 层*事务指针。
	// 这些需要链接。

	// When the Handler receives COMMIT/ROLLBACK, it should call Commit/Rollback on `conn.transaction`.
	// When the Handler prepares the sql.Context for query execution, it should set the `sql.Context`'s
	// transaction field to the GMS-compatible wrapper of `conn.transaction`.
	//
	// 当 Handler 接收到 COMMIT/ROLLBACK 时，它应该调用 `conn.transaction` 上的 Commit/Rollback。
	// 当 Handler 为查询执行准备 sql.Context 时，它应该将 `sql.Context` 的
	// 事务字段设置为 `conn.transaction` 的 GMS 兼容包装器。

	// The SetTransaction method on sql.Session might be called by GMS if it manages transactions internally (less likely for engines).
	// Let's leave this method with a warning for now, as our transaction management is external to GMS's core loop.
	//
	// 如果 GMS 内部管理事务（对于引擎来说不太可能），则可能会调用 sql.Session 上的 SetTransaction 方法。
	// 暂时将此方法保留警告，因为我们的事务管理在 GMS 核心循环之外。
	log.Warn("GuocedbSession SetTransaction called (Likely not used directly by Guocedb logic)") // 调用 GuocedbSession SetTransaction（Guocedb 逻辑可能不会直接使用）。
	// If GMS *does* call this, we'd need to wrap `tx` and replace `c.transaction`. This is complex.
	// If GMS 确实调用了此方法，我们需要包装 `tx` 并替换 `c.transaction`。这很复杂。
}

// UseDatabase implements sql.Session.UseDatabase.
// UseDatabase 实现 sql.Session.UseDatabase。
// This is called by GMS when a USE statement is executed.
// GMS 执行 USE 语句时调用此方法。
func (s *GuocedbSession) UseDatabase(ctx *sql.Context, dbName string) error {
	log.Debug("Session ID %d: Using database '%s'", s.connID, dbName) // 会话使用数据库。

	// Check if the database exists using our compute catalog
	// 使用我们的计算 catalog 检查数据库是否存在
	_, err := s.catalog.Database(context.Background(), dbName) // Use background context or adapt GMS ctx
	if err != nil {
		// The error should be our ErrDatabaseNotFound or something else mapped from storage.
		// We need to return a GMS error back to GMS.
		//
		// 错误应该是我们的 ErrDatabaseNotFound 或其他从存储映射的错误。
		// 我们需要将 GMS 错误返回给 GMS。
		if errors.Is(err, errors.ErrDatabaseNotFound) {
			log.Warn("Database '%s' not found for USE statement.", dbName) // USE 语句未找到数据库。
			return sql.ErrDatabaseNotFound.New(dbName) // Return GMS error
		}
		log.Error("Error checking database '%s' for USE statement: %v", dbName, err) // 检查数据库失败。
		return fmt.Errorf("error checking database '%s': %w", dbName, err) // Propagate other errors
	}

	// Database exists, set it as the current database in the session
	// 数据库存在，将其设置为会话中的当前数据库
	s.SetCurrentDatabase(dbName) // Use the setter method

	log.Debug("Successfully set current database to '%s' for session ID %d", dbName, s.connID) // 成功设置当前数据库。
	return nil
}

// SetVariable implements sql.Session.SetVariable.
// SetVariable 实现 sql.Session.SetVariable。
// This is called by GMS for SET statements.
// GMS 执行 SET 语句时调用此方法。
func (s *GuocedbSession) SetVariable(ctx *sql.Context, name string, value interface{}) error {
	log.Debug("Session ID %d: Setting variable '%s' = %v", s.connID, name, value) // 会话设置变量。
	// Delegate to GMS session variables manager
	// 委托给 GMS 会话变量管理器
	return s.sessionVars.Set(ctx, name, value) // GMS sessionVars handles setting variables
}

// GetVariable implements sql.Session.GetVariable.
// GetVariable 实现 sql.Session.GetVariable。
// This is called by GMS for variable access.
// GMS 访问变量时调用此方法。
func (s *GuocedbSession) GetVariable(ctx *sql.Context, name string) (interface{}, error) {
	log.Debug("Session ID %d: Getting variable '%s'", s.connID, name) // 会话获取变量。
	// Delegate to GMS session variables manager
	// 委托给 GMS 会话变量管理器
	value, err := s.sessionVars.Get(ctx, name) // GMS sessionVars handles getting variables
	if err != nil {
		log.Error("Failed to get variable '%s' for session ID %d: %v", name, s.connID, err) // 获取变量失败。
		// GMS GetVariable returns sql.ErrUnknownVariable or similar.
		// GMS GetVariable 返回 sql.ErrUnknownVariable 或类似错误。
		return nil, err // Return GMS error directly
	}
	log.Debug("Got variable '%s' = %v for session ID %d", name, value, s.connID) // 获取变量。
	return value, nil
}

// GetAllVariables implements sql.Session.GetAllVariables.
// GetAllVariables 实现 sql.Session.GetAllVariables。
func (s *GuocedbSession) GetAllVariables(ctx *sql.Context) ([]sql.SessionVariable, error) {
	log.Debug("Session ID %d: Getting all variables", s.connID) // 会话获取所有变量。
	// Delegate to GMS session variables manager
	// 委托给 GMS 会话变量管理器
	return s.sessionVars.GetAll() // GMS sessionVars handles getting all variables
}

// GetCatalog implements sql.Session.GetCatalog.
// GMS sql.Context factory calls this to get the catalog for the context.
//
// GetCatalog 实现 sql.Session.GetCatalog。
// GMS sql.Context 工厂调用此方法获取 context 的 catalog。
func (s *GuocedbSession) GetCatalog() sql.Catalog {
	log.Debug("GuocedbSession GetCatalog called for session ID %d", s.connID) // 调用 GuocedbSession GetCatalog。
	// Return the GMS-compatible sql.Catalog from our compute catalog
	// 返回计算 catalog 中的 GMS 兼容 sql.Catalog
	// Need to call GetCatalogAsSQL on s.catalog.
	// Since GetCatalogAsSQL takes context, and sql.Session.GetCatalog doesn't,
	// we might need to pass a background context or reconsider the interface.
	//
	// 需要在 s.catalog 上调用 GetCatalogAsSQL。
	// 由于 GetCatalogAsSQL 接受 context，而 sql.Session.GetCatalog 不接受，
	// 我们可能需要传递 background context 或重新考虑接口。

	// For simplicity, use context.Background() for the call to GetCatalogAsSQL.
	// This might not be ideal for real-world contexts.
	//
	// 为了简化，对 GetCatalogAsSQL 调用使用 context.Background()。
	// 这对于实际 context 可能不太理想。
	sqlCatalog, err := s.catalog.GetCatalogAsSQL(context.Background())
	if err != nil {
		log.Error("Failed to get GMS sql.Catalog from compute catalog during GetCatalog: %v", err) // 获取 GMS sql.Catalog 失败。
		// This is a critical error during context creation. How to handle?
		// 这是一个 context 创建期间的关键错误。如何处理？
		// Returning nil might cause panic in GMS.
		// 返回 nil 可能导致 GMS panic。
		// Let's log and panic for now, or return a dummy catalog.
		// 暂时记录并 panic，或返回虚拟 catalog。
		panic(fmt.Sprintf("Failed to get GMS sql.Catalog: %v", err)) // Panic for now on critical error
	}

	return sqlCatalog // Return the GMS sql.Catalog
}

// GetTransactionManager implements sql.Session.GetTransactionManager.
// GMS sql.Context factory calls this to get the transaction manager for the context.
//
// GetTransactionManager 实现 sql.Session.GetTransactionManager。
// GMS sql.Context 工厂调用此方法获取 context 的事务管理器。
func (s *GuocedbSession) GetTransactionManager() sql.TransactionManager {
	log.Debug("GuocedbSession GetTransactionManager called for session ID %d", s.connID) // 调用 GuocedbSession GetTransactionManager。
	// We need to expose our compute_transaction.TransactionManager as a GMS sql.TransactionManager.
	// Need a wrapper for compute_transaction.TransactionManager that implements sql.TransactionManager.
	// Let's create a wrapper.
	//
	// 我们需要将我们的 compute_transaction.TransactionManager 暴露为 GMS sql.TransactionManager。
	// 需要一个包装 compute_transaction.TransactionManager 并实现 sql.TransactionManager 的包装器。
	// 创建一个包装器。
	return NewGmsTransactionManagerWrapper(s.txManager) // Create and return the wrapper
}

// NewGmsTransactionManagerWrapper wraps our compute_transaction.TransactionManager
// to implement GMS sql.TransactionManager.
//
// NewGmsTransactionManagerWrapper 包装我们的 compute_transaction.TransactionManager
// 以实现 GMS sql.TransactionManager。
type GmsTransactionManagerWrapper struct {
	actualTxManager compute_transaction.TransactionManager
}

// NewGmsTransactionManagerWrapper creates a new wrapper.
// NewGmsTransactionManagerWrapper 创建一个新的包装器。
func NewGmsTransactionManagerWrapper(txManager compute_transaction.TransactionManager) sql.TransactionManager {
	log.Debug("Creating GmsTransactionManagerWrapper") // 创建 GmsTransactionManagerWrapper。
	return &GmsTransactionManagerWrapper{actualTxManager: txManager}
}

// Begin implements sql.TransactionManager.Begin.
// Delegates to the actual compute_transaction.TransactionManager.
// Begin 实现 sql.TransactionManager.Begin。
// 委托给实际的 compute_transaction.TransactionManager。
func (w *GmsTransactionManagerWrapper) Begin(ctx *sql.Context) (sql.Transaction, error) {
	log.Debug("GmsTransactionManagerWrapper Begin called.") // 调用 GmsTransactionManagerWrapper Begin。
	// Delegate to the actual transaction manager.
	// The actualTxManager.Begin returns interfaces.Transaction.
	//
	// 委托给实际的事务管理器。
	// actualTxManager.Begin 返回 interfaces.Transaction。
	actualTx, err := w.actualTxManager.Begin(context.Background()) // Use background context or adapt GMS ctx
	if err != nil {
		log.Error("Failed to begin transaction via actual tx manager: %v", err) // 通过实际 tx 管理器开启事务失败。
		// Map error if needed
		// 如果需要，映射错误
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Wrap the interfaces.Transaction into a sql.Transaction.
	// Need a wrapper for interfaces.Transaction that implements sql.Transaction.
	// Let's assume interfaces.Transaction itself implements sql.Transaction or can be cast.
	//
	// 将 interfaces.Transaction 包装到 sql.Transaction 中。
	// 需要一个包装 interfaces.Transaction 并实现 sql.Transaction 的包装器。
	// 假设 interfaces.Transaction 本身实现 sql.Transaction 或可以进行类型转换。

	// Check if the underlying transaction implements sql.Transaction
	// 检查底层事务是否实现 sql.Transaction
	gmsTx, ok := actualTx.UnderlyingTx().(sql.Transaction)
	if !ok {
		log.Error("Underlying interfaces.Transaction does not provide a GMS sql.Transaction interface for Begin.") // 底层 interfaces.Transaction 未提供 GMS sql.Transaction 接口。
		// Rollback the transaction that was just started.
		// 回滚刚刚开启的事务。
		rollbackErr := actualTx.Rollback(context.Background()) // Use background context
		if rollbackErr != nil {
			log.Error("Failed to rollback transaction after Begin wrapping error: %v", rollbackErr) // 在 Begin 包装错误后回滚事务失败。
			// This is bad. Log and continue returning the original error.
			// 这很糟糕。记录并继续返回原始错误。
		}
		return nil, errors.ErrInternal.New("transaction manager not configured correctly")
	}


	log.Debug("Transaction begun and wrapped successfully.") // 事务已成功开启和包装。
	return gmsTx // Return the underlying GMS sql.Transaction
}


// Commit implements sql.TransactionManager.Commit.
// Delegates to the actual interfaces.Transaction.
// Commit 实现 sql.TransactionManager.Commit。
// 委托给实际的 interfaces.Transaction。
// Note: This might be called by GMS for implicit commits or if GMS manages the transaction lifecycle.
// In Guocedb, we likely manage Commit/Rollback via our Handler for explicit statements.
//
// 注意：GMS 可能会为隐式提交或如果 GMS 管理事务生命周期而调用此方法。
// 在 Guocedb 中，我们可能通过 Handler 处理显式语句来管理 Commit/Rollback。
func (w *GmsTransactionManagerWrapper) Commit(ctx *sql.Context, tx sql.Transaction) error {
	log.Debug("GmsTransactionManagerWrapper Commit called.") // 调用 GmsTransactionManagerWrapper Commit。
	// We need to get our interfaces.Transaction from the GMS sql.Transaction.
	// Need a way to unwrap the GMS sql.Transaction if it's our wrapper, or cast if our interfaces.Transaction is used directly.
	//
	// 我们需要从 GMS sql.Transaction 获取我们的 interfaces.Transaction。
	// 如果是我们的包装器，需要解包 GMS sql.Transaction；如果直接使用 interfaces.Transaction，则进行类型转换。

	// This indicates that the GMS sql.Transaction passed to Commit/Rollback must be
	// related back to the interfaces.Transaction object created by our Begin method.
	//
	// 这表明传递给 Commit/Rollback 的 GMS sql.Transaction 必须与
	// 我们的 Begin 方法创建的 interfaces.Transaction 对象相关联。
	// The GMS sql.Transaction interface has an Id() method. Maybe we can map IDs?
	// GMS sql.Transaction 接口有一个 Id() 方法。也许我们可以映射 ID？
	// Or the sql.Transaction wrapper we returned from Begin needs a method to get the underlying interfaces.Transaction.
	// 或者从 Begin 返回的 sql.Transaction 包装器需要一个方法来获取底层的 interfaces.Transaction。

	// Let's assume the sql.Transaction returned by our Begin wrapper has an `Unwrap()` method to get interfaces.Transaction.
	//
	// 假设我们的 Begin 包装器返回的 sql.Transaction 有一个 `Unwrap()` 方法获取 interfaces.Transaction。
	// Or, simpler: the sql.Transaction returned by NewGmsTransactionWrapper's Begin method is the one used by GMS Context.
	// We need to find the corresponding interfaces.Transaction.
	//
	// 或者，更简单：NewGmsTransactionManagerWrapper 的 Begin 方法返回的 sql.Transaction 是 GMS Context 使用的。
	// 我们需要找到相应的 interfaces.Transaction。

	// This pattern suggests that the GmsTransactionManagerWrapper.Begin method should
	// return a custom sql.Transaction implementation that *holds* the interfaces.Transaction
	// and delegates Commit/Rollback calls.
	//
	// 这个模式表明 GmsTransactionManagerWrapper.Begin 方法应该
	// 返回一个自定义的 sql.Transaction 实现，该实现*持有* interfaces.Transaction
	// 并委托 Commit/Rollback 调用。

	// Let's implement a GmsTransactionWrapper struct that implements sql.Transaction and wraps interfaces.Transaction.
	// 在此处实现一个实现 sql.Transaction 并包装 interfaces.Transaction 的 GmsTransactionWrapper 结构体。

	// Get the underlying interfaces.Transaction from the GMS sql.Transaction (assuming it's our wrapper)
	// 从 GMS sql.Transaction 获取底层的 interfaces.Transaction（假设它是我们的包装器）
	ourTx, ok := tx.(*GmsTransactionWrapper) // Try casting to our wrapper type
	if !ok {
		log.Error("sql.Transaction passed to Commit is not our wrapper type: %T", tx) // 传递给 Commit 的 sql.Transaction 不是我们的包装器类型。
		return errors.ErrInternal.New("invalid transaction object for commit")
	}

	// Delegate Commit to the actual interfaces.Transaction
	// 委托 Commit 给实际的 interfaces.Transaction
	// Use background context or adapt GMS ctx
	err := ourTx.actualTx.Commit(context.Background())
	if err != nil {
		log.Error("Failed to commit transaction via actual tx: %v", err) // 通过实际 tx 提交事务失败。
		// Map error if needed
		// 如果需要，映射错误
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Debug("Transaction committed successfully.") // 事务已成功提交。
	return nil
}

// Rollback implements sql.TransactionManager.Rollback.
// Delegates to the actual interfaces.Transaction.
// Rollback 实现 sql.TransactionManager.Rollback。
// 委托给实际的 interfaces.Transaction。
// Note: This might be called by GMS for implicit rollbacks or if GMS manages the transaction lifecycle.
// In Guocedb, we likely manage Commit/Rollback via our Handler for explicit statements.
//
// 注意：GMS 可能会为隐式回滚或如果 GMS 管理事务生命周期而调用此方法。
// 在 Guocedb 中，我们可能通过 Handler 处理显式语句来管理 Commit/Rollback。
func (w *GmsTransactionManagerWrapper) Rollback(ctx *sql.Context, tx sql.Transaction) error {
	log.Debug("GmsTransactionManagerWrapper Rollback called.") // 调用 GmsTransactionManagerWrapper Rollback。
	// Get the underlying interfaces.Transaction from the GMS sql.Transaction (assuming it's our wrapper)
	// 从 GMS sql.Transaction 获取底层的 interfaces.Transaction（假设它是我们的包装器）
	ourTx, ok := tx.(*GmsTransactionWrapper) // Try casting to our wrapper type
	if !ok {
		log.Error("sql.Transaction passed to Rollback is not our wrapper type: %T", tx) // 传递给 Rollback 的 sql.Transaction 不是我们的包装器类型。
		return errors.ErrInternal.New("invalid transaction object for rollback")
	}

	// Delegate Rollback to the actual interfaces.Transaction
	// 委托 Rollback 给实际的 interfaces.Transaction
	// Use background context or adapt GMS ctx
	err := ourTx.actualTx.Rollback(context.Background())
	if err != nil {
		log.Error("Failed to rollback transaction via actual tx: %v", err) // 通过实际 tx 回滚事务失败。
		// Map error if needed
		// 如果需要，映射错误
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	log.Debug("Transaction rolled back successfully.") // 事务已成功回滚。
	return nil
}

// --- Wrapper for interfaces.Transaction to implement sql.Transaction ---
// --- interfaces.Transaction 的包装器，实现 sql.Transaction ---

// GmsTransactionWrapper wraps our interfaces.Transaction to implement GMS sql.Transaction.
// GmsTransactionWrapper 包装我们的 interfaces.Transaction 以实现 GMS sql.Transaction。
type GmsTransactionWrapper struct {
	// actualTx is the underlying interfaces.Transaction implementation.
	// actualTx 是底层的 interfaces.Transaction 实现。
	actualTx interfaces.Transaction
}

// NewGmsTransactionWrapper creates a new wrapper for interfaces.Transaction.
// NewGmsTransactionWrapper 创建一个新的 interfaces.Transaction 包装器。
func NewGmsTransactionWrapper(tx interfaces.Transaction) sql.Transaction {
	log.Debug("Creating GmsTransactionWrapper for interfaces.Transaction (ID: %s)", tx.ID()) // 创建 GmsTransactionWrapper。
	return &GmsTransactionWrapper{actualTx: tx}
}

// ID implements sql.Transaction.Id.
// ID 实现 sql.Transaction.Id。
// Delegates to the underlying interfaces.Transaction.Id().
// 委托给底层的 interfaces.Transaction.Id()。
func (w *GmsTransactionWrapper) ID() uint64 {
	// Assuming interfaces.Transaction has an ID method returning uint64 or similar.
	// If not, generate a unique ID or adapt. Let's assume it does.
	// 假设 interfaces.Transaction 有一个返回 uint64 或类似类型的 ID 方法。
	// 如果没有，生成一个唯一 ID 或适配。假设它有。
	// The ID method on interfaces.Transaction returns string. Need to convert or handle.
	// Let's add an ID() uint64 to interfaces.Transaction or find another way.
	// Or, GMS sql.Transaction ID might just be for internal GMS use.
	// Let's return a dummy ID for now or rely on the underlyingTx.Id() if available.
	//
	// interfaces.Transaction 上的 ID 方法返回字符串。需要转换或处理。
	// 添加一个 ID() uint64 到 interfaces.Transaction 或找到其他方法。
	// 或者，GMS sql.Transaction ID 可能仅供 GMS 内部使用。
	// 现在返回一个虚拟 ID，或者如果 available 则依赖 underlyingTx.Id()。
	// For BadgerTx, ID is uint64. Assuming interfaces.Transaction.ID() now returns uint64.
	// 对于 BadgerTx，ID 是 uint64。假设 interfaces.Transaction.ID() 现在返回 uint64。
	// Check interfaces.Transaction definition. Yes, it returns string. Need to convert.
	// 检查 interfaces.Transaction 定义。是的，它返回字符串。需要转换。
	// The sql.Transaction ID is uint64. This is a mismatch.
	// We need to provide a uint64 ID that GMS can use. Maybe a sequential counter in the manager?
	// Or wrap the BadgerTx ID (uint64) directly?
	// Let's assume BadgerTx ID is uint64 and interfaces.Transaction exposes it or we access it.
	//
	// sql.Transaction ID 是 uint64。这是一个不匹配。
	// 我们需要提供一个 GMS 可以使用的 uint64 ID。也许管理器中的序列计数器？
	// 或者直接包装 BadgerTx ID (uint64)？
	// 假设 BadgerTx ID 是 uint64，并且 interfaces.Transaction 暴露它或者我们访问它。
	// 检查 BadgerTx。它有一个 *badger.Tx 字段。Badger tx doesn't have a public uint64 ID.
	// 检查 BadgerTx。它有一个 *badger.Tx 字段。Badger tx 没有公共的 uint64 ID。
	// Let's return a dummy ID for now. GMS might not strictly rely on this ID for core logic.
	// 暂时返回一个虚拟 ID。GMS 可能不严格依赖此 ID 进行核心逻辑。
	// sql.Transaction interface just requires the Id() method. The value can be anything unique enough.
	// sql.Transaction 接口只需要 Id() 方法。值可以是任何足够独特的值。
	// Using a counter in the manager or connection might work.
	// Maybe the interfaces.Transaction ID (string) can be converted/hashed?
	// Let's return 0 for now, or a hash if the string ID is complex.
	//
	// 在管理器或连接中使用计数器可能有效。
	// 也许 interfaces.Transaction ID (string) 可以转换/哈希？
	// 暂时返回 0，或者如果字符串 ID 复杂则返回哈希值。
	// Using the string ID's hash might be unstable. Let's return a counter managed by the wrapper factory?
	// Or maybe the Connection or Session manages a transaction counter.
	// Let's add a method to interfaces.Transaction to get a uint64 ID compatible with sql.Transaction.
	// Or, simpler, have the GmsTransactionManagerWrapper assign sequential IDs.

	// For now, return a placeholder ID.
	// 暂时返回一个占位符 ID。
	log.Warn("GmsTransactionWrapper ID called (Placeholder)") // 调用 GmsTransactionWrapper ID（占位符）。
	return 0 // Placeholder ID
}

// Commit implements sql.Transaction.Commit.
// Delegates to the underlying interfaces.Transaction.
// Commit 实现 sql.Transaction.Commit。
// 委托给底层的 interfaces.Transaction。
// Note: This might be called by GMS directly on the transaction object if it manages lifecycle.
// In Guocedb, our Handler explicitly calls Commit/Rollback on the Connection's transaction.
// If GMS calls this, it indicates a different flow than anticipated.
//
// 注意：如果 GMS 管理生命周期，可能会直接在事务对象上调用此方法。
// 在 Guocedb 中，我们的 Handler 显式调用 Connection 事务上的 Commit/Rollback。
// 如果 GMS 调用此方法，则表示流程与预期不同。
func (w *GmsTransactionWrapper) Commit(ctx *sql.Context) error {
	log.Debug("GmsTransactionWrapper Commit called for tx ID %d", w.ID()) // 调用 GmsTransactionWrapper Commit。
	// Delegate Commit to the actual interfaces.Transaction
	// Use background context or adapt GMS ctx
	err := w.actualTx.Commit(context.Background())
	if err != nil {
		log.Error("Failed to commit transaction ID %d via actual tx: %v", w.ID(), err) // 通过实际 tx 提交事务失败。
		// Map error if needed
		// 如果需要，映射错误
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	log.Debug("Transaction ID %d committed successfully.", w.ID()) // 事务 ID %d 已成功提交。
	return nil
}

// Rollback implements sql.Transaction.Rollback.
// Delegates to the underlying interfaces.Transaction.
// Rollback 实现 sql.Transaction.Rollback。
// 委托给底层的 interfaces.Transaction。
// Note: This might be called by GMS directly on the transaction object.
//
// 注意：GMS 可能直接在事务对象上调用此方法。
func (w *GmsTransactionWrapper) Rollback(ctx *sql.Context) error {
	log.Debug("GmsTransactionWrapper Rollback called for tx ID %d", w.ID()) // 调用 GmsTransactionWrapper Rollback。
	// Delegate Rollback to the actual interfaces.Transaction
	// Use background context or adapt GMS ctx
	err := w.actualTx.Rollback(context.Background())
	if err != nil {
		log.Error("Failed to rollback transaction ID %d via actual tx: %v", w.ID(), err) // 通过实际 tx 回滚事务失败。
		// Map error if needed
		// 如果需要，映射错误
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	log.Debug("Transaction ID %d rolled back successfully.", w.ID()) // 事务 ID %d 已成功回滚。
	return nil
}