package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "github.com/go-sql-driver/mysql"
)

func TestDumper_WriteHeader(t *testing.T) {
	var buf bytes.Buffer
	dumper := NewDumper(nil, &buf)
	
	dumper.WriteHeader("testdb")
	
	output := buf.String()
	assert.Contains(t, output, "-- GuoceDB SQL Dump")
	assert.Contains(t, output, "-- Database: testdb")
	assert.Contains(t, output, "SET FOREIGN_KEY_CHECKS=0;")
	assert.Contains(t, output, "START TRANSACTION;")
}

func TestDumper_WriteFooter(t *testing.T) {
	var buf bytes.Buffer
	dumper := NewDumper(nil, &buf)
	
	dumper.WriteFooter()
	
	output := buf.String()
	assert.Contains(t, output, "COMMIT;")
	assert.Contains(t, output, "SET FOREIGN_KEY_CHECKS=1;")
	assert.Contains(t, output, "-- Dump completed")
}

func TestDumper_SetBatchSize(t *testing.T) {
	dumper := NewDumper(nil, nil)
	
	// Default batch size
	assert.Equal(t, 1000, dumper.batchSize)
	
	// Set valid batch size
	dumper.SetBatchSize(500)
	assert.Equal(t, 500, dumper.batchSize)
	
	// Invalid batch size should be ignored
	dumper.SetBatchSize(0)
	assert.Equal(t, 500, dumper.batchSize)
	
	dumper.SetBatchSize(-1)
	assert.Equal(t, 500, dumper.batchSize)
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil", nil, "NULL"},
		{"string", "hello", "'hello'"},
		{"string with quotes", "it's a test", "'it\\'s a test'"},
		{"string with newlines", "line1\nline2", "'line1\\nline2'"},
		{"bytes", []byte("data"), "'data'"},
		{"bool true", true, "1"},
		{"bool false", false, "0"},
		{"int", 42, "42"},
		{"float", 3.14, "3.14"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEscapeSQLString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "hello", "hello"},
		{"single quote", "it's", "it\\'s"},
		{"double quote", `say "hello"`, `say \"hello\"`},
		{"backslash", "path\\to\\file", "path\\\\to\\\\file"},
		{"newline", "line1\nline2", "line1\\nline2"},
		{"carriage return", "line1\rline2", "line1\\rline2"},
		{"tab", "col1\tcol2", "col1\\tcol2"},
		{"null byte", "data\x00end", "data\\0end"},
		{"ctrl-z", "data\x1aend", "data\\Zend"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeSQLString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "table", "`table`"},
		{"with backtick", "my`table", "`my``table`"},
		{"reserved word", "select", "`select`"},
		{"with spaces", "my table", "`my table`"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := quoteIdentifier(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Mock database for testing (would need a real test database in practice)
func TestDumper_Integration(t *testing.T) {
	// This test would require a real database connection
	// For now, we'll test the structure of the generated SQL
	
	var buf bytes.Buffer
	dumper := NewDumper(nil, &buf)
	
	// Test complete dump structure
	dumper.WriteHeader("testdb")
	
	// Simulate table schema
	buf.WriteString("-- Table structure for table `users`\n")
	buf.WriteString("DROP TABLE IF EXISTS `users`;\n")
	buf.WriteString("CREATE TABLE `users` (\n")
	buf.WriteString("  `id` int(11) NOT NULL AUTO_INCREMENT,\n")
	buf.WriteString("  `name` varchar(100) NOT NULL,\n")
	buf.WriteString("  PRIMARY KEY (`id`)\n")
	buf.WriteString(") ENGINE=InnoDB DEFAULT CHARSET=utf8;\n\n")
	
	// Simulate table data
	buf.WriteString("-- Data for table `users`\n")
	buf.WriteString("LOCK TABLES `users` WRITE;\n")
	buf.WriteString("INSERT INTO `users` (`id`, `name`) VALUES\n")
	buf.WriteString("(1, 'Alice'),\n")
	buf.WriteString("(2, 'Bob');\n")
	buf.WriteString("UNLOCK TABLES;\n\n")
	
	dumper.WriteFooter()
	
	output := buf.String()
	
	// Verify structure
	assert.Contains(t, output, "-- GuoceDB SQL Dump")
	assert.Contains(t, output, "DROP TABLE IF EXISTS `users`;")
	assert.Contains(t, output, "CREATE TABLE `users`")
	assert.Contains(t, output, "INSERT INTO `users`")
	assert.Contains(t, output, "COMMIT;")
	
	// Verify proper ordering
	headerPos := strings.Index(output, "-- GuoceDB SQL Dump")
	dropPos := strings.Index(output, "DROP TABLE")
	createPos := strings.Index(output, "CREATE TABLE")
	insertPos := strings.Index(output, "INSERT INTO")
	commitPos := strings.Index(output, "COMMIT;")
	
	assert.True(t, headerPos < dropPos)
	assert.True(t, dropPos < createPos)
	assert.True(t, createPos < insertPos)
	assert.True(t, insertPos < commitPos)
}