# Implementation Plan: Daemon Proxy Stability Improvements

**Branch**: `017-proxy-stability` | **Date**: 2026-03-08 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/017-proxy-stability/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Improve daemon proxy stability to eliminate crashes, memory leaks, and resource exhaustion that currently force users to rely on direct mode as an "escape hatch." This is a prerequisite for v3.1.0 dynamic switching features. The approach includes panic recovery middleware, comprehensive health monitoring, structured logging, automatic restart mechanisms, connection pool management, and concurrency limits. Success is measured by 24-hour uptime, stable memory/goroutine counts, and graceful handling of 100 concurrent requests.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: net/http (stdlib), runtime (metrics), debug (stack traces), existing internal packages (config, proxy, daemon, web)
**Storage**: JSON config at ~/.zen/zen.json (existing), in-memory metrics (no persistence)
**Testing**: Go testing (table-driven tests), httptest for HTTP handlers, existing test infrastructure
**Target Platform**: macOS, Linux, Windows (cross-platform daemon)
**Project Type**: CLI tool with embedded daemon (proxy + web server)
**Performance Goals**: P99 latency <100ms, 100 concurrent requests sustained for 5 minutes, <10% memory growth over 24 hours
**Constraints**: No external dependencies beyond stdlib, backward compatible with existing config schema, dev/prod port isolation (29840/29841 vs 19840/19841)
**Scale/Scope**: Single-user daemon, 10-50 requests/hour normal load, 100 concurrent peak load, 24+ hour uptime target

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Initial Check (Before Phase 0) ✅ PASS

All gates passed. See initial evaluation above.

### Post-Design Check (After Phase 1) ✅ PASS

**Re-evaluated after completing research.md, data-model.md, contracts/api.md, and quickstart.md**

### Principle I: Test-Driven Development ✅ PASS
- **Status**: COMPLIANT
- **Design Impact**: All new components (metrics, logger, limiter, health API) have clear test strategies documented in quickstart.md
- **Action**: TDD examples provided for each component with table-driven test patterns

### Principle II: Simplicity & YAGNI ✅ PASS
- **Status**: COMPLIANT
- **Design Impact**: Research decisions favor stdlib over external dependencies (no zap/zerolog, no Prometheus, in-memory metrics)
- **Validation**: No speculative features added. Metrics are simple counters and ring buffers, not complex time-series databases.

### Principle III: Config Migration Safety ✅ PASS
- **Status**: COMPLIANT
- **Design Impact**: Confirmed no config schema changes. All improvements are runtime-only.
- **Validation**: data-model.md shows all entities are in-memory, no JSON schema modifications.

### Principle IV: Branch Protection & Commit Discipline ✅ PASS
- **Status**: COMPLIANT
- **Design Impact**: Project structure shows clear component boundaries for atomic commits
- **Validation**: quickstart.md documents commit strategy (one component per commit)

### Principle V: Minimal Artifacts ✅ PASS
- **Status**: COMPLIANT
- **Design Impact**: All design artifacts in specs/017-proxy-stability/, no root pollution
- **Validation**: Project structure shows no new files in root, all changes in internal/

### Principle VI: Test Coverage Enforcement ✅ PASS
- **Status**: COMPLIANT
- **Design Impact**: New internal/httpx package targets 80% coverage, existing packages maintain thresholds
- **Validation**: quickstart.md includes coverage verification checklist before PR

### Summary
**All gates PASS after Phase 1 design** - No constitution violations. Design maintains TDD, simplicity, and coverage principles. Ready for Phase 2 (Task Generation via `/speckit.tasks`).

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/
├── root.go              # Version constant (update before release)
├── daemon.go            # Daemon start/stop commands (add auto-restart wrapper)
└── [other commands]

internal/
├── httpx/               # NEW: Panic recovery middleware
│   ├── recovery.go
│   └── recovery_test.go
├── daemon/
│   ├── api.go           # MODIFY: Add /health endpoint with detailed metrics
│   ├── server.go        # MODIFY: Add graceful shutdown, background worker management
│   ├── server_test.go   # MODIFY: Add health endpoint tests
│   ├── metrics.go       # NEW: Metrics collection (requests, latency, errors, resources)
│   ├── metrics_test.go  # NEW: Metrics tests
│   ├── logger.go        # NEW: Structured JSON logging
│   └── logger_test.go   # NEW: Logger tests
├── proxy/
│   ├── server.go        # MODIFY: Add streaming write error handling, metrics hooks
│   ├── provider.go      # MODIFY: Unified HTTP transport, connection pool management
│   ├── provider_test.go # MODIFY: Connection pool cleanup tests
│   ├── profile_proxy.go # MODIFY: Cache invalidation with connection cleanup
│   ├── healthcheck.go   # MODIFY: Client cleanup after health checks
│   └── limiter.go       # NEW: Concurrency limiter (semaphore-based, 100 limit)
├── web/
│   └── server.go        # MODIFY: Add panic recovery middleware
└── config/
    └── [no changes]     # No schema modifications

tests/
├── integration/         # NEW: 24-hour stability test, 100 concurrent load test
└── e2e/                 # NEW: Auto-restart verification, panic isolation test

scripts/
└── dev.sh               # Existing dev daemon management
```

**Structure Decision**: Single Go project with existing internal package structure. New `internal/httpx` package for shared middleware. Metrics and logging added to `internal/daemon`. Concurrency limiter added to `internal/proxy`. No config schema changes, all improvements are runtime behavior.

## Complexity Tracking

**No violations** - All constitution gates pass. This section is not applicable.
