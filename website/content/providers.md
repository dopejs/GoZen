---
sidebar_position: 2
title: Providers
---

# Provider Management

A provider represents an API endpoint configuration including base URL, auth token, model name, and more.

## Configuration Example

```json
{
  "providers": {
    "my-provider": {
      "base_url": "https://api.example.com",
      "auth_token": "sk-xxx",
      "model": "claude-sonnet-4-5",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000",
        "MAX_THINKING_TOKENS": "50000"
      },
      "codex_env_vars": {
        "CODEX_SOME_VAR": "value"
      },
      "opencode_env_vars": {
        "OPENCODE_EXPERIMENTAL_OUTPUT_TOKEN_MAX": "64000"
      }
    }
  }
}
```

## Environment Variables

Each provider can have per-CLI environment variables:

### Common Claude Code Environment Variables

| Variable | Description |
|----------|-------------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | Maximum output tokens |
| `MAX_THINKING_TOKENS` | Extended thinking budget |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | Maximum context window |
| `BASH_DEFAULT_TIMEOUT_MS` | Bash default timeout |
