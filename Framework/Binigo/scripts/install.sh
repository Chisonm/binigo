#!/bin/bash

# Binigo Framework Installer
# Install with: curl -sSL https://raw.githubusercontent.com/yourusername/binigo/main/install.sh | bash

set -e

BINIGO_VERSION="latest"
INSTALL_DIR="$HOME/.binigo"
BIN_DIR="$INSTALL_DIR/bin"

echo "🚀 Installing Binigo Framework..."
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or higher first."
    echo "Visit: https://golang.org/dl/"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.21"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "❌ Go version $REQUIRED_VERSION or higher is required. You have $GO_VERSION"
    exit 1
fi

echo "✅ Go version $GO_VERSION detected"
echo ""

# Create installation directory
mkdir -p "$BIN_DIR"

# Download and install binigo CLI
echo "📦 Installing Binigo CLI..."
go install github.com/yourusername/binigo/cmd/binigo@latest

# Check if installation was successful
if [ $? -eq 0 ]; then
    echo "✅ Binigo CLI installed successfully"
else
    echo "❌ Failed to install Binigo CLI"
    exit 1
fi

# Add to PATH if not already present
SHELL_CONFIG=""
if [ -n "$BASH_VERSION" ]; then
    SHELL_CONFIG="$HOME/.bashrc"
elif [ -n "$ZSH_VERSION" ]; then
    SHELL_CONFIG="$HOME/.zshrc"
fi

if [ -n "$SHELL_CONFIG" ]; then
    GOPATH=$(go env GOPATH)
    GOBIN="$GOPATH/bin"
    
    if [[ ":$PATH:" != *":$GOBIN:"* ]]; then
        echo "" >> "$SHELL_CONFIG"
        echo "# Binigo Framework" >> "$SHELL_CONFIG"
        echo "export PATH=\"\$PATH:$GOBIN\"" >> "$SHELL_CONFIG"
        echo "✅ Added Binigo to PATH in $SHELL_CONFIG"
        echo "   Please run: source $SHELL_CONFIG"
    fi
fi

echo ""
echo "🎉 Binigo Framework installed successfully!"
echo ""
echo "📚 Quick Start:"
echo "   1. Create a new project:"
echo "      binigo new myapp"
echo ""
echo "   2. Navigate to your project:"
echo "      cd myapp"
echo ""
echo "   3. Start development server:"
echo "      binigo serve"
echo ""
echo "📖 Documentation: https://github.com/yourusername/binigo"
echo "💬 Community: https://github.com/yourusername/binigo/discussions"
echo ""