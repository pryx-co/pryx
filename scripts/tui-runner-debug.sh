#!/bin/bash

# Debug version of tui-runner that logs everything to a file

# Colors
GREEN=$(tput setaf 2 2>/dev/null || printf '\033[0;32m')
BLUE=$(tput setaf 4 2>/dev/null || printf '\033[0;34m')
RED=$(tput setaf 1 2>/dev/null || printf '\033[0;31m')
YELLOW=$(tput setaf 3 2>/dev/null || printf '\033[0;33m')
NC=$(tput sgr0 2>/dev/null || printf '\033[0m')

# Log file
LOG_FILE="/tmp/pryx-tui-debug.log"
echo "=== Pryx TUI Debug Log - $(date) ===" > "$LOG_FILE"

log() {
    echo "$1" | tee -a "$LOG_FILE"
}

log_err() {
    echo "$1" | tee -a "$LOG_FILE" >&2
}

log "${BLUE}Starting Pryx TUI + Runtime (Debug Mode)...${NC}"
log "Log file: $LOG_FILE"

# Store PIDs
PIDS=()

# Cleanup function
cleanup() {
    log "\n${RED}Shutting down services...${NC}"
    for pid in "${PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid"
        fi
    done
    wait
    log "${GREEN}Shutdown complete.${NC}"
    log "Full log available at: $LOG_FILE"
}

on_exit() {
    local exit_code=$?
    trap - EXIT
    cleanup
    exit "$exit_code"
}

on_sigint() {
    trap - EXIT SIGINT SIGTERM
    cleanup
    exit 130
}

on_sigterm() {
    trap - EXIT SIGINT SIGTERM
    cleanup
    exit 143
}

trap on_exit EXIT
trap on_sigint SIGINT
trap on_sigterm SIGTERM

# Check if runtime port is already in use
RUNTIME_PORT=${PRYX_RUNTIME_PORT:-3000}
if lsof -Pi :$RUNTIME_PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
    log "${YELLOW}[Runtime]${NC} Port $RUNTIME_PORT is already in use. Runtime may already be running."
else
    # Start Runtime (Go)
    log "${GREEN}[Runtime]${NC} Starting Go Core on :$RUNTIME_PORT..."
    cd apps/runtime || exit 1
    
    # Build first to catch compile errors
    log "${BLUE}[Runtime]${NC} Building..."
    if ! go build -o /tmp/pryx-core-dev ./cmd/pryx-core 2>&1 | tee -a "$LOG_FILE"; then
        log_err "${RED}[Runtime]${NC} Failed to build runtime!${NC}"
        log_err "Check log at: $LOG_FILE"
        exit 1
    fi
    
    # Run the built binary with output to log
    log "${BLUE}[Runtime]${NC} Starting binary..."
    /tmp/pryx-core-dev 2>&1 | tee -a "$LOG_FILE" &
    RUNTIME_PID=$!
    PIDS+=($RUNTIME_PID)
    cd ../..

    # Wait for Runtime to be ready
    log "${BLUE}Waiting for Runtime to start...${NC}"
    for i in {1..30}; do
        if ! kill -0 "$RUNTIME_PID" 2>/dev/null; then
            log_err "${RED}Runtime failed to start!${NC}"
            log_err "Check log at: $LOG_FILE"
            exit 1
        fi
        
        # Check if port is listening
        if lsof -Pi :$RUNTIME_PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
            log "${GREEN}[Runtime]${NC} Ready on port $RUNTIME_PORT"
            break
        fi
        
        if [ $i -eq 30 ]; then
            log "${YELLOW}[Runtime]${NC} Timeout waiting for port. Continuing anyway...${NC}"
        fi
        
        sleep 0.5
    done
fi

# Check if TUI binary exists and is up to date
cd apps/tui || exit 1

# Check if we need to rebuild
if [ ! -f "./pryx-tui" ] || [ "./index.tsx" -nt "./pryx-tui" ] || [ "./src/components/App.tsx" -nt "./pryx-tui" ]; then
    log "${GREEN}[TUI]${NC} Building Terminal UI..."
    
    # Install dependencies if needed
    if [ ! -d "node_modules" ]; then
        log "${BLUE}[TUI]${NC} Installing dependencies...${NC}"
        bun install 2>&1 | tee -a "$LOG_FILE"
    fi
    
    # Build with full error output
    log "${BLUE}[TUI]${NC} Running bun build..."
    if ! bun run build 2>&1 | tee -a "$LOG_FILE"; then
        log_err "${RED}[TUI]${NC} Build failed!${NC}"
        log_err "Check log at: $LOG_FILE"
        cd ../..
        cleanup
        exit 1
    fi
    log "${GREEN}[TUI]${NC} Build successful"
else
    log "${GREEN}[TUI]${NC} Using existing build"
fi

# Start TUI with all output logged
log "${GREEN}[TUI]${NC} Starting Terminal UI..."
log "${YELLOW}Note: TUI output will be logged to $LOG_FILE${NC}"
cd ../..

# Run TUI but capture all output
./apps/tui/pryx-tui 2>&1 | tee -a "$LOG_FILE"
