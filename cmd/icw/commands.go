package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/jakobsen/icw/internal/component"
	"github.com/jakobsen/icw/internal/config"
	"github.com/jakobsen/icw/internal/svn"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(treeCmd)
	rootCmd.AddCommand(addCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Sync workspace with repository (checkout components)",
	Long:  `Updates the workspace by checking out components from the repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdate()
	},
}

func runUpdate() error {
	// Find workspace root
	root, err := config.FindWorkspaceRoot()
	if err != nil {
		// No workspace found, prompt to create one
		cwd, _ := os.Getwd()
		color.Yellow("No workspace.config found in %s or parent directories", cwd)
		fmt.Print("Create a new workspace here? [Y/n] ")

		var response string
		fmt.Scanln(&response)
		response = filepath.Clean(response)

		if response == "" || response == "y" || response == "Y" {
			if err := config.CreateWorkspaceConfig(cwd); err != nil {
				return fmt.Errorf("failed to create workspace: %w", err)
			}
			color.Green("Created workspace.config in %s", cwd)
			color.Yellow("Please edit workspace.config to add components, then run 'icw update' again")
			return nil
		}
		return fmt.Errorf("not in a workspace")
	}

	color.Cyan("Workspace root: %s", root)

	// Create workspace
	ws := component.NewWorkspace(root)

	// Parse workspace.config
	parser := config.NewParser(ws)
	if err := parser.ParseWorkspaceConfig(ws.Config); err != nil {
		return fmt.Errorf("failed to parse workspace.config: %w", err)
	}

	// Check if we have any components
	if len(ws.Components) == 0 {
		color.Yellow("No components defined in workspace.config")
		return nil
	}

	color.Green("Found %d component(s) in workspace.config", len(ws.Components))

	// Create SVN client
	svnClient, err := svn.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create SVN client: %w", err)
	}

	color.Cyan("Using repository: %s", svnClient.Repo)

	// Checkout/update each component
	for name, comp := range ws.Components {
		if comp.VCS == "local" {
			color.Blue("  [SKIP] %s (local reference)", name)
			continue
		}

		destPath := filepath.Join(root, comp.Path)

		if comp.VCS == "svn" {
			// Check if already checked out
			if svn.IsWorkingCopy(destPath) {
				color.Yellow("  [UPDATE] %s (%s)", name, comp.Branch)
				if err := svnClient.Update(destPath); err != nil {
					color.Red("    Failed: %v", err)
					continue
				}
			} else {
				color.Green("  [CHECKOUT] %s (%s)", name, comp.Branch)
				// Create parent directory if needed
				parentDir := filepath.Dir(destPath)
				if err := os.MkdirAll(parentDir, 0755); err != nil {
					color.Red("    Failed to create directory: %v", err)
					continue
				}

				if err := svnClient.Checkout(comp.Path, comp.Branch, destPath); err != nil {
					color.Red("    Failed: %v", err)
					continue
				}
			}
		} else if comp.VCS == "git" {
			color.Yellow("  [TODO] %s (git support not yet implemented)", name)
		}
	}

	color.Green("\nUpdate complete!")
	return nil
}

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st"},
	Short:   "Show status between workspace and repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		color.Green("Running status command...")
		// TODO: Implement status logic
		return fmt.Errorf("not yet implemented")
	},
}

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Display dependency tree with HDL files",
	RunE: func(cmd *cobra.Command, args []string) error {
		color.Green("Running tree command...")
		// TODO: Implement tree logic
		return fmt.Errorf("not yet implemented")
	},
}

var addCmd = &cobra.Command{
	Use:   "add <component_path> <repo_target>",
	Short: "Add component to repository",
	Long: `Add a new component to the repository.
Example: icw add digital/my_module digital
repo_target format: <analog|digital|setup|process>[/category]`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		componentPath := args[0]
		repoTarget := args[1]
		color.Green("Adding component: %s to %s", componentPath, repoTarget)
		// TODO: Implement add logic
		return fmt.Errorf("not yet implemented")
	},
}
