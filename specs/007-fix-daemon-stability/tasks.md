# Tasks: Fix Daemon Proxy Stability

**Input**: Design documents from `/specs/007-fix-daemon-stability/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included (TDD is NON-NEGOTIABLE per constitution Principle I)

**Organization**: Tasks are grouped by user story. US1 (Stable Proxy Port) and US3 (Duration Fix) are independent. US2 (Daemon Recovery) depends on US1 (file lock and port pinning are prerequisites for reliable recovery).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: No new project structure needed — this feature modifies existing files. Setup ensures the working environment is ready.

- [X] T001 Verify dev environment builds cleanly: run `go build ./...` and `go test ./...` from repo root
- [X] T002 Verify current test coverage baselines by running `go test -cover` for `internal/daemon`, `internal/proxy`, `internal/config`, `internal/bot`, `internal/web`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared infrastructure used by multiple user stories

**CRITICAL**: US1 and US2 both depend on the file lock mechanism. Build it here first.

### Tests

- [X] T003 [P] Write tests for `DaemonLockPath()`, `AcquireDaemonLock()`, `ReleaseDaemonLock()` in `internal/daemon/daemon_test.go` — test: lock acquired successfully, second lock attempt blocks/returns EWOULDBLOCK, lock released on file close

### Implementation

- [X] T004 Implement `DaemonLockPath()` returning `filepath.Join(config.ConfigDirPath(), "zend.lock")` in `internal/daemon/daemon.go`
- [X] T005 Implement `AcquireDaemonLock()` and `ReleaseDaemonLock()` using `syscall.Flock(LOCK_EX|LOCK_NB)` in `internal/daemon/daemon.go` — return lock file handle on success, `ErrLockContention` on EWOULDBLOCK; caller can then do blocking `Flock(LOCK_EX)` on the handle
- [X] T006 Verify T003 tests pass: `go test ./internal/daemon/... -run TestDaemonLock`

**Checkpoint**: File lock primitives ready — US1 and US2 can now proceed

---

## Phase 3: User Story 1 — Stable Proxy Port Across Restarts (Priority: P0) MVP

**Goal**: The proxy port is pinned to a deterministic value (default 19841), persisted in config, and never changes across restarts. Port conflicts are detected with process identification. Users can change the port via `zen config set proxy_port`.

**Independent Test**: Start daemon, note port, stop & restart, verify port is identical. Also: occupy port with non-zen process, attempt start, verify error message names the conflicting process.

### Tests for User Story 1

> **Write these tests FIRST, ensure they FAIL before implementation**

- [X] T007 [P] [US1] Write test for `EnsureProxyPort()` in `internal/config/store_test.go` — test: when ProxyPort==0, persists DefaultProxyPort; when ProxyPort!=0, no-op; verify config file contains `proxy_port` after call
- [X] T008 [P] [US1] Write test for `GetProcessOnPort()` in `internal/daemon/process_unix_test.go` — test: returns PID and process name for a known listening port (use a test TCP listener), returns error for unused port
- [X] T009 [P] [US1] Write test for `IsZenProcess()` in `internal/daemon/process_unix_test.go` — test: returns true for process names containing "zen" or "gozen", false for others
- [X] T010 [P] [US1] Write test for smart port binding in `internal/daemon/server_test.go` — test: `startProxy()` binds successfully on free port; fails with descriptive error when port is occupied by non-zen process
- [X] T011 [P] [US1] Write test for `zen config set` subcommand in `cmd/config_test.go` — test: valid proxy_port saves to config; invalid port returns error; unknown key returns error; value type mismatch returns error
- [X] T012 [P] [US1] Write test for settings API returning `proxy_port` in `internal/web/api_settings_test.go` — test: GET /api/v1/settings response includes `proxy_port` field with correct value

### Implementation for User Story 1

- [X] T013 [US1] Add `EnsureProxyPort()` method to Store in `internal/config/store.go` — if ProxyPort==0, set to DefaultProxyPort and save; add convenience wrapper in `internal/config/compat.go`
- [X] T014 [US1] Call `config.EnsureProxyPort()` at the start of `Daemon.Start()` in `internal/daemon/server.go` (before `config.GetProxyPort()`) to ensure port is persisted on first run
- [X] T015 [P] [US1] Implement `GetProcessOnPort(port int) (pid int, name string, err error)` in `internal/daemon/process_unix.go` — shell out to `lsof -i :PORT -sTCP:LISTEN -t` then `ps -p PID -o comm=`
- [X] T016 [P] [US1] Implement `IsZenProcess(processName string) bool` in `internal/daemon/process_unix.go` — check if name contains "zen" or "gozen"
- [X] T017 [US1] Add smart port conflict handling to `startProxy()` in `internal/daemon/server.go` — on `net.Listen` failure: call `GetProcessOnPort()`, if zen process then kill it and retry bind (if kill fails with EPERM, report "Cannot kill process PID (permission denied). Try: sudo kill PID"), if non-zen then return error with process name and port
- [X] T018 [US1] Integrate file lock into `ensureDaemonRunning()` in `cmd/root.go` — acquire lock before starting daemon, release after `waitForDaemonReady()`; if lock contended, wait then check if daemon is alive
- [X] T019 [US1] Integrate file lock into `startDaemonBackground()` in `cmd/daemon.go` — same pattern as T018 for the `zen daemon start` code path
- [X] T020 [US1] Add `configSetCmd` Cobra subcommand in `cmd/config.go` — accepts `<key> <value>`, whitelist: `proxy_port`; validates port 1024-65535; saves via `config.SetProxyPort()`; if daemon running, stop & restart; print warning about restarting client processes
- [X] T021 [US1] Register `configSetCmd` under `configCmd` in `cmd/config.go` `init()` function
- [X] T022 [US1] Add `ProxyPort` field to `settingsResponse` struct and populate it in `getSettings()` handler in `internal/web/api_settings.go`
- [X] T023 [US1] Add read-only proxy port display to Web UI in `web/src/pages/settings/tabs/GeneralSettings.tsx` — show current `proxy_port` as disabled Input with helper text: "To change the proxy port, use `zen config set proxy_port <port>` in the terminal."
- [X] T024 [US1] Ensure `proxy_port` exists in `Settings` type in `web/src/types/api.ts`
- [X] T025 [US1] Verify all T007-T012 tests pass: `go test ./internal/config/... ./internal/daemon/... ./internal/web/... ./cmd/...`

**Checkpoint**: Proxy port is stable across restarts, configurable via CLI, read-only in Web UI. Port conflicts detected with process identification.

---

## Phase 4: User Story 2 — Automatic Daemon Recovery After System Sleep (Priority: P0)

**Goal**: When the daemon dies (e.g., SIGTERM during sleep) and the client exits with a connection error, the `zen` wrapper automatically restarts the daemon and re-launches the client. Coordinated via file lock.

**Independent Test**: Start a client session via `zen`, kill the daemon process, observe that `zen` automatically restarts the daemon and re-launches the client.

**Depends on**: Phase 2 (file lock), Phase 3 (stable port — recovery requires port to remain the same)

### Tests for User Story 2

> **Write these tests FIRST, ensure they FAIL before implementation**

- [X] T026 [P] [US2] Write test for `isConnectionError()` helper in `cmd/root_test.go` — test: returns true for stderr containing "connection refused", "connection reset", "ECONNREFUSED"; returns false for other error text; returns false for empty stderr
- [X] T027 [P] [US2] Write test for `runClientWithRetry()` logic in `cmd/root_test.go` — test: on connection error with dead daemon, restarts daemon and retries (mock daemon check and exec); on non-connection error, forwards exit code immediately; on second failure, forwards exit code without further retry; on exit code 0, returns success

### Implementation for User Story 2

- [X] T028 [US2] Extract client process execution from `startViaDaemon()` into a `runClient()` helper in `cmd/root.go` — takes cliPath, args, env; returns exit code, stderr output, and error; captures stderr via `io.MultiWriter(os.Stderr, &stderrBuf)` to tee output to both the terminal and a buffer for connection error detection
- [X] T029 [US2] Implement `isConnectionError(stderr string) bool` in `cmd/root.go` — check for patterns: "connection refused", "connection reset", "connection error", "ECONNREFUSED", "ECONNRESET", "ETIMEDOUT" (case-insensitive)
- [X] T030 [US2] Add retry loop to `startViaDaemon()` in `cmd/root.go` — after `runClient()` returns non-zero with connection error: check `daemon.IsDaemonRunning()`, if dead then log recovery message, call `ensureDaemonRunning()` (which uses file lock from T018), re-run `runClient()`; limit to 1 retry
- [X] T031 [US2] Add recovery logging — when daemon restart is triggered, log "Daemon connection lost. Restarting daemon..." to stderr so user sees recovery in progress
- [X] T031b [US2] Add daemon-side restart audit logging in `internal/daemon/server.go` — when the daemon starts and detects it replaced a stale process (via port conflict kill in T017), log "Daemon restarted (replaced stale process PID)" to `zend.log` for auditability (FR-010)
- [X] T032 [US2] Verify all T026-T027 tests pass: `go test ./cmd/... -run TestConnectionError -run TestRunClientWithRetry`

**Checkpoint**: After daemon death, `zen` wrapper auto-restarts daemon and re-launches client with single retry.

---

## Phase 5: User Story 3 — Accurate Request Duration in Monitoring (Priority: P1)

**Goal**: The `duration_ms` field in monitoring data outputs actual milliseconds instead of nanoseconds.

**Independent Test**: Make a request through proxy, query monitoring endpoint, verify `duration_ms` is a reasonable value (e.g., 2500 for 2.5s, not 2500000000).

**Depends on**: None (independent of US1 and US2)

### Tests for User Story 3

> **Write these tests FIRST, ensure they FAIL before implementation**

- [X] T033 [P] [US3] Write test for `RequestRecord` JSON serialization in `internal/proxy/request_monitor_test.go` — create a record with DurationMs=2500, marshal to JSON, verify `duration_ms` field is 2500 (not nanoseconds)
- [X] T034 [P] [US3] Write test for `ProviderAttempt` JSON serialization in `internal/proxy/request_monitor_test.go` — same pattern, verify `duration_ms` is milliseconds
- [X] T035 [P] [US3] Write test for `MatchLog` JSON serialization in `internal/bot/matcher_test.go` — create a MatchLog with DurationMs=1500, marshal to JSON, verify `duration_ms` field is 1500

### Implementation for User Story 3

- [X] T036 [P] [US3] Change `RequestRecord.Duration` field from `time.Duration` to `DurationMs int64` (keep JSON tag `duration_ms`) in `internal/proxy/request_monitor.go`
- [X] T037 [P] [US3] Change `ProviderAttempt.Duration` field from `time.Duration` to `DurationMs int64` (keep JSON tag `duration_ms`) in `internal/proxy/request_monitor.go`
- [X] T038 [P] [US3] Change `MatchLog.Duration` field from `time.Duration` to `DurationMs int64` (keep JSON tag `duration_ms`) in `internal/bot/matcher.go`
- [X] T039 [US3] Update all `RequestRecord` creation sites in `internal/proxy/server.go` — change `Duration: duration` to `DurationMs: duration.Milliseconds()` at lines ~789, ~810, ~869
- [X] T040 [US3] Update `buildFailoverChain()` in `internal/proxy/server.go` — change `Duration: f.Elapsed` to `DurationMs: f.Elapsed.Milliseconds()`
- [X] T041 [US3] Update `recordMatchLog()` in `internal/bot/matcher.go` — change `log.Duration = duration` to `log.DurationMs = duration.Milliseconds()`
- [X] T042 [US3] Fix any compilation errors from the field rename — search for all references to `.Duration` on RequestRecord, ProviderAttempt, and MatchLog types and update to `.DurationMs`
- [X] T043 [US3] Verify all T033-T035 tests pass: `go test ./internal/proxy/... ./internal/bot/... -run TestDuration -run TestMatchLog`

**Checkpoint**: All `duration_ms` values in monitoring data are actual milliseconds. Monitoring dashboard shows human-readable durations.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, coverage checks, and build verification

- [X] T044 Run full test suite: `go test ./...` — all tests must pass
- [X] T045 Verify test coverage meets thresholds: `go test -cover ./internal/daemon/` (≥50%), `go test -cover ./internal/proxy/` (≥80%), `go test -cover ./internal/config/` (≥80%), `go test -cover ./internal/bot/` (≥80%), `go test -cover ./internal/web/` (≥80%)
- [X] T046 Build web UI and verify: `cd web && pnpm install && pnpm test && pnpm build`
- [X] T047 Run `go build ./...` to verify clean compilation
- [X] T048 Run quickstart.md test scenarios via automated e2e integration tests (`go test -tags integration ./tests/`)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 — builds file lock primitives
- **Phase 3 (US1 - Port Stability)**: Depends on Phase 2 — uses file lock
- **Phase 4 (US2 - Daemon Recovery)**: Depends on Phase 2 + Phase 3 — uses file lock AND stable port
- **Phase 5 (US3 - Duration Fix)**: Depends on Phase 1 ONLY — independent of US1/US2, can run in parallel with Phase 3
- **Phase 6 (Polish)**: Depends on all desired user stories being complete

### User Story Dependencies

```
Phase 1 (Setup)
    ↓
Phase 2 (Foundational: File Lock)
    ↓                    ↘
Phase 3 (US1: Port)     Phase 5 (US3: Duration) ← can start after Phase 1
    ↓
Phase 4 (US2: Recovery)
    ↓                    ↙
Phase 6 (Polish)
```

### Parallel Opportunities

**Phase 2**: T003 (test) is independent
**Phase 3 (US1)**: T007-T012 tests can all run in parallel. T015+T016 implementations are in different functions, can run in parallel.
**Phase 4 (US2)**: T026+T027 tests can run in parallel.
**Phase 5 (US3)**: T033-T035 tests can run in parallel. T036-T038 struct changes can run in parallel.
**Cross-phase**: US3 (Phase 5) can run entirely in parallel with US1 (Phase 3) since they touch different files.

---

## Parallel Example: User Story 3 (Duration Fix)

```text
# All test tasks in parallel (different test files):
T033: "Write test for RequestRecord JSON serialization in internal/proxy/request_monitor_test.go"
T034: "Write test for ProviderAttempt JSON serialization in internal/proxy/request_monitor_test.go"
T035: "Write test for MatchLog JSON serialization in internal/bot/matcher_test.go"

# All struct changes in parallel (different structs, could be same file but no conflicts):
T036: "Change RequestRecord.Duration in internal/proxy/request_monitor.go"
T037: "Change ProviderAttempt.Duration in internal/proxy/request_monitor.go"
T038: "Change MatchLog.Duration in internal/bot/matcher.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 + User Story 3)

1. Complete Phase 1: Setup (verify build)
2. Complete Phase 2: Foundational (file lock)
3. Complete Phase 5: US3 Duration Fix (quick win, independent, high impact)
4. Complete Phase 3: US1 Stable Port (core stability fix)
5. **STOP and VALIDATE**: Test port stability and duration values independently
6. This alone resolves 2 of 3 root causes

### Full Delivery (Add User Story 2)

7. Complete Phase 4: US2 Daemon Recovery (depends on US1)
8. Complete Phase 6: Polish
9. **VALIDATE**: Full end-to-end testing with daemon kill scenarios

### Suggested Order for Single Developer

Phase 1 → Phase 2 → Phase 5 (US3, quick win) → Phase 3 (US1) → Phase 4 (US2) → Phase 6

---

## Notes

- [P] tasks = different files, no dependencies between them
- [Story] label maps task to specific user story for traceability
- TDD enforced: write test → verify it fails → implement → verify it passes
- Commit after each task or logical group (constitution Principle IV)
- US3 (Duration Fix) is the fastest win — can be completed first in ~30 minutes
- US1 (Port Stability) is the most complex — involves config, daemon, CLI, and Web UI changes
- US2 (Recovery) is moderate complexity but depends on US1's port stability and file lock
