#!/bin/sh
set -eu

# GoZen deploy script - builds binaries for GitHub Releases
# Usage:
#   ./deploy.sh                Build for current platform
#   ./deploy.sh --all          Build for all supported platforms
#
# Binaries are placed in dist/ with naming: zen-<os>-<arch>
# Upload them to GitHub Releases. install.sh will download them by this name.

VERSION="$(grep 'var Version' cmd/root.go | sed 's/.*"\(.*\)"/\1/')"
DIST_DIR="dist"
PROJECT="zen"

info()  { printf "\033[1;34m==>\033[0m %s\n" "$1"; }
ok()    { printf "\033[1;32m==>\033[0m %s\n" "$1"; }
err()   { printf "\033[1;31mError:\033[0m %s\n" "$1" >&2; }

if ! command -v go >/dev/null 2>&1; then
  err "Go compiler not found. Install Go from https://go.dev/dl/"
  exit 1
fi

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

build_binary() {
  _os="$1"
  _arch="$2"
  _out="${DIST_DIR}/${PROJECT}-${_os}-${_arch}"

  info "Building ${_os}/${_arch}..."
  GOOS="$_os" GOARCH="$_arch" go build -ldflags="-s -w" -o "$_out" .
  ok "Created ${_out}"
}

case "${1:-}" in
  --all)
    build_binary darwin  amd64
    build_binary darwin  arm64
    build_binary linux   amd64
    build_binary linux   arm64
    ;;
  *)
    _os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    _arch="$(uname -m)"
    case "$_arch" in
      x86_64)  _arch="amd64" ;;
      aarch64) _arch="arm64" ;;
    esac
    build_binary "$_os" "$_arch"
    ;;
esac

printf "\n"
ok "Build complete (v${VERSION})"
info "Upload binaries in ${DIST_DIR}/ to GitHub Releases tag v${VERSION}"
