#!/bin/bash
# Configure sudo for passwordless execution of dbusunit by user mte

set -e

echo "Configuring sudo for passwordless dbusunit execution..."
echo ""

# Get the absolute path to the dbusunit binary
DBUSUNIT_PATH="/home/mte/rmkhaklab/halko/bin/dbusunit"

if [ ! -f "$DBUSUNIT_PATH" ]; then
    echo "ERROR: dbusunit binary not found at $DBUSUNIT_PATH"
    echo "Please build it first: make build"
    exit 1
fi

# Create sudoers configuration file
SUDOERS_FILE="/etc/sudoers.d/halko-dbusunit"

echo "This script will create: $SUDOERS_FILE"
echo "It allows user 'mte' to run dbusunit without password."
echo ""
read -p "Continue? (y/n) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 1
fi

# Create the sudoers file
cat << EOF | sudo tee "$SUDOERS_FILE" > /dev/null
# Allow user mte to run dbusunit without password
# This is needed for the tmux debug environment
# File created by scripts/setup-sudo-dbusunit.sh

mte ALL=(root) NOPASSWD: $DBUSUNIT_PATH
EOF

# Set proper permissions (sudoers files must be 0440)
sudo chmod 0440 "$SUDOERS_FILE"

# Validate the sudoers file
if sudo visudo -c -f "$SUDOERS_FILE"; then
    echo ""
    echo "✓ Sudo configuration created successfully!"
    echo ""
    echo "User 'mte' can now run:"
    echo "  sudo $DBUSUNIT_PATH"
    echo ""
    echo "Without being prompted for a password."
else
    echo ""
    echo "ERROR: Sudoers file validation failed!"
    echo "Removing invalid file..."
    sudo rm -f "$SUDOERS_FILE"
    exit 1
fi
