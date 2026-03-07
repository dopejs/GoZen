# Tasks: Fix OpenAI Responses API Transform

**Input**: Design documents from `/specs/014-fix-responses-api-transform/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, quickstart.md

**Tests**: TDD is NON-NEGOTIABLE per constitution. All tests written first and verified to FAIL before implementation.

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Verify baseline before any changes

- [x] T001 Run `go test ./internal/proxy/... -cover` to verify all existing tests pass and record baseline coverage

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Build shared transform functions needed by ALL user stories

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Tests (RED) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T002 [P] Write table-driven tests for `isResponsesAPIRequired` in `internal/proxy/server_test.go`: (1) body containing `"input is required"` → true, (2) body with `"server error"` → false, (3) body with `"invalid_request_error"` → false, (4) empty body → false, (5) malformed JSON → false
- [x] T003 [P] Write table-driven tests for `ChatCompletionsToResponsesAPI` in `internal/proxy/transform/responses_test.go`: (1) messages→input rename, (2) max_completion_tokens→max_output_tokens, (3) tool flattening (unwrap function wrapper), (4) removal of unsupported fields (n, logprobs, stream_options, presence_penalty, frequency_penalty, seed, response_format), (5) passthrough of model/stream/temperature/top_p/tool_choice/stop, (6) store field set to false in output

### Implementation (GREEN)

- [x] T004 Implement `isResponsesAPIRequired(body []byte) bool` in `internal/proxy/server.go`: substring match on `"input is required"` in response body, return true if found (all T002 tests pass)
- [x] T005 Implement `ChatCompletionsToResponsesAPI(body []byte) ([]byte, error)` in `internal/proxy/transform/responses.go`: rename messages→input, rename max_completion_tokens→max_output_tokens, flatten tools (extract name/description/parameters from nested function wrapper to top level), delete Chat Completions-only fields, set `store: false` to prevent server-side storage on provider (all T003 tests pass)

**Checkpoint**: Foundation functions ready — `isResponsesAPIRequired` detects when retry is needed, `ChatCompletionsToResponsesAPI` transforms the request body

---

## Phase 3: User Story 1 — Responses API Retry Logic (Priority: P1) 🎯 MVP

**Goal**: When Chat Completions returns "input is required", automatically retry with Responses API format

**Independent Test**: Mock server returns 500 "input is required" on `/chat/completions` and 200 Responses API format on `/responses` → proxy retries and succeeds

> **Traceability note**: T006-T007 build the base `ResponsesAPIToAnthropic` transform which also satisfies US3 acceptance scenario 1 (non-streaming response field mapping). US3 Phase 5 extends this with tool call handling (scenario 2).

### Tests (RED) ⚠️

- [x] T006 [US1] Write table-driven tests for `ResponsesAPIToAnthropic` non-streaming response transform in `internal/proxy/transform/responses_test.go`: (1) text message output → Anthropic content with type "text", (2) status "completed"→stop_reason "end_turn", (3) status "incomplete"→stop_reason "max_tokens", (4) usage field mapping (input_tokens/output_tokens), (5) missing usage → zero defaults, (6) empty output array → empty content
- [x] T007 [US1] Implement `ResponsesAPIToAnthropic(body []byte) ([]byte, error)` in `internal/proxy/transform/responses.go`: extract text from output[].content[].output_text, map status→stop_reason, map usage, build Anthropic response JSON (all T006 tests pass)
- [x] T008 [US1] Write E2E retry tests in `internal/proxy/server_test.go` with 3 subtests: (1) `retry_success`: mock server returns 500 "input is required" on `/chat/completions`, 200 Responses API JSON on `/responses` → verify proxy returns 200 with Anthropic format response, (2) `no_retry_on_other_errors`: mock returns 500 "server error" on `/chat/completions` → verify NO request to `/responses`, normal failover, (3) `retry_failure_reports_responses_api_error`: mock returns 500 "input is required" on `/chat/completions` AND 401 on `/responses` → verify proxy reports 401 error from Responses API attempt

### Implementation (GREEN)

- [x] T009 [US1] Implement retry logic in `internal/proxy/server.go`: in `tryProviders` 500 handler, before `isRequestRelatedError` check, call `isResponsesAPIRequired(errBody)`. If true: (1) log retry at debug level per FR-008, (2) call new `retryWithResponsesAPI` method that transforms body via `ChatCompletionsToResponsesAPI`, builds request to `/v1/responses` path (with dedup), sends request, (3) on success call `copyResponseFromResponsesAPI` (non-streaming path: read body, call `ResponsesAPIToAnthropic`, write Anthropic response), (4) on failure return false to continue failover with the Responses API error
- [x] T010 [US1] Verify all US1 tests pass with `go test ./internal/proxy/... -v -run "TestResponsesAPI|TestRetry|TestIsResponsesAPI"` and zero regressions with `go test ./internal/proxy/...`

**Checkpoint**: US1 complete — retry mechanism works for non-streaming requests. Chat Completions providers unchanged.

---

## Phase 4: User Story 2 — Streaming Responses via Responses API (Priority: P1)

**Goal**: Transform Responses API SSE events to Anthropic SSE format for streaming requests

**Independent Test**: Mock server returns streaming Responses API SSE events on `/responses` → proxy transforms to Anthropic SSE events (message_start, content_block_delta, message_stop)

### Tests (RED) ⚠️

- [x] T011 [US2] Write test for `transformResponsesAPIToAnthropic` SSE stream transform in `internal/proxy/transform/stream_test.go`: (a) text streaming: input is Responses API SSE events (response.created, response.output_item.added, response.output_text.delta×3, response.output_item.done, response.completed), verify output contains Anthropic SSE events (message_start with model/id, content_block_start with type "text", content_block_delta with text_delta×3, content_block_stop, message_delta with stop_reason and usage, message_stop); (b) tool call streaming: input includes response.output_item.added (type function_call), response.function_call_arguments.delta×2, response.output_item.done → verify output contains content_block_start (type tool_use), content_block_delta (input_json_delta×2), content_block_stop
- [x] T012 [US2] Write E2E streaming retry test in `internal/proxy/server_test.go`: mock server returns 500 "input is required" on `/chat/completions`, returns `text/event-stream` with Responses API SSE events on `/responses` → verify proxy returns SSE stream with Anthropic-format events to client

### Implementation (GREEN)

- [x] T013 [US2] Implement `transformResponsesAPIToAnthropic` in `internal/proxy/transform/stream.go`: parse `event:` + `data:` lines (Responses API format), convert response.created → message_start, response.output_text.delta → content_block_delta (text_delta), response.function_call_arguments.delta → content_block_delta (input_json_delta), response.output_item.added → content_block_start, response.output_item.done → content_block_stop, response.completed → message_delta + message_stop (with usage from completed event)
- [x] T014 [US2] Wire streaming in `copyResponseFromResponsesAPI` in `internal/proxy/server.go`: detect `text/event-stream` Content-Type, create `StreamTransformer` with `ProviderFormat: "openai-responses"`, pipe through `transformResponsesAPIToAnthropic`, flush chunks to client (all T012 tests pass)
- [x] T015 [US2] Verify all US2 tests pass with `go test ./internal/proxy/... -v -run "TestResponsesAPIStream|TestStreamRetry"` and zero regressions

**Checkpoint**: US2 complete — streaming works through Responses API providers. Existing Chat Completions streaming unchanged.

---

## Phase 5: User Story 3 — Tool Call Response Transform (Priority: P2)

**Goal**: Correctly transform Responses API function_call output items to Anthropic tool_use content blocks

**Independent Test**: Mock server returns Responses API response with function_call output → proxy transforms to Anthropic response with tool_use content block

### Tests (RED) ⚠️

- [x] T016 [P] [US3] Write tests for function_call → tool_use mapping in `internal/proxy/transform/responses_test.go`: (1) single function_call with call_id/name/arguments → Anthropic tool_use with id/name/input (parsed JSON), (2) mixed output: message + function_call → content array with text + tool_use, (3) function_call with status "completed" → stop_reason "tool_use", (4) malformed arguments JSON string → input as empty object
- [x] T017 [P] [US3] Write E2E test with tool call in `internal/proxy/server_test.go`: mock returns 500 "input is required" on `/chat/completions`, returns Responses API JSON with function_call output on `/responses` → verify proxy returns Anthropic response with tool_use content block

### Implementation (GREEN)

- [x] T018 [US3] Add function_call handling to `ResponsesAPIToAnthropic` in `internal/proxy/transform/responses.go`: iterate output array, for type "function_call" create Anthropic tool_use block with id=call_id, name=name, input=JSON.parse(arguments), set stop_reason to "tool_use" if any function_call present (all T016 tests pass)
- [x] T019 [US3] Verify all US3 tests pass with `go test ./internal/proxy/... -v -run "TestResponsesAPIToolCall|TestToolCall"` and zero regressions

**Checkpoint**: All user stories complete — text, streaming, and tool call responses work through Responses API providers.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Coverage verification and final validation

- [x] T020 Verify test coverage meets CI thresholds: run `go test -cover ./internal/proxy/` (≥80%) and `go test -cover ./internal/proxy/transform/` (≥80%), add targeted tests if below threshold
- [x] T021 Run full test suite `go test ./...` and verify zero regressions across all packages

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2 — core retry logic
- **US2 (Phase 4)**: Depends on Phase 3 — needs retry mechanism to test streaming path
- **US3 (Phase 5)**: Depends on Phase 3 — needs retry mechanism and ResponsesAPIToAnthropic base
- **Polish (Phase 6)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: Depends on Foundational (Phase 2) only — MVP deliverable
- **US2 (P1)**: Depends on US1 (needs retry + copyResponseFromResponsesAPI wired)
- **US3 (P2)**: Depends on US1 (extends ResponsesAPIToAnthropic with tool_use handling)

### Within Each User Story

- Tests MUST be written and verified to FAIL before implementation (TDD)
- Transform functions before E2E wiring
- Core implementation before edge cases
- Commit after each task or logical group

### Parallel Opportunities

- T002 [P] and T003 [P] can run in parallel (different files: server_test.go vs responses_test.go)
- T016 [P] and T017 [P] can run in parallel (different files: responses_test.go vs server_test.go)

---

## Parallel Example: Phase 2

```bash
# Launch foundational tests in parallel (different files):
Task T002: "Tests for isResponsesAPIRequired in server_test.go"
Task T003: "Tests for ChatCompletionsToResponsesAPI in responses_test.go"
```

## Parallel Example: US3

```bash
# Launch US3 tests in parallel (different files):
Task T016: "Tests for function_call→tool_use in responses_test.go"
Task T017: "E2E test with tool call in server_test.go"
```

---

## Implementation Strategy

### MVP First (US1 Only — Phases 1-3)

1. Complete Phase 1: Setup (baseline verification)
2. Complete Phase 2: Foundational (request transform + error detection)
3. Complete Phase 3: US1 (retry logic + non-streaming response transform)
4. **STOP and VALIDATE**: Non-streaming retry works, Chat Completions providers unchanged
5. Can deploy — Responses API providers work for non-streaming requests

### Incremental Delivery

1. Setup + Foundational → Transform functions ready
2. Add US1 → Non-streaming retry works → Deploy (MVP!)
3. Add US2 → Streaming works → Deploy (full P1 coverage)
4. Add US3 → Tool calls work → Deploy (complete feature)
5. Polish → Coverage verified, zero regressions confirmed

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- Constitution Principle I (TDD) is NON-NEGOTIABLE: tests must fail before implementation
- No config schema changes needed — no version bump
- `isResponsesAPIRequired` uses substring match per research.md Decision 1
- Retry is internal to provider attempt per research.md Decision 2 — does NOT consume failover slot
- Two-step transform (Anthropic→CC→Responses API) per research.md Decision 3
- Commit after each task per constitution Principle IV
