# Tasks: Manual Provider Unavailability Marking

**Input**: Design documents from `/specs/015-mark-provider-unavailable/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/api.md

**Tests**: Included per Constitution Principle I (TDD is NON-NEGOTIABLE). Tests written first, verified to fail, then implementation to make them pass.

**Organization**: Tasks grouped by user story. Foundational phase covers config layer + proxy core (required by all stories).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Config Schema)

**Purpose**: Add UnavailableMarking type and extend OpenCCConfig with disabled_providers field

> **Note**: Phase 1 creates type definitions needed for test compilation in Phase 2. No logic is implemented until tests exist.

- [X] T001 Add UnavailableMarking struct with Type/CreatedAt/ExpiresAt fields, IsExpired()/IsActive() methods, NewUnavailableMarking() constructor, and expiration type constants ("today"/"month"/"permanent") in internal/config/config.go
- [X] T002 Add DisabledProviders map[string]*UnavailableMarking field to OpenCCConfig with json tag "disabled_providers,omitempty", bump CurrentConfigVersion from 13 to 14, and add DisabledProviders to UnmarshalJSON raw struct in internal/config/config.go

---

## Phase 2: Foundational (Config Store + Compat)

**Purpose**: Store methods for disable/enable operations. MUST complete before any user story.

**CRITICAL**: No user story work can begin until this phase is complete.

- [X] T003 Write table-driven tests for UnavailableMarking: IsExpired() returns false for unexpired today/month, true for expired, false for permanent; IsActive() returns true when not expired; NewUnavailableMarking() calculates correct ExpiresAt for each type in internal/config/config_test.go
- [X] T004 Write table-driven tests for Store methods: DisableProvider() persists marking, EnableProvider() removes marking, GetDisabledProviders() returns only active markings, IsProviderDisabled() checks active status, DisableProvider on nonexistent provider returns error. Per Constitution Principle III, also test: (a) v13 config without disabled_providers field parses to empty map, (b) v14 config with disabled_providers round-trips correctly via marshal/unmarshal, (c) all existing fields preserved after v13→v14 migration save in internal/config/config_test.go
- [X] T005 Implement DisableProvider(name, markingType string), EnableProvider(name string), GetDisabledProviders() map, IsProviderDisabled(name string) bool methods on Store in internal/config/store.go
- [X] T006 Update DeleteProvider() in Store to also delete the provider's entry from DisabledProviders map, and update ensureConfig() to initialize DisabledProviders map if nil in internal/config/store.go
- [X] T007 [P] Add convenience wrappers DisableProvider(), EnableProvider(), GetDisabledProviders(), IsProviderDisabled() in internal/config/compat.go following existing pattern (wrap DefaultStore() calls)

**Checkpoint**: Config layer complete — disable/enable operations work, persist, and survive reloads.

---

## Phase 3: User Story 3 - Error on All-Unavailable and Scenario Fallback (Priority: P1)

**Goal**: Proxy skips disabled providers (checked via `config.IsProviderDisabled()`), returns 503 error when all are disabled, scenario routes fall back to profile defaults when all route providers are disabled.

**Independent Test**: Mark all providers disabled via config, send proxy request, verify 503 error response with descriptive message.

### Tests for US3

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T008 [P] [US3] Write test: tryProviders skips a provider when config.IsProviderDisabled(name) returns true and other providers are available in internal/proxy/server_test.go
- [X] T009 [P] [US3] Write test: ServeHTTP returns 503 with all_providers_unavailable error JSON when all providers in default chain are disabled per config.IsProviderDisabled() in internal/proxy/server_test.go
- [X] T010 [P] [US3] Write test: ServeHTTP falls back from scenario route to default providers when all scenario providers are disabled per config, and returns 503 if defaults are also all disabled in internal/proxy/server_test.go

### Implementation for US3

- [X] T011 [US3] Add isProviderDisabled(name) helper method on ProxyServer that calls config.IsProviderDisabled() with lazy evaluation (no cached bool, no sync needed) in internal/proxy/server.go
- [X] T012 [US3] Add filterDisabledProviders() helper that checks each provider via isProviderDisabled() and returns (available []*Provider, allDisabled bool) in internal/proxy/server.go
- [X] T013 [US3] Modify tryProviders() to skip providers where isProviderDisabled(p.Name) returns true (skip with log message including "manually disabled", do not count as failure); log entry must be visible in request monitoring in internal/proxy/server.go
- [X] T014 [US3] Add all-providers-unavailable pre-check in ServeHTTP(): before calling tryProviders, call filterDisabledProviders(); if allDisabled, return 503 JSON error per contracts/api.md in internal/proxy/server.go
- [X] T015 [US3] Modify scenario fallback path in ServeHTTP(): when scenario route tryProviders returns false AND all scenario providers were disabled, filter defaults via filterDisabledProviders() before fallback and return 503 if all defaults also disabled in internal/proxy/server.go

**Checkpoint**: Proxy correctly skips disabled providers, returns 503 when all disabled, scenario fallback works.

---

## Phase 4: User Story 1 - Mark Provider Unavailable via Web UI (Priority: P1)

**Goal**: Users can disable/enable providers through the Web UI with duration selection.

**Independent Test**: Open Web UI, click disable on a provider, select "today", verify provider shows disabled badge, send proxy request and verify provider is skipped.

### Tests for US1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T016 [P] [US1] Write tests for POST /api/v1/providers/{name}/disable (valid types, invalid type, nonexistent provider) and POST /api/v1/providers/{name}/enable (success, nonexistent provider) in internal/web/api_providers_test.go
- [X] T017 [P] [US1] Write tests for GET /api/v1/providers/disabled (returns active markings, excludes expired) and extended GET /api/v1/providers (includes disabled field) in internal/web/api_providers_test.go

### Implementation for US1

- [X] T018 [US1] Implement handleProviderDisable() and handleProviderEnable() handlers in internal/web/api_providers.go per contracts/api.md response format
- [X] T019 [US1] Implement handleDisabledProviders() handler for GET /api/v1/providers/disabled in internal/web/api_providers.go
- [X] T020 [US1] Extend providerResponse struct with Disabled field and update toProviderResponse()/listProviders() to include active disabled status in internal/web/api_providers.go
- [X] T021 [US1] Register routes: POST /api/v1/providers/{name}/disable, POST /api/v1/providers/{name}/enable, GET /api/v1/providers/disabled in internal/web/server.go (handle path routing in existing handleProvider func)
- [X] T022 [US1] Add disableProvider(name, type) and enableProvider(name) API functions to web/src/hooks/use-providers.ts
- [X] T023 [US1] Add disable/enable dropdown action (today/month/permanent options) and enable button to provider list page, trigger API call and refresh list on success in web/src/ (provider page component)

**Checkpoint**: Web UI can disable/enable providers with duration selection.

---

## Phase 5: User Story 2 - Mark Provider Unavailable via CLI (Priority: P1)

**Goal**: Users can disable/enable providers from the terminal.

**Independent Test**: Run `zen disable my-provider --today`, verify output message, run `zen disable --list` to see it, run `zen enable my-provider` to clear.

### Tests for US2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T024 [P] [US2] Write tests for zen disable command: marks provider with correct type (default today, --month, --permanent), error on nonexistent provider, --list shows disabled providers in cmd/cmd_test.go
- [X] T025 [P] [US2] Write tests for zen enable command: clears marking, error on nonexistent provider, already-enabled provider is no-op in cmd/cmd_test.go

### Implementation for US2

- [X] T026 [US2] Create zen disable command with --today (default), --month, --permanent flags and --list flag; validate provider exists, call config.DisableProvider(), print confirmation with expiration time in cmd/disable.go
- [X] T027 [US2] Create zen enable command; validate provider exists, call config.EnableProvider(), print confirmation in cmd/enable.go
- [X] T028 [US2] Register disableCmd and enableCmd in init() of cmd/root.go

**Checkpoint**: CLI can disable/enable providers and list disabled status.

---

## Phase 6: User Story 4 - Direct Provider Selection Bypasses Unavailability (Priority: P2)

**Goal**: `zen use <provider>` works regardless of unavailability marking.

**Independent Test**: Disable a provider, run `zen use <provider>`, verify it launches without error.

- [X] T029 [US4] Write test verifying that `zen use` with a disabled provider still calls ExportProviderToEnv successfully (no disabled check in use path) in cmd/cmd_test.go
- [X] T030 [US4] Verify cmd/use.go does NOT call IsProviderDisabled — no code change needed if confirmed; if a check exists, remove it in cmd/use.go

**Checkpoint**: `zen use` bypasses unavailability marking — backward compatible.

---

## Phase 7: User Story 5 - Visibility of Unavailability Status (Priority: P2)

**Goal**: Clear visual indicators for disabled providers in Web UI and CLI.

**Independent Test**: Disable providers with different durations, check Web UI provider list for badges, check health page for separate status.

- [X] T031 [US5] Add disabled status badge component (shows type + expiration) to provider card/row in Web UI provider list page in web/src/ (provider page component)
- [X] T032 [US5] Add manual unavailability indicator distinct from auto-health status in health monitoring page in web/src/ (monitoring page component)
- [X] T033 [US5] Update existing `zen list` command output to show a disabled indicator (e.g., "[disabled: today]") next to provider names that are currently disabled in cmd/list.go
- [X] T034 [US5] Show disabled providers count or indicator in Web UI sidebar/header if any providers are currently disabled in web/src/components/layout/ (optional enhancement)

**Checkpoint**: Disabled providers are visually distinct in all views.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, coverage, cleanup

- [X] T035 Run `go test -cover ./internal/config/...` and verify ≥80% coverage; add tests if below threshold
- [X] T036 Run `go test -cover ./internal/proxy/...` and verify ≥80% coverage; add tests if below threshold
- [X] T037 Run `go test -cover ./internal/web/...` and verify ≥80% coverage; add tests if below threshold
- [X] T038 [P] Run `pnpm run test:coverage` in web/ and verify branch coverage ≥70%
- [X] T039 Run full `go test ./...` and verify all tests pass
- [ ] T040 Build and restart dev daemon (`./scripts/dev.sh restart`), manually test disable/enable flow end-to-end via Web UI and CLI

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS all user stories
- **US3 (Phase 3)**: Depends on Phase 2 — proxy behavior is needed for US1/US2 integration testing
- **US1 (Phase 4)**: Depends on Phase 2 (config layer) and Phase 3 (proxy filtering)
- **US2 (Phase 5)**: Depends on Phase 2 (config layer) and Phase 3 (proxy filtering)
- **US4 (Phase 6)**: Depends on Phase 2 (config layer) — verify no interference
- **US5 (Phase 7)**: Depends on Phase 4 (Web API/frontend) for UI integration
- **Polish (Phase 8)**: Depends on all user story phases complete

### User Story Dependencies

- **US3 (P1)**: First story — defines proxy behavior all others depend on
- **US1 (P1)**: Can start after US3; independent from US2
- **US2 (P1)**: Can start after US3; independent from US1; **can run in parallel with US1**
- **US4 (P2)**: Can start after Phase 2; independent from US1/US2/US3
- **US5 (P2)**: Depends on US1 (needs Web UI components to add badges to)

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD)
- Config/model before service/handler
- Handler before route registration
- Backend before frontend
- Commit after each task

### Parallel Opportunities

- T005 and T007: different files (store.go vs compat.go), can run in parallel after T003/T004
- T008, T009, T010: all proxy tests in same file, write sequentially but can be planned in parallel
- T016, T017: web API tests, can be written in parallel
- T024, T025: CLI tests, can be written in parallel
- US1 (Phase 4) and US2 (Phase 5): completely independent, can run in parallel
- US4 (Phase 6): can run in parallel with US1/US2

---

## Parallel Example: User Stories 1 & 2 (after Phase 3 complete)

```
# Developer A: Web UI (US1)
T016 → T017 → T018 → T019 → T020 → T021 → T022 → T023

# Developer B: CLI (US2) — runs in parallel with Developer A
T024 → T025 → T026 → T027 → T028
```

---

## Implementation Strategy

### MVP First (US3 + US1 or US2) 🎯

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T007)
3. Complete Phase 3: US3 - Proxy behavior (T008-T015)
4. **STOP and VALIDATE**: Proxy correctly skips disabled providers and returns 503 error
5. Complete Phase 4: US1 - Web UI OR Phase 5: US2 - CLI (either delivers user-facing value)
6. **MVP READY**: Users can disable providers through one interface

### Incremental Delivery

1. Setup + Foundational → Config layer works
2. Add US3 → Proxy behavior verified → Core complete
3. Add US1 → Web UI control → Demo-ready
4. Add US2 → CLI control → Full feature parity
5. Add US4 → Verify backward compat
6. Add US5 → Polish visibility
7. Coverage check + E2E → Release-ready

### Task Count Summary

| Phase | Tasks | Parallel Opportunities |
|-------|-------|----------------------|
| Phase 1: Setup | 2 | 0 (sequential) |
| Phase 2: Foundational | 5 | 1 (T005∥T007) |
| Phase 3: US3 | 8 | 3 (T008∥T009∥T010) |
| Phase 4: US1 | 8 | 2 (T016∥T017) |
| Phase 5: US2 | 5 | 2 (T024∥T025) |
| Phase 6: US4 | 2 | 0 |
| Phase 7: US5 | 4 | 0 |
| Phase 8: Polish | 6 | 1 (T038∥others) |
| **Total** | **40** | |
