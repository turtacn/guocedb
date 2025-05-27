package log

import (
	"fmt"
	"log" // Standard Go logger package // 标准 Go 日志包
	"os"
	"strings"
	"sync" // Needed if SetGlobalLogger can be called concurrently // 如果 SetGlobalLogger 可能被并发调用则需要
)

// LogLevel defines the logging severity levels.
// LogLevel 定义日志严重性级别。
type LogLevel int

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	// DebugLevel 日志通常量很大，并且通常在生产环境中禁用。
	DebugLevel LogLevel = iota
	// InfoLevel is the default logging priority.
	// InfoLevel 是默认的日志记录优先级。
	InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	// WarnLevel 日志比 Info 更重要，但不需要单独的人工审查。
	WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	// ErrorLevel 日志是高优先级的。如果应用程序运行平稳，则不应生成任何错误级别的日志。
	ErrorLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	// FatalLevel 记录一条消息，然后调用 os.Exit(1)。
	FatalLevel
	// DisabledLevel disables logging.
	// DisabledLevel 禁用日志记录。
	DisabledLevel
)

// Logger interface defines the standard logging methods for GuoceDB.
// All log methods ending in 'f' accept a format string and arguments, similar to fmt.Printf.
// The With method allows adding structured context (key-value pairs) to the logger.
// Logger 接口定义了 GuoceDB 的标准日志记录方法。
// 所有以 'f' 结尾的日志方法都接受格式字符串和参数，类似于 fmt.Printf。
// With 方法允许向日志记录器添加结构化上下文（键值对）。
type Logger interface {
	// Debugf logs messages useful for debugging.
	// Debugf 记录对调试有用的消息。
	Debugf(format string, args ...interface{})
	// Infof logs general informational messages.
	// Infof 记录常规信息性消息。
	Infof(format string, args ...interface{})
	// Warnf logs warnings that might indicate potential issues.
	// Warnf 记录可能指示潜在问题的警告。
	Warnf(format string, args ...interface{})
	// Errorf logs errors that indicate problems.
	// Errorf 记录指示问题的错误。
	Errorf(format string, args ...interface{})
	// Fatalf logs an error message and then exits the application (os.Exit(1)).
	// Fatalf 记录错误消息，然后退出应用程序 (os.Exit(1))。
	Fatalf(format string, args ...interface{})

	// With returns a new logger instance with the specified key-value pairs added as context.
	// Keys should ideally be strings. Alternating keys and values are expected.
	// With 返回一个新的日志记录器实例，其中添加了指定的键值对作为上下文。
	// 键最好是字符串。期望交替出现键和值。
	With(args ...interface{}) Logger

	// GetLevel returns the current logging level.
	// GetLevel 返回当前的日志记录级别。
	GetLevel() LogLevel
}

// --- Global Logger ---

// globalLogger holds the current logger instance for the application.
// It defaults to a no-op logger until configured otherwise.
// Access is protected by a mutex to allow safe concurrent calls to SetGlobalLogger,
// although typically it's set only once at startup.
// globalLogger 保存应用程序的当前日志记录器实例。
// 在配置之前，它默认为无操作日志记录器。
// 访问受互斥锁保护，以允许对 SetGlobalLogger 进行安全的并发调用，
// 尽管通常它仅在启动时设置一次。
var (
	globalLogger Logger = &noopLogger{} // Default to no-op // 默认为无操作
	globalMu     sync.RWMutex
)

// SetGlobalLogger replaces the default logger with the provided one.
// This should typically be called once during application initialization before
// significant concurrent logging begins. It is safe for concurrent use.
// SetGlobalLogger 用提供的日志记录器替换默认日志记录器。
// 通常应在应用程序初始化期间，在大量并发日志记录开始之前调用一次。
// 它可以安全地并发使用。
func SetGlobalLogger(logger Logger) {
	globalMu.Lock()
	defer globalMu.Unlock()
	if logger == nil {
		// Fallback to no-op logger if nil is provided
		// 如果提供 nil，则回退到无操作日志记录器
		globalLogger = &noopLogger{}
		return
	}
	globalLogger = logger
}

// GetLogger returns the currently configured global logger.
// Useful if direct access to the logger instance (e.g., for With) is needed.
// It is safe for concurrent use.
// GetLogger 返回当前配置的全局日志记录器。
// 如果需要直接访问日志记录器实例（例如，用于 With），则很有用。
// 它可以安全地并发使用。
func GetLogger() Logger {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalLogger
}

// --- Convenience Functions ---
// These functions call the corresponding methods on the global logger.
// --- 便捷函数 ---
// 这些函数调用全局日志记录器上的相应方法。

// Debugf logs a message at DebugLevel using the global logger.
// Debugf 使用全局日志记录器在 DebugLevel 记录一条消息。
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...) // Use GetLogger for safe access // 使用 GetLogger 进行安全访问
}

// Infof logs a message at InfoLevel using the global logger.
// Infof 使用全局日志记录器在 InfoLevel 记录一条消息。
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warnf logs a message at WarnLevel using the global logger.
// Warnf 使用全局日志记录器在 WarnLevel 记录一条消息。
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Errorf logs a message at ErrorLevel using the global logger.
// Errorf 使用全局日志记录器在 ErrorLevel 记录一条消息。
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatalf logs a message at FatalLevel using the global logger and exits.
// Fatalf 使用全局日志记录器在 FatalLevel 记录一条消息并退出。
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

// With creates a new logger instance derived from the global logger,
// adding the specified key-value pairs as context.
// With 从全局日志记录器派生出一个新的日志记录器实例，
// 添加指定的键值对作为上下文。
func With(args ...interface{}) Logger {
	return GetLogger().With(args...)
}

// --- No-op Logger Implementation ---
// Implements the Logger interface but performs no actions.
// --- 无操作日志记录器实现 ---
// 实现 Logger 接口，但不执行任何操作。

type noopLogger struct{}

func (l *noopLogger) Debugf(format string, args ...interface{}) {}
func (l *noopLogger) Infof(format string, args ...interface{})  {}
func (l *noopLogger) Warnf(format string, args ...interface{})  {}
func (l *noopLogger) Errorf(format string, args ...interface{}) {}
func (l *noopLogger) Fatalf(format string, args ...interface{}) {
	// Even a no-op logger should respect Fatal's exit behavior,
	// though it won't log the message.
	// 即使是无操作日志记录器也应遵守 Fatal 的退出行为，
	// 尽管它不会记录消息。
	os.Exit(1)
}
func (l *noopLogger) With(args ...interface{}) Logger {
	return l // Return self, as context is ignored // 返回自身，因为上下文被忽略
}
func (l *noopLogger) GetLevel() LogLevel {
	return DisabledLevel // No-op implies disabled level // 无操作意味着禁用级别
}

// --- Standard Library Logger Implementation ---
// Provides a basic Logger implementation using Go's standard `log` package.
// --- 标准库日志记录器实现 ---
// 使用 Go 的标准 `log` 包提供基本的 Logger 实现。

// stdLogger wraps the standard Go log.Logger.
// stdLogger 包装标准的 Go log.Logger。
type stdLogger struct {
	stdlog *log.Logger // Underlying standard logger // 底层标准记录器
	level  LogLevel    // Minimum level to log // 要记录的最低级别
	// Store context added via With as key-value pairs
	// 将通过 With 添加的上下文存储为键值对
	context []interface{}
}

// NewStdLogger creates a new Logger instance that writes to os.Stderr
// with the specified minimum logging level.
// NewStdLogger 创建一个新的 Logger 实例，该实例使用指定的最低日志记录级别
// 写入 os.Stderr。
func NewStdLogger(level LogLevel) Logger {
	// Default flags: Date, Time, Microseconds, Short file name (caller)
	// 默认标志：日期、时间、微秒、短文件名（调用者）
	flags := log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile
	return &stdLogger{
		stdlog: log.New(os.Stderr, "", flags),
		level:  level,
	}
}

// NewStdLoggerWithWriter creates a new Logger instance that writes to the specified writer.
// NewStdLoggerWithWriter 创建一个新的 Logger 实例，该实例写入指定的写入器。
func NewStdLoggerWithWriter(writer *os.File, level LogLevel) Logger {
	flags := log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile
	return &stdLogger{
		stdlog: log.New(writer, "", flags),
		level:  level,
	}
}

// Helper to format context and message
// 格式化上下文和消息的辅助函数
func (l *stdLogger) formatMsg(levelStr string, format string, args ...interface{}) string {
	msg := fmt.Sprintf(format, args...)
	if len(l.context) == 0 {
		return fmt.Sprintf("[%s] %s", levelStr, msg)
	}

	var ctxBuilder strings.Builder
	ctxBuilder.WriteString("[") // Start context block // 开始上下文块
	for i := 0; i < len(l.context); i += 2 {
		key := l.context[i]
		var val interface{} = "(MISSING)" // Placeholder if key only // 如果只有键则为占位符
		if i+1 < len(l.context) {
			val = l.context[i+1]
		}
		if i > 0 {
			ctxBuilder.WriteString(" ") // Separator // 分隔符
		}
		// Simple formatting, escape strings if needed?
		// 简单的格式化，如果需要是否转义字符串？
		// Use %q for strings to handle spaces and special chars
		// 对字符串使用 %q 来处理空格和特殊字符
		if s, ok := val.(string); ok {
			ctxBuilder.WriteString(fmt.Sprintf("%v=%q", key, s))
		} else {
			ctxBuilder.WriteString(fmt.Sprintf("%v=%v", key, val))
		}

	}
	ctxBuilder.WriteString("]") // End context block // 结束上下文块
	// Example: [INFO] Request received [reqID="abc" user=123]
	// 示例：[INFO] Request received [reqID="abc" user=123]
	return fmt.Sprintf("[%s] %s %s", levelStr, msg, ctxBuilder.String())
}

// log outputs the message if the level is sufficient. Calldepth is adjusted.
// log 如果级别足够，则输出消息。调整 Calldepth。
func (l *stdLogger) log(level LogLevel, levelStr string, format string, args ...interface{}) {
	if l.level <= level {
		// Calldepth 3: log() -> Debugf/Infof/... -> stdlog.Output()
		// Calldepth 3: log() -> Debugf/Infof/... -> stdlog.Output()
		l.stdlog.Output(3, l.formatMsg(levelStr, format, args...))
	}
}

func (l *stdLogger) Debugf(format string, args ...interface{}) {
	l.log(DebugLevel, "DEBUG", format, args...)
}

func (l *stdLogger) Infof(format string, args ...interface{}) {
	l.log(InfoLevel, "INFO", format, args...)
}

func (l *stdLogger) Warnf(format string, args ...interface{}) {
	l.log(WarnLevel, "WARN", format, args...)
}

func (l *stdLogger) Errorf(format string, args ...interface{}) {
	l.log(ErrorLevel, "ERROR", format, args...)
}

func (l *stdLogger) Fatalf(format string, args ...interface{}) {
	if l.level <= FatalLevel {
		// Calldepth 3: Fatalf() -> stdlog.Output()
		// Calldepth 3: Fatalf() -> stdlog.Output()
		l.stdlog.Output(3, l.formatMsg("FATAL", format, args...))
	}
	os.Exit(1) // Exit regardless of level check, following Fatal contract // 无论级别检查如何都退出，遵循 Fatal 约定
}

func (l *stdLogger) With(args ...interface{}) Logger {
	// Ensure args are key-value pairs (even length) - simple append here
	// 确保 args 是键值对（偶数长度）- 这里简单追加
	// A more robust implementation might validate keys are strings and length is even.
	// 更健壮的实现可能会验证键是字符串且长度是偶数。
	newContext := make([]interface{}, 0, len(l.context)+len(args))
	newContext = append(newContext, l.context...)
	newContext = append(newContext, args...)

	// Return a *new* logger instance with the combined context
	// 返回一个带有组合上下文的 *新* 日志记录器实例
	// Share the underlying log.Logger and level, but copy context
	// 共享底层的 log.Logger 和级别，但复制上下文
	return &stdLogger{
		stdlog:  l.stdlog,
		level:   l.level,
		context: newContext,
	}
}

func (l *stdLogger) GetLevel() LogLevel {
	return l.level
}

// --- Level Parsing ---
// --- 级别解析 ---

// StringToLevel converts a log level string (case-insensitive) to LogLevel.
// Returns InfoLevel if the string is not recognized.
// StringToLevel 将日志级别字符串（不区分大小写）转换为 LogLevel。
// 如果字符串无法识别，则返回 InfoLevel。
func StringToLevel(levelStr string) LogLevel {
	switch strings.ToLower(strings.TrimSpace(levelStr)) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	case "disabled", "none", "": // Treat empty string as disabled // 将空字符串视为禁用
		return DisabledLevel
	default:
		// Log a warning if an invalid level is provided?
		// 如果提供了无效级别，是否记录警告？
		// Using the default logger here might cause recursion if it's not set yet.
		// 在此处使用默认记录器可能会导致递归（如果尚未设置）。
		// fmt.Fprintf(os.Stderr, "Warning: Unrecognized log level '%s', defaulting to INFO\n", levelStr)
		return InfoLevel // Default to Info if unrecognized // 如果无法识别，则默认为 Info
	}
}

// LevelToString converts LogLevel to its string representation.
// LevelToString 将 LogLevel 转换为其字符串表示形式。
func LevelToString(level LogLevel) string {
	switch level {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	case DisabledLevel:
		return "DISABLED"
	default:
		return "UNKNOWN"
	}
}
