#!/bin/bash
# Minimal rye-fyne uninstallation script for macOS

INSTALL_DIR="$HOME/.local/bin"

if [ -f "$INSTALL_DIR/rye-fyne" ]; then
    echo "Removing rye-fyne..."
    rm "$INSTALL_DIR/rye-fyne"
    
    # Remove from PATH in .zshrc if it exists
    if [ -f ~/.zshrc ] && grep -q "$INSTALL_DIR" ~/.zshrc; then
        sed -i '' "/export PATH.*$INSTALL_DIR/d" ~/.zshrc
        echo "Removed from PATH in ~/.zshrc"
    fi
    
    echo "rye-fyne uninstalled successfully!"
else
    echo "rye-fyne not found in $INSTALL_DIR"
fi
