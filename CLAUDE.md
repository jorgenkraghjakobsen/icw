# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

ICW (IC Workspace Management Tool) is a Perl-based workspace management system for IC design projects. It manages dependencies between analog and digital components stored in Subversion repositories, with support for VHDL, Verilog, and SystemVerilog files.

## Environment Setup

Before using ICW, set the repository environment variable:
```bash
export ICW_REPO=repo_name
```

This variable is required and specifies which Subversion repository to use (default: `icworks_public`).

## Key Commands

### Installation
```bash
make install                 # Install to ~/bin and bash completion
```

### Workspace Operations
```bash
icw update                   # Sync workspace with repository (checkout components)
icw status                   # Show status between workspace and repository
icw st                       # Alias for status
icw tree                     # Display dependency tree with HDL files
icw wipe                     # Reset workspace to clean checkout
```

### Component Management
```bash
icw add <component_path> <repo_target>
# Example: icw add digital/my_module digital
# repo_target format: <analog|digital|setup|process|tools>[/category]

icw depend-ng                # Generate dependency lists for build systems
icw depend-ng -s comp1,comp2 # Stop recursion at specific components
```

### Release Management
```bash
icw release -t <tag_name> -m "<message>"  # Release component with dependencies
icw release -t <tag_name> -m "<message>" -d  # Dry run
icw dumpdepend <component> <revision> <format> [path]
icw dd <component> <revision> <format> [path]  # Alias
# Formats: modelsim, dc, incisiv, list
```

### Version Control
```bash
icw tag                      # Update version and push to Git (dev only)
icw -v                       # Show version
icw -u                       # Update icw to latest from repo
icw -r                       # Return workspace root path
```

## Architecture

### Component System

ICW manages four component types, each mapped to workspace locations:
- **analog**: Analog/mixed-signal IP (stored in `analog/`)
- **digital**: Digital HDL IP (stored in `digital/`)
- **setup**: Setup/configuration scripts (stored in `setup/`)
- **process**: Process technology files (stored in `process_setup/`)

### Configuration Files

**workspace.config**: Workspace-level configuration at repository root. Defines components to check out using:
```perl
use component("path/to/component", "type", "branch/tag");
use ref("/path/to/local/component");  # Reference local components
```

**depend.config**: Component-level dependency configuration. Each component can have a `depend.config` listing its dependencies.

### HDL File Classification

Digital components (VHDL/Verilog/SystemVerilog) are automatically classified:

**RTL files** (synthesis):
- Architecture names: rtl, impl, structural, behavioral
- Files: `*.v`, `*.sv`, `*.svh` (excluding `*_tb.sv`)

**Behavioral files** (simulation):
- Architecture names: testbench, asim, sim
- Files: `*_tb.v`, `*_tb.sv`

**Package files**: VHDL packages and package bodies

### Dependency Resolution

The tool recursively resolves dependencies:
1. Reads workspace.config from workspace root
2. For each component, checks for depend.config
3. Recursively processes dependencies (detects circular dependencies)
4. Checks out components from SVN to appropriate workspace locations
5. Updates component-specific files (cds.lib, local.lib, symlinks)

### Release Process

Releasing a component (icw:764-784):
1. Recursively releases all dependencies first
2. Checks if release tag already exists
3. Copies branch/trunk to tags/<release> in SVN
4. Updates depend.config in the release to point to released sub-components
5. Ensures all dependencies use the same release tag

## Development Notes

### Main Script Structure

The `icw` file (main script) is structured as:
- Global variables and configuration (lines 1-42)
- Component management functions (lines 44-96)
- Config file parsing (lines 100-125)
- Release management (lines 132-181)
- Dependency tree printing (lines 187-231)
- SVN operations (lines 238-264)
- HDL file discovery (lines 283-335)
- Command handlers (lines 566-1265)

### Version Management

The version is embedded in the script:
- `$version_number`: Integer version (line 29)
- `$icw_version`: Full version string with date and message (line 30)
- Update via `icw tag` command which increments version and commits to Git

### SVN Integration

- Default SVN URL: `svn://anyvej11.dk`
- Components stored at: `$svn_url/$repo/components/`
- Uses system SVN client at `/usr/bin/svn`
- Username: Current user from `$ENV{'USER'}`

### Bash Completion

Completion script at `completions/icw_bashcompletion.sh` provides:
- Command name completion
- Directory completion for `add` command
- Component type completion (setup, digital, analog) for `add` target

## File Locations

- **Main executable**: `icw` (Perl script)
- **Installation target**: `~/bin/icw`
- **Bash completion**: `/usr/local/share/bash-completion/completions/icw`
- **Workspace config**: `workspace.config` (at workspace root)
- **Generated files**: `cds.lib`, `local.lib`, `depend.config-*` (workspace root)

## Common Workflows

### Creating a New Workspace
1. Create directory and cd into it
2. Run `icw update` - prompts to create workspace.config
3. Edit workspace.config to add components
4. Run `icw update` again to checkout components

### Adding a Component to Repository
1. Create component in appropriate directory (analog/, digital/, setup/)
2. Run `icw add <component_path> <repo_target>` from workspace root
3. Component is added to SVN trunk and checked out

### Generating Build Dependencies
From within a component directory:
```bash
icw depend-ng > sources.mk         # Makefile format
icw depend-ng -f tcl > sources.tcl # TCL format
```
