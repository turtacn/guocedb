// Package log 实现GuoceDB的统一日志系统
// Package log implements unified logging system for GuoceDB
package log

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/turtacn/guocedb/common/constants"
	"github.com/turtacn/guocedb/common/errors"
)

// LogLevel 日志级别枚举
// LogLevel enumeration for log levels
type LogLevel int

const (
	// LevelTrace 跟踪级别
	// LevelTrace trace level
	LevelTrace LogLevel = iota
	// LevelDebug 调试级别
	// LevelDebug debug level
	LevelDebug
	// LevelInfo 信息级别
	// LevelInfo info level
	LevelInfo
	// LevelWarn 警告级别
	// LevelWarn warning level
	LevelWarn
	// LevelError 错误级别
	// LevelError error level
	LevelError
	// LevelFatal 致命错误级别
	// LevelFatal fatal level
	LevelFatal
)

// String 返回日志级别的字符串表示
// String returns string representation of log level
func (l LogLevel) String() string {
	switch l {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return fmt.Sprintf("UNKNOWN_LEVEL(%d)", int(l))
	}
}

// ParseLogLevel 从字符串解析日志级别
// ParseLogLevel parses log level from string
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "TRACE":
		return LevelTrace
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "FATAL":
		return LevelFatal
	default:
		return LevelInfo
	}
}

// LogFormat 日志格式枚举
// LogFormat enumeration for log formats
type LogFormat int

const (
	// FormatText 文本格式
	// FormatText text format
	FormatText LogFormat = iota
	// FormatJSON JSON格式
	// FormatJSON JSON format
	FormatJSON
)

// String 返回日志格式的字符串表示
// String returns string representation of log format
func (f LogFormat) String() string {
	switch f {
	case FormatText:
		return "text"
	case FormatJSON:
		return "json"
	default:
		return "text"
	}
}

// ParseLogFormat 从字符串解析日志格式
// ParseLogFormat parses log format from string
func ParseLogFormat(format string) LogFormat {
	switch strings.ToLower(format) {
	case "json":
		return FormatJSON
	case "text":
		return FormatText
	default:
		return FormatText
	}
}

// Fields 日志字段类型
// Fields type for log fields
type Fields map[string]interface{}

// LogEntry 日志条目结构
// LogEntry structure for log entries
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
	Fields    Fields    `json:"fields,omitempty"`
	Caller    string    `json:"caller,omitempty"`
	Stack     string    `json:"stack,omitempty"`
	TraceID   string    `json:"trace_id,omitempty"`
	SpanID    string    `json:"span_id,omitempty"`
}

// String 返回日志条目的字符串表示
// String returns string representation of log entry
func (e *LogEntry) String() string {
	caller := ""
	if e.Caller != "" {
		caller = fmt.Sprintf(" [%s]", e.Caller)
	}

	traceInfo := ""
	if e.TraceID != "" {
		traceInfo = fmt.Sprintf(" [trace:%s", e.TraceID)
		if e.SpanID != "" {
			traceInfo += fmt.Sprintf(",span:%s", e.SpanID)
		}
		traceInfo += "]"
	}

	fields := ""
	if len(e.Fields) > 0 {
		parts := make([]string, 0, len(e.Fields))
		for k, v := range e.Fields {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
		fields = fmt.Sprintf(" {%s}", strings.Join(parts, " "))
	}

	return fmt.Sprintf("%s [%s]%s%s %s%s",
		e.Timestamp.Format("2006-01-02 15:04:05.000"),
		e.Level.String(),
		caller,
		traceInfo,
		e.Message,
		fields)
}

// ToJSON 将日志条目转换为JSON格式
// ToJSON converts log entry to JSON format
func (e *LogEntry) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Logger 日志接口
// Logger interface for logging
type Logger interface {
	// SetLevel 设置日志级别
	// SetLevel sets log level
	SetLevel(level LogLevel)

	// GetLevel 获取日志级别
	// GetLevel gets log level
	GetLevel() LogLevel

	// SetFormat 设置日志格式
	// SetFormat sets log format
	SetFormat(format LogFormat)

	// GetFormat 获取日志格式
	// GetFormat gets log format
	GetFormat() LogFormat

	// WithField 添加单个字段
	// WithField adds single field
	WithField(key string, value interface{}) Logger

	// WithFields 添加多个字段
	// WithFields adds multiple fields
	WithFields(fields Fields) Logger

	// WithContext 从上下文中提取字段
	// WithContext extracts fields from context
	WithContext(ctx context.Context) Logger

	// WithCaller 添加调用者信息
	// WithCaller adds caller information
	WithCaller(skip int) Logger

	// Trace 输出跟踪级别日志
	// Trace outputs trace level log
	Trace(args ...interface{})

	// Tracef 输出格式化跟踪级别日志
	// Tracef outputs formatted trace level log
	Tracef(format string, args ...interface{})

	// Debug 输出调试级别日志
	// Debug outputs debug level log
	Debug(args ...interface{})

	// Debugf 输出格式化调试级别日志
	// Debugf outputs formatted debug level log
	Debugf(format string, args ...interface{})

	// Info 输出信息级别日志
	// Info outputs info level log
	Info(args ...interface{})

	// Infof 输出格式化信息级别日志
	// Infof outputs formatted info level log
	Infof(format string, args ...interface{})

	// Warn 输出警告级别日志
	// Warn outputs warning level log
	Warn(args ...interface{})

	// Warnf 输出格式化警告级别日志
	// Warnf outputs formatted warning level log
	Warnf(format string, args ...interface{})

	// Error 输出错误级别日志
	// Error outputs error level log
	Error(args ...interface{})

	// Errorf 输出格式化错误级别日志
	// Errorf outputs formatted error level log
	Errorf(format string, args ...interface{})

	// Fatal 输出致命错误级别日志并退出程序
	// Fatal outputs fatal level log and exits program
	Fatal(args ...interface{})

	// Fatalf 输出格式化致命错误级别日志并退出程序
	// Fatalf outputs formatted fatal level log and exits program
	Fatalf(format string, args ...interface{})

	// Close 关闭日志器
	// Close closes the logger
	Close() error
}

// RotateConfig 日志轮转配置
// RotateConfig configuration for log rotation
type RotateConfig struct {
	MaxSize    int  `json:"max_size"`    // 最大文件大小（MB）Maximum file size in MB
	MaxBackups int  `json:"max_backups"` // 最大备份文件数量 Maximum number of backup files
	MaxAge     int  `json:"max_age"`     // 最大保留天数 Maximum retention days
	Compress   bool `json:"compress"`    // 是否压缩 Whether to compress
}

// DefaultRotateConfig 默认轮转配置
// DefaultRotateConfig default rotation configuration
var DefaultRotateConfig = &RotateConfig{
	MaxSize:    constants.DefaultLogMaxSize,
	MaxBackups: constants.DefaultLogMaxBackups,
	MaxAge:     constants.DefaultLogMaxAge,
	Compress:   constants.DefaultLogCompress,
}

// Config 日志配置
// Config configuration for logger
type Config struct {
	Level      LogLevel      `json:"level"`       // 日志级别 Log level
	Format     LogFormat     `json:"format"`      // 日志格式 Log format
	Output     string        `json:"output"`      // 输出目标 Output target
	File       string        `json:"file"`        // 日志文件路径 Log file path
	Async      bool          `json:"async"`       // 是否异步 Whether async
	BufferSize int           `json:"buffer_size"` // 缓冲区大小 Buffer size
	Rotate     *RotateConfig `json:"rotate"`      // 轮转配置 Rotation configuration
}

// DefaultConfig 默认日志配置
// DefaultConfig default logger configuration
var DefaultConfig = &Config{
	Level:      ParseLogLevel(constants.DefaultLogLevel),
	Format:     ParseLogFormat(constants.DefaultLogFormat),
	Output:     "file",
	File:       constants.DefaultLogFile,
	Async:      true,
	BufferSize: 1000,
	Rotate:     DefaultRotateConfig,
}

// standardLogger 标准日志实现
// standardLogger standard logger implementation
type standardLogger struct {
	mu         sync.RWMutex
	level      LogLevel
	format     LogFormat
	output     io.WriteCloser
	fields     Fields
	caller     bool
	callerSkip int
	async      bool
	buffer     chan *LogEntry
	done       chan struct{}
	wg         sync.WaitGroup
}

// NewLogger 创建新的日志器
// NewLogger creates new logger
func NewLogger(config *Config) (Logger, error) {
	if config == nil {
		config = DefaultConfig
	}

	var output io.WriteCloser

	switch config.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	case "file":
		if config.File == "" {
			config.File = constants.DefaultLogFile
		}

		// 确保日志目录存在
		// Ensure log directory exists
		dir := filepath.Dir(config.File)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, errors.WrapError(errors.ErrCodeSystemFailure, "Failed to create log directory", err)
		}

		file, err := os.OpenFile(config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, errors.WrapError(errors.ErrCodeSystemFailure, "Failed to open log file", err)
		}
		output = file
	default:
		output = os.Stdout
	}

	logger := &standardLogger{
		level:  config.Level,
		format: config.Format,
		output: output,
		fields: make(Fields),
		async:  config.Async,
		done:   make(chan struct{}),
	}

	if config.Async {
		bufferSize := config.BufferSize
		if bufferSize <= 0 {
			bufferSize = 1000
		}
		logger.buffer = make(chan *LogEntry, bufferSize)
		logger.startAsyncWriter()
	}

	return logger, nil
}

// startAsyncWriter 启动异步写入器
// startAsyncWriter starts async writer
func (l *standardLogger) startAsyncWriter() {
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		for {
			select {
			case entry := <-l.buffer:
				l.writeEntry(entry)
			case <-l.done:
				// 处理剩余的日志条目
				// Process remaining log entries
				for {
					select {
					case entry := <-l.buffer:
						l.writeEntry(entry)
					default:
						return
					}
				}
			}
		}
	}()
}

// SetLevel 设置日志级别
// SetLevel sets log level
func (l *standardLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel 获取日志级别
// GetLevel gets log level
func (l *standardLogger) GetLevel() LogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// SetFormat 设置日志格式
// SetFormat sets log format
func (l *standardLogger) SetFormat(format LogFormat) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.format = format
}

// GetFormat 获取日志格式
// GetFormat gets log format
func (l *standardLogger) GetFormat() LogFormat {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.format
}

// clone 克隆日志器
// clone clones the logger
func (l *standardLogger) clone() *standardLogger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	fields := make(Fields)
	for k, v := range l.fields {
		fields[k] = v
	}

	return &standardLogger{
		level:      l.level,
		format:     l.format,
		output:     l.output,
		fields:     fields,
		caller:     l.caller,
		callerSkip: l.callerSkip,
		async:      l.async,
		buffer:     l.buffer,
		done:       l.done,
	}
}

// WithField 添加单个字段
// WithField adds single field
func (l *standardLogger) WithField(key string, value interface{}) Logger {
	logger := l.clone()
	logger.fields[key] = value
	return logger
}

// WithFields 添加多个字段
// WithFields adds multiple fields
func (l *standardLogger) WithFields(fields Fields) Logger {
	logger := l.clone()
	for k, v := range fields {
		logger.fields[k] = v
	}
	return logger
}

// WithContext 从上下文中提取字段
// WithContext extracts fields from context
func (l *standardLogger) WithContext(ctx context.Context) Logger {
	logger := l.clone()

	// 提取跟踪ID
	// Extract trace ID
	if traceID := getTraceIDFromContext(ctx); traceID != "" {
		logger.fields["trace_id"] = traceID
	}

	// 提取跨度ID
	// Extract span ID
	if spanID := getSpanIDFromContext(ctx); spanID != "" {
		logger.fields["span_id"] = spanID
	}

	// 提取用户ID
	// Extract user ID
	if userID := getUserIDFromContext(ctx); userID != "" {
		logger.fields["user_id"] = userID
	}

	// 提取请求ID
	// Extract request ID
	if requestID := getRequestIDFromContext(ctx); requestID != "" {
		logger.fields["request_id"] = requestID
	}

	return logger
}

// WithCaller 添加调用者信息
// WithCaller adds caller information
func (l *standardLogger) WithCaller(skip int) Logger {
	logger := l.clone()
	logger.caller = true
	logger.callerSkip = skip + 1
	return logger
}

// log 输出日志
// log outputs log
func (l *standardLogger) log(level LogLevel, args ...interface{}) {
	if level < l.GetLevel() {
		return
	}

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   fmt.Sprint(args...),
		Fields:    make(Fields),
	}

	// 复制字段
	// Copy fields
	l.mu.RLock()
	for k, v := range l.fields {
		entry.Fields[k] = v
	}

	// 添加调用者信息
	// Add caller information
	if l.caller {
		if caller := getCaller(l.callerSkip + 2); caller != "" {
			entry.Caller = caller
		}
	}
	l.mu.RUnlock()

	// 提取特殊字段
	// Extract special fields
	if traceID, ok := entry.Fields["trace_id"].(string); ok {
		entry.TraceID = traceID
		delete(entry.Fields, "trace_id")
	}

	if spanID, ok := entry.Fields["span_id"].(string); ok {
		entry.SpanID = spanID
		delete(entry.Fields, "span_id")
	}

	// 对于错误级别，添加堆栈信息
	// Add stack trace for error level
	if level >= LevelError {
		entry.Stack = getStackTrace(2)
	}

	if len(entry.Fields) == 0 {
		entry.Fields = nil
	}

	l.writeLog(entry)
}

// logf 输出格式化日志
// logf outputs formatted log
func (l *standardLogger) logf(level LogLevel, format string, args ...interface{}) {
	if level < l.GetLevel() {
		return
	}

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   fmt.Sprintf(format, args...),
		Fields:    make(Fields),
	}

	// 复制字段
	// Copy fields
	l.mu.RLock()
	for k, v := range l.fields {
		entry.Fields[k] = v
	}

	// 添加调用者信息
	// Add caller information
	if l.caller {
		if caller := getCaller(l.callerSkip + 2); caller != "" {
			entry.Caller = caller
		}
	}
	l.mu.RUnlock()

	// 提取特殊字段
	// Extract special fields
	if traceID, ok := entry.Fields["trace_id"].(string); ok {
		entry.TraceID = traceID
		delete(entry.Fields, "trace_id")
	}

	if spanID, ok := entry.Fields["span_id"].(string); ok {
		entry.SpanID = spanID
		delete(entry.Fields, "span_id")
	}

	// 对于错误级别，添加堆栈信息
	// Add stack trace for error level
	if level >= LevelError {
		entry.Stack = getStackTrace(2)
	}

	if len(entry.Fields) == 0 {
		entry.Fields = nil
	}

	l.writeLog(entry)
}

// writeLog 写入日志
// writeLog writes log
func (l *standardLogger) writeLog(entry *LogEntry) {
	if l.async && l.buffer != nil {
		select {
		case l.buffer <- entry:
		default:
			// 缓冲区满，直接写入
			// Buffer full, write directly
			l.writeEntry(entry)
		}
	} else {
		l.writeEntry(entry)
	}
}

// writeEntry 写入日志条目
// writeEntry writes log entry
func (l *standardLogger) writeEntry(entry *LogEntry) {
	var data []byte
	var err error

	format := l.GetFormat()
	if format == FormatJSON {
		data, err = entry.ToJSON()
		if err != nil {
			data = []byte(fmt.Sprintf(`{"error":"failed to marshal log entry: %v"}`, err))
		}
		data = append(data, '\n')
	} else {
		data = []byte(entry.String() + "\n")
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.output != nil {
		l.output.Write(data)
	}
}

// Trace 输出跟踪级别日志
// Trace outputs trace level log
func (l *standardLogger) Trace(args ...interface{}) {
	l.log(LevelTrace, args...)
}

// Tracef 输出格式化跟踪级别日志
// Tracef outputs formatted trace level log
func (l *standardLogger) Tracef(format string, args ...interface{}) {
	l.logf(LevelTrace, format, args...)
}

// Debug 输出调试级别日志
// Debug outputs debug level log
func (l *standardLogger) Debug(args ...interface{}) {
	l.log(LevelDebug, args...)
}

// Debugf 输出格式化调试级别日志
// Debugf outputs formatted debug level log
func (l *standardLogger) Debugf(format string, args ...interface{}) {
	l.logf(LevelDebug, format, args...)
}

// Info 输出信息级别日志
// Info outputs info level log
func (l *standardLogger) Info(args ...interface{}) {
	l.log(LevelInfo, args...)
}

// Infof 输出格式化信息级别日志
// Infof outputs formatted info level log
func (l *standardLogger) Infof(format string, args ...interface{}) {
	l.logf(LevelInfo, format, args...)
}

// Warn 输出警告级别日志
// Warn outputs warning level log
func (l *standardLogger) Warn(args ...interface{}) {
	l.log(LevelWarn, args...)
}

// Warnf 输出格式化警告级别日志
// Warnf outputs formatted warning level log
func (l *standardLogger) Warnf(format string, args ...interface{}) {
	l.logf(LevelWarn, format, args...)
}

// Error 输出错误级别日志
// Error outputs error level log
func (l *standardLogger) Error(args ...interface{}) {
	l.log(LevelError, args...)
}

// Errorf 输出格式化错误级别日志
// Errorf outputs formatted error level log
func (l *standardLogger) Errorf(format string, args ...interface{}) {
	l.logf(LevelError, format, args...)
}

// Fatal 输出致命错误级别日志并退出程序
// Fatal outputs fatal level log and exits program
func (l *standardLogger) Fatal(args ...interface{}) {
	l.log(LevelFatal, args...)
	l.Close()
	os.Exit(1)
}

// Fatalf 输出格式化致命错误级别日志并退出程序
// Fatalf outputs formatted fatal level log and exits program
func (l *standardLogger) Fatalf(format string, args ...interface{}) {
	l.logf(LevelFatal, format, args...)
	l.Close()
	os.Exit(1)
}

// Close 关闭日志器
// Close closes the logger
func (l *standardLogger) Close() error {
	if l.async && l.done != nil {
		close(l.done)
		l.wg.Wait()
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.output != nil && l.output != os.Stdout && l.output != os.Stderr {
		return l.output.Close()
	}

	return nil
}

// 工具函数 Utility functions

// getCaller 获取调用者信息
// getCaller gets caller information
func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}

	// 只保留文件名
	// Keep only filename
	file = filepath.Base(file)
	return fmt.Sprintf("%s:%d", file, line)
}

// getStackTrace 获取堆栈跟踪
// getStackTrace gets stack trace
func getStackTrace(skip int) string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip, pcs[:])

	var traces []string
	for i := 0; i < n; i++ {
		pc := pcs[i]
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		file, line := fn.FileLine(pc)
		traces = append(traces, fmt.Sprintf("%s:%d %s", filepath.Base(file), line, fn.Name()))
	}

	return strings.Join(traces, "\n")
}

// 上下文提取函数 Context extraction functions

// getTraceIDFromContext 从上下文中获取跟踪ID
// getTraceIDFromContext gets trace ID from context
func getTraceIDFromContext(ctx context.Context) string {
	if value := ctx.Value("trace_id"); value != nil {
		if traceID, ok := value.(string); ok {
			return traceID
		}
	}
	return ""
}

// getSpanIDFromContext 从上下文中获取跨度ID
// getSpanIDFromContext gets span ID from context
func getSpanIDFromContext(ctx context.Context) string {
	if value := ctx.Value("span_id"); value != nil {
		if spanID, ok := value.(string); ok {
			return spanID
		}
	}
	return ""
}

// getUserIDFromContext 从上下文中获取用户ID
// getUserIDFromContext gets user ID from context
func getUserIDFromContext(ctx context.Context) string {
	if value := ctx.Value("user_id"); value != nil {
		if userID, ok := value.(string); ok {
			return userID
		}
	}
	return ""
}

// getRequestIDFromContext 从上下文中获取请求ID
// getRequestIDFromContext gets request ID from context
func getRequestIDFromContext(ctx context.Context) string {
	if value := ctx.Value("request_id"); value != nil {
		if requestID, ok := value.(string); ok {
			return requestID
		}
	}
	return ""
}

// 全局日志器 Global logger
var (
	globalLogger Logger
	globalMu     sync.RWMutex
)

// InitGlobalLogger 初始化全局日志器
// InitGlobalLogger initializes global logger
func InitGlobalLogger(config *Config) error {
	logger, err := NewLogger(config)
	if err != nil {
		return err
	}

	globalMu.Lock()
	defer globalMu.Unlock()

	if globalLogger != nil {
		globalLogger.Close()
	}

	globalLogger = logger
	return nil
}

// GetGlobalLogger 获取全局日志器
// GetGlobalLogger gets global logger
func GetGlobalLogger() Logger {
	globalMu.RLock()
	defer globalMu.RUnlock()

	if globalLogger == nil {
		// 使用默认配置创建日志器
		// Create logger with default config
		logger, _ := NewLogger(DefaultConfig)
		globalLogger = logger
	}

	return globalLogger
}

// 全局日志函数 Global logging functions

// Trace 输出跟踪级别日志
// Trace outputs trace level log
func Trace(args ...interface{}) {
	GetGlobalLogger().WithCaller(1).Trace(args...)
}

// Tracef 输出格式化跟踪级别日志
// Tracef outputs formatted trace level log
func Tracef(format string, args ...interface{}) {
	GetGlobalLogger().WithCaller(1).Tracef(format, args...)
}

// Debug 输出调试级别日志
// Debug outputs debug level log
func Debug(args ...interface{}) {
	GetGlobalLogger().WithCaller(1).Debug(args...)
}

// Debugf 输出格式化调试级别日志
// Debugf outputs formatted debug level log
func Debugf(format string, args ...interface{}) {
	GetGlobalLogger().WithCaller(1).Debugf(format, args...)
}

// Info 输出信息级别日志
// Info outputs info level log
func Info(args ...interface{}) {
	GetGlobalLogger().WithCaller(1).Info(args...)
}

// Infof 输出格式化信息级别日志
// Infof outputs formatted info level log
func Infof(format string, args ...interface{}) {
	GetGlobalLogger().WithCaller(1).Infof(format, args...)
}

// Warn 输出警告级别日志
// Warn outputs warning level log
func Warn(args ...interface{}) {
	GetGlobalLogger().WithCaller(1).Warn(args...)
}

// Warnf 输出格式化警告级别日志
// Warnf outputs formatted warning level log
func Warnf(format string, args ...interface{}) {
	GetGlobalLogger().WithCaller(1).Warnf(format, args...)
}

// Error 输出错误级别日志
// Error outputs error level log
func Error(args ...interface{}) {
	GetGlobalLogger().WithCaller(1).Error(args...)
}

// Errorf 输出格式化错误级别日志
// Errorf outputs formatted error level log
func Errorf(format string, args ...interface{}) {
	GetGlobalLogger().WithCaller(1).Errorf(format, args...)
}

// Fatal 输出致命错误级别日志并退出程序
// Fatal outputs fatal level log and exits program
func Fatal(args ...interface{}) {
	GetGlobalLogger().WithCaller(1).Fatal(args...)
}

// Fatalf 输出格式化致命错误级别日志并退出程序
// Fatalf outputs formatted fatal level log and exits program
func Fatalf(format string, args ...interface{}) {
	GetGlobalLogger().WithCaller(1).Fatalf(format, args...)
}

// WithField 添加单个字段
// WithField adds single field
func WithField(key string, value interface{}) Logger {
	return GetGlobalLogger().WithField(key, value)
}

// WithFields 添加多个字段
// WithFields adds multiple fields
func WithFields(fields Fields) Logger {
	return GetGlobalLogger().WithFields(fields)
}

// WithContext 从上下文中提取字段
// WithContext extracts fields from context
func WithContext(ctx context.Context) Logger {
	return GetGlobalLogger().WithContext(ctx)
}

// WithCaller 添加调用者信息
// WithCaller adds caller information
func WithCaller(skip int) Logger {
	return GetGlobalLogger().WithCaller(skip + 1)
}

// SetLevel 设置全局日志级别
// SetLevel sets global log level
func SetLevel(level LogLevel) {
	GetGlobalLogger().SetLevel(level)
}

// GetLevel 获取全局日志级别
// GetLevel gets global log level
func GetLevel() LogLevel {
	return GetGlobalLogger().GetLevel()
}

// SetFormat 设置全局日志格式
// SetFormat sets global log format
func SetFormat(format LogFormat) {
	GetGlobalLogger().SetFormat(format)
}

// GetFormat 获取全局日志格式
// GetFormat gets global log format
func GetFormat() LogFormat {
	return GetGlobalLogger().GetFormat()
}

// 轮转日志实现 Rotating log implementation

// rotatingWriter 轮转写入器
// rotatingWriter rotating writer
type rotatingWriter struct {
	mu        sync.Mutex
	filename  string
	file      *os.File
	config    *RotateConfig
	size      int64
	backupNum int
}

// NewRotatingWriter 创建轮转写入器
// NewRotatingWriter creates rotating writer
func NewRotatingWriter(filename string, config *RotateConfig) (io.WriteCloser, error) {
	if config == nil {
		config = DefaultRotateConfig
	}

	// 确保目录存在
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, errors.WrapError(errors.ErrCodeSystemFailure, "Failed to create log directory", err)
	}

	writer := &rotatingWriter{
		filename: filename,
		config:   config,
	}

	if err := writer.openFile(); err != nil {
		return nil, err
	}

	return writer, nil
}

// openFile 打开文件
// openFile opens file
func (w *rotatingWriter) openFile() error {
	info, err := os.Stat(w.filename)
	if err == nil {
		w.size = info.Size()
	}

	file, err := os.OpenFile(w.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return errors.WrapError(errors.ErrCodeSystemFailure, "Failed to open log file", err)
	}

	w.file = file
	return nil
}

// Write 写入数据
// Write writes data
func (w *rotatingWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	writeLen := int64(len(p))

	// 检查是否需要轮转
	// Check if rotation needed
	if w.size+writeLen > int64(w.config.MaxSize)*1024*1024 {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = w.file.Write(p)
	w.size += int64(n)

	return n, err
}

// rotate 执行轮转
// rotate performs rotation
func (w *rotatingWriter) rotate() error {
	if w.file != nil {
		w.file.Close()
	}

	// 重命名现有文件
	// Rename existing files
	for i := w.config.MaxBackups; i > 0; i-- {
		oldName := fmt.Sprintf("%s.%d", w.filename, i)
		newName := fmt.Sprintf("%s.%d", w.filename, i+1)

		if i == w.config.MaxBackups {
			// 删除最老的文件
			// Remove oldest file
			os.Remove(oldName)
		} else {
			// 重命名文件
			// Rename file
			if _, err := os.Stat(oldName); err == nil {
				os.Rename(oldName, newName)
			}
		}
	}

	// 将当前文件重命名为.1
	// Rename current file to .1
	backupName := fmt.Sprintf("%s.1", w.filename)
	if err := os.Rename(w.filename, backupName); err != nil {
		return errors.WrapError(errors.ErrCodeSystemFailure, "Failed to rotate log file", err)
	}

	// 压缩备份文件
	// Compress backup file
	if w.config.Compress {
		go w.compressFile(backupName)
	}

	// 清理过期文件
	// Clean up expired files
	go w.cleanupOldFiles()

	// 重新打开文件
	// Reopen file
	w.size = 0
	return w.openFile()
}

// compressFile 压缩文件
// compressFile compresses file
func (w *rotatingWriter) compressFile(filename string) {
	// 这里可以实现文件压缩逻辑
	// File compression logic can be implemented here
	// 为了简化，这里只是一个占位符
	// This is just a placeholder for simplification
}

// cleanupOldFiles 清理过期文件
// cleanupOldFiles cleans up expired files
func (w *rotatingWriter) cleanupOldFiles() {
	if w.config.MaxAge <= 0 {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -w.config.MaxAge)

	dir := filepath.Dir(w.filename)
	base := filepath.Base(w.filename)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, base+".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(dir, name))
		}
	}
}

// Close 关闭写入器
// Close closes writer
func (w *rotatingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Close()
	}

	return nil
}

// 多输出日志器 Multi-output logger

// multiLogger 多输出日志器
// multiLogger multi-output logger
type multiLogger struct {
	loggers []Logger
}

// NewMultiLogger 创建多输出日志器
// NewMultiLogger creates multi-output logger
func NewMultiLogger(loggers ...Logger) Logger {
	return &multiLogger{
		loggers: loggers,
	}
}

// SetLevel 设置日志级别
// SetLevel sets log level
func (m *multiLogger) SetLevel(level LogLevel) {
	for _, logger := range m.loggers {
		logger.SetLevel(level)
	}
}

// GetLevel 获取日志级别
// GetLevel gets log level
func (m *multiLogger) GetLevel() LogLevel {
	if len(m.loggers) > 0 {
		return m.loggers[0].GetLevel()
	}
	return LevelInfo
}

// SetFormat 设置日志格式
// SetFormat sets log format
func (m *multiLogger) SetFormat(format LogFormat) {
	for _, logger := range m.loggers {
		logger.SetFormat(format)
	}
}

// GetFormat 获取日志格式
// GetFormat gets log format
func (m *multiLogger) GetFormat() LogFormat {
	if len(m.loggers) > 0 {
		return m.loggers[0].GetFormat()
	}
	return FormatText
}

// WithField 添加单个字段
// WithField adds single field
func (m *multiLogger) WithField(key string, value interface{}) Logger {
	loggers := make([]Logger, len(m.loggers))
	for i, logger := range m.loggers {
		loggers[i] = logger.WithField(key, value)
	}
	return &multiLogger{loggers: loggers}
}

// WithFields 添加多个字段
// WithFields adds multiple fields
func (m *multiLogger) WithFields(fields Fields) Logger {
	loggers := make([]Logger, len(m.loggers))
	for i, logger := range m.loggers {
		loggers[i] = logger.WithFields(fields)
	}
	return &multiLogger{loggers: loggers}
}

// WithContext 从上下文中提取字段
// WithContext extracts fields from context
func (m *multiLogger) WithContext(ctx context.Context) Logger {
	loggers := make([]Logger, len(m.loggers))
	for i, logger := range m.loggers {
		loggers[i] = logger.WithContext(ctx)
	}
	return &multiLogger{loggers: loggers}
}

// WithCaller 添加调用者信息
// WithCaller adds caller information
func (m *multiLogger) WithCaller(skip int) Logger {
	loggers := make([]Logger, len(m.loggers))
	for i, logger := range m.loggers {
		loggers[i] = logger.WithCaller(skip)
	}
	return &multiLogger{loggers: loggers}
}

// Trace 输出跟踪级别日志
// Trace outputs trace level log
func (m *multiLogger) Trace(args ...interface{}) {
	for _, logger := range m.loggers {
		logger.Trace(args...)
	}
}

// Tracef 输出格式化跟踪级别日志
// Tracef outputs formatted trace level log
func (m *multiLogger) Tracef(format string, args ...interface{}) {
	for _, logger := range m.loggers {
		logger.Tracef(format, args...)
	}
}

// Debug 输出调试级别日志
// Debug outputs debug level log
func (m *multiLogger) Debug(args ...interface{}) {
	for _, logger := range m.loggers {
		logger.Debug(args...)
	}
}

// Debugf 输出格式化调试级别日志
// Debugf outputs formatted debug level log
func (m *multiLogger) Debugf(format string, args ...interface{}) {
	for _, logger := range m.loggers {
		logger.Debugf(format, args...)
	}
}

// Info 输出信息级别日志
// Info outputs info level log
func (m *multiLogger) Info(args ...interface{}) {
	for _, logger := range m.loggers {
		logger.Info(args...)
	}
}

// Infof 输出格式化信息级别日志
// Infof outputs formatted info level log
func (m *multiLogger) Infof(format string, args ...interface{}) {
	for _, logger := range m.loggers {
		logger.Infof(format, args...)
	}
}

// Warn 输出警告级别日志
// Warn outputs warning level log
func (m *multiLogger) Warn(args ...interface{}) {
	for _, logger := range m.loggers {
		logger.Warn(args...)
	}
}

// Warnf 输出格式化警告级别日志
// Warnf outputs formatted warning level log
func (m *multiLogger) Warnf(format string, args ...interface{}) {
	for _, logger := range m.loggers {
		logger.Warnf(format, args...)
	}
}

// Error 输出错误级别日志
// Error outputs error level log
func (m *multiLogger) Error(args ...interface{}) {
	for _, logger := range m.loggers {
		logger.Error(args...)
	}
}

// Errorf 输出格式化错误级别日志
// Errorf outputs formatted error level log
func (m *multiLogger) Errorf(format string, args ...interface{}) {
	for _, logger := range m.loggers {
		logger.Errorf(format, args...)
	}
}

// Fatal 输出致命错误级别日志并退出程序
// Fatal outputs fatal level log and exits program
func (m *multiLogger) Fatal(args ...interface{}) {
	for _, logger := range m.loggers {
		logger.Fatal(args...)
	}
}

// Fatalf 输出格式化致命错误级别日志并退出程序
// Fatalf outputs formatted fatal level log and exits program
func (m *multiLogger) Fatalf(format string, args ...interface{}) {
	for _, logger := range m.loggers {
		logger.Fatalf(format, args...)
	}
}

// Close 关闭日志器
// Close closes the logger
func (m *multiLogger) Close() error {
	var lastErr error
	for _, logger := range m.loggers {
		if err := logger.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// 性能监控日志器 Performance monitoring logger

// PerfLogger 性能日志器
// PerfLogger performance logger
type PerfLogger struct {
	Logger
	slowThreshold time.Duration
}

// NewPerfLogger 创建性能日志器
// NewPerfLogger creates performance logger
func NewPerfLogger(logger Logger, slowThreshold time.Duration) *PerfLogger {
	return &PerfLogger{
		Logger:        logger,
		slowThreshold: slowThreshold,
	}
}

// LogDuration 记录执行时长
// LogDuration logs execution duration
func (p *PerfLogger) LogDuration(operation string, start time.Time, fields ...Fields) {
	duration := time.Since(start)

	logFields := Fields{
		"operation":   operation,
		"duration":    duration.String(),
		"duration_ms": duration.Milliseconds(),
	}

	// 合并额外字段
	// Merge additional fields
	for _, f := range fields {
		for k, v := range f {
			logFields[k] = v
		}
	}

	logger := p.WithFields(logFields)

	if duration > p.slowThreshold {
		logger.Warn("Slow operation detected")
	} else {
		logger.Debug("Operation completed")
	}
}

// Cleanup 清理函数
// Cleanup cleanup function
func Cleanup() {
	globalMu.Lock()
	defer globalMu.Unlock()

	if globalLogger != nil {
		globalLogger.Close()
		globalLogger = nil
	}
}

// init 初始化函数
// init initialization function
func init() {
	// 创建默认全局日志器
	// Create default global logger
	if err := InitGlobalLogger(DefaultConfig); err != nil {
		// 如果创建失败，使用标准输出
		// If creation fails, use stdout
		config := &Config{
			Level:  LevelInfo,
			Format: FormatText,
			Output: "stdout",
			Async:  false,
		}
		InitGlobalLogger(config)
	}
}
