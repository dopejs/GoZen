# Tasks: Provider & Model Tag in Proxy Responses

**Input**: Design documents from `/specs/005-provider-model-tag/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Included — TDD is NON-NEGOTIABLE per project constitution. Tests written first for all new logic.

**Organization**: Tasks grouped by user story. ~5 Go files modified + ~3 Web UI files modified (~150-200 lines production code + tests).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Branch verification and baseline establishment

- [ ] T001 Verify branch `005-provider-model-tag` is checked out and builds cleanly with `go build ./...`
- [ ] T002 Run baseline test suite `go test ./...` and record pass count and coverage for `internal/proxy/`, `internal/config/`, `internal/web/`

**Checkpoint**: Baseline established — all existing tests pass before any changes

---

## Phase 2: Foundational — Config Infrastructure (FR-001, FR-011)

**Purpose**: Add `ShowProviderTag` config field and API support — MUST complete before tag injection (US1/US2) or Web UI (US3) can work

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Tests for Foundational ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T003 [P] Write test for `ShowProviderTag` config field: v10→v11 migration, default `false`, marshal round-trip in `internal/config/config_test.go`
- [ ] T004 [P] Write test for `GetShowProviderTag()` / `SetShowProviderTag()` in `internal/config/store_test.go`
- [ ] T005 [P] Write test for settings API GET/PUT with `show_provider_tag` field in `internal/web/api_settings_test.go`

### Implementation for Foundational

- [ ] T006 Add `ShowProviderTag bool` field to `OpenCCConfig` in `internal/config/config.go` and bump `CurrentConfigVersion` from 10 to 11
- [ ] T007 Add `GetShowProviderTag()` / `SetShowProviderTag()` methods to Store in `internal/config/store.go` and convenience functions in `internal/config/compat.go`
- [ ] T008 Add `ShowProviderTag` to `settingsResponse` and `settingsRequest` (as `*bool`) in `internal/web/api_settings.go`, wire getter/setter in `getSettings()` / `updateSettings()`
- [ ] T009 Verify all foundational tests pass (T003, T004, T005)

**Checkpoint**: Config field exists, API can read/write it, tests pass. Tag injection and Web UI can now proceed.

---

## Phase 3: User Story 1 — Non-Streaming Response Tag Injection (Priority: P1) 🎯 MVP

**Goal**: Prepend `[provider: <name>, model: <model>]\n` to the first text content block in non-streaming responses (both Anthropic Messages and OpenAI Chat Completions formats)

**Independent Test**: Enable the tag in settings, send a non-streaming request through the proxy, verify the first text content block starts with `[provider: xxx, model: xxx]\n`.

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T010 [US1] Write test for non-streaming Anthropic tag injection: enabled → tag prepended to first text block; disabled → body unmodified in `internal/proxy/server_test.go`
- [ ] T011 [US1] Write test for non-streaming OpenAI tag injection: enabled → tag prepended to `choices[0].message.content`; disabled → body unmodified in `internal/proxy/server_test.go`
- [ ] T012 [US1] Write test for edge cases: tool-use-only response (no tag), empty content array (no tag), non-2xx response (no tag), failover scenario (provider A fails → provider B succeeds → tag shows provider B's name and model) in `internal/proxy/server_test.go`

### Implementation for User Story 1

- [ ] T013 [US1] Implement `injectProviderTag()` helper function for non-streaming responses in `internal/proxy/server.go` — parses JSON body, identifies format (Anthropic content array vs OpenAI choices), prepends tag to first text content, re-marshals
- [ ] T014 [US1] Integrate tag injection into `copyResponse()` non-streaming path in `internal/proxy/server.go` — call `injectProviderTag()` after transformation, before writing body; gate on `config.GetShowProviderTag()` and 2xx status
- [ ] T015 [US1] Verify all US1 tests pass (T010, T011, T012)

**Checkpoint**: Non-streaming tag injection works for both Anthropic and OpenAI formats, gated by config setting

---

## Phase 4: User Story 2 — SSE Streaming Response Tag Injection (Priority: P1)

**Goal**: Prepend `[provider: <name>, model: <model>]\n` to the first text delta in SSE-streamed responses (both Anthropic and OpenAI formats)

**Independent Test**: Enable the tag, send a streaming request, verify the first content delta event contains the tag text before the actual content begins.

### Tests for User Story 2 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T016 [US2] Write test for Anthropic SSE tag injection: `message_start` → model extracted, first `content_block_delta` with `text_delta` → tag prepended in `internal/proxy/server_test.go`
- [ ] T017 [US2] Write test for OpenAI Chat Completions SSE tag injection: first chunk → model extracted, first `delta.content` → tag prepended in `internal/proxy/server_test.go`
- [ ] T018 [US2] Write test for OpenAI Responses API SSE tag injection (transformed from Anthropic via StreamTransformer): `response.created` → model extracted, first `response.output_text.delta` → tag prepended to `delta` field in `internal/proxy/server_test.go`
- [ ] T019 [US2] Write test for SSE edge cases: tool-use-only stream (no tag), tag disabled (passthrough) in `internal/proxy/server_test.go`

### Implementation for User Story 2

- [ ] T020 [US2] Implement `tagInjectingReader` struct in `internal/proxy/server.go` — wraps an `io.Reader`, buffers SSE events, extracts model from early events (`message_start` / first OpenAI chunk / `response.created`), prepends tag to first text delta, then passes through remaining data. Must handle three SSE variants: native Anthropic (`content_block_delta`), native OpenAI Chat Completions (`delta.content`), and transformed OpenAI Responses API (`response.output_text.delta`)
- [ ] T021 [US2] Integrate `tagInjectingReader` into `copyResponse()` SSE streaming path in `internal/proxy/server.go` — wrap the reader AFTER optional StreamTransformer (ensuring FR-010 post-transformation ordering) when `config.GetShowProviderTag()` is enabled and status is 2xx
- [ ] T022 [US2] Verify all US2 tests pass (T016, T017, T018, T019)

**Checkpoint**: SSE streaming tag injection works for both Anthropic and OpenAI formats, composable with existing StreamTransformer

---

## Phase 5: User Story 3 — Web UI Toggle for Tag Feature (Priority: P2)

**Goal**: Add a toggle switch in Web UI General Settings to enable/disable the provider/model tag without editing the config file

**Independent Test**: Open Web UI settings, toggle the "Show provider info in responses" switch, save, verify the setting persists and takes effect on the next proxy request.

### Tests for User Story 3 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T023 [US3] Write test for GeneralSettings component: verify toggle renders with default OFF, toggling ON and saving sends `show_provider_tag: true` to API in `web/src/pages/settings/tabs/GeneralSettings.test.tsx`

### Implementation for User Story 3

- [ ] T024 [US3] Add `show_provider_tag?: boolean` to `Settings` interface in `web/src/types/api.ts`
- [ ] T025 [P] [US3] Add translation keys for "Show provider info in responses" label and description in `web/src/i18n/locales/en.json`, `zh-CN.json`, `zh-TW.json`, `es.json`, `ja.json`, `ko.json`
- [ ] T026 [US3] Add Switch toggle for `show_provider_tag` in `web/src/pages/settings/tabs/GeneralSettings.tsx` — read from settings query, include in save mutation
- [ ] T027 [US3] Verify US3 tests pass and toggle works in dev Web UI at `http://localhost:29840`

**Checkpoint**: Web UI toggle controls the tag feature, persists across page reloads, takes effect without daemon restart

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Coverage verification, latency validation, and full regression

- [ ] T028 Run full test suite `go test ./...` — all pass
- [ ] T029 [P] Check coverage for `internal/proxy/` — target ≥80%
- [ ] T030 [P] Check coverage for `internal/config/` — target ≥80%
- [ ] T031 [P] Check coverage for `internal/web/` — target ≥80%
- [ ] T032 Run Web UI tests `cd web && npm run test` — all pass
- [ ] T033 Verify tag injection latency: add a Go benchmark test for `injectProviderTag()` and `tagInjectingReader` in `internal/proxy/server_test.go` — confirm <5ms per SC-003
- [ ] T034 Run quickstart.md manual verification steps (if dev environment available)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories (config field + API must exist first)
- **US1 (Phase 3)**: Depends on Foundational — non-streaming tag injection in `server.go`
- **US2 (Phase 4)**: Depends on Foundational — SSE streaming tag injection in `server.go`; depends on US1 (shared `server.go`, tag format helper)
- **US3 (Phase 5)**: Depends on Foundational — Web UI toggle (different files from US1/US2, but needs API field)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Depends only on Foundational — `internal/proxy/server.go` + `internal/proxy/server_test.go`
- **US2 (P1)**: Depends on US1 — shares `server.go` (builds on tag format helper from US1), different code paths but same file
- **US3 (P2)**: Depends only on Foundational — `web/src/` files + `internal/web/api_settings.go` (already done in Foundational)

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD — NON-NEGOTIABLE)
- Implementation makes tests pass
- Verification confirms tests pass
- Story complete before moving to next priority

### Parallel Opportunities

- T003, T004, T005 can all run in parallel (different test files)
- T006, T007, T008 run sequentially (same package, interdependent)
- US1 and US3 can run in parallel after Foundational (different files)
- US2 depends on US1 (same file, builds on helper)
- T029, T030, T031 can run in parallel (read-only coverage checks)

---

## Parallel Example: Foundational Phase

```bash
# Tests in parallel (different files):
Task T003: "Config field migration test in internal/config/config_test.go"
Task T004: "Store getter/setter test in internal/config/store_test.go"
Task T005: "Settings API test in internal/web/api_settings_test.go"

# Implementation sequentially (interdependent):
Task T006 → T007 → T008 → T009 (verify)
```

## Parallel Example: After Foundational

```bash
# US1 and US3 can run in parallel (different files):
Worker A: T010 → T011 → T012 → T013 → T014 → T015 (server.go)
Worker B: T023 → T024 → T025 → T026 → T027 (web/src/)

# US2 follows US1 (same file):
After Worker A: T016 → T017 → T018 → T019 → T020 → T021 → T022
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (baseline)
2. Complete Phase 2: Foundational — config field + API support
3. Complete Phase 3: User Story 1 — non-streaming tag injection
4. **STOP and VALIDATE**: Test with a non-streaming request; verify tag appears in response
5. This delivers the core value for users who check provider/model info

### Incremental Delivery

1. Setup → Baseline established
2. Foundational → Config + API ready
3. US1 → Non-streaming tag injection → Core value delivered (MVP!)
4. US2 → SSE streaming tag injection → Full proxy coverage (most real-world usage)
5. US3 → Web UI toggle → User-friendly control
6. Polish → Coverage verified, latency validated, full regression clean

### Recommended Execution Order

Since US1 and US2 share `server.go`:

1. Setup (Phase 1)
2. Foundational (Phase 2) — do first, unlocks everything
3. US1 (Phase 3) + US3 (Phase 5) in parallel — different files
4. US2 (Phase 4) after US1 — same file, builds on US1 helper
5. Polish (Phase 6)

---

## Notes

- Total: 34 tasks (2 setup, 7 foundational, 6 US1, 7 US2, 5 US3, 7 polish)
- All changes in existing files — no new files created
- ~150-200 lines of production code across ~8 files (5 Go + 3 Web UI)
- TDD enforced: tests T003-T005, T010-T012, T016-T019, T023 must fail before implementation
- Config version bump 10→11 with no-op migration (boolean defaults to false)
- `*bool` in settingsRequest to distinguish "not sent" from "set to false"
- FR-010 compliance: tag injection happens AFTER format transformation in `copyResponse()` — explicitly stated in T021
- FR-008 compliance: model extracted from response body (reflects actual upstream model); failover tested in T012
- Three SSE format variants explicitly covered: native Anthropic, native OpenAI Chat Completions, transformed OpenAI Responses API (T016, T017, T018)
- SC-003 latency validated via benchmark in T033
