#!/bin/bash
# Run end-to-end integration tests for daemon stability features.
#
# These tests build the actual zen binary, start a real daemon on isolated ports,
# and exercise the quickstart.md test scenarios automatically.
#
# Usage:
#   ./scripts/run_e2e.sh           # Run all e2e tests
#   ./scripts/run_e2e.sh -run Port # Run tests matching "Port"
#   ./scripts/run_e2e.sh -v        # Verbose output

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

echo "Running end-to-end daemon stability tests..."
echo "  (building binary, starting daemon on ephemeral ports)"
echo ""

go test -tags integration -v -timeout 120s ./tests/ "$@"
