#!/bin/bash
# Test bot task list queries in various languages

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEV_CONFIG_DIR="$PROJECT_ROOT/.dev-config"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Get web port from config or use default
WEB_PORT=${WEB_PORT:-19842}

# Test queries in various languages
QUERIES=(
    # English
    "list"
    "tasks"
    "status"
    "what tasks are running"
    "show me the tasks"
    "any active processes"

    # Chinese
    "任务"
    "列表"
    "状态"
    "有哪些任务"
    "现在有哪些任务"
    "目前有什么任务"
    "查看任务"
    "显示进程"

    # Japanese
    "タスク"
    "タスク一覧"
    "実行中のタスク"

    # Korean
    "작업 목록"
    "실행 중인 작업"

    # Spanish
    "tareas"
    "lista de tareas"
    "mostrar tareas"
)

echo -e "${YELLOW}Testing bot task list queries...${NC}"
echo "Web port: $WEB_PORT"
echo ""

# First, register a fake session to have something to show
echo -e "${YELLOW}Registering test session...${NC}"
curl -s -X POST "http://127.0.0.1:$WEB_PORT/api/v1/daemon/sessions" \
    -H "Content-Type: application/json" \
    -d '{"session_id":"test-session-123","profile":"default","client_type":"claude-code"}' \
    > /dev/null 2>&1 || true
echo ""

SESSION_ID=""

for query in "${QUERIES[@]}"; do
    echo -e "${YELLOW}Query:${NC} $query"

    # Build request
    if [ -z "$SESSION_ID" ]; then
        REQUEST="{\"message\":\"$query\"}"
    else
        REQUEST="{\"message\":\"$query\",\"session_id\":\"$SESSION_ID\"}"
    fi

    # Send request and capture response
    RESPONSE=$(curl -s -X POST "http://127.0.0.1:$WEB_PORT/api/v1/bot/chat" \
        -H "Content-Type: application/json" \
        -d "$REQUEST" 2>&1)

    # Extract session ID from first response
    if [ -z "$SESSION_ID" ]; then
        SESSION_ID=$(echo "$RESPONSE" | grep -o '"session_id":"[^"]*"' | head -1 | cut -d'"' -f4)
    fi

    # Extract content from SSE response
    CONTENT=$(echo "$RESPONSE" | grep 'event: done' -A1 | grep 'data:' | sed 's/data: //' | jq -r '.content' 2>/dev/null || echo "$RESPONSE")

    # Check if response mentions tasks/processes
    if echo "$CONTENT" | grep -qiE '进程|任务|process|task|session|连接|idle|busy|waiting|没有'; then
        echo -e "${GREEN}✓ Response mentions tasks/processes${NC}"
    else
        echo -e "${RED}✗ Response does NOT mention tasks/processes${NC}"
    fi

    # Show first 200 chars of response
    echo -e "Response: ${CONTENT:0:200}..."
    echo ""

    # Clear session for next query to test fresh
    curl -s -X POST "http://127.0.0.1:$WEB_PORT/api/v1/bot/chat" \
        -H "Content-Type: application/json" \
        -d "{\"session_id\":\"$SESSION_ID\",\"clear\":true}" > /dev/null 2>&1

    sleep 0.5
done

echo -e "${GREEN}Done!${NC}"
