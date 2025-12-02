package component

import (
	"strings"
	"testing"
)

func TestAddComponent(t *testing.T) {
	ws := NewWorkspace("/tmp/test")

	// Test adding a new component
	comp1 := &Component{
		Name:       "digital/module1",
		Path:       "digital/module1",
		Type:       TypeDigital,
		Branch:     "trunk",
		VCS:        "svn",
		DeclaredBy: "workspace.config",
	}

	err := ws.AddComponent(comp1)
	if err != nil {
		t.Errorf("Failed to add first component: %v", err)
	}

	// Test adding the same component with same branch - should not error
	comp2 := &Component{
		Name:       "digital/module1",
		Path:       "digital/module1",
		Type:       TypeDigital,
		Branch:     "trunk",
		VCS:        "svn",
		DeclaredBy: "digital/module2",
	}

	err = ws.AddComponent(comp2)
	if err != nil {
		t.Errorf("Failed to add same component with same branch: %v", err)
	}

	// Verify DeclaredBy was updated to include both sources
	stored, ok := ws.GetComponent("digital/module1")
	if !ok {
		t.Error("Component not found after adding")
	}
	if !strings.Contains(stored.DeclaredBy, "workspace.config") || !strings.Contains(stored.DeclaredBy, "digital/module2") {
		t.Errorf("DeclaredBy should contain both sources, got: %s", stored.DeclaredBy)
	}

	// Test adding the same component with different branch - should error
	comp3 := &Component{
		Name:       "digital/module1",
		Path:       "digital/module1",
		Type:       TypeDigital,
		Branch:     "tags/v1.0",
		VCS:        "svn",
		DeclaredBy: "digital/module3",
	}

	err = ws.AddComponent(comp3)
	if err == nil {
		t.Error("Expected error when adding component with conflicting branch")
	}

	// Verify it's a BranchConflictError
	if conflictErr, ok := err.(*BranchConflictError); ok {
		if conflictErr.Component != "digital/module1" {
			t.Errorf("Expected component name 'digital/module1', got %s", conflictErr.Component)
		}
		if conflictErr.Existing != "trunk" {
			t.Errorf("Expected existing branch 'trunk', got %s", conflictErr.Existing)
		}
		if conflictErr.New != "tags/v1.0" {
			t.Errorf("Expected new branch 'tags/v1.0', got %s", conflictErr.New)
		}
		if !strings.Contains(conflictErr.ExistingSource, "workspace.config") {
			t.Errorf("Expected ExistingSource to contain 'workspace.config', got %s", conflictErr.ExistingSource)
		}
		if conflictErr.NewSource != "digital/module3" {
			t.Errorf("Expected NewSource 'digital/module3', got %s", conflictErr.NewSource)
		}
	} else {
		t.Errorf("Expected BranchConflictError, got %T", err)
	}
}

func TestBranchConflictErrorMessage(t *testing.T) {
	err := &BranchConflictError{
		Component:      "digital/spi",
		Existing:       "trunk",
		ExistingSource: "workspace.config",
		New:            "tags/v2.0",
		NewSource:      "digital/top",
	}

	msg := err.Error()

	// Verify message contains key information
	if !strings.Contains(msg, "digital/spi") {
		t.Error("Error message should contain component name")
	}
	if !strings.Contains(msg, "trunk") {
		t.Error("Error message should contain existing branch")
	}
	if !strings.Contains(msg, "tags/v2.0") {
		t.Error("Error message should contain new branch")
	}
	if !strings.Contains(msg, "workspace.config") {
		t.Error("Error message should contain existing source")
	}
	if !strings.Contains(msg, "digital/top") {
		t.Error("Error message should contain new source")
	}
}
