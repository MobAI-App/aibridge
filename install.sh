#!/bin/bash
set -e

# AiBridge Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/MobAI-App/aibridge/main/install.sh | bash

REPO="MobAI-App/aibridge"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case "$OS" in
    darwin|linux)
        ;;
    *)
        echo "Unsupported OS: $OS"
        echo "For Windows, download from: https://github.com/$REPO/releases"
        exit 1
        ;;
esac

# Get latest version
echo "Fetching latest version..."
VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    echo "Failed to get latest version"
    exit 1
fi

echo "Installing aibridge v$VERSION for $OS/$ARCH..."

# Download
FILENAME="aibridge_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/v$VERSION/$FILENAME"

TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

echo "Downloading $URL..."
curl -fsSL "$URL" -o "$TEMP_DIR/$FILENAME"

# Extract
echo "Extracting..."
tar -xzf "$TEMP_DIR/$FILENAME" -C "$TEMP_DIR"

# Install
echo "Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$TEMP_DIR/aibridge" "$INSTALL_DIR/"
else
    sudo mv "$TEMP_DIR/aibridge" "$INSTALL_DIR/"
fi

chmod +x "$INSTALL_DIR/aibridge"

echo ""
echo "✓ aibridge v$VERSION installed successfully!"
echo ""
echo "Run 'aibridge --help' to get started."
