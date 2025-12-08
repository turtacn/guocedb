package commands

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	_ "github.com/go-sql-driver/mysql"
)

// DiagnosticInfo contains comprehensive diagnostic information.
type DiagnosticInfo struct {
	Timestamp       time.Time              `json:"timestamp"`
	System          SystemInfo             `json:"system"`
	Runtime         RuntimeInfo            `json:"runtime"`
	Server          *ServerInfo            `json:"server,omitempty"`
	Databases       []DatabaseInfo         `json:"databases,omitempty"`
	Config          map[string]interface{} `json:"config,omitempty"`
	RecentLogs      []string               `json:"recent_logs,omitempty"`
	ConnectionError string                 `json:"connection_error,omitempty"`
}

// SystemInfo contains system-level information.
type SystemInfo struct {
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	NumCPU    int    `json:"num_cpu"`
	Hostname  string `json:"hostname"`
	GoVersion string `json:"go_version"`
}

// RuntimeInfo contains Go runtime information.
type RuntimeInfo struct {
	NumGoroutine int    `json:"num_goroutine"`
	MemAlloc     uint64 `json:"mem_alloc"`
	MemSys       uint64 `json:"mem_sys"`
	NumGC        uint32 `json:"num_gc"`
	LastGC       string `json:"last_gc"`
}

// ServerInfo contains server-specific information.
type ServerInfo struct {
	Version   string            `json:"version"`
	Uptime    int64             `json:"uptime"`
	Variables map[string]string `json:"variables"`
	Status    map[string]string `json:"status"`
}

// DatabaseInfo contains database-specific information.
type DatabaseInfo struct {
	Name   string `json:"name"`
	Tables int    `json:"tables"`
	Size   string `json:"size"`
}

// NewDiagnosticCmd creates the diagnostic command.
func NewDiagnosticCmd() *cobra.Command {
	var (
		output        string
		includeConfig bool
		includeLogs   bool
		addr          string
	)
	
	cmd := &cobra.Command{
		Use:   "diagnostic",
		Short: "Collect diagnostic information",
		Long: `Collect comprehensive diagnostic information about the system,
runtime, and server status for troubleshooting purposes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiagnostic(addr, output, includeConfig, includeLogs)
		},
	}
	
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file (default: stdout)")
	cmd.Flags().BoolVar(&includeConfig, "include-config", true, "include configuration information")
	cmd.Flags().BoolVar(&includeLogs, "include-logs", false, "include recent log entries")
	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:3306", "server address")
	
	return cmd
}

func runDiagnostic(addr, output string, includeConfig, includeLogs bool) error {
	diag := &DiagnosticInfo{
		Timestamp: time.Now(),
		System:    collectSystemInfo(),
		Runtime:   collectRuntimeInfo(),
	}
	
	// Try to connect to server and collect information
	if db, err := sql.Open("mysql", fmt.Sprintf("root@tcp(%s)/", addr)); err == nil {
		defer db.Close()
		if err := db.Ping(); err == nil {
			diag.Server = collectServerInfo(db)
			diag.Databases = collectDatabaseInfo(db)
		} else {
			diag.ConnectionError = fmt.Sprintf("Failed to ping server: %v", err)
		}
	} else {
		diag.ConnectionError = fmt.Sprintf("Failed to connect: %v", err)
	}
	
	if includeConfig {
		diag.Config = collectConfigInfo()
	}
	
	if includeLogs {
		diag.RecentLogs = collectRecentLogs()
	}
	
	// Output diagnostic information
	var w io.Writer = os.Stdout
	if output != "" {
		f, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		w = f
	}
	
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(diag); err != nil {
		return fmt.Errorf("failed to encode diagnostic information: %w", err)
	}
	
	if output != "" {
		fmt.Fprintf(os.Stderr, "Diagnostic information written to %s\n", output)
	}
	
	return nil
}

func collectSystemInfo() SystemInfo {
	hostname, _ := os.Hostname()
	return SystemInfo{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		NumCPU:    runtime.NumCPU(),
		Hostname:  hostname,
		GoVersion: runtime.Version(),
	}
}

func collectRuntimeInfo() RuntimeInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	lastGC := "never"
	if m.LastGC > 0 {
		lastGC = time.Unix(0, int64(m.LastGC)).Format(time.RFC3339)
	}
	
	return RuntimeInfo{
		NumGoroutine: runtime.NumGoroutine(),
		MemAlloc:     m.Alloc,
		MemSys:       m.Sys,
		NumGC:        m.NumGC,
		LastGC:       lastGC,
	}
}

func collectServerInfo(db *sql.DB) *ServerInfo {
	info := &ServerInfo{
		Variables: make(map[string]string),
		Status:    make(map[string]string),
	}
	
	// Get version
	if err := db.QueryRow("SELECT VERSION()").Scan(&info.Version); err != nil {
		info.Version = "unknown"
	}
	
	// Get uptime
	var uptimeVar, uptimeValue string
	if err := db.QueryRow("SHOW STATUS LIKE 'Uptime'").Scan(&uptimeVar, &uptimeValue); err == nil {
		if uptime, err := parseUptime(uptimeValue); err == nil {
			info.Uptime = uptime
		}
	}
	
	// Collect server variables
	rows, err := db.Query("SHOW VARIABLES")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name, value string
			if err := rows.Scan(&name, &value); err == nil {
				info.Variables[name] = value
			}
		}
	}
	
	// Collect server status
	rows, err = db.Query("SHOW STATUS")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name, value string
			if err := rows.Scan(&name, &value); err == nil {
				info.Status[name] = value
			}
		}
	}
	
	return info
}

func collectDatabaseInfo(db *sql.DB) []DatabaseInfo {
	var databases []DatabaseInfo
	
	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		return databases
	}
	defer rows.Close()
	
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			continue
		}
		
		dbInfo := DatabaseInfo{
			Name: dbName,
		}
		
		// Count tables in this database
		var tableCount int
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '%s'", dbName)
		if err := db.QueryRow(countQuery).Scan(&tableCount); err == nil {
			dbInfo.Tables = tableCount
		}
		
		// Get database size (this is a simplified approach)
		dbInfo.Size = "unknown"
		
		databases = append(databases, dbInfo)
	}
	
	return databases
}

func collectConfigInfo() map[string]interface{} {
	// Collect environment variables related to GuoceDB
	config := make(map[string]interface{})
	
	envVars := []string{
		"GUOCEDB_HOST",
		"GUOCEDB_PORT",
		"GUOCEDB_DATA_DIR",
		"GUOCEDB_LOG_LEVEL",
		"GUOCEDB_LOG_FORMAT",
		"GUOCEDB_READ_TIMEOUT",
		"GUOCEDB_WRITE_TIMEOUT",
		"GUOCEDB_MAX_CONNECTIONS",
	}
	
	for _, envVar := range envVars {
		if value := os.Getenv(envVar); value != "" {
			config[envVar] = value
		}
	}
	
	return config
}

func collectRecentLogs() []string {
	// This is a placeholder implementation
	// In a real system, this would read from log files or a logging system
	return []string{
		"Log collection not implemented",
		"This would contain recent log entries",
	}
}