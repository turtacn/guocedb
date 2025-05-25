// Package log provides a unified logging interface and default implementation for Guocedb.
// This ensures consistent logging practices across the entire project, allowing for
// centralized configuration, output redirection, and easy integration with various
// logging backends.
//
// 此包为 Guocedb 提供了一个统一的日志接口和默认实现。
// 这确保了整个项目中日志实践的一致性，允许集中配置、输出重定向，
// 并方便与各种日志后端集成。
package log

import (
	"fmt"
	"io"
	"log"  // Standard Go logger
	"os"   // For file operations
	"sync" // For mutex to protect logger
	"time" // For timestamping logs

	"github.com/turtacn/guocedb/common/constants"  // For default log file path
	"github.com/turtacn/guocedb/common/types/enum" // For LogLevel enum
)

// Logger is the interface that defines the logging capabilities for Guocedb.
// Any component requiring logging should depend on this interface, not a concrete implementation.
//
// Logger 接口定义了 Guocedb 的日志记录能力。
// 任何需要日志记录的组件都应该依赖此接口，而不是具体的实现。
type Logger interface {
	// Debug logs a message at the DEBUG level.
	// 记录 DEBUG 级别的消息。
	Debug(format string, args ...interface{})
	// Info logs a message at the INFO level.
	// 记录 INFO 级别的消息。
	Info(format string, args ...interface{})
	// Warn logs a message at the WARN level.
	// 记录 WARN 级别的消息。
	Warn(format string, args ...interface{})
	// Error logs a message at the ERROR level.
	// 记录 ERROR 级别的消息。
	Error(format string, args ...interface{})
	// Fatal logs a message at the FATAL level, then exits the application.
	// 记录 FATAL 级别的消息，然后退出应用程序。
	Fatal(format string, args ...interface{})
	// SetLevel sets the current logging level. Messages below this level will be ignored.
	// 设置当前的日志级别。低于此级别的消息将被忽略。
	SetLevel(level enum.LogLevel)
	// GetLevel returns the current logging level.
	// 获取当前的日志级别。
	GetLevel() enum.LogLevel
}

// defaultLogger is a simple implementation of the Logger interface using the standard Go `log` package.
// It supports different logging levels and can write to a file or standard output.
//
// defaultLogger 是 Logger 接口的简单实现，使用了 Go 标准库的 `log` 包。
// 它支持不同的日志级别，并且可以写入文件或标准输出。
type defaultLogger struct {
	mu     sync.RWMutex  // Mutex to protect log level and output writer
	level  enum.LogLevel // Current active logging level
	logger *log.Logger   // Underlying standard Go logger
	output io.Writer     // Where the logs are written (e.g., os.Stdout, file)
	file   *os.File      // File handle if logging to a file
}

var (
	// globalLogger is the singleton instance of the default logger.
	// It's initialized once and used throughout the application.
	//
	// globalLogger 是默认日志器的单例实例。
	// 它只初始化一次，并在整个应用程序中使用。
	globalLogger *defaultLogger
	// once ensures that the globalLogger is initialized only once.
	// once 确保 globalLogger 只初始化一次。
	once sync.Once
)

// InitLogger initializes the global logger. It should be called once at application startup.
// If filePath is empty, logs will be written to os.Stdout.
//
// InitLogger 初始化全局日志器。它应该在应用程序启动时只调用一次。
// 如果 filePath 为空，日志将写入 os.Stdout。
func InitLogger(level enum.LogLevel, filePath string) error {
	var initErr error
	once.Do(func() {
		var output io.Writer
		var file *os.File
		if filePath != "" {
			var err error
			file, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				initErr = fmt.Errorf("failed to open log file %s: %w", filePath, err)
				return // Exit func for once.Do
			}
			output = io.MultiWriter(os.Stdout, file) // Log to both stdout and file
		} else {
			output = os.Stdout
		}

		stdLogger := log.New(output, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
		globalLogger = &defaultLogger{
			level:  level,
			logger: stdLogger,
			output: output,
			file:   file,
		}
		globalLogger.Info("Logger initialized. Level: %s, Output: %s", level.String(), filePath)
	})
	return initErr
}

// GetLogger returns the singleton instance of the global logger.
// It panics if InitLogger has not been called.
//
// GetLogger 返回全局日志器的单例实例。
// 如果 InitLogger 尚未调用，它会 panic。
func GetLogger() Logger {
	if globalLogger == nil {
		// This should not happen in a properly initialized application.
		// For robustness in early development, we can try to initialize a minimal logger.
		// In production, this indicates a startup issue.
		_ = InitLogger(enum.LogLevel_INFO, constants.DefaultLogFilePath)
		if globalLogger == nil { // Still nil? Something is critically wrong.
			panic("logger not initialized. Call InitLogger() first.")
		}
	}
	return globalLogger
}

// CloseLogger gracefully closes the log file if it was opened.
// It should be called before application shutdown.
//
// CloseLogger 优雅地关闭日志文件（如果已打开）。
// 它应该在应用程序关闭前调用。
func CloseLogger() error {
	if globalLogger != nil {
		globalLogger.mu.Lock()
		defer globalLogger.mu.Unlock()
		if globalLogger.file != nil {
			err := globalLogger.file.Close()
			globalLogger.file = nil // Clear file handle
			if err != nil {
				return fmt.Errorf("failed to close log file: %w", err)
			}
		}
	}
	return nil
}

// logf formats and logs a message at the specified level.
//
// logf 以指定的级别格式化并记录消息。
func (l *defaultLogger) logf(level enum.LogLevel, format string, args ...interface{}) {
	l.mu.RLock() // Use RLock for read access to level
	currentLevel := l.level
	l.mu.RUnlock()

	if level < currentLevel {
		return // Message level is below the active logging level, so ignore it
	}

	// Prepare the log prefix with timestamp, level, and file/line info
	prefix := fmt.Sprintf("%s [%s] ", time.Now().Format("2006-01-02 15:04:05.000"), level.String())

	// Use the underlying standard logger to print the message with file/line info
	// The standard logger's Lshortfile flag will automatically add file:line.
	// We prepend our custom prefix.
	l.logger.Printf(prefix+format, args...)
}

// Debug logs a message at the DEBUG level.
func (l *defaultLogger) Debug(format string, args ...interface{}) {
	l.logf(enum.LogLevel_DEBUG, format, args...)
}

// Info logs a message at the INFO level.
func (l *defaultLogger) Info(format string, args ...interface{}) {
	l.logf(enum.LogLevel_INFO, format, args...)
}

// Warn logs a message at the WARN level.
func (l *defaultLogger) Warn(format string, args ...interface{}) {
	l.logf(enum.LogLevel_WARN, format, args...)
}

// Error logs a message at the ERROR level.
func (l *defaultLogger) Error(format string, args ...interface{}) {
	l.logf(enum.LogLevel_ERROR, format, args...)
}

// Fatal logs a message at the FATAL level, then exits the application.
func (l *defaultLogger) Fatal(format string, args ...interface{}) {
	l.logf(enum.LogLevel_FATAL, format, args...)
	// For fatal errors, we should ensure all buffers are flushed and then exit.
	// os.Exit(1) is standard for abnormal termination.
	if l.file != nil {
		_ = l.file.Sync() // Attempt to flush file buffer
	}
	os.Exit(1)
}

// SetLevel sets the current logging level.
func (l *defaultLogger) SetLevel(level enum.LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
	l.Info("Log level changed to %s", level.String())
}

// GetLevel returns the current logging level.
func (l *defaultLogger) GetLevel() enum.LogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}
