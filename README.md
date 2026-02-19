# GoZen

<p align="center">
  <img src="https://raw.githubusercontent.com/dopejs/GoZen/main/assets/gozen.svg" alt="GoZen Logo" width="120">
</p>

[简体中文](docs/README.zh-CN.md) | [繁體中文](docs/README.zh-TW.md) | [Español](docs/README.es.md)

> **Go Zen** — enter a zen-like flow state for programming. **Goes Env** — seamless environment switching.

Multi-CLI environment switcher for Claude Code, Codex, and OpenCode with API proxy auto-failover.

## Features

- **Multi-CLI Support** — Supports Claude Code, Codex, and OpenCode, configurable per project
- **Multi-Config Management** — Manage all API configurations in `~/.zen/zen.json`
- **Unified Daemon** — Single `zend` process hosts both the proxy server and the Web UI
- **Proxy Failover** — Built-in HTTP proxy that automatically switches to backup providers when the primary is unavailable
- **Scenario Routing** — Intelligent routing based on request characteristics (thinking, image, longContext, etc.)
- **Project Bindings** — Bind directories to specific profiles and CLIs for project-level auto-configuration
- **Environment Variables** — Configure CLI-specific environment variables at the provider level
- **Web Management UI** — Browser-based visual management with password-protected access
- **Web UI Security** — Auto-generated access password, session-based auth, RSA encryption for token transport
- **Config Sync** — Sync providers, profiles, and settings across devices via WebDAV, S3, GitHub Gist, or GitHub Repo with AES-256-GCM encryption
- **Version Update Check** — Automatic non-blocking check for new versions on startup (24h cache)
- **Self-Update** — One-command upgrade via `zen upgrade` with semver version matching (supports prerelease versions)
- **Shell Completion** — Supports zsh / bash / fish

### v3.0 New Features

- **Usage Tracking** — Track token usage and costs per provider, model, and project
- **Budget Control** — Set daily/weekly/monthly spending limits with warn/downgrade/block actions
- **Provider Health Monitoring** — Real-time health checks with latency and error rate tracking
- **Smart Load Balancing** — Multiple strategies: failover, round-robin, least-latency, least-cost
- **Webhooks** — Notifications for budget alerts, provider status changes, and daily summaries (Slack/Discord/Generic)
- **Context Compression** — Automatic context compression when token count exceeds threshold
- **Middleware Pipeline** — Pluggable middleware for request/response transformation
- **Agent Infrastructure** — Built-in support for agent-based workflows with session management

## Installation

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh
```

Uninstall:

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/uninstall.sh | sh
```

## Quick Start

```sh
# Add your first provider
zen config add provider

# Launch (using default profile)
zen

# Use a specific profile
zen -p work

# Use a specific CLI
zen --cli codex
```

## Command Reference

| Command | Description |
|---------|-------------|
| `zen` | Launch CLI (using project binding or default config) |
| `zen -p <profile>` | Launch with a specific profile |
| `zen -p` | Interactively select a profile |
| `zen --cli <cli>` | Use a specific CLI (claude/codex/opencode) |
| `zen -y` / `zen --yes` | Auto-approve CLI permissions (claude `--permission-mode acceptEdits`, codex `-a never`) |
| `zen use <provider>` | Directly use a specific provider (no proxy) |
| `zen pick` | Interactively select a provider to launch |
| `zen list` | List all providers and profiles |
| `zen config` | Show config subcommands |
| `zen config add provider` | Add a new provider |
| `zen config add profile` | Add a new profile |
| `zen config default-client` | Set the default CLI client |
| `zen config default-profile` | Set the default profile |
| `zen config reset-password` | Reset the Web UI access password |
| `zen config sync` | Pull config from remote sync backend |
| `zen daemon start` | Start the zend daemon |
| `zen daemon stop` | Stop the daemon |
| `zen daemon restart` | Restart the daemon |
| `zen daemon status` | Show daemon status |
| `zen daemon enable` | Install daemon as system service |
| `zen daemon disable` | Uninstall daemon system service |
| `zen bind <profile>` | Bind current directory to a profile |
| `zen bind --cli <cli>` | Bind current directory to a specific CLI |
| `zen unbind` | Remove binding for current directory |
| `zen status` | Show binding status for current directory |
| `zen web` | Open the Web management UI in browser |
| `zen upgrade` | Upgrade to the latest version |
| `zen version` | Show version |

## Daemon Architecture

In v3.0, GoZen uses a unified daemon process (`zend`) that hosts both the HTTP proxy and the Web UI:

- **Proxy server** runs on port `19841` (configurable via `proxy_port`)
- **Web UI** runs on port `19840` (configurable via `web_port`)
- The daemon starts automatically when you run `zen` or `zen web`
- Config changes are hot-reloaded via file watching
- Sync auto-push (debounced 2s) and auto-pull are handled by the daemon

```sh
# Manual daemon management
zen daemon start          # Start the daemon
zen daemon stop           # Stop the daemon
zen daemon restart        # Restart the daemon
zen daemon status         # Check daemon status

# System service (auto-start on boot)
zen daemon enable         # Install as system service
zen daemon disable        # Remove system service
```

## Multi-CLI Support

zen supports three AI coding assistant CLIs:

| CLI | Description | API Format |
|-----|-------------|------------|
| `claude` | Claude Code (default) | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

### Set Default CLI

```sh
zen config default-client

# Or via Web UI
zen web  # Settings page
```

### Per-Project CLI

```sh
cd ~/work/project
zen bind --cli codex  # Use Codex for this directory
```

### Temporary CLI Override

```sh
zen --cli opencode  # Use OpenCode for this session
```

## Profile Management

A profile is an ordered list of providers used for failover.

### Configuration Example

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-main", "anthropic-backup"]
    },
    "work": {
      "providers": ["company-api"],
      "routing": {
        "think": {"providers": [{"name": "thinking-api"}]}
      }
    }
  }
}
```

### Using Profiles

```sh
# Use default profile
zen

# Use a specific profile
zen -p work

# Interactive selection
zen -p
```

## Project Bindings

Bind directories to specific profiles and/or CLIs for project-level auto-configuration.

```sh
cd ~/work/company-project

# Bind profile
zen bind work-profile

# Bind CLI
zen bind --cli codex

# Bind both
zen bind work-profile --cli codex

# Check status
zen status

# Remove binding
zen unbind
```

**Priority**: Command-line args > Project binding > Global default

## Web Management UI

```sh
# Open in browser (auto-starts daemon if needed)
zen web
```

Web UI features:
- Provider and Profile management
- Project binding management
- Global settings (default client, default profile, ports)
- Config sync settings
- Request log viewer with auto-refresh
- Model field autocomplete

### Web UI Security

When the daemon starts for the first time, it auto-generates an access password. Non-local requests (outside 127.0.0.1/::1) require login.

- **Session-based auth** with configurable expiry
- **Brute-force protection** with exponential backoff
- **RSA encryption** for sensitive token transport (API keys encrypted in-browser before sending)
- Local access (127.0.0.1) bypasses authentication

```sh
# Reset the Web UI password
zen config reset-password

# Change password via Web UI
zen web  # Settings → Change Password
```

## Config Sync

Sync providers, profiles, default profile, and default client across devices. Auth tokens are encrypted with AES-256-GCM (PBKDF2-SHA256 key derivation) before upload.

Supported backends:
- **WebDAV** — Any WebDAV server (e.g. Nextcloud, ownCloud)
- **S3** — AWS S3 or S3-compatible storage (e.g. MinIO, Cloudflare R2)
- **GitHub Gist** — Private gist (requires PAT with `gist` scope)
- **GitHub Repo** — Repository file via Contents API (requires PAT with `repo` scope)

### Setup via Web UI

```sh
zen web  # Settings → Config Sync
```

### Manual Pull via CLI

```sh
zen config sync
```

### Conflict Resolution

- Per-entity timestamp merge: newer modification wins
- Deleted entities use tombstones (expire after 30 days)
- Scalars (default profile/client): newer timestamp wins

## Environment Variables

Each provider can have CLI-specific environment variables:

```json
{
  "providers": {
    "my-provider": {
      "base_url": "https://api.example.com",
      "auth_token": "sk-xxx",
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

### Common Claude Code Environment Variables

| Variable | Description |
|----------|-------------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | Max output tokens |
| `MAX_THINKING_TOKENS` | Extended thinking budget |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | Max context window |
| `BASH_DEFAULT_TIMEOUT_MS` | Bash default timeout |

## Scenario Routing

Automatically route requests to different providers based on request characteristics:

| Scenario | Trigger Condition |
|----------|-------------------|
| `think` | Thinking mode enabled |
| `image` | Contains image content |
| `longContext` | Content exceeds threshold |
| `webSearch` | Uses web_search tool |
| `background` | Uses Haiku model |

**Fallback mechanism**: If all providers in a scenario config fail, it automatically falls back to the profile's default providers.

Configuration example:

```json
{
  "profiles": {
    "smart": {
      "providers": ["main-api"],
      "long_context_threshold": 60000,
      "routing": {
        "think": {
          "providers": [{"name": "thinking-api", "model": "claude-opus-4-5"}]
        },
        "longContext": {
          "providers": [{"name": "long-context-api"}]
        }
      }
    }
  }
}
```

## Usage Tracking & Budget Control

Track API usage and set spending limits:

```json
{
  "pricing": {
    "claude-sonnet-4-20250514": {"input_per_million": 3.0, "output_per_million": 15.0},
    "claude-opus-4-20250514": {"input_per_million": 15.0, "output_per_million": 75.0}
  },
  "budgets": {
    "daily": {"amount": 10.0, "action": "warn"},
    "monthly": {"amount": 100.0, "action": "block"},
    "per_project": true
  }
}
```

Budget actions: `warn` (log warning), `downgrade` (switch to cheaper model), `block` (reject requests).

## Provider Health Monitoring

Automatic health checks with metrics tracking:

```json
{
  "health_check": {
    "enabled": true,
    "interval_secs": 60,
    "timeout_secs": 10
  }
}
```

View provider health via Web UI or API: `GET /api/v1/health/providers`

## Smart Load Balancing

Configure load balancing strategy per profile:

```json
{
  "profiles": {
    "balanced": {
      "providers": ["provider-a", "provider-b", "provider-c"],
      "strategy": "least-latency"
    }
  }
}
```

Strategies:
- `failover` — Try providers in order (default)
- `round-robin` — Distribute requests evenly
- `least-latency` — Route to fastest provider
- `least-cost` — Route to cheapest provider for the model

## Webhooks

Get notified about important events:

```json
{
  "webhooks": [
    {
      "name": "slack-alerts",
      "url": "https://hooks.slack.com/services/xxx",
      "events": ["budget_warning", "budget_exceeded", "provider_down", "provider_up"],
      "enabled": true
    }
  ]
}
```

Events: `budget_warning`, `budget_exceeded`, `provider_down`, `provider_up`, `failover`, `daily_summary`

## Context Compression

Automatically compress context when it exceeds a threshold:

```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 100000,
    "target_ratio": 0.5
  }
}
```

## Middleware Pipeline

Transform requests and responses with pluggable middleware:

```json
{
  "middleware": {
    "enabled": true,
    "middlewares": [
      {"name": "context-injection", "enabled": true, "config": {"inject_cursorrules": true}},
      {"name": "rate-limiter", "enabled": true, "config": {"requests_per_minute": 60}}
    ]
  }
}
```

Built-in middleware: `context-injection`, `request-logger`, `rate-limiter`, `compression`

## Config Files

| File | Description |
|------|-------------|
| `~/.zen/zen.json` | Main configuration file |
| `~/.zen/zend.log` | Daemon log |
| `~/.zen/zend.pid` | Daemon PID file |
| `~/.zen/logs.db` | Request log database (SQLite) |

### Full Configuration Example

```json
{
  "version": 8,
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

## Upgrade

```sh
# Latest version
zen upgrade

# Specific version
zen upgrade 3.0
zen upgrade 3.0.0

# Prerelease version
zen upgrade 3.0.0-alpha.1
```

## Migrating from Older Versions

GoZen automatically migrates configurations from previous versions:
- `~/.opencc/opencc.json` → `~/.zen/zen.json` (from OpenCC v1.x)
- `~/.cc_envs/` → `~/.zen/zen.json` (from legacy format)

## Development

```sh
# Build
go build -o zen .

# Test
go test ./...
```

Release: Push a tag and GitHub Actions will build automatically.

```sh
git tag v3.0.0
git push origin v3.0.0
```

## License

MIT
