package svn

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jakobsen/icw/internal/auth"
)

// Client represents an SVN client
type Client struct {
	URL      string // Base SVN URL (e.g., svn://anyvej11.dk)
	Repo     string // Repository name (from ICW_REPO env var)
	Username string // SVN username
	Password string // SVN password (from ICW_SVN_PASSWORD env var, optional)
}

// buildAuthArgs returns common authentication arguments for svn commands
func (c *Client) buildAuthArgs() []string {
	args := []string{"--username", c.Username, "--non-interactive", "--trust-server-cert"}
	if c.Password != "" {
		args = append(args, "--password", c.Password)
	}
	return args
}

// NewClient creates a new SVN client
func NewClient() (*Client, error) {
	return NewClientWithConfig("", "")
}

// NewClientWithConfig creates a new SVN client with explicit configuration
// If repo or svnURL are empty, falls back to environment variables
func NewClientWithConfig(repo, svnURL string) (*Client, error) {
	// Get repository from parameter, env var, or error
	if repo == "" {
		repo = os.Getenv("ICW_REPO")
	}
	if repo == "" {
		return nil, fmt.Errorf("ICW_REPO not set\nPlease set it with: export ICW_REPO=repo_name\nOr add to workspace.config: set repo \"repo_name\"")
	}

	// Get SVN server URL from parameter, env var, or use default
	if svnURL == "" {
		svnURL = os.Getenv("ICW_SVN_URL")
	}
	if svnURL == "" {
		// Auto-detect SVN URL based on hostname
		hostname, err := os.Hostname()
		if err == nil && hostname == "g9" {
			// On g9 server, use local svnserve
			svnURL = "svn://g9"
		} else {
			// Default to remote server
			svnURL = "svn://anyvej11.dk"
		}
	}

	username := os.Getenv("USER")
	if username == "" {
		username = "anonymous"
	}

	// Get password from multiple sources (env var, stored credentials)
	password, err := auth.GetPassword()
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	return &Client{
		URL:      svnURL,
		Repo:     repo,
		Username: username,
		Password: password,
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

// Cat reads a file directly from the repository without checking it out
func (c *Client) Cat(componentPath, branch, filename string) (string, error) {
	// Construct URL to the file in the repository
	url := fmt.Sprintf("%s/%s/components/%s/%s/%s", c.URL, c.Repo, componentPath, branch, filename)

	args := append([]string{"cat", url}, c.buildAuthArgs()...)
	cmd := exec.Command("svn", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// File might not exist, which is OK for depend.config
		return "", fmt.Errorf("svn cat failed: %w", err)
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

// TestConnection tests connectivity to the SVN server and repository
func (c *Client) TestConnection() error {
	// Try to list the repository root
	repoURL := fmt.Sprintf("%s/%s", c.URL, c.Repo)
	args := append([]string{"list", repoURL}, c.buildAuthArgs()...)
	args = append(args, "--depth", "immediates")
	cmd := exec.Command("svn", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w\n%s", repoURL, err, output)
	}

	return nil
}

// ListComponents lists available components in the repository
func (c *Client) ListComponents() ([]string, error) {
	componentsURL := fmt.Sprintf("%s/%s/components", c.URL, c.Repo)
	args := append([]string{"list", componentsURL}, c.buildAuthArgs()...)
	cmd := exec.Command("svn", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list components: %w\n%s", err, output)
	}

	// Parse output into component list
	var components []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && line != "/" {
			// Remove trailing slash
			components = append(components, strings.TrimSuffix(line, "/"))
		}
	}

	return components, nil
}

// ListComponentsByType lists components of a specific type (analog, digital, setup, process)
func (c *Client) ListComponentsByType(componentType string) ([]string, error) {
	typeURL := fmt.Sprintf("%s/%s/components/%s", c.URL, c.Repo, componentType)
	args := append([]string{"list", typeURL}, c.buildAuthArgs()...)
	args = append(args, "--depth", "infinity")
	cmd := exec.Command("svn", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list %s components: %w\n%s", componentType, err, output)
	}

	// Parse output into component list
	var components []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && line != "/" && strings.HasSuffix(line, "/") {
			// Only include directories (end with /)
			component := strings.TrimSuffix(line, "/")
			// Filter out trunk/tags/branches subdirectories
			if !strings.Contains(component, "/") {
				components = append(components, componentType+"/"+component)
			}
		}
	}

	return components, nil
}

// ListBranches lists all branches for a component
func (c *Client) ListBranches(componentPath string) ([]string, error) {
	branchesURL := fmt.Sprintf("%s/%s/components/%s/branches", c.URL, c.Repo, componentPath)
	args := append([]string{"list", branchesURL}, c.buildAuthArgs()...)
	cmd := exec.Command("svn", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w\n%s", err, output)
	}

	// Parse output into branch list
	var branches []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && line != "/" {
			// Remove trailing slash
			branches = append(branches, strings.TrimSuffix(line, "/"))
		}
	}

	return branches, nil
}

// ListTags lists all tags for a component
func (c *Client) ListTags(componentPath string) ([]string, error) {
	tagsURL := fmt.Sprintf("%s/%s/components/%s/tags", c.URL, c.Repo, componentPath)
	args := append([]string{"list", tagsURL}, c.buildAuthArgs()...)
	cmd := exec.Command("svn", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w\n%s", err, output)
	}

	// Parse output into tag list
	var tags []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && line != "/" {
			// Remove trailing slash
			tags = append(tags, strings.TrimSuffix(line, "/"))
		}
	}

	return tags, nil
}

// ComponentInfo holds detailed information about a component
type ComponentInfo struct {
	Path     string
	HasTrunk bool
	Branches []string
	Tags     []string
}

// GetComponentInfo retrieves detailed information about a component
func (c *Client) GetComponentInfo(componentPath string) (*ComponentInfo, error) {
	info := &ComponentInfo{
		Path: componentPath,
	}

	// Check if trunk exists
	trunkURL := fmt.Sprintf("%s/%s/components/%s/trunk", c.URL, c.Repo, componentPath)
	args := append([]string{"list", trunkURL}, c.buildAuthArgs()...)
	args = append(args, "--depth", "empty")
	cmd := exec.Command("svn", args...)
	if err := cmd.Run(); err == nil {
		info.HasTrunk = true
	}

	// Get branches
	branches, err := c.ListBranches(componentPath)
	if err == nil {
		info.Branches = branches
	}

	// Get tags
	tags, err := c.ListTags(componentPath)
	if err == nil {
		info.Tags = tags
	}

	return info, nil
}

// FindComponentsByPattern finds components matching a glob pattern
func (c *Client) FindComponentsByPattern(pattern string) ([]string, error) {
	// Split pattern into type and name parts
	// e.g., "digital/dig*" -> type="digital", namePattern="dig*"
	parts := strings.SplitN(pattern, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("pattern must be in format type/pattern (e.g., digital/dig*)")
	}

	componentType := parts[0]
	namePattern := parts[1]

	// Get all components of the specified type
	components, err := c.ListComponentsByType(componentType)
	if err != nil {
		return nil, err
	}

	// Filter components matching the pattern
	var matches []string
	for _, comp := range components {
		// Extract just the component name (after type/)
		compParts := strings.SplitN(comp, "/", 2)
		if len(compParts) == 2 {
			compName := compParts[1]
			// Simple glob matching (* at beginning, end, or both)
			if matchGlob(compName, namePattern) {
				matches = append(matches, comp)
			}
		}
	}

	return matches, nil
}

// matchGlob performs simple glob pattern matching
// Supports: *, prefix*, *suffix, *middle*, exact
func matchGlob(name, pattern string) bool {
	// No wildcards - exact match
	if !strings.Contains(pattern, "*") {
		return name == pattern
	}

	// Count wildcards
	wildcardCount := strings.Count(pattern, "*")

	if wildcardCount == 1 {
		if strings.HasPrefix(pattern, "*") {
			// *suffix
			suffix := strings.TrimPrefix(pattern, "*")
			return strings.HasSuffix(name, suffix)
		} else if strings.HasSuffix(pattern, "*") {
			// prefix*
			prefix := strings.TrimSuffix(pattern, "*")
			return strings.HasPrefix(name, prefix)
		}
	} else if wildcardCount == 2 && strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		// *middle*
		middle := strings.Trim(pattern, "*")
		return strings.Contains(name, middle)
	}

	return false
}
