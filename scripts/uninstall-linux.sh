#!/bin/bash
# Minimal rye-fyne uninstallation script for Linux

INSTALL_DIR="$HOME/.local/bin"

if [ -f "$INSTALL_DIR/rye-fyne" ]; then
    echo "Removing rye-fyne..."
    rm "$INSTALL_DIR/rye-fyne"
    
    # Remove from PATH in .bashrc if it exists
    if [ -f ~/.bashrc ] && grep -q "$INSTALL_DIR" ~/.bashrc; then
        sed -i "/export PATH.*$INSTALL_DIR/d" ~/.bashrc
        echo "Removed from PATH in ~/.bashrc"
    fi
    
    echo "rye-fyne uninstalled successfully!"
else
    echo "rye-fyne not found in $INSTALL_DIR"
fi
