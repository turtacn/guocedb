// Package audit contains the auditing subsystem.
// It is responsible for logging security-relevant events for monitoring and compliance.
//
// audit 包包含审计子系统。
// 它负责记录与安全相关的事件，以便进行监视和合规性检查。
package audit

import (
	"context"
	"time"
	"fmt"

	"github.com/turtacn/guocedb/common/log" // Leverage the common logging system
	"github.com/turtacn/guocedb/security/authn" // Import authentication identity
	"github.com/turtacn/guocedb/security/authz" // Import authorization types
)

// EventType represents the type of auditable event.
// EventType 表示可审计事件的类型。
type EventType string

// Define common auditable event types
// 定义常见的可审计事件类型
const (
	EventTypeLoginSuccess    EventType = "LOGIN_SUCCESS"
	EventTypeLoginFailure    EventType = "LOGIN_FAILURE"
	EventTypePermissionDenied EventType = "PERMISSION_DENIED"
	EventTypeDDL             EventType = "DDL" // Data Definition Language (CREATE, DROP, ALTER)
	EventTypeDML             EventType = "DML" // Data Manipulation Language (INSERT, UPDATE, DELETE)
	EventTypeQuery           EventType = "QUERY" // Read queries (SELECT)
	EventTypeAdmin           EventType = "ADMIN_OPERATION" // Admin commands
	// TODO: Add more event types
	// TODO: 添加更多事件类型
)

// Event represents a single auditable event.
// Event 表示单个可审计事件。
type Event struct {
	Timestamp time.Time       // When the event occurred / 事件发生时间
	Type      EventType       // Type of event / 事件类型
	Identity  authn.Identity  // The user or entity that performed the action / 执行操作的用户或实体
	Action    authz.Action    // The specific action attempted (relevant for DDL, DML, PERMISSION_DENIED) / 尝试的特定操作
	Resource  authz.Resource  // The resource affected by the action (relevant for DDL, DML, PERMISSION_DENIED) / 受操作影响的资源
	Query     string          // The SQL query text (relevant for DDL, DML, QUERY) / SQL 查询文本
	Success   bool            // Whether the action was successful / 操作是否成功
	Error     error           // Error details if the action failed / 如果操作失败，提供错误详情
	Details   map[string]interface{} // Additional event details / 其他事件详情
	// TODO: Add fields for client IP, connection ID, etc.
	// TODO: 添加客户端 IP、连接 ID 等字段。
}


// Auditor is the interface for the auditing subsystem.
// Auditor 是审计子系统的接口。
type Auditor interface {
	// LogEvent records a single auditable event.
	// Implementations might write to a file, database, or send to a remote service.
	//
	// LogEvent 记录单个可审计事件。
	// 实现可以将事件写入文件、数据库或发送到远程服务。
	LogEvent(ctx context.Context, event Event) error

	// TODO: Add methods for configuring auditing, filtering events, etc.
	// TODO: 添加用于配置审计、过滤事件等的方法。
}

// TODO: Implement concrete Auditor implementations here,
// such as FileAuditor, DatabaseAuditor, etc.
//
// TODO: 在此处实现具体的 Auditor 实现，
// 例如 FileAuditor、DatabaseAuditor 等。

// DefaultAuditor logs audit events using the common logging system.
// This is a basic implementation for demonstration.
//
// DefaultAuditor 使用通用日志系统记录审计事件。
// 这是用于演示的基本实现。
type DefaultAuditor struct {
	// Uses the common logger
	// 使用通用日志记录器
}

// NewDefaultAuditor creates a new DefaultAuditor instance.
// NewDefaultAuditor 创建一个新的 DefaultAuditor 实例。
func NewDefaultAuditor() Auditor {
	log.Info("Initializing default auditor (logging to common logger).") // 初始化默认审计器（记录到通用日志记录器）。
	return &DefaultAuditor{}
}

// LogEvent logs an auditable event using the common logging system.
// LogEvent 使用通用日志系统记录可审计事件。
func (a *DefaultAuditor) LogEvent(ctx context.Context, event Event) error {
	// Format the event details into a log message.
	// 将事件详情格式化为日志消息。
	logMsg := fmt.Sprintf("AUDIT: Type=%s, User='%s', Success=%t", event.Type, event.Identity.GetUsername(), event.Success)

	if event.Action != "" {
		logMsg += fmt.Sprintf(", Action='%s'", event.Action)
	}
	if event.Resource != nil {
		logMsg += fmt.Sprintf(", Resource='%s'", event.Resource.String())
	}
	if event.Query != "" {
		// Truncate long queries for logging
		// 截断长查询以进行日志记录
		queryLog := event.Query
		if len(queryLog) > 100 { // Example limit
			queryLog = queryLog[:100] + "..."
		}
		logMsg += fmt.Sprintf(", Query='%s'", queryLog)
	}
	if event.Error != nil {
		logMsg += fmt.Sprintf(", Error='%v'", event.Error)
	}
	if len(event.Details) > 0 {
		// Simple logging of details map
		// 简单记录详情 map
		logMsg += fmt.Sprintf(", Details=%+v", event.Details)
	}

	// Use the common logger to log the formatted message.
	// 使用通用日志记录器记录格式化的消息。
	// Choose log level based on event type or success/failure.
	// 根据事件类型或成功/失败选择日志级别。
	if !event.Success {
		log.Error(logMsg) // Log failures as errors
	} else if event.Type == EventTypeLoginSuccess || event.Type == EventTypeDDL {
		log.Info(logMsg) // Log important successful events as info
	} else {
		log.Debug(logMsg) // Log other successful events as debug
	}

	// In this simple implementation, logging to the common logger always succeeds unless the logger itself fails.
	// 在这个简单的实现中，除非日志记录器本身失败，否则记录到通用日志记录器总是成功的。
	return nil // Assuming logging succeeded
}