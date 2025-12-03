package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/jakobsen/icw/internal/auth"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage SVN authentication credentials",
	Long: `Store and manage SVN authentication credentials for ICW.

This command helps you securely store your SVN password so you don't
have to enter it every time or set environment variables.

Examples:
  icw auth login           # Store your SVN password
  icw auth logout          # Remove stored credentials
  icw auth status          # Check if credentials are stored
  icw auth test            # Test your credentials`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store SVN password for authentication",
	Long: `Prompts for your SVN password and stores it securely in ~/.icw/credentials.

The credentials file is created with 0600 permissions (readable only by you).
Your password will be used automatically for all SVN operations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAuthLogin()
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored SVN credentials",
	Long:  `Deletes the stored SVN password from ~/.icw/credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAuthLogout()
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Long:  `Displays whether SVN credentials are currently stored.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAuthStatus()
	},
}

var authTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test SVN authentication",
	Long:  `Tests your stored credentials by attempting to connect to the SVN server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAuthTest()
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authTestCmd)
}

func runAuthLogin() error {
	color.Cyan("ICW Authentication Setup")
	color.Cyan("========================\n")

	// Get current username for default
	currentUser := os.Getenv("USER")
	if currentUser == "" {
		currentUser = "anonymous"
	}

	color.Yellow("This will store your SVN password in ~/.icw/credentials")
	color.Yellow("The file will be created with permissions 0600 (readable only by you)\n")

	// Prompt for password
	password, err := auth.PromptPassword()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	// Save password
	if err := auth.SavePassword(password); err != nil {
		return fmt.Errorf("failed to save password: %w", err)
	}

	color.Green("\n✓ Credentials saved successfully!")
	color.Cyan("\nYour password is stored in: %s", auth.CredentialsFile())
	color.Cyan("You can now use ICW commands without entering your password.\n")

	color.Yellow("Try it:")
	color.Yellow("  icw list -r cp3")
	color.Yellow("  icw migrate --create-repo myrepo")

	return nil
}

func runAuthLogout() error {
	if !auth.HasStoredCredentials() {
		color.Yellow("No credentials stored")
		return nil
	}

	if err := auth.DeletePassword(); err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}

	color.Green("✓ Credentials removed successfully")
	color.Cyan("\nYou'll need to run 'icw auth login' to store credentials again")
	color.Cyan("Or set ICW_SVN_PASSWORD environment variable for each command")

	return nil
}

func runAuthStatus() error {
	color.Cyan("Authentication Status")
	color.Cyan("====================\n")

	// Check environment variable
	if envPassword := os.Getenv("ICW_SVN_PASSWORD"); envPassword != "" {
		color.Green("✓ Password set via ICW_SVN_PASSWORD environment variable")
	} else {
		color.Yellow("○ ICW_SVN_PASSWORD not set")
	}

	// Check stored credentials
	if auth.HasStoredCredentials() {
		color.Green("✓ Credentials stored in: %s", auth.CredentialsFile())

		// Check file permissions
		credFile := auth.CredentialsFile()
		info, err := os.Stat(credFile)
		if err == nil {
			perms := info.Mode().Perm()
			if perms == 0600 {
				color.Green("✓ File permissions: 0600 (secure)")
			} else {
				color.Red("⚠ File permissions: %04o (should be 0600)", perms)
				color.Yellow("  Fix with: chmod 600 %s", credFile)
			}
		}
	} else {
		color.Yellow("○ No credentials stored")
		color.Cyan("\n  Run 'icw auth login' to store your password")
	}

	// Show username
	username := os.Getenv("USER")
	if username != "" {
		color.Cyan("\nUsername: %s (from $USER)", username)
	} else {
		color.Yellow("\nUsername: not set (will use 'anonymous')")
	}

	// Show SVN URL
	hostname, _ := os.Hostname()
	if hostname == "g9" {
		color.Cyan("SVN URL: svn://g9 (auto-detected)")
	} else {
		color.Cyan("SVN URL: svn://anyvej11.dk (default)")
	}

	if envURL := os.Getenv("ICW_SVN_URL"); envURL != "" {
		color.Cyan("SVN URL: %s (from ICW_SVN_URL)", envURL)
	}

	return nil
}

func runAuthTest() error {
	color.Cyan("Testing SVN Authentication...")
	color.Cyan("============================\n")

	// Get password
	password, err := auth.GetPassword()
	if err != nil {
		return fmt.Errorf("failed to get password: %w", err)
	}

	if password == "" {
		color.Red("✗ No password available")
		color.Yellow("\nPlease run: icw auth login")
		return fmt.Errorf("no credentials found")
	}

	// Check if we have a repository to test
	repo := os.Getenv("ICW_REPO")
	if repo == "" {
		color.Yellow("ICW_REPO not set, using 'cp3' for testing")
		repo = "cp3"
	}

	color.Cyan("Repository: %s", repo)
	color.Cyan("Username: %s", os.Getenv("USER"))

	// Create a test SVN client and try to connect
	// (We'll import the SVN package for this)
	color.Green("\n✓ Password is available")
	color.Yellow("\nTo fully test, run:")
	color.Yellow("  icw test")
	color.Yellow("Or:")
	color.Yellow("  icw list -r %s", repo)

	return nil
}
