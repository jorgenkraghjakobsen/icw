# Setting Up ICW on g9 Server

Complete setup guide for ICW on the g9 server including bash completion and authentication.

## Quick Setup (Complete)

On the g9 server:

```bash
# 1. Install bash completion
cd /path/to/icw
./setup_completion_g9.sh
source ~/.bashrc

# 2. Set up authentication (one time)
icw auth login
# Enter your SVN password when prompted

# 3. Done! Test it
icw list -r cp3
```

## Step-by-Step Setup

### Step 1: Bash Completion

```bash
cd /path/to/icw
./setup_completion_g9.sh
```

This script will:
1. Try system-wide installation (if you have sudo)
2. Fall back to user-local installation in `~/.bash_completion.d/`
3. Automatically update your `~/.bashrc` if needed

Then reload your shell:
```bash
source ~/.bashrc
```

### Step 2: Authentication (NEW! ✨)

**Simple one-command setup:**

```bash
icw auth login
```

Enter your password when prompted. It's securely stored in `~/.icw/credentials` (permissions 0600).

**That's it!** No more environment variables needed.

Or start a new bash session.

## Manual Setup (if script doesn't work)

### Option 1: System-wide (requires sudo)

```bash
sudo cp completions/icw_bashcompletion.sh /usr/local/share/bash-completion/completions/icw
```

Completion will be active in new shells automatically.

### Option 2: User-local (no sudo required)

```bash
# Create completion directory
mkdir -p ~/.bash_completion.d

# Copy completion file
cp completions/icw_bashcompletion.sh ~/.bash_completion.d/icw

# Add to ~/.bashrc (only if not already there)
cat >> ~/.bashrc << 'EOF'

# Load bash completions from ~/.bash_completion.d
if [ -d "$HOME/.bash_completion.d" ]; then
    for completion_file in "$HOME/.bash_completion.d"/*; do
        [ -r "$completion_file" ] && source "$completion_file"
    done
fi
EOF

# Reload bashrc
source ~/.bashrc
```

## Verify Installation

Test completion by typing:
```bash
icw <TAB><TAB>
```

You should see all available commands:
```
add      hdl      list     ls       migrate  st       status   test     tree     update   version
```

## Available Commands on g9

All ICW commands work on g9, including:

**Authentication:**
- `icw auth login` - Store your password (one time setup)
- `icw auth status` - Check authentication status
- `icw auth logout` - Remove stored credentials
- `icw auth test` - Test your credentials

**Repository Listing:**
- `icw list -r <repo>` - List components from any repository
- `icw list -r <repo> -t <type>` - Filter by type (analog/digital/setup/process)

**Migration (g9 only):**
- `icw migrate` - Interactive migration mode
- `icw migrate --create-repo <name>` - Create new repository
- `icw migrate --add-user <user> --to <repo>` - Add user to repository
- `icw migrate --from <src> --to <dst>` - Full migration between repos

**Other Commands:**
- `update`, `status`, `tree`, `hdl`, `add`, `version`, `test`

## Step 3: Environment Setup (Optional)

Set a default repository (optional, you can always use `-r` flag):

```bash
export ICW_REPO=cp3  # or your repo name
```

Add to `~/.bashrc` on g9 for persistence:
```bash
echo 'export ICW_REPO=cp3' >> ~/.bashrc
source ~/.bashrc
```

## Verify Setup

Test everything works:

```bash
# Check authentication status
icw auth status

# List components (should work without password prompt)
icw list -r cp3

# Test tab completion
icw <TAB><TAB>
icw auth <TAB><TAB>
icw list -r cp3 -<TAB>
```

## Troubleshooting

### Completion not working

1. Check if completion is loaded:
   ```bash
   complete -p icw
   ```
   Should output: `complete -F _icw_complete icw`

2. Manually source the completion:
   ```bash
   source ~/.bash_completion.d/icw
   # or
   source /usr/local/share/bash-completion/completions/icw
   ```

3. Check if bash-completion is installed:
   ```bash
   apt list --installed | grep bash-completion  # Debian/Ubuntu
   yum list installed | grep bash-completion    # RHEL/CentOS
   ```

### MAW/migrate errors on other servers

The `migrate` command only works on g9 because it requires access to the MAW backend system. If you run it on other servers (like t14), you'll see:

```
MAW client error: ...
Note: MAW operations must run on g9 server
```

This is expected behavior. Use the other commands (update, status, tree, etc.) on any server.

## Quick Command Reference

| Command | Description | Works on g9 only? |
|---------|-------------|-------------------|
| `icw update` | Sync workspace with repository | No |
| `icw status` / `icw st` | Show status | No |
| `icw tree` | Display dependency tree | No |
| `icw hdl` | Display tree with HDL files | No |
| `icw add <path> <type>` | Add component | No |
| `icw list` / `icw ls` | List components | No |
| `icw test` | Test SVN connection | No |
| `icw version` | Show version | No |
| `icw migrate` | Migrate repositories | **Yes** |
