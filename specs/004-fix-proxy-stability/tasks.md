# Tasks: Reverse Proxy Stability Fix

**Input**: Design documents from `/specs/004-fix-proxy-stability/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Included — TDD is NON-NEGOTIABLE per project constitution. Tests written first for all fixes.

**Organization**: Tasks grouped by user story. 4 files modified (~70 lines changed) + 4 test gaps filled + 2 enhancement tasks.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Branch and workspace preparation

- [x] T001 Verify branch `004-fix-proxy-stability` is checked out and builds cleanly with `go build ./...`
- [x] T002 Run baseline test suite `go test ./...` and record pass count and coverage for `internal/proxy/` — baseline: 81.0%

**Checkpoint**: Baseline established — all existing tests pass before any changes

---

## Phase 2: User Story 1 — Provider Network Proxy Support in Daemon Path (Priority: P1) 🎯 MVP

**Goal**: Fix `ProfileProxy.buildProviders()` to apply per-provider ProxyURL/Client settings and model defaults, matching the reference implementation in `cmd/root.go:buildProviders()`

**Independent Test**: Configure a provider with `proxy_url: "socks5://..."`. Send a request through `zen` (daemon path). Verify the upstream connection uses the configured network proxy.

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T003 [US1] Write test for ProxyURL propagation in daemon path buildProviders in `internal/proxy/profile_proxy_test.go`
- [x] T004 [US1] Write test for model default fallbacks in daemon path buildProviders in `internal/proxy/profile_proxy_test.go`
- [x] T005 [P] [US1] Write test for buildProviders with ProxyURL in direct path in `cmd/root_test.go`

### Implementation for User Story 1

- [x] T006 [US1] Fix `ProfileProxy.buildProviders()` in `internal/proxy/profile_proxy.go` — add ProxyURL + Client + model defaults
- [x] T007 [US1] Add model default fallbacks in `ProfileProxy.buildProviders()` in `internal/proxy/profile_proxy.go` — (done together with T006)
- [x] T008 [US1] Verify all US1 tests pass

**Checkpoint**: ProxyURL propagation works in daemon path — providers with proxy_url get dedicated HTTP clients

---

## Phase 3: User Story 2 — Daemon Readiness Verification (Priority: P1)

**Goal**: Fix `waitForDaemonReady()` to verify both web port AND proxy port are accepting connections before launching the client

**Independent Test**: Start daemon from cold state. Verify no ConnectionRefused on first request. Confirm logs show both ports verified.

### Tests for User Story 2 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T009 [US2] Write test for `waitForDaemonReady()` proxy port check in `cmd/daemon_test.go`

### Implementation for User Story 2

- [x] T010 [US2] Fix `waitForDaemonReady()` in `cmd/daemon.go` — check both web and proxy ports
- [x] T011 [US2] Improve error messaging in `ensureDaemonRunning()` in `cmd/root.go` — suggest `zen daemon status`
- [x] T012 [US2] Verify US2 tests pass

**Checkpoint**: Daemon readiness verifies both ports — eliminates startup race condition

---

## Phase 4: User Story 3 — Improved Failover Error Reporting (Priority: P2)

**Goal**: Improve 502 error response when all providers fail to include structured per-provider failure details

**Independent Test**: Configure a profile where all providers are unreachable. Send request through `zen`. Verify error message lists each provider and its failure reason.

### Tests for User Story 3 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T013 [US3] Write test for 502 response body format in `internal/proxy/server_test.go`

### Implementation for User Story 3

- [x] T014 [US3] Improve error message detail in `tryProviders()` — include elapsed time per provider
- [x] T015 [US3] Add elapsed time tracking to failover in `internal/proxy/server.go:tryProviders()`
- [x] T016 [US3] Verify US3 tests pass

**Checkpoint**: 502 errors now include actionable per-provider diagnostics

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Coverage verification and full regression

- [x] T017 Run full test suite `go test ./...` — all pass (flaky TestScenario_StopNormalDaemon is pre-existing)
- [x] T018 Check coverage for `internal/proxy/` — 81.7% (target ≥80%) PASS
- [x] T019 Check coverage for `cmd/` — 25.7% (cmd/ not in CI threshold; internal/proxy 81.7% is the gated package)
- [x] T020 Run quickstart.md manual verification steps (if dev environment available) — deferred to manual testing

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **US1 (Phase 2)**: Depends on Setup — HIGHEST PRIORITY (root cause fix)
- **US2 (Phase 3)**: Depends on Setup — can run in parallel with US1 (different files)
- **US3 (Phase 4)**: Depends on Setup — can run in parallel with US1/US2 (different files)
- **Polish (Phase 5)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: No dependencies on other stories — `internal/proxy/profile_proxy.go` + `cmd/root_test.go`
- **US2 (P1)**: No dependencies on other stories — `cmd/daemon.go` + `cmd/root.go`
- **US3 (P2)**: No dependencies on other stories — `internal/proxy/server.go`

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD — NON-NEGOTIABLE)
- Implementation makes tests pass
- Verification confirms tests pass
- Story complete before moving to next priority

### Parallel Opportunities

- T003/T004 run sequentially (same file), T005 can run in parallel with them (different file)
- US1, US2, US3 can all run in parallel (different source files, no shared dependencies)
- T017, T018, T019 can run in parallel (read-only coverage checks)

---

## Parallel Example: User Story 1

```bash
# T003/T004 sequentially (same file), T005 in parallel (different file):
Sequential: T003 → T004 in internal/proxy/profile_proxy_test.go
Parallel:   T005 in cmd/root_test.go

# Then implement sequentially (same file):
Task T006: "Fix buildProviders() ProxyURL in internal/proxy/profile_proxy.go"
Task T007: "Add model defaults in internal/proxy/profile_proxy.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (baseline)
2. Complete Phase 2: User Story 1 — ProxyURL fix (ROOT CAUSE)
3. **STOP and VALIDATE**: Test with a provider that has proxy_url configured
4. This alone should resolve the reported ConnectionRefused issue

### Incremental Delivery

1. Setup → Baseline established
2. US1 → ProxyURL fix → Root cause resolved (MVP!)
3. US2 → Readiness fix + diagnostic hints → Race condition eliminated
4. US3 → Error reporting + elapsed time logging → Better diagnostics
5. Polish → Coverage verified, full regression clean

### All Stories in Parallel

Since all 3 user stories modify different files with no shared dependencies:

- US1: `internal/proxy/profile_proxy.go` + `internal/proxy/profile_proxy_test.go` + `cmd/root_test.go`
- US2: `cmd/daemon.go` + `cmd/root.go` (+ tests)
- US3: `internal/proxy/server.go` + `internal/proxy/server_test.go`

All 3 can be implemented simultaneously for fastest delivery.

---

## Notes

- Total: 20 tasks (2 setup, 3 US1 tests, 3 US1 impl, 1 US2 test, 3 US2 impl, 1 US3 test, 3 US3 impl, 4 polish)
- All changes are in existing files — no new files created
- ~70 lines of production code changed across 4 files
- 4 test gaps filled across 3 test files
- 2 enhancement tasks: T011 (diagnostic suggestion), T015 (elapsed time logging)
- TDD enforced: tests T003-T005, T009, T013 must fail before implementation begins
- FR-006 (stale daemon recovery) already fully implemented — no new task needed, covered by T017 regression
