#!/usr/bin/env bash
# Pryx One-Liner Installer
# Usage: curl -fsSL https://pryx.dev/install | bash
#
set -e

log()   { echo '[pryx] '; }
warn()  { echo '[pryx] '; }
error() { echo '[pryx] ERROR: ' >&2; exit 1; }

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

get_latest_version() {
    curl -fsSL "https://api.github.com/repos/pryx-dev/pryx/releases/latest" 2>/dev/null | grep '"tag_name"' | sed -E 's/.*"v?([^"]+)".*/\1/' | head -1
}

install_pryx() {
    local platform version install_dir download_url tmp_dir

    command -v curl >/dev/null 2>&1 || error "Missing required command: curl"
    command -v tar >/dev/null 2>&1 || error "Missing required command: tar"

    platform="$(detect_platform)"
    version="${PRYX_VERSION:-$(get_latest_version)}"
    install_dir="${PRYX_INSTALL_DIR:-$HOME/.local/bin}"
    download_url="https://github.com/pryx-dev/pryx/releases/download/v${version}/pryx-${platform}.tar.gz"

    log "Installing Pryx v${version} for ${platform}..."
    log "Install directory: ${install_dir}"

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
    install -m 755 "$tmp_dir/pryx" "$install_dir/pryx" 2>/dev/null || cp "$tmp_dir/pryx" "$install_dir/pryx" && chmod +x "$install_dir/pryx"
    
    if [[ -f "$tmp_dir/pryx-core" ]]; then
        install -m 755 "$tmp_dir/pryx-core" "$install_dir/pryx-core" 2>/dev/null || cp "$tmp_dir/pryx-core" "$install_dir/pryx-core" && chmod +x "$install_dir/pryx-core"
    fi
    
    log "Pryx installed successfully!"
    echo "  Run 'pryx' to get started"
    echo "  Run 'pryx doctor' to verify installation"
}

echo "Pryx Installer"
echo "Sovereign AI agent with local-first control center"
echo ""
install_pryx
