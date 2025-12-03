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

	// Create repository using svnadmin
	cmd := exec.Command("svnadmin", "create", repoName)
	cmd.Dir = c.repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	// Write svnserve.conf
	svnConf := fmt.Sprintf("%s/conf/svnserve.conf", repoFullPath)
	f, err := os.Create(svnConf)
	if err != nil {
		return fmt.Errorf("failed to create svnserve.conf: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, `[general]
realm = %s
anon-access = none
auth-access = write

[sasl]
use-sasl = true
`, repoName)
	if err != nil {
		return fmt.Errorf("failed to write svnserve.conf: %w", err)
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
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
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

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add user to repo: %w", err)
	}

	return nil
}

// RemoveUserFromRepo removes a user from a repository
func (c *Client) RemoveUserFromRepo(repo, username string) error {
	if c.sasldbPath == "" {
		c.sasldbPath = "/etc/svn_repos_sasldb"
	}

	cmd := exec.Command("sudo", "saslpasswd2", "-d", "-f", c.sasldbPath, "-u", repo, username)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove user from repo: %w", err)
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

	// Archive repository
	cmd := exec.Command("cp", "-r", repoFullPath, archivePath+"/"+repoName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to archive repository: %w", err)
	}

	// Remove repository
	cmd = exec.Command("rm", "-rf", repoFullPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove repository: %w", err)
	}

	return nil
}
