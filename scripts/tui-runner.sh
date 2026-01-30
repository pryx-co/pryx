#!/bin/bash

GREEN=$(tput setaf 2 2>/dev/null || printf '\033[0;32m')
BLUE=$(tput setaf 4 2>/dev/null || printf '\033[0;34m')
RED=$(tput setaf 1 2>/dev/null || printf '\033[0;31m')
YELLOW=$(tput setaf 3 2>/dev/null || printf '\033[0;33m')
NC=$(tput sgr0 2>/dev/null || printf '\033[0m')

echo -e "${BLUE}Starting Pryx TUI + Runtime...${NC}"

PIDS=()

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

PORT_FILE="$HOME/.pryx/runtime.port"
if [ -f "$PORT_FILE" ]; then
    RUNTIME_PID=$(lsof -t "$PORT_FILE" 2>/dev/null | head -1)
    if [ -n "$RUNTIME_PID" ] && kill -0 "$RUNTIME_PID" 2>/dev/null; then
        echo -e "${YELLOW}[Runtime]${NC} Runtime appears to already be running (PID: $RUNTIME_PID)"
        echo -e "${YELLOW}[Runtime]${NC} Port file exists at $PORT_FILE"
        RUNTIME_ALREADY_RUNNING=1
    fi
fi

if [ -z "$RUNTIME_ALREADY_RUNNING" ]; then
    rm -f "$PORT_FILE"
    
    echo -e "${GREEN}[Runtime]${NC} Starting Go Core with dynamic port allocation..."
    cd apps/runtime || exit 1
    
    if ! go build -o /tmp/pryx-core-dev ./cmd/pryx-core; then
        echo -e "${RED}[Runtime]${NC} Failed to build runtime!${NC}"
        exit 1
    fi
    
    mkdir -p "$HOME/.pryx/logs"
    /tmp/pryx-core-dev > "$HOME/.pryx/logs/runtime.log" 2>&1 &
    RUNTIME_PID=$!
    PIDS+=($RUNTIME_PID)
    cd ../..

    echo -e "${BLUE}Waiting for Runtime to start and allocate port...${NC}"
    for i in {1..30}; do
        if ! kill -0 "$RUNTIME_PID" 2>/dev/null; then
            echo -e "${RED}Runtime failed to start!${NC}"
            exit 1
        fi
        
        if [ -f "$PORT_FILE" ]; then
            RUNTIME_PORT=$(cat "$PORT_FILE" 2>/dev/null)
            if [ -n "$RUNTIME_PORT" ]; then
                echo -e "${GREEN}[Runtime]${NC} Ready on port $RUNTIME_PORT"
                break
            fi
        fi
        
        if [ $i -eq 30 ]; then
            echo -e "${YELLOW}[Runtime]${NC} Timeout waiting for port file. Continuing anyway...${NC}"
        fi
        
        sleep 0.5
    done
fi

if [ -f "$PORT_FILE" ]; then
    RUNTIME_PORT=$(cat "$PORT_FILE" 2>/dev/null)
    if [ -n "$RUNTIME_PORT" ]; then
        export PRYX_API_URL="http://localhost:${RUNTIME_PORT}"
        export PRYX_WS_URL="ws://localhost:${RUNTIME_PORT}/ws"
        echo -e "${BLUE}Runtime configured:${NC}"
        echo -e "  API: ${PRYX_API_URL}"
        echo -e "  WebSocket: ${PRYX_WS_URL}"
    fi
fi

cd apps/tui || exit 1

if [ ! -d "node_modules" ]; then
    echo -e "${BLUE}[TUI]${NC} Installing dependencies...${NC}"
    bun install --silent
fi

echo -e "${GREEN}[TUI]${NC} Starting Terminal UI (development mode)..."
echo -e "${BLUE}[TUI]${NC} Note: Using 'bun run dev' for proper OpenTUI preload support${NC}"
bun run dev
