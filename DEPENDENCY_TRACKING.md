# Dependency Tracking and Conflict Detection

## Overview

The `icw update` command now automatically follows and checks out dependencies defined in `depend.config` files within each component. It also detects and reports version conflicts when multiple components require the same dependency with different versions/branches.

## Implementation Details

### 1. Dependency Resolution

When `icw update` runs, it:

1. Parses `workspace.config` to get the initial set of components
2. Checks out each component from SVN
3. For each checked-out component, looks for a `depend.config` file
4. Parses any found `depend.config` files to discover dependencies
5. Recursively checks out all dependencies
6. Uses a queue-based approach to process components in breadth-first order

### 2. Version Conflict Detection

The system tracks which component declares each dependency and detects conflicts when:

- Component A requires `digital/spi_master` at `trunk`
- Component B requires `digital/spi_master` at `tags/v2.0`

When a conflict is detected:

1. The update process stops immediately
2. A detailed error message shows:
   - The conflicting component name
   - The first declaration (source component and branch)
   - The conflicting declaration (source component and branch)

### 3. Component Tracking

Each component now tracks:

- **DeclaredBy**: Which component or config file declared this dependency
- **Dependencies**: List of components this component depends on
- **Resolved**: Whether conflicts have been resolved (for future use)

### 4. Circular Dependency Protection

The parser maintains a `processed` map to track which components have already been processed, preventing infinite loops in case of circular dependencies.

## Example Usage

### Success Case

```bash
$ icw update
Workspace root: /home/user/myproject
Found 1 component(s) in workspace.config
Using repository: icworks

  [CHECKOUT] digital/top (trunk)
    Found 2 dependencies
  [CHECKOUT] digital/spi_master (trunk)
  [CHECKOUT] analog/bias (tags/v1.0)

Update complete!
Processed 3 component(s) total
```

### Conflict Case

```bash
$ icw update
Workspace root: /home/user/myproject
Found 2 component(s) in workspace.config
Using repository: icworks

  [CHECKOUT] digital/module1 (trunk)
    Found 1 dependencies
  [CHECKOUT] digital/spi_master (trunk)
  [CHECKOUT] digital/module2 (trunk)
    Found 1 dependencies
    ERROR: dependency conflict: branch mismatch for component 'digital/spi_master'
  First declared by: digital/module1 requesting 'trunk'
  Also declared by: digital/module2 requesting 'tags/v2.0'
Error: version conflict detected: ...
```

## Files Modified

1. **internal/component/types.go**
   - Added `DeclaredBy` field to `Component` struct
   - Enhanced `BranchConflictError` with source tracking
   - Updated `AddComponent()` to detect and report conflicts with source info

2. **internal/config/parser.go**
   - Added `processed` map to `Parser` struct
   - Implemented `ParseDependConfig()` method
   - Updated component parsing to set `DeclaredBy` field

3. **cmd/icw/commands.go**
   - Refactored `runUpdate()` to use queue-based processing
   - Added dependency discovery and recursive checkout
   - Added conflict detection and error reporting

## Testing

Comprehensive tests added:

- **internal/component/types_test.go**
  - Tests component addition with same/different branches
  - Tests conflict error generation
  - Tests `DeclaredBy` tracking for multiple sources

- **internal/config/parser_test.go**
  - Tests parsing `depend.config` files
  - Tests conflict detection during dependency parsing
  - Tests circular dependency protection
  - Tests missing `depend.config` handling

All tests pass:
```bash
$ go test ./...
ok      github.com/jakobsen/icw/internal/component
ok      github.com/jakobsen/icw/internal/config
```

## Future Enhancements

Possible improvements:

1. Add a `--force` flag to allow resolving conflicts by choosing a specific version
2. Implement dependency graph visualization (`icw tree` command)
3. Add support for version ranges/constraints
4. Cache dependency information to speed up subsequent updates
5. Add `icw status` to show which components have conflicts before updating
