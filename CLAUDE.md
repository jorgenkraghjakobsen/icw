# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

ICW (IC Workspace Management Tool) is a workspace management system for IC design projects. It manages dependencies between analog and digital components. The tool is being migrated from Perl to Go for better performance and maintainability.

- **Design components** (analog/digital/setup/process): Stored in Subversion
- **Software tools**: Stored in Git
- **Language**: Go (migrating from Perl)

## Environment Setup

**REQUIRED**: Before using ICW, set the repository environment variable:
```bash
export ICW_REPO=repo_name
```

This variable is **mandatory** and specifies which Subversion repository to use. ICW will refuse to run without it.

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

### Authentication
```bash
icw auth login               # Store SVN password securely
icw auth status              # Check authentication status
icw auth logout              # Remove stored credentials
icw auth test                # Test SVN connection
```

### Release Management
```bash
icw release -t <tag_name> -m "<message>"     # Release component with dependencies
icw release -t <tag_name> -m "<message>" -d  # Dry run (preview)

# Release creates tags for component and all dependencies
# Updates depend.config in tags to point to released versions
# Run from within the component directory

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

### Architecture (Go Implementation)

The Go implementation is organized as:
- **cmd/icw/**: Command-line interface and command handlers
  - `main.go`: Root command and CLI setup
  - `commands.go`: Command implementations (update, release, add, etc.)
  - `auth.go`: Authentication command handlers
  - `migrate.go`: Migration utilities

- **internal/**: Core packages
  - `svn/`: SVN client with authentication
  - `component/`: Component and workspace management
  - `config/`: Configuration file parsing
  - `hdl/`: HDL file discovery and classification
  - `auth/`: Secure credential storage
  - `version/`: Version information

### Key Features Implemented

✅ **Authentication System** (`icw auth`)
- Secure password storage in `~/.icw/credentials`
- Environment variable support (`ICW_SVN_PASSWORD`)
- Auto-authentication for all SVN operations

✅ **Release Management** (`icw release`)
- Recursive dependency release
- Automatic depend.config updates in tags
- Dry-run mode for preview
- Idempotent operations

✅ **Component Addition** (`icw add`)
- Creates SVN directory structure
- Imports local component to trunk
- Converts to SVN working copy
- Supports categories (e.g., digital/muxes)

✅ **Workspace Updates** (`icw update`)
- Automatic branch/tag switching
- Dependency resolution
- Conflict detection

✅ **Enhanced Tree Display** (`icw tree`)
- Aligned column formatting
- Branch/tag display
- Component type indicators

### SVN Integration

- Default SVN URL: `svn://anyvej11.dk` (auto-detects `svn://g9` on g9 server)
- Components stored at: `$SVN_URL/$REPO/components/`
- Uses system SVN client at `/usr/bin/svn`
- Username: Current user from `$USER` environment variable
- Password: From `~/.icw/credentials` or `ICW_SVN_PASSWORD`

### Bash Completion

Completion script at `completions/icw_bashcompletion.sh` provides:
- Command name completion
- Flag completion
- Directory completion for `add` command
- Component type completion

## File Locations

- **Main executable**: `icw` (Go binary)
- **Installation target**: `~/bin/icw`
- **Bash completion**: `/usr/local/share/bash-completion/completions/icw`
- **Credentials**: `~/.icw/credentials` (mode 0600)
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
