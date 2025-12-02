package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/jakobsen/icw/internal/component"
	"github.com/jakobsen/icw/internal/config"
	"github.com/jakobsen/icw/internal/svn"
	"github.com/jakobsen/icw/internal/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(treeCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(listCmd)

	// Add flags for list command
	listCmd.Flags().StringP("type", "t", "", "Filter by component type (analog, digital, setup, process)")
	listCmd.Flags().BoolP("branches", "b", false, "Show branches for component")
	listCmd.Flags().BoolP("tags", "g", false, "Show tags for component")
	listCmd.Flags().BoolP("all", "a", false, "Show all details (branches and tags)")
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display detailed version information including build date and commit.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Info())
	},
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
		response = strings.TrimSpace(strings.ToLower(response))

		// Default to "yes" if empty (user pressed Enter) or explicit yes
		if response == "" || response == "y" || response == "yes" {
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

	// Create SVN client with config from workspace.config
	svnClient, err := svn.NewClientWithConfig(parser.Repo, parser.SvnURL)
	if err != nil {
		return fmt.Errorf("failed to create SVN client: %w", err)
	}

	color.Cyan("Using repository: %s", svnClient.Repo)
	if parser.Repo != "" {
		color.Cyan("  (from workspace.config)")
	}

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

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test SVN server and repository configuration",
	Long: `Verify connectivity to SVN server and repository access.

Environment Variables:
  ICW_REPO     Repository name (required)
  ICW_SVN_URL  SVN server URL (default: svn://anyvej11.dk)
  USER         Username for SVN authentication

Examples:
  export ICW_REPO=icworks
  icw test

  export ICW_REPO=myrepo
  export ICW_SVN_URL=svn://myserver.com
  icw test`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTest()
	},
}

func runTest() error {
	color.Cyan("=== ICW Configuration Test ===\n")

	// Try to read workspace.config if available
	var configRepo, configURL string
	root, err := config.FindWorkspaceRoot()
	if err == nil {
		color.Yellow("Found workspace.config at: %s", root)
		ws := component.NewWorkspace(root)
		parser := config.NewParser(ws)
		if err := parser.ParseWorkspaceConfig(ws.Config); err == nil {
			configRepo = parser.Repo
			configURL = parser.SvnURL
			if configRepo != "" {
				color.Green("  ✓ Repository from config: %s", configRepo)
			}
			if configURL != "" {
				color.Green("  ✓ SVN URL from config: %s", configURL)
			}
		}
		fmt.Println()
	}

	// Check environment variables
	color.Yellow("Checking environment variables...")
	envRepo := os.Getenv("ICW_REPO")
	envURL := os.Getenv("ICW_SVN_URL")
	user := os.Getenv("USER")

	// Determine effective repo (env var overrides config)
	repo := envRepo
	if repo == "" {
		repo = configRepo
	}

	if repo == "" {
		color.Red("  ✗ ICW_REPO: not set")
		fmt.Println("\nPlease set ICW_REPO environment variable:")
		fmt.Println("  export ICW_REPO=your_repo_name")
		fmt.Println("Or add to workspace.config:")
		fmt.Println("  set repo \"your_repo_name\"")
		return fmt.Errorf("ICW_REPO not set")
	}

	if envRepo != "" {
		color.Green("  ✓ ICW_REPO: %s (from environment)", repo)
	} else {
		color.Green("  ✓ ICW_REPO: %s (from workspace.config)", repo)
	}

	// Determine effective SVN URL (env var overrides config)
	svnURL := envURL
	if svnURL == "" {
		svnURL = configURL
	}

	if svnURL == "" {
		color.Yellow("  ○ ICW_SVN_URL: using default (svn://anyvej11.dk)")
	} else {
		if envURL != "" {
			color.Green("  ✓ ICW_SVN_URL: %s (from environment)", svnURL)
		} else {
			color.Green("  ✓ ICW_SVN_URL: %s (from workspace.config)", svnURL)
		}
	}

	if user == "" {
		color.Yellow("  ○ USER: using default (anonymous)")
	} else {
		color.Green("  ✓ USER: %s", user)
	}

	// Create SVN client
	fmt.Println()
	color.Yellow("Creating SVN client...")
	svnClient, err := svn.NewClientWithConfig(repo, svnURL)
	if err != nil {
		color.Red("  ✗ Failed: %v", err)
		return err
	}
	color.Green("  ✓ SVN URL: %s", svnClient.URL)
	color.Green("  ✓ Repository: %s", svnClient.Repo)
	color.Green("  ✓ Username: %s", svnClient.Username)

	// Test connection
	fmt.Println()
	color.Yellow("Testing SVN server connection...")
	if err := svnClient.TestConnection(); err != nil {
		color.Red("  ✗ Connection failed: %v", err)
		return err
	}
	color.Green("  ✓ Connection successful!")

	// List components
	fmt.Println()
	color.Yellow("Listing available components...")
	components, err := svnClient.ListComponents()
	if err != nil {
		color.Red("  ✗ Failed to list components: %v", err)
		return err
	}

	if len(components) == 0 {
		color.Yellow("  ○ No components found in repository")
	} else {
		color.Green("  ✓ Found %d components:", len(components))
		for _, comp := range components {
			fmt.Printf("    - %s\n", comp)
		}
	}

	fmt.Println()
	color.Green("=== All tests passed! ===")
	return nil
}

var listCmd = &cobra.Command{
	Use:   "list [component]",
	Aliases: []string{"ls"},
	Short: "List components and their branches/tags",
	Long: `List components in the repository with optional filtering and details.

Examples:
  icw list                       # List all components
  icw list -t digital            # List only digital components
  icw list digital/my_module     # Show details for specific component
  icw list digital/my_module -b  # Show branches only
  icw list digital/my_module -g  # Show tags only
  icw list digital/my_module -a  # Show all details (branches and tags)
  icw list digital/dig*          # Show all components matching pattern
  icw list "digital/*cp3"        # Pattern with quotes (shell glob protection)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList(cmd, args)
	},
}

func runList(cmd *cobra.Command, args []string) error {
	// Get flags
	componentType, _ := cmd.Flags().GetString("type")
	showBranches, _ := cmd.Flags().GetBool("branches")
	showTags, _ := cmd.Flags().GetBool("tags")
	showAll, _ := cmd.Flags().GetBool("all")

	// Try to read workspace.config for repo configuration
	var configRepo, configURL string
	root, err := config.FindWorkspaceRoot()
	if err == nil {
		ws := component.NewWorkspace(root)
		parser := config.NewParser(ws)
		if err := parser.ParseWorkspaceConfig(ws.Config); err == nil {
			configRepo = parser.Repo
			configURL = parser.SvnURL
		}
	}

	// Create SVN client
	svnClient, err := svn.NewClientWithConfig(configRepo, configURL)
	if err != nil {
		return fmt.Errorf("failed to create SVN client: %w", err)
	}

	// If a specific component is requested as positional argument
	if len(args) > 0 {
		componentPath := args[0]
		// Check if it contains a glob pattern
		if strings.Contains(componentPath, "*") {
			return showMatchingComponents(svnClient, componentPath, showBranches, showTags, showAll)
		}
		return showComponentDetails(svnClient, componentPath, showBranches, showTags, showAll)
	}

	// Check if -t contains a full component path (contains /)
	// If so, show details instead of listing
	if componentType != "" && strings.Contains(componentType, "/") {
		// Check if it contains a glob pattern
		if strings.Contains(componentType, "*") {
			return showMatchingComponents(svnClient, componentType, showBranches, showTags, showAll)
		}
		return showComponentDetails(svnClient, componentType, showBranches, showTags, showAll)
	}

	// List components
	color.Cyan("Repository: %s", svnClient.Repo)
	fmt.Println()

	if componentType != "" {
		// List components of specific type
		color.Yellow("Listing %s components...", componentType)
		components, err := svnClient.ListComponentsByType(componentType)
		if err != nil {
			return fmt.Errorf("failed to list components: %w", err)
		}

		if len(components) == 0 {
			color.Yellow("No %s components found", componentType)
			return nil
		}

		for _, comp := range components {
			fmt.Printf("  %s\n", comp)
		}
		fmt.Println()
		color.Green("Total: %d components", len(components))
	} else {
		// List all components by type
		types := []string{"analog", "digital", "setup", "process"}
		totalCount := 0

		for _, typ := range types {
			components, err := svnClient.ListComponentsByType(typ)
			if err != nil {
				color.Yellow("  [%s] Could not list: %v", typ, err)
				continue
			}

			if len(components) > 0 {
				color.Cyan("[%s]", strings.ToUpper(typ))
				for _, comp := range components {
					fmt.Printf("  %s\n", comp)
				}
				fmt.Println()
				totalCount += len(components)
			}
		}

		color.Green("Total: %d components", totalCount)
	}

	return nil
}

func showComponentDetails(svnClient *svn.Client, componentPath string, showBranches, showTags, showAll bool) error {
	color.Cyan("Component: %s", componentPath)
	fmt.Println()

	// Get component info
	info, err := svnClient.GetComponentInfo(componentPath)
	if err != nil {
		return fmt.Errorf("failed to get component info: %w", err)
	}

	// Show trunk
	if info.HasTrunk {
		color.Green("✓ trunk")
	} else {
		color.Yellow("✗ trunk (not found)")
	}
	fmt.Println()

	// Determine what to show
	displayBranches := showBranches || showAll || (!showBranches && !showTags && !showAll)
	displayTags := showTags || showAll || (!showBranches && !showTags && !showAll)

	// Show branches
	if displayBranches {
		if len(info.Branches) > 0 {
			color.Cyan("Branches (%d):", len(info.Branches))
			for _, branch := range info.Branches {
				fmt.Printf("  %s\n", branch)
			}
		} else {
			color.Yellow("Branches: none")
		}
		fmt.Println()
	}

	// Show tags
	if displayTags {
		if len(info.Tags) > 0 {
			color.Cyan("Tags (%d):", len(info.Tags))
			for _, tag := range info.Tags {
				fmt.Printf("  %s\n", tag)
			}
		} else {
			color.Yellow("Tags: none")
		}
	}

	return nil
}

func showMatchingComponents(svnClient *svn.Client, pattern string, showBranches, showTags, showAll bool) error {
	color.Cyan("Pattern: %s", pattern)
	fmt.Println()

	// Find matching components
	matches, err := svnClient.FindComponentsByPattern(pattern)
	if err != nil {
		return fmt.Errorf("failed to find matching components: %w", err)
	}

	if len(matches) == 0 {
		color.Yellow("No components match pattern: %s", pattern)
		return nil
	}

	color.Green("Found %d matching component(s):", len(matches))
	fmt.Println()

	// If no flags are set, just list the component names
	if !showBranches && !showTags && !showAll {
		for _, comp := range matches {
			fmt.Printf("  %s\n", comp)
		}
		return nil
	}

	// Show details for each matching component
	for i, comp := range matches {
		if i > 0 {
			fmt.Println("---")
			fmt.Println()
		}
		if err := showComponentDetails(svnClient, comp, showBranches, showTags, showAll); err != nil {
			color.Red("Error getting details for %s: %v", comp, err)
		}
	}

	return nil
}
