#!/bin/bash

# Colors
GREEN=$(tput setaf 2 2>/dev/null || printf '\033[0;32m')
BLUE=$(tput setaf 4 2>/dev/null || printf '\033[0;34m')
RED=$(tput setaf 1 2>/dev/null || printf '\033[0;31m')
NC=$(tput sgr0 2>/dev/null || printf '\033[0m')

echo -e "${BLUE}Starting Pryx Local Development Stack...${NC}"

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
    exit 0
}

# Trap signals
trap cleanup SIGINT SIGTERM

# Start Runtime (Go)
echo -e "${GREEN}[Runtime]${NC} Starting Go Core..."
cd apps/runtime || exit 1
go run ./cmd/pryx-core &
PIDS+=($!)
cd ../..

# Start Web (Bun)
echo -e "${GREEN}[Web]${NC} Starting Web Interface..."
cd apps/web || exit 1
bun run dev &
PIDS+=($!)
cd ../..

# Wait for processes
wait
