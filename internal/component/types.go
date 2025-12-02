package component

import "strings"

// ComponentType represents the type of a component
type ComponentType string

const (
	TypeAnalog  ComponentType = "analog"
	TypeDigital ComponentType = "digital"
	TypeSetup   ComponentType = "setup"
	TypeProcess ComponentType = "process"
	TypeTools   ComponentType = "tools" // For software tools in Git
)

// Component represents a design component or software tool
type Component struct {
	Name   string        // Component identifier (e.g., "digital/my_module")
	Path   string        // Path in workspace
	Type   ComponentType // Type of component
	Branch string        // SVN branch/tag or Git branch (e.g., "trunk", "tags/v1.0", "main")
	VCS    string        // Version control system: "svn" or "git"

	// Dependencies
	Dependencies []*Component

	// For tracking conflicts
	Resolved   bool
	DeclaredBy string // Name of component that declared this dependency (for conflict reporting)
}

// Workspace represents the entire workspace configuration
type Workspace struct {
	Root       string                // Workspace root directory
	Components map[string]*Component // Components indexed by name
	Config     string                // Path to workspace.config
}

// NewWorkspace creates a new workspace instance
func NewWorkspace(root string) *Workspace {
	return &Workspace{
		Root:       root,
		Components: make(map[string]*Component),
		Config:     root + "/workspace.config",
	}
}

// AddComponent adds a component to the workspace
// Returns error if there's a branch conflict
func (w *Workspace) AddComponent(comp *Component) error {
	if existing, ok := w.Components[comp.Name]; ok {
		// Component already exists, check for branch conflicts
		if existing.Branch != comp.Branch {
			return &BranchConflictError{
				Component:      comp.Name,
				Existing:       existing.Branch,
				ExistingSource: existing.DeclaredBy,
				New:            comp.Branch,
				NewSource:      comp.DeclaredBy,
			}
		}
		// Same branch, no conflict - but update DeclaredBy to include both sources if different
		if comp.DeclaredBy != "" && existing.DeclaredBy != comp.DeclaredBy {
			if existing.DeclaredBy == "" {
				existing.DeclaredBy = comp.DeclaredBy
			} else if !strings.Contains(existing.DeclaredBy, comp.DeclaredBy) {
				existing.DeclaredBy = existing.DeclaredBy + ", " + comp.DeclaredBy
			}
		}
		return nil
	}

	// Add new component
	w.Components[comp.Name] = comp
	return nil
}

// GetComponent retrieves a component by name
func (w *Workspace) GetComponent(name string) (*Component, bool) {
	comp, ok := w.Components[name]
	return comp, ok
}

// BranchConflictError represents a branch mismatch error
type BranchConflictError struct {
	Component       string
	Existing        string
	ExistingSource  string
	New             string
	NewSource       string
}

func (e *BranchConflictError) Error() string {
	msg := "branch mismatch for component '" + e.Component + "'\n"
	if e.ExistingSource != "" {
		msg += "  First declared by: " + e.ExistingSource + " requesting '" + e.Existing + "'\n"
	} else {
		msg += "  First declared: '" + e.Existing + "'\n"
	}
	if e.NewSource != "" {
		msg += "  Also declared by: " + e.NewSource + " requesting '" + e.New + "'"
	} else {
		msg += "  Also declared: '" + e.New + "'"
	}
	return msg
}
