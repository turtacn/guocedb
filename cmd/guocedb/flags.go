package main

import (
	"github.com/spf13/cobra"
)

var (
	cfgFile       string
	dataDir       string
	port          int
	host          string
	logLevel      string
	enableAuth    bool
	enableMetrics bool
)

// initFlags sets up command-line flags.
func initFlags(cmd *cobra.Command) {
	// Persistent flags (available to all commands)
	pFlags := cmd.PersistentFlags()
	pFlags.StringVarP(&cfgFile, "config", "c", "", "config file path")

	// Regular flags (only for server command)
	flags := cmd.Flags()

	// Server settings
	flags.StringVar(&host, "host", "0.0.0.0", "server listen host")
	flags.IntVarP(&port, "port", "p", 3306, "server listen port")

	// Storage settings
	flags.StringVarP(&dataDir, "data-dir", "d", "./data", "data directory")

	// Logging settings
	flags.StringVar(&logLevel, "log-level", "info", "log level (debug|info|warn|error)")

	// Feature flags
	flags.BoolVar(&enableAuth, "auth", false, "enable authentication")
	flags.BoolVar(&enableMetrics, "metrics", true, "enable metrics endpoint")
}

// buildRootCmd creates the root command.
func buildRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "guocedb",
		Short: "GuoceDB - A MySQL-compatible distributed database",
		Long: `GuoceDB is a MySQL-compatible database built with Go.
It provides ACID transactions, SQL query support, and horizontal scalability.`,
		RunE: runServer,
	}

	initFlags(cmd)

	// Subcommands
	cmd.AddCommand(buildVersionCmd())
	cmd.AddCommand(buildCheckCmd())

	return cmd
}

// buildVersionCmd creates the version subcommand.
func buildVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			printVersion()
		},
	}
}

// buildCheckCmd creates the check subcommand.
func buildCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check configuration validity",
		RunE: func(cmd *cobra.Command, args []string) error {
			return checkConfig()
		},
	}
}
