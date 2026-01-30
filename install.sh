#!/usr/bin/env bash
# Pryx One-Liner Installer
# Usage: curl -fsSL https://pryx.dev/install | bash
#
# This installer works on all devices (macOS, Linux) and is the RECOMMENDED
# installation method. It downloads the latest Pryx release, installs it to
# ~/.local/bin (or /usr/local/bin with sudo), and adds it to your PATH.
#
# Options:
#   PRYX_INSTALL_DIR - Custom installation directory (default: ~/.local/bin)
#   PRYX_VERSION     - Specific version to install (default: latest)
#   PRYX_NO_MODIFY_PATH - Set to 1 to skip PATH modification

set -e

# Colors
RED=$(tput setaf 1 2>/dev/null || printf '\033[0;31m')
GREEN=$(tput setaf 2 2>/dev/null || printf '\033[0;32m')
YELLOW=$(tput setaf 3 2>/dev/null || printf '\033[0;33m')
BLUE=$(tput setaf 4 2>/dev/null || printf '\033[0;34m')
NC=$(tput sgr0 2>/dev/null || printf '\033[0m')

log()   { echo -e "${GREEN}[pryx]${NC} $1"; }
warn()  { echo -e "${YELLOW}[pryx]${NC} $1"; }
error() { echo -e "${RED}[pryx]${NC} $1" >&2; exit 1; }

# Detect platform
detect_platform() {
    local os arch
    os="$(uname -s)"
    arch="$(uname -m)"
    
    case "$os" in
        Darwin) os="darwin" ;;
        Linux) os="linux" ;;
        *) error "Unsupported OS: $os" ;;
    esac
    
    case "$arch" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *) error "Unsupported architecture: $arch" ;;
    esac
    
    echo "${os}-${arch}"
}

# Determine install directory
get_install_dir() {
    if [[ -n "$PRYX_INSTALL_DIR" ]]; then
        echo "$PRYX_INSTALL_DIR"
    elif [[ -w "/usr/local/bin" ]]; then
        echo "/usr/local/bin"
    else
        echo "$HOME/.local/bin"
    fi
}

# Get latest version from GitHub releases
get_latest_version() {
    if [[ -n "$PRYX_VERSION" ]]; then
        echo "$PRYX_VERSION"
        return
    fi
    
    curl -fsSL "https://api.github.com/repos/pryx-dev/pryx/releases/latest" \
        | grep '"tag_name"' \
        | sed -E 's/.*"v?([^"]+)".*/\1/' \
        | head -1
}

# Download and install
install_pryx() {
    local platform version install_dir download_url tmp_dir
    
    platform="$(detect_platform)"
    version="${PRYX_VERSION:-$(get_latest_version)}"
    install_dir="$(get_install_dir)"
    download_url="https://github.com/pryx-dev/pryx/releases/download/v${version}/pryx-${platform}.tar.gz"
    
    log "Installing Pryx v${version} for ${platform}..."
    log "Install directory: ${install_dir}"
    
    # Create install dir
    mkdir -p "$install_dir"
    
    # Download to temp
    tmp_dir="$(mktemp -d)"
    trap "rm -rf $tmp_dir" EXIT
    
    log "Downloading from ${download_url}..."
    if ! curl -fsSL "$download_url" -o "$tmp_dir/pryx.tar.gz"; then
        error "Failed to download Pryx. Check your internet connection."
    fi
    
    # Extract
    log "Extracting..."
    tar -xzf "$tmp_dir/pryx.tar.gz" -C "$tmp_dir"
    
    # Install binaries
    log "Installing binaries..."
    install -m 755 "$tmp_dir/pryx" "$install_dir/pryx" 2>/dev/null || \
        cp "$tmp_dir/pryx" "$install_dir/pryx" && chmod +x "$install_dir/pryx"
    
    if [[ -f "$tmp_dir/pryx-core" ]]; then
        install -m 755 "$tmp_dir/pryx-core" "$install_dir/pryx-core" 2>/dev/null || \
            cp "$tmp_dir/pryx-core" "$install_dir/pryx-core" && chmod +x "$install_dir/pryx-core"
    fi
    
    # Update PATH
    update_path "$install_dir"
    
    log "✓ Pryx installed successfully!"
    echo ""
    echo -e "  ${BLUE}Run 'pryx' to get started${NC}"
    echo -e "  ${BLUE}Run 'pryx doctor' to verify installation${NC}"
    echo ""
}

# Add to PATH
update_path() {
    local install_dir="$1"
    
    if [[ "$PRYX_NO_MODIFY_PATH" == "1" ]]; then
        return
    fi
    
    # Check if already in PATH
    if echo "$PATH" | tr ':' '\n' | grep -qx "$install_dir"; then
        return
    fi
    
    local shell_rc=""
    case "$SHELL" in
        */zsh) shell_rc="$HOME/.zshrc" ;;
        */bash) 
            if [[ -f "$HOME/.bashrc" ]]; then
                shell_rc="$HOME/.bashrc"
            else
                shell_rc="$HOME/.profile"
            fi
            ;;
        */fish) shell_rc="$HOME/.config/fish/config.fish" ;;
    esac
    
    if [[ -n "$shell_rc" ]]; then
        echo "" >> "$shell_rc"
        echo "# Pryx" >> "$shell_rc"
        echo "export PATH=\"$install_dir:\$PATH\"" >> "$shell_rc"
        log "Added ${install_dir} to PATH in ${shell_rc}"
        warn "Restart your shell or run: export PATH=\"$install_dir:\$PATH\""
    fi
}

# Main
main() {
    echo -e "${BLUE}"
    echo "  ____                    "
    echo " |  _ \ _ __ _   ___  __  "
    echo " | |_) | '__| | | \ \/ /  "
    echo " |  __/| |  | |_| |>  <   "
    echo " |_|   |_|   \__, /_/\_\  "
    echo "             |___/        "
    echo -e "${NC}"
    echo "The AI-powered coding assistant"
    echo ""
    
    install_pryx
}

main "$@"
