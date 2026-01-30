#!/usr/bin/env bash
# Pryx Native Bundle Build Script
# Usage: ./scripts/build-bundles.sh [target]
# Targets: macos, windows, linux, all
#
# Prerequisites:
# - Tauri CLI: cargo install tauri-cli --version "^2.0.0"
# - macOS: Xcode command line tools (xcode-select --install)
# - Linux: appimagetool for AppImage, fakeroot for DEB
# - Windows: NSIS for installers

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
HOST_DIR="$ROOT_DIR/apps/host"
RUNTIME_DIR="$ROOT_DIR/apps/runtime"

# Colors
RED=$(tput setaf 1 2>/dev/null || printf '\033[0;31m')
GREEN=$(tput setaf 2 2>/dev/null || printf '\033[0;32m')
YELLOW=$(tput setaf 3 2>/dev/null || printf '\033[0;33m')
BLUE=$(tput setaf 4 2>/dev/null || printf '\033[0;34m')
NC=$(tput sgr0 2>/dev/null || printf '\033[0m')

log() {
    echo -e "${GREEN}[BUILD]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
    exit 1
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."

    if ! command -v cargo >/dev/null 2>&1; then
        error "Rust/Cargo is required. Install from https://rustup.rs/"
    fi

    if ! command -v tauri >/dev/null 2>&1; then
        warn "Tauri CLI not found. Installing..."
        cargo install tauri-cli --version "^2.0.0" || error "Failed to install Tauri CLI"
    fi

    log "Prerequisites check passed"
}

# Setup Tauri project if needed
setup_tauri() {
    log "Setting up Tauri project..."

    if [ ! -f "${HOST_DIR}/src-tauri/src/main.rs" ]; then
        warn "Tauri project not fully initialized. Creating minimal setup..."

        # Create basic Tauri structure
        mkdir -p "${HOST_DIR}/src-tauri/src"
        mkdir -p "${HOST_DIR}/src-tauri/icons"

        # Copy icons if they exist
        if [ -d "${HOST_DIR}/icons" ]; then
            cp "${HOST_DIR}/icons"/*.png "${HOST_DIR}/src-tauri/icons/" 2>/dev/null || true
            cp "${HOST_DIR}/icons"/*.icns "${HOST_DIR}/src-tauri/icons/" 2>/dev/null || true
            cp "${HOST_DIR}/icons"/*.ico "${HOST_DIR}/src-tauri/icons/" 2>/dev/null || true
        fi

        # Create minimal main.rs
        cat > "${HOST_DIR}/src-tauri/src/main.rs" << 'EOFMAIN'
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use tauri::Manager;

#[tauri::command]
fn greet(name: &str) -> String {
    format!("Hello, {}!", name)
}

#[tauri::command]
async fn close_splashscreen(window: tauri::Window) {
    if let Some(splashscreen) = window.get_webview_window("splashscreen") {
        splashscreen.close().unwrap();
    }
    window.get_webview_window("main").unwrap().show().unwrap();
}

fn main() {
    tauri::Builder::default()
        .invoke_handler(tauri::generate_handler![greet, close_splashscreen])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
EOFMAIN

        # Copy tauri.conf.json to src-tauri
        if [ -f "${HOST_DIR}/tauri.conf.json" ]; then
            cp "${HOST_DIR}/tauri.conf.json" "${HOST_DIR}/src-tauri/"
        fi

        # Create Cargo.toml for src-tauri
        cat > "${HOST_DIR}/src-tauri/Cargo.toml" << 'EOFCARGO'
[package]
name = "pryx-host-tauri"
version = "0.1.0"
description = "Pryx Host Application"
authors = ["Pryx"]
edition = "2021"
rust-version = "1.70.0"

[dependencies]
tauri = { version = "2", features = ["tray-icon"] }
tauri-plugin-shell = "2"
tauri-plugin-opener = "2"
tauri-plugin-deep-link = "2"
tauri-plugin-dialog = "2"
tauri-plugin-notification = "2"
tauri-plugin-clipboard-manager = "2"
tauri-plugin-updater = "2"
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
tokio = { version = "1.0", features = ["full"] }
log = "0.4"
anyhow = "1.0"
libc = "0.2"
thiserror = "1.0"

[build-dependencies]
tauri-build = { version = "2", features = [] }

[features]
default = ["custom-protocol"]
custom-protocol = ["tauri/custom-protocol"]

[profile.release]
panic = "abort"
codegen-units = 1
lto = true
opt-level = "z"
strip = true
EOFCARGO

        # Create build.rs
        cat > "${HOST_DIR}/src-tauri/build.rs" << 'EOFBUILD'
fn main() {
    tauri_build::build()
}
EOFBUILD

        log "Tauri project structure created"
    else
        log "Tauri project already initialized"
    fi
}

# Build Go runtime first
build_runtime() {
    log "Building Go runtime (pryx-core)..."
    cd "$RUNTIME_DIR"

    if [ -f "go.mod" ]; then
        go build -o "$HOST_DIR/pryx-core" ./cmd/pryx-core
        log "Runtime built: $HOST_DIR/pryx-core"
    else
        warn "Go runtime not found, skipping..."
    fi
}

# Build Tauri bundles
build_tauri() {
    local target="$1"
    log "Building Tauri bundles for: $target"
    cd "$HOST_DIR/src-tauri"

    # Setup Tauri first
    setup_tauri

    if [[ "$target" == "macos" ]]; then
        cargo tauri build --bundles dmg,app
    elif [[ "$target" == "windows" ]]; then
        cargo tauri build --bundles nsis,msi
    elif [[ "$target" == "linux" ]]; then
        cargo tauri build --bundles appimage,deb
    else
        cargo tauri build
    fi

    log "Bundles created in: $HOST_DIR/target/release/bundle/"
}

# Main
TARGET="${1:-all}"

echo -e "${BLUE}"
echo "  ╔══════════════════════════════════════╗"
echo "  ║   Pryx Native Bundle Builder        ║"
echo "  ╚══════════════════════════════════════╝"
echo -e "${NC}"
echo ""

log "Starting Pryx bundle build: $TARGET"

check_prerequisites

if [[ "$TARGET" == "all" ]]; then
    log "Building bundles for all platforms..."
    echo ""
    log "Note: Cross-platform builds require additional setup:"
    echo "  - macOS: Available now (Xcode required)"
    echo "  - Windows: Requires Windows VM or cross-compilation setup"
    echo "  - Linux: Requires Linux VM or Docker"
    echo ""

    build_runtime

    if [[ "$(uname -s)" == "Darwin" ]]; then
        build_tauri "macos"
    else
        warn "Skipping macOS build (not running on macOS)"
    fi

    if [[ "$(uname -s)" == "Linux" ]]; then
        build_tauri "linux"
    else
        warn "Skipping Linux build (not running on Linux)"
    fi

    if [[ "$(uname -s)" == "MINGW"* ]] || [[ "$(uname -s)" == "MSYS"* ]]; then
        build_tauri "windows"
    else
        warn "Skipping Windows build (not running on Windows)"
    fi
else
    build_tauri "$TARGET"
fi

echo ""
log "Build complete!"
ls -la "$HOST_DIR/target/release/bundle/" 2>/dev/null || warn "Bundle directory not found (first build?)"
