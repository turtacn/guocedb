// Package commands provides CLI command implementations.
package commands

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// Version information injected at build time
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

// NewVersionCmd creates the version command.
func NewVersionCmd() *cobra.Command {
	var short bool
	
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			if short {
				fmt.Println(Version)
				return
			}
			
			fmt.Printf("GuoceDB version %s\n", Version)
			fmt.Printf("  Git commit: %s\n", GitCommit)
			fmt.Printf("  Build time: %s\n", BuildTime)
			fmt.Printf("  Go version: %s\n", runtime.Version())
			fmt.Printf("  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
	
	cmd.Flags().BoolVar(&short, "short", false, "print version only")
	
	return cmd
}