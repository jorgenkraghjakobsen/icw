# ICW Repo Migration Plan

## User Story

As a project manager preparing for a new tape-out (CP4), I want to create a new repository based on an existing one (CP3), selectively migrating components with control over revision history, so that I can start the new project with a clean foundation while preserving important work from the previous tape-out.

## Requirements

### 1. Repository Creation
- Create new repository on g9 system
- Copy user access rights from source repository (CP3 → CP4)
- Set up standard directory structure (analog/, digital/, setup/, process_setup/)

### 2. Component Selection
Two methods for selecting components to migrate:

**Option A: From workspace.config**
- Parse existing workspace.config from source repo
- Present list of components with checkboxes
- User selects which to migrate

**Option B: Browse and Select**
- Browse source repository components by type
- Show component hierarchy
- Multi-select components to migrate

### 3. Migration Options

For each component, choose:
- **Full history**: Copy entire SVN history (all revisions, branches, tags)
- **Latest version**: Copy only latest version from specified branch/tag
- **Specific version**: Copy from a specific tag/branch

### 4. Dependency Handling

When migrating components:
- Detect dependencies from depend.config files
- Warn about missing dependencies
- Option to auto-include dependencies
- Update depend.config references to new repo

## Proposed Command Structure

```bash
# Interactive migration wizard
icw migrate

# Or with parameters
icw migrate --from-repo cp3 --to-repo cp4 [options]
```

### Command Options

```
--from-repo <name>        Source repository name
--to-repo <name>          Target repository name (creates if doesn't exist)
--from-workspace <path>   Use workspace.config as source
--interactive            Interactive component selection (default)
--full-history           Migrate with full SVN history
--latest-only            Migrate latest version only
--from-branch <name>     Source branch (default: trunk)
--dry-run               Show what would be migrated without doing it
--with-deps             Automatically include dependencies
```

## Workflow Design

### Phase 1: Repository Setup

```
┌─────────────────────────────────────────┐
│  1. Repository Setup                     │
├─────────────────────────────────────────┤
│  ┌────────────────────────────────┐     │
│  │ Check source repo (cp3)        │     │
│  │ - Verify access                │     │
│  │ - Check g9 system connectivity │     │
│  └────────────┬───────────────────┘     │
│               v                          │
│  ┌────────────────────────────────┐     │
│  │ Create target repo (cp4)       │     │
│  │ - Create repo on g9            │     │
│  │ - Copy user permissions        │     │
│  │ - Create directory structure   │     │
│  └────────────┬───────────────────┘     │
│               v                          │
│  ┌────────────────────────────────┐     │
│  │ Verify repo created            │     │
│  └────────────────────────────────┘     │
└─────────────────────────────────────────┘
```

### Phase 2: Component Selection

```
┌─────────────────────────────────────────┐
│  2. Component Selection                  │
├─────────────────────────────────────────┤
│  ┌────────────────────────────────┐     │
│  │ List available components      │     │
│  │ From: workspace.config or      │     │
│  │       repository browse         │     │
│  └────────────┬───────────────────┘     │
│               v                          │
│  ┌────────────────────────────────┐     │
│  │ Show component list with:      │     │
│  │ [x] analog/bias (trunk)        │     │
│  │ [x] digital/top (tags/v2.0)    │     │
│  │ [ ] digital/old_module         │     │
│  │ [x] setup/digital_env          │     │
│  └────────────┬───────────────────┘     │
│               v                          │
│  ┌────────────────────────────────┐     │
│  │ Check dependencies             │     │
│  │ - Parse depend.config          │     │
│  │ - Warn about missing deps      │     │
│  │ - Offer to include them        │     │
│  └────────────┬───────────────────┘     │
│               v                          │
│  ┌────────────────────────────────┐     │
│  │ Confirm selection              │     │
│  └────────────────────────────────┘     │
└─────────────────────────────────────────┘
```

### Phase 3: Migration Strategy

```
┌─────────────────────────────────────────┐
│  3. Migration Strategy Selection         │
├─────────────────────────────────────────┤
│  For each component, choose:             │
│                                          │
│  ○ Full History Migration               │
│    - Copy all revisions                 │
│    - Copy all branches                  │
│    - Copy all tags                      │
│    - Preserves complete history         │
│    - Slower, larger                     │
│                                          │
│  ○ Latest Version Only                  │
│    - Copy current state only            │
│    - From specified branch/tag          │
│    - No history                         │
│    - Faster, smaller                    │
│                                          │
│  ○ Specific Version                     │
│    - Copy from specific tag/branch      │
│    - No history                         │
│    - Good for released versions         │
└─────────────────────────────────────────┘
```

### Phase 4: Migration Execution

```
┌─────────────────────────────────────────┐
│  4. Execute Migration                    │
├─────────────────────────────────────────┤
│  For each selected component:            │
│                                          │
│  1. Create component structure in        │
│     target repo (trunk/tags/branches)    │
│                                          │
│  2. If full history:                     │
│     - Use svn copy with history          │
│     - Preserve all revisions             │
│                                          │
│  3. If latest only:                      │
│     - Export from source                 │
│     - Import to target                   │
│     - Single revision                    │
│                                          │
│  4. Update depend.config files:          │
│     - Change repo references             │
│     - Maintain version references        │
│                                          │
│  5. Verify migration:                    │
│     - Check files copied                 │
│     - Verify structure                   │
└─────────────────────────────────────────┘
```

## Technical Implementation Plan

### 1. New Module: internal/migrate

```go
package migrate

type MigrationConfig struct {
    SourceRepo      string
    TargetRepo      string
    Components      []ComponentMigration
    FullHistory     bool
    WithDependencies bool
}

type ComponentMigration struct {
    Name           string
    Path           string
    Type           string
    SourceBranch   string
    TargetBranch   string
    IncludeHistory bool
}
```

### 2. G9 System Integration

Need to interface with existing g9 system:
```go
// internal/g9/client.go
type G9Client struct {
    BaseURL string
    Auth    AuthConfig
}

func (c *G9Client) CreateRepo(name string) error
func (c *G9Client) CopyUsers(fromRepo, toRepo string) error
func (c *G9Client) GetRepoUsers(repo string) ([]User, error)
```

### 3. Migration Engine

```go
// internal/migrate/engine.go
type MigrationEngine struct {
    Source      *svn.Client
    Target      *svn.Client
    G9          *g9.Client
    Config      *MigrationConfig
}

func (e *MigrationEngine) Execute() error {
    // 1. Create target repo
    // 2. Setup structure
    // 3. Migrate components
    // 4. Update dependencies
    // 5. Verify
}
```

### 4. Command Implementation

```go
// cmd/icw/migrate.go
var migrateCmd = &cobra.Command{
    Use:   "migrate",
    Short: "Migrate components between repositories",
    Long:  `Interactive tool for migrating components from one repository to another`,
    RunE:  runMigrate,
}

func runMigrate() error {
    // Interactive wizard
    // 1. Get source/target repos
    // 2. Select components
    // 3. Choose migration strategy
    // 4. Confirm and execute
}
```

## Data Flow

```
Source Repo (CP3)                Target Repo (CP4)
┌──────────────┐                ┌──────────────┐
│              │                │              │
│ Components   │                │ New Repo     │
│ - List       │                │ - Structure  │
│ - History    │                │ - Users      │
│ - Metadata   │                │              │
└──────┬───────┘                └──────┬───────┘
       │                               │
       │  ┌─────────────────────┐     │
       └─>│ Migration Engine    │─────┘
          │                     │
          │ - Parse workspace   │
          │ - Select components │
          │ - Copy data         │
          │ - Update configs    │
          └─────────────────────┘
```

## SVN Operations Required

### Full History Migration
```bash
# SVN to SVN copy with history
svn copy SOURCE_URL TARGET_URL -m "Migrate component"
```

### Latest Version Migration
```bash
# Export from source (no .svn)
svn export SOURCE_URL local_temp

# Import to target
svn import local_temp TARGET_URL -m "Import component"
```

## User Interface Mock-ups

### Interactive Selection

```
ICW Repository Migration Tool
==============================

Source repository: cp3
Target repository: cp4 (will be created)

Loading components from cp3...

Select components to migrate:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

ANALOG COMPONENTS:
  [x] analog/bias              (trunk)        [Full History]
  [x] analog/bandgap_1v2       (tags/v2.0)    [Latest Only]
  [ ] analog/old_opamp         (trunk)

DIGITAL COMPONENTS:
  [x] digital/top              (trunk)        [Full History]
  [x] digital/spi_master       (tags/v1.5)    [Latest Only]
  [ ] digital/deprecated_uart  (trunk)

SETUP COMPONENTS:
  [x] setup/digital_env        (trunk)        [Latest Only]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Selected: 5 components
Options:
  [x] Include dependencies automatically
  [x] Update depend.config references
  [ ] Dry run (show what would happen)

Continue? [Y/n]
```

### Progress Display

```
Migrating components from cp3 to cp4...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[✓] Created repository cp4
[✓] Copied user permissions (15 users)
[✓] Created directory structure

Migrating components:
[✓] analog/bias              (with full history)
[✓] analog/bandgap_1v2       (latest from tags/v2.0)
[→] digital/top              (copying history...)
[ ] digital/spi_master
[ ] setup/digital_env

Progress: 2/5 components (40%)
```

## Considerations and Edge Cases

### 1. Large Repositories
- **Issue**: Full history migration can be slow for large repos
- **Solution**: Show progress, allow background operation, provide time estimates

### 2. Dependency Conflicts
- **Issue**: Component A depends on Component B, but B is not selected
- **Solution**: Warn user, offer to include B, or allow ignoring

### 3. Branch/Tag Mismatches
- **Issue**: Source uses tags/v1.0, target might want trunk
- **Solution**: Allow branch/tag mapping during selection

### 4. Naming Conflicts
- **Issue**: Component already exists in target
- **Solution**: Warn, offer to skip, overwrite, or rename

### 5. Network Failures
- **Issue**: Migration interrupted mid-way
- **Solution**: Transaction log, ability to resume, rollback on failure

### 6. Permission Issues
- **Issue**: User lacks permission on source or target
- **Solution**: Validate permissions before starting, clear error messages

### 7. Repository Schema Differences
- **Issue**: Source and target repos have different structures
- **Solution**: Normalize paths during migration

## Testing Strategy

### Unit Tests
- Component selection logic
- Dependency resolution
- Config file updates

### Integration Tests
- Create test repos
- Migrate sample components
- Verify results

### Manual Testing
- Full workflow with real data
- Edge cases
- User experience

## Rollout Plan

### Phase 1: Core Migration
- Implement basic migration (latest version only)
- Simple component selection
- No dependency handling

### Phase 2: Full History
- Add full history migration option
- Test with large components

### Phase 3: Dependencies
- Automatic dependency detection
- Dependency conflict resolution
- Config file updates

### Phase 4: UI/UX
- Interactive selection
- Progress display
- Better error messages

### Phase 5: Advanced Features
- Workspace.config import
- Batch operations
- Migration templates

## Open Questions

1. **G9 System API**: What's the interface to the g9 system for creating repos and managing users?
   - HTTP API?
   - Command-line tool?
   - Direct database access?

2. **User Authentication**: How do we authenticate with g9 and SVN repositories?
   - Same credentials?
   - Service account?
   - Token-based?

3. **Repository Naming**: Any conventions for repository names?
   - cp3, cp4 pattern?
   - Project codes?
   - Date-based?

4. **Component Versioning**: When migrating, should we:
   - Create new tags in target?
   - Keep original tag names?
   - Reset version numbers?

5. **Workspace.config Update**: After migration, should we:
   - Automatically create workspace.config in target?
   - Update existing workspace.config files?
   - Leave it to user?

## Success Criteria

✅ Can create new repository on g9
✅ Can copy user permissions
✅ Can select components interactively
✅ Can migrate with full history
✅ Can migrate latest version only
✅ Can detect and handle dependencies
✅ Can update depend.config references
✅ Provides clear progress and error messages
✅ Can recover from failures
✅ Completes CP3→CP4 migration successfully

## Timeline Estimate

- **Phase 1** (Core Migration): 2-3 days
- **Phase 2** (Full History): 1 day
- **Phase 3** (Dependencies): 2 days
- **Phase 4** (UI/UX): 2 days
- **Phase 5** (Advanced): 2-3 days
- **Testing & Refinement**: 2-3 days

**Total**: ~2 weeks for full implementation

## Next Steps

1. **Review this plan** - Discuss and refine requirements
2. **Answer open questions** - Clarify g9 system integration
3. **Prioritize phases** - Decide which features are MVP
4. **Start implementation** - Begin with Phase 1
