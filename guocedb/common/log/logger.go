package log

import (
	"os"
	"github.com/sirupsen/logrus"
)

// Logger defines a standardized logging interface for the project.
type Logger interface {
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
}

// logrusWrapper is a wrapper around logrus.Entry that implements our Logger interface.
type logrusWrapper struct {
	*logrus.Entry
}

// NewLogger creates a new logger instance with default settings.
func NewLogger() Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel) // Default log level

	return &logrusWrapper{logrus.NewEntry(log)}
}

// WithField adds a structured field to the log entry.
func (l *logrusWrapper) WithField(key string, value interface{}) Logger {
	return &logrusWrapper{l.Entry.WithField(key, value)}
}

// WithFields adds multiple structured fields to the log entry.
func (l *logrusWrapper) WithFields(fields map[string]interface{}) Logger {
	return &logrusWrapper{l.Entry.WithFields(logrus.Fields(fields))}
}

// WithError adds an error to the log entry.
func (l *logrusWrapper) WithError(err error) Logger {
	return &logrusWrapper{l.Entry.WithError(err)}
}

var (
	// defaultLogger is the pre-configured global logger instance.
	defaultLogger = NewLogger().(*logrusWrapper)
)

// SetLevel sets the logging level for the default logger.
func SetLevel(level string) {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		defaultLogger.Warnf("Invalid log level '%s', using 'info' instead.", level)
		lvl = logrus.InfoLevel
	}
	defaultLogger.Logger.SetLevel(lvl)
}

// Delegating functions to the default logger for easy global access.
func WithField(key string, value interface{}) Logger { return defaultLogger.WithField(key, value) }
func WithFields(fields map[string]interface{}) Logger { return defaultLogger.WithFields(fields) }
func WithError(err error) Logger { return defaultLogger.WithError(err) }
func Debugf(format string, args ...interface{}) { defaultLogger.Debugf(format, args...) }
func Infof(format string, args ...interface{}) { defaultLogger.Infof(format, args...) }
func Warnf(format string, args ...interface{}) { defaultLogger.Warnf(format, args...) }
func Errorf(format string, args ...interface{}) { defaultLogger.Errorf(format, args...) }
func Fatalf(format string, args ...interface{}) { defaultLogger.Fatalf(format, args...) }
func Panicf(format string, args ...interface{}) { defaultLogger.Panicf(format, args...) }

// TODO: Add integration with distributed tracing (e.g., OpenTelemetry) to include trace IDs in logs.
