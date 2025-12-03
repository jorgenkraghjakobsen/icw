# ICW Quick Start Guide

Get up and running with ICW in 3 simple steps.

## For g9 Server Users

### 1. Install Bash Completion

```bash
cd /path/to/icw
./setup_completion_g9.sh
source ~/.bashrc
```

### 2. Set Up Authentication

```bash
icw auth login
```

Enter your SVN password when prompted. Done!

### 3. Start Using ICW

```bash
# List components from any repository
icw list -r cp3
icw list -r cp4 -t digital

# Create a new repository
icw migrate --create-repo mynewrepo

# Add users
icw migrate --add-user your_name --to mynewrepo
```

That's it! 🎉

---

## Common Tasks

### Listing Components

```bash
# List all components
icw list -r cp3

# List by type
icw list -r cp3 -t digital
icw list -r cp3 -t analog

# Component details
icw list -r cp3 digital/uart -a

# Pattern matching
icw list -r cp3 "digital/*"
```

### Managing Repositories

```bash
# Create repository
icw migrate --create-repo myrepo

# Add users
icw migrate --add-user alice --to myrepo
icw migrate --add-user bob --to myrepo

# View repositories
icw migrate
```

### Workspace Operations

```bash
# Update workspace (checks out components)
icw update

# Show status
icw status

# Show dependency tree
icw tree

# Show dependency tree with HDL files
icw hdl
```

### Authentication Management

```bash
# Check status
icw auth status

# Store new password
icw auth login

# Remove stored password
icw auth logout

# Test credentials
icw auth test
```

---

## Tab Completion Examples

```bash
icw <TAB><TAB>
# Shows all commands

icw list -<TAB>
# Shows: -r --repo -t --type -b --branches -g --tags -a --all

icw auth <TAB>
# Shows: login logout status test

icw migrate --<TAB>
# Shows: --create-repo --from --to --add-user --dry-run
```

---

## First-Time Setup Checklist

- [ ] Install bash completion: `./setup_completion_g9.sh`
- [ ] Reload shell: `source ~/.bashrc`
- [ ] Set up authentication: `icw auth login`
- [ ] Test it works: `icw list -r cp3`
- [ ] Set up sudo for migrate (if needed): See MIGRATE_SETUP.md
- [ ] (Optional) Set default repo: `export ICW_REPO=cp3`

---

## Need Help?

### Documentation

- **AUTH_GUIDE.md** - Complete authentication guide
- **G9_SETUP.md** - Detailed g9 setup
- **MIGRATE_SETUP.md** - Migration command setup
- **COMPLETION_FEATURES.md** - Bash completion features
- **SVN_AUTH_SETUP.md** - Advanced SVN auth options

### Common Issues

**"Can't get username or password"**
```bash
icw auth login
```

**Tab completion not working**
```bash
source ~/.bash_completion.d/icw
complete -p icw  # Verify it's loaded
```

**Migrate command fails with sudo error**
See MIGRATE_SETUP.md for sudo setup

### Quick Commands

```bash
icw --help           # General help
icw auth --help      # Auth help
icw list --help      # List help
icw migrate --help   # Migrate help
```

---

## Environment Variables (Optional)

| Variable | Purpose | Example |
|----------|---------|---------|
| `ICW_REPO` | Default repository | `export ICW_REPO=cp3` |
| `ICW_SVN_URL` | Override SVN URL | `export ICW_SVN_URL=svn://custom` |
| `ICW_SVN_PASSWORD` | Password (for scripts) | `export ICW_SVN_PASSWORD=pass` |

**Note:** Using `icw auth login` is recommended over environment variables!

---

## Summary

**Three steps to get started:**

1. `./setup_completion_g9.sh` - Install completion
2. `icw auth login` - Set up authentication
3. `icw list -r cp3` - Start using ICW!

**Daily usage:**

```bash
icw list -r <repo>                    # List components
icw migrate --create-repo <name>      # Create repos
icw migrate --add-user <user> --to <repo>  # Add users
icw update                            # Update workspace
```

Simple, fast, powerful! 🚀
