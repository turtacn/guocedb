// Package log 提供了 GuoceDB 的结构化日志系统
// Package log provides structured logging system for GuoceDB
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
	"sync/atomic"
	"time"

	"github.com/guocedb/guocedb/common/constants"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ===== 日志级别 Log Levels =====

// Level 日志级别类型
// Level log level type
type Level int32

const (
	// DebugLevel 调试级别
	// DebugLevel debug level
	DebugLevel Level = iota
	// InfoLevel 信息级别
	// InfoLevel info level
	InfoLevel
	// WarnLevel 警告级别
	// WarnLevel warning level
	WarnLevel
	// ErrorLevel 错误级别
	// ErrorLevel error level
	ErrorLevel
	// FatalLevel 致命错误级别
	// FatalLevel fatal error level
	FatalLevel
	// PanicLevel 恐慌级别
	// PanicLevel panic level
	PanicLevel
)

// String 返回日志级别的字符串表示
// String returns string representation of log level
func (l Level) String() string {
	switch l {
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
	case PanicLevel:
		return "PANIC"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel 解析日志级别字符串
// ParseLevel parses log level string
func ParseLevel(level string) (Level, error) {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DebugLevel, nil
	case "INFO":
		return InfoLevel, nil
	case "WARN", "WARNING":
		return WarnLevel, nil
	case "ERROR":
		return ErrorLevel, nil
	case "FATAL":
		return FatalLevel, nil
	case "PANIC":
		return PanicLevel, nil
	default:
		return InfoLevel, fmt.Errorf("invalid log level: %s", level)
	}
}

// ===== 日志字段 Log Fields =====

// Fields 日志字段类型
// Fields log fields type
type Fields map[string]interface{}

// ===== 日志格式化器 Log Formatter =====

// Formatter 日志格式化器接口
// Formatter log formatter interface
type Formatter interface {
	Format(entry *Entry) ([]byte, error)
}

// TextFormatter 文本格式化器
// TextFormatter text formatter
type TextFormatter struct {
	// TimestampFormat 时间戳格式
	// TimestampFormat timestamp format
	TimestampFormat string
	// DisableColors 禁用颜色
	// DisableColors disable colors
	DisableColors bool
	// FullTimestamp 显示完整时间戳
	// FullTimestamp show full timestamp
	FullTimestamp bool
	// DisableTimestamp 禁用时间戳
	// DisableTimestamp disable timestamp
	DisableTimestamp bool
}

// Format 格式化日志条目
// Format formats log entry
func (f *TextFormatter) Format(entry *Entry) ([]byte, error) {
	var b strings.Builder

	// 时间戳
	if !f.DisableTimestamp {
		timestamp := entry.Time.Format(f.TimestampFormat)
		if f.TimestampFormat == "" {
			timestamp = entry.Time.Format(time.RFC3339)
		}
		b.WriteString(timestamp)
		b.WriteString(" ")
	}

	// 日志级别
	levelText := strings.ToUpper(entry.Level.String())
	if !f.DisableColors {
		levelText = f.colorize(levelText, entry.Level)
	}
	b.WriteString("[")
	b.WriteString(levelText)
	b.WriteString("] ")

	// 消息
	b.WriteString(entry.Message)

	// 字段
	if len(entry.Fields) > 0 {
		b.WriteString(" ")
		first := true
		for k, v := range entry.Fields {
			if !first {
				b.WriteString(" ")
			}
			b.WriteString(k)
			b.WriteString("=")
			b.WriteString(fmt.Sprintf("%v", v))
			first = false
		}
	}

	// 调用信息
	if entry.Caller != "" {
		b.WriteString(" ")
		b.WriteString(entry.Caller)
	}

	b.WriteString("\n")
	return []byte(b.String()), nil
}

// colorize 为日志级别添加颜色
// colorize adds color to log level
func (f *TextFormatter) colorize(text string, level Level) string {
	if f.DisableColors {
		return text
	}

	var color string
	switch level {
	case DebugLevel:
		color = "\033[36m" // Cyan
	case InfoLevel:
		color = "\033[32m" // Green
	case WarnLevel:
		color = "\033[33m" // Yellow
	case ErrorLevel:
		color = "\033[31m" // Red
	case FatalLevel, PanicLevel:
		color = "\033[35m" // Magenta
	default:
		color = "\033[0m" // Reset
	}
	return color + text + "\033[0m"
}

// JSONFormatter JSON格式化器
// JSONFormatter JSON formatter
type JSONFormatter struct {
	// TimestampFormat 时间戳格式
	// TimestampFormat timestamp format
	TimestampFormat string
	// DisableTimestamp 禁用时间戳
	// DisableTimestamp disable timestamp
	DisableTimestamp bool
	// PrettyPrint 美化输出
	// PrettyPrint pretty print
	PrettyPrint bool
}

// Format 格式化日志条目为JSON
// Format formats log entry as JSON
func (f *JSONFormatter) Format(entry *Entry) ([]byte, error) {
	data := make(Fields, len(entry.Fields)+4)

	// 复制字段
	for k, v := range entry.Fields {
		data[k] = v
	}

	// 添加标准字段
	if !f.DisableTimestamp {
		timestampFormat := f.TimestampFormat
		if timestampFormat == "" {
			timestampFormat = time.RFC3339Nano
		}
		data["time"] = entry.Time.Format(timestampFormat)
	}

	data["level"] = entry.Level.String()
	data["msg"] = entry.Message

	if entry.Caller != "" {
		data["caller"] = entry.Caller
	}

	// 转换为JSON
	var output []byte
	var err error
	if f.PrettyPrint {
		output, err = json.MarshalIndent(data, "", "  ")
	} else {
		output, err = json.Marshal(data)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to marshal log entry: %w", err)
	}

	return append(output, '\n'), nil
}

// ===== 日志条目 Log Entry =====

// Entry 日志条目
// Entry log entry
type Entry struct {
	Logger  *Logger
	Level   Level
	Time    time.Time
	Message string
	Fields  Fields
	Caller  string
	Context context.Context
}

// WithField 添加单个字段
// WithField adds single field
func (e *Entry) WithField(key string, value interface{}) *Entry {
	return e.WithFields(Fields{key: value})
}

// WithFields 添加多个字段
// WithFields adds multiple fields
func (e *Entry) WithFields(fields Fields) *Entry {
	newFields := make(Fields, len(e.Fields)+len(fields))
	for k, v := range e.Fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &Entry{
		Logger:  e.Logger,
		Level:   e.Level,
		Time:    e.Time,
		Message: e.Message,
		Fields:  newFields,
		Caller:  e.Caller,
		Context: e.Context,
	}
}

// WithContext 设置上下文
// WithContext sets context
func (e *Entry) WithContext(ctx context.Context) *Entry {
	return &Entry{
		Logger:  e.Logger,
		Level:   e.Level,
		Time:    e.Time,
		Message: e.Message,
		Fields:  e.Fields,
		Caller:  e.Caller,
		Context: ctx,
	}
}

// ===== 日志钩子 Log Hooks =====

// Hook 日志钩子接口
// Hook log hook interface
type Hook interface {
	Levels() []Level
	Fire(*Entry) error
}

// ===== 日志输出 Log Output =====

// Output 日志输出接口
// Output log output interface
type Output interface {
	Write(entry *Entry) error
	Close() error
}

// WriterOutput 基于 io.Writer 的输出
// WriterOutput writer-based output
type WriterOutput struct {
	Writer    io.Writer
	Formatter Formatter
	mu        sync.Mutex
}

// Write 写入日志
// Write writes log
func (w *WriterOutput) Write(entry *Entry) error {
	formatted, err := w.Formatter.Format(entry)
	if err != nil {
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	_, err = w.Writer.Write(formatted)
	return err
}

// Close 关闭输出
// Close closes output
func (w *WriterOutput) Close() error {
	if closer, ok := w.Writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// FileOutput 文件输出
// FileOutput file output
type FileOutput struct {
	*WriterOutput
	writer *lumberjack.Logger
}

// NewFileOutput 创建文件输出
// NewFileOutput creates file output
func NewFileOutput(filename string, formatter Formatter, maxSize, maxBackups, maxAge int) *FileOutput {
	writer := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,    // megabytes
		MaxBackups: maxBackups,
		MaxAge:     maxAge, // days
		Compress:   true,
	}

	return &FileOutput{
		WriterOutput: &WriterOutput{
			Writer:    writer,
			Formatter: formatter,
		},
		writer: writer,
	}
}

// ===== 异步写入器 Async Writer =====

// AsyncOutput 异步输出
// AsyncOutput async output
type AsyncOutput struct {
	output    Output
	entries   chan *Entry
	done      chan struct{}
	wg        sync.WaitGroup
	bufferLen int
}

// NewAsyncOutput 创建异步输出
// NewAsyncOutput creates async output
func NewAsyncOutput(output Output, bufferLen int) *AsyncOutput {
	a := &AsyncOutput{
		output:    output,
		entries:   make(chan *Entry, bufferLen),
		done:      make(chan struct{}),
		bufferLen: bufferLen,
	}

	a.wg.Add(1)
	go a.run()

	return a
}

// Write 异步写入日志
// Write writes log asynchronously
func (a *AsyncOutput) Write(entry *Entry) error {
	select {
	case a.entries <- entry:
		return nil
	default:
		// 缓冲区满，丢弃日志
		return fmt.Errorf("async buffer full")
	}
}

// Close 关闭异步输出
// Close closes async output
func (a *AsyncOutput) Close() error {
	close(a.done)
	a.wg.Wait()
	close(a.entries)
	return a.output.Close()
}

// run 运行异步写入循环
// run runs async write loop
func (a *AsyncOutput) run() {
	defer a.wg.Done()

	for {
		select {
		case entry := <-a.entries:
			if entry != nil {
				_ = a.output.Write(entry)
			}
		case <-a.done:
			// 处理剩余的日志
			for entry := range a.entries {
				if entry != nil {
					_ = a.output.Write(entry)
				}
			}
			return
		}
	}
}

// ===== 日志记录器 Logger =====

// Logger 日志记录器
// Logger log recorder
type Logger struct {
	outputs      []Output
	hooks        []Hook
	level        int32 // atomic
	reportCaller bool
	mu           sync.RWMutex
	fields       Fields
	exitFunc     func(int)
}

// New 创建新的日志记录器
// New creates new logger
func New() *Logger {
	return &Logger{
		outputs:  make([]Output, 0),
		hooks:    make([]Hook, 0),
		level:    int32(InfoLevel),
		fields:   make(Fields),
		exitFunc: os.Exit,
	}
}

// SetLevel 设置日志级别
// SetLevel sets log level
func (l *Logger) SetLevel(level Level) {
	atomic.StoreInt32(&l.level, int32(level))
}

// GetLevel 获取日志级别
// GetLevel gets log level
func (l *Logger) GetLevel() Level {
	return Level(atomic.LoadInt32(&l.level))
}

// SetReportCaller 设置是否报告调用者
// SetReportCaller sets whether to report caller
func (l *Logger) SetReportCaller(report bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.reportCaller = report
}

// AddOutput 添加输出
// AddOutput adds output
func (l *Logger) AddOutput(output Output) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.outputs = append(l.outputs, output)
}

// AddHook 添加钩子
// AddHook adds hook
func (l *Logger) AddHook(hook Hook) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hooks = append(l.hooks, hook)
}

// WithField 创建带字段的条目
// WithField creates entry with field
func (l *Logger) WithField(key string, value interface{}) *Entry {
	return l.newEntry().WithField(key, value)
}

// WithFields 创建带多个字段的条目
// WithFields creates entry with fields
func (l *Logger) WithFields(fields Fields) *Entry {
	return l.newEntry().WithFields(fields)
}

// WithContext 创建带上下文的条目
// WithContext creates entry with context
func (l *Logger) WithContext(ctx context.Context) *Entry {
	return l.newEntry().WithContext(ctx)
}

// newEntry 创建新条目
// newEntry creates new entry
func (l *Logger) newEntry() *Entry {
	entry := &Entry{
		Logger: l,
		Time:   time.Now(),
		Fields: make(Fields, len(l.fields)),
	}

	// 复制全局字段
	l.mu.RLock()
	for k, v := range l.fields {
		entry.Fields[k] = v
	}
	l.mu.RUnlock()

	return entry
}

// log 记录日志
// log logs message
func (l *Logger) log(level Level, args ...interface{}) {
	if level < l.GetLevel() {
		return
	}

	entry := l.newEntry()
	entry.Level = level
	entry.Message = fmt.Sprint(args...)

	// 获取调用者信息
	if l.reportCaller {
		entry.Caller = getCaller()
	}

	// 触发钩子
	l.fireHooks(entry)

	// 写入输出
	l.write(entry)

	// 处理 Fatal 和 Panic
	if level == FatalLevel {
		l.exitFunc(1)
	} else if level == PanicLevel {
		panic(entry.Message)
	}
}

// logf 记录格式化日志
// logf logs formatted message
func (l *Logger) logf(level Level, format string, args ...interface{}) {
	if level < l.GetLevel() {
		return
	}

	entry := l.newEntry()
	entry.Level = level
	entry.Message = fmt.Sprintf(format, args...)

	// 获取调用者信息
	if l.reportCaller {
		entry.Caller = getCaller()
	}

	// 触发钩子
	l.fireHooks(entry)

	// 写入输出
	l.write(entry)

	// 处理 Fatal 和 Panic
	if level == FatalLevel {
		l.exitFunc(1)
	} else if level == PanicLevel {
		panic(entry.Message)
	}
}

// fireHooks 触发钩子
// fireHooks fires hooks
func (l *Logger) fireHooks(entry *Entry) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, hook := range l.hooks {
		for _, level := range hook.Levels() {
			if level == entry.Level {
				if err := hook.Fire(entry); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to fire hook: %v\n", err)
				}
				break
			}
		}
	}
}

// write 写入输出
// write writes to outputs
func (l *Logger) write(entry *Entry) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, output := range l.outputs {
		if err := output.Write(entry); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write log: %v\n", err)
		}
	}
}

// Debug 记录调试日志
// Debug logs debug message
func (l *Logger) Debug(args ...interface{}) {
	l.log(DebugLevel, args...)
}

// Debugf 记录格式化调试日志
// Debugf logs formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logf(DebugLevel, format, args...)
}

// Info 记录信息日志
// Info logs info message
func (l *Logger) Info(args ...interface{}) {
	l.log(InfoLevel, args...)
}

// Infof 记录格式化信息日志
// Infof logs formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logf(InfoLevel, format, args...)
}

// Warn 记录警告日志
// Warn logs warning message
func (l *Logger) Warn(args ...interface{}) {
	l.log(WarnLevel, args...)
}

// Warnf 记录格式化警告日志
// Warnf logs formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logf(WarnLevel, format, args...)
}

// Error 记录错误日志
// Error logs error message
func (l *Logger) Error(args ...interface{}) {
	l.log(ErrorLevel, args...)
}

// Errorf 记录格式化错误日志
// Errorf logs formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logf(ErrorLevel, format, args...)
}

// Fatal 记录致命错误日志并退出
// Fatal logs fatal message and exits
func (l *Logger) Fatal(args ...interface{}) {
	l.log(FatalLevel, args...)
}

// Fatalf 记录格式化致命错误日志并退出
// Fatalf logs formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logf(FatalLevel, format, args...)
}

// Panic 记录恐慌日志并触发panic
// Panic logs panic message and panics
func (l *Logger) Panic(args ...interface{}) {
	l.log(PanicLevel, args...)
}

// Panicf 记录格式化恐慌日志并触发panic
// Panicf logs formatted panic message and panics
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.logf(PanicLevel, format, args...)
}

// ===== 辅助函数 Helper Functions =====

// getCaller 获取调用者信息
// getCaller gets caller information
func getCaller() string {
	_, file, line, ok := runtime.Caller(4)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// ===== 默认日志记录器 Default Logger =====

var defaultLogger = New()

func init() {
	// 设置默认输出为标准输出
	defaultLogger.AddOutput(&WriterOutput{
		Writer: os.Stdout,
		Formatter: &TextFormatter{
			FullTimestamp: true,
		},
	})
}

// SetDefaultLogger 设置默认日志记录器
// SetDefaultLogger sets default logger
func SetDefaultLogger(logger *Logger) {
	defaultLogger = logger
}

// GetDefaultLogger 获取默认日志记录器
// GetDefaultLogger gets default logger
func GetDefaultLogger() *Logger {
	return defaultLogger
}

// ===== 包级别函数 Package Level Functions =====

// SetLevel 设置默认日志级别
// SetLevel sets default log level
func SetLevel(level Level) {
	defaultLogger.SetLevel(level)
}

// GetLevel 获取默认日志级别
// GetLevel gets default log level
func GetLevel() Level {
	return defaultLogger.GetLevel()
}

// WithField 创建带字段的条目
// WithField creates entry with field
func WithField(key string, value interface{}) *Entry {
	return defaultLogger.WithField(key, value)
}

// WithFields 创建带多个字段的条目
// WithFields creates entry with fields
func WithFields(fields Fields) *Entry {
	return defaultLogger.WithFields(fields)
}

// WithContext 创建带上下文的条目
// WithContext creates entry with context
func WithContext(ctx context.Context) *Entry {
	return defaultLogger.WithContext(ctx)
}

// Debug 记录调试日志
// Debug logs debug message
func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

// Debugf 记录格式化调试日志
// Debugf logs formatted debug message
func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

// Info 记录信息日志
// Info logs info message
func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

// Infof 记录格式化信息日志
// Infof logs formatted info message
func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

// Warn 记录警告日志
// Warn logs warning message
func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

// Warnf 记录格式化警告日志
// Warnf logs formatted warning message
func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

// Error 记录错误日志
// Error logs error message
func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

// Errorf 记录格式化错误日志
// Errorf logs formatted error message
func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

// Fatal 记录致命错误日志并退出
// Fatal logs fatal message and exits
func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

// Fatalf 记录格式化致命错误日志并退出
// Fatalf logs formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

// Panic 记录恐慌日志并触发panic
// Panic logs panic message and panics
func Panic(args ...interface{}) {
	defaultLogger.Panic(args...)
}

// Panicf 记录格式化恐慌日志并触发panic
// Panicf logs formatted panic message and panics
func Panicf(format string, args ...interface{}) {
	defaultLogger.Panicf(format, args...)
}

// ===== 特定场景的日志函数 Scenario-specific Log Functions =====

// SQL 记录SQL查询日志
// SQL logs SQL query
func SQL(query string, duration time.Duration, err error) {
	entry := WithFields(Fields{
		"query":    query,
		"duration": duration.String(),
	})

	if err != nil {
		entry.WithField("error", err.Error()).Error("SQL query failed")
	} else {
		entry.Info("SQL query executed")
	}
}

// Transaction 记录事务日志
// Transaction logs transaction
func Transaction(txID string, action string, err error) {
	entry := WithFields(Fields{
		"tx_id":  txID,
		"action": action,
	})

	if err != nil {
		entry.WithField("error", err.Error()).Error("Transaction failed")
	} else {
		entry.Info("Transaction completed")
	}
}

// Access 记录访问日志
// Access logs access
func Access(method, path string, statusCode int, duration time.Duration) {
	WithFields(Fields{
		"method":      method,
		"path":        path,
		"status_code": statusCode,
		"duration":    duration.String(),
	}).Info("Access log")
}

// Performance 记录性能日志
// Performance logs performance
func Performance(operation string, duration time.Duration, details Fields) {
	fields := Fields{
		"operation": operation,
		"duration":  duration.String(),
	}
	for k, v := range details {
		fields[k] = v
	}
	WithFields(fields).Debug("Performance log")
}

// ===== 配置辅助函数 Configuration Helper Functions =====

// ConfigureFromEnv 从环境变量配置日志
// ConfigureFromEnv configures logger from environment variables
func ConfigureFromEnv() error {
	// 日志级别
	if levelStr := os.Getenv("GUOCEDB_LOG_LEVEL"); levelStr != "" {
		level, err := ParseLevel(levelStr)
		if err != nil {
			return err
		}
		SetLevel(level)
	}

	// 日志格式
	if format := os.Getenv("GUOCEDB_LOG_FORMAT"); format != "" {
		var formatter Formatter
		switch format {
		case "json":
			formatter = &JSONFormatter{}
		case "text":
			formatter = &TextFormatter{FullTimestamp: true}
		default:
			return fmt.Errorf("unknown log format: %s", format)
		}

		// 重新配置输出
		defaultLogger.outputs = []Output{
			&WriterOutput{
				Writer:    os.Stdout,
				Formatter: formatter,
			},
		}
	}

	// 日志文件
	if logFile := os.Getenv("GUOCEDB_LOG_FILE"); logFile != "" {
		fileOutput := NewFileOutput(
			logFile,
			&JSONFormatter{},
			100, // 100MB
			7,   // 7 backups
			30,  // 30 days
		)
		defaultLogger.AddOutput(fileOutput)
	}

	// 是否报告调用者
	if reportCaller := os.Getenv("GUOCEDB_LOG_CALLER"); reportCaller == "true" {
		defaultLogger.SetReportCaller(true)
	}

	return nil
}

// ===== 性能优化的日志池 Performance-optimized Log Pool =====

var entryPool = sync.Pool{
	New: func() interface{} {
		return &Entry{
			Fields: make(Fields),
		}
	},
}

// getPooledEntry 从池中获取条目
// getPooledEntry gets entry from pool
func getPooledEntry() *Entry {
	entry := entryPool.Get().(*Entry)
	entry.Time = time.Now()
	return entry
}

// putPooledEntry 将条目放回池中
// putPooledEntry puts entry back to pool
func putPooledEntry(entry *Entry) {
	// 清理条目
	entry.Logger = nil
	entry.Level = 0
	entry.Message = ""
	entry.Caller = ""
	entry.Context = nil

	// 清理字段
	for k := range entry.Fields {
		delete(entry.Fields, k)
	}

	entryPool.Put(entry)
}

// 需要导入的包（在实际文件中应该在顶部）
// import (
// 	"encoding/json"
// )
import "encoding/json"
