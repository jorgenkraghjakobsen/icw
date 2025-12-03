# SVN Authentication Setup for ICW

ICW now supports SVN password authentication via environment variables for seamless operation on g9.

## Quick Setup (g9 Server)

Add to your `~/.bashrc` on g9:

```bash
# ICW SVN Authentication
export ICW_SVN_PASSWORD=your_password_here
```

Then reload:
```bash
source ~/.bashrc
```

## Environment Variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| `ICW_REPO` | Repository name | **Yes** | `cp3`, `cp4`, `mynewrepo` |
| `ICW_SVN_URL` | SVN server URL | No* | `svn://g9` (auto-detected on g9) |
| `ICW_SVN_PASSWORD` | SVN password | No** | `your_password` |
| `USER` | Username | No*** | `jakobsen` (auto-detected) |

\* Auto-detects `svn://g9` when hostname is "g9", otherwise uses `svn://anyvej11.dk`
\*\* Required for SASL-authenticated repositories (like on g9)
\*\*\* Uses current `$USER` if not specified

## Usage Examples

### Option 1: Set Once in ~/.bashrc (Recommended)

```bash
# Add to ~/.bashrc on g9
export ICW_SVN_PASSWORD=gettonui9

# Then use ICW normally
icw list -r cp3
icw list -r cp4 -t digital
icw migrate --from cp3 --to cp4
```

### Option 2: Set Per-Session

```bash
# Set for current session
export ICW_SVN_PASSWORD=gettonui9

# Use ICW
icw list -r cp3 -t analog
```

### Option 3: Set Per-Command

```bash
# One-time command
ICW_SVN_PASSWORD=gettonui9 icw list -r cp3

# Multiple commands
ICW_SVN_PASSWORD=gettonui9 icw list -r cp3 -t digital
ICW_SVN_PASSWORD=gettonui9 icw list -r cp4 -t analog
```

## Complete g9 Setup Example

```bash
# Edit ~/.bashrc
nano ~/.bashrc

# Add these lines at the end:
export ICW_SVN_PASSWORD=gettonui9

# Save and exit (Ctrl+X, Y, Enter)

# Reload bashrc
source ~/.bashrc

# Test it works
icw list -r cp3
icw list -r cp4 -t digital
```

## Security Considerations

### Is This Secure?

**For g9 local development**: Storing the password in `~/.bashrc` is acceptable because:
- G9 is an internal development server
- Your home directory is protected by file permissions
- The password is used for read-only operations on team repositories
- Alternative (SSH keys) requires server configuration changes

**For production or sensitive systems**: Consider:
- Using SVN with SSH keys instead of password
- Configuring SVN credential helpers
- Not storing passwords in environment variables

### File Permissions

Ensure your bashrc is not readable by others:
```bash
chmod 600 ~/.bashrc
ls -la ~/.bashrc
# Should show: -rw------- (only you can read/write)
```

## Auto-Detection Features

### Hostname Detection

ICW automatically detects when running on g9:

| Hostname | SVN URL Used | Notes |
|----------|--------------|-------|
| `g9` | `svn://g9` | Local svnserve |
| Other | `svn://anyvej11.dk` | Remote server |

You can override with `ICW_SVN_URL`:
```bash
export ICW_SVN_URL=svn://custom-server.com
```

### Repository Listing

The `--repo` flag allows listing from any repository:

```bash
# List components from different repos
icw list -r cp3              # Lists from cp3
icw list -r cp4 -t digital   # Lists digital components from cp4
icw list -r mynewrepo        # Lists from your new repo
```

## Troubleshooting

### "Can't get username or password"

**Problem**: SVN can't authenticate

**Solution**: Set the password:
```bash
export ICW_SVN_PASSWORD=your_password
```

### "Unable to connect to repository"

**Problem**: Wrong URL or repository doesn't exist

**Check**:
```bash
# Verify repository exists
ls -la /data_v1/svn/repos/

# Verify hostname detection
hostname  # Should show "g9"

# Test SVN directly
svn list svn://g9/cp3 --username $USER --password $ICW_SVN_PASSWORD
```

### Password Not Working

**Verify password is set**:
```bash
echo $ICW_SVN_PASSWORD
```

Should show your password. If empty:
```bash
export ICW_SVN_PASSWORD=gettonui9
```

### Still Getting Errors

**Test authentication manually**:
```bash
# Try listing with explicit password
svn list svn://g9/cp3/components/digital --username jakobsen --password gettonui9
```

If this works but ICW doesn't:
1. Check ICW is using the right binary: `which icw` (should be `~/bin/icw`)
2. Check password is set: `echo $ICW_SVN_PASSWORD`
3. Re-source bashrc: `source ~/.bashrc`

## Migration Workflow with Authentication

Complete workflow for migrating between repositories:

```bash
# 1. Set up authentication (once)
echo 'export ICW_SVN_PASSWORD=gettonui9' >> ~/.bashrc
source ~/.bashrc

# 2. Explore source repository
icw list -r cp3
icw list -r cp3 -t digital
icw list -r cp3 digital/uart -a

# 3. Create new repository
icw migrate --create-repo cp5

# 4. Add users
icw migrate --add-user jakobsen --to cp5

# 5. Check target (should be empty)
icw list -r cp5

# 6. Perform migration (when implemented)
icw migrate --from cp3 --to cp5 --dry-run

# 7. Verify
icw list -r cp5
```

## Advanced: Multiple Passwords for Different Repos

If different repositories have different passwords (unusual), use per-command approach:

```bash
# Repo1 with password1
ICW_SVN_PASSWORD=pass1 icw list -r repo1

# Repo2 with password2
ICW_SVN_PASSWORD=pass2 icw list -r repo2
```

Or create wrapper scripts:

```bash
# ~/bin/icw-cp3
#!/bin/bash
ICW_SVN_PASSWORD=pass_for_cp3 ~/bin/icw "$@"

# ~/bin/icw-cp4
#!/bin/bash
ICW_SVN_PASSWORD=pass_for_cp4 ~/bin/icw "$@"
```

## Summary

**Quick start for g9:**
```bash
# One-time setup
echo 'export ICW_SVN_PASSWORD=gettonui9' >> ~/.bashrc
source ~/.bashrc

# Now use ICW freely
icw list -r cp3
icw list -r cp4 -t digital
icw migrate --from cp3 --to cp4
```

That's it! 🎉
