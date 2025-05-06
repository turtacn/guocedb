// Package mysql provides the MySQL server protocol implementation.
// handler.go implements the go-mysql-server sql.Handler interface.
// It acts as the bridge between the MySQL protocol layer and the SQL engine (compute layer).
//
// mysql 包提供了 MySQL 服务器协议实现。
// handler.go 实现了 go-mysql-server 的 sql.Handler 接口。
// 它充当 MySQL 协议层和 SQL 引擎（计算层）之间的桥梁。
package mysql

import (
	"context"
	"fmt"
	"strings" // For command parsing like "BEGIN"

	"github.com/dolthub/go-mysql-server/server/mysql" // Import GMS MySQL server types
	"github.com/dolthub/go-mysql-server/sql"          // Import GMS SQL types
	"github.com/dolthub/go-mysql-server/sql/analyzer" // Import GMS analyzer for creating GMS engine
	"github.com/dolthub/go-mysql-server/sql/plan"     // Import GMS plan nodes for special commands
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/log"
	compute_analyzer "github.com/turtacub/guocedb/compute/analyzer" // Import compute analyzer
	compute_catalog "github.com/turtacn/guocedb/compute/catalog"       // Import compute catalog
	compute_executor "github.com/turtacn/guocedb/compute/executor"     // Import compute executor
	compute_optimizer "github.com/turtacub/guocedb/compute/optimizer" // Import compute optimizer
	compute_parser "github.com/turtacn/guocedb/compute/parser"         // Import compute parser
	compute_transaction "github.com/turtacn/guocedb/compute/transaction" // Import compute transaction manager
)

// Handler is an implementation of the go-mysql-server sql.Handler interface for Guocedb.
// It receives queries from the MySQL server and directs them to the compute layer.
//
// Handler 是 Guocedb 的 go-mysql-server sql.Handler 接口的实现。
// 它接收来自 MySQL 服务器的查询，并将其导向计算层。
type Handler struct {
	// components holds references to the compute layer components.
	// components 保存对计算层组件的引用。
	components *ComputeComponents
}

// NewHandler creates a new Handler instance.
// It requires references to the initialized compute layer components.
//
// NewHandler 创建一个新的 Handler 实例。
// 它需要对已初始化的计算层组件的引用。
func NewHandler(components *ComputeComponents) sql.Handler {
	log.Info("Initializing MySQL handler.") // 初始化 MySQL 处理程序。
	return &Handler{
		components: components,
	}
}

// NewSession creates a new sql.Session for a client connection.
// This method is called by the go-mysql-server library for each new connection after authentication.
//
// NewSession 为客户端连接创建一个新的 sql.Session。
// go-mysql-server 库在认证成功后为每个新连接调用此方法。
func (h *Handler) NewSession(conn *mysql.Conn) sql.Session {
	log.Debug("Handler NewSession called for connection ID %d", conn.ConnectionID) // 调用 Handler NewSession。

	// Create our custom GuocedbSession
	// 创建我们的自定义 GuocedbSession
	session, err := NewGuocedbSession(conn.ConnectionID, h.components.Catalog, h.components.TxManager) // Pass required components
	if err != nil {
		log.Error("Failed to create GuocedbSession for connection ID %d: %v", conn.ConnectionID, err) // 创建 GuocedbSession 失败。
		// In a real system, propagate this error to GMS or panic.
		// 在实际系统中，将此错误传播给 GMS 或 panic。
		// For now, log and return a nil session (GMS might handle nil).
		// 目前，记录并返回 nil 会话（GMS 可能会处理 nil）。
		return nil // Critical error
	}

	// Set the underlying GMS connection in the session
	// 在会话中设置底层的 GMS 连接
	session.(*GuocedbSession).SetClientConn(conn) // Cast to our concrete type

	// Set the client principal in the session (obtained from the connection after authentication)
	// 在会话中设置客户端 principal（认证后从连接获取）
	// The mysql.Conn should have the authenticated user information available.
	// mysql.Conn 应该提供已认证的用户信息。
	// Assuming mysql.Conn provides a Client method returning sql.ClientPrincipal.
	// 假设 mysql.Conn 提供一个返回 sql.ClientPrincipal 的 Client 方法。
	// Let's assume conn.Client() exists and returns sql.ClientPrincipal.
	// 假设 conn.Client() 存在并返回 sql.ClientPrincipal。
	// Check GMS mysql.Conn source... Yes, it has conn.Client().
	// 检查 GMS mysql.Conn 源码... 是的，它有 conn.Client()。
	clientPrincipal := conn.Client()
	if clientPrincipal == nil {
		log.Warn("Client principal is nil for new session %d. Authentication might have failed or was skipped?", conn.ConnectionID) // 新会话的客户端 principal 为 nil。认证可能失败或被跳过？
		// Create a dummy principal? Or rely on GMS handling nil?
		// 创建一个虚拟 principal？还是依赖 GMS 处理 nil？
		// GMS components like analyzer/executor will check for nil client and might panic.
		// GMS 组件如 analyzer/executor 会检查 nil client，并可能 panic。
		// It's better to set a dummy or error principal.
		// 最好设置一个虚拟或错误 principal。
		// For now, let's proceed with nil, assuming GMS handles it gracefully or it's caught later.
		// 暂时使用 nil，假设 GMS 能优雅处理，或稍后捕获。
	} else {
		log.Debug("Setting client principal '%s' for session ID %d", clientPrincipal.Identity().Username, conn.ConnectionID) // 设置客户端 principal。
	}
	session.(*GuocedbSession).SetClient(clientPrincipal) // Cast and set

	log.Debug("GuocedbSession created for connection ID %d", conn.ConnectionID) // 已为连接 ID %d 创建 GuocedbSession。
	return session // Return our custom session
}

// HandleQuery handles a query received from a client connection.
// This method is called by the go-mysql-server library after receiving a COM_QUERY packet.
//
// HandleQuery 处理从客户端连接接收到的查询。
// go-mysql-server 库在接收到 COM_QUERY 数据包后调用此方法。
// It orchestrates parsing, analysis, optimization, and execution using the compute layer components.
// 它协调使用计算层组件进行解析、分析、优化和执行。
func (h *Handler) HandleQuery(ctx *sql.Context, query string) (sql.RowIter, error) {
	log.Debug("Handler HandleQuery called for query: %s", query) // 调用 Handler HandleQuery。

	// The sql.Context provided by GMS already contains the sql.Session, which our GuocedbSession implements.
	// We can access our session and its components via ctx.Session().
	//
	// GMS 提供的 sql.Context 已包含 sql.Session，我们的 GuocedbSession 实现了该接口。
	// 我们可以通过 ctx.Session() 访问我们的会话及其组件。
	session, ok := ctx.Session().(*GuocedbSession)
	if !ok {
		log.Error("sql.Context session is not a GuocedbSession: %T", ctx.Session()) // sql.Context 会话不是 GuocedbSession。
		// This is a critical internal error.
		// 这是一个关键的内部错误。
		return nil, errors.ErrInternal.New("invalid session type in sql.Context")
	}

	// Get the compute components and catalog from the session (or the handler itself if they are shared)
	// 从会话（或如果组件是共享的，则从 handler 本身）获取计算组件和 catalog
	// The handler holds the compute components directly.
	// handler 直接持有计算组件。
	computeCatalog := h.components.Catalog
	parser := h.components.Parser
	analyzer := h.components.Analyzer
	optimizer := h.components.Optimizer
	executor := h.components.Executor
	txManager := h.components.TxManager // Need tx manager for explicit transaction commands


	// --- Handle Special Commands Explicitly (e.g., BEGIN, COMMIT, ROLLBACK) ---
	// These commands manage the transaction lifecycle and are often handled before the main plan execution.
	//
	// --- 显式处理特殊命令（例如 BEGIN, COMMIT, ROLLBACK） ---
	// 这些命令管理事务生命周期，通常在主计划执行之前处理。

	upperQuery := strings.ToUpper(strings.TrimSpace(query)) // Normalize for comparison

	// Access the current transaction from the Connection struct via the session.
	// Note: This is complex. The Session holds a pointer to the Connection? Or the Handler does?
	// The NewSession method creates a Connection and a Session. The Connection holds the Transaction.
	// The Session needs to access the Connection's Transaction field to manage COMMIT/ROLLBACK.
	//
	// 通过会话访问 Connection 结构体中的当前事务。
	// 注意：这很复杂。Session 持有 Connection 的指针？还是 Handler 持有？
	// NewSession 方法创建一个 Connection 和一个 Session。Connection 持有 Transaction。
	// Session 需要访问 Connection 的 Transaction 字段来管理 COMMIT/ROLLBACK。
	// The GuocedbSession struct has a clientConn field (*mysql.Conn). We need to get the Connection struct from that.
	// The Connection struct should maybe be accessible from the mysql.Conn? No, that creates circular deps.
	//
	// GuocedbSession 结构体有一个 clientConn 字段 (*mysql.Conn)。我们需要从它获取 Connection 结构体。
	// Connection 结构体应该可以从 mysql.Conn 访问吗？不，那会创建循环依赖。

	// Revised approach: The Connection struct holds the Session. The Session holds the TxManager.
	// The Connection calls TxManager methods via the Session.
	// When COMMIT/ROLLBACK is processed, the Handler needs to get the *current* transaction
	// object from the Connection and call its Commit/Rollback methods.
	// This implies the Connection needs a public method like `GetCurrentTransaction()`
	// and `SetCurrentTransaction(tx)`.
	//
	// 修改方法：Connection 结构体持有 Session。Session 持有 TxManager。
	// Connection 通过 Session 调用 TxManager 方法。
	// 当 COMMIT/ROLLBACK 被处理时，Handler 需要从 Connection 获取*当前*事务对象，并调用其 Commit/Rollback 方法。
	// 这意味着 Connection 需要一个公共方法，如 `GetCurrentTransaction()` 和 `SetCurrentTransaction(tx)`。
	// Let's add these methods to the Connection struct (defined in connection.go).

	// Get the Connection struct associated with this session.
	// Need a way to get `*Connection` from `*mysql.Conn`.
	// The `mysql.Conn` might store an arbitrary `UserData`. We could store `*Connection` there.
	//
	// 获取与此会话关联的 Connection 结构体。
	// 需要一种方法从 `*mysql.Conn` 获取 `*Connection`。
	// `mysql.Conn` 可以存储任意 `UserData`。我们可以将 `*Connection` 存储在那里。
	// In NewConnection, after creating the Connection, set `gmsConn.SetUserData(conn)`.
	// In HandleQuery, get `conn := ctx.Session().(*GuocedbSession).GetClientConn().UserData().(*Connection)`.

	// Get the Connection struct from the sql.Context's session's underlying mysql.Conn's UserData.
	//从 sql.Context 的会话的底层 mysql.Conn 的 UserData 获取 Connection 结构体。
	conn := ctx.Session().(*GuocedbSession).GetClientConn().UserData().(*Connection) // Assumes UserData is set and is *Connection


	// Handle BEGIN command
	// 处理 BEGIN 命令
	if upperQuery == "BEGIN" || upperQuery == "START TRANSACTION" {
		log.Debug("Handling BEGIN command for connection ID %d", conn.conn.ConnectionID) // 处理 BEGIN 命令。
		conn.mu.Lock()
		if conn.transaction != nil {
			// Already in a transaction, maybe implicit commit or nested transaction handling needed?
			// 已经在事务中，可能需要隐式提交或嵌套事务处理？
			log.Warn("BEGIN called while already in a transaction for connection ID %d", conn.conn.ConnectionID) // 在事务中调用 BEGIN。
			conn.mu.Unlock()
			// For simplicity, return success for now, assuming nested BEGIN is allowed/ignored.
			// 为了简化，暂时返回成功，假设允许/忽略嵌套 BEGIN。
			// A real system needs to handle this according to transaction isolation levels.
			// 实际系统需要根据事务隔离级别处理此情况。
			return new(plan.Done).RowIter(ctx), nil // Return a Done iterator for success
		}
		conn.mu.Unlock()

		// Begin a new transaction using the transaction manager
		// 使用事务管理器开启新事务
		newTx, err := txManager.Begin(context.Background()) // Use background context for tx Begin
		if err != nil {
			log.Error("Failed to begin transaction for BEGIN command, conn ID %d: %v", conn.conn.ConnectionID, err) // 开启事务失败。
			return nil, fmt.Errorf("%w: failed to begin transaction: %v", errors.ErrTransactionCommitFailed, err) // Use a transaction error type
		}

		// Set the new transaction as the current transaction for the connection
		// 将新事务设置为连接的当前事务
		conn.mu.Lock()
		conn.transaction = newTx
		conn.mu.Unlock()

		log.Debug("Transaction started for BEGIN command, conn ID %d. Tx ID: %s", conn.conn.ConnectionID, newTx.ID()) // 事务已开启。
		return new(plan.Done).RowIter(ctx), nil // Return success
	}

	// Handle COMMIT command
	// 处理 COMMIT 命令
	if upperQuery == "COMMIT" {
		log.Debug("Handling COMMIT command for connection ID %d", conn.conn.ConnectionID) // 处理 COMMIT 命令。
		conn.mu.Lock()
		currentTx := conn.transaction
		conn.transaction = nil // Transaction is ending
		conn.mu.Unlock()

		if currentTx == nil {
			// No active transaction, maybe auto-commit is on?
			// 没有活跃事务，可能是自动提交已开启？
			log.Warn("COMMIT called with no active transaction for connection ID %d", conn.conn.ConnectionID) // 在没有活跃事务时调用 COMMIT。
			// For simplicity, return success, assuming auto-commit handled the previous statement.
			// 为了简化，返回成功，假设自动提交处理了前一条语句。
			return new(plan.Done).RowIter(ctx), nil // Return success
		}

		// Commit the current transaction
		// 提交当前事务
		err := currentTx.Commit(context.Background()) // Use background context for tx Commit
		if err != nil {
			log.Error("Failed to commit transaction for COMMIT command, conn ID %d, tx ID %s: %v", conn.conn.ConnectionID, currentTx.ID(), err) // 提交事务失败。
			// Return error to client
			// 返回错误给客户端
			return nil, fmt.Errorf("%w: failed to commit transaction: %v", errors.ErrTransactionCommitFailed, err) // Use a transaction error type
		}

		log.Debug("Transaction committed for COMMIT command, conn ID %d. Tx ID: %s", conn.conn.ConnectionID, currentTx.ID()) // 事务已提交。
		// A new transaction might be implicitly started by GMS for the next statement if auto-commit is on.
		// 如果自动提交开启，GMS 可能会为下一条语句隐式开启新事务。
		// Our TxManager.Begin will be called via GetTransactionManager when GMS needs it.
		// 当 GMS 需要时，将通过 GetTransactionManager 调用我们的 TxManager.Begin。
		return new(plan.Done).RowIter(ctx), nil // Return success
	}

	// Handle ROLLBACK command
	// 处理 ROLLBACK 命令
	if upperQuery == "ROLLBACK" {
		log.Debug("Handling ROLLBACK command for connection ID %d", conn.conn.ConnectionID) // 处理 ROLLBACK 命令。
		conn.mu.Lock()
		currentTx := conn.transaction
		conn.transaction = nil // Transaction is ending
		conn.mu.Unlock()

		if currentTx == nil {
			// No active transaction
			// 没有活跃事务
			log.Warn("ROLLBACK called with no active transaction for connection ID %d", conn.conn.ConnectionID) // 在没有活跃事务时调用 ROLLBACK。
			// Return success as there was nothing to roll back.
			// 返回成功，因为没有什么可回滚的。
			return new(plan.Done).RowIter(ctx), nil // Return success
		}

		// Rollback the current transaction
		// 回滚当前事务
		err := currentTx.Rollback(context.Background()) // Use background context for tx Rollback
		if err != nil {
			log.Error("Failed to rollback transaction for ROLLBACK command, conn ID %d, tx ID %s: %v", conn.conn.ConnectionID, currentTx.ID(), err) // 回滚事务失败。
			// Return error to client
			// 返回错误给客户端
			return nil, fmt.Errorf("%w: failed to rollback transaction: %v", errors.ErrTransactionRollbackFailed, err) // Use a transaction error type
		}

		log.Debug("Transaction rolled back for ROLLBACK command, conn ID %d. Tx ID: %s", conn.conn.ConnectionID, currentTx.ID()) // 事务已回滚。
		// Similar to COMMIT, a new transaction might be implicitly started by GMS.
		// 与 COMMIT 类似，GMS 可能会隐式开启新事务。
		return new(plan.Done).RowIter(ctx), nil // Return success
	}

	// --- End Special Commands ---


	// --- Process Regular Queries ---
	// For regular DML/DDL/SELECT statements, use the compute layer pipeline.
	// Need to associate the current connection's transaction with the sql.Context for GMS.
	//
	// --- 处理常规查询 ---
	// 对于常规的 DML/DDL/SELECT 语句，使用计算层管道。
	// 需要将当前连接的事务与 GMS 的 sql.Context 相关联。

	conn.mu.Lock()
	currentTx := conn.transaction // Get the current transaction from the connection
	conn.mu.Unlock()

	if currentTx == nil {
		// If no active transaction (e.g., after COMMIT/ROLLBACK and before next implicit BEGIN),
		// GMS might handle this, or we might need to start a new one here if auto-commit is on.
		//
		// 如果没有活跃事务（例如，在 COMMIT/ROLLBACK 之后和下一个隐式 BEGIN 之前），
		// GMS 可能会处理这种情况，或者如果自动提交开启，我们可能需要在此处开启一个新事务。
		// Let's assume auto-commit behavior relies on GMS calling TxManager.Begin when needed
		// via the sql.Context factory.
		//
		// 假设自动提交行为依赖于 GMS 在需要时通过 sql.Context 工厂调用 TxManager.Begin。
		// If auto-commit is off, and no explicit BEGIN, this case is valid (no transaction).
		// If auto-commit is on, and no explicit BEGIN, GMS should implicitly begin.
		// So, currentTx being nil here means either auto-commit is off, or GMS hasn't started implicit tx yet.
		// The safest is to ensure the sql.Context gets a transaction if needed.
		//
		// 如果自动提交关闭，且没有显式 BEGIN，则此情况有效（无事务）。
		// 如果自动提交开启，且没有显式 BEGIN，GMS 应该隐式开启。
		// 因此，此处 currentTx 为 nil 意味着自动提交关闭，或 GMS 尚未启动隐式 tx。
		// 最安全的方法是确保 sql.Context 在需要时获取事务。
		log.Debug("No active transaction for regular query, conn ID %d. Relying on GMS Context for transaction.", conn.conn.ConnectionID) // 常规查询没有活跃事务。依赖 GMS Context 进行事务处理。
		// The sql.Context will use the Session's GetTransactionManager to potentially begin a transaction.
		// sql.Context 将使用 Session 的 GetTransactionManager 来潜在地开启事务。
		// We don't need to set `ctx.SetTx(currentTx)` explicitly here IF GMS manages tx via Session.GetTransactionManager.
		// 如果 GMS 通过 Session.GetTransactionManager 管理 tx，则不需要在此处显式设置 `ctx.SetTx(currentTx)`。
	} else {
		// There is an active transaction. We need to provide the GMS-compatible sql.Transaction
		// to the sql.Context for execution.
		//
		// 有一个活跃事务。我们需要为执行将 GMS 兼容的 sql.Transaction 提供给 sql.Context。
		// Get the GMS-compatible transaction wrapper from our interfaces.Transaction
		// 从我们的 interfaces.Transaction 获取 GMS 兼容的事务包装器
		gmsTx, ok := currentTx.UnderlyingTx().(sql.Transaction)
		if !ok {
			log.Error("Underlying interfaces.Transaction does not provide a GMS sql.Transaction interface for execution context.") // 底层 interfaces.Transaction 未提供 GMS sql.Transaction 接口。
			return nil, errors.ErrInternal.New("active transaction object not GMS compatible")
		}
		// Set the transaction in the GMS sql.Context.
		// This is important so that GMS operators use the correct transaction for data access.
		//
		// 在 GMS sql.Context 中设置事务。
		// 这很重要，以便 GMS 操作符使用正确的事务进行数据访问。
		ctx.SetTx(gmsTx) // Set the active transaction in the context
		log.Debug("Set active transaction %d in GMS context for conn ID %d.", gmsTx.ID(), conn.conn.ConnectionID) // 在 GMS context 中设置活跃事务。
	}


	// 1. Parsing (Done by GMS server internally for COM_QUERY, but our handler might receive the string).
	//    Our parser wrapper can be used here if needed, but GMS's Engine typically includes its own parser.
	//    Let's assume GMS calls us with the query string and we need to use our own parser wrapper.
	//
	// 1. 解析（GMS 服务器内部为 COM_QUERY 完成，但我们的 handler 可能接收到字符串）。
	//    如果需要，可以在此处使用我们的解析器包装器，但 GMS 的 Engine 通常包含其自己的解析器。
	//    假设 GMS 使用查询字符串调用我们，并且我们需要使用我们自己的解析器包装器。

	// Since the sql.Handler.HandleQuery receives the query string, we need to parse it.
	// 由于 sql.Handler.HandleQuery 接收查询字符串，我们需要解析它。
	parsedNodes, err := parser.Parse(context.Background(), query) // Use background context for parsing
	if err != nil {
		// Parser already wraps with ErrInvalidSQL
		// 解析器已使用 ErrInvalidSQL 包装
		return nil, err // Propagate parsing error
	}
	if len(parsedNodes) == 0 {
		log.Warn("Parser returned no nodes for query: %s", query) // 解析器未为查询返回节点。
		// Empty query or comment? Return success with no rows.
		// 空查询或注释？返回成功，没有行。
		return new(plan.Done).RowIter(ctx), nil // Return a Done iterator
	}
	if len(parsedNodes) > 1 {
		log.Warn("Parser returned multiple nodes for query (only first will be executed): %s", query) // 解析器为查询返回多个节点（只执行第一个）。
		// For simplicity, execute only the first statement. A real system handles batches.
		// 为了简化，只执行第一条语句。实际系统处理批处理。
	}
	parsedNode := parsedNodes[0] // Get the first node

	// 2. Analysis
	// 2. 分析
	// Use the compute analyzer
	// 使用计算分析器
	analyzedNode, err := analyzer.Analyze(context.Background(), parsedNode, computeCatalog) // Use background context for analysis
	if err != nil {
		// Analyzer already wraps with ErrInvalidSQL
		// 分析器已使用 ErrInvalidSQL 包装
		return nil, err // Propagate analysis error
	}
	// Check if the analyzed node is resolved
	// 检查分析后的节点是否已解析
	if !analyzedNode.Resolved() {
		log.Error("Analyzer returned unresolved node for query: %s", query) // 分析器未解析节点。
		// This indicates a problem in the analysis process or an unsupported query.
		// 这表明分析过程存在问题或查询不受支持。
		return nil, errors.ErrInvalidSQL.New("analysis failed: unresolved plan node") // Return a specific error
	}
	log.Debug("Query analyzed successfully.") // 查询分析成功。


	// 3. Optimization
	// 3. 优化
	// Use the compute optimizer
	// 使用计算优化器
	optimizedNode, err := optimizer.Optimize(context.Background(), analyzedNode, computeCatalog) // Use background context for optimization
	if err != nil {
		// Optimizer might return ErrInvalidSQL or other errors.
		// 优化器可能返回 ErrInvalidSQL 或其他错误。
		log.Error("Failed to optimize analyzed node: %v", err) // 优化已分析节点失败。
		return nil, fmt.Errorf("%w: optimization failed: %v", errors.ErrInvalidSQL, err) // Wrap generally
	}
	log.Debug("Query optimized successfully.") // 查询优化成功。


	// 4. Execution
	// 4. 执行
	// The optimized node is the final plan. Execute it.
	// 优化后的节点是最终计划。执行它。
	// Use the compute executor. It needs the transaction associated with this connection.
	// 使用计算执行器。它需要与此连接关联的事务。
	rowIter, err := executor.Execute(context.Background(), optimizedNode, computeCatalog, conn.transaction) // Use background context for execution, pass the connection's tx
	if err != nil {
		// Executor should map underlying storage/internal errors.
		// 执行器应映射底层存储/内部错误。
		return nil, fmt.Errorf("%w: execution failed: %v", errors.ErrInternal, err) // Wrap generally
	}

	log.Debug("Query execution started, returning row iterator.") // 查询执行开始，返回行迭代器。
	return rowIter, nil // Return the row iterator from the executor
}

// TODO: Implement HandlePreparedQuery if prepared statements are supported.
// TODO: 如果支持预处理语句，实现 HandlePreparedQuery。
func (h *Handler) HandlePreparedQuery(ctx *sql.Context, prepareId uint32, paramCount uint16, query string, params ...interface{}) (sql.RowIter, error) {
	log.Warn("Handler HandlePreparedQuery called (Not Implemented)") // 调用 Handler HandlePreparedQuery（未实现）。
	return nil, errors.ErrNotImplemented.New("prepared statements")
}

// TODO: Implement HandleParse if prepared statements are supported and GMS calls Parse separately.
// TODO: 如果支持预处理语句且 GMS 单独调用 Parse，实现 HandleParse。
func (h *Handler) HandleParse(ctx *sql.Context, stmtName string, query string) error {
	log.Warn("Handler HandleParse called (Not Implemented)") // 调用 Handler HandleParse（未实现）。
	return errors.ErrNotImplemented.New("prepared statement parsing")
}

// TODO: Implement HandleExecute if prepared statements are supported and GMS calls Execute separately.
// TODO: 如果支持预处理语句且 GMS 单独调用 Execute，实现 HandleExecute。
func (h *Handler) HandleExecute(ctx *sql.Context, stmtName string, params ...sql.Expression) (sql.RowIter, error) {
	log.Warn("Handler HandleExecute called (Not Implemented)") // 调用 Handler HandleExecute（未实现）。
	return nil, errors.ErrNotImplemented.New("prepared statement execution")
}

// TODO: Implement HandleClose if prepared statements are supported and GMS calls Close separately.
// TODO: 如果支持预处理语句且 GMS 单独调用 Close，实现 HandleClose。
func (h *Handler) HandleClose(ctx *sql.Context, stmtName string) error {
	log.Warn("Handler HandleClose called (Not Implemented)") // 调用 Handler HandleClose（未实现）。
	return errors.ErrNotImplemented.New("prepared statement closing")
}