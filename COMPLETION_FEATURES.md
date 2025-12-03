# ICW Bash Completion Features

Complete bash completion support for all ICW commands, flags, and options.

## Installation

### Local Installation
```bash
make install
```

### Manual Installation (any server)
```bash
# User-local (recommended)
./setup_completion_g9.sh

# Or manually
mkdir -p ~/.bash_completion.d
cp completions/icw_bashcompletion.sh ~/.bash_completion.d/icw
source ~/.bashrc
```

## Completion Features

### Command Completion
Type `icw <TAB><TAB>` to see all available commands:
```
add         help        ls          status      tree        version
completion  hdl         migrate     st          update
```

### Flag Completion

#### List Command (`icw list` / `icw ls`)
- `icw list -<TAB>` shows:
  - `-t, --type` - Filter by component type
  - `-b, --branches` - Show branches for component
  - `-g, --tags` - Show tags for component
  - `-a, --all` - Show all details
  - `-h, --help` - Help

- `icw list --type <TAB>` shows:
  - `analog` `digital` `setup` `process`

**Examples:**
```bash
icw list -<TAB>              # Shows all flags
icw list --type <TAB>        # Shows: analog digital setup process
icw list --branches          # Auto-completes flag name
icw ls -a                    # Works with alias too
```

#### Migrate Command (`icw migrate`)
- `icw migrate -<TAB>` shows:
  - `--create-repo` - Create a new repository
  - `--from` - Source repository
  - `--to` - Target repository
  - `--dry-run` - Show what would be done
  - `-h, --help` - Help

**Examples:**
```bash
icw migrate --<TAB>          # Shows all flags
icw migrate --create-repo <TAB>  # Type repo name
icw migrate --from cp3 --to <TAB>  # Type target repo
icw migrate --dry-run        # Auto-completes
```

#### Add Command (`icw add`)
- `icw add <TAB>` - Shows directories (first argument)
- `icw add digital/ <TAB>` - Shows component types:
  - `analog` `digital` `setup` `process`

**Examples:**
```bash
icw add <TAB>                # Shows directories
icw add digital/<TAB>        # Shows subdirectories
icw add digital/mymod <TAB>  # Shows: analog digital setup process
```

#### Other Commands
All commands support:
- `-h, --help` - Show help
- Global `-v, --version` - Show version (at root level)

**Examples:**
```bash
icw update -<TAB>            # Shows: -h --help
icw status --<TAB>           # Shows: -h --help
icw tree -h                  # Auto-completes
```

### Help and Completion Commands

#### Help Command
- `icw help <TAB>` - Shows all commands for help topics

**Example:**
```bash
icw help <TAB>               # Shows all commands
icw help migrate             # Get help on migrate
```

#### Completion Command
- `icw completion <TAB>` - Shows shell types:
  - `bash` `zsh` `fish` `powershell`

**Example:**
```bash
icw completion <TAB>         # Shows: bash zsh fish powershell
icw completion bash > /etc/bash_completion.d/icw
```

## Smart Completion Examples

### Typical Workflows

**Listing components:**
```bash
icw list --type d<TAB>       # Completes to: digital
icw ls -t analog -b          # All flags auto-complete
```

**Migration workflow:**
```bash
icw migrate --cr<TAB>        # Completes to: --create-repo
icw migrate --create-repo cp4
icw migrate --from cp3 --to cp4 --dr<TAB>  # Completes to: --dry-run
```

**Adding components:**
```bash
icw add dig<TAB>             # Shows matching directories
icw add digital/mycomp <TAB> # Shows: analog digital setup process
icw add digital/mycomp digital  # Complete command
```

**Checking status:**
```bash
icw st<TAB>                  # Completes to: status (or st)
icw status                   # No arguments, shows workspace status
```

## Completion Intelligence

The completion system provides:

1. **Context-aware suggestions**: Different completions based on command and position
2. **Flag value completion**: After flags like `--type`, suggests valid values
3. **Alias support**: Works with command aliases (`st`, `ls`)
4. **Smart argument detection**: Knows when to show directories vs component types
5. **Global vs local flags**: Shows appropriate flags for each command

## Testing Completion

To test if completion is working:

```bash
# Test command completion
icw <TAB><TAB>

# Test flag completion
icw list --<TAB><TAB>

# Test value completion
icw list --type <TAB><TAB>

# Test add completion
icw add <TAB>

# Verify completion is loaded
complete -p icw
# Should show: complete -F _icw_complete icw
```

## Completion Coverage

| Command | Flags | Value Completion | Argument Completion |
|---------|-------|------------------|---------------------|
| `list`, `ls` | ✅ `-t/-b/-g/-a` | ✅ Component types | ⚫ Files |
| `migrate` | ✅ `--create-repo/--from/--to/--dry-run` | ⚫ Repo names | ❌ |
| `add` | ✅ `--help` | ❌ | ✅ Dirs + types |
| `update` | ✅ `--help` | ❌ | ❌ |
| `status`, `st` | ✅ `--help` | ❌ | ❌ |
| `tree` | ✅ `--help` | ❌ | ❌ |
| `hdl` | ✅ `--help` | ❌ | ❌ |
| `test` | ✅ `--help` | ❌ | ❌ |
| `version` | ✅ `--help` | ❌ | ❌ |
| `completion` | ✅ `--help` | ✅ Shell types | ❌ |
| `help` | ✅ `--help` | ❌ | ✅ Commands |

**Legend:**
- ✅ Full support
- ⚫ Partial support (lets user type freely)
- ❌ Not applicable

## Future Enhancements

Possible future improvements:
- Dynamic repository name completion for `--from/--to/--create-repo`
- Component path completion from SVN repository
- Branch/tag name completion
- Recently used values for quick access

## Troubleshooting

**Completion not working:**
```bash
# Reload completion
source ~/.bash_completion.d/icw

# Check if loaded
complete -p icw

# Re-install
./setup_completion_g9.sh
```

**Wrong suggestions:**
```bash
# Clear bash completion cache
hash -r

# Restart bash
exec bash
```

**Completion shows nothing:**
- Ensure ICW is in your PATH: `which icw`
- Check completion file exists: `ls -la ~/.bash_completion.d/icw`
- Verify bash-completion package is installed: `apt list --installed | grep bash-completion`
