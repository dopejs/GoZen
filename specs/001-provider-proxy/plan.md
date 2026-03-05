# Implementation Plan: Provider Proxy Settings

**Branch**: `001-provider-proxy` | **Date**: 2026-02-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-provider-proxy/spec.md`

## Summary

Add per-provider proxy configuration (HTTP/HTTPS/SOCKS5) so GoZen can
route upstream API requests through a user-specified proxy server. The
proxy URL is stored as a new `proxy_url` field on `ProviderConfig`. The
daemon's proxy server creates per-provider `http.Transport` instances
with the configured proxy. For `zen use`, proxy env vars are exported
to the spawned CLI process. The field is excluded from config sync.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: `net/http`, `net/url`, `golang.org/x/net/proxy` (for SOCKS5)
**Storage**: JSON config at `~/.zen/zen.json`
**Testing**: `go test ./...` (table-driven tests)
**Target Platform**: macOS, Linux (cross-platform CLI)
**Project Type**: CLI + daemon + web UI
**Performance Goals**: No measurable latency increase for providers without proxy
**Constraints**: Per-provider HTTP client/transport; no shared global proxy
**Scale/Scope**: Single new field on existing config type, ~8 files modified

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. TDD (NON-NEGOTIABLE) | PASS | Tests planned for: proxy URL validation, config migration, ExportToEnv proxy vars, transport creation |
| II. Simplicity & YAGNI | PASS | Single field addition, per-provider transport is the minimum viable approach |
| III. Config Migration Safety | PASS | Will bump `CurrentConfigVersion` to 9, add `proxy_url` field with `omitempty`, backward-compatible (empty = no proxy) |
| IV. Branch Protection | PASS | Working on feature branch `001-provider-proxy` |
| V. Minimal Artifacts | PASS | No example files, no summary docs |

No violations to justify.

## Project Structure

### Documentation (this feature)

```text
specs/001-provider-proxy/
├── plan.md
├── research.md
├── data-model.md
├── contracts/
│   └── api-provider-contract.md
└── tasks.md
```

### Source Code (files to modify)

```text
internal/
├── config/
│   ├── config.go          # Add ProxyURL field to ProviderConfig, bump version
│   ├── config_test.go     # Migration tests, validation tests
│   └── store.go           # (no changes needed — unmarshal handles new field)
├── proxy/
│   ├── provider.go        # Add ProxyURL field to Provider struct
│   ├── server.go          # Per-provider http.Client with proxy transport
│   ├── healthcheck.go     # Use proxy transport for health checks
│   └── proxy_test.go      # Transport creation tests
├── sync/
│   └── manager.go         # Exclude proxy_url from sync payload
web/
├── src/
│   ├── types/api.ts       # Add proxy_url to Provider interface
│   └── pages/providers/
│       └── edit.tsx        # Add proxy URL input field
tui/
└── editor.go              # Add proxy URL field to TUI editor
cmd/
├── use.go                 # Export proxy env vars via ExportToEnv
└── root.go                # Pass ProxyURL when building proxy.Provider
```

**Structure Decision**: Existing Go project structure. All changes are
additions to existing files. No new packages or directories needed.

## Complexity Tracking

No violations to justify.
