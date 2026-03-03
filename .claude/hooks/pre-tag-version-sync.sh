#!/bin/bash
# Pre-tool hook: ensure cmd/root.go Version matches the git tag being created.
# If mismatched, update root.go, stage + commit, then allow the tag.
set -euo pipefail

INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

# Only intercept "git tag v*" commands
if ! echo "$COMMAND" | grep -qE '^git tag v[0-9]'; then
  exit 0
fi

# Extract the tag version (strip leading "v")
TAG_VERSION=$(echo "$COMMAND" | grep -oE 'v[0-9][^ ]*' | head -1 | sed 's/^v//')
if [ -z "$TAG_VERSION" ]; then
  exit 0
fi

CWD=$(echo "$INPUT" | jq -r '.cwd // empty')
ROOT_GO="${CWD}/cmd/root.go"

if [ ! -f "$ROOT_GO" ]; then
  exit 0
fi

# Read current version from root.go
CURRENT=$(sed -n 's/^var Version = "\(.*\)"/\1/p' "$ROOT_GO")

if [ "$CURRENT" = "$TAG_VERSION" ]; then
  exit 0
fi

# Version mismatch — update root.go
sed -i.bak "s/var Version = \".*\"/var Version = \"${TAG_VERSION}\"/" "$ROOT_GO"
rm -f "${ROOT_GO}.bak"

# Stage and commit the change
git -C "$CWD" add cmd/root.go
git -C "$CWD" commit -m "chore: bump version to ${TAG_VERSION}"

# Let Claude know what happened
cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "allow"
  },
  "message": "Updated cmd/root.go version from ${CURRENT} to ${TAG_VERSION} and committed."
}
EOF
