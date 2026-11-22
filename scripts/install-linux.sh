#!/bin/bash
# Minimal rye-fyne installation script for Linux

set -e

REPO="refaktor/rye-fyne"
INSTALL_DIR="$HOME/.local/bin"

# Get latest version
VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

# Download and install
echo "Installing rye-fyne $VERSION..."
mkdir -p "$INSTALL_DIR"
curl -sL "https://github.com/$REPO/releases/download/$VERSION/rye-fyne-linux-amd64.tar.gz" | tar -xz -C /tmp
mv /tmp/rye-fyne "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/rye-fyne"

# Add to PATH if needed
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> ~/.bashrc
    echo "Added to PATH. Restart your terminal or run: source ~/.bashrc"
fi

echo "rye-fyne installed successfully!"
echo "Usage: rye-fyne [your-app.rye]"
