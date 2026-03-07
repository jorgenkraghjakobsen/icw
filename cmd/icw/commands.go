package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/jakobsen/icw/internal/component"
	"github.com/jakobsen/icw/internal/config"
	"github.com/jakobsen/icw/internal/hdl"
	"github.com/jakobsen/icw/internal/svn"
	"github.com/jakobsen/icw/internal/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(treeCmd)
	rootCmd.AddCommand(hdlCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(releaseCmd)

	// Add flags for list command
	listCmd.Flags().StringP("type", "t", "", "Filter by component type (analog, digital, setup, process)")
	listCmd.Flags().BoolP("branches", "b", false, "Show branches for component")
	listCmd.Flags().BoolP("tags", "g", false, "Show tags for component")
	listCmd.Flags().BoolP("all", "a", false, "Show all details (branches and tags)")
	listCmd.Flags().StringP("repo", "r", "", "Repository to list from (overrides ICW_REPO/workspace.config)")

	// Add flags for release command
	releaseCmd.Flags().StringP("tag", "t", "", "Release tag name (required)")
	releaseCmd.Flags().StringP("message", "m", "", "Commit message (required)")
	releaseCmd.Flags().BoolP("dry-run", "d", false, "Dry run mode (show what would be done)")
	releaseCmd.MarkFlagRequired("tag")
	releaseCmd.MarkFlagRequired("message")
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

	// Parallel update with concurrent work queue
	var mu sync.Mutex          // protects checkedOut, parser state, and output
	checkedOut := make(map[string]bool)
	sem := make(chan struct{}, 8) // limit concurrent SVN operations
	var wg sync.WaitGroup
	var firstErr error           // capture first conflict error

	var processComponent func(comp *component.Component)
	processComponent = func(comp *component.Component) {
		defer wg.Done()

		// Check if already processed or if we hit a conflict
		mu.Lock()
		if checkedOut[comp.Name] || firstErr != nil {
			mu.Unlock()
			return
		}
		checkedOut[comp.Name] = true
		mu.Unlock()

		// Skip local references
		if comp.VCS == "local" {
			mu.Lock()
			color.Blue("  [SKIP] %s (local reference)", comp.Name)
			mu.Unlock()
			return
		}

		destPath := filepath.Join(root, comp.Path)

		if comp.VCS == "svn" {
			// Acquire semaphore to limit concurrent SVN operations
			sem <- struct{}{}

			var svnOutput string
			var svnErr error

			if svn.IsWorkingCopy(destPath) {
				currentBranch, err := svnClient.GetBranch(destPath)
				if err != nil {
					mu.Lock()
					color.Yellow("  [UPDATE] %s (%s) - couldn't detect current branch", comp.Name, comp.Branch)
					mu.Unlock()
					svnOutput, svnErr = svnClient.Update(destPath)
				} else if currentBranch != comp.Branch {
					mu.Lock()
					color.Yellow("  [SWITCH] %s (%s -> %s)", comp.Name, currentBranch, comp.Branch)
					mu.Unlock()
					svnOutput, svnErr = svnClient.Switch(destPath, comp.Path, comp.Branch)
				} else {
					mu.Lock()
					color.Yellow("  [UPDATE] %s (%s)", comp.Name, comp.Branch)
					mu.Unlock()
					svnOutput, svnErr = svnClient.Update(destPath)
				}
			} else {
				mu.Lock()
				color.Green("  [CHECKOUT] %s (%s)", comp.Name, comp.Branch)
				mu.Unlock()
				parentDir := filepath.Dir(destPath)
				if err := os.MkdirAll(parentDir, 0755); err != nil {
					<-sem
					mu.Lock()
					color.Red("    Failed to create directory: %v", err)
					mu.Unlock()
					return
				}
				svnOutput, svnErr = svnClient.Checkout(comp.Path, comp.Branch, destPath)
			}

			<-sem // Release semaphore after SVN operation

			// Print SVN output atomically
			mu.Lock()
			if svnOutput != "" {
				fmt.Print(svnOutput)
			}
			if svnErr != nil {
				color.Red("    Failed: %v", svnErr)
				mu.Unlock()
				return
			}
			mu.Unlock()

			// Parse dependencies (modifies shared parser/workspace state)
			dependConfigPath := filepath.Join(destPath, "depend.config")
			mu.Lock()
			dependencies, err := parser.ParseDependConfig(comp, dependConfigPath)
			if err != nil {
				if strings.Contains(err.Error(), "dependency conflict") || strings.Contains(err.Error(), "branch mismatch") {
					color.Red("    ERROR: %v", err)
					if firstErr == nil {
						firstErr = fmt.Errorf("version conflict detected: %w", err)
					}
					mu.Unlock()
					return
				}
				color.Red("    Warning: Failed to parse dependencies: %v", err)
				mu.Unlock()
				return
			}

			if len(dependencies) > 0 {
				color.Cyan("    Found %d dependencies", len(dependencies))
			}
			mu.Unlock()

			// Launch goroutines for discovered dependencies
			for _, dep := range dependencies {
				wg.Add(1)
				go processComponent(dep)
			}

		} else if comp.VCS == "git" {
			mu.Lock()
			color.Yellow("  [TODO] %s (git support not yet implemented)", comp.Name)
			mu.Unlock()
		}
	}

	// Launch all top-level components in parallel
	for _, comp := range ws.Components {
		wg.Add(1)
		go processComponent(comp)
	}

	wg.Wait()

	if firstErr != nil {
		return firstErr
	}

	// Generate Cadence library files
	generateCadenceLibs(root, ws)

	color.Green("\nUpdate complete!")
	color.Green("Processed %d component(s) total", len(checkedOut))
	return nil
}

// generateCadenceLibs creates local.lib with DEFINE lines for analog/local
// components and copies cds.lib from setup/analog_certus to workspace root.
func generateCadenceLibs(root string, ws *component.Workspace) {
	localLibPath := filepath.Join(root, "local.lib")

	// Remove existing local.lib (fresh generation each time)
	os.Remove(localLibPath)

	// Build local.lib content
	var lines []string
	for _, comp := range ws.Components {
		basename := filepath.Base(comp.Path)
		if comp.VCS == "local" {
			lines = append(lines, fmt.Sprintf("DEFINE %s %s", basename, comp.Path))
		} else if comp.Type == component.TypeAnalog {
			lines = append(lines, fmt.Sprintf("DEFINE %s analog/%s #revision %s from %s", basename, basename, comp.Branch, comp.Path))
		}
	}

	if len(lines) > 0 {
		content := strings.Join(lines, "\n") + "\n"
		if err := os.WriteFile(localLibPath, []byte(content), 0644); err != nil {
			color.Red("Failed to write local.lib: %v", err)
		} else {
			color.Green("Generated local.lib with %d entries", len(lines))
		}
	}

	// Copy files from setup/analog_certus if they exist
	setupFiles := []struct {
		src, dst, label string
	}{
		{"cds.lib", "cds.lib", "cds.lib"},
		{"cdsinit", ".cdsinit", ".cdsinit"},
	}
	for _, f := range setupFiles {
		srcPath := filepath.Join(root, "setup", "analog_certus", f.src)
		dstPath := filepath.Join(root, f.dst)
		if src, err := os.Open(srcPath); err == nil {
			if dst, err := os.Create(dstPath); err == nil {
				if _, err := io.Copy(dst, src); err != nil {
					color.Red("Failed to copy %s: %v", f.label, err)
				} else {
					color.Green("Copied %s from setup/analog_certus/", f.label)
				}
				dst.Close()
			} else {
				color.Red("Failed to create %s: %v", f.label, err)
			}
			src.Close()
		}
	}
}

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st"},
	Short:   "Show status between workspace and repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus()
	},
}

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Display dependency tree from config files",
	Long:  `Display component dependency tree showing workspace structure and component relationships based on workspace.config and depend.config files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTree()
	},
}

var hdlCmd = &cobra.Command{
	Use:   "hdl",
	Short: "Display dependency tree with HDL files",
	Long:  `Display component dependency tree with detailed HDL file listings, categorized by type (package, rtl, behav).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHdl()
	},
}

var addCmd = &cobra.Command{
	Use:   "add <component_path> <repo_target>",
	Short: "Add component to repository",
	Long: `Add a new component to the repository.

The component must exist in your local workspace before adding to SVN.
After adding, the component will be imported to SVN trunk.

Component Path:
  Path to the component relative to workspace root (e.g., "digital/my_module")
  or just the component name if you're in the correct directory (e.g., "fpga_template")

Repository Target:
  - Simple type: digital, analog, setup, process
  - Type with category: digital/muxes, analog/opamps, etc.

  The category helps organize related components together.

Examples:
  icw add digital/fpga_template digital           # Add to digital type
  icw add fpga_template digital                   # Same, if in digital/ dir
  icw add digital/mux8to1 digital/muxes          # Add to digital/muxes category
  icw add analog/opamp analog/amplifiers         # Add to analog/amplifiers category`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAdd(args[0], args[1])
	},
}

var releaseCmd = &cobra.Command{
	Use:   "release -t <tag> -m <message>",
	Short: "Create a tagged release of a component",
	Long: `Create a tagged release of the current component and all its dependencies.

This command must be run from within a component directory. It will:
  1. Recursively release all dependencies first
  2. Check if the release tag already exists (skip if yes)
  3. Create an SVN tag from the current branch
  4. Update depend.config in the tag to point to released dependencies

The release process ensures all components in the dependency tree use
the same release tag, creating a consistent, reproducible release.

Flags:
  -t, --tag string       Release tag name (e.g., v1.0.0)
  -m, --message string   Commit message
  -d, --dry-run          Show what would be done without making changes

Examples:
  icw release -t v1.0.0 -m "First stable release"
  icw release -t v1.0.1 -m "Bug fix release"
  icw release -t v2.0.0-beta -m "Beta release" -d   # Dry run first`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tagName, _ := cmd.Flags().GetString("tag")
		message, _ := cmd.Flags().GetString("message")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		return runRelease(tagName, message, dryRun)
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
  icw list -r cp3                # List components from cp3 repository
  icw list -r cp4 -t analog      # List analog components from cp4
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
	repoFlag, _ := cmd.Flags().GetString("repo")

	// Determine which repository to use
	var configRepo, configURL string

	if repoFlag != "" {
		// Use repository specified via --repo flag
		configRepo = repoFlag
	} else {
		// Try to read workspace.config for repo configuration
		root, err := config.FindWorkspaceRoot()
		if err == nil {
			ws := component.NewWorkspace(root)
			parser := config.NewParser(ws)
			if err := parser.ParseWorkspaceConfig(ws.Config); err == nil {
				configRepo = parser.Repo
				configURL = parser.SvnURL
			}
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

// treeNode holds a resolved component and its children for printing
type treeNode struct {
	comp     *component.Component
	children []*treeNode
}

// dependCache provides thread-safe caching of depend.config content
type dependCache struct {
	mu    sync.Mutex
	cache map[string]string // key: "path:branch" -> depend.config content ("" = no deps)
}

func newDependCache() *dependCache {
	return &dependCache{cache: make(map[string]string)}
}

// get returns cached content and whether it was found
func (dc *dependCache) get(path, branch string) (string, bool) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	v, ok := dc.cache[path+":"+branch]
	return v, ok
}

// set stores content in the cache
func (dc *dependCache) set(path, branch, content string) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.cache[path+":"+branch] = content
}

// fetchDependConfig reads depend.config from local checkout or SVN, using cache
func fetchDependConfig(comp *component.Component, workspaceRoot string, svnClient *svn.Client, cache *dependCache) string {
	if comp.VCS == "local" {
		return ""
	}

	if content, ok := cache.get(comp.Path, comp.Branch); ok {
		return content
	}

	var content string

	// Check local first
	dependConfigPath := filepath.Join(workspaceRoot, comp.Path, "depend.config")
	if _, err := os.Stat(dependConfigPath); err == nil {
		data, err := os.ReadFile(dependConfigPath)
		if err == nil {
			content = string(data)
		}
	} else {
		result, err := svnClient.Cat(comp.Path, comp.Branch, "depend.config")
		if err == nil {
			content = result
		}
	}

	cache.set(comp.Path, comp.Branch, content)
	return content
}

// resolveTree builds the full tree structure, fetching dependencies in parallel at each level
func resolveTree(comps []*component.Component, workspaceRoot string, svnClient *svn.Client, cache *dependCache) []*treeNode {
	nodes := make([]*treeNode, len(comps))

	// Fetch all depend.configs at this level in parallel
	type fetchResult struct {
		idx  int
		deps []*component.Component
	}
	results := make([]fetchResult, len(comps))

	var wg sync.WaitGroup
	for i, comp := range comps {
		nodes[i] = &treeNode{comp: comp}
		wg.Add(1)
		go func(idx int, c *component.Component) {
			defer wg.Done()
			content := fetchDependConfig(c, workspaceRoot, svnClient, cache)
			var deps []*component.Component
			if content != "" {
				deps = parseDependConfigContentForTree(content)
			}
			results[idx] = fetchResult{idx: idx, deps: deps}
		}(i, comp)
	}
	wg.Wait()

	// Recursively resolve children
	for _, r := range results {
		if len(r.deps) > 0 {
			nodes[r.idx].children = resolveTree(r.deps, workspaceRoot, svnClient, cache)
		}
	}

	return nodes
}

// printTree prints the resolved tree
func printTree(nodes []*treeNode, indent int) {
	for _, n := range nodes {
		indentStr := strings.Repeat(" ", indent)
		nameCol := fmt.Sprintf("%s%-*s", indentStr, 45-indent, n.comp.Name)
		branchCol := fmt.Sprintf("%-20s", n.comp.Branch)
		fmt.Printf("%s  %s  [%s]\n", nameCol, branchCol, n.comp.Type)
		if len(n.children) > 0 {
			printTree(n.children, indent+2)
		}
	}
}

func runTree() error {
	// Find workspace root
	root, err := config.FindWorkspaceRoot()
	if err != nil {
		return fmt.Errorf("not in a workspace: %w", err)
	}

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

	// Create SVN client for fetching depend.config from repository
	svnClient, err := svn.NewClientWithConfig(parser.Repo, parser.SvnURL)
	if err != nil {
		return fmt.Errorf("failed to create SVN client: %w", err)
	}

	// Collect top-level components
	topLevel := make([]*component.Component, 0, len(ws.Components))
	for _, comp := range ws.Components {
		topLevel = append(topLevel, comp)
	}

	// Resolve full tree (parallel fetches + caching)
	cache := newDependCache()
	nodes := resolveTree(topLevel, root, svnClient, cache)

	// Print dependency tree
	color.Cyan("Dependency tree for workspace\n")
	printTree(nodes, 0)

	return nil
}

func runHdl() error {
	// Find workspace root
	root, err := config.FindWorkspaceRoot()
	if err != nil {
		return fmt.Errorf("not in a workspace: %w", err)
	}

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

	// Load all dependencies
	for _, comp := range ws.Components {
		if comp.VCS == "local" {
			continue
		}
		destPath := filepath.Join(root, comp.Path)
		dependConfigPath := filepath.Join(destPath, "depend.config")
		parser.ParseDependConfig(comp, dependConfigPath)
	}

	// Print dependency tree with HDL files
	color.Cyan("Dependency tree with HDL files\n")

	// Track which components we've already printed to avoid duplicates
	printed := make(map[string]bool)

	// Print each top-level component and its dependencies
	for _, comp := range ws.Components {
		printComponentTreeWithHDL(comp, root, 0, printed)
	}

	return nil
}


func parseDependConfigContentForTree(content string) []*component.Component {
	var dependencies []*component.Component

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse component line - use simple parsing
		comp := parseComponentLineSimple(line)
		if comp != nil {
			dependencies = append(dependencies, comp)
		}
	}

	return dependencies
}

func parseComponentLineSimple(line string) *component.Component {
	// Match: use component("path", "type", "branch")
	re := regexp.MustCompile(`use\s+component\s*\(\s*"([^"]+)"\s*,\s*"([^"]+)"\s*,\s*"([^"]+)"\s*\)`)
	if matches := re.FindStringSubmatch(line); matches != nil {
		return &component.Component{
			Name:   matches[1],
			Path:   matches[1],
			Type:   component.ComponentType(matches[2]),
			Branch: matches[3],
			VCS:    "svn",
		}
	}

	// Match: use component("path", "type")
	re = regexp.MustCompile(`use\s+component\s*\(\s*"([^"]+)"\s*,\s*"([^"]+)"\s*\)`)
	if matches := re.FindStringSubmatch(line); matches != nil {
		return &component.Component{
			Name:   matches[1],
			Path:   matches[1],
			Type:   component.ComponentType(matches[2]),
			Branch: "trunk",
			VCS:    "svn",
		}
	}

	// Match: use ref("path")
	re = regexp.MustCompile(`use\s+ref\s*\(\s*"([^"]+)"\s*\)`)
	if matches := re.FindStringSubmatch(line); matches != nil {
		return &component.Component{
			Name:   matches[1],
			Path:   matches[1],
			Type:   component.TypeDigital, // Default
			Branch: "local",
			VCS:    "local",
		}
	}

	return nil
}

func printComponentTreeWithHDL(comp *component.Component, workspaceRoot string, indent int, printed map[string]bool) {
	// Skip if already printed
	if printed[comp.Name] {
		return
	}
	printed[comp.Name] = true

	// Print component info
	indentStr := strings.Repeat(" ", indent)
	fmt.Printf("%s%s (%s), %s\n", indentStr, comp.Name, comp.Path, comp.Branch)

	// Only discover HDL files for digital components
	if comp.Type == component.TypeDigital && comp.VCS != "local" {
		componentPath := filepath.Join(workspaceRoot, comp.Path)
		hdlFiles, err := hdl.DiscoverFiles(componentPath)
		if err == nil {
			// Print package files
			if len(hdlFiles.Package) > 0 {
				fileList := shortenPaths(hdlFiles.Package, workspaceRoot)
				fmt.Printf("%s  - package: %s\n", indentStr, strings.Join(fileList, " "))
			}

			// Print RTL files
			if len(hdlFiles.RTL) > 0 {
				fileList := shortenPaths(hdlFiles.RTL, workspaceRoot)
				fmt.Printf("%s  - rtl: %s\n", indentStr, strings.Join(fileList, " "))
			}

			// Print behavioral files
			if len(hdlFiles.Behav) > 0 {
				fileList := shortenPaths(hdlFiles.Behav, workspaceRoot)
				fmt.Printf("%s  - behav: %s\n", indentStr, strings.Join(fileList, " "))
			}
		}
	}

	// Recursively print dependencies
	for _, dep := range comp.Dependencies {
		printComponentTreeWithHDL(dep, workspaceRoot, indent+2, printed)
	}
}

func shortenPaths(paths []string, workspaceRoot string) []string {
	shortened := make([]string, len(paths))
	for i, path := range paths {
		// Replace workspace root with "..."
		shortened[i] = strings.Replace(path, workspaceRoot, "...", 1)
	}
	return shortened
}

// copyDir recursively copies directory contents from src to dst
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Skip .svn directories
		if info.IsDir() && info.Name() == ".svn" {
			return filepath.SkipDir
		}

		// Construct destination path
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return err
		}

		// Preserve permissions
		return os.Chmod(dstPath, info.Mode())
	})
}

func runStatus() error {
	// Find workspace root
	root, err := config.FindWorkspaceRoot()
	if err != nil {
		return fmt.Errorf("not in a workspace: %w", err)
	}

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

	// Create SVN client
	svnClient, err := svn.NewClientWithConfig(parser.Repo, parser.SvnURL)
	if err != nil {
		return fmt.Errorf("failed to create SVN client: %w", err)
	}

	color.Cyan("Workspace status:\n")

	// Check each component
	hasChanges := false
	for _, comp := range ws.Components {
		if comp.VCS == "local" {
			continue
		}

		destPath := filepath.Join(root, comp.Path)

		// Check if component is checked out
		if !svn.IsWorkingCopy(destPath) {
			color.Yellow("[NOT CHECKED OUT] %s", comp.Name)
			hasChanges = true
			continue
		}

		// Get SVN status
		status, err := svnClient.Status(destPath)
		if err != nil {
			color.Red("[ERROR] %s: %v", comp.Name, err)
			continue
		}

		// Check if there are any changes
		if strings.TrimSpace(status) != "" {
			color.Yellow("[MODIFIED] %s (%s)", comp.Name, comp.Branch)
			// Print the status with indentation
			lines := strings.Split(strings.TrimSpace(status), "\n")
			for _, line := range lines {
				fmt.Printf("  %s\n", line)
			}
			hasChanges = true
		} else {
			color.Green("[CLEAN] %s (%s)", comp.Name, comp.Branch)
		}
	}

	if !hasChanges {
		fmt.Println()
		color.Green("Workspace is clean - no changes detected")
	}

	return nil
}

func runAdd(componentPath, repoTarget string) error {
	// Find workspace root
	root, err := config.FindWorkspaceRoot()
	if err != nil {
		return fmt.Errorf("not in a workspace: %w", err)
	}

	color.Cyan("=== Adding Component to Repository ===\n")

	// Parse workspace.config to get repository info
	ws := component.NewWorkspace(root)
	parser := config.NewParser(ws)
	if err := parser.ParseWorkspaceConfig(ws.Config); err != nil {
		return fmt.Errorf("failed to parse workspace.config: %w", err)
	}

	// Create SVN client
	svnClient, err := svn.NewClientWithConfig(parser.Repo, parser.SvnURL)
	if err != nil {
		return fmt.Errorf("failed to create SVN client: %w", err)
	}

	// Parse repo_target to extract type and category
	// Format: <type> or <type/category>
	// Examples: "digital", "digital/muxes", "analog/opamps"
	parts := strings.Split(repoTarget, "/")
	componentType := parts[0]
	var category string
	if len(parts) > 1 {
		category = strings.Join(parts[1:], "/")
	}

	// Validate component type
	validTypes := map[string]bool{
		"digital": true,
		"analog":  true,
		"setup":   true,
		"process": true,
	}
	if !validTypes[componentType] {
		return fmt.Errorf("invalid component type '%s', must be one of: digital, analog, setup, process", componentType)
	}

	// Resolve component path
	// If componentPath doesn't contain a type prefix, try to infer it from current directory
	var fullComponentPath string
	var componentName string

	// Handle "." as current directory
	if componentPath == "." {
		cwd, _ := os.Getwd()
		relPath, err := filepath.Rel(root, cwd)
		if err != nil || relPath == "." {
			return fmt.Errorf("cannot use '.' from workspace root, please specify component path")
		}
		fullComponentPath = relPath
		componentName = filepath.Base(cwd)
	} else if strings.Contains(componentPath, "/") {
		// Full path provided (e.g., "digital/fpga_template")
		fullComponentPath = componentPath
		pathParts := strings.Split(componentPath, "/")
		componentName = pathParts[len(pathParts)-1]
	} else {
		// Just component name provided (e.g., "fpga_template")
		// Try to infer from current directory
		cwd, _ := os.Getwd()
		relPath, err := filepath.Rel(root, cwd)
		if err == nil && relPath != "." {
			// We're in a subdirectory of the workspace
			fullComponentPath = filepath.Join(relPath, componentPath)
		} else {
			// Assume it's in the component type directory
			fullComponentPath = filepath.Join(componentType, componentPath)
		}
		componentName = componentPath
	}

	// Check if component exists locally
	localPath := filepath.Join(root, fullComponentPath)
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("component not found at: %s", localPath)
	}

	color.Green("Component path: %s", fullComponentPath)
	color.Green("Local path: %s", localPath)
	color.Green("Repository: %s", svnClient.Repo)
	color.Green("Type: %s", componentType)
	if category != "" {
		color.Green("Category: %s", category)
	}
	fmt.Println()

	// Construct SVN path
	var svnComponentPath string
	if category != "" {
		svnComponentPath = fmt.Sprintf("%s/%s/%s", componentType, category, componentName)
	} else {
		svnComponentPath = fmt.Sprintf("%s/%s", componentType, componentName)
	}

	// Check if local component is already under SVN control
	if svn.IsWorkingCopy(localPath) {
		return fmt.Errorf("component is already under SVN control: %s", localPath)
	}

	color.Yellow("Step 1: Creating directory structure in SVN...")
	color.Cyan("  SVN path: %s", svnComponentPath)

	// Create directory structure (trunk, tags, branches) in SVN
	trunkURL := fmt.Sprintf("%s/%s/components/%s/trunk", svnClient.URL, svnClient.Repo, svnComponentPath)
	if err := svnClient.CreateComponentDirs(svnComponentPath); err != nil {
		return fmt.Errorf("failed to create SVN directories: %w", err)
	}
	color.Green("  ✓ Directory structure created\n")

	// Checkout trunk to temp directory, then move .svn into existing directory
	// This avoids renaming the user's directory (which breaks shell cwd)
	color.Yellow("Step 2: Creating SVN working copy...")
	tmpPath := localPath + ".svn-tmp"
	if output, err := svnClient.Checkout(svnComponentPath, "trunk", tmpPath); err != nil {
		fmt.Print(output)
		os.RemoveAll(tmpPath)
		return fmt.Errorf("failed to checkout trunk: %w", err)
	} else if output != "" {
		fmt.Print(output)
	}

	// Move .svn from temp to component directory
	svnDir := filepath.Join(tmpPath, ".svn")
	targetSvnDir := filepath.Join(localPath, ".svn")
	if err := os.Rename(svnDir, targetSvnDir); err != nil {
		os.RemoveAll(tmpPath)
		return fmt.Errorf("failed to initialize working copy: %w", err)
	}
	os.RemoveAll(tmpPath)
	color.Green("  ✓ Working copy created\n")

	// Add all files to SVN
	color.Yellow("Step 3: Adding files to SVN...")
	if err := svnClient.AddAll(localPath); err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}
	color.Green("  ✓ Files added\n")

	// Commit files
	color.Yellow("Step 4: Committing files to repository...")
	commitMsg := fmt.Sprintf("Initial import of %s", componentName)
	if err := svnClient.Commit(localPath, commitMsg); err != nil {
		return fmt.Errorf("failed to commit files: %w", err)
	}
	color.Green("  ✓ Files committed\n")

	color.Green("=== Component Added Successfully! ===\n")
	color.Cyan("Component '%s' has been added to repository '%s'", componentName, svnClient.Repo)
	color.Cyan("SVN path: %s", svnComponentPath)
	color.Cyan("Trunk URL: %s", trunkURL)
	color.Cyan("Local path is now an SVN working copy: %s", localPath)
	fmt.Println()
	color.Yellow("You can now:")
	color.Yellow("  - Make changes and commit: icw commit")
	color.Yellow("  - Update from repository: icw update")
	color.Yellow("  - Add to workspace.config: use component(\"%s\", \"%s\", \"trunk\")", fullComponentPath, componentType)

	return nil
}

func runRelease(tagName, message string, dryRun bool) error {
	// Find workspace root
	root, err := config.FindWorkspaceRoot()
	if err != nil {
		return fmt.Errorf("not in a workspace: %w", err)
	}

	// Parse workspace.config to build component tree
	ws := component.NewWorkspace(root)
	parser := config.NewParser(ws)
	if err := parser.ParseWorkspaceConfig(ws.Config); err != nil {
		return fmt.Errorf("failed to parse workspace.config: %w", err)
	}

	// Load dependencies for all components
	for _, comp := range ws.Components {
		if comp.VCS == "local" {
			continue
		}
		destPath := filepath.Join(root, comp.Path)
		dependConfigPath := filepath.Join(destPath, "depend.config")

		// Parse dependencies (ignoring errors if file doesn't exist)
		_, err := parser.ParseDependConfig(comp, dependConfigPath)
		if err != nil {
			// Check if it's a conflict error
			if strings.Contains(err.Error(), "dependency conflict") || strings.Contains(err.Error(), "branch mismatch") {
				return fmt.Errorf("version conflict detected: %w", err)
			}
			// Ignore other errors (e.g., file not found)
		}
	}

	// Create SVN client
	svnClient, err := svn.NewClientWithConfig(parser.Repo, parser.SvnURL)
	if err != nil {
		return fmt.Errorf("failed to create SVN client: %w", err)
	}

	// Detect current component from directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Extract component name from current directory
	componentName := filepath.Base(cwd)

	// Find the component in the workspace
	var currentComp *component.Component
	for _, comp := range ws.Components {
		// Match by full path or exact name match
		compBaseName := filepath.Base(comp.Path)
		if comp.Name == componentName || compBaseName == componentName || comp.Path == componentName {
			currentComp = comp
			break
		}
	}

	if currentComp == nil {
		return fmt.Errorf("current directory '%s' is not a known component in workspace.config", componentName)
	}

	if dryRun {
		color.Yellow("=== DRY RUN MODE ===")
		color.Yellow("Showing what would be done without making changes\n")
	} else {
		color.Cyan("=== Releasing Component ===\n")
	}

	// Release component and all dependencies
	if err := releaseComponent(svnClient, currentComp, tagName, message, "", dryRun, root, parser); err != nil {
		return err
	}

	if dryRun {
		fmt.Println()
		color.Green("=== Dry Run Complete ===")
		color.Cyan("Run without -d flag to actually create the release")
	} else {
		fmt.Println()
		color.Green("=== Release Complete: %s ===", tagName)
	}

	return nil
}

func releaseComponent(
	svnClient *svn.Client,
	comp *component.Component,
	tagName string,
	message string,
	indent string,
	dryRun bool,
	workspaceRoot string,
	parser *config.Parser,
) error {
	// 1. Release dependencies first (recursive, depth-first)
	if len(comp.Dependencies) > 0 {
		for _, dep := range comp.Dependencies {
			if err := releaseComponent(svnClient, dep, tagName, message, indent+"  ", dryRun, workspaceRoot, parser); err != nil {
				return err
			}
		}
	}

	// 2. Check if component is already released
	exists, err := svnClient.TagExists(comp.Path, tagName)
	if err != nil {
		return fmt.Errorf("failed to check if tag exists: %w", err)
	}

	if exists {
		color.Yellow("%s%s %s already exists", indent, comp.Name, tagName)
		return nil
	}

	// 3. Create release tag
	if dryRun {
		color.Cyan("%s[DRY RUN] Would create tag: %s/tags/%s from %s", indent, comp.Path, tagName, comp.Branch)
	} else {
		color.Green("%sReleasing %s (%s) -> tags/%s", indent, comp.Name, comp.Path, tagName)
		if err := svnClient.CreateTag(comp.Path, comp.Branch, tagName, message); err != nil {
			return fmt.Errorf("failed to create tag for %s: %w", comp.Name, err)
		}
		color.Green("%s  ✓ Created tag", indent)
	}

	// 4. Update depend.config in release tag if component has dependencies
	if len(comp.Dependencies) > 0 {
		if dryRun {
			color.Cyan("%s[DRY RUN] Would update depend.config in tag", indent)
			for _, dep := range comp.Dependencies {
				color.Cyan("%s    - %s -> tags/%s", indent, dep.Name, tagName)
			}
		} else {
			// Generate new depend.config pointing to released dependencies
			// Use just the component name (without path) for temp file
			compBaseName := filepath.Base(comp.Name)
			tempFile := filepath.Join(workspaceRoot, fmt.Sprintf("depend.config-%s-%s", compBaseName, tagName))
			f, err := os.Create(tempFile)
			if err != nil {
				return fmt.Errorf("failed to create temp depend.config: %w", err)
			}

			for _, dep := range comp.Dependencies {
				_, err := fmt.Fprintf(f, "use component(\"%s\", \"%s\", \"tags/%s\");\n", dep.Path, dep.Type, tagName)
				if err != nil {
					f.Close()
					return fmt.Errorf("failed to write depend.config: %w", err)
				}
			}
			f.Close()

			// Delete old depend.config from tag
			dependConfigURL := fmt.Sprintf("%s/%s/components/%s/tags/%s/depend.config",
				svnClient.URL, svnClient.Repo, comp.Path, tagName)

			if err := svnClient.DeleteFile(dependConfigURL, message+" (update depend.config 1/2)"); err != nil {
				// Ignore error if file doesn't exist
				if !strings.Contains(err.Error(), "non-existent") {
					color.Yellow("%s  Warning: couldn't delete old depend.config: %v", indent, err)
				}
			}

			// Import new depend.config
			if err := svnClient.ImportFile(tempFile, dependConfigURL, message+" (update depend.config 2/2)"); err != nil {
				return fmt.Errorf("failed to import depend.config: %w", err)
			}

			// Clean up temp file
			os.Remove(tempFile)

			color.Green("%s  ✓ Updated depend.config", indent)
		}
	}

	return nil
}
