# Tasks: Proxy Transform Layer Correctness

**Input**: Design documents from `/specs/018-proxy-transform-correctness/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅

**Tests**: TDD required per Constitution Principle I — test tasks are included and MUST be written first.

**Organization**: Tasks grouped by user story (Phases 1-5 from plan.md map to US1-US5).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to
- Paths are relative to repository root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish protocol format constants — the foundation all other phases depend on.

- [X] T001 Add `FormatAnthropicMessages`, `FormatOpenAIChat`, `FormatOpenAIResponses` constants to `internal/proxy/transform/transform.go`
- [X] T002 Update `NeedsTransform` in `internal/proxy/transform/transform.go` to handle new format strings
- [X] T003 Update `detectClientFormat` in `internal/proxy/profile_proxy.go` to return new format constants instead of `"openai"`/`"anthropic"`

**Checkpoint**: Format constants defined and detection updated — all downstream phases can now use the new identifiers.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Remove debug file I/O from hot path — independent of all user stories, unblocks clean builds.

**⚠️ CRITICAL**: Complete before any streaming or transform work begins.

- [X] T004 Delete `init()` function, `debugLogger` var, and all `debugLogger.Printf(...)` call sites from `internal/proxy/transform/anthropic.go`
- [X] T005 Verify `go build ./...` passes after logging removal

**Checkpoint**: No file I/O in transform hot path — clean build confirmed.

---

## Phase 3: User Story 1 — Protocol-Aware Request Transformation (Priority: P0) 🎯 MVP

**Goal**: Proxy correctly routes `openai-chat` and `openai-responses` client formats through separate transformation paths.

**Independent Test**: Send Chat Completions and Responses API requests through proxy to Anthropic provider; validate response shapes match expected protocol schemas.

### Tests for User Story 1

> **Write these tests FIRST, ensure they FAIL before implementation**

- [X] T006 [P] [US1] Add table-driven tests for `detectClientFormat` covering all three format return values in `internal/proxy/profile_proxy_test.go`
- [X] T007 [P] [US1] Add table-driven tests for `NeedsTransform` with new format constants in `internal/proxy/transform/transform_test.go`
- [X] T008 [P] [US1] Add tests for `StreamTransformer` routing: `openai-chat`→anthropic and `openai-responses`→anthropic paths in `internal/proxy/transform/stream_test.go`

### Implementation for User Story 1

- [X] T009 [US1] Update `StreamTransformer.TransformSSEStream` routing in `internal/proxy/transform/stream.go` to branch on `openai-chat` vs `openai-responses` (currently both map to same path)
- [X] T010 [US1] Update `AnthropicTransformer.TransformRequest` in `internal/proxy/transform/anthropic.go` to handle `openai-chat` and `openai-responses` client formats distinctly
- [X] T011 [US1] Update `AnthropicTransformer.TransformResponse` in `internal/proxy/transform/anthropic.go` to produce correct response shape per client format
- [X] T012 [US1] Update `OpenAITransformer.TransformRequest` in `internal/proxy/transform/openai.go` to handle `anthropic-messages` client format
- [X] T013 [US1] Update `OpenAITransformer.TransformResponse` in `internal/proxy/transform/openai.go` to handle `anthropic-messages` client format

**Checkpoint**: Chat Completions and Responses API requests produce correct response shapes independently.

---

## Phase 4: User Story 2 — Complete Tool Call Transformation (Priority: P0)

**Goal**: Bidirectional tool call transformation works correctly in both streaming and non-streaming modes.

**Independent Test**: Send requests with tool definitions through proxy; verify tool invocations and results transform correctly in both directions and streaming modes.

### Tests for User Story 2

> **Write these tests FIRST, ensure they FAIL before implementation**

- [X] T014 [P] [US2] Add table-driven tests for OpenAI Chat `tools` → Anthropic `tools` request transformation in `internal/proxy/transform/openai_test.go`
- [X] T015 [P] [US2] Add table-driven tests for Anthropic `tool_use` → OpenAI Chat `tool_calls` response transformation in `internal/proxy/transform/anthropic_test.go`
- [X] T016 [P] [US2] Add streaming tests: Anthropic `content_block_start(tool_use)` + `input_json_delta` → OpenAI Chat `tool_calls` deltas in `internal/proxy/transform/stream_test.go`
- [X] T017 [P] [US2] Add streaming tests: Anthropic `input_json_delta` → OpenAI Responses `response.function_call_arguments.delta` in `internal/proxy/transform/stream_test.go`

### Implementation for User Story 2

- [X] T018 [US2] Fix/verify OpenAI Chat `tools` array → Anthropic `tools` with `input_schema` in `internal/proxy/transform/openai.go`
- [X] T019 [US2] Fix/verify Anthropic `tool_use` content blocks → OpenAI Chat `tool_calls` array in `internal/proxy/transform/anthropic.go`
- [X] T020 [US2] Implement streaming: Anthropic `content_block_start(tool_use)` → OpenAI Chat `tool_calls[].id` + `function.name` delta in `internal/proxy/transform/stream.go`
- [X] T021 [US2] Implement streaming: Anthropic `input_json_delta` → OpenAI Chat `tool_calls[].function.arguments` delta in `internal/proxy/transform/stream.go`
- [X] T022 [US2] Implement streaming: Anthropic `content_block_start(tool_use)` → OpenAI Responses `response.output_item.added` in `internal/proxy/transform/stream.go`
- [X] T023 [US2] Implement streaming: Anthropic `input_json_delta` → OpenAI Responses `response.function_call_arguments.delta` in `internal/proxy/transform/stream.go`

**Checkpoint**: Tool calls transform correctly in all directions and streaming modes.

---

## Phase 5: User Story 3 — Safe SSE Stream Error Handling (Priority: P0)

**Goal**: Upstream stream errors propagate to client as protocol-native error events; no fake completions.

**Independent Test**: Simulate truncated/errored upstream streams; verify proxy emits error events instead of completion events.

### Tests for User Story 3

> **Write these tests FIRST, ensure they FAIL before implementation**

- [X] T024 [P] [US3] Add test: truncated reader causes `scanner.Err()` → error event emitted (not `message_stop`) for `transformAnthropicToOpenAI` in `internal/proxy/transform/stream_test.go`
- [X] T025 [P] [US3] Add test: truncated reader causes `scanner.Err()` → error event emitted for `transformOpenAIToAnthropic` in `internal/proxy/transform/stream_test.go`
- [X] T026 [P] [US3] Add test: truncated reader causes `scanner.Err()` → error event emitted for `transformResponsesAPIToAnthropic` in `internal/proxy/transform/stream_test.go`
- [X] T027 [P] [US3] Add test: clean EOF causes `scanner.Err() == nil` → correct completion event emitted for each streaming path in `internal/proxy/transform/stream_test.go`

### Implementation for User Story 3

- [X] T028 [US3] Add `writeStreamError(pw io.Writer, clientFormat string, err error)` helper in `internal/proxy/transform/stream.go` emitting protocol-native error events per data-model.md
- [X] T029 [US3] Add `scanner.Err()` check after loop in `transformAnthropicToOpenAI` in `internal/proxy/transform/stream.go`; call `writeStreamError` if non-nil, skip completion event
- [X] T030 [US3] Add `scanner.Err()` check after loop in `transformOpenAIToAnthropic` in `internal/proxy/transform/stream.go`; call `writeStreamError` if non-nil, skip completion event
- [X] T031 [US3] Add `scanner.Err()` check after loop in `transformResponsesAPIToAnthropic` in `internal/proxy/transform/stream.go`; call `writeStreamError` if non-nil, skip completion event

**Checkpoint**: All three streaming paths propagate errors correctly; no fake completions on broken streams.

---

## Phase 6: User Story 4 — Transform Error Classification (Priority: P1)

**Goal**: Transform failures return HTTP 500 to client without marking provider unhealthy.

**Independent Test**: Inject transform failures; verify provider health unchanged and correct error response returned.

### Tests for User Story 4

> **Write these tests FIRST, ensure they FAIL before implementation**

- [X] T032 [P] [US4] Add test: request transform error → HTTP 500 returned, provider `MarkFailed` NOT called in `internal/proxy/server_test.go`
- [X] T033 [P] [US4] Add test: response transform error → HTTP 500 returned, provider health unchanged in `internal/proxy/server_test.go`

### Implementation for User Story 4

- [X] T034 [US4] Define `TransformError` type with `Op string` and `Err error` fields in `internal/proxy/server.go`
- [X] T035 [US4] Update `forwardRequest` in `internal/proxy/server.go` to return `&TransformError{Op: "request", Err: err}` when `TransformRequest` fails (instead of silently continuing)
- [X] T036 [US4] Update `tryProviders` in `internal/proxy/server.go` to detect `TransformError` via `errors.As`; return HTTP 500 with `{"error":{"type":"transform_error","message":"..."}}` without calling any provider health methods

**Checkpoint**: Transform errors isolated from provider health; providers not penalized for local bugs.

---

## Phase 7: User Story 5 — Remove Transform Hot Path Logging (Priority: P1)

**Already handled in Phase 2 (T004)**. This phase validates the removal is complete and no regressions introduced.

**Independent Test**: Run proxy under load; verify no `~/.zen-dev/transform.log` created and no file I/O in transform path.

### Tests for User Story 5

> **Write these tests FIRST, ensure they FAIL before implementation**

- [X] T037 [US5] Add test: verify no `debugLogger` references exist in `internal/proxy/transform/` package via programmatic check in `internal/proxy/transform/transform_test.go`

### Validation for User Story 5

- [X] T038 [US5] Verify no references to `debugLogger` remain in `internal/proxy/transform/` via `grep -r debugLogger internal/proxy/transform/`
- [X] T039 [US5] Verify `~/.zen-dev/transform.log` is not created on proxy startup after changes

**Checkpoint**: Zero file I/O in transform hot path confirmed.

---

## Phase 8: Polish & Cross-Cutting Concerns

- [X] T040 [P] Run `go test ./internal/proxy/... -cover` and verify `internal/proxy/transform` coverage ≥ 80%
- [X] T041 [P] Run `go test ./internal/proxy/... -race` to verify no data races introduced
- [X] T042 Run `go build ./...` for final clean build verification
- [X] T043 Review `git status` and remove any generated/temporary files before committing

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: No dependencies — can run in parallel with Phase 1
- **Phase 3 (US1)**: Depends on Phase 1 completion (needs format constants)
- **Phase 4 (US2)**: Depends on Phase 3 completion (needs protocol routing)
- **Phase 5 (US3)**: Depends on Phase 1 completion (needs format constants for error events)
- **Phase 6 (US4)**: Independent of Phases 3–5 — can start after Phase 1
- **Phase 7 (US5)**: Already done in Phase 2 — validation only
- **Phase 8 (Polish)**: Depends on all phases complete

### User Story Dependencies

- **US1 (Protocol Routing)**: Blocks US2 and US3 (streaming paths depend on correct format routing)
- **US2 (Tool Calls)**: Depends on US1
- **US3 (SSE Errors)**: Depends on US1 (needs `writeStreamError` to know client format)
- **US4 (Error Classification)**: Independent — can run in parallel with US1–US3
- **US5 (Logging)**: Done in Phase 2

### Parallel Opportunities

- T001, T002, T003 can run in parallel (different files)
- T004, T005 can run in parallel with Phase 1
- T006, T007, T008 (US1 tests) can run in parallel
- T014, T015, T016, T017 (US2 tests) can run in parallel
- T024, T025, T026, T027 (US3 tests) can run in parallel
- T032, T033 (US4 tests) can run in parallel
- T040, T041 (coverage + race) can run in parallel

---

## Parallel Example: User Story 3 Tests

```bash
# Write all US3 tests in parallel (different test cases, same file):
Task: T024 - truncated reader test for transformAnthropicToOpenAI
Task: T025 - truncated reader test for transformOpenAIToAnthropic
Task: T026 - truncated reader test for transformResponsesAPIToAnthropic
Task: T027 - clean EOF completion event tests
```

---

## Implementation Strategy

### MVP First (US1 + US3 — Core Correctness)

1. Complete Phase 1: Format constants
2. Complete Phase 2: Remove debug logging
3. Complete Phase 3: Protocol routing (US1)
4. Complete Phase 5: SSE error safety (US3)
5. **STOP and VALIDATE**: Proxy correctly routes protocols and handles stream errors

### Incremental Delivery

1. Phase 1+2 → Clean foundation
2. Phase 3 (US1) → Protocol routing correct
3. Phase 4 (US2) → Tool calls work
4. Phase 5 (US3) → Streams fail safely
5. Phase 6 (US4) → Provider health protected
6. Phase 8 → Coverage verified, PR ready

---

## Notes

- [P] tasks = different files or independent test cases, no blocking dependencies
- TDD required: all test tasks (T006–T008, T014–T017, T024–T027, T032–T033, T037) MUST be written and confirmed failing before their implementation tasks
- Commit after each phase checkpoint
- Constitution Principle VI: verify `internal/proxy/transform` coverage ≥ 80% before opening PR
