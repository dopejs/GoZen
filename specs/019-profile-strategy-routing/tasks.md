# Tasks: Profile Strategy-Aware Provider Routing

**Input**: Design documents from `/specs/019-profile-strategy-routing/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: This feature follows TDD (Test-Driven Development) as mandated by the project constitution. All test tasks are included and MUST be completed before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Go project**: `internal/`, `cmd/`, `tests/` at repository root
- All paths are absolute from repository root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Verify existing project structure matches plan.md (internal/proxy/, internal/config/)
- [X] T002 Verify Go 1.21+ installed and dependencies available (net/http, sync, time, encoding/json)
- [X] T003 [P] Verify SQLite LogDB exists at ~/.zen/logs.db with requests table

**Checkpoint**: Project structure validated, ready for foundational work

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Add GetProviderLatencyMetrics() method signature to LogDB interface in internal/proxy/logdb.go
- [X] T005 [P] Verify ProfileConfig.Strategy field exists in internal/config/config.go (added v1.4.0)
- [X] T006 [P] Verify LoadBalanceStrategy enum exists with all 4 values in internal/config/config.go
- [X] T007 [P] Verify LoadBalancer.Select() signature accepts strategy parameter in internal/proxy/loadbalancer.go
- [X] T008 Add logging infrastructure for strategy decisions in internal/proxy/loadbalancer.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Least-Latency Strategy Routing (Priority: P1) 🎯 MVP

**Goal**: Route requests to the provider with lowest average latency, ensuring fastest possible response

**Independent Test**: Configure profile with least-latency strategy, send requests to providers with known latencies (A: 100ms, B: 50ms, C: 200ms), verify provider B is selected first

### Tests for User Story 1 (TDD - Write First, Ensure FAIL)

- [X] T009 [P] [US1] Write unit test for LogDB.GetProviderLatencyMetrics() in internal/proxy/logdb_test.go (test query with 100-request window, minimum 10 samples)
- [X] T010 [P] [US1] Write unit test for LoadBalancer.selectLeastLatency() in internal/proxy/loadbalancer_test.go (test sorting by latency, healthy providers first)
- [X] T011 [P] [US1] Write unit test for insufficient samples handling in internal/proxy/loadbalancer_test.go (test providers with <10 samples excluded)
- [X] T012 [P] [US1] Write integration test for least-latency routing in internal/proxy/profile_proxy_test.go (test end-to-end profile → strategy → provider selection)

**Checkpoint**: All US1 tests written and FAILING - ready for implementation

### Implementation for User Story 1

- [X] T013 [US1] Implement LogDB.GetProviderLatencyMetrics() in internal/proxy/logdb.go (SQL query: AVG(latency_ms) over last 100 requests, HAVING COUNT(*) >= 10)
- [X] T014 [US1] Modify LoadBalancer.selectLeastLatency() to query metrics and sort providers in internal/proxy/loadbalancer.go
- [X] T015 [US1] Add insufficient sample handling to selectLeastLatency() in internal/proxy/loadbalancer.go (exclude providers with <10 samples, append to end)
- [X] T016 [US1] Add strategy decision logging in LoadBalancer.Select() for least-latency in internal/proxy/loadbalancer.go (log: profile, strategy, selected, reason, candidates)
- [X] T017 [US1] Modify ProfileProxy.ServeHTTP() to pass profile strategy to LoadBalancer in internal/proxy/profile_proxy.go
- [X] T018 [US1] Run all US1 tests and verify they PASS (go test -v ./internal/proxy -run ".*Latency.*")

**Checkpoint**: User Story 1 fully functional - least-latency routing works independently

---

## Phase 4: User Story 2 - Least-Cost Strategy Routing (Priority: P2)

**Goal**: Route requests to the provider with lowest cost per token, minimizing API expenses

**Independent Test**: Configure profile with least-cost strategy, send requests to providers with known costs (A: $0.01/1K, B: $0.005/1K, C: $0.02/1K), verify provider B is selected first

### Tests for User Story 2 (TDD - Write First, Ensure FAIL)

- [X] T019 [P] [US2] Write unit test for LoadBalancer.selectLeastCost() in internal/proxy/loadbalancer_test.go (test sorting by cost, healthy providers first)
- [X] T020 [P] [US2] Write unit test for cost tiebreaker in internal/proxy/loadbalancer_test.go (test identical costs fall back to configured order)
- [X] T021 [P] [US2] Write integration test for least-cost routing in internal/proxy/profile_proxy_test.go (test end-to-end profile → strategy → provider selection)

**Checkpoint**: All US2 tests written and FAILING - ready for implementation

### Implementation for User Story 2

- [X] T022 [US2] Verify LoadBalancer.selectLeastCost() exists and uses pricing data in internal/proxy/loadbalancer.go (already implemented, may need adjustments)
- [X] T023 [US2] Add strategy decision logging in LoadBalancer.Select() for least-cost in internal/proxy/loadbalancer.go (log: profile, strategy, selected, reason="lowest cost: $X/1M tokens", candidates)
- [X] T024 [US2] Verify ProfileProxy passes strategy to LoadBalancer for least-cost in internal/proxy/profile_proxy.go (should work from US1 implementation)
- [X] T025 [US2] Run all US2 tests and verify they PASS (go test -v ./internal/proxy -run ".*Cost.*")

**Checkpoint**: User Stories 1 AND 2 both work independently - latency and cost routing functional

---

## Phase 5: User Story 3 - Round-Robin Strategy Routing (Priority: P3)

**Goal**: Distribute requests evenly across all healthy providers, ensuring balanced load distribution

**Independent Test**: Configure profile with round-robin strategy, send 9 requests to 3 providers, verify each receives exactly 3 requests

### Tests for User Story 3 (TDD - Write First, Ensure FAIL)

- [X] T026 [P] [US3] Write unit test for LoadBalancer.selectRoundRobin() in internal/proxy/loadbalancer_test.go (test even distribution, atomic counter increment)
- [X] T027 [P] [US3] Write unit test for round-robin with unhealthy provider in internal/proxy/loadbalancer_test.go (test skips unhealthy, distributes among healthy)
- [X] T028 [P] [US3] Write concurrency test for round-robin counter in internal/proxy/loadbalancer_test.go (test 100 concurrent calls, no race conditions)
- [X] T029 [P] [US3] Write integration test for round-robin routing in internal/proxy/profile_proxy_test.go (test 9 requests → 3 per provider)

**Checkpoint**: All US3 tests written and FAILING - ready for implementation

### Implementation for User Story 3

- [X] T030 [US3] Verify LoadBalancer.selectRoundRobin() exists and uses atomic counter in internal/proxy/loadbalancer.go (already implemented, may need adjustments)
- [X] T031 [US3] Add strategy decision logging in LoadBalancer.Select() for round-robin in internal/proxy/loadbalancer.go (log: profile, strategy, selected, reason="round-robin: index N", candidates)
- [X] T032 [US3] Verify ProfileProxy passes strategy to LoadBalancer for round-robin in internal/proxy/profile_proxy.go (should work from US1 implementation)
- [X] T033 [US3] Run all US3 tests including race detector (go test -race -v ./internal/proxy -run ".*RoundRobin.*")

**Checkpoint**: User Stories 1, 2, AND 3 all work independently - latency, cost, and round-robin routing functional

---

## Phase 6: User Story 4 - Weighted Strategy Routing (Priority: P3)

**Goal**: Distribute requests according to configured weights, allowing users to prefer certain providers

**Independent Test**: Configure profile with weighted strategy (A:70, B:20, C:10), send 100 requests, verify distribution matches weights within 15% variance

### Tests for User Story 4 (TDD - Write First, Ensure FAIL)

- [X] T034 [P] [US4] Write unit test for LoadBalancer.selectWeighted() in internal/proxy/loadbalancer_test.go (test weighted distribution, healthy providers only)
- [X] T035 [P] [US4] Write unit test for weighted recalculation in internal/proxy/loadbalancer_test.go (test weights recalculated when provider becomes unhealthy)
- [X] T036 [P] [US4] Write unit test for weighted fallback in internal/proxy/loadbalancer_test.go (test no weights configured → equal weights)
- [X] T037 [P] [US4] Write integration test for weighted routing in internal/proxy/profile_proxy_test.go (test 100 requests → distribution within 15% of weights)

**Checkpoint**: All US4 tests written and FAILING - ready for implementation

### Implementation for User Story 4

- [X] T038 [US4] Implement LoadBalancer.selectWeighted() in internal/proxy/loadbalancer.go (weighted random selection, recalculate on health change)
- [X] T039 [US4] Add weighted strategy to LoadBalancer.Select() switch statement in internal/proxy/loadbalancer.go
- [X] T040 [US4] Add strategy decision logging in LoadBalancer.Select() for weighted in internal/proxy/loadbalancer.go (log: profile, strategy, selected, reason="weighted: X%", candidates)
- [X] T041 [US4] Add weighted strategy constant to LoadBalanceStrategy enum in internal/config/config.go (if not already present)
- [X] T042 [US4] Verify ProfileProxy passes strategy to LoadBalancer for weighted in internal/proxy/profile_proxy.go (should work from US1 implementation)
- [X] T043 [US4] Run all US4 tests and verify they PASS (go test -v ./internal/proxy -run ".*Weighted.*")

**Checkpoint**: All 4 user stories work independently - complete strategy routing implementation

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T044 [P] Add comprehensive error handling tests in internal/proxy/loadbalancer_test.go (test DB query failure, invalid strategy, metric unavailable)
- [X] T045 [P] Add edge case tests in internal/proxy/loadbalancer_test.go (test all providers identical metrics, all providers unhealthy, concurrent config reload)
- [X] T046 [P] Verify metric cache concurrency safety in internal/proxy/loadbalancer_test.go (test RWMutex usage, read-only snapshots)
- [X] T047 [P] Add performance benchmarks in internal/proxy/loadbalancer_test.go (benchmark strategy evaluation <5ms target)
- [X] T048 Verify backward compatibility in internal/proxy/profile_proxy_test.go (test empty strategy defaults to failover, existing configs work)
- [X] T049 Run full test suite with coverage (go test -cover ./internal/proxy, target ≥80%) — 82.5% achieved
- [X] T050 Run full test suite with race detector (go test -race ./internal/proxy)
- [X] T051 [P] Update CLAUDE.md Active Technologies section with feature details
- [X] T052 [P] Verify quickstart.md validation steps in specs/019-profile-strategy-routing/quickstart.md
- [X] T053 Run integration tests against dev daemon (./scripts/dev.sh && go test ./tests/integration)

**Checkpoint**: All tests passing, coverage ≥80%, no race conditions, ready for PR

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3 → P3)
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Independent of US1 (uses different strategy logic)
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Independent of US1/US2 (uses different strategy logic)
- **User Story 4 (P3)**: Can start after Foundational (Phase 2) - Independent of US1/US2/US3 (uses different strategy logic)

**Key Insight**: All 4 user stories are FULLY INDEPENDENT after Foundational phase. They can be implemented in parallel by different developers.

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD)
- Implementation tasks must be completed in order (dependencies noted in task descriptions)
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all 4 user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- All Polish tasks marked [P] can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together (TDD - write first):
Task T009: "Write unit test for LogDB.GetProviderLatencyMetrics() in internal/proxy/logdb_test.go"
Task T010: "Write unit test for LoadBalancer.selectLeastLatency() in internal/proxy/loadbalancer_test.go"
Task T011: "Write unit test for insufficient samples handling in internal/proxy/loadbalancer_test.go"
Task T012: "Write integration test for least-latency routing in internal/proxy/profile_proxy_test.go"

# Verify all tests FAIL (expected - no implementation yet)

# Then implement sequentially (dependencies exist):
Task T013: "Implement LogDB.GetProviderLatencyMetrics()"
Task T014: "Modify LoadBalancer.selectLeastLatency()"
Task T015: "Add insufficient sample handling"
Task T016: "Add strategy decision logging"
Task T017: "Modify ProfileProxy.ServeHTTP()"
Task T018: "Run all US1 tests and verify PASS"
```

---

## Parallel Example: All User Stories (After Foundational)

```bash
# Once Phase 2 (Foundational) is complete, launch all user stories in parallel:

# Developer A: User Story 1 (Least-Latency)
Tasks T009-T018

# Developer B: User Story 2 (Least-Cost)
Tasks T019-T025

# Developer C: User Story 3 (Round-Robin)
Tasks T026-T033

# Developer D: User Story 4 (Weighted)
Tasks T034-T043

# All stories complete independently, integrate seamlessly
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T008) - CRITICAL
3. Complete Phase 3: User Story 1 (T009-T018)
4. **STOP and VALIDATE**: Test least-latency routing independently
5. Deploy/demo if ready - users get fastest response routing

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP - latency optimization!)
3. Add User Story 2 → Test independently → Deploy/Demo (cost optimization!)
4. Add User Story 3 → Test independently → Deploy/Demo (load balancing!)
5. Add User Story 4 → Test independently → Deploy/Demo (advanced control!)
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (T001-T008)
2. Once Foundational is done:
   - Developer A: User Story 1 (T009-T018) - Least-Latency
   - Developer B: User Story 2 (T019-T025) - Least-Cost
   - Developer C: User Story 3 (T026-T033) - Round-Robin
   - Developer D: User Story 4 (T034-T043) - Weighted
3. Stories complete and integrate independently
4. Team completes Polish together (T044-T053)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- TDD is MANDATORY: Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Run with race detector: `go test -race ./internal/proxy`
- Target coverage: ≥80% (go test -cover ./internal/proxy)
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence

---

## Task Count Summary

- **Total Tasks**: 53
- **Setup (Phase 1)**: 3 tasks
- **Foundational (Phase 2)**: 5 tasks (BLOCKING)
- **User Story 1 (Phase 3)**: 10 tasks (4 tests + 6 implementation)
- **User Story 2 (Phase 4)**: 7 tasks (3 tests + 4 implementation)
- **User Story 3 (Phase 5)**: 8 tasks (4 tests + 4 implementation)
- **User Story 4 (Phase 6)**: 10 tasks (4 tests + 6 implementation)
- **Polish (Phase 7)**: 10 tasks (cross-cutting)

**Parallel Opportunities**: 28 tasks marked [P] can run in parallel within their phase

**MVP Scope**: Phases 1-3 only (18 tasks) delivers least-latency routing - immediate user value

**Independent Test Criteria**:
- US1: Provider with lowest latency selected first (50ms beats 100ms beats 200ms)
- US2: Provider with lowest cost selected first ($0.005 beats $0.01 beats $0.02)
- US3: 9 requests distributed evenly (3 per provider)
- US4: 100 requests match weights within 15% (70/20/10 distribution)
