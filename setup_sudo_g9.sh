#!/bin/bash
# Setup script to add svnadmin to sudoers NOPASSWD list on g9

echo "ICW Sudo Configuration for g9"
echo "=============================="
echo ""
echo "This script will add svnadmin to the NOPASSWD sudo list."
echo "You'll need to enter your password once to configure sudo."
echo ""

# Check if running on g9
hostname=$(hostname)
if [ "$hostname" != "g9" ]; then
    echo "Warning: Not running on g9 (current: $hostname)"
    echo "This configuration is specifically for the g9 server."
    read -p "Continue anyway? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Find svnadmin
SVNADMIN=$(which svnadmin)
if [ -z "$SVNADMIN" ]; then
    echo "Error: svnadmin not found in PATH"
    exit 1
fi

echo "Found svnadmin at: $SVNADMIN"
echo ""

# Create sudoers entry
SUDOERS_FILE="/etc/sudoers.d/icw-svnadmin"
SUDOERS_CONTENT="# Allow ICW to manage SVN repositories
jakobsen ALL=(ALL) NOPASSWD: $SVNADMIN"

echo "Will create: $SUDOERS_FILE"
echo "Content:"
echo "$SUDOERS_CONTENT"
echo ""

read -p "Proceed with sudo configuration? [Y/n] " -n 1 -r
echo
if [[ $REPLY =~ ^[Nn]$ ]]; then
    echo "Aborted."
    exit 0
fi

# Write sudoers file
echo "$SUDOERS_CONTENT" | sudo tee "$SUDOERS_FILE" > /dev/null

# Set correct permissions
sudo chmod 0440 "$SUDOERS_FILE"

# Validate sudoers syntax
if sudo visudo -c -f "$SUDOERS_FILE" 2>&1 | grep -q "parsed OK"; then
    echo ""
    echo "✓ Sudo configuration successfully installed!"
    echo "✓ File: $SUDOERS_FILE"
    echo ""
    echo "You can now use: icw migrate --create-repo <name>"
else
    echo ""
    echo "✗ Error: Invalid sudoers syntax"
    sudo rm -f "$SUDOERS_FILE"
    exit 1
fi
