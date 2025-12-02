# ICW Tree and HDL Commands

## Overview

ICW provides two commands for visualizing component dependencies:

- **`icw tree`**: Displays a clean dependency tree showing component relationships from config files
- **`icw hdl`**: Displays the dependency tree with detailed HDL file listings

## icw tree

### Description

Shows the component dependency structure based on `workspace.config` and `depend.config` files. This gives a clean view of component relationships without file-level details.

### Usage

```bash
icw tree
```

### Output Format

```
Dependency tree for workspace

digital/top (trunk) [digital]
  digital/spi_master (trunk) [digital]
  analog/bias (tags/v1.0) [analog]
  setup/digital_env (trunk) [setup]
```

Each line shows:
- Component name
- Branch/tag
- Component type in brackets

### When to Use

- Quick overview of component dependencies
- Understanding workspace structure
- Checking which components depend on what
- Verifying dependency configuration

## icw hdl

### Description

Shows the complete dependency tree with detailed HDL file listings for digital components. Files are categorized by type (package, rtl, behav).

### Usage

```bash
icw hdl
```

### Output Format

```
Dependency tree with HDL files

digital/top (digital/top), trunk
  - rtl: .../digital/top/top.v .../digital/top/fsm.v
  - behav: .../digital/top/top_tb.v
  digital/spi_master (digital/spi_master), trunk
    - rtl: .../digital/spi_master/spi_master.v
    - behav: .../digital/spi_master/spi_master_tb.v
  analog/bias (analog/bias), tags/v1.0
```

Each component shows:
- Component name and path
- Branch/tag
- HDL files categorized as:
  - `package`: VHDL package files
  - `rtl`: Synthesizable RTL files
  - `behav`: Behavioral/testbench files

### When to Use

- Generating build file lists
- Verifying which HDL files are in each component
- Understanding what will be compiled
- Checking file organization

## Related Commands

- **icw update**: Checks out components and dependencies
- **icw status** (alias: `st`): Shows workspace modification status
- **icw list**: Lists available components in repository
- **icw depend-ng**: Generates dependency lists for build tools (not yet implemented)
