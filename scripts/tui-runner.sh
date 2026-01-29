#!/bin/bash

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m'

echo -e "${BLUE}Starting Pryx TUI + Runtime...${NC}"

# Store PIDs
PIDS=()

# Cleanup function
cleanup() {
    echo -e "\n${RED}Shutting down services...${NC}"
    for pid in "${PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid"
        fi
    done
    wait
    echo -e "${GREEN}Shutdown complete.${NC}"
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
    echo -e "${YELLOW}[Runtime]${NC} Port $RUNTIME_PORT is already in use. Runtime may already be running."
else
    # Start Runtime (Go)
    echo -e "${GREEN}[Runtime]${NC} Starting Go Core on :$RUNTIME_PORT..."
    cd apps/runtime || exit 1
    
    # Build first to catch compile errors
    if ! go build -o /tmp/pryx-core-dev ./cmd/pryx-core; then
        echo -e "${RED}[Runtime]${NC} Failed to build runtime!${NC}"
        exit 1
    fi
    
    # Run the built binary
    /tmp/pryx-core-dev &
    RUNTIME_PID=$!
    PIDS+=($RUNTIME_PID)
    cd ../..

    # Wait for Runtime to be ready
    echo -e "${BLUE}Waiting for Runtime to start...${NC}"
    for i in {1..30}; do
        if ! kill -0 "$RUNTIME_PID" 2>/dev/null; then
            echo -e "${RED}Runtime failed to start!${NC}"
            exit 1
        fi
        
        # Check if port is listening
        if lsof -Pi :$RUNTIME_PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
            echo -e "${GREEN}[Runtime]${NC} Ready on port $RUNTIME_PORT"
            break
        fi
        
        if [ $i -eq 30 ]; then
            echo -e "${YELLOW}[Runtime]${NC} Timeout waiting for port. Continuing anyway...${NC}"
        fi
        
        sleep 0.5
    done
fi

# Check if TUI binary exists and is up to date
cd apps/tui || exit 1

# Check if we need to rebuild
if [ ! -f "./pryx-tui" ] || [ "./index.tsx" -nt "./pryx-tui" ] || [ "./src/components/App.tsx" -nt "./pryx-tui" ]; then
    echo -e "${GREEN}[TUI]${NC} Building Terminal UI..."
    
    # Install dependencies if needed
    if [ ! -d "node_modules" ]; then
        echo -e "${BLUE}[TUI]${NC} Installing dependencies...${NC}"
        bun install --silent
    fi
    
    # Build with error output visible
    if ! bun run build 2>&1; then
        echo -e "${RED}[TUI]${NC} Build failed!${NC}"
        cd ../..
        cleanup
        exit 1
    fi
    echo -e "${GREEN}[TUI]${NC} Build successful"
else
    echo -e "${GREEN}[TUI]${NC} Using existing build"
fi

# Start TUI
echo -e "${GREEN}[TUI]${NC} Starting Terminal UI..."
cd ../..
./apps/tui/pryx-tui
