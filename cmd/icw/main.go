package main

import (
	"fmt"
	"os"

	"github.com/jakobsen/icw/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "icw",
	Short: "IC Workspace Management Tool",
	Long: `ICW manages dependencies between analog and digital components.
Design components are stored in Subversion, software tools in Git.

Environment Variables:
  ICW_REPO     Repository name (required)
  ICW_SVN_URL  SVN server URL (default: svn://anyvej11.dk)
  USER         Username for SVN authentication

Quick Start:
  export ICW_REPO=icworks
  icw test        # Test server connection
  icw update      # Sync workspace with repository`,
	Version:                    version.Short(),
	SuggestionsMinimumDistance: 2,
	SilenceErrors:              false,
	SilenceUsage:               false,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
