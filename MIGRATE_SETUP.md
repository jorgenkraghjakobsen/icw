# Setting Up ICW Migrate Command on g9

Complete setup for repository migration on g9 server.

## Quick Setup (Complete)

```bash
# 1. Set up authentication
icw auth login
# Enter your SVN password

# 2. Set up sudo for svnadmin (one time)
echo '# Allow ICW to manage SVN repositories
jakobsen ALL=(ALL) NOPASSWD: /usr/bin/svnadmin' | sudo tee /etc/sudoers.d/icw-svnadmin > /dev/null && sudo chmod 0440 /etc/sudoers.d/icw-svnadmin && echo "✓ Sudo configured for svnadmin"

# 3. Test it
icw migrate --create-repo test-repo
icw list -r test-repo
```

## Detailed Setup

### Step 1: Authentication (Required)

**Use the new `icw auth` command:**

```bash
icw auth login
```

Enter your password once. It's securely stored and used automatically.

**Alternative** (old way, still works):
```bash
echo 'export ICW_SVN_PASSWORD=your_password' >> ~/.bashrc
source ~/.bashrc
```

### Step 2: Sudo Access (Required)

#### Option 1: One-line setup
Run this command on g9:
```bash
echo '# Allow ICW to manage SVN repositories
jakobsen ALL=(ALL) NOPASSWD: /usr/bin/svnadmin' | sudo tee /etc/sudoers.d/icw-svnadmin > /dev/null && sudo chmod 0440 /etc/sudoers.d/icw-svnadmin && echo "✓ Sudo configured for svnadmin"
```

### Option 2: Manual setup
```bash
# On g9 server, create the sudoers file
sudo nano /etc/sudoers.d/icw-svnadmin
```

Add this content:
```
# Allow ICW to manage SVN repositories
jakobsen ALL=(ALL) NOPASSWD: /usr/bin/svnadmin
```

Set correct permissions:
```bash
sudo chmod 0440 /etc/sudoers.d/icw-svnadmin
```

Validate:
```bash
sudo visudo -c -f /etc/sudoers.d/icw-svnadmin
```

Should show: "parsed OK"

## Verify Setup

Test that sudo works without password:
```bash
sudo svnadmin --version
```

Should show version info without asking for a password.

## Test Everything Works

```bash
# Check authentication
icw auth status

# Interactive mode - shows available repos
icw migrate

# Create a new repository
icw migrate --create-repo my-new-repo

# Add yourself to the repo
icw migrate --add-user jakobsen --to my-new-repo

# List components (should be empty)
icw list -r my-new-repo

# Verify it was created
ls -la /data_v1/svn/repos/my-new-repo
```

## Already Configured Commands

You already have passwordless sudo for these commands (from existing config):
- `/usr/bin/apt`, `/usr/bin/apt-get`
- `/usr/bin/systemctl`, `/usr/bin/journalctl`
- `/usr/bin/mkdir`, `/usr/bin/chown`, `/usr/bin/chmod`
- `/usr/bin/tar`, `/usr/bin/tee`
- `/usr/bin/cp`, `/usr/bin/mv`, `/usr/bin/rm`
- `/usr/sbin/nginx`, `/usr/bin/rsync`
- `/usr/bin/testparm`, `/usr/sbin/smbpasswd`, `/usr/bin/du`

Now adding:
- `/usr/bin/svnadmin` ✨ (for ICW migrate)

## Why Sudo is Needed

The SVN repository directory `/data_v1/svn/repos` is owned by root:
```
drwxr-xr-x 11 root root 4096 Dec  3 13:53 /data_v1/svn/repos
```

Repository creation requires:
1. `sudo svnadmin create` - Create the repository structure
2. `sudo tee` - Write svnserve.conf (already has NOPASSWD)

## Troubleshooting

### "sudo: a password is required"
The sudoers file hasn't been created yet. Follow the setup steps above.

### "parsed OK" not shown
There's a syntax error in the sudoers file. Check for typos.

### Repository still won't create
Check permissions:
```bash
ls -la /data_v1/svn/repos/
sudo -l | grep svnadmin
```

### Want to test without creating repos
Use dry-run mode:
```bash
~/bin/icw migrate --from cp3 --to cp4 --dry-run
```

## Security Note

The sudoers configuration is limited to:
- **User**: Only `jakobsen`
- **Command**: Only `/usr/bin/svnadmin` (full path)
- **Purpose**: SVN repository management for ICW

This follows the principle of least privilege - only the minimal permissions needed for the migrate function.
