# Implementation Plan: Fix Proxy Port Stability

**Branch**: `002-fix-proxy-port` | **Date**: 2026-02-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-fix-proxy-port/spec.md`

## Summary

The legacy (non-daemon) proxy startup path in `cmd/root.go` hardcodes
`127.0.0.1:0` as the listen address, causing the OS to assign a random
port on every start. The fix is to read the configured proxy port from
the config store (same logic the daemon path already uses) and bind to
that specific address. Port 0 or unset values fall back to the default
(19841).

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: `net`, `net/http`, `internal/config`, `internal/proxy`
**Storage**: JSON config at `~/.zen/zen.json`
**Testing**: `go test` with table-driven tests
**Target Platform**: macOS, Linux (cross-platform CLI)
**Project Type**: CLI tool
**Performance Goals**: N/A (bug fix, no performance change)
**Constraints**: Must not change daemon path behavior; must not change config schema
**Scale/Scope**: 1 file modified (`cmd/root.go`), ~10 lines changed

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. TDD | ✅ Pass | Tests will be written first for port resolution logic |
| II. Simplicity & YAGNI | ✅ Pass | Minimal change: replace hardcoded `:0` with config read |
| III. Config Migration | ✅ Pass | No config schema changes needed |
| IV. Branch Protection | ✅ Pass | Working on feature branch `002-fix-proxy-port` |
| V. Minimal Artifacts | ✅ Pass | No extra files created |
| VI. Test Coverage | ✅ Pass | Will verify coverage thresholds after implementation |

## Project Structure

### Documentation (this feature)

```text
specs/002-fix-proxy-port/
├── spec.md
├── plan.md              # This file
├── research.md          # Skipped (no unknowns)
└── checklists/
    └── requirements.md
```

### Source Code (affected files)

```text
cmd/root.go                      # startLegacyProxy() — bug location
internal/proxy/server.go         # StartProxy(), StartProxyWithRouting()
internal/proxy/server_test.go    # New tests for port binding
```

**Structure Decision**: Existing Go project structure. No new files
except test additions in existing test files.

## Research

No research phase needed. The root cause is identified and the fix
is straightforward:

- **Root cause**: `startLegacyProxy()` in `cmd/root.go` passes
  `"127.0.0.1:0"` to `proxy.StartProxy()` and
  `proxy.StartProxyWithRouting()`, causing OS-assigned random ports.
- **Fix**: Read `config.GetProxyPort()` and format the address as
  `fmt.Sprintf("127.0.0.1:%d", proxyPort)` before passing it.
- **Reference**: The daemon path in `internal/daemon/server.go:199`
  already does this correctly.

## Design

### Change 1: `cmd/root.go` — `startLegacyProxy()`

Replace the two hardcoded `"127.0.0.1:0"` calls with the configured
port:

```
proxyPort := config.GetProxyPort()
addr := fmt.Sprintf("127.0.0.1:%d", proxyPort)
```

Then pass `addr` to both `proxy.StartProxyWithRouting()` and
`proxy.StartProxy()`.

### Change 2: Port 0 guard in `config.GetProxyPort()`

Verify that `GetProxyPort()` already returns `DefaultProxyPort`
(19841) when the config value is 0 or unset. If it does (confirmed
from code review), no change needed here.

### No changes needed

- `internal/proxy/server.go`: `StartProxy()` and
  `StartProxyWithRouting()` already accept a `listenAddr` parameter
  and bind to it. No modification required — the bug is in the
  caller, not the callee.
- Config schema: No changes. `proxy_port` field already exists.
- Daemon path: Already correct. No changes.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
