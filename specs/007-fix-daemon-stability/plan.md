# Implementation Plan: Fix Daemon Proxy Stability

**Branch**: `007-fix-daemon-stability` | **Date**: 2026-03-04 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/007-fix-daemon-stability/spec.md`

## Summary

Fix three daemon proxy stability issues: (1) pin proxy port to default 19841 with smart port conflict detection, configurable via `zen config set`; (2) add wrapper-level daemon auto-recovery when client exits due to dead daemon, coordinated via file lock; (3) fix `duration_ms` serialization to output milliseconds instead of nanoseconds in proxy monitor and bot matcher.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: Cobra (CLI), net/http (proxy/web), syscall (file lock, process management), React+TypeScript (Web UI)
**Storage**: JSON config at `~/.zen/zen.json`, SQLite for proxy logs
**Testing**: `go test ./...` (table-driven), Vitest (web UI)
**Target Platform**: macOS (primary), Linux (secondary)
**Project Type**: CLI + background daemon + web UI
**Performance Goals**: Daemon restart < 5 seconds, wrapper recovery < 10 seconds
**Constraints**: Zero manual intervention on daemon death, proxy port must never change silently
**Scale/Scope**: Single-user local tool, ~5 files modified in Go backend, ~2 files in web frontend

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. TDD (NON-NEGOTIABLE) | PASS | Tests will be written first for port pinning, file lock, duration fix, and wrapper recovery logic |
| II. Simplicity & YAGNI | PASS | Minimal changes: wrapper retry (not transparent proxy), file lock (not distributed lock), simple `zen config set` (not full settings framework) |
| III. Config Migration Safety | PASS | No schema change needed — `proxy_port` field already exists in `OpenCCConfig`. The fix ensures it's always persisted on first start, but the JSON shape is unchanged. No version bump required. |
| IV. Branch Protection | PASS | Working on feature branch `007-fix-daemon-stability`, will PR to main |
| V. Minimal Artifacts | PASS | No summary docs created. Spec artifacts are in `specs/` as required by speckit workflow |
| VI. Test Coverage (NON-NEGOTIABLE) | PASS | Must maintain: daemon ≥ 50%, proxy ≥ 80%, config ≥ 80%, web ≥ 80%, bot ≥ 80% |

**Gate result: PASS** — no violations, no complexity tracking needed.

## Project Structure

### Documentation (this feature)

```text
specs/007-fix-daemon-stability/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
cmd/
├── root.go              # MODIFY: add wrapper retry loop in startViaDaemon()
├── config.go            # MODIFY: add `zen config set` subcommand
└── daemon.go            # MODIFY: add file lock to startDaemonBackground()

internal/
├── config/
│   ├── config.go        # READ-ONLY: DefaultProxyPort constant (no changes)
│   ├── store.go         # MODIFY: add EnsureProxyPort() to persist default on first run
│   └── compat.go        # MODIFY: add EnsureProxyPort() convenience wrapper
├── daemon/
│   ├── daemon.go         # MODIFY: add file lock functions, port conflict detection
│   ├── server.go         # MODIFY: call EnsureProxyPort() on Start(), smart port binding
│   └── process_unix.go   # MODIFY: add process identification for port conflict detection
├── proxy/
│   └── request_monitor.go # MODIFY: fix Duration serialization (custom MarshalJSON or int64 ms field)
└── bot/
    └── matcher.go         # MODIFY: fix Duration serialization (same pattern)

web/
└── src/
    ├── pages/settings/tabs/GeneralSettings.tsx  # MODIFY: add read-only proxy port display
    └── types/api.ts                             # MODIFY: ensure proxy_port in Settings type

internal/web/
└── api_settings.go      # MODIFY: include proxy_port in settingsResponse (read-only)
```

**Structure Decision**: Existing single-project Go structure. No new packages needed — all changes fit cleanly into existing `cmd/`, `internal/daemon/`, `internal/config/`, `internal/proxy/`, `internal/bot/`, and `web/` directories.

## Constitution Re-Check (Post-Design)

| Principle | Status | Notes |
|-----------|--------|-------|
| I. TDD | PASS | Duration fix has clear test: serialize → check value is ms not ns. Port pinning: start/stop/start → assert same port. File lock: concurrent goroutines → assert single startup. Wrapper retry: mock dead daemon → assert re-launch. |
| II. Simplicity & YAGNI | PASS | `int64` field (not custom type), `syscall.Flock` (not distributed lock), `lsof` shell-out (not /proc parsing), whitelist config set (not generic JSON mutation). All minimal. |
| III. Config Migration | PASS | No schema change confirmed. `proxy_port` field already exists. `EnsureProxyPort()` only writes when value is 0 — no migration needed. |
| IV. Branch Protection | PASS | All work on `007-fix-daemon-stability` branch. |
| V. Minimal Artifacts | PASS | Spec artifacts only. No summary docs. |
| VI. Test Coverage | PASS | Affected packages: `daemon` (50%), `proxy` (80%), `config` (80%), `bot` (80%), `web` (80%). Tests planned for all modified code. |

**Post-design gate: PASS**

## Complexity Tracking

> No violations — table intentionally left empty.
