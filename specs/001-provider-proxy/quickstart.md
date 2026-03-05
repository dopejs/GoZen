# Quickstart: Provider Proxy Settings

**Feature**: 001-provider-proxy

## What This Feature Does

Adds an optional `proxy_url` field to each provider configuration in
GoZen. When set, all API requests to that provider are routed through
the specified proxy server (HTTP, HTTPS, or SOCKS5).

## How To Use

### 1. Edit config directly

Add `proxy_url` to a provider in `~/.zen/zen.json`:

```json
{
  "providers": {
    "work": {
      "base_url": "https://api.anthropic.com",
      "auth_token": "sk-ant-...",
      "proxy_url": "socks5://proxy.corp:1080"
    }
  }
}
```

### 2. Via TUI

Run `zen config` → edit a provider → fill in the Proxy URL field.

### 3. Via Web UI

Open the provider editor in the Web UI → enter the proxy URL.

### 4. Direct use (zen use)

`zen use work` automatically exports the correct proxy environment
variables (`HTTP_PROXY`/`HTTPS_PROXY` for HTTP proxies, `ALL_PROXY`
for SOCKS5) to the spawned CLI process.

## Supported Proxy Types

| Scheme | Example | Notes |
|--------|---------|-------|
| `http` | `http://proxy:8080` | HTTP CONNECT proxy |
| `https` | `https://proxy:8443` | HTTPS CONNECT proxy |
| `socks5` | `socks5://proxy:1080` | SOCKS5 proxy (DNS resolved at proxy) |

Embedded credentials are supported: `http://user:pass@proxy:8080`

## Key Behaviors

- **Per-provider**: Different providers can use different proxies
- **Not synced**: Proxy settings are device-local (excluded from sync)
- **Logged**: Proxy URL appears in structured logs (credentials masked)
- **Validated**: Invalid proxy URLs are rejected on save
