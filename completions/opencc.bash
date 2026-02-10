# Bash completion for opencc - Claude Code environment switcher

_opencc_complete() {
  local cur="${COMP_WORDS[COMP_CWORD]}"

  # Only complete the first argument (config name)
  if [ "$COMP_CWORD" -eq 1 ]; then
    # Don't complete when input is empty
    [ -z "$cur" ] && return

    local configs=""
    for f in "$HOME/.cc_envs"/*.env; do
      [ -f "$f" ] && configs="$configs $(basename "$f" .env)"
    done
    COMPREPLY=($(compgen -W "$configs" -- "$cur"))
  fi
}

complete -F _opencc_complete opencc
