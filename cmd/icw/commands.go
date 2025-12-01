package main

import (
	"fmt"

	"github.com/fatih/color"
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
		color.Green("Running update command...")
		// TODO: Implement update logic
		return fmt.Errorf("not yet implemented")
	},
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
