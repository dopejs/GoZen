#!/bin/sh
set -eu

# opencc installer - Claude Code environment switcher
# Usage:
#   ./install.sh              Install opencc
#   ./install.sh --uninstall  Remove opencc (preserves ~/.cc_envs)

BIN_TARGET="/usr/local/bin/opencc"
CC_ENVS_DIR="$HOME/.cc_envs"

# --- Helpers ---

info()  { printf "\033[1;34m==>\033[0m %s\n" "$1"; }
ok()    { printf "\033[1;32m==>\033[0m %s\n" "$1"; }
warn()  { printf "\033[1;33mWarning:\033[0m %s\n" "$1"; }
err()   { printf "\033[1;31mError:\033[0m %s\n" "$1" >&2; }

need_sudo() {
  if [ "$(id -u)" -ne 0 ]; then
    if command -v sudo >/dev/null 2>&1; then
      echo "sudo"
    else
      err "Need root privileges. Run with sudo or as root."
      exit 1
    fi
  fi
}

# Resolve script directory (works for local install, not curl|sh)
SCRIPT_DIR=""
resolve_script_dir() {
  if [ -f "${0:-}" ] && [ -d "$(dirname "$0")" ]; then
    SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
  fi
}

# --- Completion install paths ---

zsh_comp_dir() {
  # Ask interactive zsh for its actual fpath (includes oh-my-zsh etc.)
  # Skip SIP-protected /usr/share
  if command -v zsh >/dev/null 2>&1; then
    _fpath_list="$(zsh -ic 'for d in $fpath; do echo "$d"; done' 2>/dev/null)"
    for _d in $_fpath_list; do
      case "$_d" in
        /usr/share/*) continue ;;
        */site-functions)
          echo "$_d"
          return 0
          ;;
      esac
    done
  fi
  echo "/usr/local/share/zsh/site-functions"
}

bash_comp_dir() {
  if [ "$(uname -s)" = "Darwin" ]; then
    if [ -d "/opt/homebrew/etc/bash_completion.d" ]; then
      echo "/opt/homebrew/etc/bash_completion.d"
    elif [ -d "/usr/local/etc/bash_completion.d" ]; then
      echo "/usr/local/etc/bash_completion.d"
    fi
  else
    if [ -d "/usr/share/bash-completion/completions" ]; then
      echo "/usr/share/bash-completion/completions"
    elif [ -d "/etc/bash_completion.d" ]; then
      echo "/etc/bash_completion.d"
    fi
  fi
}

fish_comp_dir() {
  echo "$HOME/.config/fish/completions"
}

# --- Install ---

install_main_script() {
  SUDO="$(need_sudo)"
  info "Installing opencc to $BIN_TARGET"
  if [ -n "$SCRIPT_DIR" ] && [ -f "$SCRIPT_DIR/opencc" ]; then
    $SUDO cp "$SCRIPT_DIR/opencc" "$BIN_TARGET"
  else
    err "Cannot find opencc script. Run install.sh from the project directory."
    exit 1
  fi
  $SUDO chmod +x "$BIN_TARGET"
  ok "Installed $BIN_TARGET"
}

install_completions() {
  SUDO="$(need_sudo)"

  # zsh
  if command -v zsh >/dev/null 2>&1; then
    zdir="$(zsh_comp_dir)"
    if [ -n "$zdir" ]; then
      if [ -n "$SCRIPT_DIR" ] && [ -f "$SCRIPT_DIR/completions/_opencc" ]; then
        info "Installing zsh completion to $zdir/_opencc"
        $SUDO mkdir -p "$zdir"
        $SUDO cp "$SCRIPT_DIR/completions/_opencc" "$zdir/_opencc"
        # Clear zsh completion cache so new completion is discovered
        rm -f "$HOME"/.zcompdump*
        ok "zsh completion installed"
      else
        warn "completions/_opencc not found, skipping zsh completion"
      fi
    else
      warn "zsh site-functions directory not found, skipping zsh completion"
    fi
  fi

  # bash
  if command -v bash >/dev/null 2>&1; then
    bdir="$(bash_comp_dir)"
    if [ -n "$bdir" ]; then
      if [ -n "$SCRIPT_DIR" ] && [ -f "$SCRIPT_DIR/completions/opencc.bash" ]; then
        info "Installing bash completion to $bdir/opencc"
        $SUDO cp "$SCRIPT_DIR/completions/opencc.bash" "$bdir/opencc"
        ok "bash completion installed"
      else
        warn "completions/opencc.bash not found, skipping bash completion"
      fi
    else
      warn "bash-completion directory not found, skipping bash completion"
    fi
  fi

  # fish
  if command -v fish >/dev/null 2>&1; then
    fdir="$(fish_comp_dir)"
    if [ -n "$SCRIPT_DIR" ] && [ -f "$SCRIPT_DIR/completions/opencc.fish" ]; then
      info "Installing fish completion to $fdir/opencc.fish"
      mkdir -p "$fdir"
      cp "$SCRIPT_DIR/completions/opencc.fish" "$fdir/opencc.fish"
      ok "fish completion installed"
    else
      warn "completions/opencc.fish not found, skipping fish completion"
    fi
  fi
}

create_envs_dir() {
  if [ ! -d "$CC_ENVS_DIR" ]; then
    info "Creating $CC_ENVS_DIR"
    mkdir -p "$CC_ENVS_DIR"
    ok "Created $CC_ENVS_DIR"
  else
    info "$CC_ENVS_DIR already exists"
  fi
}

# --- Interactive first config ---

setup_first_config() {
  # Skip if configs already exist or if stdin is not a terminal
  for f in "$CC_ENVS_DIR"/*.env; do
    [ -f "$f" ] && return 0
  done

  if [ ! -t 0 ]; then
    info "Run interactively to create your first configuration"
    return 0
  fi

  printf "\n"
  info "No configurations found. Let's create your first one."
  printf "\n"

  printf "  Config name (e.g. work, personal): "
  read -r config_name
  if [ -z "$config_name" ]; then
    warn "Skipped config creation"
    return 0
  fi

  printf "  ANTHROPIC_BASE_URL: "
  read -r base_url

  printf "  ANTHROPIC_AUTH_TOKEN: "
  read -r auth_token

  printf "  ANTHROPIC_MODEL [claude-sonnet-4-20250514]: "
  read -r model
  model="${model:-claude-sonnet-4-20250514}"

  env_file="$CC_ENVS_DIR/${config_name}.env"
  cat > "$env_file" <<ENVEOF
ANTHROPIC_BASE_URL=$base_url
ANTHROPIC_AUTH_TOKEN=$auth_token
ANTHROPIC_MODEL=$model
ENVEOF

  ok "Created $env_file"
  printf "  Try it: opencc %s\n" "$config_name"
}

# --- Uninstall ---

do_uninstall() {
  SUDO="$(need_sudo)"
  info "Uninstalling opencc..."

  # Remove main script
  if [ -f "$BIN_TARGET" ]; then
    $SUDO rm -f "$BIN_TARGET"
    ok "Removed $BIN_TARGET"
  fi

  # Remove zsh completion
  if command -v zsh >/dev/null 2>&1; then
    zdir="$(zsh_comp_dir)"
    if [ -n "$zdir" ] && [ -f "$zdir/_opencc" ]; then
      $SUDO rm -f "$zdir/_opencc"
      ok "Removed $zdir/_opencc"
    fi
  fi

  # Remove bash completion
  if command -v bash >/dev/null 2>&1; then
    bdir="$(bash_comp_dir)"
    if [ -n "$bdir" ] && [ -f "$bdir/opencc" ]; then
      $SUDO rm -f "$bdir/opencc"
      ok "Removed $bdir/opencc"
    fi
  fi

  # Remove fish completion
  fdir="$(fish_comp_dir)"
  if [ -f "$fdir/opencc.fish" ]; then
    rm -f "$fdir/opencc.fish"
    ok "Removed $fdir/opencc.fish"
  fi

  printf "\n"
  ok "opencc has been uninstalled"
  info "~/.cc_envs/ was preserved (your configurations)"
}

# --- Main ---

resolve_script_dir

case "${1:-}" in
  --uninstall)
    do_uninstall
    exit 0
    ;;
  "")
    ;;
  *)
    err "Unknown option: $1"
    echo "Usage: $0 [--uninstall]" >&2
    exit 1
    ;;
esac

printf "\n"
info "Installing opencc - Claude Code environment switcher"
printf "\n"

install_main_script
create_envs_dir
install_completions
setup_first_config

printf "\n"
ok "Installation complete!"
printf "  Run 'opencc -l' to list configurations\n"
printf "  Run 'opencc <name>' to start claude with a configuration\n"