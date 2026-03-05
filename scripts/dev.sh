#!/bin/bash
# Development script for running GoZen with isolated config
#
# Usage:
#   ./scripts/dev.sh              # Start dev daemon (ports 29840/29841)
#   ./scripts/dev.sh stop         # Stop dev daemon
#   ./scripts/dev.sh status       # Check dev daemon status
#   ./scripts/dev.sh web          # Start frontend dev server
#   ./scripts/dev.sh all          # Start both daemon and frontend
#   ./scripts/dev.sh zen [args]   # Run zen command with dev config (e.g., zen daemon start)

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
    echo -e "${YELLOW}Building frontend...${NC}"
    cd "$PROJECT_ROOT/web"
    npm run build
    cd "$PROJECT_ROOT"

    echo -e "${YELLOW}Building daemon...${NC}"
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

# Start a client with dev proxy
start_client() {
    local client="${1:-claude}"
    local extra_args="${@:2}"

    echo -e "${YELLOW}Starting $client with dev proxy...${NC}"
    echo -e "  Proxy: http://127.0.0.1:$DEV_PROXY_PORT"
    echo ""

    case "$client" in
        claude)
            ANTHROPIC_BASE_URL="http://127.0.0.1:$DEV_PROXY_PORT" claude $extra_args
            ;;
        codex)
            OPENAI_BASE_URL="http://127.0.0.1:$DEV_PROXY_PORT" codex $extra_args
            ;;
        opencode)
            OPENCODE_API_BASE="http://127.0.0.1:$DEV_PROXY_PORT" opencode $extra_args
            ;;
        *)
            echo -e "${RED}Unknown client: $client${NC}"
            echo "Supported clients: claude, codex, opencode"
            exit 1
            ;;
    esac
}

# Run zen command with dev config
run_zen() {
    # Build if binary doesn't exist
    if [ ! -f "$PROJECT_ROOT/bin/zen-dev" ]; then
        build_daemon
    fi
    init_config

    GOZEN_CONFIG_DIR="$DEV_CONFIG_DIR" "$PROJECT_ROOT/bin/zen-dev" "$@"
}

# Run bot test harness
run_bot_test() {
    echo -e "${YELLOW}Building bot test harness...${NC}"
    cd "$PROJECT_ROOT"
    go build -o "$PROJECT_ROOT/bin/bottest" ./cmd/bottest
    echo -e "${GREEN}Starting bot chat test...${NC}"
    echo ""
    GOZEN_CONFIG_DIR="$DEV_CONFIG_DIR" "$PROJECT_ROOT/bin/bottest"
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
    zen)
        run_zen "${@:2}"
        ;;
    bot)
        run_bot_test
        ;;
    client)
        start_client "${@:2}"
        ;;
    claude)
        start_client claude "${@:2}"
        ;;
    codex)
        start_client codex "${@:2}"
        ;;
    opencode)
        start_client opencode "${@:2}"
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|web|all|build|zen|bot|client|claude|codex|opencode}"
        echo ""
        echo "Commands:"
        echo "  start     Start dev daemon (ports $DEV_WEB_PORT/$DEV_PROXY_PORT)"
        echo "  stop      Stop dev daemon"
        echo "  restart   Restart dev daemon"
        echo "  status    Check dev daemon status"
        echo "  web       Start frontend dev server"
        echo "  all       Start both daemon and frontend"
        echo "  build     Build dev binary only"
        echo "  zen       Run zen command with dev config (e.g., zen daemon start)"
        echo "  bot       Run bot chat test harness"
        echo "  client    Start a client with dev proxy (e.g., client claude)"
        echo "  claude    Shortcut for 'client claude'"
        echo "  codex     Shortcut for 'client codex'"
        echo "  opencode  Shortcut for 'client opencode'"
        exit 1
        ;;
esac
