# Implementation Plan: Reverse Proxy Stability Fix

**Branch**: `004-fix-proxy-stability` | **Date**: 2026-03-02 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/004-fix-proxy-stability/spec.md`

## Summary

Fix three issues causing ConnectionRefused errors in the daemon proxy path:
1. **ProxyURL bug**: `ProfileProxy.buildProviders()` ignores per-provider ProxyURL/Client settings, causing providers that need network proxies (SOCKS5/HTTP) to silently fail.
2. **Readiness race**: `waitForDaemonReady()` only checks the web port (19840), not the proxy port (19841), allowing the client to launch before the proxy is ready.
3. **Error reporting**: When all providers fail, the 502 response lacks structured detail for diagnostics.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: `net/http`, `net/url`, `golang.org/x/net/proxy` (SOCKS5), Cobra (CLI)
**Storage**: JSON config at `~/.zen/zen.json`
**Testing**: `go test ./...` (table-driven tests, TDD per constitution)
**Target Platform**: macOS, Linux (CLI tool)
**Project Type**: CLI + daemon (reverse proxy with failover)
**Performance Goals**: Failover latency <5s, daemon startup <5s
**Constraints**: No breaking changes to config schema (no version bump needed)
**Scale/Scope**: 3 files modified, ~50 lines changed + tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. TDD (NON-NEGOTIABLE) | PASS | Tests written first for all 3 fixes |
| II. Simplicity & YAGNI | PASS | Minimal changes — align daemon path with existing direct path |
| III. Config Migration Safety | PASS | No config schema changes needed |
| IV. Branch Protection & Commit | PASS | Feature branch, individual commits per fix |
| V. Minimal Artifacts | PASS | No summary docs created |
| VI. Test Coverage (NON-NEGOTIABLE) | PASS | Coverage targets: internal/proxy ≥80%, cmd ≥50% |

No violations — no Complexity Tracking needed.

## Project Structure

### Documentation (this feature)

```text
specs/004-fix-proxy-stability/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (files modified)

```text
internal/proxy/
├── profile_proxy.go       # FIX: buildProviders() — add ProxyURL + Client
├── profile_proxy_test.go  # NEW: test ProxyURL propagation
├── server.go              # REVIEW: tryProviders 502 error format
└── provider.go            # EXISTING: Provider struct, NewHTTPClientWithProxy

cmd/
├── daemon.go              # FIX: waitForDaemonReady() — check both ports
└── root.go                # REFERENCE: correct buildProviders() implementation

internal/daemon/
└── server.go              # REVIEW: startProxy() ordering
```

**Structure Decision**: This is a bug fix across existing files. No new packages or structural changes needed. The fix aligns `profile_proxy.go` with the reference implementation in `cmd/root.go`.
