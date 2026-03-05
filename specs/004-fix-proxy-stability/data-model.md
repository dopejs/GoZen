# Data Model: Reverse Proxy Stability Fix

**Feature**: 004-fix-proxy-stability
**Date**: 2026-03-02

## Entities

No new entities introduced. This fix modifies the runtime construction of existing entities.

### Provider (existing — `internal/proxy/provider.go`)

| Field | Type | Description | Changed? |
|-------|------|-------------|----------|
| Name | string | Provider identifier | No |
| Type | string | "anthropic" or "openai" | No |
| BaseURL | *url.URL | Upstream API endpoint | No |
| Token | string | Auth token | No |
| Model | string | Default model ID | No |
| ReasoningModel | string | Reasoning model variant | No |
| HaikuModel | string | Haiku model variant | No |
| OpusModel | string | Opus model variant | No |
| SonnetModel | string | Sonnet model variant | No |
| ProxyURL | string | Network proxy URL (SOCKS5/HTTP) | **Now set in daemon path** |
| Client | *http.Client | Per-provider HTTP client | **Now created in daemon path** |
| Healthy | bool | Health state | No |
| AuthFailed | bool | Auth failure flag | No |
| FailedAt | time.Time | Last failure timestamp | No |
| Backoff | time.Duration | Current backoff duration | No |

### Changes Summary

The `ProxyURL` and `Client` fields already exist on the `Provider` struct. The fix ensures `ProfileProxy.buildProviders()` populates them when constructing providers, matching the behavior of `cmd/root.go:buildProviders()`.

No schema changes. No config migration. No new types.
