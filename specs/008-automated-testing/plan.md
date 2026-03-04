# Implementation Plan: Comprehensive Automated Testing Infrastructure

**Branch**: `008-automated-testing` | **Date**: 2026-03-04 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/008-automated-testing/spec.md`

## Summary

Build a comprehensive automated testing infrastructure for GoZen that fills critical gaps between existing unit tests (80%+ coverage) and production behavior. The implementation adds: (1) integration tests verifying Web API config changes propagate to the running proxy, (2) e2e tests for provider failover and scenario routing through the real binary, (3) process stability tests for server deployment scenarios, (4) frontend page component tests, (5) Claude Code testing skills for developer workflow, and (6) a unified test runner with CI integration.

## Technical Context

**Language/Version**: Go 1.21+ (backend), TypeScript (frontend, React 18)
**Primary Dependencies**: `net/http`, `net/http/httptest`, `os/exec`, `encoding/json` (Go tests); vitest 4, @testing-library/react, MSW (frontend tests)
**Storage**: JSON config at `~/.zen/zen.json` (test isolation via `GOZEN_CONFIG_DIR` and ephemeral ports)
**Testing**: `go test` with `-tags=integration` build tags (Go); `pnpm test` / vitest (frontend)
**Target Platform**: macOS (primary), Linux (CI — Ubuntu)
**Project Type**: CLI tool with daemon, reverse proxy, web UI, and bot gateway
**Performance Goals**: All integration + e2e tests complete within 3 minutes; daemon survives 500+ mixed requests without >50MB memory growth
**Constraints**: Tests must use mock HTTP servers only (no real API credentials); e2e tests non-blocking in CI
**Scale/Scope**: ~30 new Go test functions, ~15 new frontend test files, 5 Claude Code skill files, Makefile updates, CI updates

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. TDD (NON-NEGOTIABLE) | PASS | This feature IS tests — all test code written first by nature. Skills are Markdown (no production code to TDD). |
| II. Simplicity & YAGNI | PASS | Tests reuse existing `testEnv` pattern from `tests/e2e_daemon_test.go` and `ProxyTestConfig` from `test/integration/proxy_test.go`. No new abstractions beyond consolidating duplicate helpers. Mock servers use `httptest.NewServer` directly. |
| III. Config Migration Safety | PASS | No config schema changes — tests only read/write existing config format. |
| IV. Branch Protection & Commit Discipline | PASS | All work on `008-automated-testing` branch. Individual commits per task group. |
| V. Minimal Artifacts | PASS | No summary docs. Skills are functional `.claude/commands/` files. Test runner is a Makefile (extends existing). |
| VI. Test Coverage (NON-NEGOTIABLE) | PASS | Frontend tests must maintain 70% threshold. Go coverage thresholds unchanged (new tests improve coverage, not reduce). |

**Pre-Phase 0 gate: PASS** — No violations.

## Project Structure

### Documentation (this feature)

```text
specs/008-automated-testing/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0: test infrastructure research
├── data-model.md        # Phase 1: mock server and test entity design
├── quickstart.md        # Phase 1: manual verification scenarios
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (repository root)

```text
# Go integration tests (existing directory, new files)
test/integration/
├── proxy_test.go           # EXISTING: 8 tests — ADD: config hot-reload, failover chain, scenario routing
├── web_test.go             # EXISTING: 12 tests — ADD: config persistence verification
├── daemon_test.go          # EXISTING: 12 tests — ADD: SIGTERM/SIGKILL recovery, idempotent start
├── helpers_test.go         # NEW: shared test helpers (consolidate duplicate findProjectRoot, findFreePort, setupTest)
└── mock_provider_test.go   # NEW: configurable mock Anthropic/OpenAI provider server

# Go e2e tests (existing directory, new files)
tests/
├── doc.go                  # EXISTING: build tag stub
├── e2e_daemon_test.go      # EXISTING: 6 tests from spec 007
├── e2e_proxy_test.go       # NEW: provider failover, scenario routing through real binary
└── e2e_stress_test.go      # NEW: 500+ request stress test, memory monitoring

# Frontend page tests (new test files alongside existing)
web/src/pages/
├── monitoring/
│   ├── index.tsx           # EXISTING
│   └── index.test.tsx      # NEW: render, auto-refresh, filters, detail modal
├── providers/
│   ├── index.tsx           # EXISTING
│   ├── index.test.tsx      # NEW: list rendering, add button
│   ├── edit.tsx            # EXISTING
│   └── edit.test.tsx       # NEW: form validation, edit flow
├── profiles/
│   ├── index.tsx           # EXISTING
│   ├── index.test.tsx      # NEW: list rendering
│   ├── edit.tsx            # EXISTING
│   └── edit.test.tsx       # NEW: provider reordering, routing config
└── settings/
    ├── index.tsx           # EXISTING
    └── tabs/
        ├── GeneralSettings.tsx      # EXISTING
        ├── GeneralSettings.test.tsx # NEW: read-only proxy port, settings display
        ├── PasswordSettings.tsx     # EXISTING
        └── PasswordSettings.test.tsx # NEW: password change flow

# Claude Code testing skills (new files)
.claude/commands/
├── test.run.md             # NEW: unit test runner with coverage
├── test.integration.md     # NEW: integration + e2e test runner
├── test.web.md             # NEW: frontend test runner
├── test.all.md             # NEW: all tiers consolidated
└── test.write.md           # NEW: test skeleton generator

# Build tooling (modify existing)
Makefile                    # MODIFY: add test-unit, test-integration, test-e2e, test-web, test-all targets
.github/workflows/ci.yml   # MODIFY: add non-blocking e2e test job
```

**Structure Decision**: Extend existing test directories (`test/integration/`, `tests/`, `web/src/`) rather than creating new top-level directories. This follows the established project layout and keeps related tests together. Shared helpers are consolidated into `helpers_test.go` to reduce duplication across the three existing test config types (`TestConfig`, `ProxyTestConfig`, `WebTestConfig`).

## Complexity Tracking

No constitution violations — table not needed.
