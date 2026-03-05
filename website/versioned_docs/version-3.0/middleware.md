---
sidebar_position: 15
title: Middleware
---

# Middleware Pipeline

Transform GoZen into a programmable AI API gateway with pluggable middleware.

## Overview

Requests and responses flow through a middleware chain. Each middleware can intercept, modify, or enhance traffic.

## Configuration

```json
{
  "middleware": {
    "enabled": true,
    "middlewares": [
      {
        "name": "context-injection",
        "enabled": true,
        "config": {
          "inject_cursorrules": true,
          "inject_claude_md": true
        }
      },
      {
        "name": "rate-limiter",
        "enabled": true,
        "config": {
          "requests_per_minute": 60
        }
      }
    ]
  }
}
```

## Built-in Middleware

### context-injection

Auto-inject project context files into requests.

```json
{
  "name": "context-injection",
  "config": {
    "inject_cursorrules": true,
    "inject_claude_md": true,
    "max_inject_tokens": 2000
  }
}
```

### request-logger

Enhanced request/response logging.

```json
{
  "name": "request-logger",
  "config": {
    "log_headers": false,
    "log_body": true,
    "max_body_size": 1000
  }
}
```

### rate-limiter

Per-session rate limiting.

```json
{
  "name": "rate-limiter",
  "config": {
    "requests_per_minute": 60,
    "burst": 10
  }
}
```

### compression

Context compression as middleware (wraps the compression feature).

```json
{
  "name": "compression",
  "config": {
    "threshold_tokens": 50000
  }
}
```

## Plugin Sources

| Source | Description |
|--------|-------------|
| `builtin` | Compiled into GoZen |
| `local` | Go plugins (.so) from disk |
| `remote` | Downloaded from URL |

### Local Plugin

```json
{
  "name": "my-plugin",
  "source": "local",
  "path": "/path/to/plugin.so"
}
```

### Remote Plugin

```json
{
  "name": "community-plugin",
  "source": "remote",
  "url": "https://example.com/plugin-manifest.json"
}
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/middleware` | GET | List all middleware |
| `/api/v1/middleware` | PUT | Update middleware config |
| `/api/v1/middleware/{name}/enable` | POST | Enable middleware |
| `/api/v1/middleware/{name}/disable` | POST | Disable middleware |
| `/api/v1/middleware/reload` | POST | Reload all middleware |

## Developing Custom Middleware

Creating custom middleware plugins requires implementing the Middleware interface in Go. Documentation for middleware development is coming soon.
