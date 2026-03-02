# Implementation Plan: Reverse Proxy Stability Fix

**Branch**: `004-fix-proxy-stability` | **Date**: 2026-03-02 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/004-fix-proxy-stability/spec.md`

## Summary

Fix three issues causing ConnectionRefused errors in the daemon proxy path:
1. **ProxyURL bug** (root cause): `ProfileProxy.buildProviders()` ignores per-provider ProxyURL/Client settings, causing providers that need network proxies (SOCKS5/HTTP) to silently fail. All providers fail → user sees ConnectionRefused.
2. **Readiness race**: `waitForDaemonReady()` only checks the web port (19840), not the proxy port (19841), allowing the client to launch before the proxy is confirmed ready.
3. **Error reporting**: When all providers fail, the 502 response lacks structured per-provider detail for diagnostics.

Additionally: add model default fallbacks in daemon path to match the direct path.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: `net/http`, `net/url`, `golang.org/x/net/proxy` (SOCKS5), Cobra (CLI)
**Storage**: JSON config at `~/.zen/zen.json` (no schema changes needed)
**Testing**: `go test ./...` (table-driven, TDD per constitution)
**Target Platform**: macOS, Linux (CLI + daemon)
**Project Type**: CLI + daemon (reverse proxy with failover)
**Performance Goals**: Failover latency <5s, daemon startup <5s
**Constraints**: No breaking changes to config schema, no format changes to 502 errors
**Scale/Scope**: 3 files modified (~50 lines changed) + 4 test gaps filled

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. TDD (NON-NEGOTIABLE) | PASS | Tests written first for all fixes; 4 identified test gaps filled |
| II. Simplicity & YAGNI | PASS | Minimal changes — align daemon path with existing direct path code |
| III. Config Migration Safety | PASS | No config schema changes needed — ProxyURL field already exists |
| IV. Branch Protection & Commit | PASS | Feature branch, individual commits per fix |
| V. Minimal Artifacts | PASS | No summary docs created |
| VI. Test Coverage (NON-NEGOTIABLE) | PASS | Coverage targets: internal/proxy ≥80%, cmd ≥50% |

**Post-design re-check**: All gates still pass. No new entities, no schema changes, no new abstractions.

## Project Structure

### Documentation (this feature)

```text
specs/004-fix-proxy-stability/
├── plan.md              # This file
├── research.md          # Phase 0 — root cause analysis, 5 research questions resolved
├── data-model.md        # Phase 1 — no new entities, documents Provider field usage fix
├── quickstart.md        # Phase 1 — build/test/manual verification steps
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (files modified)

```text
internal/proxy/
├── profile_proxy.go       # FIX: buildProviders() — add ProxyURL + Client + model defaults
├── profile_proxy_test.go  # TEST: ProxyURL propagation, model defaults
├── server.go              # FIX: improve 502 error message detail
└── server_test.go         # TEST: validate 502 response body format

cmd/
├── daemon.go              # FIX: waitForDaemonReady() — check both ports
└── root_test.go           # TEST: buildProviders with ProxyURL
```

**Structure Decision**: Bug fix across existing files. No new packages, no structural changes. The fix aligns `profile_proxy.go` with the reference implementation in `cmd/root.go`.

## Test Gap Coverage

From the spec's test gap analysis (150+ existing tests reviewed):

| Gap | Fix Location | Test Location |
|-----|-------------|---------------|
| ProxyURL propagation in daemon path | `profile_proxy.go` | `profile_proxy_test.go` |
| buildProviders with ProxyURL (direct path) | Already correct | `root_test.go` |
| waitForDaemonReady proxy port check | `daemon.go` | `daemon.go` (test helper) |
| 502 response body format validation | `server.go` | `server_test.go` |
