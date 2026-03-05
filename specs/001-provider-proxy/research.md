# Research: Provider Proxy Settings

**Date**: 2026-02-28
**Feature**: 001-provider-proxy

## R1: Go HTTP Proxy Transport

**Decision**: Use `http.Transport.Proxy` for HTTP/HTTPS proxies and
`golang.org/x/net/proxy` for SOCKS5.

**Rationale**: Go's standard `net/http` package natively supports
HTTP/HTTPS proxies via `http.Transport{Proxy: http.ProxyURL(u)}`. For
SOCKS5, Go's stdlib does not include a SOCKS5 dialer, but
`golang.org/x/net/proxy` provides `proxy.SOCKS5()` which returns a
`proxy.Dialer`. This can be wrapped into a `DialContext` function on
`http.Transport`.

**Alternatives considered**:
- `proxy.FromEnvironment()`: Uses `HTTP_PROXY`/`HTTPS_PROXY` env vars.
  Rejected because we need per-provider proxy, not global env vars.
- Third-party libraries (e.g., `h12.io/socks`): Unnecessary when
  `golang.org/x/net/proxy` is the official Go sub-repo.

## R2: Per-Provider vs Shared HTTP Client

**Decision**: Create a per-provider `*http.Client` with its own
`*http.Transport` when `ProxyURL` is set. Providers without a proxy
continue to use the shared default client (current behavior).

**Rationale**: The current `ProxyServer` uses a single `*http.Client`
for all providers (`s.Client`). To support per-provider proxies, each
provider needs its own transport. The cleanest approach is to store a
`*http.Client` on the `Provider` struct itself, created during provider
construction. The `ProxyServer.forwardRequest` method then uses
`p.Client.Do(req)` instead of `s.Client.Do(req)`, falling back to
`s.Client` when `p.Client` is nil.

**Alternatives considered**:
- Transport map on ProxyServer keyed by provider name: More complex,
  harder to keep in sync with provider lifecycle.
- Single transport with per-request proxy function: The `Proxy` function
  on `http.Transport` is called per-request but cannot use different
  transports for different dial protocols (SOCKS5 needs a different
  dialer).

## R3: SOCKS5 DNS Resolution

**Decision**: `socks5://` scheme resolves DNS at the proxy (SOCKS5h
behavior).

**Rationale**: Users behind restrictive firewalls typically need DNS
resolution at the proxy side (SOCKS5h). Go's `golang.org/x/net/proxy`
SOCKS5 dialer resolves DNS at the proxy by default when using
`proxy.SOCKS5("tcp", addr, auth, proxy.Direct)` — it sends the
hostname, not a resolved IP.

**Alternatives considered**:
- Support both `socks5://` (local DNS) and `socks5h://` (proxy DNS):
  Adds complexity. SOCKS5h is the dominant use case. If needed later,
  can be added as a minor version bump.

## R4: Config Sync Exclusion Strategy

**Decision**: Clear `ProxyURL` before marshaling in `buildLocalPayload`.

**Rationale**: The sync module marshals the full `ProviderConfig` via
`json.Marshal(pc)`. To exclude `proxy_url`, we clone the provider
config and clear the `ProxyURL` field before marshaling. This is
simpler than adding a custom `MarshalJSON` with conditional field
exclusion.

**Alternatives considered**:
- Custom JSON struct tag (e.g., `-`): Would also exclude it from the
  main config file, breaking the feature.
- Separate `SyncProviderConfig` type: Over-engineering for a single
  excluded field.

## R5: Proxy URL Credential Masking for Logs

**Decision**: Mask credentials using `url.URL.Redacted()` which
replaces the password with `xxxxx`.

**Rationale**: Go's `url.URL.Redacted()` method (available since
Go 1.15) returns the URL string with the password replaced by
`xxxxx`. This is a standard approach, zero extra code, and consistent
with Go conventions.

**Alternatives considered**:
- Custom masking (e.g., `***`): Non-standard, extra code for no benefit.
- Omit entire userinfo: Loses diagnostic value (knowing that auth is
  configured is useful for debugging).

## R6: Proxy URL Validation

**Decision**: Validate on save via `url.Parse()` + scheme check.

**Rationale**: Validate when the user saves the config (in TUI/Web UI
and in `SetProvider`). Parse with `url.Parse()`, check scheme is one of
`http`, `https`, `socks5`, and verify host is non-empty. This catches
errors early rather than at request time.

**Alternatives considered**:
- Validate on use (lazy): Users would see cryptic connection errors
  instead of clear validation messages.
- Regex validation: Fragile, doesn't cover all URL edge cases.
