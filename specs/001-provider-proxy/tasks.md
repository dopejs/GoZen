# Tasks: Provider Proxy Settings

**Input**: Design documents from `/specs/001-provider-proxy/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included — Constitution Principle I (TDD) requires tests for new features.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Add the `proxy_url` field to config and proxy types, bump config version

- [x] T001 Add `ProxyURL` field to `ProviderConfig` struct and bump `CurrentConfigVersion` to 9 in internal/config/config.go
- [x] T002 Add `ProxyURL` and `Client` fields to `Provider` struct in internal/proxy/provider.go
- [x] T003 Add `ValidateProxyURL` function to internal/config/config.go — validate scheme (http/https/socks5), parseable URL, non-empty host; empty string is valid
- [x] T004 Add `golang.org/x/net/proxy` dependency via `go get golang.org/x/net/proxy`

**Checkpoint**: Config struct extended, validation function available, dependency added

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core proxy transport creation and credential masking used by all stories

**CRITICAL**: No user story work can begin until this phase is complete

### Tests for Foundational

- [x] T005 [P] Write table-driven tests for `ValidateProxyURL` in internal/config/config_test.go — cases: valid http/https/socks5, invalid scheme (ftp), malformed URL, empty string, IP address, embedded credentials, missing host
- [x] T006 [P] Write table-driven tests for config migration (v8 config without `proxy_url` loads correctly, round-trip marshal preserves field) in internal/config/config_test.go
- [x] T007 [P] Write table-driven tests for `MaskProxyURL` helper (mask credentials, no credentials, empty string) in internal/config/config_test.go

### Implementation for Foundational

- [x] T008 Add `MaskProxyURL` helper function to internal/config/config.go — use `url.URL.Redacted()` to mask credentials
- [x] T009 Update `Clone()` method on `ProviderConfig` to copy `ProxyURL` field in internal/config/config.go
- [x] T010 Add `NewHTTPClientWithProxy` function to internal/proxy/provider.go — create `*http.Client` with `http.Transport` configured for the proxy URL; handle http/https via `http.ProxyURL()` and socks5 via `golang.org/x/net/proxy.SOCKS5()` dialer

**Checkpoint**: Foundation ready — proxy URL validation, transport creation, and credential masking are testable

---

## Phase 3: User Story 1 - Configure a proxy for a provider (Priority: P1) MVP

**Goal**: Provider requests from GoZen's daemon proxy route through the configured proxy server

**Independent Test**: Configure a provider with `proxy_url`, start the daemon, make a request, verify it goes through the proxy

### Tests for User Story 1

- [x] T011 [P] [US1] Write test for `NewHTTPClientWithProxy` — verify http.Transport.Proxy is set for http/https scheme and SOCKS5 dialer is used for socks5 scheme, in internal/proxy/proxy_test.go
- [x] T012 [P] [US1] Write test that `forwardRequest` uses `p.Client` when set (not `s.Client`) in internal/proxy/proxy_test.go

### Implementation for User Story 1

- [x] T013 [US1] Update provider construction in cmd/root.go — pass `ProxyURL` from `ProviderConfig` to `proxy.Provider`, call `NewHTTPClientWithProxy` to set `Provider.Client` when `ProxyURL` is non-empty
- [x] T014 [US1] Update `forwardRequest` in internal/proxy/server.go — use `p.Client.Do(req)` when `p.Client != nil`, fall back to `s.Client.Do(req)` otherwise
- [x] T015 [US1] Update `logStructured` calls in internal/proxy/server.go — include masked proxy URL in log message when provider has a proxy configured (use `MaskProxyURL`)
- [x] T016 [US1] Update `HealthChecker.CheckProvider` in internal/proxy/healthcheck.go — accept and use provider's proxy URL to create a per-check HTTP client with the proxy transport

**Checkpoint**: US1 complete — daemon proxy routes requests through per-provider proxy

---

## Phase 4: User Story 2 - Direct connection with `zen use` (Priority: P2)

**Goal**: `zen use <provider>` exports proxy env vars so the CLI connects through the proxy

**Independent Test**: Configure a provider with `proxy_url`, run `zen use`, check spawned process env vars

### Tests for User Story 2

- [x] T017 [P] [US2] Write table-driven tests for `ExportProxyToEnv` — cases: http sets HTTP_PROXY+HTTPS_PROXY, socks5 sets ALL_PROXY, empty clears nothing, in internal/config/config_test.go

### Implementation for User Story 2

- [x] T018 [US2] Add `ExportProxyToEnv` method to `ProviderConfig` in internal/config/config.go — set HTTP_PROXY/HTTPS_PROXY for http/https, ALL_PROXY for socks5; no-op when ProxyURL is empty
- [x] T019 [US2] Call `ExportProxyToEnv` from `ExportToEnv` method in internal/config/config.go — add call at end of existing `ExportToEnv` so `zen use` path picks it up automatically

**Checkpoint**: US2 complete — `zen use` propagates proxy settings to CLI process

---

## Phase 5: User Story 3 - Manage proxy settings via TUI and Web UI (Priority: P3)

**Goal**: Users can configure proxy URL through the TUI editor and Web UI

**Independent Test**: Open TUI/Web UI, edit a provider, enter proxy URL, save, verify in zen.json

### Implementation for User Story 3

- [x] T020 [P] [US3] Add `fieldProxyURL` to TUI editor in tui/editor.go — add new field constant, text input with placeholder "socks5://proxy:1080", prompt "  Proxy URL:        ", placed after fieldSonnetModel
- [x] T021 [US3] Wire `fieldProxyURL` in TUI editor save/load in tui/editor.go — load ProxyURL from config when editing, validate with `ValidateProxyURL` on save, show error for invalid URLs
- [x] T022 [P] [US3] Add `proxy_url` to Provider TypeScript interface in web/src/types/api.ts
- [x] T023 [US3] Add proxy URL input field to Web UI provider editor in web/src/pages/providers/edit.tsx — optional field with placeholder and client-side scheme validation

**Checkpoint**: US3 complete — proxy URL configurable through both TUI and Web UI

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Sync exclusion, final validation, build check

- [x] T024 Exclude `proxy_url` from sync payload in internal/sync/manager.go — in `buildLocalPayload`, clone provider config and clear ProxyURL before marshaling; in pull/merge, preserve local ProxyURL values
- [x] T025 [P] Write test for sync exclusion — verify marshaled sync payload has no `proxy_url` field, and pull preserves local proxy values, in internal/sync/manager_test.go
- [x] T026 Run `go build ./...` and `go test ./...` to verify all tests pass and no build errors
- [ ] T027 Test manually: configure a provider with proxy, run `zen` and `zen use`, verify connectivity

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 completion — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2 — core proxy routing
- **US2 (Phase 4)**: Depends on Phase 2 only — independent of US1
- **US3 (Phase 5)**: Depends on Phase 2 only — independent of US1 and US2
- **Polish (Phase 6)**: Depends on all stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Depends on Foundational only — no dependencies on other stories
- **User Story 2 (P2)**: Depends on Foundational only — independent of US1
- **User Story 3 (P3)**: Depends on Foundational only — independent of US1/US2

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD per Constitution Principle I)
- Models/types before logic
- Core implementation before logging/polish

### Parallel Opportunities

- T005, T006, T007 can run in parallel (different test concerns, same file but different test functions)
- T011, T012 can run in parallel (different test functions)
- T020, T022 can run in parallel (TUI and Web UI are different codebases)
- US1, US2, US3 can all start in parallel after Foundational phase completes

---

## Parallel Example: User Story 1

```text
# After Foundational phase completes, launch US1 tests in parallel:
T011: "Write test for NewHTTPClientWithProxy in internal/proxy/proxy_test.go"
T012: "Write test for forwardRequest using p.Client in internal/proxy/proxy_test.go"

# Then implement sequentially:
T013 → T014 → T015 → T016
```

## Parallel Example: User Story 3

```text
# TUI and Web UI tasks in parallel:
T020: "Add fieldProxyURL to TUI editor in tui/editor.go"
T022: "Add proxy_url to Provider TypeScript interface in web/src/types/api.ts"

# Then wire up:
T021 (depends on T020)
T023 (depends on T022)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T004)
2. Complete Phase 2: Foundational (T005-T010)
3. Complete Phase 3: User Story 1 (T011-T016)
4. **STOP and VALIDATE**: Daemon proxy routes through configured proxy
5. Commit and verify with `go test ./...`

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 → Daemon proxy routing works → Commit (MVP)
3. Add US2 → `zen use` propagates proxy env vars → Commit
4. Add US3 → TUI and Web UI support → Commit
5. Polish → Sync exclusion, final tests → Commit

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- Constitution requires TDD — write tests first, verify they fail, then implement
- Commit after each completed story (per Constitution Principle IV)
- `go test ./...` must pass before each commit
