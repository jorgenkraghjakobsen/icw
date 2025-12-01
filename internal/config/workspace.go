package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindWorkspaceRoot searches for workspace.config starting from current directory
// and walking up the directory tree
func FindWorkspaceRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	dir := cwd
	for {
		configPath := filepath.Join(dir, "workspace.config")
		if _, err := os.Stat(configPath); err == nil {
			// Found workspace.config
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding workspace.config
			return "", fmt.Errorf("not in an ICW workspace (no workspace.config found)")
		}
		dir = parent
	}
}

// WorkspaceExists checks if a workspace.config exists in the given directory
func WorkspaceExists(dir string) bool {
	configPath := filepath.Join(dir, "workspace.config")
	_, err := os.Stat(configPath)
	return err == nil
}

// CreateWorkspaceConfig creates a new workspace.config with example content
func CreateWorkspaceConfig(dir string) error {
	configPath := filepath.Join(dir, "workspace.config")

	// Check if it already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("workspace.config already exists")
	}

	// Create example config
	example := `# ICW Workspace Configuration
#
# Syntax:
#   use component("path/to/component", "type", "branch")
#   use component("path/to/component", "type")          # defaults to trunk
#   use component("path/to/component")                  # infers type from path
#   use ref("/absolute/path/to/local/component")        # local reference
#
# Types: analog, digital, setup, process, tools
# VCS: analog/digital/setup/process use SVN, tools use Git
#
# Examples:
#   use component("analog/bias", "analog", "trunk")
#   use component("digital/top", "digital", "tags/v1.0")
#   use component("setup/analog", "setup")
#   use component("tools/cad_utils", "tools", "main")   # Git branch

`

	if err := os.WriteFile(configPath, []byte(example), 0644); err != nil {
		return fmt.Errorf("failed to create workspace.config: %w", err)
	}

	return nil
}
