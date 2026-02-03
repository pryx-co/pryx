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
BLUE=$(tput setaf 4 2>/dev/null || printf '\033[0;34m')
NC=$(tput sgr0 2>/dev/null || printf '\033[0m')

log()   { echo -e "${GREEN}[pryx]${NC} $1"; }
warn()  { echo -e "${YELLOW}[pryx]${NC} $1"; }
error() { echo -e "${RED}[pryx]${NC} $1" >&2; exit 1; }

# Configuration
VERSION="${PRYX_VERSION:-latest}"
DOWNLOAD_BASE="https://github.com/pryx-dev/pryx/releases"
INSTALL_DIR="${PRYX_INSTALL_DIR:-${HOME}/.pryx/bin}"

# Platform detection
OS_TYPE="$(uname -s)"
ARCH_TYPE="$(uname -m)"

usage() {
    echo "Usage: curl -fsSL https://get.pryx.ai/install.sh | bash -s -- [options]"
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

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --install-dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            --version)
                VERSION="$2"
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

require_command() {
    if ! command -v "$1" >/dev/null 2>&1; then
        error "Missing required command: $1"
    fi
}

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

get_install_dir() {
    if [[ -n "${PRYX_INSTALL_DIR:-}" ]]; then
        echo "$PRYX_INSTALL_DIR"
    elif [[ -w "/usr/local/bin" ]]; then
        echo "/usr/local/bin"
    else
        echo "$HOME/.pryx/bin"
    fi
}

get_latest_version() {
    if [[ -n "${PRYX_VERSION:-}" ]]; then
        echo "$PRYX_VERSION"
        return
    fi

    curl -fsSL "https://api.github.com/repos/pryx-dev/pryx/releases/latest" \
        | grep '"tag_name"' \
        | sed -E 's/.*"v?([^"]+)".*/\1/' \
        | head -1
}

download_pryx_core() {
    local platform version install_dir download_url tmp_dir

    require_command curl
    require_command tar

    platform="$(detect_platform)"
    version="${PRYX_VERSION:-$(get_latest_version)}"
    install_dir="$(get_install_dir)"
    download_url="${DOWNLOAD_BASE}/download/v${version}/pryx-${platform}.tar.gz"

    log "Installing Pryx v${version} for ${platform}..."
    log "Install directory: ${install_dir}"

    if [[ -n "${PRYX_INSTALL_DIR:-}" ]]; then
        if [[ -d "$install_dir" && ! -w "$install_dir" ]]; then
            error "Install directory is not writable: ${install_dir}"
        fi
        if [[ ! -d "$install_dir" && ! -w "$(dirname "$install_dir")" ]]; then
            error "Install directory is not writable: ${install_dir}"
        fi
    fi

    if [[ -d "$install_dir" && ! -w "$install_dir" ]]; then
        warn "Install directory not writable: ${install_dir}"
        install_dir="$HOME/.pryx/bin"
        warn "Falling back to ${install_dir}"
    fi

    if [[ ! -d "$install_dir" && ! -w "$(dirname "$install_dir")" ]]; then
        error "Install directory is not writable: ${install_dir}"
    fi

    if [[ "${PRYX_DRY_RUN:-}" == "1" ]]; then
        echo "dry-run: mkdir -p \"$install_dir\""
        echo "dry-run: download ${download_url}"
        echo "dry-run: extract to temp dir"
        echo "dry-run: install binaries to ${install_dir}"
        if [[ "${PRYX_NO_MODIFY_PATH:-}" != "1" ]]; then
            echo "dry-run: update PATH in shell rc"
        fi
        exit 0
    fi

    mkdir -p "$install_dir"

    tmp_dir="$(mktemp -d)"
    trap "rm -rf $tmp_dir" EXIT

    log "Downloading from ${download_url}..."
    if ! curl -fsSL "$download_url" -o "$tmp_dir/pryx.tar.gz"; then
        error "Failed to download Pryx. Check your internet connection."
    fi

    log "Extracting..."
    tar -xzf "$tmp_dir/pryx.tar.gz" -C "$tmp_dir"

    log "Installing binaries..."
    install -m 755 "$tmp_dir/pryx" "$install_dir/pryx" 2>/dev/null || \
        cp "$tmp_dir/pryx" "$install_dir/pryx" && chmod +x "$install_dir/pryx"

    if [[ -f "$tmp_dir/pryx-core" ]]; then
        install -m 755 "$tmp_dir/pryx-core" "$install_dir/pryx-core" 2>/dev/null || \
            cp "$tmp_dir/pryx-core" "$install_dir/pryx-core" && chmod +x "$install_dir/pryx-core"
    fi

    INSTALL_DIR="$install_dir"

    log "✓ Pryx installed successfully!"
}

# Function to install to PATH
install_to_path() {
    local install_path="$1"
    
    if [[ "${PRYX_NO_MODIFY_PATH:-}" == "1" ]]; then
        return
    fi

    echo -e "${YELLOW}Adding to PATH...${NC}"
    
    # Determine shell config file
    local shell_config=""
    if [ -n "$SHELL" ]; then
        shell_config="$HOME/.bashrc"
        if [ -f "$HOME/.zshrc" ]; then
            shell_config="$HOME/.zshrc"
        fi
    fi
    
    # Create/update config if not already there
    if [ -f "$shell_config" ]; then
        if ! grep -q "pryx" "$shell_config"; then
            echo "export PATH=\"$install_path:\$PATH\"" >> "$shell_config"
            echo -e "${GREEN}Added to $shell_config${NC}"
        fi
    else
        echo "export PATH=\"$install_path:\$PATH\"" > "$shell_config"
        echo -e "${GREEN}Created $shell_config${NC}"
    fi
    
    export PATH="$install_path:$PATH"
    echo -e "${GREEN}PATH updated: $PATH${NC}"
}

# Function to verify installation
verify_installation() {
    echo ""
    echo -e "${GREEN}=== Verifying Installation ===${NC}"
    
    if [ -f "$INSTALL_DIR/pryx" ]; then
        echo -e "${GREEN}✓ Binary found at: $INSTALL_DIR/pryx${NC}"
    else
        echo -e "${RED}✗ Binary not found: pryx${NC}"
        return 1
    fi
    
    if [ -x "$INSTALL_DIR/pryx" ]; then
        echo -e "${GREEN}✓ Binary is executable${NC}"
    else
        echo -e "${RED}✗ Binary is not executable${NC}"
        return 1
    fi
    
    if [ -f "$INSTALL_DIR/pryx-core" ]; then
        echo -e "${GREEN}✓ Binary found at: $INSTALL_DIR/pryx-core${NC}"
    else
        echo -e "${RED}✗ Binary not found: pryx-core${NC}"
        return 1
    fi
    
    if [ -x "$INSTALL_DIR/pryx-core" ]; then
        echo -e "${GREEN}✓ Binary is executable${NC}"
    else
        echo -e "${RED}✗ Binary is not executable${NC}"
        return 1
    fi

    if command -v pryx &> /dev/null; then
        echo -e "${GREEN}✓ pryx command found in PATH${NC}"
    else
        echo -e "${YELLOW}⚠ pryx command not in PATH${NC}"
    fi
    
    if command -v pryx-core &> /dev/null; then
        echo -e "${GREEN}✓ pryx-core command found in PATH${NC}"
    else
        echo -e "${YELLOW}⚠ pryx-core command not in PATH${NC}"
    fi

    echo -e "${GREEN}=== Verification Complete ===${NC}"
    return 0
}

# Main installation flow
main() {
    parse_args "$@"
    echo -e "${GREEN}Pryx One-Liner Installer${NC}"
    echo -e "${GREEN}Platform: ${OS_TYPE} (${ARCH_TYPE})${NC}"
    echo -e "${GREEN}Version: $VERSION${NC}"
    echo "Sovereign AI agent with local-first control center"
    echo ""
    download_pryx_core
    if [[ "${PRYX_DRY_RUN:-}" == "1" ]]; then
        echo "dry-run: update PATH for ${INSTALL_DIR}"
        exit 0
    fi

    install_to_path "$INSTALL_DIR"
    
    if verify_installation; then
        echo ""
        echo -e "${GREEN}Installation successful!${NC}"
        echo ""
        if [[ "${PRYX_NO_ONBOARD:-}" != "1" ]]; then
            echo -e "${YELLOW}To get started, run:${NC}"
            echo -e "${GREEN}  pryx${NC}"
        fi
        echo ""
        if [[ "${PRYX_NO_ONBOARD:-}" != "1" ]]; then
            echo -e "${YELLOW}To verify installation, run:${NC}"
            echo -e "${GREEN}  pryx doctor${NC}"
        fi
        echo ""
        echo -e "${YELLOW}Configuration directory:${NC}"
        echo -e "${GREEN}  $HOME/.pryx${NC}"
        exit 0
    else
        echo -e "${RED}Installation verification failed${NC}"
        exit 1
    fi
}

main "$@"
