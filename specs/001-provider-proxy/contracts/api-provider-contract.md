# API Contract: Provider Configuration

**Date**: 2026-02-28
**Feature**: 001-provider-proxy

## REST API: Provider CRUD

The Web UI provider API endpoints at `/api/v1/providers` accept and
return the `Provider` type. The `proxy_url` field is added to the
existing schema.

### GET /api/v1/providers

Returns all providers. Each provider now includes `proxy_url`.

```json
{
  "providers": [
    {
      "name": "work",
      "type": "anthropic",
      "base_url": "https://api.anthropic.com",
      "auth_token": "sk-ant-...",
      "proxy_url": "socks5://proxy.corp:1080",
      "model": "claude-sonnet-4-5"
    }
  ]
}
```

### POST /api/v1/providers

Create or update a provider. `proxy_url` is optional.

**Request body**:
```json
{
  "name": "work",
  "base_url": "https://api.anthropic.com",
  "auth_token": "sk-ant-...",
  "proxy_url": "http://proxy:8080"
}
```

**Validation errors** (HTTP 400):
- `"proxy_url: unsupported scheme \"ftp\" (must be http, https, or socks5)"`
- `"proxy_url: invalid URL format"`
- `"proxy_url: missing host"`

### PUT /api/v1/providers/:name

Same request/response contract as POST.

## CLI Contract: zen use

`zen use <provider>` sets these environment variables when the provider
has `proxy_url` configured:

| Proxy Scheme | Environment Variables Set |
|-------------|--------------------------|
| `http://...` | `HTTP_PROXY`, `HTTPS_PROXY` |
| `https://...` | `HTTP_PROXY`, `HTTPS_PROXY` |
| `socks5://...` | `ALL_PROXY` |

When `proxy_url` is empty, no proxy environment variables are set
(unchanged behavior).

## Config Sync Contract

The `proxy_url` field is **excluded** from sync payloads. When a
remote config is pulled, local `proxy_url` values are preserved.
The sync payload `ProviderConfig` representation omits the field.
