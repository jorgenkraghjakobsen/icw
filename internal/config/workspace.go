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
	example := `################################################################################
# ICW Workspace Configuration
################################################################################
#
# REQUIRED: Set environment variable before running icw:
#   export ICW_REPO=your_repo_name
#
# Syntax:
#   use component("path/to/component", "type", "branch")
#   use component("path/to/component", "type")          # defaults to trunk
#   use component("path/to/component")                  # infers type from path
#   use ref("/absolute/path/to/local/component")        # local reference
#
# Component Types:
#   analog   - Analog/mixed-signal components (SVN)
#   digital  - Digital HDL components (SVN)
#   setup    - Setup/configuration scripts (SVN)
#   process  - Process technology files (SVN)
#   tools    - Software tools and scripts (Git)
#
# Branch Formats:
#   SVN: trunk, tags/v1.0.0, branches/feature_name
#   Git: main, develop, tags/v2.1.0, feature/new-feature
#
# Dependencies:
#   Component dependencies are automatically resolved from depend.config
#   files in each component directory.
#
################################################################################
# Examples
################################################################################
#
# Released analog component:
#   use component("analog/bandgap_1v2", "analog", "tags/v2.1.0")
#
# Development version:
#   use component("analog/opamp_folded", "analog", "trunk")
#
# Feature branch:
#   use component("digital/spi_master", "digital", "branches/feature_x")
#
# Auto-detect type from path:
#   use component("setup/analog")
#
# Git-based tools:
#   use component("tools/layout_scripts", "tools", "main")
#
# Local development component:
#   use ref("/home/user/local_dev/custom_cell")
#
################################################################################
# Your Components - Add your components below
################################################################################

# Uncomment and edit these examples:
# use component("analog/bias", "analog", "trunk")
# use component("digital/top", "digital", "tags/v1.0")
# use component("setup/analog", "setup")

`

	if err := os.WriteFile(configPath, []byte(example), 0644); err != nil {
		return fmt.Errorf("failed to create workspace.config: %w", err)
	}

	return nil
}
