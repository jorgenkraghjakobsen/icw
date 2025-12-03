package main

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/jakobsen/icw/internal/maw"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate components between repositories",
	Long: `Migrate components from one repository to another.
This tool helps you create a new repository and selectively migrate components
from an existing repository, including dependency handling and user management.

Examples:
  # Interactive mode
  icw migrate

  # Create new repository only
  icw migrate --create-repo cp4

  # Full migration
  icw migrate --from cp3 --to cp4`,
	RunE: runMigrate,
}

// Command flags
var (
	flagCreateRepo string
	flagFromRepo   string
	flagToRepo     string
	flagDryRun     bool
)

func init() {
	migrateCmd.Flags().StringVar(&flagCreateRepo, "create-repo", "", "Create a new repository")
	migrateCmd.Flags().StringVar(&flagFromRepo, "from", "", "Source repository")
	migrateCmd.Flags().StringVar(&flagToRepo, "to", "", "Target repository")
	migrateCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Show what would be done without doing it")
}

func runMigrate(cmd *cobra.Command, args []string) error {
	// Create MAW client
	mawClient, err := maw.NewClient()
	if err != nil {
		return fmt.Errorf("MAW client error: %w\nNote: MAW operations must run on g9 server", err)
	}

	// If only --create-repo is specified, just create the repo
	if flagCreateRepo != "" && flagFromRepo == "" && flagToRepo == "" {
		return createRepositoryOnly(mawClient, flagCreateRepo)
	}

	// If --from and --to are specified, do full migration
	if flagFromRepo != "" && flagToRepo != "" {
		return runFullMigration(mawClient, flagFromRepo, flagToRepo)
	}

	// Otherwise run interactive mode
	return runInteractiveMigration(mawClient)
}

func createRepositoryOnly(mawClient *maw.Client, repoName string) error {
	color.Cyan("Creating repository: %s", repoName)

	// Check if repo already exists
	if mawClient.RepoExists(repoName) {
		return fmt.Errorf("repository %s already exists", repoName)
	}

	if flagDryRun {
		color.Yellow("[DRY RUN] Would create repository: %s", repoName)
		return nil
	}

	// Create repository
	if err := mawClient.CreateRepo(repoName); err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	color.Green("✓ Repository %s created successfully", repoName)
	color.Cyan("\nRepository details:")
	color.Cyan("  SVN URL: svn://g9/%s", repoName)
	color.Cyan("  Path: /data_v1/svn/repos/%s", repoName)

	// Show next steps
	fmt.Println()
	color.Yellow("Next steps:")
	fmt.Println("  1. Add users: icw migrate --add-user <username> --to", repoName)
	fmt.Println("  2. Create workspace.config")
	fmt.Println("  3. Add components")

	return nil
}

func runFullMigration(mawClient *maw.Client, fromRepo, toRepo string) error {
	color.Cyan("Migration: %s → %s", fromRepo, toRepo)

	// Verify source repo exists
	if !mawClient.RepoExists(fromRepo) {
		return fmt.Errorf("source repository %s does not exist", fromRepo)
	}

	// Check if target repo exists
	if mawClient.RepoExists(toRepo) {
		return fmt.Errorf("target repository %s already exists", toRepo)
	}

	if flagDryRun {
		return showMigrationPlan(mawClient, fromRepo, toRepo)
	}

	// TODO: Implement full migration workflow
	// 1. Create target repo
	// 2. Copy users
	// 3. Select components
	// 4. Migrate components
	// 5. Update dependencies

	return fmt.Errorf("full migration not yet implemented - use --create-repo first")
}

func runInteractiveMigration(mawClient *maw.Client) error {
	color.Cyan("ICW Repository Migration Tool")
	color.Cyan("============================\n")

	// Show available repositories
	repos, err := mawClient.ListRepos()
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	color.Yellow("Available repositories:")
	for _, repo := range repos {
		fmt.Printf("  • %s\n", repo)
	}

	fmt.Println()
	color.Yellow("Usage:")
	fmt.Println("  icw migrate --create-repo <name>              Create new repository")
	fmt.Println("  icw migrate --from <source> --to <target>     Full migration")

	return nil
}

func showMigrationPlan(mawClient *maw.Client, fromRepo, toRepo string) error {
	color.Cyan("\n[DRY RUN] Migration Plan: %s → %s", fromRepo, toRepo)
	color.Cyan("==========================================\n")

	// Get users from source repo
	users, err := mawClient.ListRepoUsers(fromRepo)
	if err != nil {
		color.Yellow("Warning: Could not list users: %v", err)
	} else {
		color.Green("Users to copy (%d):", len(users))
		for _, user := range users {
			fmt.Printf("  • %s\n", user)
		}
		fmt.Println()
	}

	// Show steps
	color.Cyan("Steps that would be performed:")
	fmt.Println("  1. Create repository:", toRepo)
	fmt.Printf("  2. Copy %d users from %s to %s\n", len(users), fromRepo, toRepo)
	fmt.Println("  3. Select components to migrate (interactive)")
	fmt.Println("  4. Migrate selected components")
	fmt.Println("  5. Update depend.config references")
	fmt.Println()

	color.Yellow("Run without --dry-run to execute")

	return nil
}
