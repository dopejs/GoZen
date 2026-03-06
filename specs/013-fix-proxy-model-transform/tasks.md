# Tasks: Fix Proxy Model Transform for Cross-Format Providers

**Input**: Design documents from `/specs/013-fix-proxy-model-transform/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Required — TDD is NON-NEGOTIABLE per project constitution. Tests must be written first and verified to FAIL before implementation.

**Organization**: Tasks are grouped by user story. US1 and US2 are independent bugs (different files) and can be implemented in parallel. US3 is an integration verification that depends on both.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Verify green baseline before making changes

- [x] T001 Run existing proxy and transform test suites to establish green baseline (`go test ./internal/proxy/... && go test ./internal/proxy/transform/...`)

---

## Phase 2: User Story 1 - Path Deduplication (Priority: P1) 🎯 MVP

**Goal**: Fix path duplication when `base_url` already includes `/v1`, preventing double `/v1/v1/...` in the final URL.

**Independent Test**: Send a request through a provider with `/v1` in base_url and verify the final URL has a single `/v1/chat/completions`.

### Tests for User Story 1 (Red Phase)

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T002 [US1] Write failing table-driven test cases for path deduplication in `internal/proxy/server_test.go` — cover scenarios: (1) base_url `https://host/v1` + targetPath `/v1/chat/completions` → no duplication, (2) base_url `https://host` + targetPath `/v1/chat/completions` → correctly appended, (3) base_url `https://host/v1/` (trailing slash) + targetPath `/v1/chat/completions` → no duplication, (4) base_url `https://host/api/v1` + targetPath `/v1/chat/completions` → correct handling, (5) Codex client path `/v1/chat/completions` through OpenAI provider with `/v1` base_url → no duplication

### Implementation for User Story 1 (Green Phase)

- [x] T003 [US1] Implement path deduplication in `forwardRequest()` in `internal/proxy/server.go` (~line 465) — strip `/v1` prefix from `targetPath` when `base_url` path already ends with `/v1`, using `strings.TrimSuffix`/`strings.HasSuffix`/`strings.HasPrefix` pattern from plan.md; add debug log line showing original and deduplicated path
- [x] T004 [US1] Run US1 tests and full proxy test suite to verify all pass with zero regressions in `internal/proxy/server_test.go`

**Checkpoint**: Path deduplication fixed — providers with `/v1` in base_url no longer return 404.

---

## Phase 3: User Story 2 - Type-Aware Default Model Filling (Priority: P1)

**Goal**: Prevent `buildProviders()` from filling Anthropic default model names into OpenAI-type providers, allowing `mapModel()` to fall through to the provider's own `model` field.

**Independent Test**: Send `model: claude-sonnet-4-6` through a provider with only `model: "gpt-5.3-codex"` and verify it maps to `gpt-5.3-codex`.

### Tests for User Story 2 (Red Phase)

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T005 [US2] Write failing test cases for type-aware default model filling in `internal/proxy/server_test.go` — cover scenarios: (1) OpenAI provider with only `model` field, no tier-specific models → tier fields remain empty after `buildProviders()`, (2) OpenAI provider with explicit `sonnet_model` → explicit value preserved, (3) Anthropic provider with no tier-specific models → defaults filled (backward compat), (4) OpenAI provider with empty `model` → original model name preserved
- [x] T006 [US2] Write failing test cases for `mapModel()` fallthrough behavior in `internal/proxy/server_test.go` — cover scenarios: (1) OpenAI provider, empty sonnet_model, request `claude-sonnet-4-6` → falls through to `p.Model`, (2) OpenAI provider, explicit sonnet_model `gpt-5.4`, request `claude-sonnet-4-6` → returns `gpt-5.4`, (3) OpenAI provider, request `claude-opus-4-6` → falls through to `p.Model`

### Implementation for User Story 2 (Green Phase)

- [x] T007 [US2] Add provider type guard to `buildProviders()` in `internal/proxy/profile_proxy.go` (~line 186) — wrap each default model assignment with `isAnthropic := pc.GetType() == config.ProviderTypeAnthropic` check, only fill Anthropic defaults when `isAnthropic` is true; add debug log line showing model mapping result
- [x] T008 [US2] Run US2 tests and full proxy test suite to verify all pass with zero regressions in `internal/proxy/server_test.go`

**Checkpoint**: OpenAI providers no longer get Anthropic default models — `mapModel()` correctly falls through to `p.Model`.

---

## Phase 4: User Story 3 - End-to-End Cross-Format Verification (Priority: P2)

**Goal**: Validate the complete Anthropic↔OpenAI transformation pipeline works after both bug fixes, including streaming responses.

**Independent Test**: Send non-streaming and streaming Anthropic requests through an OpenAI provider and verify correct response transformation.

### Tests for User Story 3

- [ ] T009 [US3] Write integration test for non-streaming Anthropic→OpenAI full pipeline in `internal/proxy/server_test.go` — mock OpenAI backend, send Anthropic-format request, verify: correct path (no /v1 duplication), correct model mapping (falls through to provider model), model mapping applied before format transformation (FR-004), response transformed back to Anthropic format with `id`, `model`, `content`, `usage` fields
- [ ] T010 [US3] Write integration test for streaming Anthropic→OpenAI SSE pipeline in `internal/proxy/server_test.go` — mock OpenAI SSE backend, send streaming Anthropic request, verify SSE events are correctly transformed from OpenAI to Anthropic format
- [ ] T011 [US3] Write edge case tests in `internal/proxy/server_test.go` — cover: (1) upstream returns 4xx/5xx error → error properly propagated, (2) request body has no `model` field → body passed through unchanged

**Checkpoint**: Full Anthropic↔OpenAI pipeline verified for both streaming and non-streaming.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Coverage, manual validation, and final checks

- [ ] T012 Run coverage check with `go test -cover ./internal/proxy/...` and verify ≥80% for proxy package
- [ ] T013 [P] Restart dev daemon (`./scripts/dev.sh restart`) and run quickstart.md E2E curl commands for manual verification against real provider
- [ ] T014 [P] Review all changes for backward compatibility — verify existing Anthropic provider configs produce identical behavior

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — must complete first to establish baseline
- **US1 (Phase 2)**: Depends on Setup — can run in parallel with US2 (different files: `server.go` vs `profile_proxy.go`)
- **US2 (Phase 3)**: Depends on Setup — can run in parallel with US1 (different files)
- **US3 (Phase 4)**: Depends on BOTH US1 and US2 being complete (integration test requires both fixes)
- **Polish (Phase 5)**: Depends on US3 completion

### User Story Dependencies

- **User Story 1 (P1)**: Independent — modifies `internal/proxy/server.go` only
- **User Story 2 (P1)**: Independent — modifies `internal/proxy/profile_proxy.go` only
- **User Story 3 (P2)**: Depends on US1 + US2 — integration verification across both fixes

### Within Each User Story (TDD Cycle)

1. Write tests → verify they FAIL (Red)
2. Implement fix (Green)
3. Run full test suite → verify zero regressions
4. Commit

### Parallel Opportunities

- **US1 and US2 implementation tasks can run in parallel** — T003 (`server.go`) and T007 (`profile_proxy.go`) modify different files with no dependencies
- **Test-writing tasks (T002 vs T005/T006) must be sequential** — both write to `server_test.go`, so they cannot run concurrently
- T005 and T006 can be written together as they're both US2 test functions (same file, different test functions)
- T013 and T014 can run in parallel (different activities)

---

## Parallel Example: US1 + US2

```
# These two story phases can execute concurrently:

Stream A (US1 — server.go):
  T002 → T003 → T004

Stream B (US2 — profile_proxy.go):
  T005 → T006 → T007 → T008

# Then converge for US3:
  T009 → T010 → T011

# Then polish:
  T012, T013 [P], T014 [P]
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (verify green baseline)
2. Complete Phase 2: User Story 1 (path deduplication)
3. **STOP and VALIDATE**: Test US1 independently — providers with `/v1` in base_url should work
4. This alone fixes the 404 error for `cctq-codex` provider

### Incremental Delivery

1. Setup → baseline green
2. US1 → path dedup fixed → commit
3. US2 → model defaults fixed → commit
4. US3 → full pipeline verified → commit
5. Polish → coverage ≥80%, manual E2E → ready for PR

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- US1 and US2 are independent bugs — can be fixed in either order
- All tests follow existing table-driven patterns per research.md (R4)
- ~30 lines of code changes, ~100 lines of new tests (per plan.md estimate)
- No config schema changes needed — no migration required
- Per Constitution Principle IV: commit after each completed TDD cycle (tests + implementation per user story)
