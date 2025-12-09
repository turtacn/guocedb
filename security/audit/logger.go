// Package audit provides audit logging for GuoceDB.
package audit

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

// AuditLogger handles writing audit events to a log.
type AuditLogger struct {
	writer    io.Writer
	bufWriter *bufio.Writer
	mu        sync.Mutex
	async     bool
	eventChan chan *AuditEvent
	done      chan struct{}
	
	// Filter configuration
	excludeIPs []string
}

// AuditConfig configures the audit logger.
type AuditConfig struct {
	FilePath    string
	Async       bool
	BufferSize  int
	ExcludeIPs  []string
	IncludeStmt bool
}

// NewAuditLogger creates a new audit logger with the given configuration.
func NewAuditLogger(config AuditConfig) (*AuditLogger, error) {
	var writer io.Writer
	
	if config.FilePath == "" || config.FilePath == "stdout" {
		writer = os.Stdout
	} else {
		f, err := os.OpenFile(config.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			return nil, err
		}
		writer = f
	}
	
	logger := &AuditLogger{
		writer:     writer,
		bufWriter:  bufio.NewWriter(writer),
		async:      config.Async,
		excludeIPs: config.ExcludeIPs,
	}
	
	if config.Async {
		bufSize := config.BufferSize
		if bufSize <= 0 {
			bufSize = 1000
		}
		logger.eventChan = make(chan *AuditEvent, bufSize)
		logger.done = make(chan struct{})
		go logger.processLoop()
	}
	
	return logger, nil
}

// Log records an audit event.
func (l *AuditLogger) Log(event *AuditEvent) {
	if l.shouldFilter(event) {
		return
	}
	
	if l.async {
		select {
		case l.eventChan <- event:
		default:
			// Channel full, write synchronously to avoid dropping events
			l.writeEvent(event)
		}
	} else {
		l.writeEvent(event)
	}
}

// writeEvent writes a single event to the log.
func (l *AuditLogger) writeEvent(event *AuditEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	data, err := json.Marshal(event)
	if err != nil {
		// Log marshal error but don't fail
		return
	}
	
	l.bufWriter.Write(data)
	l.bufWriter.WriteByte('\n')
	l.bufWriter.Flush()
}

// processLoop processes events asynchronously.
func (l *AuditLogger) processLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case event := <-l.eventChan:
			l.writeEvent(event)
			
		case <-ticker.C:
			// Periodic flush
			l.mu.Lock()
			l.bufWriter.Flush()
			l.mu.Unlock()
			
		case <-l.done:
			// Process remaining events
			for {
				select {
				case event := <-l.eventChan:
					l.writeEvent(event)
				default:
					return
				}
			}
		}
	}
}

// shouldFilter determines if an event should be filtered out.
func (l *AuditLogger) shouldFilter(event *AuditEvent) bool {
	// Filter by IP
	for _, excludeIP := range l.excludeIPs {
		if event.ClientIP == excludeIP {
			return true
		}
	}
	
	return false
}

// Close flushes and closes the audit logger.
func (l *AuditLogger) Close() error {
	if l.async {
		close(l.done)
		// Wait a bit for processing to complete
		time.Sleep(100 * time.Millisecond)
	}
	
	l.mu.Lock()
	defer l.mu.Unlock()
	
	l.bufWriter.Flush()
	
	if closer, ok := l.writer.(io.Closer); ok {
		return closer.Close()
	}
	
	return nil
}

// GetEvents retrieves audit events within a time range (for testing/analysis).
// This is a simple implementation that reads from a file.
func GetEvents(filePath string, start, end time.Time) ([]*AuditEvent, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	
	var events []*AuditEvent
	scanner := bufio.NewScanner(f)
	
	for scanner.Scan() {
		var event AuditEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		
		if (event.Timestamp.Equal(start) || event.Timestamp.After(start)) &&
			(event.Timestamp.Equal(end) || event.Timestamp.Before(end)) {
			events = append(events, &event)
		}
	}
	
	return events, scanner.Err()
}
