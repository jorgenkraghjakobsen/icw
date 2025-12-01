# ICW Version Management

## Current Version System

ICW uses **Git tags** for semantic versioning. The version is automatically injected at build time using Go's `-ldflags`.

### Version Format

- **Semantic versioning**: `vMAJOR.MINOR.PATCH` (e.g., `v2.0.0`)
- **Development builds**: Git commit hash with `-dirty` suffix if uncommitted changes exist
- **Build info**: Includes commit hash, build date, Go version, and platform

### Checking Version

```bash
# Short version (for scripts)
icw --version
# Output: icw version v2.0.0 (5b734d6)

# Detailed version (human-readable)
icw version
# Output:
# icw version v2.0.0 (linux/amd64)
# Built: 2025-12-01T11:40:53Z
# Commit: 5b734d6
# Go: go1.21.8
```

## Creating a New Release

### 1. Commit all changes
```bash
git add .
git commit -m "Description of changes"
```

### 2. Create annotated tag
```bash
git tag -a vX.Y.Z -m "Release vX.Y.Z

Features:
- Feature 1
- Feature 2

Bug fixes:
- Fix 1
"
```

### 3. Push tag to remote
```bash
git push origin vX.Y.Z
```

### 4. Build release binary
```bash
make build
```

The version will be automatically set to `vX.Y.Z` from the git tag.

## Version Information at Build Time

The Makefile injects these variables:

- `Version`: Git tag or commit hash (`git describe --tags`)
- `Commit`: Short commit hash (`git rev-parse --short HEAD`)
- `BuildDate`: UTC timestamp (`YYYY-MM-DDTHH:MM:SSZ`)
- `GoVersion`: Go compiler version (from `runtime.Version()`)

## Migration from Perl

The old Perl version hardcoded version numbers in the source:
```perl
my $version_number = 127;
my $icw_version = '127 : Tue May 21 04:30:54 PM CEST 2024 : Added date to tag';
```

The new Go version uses build-time injection, which:
- ✅ Eliminates version number commits
- ✅ Automatically tracks git state
- ✅ Provides more detailed build information
- ✅ Supports semantic versioning with git tags
- ✅ Shows `-dirty` flag for uncommitted changes

## Examples

### Clean tagged build
```bash
$ git tag -a v2.0.0 -m "Release"
$ make build
Building icw v2.0.0 (5b734d6)...

$ ./icw version
icw version v2.0.0 (linux/amd64)
Built: 2025-12-01T11:40:53Z
Commit: 5b734d6
```

### Development build
```bash
$ make build
Building icw v2.0.0-dirty (5b734d6)...

$ ./icw version
icw version v2.0.0-dirty (linux/amd64)  # -dirty = uncommitted changes
```

### Pre-release/Development commits after tag
```bash
$ git commit -m "New feature"
$ make build
Building icw v2.0.0-1-g6c8a9f2 (6c8a9f2)...
#            ^     ^  ^
#            |     |  commit hash
#            |     commits since v2.0.0
#            last tag
```

This follows standard Git describe format and makes it easy to track exactly which code was built.
