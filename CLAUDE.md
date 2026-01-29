# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

ICW (IC Workspace Management Tool) is a workspace management system for IC design projects. It manages dependencies between analog and digital components stored in Subversion.

- **Design components** (analog/digital/setup/process): Stored in Subversion
- **Software tools**: Stored in Git
- **Language**: Go (uses Cobra for CLI)

## Build & Test Commands

```bash
make build                   # Build the icw binary
make install                 # Install to ~/bin and bash completion
make test                    # Run all tests
make clean                   # Remove built binary
make deps                    # Download Go dependencies

# Run a specific test
go test -v ./internal/config -run TestParseDependConfig

# Run tests for a package
go test -v ./internal/config/...
go test -v ./internal/hdl/...

# Run all tests with coverage
go test -v -cover ./...
```

## Environment Setup

**REQUIRED**: Before using ICW, set the repository environment variable:
```bash
export ICW_REPO=repo_name
```

This variable is **mandatory** and specifies which Subversion repository to use. Alternatively, set `repo` in `workspace.config`.

## Architecture

### Package Structure

```
cmd/icw/
  main.go       # Root Cobra command setup
  commands.go   # All command implementations (update, release, add, tree, etc.)
  auth.go       # Authentication subcommands (login, logout, status, test)

internal/
  svn/          # SVN client - wraps system svn command with auth
  component/    # Component and Workspace types, branch conflict detection
  config/       # Parser for workspace.config and depend.config files
  hdl/          # HDL file discovery (classifies VHDL/Verilog as rtl/behav/package)
  auth/         # Credential storage (~/.icw/credentials)
  version/      # Build version info injected via ldflags
```

### Key Types

- `component.Workspace`: Holds root path, component map, and config path
- `component.Component`: Name, path, type, branch, VCS, dependencies, conflict tracking
- `config.Parser`: Parses workspace.config and depend.config, tracks processed components
- `svn.Client`: SVN operations with authentication (checkout, update, switch, tag, etc.)

### Configuration Files

**workspace.config** (workspace root):
```perl
set repo "repo_name"         # Optional: override ICW_REPO
set svnurl "svn://server"    # Optional: override default SVN URL
use component("path", "type", "branch/tag");
use ref("/path/to/local");   # Reference local components
```

**depend.config** (per-component): Same syntax, declares dependencies.

### Dependency Resolution

1. Parse workspace.config from workspace root
2. For each component, parse depend.config if it exists
3. Recursively process dependencies (circular detection via `processedConfigs` map)
4. Detect branch conflicts when same component declared with different branches
5. Checkout/switch components from SVN

### Release Process (`releaseComponent` in commands.go:1283)

1. Recursively release all dependencies first (depth-first)
2. Check if tag already exists (skip if yes)
3. Create SVN tag: `svn copy` from branch to `tags/<tagName>`
4. Update depend.config in tag to point to released dependencies

### SVN Integration

- Default URL: `svn://anyvej11.dk` (auto-detects `svn://g9` on g9 hostname)
- Components at: `$SVN_URL/$REPO/components/<type>/<name>/{trunk,tags,branches}/`
- Auth: `~/.icw/credentials` or `ICW_SVN_PASSWORD` env var

### HDL Classification (internal/hdl/discover.go)

Digital components auto-classify HDL files:
- **RTL**: Architecture names rtl/impl/structural/behavioral; `*.v`, `*.sv` (not `*_tb.*`)
- **Behavioral**: Architecture names testbench/asim/sim; `*_tb.v`, `*_tb.sv`
- **Package**: VHDL packages and package bodies

## Key Commands Reference

```bash
icw update                   # Sync workspace (checkout/switch components)
icw status | st              # Show workspace status
icw tree                     # Display dependency tree
icw hdl                      # Tree with HDL file listings
icw list [-t type] [-r repo] # List components in repository
icw add <path> <type>        # Add component to SVN
icw release -t <tag> -m <msg> [-d]  # Release with dependencies
icw auth login|logout|status|test   # Credential management
icw test                     # Test SVN connection
```
