---

description: "Task list for daemon proxy stability improvements"
---

# Tasks: Daemon Proxy Stability Improvements

**Input**: Design documents from `/specs/017-proxy-stability/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are included per Constitution Principle I (TDD is NON-NEGOTIABLE). All tests must be written FIRST and verified to FAIL before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Go project structure: `internal/`, `cmd/`, `tests/`
- Paths follow existing GoZen structure from plan.md

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Create internal/httpx package directory structure
- [x] T002 Create internal/daemon/metrics.go and logger.go stub files
- [x] T003 [P] Create internal/proxy/limiter.go stub file
- [x] T004 [P] Create tests/integration directory for stability tests

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T005 Verify existing panic recovery middleware in internal/httpx/recovery.go is complete
- [x] T006 Verify existing health API in internal/daemon/api.go is complete
- [x] T007 Verify existing connection pool management in internal/proxy/provider.go is complete
- [x] T008 Verify existing graceful shutdown in internal/daemon/server.go is complete

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Continuous Reliable Service (Priority: P1) 🎯 MVP

**Goal**: Daemon runs for 24 hours without crashes, recovers from panics, maintains stable memory/goroutines, auto-restarts on unrecoverable errors

**Independent Test**: Run daemon for 24 hours under normal load (10-50 req/hr). Verify no crashes, memory growth <10%, goroutine count stable, auto-restart works.

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T009 [P] [US1] Write panic recovery test in internal/httpx/recovery_test.go (verify panic doesn't crash daemon)
- [x] T010 [P] [US1] Write memory stability test in tests/integration/stability_test.go (24-hour memory growth <10%)
- [x] T011 [P] [US1] Write goroutine stability test in tests/integration/stability_test.go (no goroutine leaks)
- [x] T012 [P] [US1] Write auto-restart test in tests/integration/daemon_restart_test.go (verify restart within 5s)

### Implementation for User Story 1

- [x] T013 [US1] Implement auto-restart wrapper in cmd/daemon.go (exponential backoff, max 5 restarts)
- [x] T014 [US1] Add goroutine leak detection monitor in internal/daemon/server.go (baseline comparison, 1-minute ticker)
- [x] T015 [US1] Implement goroutine stack dump on leak detection in internal/daemon/server.go (runtime.Stack)
- [x] T016 [US1] Add context cancellation for background workers in internal/daemon/server.go (runCtx, runCancel, bgWG)
- [x] T017 [US1] Verify panic recovery integration in internal/web/server.go and internal/daemon/server.go (httpx.Recover middleware)
- [x] T018 [US1] Add session cleanup loop context cancellation in internal/daemon/server.go (use runCtx)

**Checkpoint**: At this point, User Story 1 should be fully functional - daemon runs 24 hours without crashes

---

## Phase 4: User Story 2 - Transparent Health Visibility (Priority: P2)

**Goal**: Users can query health endpoint to see daemon status (healthy/degraded/unhealthy) with runtime metrics and provider health

**Independent Test**: Query /api/v1/daemon/health under various conditions (normal, high load, provider failures). Verify status accurately reflects daemon state.

### Tests for User Story 2

- [ ] T019 [P] [US2] Write health endpoint test in internal/daemon/server_test.go (verify 200 response with correct schema)
- [ ] T020 [P] [US2] Write degraded status test in internal/daemon/server_test.go (goroutines >1000 or memory >500MB)
- [ ] T021 [P] [US2] Write unhealthy status test in internal/daemon/server_test.go (all providers failing)
- [ ] T022 [P] [US2] Write health endpoint performance test in tests/integration/health_test.go (response <100ms under load)

### Implementation for User Story 2

- [x] T023 [US2] Verify /api/v1/daemon/health endpoint exists in internal/daemon/api.go (already implemented in buffer)
- [x] T024 [US2] Verify daemonHealthResponse struct in internal/daemon/api.go (status, memory, goroutines, providers)
- [x] T025 [US2] Verify health status logic in internal/daemon/api.go (healthy/degraded/unhealthy determination)
- [x] T026 [US2] Verify provider health integration in internal/daemon/api.go (GetAllStatus from health checker)
- [x] T027 [US2] Add health endpoint to proxy server in internal/daemon/server.go (already registered in buffer)

**Checkpoint**: At this point, User Story 2 should be fully functional - health endpoint provides accurate status

---

## Phase 5: User Story 3 - Observable Request Performance (Priority: P2)

**Goal**: Users can query metrics endpoint to see request statistics, latency percentiles, error breakdowns, resource peaks

**Independent Test**: Generate 1000 requests over 10 minutes. Query /api/v1/daemon/metrics and verify accurate counts, percentiles, error grouping.

### Tests for User Story 3

- [x] T028 [P] [US3] Write metrics collection test in internal/daemon/metrics_test.go (record requests, calculate percentiles)
- [x] T029 [P] [US3] Write metrics endpoint test in internal/daemon/server_test.go (verify 200 response with correct schema)
- [x] T030 [P] [US3] Write percentile calculation test in internal/daemon/metrics_test.go (P50/P95/P99 accuracy)
- [x] T031 [P] [US3] Write error grouping test in internal/daemon/metrics_test.go (by provider and type)

### Implementation for User Story 3

- [x] T032 [P] [US3] Create Metrics struct in internal/daemon/metrics.go (counters, ring buffer, error maps)
- [x] T033 [P] [US3] Create MetricsStats struct in internal/daemon/metrics.go (response schema)
- [x] T034 [US3] Implement NewMetrics constructor in internal/daemon/metrics.go (initialize maps and ring buffer)
- [x] T035 [US3] Implement RecordRequest method in internal/daemon/metrics.go (update counters, latency buffer, errors)
- [x] T036 [US3] Implement GetPercentile method in internal/daemon/metrics.go (sort ring buffer, calculate P50/P95/P99)
- [x] T037 [US3] Implement GetStats method in internal/daemon/metrics.go (aggregate all metrics)
- [x] T038 [US3] Add metrics instance to Daemon struct in internal/daemon/server.go
- [x] T039 [US3] Implement GET /api/v1/daemon/metrics endpoint in internal/daemon/api.go
- [x] T040 [US3] Integrate metrics recording in internal/proxy/server.go (record latency, success/failure, provider, error type)
- [x] T041 [US3] Add resource peak tracking in internal/daemon/server.go (goroutines, memory)

**Checkpoint**: At this point, User Story 3 should be fully functional - metrics endpoint provides accurate statistics

---

## Phase 6: User Story 4 - Resilient Under Load (Priority: P1)

**Goal**: Daemon handles 100 concurrent requests for 5 minutes without crashes, enforces timeouts, manages connection pools, prevents resource exhaustion

**Independent Test**: Send 100 concurrent requests continuously for 5 minutes. Verify all complete, daemon responsive, resources within bounds.

### Tests for User Story 4

- [x] T042 [P] [US4] Write concurrency limiter test in internal/proxy/limiter_test.go (verify 100 limit, blocking behavior)
- [x] T043 [P] [US4] Write load test in tests/integration/load_test.go (100 concurrent for 5 minutes)
- [x] T044 [P] [US4] Write timeout test in internal/proxy/server_test.go (verify request cancellation after timeout)
- [x] T045 [P] [US4] Write connection pool test in internal/proxy/provider_test.go (verify cleanup on invalidation)

### Implementation for User Story 4

- [x] T046 [P] [US4] Create Limiter struct in internal/proxy/limiter.go (semaphore channel)
- [x] T047 [P] [US4] Implement NewLimiter constructor in internal/proxy/limiter.go (create buffered channel with size 100)
- [x] T048 [P] [US4] Implement Acquire method in internal/proxy/limiter.go (block until slot available)
- [x] T049 [P] [US4] Implement Release method in internal/proxy/limiter.go (release slot)
- [x] T050 [US4] Add limiter to ProxyServer struct in internal/proxy/server.go
- [x] T051 [US4] Integrate limiter in ProxyServer.ServeHTTP in internal/proxy/server.go (Acquire/defer Release)
- [x] T052 [US4] Verify request timeout enforcement in internal/proxy/server.go (context.WithTimeout 120s)
- [x] T053 [US4] Verify connection pool cleanup in internal/proxy/profile_proxy.go (already implemented in buffer)
- [x] T054 [US4] Verify streaming write error handling in internal/proxy/server.go (already implemented in buffer)

**Checkpoint**: At this point, User Story 4 should be fully functional - daemon handles 100 concurrent requests gracefully

---

## Phase 7: User Story 5 - Structured Diagnostic Logging (Priority: P3)

**Goal**: All critical events logged in JSON format with timestamp, level, event type, context fields. Selective logging (errors + slow requests >1s).

**Independent Test**: Trigger various scenarios (startup, requests, errors, panics, resource warnings). Verify logs contain structured JSON with complete context.

### Tests for User Story 5

- [ ] T055 [P] [US5] Write structured logger test in internal/daemon/logger_test.go (verify JSON format)
- [ ] T056 [P] [US5] Write log event test in internal/daemon/logger_test.go (verify timestamp, level, event, fields)
- [ ] T057 [P] [US5] Write selective logging test in internal/daemon/logger_test.go (only errors and slow requests)

### Implementation for User Story 5

- [ ] T058 [P] [US5] Create StructuredLogger struct in internal/daemon/logger.go (wraps stdlib logger)
- [ ] T059 [P] [US5] Implement NewStructuredLogger constructor in internal/daemon/logger.go
- [ ] T060 [P] [US5] Implement Info method in internal/daemon/logger.go (JSON format with timestamp, level, event, fields)
- [ ] T061 [P] [US5] Implement Warn method in internal/daemon/logger.go (JSON format)
- [ ] T062 [P] [US5] Implement Error method in internal/daemon/logger.go (JSON format)
- [ ] T063 [P] [US5] Implement Debug method in internal/daemon/logger.go (JSON format)
- [ ] T064 [US5] Add structured logger to Daemon struct in internal/daemon/server.go
- [ ] T065 [US5] Log daemon_started event in internal/daemon/server.go (PID, ports, version)
- [ ] T066 [US5] Log daemon_shutdown event in internal/daemon/server.go (uptime, reason)
- [ ] T067 [US5] Log request_received event in internal/proxy/server.go (only if error or duration >1s)
- [ ] T068 [US5] Log provider_failed event in internal/proxy/server.go (session, provider, error, duration)
- [ ] T069 [US5] Log panic_recovered event in internal/httpx/recovery.go (error, stack, path)
- [ ] T070 [US5] Log goroutine_leak_detected event in internal/daemon/server.go (baseline, current)
- [ ] T071 [US5] Log daemon_crashed_restarting event in cmd/daemon.go (restart_count, backoff, error)

**Checkpoint**: At this point, User Story 5 should be fully functional - all critical events logged in structured JSON

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T072 [P] Update CLAUDE.md Active Technologies section with new packages (httpx, metrics, logger, limiter)
- [ ] T073 [P] Update CLAUDE.md Recent Changes section with stability improvements summary
- [ ] T074 Run go test ./... to verify all tests pass
- [ ] T075 Run go test -cover ./internal/{daemon,proxy,web,httpx} to verify coverage thresholds
- [ ] T076 Run quickstart.md validation checklist (health <100ms, metrics <100ms, 24-hour stability)
- [ ] T077 [P] Code cleanup: remove any debug logging or temporary test endpoints
- [ ] T078 [P] Verify all error messages are user-friendly and actionable
- [ ] T079 Run ./scripts/dev.sh restart to verify dev daemon starts cleanly
- [ ] T080 Manual testing: Send 100 concurrent requests and verify metrics accuracy

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Independent of US1
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - Independent of US1/US2
- **User Story 4 (P1)**: Can start after Foundational (Phase 2) - Independent of US1/US2/US3
- **User Story 5 (P3)**: Can start after Foundational (Phase 2) - Independent of US1/US2/US3/US4

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Core structs before methods
- Methods before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational verification tasks can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Implementation tasks within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Write panic recovery test in internal/httpx/recovery_test.go"
Task: "Write memory stability test in tests/integration/stability_test.go"
Task: "Write goroutine stability test in tests/integration/stability_test.go"
Task: "Write auto-restart test in tests/integration/daemon_restart_test.go"

# After tests fail, no parallel implementation tasks in US1 (sequential dependencies)
```

## Parallel Example: User Story 3

```bash
# Launch all tests for User Story 3 together:
Task: "Write metrics collection test in internal/daemon/metrics_test.go"
Task: "Write metrics endpoint test in internal/daemon/server_test.go"
Task: "Write percentile calculation test in internal/daemon/metrics_test.go"
Task: "Write error grouping test in internal/daemon/metrics_test.go"

# Launch struct creation tasks together:
Task: "Create Metrics struct in internal/daemon/metrics.go"
Task: "Create MetricsStats struct in internal/daemon/metrics.go"
```

## Parallel Example: User Story 5

```bash
# Launch all tests for User Story 5 together:
Task: "Write structured logger test in internal/daemon/logger_test.go"
Task: "Write log event test in internal/daemon/logger_test.go"
Task: "Write selective logging test in internal/daemon/logger_test.go"

# Launch struct and method creation tasks together:
Task: "Create StructuredLogger struct in internal/daemon/logger.go"
Task: "Implement Info method in internal/daemon/logger.go"
Task: "Implement Warn method in internal/daemon/logger.go"
Task: "Implement Error method in internal/daemon/logger.go"
Task: "Implement Debug method in internal/daemon/logger.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 4 Only - Both P1)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Continuous Reliable Service)
4. Complete Phase 6: User Story 4 (Resilient Under Load)
5. **STOP and VALIDATE**: Test 24-hour stability + 100 concurrent load
6. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 + 4 (P1) → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 (P2) → Test independently → Deploy/Demo
4. Add User Story 3 (P2) → Test independently → Deploy/Demo
5. Add User Story 5 (P3) → Test independently → Deploy/Demo
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (Continuous Reliable Service)
   - Developer B: User Story 4 (Resilient Under Load)
   - Developer C: User Story 2 (Transparent Health Visibility)
   - Developer D: User Story 3 (Observable Request Performance)
   - Developer E: User Story 5 (Structured Diagnostic Logging)
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing (TDD is NON-NEGOTIABLE per Constitution)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Many tasks verify existing buffer work - these should be quick validation checks
- Focus new implementation on: metrics collection, structured logging, concurrency limiter, auto-restart, goroutine leak detection
