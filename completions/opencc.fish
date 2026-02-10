# Fish completion for opencc - Claude Code environment switcher

# Disable file completions
complete -c opencc -f

# Complete config names only as the first argument, and only when there's a prefix
complete -c opencc -f -n '__fish_is_first_token; and test -n (commandline -ct)' \
  -a '(for f in ~/.cc_envs/*.env; basename $f .env; end 2>/dev/null)'

# Options
complete -c opencc -s l -l list -d 'List available configurations'
complete -c opencc -s h -l help -d 'Show help'
complete -c opencc -l version -d 'Show version'
complete -c opencc -l setup-completion -d 'Print shell completion code'
