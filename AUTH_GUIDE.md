# ICW Authentication Guide

Simple, secure authentication for ICW - no more environment variables!

## Quick Start (Recommended)

**One command to set up authentication:**

```bash
icw auth login
```

That's it! Enter your password once, and you're done. Your password is securely stored in `~/.icw/credentials` (readable only by you).

## Using ICW with Authentication

### First Time Setup

```bash
# 1. Store your password (one time only)
icw auth login
# Enter SVN password: ********

# 2. Use ICW normally - no password needed!
icw list -r cp3
icw list -r cp4 -t digital
icw migrate --create-repo mynewrepo
```

### Daily Usage

After running `icw auth login` once, just use ICW commands normally:

```bash
icw list -r cp3              # Works automatically
icw list -r cp4 -t digital   # No password needed
icw migrate --create-repo myrepo
icw update
```

## Auth Commands

| Command | Description | Example |
|---------|-------------|---------|
| `icw auth login` | Store your password | First-time setup |
| `icw auth status` | Check authentication status | See what's configured |
| `icw auth logout` | Remove stored password | Clear credentials |
| `icw auth test` | Test your credentials | Verify setup |

## Examples

### Store Password

```bash
$ icw auth login
ICW Authentication Setup
========================

This will store your SVN password in ~/.icw/credentials
The file will be created with permissions 0600 (readable only by you)

Enter SVN password: ********

✓ Credentials saved successfully!

Your password is stored in: /home/jakobsen/.icw/credentials
You can now use ICW commands without entering your password.

Try it:
  icw list -r cp3
  icw migrate --create-repo myrepo
```

### Check Status

```bash
$ icw auth status
Authentication Status
====================

○ ICW_SVN_PASSWORD not set
✓ Credentials stored in: /home/jakobsen/.icw/credentials
✓ File permissions: 0600 (secure)

Username: jakobsen (from $USER)
SVN URL: svn://g9 (auto-detected)
```

### Remove Credentials

```bash
$ icw auth logout
✓ Credentials removed successfully

You'll need to run 'icw auth login' to store credentials again
Or set ICW_SVN_PASSWORD environment variable for each command
```

## How It Works

### Priority Order

ICW looks for your password in this order:

1. **ICW_SVN_PASSWORD environment variable** (for scripts/automation)
2. **Stored credentials** (`~/.icw/credentials`)
3. **Prompt** (for interactive commands, if neither above is set)

### Security

- Credentials are stored in `~/.icw/credentials` with **0600 permissions** (only you can read)
- File is in your home directory (`~/.icw/`)
- Password is stored in plain text locally (same security as SSH keys)
- Only used for local development on g9 server

### File Location

```
~/.icw/
└── credentials    # Your SVN password (0600 permissions)
```

## Comparison: Old vs New

### ❌ Old Way (Environment Variable)

```bash
# Had to set this in ~/.bashrc
export ICW_SVN_PASSWORD=gettonui9

# Or set every time
ICW_SVN_PASSWORD=gettonui9 icw list -r cp3
```

Problems:
- Password visible in `~/.bashrc`
- Shows up in `ps` output
- Easy to forget to set
- Not beginner-friendly

### ✅ New Way (`icw auth`)

```bash
# One-time setup
icw auth login
# Enter password once

# Then just use ICW
icw list -r cp3
icw migrate --create-repo myrepo
```

Benefits:
- Simple, user-friendly
- Secure file permissions
- One-time setup
- Just works™

## Advanced Usage

### Scripting / Automation

For scripts, you can still use environment variables:

```bash
#!/bin/bash
export ICW_SVN_PASSWORD=secret
icw list -r cp3
icw list -r cp4
```

Or use stored credentials (recommended):

```bash
#!/bin/bash
# Just use icw - credentials are already stored
icw list -r cp3
icw list -r cp4
```

### Multiple Machines

You need to run `icw auth login` on each machine where you use ICW:

```bash
# On g9
icw auth login

# On another server
icw auth login
```

### Different Passwords for Different Repos

Currently, ICW stores one password for all repositories. If you need different passwords:

```bash
# Use environment variable per command
ICW_SVN_PASSWORD=pass1 icw list -r repo1
ICW_SVN_PASSWORD=pass2 icw list -r repo2
```

## Troubleshooting

### "failed to get credentials" error

**Solution**: Run `icw auth login` to store your password

### Check if credentials are stored

```bash
icw auth status
```

### Password not working

1. Check status: `icw auth status`
2. Remove old credentials: `icw auth logout`
3. Store new credentials: `icw auth login`
4. Test: `icw list -r cp3`

### Permission denied on credentials file

```bash
chmod 600 ~/.icw/credentials
```

### Want to see the stored password

```bash
cat ~/.icw/credentials
```

Note: Your password is stored in plain text (same as SSH private keys)

## Migration from Environment Variable

If you're currently using `ICW_SVN_PASSWORD` in `~/.bashrc`:

### Step 1: Store credentials

```bash
icw auth login
# Enter your password
```

### Step 2: Test it works

```bash
icw list -r cp3
```

### Step 3: Remove from ~/.bashrc (optional)

```bash
# Edit ~/.bashrc and remove this line:
# export ICW_SVN_PASSWORD=gettonui9

# Reload
source ~/.bashrc
```

The environment variable method still works, but `icw auth` is easier!

## Tab Completion

```bash
icw auth <TAB>
# Shows: login logout status test

icw auth login <TAB>
# Ready for command
```

## Summary

**TL;DR**: Run `icw auth login` once, then use ICW normally. Your password is securely stored and automatically used.

```bash
# Setup (once)
icw auth login

# Use (always)
icw list -r cp3
icw migrate --create-repo myrepo
```

Simple. Secure. User-friendly. 🎉
