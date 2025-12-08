// Package export provides database export functionality.
package export

import (
	"database/sql"
	"fmt"
	"io"
	"strings"
	"time"
)

// Dumper handles SQL dump generation.
type Dumper struct {
	db         *sql.DB
	w          io.Writer
	SchemaOnly bool
	DataOnly   bool
	batchSize  int
}

// NewDumper creates a new SQL dumper.
func NewDumper(db *sql.DB, w io.Writer) *Dumper {
	return &Dumper{
		db:        db,
		w:         w,
		batchSize: 1000,
	}
}

// SetBatchSize sets the batch size for INSERT statements.
func (d *Dumper) SetBatchSize(size int) {
	if size > 0 {
		d.batchSize = size
	}
}

// ListTables returns a list of all tables in the current database.
func (d *Dumper) ListTables() ([]string, error) {
	rows, err := d.db.Query("SHOW TABLES")
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	defer rows.Close()
	
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, name)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tables: %w", err)
	}
	
	return tables, nil
}

// WriteHeader writes the dump file header.
func (d *Dumper) WriteHeader(database string) {
	fmt.Fprintf(d.w, "-- GuoceDB SQL Dump\n")
	fmt.Fprintf(d.w, "-- Database: %s\n", database)
	fmt.Fprintf(d.w, "-- Generated: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(d.w, "-- Server version: GuoceDB\n\n")
	fmt.Fprintf(d.w, "SET FOREIGN_KEY_CHECKS=0;\n")
	fmt.Fprintf(d.w, "SET SQL_MODE='NO_AUTO_VALUE_ON_ZERO';\n")
	fmt.Fprintf(d.w, "SET AUTOCOMMIT=0;\n")
	fmt.Fprintf(d.w, "START TRANSACTION;\n\n")
}

// WriteTableSchema writes the CREATE TABLE statement for a table.
func (d *Dumper) WriteTableSchema(table string) error {
	row := d.db.QueryRow("SHOW CREATE TABLE " + quoteIdentifier(table))
	var tableName, createStmt string
	if err := row.Scan(&tableName, &createStmt); err != nil {
		return fmt.Errorf("failed to get CREATE TABLE for %s: %w", table, err)
	}
	
	fmt.Fprintf(d.w, "-- Table structure for table %s\n", quoteIdentifier(table))
	fmt.Fprintf(d.w, "DROP TABLE IF EXISTS %s;\n", quoteIdentifier(table))
	fmt.Fprintf(d.w, "%s;\n\n", createStmt)
	
	return nil
}

// WriteTableData writes the INSERT statements for a table's data.
func (d *Dumper) WriteTableData(table string) error {
	// Get column information
	rows, err := d.db.Query("SELECT * FROM " + quoteIdentifier(table))
	if err != nil {
		return fmt.Errorf("failed to select from table %s: %w", table, err)
	}
	defer rows.Close()
	
	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns for table %s: %w", table, err)
	}
	
	if len(cols) == 0 {
		return nil
	}
	
	fmt.Fprintf(d.w, "-- Data for table %s\n", quoteIdentifier(table))
	
	// Check if table has any data
	hasData := false
	values := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range values {
		ptrs[i] = &values[i]
	}
	
	batch := []string{}
	for rows.Next() {
		if !hasData {
			hasData = true
			fmt.Fprintf(d.w, "LOCK TABLES %s WRITE;\n", quoteIdentifier(table))
		}
		
		if err := rows.Scan(ptrs...); err != nil {
			return fmt.Errorf("failed to scan row from table %s: %w", table, err)
		}
		
		valStrs := make([]string, len(cols))
		for i, v := range values {
			valStrs[i] = formatValue(v)
		}
		batch = append(batch, "("+strings.Join(valStrs, ", ")+")")
		
		if len(batch) >= d.batchSize {
			if err := d.flushBatch(table, cols, batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}
	
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows from table %s: %w", table, err)
	}
	
	if len(batch) > 0 {
		if err := d.flushBatch(table, cols, batch); err != nil {
			return err
		}
	}
	
	if hasData {
		fmt.Fprintf(d.w, "UNLOCK TABLES;\n")
	}
	fmt.Fprintf(d.w, "\n")
	
	return nil
}

// flushBatch writes a batch of INSERT statements.
func (d *Dumper) flushBatch(table string, cols []string, values []string) error {
	quotedCols := make([]string, len(cols))
	for i, col := range cols {
		quotedCols[i] = quoteIdentifier(col)
	}
	
	fmt.Fprintf(d.w, "INSERT INTO %s (%s) VALUES\n", 
		quoteIdentifier(table), strings.Join(quotedCols, ", "))
	fmt.Fprintf(d.w, "%s;\n", strings.Join(values, ",\n"))
	
	return nil
}

// WriteFooter writes the dump file footer.
func (d *Dumper) WriteFooter() {
	fmt.Fprintf(d.w, "COMMIT;\n")
	fmt.Fprintf(d.w, "SET FOREIGN_KEY_CHECKS=1;\n")
	fmt.Fprintf(d.w, "SET SQL_MODE='';\n")
	fmt.Fprintf(d.w, "SET AUTOCOMMIT=1;\n")
	fmt.Fprintf(d.w, "-- Dump completed on %s\n", time.Now().Format(time.RFC3339))
}

// formatValue formats a value for SQL output.
func formatValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	
	switch val := v.(type) {
	case []byte:
		return fmt.Sprintf("'%s'", escapeSQLString(string(val)))
	case string:
		return fmt.Sprintf("'%s'", escapeSQLString(val))
	case time.Time:
		return fmt.Sprintf("'%s'", val.Format("2006-01-02 15:04:05"))
	case bool:
		if val {
			return "1"
		}
		return "0"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// escapeSQLString escapes special characters in SQL strings.
func escapeSQLString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	s = strings.ReplaceAll(s, "\x00", "\\0")
	s = strings.ReplaceAll(s, "\x1a", "\\Z")
	return s
}

// quoteIdentifier quotes SQL identifiers (table names, column names).
func quoteIdentifier(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}