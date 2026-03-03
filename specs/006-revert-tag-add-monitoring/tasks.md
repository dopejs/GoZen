# Tasks: Revert Provider Tag & Add Request Monitoring UI

**Input**: Design documents from `/specs/006-revert-tag-add-monitoring/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Tests are REQUIRED per Constitution Principle I (TDD is NON-NEGOTIABLE). All tests must be written first and verified to fail before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- Go code: `internal/proxy/`, `internal/config/`, `internal/web/`
- Web UI: `internal/web/static/` (vanilla JS, no frameworks)
- Tests: `*_test.go` files alongside implementation

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Verify Go 1.21+ and dev environment setup per quickstart.md
- [x] T002 Create branch 006-revert-tag-add-monitoring and verify clean working directory
- [x] T003 Review existing code structure in internal/proxy/server.go to understand tag injection points

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Define RequestRecord and ProviderAttempt types in internal/proxy/request_monitor.go per data-model.md
- [x] T005 Define RequestMonitor struct with sync.RWMutex and buffer in internal/proxy/request_monitor.go
- [x] T006 Define RequestFilter struct for query filtering in internal/proxy/request_monitor.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Remove Provider Tag Injection (Priority: P1) 🎯 MVP

**Goal**: Completely remove provider tag injection code to stop API errors and data pollution

**Independent Test**: Send requests with thinking blocks through proxy, verify responses are unmodified and Bedrock API returns 200 OK

### Tests for User Story 1 (TDD - Write First)

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T007 [P] [US1] Write test in internal/proxy/server_test.go: TestCopyResponse_NoTagInjection (verify no tag in response)
- [x] T008 [P] [US1] Write test in internal/proxy/server_test.go: TestCopyResponse_ThinkingBlockPreserved (verify thinking block unchanged)
- [x] T009 [P] [US1] Write test in internal/config/config_test.go: TestConfig_DeprecatedFieldIgnored (verify show_provider_tag ignored on load)

### Implementation for User Story 1

- [x] T010 [US1] Remove injectProviderTag() function from internal/proxy/server.go (lines 593-658)
- [x] T011 [US1] Remove newTagInjectingReader() and tagInjectingReader type from internal/proxy/server.go (lines 660-840)
- [x] T012 [US1] Remove tag injection calls in copyResponse() from internal/proxy/server.go (lines 533-535 for streaming, 573-577 for non-streaming)
- [x] T013 [P] [US1] Remove ShowProviderTag field from OpenCCConfig in internal/config/config.go
- [x] T014 [P] [US1] Remove GetShowProviderTag() and SetShowProviderTag() methods from internal/config/store.go
- [x] T015 [P] [US1] Remove GetShowProviderTag() and SetShowProviderTag() functions from internal/config/compat.go
- [x] T016 [P] [US1] Remove ShowProviderTag from settings API in internal/web/api_settings.go
- [x] T017 [P] [US1] Remove provider tag toggle from Web UI settings page in internal/web/static/settings.html or internal/web/static/app.js
- [x] T018 [P] [US1] Remove show_provider_tag references from Web UI JavaScript in internal/web/static/app.js
- [x] T019 [US1] Remove tag injection tests from internal/proxy/server_test.go (search for "tag" or "provider tag")
- [x] T020 [US1] Update internal/web/api_settings_test.go to remove ShowProviderTag assertions
- [x] T021 [US1] Run all tests: go test ./internal/proxy ./internal/config ./internal/web
- [x] T022 [US1] Manual verification per quickstart.md Part 1: send thinking block request, verify no tag, test all scenarios (single provider, failover, all fail, streaming)

**Checkpoint**: At this point, User Story 1 should be fully functional - no tag injection, backward compatible config loading

---

## Phase 4: User Story 2 - Real-Time Request Monitoring Page (Priority: P1)

**Goal**: Implement Web UI monitoring page displaying real-time request metadata without modifying API responses

**Independent Test**: Start daemon, make API requests, open Web UI monitoring page, verify all requests appear with correct metadata

### Tests for User Story 2 (TDD - Write First)

- [x] T023 [P] [US2] Write test in internal/proxy/request_monitor_test.go: TestRequestMonitor_Add (verify buffer append and LRU eviction)
- [x] T024 [P] [US2] Write test in internal/proxy/request_monitor_test.go: TestRequestMonitor_GetRecent (verify reverse chronological order)
- [x] T025 [P] [US2] Write test in internal/proxy/request_monitor_test.go: TestRequestMonitor_ThreadSafety (concurrent Add/GetRecent)
- [x] T026 [P] [US2] Write test in internal/web/api_requests_test.go: TestGetRequests_Success (verify API returns JSON with correct structure)
- [x] T027 [P] [US2] Write test in internal/web/api_requests_test.go: TestGetRequests_WithFilters (verify provider/status/time filters work)

### Implementation for User Story 2

- [x] T028 [P] [US2] Implement NewRequestMonitor() constructor in internal/proxy/request_monitor.go
- [x] T029 [P] [US2] Implement Add() method with LRU eviction in internal/proxy/request_monitor.go
- [x] T030 [P] [US2] Implement GetRecent() method with filtering in internal/proxy/request_monitor.go
- [x] T031 [P] [US2] Implement global request monitor singleton (GetGlobalRequestMonitor, InitGlobalRequestMonitor) in internal/proxy/request_monitor.go
- [x] T032 [US2] Add request capture in ServeHTTP() in internal/proxy/server.go: extract session ID, client type, model, request size
- [x] T033 [US2] Add request capture in tryProviders() in internal/proxy/server.go: track provider attempts and failures
- [x] T034 [US2] Add request finalization in recordUsageAndMetrics() in internal/proxy/server.go: calculate cost, add to monitor
- [x] T035 [P] [US2] Implement GET /api/v1/monitoring/requests handler in internal/web/api_requests.go
- [x] T036 [P] [US2] Register /api/v1/requests route in internal/web/server.go
- [x] T037 [P] [US2] Create monitoring page HTML in internal/web/static/monitoring.html
- [x] T038 [P] [US2] Implement request table rendering in internal/web/static/monitoring.js
- [x] T039 [P] [US2] Implement polling logic (5s interval) in internal/web/static/monitoring.js
- [x] T040 [P] [US2] Add "Requests" navigation link in internal/web/static/index.html or internal/web/static/app.js
- [x] T041 [US2] Run all tests: go test ./internal/proxy ./internal/web
- [x] T042 [US2] Manual verification per quickstart.md Part 2 & 3: send requests, check API, check Web UI, test all scenarios (single provider, failover, all fail, streaming)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - tag removed, monitoring page functional

---

## Phase 5: User Story 3 - Request Detail View (Priority: P2)

**Goal**: Add detailed request view showing token breakdown, failover history, and error messages

**Independent Test**: Click on request in monitoring feed, verify detail modal opens with complete metadata

### Tests for User Story 3 (TDD - Write First)

- [x] T043 [P] [US3] Write test in internal/web/api_requests_test.go: TestGetRequestDetail_Success (verify single request detail API)
- [x] T044 [P] [US3] Write test in internal/web/api_requests_test.go: TestGetRequestDetail_NotFound (verify 404 for invalid ID)

### Implementation for User Story 3

- [x] T045 [P] [US3] Implement GET /api/v1/monitoring/requests/:id handler in internal/web/api_requests.go
- [x] T046 [P] [US3] Add GetByID() method to RequestMonitor in internal/proxy/request_monitor.go
- [x] T047 [P] [US3] Add requestsApi.get() method in web/src/lib/api.ts
- [x] T048 [P] [US3] Implement detail modal in web/src/pages/monitoring/index.tsx (show failover chain, token breakdown)
- [x] T049 [P] [US3] Add click handler to table rows and i18n translations
- [x] T050 [US3] Run tests: go test ./internal/proxy ./internal/web
- [ ] T051 [US3] Manual verification: click request, verify detail modal shows all fields per quickstart.md scenarios

**Checkpoint**: All P1 and P2 user stories should now be independently functional

---

## Phase 6: User Story 4 - Request Filtering and Search (Priority: P3)

**Goal**: Add filter controls to narrow request list by provider, model, status, time range

**Independent Test**: Apply filters (provider, status), verify table updates to show only matching requests

### Tests for User Story 4 (TDD - Write First)

- [x] T052 [P] [US4] Write test in internal/proxy/request_monitor_test.go: TestRequestMonitor_FilterByProvider (already existed)
- [x] T053 [P] [US4] Write test in internal/proxy/request_monitor_test.go: TestRequestMonitor_FilterByStatus
- [x] T054 [P] [US4] Write test in internal/proxy/request_monitor_test.go: TestRequestMonitor_FilterByTimeRange
- [x] T055 [P] [US4] Write test in internal/proxy/request_monitor_test.go: TestRequestMonitor_FilterByModel

### Implementation for User Story 4

- [x] T056 [P] [US4] Implement filter logic in GetRecent() method in internal/proxy/request_monitor.go (already implemented)
- [x] T057 [P] [US4] Add model filter UI control (dropdown) in web/src/pages/monitoring/index.tsx
- [x] T058 [P] [US4] Implement filter state management in web/src/pages/monitoring/index.tsx
- [x] T059 [P] [US4] Update API call to include model filter query param in web/src/pages/monitoring/index.tsx
- [x] T060 [US4] Run tests: go test ./internal/proxy
- [ ] T061 [US4] Manual verification: apply filters, verify table updates correctly per quickstart.md test scenarios

**Checkpoint**: All user stories (P1, P2, P3) should now be independently functional

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T062 [P] Add comprehensive test coverage to reach 80% threshold for internal/proxy (81.6%)
- [x] T063 [P] Add comprehensive test coverage to reach 80% threshold for internal/config (85.6%)
- [x] T064 [P] Add comprehensive test coverage to reach 80% threshold for internal/web (80.3%)
- [x] T065 Run go test -cover ./internal/proxy ./internal/config ./internal/web to verify coverage
- [x] T066 [P] Add error handling for edge cases (daemon restart, buffer overflow) in internal/proxy/request_monitor.go (LRU eviction already implemented)
- [x] T067 [P] Add logging for monitoring operations in internal/proxy/request_monitor.go (logging handled by proxy layer)
- [ ] T068 [P] Optimize Web UI rendering for large request lists in internal/web/static/monitoring.js
- [ ] T069 Run full quickstart.md validation (all scenarios)
- [x] T070 Update CLAUDE.md to document monitoring feature and tag removal
- [x] T071 Code cleanup: remove unused imports, format code with gofmt
- [ ] T072 Final integration test: restart dev daemon, send 50 requests, verify all appear in UI

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - User Story 1 (P1): Can start after Foundational - No dependencies on other stories
  - User Story 2 (P1): Can start after Foundational - No dependencies on other stories (but logically after US1)
  - User Story 3 (P2): Depends on User Story 2 (needs monitoring page to exist)
  - User Story 4 (P3): Depends on User Story 2 (needs monitoring page to exist)
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - Logically after US1 but technically independent
- **User Story 3 (P2)**: Depends on User Story 2 completion (extends monitoring page)
- **User Story 4 (P3)**: Depends on User Story 2 completion (extends monitoring page)

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD)
- Removal tasks (US1) before monitoring tasks (US2)
- Core monitoring (US2) before extensions (US3, US4)
- API before Web UI
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Within User Story 1: T013-T018 (config/web removal) can run in parallel after T010-T012 (proxy removal)
- Within User Story 2: T028-T031 (monitor impl) parallel with T035-T036 (API) parallel with T037-T040 (Web UI)
- Within User Story 3: T045-T046 (API) parallel with T047-T049 (Web UI)
- Within User Story 4: All implementation tasks T056-T059 can run in parallel
- All Polish tasks marked [P] can run in parallel

---

## Parallel Example: User Story 2

```bash
# Launch all tests for User Story 2 together:
Task: "Write test TestRequestMonitor_Add in internal/proxy/request_monitor_test.go"
Task: "Write test TestRequestMonitor_GetRecent in internal/proxy/request_monitor_test.go"
Task: "Write test TestRequestMonitor_ThreadSafety in internal/proxy/request_monitor_test.go"
Task: "Write test TestGetRequests_Success in internal/web/api_requests_test.go"
Task: "Write test TestGetRequests_WithFilters in internal/web/api_requests_test.go"

# Launch all core implementation tasks together:
Task: "Implement NewRequestMonitor() in internal/proxy/request_monitor.go"
Task: "Implement Add() method in internal/proxy/request_monitor.go"
Task: "Implement GetRecent() method in internal/proxy/request_monitor.go"
Task: "Implement global singleton in internal/proxy/request_monitor.go"

# Launch API and Web UI tasks together:
Task: "Implement GET /api/v1/monitoring/requests handler in internal/web/api_requests.go"
Task: "Register route in internal/web/server.go"
Task: "Create monitoring page HTML in internal/web/static/monitoring.html"
Task: "Implement table rendering in internal/web/static/monitoring.js"
Task: "Implement polling logic in internal/web/static/monitoring.js"
Task: "Add navigation link in internal/web/static/index.html or app.js"
```

---

## Implementation Strategy

### MVP First (User Stories 1 & 2 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Remove tag injection)
4. **STOP and VALIDATE**: Test US1 independently per quickstart.md Part 1
5. Complete Phase 4: User Story 2 (Monitoring page)
6. **STOP and VALIDATE**: Test US2 independently per quickstart.md Part 2 & 3
7. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (Critical fix!)
3. Add User Story 2 → Test independently → Deploy/Demo (MVP with monitoring!)
4. Add User Story 3 → Test independently → Deploy/Demo (Enhanced debugging)
5. Add User Story 4 → Test independently → Deploy/Demo (Power user features)
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (tag removal)
   - Developer B: User Story 2 (monitoring page) - starts after US1 tests pass
3. After US1 & US2 complete:
   - Developer A: User Story 3 (detail view)
   - Developer B: User Story 4 (filters)
4. Stories complete and integrate independently

---

## Summary

**Total Tasks**: 72
- Setup: 3 tasks
- Foundational: 3 tasks
- User Story 1 (P1): 16 tasks (3 tests + 13 implementation)
- User Story 2 (P1): 20 tasks (5 tests + 15 implementation)
- User Story 3 (P2): 9 tasks (2 tests + 7 implementation)
- User Story 4 (P3): 10 tasks (4 tests + 6 implementation)
- Polish: 11 tasks

**Parallel Opportunities**: 45 tasks marked [P] can run in parallel within their phase

**Independent Test Criteria**:
- US1: Send thinking block request, verify no tag, Bedrock returns 200 OK
- US2: Send requests, open Web UI, verify all appear with correct metadata
- US3: Click request, verify detail modal shows complete metadata
- US4: Apply filters, verify table updates to show only matching requests

**Suggested MVP Scope**: User Stories 1 & 2 (tag removal + basic monitoring)

**Format Validation**: ✅ All tasks follow checklist format with checkbox, ID, [P]/[Story] labels, and file paths

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- TDD is mandatory: verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Run `./scripts/dev.sh restart` after Go code changes
- Use quickstart.md for manual verification at each checkpoint
