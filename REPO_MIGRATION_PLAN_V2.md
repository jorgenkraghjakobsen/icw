# ICW Repository Migration Plan - Updated with MAW Integration

## Executive Summary

Based on review of the existing MAW (icw-maw) system, we can integrate repository creation and user management directly with the g9 backend.

## MAW System Integration

### Existing System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      MAW System (g9)                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────────┐      ┌──────────────────────┐    │
│  │   Web Interface      │      │   Backend Service    │    │
│  │   (Go HTTP Server)   │─────>│   (svn_serve)        │    │
│  │   Port 8080          │      │                      │    │
│  └──────────────────────┘      └──────────┬───────────┘    │
│                                            │                 │
│                                            v                 │
│                          ┌──────────────────────────────┐   │
│                          │  Repository Management       │   │
│                          │  /data_v1/svn/repos/         │   │
│                          │  SASL User Authentication    │   │
│                          └──────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### MAW Backend API (from server/backend/maw.go)

Available functions we can use:

```go
// Repository Management
ListRepos() ([]string, error)                          // List all repos
CreateNewRepo(repoName string) error                   // Create new repo
DeleteRepo(repoName string) error                      // Delete repo (archives first)
GetRepoStructure(repo, user, pass string) (*Node, error) // Browse repo contents

// User Management
ListUsers() ([]string, error)                          // List all users
ListRepoUsers(repoName string) ([]string, error)       // List users for specific repo
AddUserToRepo(repo, user, password string)             // Add user to repo
RemoveUserFromRepo(repo, user string)                  // Remove user from repo
```

### Key System Details

- **Location**: g9 server
- **Repo Path**: `/data_v1/svn/repos/`
- **Archive Path**: `/data_v1/svn/deleted/` (for deleted repos)
- **Auth**: SASL-based (password in `/etc/svn_repos_sasldb`)
- **SVN URL**: `svn://localhost/REPONAME` (on g9)

## Updated Migration Plan

### Implementation Strategy

Since we have direct access to MAW's backend functions, we have two options:

#### Option A: Direct Integration (Recommended)
Import MAW backend as a Go module into ICW:

```go
import "github.com/theballmarcus/icw-maw/server/backend"

// Use MAW functions directly
backend.CreateNewRepo("cp4")
backend.ListRepoUsers("cp3")
backend.AddUserToRepo("cp4", username, password)
```

**Pros**:
- Direct, fast
- No HTTP overhead
- Type-safe
- Easy to test

**Cons**:
- Must run on g9 or have access to repo paths
- Tight coupling

#### Option B: HTTP API Client
Create HTTP client that talks to MAW web service:

```go
// HTTP client for MAW
type MAWClient struct {
    BaseURL string // http://g9:8080
}

func (c *MAWClient) CreateRepo(name string) error
func (c *MAWClient) CopyUsers(from, to string) error
```

**Pros**:
- Can run from anywhere
- Loose coupling
- Service-oriented

**Cons**:
- MAW server must expose REST API
- Network overhead
- Need to handle HTTP errors

### Recommended Approach: Hybrid

1. **For g9 operations**: Use MAW backend directly (Option A)
2. **For client operations**: Can call `icw migrate` from anywhere
3. **Detection**: Auto-detect if running on g9, otherwise require g9 access

## Detailed Migration Workflow

### Phase 1: Repository Setup

```go
func setupTargetRepository(source, target string) error {
    // 1. Check source repo exists
    repos, _ := backend.ListRepos()
    if !contains(repos, source) {
        return fmt.Errorf("source repo %s not found", source)
    }

    // 2. Create target repo
    if err := backend.CreateNewRepo(target); err != nil {
        return err
    }

    // 3. Copy users from source to target
    users, _ := backend.ListRepoUsers(source)
    for _, user := range users {
        // Note: Need to handle passwords
        // Option 1: Copy SASL DB entries directly
        // Option 2: Force password reset for users
        copyUserToRepo(user, source, target)
    }

    // 4. Create standard directory structure
    createRepoStructure(target)

    return nil
}
```

### Phase 2: Component Selection

```
icw migrate --from cp3 --to cp4
```

**Interactive UI**:
```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ ICW Repository Migration                                ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                          ┃
┃ Source: cp3                                              ┃
┃ Target: cp4 (will be created)                            ┃
┃                                                          ┃
┃ ┌─ Component Selection ─────────────────────────────┐   ┃
┃ │                                                    │   ┃
┃ │ Select migration source:                          │   ┃
┃ │   ○ From workspace.config                         │   ┃
┃ │   ● Browse repository                             │   ┃
┃ │                                                    │   ┃
┃ │ Components in cp3:                                │   ┃
┃ │                                                    │   ┃
┃ │ ANALOG (3):                                       │   ┃
┃ │   [x] analog/bias              trunk              │   ┃
┃ │   [x] analog/bandgap_1v2       tags/v2.0          │   ┃
┃ │   [ ] analog/old_opamp         trunk              │   ┃
┃ │                                                    │   ┃
┃ │ DIGITAL (5):                                      │   ┃
┃ │   [x] digital/top              trunk              │   ┃
┃ │   [x] digital/spi_master       tags/v1.5          │   ┃
┃ │   [x] digital/uart             trunk              │   ┃
┃ │   [ ] digital/old_module       trunk              │   ┃
┃ │   [ ] digital/test_only        trunk              │   ┃
┃ │                                                    │   ┃
┃ │ SETUP (2):                                        │   ┃
┃ │   [x] setup/analog             trunk              │   ┃
┃ │   [x] setup/digital            trunk              │   ┃
┃ │                                                    │   ┃
┃ │ [x] Include dependencies automatically            │   ┃
┃ │ [x] Update depend.config references               │   ┃
┃ │                                                    │   ┃
┃ └────────────────────────────────────────────────────┘   ┃
┃                                                          ┃
┃ Selected: 7 components                                   ┃
┃                                                          ┃
┃ [Continue]  [Cancel]                                     ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Phase 3: Migration Strategy Selection

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ Migration Strategy                                       ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                          ┃
┃ Choose migration method for each component:              ┃
┃                                                          ┃
┃ ○ Full History (preserves all revisions, tags)          ┃
┃   - Uses: svn copy with history                         ┃
┃   - Slower, but complete                                ┃
┃                                                          ┃
┃ ● Latest Version (snapshot at specific revision)        ┃
┃   - Uses: svn export + import                           ┃
┃   - Faster, clean start                                 ┃
┃   - Good for released versions                          ┃
┃                                                          ┃
┃ Per-component overrides:                                 ┃
┃ ┌────────────────────────────────────────────────────┐  ┃
┃ │ digital/top            [Full History ▼]            │  ┃
┃ │ digital/spi_master     [Latest Only  ▼]            │  ┃
┃ │ analog/bias            [Latest Only  ▼]            │  ┃
┃ └────────────────────────────────────────────────────┘  ┃
┃                                                          ┃
┃ [Continue]  [Back]                                       ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Phase 4: Confirmation and Execution

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ Confirm Migration                                        ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                          ┃
┃ Ready to migrate from cp3 to cp4                         ┃
┃                                                          ┃
┃ Actions:                                                 ┃
┃  1. Create repository 'cp4' on g9                        ┃
┃  2. Copy 15 users from cp3 to cp4                        ┃
┃  3. Migrate 7 components:                                ┃
┃     - 1 with full history                                ┃
┃     - 6 latest version only                              ┃
┃  4. Update 12 depend.config files                        ┃
┃                                                          ┃
┃ Estimated time: ~5 minutes                               ┃
┃ Estimated size: ~150 MB                                  ┃
┃                                                          ┃
┃ ⚠️  This cannot be easily undone                         ┃
┃                                                          ┃
┃ [Execute Migration]  [Cancel]                            ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

## Technical Implementation

### Module Structure

```
internal/migrate/
├── engine.go          # Main migration engine
├── selector.go        # Component selection UI
├── strategy.go        # Migration strategy logic
├── svn.go            # SVN operations (copy, export, import)
└── maw.go            # MAW backend integration

internal/maw/
├── client.go         # Wrapper for MAW backend
└── types.go          # MAW-related types
```

### Code Structure

```go
package migrate

import (
    "github.com/theballmarcus/icw-maw/server/backend"
    "github.com/jakobsen/icw/internal/svn"
)

type Engine struct {
    SourceRepo   string
    TargetRepo   string
    MAW          *backend.MAWBackend
    SVN          *svn.Client
    Components   []ComponentMigration
    Strategy     MigrationStrategy
}

type ComponentMigration struct {
    Name         string
    Path         string
    Type         string
    Branch       string
    Strategy     MigrationStrategy  // FullHistory or LatestOnly
    Dependencies []*ComponentMigration
}

type MigrationStrategy int

const (
    FullHistory MigrationStrategy = iota
    LatestOnly
    SpecificVersion
)

func (e *Engine) Execute() error {
    // 1. Create target repo
    if err := e.createTargetRepo(); err != nil {
        return err
    }

    // 2. Copy users
    if err := e.copyUsers(); err != nil {
        return err
    }

    // 3. Create directory structure
    if err := e.createStructure(); err != nil {
        return err
    }

    // 4. Migrate components
    for _, comp := range e.Components {
        if err := e.migrateComponent(comp); err != nil {
            return fmt.Errorf("failed to migrate %s: %w", comp.Name, err)
        }
    }

    // 5. Update depend.config files
    if err := e.updateDependencies(); err != nil {
        return err
    }

    return nil
}

func (e *Engine) migrateComponent(comp ComponentMigration) error {
    switch comp.Strategy {
    case FullHistory:
        return e.migrateWithHistory(comp)
    case LatestOnly:
        return e.migrateLatestOnly(comp)
    case SpecificVersion:
        return e.migrateSpecificVersion(comp)
    }
    return nil
}

func (e *Engine) migrateWithHistory(comp ComponentMigration) error {
    // Use SVN copy to preserve history
    sourceURL := fmt.Sprintf("svn://g9/%s/components/%s", e.SourceRepo, comp.Path)
    targetURL := fmt.Sprintf("svn://g9/%s/components/%s", e.TargetRepo, comp.Path)

    // svn copy SOURCE TARGET -m "Migrate from cp3"
    return e.SVN.CopyWithHistory(sourceURL, targetURL, "Migrate from "+e.SourceRepo)
}

func (e *Engine) migrateLatestOnly(comp ComponentMigration) error {
    // 1. Export from source (no .svn)
    sourceURL := fmt.Sprintf("svn://g9/%s/components/%s/%s",
                             e.SourceRepo, comp.Path, comp.Branch)
    tmpDir := "/tmp/icw_migrate_" + comp.Name

    if err := e.SVN.Export(sourceURL, tmpDir); err != nil {
        return err
    }
    defer os.RemoveAll(tmpDir)

    // 2. Import to target
    targetURL := fmt.Sprintf("svn://g9/%s/components/%s/trunk",
                             e.TargetRepo, comp.Path)

    msg := fmt.Sprintf("Import %s from %s %s", comp.Name, e.SourceRepo, comp.Branch)
    return e.SVN.Import(tmpDir, targetURL, msg)
}
```

### SVN Operations Required

Add these methods to `internal/svn/client.go`:

```go
// CopyWithHistory copies a component with full SVN history
func (c *Client) CopyWithHistory(sourceURL, targetURL, message string) error {
    cmd := exec.Command("svn", "copy", sourceURL, targetURL,
                       "-m", message,
                       "--username", c.Username)
    return cmd.Run()
}

// Export exports a component without SVN metadata
func (c *Client) Export(url, destPath string) error {
    cmd := exec.Command("svn", "export", url, destPath,
                       "--username", c.Username)
    return cmd.Run()
}

// Import imports a directory to SVN
func (c *Client) Import(srcPath, url, message string) error {
    cmd := exec.Command("svn", "import", srcPath, url,
                       "-m", message,
                       "--username", c.Username)
    return cmd.Run()
}
```

### User Password Handling

When copying users, we have two options:

**Option 1: Direct SASL DB Copy** (requires root on g9)
```bash
# Copy SASL password entries from cp3 to cp4
sudo sasldblistusers2 -f /etc/svn_repos_sasldb | grep '@cp3:' | \
  sed 's/@cp3:/@cp4:/' | \
  sudo saslpasswd2 -c -f /etc/svn_repos_sasldb -p
```

**Option 2: Reset Passwords** (send email to users)
```go
func (e *Engine) copyUsers() error {
    users, _ := backend.ListRepoUsers(e.SourceRepo)

    for _, user := range users {
        // Generate temporary password
        tempPass := generateTempPassword()

        // Add to target repo
        backend.AddUserToRepo(e.TargetRepo, user, tempPass)

        // Send email with password reset link
        sendPasswordResetEmail(user, e.TargetRepo, tempPass)
    }

    return nil
}
```

**Recommendation**: Option 2 (password reset) is safer and gives users a chance to review access.

## Command Line Interface

### Basic Usage

```bash
# Interactive mode (recommended)
icw migrate

# Command line mode
icw migrate --from cp3 --to cp4

# From workspace
icw migrate --from cp3 --to cp4 --workspace workspace.config

# Dry run
icw migrate --from cp3 --to cp4 --dry-run

# With specific options
icw migrate --from cp3 --to cp4 \
  --full-history \
  --with-deps \
  --components digital/top,digital/spi,analog/bias
```

### Command Flags

```
--from <repo>           Source repository name (required)
--to <repo>             Target repository name (required, will be created)
--workspace <file>      Use workspace.config for component list
--components <list>     Comma-separated list of components to migrate
--full-history          Migrate all components with full history (default: latest only)
--latest-only           Migrate latest version only
--with-deps             Automatically include dependencies
--dry-run               Show what would be migrated without doing it
--reset-passwords       Force password reset for all users (send emails)
--skip-users            Don't copy users (manual setup)
```

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1)
- [ ] Create `internal/migrate` module
- [ ] Integrate MAW backend
- [ ] Basic repo creation and user copying
- [ ] Component listing and selection
- [ ] Simple migration (latest only)

### Phase 2: SVN Operations (Week 1)
- [ ] Implement Export/Import
- [ ] Implement CopyWithHistory
- [ ] Handle directory structures
- [ ] Test with sample components

### Phase 3: Interactive UI (Week 2)
- [ ] Component selection interface
- [ ] Strategy selection per component
- [ ] Progress bars and status updates
- [ ] Error handling and rollback

### Phase 4: Dependency Management (Week 2)
- [ ] Parse depend.config files
- [ ] Detect dependencies
- [ ] Auto-include missing deps
- [ ] Update depend.config references

### Phase 5: Polish and Testing (Week 2)
- [ ] Comprehensive error messages
- [ ] Dry-run mode
- [ ] Transaction log for recovery
- [ ] Full CP3→CP4 migration test
- [ ] Documentation

## Success Criteria

- ✅ Can run `icw migrate --from cp3 --to cp4`
- ✅ Creates cp4 repo on g9
- ✅ Copies all users from cp3
- ✅ Allows selection of components interactively
- ✅ Migrates with full history or latest only
- ✅ Updates all depend.config references
- ✅ Handles dependencies automatically
- ✅ Provides clear progress and errors
- ✅ Can complete CP3→CP4 migration in <10 minutes

## Timeline

- **Week 1**: Core implementation (repo creation, basic migration)
- **Week 2**: UI, dependencies, testing
- **Week 3**: CP3→CP4 production migration

**Total**: ~3 weeks to production-ready

## Next Steps

1. **Review and approve** this plan
2. **Verify g9 access** - Can we run MAW backend functions?
3. **Test MAW integration** - Create test repo to verify
4. **Start Phase 1** - Basic infrastructure
5. **Iterative development** - Show progress after each phase

## Questions for Clarification

1. Do we have access to run MAW backend functions directly on g9?
2. Should we handle user passwords (Option 1 or 2)?
3. Any specific components to exclude from CP3→CP4?
4. Timeline pressure - is 3 weeks acceptable?
5. Who will test the migration before production?
