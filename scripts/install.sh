#!/usr/bin/env bash
#
# Pryx One-Liner Installer
# Installs pryx-core on macOS/Linux in <60 seconds
# No external deps required (Node.js, Python, Docker)
#

set -eEuo pipefail

# Colors for output
RED=$(tput setaf 1 2>/dev/null || printf '\033[0;31m')
GREEN=$(tput setaf 2 2>/dev/null || printf '\033[0;32m')
YELLOW=$(tput setaf 3 2>/dev/null || printf '\033[0;33m')
NC=$(tput sgr0 2>/dev/null || printf '\033[0m')

# Configuration
VERSION="0.1.0-dev"
DOWNLOAD_BASE="https://github.com/pryx-core/releases"
INSTALL_DIR="${HOME}/.pryx/bin"

# Platform detection
OS_TYPE="$(uname -s)"
ARCH_TYPE="$(uname -m)"

echo -e "${GREEN}Pryx One-Liner Installer${NC}"
echo -e "${GREEN}Platform: ${OS_TYPE} (${ARCH_TYPE})${NC}"
echo ""

# Function to download pryx-core
download_pryx_core() {
    local binary_name="pryx-core"
    
    # Add platform suffix if needed
    if [ "$OS_TYPE" = "Darwin" ]; then
        if [ "$ARCH_TYPE" = "arm64" ]; then
            binary_name="${binary_name}-darwin-arm64"
        elif [ "$ARCH_TYPE" = "x86_64" ]; then
            binary_name="${binary_name}-darwin-amd64"
        fi
    elif [ "$OS_TYPE" = "Linux" ]; then
        if [ "$ARCH_TYPE" = "x86_64" ]; then
            binary_name="${binary_name}-linux-amd64"
        elif [ "$ARCH_TYPE" = "aarch64" ]; then
            binary_name="${binary_name}-linux-aarch64"
        fi
    fi
    
    local download_url="${DOWNLOAD_BASE}/v${VERSION}/${binary_name}"
    
    echo -e "${YELLOW}Downloading ${binary_name}...${NC}"
    
    # Create install directory
    mkdir -p "$INSTALL_DIR"
    
    # Download binary
    if command -v curl &> /dev/null; then
        curl -fsSL --progress-bar -o "$INSTALL_DIR/pryx-core" "$download_url"
    elif command -v wget &> /dev/null; then
        wget --progress=bar:force -O "$INSTALL_DIR/pryx-core" "$download_url"
    else
        echo -e "${RED}Error: Neither curl nor wget found. Please install curl.${NC}"
        exit 1
    fi
    
    # Make executable
    chmod +x "$INSTALL_DIR/pryx-core"
    
    echo -e "${GREEN}Downloaded to: $INSTALL_DIR/pryx-core${NC}"
}

# Function to install to PATH
install_to_path() {
    local install_path="$1"
    
    echo -e "${YELLOW}Adding to PATH...${NC}"
    
    # Determine shell config file
    local shell_config=""
    if [ -n "$SHELL" ]; then
        SHELL_CONFIG="$HOME/.bashrc"
        if [ -f "$HOME/.zshrc" ]; then
            SHELL_CONFIG="$HOME/.zshrc"
        fi
    fi
    
    # Create/update config if not already there
    if [ -f "$shell_config" ]; then
        if ! grep -q "pryx" "$shell_config"; then
            echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$shell_config"
            echo -e "${GREEN}Added to $shell_config${NC}"
        fi
    else
        echo "export PATH=\"\$PATH:$INSTALL_DIR\"" > "$shell_config"
        echo -e "${GREEN}Created $shell_config${NC}"
    fi
    
    # Add to current session
    export PATH="$PATH:$INSTALL_DIR"
    echo -e "${GREEN}PATH updated: $PATH${NC}"
}

# Function to verify installation
verify_installation() {
    echo ""
    echo -e "${GREEN}=== Verifying Installation ===${NC}"
    
    # Check if binary exists
    if [ -f "$INSTALL_DIR/pryx-core" ]; then
        echo -e "${GREEN}✓ Binary found at: $INSTALL_DIR/pryx-core${NC}"
    else
        echo -e "${RED}✗ Binary not found${NC}"
        return 1
    fi
    
    # Check if binary is executable
    if [ -x "$INSTALL_DIR/pryx-core" ]; then
        echo -e "${GREEN}✓ Binary is executable${NC}"
    else
        echo -e "${RED}✗ Binary is not executable${NC}"
        return 1
    fi
    
    # Check if in PATH
    if command -v pryx-core &> /dev/null; then
        echo -e "${GREEN}✓ pryx-core command found in PATH${NC}"
    else
        echo -e "${YELLOW}⚠ pryx-core command not in PATH${NC}"
    fi
    
    # Check version
    if "$INSTALL_DIR/pryx-core" version &> /dev/null; then
        VERSION_OUTPUT="$($INSTALL_DIR/pryx-core version 2>&1 || true)"
        echo -e "${GREEN}✓ Version: $VERSION_OUTPUT${NC}"
    else
        echo -e "${YELLOW}⚠ Could not determine version${NC}"
    fi
    
    echo -e "${GREEN}=== Verification Complete ===${NC}"
    return 0
}

# Main installation flow
main() {
    echo -e "${GREEN}Pryx One-Liner Installer${NC}"
    echo -e "${GREEN}Version: $VERSION${NC}"
    echo ""
    
    # Download and install
    download_pryx_core
    
    # Install to PATH
    install_to_path "$INSTALL_DIR"
    
    # Verify installation
    if verify_installation; then
        echo ""
        echo -e "${GREEN}Installation successful!${NC}"
        echo ""
        echo -e "${YELLOW}To get started, run:${NC}"
        echo -e "${GREEN}  pryx-core${NC}"
        echo ""
        echo -e "${YELLOW}To verify installation, run:${NC}"
        echo -e "${GREEN}  pryx doctor${NC}"
        echo ""
        echo -e "${YELLOW}Configuration directory:${NC}"
        echo -e "${GREEN}  $HOME/.pryx${NC}"
        exit 0
    else
        echo -e "${RED}Installation verification failed${NC}"
        exit 1
    fi
}
