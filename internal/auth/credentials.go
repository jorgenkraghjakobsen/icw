package auth

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// CredentialsFile returns the path to the credentials file
func CredentialsFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".icw", "credentials")
}

// EnsureCredentialsDir creates the .icw directory if it doesn't exist
func EnsureCredentialsDir() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	icwDir := filepath.Join(home, ".icw")
	if err := os.MkdirAll(icwDir, 0700); err != nil {
		return fmt.Errorf("failed to create .icw directory: %w", err)
	}

	return nil
}

// SavePassword saves the SVN password to the credentials file
func SavePassword(password string) error {
	if err := EnsureCredentialsDir(); err != nil {
		return err
	}

	credFile := CredentialsFile()

	// Write password to file with restricted permissions
	f, err := os.OpenFile(credFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create credentials file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(password); err != nil {
		return fmt.Errorf("failed to write password: %w", err)
	}

	return nil
}

// LoadPassword loads the SVN password from the credentials file
func LoadPassword() (string, error) {
	credFile := CredentialsFile()

	// Check if file exists
	if _, err := os.Stat(credFile); os.IsNotExist(err) {
		return "", nil // No credentials stored
	}

	// Read password from file
	data, err := os.ReadFile(credFile)
	if err != nil {
		return "", fmt.Errorf("failed to read credentials: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}

// DeletePassword removes the stored password
func DeletePassword() error {
	credFile := CredentialsFile()

	if err := os.Remove(credFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}

	return nil
}

// PromptPassword prompts the user to enter their password (with hidden input)
func PromptPassword() (string, error) {
	fmt.Print("Enter SVN password: ")

	// Read password without echoing
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // New line after password input

	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	password := strings.TrimSpace(string(passwordBytes))
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	return password, nil
}

// PromptUsername prompts the user to enter their username
func PromptUsername(defaultUser string) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	if defaultUser != "" {
		fmt.Printf("Username [%s]: ", defaultUser)
	} else {
		fmt.Print("Username: ")
	}

	username, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read username: %w", err)
	}

	username = strings.TrimSpace(username)

	// Use default if empty
	if username == "" && defaultUser != "" {
		return defaultUser, nil
	}

	if username == "" {
		return "", fmt.Errorf("username cannot be empty")
	}

	return username, nil
}

// GetPassword returns the password from stored credentials, env var, or prompts
func GetPassword() (string, error) {
	// 1. Check environment variable first (for scripts/automation)
	if envPassword := os.Getenv("ICW_SVN_PASSWORD"); envPassword != "" {
		return envPassword, nil
	}

	// 2. Check stored credentials
	storedPassword, err := LoadPassword()
	if err != nil {
		return "", err
	}

	if storedPassword != "" {
		return storedPassword, nil
	}

	// 3. No stored credentials - return empty (caller should prompt or error)
	return "", nil
}

// HasStoredCredentials checks if credentials are stored
func HasStoredCredentials() bool {
	credFile := CredentialsFile()
	_, err := os.Stat(credFile)
	return err == nil
}
