# Tasks: Code Scenario Routing

**Input**: Design documents from `/specs/009-code-scenario-routing/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md

**Tests**: Required — constitution mandates TDD (Principle I, NON-NEGOTIABLE). Tests written first, verified to fail, then implementation to pass.

**Organization**: Tasks grouped by user story. US1 = backend routing logic (P0 MVP), US2 = Web UI configuration (P1).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2)
- Exact file paths included in all descriptions

---

## Phase 1: Foundational

**Purpose**: Add the shared `ScenarioCode` constant that both user stories depend on.

**⚠️ CRITICAL**: US1 and US2 cannot begin until this is complete.

- [ ] T001 Add `ScenarioCode Scenario = "code"` to the Scenario const block in `internal/config/config.go` (after `ScenarioBackground`, before `ScenarioDefault`)

**Checkpoint**: `go build ./...` succeeds. Existing tests still pass (`go test ./...`).

---

## Phase 2: User Story 1 — Route Coding Requests to Dedicated Provider/Model (Priority: P0) 🎯 MVP

**Goal**: The proxy detects regular coding requests (non-think, non-image, non-webSearch, non-longContext, non-background) and routes them to the `code` scenario's provider chain when configured.

**Independent Test**: Configure a profile with `think` → Provider A and `code` → Provider B. Send a thinking request → goes to A. Send a regular request → goes to B. Send a Haiku request → goes to `background`, not B.

### TDD Tests for User Story 1

> **Write these tests FIRST, verify they FAIL, then implement.**

- [ ] T002 [US1] Write TDD tests for `code` scenario detection in `internal/proxy/scenario_test.go`. Table-driven tests covering: (1) regular non-specialized request → `ScenarioCode`, (2) Haiku model request → `ScenarioBackground` not `ScenarioCode`, (3) thinking-enabled request → `ScenarioThink` not `ScenarioCode`, (4) image request → `ScenarioImage` not `ScenarioCode`, (5) webSearch request → `ScenarioWebSearch` not `ScenarioCode`, (6) backward compat: when detection returns `code` but no route configured, server falls through to default providers

### Implementation for User Story 1

- [ ] T003 [US1] Add `isCodeRequest()` function (returns `!isBackgroundRequest(body)`) and insert code scenario check into `DetectScenario()` between `isLongContext` and `isBackgroundRequest` checks in `internal/proxy/scenario.go`
- [ ] T004 [P] [US1] Add `{config.ScenarioCode, "code        (regular coding requests)"}` entry to `knownScenarios` slice in `tui/routing.go` (insert between `longContext` and `background` entries at line ~63)

**Checkpoint**: `go test ./internal/proxy/ -v` passes. All 6 new test cases green. All existing scenario tests unchanged and passing. TUI routing editor shows "code" scenario.

---

## Phase 3: User Story 2 — Configure Code Scenario via Web UI (Priority: P1)

**Goal**: The "Code" scenario appears in the Web UI profile editor's Routing tab, allowing users to configure providers and model overrides for coding requests through the UI.

**Independent Test**: Open profile edit page → Routing tab → "Code" scenario visible → add provider with model override → save → verify `zen.json` contains `"code"` route.

### Implementation for User Story 2

- [ ] T005 [US2] Add `'code'` to the `Scenario` type union, insert `'code'` into `SCENARIOS` array (between `'longContext'` and `'background'`), and add `code: 'Code'` to `SCENARIO_LABELS` map in `web/src/types/api.ts`
- [ ] T006 [P] [US2] Add `"scenarioCode": "Code"` to `web/src/i18n/locales/en.json` (after `scenarioLongContext` entry at line ~129)
- [ ] T007 [P] [US2] Add `"scenarioCode": "编程"` to `web/src/i18n/locales/zh-CN.json` (after `scenarioLongContext` entry at line ~129)
- [ ] T008 [P] [US2] Add `"scenarioCode": "編程"` to `web/src/i18n/locales/zh-TW.json` (after `scenarioLongContext` entry at line ~129)

**Checkpoint**: `pnpm run test` in `web/` passes. Profile edit page shows "Code" scenario in the Routing tab alongside existing scenarios. No changes needed to `edit.tsx` — it dynamically renders from `SCENARIOS`.

---

## Phase 4: Polish & Verification

**Purpose**: Ensure coverage thresholds are met and all tests pass.

- [ ] T009 Verify all Go tests pass and `internal/proxy` coverage ≥80% via `go test -cover ./internal/proxy/`
- [ ] T010 Verify frontend tests pass and coverage ≥70% via `pnpm run test:coverage` in `web/`

**Checkpoint**: All CI-mandated coverage thresholds met. `go test ./...` and `pnpm run test:coverage` both green.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: No dependencies — start immediately
- **US1 (Phase 2)**: Depends on Phase 1 (T001) — BLOCKS on ScenarioCode constant
- **US2 (Phase 3)**: Depends on Phase 1 (T001) — can run in PARALLEL with Phase 2
- **Polish (Phase 4)**: Depends on Phase 2 + Phase 3 completion

### User Story Dependencies

- **US1 (P0)**: Depends only on T001. Can start after Phase 1.
- **US2 (P1)**: Depends only on T001. Independent of US1. Can start after Phase 1.
- **US1 and US2 have NO cross-dependencies** — they modify completely different files.

### Within User Story 1

- T002 (tests) MUST be written and FAIL before T003 (implementation)
- T003 (scenario.go) depends on T002 (tests exist to verify)
- T004 (tui/routing.go) is [P] — can run in parallel with T002/T003 (different file)

### Within User Story 2

- T005 (api.ts) is the primary type change
- T006, T007, T008 (i18n files) are all [P] — can run in parallel with T005 and each other

### Parallel Opportunities

```
Phase 1:  T001
              ├──────────────────────────────┐
Phase 2:  T002 → T003                    Phase 3: T005
          T004 [P]                                 T006 [P]
                                                   T007 [P]
                                                   T008 [P]
              └──────────────────────────────┘
Phase 4:  T009, T010
```

---

## Implementation Strategy

### MVP First (US1 Only)

1. Complete Phase 1: T001 (ScenarioCode constant)
2. Complete Phase 2: T002 → T003, T004 (detection logic + TUI)
3. **STOP and VALIDATE**: Run `go test ./internal/proxy/ -v` — all tests pass
4. Feature is functional via JSON config and TUI at this point

### Incremental Delivery

1. T001 → Foundation ready
2. T002 → T003 + T004 → US1 complete → routing works (MVP!)
3. T005 + T006 + T007 + T008 → US2 complete → Web UI shows Code scenario
4. T009 + T010 → Coverage verified → PR-ready

---

## Notes

- [P] tasks = different files, no dependencies
- Constitution Principle I (TDD) requires T002 before T003
- US1 and US2 touch entirely different file sets — fully parallelizable
- No config version bump needed (routing map uses string keys)
- `edit.tsx` needs NO changes — it dynamically renders from `SCENARIOS` array
- Commit after each task per constitution Principle IV
