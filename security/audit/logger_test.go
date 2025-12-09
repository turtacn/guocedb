package audit

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAuditLogWrite(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "audit.log")
	
	logger, err := NewAuditLogger(AuditConfig{
		FilePath: tmpFile,
		Async:    false,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	
	event := NewAuthenticationEvent("testuser", "127.0.0.1", true)
	logger.Log(event)
	
	// Read and verify
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	content := string(data)
	if !containsString(content, "testuser") {
		t.Error("Log should contain username")
	}
	if !containsString(content, "AUTHENTICATION") {
		t.Error("Log should contain event type")
	}
	if !containsString(content, "SUCCESS") {
		t.Error("Log should contain success result")
	}
}

func TestAuditLogJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := &AuditLogger{
		writer:    &buf,
		bufWriter: bufio.NewWriter(&buf),
	}
	
	event := &AuditEvent{
		Timestamp: time.Now(),
		EventType: EventTypeQuery,
		Username:  "user1",
		Statement: "SELECT * FROM t1",
		Result:    ResultSuccess,
	}
	logger.writeEvent(event)
	
	// Verify JSON format
	var parsed map[string]interface{}
	err := json.Unmarshal(buf.Bytes()[:buf.Len()-1], &parsed) // Remove trailing newline
	if err != nil {
		t.Fatalf("Log should be valid JSON: %v", err)
	}
	
	if parsed["username"] != "user1" {
		t.Errorf("Expected username 'user1', got %v", parsed["username"])
	}
	if parsed["event_type"] != "QUERY" {
		t.Errorf("Expected event_type 'QUERY', got %v", parsed["event_type"])
	}
}

func TestAuditLogAsync(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "audit_async.log")
	
	logger, err := NewAuditLogger(AuditConfig{
		FilePath:   tmpFile,
		Async:      true,
		BufferSize: 100,
	})
	if err != nil {
		t.Fatalf("Failed to create async logger: %v", err)
	}
	
	// Log multiple events
	for i := 0; i < 10; i++ {
		event := NewAuthenticationEvent("user", "127.0.0.1", true)
		logger.Log(event)
	}
	
	// Close to ensure all events are written
	logger.Close()
	
	// Verify events were written
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	// Count lines (each event is one line)
	lines := bytes.Count(data, []byte("\n"))
	if lines != 10 {
		t.Errorf("Expected 10 events, got %d", lines)
	}
}

func TestAuditLogFilter(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "audit_filter.log")
	
	logger, err := NewAuditLogger(AuditConfig{
		FilePath:   tmpFile,
		Async:      false,
		ExcludeIPs: []string{"192.168.1.1"},
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()
	
	// This should be filtered
	event1 := NewAuthenticationEvent("user1", "192.168.1.1", true)
	logger.Log(event1)
	
	// This should be logged
	event2 := NewAuthenticationEvent("user2", "127.0.0.1", true)
	logger.Log(event2)
	
	// Read and verify
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	content := string(data)
	if containsString(content, "user1") {
		t.Error("Filtered IP should not be logged")
	}
	if !containsString(content, "user2") {
		t.Error("Non-filtered IP should be logged")
	}
}

func TestNewAuthenticationEvent(t *testing.T) {
	event := NewAuthenticationEvent("testuser", "127.0.0.1", true)
	
	if event.EventType != EventTypeAuthentication {
		t.Error("Event type should be AUTHENTICATION")
	}
	if event.Result != ResultSuccess {
		t.Error("Result should be SUCCESS for successful auth")
	}
	if event.Username != "testuser" {
		t.Error("Username should match")
	}
	
	event2 := NewAuthenticationEvent("testuser", "127.0.0.1", false)
	if event2.Result != ResultFailure {
		t.Error("Result should be FAILURE for failed auth")
	}
}

func TestNewQueryEvent(t *testing.T) {
	event := NewQueryEvent("user", "127.0.0.1", "testdb", "SELECT * FROM users", 100*time.Millisecond, 42)
	
	if event.EventType != EventTypeQuery {
		t.Error("Event type should be QUERY")
	}
	if event.Result != ResultSuccess {
		t.Error("Result should be SUCCESS")
	}
	if event.Username != "user" {
		t.Error("Username should match")
	}
	if event.Database != "testdb" {
		t.Error("Database should match")
	}
	if event.RowsAffected != 42 {
		t.Errorf("Expected 42 rows affected, got %d", event.RowsAffected)
	}
}

func TestTruncateStatement(t *testing.T) {
	longStatement := "SELECT * FROM users WHERE " + string(make([]byte, 2000))
	
	truncated := truncateStatement(longStatement, 1000)
	
	if len(truncated) > 1003 { // 1000 + "..."
		t.Errorf("Statement should be truncated to ~1000 chars, got %d", len(truncated))
	}
	
	if truncated[len(truncated)-3:] != "..." {
		t.Error("Truncated statement should end with '...'")
	}
}

func TestGetEvents(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "audit_get.log")
	
	logger, err := NewAuditLogger(AuditConfig{
		FilePath: tmpFile,
		Async:    false,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	
	// Log events at different times
	now := time.Now()
	event1 := &AuditEvent{
		Timestamp: now.Add(-2 * time.Hour),
		EventType: EventTypeQuery,
		Username:  "user1",
		Result:    ResultSuccess,
	}
	event2 := &AuditEvent{
		Timestamp: now.Add(-1 * time.Hour),
		EventType: EventTypeQuery,
		Username:  "user2",
		Result:    ResultSuccess,
	}
	event3 := &AuditEvent{
		Timestamp: now,
		EventType: EventTypeQuery,
		Username:  "user3",
		Result:    ResultSuccess,
	}
	
	logger.Log(event1)
	logger.Log(event2)
	logger.Log(event3)
	logger.Close()
	
	// Get events from last 90 minutes
	events, err := GetEvents(tmpFile, now.Add(-90*time.Minute), now.Add(time.Minute))
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}
	
	if len(events) != 2 {
		t.Errorf("Expected 2 events in range, got %d", len(events))
	}
}

// Helper functions

func containsString(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
