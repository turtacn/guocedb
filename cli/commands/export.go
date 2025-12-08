package commands

import (
	"database/sql"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/turtacn/guocedb/internal/export"
	_ "github.com/go-sql-driver/mysql"
)

// NewExportCmd creates the export command.
func NewExportCmd() *cobra.Command {
	var (
		database   string
		tables     []string
		output     string
		schemaOnly bool
		dataOnly   bool
		addr       string
	)
	
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export database to SQL dump",
		Long: `Export database schema and/or data to SQL dump format.
The output can be used to recreate the database structure and data.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(addr, database, tables, output, schemaOnly, dataOnly)
		},
	}
	
	cmd.Flags().StringVar(&database, "database", "", "database to export (required)")
	cmd.Flags().StringSliceVar(&tables, "tables", nil, "specific tables to export (default: all tables)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file (default: stdout)")
	cmd.Flags().BoolVar(&schemaOnly, "schema-only", false, "export schema only (no data)")
	cmd.Flags().BoolVar(&dataOnly, "data-only", false, "export data only (no schema)")
	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:3306", "server address")
	
	cmd.MarkFlagRequired("database")
	
	return cmd
}

func runExport(addr, database string, tables []string, output string, schemaOnly, dataOnly bool) error {
	// Validate flags
	if schemaOnly && dataOnly {
		return fmt.Errorf("cannot specify both --schema-only and --data-only")
	}
	
	// 1. Connect to the server
	dsn := fmt.Sprintf("root@tcp(%s)/%s", addr, database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	
	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("cannot connect to database %s at %s: %w", database, addr, err)
	}
	
	// 2. Determine output destination
	var w io.Writer = os.Stdout
	if output != "" {
		f, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", output, err)
		}
		defer f.Close()
		w = f
	}
	
	// 3. Create dumper
	dumper := export.NewDumper(db, w)
	dumper.SchemaOnly = schemaOnly
	dumper.DataOnly = dataOnly
	
	// 4. Get table list
	if len(tables) == 0 {
		tables, err = dumper.ListTables()
		if err != nil {
			return fmt.Errorf("failed to list tables: %w", err)
		}
		
		if len(tables) == 0 {
			fmt.Fprintf(os.Stderr, "Warning: No tables found in database %s\n", database)
			return nil
		}
	}
	
	// 5. Write dump header
	dumper.WriteHeader(database)
	
	// 6. Export each table
	for _, table := range tables {
		// Verify table exists
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ?", 
			database, table).Scan(&count)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not verify table %s exists: %v\n", table, err)
		} else if count == 0 {
			fmt.Fprintf(os.Stderr, "Warning: Table %s does not exist in database %s\n", table, database)
			continue
		}
		
		// Export schema
		if !dataOnly {
			if err := dumper.WriteTableSchema(table); err != nil {
				return fmt.Errorf("failed to export schema for table %s: %w", table, err)
			}
		}
		
		// Export data
		if !schemaOnly {
			if err := dumper.WriteTableData(table); err != nil {
				return fmt.Errorf("failed to export data for table %s: %w", table, err)
			}
		}
	}
	
	// 7. Write dump footer
	dumper.WriteFooter()
	
	if output != "" {
		fmt.Fprintf(os.Stderr, "Export completed successfully. Output written to %s\n", output)
	}
	
	return nil
}