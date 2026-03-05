# Tasks: Skill-Based Intent Recognition

**Input**: Design documents from `/specs/003-skill-intent-recognition/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included (TDD required by constitution — Principle I)

**Organization**: Tasks grouped by user story. US1 and US2 are both P1 but US1 is foundational for US2.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Config & Types)

**Purpose**: Config schema changes and core type definitions needed by all stories

- [x] T001 Add SkillsConfig, SkillDefinition types to internal/config/config.go (per data-model.md)
- [x] T002 Write config migration tests: old format parsing, round-trip, field preservation in internal/config/config_test.go
- [x] T003 Bump CurrentConfigVersion v9→v10 and add migration logic for missing skills field in internal/config/config.go (make T002 tests pass)
- [x] T004 Add compat helpers for Skills config access in internal/config/compat.go

**Checkpoint**: Config compiles, `go test ./internal/config/...` passes, v9 configs migrate cleanly to v10

---

## Phase 2: Foundational (Skill Core)

**Purpose**: Skill type, registry, and builtin definitions — blocks US1-US4

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T005 Write tests for Skill struct, validation, and SkillRegistry (NewSkill, Validate, Register, Get, List, Enable/Disable) in internal/bot/skill_test.go
- [x] T006 Implement Skill struct, validation, and SkillRegistry in internal/bot/skill.go (make T005 tests pass)
- [x] T007 [P] Write tests for builtin skill definitions (all existing intents covered: control, bind, approve, send_task, persona, forget, query_status, query_list) in internal/bot/builtin_skills_test.go
- [x] T008 [P] Implement builtin skill definitions with en/zh keywords in internal/bot/builtin_skills.go (make T007 tests pass)

**Checkpoint**: `go test ./internal/bot/... -run TestSkill` passes, all builtin skills load and validate

---

## Phase 3: User Story 1 - Skill 定义与注册 (Priority: P1) 🎯 MVP

**Goal**: 管理员可通过配置定义 Skill，系统启动时加载并注册，支持热重载

**Independent Test**: 在 zen.json 中添加自定义 Skill，重启/重载后验证 Skill 出现在注册列表中

### Tests for User Story 1

> **Write tests FIRST, ensure they FAIL before implementation**

- [x] T009 [P] [US1] Write tests for SkillRegistry loading from config (load builtin + custom, merge, validate, skip invalid) in internal/bot/skill_test.go
- [x] T010 [P] [US1] Write tests for SkillRegistry hot-reload (config change triggers re-registration) in internal/bot/skill_test.go

### Implementation for User Story 1

- [x] T011 [US1] Implement LoadFromConfig on SkillRegistry: merge builtin + custom skills from SkillsConfig in internal/bot/skill.go (make T009 pass)
- [x] T012 [US1] Implement Reload on SkillRegistry: re-merge skills on config change in internal/bot/skill.go (make T010 pass)
- [x] T013 [US1] Integrate SkillRegistry into Gateway: init on startup, reload on config change in internal/bot/gateway.go

**Checkpoint**: Skills load from config on startup, hot-reload works, `go test ./internal/bot/... -run TestSkill` passes

---

## Phase 4: User Story 2 - Skill 驱动的意图识别 (Priority: P1)

**Goal**: 用户发送自然语言消息时，系统通过 Skill 匹配识别意图，回退到 LLM 处理模糊情况

**Independent Test**: 发送"帮我暂停一下"，验证识别为 IntentControl 而非 IntentChat

### Tests for User Story 2

> **Write tests FIRST, ensure they FAIL before implementation**

- [x] T014 [P] [US2] Write tests for local matcher: keyword matching, synonym matching, fuzzy matching, score weighting in internal/bot/matcher_test.go
- [x] T015 [P] [US2] Write tests for LLM fallback matcher: prompt construction, JSON response parsing, timeout handling, error fallback in internal/bot/matcher_test.go
- [x] T016 [P] [US2] Write tests for parameter extraction via LLM: control intent extracts action/target, send_task extracts target/task in internal/bot/matcher_test.go

### Implementation for User Story 2

- [x] T017 [US2] Implement local matcher (keyword + synonym + fuzzy scoring) in internal/bot/matcher.go (make T014 pass)
- [x] T018 [US2] Implement LLM fallback matcher (prompt build, call LLMClient, parse JSON response) in internal/bot/matcher.go (make T015 pass)
- [x] T019 [US2] Implement parameter extraction via LLM (single call to extract action/target from message given intent type) in internal/bot/matcher.go (make T016 pass)
- [x] T020 [US2] Implement SkillMatcher.Match orchestrator: regex first → local skill match → LLM fallback → IntentChat in internal/bot/matcher.go
- [x] T021 [US2] Integrate SkillMatcher into NLUParser.Parse: after regex miss, call SkillMatcher before falling back to IntentChat in internal/bot/nlu.go
- [x] T022 [US2] Update existing NLU tests to verify Skill integration path (regex still wins, Skill catches natural language, fallback to chat) in internal/bot/nlu_test.go
- [x] T023 [US2] Update handlers.go: when Skill match returns intent needing params, route through parameter extraction before dispatching to handler in internal/bot/handlers.go

**Checkpoint**: "帮我暂停一下" → IntentControl, "pause" → still regex, "随便聊聊" → IntentChat. `go test ./internal/bot/...` passes

---

## Phase 5: User Story 3 - Skill 管理与调试 (Priority: P2)

**Goal**: 管理员通过 Web UI 管理 Skill（CRUD、启用/禁用）并测试意图匹配，查看匹配日志

**Independent Test**: 通过 Web API 创建自定义 Skill，调用 test 端点验证匹配结果，查看 logs

### Tests for User Story 3

> **Write tests FIRST, ensure they FAIL before implementation**

- [x] T024 [P] [US3] Write tests for MatchLog ring buffer (add, list with limit, overflow eviction) in internal/bot/matcher_test.go
- [x] T025 [P] [US3] Write tests for Skill Web API endpoints (GET list, POST create, PUT update, DELETE, PUT config, POST test, GET logs) in internal/web/api_bot_skills_test.go

### Implementation for User Story 3

- [x] T026 [US3] Implement MatchLog ring buffer in internal/bot/matcher.go (make T024 pass)
- [x] T027 [US3] Record match logs in SkillMatcher.Match after each match attempt in internal/bot/matcher.go
- [x] T028 [US3] Implement Skill CRUD API handlers (GET /api/v1/bot/skills, POST, PUT /{name}, DELETE /{name}) in internal/web/api_bot_skills.go
- [x] T029 [US3] Implement Skill config API handler (PUT /api/v1/bot/skills/config) in internal/web/api_bot_skills.go
- [x] T030 [US3] Implement Skill test API handler (POST /api/v1/bot/skills/test) in internal/web/api_bot_skills.go
- [x] T031 [US3] Implement Skill logs API handler (GET /api/v1/bot/skills/logs) in internal/web/api_bot_skills.go
- [x] T032 [US3] Register Skill API routes in internal/web/server.go
- [ ] T033 [US3] Add Skill management UI components (skill list, create/edit form, test panel, log viewer) in internal/web/static/app.js

**Checkpoint**: All 7 API endpoints functional, Web UI shows skills and test results. `go test ./internal/web/... -run TestSkill` passes

---

## Phase 6: User Story 4 - 多语言意图识别 (Priority: P3)

**Goal**: 中文和其他非英语语言的自然语言指令被正确识别，准确率差异 ≤5%

**Independent Test**: 用中文发送"查看所有进程状态"，验证识别为 IntentQueryStatus

### Tests for User Story 4

- [x] T034 [P] [US4] Write tests for multi-language keyword matching: zh keywords match Chinese input, en keywords match English input, mixed language input in internal/bot/matcher_test.go
- [x] T035 [P] [US4] Write tests for multi-language LLM fallback: Chinese natural language correctly classified in internal/bot/matcher_test.go

### Implementation for User Story 4

- [x] T036 [US4] Ensure builtin skills have comprehensive zh keyword coverage (review and expand Chinese synonyms) in internal/bot/builtin_skills.go
- [x] T037 [US4] Add language detection heuristic to local matcher for selecting keyword set in internal/bot/matcher.go (make T034 pass)
- [x] T038 [US4] Update LLM fallback prompt to include multi-language instruction and examples in internal/bot/matcher.go (make T035 pass)

**Checkpoint**: "查看所有进程状态" → IntentQueryStatus, "帮我绑定到myproject" → IntentBind. `go test ./internal/bot/... -run TestMultiLang` passes

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Coverage verification, edge cases, integration validation

- [x] T039 Verify test coverage ≥80% for internal/bot with `go test -cover ./internal/bot/`
- [x] T040 Verify test coverage ≥80% for internal/config with `go test -cover ./internal/config/`
- [x] T041 [P] Add edge case tests: short messages (emoji, single char), concurrent skill reload, multiple skills same confidence in internal/bot/matcher_test.go
- [x] T042 [P] Add edge case tests: LLM timeout fallback, invalid LLM response parsing in internal/bot/matcher_test.go
- [x] T043 [P] Add accuracy benchmark test: predefined test case set (≥20 cases covering all intents) verifying ≥85% correct recognition rate (SC-001) in internal/bot/matcher_test.go
- [x] T044 [P] Add latency benchmark test: verify local match ≤500ms for 95th percentile, LLM fallback ≤2s (SC-004) in internal/bot/matcher_test.go
- [x] T045 [P] Add frontend smoke tests for Skill management UI (skill list render, create form, test panel interaction) in internal/web/api_bot_skills_test.go
- [x] T046 Run full test suite `go test ./...` and fix any failures
- [x] T047 Run quickstart.md validation (dev daemon start, API smoke test)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (config types must exist)
- **US1 (Phase 3)**: Depends on Phase 2 (Skill types and registry)
- **US2 (Phase 4)**: Depends on Phase 3 (Skills must be registered to match against)
- **US3 (Phase 5)**: Depends on Phase 4 (matcher must exist for test/logs endpoints)
- **US4 (Phase 6)**: Depends on Phase 4 (matcher must exist for multi-language testing)
- **Polish (Phase 7)**: Depends on all story phases complete

### User Story Dependencies

- **US1 (P1)**: Foundational → can start after Phase 2
- **US2 (P1)**: Depends on US1 (needs registered Skills to match against)
- **US3 (P2)**: Depends on US2 (needs matcher for test/logs). Can parallel with US4
- **US4 (P3)**: Depends on US2 (needs matcher for multi-language). Can parallel with US3

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD)
- Types/models before logic
- Core implementation before integration
- Commit after each task

### Parallel Opportunities

- T002 + T004 (config tests and compat helpers can start after T001 types exist)
- T005 + T007 (foundational tests)
- T006 + T008 (foundational impl, after respective tests)
- T009 + T010 (US1 tests)
- T014 + T015 + T016 (US2 tests)
- T024 + T025 (US3 tests)
- T034 + T035 (US4 tests)
- US3 + US4 (after US2 complete)
- T041 + T042 + T043 + T044 + T045 (polish edge cases & benchmarks)

---

## Parallel Example: User Story 2

```bash
# Launch all US2 tests in parallel (different test functions, same file):
Task T014: "Write local matcher tests in internal/bot/matcher_test.go"
Task T015: "Write LLM fallback tests in internal/bot/matcher_test.go"
Task T016: "Write parameter extraction tests in internal/bot/matcher_test.go"

# Then implement sequentially (same file, dependent logic):
Task T017: "Implement local matcher" (makes T014 pass)
Task T018: "Implement LLM fallback" (makes T015 pass)
Task T019: "Implement parameter extraction" (makes T016 pass)
Task T020: "Implement Match orchestrator" (ties T017-T019 together)
```

---

## Implementation Strategy

### MVP First (US1 + US2)

1. Complete Phase 1: Config types + migration
2. Complete Phase 2: Skill core types + builtins
3. Complete Phase 3: US1 — Skill registration from config
4. Complete Phase 4: US2 — Matcher engine + NLU integration
5. **STOP and VALIDATE**: "帮我暂停一下" correctly identified as IntentControl
6. This is a functional MVP — natural language intent recognition works

### Incremental Delivery

1. Setup + Foundational + US1 + US2 → MVP (core matching works)
2. Add US3 → Management UI + debugging tools
3. Add US4 → Multi-language coverage
4. Polish → Coverage, edge cases, validation

---

## Notes

- [P] tasks = different files or independent test functions, no dependencies
- [Story] label maps task to specific user story for traceability
- Constitution requires TDD: write tests first, verify they fail, then implement
- Commit after each task (Principle IV)
- Config version bump (v9→v10) in Phase 1 is blocking — do it first (T001 types → T002 tests → T003 migration)
- internal/bot must maintain ≥80% test coverage (Principle VI)
- SC-001 accuracy benchmark and SC-004 latency benchmark validated in Phase 7 (T043, T044)
