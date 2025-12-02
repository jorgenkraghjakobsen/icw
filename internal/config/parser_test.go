package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jakobsen/icw/internal/component"
)

func TestParseDependConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a workspace
	ws := component.NewWorkspace(tmpDir)
	parser := NewParser(ws)

	// Create a parent component
	parent := &component.Component{
		Name:       "digital/top",
		Path:       "digital/top",
		Type:       component.TypeDigital,
		Branch:     "trunk",
		VCS:        "svn",
		DeclaredBy: "workspace.config",
	}

	// Create a depend.config file
	dependConfigPath := filepath.Join(tmpDir, "depend.config")
	dependContent := `# Test depend.config
use component("digital/spi_master", "digital", "trunk")
use component("analog/bias", "analog", "tags/v1.0")
`
	err := os.WriteFile(dependConfigPath, []byte(dependContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create depend.config: %v", err)
	}

	// Parse depend.config
	deps, err := parser.ParseDependConfig(parent, dependConfigPath)
	if err != nil {
		t.Fatalf("Failed to parse depend.config: %v", err)
	}

	// Verify we got 2 dependencies
	if len(deps) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(deps))
	}

	// Verify dependencies were added to workspace
	if len(ws.Components) != 2 {
		t.Errorf("Expected 2 components in workspace, got %d", len(ws.Components))
	}

	// Verify first dependency
	dep1, ok := ws.GetComponent("digital/spi_master")
	if !ok {
		t.Error("digital/spi_master not found in workspace")
	} else {
		if dep1.Branch != "trunk" {
			t.Errorf("Expected branch 'trunk', got '%s'", dep1.Branch)
		}
		if dep1.DeclaredBy != "digital/top" {
			t.Errorf("Expected DeclaredBy 'digital/top', got '%s'", dep1.DeclaredBy)
		}
	}

	// Verify second dependency
	dep2, ok := ws.GetComponent("analog/bias")
	if !ok {
		t.Error("analog/bias not found in workspace")
	} else {
		if dep2.Branch != "tags/v1.0" {
			t.Errorf("Expected branch 'tags/v1.0', got '%s'", dep2.Branch)
		}
		if dep2.DeclaredBy != "digital/top" {
			t.Errorf("Expected DeclaredBy 'digital/top', got '%s'", dep2.DeclaredBy)
		}
	}

	// Verify parent's dependencies list
	if len(parent.Dependencies) != 2 {
		t.Errorf("Expected parent to have 2 dependencies, got %d", len(parent.Dependencies))
	}
}

func TestParseDependConfigConflict(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a workspace
	ws := component.NewWorkspace(tmpDir)
	parser := NewParser(ws)

	// Add a component to the workspace first
	existingComp := &component.Component{
		Name:       "digital/spi_master",
		Path:       "digital/spi_master",
		Type:       component.TypeDigital,
		Branch:     "trunk",
		VCS:        "svn",
		DeclaredBy: "digital/module1",
	}
	ws.AddComponent(existingComp)

	// Create a parent component
	parent := &component.Component{
		Name:       "digital/top",
		Path:       "digital/top",
		Type:       component.TypeDigital,
		Branch:     "trunk",
		VCS:        "svn",
		DeclaredBy: "workspace.config",
	}

	// Create a depend.config file with conflicting version
	dependConfigPath := filepath.Join(tmpDir, "depend.config")
	dependContent := `# Test depend.config with conflict
use component("digital/spi_master", "digital", "tags/v2.0")
`
	err := os.WriteFile(dependConfigPath, []byte(dependContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create depend.config: %v", err)
	}

	// Parse depend.config - should return conflict error
	_, err = parser.ParseDependConfig(parent, dependConfigPath)
	if err == nil {
		t.Error("Expected conflict error, got nil")
	}

	// Verify it's a conflict error
	if err != nil && !containsString(err.Error(), "branch mismatch") {
		t.Errorf("Expected 'branch mismatch' error, got: %v", err)
	}
}

func TestParseDependConfigCircular(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a workspace
	ws := component.NewWorkspace(tmpDir)
	parser := NewParser(ws)

	// Create a component
	comp := &component.Component{
		Name:       "digital/module1",
		Path:       "digital/module1",
		Type:       component.TypeDigital,
		Branch:     "trunk",
		VCS:        "svn",
		DeclaredBy: "workspace.config",
	}

	// Parse depend.config first time
	dependConfigPath := filepath.Join(tmpDir, "depend.config")
	dependContent := `use component("digital/module2", "digital", "trunk")`
	os.WriteFile(dependConfigPath, []byte(dependContent), 0644)

	parser.ParseDependConfig(comp, dependConfigPath)

	// Parse again for the same component - should skip due to processed tracking
	deps, err := parser.ParseDependConfig(comp, dependConfigPath)
	if err != nil {
		t.Errorf("Second parse should not error: %v", err)
	}

	// Should return nil (already processed)
	if deps != nil {
		t.Errorf("Expected nil deps for already processed component, got %d deps", len(deps))
	}
}

func TestParseDependConfigMissing(t *testing.T) {
	// Create a workspace
	ws := component.NewWorkspace("/tmp/test")
	parser := NewParser(ws)

	// Create a parent component
	parent := &component.Component{
		Name:       "digital/top",
		Path:       "digital/top",
		Type:       component.TypeDigital,
		Branch:     "trunk",
		VCS:        "svn",
		DeclaredBy: "workspace.config",
	}

	// Try to parse a non-existent depend.config
	deps, err := parser.ParseDependConfig(parent, "/tmp/nonexistent/depend.config")
	if err != nil {
		t.Errorf("Should not error on missing depend.config: %v", err)
	}

	// Should return nil (no dependencies)
	if deps != nil {
		t.Errorf("Expected nil deps for missing depend.config, got %d deps", len(deps))
	}
}

func containsString(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(str) > len(substr) &&
		(str[:len(substr)] == substr || str[len(str)-len(substr):] == substr ||
		len(str) > len(substr) && findInString(str, substr)))
}

func findInString(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
