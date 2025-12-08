package commands

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	_ "github.com/go-sql-driver/mysql"
)

// ServerStatus holds server status information.
type ServerStatus struct {
	Version     string   `json:"version"`
	Uptime      int64    `json:"uptime"`
	Connections int      `json:"connections"`
	Databases   []string `json:"databases"`
	MemoryUsed  uint64   `json:"memory_used"`
	MemoryTotal uint64   `json:"memory_total"`
	DataDir     string   `json:"data_dir"`
}

// NewStatusCmd creates the status command.
func NewStatusCmd() *cobra.Command {
	var (
		format string
		addr   string
	)
	
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show server status",
		Long:  `Show the current status of the GuoceDB server including version, uptime, connections, and memory usage.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(addr, format)
		},
	}
	
	cmd.Flags().StringVar(&format, "format", "table", "output format: table|json|text")
	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:3306", "server address")
	
	return cmd
}

func runStatus(addr, format string) error {
	// 1. Connect to the server
	dsn := fmt.Sprintf("root@tcp(%s)/", addr)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("cannot connect to server at %s: %w", addr, err)
	}
	defer db.Close()
	
	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("cannot connect to server at %s: %w", addr, err)
	}
	
	// 2. Collect status information
	status := &ServerStatus{}
	
	// Get version
	if err := db.QueryRow("SELECT VERSION()").Scan(&status.Version); err != nil {
		status.Version = "unknown"
	}
	
	// Get uptime (try to get from SHOW STATUS)
	var uptimeVar, uptimeValue string
	if err := db.QueryRow("SHOW STATUS LIKE 'Uptime'").Scan(&uptimeVar, &uptimeValue); err == nil {
		if uptime, err := parseUptime(uptimeValue); err == nil {
			status.Uptime = uptime
		}
	}
	
	// Get connection count
	var connVar, connValue string
	if err := db.QueryRow("SHOW STATUS LIKE 'Threads_connected'").Scan(&connVar, &connValue); err == nil {
		if conn, err := parseConnections(connValue); err == nil {
			status.Connections = conn
		}
	}
	
	// Get database list
	rows, err := db.Query("SHOW DATABASES")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var dbName string
			if err := rows.Scan(&dbName); err == nil {
				status.Databases = append(status.Databases, dbName)
			}
		}
	}
	
	// Get memory usage from runtime
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	status.MemoryUsed = m.Alloc
	status.MemoryTotal = m.Sys
	
	// 3. Output the status
	return outputStatus(status, format)
}

func parseUptime(value string) (int64, error) {
	// Try to parse as seconds
	var uptime int64
	if _, err := fmt.Sscanf(value, "%d", &uptime); err != nil {
		return 0, err
	}
	return uptime, nil
}

func parseConnections(value string) (int, error) {
	var connections int
	if _, err := fmt.Sscanf(value, "%d", &connections); err != nil {
		return 0, err
	}
	return connections, nil
}

func outputStatus(s *ServerStatus, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(s)
		
	case "table":
		fmt.Printf("%-20s %s\n", "Version:", s.Version)
		fmt.Printf("%-20s %s\n", "Uptime:", formatUptime(s.Uptime))
		fmt.Printf("%-20s %d\n", "Connections:", s.Connections)
		fmt.Printf("%-20s %s\n", "Memory Used:", humanize.Bytes(s.MemoryUsed))
		fmt.Printf("%-20s %s\n", "Memory Total:", humanize.Bytes(s.MemoryTotal))
		if len(s.Databases) > 0 {
			fmt.Printf("%-20s %s\n", "Databases:", strings.Join(s.Databases, ", "))
		} else {
			fmt.Printf("%-20s %s\n", "Databases:", "none")
		}
		
	case "text":
		fmt.Printf("version=%s uptime=%d connections=%d memory_used=%d memory_total=%d databases=%d\n",
			s.Version, s.Uptime, s.Connections, s.MemoryUsed, s.MemoryTotal, len(s.Databases))
		
	default:
		return fmt.Errorf("unsupported format: %s (supported: table, json, text)", format)
	}
	
	return nil
}

func formatUptime(seconds int64) string {
	if seconds == 0 {
		return "unknown"
	}
	
	duration := time.Duration(seconds) * time.Second
	
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, secs)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	} else {
		return fmt.Sprintf("%ds", secs)
	}
}