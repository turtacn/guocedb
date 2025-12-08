// Package main provides the CLI entry point for GuoceDB.
package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/turtacn/guocedb/cli/commands"
)

var (
	cfgFile string
	verbose bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "guocedb",
		Short: "GuoceDB - MySQL-compatible distributed database",
		Long: `GuoceDB is a MySQL-compatible database built with Go.
It provides a distributed, scalable database solution with MySQL protocol compatibility.`,
	}
	
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	
	// Register subcommands
	rootCmd.AddCommand(
		commands.NewServeCmd(&cfgFile),
		commands.NewStatusCmd(),
		commands.NewExportCmd(),
		commands.NewDiagnosticCmd(),
		commands.NewVersionCmd(),
	)
	
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}