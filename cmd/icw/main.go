package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "2.0.0"
	buildDate = "2024-12-01"
)

var rootCmd = &cobra.Command{
	Use:   "icw",
	Short: "IC Workspace Management Tool",
	Long: `ICW manages dependencies between analog and digital components.
Design components are stored in Subversion, software tools in Git.`,
	Version: fmt.Sprintf("%s (%s)", version, buildDate),
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
