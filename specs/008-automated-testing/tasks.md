# Tasks: Comprehensive Automated Testing Infrastructure

**Input**: Design documents from `/specs/008-automated-testing/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, quickstart.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Consolidate duplicated test helpers and create the reusable mock provider server that all user stories depend on.

- [x] T001 Create shared test helpers by extracting `findProjectRoot()`, `findFreePort()`, and `setupBaseTest()` into `test/integration/helpers_test.go` — consolidating duplicates from `proxy_test.go`, `web_test.go`, and `daemon_test.go` (R1)
- [x] T002 [P] Refactor `ProxyTestConfig` in `test/integration/proxy_test.go` to embed the new `BaseTestConfig` from `helpers_test.go`, removing duplicated helper functions
- [x] T003 [P] Refactor `WebTestConfig` in `test/integration/web_test.go` to embed the new `BaseTestConfig` from `helpers_test.go`, removing duplicated helper functions
- [x] T004 [P] Refactor `TestConfig` in `test/integration/daemon_test.go` to embed the new `BaseTestConfig` from `helpers_test.go`, removing duplicated helper functions

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Mock provider server and shared e2e helpers that MUST be complete before any user story tests can be written.

**⚠️ CRITICAL**: No Go integration/e2e user story work (US1–US3) can begin until this phase is complete. US4 (skills) and US5 (frontend tests) have no dependency on this phase.

- [x] T005 Create configurable `MockProvider` and `MockResponse` structs in `test/integration/mock_provider_test.go` — wraps `httptest.Server` with FIFO response queue, default response, request counting, latency injection, and Anthropic `/v1/messages` response format (R2, data-model.md entities 2–3)
- [x] T006 Create shared e2e test helpers in `tests/helpers_test.go` — extract `MockProvider` usage patterns into functions that `e2e_proxy_test.go` and `e2e_stress_test.go` can share with the `testEnv` struct from `e2e_daemon_test.go`

**Checkpoint**: Foundation ready — mock provider and shared helpers available for all user story phases.

---

## Phase 3: User Story 1 — Web UI Config Persistence and Hot-Reload (Priority: P0) 🎯 MVP

**Goal**: Verify that config changes via Web API are persisted to `zen.json` and hot-reloaded into the running proxy without restart.

**Independent Test**: `go test -tags=integration -run TestIntegration_Config -v ./test/integration/...`

### Implementation for User Story 1

- [x] T007 [US1] Add `TestIntegration_ConfigHotReload_AddProvider` in `test/integration/proxy_test.go` — start daemon with provider A (mock), add provider B via `POST /api/v1/providers`, verify proxy routes to B when A is unavailable (FR-004, SC-001)
- [x] T008 [US1] Add `TestIntegration_ConfigHotReload_ChangeFailoverOrder` in `test/integration/proxy_test.go` — start daemon with profile ["A","B"], update to ["B","A"] via `PUT /api/v1/profiles/default`, verify B receives traffic first (FR-005, SC-001)
- [x] T009 [US1] Add `TestIntegration_ConfigHotReload_RemoveProvider` in `test/integration/proxy_test.go` — remove provider via `DELETE /api/v1/providers/:name`, verify proxy no longer routes to it (FR-006)
- [x] T010 [US1] Add `TestIntegration_ConfigHotReload_EditProviderURL` in `test/integration/proxy_test.go` — edit provider A's base_url to point to a different mock server via `PUT /api/v1/providers/A`, verify proxy routes to new URL (R4, acceptance scenario 5)
- [x] T011 [US1] Add `TestIntegration_ConfigPersistence_SettingsUpdate` in `test/integration/web_test.go` — update settings via `PUT /api/v1/settings`, verify `zen.json` reflects changes and proxy picks them up (FR-007)

**Checkpoint**: Config hot-reload flow verified end-to-end. All 5 acceptance scenarios covered.

---

## Phase 4: User Story 2 — Proxy Stability Under Adverse Conditions (Priority: P0)

**Goal**: Verify provider failover, scenario routing, and stress resilience through the real binary.

**Independent Test**: `go test -tags=integration -run "TestE2E_Provider|TestE2E_Scenario|TestE2E_Stress" -v -timeout 180s ./tests/...`

### Implementation for User Story 2

- [x] T012 [P] [US2] Add `TestE2E_ProviderFailover_TwoProviders` in `tests/e2e_proxy_test.go` — daemon with providers [A(503), B(200)], send request, verify failover A→B returns success (FR-008, SC-002)
- [x] T013 [P] [US2] Add `TestE2E_ProviderFailover_ThreeProviders` in `tests/e2e_proxy_test.go` — daemon with [A(503), B(503), C(200)], verify full chain A→B→C (FR-008, SC-002)
- [x] T014 [US2] Add `TestE2E_ProviderFailover_AllDown` in `tests/e2e_proxy_test.go` — all providers return 500, verify proxy returns error (not hang/crash) and daemon stays running (acceptance scenario 3)
- [x] T015 [US2] Add `TestE2E_ProviderFailover_RateLimited` in `tests/e2e_proxy_test.go` — provider returns 429, verify proxy fails over to next provider without crashing (acceptance scenario 6)
- [x] T016 [US2] Add `TestE2E_ScenarioRouting_ThinkingMode` in `tests/e2e_proxy_test.go` — configure think-capable and standard providers with scenario routing, send request with extended thinking params, verify routing to think-capable provider (FR-009, SC-003)
- [x] T017 [US2] Add `TestE2E_ClientDisconnect` in `tests/e2e_proxy_test.go` — send request, cancel context mid-flight, verify daemon stays stable for subsequent requests (FR-011, acceptance scenario 7)
- [x] T018 [US2] Add `TestE2E_StressTest` in `tests/e2e_stress_test.go` — send 500+ requests with mixed success/failure patterns, monitor memory via `ps -o rss= -p <PID>`, verify no >50MB growth (FR-010, SC-004, R5)

**Checkpoint**: Proxy stability verified — failover, scenario routing, client disconnect, and stress resilience all pass.

---

## Phase 5: User Story 3 — Process Stability for Server Deployment (Priority: P1)

**Goal**: Verify daemon handles SIGTERM, SIGKILL, stale PID recovery, and idempotent startup for unattended server scenarios.

**Independent Test**: `go test -tags=integration -run "TestE2E_Process" -v ./tests/...`

### Implementation for User Story 3

- [x] T019 [P] [US3] Add `TestE2E_ProcessStability_GracefulShutdown` in `tests/e2e_proxy_test.go` — send SIGTERM, verify in-flight requests complete, PID file removed, port released, clean exit (FR-012, SC-005)
- [x] T020 [P] [US3] Add `TestE2E_ProcessStability_KillRecovery` in `tests/e2e_proxy_test.go` — send SIGKILL, verify restart succeeds: stale PID cleaned, lock acquired, same ports (FR-013, SC-005)
- [x] T021 [US3] Add `TestE2E_ProcessStability_IdempotentStart` in `tests/e2e_proxy_test.go` — start daemon, attempt second start, verify detection of existing instance and no duplicate (FR-014)
- [x] T022 [US3] Add `TestE2E_ProcessStability_ConfigReloadUnderLoad` in `tests/e2e_proxy_test.go` — start daemon, send requests while config watcher triggers reload, verify no dropped connections and web UI remains accessible (acceptance scenario 4)
- [x] T023 [US3] Add reference process supervisor configurations in `docs/server-deployment.md` — include launchd plist (macOS) and systemd unit (Linux) examples demonstrating how to run GoZen as a server-mode service (FR-015, documentation only)

**Checkpoint**: Daemon is supervisor-friendly — clean PID handling, idempotent startup, no port conflicts on restart.

---

## Phase 6: User Story 4 — Testing Skills for Development Workflow (Priority: P1)

**Goal**: Create Claude Code slash commands that guide developers through writing and running automated tests.

**Independent Test**: Invoke each skill from Claude Code CLI and verify correct output.

### Implementation for User Story 4

- [x] T024 [P] [US4] Create `/test.run` skill in `.claude/commands/test.run.md` — detect modified Go packages, run `go test` with race detection and coverage, compare against CI thresholds (80% core / 50% supporting), report pass/fail with test names (FR-021, FR-026)
- [x] T025 [P] [US4] Create `/test.integration` skill in `.claude/commands/test.integration.md` — build binary, run `go test -tags=integration ./test/integration/... ./tests/...`, report results with daemon startup/shutdown status (FR-022, FR-026)
- [x] T026 [P] [US4] Create `/test.web` skill in `.claude/commands/test.web.md` — run `pnpm test` in `web/` with coverage, report page-level and hook-level results, flag drops below 70% (FR-023, FR-026)
- [x] T027 [P] [US4] Create `/test.all` skill in `.claude/commands/test.all.md` — run all test tiers (unit, integration, e2e, web) in sequence, produce consolidated pass/fail summary, suggest `/commit` on full pass (FR-024, FR-026, FR-027)
- [x] T028 [P] [US4] Create `/test.write` skill in `.claude/commands/test.write.md` — analyze `git diff` against base branch, identify test files and patterns, generate skeleton test cases following project TDD conventions (FR-025, FR-026)

**Checkpoint**: All 5 testing skills installable and invocable from Claude Code CLI.

---

## Phase 7: User Story 5 — Frontend Page Component Testing (Priority: P2)

**Goal**: Add page-level component tests for critical Web UI pages using vitest + @testing-library/react + MSW.

**Independent Test**: `cd web && pnpm test -- --run`

### Implementation for User Story 5

- [x] T029 [P] [US5] Create monitoring page tests in `web/src/pages/monitoring/index.test.tsx` — render with MSW mock data, verify request table rendering (timestamps, durations, status), auto-refresh toggle, filter interactions, detail modal display (FR-016, SC-006)
- [x] T030 [P] [US5] Create providers list page tests in `web/src/pages/providers/index.test.tsx` — render with mock providers, verify list rendering, add button interaction (FR-017, SC-006)
- [x] T031 [P] [US5] Create provider edit page tests in `web/src/pages/providers/edit.test.tsx` — render edit form, verify form validation, submit flow with correct API payload (FR-017, SC-006)
- [x] T032 [P] [US5] Create profiles list page tests in `web/src/pages/profiles/index.test.tsx` — render with mock profiles, verify list rendering (FR-018, SC-006)
- [x] T033 [P] [US5] Create profile edit page tests in `web/src/pages/profiles/edit.test.tsx` — render edit form, verify provider reordering, routing config (FR-018, SC-006)
- [x] T034 [P] [US5] Create general settings tab tests in `web/src/pages/settings/tabs/GeneralSettings.test.tsx` — render general tab, verify proxy port displayed read-only with correct value from settings API (FR-019, SC-006)
- [x] T035 [P] [US5] Create password settings tab tests in `web/src/pages/settings/tabs/PasswordSettings.test.tsx` — render password tab, verify password change form validation and submit flow (FR-019, SC-006)
- [x] T036 [US5] Verify frontend coverage remains above 70% threshold after adding all page tests — run `pnpm test:coverage` in `web/` (FR-020, SC-006)

**Checkpoint**: All critical page components have test coverage. Overall web UI coverage stays above 70%.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Unified test runner, CI updates, and validation.

- [x] T037 Add Makefile test targets: `test-unit` (go test ./...), `test-integration` (go test -tags=integration ./test/integration/...), `test-e2e` (go test -tags=integration -timeout 180s ./tests/...), `test-web` (cd web && pnpm test), `test-all` (all tiers in sequence) in `Makefile` (FR-028, SC-007)
- [x] T038 Update CI pipeline in `.github/workflows/ci.yml` — add new `e2e` job that runs `go test -tags=integration -timeout 180s ./tests/...` with `continue-on-error: true` (non-blocking), keeping existing integration tests in `go` job as blocking (FR-030, R7)
- [x] T039 Run quickstart.md validation — execute all 7 manual verification scenarios from `specs/008-automated-testing/quickstart.md` and verify expected outcomes (SC-007, SC-008)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (T001–T004) — BLOCKS Go test stories (US1–US3)
- **US1 (Phase 3)**: Depends on Phase 2 (needs MockProvider from T005)
- **US2 (Phase 4)**: Depends on Phase 2 (needs MockProvider from T005, e2e helpers from T006)
- **US3 (Phase 5)**: Depends on Phase 2 (needs e2e helpers from T006)
- **US4 (Phase 6)**: No code dependencies — can start after Phase 1 (skills are Markdown files)
- **US5 (Phase 7)**: No Go dependencies — can start after Phase 1 (uses existing MSW setup)
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P0)** ← Foundational: Uses `MockProvider` in integration tests
- **US2 (P0)** ← Foundational: Uses `MockProvider` in e2e tests
- **US3 (P1)** ← Foundational: Uses `testEnv` helpers in e2e tests
- **US4 (P1)** ← None: Markdown skill files, no code dependencies
- **US5 (P2)** ← None: Frontend tests use existing MSW setup, independent of Go tests

### Within Each User Story

- Integration/e2e tests written using TDD approach (tests define expected behavior, then verified against running daemon)
- Each test function is self-contained: start daemon, run assertions, cleanup
- Commit after each task or logical group

### Parallel Opportunities

- T002, T003, T004 can run in parallel (refactoring independent config structs)
- T005, T006 can run in parallel (mock provider vs e2e helpers — different directories)
- T012, T013 can run in parallel (different test functions in same file, no shared state)
- T019, T020 can run in parallel (different test functions)
- T024–T028 can ALL run in parallel (independent `.claude/commands/` files)
- T029–T035 can ALL run in parallel (independent page test files)
- US4 and US5 can run in parallel with US1–US3 (no code dependencies)

---

## Parallel Example: User Story 2

```bash
# Launch parallel e2e failover tests (different test functions, no shared state):
Task: "TestE2E_ProviderFailover_TwoProviders in tests/e2e_proxy_test.go"
Task: "TestE2E_ProviderFailover_ThreeProviders in tests/e2e_proxy_test.go"

# Launch parallel process stability tests:
Task: "TestE2E_ProcessStability_GracefulShutdown in tests/e2e_proxy_test.go"
Task: "TestE2E_ProcessStability_KillRecovery in tests/e2e_proxy_test.go"
```

## Parallel Example: User Story 5

```bash
# Launch ALL page tests in parallel (completely independent files):
Task: "monitoring/index.test.tsx"
Task: "providers/index.test.tsx"
Task: "providers/edit.test.tsx"
Task: "profiles/index.test.tsx"
Task: "profiles/edit.test.tsx"
Task: "settings/tabs/GeneralSettings.test.tsx"
Task: "settings/tabs/PasswordSettings.test.tsx"
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2)

1. Complete Phase 1: Setup (helper consolidation)
2. Complete Phase 2: Foundational (MockProvider + e2e helpers)
3. Complete Phase 3: US1 — Config hot-reload tests
4. Complete Phase 4: US2 — Proxy stability tests
5. **STOP and VALIDATE**: Run `make test-integration && make test-e2e` — core reliability verified

### Incremental Delivery

1. Setup + Foundational → Shared infrastructure ready
2. Add US1 → Config persistence verified → First quality gate
3. Add US2 → Proxy stability verified → Core reliability proven
4. Add US3 → Process stability verified → Server deployment ready
5. Add US4 → Testing skills available → Developer workflow established
6. Add US5 → Frontend coverage added → Full test pyramid complete
7. Polish → CI + Makefile + validation → Ready for merge

### Parallel Track Strategy

Since US4 (skills) and US5 (frontend tests) have no Go code dependencies:

- **Track A** (Go): Phase 1 → Phase 2 → US1 → US2 → US3
- **Track B** (Skills): US4 (can start immediately after Phase 1)
- **Track C** (Frontend): US5 (can start immediately)
- **Merge**: Polish phase after all tracks complete

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently testable via the listed test command
- All Go tests use `//go:build integration` build tag — won't run with `go test ./...`
- All frontend tests use existing MSW handlers in `web/src/test/setup.ts`
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
