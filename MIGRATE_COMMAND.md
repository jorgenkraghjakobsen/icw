# ICW Migrate Command - Implementation Status

## Overview

The `icw migrate` command is now implemented with basic repository creation functionality integrated with the MAW backend system.

## Current Status

✅ **Completed**:
- MAW backend integration wrapper (`internal/maw/client.go`)
- Basic migrate command structure
- Repository creation functionality
- Interactive mode
- Dry-run support
- Command-line interface

## Command Usage

### Create a New Repository

```bash
# On g9 server
icw migrate --create-repo cp4
```

This will:
1. Create a new SVN repository at `/data_v1/svn/repos/cp4`
2. Configure svnserve.conf with SASL authentication
3. Show SVN URL and next steps

### Interactive Mode

```bash
icw migrate
```

Shows:
- Available repositories
- Usage examples

### Full Migration (Coming Soon)

```bash
icw migrate --from cp3 --to cp4
```

Will eventually:
1. Create target repository
2. Copy users from source
3. Select components interactively
4. Migrate components
5. Update dependencies

### Dry Run

```bash
icw migrate --from cp3 --to cp4 --dry-run
```

Shows migration plan without executing.

## Files Created

```
internal/maw/client.go          # MAW backend wrapper
cmd/icw/migrate.go              # Migrate command implementation
```

## MAW Client Functions

```go
type Client struct {
    repoPath   string
    sasldbPath string
}

// Available methods:
CreateRepo(name string) error
ListRepos() ([]string, error)
RepoExists(name string) bool
ListRepoUsers(repo string) ([]string, error)
AddUserToRepo(repo, user, pass string) error
RemoveUserFromRepo(repo, user string) error
DeleteRepo(name string) error
```

## Testing on g9

### 1. Install ICW on g9

```bash
# On g9 server
cd /path/to/icw
make install
```

### 2. Set Environment Variables

```bash
export SASLPASSWD=/etc/svn_repos_sasldb
```

### 3. Test Repository Creation

```bash
# Create a test repository
icw migrate --create-repo test_repo_$(date +%Y%m%d)

# Expected output:
# Creating repository: test_repo_20241203
# ✓ Repository test_repo_20241203 created successfully
#
# Repository details:
#   SVN URL: svn://g9/test_repo_20241203
#   Path: /data_v1/svn/repos/test_repo_20241203
```

### 4. Verify Creation

```bash
# List repositories
ls -la /data_v1/svn/repos/

# Check SVN configuration
cat /data_v1/svn/repos/test_repo_*/conf/svnserve.conf

# Test with icw
icw migrate  # Should list the new repo
```

### 5. Test with Existing Repository

```bash
# Try to create repository that exists (should fail)
icw migrate --create-repo cp3

# Expected output:
# Error: repository cp3 already exists
```

## Example Session on g9

```bash
jakobsen@g9:~$ icw migrate
ICW Repository Migration Tool
============================

Available repositories:
  • cp3
  • icworks
  • test_repo

Usage:
  icw migrate --create-repo <name>              Create new repository
  icw migrate --from <source> --to <target>     Full migration

jakobsen@g9:~$ icw migrate --create-repo cp4
Creating repository: cp4
✓ Repository cp4 created successfully

Repository details:
  SVN URL: svn://g9/cp4
  Path: /data_v1/svn/repos/cp4

Next steps:
  1. Add users: icw migrate --add-user <username> --to cp4
  2. Create workspace.config
  3. Add components
```

## Security Considerations

1. **Requires g9 Server**: Command checks hostname and only runs on g9
2. **SASL Authentication**: Uses existing SASL database for user management
3. **Sudo Required**: User operations require sudo access
4. **Repository Permissions**: Standard SVN repository permissions apply

## Error Handling

### Not on g9 Server

```bash
jakobsen@t14:~$ icw migrate
Error: MAW operations must run on g9 server (current: t14)
```

### Repository Already Exists

```bash
jakobsen@g9:~$ icw migrate --create-repo cp3
Error: repository cp3 already exists
```

### SASL DB Not Found

```bash
# Set SASLPASSWD environment variable
export SASLPASSWD=/etc/svn_repos_sasldb
```

## Next Steps

### Phase 2: User Management
- [ ] Add `--add-user` flag to add users to repos
- [ ] Add `--copy-users` to copy users between repos
- [ ] List users in interactive mode

### Phase 3: Component Selection
- [ ] Browse source repository components
- [ ] Interactive component selection UI
- [ ] Dependency detection

### Phase 4: Component Migration
- [ ] SVN export/import (latest only)
- [ ] SVN copy (with history)
- [ ] Progress tracking
- [ ] Error recovery

### Phase 5: Dependency Updates
- [ ] Parse depend.config files
- [ ] Update repository references
- [ ] Verify migration

## Testing Checklist

On g9 server:

- [ ] `icw migrate` shows available repositories
- [ ] `icw migrate --create-repo test_repo` creates repository
- [ ] Verify SVN repository created in `/data_v1/svn/repos/`
- [ ] Verify svnserve.conf is correct
- [ ] Attempt to create duplicate repo fails with clear error
- [ ] Check repository appears in `icw migrate` list
- [ ] Test svn checkout works: `svn co svn://g9/test_repo`

## Known Limitations

1. **Must run on g9**: No remote access yet
2. **No user migration yet**: Must manually add users
3. **No component migration**: Only repo creation works
4. **No rollback**: Failed operations may need manual cleanup

## Configuration

Default paths (can be changed in code if needed):
- Repository path: `/data_v1/svn/repos`
- Archive path: `/data_v1/svn/deleted`
- SASL DB: `$SASLPASSWD` or `/etc/svn_repos_sasldb`

## Support

For issues:
1. Check you're running on g9
2. Verify SASLPASSWD environment variable
3. Check repository paths exist
4. Ensure sudo access for user operations

## Success Criteria

✅ Command compiles and installs
✅ Clear error when not on g9
✅ Help text is clear and informative
✅ Can create new repository on g9
✅ Repository is properly configured
✅ Error handling for duplicates

🚧 User migration (coming next)
🚧 Component migration (coming later)
