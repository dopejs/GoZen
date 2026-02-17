---
sidebar_position: 8
title: Config Reference
---

# Configuration Reference

## File Locations

| File | Description |
|------|-------------|
| `~/.zen/zen.json` | Main configuration file |
| `~/.zen/zend.log` | Daemon log |
| `~/.zen/zend.pid` | Daemon PID file |
| `~/.zen/logs.db` | Request log database (SQLite) |

## Full Configuration Example

```json
{
  "version": 7,
  "default_profile": "default",
  "default_client": "claude",
  "proxy_port": 19841,
  "web_port": 19840,
  "providers": {
    "anthropic": {
      "base_url": "https://api.anthropic.com",
      "auth_token": "sk-ant-xxx",
      "model": "claude-sonnet-4-5",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000"
      }
    }
  },
  "profiles": {
    "default": {
      "providers": ["anthropic"]
    }
  },
  "project_bindings": {
    "/path/to/project": {
      "profile": "work",
      "client": "codex"
    }
  }
}
```

## Field Reference

| Field | Description |
|-------|-------------|
| `version` | Config file version number |
| `default_profile` | Default profile name |
| `default_client` | Default CLI client (claude/codex/opencode) |
| `proxy_port` | Proxy server port (default: 19841) |
| `web_port` | Web management interface port (default: 19840) |
| `providers` | Provider configuration collection |
| `profiles` | Profile configuration collection |
| `project_bindings` | Project binding configuration |
| `sync` | Config sync settings (optional) |
