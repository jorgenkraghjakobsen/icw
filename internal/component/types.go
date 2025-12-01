package component

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
	Resolved bool
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
				Component: comp.Name,
				Existing:  existing.Branch,
				New:       comp.Branch,
			}
		}
		// Same branch, no conflict
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
	Component string
	Existing  string
	New       string
}

func (e *BranchConflictError) Error() string {
	return "branch mismatch for " + e.Component + ": " + e.Existing + " vs " + e.New
}
