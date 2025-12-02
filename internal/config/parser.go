package config

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/jakobsen/icw/internal/component"
)

// Parser handles parsing of workspace.config and depend.config files
type Parser struct {
	workspace *component.Workspace
	Repo      string // Repository name from config file
	SvnURL    string // SVN URL from config file
	processed map[string]bool // Track processed components to avoid infinite loops
}

// NewParser creates a new config parser
func NewParser(ws *component.Workspace) *Parser {
	return &Parser{
		workspace: ws,
		processed: make(map[string]bool),
	}
}

// ParseWorkspaceConfig parses the workspace.config file
func (p *Parser) ParseWorkspaceConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open config: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for repository configuration
		if p.parseRepoConfig(line) {
			continue
		}

		// Parse component declaration
		comp, err := p.parseComponentLine(line)
		if err != nil {
			return fmt.Errorf("line %d: %w", lineNum, err)
		}

		if comp != nil {
			// Components from workspace.config are declared by the workspace itself
			comp.DeclaredBy = "workspace.config"

			if err := p.workspace.AddComponent(comp); err != nil {
				return fmt.Errorf("line %d: %w", lineNum, err)
			}
		}
	}

	return scanner.Err()
}

// parseRepoConfig parses repository configuration directives
// Returns true if the line was a config directive
func (p *Parser) parseRepoConfig(line string) bool {
	// Pattern for: set repo "repo_name"
	repoPattern := regexp.MustCompile(`set\s+repo\s+"([^"]+)"`)
	if matches := repoPattern.FindStringSubmatch(line); matches != nil {
		p.Repo = matches[1]
		return true
	}

	// Pattern for: set svn_url "svn://server"
	urlPattern := regexp.MustCompile(`set\s+svn_url\s+"([^"]+)"`)
	if matches := urlPattern.FindStringSubmatch(line); matches != nil {
		p.SvnURL = matches[1]
		return true
	}

	return false
}

// parseComponentLine parses a single line from the config file
// Supports formats:
//   use component("path/to/component", "type", "branch")
//   use component("path/to/component", "type")  # defaults to trunk
//   use component("path/to/component")          # infers type from path
//   use ref("path/to/local")                    # local reference
func (p *Parser) parseComponentLine(line string) (*component.Component, error) {
	// Pattern for: use component("path", "type", "branch")
	fullPattern := regexp.MustCompile(`use\s+component\s*\(\s*"([^"]+)"\s*,\s*"([^"]+)"\s*,\s*"([^"]+)"\s*\)`)

	// Pattern for: use component("path", "type")
	typePattern := regexp.MustCompile(`use\s+component\s*\(\s*"([^"]+)"\s*,\s*"([^"]+)"\s*\)`)

	// Pattern for: use component("path")
	pathPattern := regexp.MustCompile(`use\s+component\s*\(\s*"([^"]+)"\s*\)`)

	// Pattern for: use ref("path")
	refPattern := regexp.MustCompile(`use\s+ref\s*\(\s*"([^"]+)"\s*\)`)

	// Try full pattern first
	if matches := fullPattern.FindStringSubmatch(line); matches != nil {
		return &component.Component{
			Name:   matches[1],
			Path:   matches[1],
			Type:   component.ComponentType(matches[2]),
			Branch: matches[3],
			VCS:    inferVCS(component.ComponentType(matches[2])),
		}, nil
	}

	// Try type pattern
	if matches := typePattern.FindStringSubmatch(line); matches != nil {
		return &component.Component{
			Name:   matches[1],
			Path:   matches[1],
			Type:   component.ComponentType(matches[2]),
			Branch: "trunk", // Default branch for SVN
			VCS:    inferVCS(component.ComponentType(matches[2])),
		}, nil
	}

	// Try path-only pattern
	if matches := pathPattern.FindStringSubmatch(line); matches != nil {
		compType := inferTypeFromPath(matches[1])
		return &component.Component{
			Name:   matches[1],
			Path:   matches[1],
			Type:   compType,
			Branch: "trunk",
			VCS:    inferVCS(compType),
		}, nil
	}

	// Try ref pattern (local reference)
	if matches := refPattern.FindStringSubmatch(line); matches != nil {
		// For local refs, we don't check them out
		// Just record them for dependency resolution
		compType := inferTypeFromPath(matches[1])
		return &component.Component{
			Name:   matches[1],
			Path:   matches[1],
			Type:   compType,
			Branch: "local",
			VCS:    "local",
		}, nil
	}

	// Unknown format
	if strings.Contains(line, "use") {
		return nil, fmt.Errorf("invalid component syntax: %s", line)
	}

	// Not a component line, skip it
	return nil, nil
}

// inferTypeFromPath infers component type from its path
func inferTypeFromPath(path string) component.ComponentType {
	if strings.HasPrefix(path, "analog/") {
		return component.TypeAnalog
	}
	if strings.HasPrefix(path, "digital/") {
		return component.TypeDigital
	}
	if strings.HasPrefix(path, "setup/") {
		return component.TypeSetup
	}
	if strings.HasPrefix(path, "process/") || strings.HasPrefix(path, "process_setup/") {
		return component.TypeProcess
	}
	if strings.HasPrefix(path, "tools/") || strings.HasPrefix(path, "software/") {
		return component.TypeTools
	}
	// Default to digital
	return component.TypeDigital
}

// inferVCS determines which VCS to use based on component type
func inferVCS(compType component.ComponentType) string {
	switch compType {
	case component.TypeTools:
		return "git" // Software tools use Git
	case component.TypeAnalog, component.TypeDigital, component.TypeSetup, component.TypeProcess:
		return "svn" // Design components use SVN
	default:
		return "svn" // Default to SVN
	}
}

// ParseDependConfig parses a component's depend.config file and resolves dependencies
// Returns a slice of dependency components found
func (p *Parser) ParseDependConfig(parent *component.Component, dependConfigPath string) ([]*component.Component, error) {
	// Check if we've already processed this component to avoid circular dependencies
	if p.processed[parent.Name] {
		return nil, nil // Already processed, skip
	}
	p.processed[parent.Name] = true

	// Check if depend.config exists
	if _, err := os.Stat(dependConfigPath); os.IsNotExist(err) {
		return nil, nil // No dependencies
	}

	file, err := os.Open(dependConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open depend.config: %w", err)
	}
	defer file.Close()

	var dependencies []*component.Component
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse component declaration (same syntax as workspace.config)
		comp, err := p.parseComponentLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		if comp != nil {
			// Set the DeclaredBy field to track where this dependency came from
			comp.DeclaredBy = parent.Name

			// Add component to workspace (with conflict detection)
			if err := p.workspace.AddComponent(comp); err != nil {
				// Check if it's a branch conflict
				if conflictErr, ok := err.(*component.BranchConflictError); ok {
					return nil, fmt.Errorf("dependency conflict: %w", conflictErr)
				}
				return nil, err
			}

			// Add to parent's dependencies
			parent.Dependencies = append(parent.Dependencies, comp)
			dependencies = append(dependencies, comp)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return dependencies, nil
}
