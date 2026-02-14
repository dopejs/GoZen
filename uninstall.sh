#!/bin/sh
set -eu
# GoZen uninstaller - delegates to install.sh --uninstall
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/uninstall.sh | sh

SCRIPT_URL="https://raw.githubusercontent.com/dopejs/gozen/main/install.sh"

if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$SCRIPT_URL" | sh -s -- --uninstall
elif command -v wget >/dev/null 2>&1; then
  wget -qO- "$SCRIPT_URL" | sh -s -- --uninstall
else
  printf "\033[1;31mError:\033[0m curl or wget is required\n" >&2
  exit 1
fi
