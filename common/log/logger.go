// Package log defines the unified logging interface and default implementation for the Guocedb project.
// This file is a core component of the system's observability, ensuring consistent, configurable,
// and analyzable structured logging across all modules.
package log

import (
	"log"  // Standard library log.
	"os"   // For standard error output and file operations.
	"sync" // For ensuring thread-safe logging.

	"github.com/turtacn/guocedb/common/constants"  // Import constants for default log settings.
	"github.com/turtacn/guocedb/common/types/enum" // Import enum for log levels and component types.

	"github.com/natefinch/lumberjack" // For log rotation.
	"go.uber.org/zap"                 // For structured logging.
	"go.uber.org/zap/zapcore"         // For Zap core functionalities.
)

// Logger is the interface for Guocedb's unified logger.
// All modules should use this interface for logging to ensure consistency.
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field) // Fatal logs and then calls os.Exit(1).
	With(fields ...zap.Field) Logger       // Creates a child logger with added fields.
	SetLevel(level enum.LogLevel)          // Sets the minimum logging level.
}

// guocedbLogger implements the Logger interface using Zap.
type guocedbLogger struct {
	zapLogger *zap.Logger
	atom      zap.AtomicLevel // AtomicLevel allows dynamic log level changes.
	mu        sync.RWMutex    // Mutex for protecting logger modifications.
}

// globalLogger is the singleton instance of the Guocedb logger.
var globalLogger *guocedbLogger
var once sync.Once

// InitLogger initializes the global logger instance.
// It should be called once at application startup.
// If logFilePath is empty, logs will be written to os.Stdout.
// If level is not provided or invalid, constants.DefaultLogLevel will be used.
func InitLogger(logFilePath string, level string) {
	once.Do(func() {
		// Determine log level
		var logLevel zapcore.Level
		parsedLevel, err := enum.ParseLogLevel(level)
		if err != nil {
			log.Printf("Failed to parse log level '%s', using default '%s'", level, constants.DefaultLogLevel)
			parsedLevel, _ = enum.ParseLogLevel(constants.DefaultLogLevel) // Use default if parsing fails
		}
		logLevel = toZapLevel(parsedLevel)

		atom := zap.NewAtomicLevelAt(logLevel)

		// Configure console output
		consoleEncoderCfg := zap.NewDevelopmentEncoderConfig()
		consoleEncoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder // Colored output for console
		consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderCfg)
		consoleWriter := zapcore.AddSync(os.Stdout)
		consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, atom)

		var cores []zapcore.Core
		cores = append(cores, consoleCore) // Always include console output

		// Configure file output if path is provided
		if logFilePath != "" {
			fileEncoderCfg := zap.NewProductionEncoderConfig()
			fileEncoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder // ISO8601 for file logs
			fileEncoder := zapcore.NewJSONEncoder(fileEncoderCfg)  // JSON format for easier parsing by tools
			fileWriter := zapcore.AddSync(&lumberjack.Logger{
				Filename:   logFilePath,
				MaxSize:    constants.LogFileMaxSizeMB, // megabytes
				MaxBackups: constants.LogFileMaxBackups,
				MaxAge:     constants.LogFileMaxAgeDays, // days
				Compress:   true,                        // zip old logs
			})
			fileCore := zapcore.NewCore(fileEncoder, fileWriter, atom)
			cores = append(cores, fileCore)
		}

		// Combine cores if multiple outputs are configured
		core := zapcore.NewTee(cores...)

		zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
		globalLogger = &guocedbLogger{
			zapLogger: zapLogger,
			atom:      atom,
		}
		zap.ReplaceGlobals(zapLogger) // Replace global Zap logger for convenience
	})
}

// GetLogger returns the global Logger instance.
// It should be called after InitLogger. If not, it will return a no-op logger to prevent panics.
func GetLogger() Logger {
	if globalLogger == nil {
		// Fallback to a discard logger if not initialized to prevent nil panics
		// In a real application, you might want to panic here or have a more robust default.
		return &noOpLogger{}
	}
	return globalLogger
}

// toZapLevel converts Guocedb's LogLevel enum to Zap's zapcore.Level.
func toZapLevel(level enum.LogLevel) zapcore.Level {
	switch level {
	case enum.LogLevelDebug:
		return zapcore.DebugLevel
	case enum.LogLevelInfo:
		return zapcore.InfoLevel
	case enum.LogLevelWarn:
		return zapcore.WarnLevel
	case enum.LogLevelError:
		return zapcore.ErrorLevel
	case enum.LogLevelFatal:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel // Default to Info if unknown
	}
}

// toGuocedbLevel converts Zap's zapcore.Level to Guocedb's LogLevel enum.
func toGuocedbLevel(level zapcore.Level) enum.LogLevel {
	switch level {
	case zapcore.DebugLevel:
		return enum.LogLevelDebug
	case zapcore.InfoLevel:
		return enum.LogLevelInfo
	case zapcore.WarnLevel:
		return enum.LogLevelWarn
	case zapcore.ErrorLevel:
		return enum.LogLevelError
	case zapcore.FatalLevel:
		return enum.LogLevelFatal
	default:
		return enum.LogLevelInfo // Default to Info if unknown
	}
}

// Debug logs a message at Debug level.
func (l *guocedbLogger) Debug(msg string, fields ...zap.Field) {
	l.zapLogger.Debug(msg, fields...)
}

// Info logs a message at Info level.
func (l *guocedbLogger) Info(msg string, fields ...zap.Field) {
	l.zapLogger.Info(msg, fields...)
}

// Warn logs a message at Warn level.
func (l *guocedbLogger) Warn(msg string, fields ...zap.Field) {
	l.zapLogger.Warn(msg, fields...)
}

// Error logs a message at Error level.
func (l *guocedbLogger) Error(msg string, fields ...zap.Field) {
	l.zapLogger.Error(msg, fields...)
}

// Fatal logs a message at Fatal level and then calls os.Exit(1).
func (l *guocedbLogger) Fatal(msg string, fields ...zap.Field) {
	l.zapLogger.Fatal(msg, fields...)
}

// With creates a child logger with added fields.
func (l *guocedbLogger) With(fields ...zap.Field) Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return &guocedbLogger{
		zapLogger: l.zapLogger.With(fields...),
		atom:      l.atom, // Share the same atomic level for consistent level changes
	}
}

// SetLevel dynamically sets the minimum logging level for the logger.
func (l *guocedbLogger) SetLevel(level enum.LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.atom.SetLevel(toZapLevel(level))
}

// noOpLogger is a no-operation logger for scenarios where the logger isn't initialized.
type noOpLogger struct{}

func (*noOpLogger) Debug(msg string, fields ...zap.Field) {}
func (*noOpLogger) Info(msg string, fields ...zap.Field)  {}
func (*noOpLogger) Warn(msg string, fields ...zap.Field)  {}
func (*noOpLogger) Error(msg string, fields ...zap.Field) {}
func (*noOpLogger) Fatal(msg string, fields ...zap.Field) { os.Exit(1) } // Fatal still exits
func (l *noOpLogger) With(fields ...zap.Field) Logger     { return l }
func (*noOpLogger) SetLevel(level enum.LogLevel)          {}

//Personal.AI order the ending
