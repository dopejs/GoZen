#!/bin/bash
# Install Git hooks for GoZen development

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HOOKS_DIR="$REPO_ROOT/.git/hooks"

echo "Installing Git hooks..."

# Install pre-commit hook
cp "$SCRIPT_DIR/pre-commit-hook.sh" "$HOOKS_DIR/pre-commit"
chmod +x "$HOOKS_DIR/pre-commit"

echo "✓ Pre-commit hook installed successfully"
echo ""
echo "Branch naming format: <type>/<description>"
echo "Allowed types: feat, fix, docs, refactor, test, chore, perf, ci, build, revert"
echo ""
echo "Examples:"
echo "  - feat/user-authentication"
echo "  - fix/payment-processing-bug"
echo "  - docs/api-documentation"
