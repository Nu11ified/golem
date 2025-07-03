#!/bin/bash

# Golem Framework Installation Script
# This script downloads and installs the latest golem binary

set -e

# Configuration
REPO="Nu11ified/golem"
BINARY_NAME="golem"
INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# Detect OS and architecture
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case $os in
        linux*)
            OS="linux"
            ;;
        darwin*)
            OS="darwin"
            ;;
        mingw*|msys*|cygwin*)
            OS="windows"
            ;;
        *)
            error "Unsupported OS: $os"
            ;;
    esac
    
    case $arch in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $arch"
            ;;
    esac
    
    info "Detected platform: $OS/$ARCH"
}

# Get latest release version from GitHub
get_latest_version() {
    info "Fetching latest version..."
    
    if command -v curl >/dev/null 2>&1; then
        LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        LATEST_VERSION=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        error "Neither curl nor wget found. Please install one of them."
    fi
    
    if [ -z "$LATEST_VERSION" ]; then
        error "Failed to get latest version"
    fi
    
    info "Latest version: $LATEST_VERSION"
}

# Download and install binary
install_binary() {
    local filename="$BINARY_NAME-$LATEST_VERSION-$OS-$ARCH"
    local archive_name
    
    if [ "$OS" = "windows" ]; then
        filename="$filename.exe"
        archive_name="$BINARY_NAME-$LATEST_VERSION-$OS-$ARCH.zip"
    else
        archive_name="$BINARY_NAME-$LATEST_VERSION-$OS-$ARCH.tar.gz"
    fi
    
    local download_url="https://github.com/$REPO/releases/download/$LATEST_VERSION/$archive_name"
    local temp_dir=$(mktemp -d)
    
    info "Downloading $archive_name..."
    
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$temp_dir/$archive_name" "$download_url"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$temp_dir/$archive_name" "$download_url"
    else
        error "Neither curl nor wget found. Please install one of them."
    fi
    
    info "Extracting binary..."
    
    if [ "$OS" = "windows" ]; then
        unzip -q "$temp_dir/$archive_name" -d "$temp_dir"
    else
        tar -xzf "$temp_dir/$archive_name" -C "$temp_dir"
    fi
    
    # Check if we need sudo for installation
    if [ -w "$INSTALL_DIR" ]; then
        cp "$temp_dir/$filename" "$INSTALL_DIR/$BINARY_NAME"
    else
        warning "Need sudo privileges to install to $INSTALL_DIR"
        sudo cp "$temp_dir/$filename" "$INSTALL_DIR/$BINARY_NAME"
    fi
    
    # Make executable
    if [ "$OS" != "windows" ]; then
        if [ -w "$INSTALL_DIR/$BINARY_NAME" ]; then
            chmod +x "$INSTALL_DIR/$BINARY_NAME"
        else
            sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
        fi
    fi
    
    # Clean up
    rm -rf "$temp_dir"
    
    success "Golem installed successfully to $INSTALL_DIR/$BINARY_NAME"
}

# Verify installation
verify_installation() {
    if command -v $BINARY_NAME >/dev/null 2>&1; then
        local installed_version=$($BINARY_NAME version 2>/dev/null || echo "unknown")
        success "Installation verified: $installed_version"
        
        info "You can now use golem:"
        echo "  golem new my-app    # Create a new project"
        echo "  golem dev           # Start development server"
        echo "  golem build         # Build for production"
        echo "  golem help          # Show help"
    else
        error "Installation failed. $BINARY_NAME not found in PATH."
    fi
}

# Main execution
main() {
    echo -e "${BLUE}🗿 Golem Framework Installation Script${NC}"
    echo ""
    
    detect_platform
    get_latest_version
    install_binary
    verify_installation
    
    echo ""
    success "Installation complete! Happy coding with Golem! 🚀"
}

# Run main function
main "$@" 