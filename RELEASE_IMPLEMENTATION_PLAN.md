# ICW Release Command Implementation Plan

## Overview

The release command creates tagged releases of components and their dependencies in the SVN repository. It's a critical feature for managing component versions across the IC design workflow.

## Current Perl Implementation Analysis

### Command Interface

**Usage:**
```bash
icw release -t <tag_name> -m "<message>" [-d]
```

**Flags:**
- `-t <tag>`: Release tag name (required)
- `-m <message>`: Commit message (required)
- `-d`: Dry run mode (optional)

**Execution Context:**
- Must be run from within a component directory
- Reads workspace.config to build component dependency tree
- Automatically detects current component from directory name

### Release Process (Recursive)

The `release_component()` function implements a recursive release process:

#### Step 1: Release Dependencies First (Recursive)
```perl
foreach my $c (@{$components{$comp}{depend}}) {
    release_component($c->{name}, $release, $message, $indent.'  ', $dryrun);
}
```
- Recursively releases all dependencies before the component
- Ensures all sub-components use the same release tag
- Depth-first traversal with indentation for visual feedback

#### Step 2: Check if Already Released
```perl
$svn_cmd_str = "$svn ls $svn_url/$repo/components/$components{$comp}{path}/tags/$release"
```
- Checks if tag already exists
- Skips if already released (idempotent operation)
- Prints status: `component_name release_tag already exists`

#### Step 3: Copy Branch to Release Tag
```perl
svn copy --parents -m <message>
  $svn_url/$repo/components/$components{$comp}{path}/$components{$comp}{branch}
  $svn_url/$repo/components/$components{$comp}{path}/tags/$release
```
- Creates SVN tag from current branch (usually trunk)
- Uses `--parents` to create intermediate directories
- Server-side operation (no local checkout needed)

#### Step 4: Update depend.config in Release
If component has dependencies:

**4a. Generate new depend.config:**
```perl
use component("path/to/dep1", "digital", "tags/release_tag");
use component("path/to/dep2", "analog", "tags/release_tag");
```
- Points all dependencies to the same release tag
- Ensures consistent versions across the release

**4b. Replace depend.config in release tag:**
```perl
svn delete -m <message> .../tags/release/depend.config
svn import -m <message> depend.config-temp .../tags/release/depend.config
```
- Deletes old depend.config from tag
- Imports new one with updated dependency versions
- Creates temporary local file: `depend.config-<comp>-<release>`

## What to Implement in Go

### 1. Release Command Structure

```go
icw release -t <tag> -m "<message>" [-d|--dry-run]
```

**Command handler should:**
- Find workspace root
- Parse workspace.config to build component tree
- Detect current component from directory
- Validate tag name and message
- Call recursive release function
- Support dry-run mode

### 2. Core Release Function

```go
func releaseComponent(
    svnClient *svn.Client,
    comp *component.Component,
    tagName string,
    message string,
    indent string,
    dryRun bool,
) error
```

**Implementation steps:**
1. Release all dependencies recursively (depth-first)
2. Check if tag already exists (skip if yes)
3. SVN copy branch to tags/tagName
4. If has dependencies:
   - Generate new depend.config
   - Delete old depend.config from tag
   - Import new depend.config to tag

### 3. SVN Client Methods Needed

Add to `internal/svn/client.go`:

```go
// Check if a tag exists
func (c *Client) TagExists(componentPath, tagName string) (bool, error)

// Copy branch to tag (SVN copy command)
func (c *Client) CreateTag(componentPath, branch, tagName, message string) error

// Delete file from repository URL
func (c *Client) DeleteFile(fileURL, message string) error

// Import file to repository URL
func (c *Client) ImportFile(localPath, destURL, message string) error
```

### 4. Dependency Tree Building

Reuse existing component parsing:
- Parse workspace.config
- For each component, parse depend.config
- Build component dependency graph
- Handle circular dependency detection (already implemented)

### 5. Output and User Feedback

**Normal mode:**
```
Releasing component: digital/fpga_template
  Releasing dependency: digital/i2c_if
    digital/i2c_if v1.0.0 already exists
  Releasing dependency: digital/uart_if
    ✓ Created tag: digital/uart_if/tags/v1.0.0
    ✓ Updated depend.config
  ✓ Created tag: digital/fpga_template/tags/v1.0.0
  ✓ Updated depend.config

Release complete: v1.0.0
```

**Dry-run mode:**
```
[DRY RUN] Would release: digital/fpga_template
  [DRY RUN] Would release dependency: digital/i2c_if
    svn copy ... i2c_if/trunk i2c_if/tags/v1.0.0
  [DRY RUN] Would release dependency: digital/uart_if
    svn copy ... uart_if/trunk uart_if/tags/v1.0.0
    svn delete ... uart_if/tags/v1.0.0/depend.config
    svn import ... depend.config-uart_if-v1.0.0 ...
  svn copy ... fpga_template/trunk fpga_template/tags/v1.0.0
  svn delete ... fpga_template/tags/v1.0.0/depend.config
  svn import ... depend.config-fpga_template-v1.0.0 ...
```

## Implementation Priority

### Phase 1: Basic Release (No Dependencies)
1. ✓ Parse command-line flags (-t, -m, -d)
2. ✓ Detect current component
3. ✓ Check if tag exists
4. ✓ Create tag from branch
5. ✓ Basic output

### Phase 2: Dependency Handling
1. ✓ Parse workspace.config and depend.config
2. ✓ Build dependency tree
3. ✓ Recursive release
4. ✓ Generate new depend.config
5. ✓ Update depend.config in tag

### Phase 3: Error Handling & Polish
1. ✓ Proper error messages
2. ✓ Dry-run mode
3. ✓ Idempotent operation (skip if exists)
4. ✓ Rollback on failure
5. ✓ Progress indicators

## Key Design Decisions

### 1. Server-Side Operations
- All SVN operations are server-side (no local checkouts)
- Uses `svn copy`, `svn delete`, `svn import` with URLs
- Faster and cleaner than checkout-modify-commit

### 2. Recursive Depth-First
- Dependencies released before parent
- Ensures consistent release versions
- Visual indentation shows hierarchy

### 3. Idempotent
- Check if tag exists before creating
- Safe to re-run if partially failed
- No duplicate tags

### 4. Atomic per Component
- Each component release is one SVN operation
- depend.config update is separate operation
- Clear error boundaries

## Testing Strategy

### Test Cases

1. **Simple component (no dependencies)**
   - Create tag
   - Verify tag exists in SVN
   - Verify content matches branch

2. **Component with dependencies**
   - Release recursively
   - Verify all tags created
   - Verify depend.config updated in parent

3. **Already released**
   - Release same tag twice
   - Should skip gracefully

4. **Dry-run mode**
   - Run with -d flag
   - Verify no tags created
   - Verify output shows what would happen

5. **Component on branch (not trunk)**
   - Release from feature branch
   - Verify tag created from correct branch

## Example Usage

```bash
# Release current component
cd ~/work/asic/workspace/test-cp4/digital/fpga_template
icw release -t v1.0.0 -m "First stable release"

# Dry run first
icw release -t v1.0.1 -m "Bug fix release" -d

# Release from branch
icw release -t v2.0.0-beta -m "Beta release from feature branch"
```

## Related Commands

### icw tag (for ICW itself)
The `icw tag` command (lines 575-623 in Perl) is different:
- Updates version number in icw source
- Commits to Git (not SVN)
- For versioning the ICW tool itself
- Not needed for Go version (use Git tags directly)

### Future: icw dumpdepend
The `dumpdepend` command exports dependency lists for build tools:
- Already implemented as `icw depend-ng` in Go
- Generates Makefile or TCL format
- Lower priority than release

## Migration Notes

**From Perl to Go:**
- Use cobra for CLI (consistent with other commands)
- Reuse existing component/config parsing
- Add SVN methods as needed
- Use color package for output (already used)
- Better error handling than Perl version

**Backward Compatibility:**
- Same command-line interface
- Same SVN repository structure
- Same depend.config format
- Can release components created by Perl version
- Perl version can read releases created by Go version
