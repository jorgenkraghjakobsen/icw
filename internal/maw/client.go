package maw

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Client wraps MAW backend functionality
type Client struct {
	repoPath    string
	sasldbPath  string
}

// NewClient creates a new MAW client
func NewClient() (*Client, error) {
	// Check if running on g9
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	if hostname != "g9" {
		return nil, fmt.Errorf("MAW operations must run on g9 server (current: %s)", hostname)
	}

	return &Client{
		repoPath:   "/data_v1/svn/repos",
		sasldbPath: os.Getenv("SASLPASSWD"),
	}, nil
}

// CreateRepo creates a new SVN repository
func (c *Client) CreateRepo(repoName string) error {
	// Check if repo already exists
	repoFullPath := fmt.Sprintf("%s/%s", c.repoPath, repoName)
	if _, err := os.Stat(repoFullPath); err == nil {
		return fmt.Errorf("repository %s already exists", repoName)
	}

	// Check if we have write permissions to the repos directory
	if _, err := os.Stat(c.repoPath); err != nil {
		return fmt.Errorf("cannot access repository directory %s: %w", c.repoPath, err)
	}

	// Create repository using svnadmin with sudo (repos directory is owned by root)
	cmd := exec.Command("sudo", "svnadmin", "create", repoFullPath)

	// Capture both stdout and stderr for better error messages
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			return fmt.Errorf("failed to create repository: %w\nOutput: %s", err, strings.TrimSpace(string(output)))
		}
		return fmt.Errorf("failed to create repository: %w", err)
	}

	// Write svnserve.conf using sudo (since we created the repo with sudo)
	svnConf := fmt.Sprintf("%s/conf/svnserve.conf", repoFullPath)
	confContent := fmt.Sprintf(`[general]
realm = %s
anon-access = none
auth-access = write

[sasl]
use-sasl = true
`, repoName)

	// Use sudo to write the config file
	cmd = exec.Command("sudo", "tee", svnConf)
	cmd.Stdin = strings.NewReader(confContent)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to write svnserve.conf: %w\nOutput: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

// ListRepos returns a list of all repositories
func (c *Client) ListRepos() ([]string, error) {
	files, err := os.ReadDir(c.repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read repository directory: %w", err)
	}

	var repos []string
	for _, file := range files {
		if file.IsDir() {
			repos = append(repos, file.Name())
		}
	}

	return repos, nil
}

// RepoExists checks if a repository exists
func (c *Client) RepoExists(repoName string) bool {
	repoFullPath := fmt.Sprintf("%s/%s", c.repoPath, repoName)
	_, err := os.Stat(repoFullPath)
	return err == nil
}

// ListRepoUsers returns users for a specific repository
func (c *Client) ListRepoUsers(repoName string) ([]string, error) {
	if c.sasldbPath == "" {
		c.sasldbPath = "/etc/svn_repos_sasldb"
	}

	cmd := exec.Command("sasldblistusers2", "-f", c.sasldbPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w\nOutput: %s", err, strings.TrimSpace(string(output)))
	}

	// Parse output and filter by repo
	var repoUsers []string
	userList := strings.Split(string(output), "\n")
	for _, user := range userList {
		if strings.Contains(user, "@"+repoName+":") {
			// Extract username before the colon
			parts := strings.Split(user, ":")
			if len(parts) > 0 {
				username := strings.TrimSpace(parts[0])
				repoUsers = append(repoUsers, username)
			}
		}
	}

	return repoUsers, nil
}

// AddUserToRepo adds a user to a repository
func (c *Client) AddUserToRepo(repo, username, password string) error {
	if c.sasldbPath == "" {
		c.sasldbPath = "/etc/svn_repos_sasldb"
	}

	cmd := exec.Command("sudo", "saslpasswd2", "-c", "-f", c.sasldbPath, "-u", repo, username)
	cmd.Stdin = strings.NewReader(password + "\n")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add user to repo: %w\nOutput: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

// RemoveUserFromRepo removes a user from a repository
func (c *Client) RemoveUserFromRepo(repo, username string) error {
	if c.sasldbPath == "" {
		c.sasldbPath = "/etc/svn_repos_sasldb"
	}

	cmd := exec.Command("sudo", "saslpasswd2", "-d", "-f", c.sasldbPath, "-u", repo, username)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove user from repo: %w\nOutput: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

// DeleteRepo deletes a repository (archives it first)
func (c *Client) DeleteRepo(repoName string) error {
	if !c.RepoExists(repoName) {
		return fmt.Errorf("repository %s does not exist", repoName)
	}

	repoFullPath := fmt.Sprintf("%s/%s", c.repoPath, repoName)
	archivePath := "/data_v1/svn/deleted"

	// Ensure archive directory exists
	if err := os.MkdirAll(archivePath, 0755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	// Archive repository using sudo
	cmd := exec.Command("sudo", "cp", "-r", repoFullPath, archivePath+"/"+repoName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to archive repository: %w\nOutput: %s", err, strings.TrimSpace(string(output)))
	}

	// Remove repository using sudo
	cmd = exec.Command("sudo", "rm", "-rf", repoFullPath)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove repository: %w\nOutput: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}
