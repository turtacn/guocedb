// Package log provides a unified logging interface for guocedb.
package log

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

// Logger defines the interface for logging.
type Logger interface {
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
	WithContext(ctx context.Context) Logger

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
}

// logger is a wrapper around logrus.
type logger struct {
	*logrus.Entry
}

var defaultLogger Logger

func init() {
	// Initialize the default logger
	defaultLogger = NewLogger("info", "json")
}

// NewLogger creates a new logger instance.
func NewLogger(level string, format string) Logger {
	log := logrus.New()

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)

	if format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	log.SetOutput(os.Stdout)

	return &logger{logrus.NewEntry(log)}
}

// GetLogger returns the default logger instance.
func GetLogger() Logger {
	return defaultLogger
}

// WithField adds a field to the log entry.
func (l *logger) WithField(key string, value interface{}) Logger {
	return &logger{l.Entry.WithField(key, value)}
}

// WithFields adds multiple fields to the log entry.
func (l *logger) WithFields(fields map[string]interface{}) Logger {
	return &logger{l.Entry.WithFields(logrus.Fields(fields))}
}

// WithError adds an error to the log entry.
func (l *logger) WithError(err error) Logger {
	return &logger{l.Entry.WithError(err)}
}

// WithContext adds a context to the log entry.
// Placeholder for distributed tracing.
func (l *logger) WithContext(ctx context.Context) Logger {
	// In a real implementation, we would extract trace and span IDs from the context.
	// For example:
	// if span := opentracing.SpanFromContext(ctx); span != nil {
	//     return &logger{l.Entry.WithField("traceID", ...)}
	// }
	return l
}
