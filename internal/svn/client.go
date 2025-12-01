package svn

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Client represents an SVN client
type Client struct {
	URL      string // Base SVN URL (e.g., svn://anyvej11.dk)
	Repo     string // Repository name (from ICW_REPO env var)
	Username string // SVN username
}

// NewClient creates a new SVN client
func NewClient() (*Client, error) {
	repo := os.Getenv("ICW_REPO")
	if repo == "" {
		repo = "icworks_public" // Default repo
	}

	username := os.Getenv("USER")
	if username == "" {
		username = "anonymous"
	}

	return &Client{
		URL:      "svn://anyvej11.dk",
		Repo:     repo,
		Username: username,
	}, nil
}

// Checkout checks out a component from SVN
func (c *Client) Checkout(componentPath, branch, destPath string) error {
	// Build SVN URL: svn://anyvej11.dk/repo/components/path/branch
	svnURL := fmt.Sprintf("%s/%s/components/%s/%s", c.URL, c.Repo, componentPath, branch)

	// Run svn checkout
	cmd := exec.Command("svn", "checkout", svnURL, destPath, "--username", c.Username)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("svn checkout failed: %w", err)
	}

	return nil
}

// Update updates an existing SVN working copy
func (c *Client) Update(path string) error {
	cmd := exec.Command("svn", "update", path, "--username", c.Username)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("svn update failed: %w", err)
	}

	return nil
}

// Status returns the status of a working copy
func (c *Client) Status(path string) (string, error) {
	cmd := exec.Command("svn", "status", path, "--username", c.Username)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("svn status failed: %w", err)
	}

	return string(output), nil
}

// Info returns information about a working copy or URL
func (c *Client) Info(path string) (string, error) {
	cmd := exec.Command("svn", "info", path, "--username", c.Username)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("svn info failed: %w", err)
	}

	return string(output), nil
}

// Add adds a new component to SVN
func (c *Client) Add(componentPath, componentType string) error {
	// Create directory structure in SVN
	mkdirURL := fmt.Sprintf("%s/%s/components/%s/%s", c.URL, c.Repo, componentType, componentPath)

	// Create component directory structure (trunk, tags, branches)
	cmd := exec.Command("svn", "mkdir",
		mkdirURL,
		mkdirURL+"/trunk",
		mkdirURL+"/tags",
		mkdirURL+"/branches",
		"-m", fmt.Sprintf("Created component %s", componentPath),
		"--username", c.Username)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("svn mkdir failed: %w\n%s", err, output)
	}

	return nil
}

// IsWorkingCopy checks if a path is an SVN working copy
func IsWorkingCopy(path string) bool {
	svnDir := path + "/.svn"
	info, err := os.Stat(svnDir)
	return err == nil && info.IsDir()
}

// GetBranch returns the current branch/tag of a working copy
func (c *Client) GetBranch(path string) (string, error) {
	output, err := c.Info(path)
	if err != nil {
		return "", err
	}

	// Parse URL from svn info output
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "URL:") {
			url := strings.TrimSpace(strings.TrimPrefix(line, "URL:"))
			// Extract branch from URL (e.g., .../trunk or .../tags/v1.0)
			parts := strings.Split(url, "/")
			if len(parts) >= 2 {
				// Last part is the branch/tag name
				lastPart := parts[len(parts)-1]
				if lastPart == "trunk" {
					return "trunk", nil
				}
				// Check if it's in tags or branches
				if len(parts) >= 3 {
					secondLast := parts[len(parts)-2]
					if secondLast == "tags" || secondLast == "branches" {
						return secondLast + "/" + lastPart, nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("could not determine branch from svn info")
}
