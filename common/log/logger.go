// Package log provides a unified logging interface and implementation.
// log 包提供了统一的日志接口和实现。
package log

import (
	"log" // Using standard log for simplicity initially
	"os"
)

// Logger is the unified logging interface.
// Logger 是统一的日志接口。
type Logger interface {
	Debug(format string, v ...interface{}) // Debug logs debug messages. / Debug 记录调试消息。
	Info(format string, v ...interface{})  // Info logs informational messages. / Info 记录信息消息。
	Warn(format string, v ...interface{})  // Warn logs warning messages. / Warn 记录警告消息。
	Error(format string, v ...interface{}) // Error logs error messages. / Error 记录错误消息。
	Fatal(format string, v ...interface{}) // Fatal logs fatal messages and exits. / Fatal 记录致命消息并退出。
	Printf(format string, v ...interface{}) // Printf logs formatted messages. / Printf 记录格式化消息。
}

// globalLogger is the package-level logger instance.
// globalLogger 是包级别的日志实例。
var globalLogger Logger = &standardLogger{} // Default to standard logger

// SetLogger sets the global logger instance.
// SetLogger 设置全局日志实例。
func SetLogger(l Logger) {
	globalLogger = l
}

// Debug logs a debug message using the global logger.
// Debug 使用全局日志记录器记录调试消息。
func Debug(format string, v ...interface{}) {
	globalLogger.Debug(format, v...)
}

// Info logs an info message using the global logger.
// Info 使用全局日志记录器记录信息消息。
func Info(format string, v ...interface{}) {
	globalLogger.Info(format, v...)
}

// Warn logs a warning message using the global logger.
// Warn 使用全局日志记录器记录警告消息。
func Warn(format string, v ...interface{}) {
	globalLogger.Warn(format, v...)
}

// Error logs an error message using the global logger.
// Error 使用全局日志记录器记录错误消息。
func Error(format string, v ...interface{}) {
	globalLogger.Error(format, v...)
}

// Fatal logs a fatal message using the global logger and exits.
// Fatal 使用全局日志记录器记录致命消息并退出。
func Fatal(format string, v ...interface{}) {
	globalLogger.Fatal(format, v...)
}

// Printf logs a formatted message using the global logger.
// Printf 使用全局日志记录器记录格式化消息。
func Printf(format string, v ...interface{}) {
	globalLogger.Printf(format, v...)
}

// standardLogger is a simple implementation using Go's standard log package.
// standardLogger 是一个使用 Go 标准 log 包的简单实现。
type standardLogger struct{}

// Debug logs a debug message.
// Debug 记录调试消息。
func (l *standardLogger) Debug(format string, v ...interface{}) {
	// Standard log doesn't have levels easily. Prefix manually.
	// 标准日志不容易实现分级。手动添加前缀。
	log.Printf("[DEBUG] "+format, v...)
}

// Info logs an info message.
// Info 记录信息消息。
func (l *standardLogger) Info(format string, v ...interface{}) {
	log.Printf("[INFO] "+format, v...)
}

// Warn logs a warning message.
// Warn 记录警告消息。
func (l *standardLogger) Warn(format string, v ...interface{}) {
	log.Printf("[WARN] "+format, v...)
}

// Error logs an error message.
// Error 记录错误消息。
func (l *standardLogger) Error(format string, v ...interface{}) {
	log.Printf("[ERROR] "+format, v...)
}

// Fatal logs a fatal message and exits.
// Fatal 记录致命消息并退出。
func (l *standardLogger) Fatal(format string, v ...interface{}) {
	log.Fatalf("[FATAL] "+format, v...)
}

// Printf logs a formatted message.
// Printf 记录格式化消息。
func (l *standardLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// InitLogger initializes the logger based on configuration (future).
// InitLogger 根据配置初始化日志记录器（未来）。
func InitLogger() {
	// TODO: Implement logger initialization based on configuration (e.g., log file, level).
	// Currently uses standard output and fixed levels.
	//
	// TODO: 根据配置（例如日志文件、级别）实现日志记录器初始化。
	// 当前使用标准输出和固定级别。
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	Info("Logger initialized using standard log.") // 使用标准日志初始化记录器。
}