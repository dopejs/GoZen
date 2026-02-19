#!/bin/bash
# Development script for running GoZen with isolated config
#
# Usage:
#   ./scripts/dev.sh              # Start dev daemon (ports 29840/29841)
#   ./scripts/dev.sh stop         # Stop dev daemon
#   ./scripts/dev.sh status       # Check dev daemon status
#   ./scripts/dev.sh web          # Start frontend dev server
#   ./scripts/dev.sh all          # Start both daemon and frontend

set -e

# Dev environment config
DEV_CONFIG_DIR="${HOME}/.zen-dev"
DEV_WEB_PORT=29840
DEV_PROXY_PORT=29841

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Build the binary
build_daemon() {
    echo -e "${YELLOW}Building daemon...${NC}"
    cd "$PROJECT_ROOT"
    go build -o "$PROJECT_ROOT/bin/zen-dev" .
    echo -e "${GREEN}Built: $PROJECT_ROOT/bin/zen-dev${NC}"
}

# Initialize dev config if not exists
init_config() {
    if [ ! -f "$DEV_CONFIG_DIR/zen.json" ]; then
        echo -e "${YELLOW}Initializing dev config at $DEV_CONFIG_DIR${NC}"
        mkdir -p "$DEV_CONFIG_DIR"
        cat > "$DEV_CONFIG_DIR/zen.json" << EOF
{
  "version": 6,
  "web_port": $DEV_WEB_PORT,
  "proxy_port": $DEV_PROXY_PORT,
  "providers": {},
  "profiles": {
    "default": {
      "providers": []
    }
  }
}
EOF
        echo -e "${GREEN}Dev config created. Add providers via Web UI at http://localhost:$DEV_WEB_PORT${NC}"
    fi
}

# Start dev daemon
start_daemon() {
    build_daemon
    init_config

    # Check if already running
    if [ -f "$DEV_CONFIG_DIR/zend.pid" ]; then
        PID=$(cat "$DEV_CONFIG_DIR/zend.pid")
        if kill -0 "$PID" 2>/dev/null; then
            echo -e "${YELLOW}Dev daemon already running (PID $PID)${NC}"
            return 0
        fi
    fi

    echo -e "${YELLOW}Starting dev daemon...${NC}"
    GOZEN_CONFIG_DIR="$DEV_CONFIG_DIR" "$PROJECT_ROOT/bin/zen-dev" daemon start
    echo -e "${GREEN}Dev daemon started${NC}"
    echo -e "  Web UI:  http://127.0.0.1:$DEV_WEB_PORT"
    echo -e "  Proxy:   http://127.0.0.1:$DEV_PROXY_PORT"
}

# Stop dev daemon
stop_daemon() {
    if [ ! -f "$DEV_CONFIG_DIR/zend.pid" ]; then
        echo -e "${YELLOW}Dev daemon not running${NC}"
        return 0
    fi

    echo -e "${YELLOW}Stopping dev daemon...${NC}"
    GOZEN_CONFIG_DIR="$DEV_CONFIG_DIR" "$PROJECT_ROOT/bin/zen-dev" daemon stop 2>/dev/null || true
    echo -e "${GREEN}Dev daemon stopped${NC}"
}

# Show daemon status
show_status() {
    if [ -f "$PROJECT_ROOT/bin/zen-dev" ]; then
        GOZEN_CONFIG_DIR="$DEV_CONFIG_DIR" "$PROJECT_ROOT/bin/zen-dev" daemon status
    else
        echo -e "${RED}Dev binary not built. Run: ./scripts/dev.sh${NC}"
    fi
}

# Start frontend dev server
start_web() {
    echo -e "${YELLOW}Starting frontend dev server...${NC}"
    cd "$PROJECT_ROOT/web"

    # Update vite proxy to use dev ports
    VITE_API_PORT=$DEV_WEB_PORT npm run dev
}

# Main
case "${1:-start}" in
    start)
        start_daemon
        ;;
    stop)
        stop_daemon
        ;;
    restart)
        stop_daemon
        sleep 1
        start_daemon
        ;;
    status)
        show_status
        ;;
    web)
        start_web
        ;;
    all)
        start_daemon
        echo ""
        start_web
        ;;
    build)
        build_daemon
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|web|all|build}"
        exit 1
        ;;
esac
