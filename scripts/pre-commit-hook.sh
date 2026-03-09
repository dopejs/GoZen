#!/bin/sh
# Pre-commit hook to validate branch names
# Format: <type>/<description>
# Allowed types: feat, fix, docs, refactor, test, chore, perf, ci, build, revert

# Get current branch name
branch=$(git rev-parse --abbrev-ref HEAD)

# Protected branches (skip validation)
if [ "$branch" = "main" ] || [ "$branch" = "master" ] || [ "$branch" = "develop" ]; then
    exit 0
fi

# Skip validation during rebase/merge
if [ -d .git/rebase-merge ] || [ -d .git/rebase-apply ] || [ -f .git/MERGE_HEAD ]; then
    exit 0
fi

# Validate branch name format
valid_pattern="^(feat|fix|docs|refactor|test|chore|perf|ci|build|revert)/[a-z0-9-]+$"

if ! echo "$branch" | grep -qE "$valid_pattern"; then
    echo ""
    echo "❌ Branch name validation failed"
    echo ""
    echo "Branch: $branch"
    echo ""
    echo "Error: Branch name does not match the required pattern: <type>/<description>"
    echo ""
    echo "Valid branch name format: <type>/<description>"
    echo "Allowed types: feat, fix, docs, refactor, test, chore, perf, ci, build, revert"
    echo ""
    echo "Examples:"
    echo "  - feat/user-authentication"
    echo "  - fix/payment-processing-bug"
    echo "  - docs/api-documentation"
    echo ""
    exit 1
fi

exit 0
