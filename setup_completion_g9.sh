#!/bin/bash
# Setup script for ICW bash completion on g9 server
# This script can be run with or without sudo privileges

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPLETION_FILE="$SCRIPT_DIR/completions/icw_bashcompletion.sh"
BINARY_NAME="icw"

echo "ICW Bash Completion Setup for g9"
echo "=================================="
echo ""

# Check if completion file exists
if [ ! -f "$COMPLETION_FILE" ]; then
    echo "Error: Completion file not found at $COMPLETION_FILE"
    exit 1
fi

echo "Found completion file: $COMPLETION_FILE"
echo ""

# Try system-wide installation first
if [ -w /usr/local/share/bash-completion/completions ] || [ "$EUID" -eq 0 ]; then
    echo "Installing system-wide completion..."
    SYSTEM_COMPLETION_DIR="/usr/local/share/bash-completion/completions"

    if [ ! -d "$SYSTEM_COMPLETION_DIR" ]; then
        echo "Creating $SYSTEM_COMPLETION_DIR..."
        sudo mkdir -p "$SYSTEM_COMPLETION_DIR"
    fi

    sudo cp "$COMPLETION_FILE" "$SYSTEM_COMPLETION_DIR/$BINARY_NAME"
    echo "✓ Installed to: $SYSTEM_COMPLETION_DIR/$BINARY_NAME"
    echo ""
    echo "Completion will be available in new bash sessions."
    echo "To activate in current session, run:"
    echo "  source $SYSTEM_COMPLETION_DIR/$BINARY_NAME"

else
    # User-local installation
    echo "No sudo access. Installing user-local completion..."
    USER_COMPLETION_DIR="$HOME/.bash_completion.d"

    mkdir -p "$USER_COMPLETION_DIR"
    cp "$COMPLETION_FILE" "$USER_COMPLETION_DIR/$BINARY_NAME"
    echo "✓ Installed to: $USER_COMPLETION_DIR/$BINARY_NAME"
    echo ""

    # Check if bashrc sources completions
    BASHRC="$HOME/.bashrc"
    COMPLETION_LOADER="
# Load bash completions from ~/.bash_completion.d
if [ -d \"\$HOME/.bash_completion.d\" ]; then
    for completion_file in \"\$HOME/.bash_completion.d\"/*; do
        [ -r \"\$completion_file\" ] && source \"\$completion_file\"
    done
fi"

    if ! grep -q "\.bash_completion\.d" "$BASHRC" 2>/dev/null; then
        echo "Adding completion loader to $BASHRC..."
        echo "$COMPLETION_LOADER" >> "$BASHRC"
        echo "✓ Updated $BASHRC"
    else
        echo "✓ Completion loader already in $BASHRC"
    fi

    echo ""
    echo "To activate completion, run:"
    echo "  source $USER_COMPLETION_DIR/$BINARY_NAME"
    echo "Or start a new bash session."
fi

echo ""
echo "=================================="
echo "Setup complete!"
echo ""
echo "Next step: Set up authentication"
echo "  icw auth login"
echo ""
echo "Then test it:"
echo "  icw list -r cp3"
echo ""
echo "Note: The 'migrate' command only works on g9 server."
echo "      All other commands work on any server."
