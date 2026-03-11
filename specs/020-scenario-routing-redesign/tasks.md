# Tasks: Scenario Routing Architecture Redesign

**Input**: Design documents from `/specs/020-scenario-routing-redesign/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Implementation Strategy**: Complete refactoring. Existing scenario detection code (`internal/proxy/scenario.go`) will be replaced with new architecture.

**Tests**: This project follows TDD (Constitution I: NON-NEGOTIABLE). All tests MUST be written FIRST and verified to FAIL before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

**Key Design Decisions** (finalized 2026-03-10):
1. **Scenario Key Naming**: Support camelCase, kebab-case, and snake_case; normalize internally to camelCase
2. **Scenario Type**: `type Scenario = string` (type alias) with constants for builtin scenarios
3. **Config Structure**: New `RoutePolicy` type replacing `ScenarioRoute`, v14 → v15 migration
4. **Protocol Detection**: Priority: URL path → X-Zen-Client header → body structure → default openai_chat
5. **Implementation**: Complete refactoring (replace scenario.go, not modify)

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

GoZen uses Go project structure:
- `internal/proxy/` - Proxy routing logic
- `internal/config/` - Configuration management
- `internal/middleware/` - Middleware interface
- `tests/integration/` - Integration tests

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Create routing-specific file structure in internal/proxy/
- [X] T002 [P] Add routing types to internal/proxy/routing_decision.go
- [X] T003 [P] Update RequestContext in internal/middleware/interface.go with routing fields

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Bump CurrentConfigVersion to 15 in internal/config/config.go
- [X] T005 [P] Change Scenario type to string alias and add RoutePolicy type in internal/config/config.go
- [X] T006 [P] Add scenario key normalization function (camelCase) in internal/proxy/routing_classifier.go
- [X] T007 Implement config validation function ValidateRoutingConfig in internal/config/store.go
- [X] T008 [P] Add structured logging functions for routing decisions in internal/daemon/logger.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Protocol-Agnostic Scenario Detection (Priority: P1) 🎯 MVP

**Goal**: Enable scenario routing to work consistently across Anthropic Messages, OpenAI Chat, and OpenAI Responses protocols

**Independent Test**: Send equivalent requests (same semantic content) via different API protocols and verify they route to the same provider/model

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T009 [P] [US1] Write test for Anthropic Messages normalization in internal/proxy/routing_normalize_test.go
- [X] T010 [P] [US1] Write test for OpenAI Chat normalization in internal/proxy/routing_normalize_test.go
- [X] T011 [P] [US1] Write test for OpenAI Responses normalization in internal/proxy/routing_normalize_test.go
- [X] T012 [P] [US1] Write test for malformed request handling in internal/proxy/routing_normalize_test.go
- [X] T013 [P] [US1] Write test for feature extraction in internal/proxy/routing_normalize_test.go
- [X] T014 [P] [US1] Write integration test for protocol-agnostic routing in tests/integration/routing_protocol_test.go

### Implementation for User Story 1

- [X] T015 [P] [US1] Create NormalizedRequest type in internal/proxy/routing_normalize.go
- [X] T016 [P] [US1] Create RequestFeatures type in internal/proxy/routing_normalize.go
- [X] T017 [US1] Implement DetectProtocol function (URL path → header → body → default) in internal/proxy/routing_normalize.go
- [X] T018 [US1] Implement Normalize function for Anthropic Messages in internal/proxy/routing_normalize.go
- [X] T019 [US1] Implement Normalize function for OpenAI Chat in internal/proxy/routing_normalize.go
- [X] T020 [US1] Implement Normalize function for OpenAI Responses in internal/proxy/routing_normalize.go
- [X] T021 [US1] Implement ExtractFeatures function in internal/proxy/routing_normalize.go
- [X] T022 [US1] Implement token counting for long-context detection in internal/proxy/routing_normalize.go
- [X] T023 [US1] Update ProxyServer.ServeHTTP to populate RequestContext.RequestFormat in internal/proxy/server.go
- [X] T024 [US1] Update ProxyServer.ServeHTTP to populate RequestContext.NormalizedRequest in internal/proxy/server.go
- [X] T025 [US1] Add error handling for normalization failures (route to default) in internal/proxy/server.go

**Checkpoint**: At this point, User Story 1 should be fully functional - requests normalize correctly across all three protocols

---

## Phase 4: User Story 2 - Middleware-Driven Custom Routing (Priority: P1)

**Goal**: Allow middleware to explicitly set routing decisions without manipulating request body shapes

**Independent Test**: Create a test middleware that sets a custom scenario (e.g., "plan") and verify the request routes to the configured provider for that scenario

### Tests for User Story 2

- [X] T026 [P] [US2] Write test for middleware decision precedence in internal/proxy/routing_resolver_test.go
- [X] T027 [P] [US2] Write test for builtin classifier fallback in internal/proxy/routing_classifier_test.go
- [X] T028 [P] [US2] Write test for routing hints integration in internal/proxy/routing_classifier_test.go
- [X] T029 [P] [US2] Write integration test for middleware-driven routing in tests/integration/routing_middleware_test.go

### Implementation for User Story 2

- [X] T030 [P] [US2] Implement BuiltinClassifier.Classify function in internal/proxy/routing_classifier.go
- [X] T031 [P] [US2] Implement confidence scoring in internal/proxy/routing_classifier.go
- [X] T032 [US2] Implement ResolveRoutingDecision function in internal/proxy/routing_resolver.go
- [X] T033 [US2] Implement routing hints integration in builtin classifier in internal/proxy/routing_classifier.go
- [X] T034 [US2] Update ProxyServer.ServeHTTP to call middleware pipeline before routing in internal/proxy/server.go
- [X] T035 [US2] Update ProxyServer.ServeHTTP to resolve routing decision after middleware in internal/proxy/server.go
- [X] T036 [US2] Add logging for routing decisions in internal/proxy/server.go

**Checkpoint**: At this point, User Stories 1 AND 2 should both work - middleware can override builtin classifier

---

## Phase 5: User Story 3 - Open Scenario Namespace (Priority: P2)

**Goal**: Allow users to define custom scenario routes in config without modifying source code

**Independent Test**: Add a custom scenario route to the config, have middleware emit that scenario, and verify the request routes correctly

### Tests for User Story 3

- [X] T037 [P] [US3] Write test for custom scenario route lookup in internal/proxy/routing_resolver_test.go
- [X] T038 [P] [US3] Write test for scenario key normalization in internal/proxy/routing_classifier_test.go
- [X] T039 [P] [US3] Write test for unknown scenario fallback in internal/proxy/routing_resolver_test.go
- [X] T040 [P] [US3] Write test for config validation with custom routes in internal/config/config_test.go

### Implementation for User Story 3

- [X] T041 [P] [US3] Implement NormalizeScenarioKey function in internal/proxy/routing_classifier.go
- [X] T042 [US3] Implement ResolveRoutePolicy function in internal/proxy/routing_resolver.go
- [X] T043 [US3] Update config validation to accept custom scenario keys in internal/config/store.go
- [X] T044 [US3] Update ProxyServer.ServeHTTP to use ResolveRoutePolicy in internal/proxy/server.go
- [X] T045 [US3] Add fallback to default route for unknown scenarios in internal/proxy/server.go

**Checkpoint**: All user stories 1-3 should now work - custom scenarios can be configured and routed

---

## Phase 6: User Story 4 - Per-Scenario Routing Policies (Priority: P2)

**Goal**: Allow each scenario route to have its own strategy, weights, and model overrides

**Independent Test**: Configure different strategies for different scenarios and verify each scenario uses its own policy

### Tests for User Story 4

- [X] T046 [P] [US4] Write test for per-scenario strategy application in internal/proxy/loadbalancer_test.go
- [X] T047 [P] [US4] Write test for per-scenario weights in internal/proxy/loadbalancer_test.go
- [X] T048 [P] [US4] Write test for per-scenario model overrides in internal/proxy/profile_proxy_test.go
- [ ] T049 [P] [US4] Write test for per-scenario threshold override in internal/proxy/routing_classifier_test.go
- [ ] T050 [P] [US4] Write integration test for per-scenario policies in tests/integration/routing_policy_test.go

### Implementation for User Story 4

- [ ] T051 [US4] Update LoadBalancer.Select to accept route-specific strategy in internal/proxy/loadbalancer.go
- [ ] T052 [US4] Update LoadBalancer.Select to accept route-specific weights in internal/proxy/loadbalancer.go
- [ ] T053 [US4] Update ProfileProxy to apply route-specific model overrides in internal/proxy/profile_proxy.go
- [ ] T054 [US4] Update scenario detection to use route-specific threshold in internal/proxy/routing_classifier.go
- [X] T055 [US4] Update ProxyServer.ServeHTTP to pass route policy to load balancer in internal/proxy/server.go

**Checkpoint**: All user stories 1-4 should work - each scenario can have independent routing policy

---

## Phase 7: User Story 5 - Strong Config Validation (Priority: P3)

**Goal**: Reject invalid routing configurations at load time with clear error messages

**Independent Test**: Attempt to load various invalid configs and verify each fails with a specific error message

### Tests for User Story 5

- [ ] T056 [P] [US5] Write test for non-existent provider validation in internal/config/config_test.go
- [ ] T057 [P] [US5] Write test for empty provider list validation in internal/config/config_test.go
- [ ] T058 [P] [US5] Write test for invalid weights validation in internal/config/config_test.go
- [ ] T059 [P] [US5] Write test for invalid strategy validation in internal/config/config_test.go
- [ ] T060 [P] [US5] Write test for scenario key format validation in internal/config/config_test.go

### Implementation for User Story 5

- [ ] T061 [US5] Implement provider existence validation in ValidateRoutingConfig in internal/config/store.go
- [ ] T062 [US5] Implement empty provider list validation in ValidateRoutingConfig in internal/config/store.go
- [ ] T063 [US5] Implement weights validation in ValidateRoutingConfig in internal/config/store.go
- [ ] T064 [US5] Implement strategy validation in ValidateRoutingConfig in internal/config/store.go
- [ ] T065 [US5] Implement scenario key format validation in ValidateRoutingConfig in internal/config/store.go
- [ ] T066 [US5] Call ValidateRoutingConfig in Store.loadLocked in internal/config/store.go
- [ ] T067 [US5] Add structured error messages for validation failures in internal/config/store.go

**Checkpoint**: All user stories 1-5 should work - invalid configs are rejected at load time

---

## Phase 8: User Story 6 - Routing Observability (Priority: P3)

**Goal**: Emit structured logs that explain why each request was routed to a specific provider and model

**Independent Test**: Process requests and verify the expected log entries are emitted with correct fields

### Tests for User Story 6

- [ ] T068 [P] [US6] Write test for middleware decision logging in internal/proxy/server_test.go
- [ ] T069 [P] [US6] Write test for builtin classifier logging in internal/proxy/server_test.go
- [ ] T070 [P] [US6] Write test for fallback logging in internal/proxy/server_test.go
- [ ] T071 [P] [US6] Write test for provider selection logging in internal/proxy/server_test.go

### Implementation for User Story 6

- [ ] T072 [US6] Implement LogRoutingDecision function in internal/daemon/logger.go
- [ ] T073 [US6] Implement LogRoutingFallback function in internal/daemon/logger.go
- [ ] T074 [US6] Add routing decision logging in ProxyServer.ServeHTTP in internal/proxy/server.go
- [ ] T075 [US6] Add fallback logging in ProxyServer.ServeHTTP in internal/proxy/server.go
- [ ] T076 [US6] Add provider selection logging in ProxyServer.ServeHTTP in internal/proxy/server.go
- [ ] T077 [US6] Add request features logging in ProxyServer.ServeHTTP in internal/proxy/server.go

**Checkpoint**: All user stories complete - routing decisions are fully observable

---

## Phase 9: Config Migration & Backward Compatibility

**Purpose**: Ensure v14 configs migrate automatically to v15 with RoutePolicy structure

### Tests for Config Migration

- [ ] T078 [P] Write test for v14→v15 config migration (ScenarioRoute → RoutePolicy) in internal/config/config_test.go
- [ ] T079 [P] Write test for scenario key normalization (kebab-case → camelCase) in internal/config/config_test.go
- [ ] T080 [P] Write test for builtin scenario preservation in internal/proxy/routing_classifier_test.go
- [ ] T081 [P] Write test for config round-trip (marshal/unmarshal) in internal/config/config_test.go

### Implementation for Config Migration

- [ ] T082 Implement RoutePolicy.UnmarshalJSON with v14 ScenarioRoute detection in internal/config/config.go
- [ ] T083 Implement legacy ScenarioRoute to RoutePolicy conversion (add default values for new fields) in internal/config/config.go
- [ ] T083.1 Verify profile-level strategy/weights/threshold fields preserved during v14→v15 migration in internal/config/config.go
- [ ] T084 Implement scenario key normalization (web-search → webSearch) in internal/proxy/routing_classifier.go
- [ ] T085 Update Store.saveLocked to write version 15 in internal/config/store.go
- [ ] T086 [P] Update TUI routing.go to support custom scenario keys
- [ ] T087 [P] Update Web UI types/api.ts to change Scenario type to string
- [ ] T088 [P] Update Web UI pages/profiles/edit.tsx to support custom scenarios

**Checkpoint**: Legacy configs migrate automatically, custom scenarios work in UI

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T089 [P] Update CLAUDE.md with new routing patterns
- [ ] T090 [P] Update docs/scenario-routing-architecture.md with implementation details
- [ ] T091 [P] Remove or deprecate old scenario.go file
- [ ] T092 Code cleanup and refactoring across routing files
- [ ] T093 Performance profiling for normalization and classification
- [ ] T094 [P] Add edge case tests for concurrent requests in tests/integration/
- [ ] T095 [P] Add edge case tests for session cache interaction in tests/integration/
- [ ] T096 [P] Add comprehensive E2E tests for all builtin scenarios in tests/e2e_proxy_test.go
- [ ] T097 Run quickstart.md validation scenarios
- [ ] T098 Verify test coverage ≥ 80% for internal/proxy and internal/config

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-8)**: All depend on Foundational phase completion
  - US1 (Protocol-Agnostic) → No dependencies on other stories
  - US2 (Middleware-Driven) → Depends on US1 (needs normalization)
  - US3 (Open Namespace) → Depends on US2 (needs decision resolution)
  - US4 (Per-Scenario Policies) → Depends on US3 (needs route resolution)
  - US5 (Config Validation) → Can start after Foundational (independent)
  - US6 (Observability) → Can start after US1 (needs routing flow)
- **Config Migration (Phase 9)**: Depends on US3 completion (needs new config types)
- **Polish (Phase 10)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational - No dependencies on other stories
- **User Story 2 (P1)**: Depends on US1 (needs NormalizedRequest)
- **User Story 3 (P2)**: Depends on US2 (needs RoutingDecision resolution)
- **User Story 4 (P2)**: Depends on US3 (needs RoutePolicy resolution)
- **User Story 5 (P3)**: Can start after Foundational - Independent of other stories
- **User Story 6 (P3)**: Depends on US1 (needs routing flow to log)

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD per Constitution I)
- Types before functions
- Core functions before integration
- Integration before logging
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- All tests for a user story marked [P] can run in parallel
- Types within a story marked [P] can run in parallel
- US5 (Config Validation) can run in parallel with US1-US4 (independent)
- US6 (Observability) can run in parallel with US2-US5 after US1 completes

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Write test for Anthropic Messages normalization in internal/proxy/routing_normalize_test.go"
Task: "Write test for OpenAI Chat normalization in internal/proxy/routing_normalize_test.go"
Task: "Write test for OpenAI Responses normalization in internal/proxy/routing_normalize_test.go"
Task: "Write test for malformed request handling in internal/proxy/routing_normalize_test.go"
Task: "Write test for feature extraction in internal/proxy/routing_normalize_test.go"

# Launch all types for User Story 1 together:
Task: "Create NormalizedRequest type in internal/proxy/routing_normalize.go"
Task: "Create RequestFeatures type in internal/proxy/routing_normalize.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Protocol-Agnostic Detection)
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

**MVP Deliverable**: Scenario routing works consistently across Anthropic, OpenAI Chat, and OpenAI Responses protocols

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo (Middleware extensibility)
4. Add User Story 3 → Test independently → Deploy/Demo (Custom scenarios)
5. Add User Story 4 → Test independently → Deploy/Demo (Per-scenario policies)
6. Add User Story 5 → Test independently → Deploy/Demo (Config validation)
7. Add User Story 6 → Test independently → Deploy/Demo (Observability)
8. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (Protocol-Agnostic)
   - Developer B: User Story 5 (Config Validation) - independent
3. After US1 completes:
   - Developer A: User Story 2 (Middleware-Driven)
   - Developer C: User Story 6 (Observability) - depends on US1
4. After US2 completes:
   - Developer A: User Story 3 (Open Namespace)
5. After US3 completes:
   - Developer A: User Story 4 (Per-Scenario Policies)
   - Developer B: Config Migration (Phase 9)
6. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- TDD is NON-NEGOTIABLE (Constitution I): Write tests FIRST, verify they FAIL, then implement
- Test coverage MUST be ≥ 80% for internal/proxy and internal/config (Constitution VI)
- Commit after each task or logical group (Constitution IV)
- Stop at any checkpoint to validate story independently
- Daemon proxy stability is P0 - all issues are blocking (Constitution VIII)
