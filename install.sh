#!/usr/bin/env bash
# Pryx One-Liner Installer
# Usage: curl -fsSL https://pryx.dev/install | bash
#
# This installer works on all devices (macOS, Linux) and is the RECOMMENDED
# installation method. It downloads the latest Pryx release, installs it to
# ~/.local/bin (or /usr/local/bin with sudo), and adds it to your PATH.
#
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

usage() {
    echo "Usage: curl -fsSL https://pryx.dev/install | bash -s -- [options]"
    echo ""
    echo "Options:"
    echo "  --install-dir <path>     Custom installation directory"
    echo "  --version <version>      Specific version to install"
    echo "  --no-path                Skip PATH modification"
    echo "  --no-onboard             Skip onboarding message"
    echo "  --dry-run                Print actions without making changes"
    echo "  --no-prompt              Non-interactive mode"
    echo "  --help                   Show this help"
    echo ""
    echo "Environment variables:"
    echo "  PRYX_INSTALL_DIR"
    echo "  PRYX_VERSION"
    echo "  PRYX_NO_MODIFY_PATH=1"
    echo "  PRYX_NO_ONBOARD=1"
    echo "  PRYX_DRY_RUN=1"
    echo "  PRYX_NO_PROMPT=1"
}

require_command() {
    if ! command -v "$1" >/dev/null 2>&1; then
        error "Missing required command: $1"
    fi
}

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

# Parse args
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --install-dir)
                PRYX_INSTALL_DIR="$2"
                shift 2
                ;;
            --version)
                PRYX_VERSION="$2"
                shift 2
                ;;
            --no-path)
                PRYX_NO_MODIFY_PATH=1
                shift
                ;;
            --no-onboard)
                PRYX_NO_ONBOARD=1
                shift
                ;;
            --dry-run)
                PRYX_DRY_RUN=1
                shift
                ;;
            --no-prompt)
                PRYX_NO_PROMPT=1
                shift
                ;;
            --help|-h)
                usage
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                ;;
        esac
    done
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

    require_command curl
    require_command tar

    platform="$(detect_platform)"
    version="${PRYX_VERSION:-$(get_latest_version)}"
    install_dir="$(get_install_dir)"
    download_url="https://github.com/pryx-dev/pryx/releases/download/v${version}/pryx-${platform}.tar.gz"

    log "Installing Pryx v${version} for ${platform}..."
    log "Install directory: ${install_dir}"

    if [[ -n "$PRYX_INSTALL_DIR" ]]; then
        if [[ -d "$install_dir" && ! -w "$install_dir" ]]; then
            error "Install directory is not writable: ${install_dir}"
        fi
        if [[ ! -d "$install_dir" && ! -w "$(dirname "$install_dir")" ]]; then
            error "Install directory is not writable: ${install_dir}"
        fi
    fi

    if [[ -d "$install_dir" && ! -w "$install_dir" ]]; then
        warn "Install directory not writable: ${install_dir}"
        install_dir="$HOME/.local/bin"
        warn "Falling back to ${install_dir}"
    fi

    if [[ ! -d "$install_dir" && ! -w "$(dirname "$install_dir")" ]]; then
        error "Install directory is not writable: ${install_dir}"
    fi

    if [[ "$PRYX_DRY_RUN" == "1" ]]; then
        echo "dry-run: mkdir -p \"$install_dir\""
        echo "dry-run: download ${download_url}"
        echo "dry-run: extract to temp dir"
        echo "dry-run: install binaries to ${install_dir}"
        if [[ "$PRYX_NO_MODIFY_PATH" != "1" ]]; then
            echo "dry-run: update PATH in shell rc"
        fi
        exit 0
    fi

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
    if [[ "$PRYX_NO_ONBOARD" != "1" ]]; then
        echo -e "  ${BLUE}Run 'pryx' to get started${NC}"
        echo -e "  ${BLUE}Run 'pryx doctor' to verify installation${NC}"
    fi
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
    parse_args "$@"
    echo -e "${BLUE}"
    echo "  ____                    "
    echo " |  _ \ _ __ _   ___  __  "
    echo " | |_) | '__| | | \ \/ /  "
    echo " |  __/| |  | |_| |>  <   "
    echo " |_|   |_|   \__, /_/\_\  "
    echo "             |___/        "
    echo -e "${NC}"
    echo "Sovereign AI agent with local-first control center"
    echo ""
    
    install_pryx
}

main "$@"
