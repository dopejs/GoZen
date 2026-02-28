# Data Model: Provider Proxy Settings

**Date**: 2026-02-28
**Feature**: 001-provider-proxy

## Modified Entities

### ProviderConfig (internal/config/config.go)

Existing entity, extended with one new field.

| Field | Type | JSON Key | Required | Default | Notes |
|-------|------|----------|----------|---------|-------|
| Type | string | `type` | No | `"anthropic"` | Existing |
| BaseURL | string | `base_url` | Yes | - | Existing |
| AuthToken | string | `auth_token` | Yes | - | Existing |
| **ProxyURL** | **string** | **`proxy_url`** | **No** | **`""`** | **NEW — proxy server URL (http/https/socks5 scheme)** |
| Model | string | `model` | No | `""` | Existing |
| ReasoningModel | string | `reasoning_model` | No | `""` | Existing |
| HaikuModel | string | `haiku_model` | No | `""` | Existing |
| OpusModel | string | `opus_model` | No | `""` | Existing |
| SonnetModel | string | `sonnet_model` | No | `""` | Existing |
| EnvVars | map[string]string | `env_vars` | No | nil | Existing |
| ClaudeEnvVars | map[string]string | `claude_env_vars` | No | nil | Existing |
| CodexEnvVars | map[string]string | `codex_env_vars` | No | nil | Existing |
| OpenCodeEnvVars | map[string]string | `opencode_env_vars` | No | nil | Existing |

### Provider (internal/proxy/provider.go)

Existing entity, extended with two new fields.

| Field | Type | Required | Default | Notes |
|-------|------|----------|---------|-------|
| ProxyURL | string | No | `""` | NEW — raw proxy URL string |
| Client | *http.Client | No | nil | NEW — per-provider HTTP client (created when ProxyURL is set) |

## Validation Rules

- `proxy_url` MUST be a valid URL parseable by `url.Parse()`.
- Scheme MUST be one of: `http`, `https`, `socks5`.
- Host MUST be non-empty.
- Empty string is valid (means no proxy).
- Embedded credentials (`user:pass@host`) are supported.

## Config Version

- `CurrentConfigVersion` bumps from 8 to 9.
- No migration logic needed: the `proxy_url` field uses `omitempty`
  and defaults to empty string. Old configs without it parse correctly.
  The version bump marks the schema change for tracking purposes.

## Sync Behavior

- `proxy_url` is excluded from config sync payloads.
- When building the sync payload, the field is cleared before marshal.
- When pulling from sync, the local `proxy_url` value is preserved
  (not overwritten by the remote config which has no `proxy_url`).

## JSON Example

```json
{
  "providers": {
    "work": {
      "type": "anthropic",
      "base_url": "https://api.anthropic.com",
      "auth_token": "sk-ant-...",
      "proxy_url": "socks5://proxy.corp:1080",
      "model": "claude-sonnet-4-5"
    },
    "home": {
      "base_url": "https://api.anthropic.com",
      "auth_token": "sk-ant-...",
      "model": "claude-sonnet-4-5"
    }
  }
}
```
