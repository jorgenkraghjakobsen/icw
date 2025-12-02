# ICW TODO List

## Future Enhancements

### HDL File Classification

#### Gate-Level Netlist Handling

**Issue**: Files like `*_gate.v` are generated gate-level netlists that should be:
- Included in certain build flows (e.g., post-synthesis simulation)
- Excluded from synthesis/compilation of the complete system

**Example**: `dig_top_cp3_gate.v`

**Proposed Solution Options**:

1. **Naming Convention Detection**
   - Detect `*_gate.v` pattern
   - Add new file type: `gate` or `netlist`
   - Allow users to specify which file types to include in different flows

2. **Explicit Exclusion File**
   - Create `.icwignore` or similar in component directories
   - List files to exclude from different build flows
   - Similar to `.gitignore` syntax

3. **File Classification Markers**
   - Add comments in HDL files: `// ICW: exclude-from-synthesis`
   - Parse and respect these markers during file discovery

4. **Build Flow Configuration**
   - Extend `depend.config` with flow-specific file lists
   - Example:
     ```
     exclude synthesis "dig_top_cp3_gate.v"
     include post-synth-sim "dig_top_cp3_gate.v"
     ```

**Affected Commands**:
- `icw tree` - Could add flags like `--flow=synthesis` to filter files
- `icw depend-ng` - Should respect flow-specific exclusions

**Priority**: Medium (works fine for now, but needed for production use)

**Status**: Deferred - Document and implement when needed

---

## Completed Features

- ✅ Dependency tracking and resolution
- ✅ Version conflict detection
- ✅ `icw tree` - Display clean dependency tree from config files
- ✅ `icw hdl` - Display dependency tree with HDL file listing
- ✅ `icw status` (alias: `st`) - Show workspace status vs repository
- ✅ HDL file classification (RTL, behavioral, packages)
- ✅ Recursive dependency checkout

## Not Yet Implemented

### High Priority
- `icw depend-ng` - Generate dependency lists for build systems

### Medium Priority
- `icw add` - Add components to repository
- `icw release` - Release component with dependencies
- `icw dumpdepend` - Dump dependencies for specific tools
- Gate-level netlist handling (see above)

### Low Priority
- `icw wipe` - Reset workspace to clean checkout
- Git support for tools components (partial implementation exists)
- Build flow configuration system
