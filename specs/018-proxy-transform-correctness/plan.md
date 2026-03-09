# Implementation Plan: Proxy Transform Layer Correctness

**Branch**: `018-proxy-transform-correctness` | **Date**: 2026-03-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/018-proxy-transform-correctness/spec.md`

## Summary

Fix the GoZen proxy transform layer to correctly distinguish between three protocol formats (`anthropic-messages`, `openai-chat`, `openai-responses`), complete bidirectional tool call transformation including streaming, make SSE stream error handling safe (no fake completions), classify transform errors separately from provider errors, and remove debug file I/O from the hot path.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: `bufio`, `encoding/json`, `io`, `net/http` (stdlib only)
**Storage**: N/A (no config schema changes)
**Testing**: `go test ./...`, table-driven tests in existing `*_test.go` files
**Target Platform**: Linux/macOS daemon process
**Project Type**: CLI tool / reverse proxy daemon
**Performance Goals**: Zero file I/O in transform hot path; existing 1000 req/s target unaffected
**Constraints**: Must not break existing transform test coverage (≥80% for `internal/proxy/transform`)
**Scale/Scope**: Affects every proxied request; changes are internal to `internal/proxy/`

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. TDD | ✅ Required | Write failing tests first for each sub-task |
| II. YAGNI | ✅ Compliant | No new abstractions; minimal interface changes |
| III. Config Migration | ✅ N/A | No config schema changes |
| IV. Branch Protection | ✅ Required | PR required; atomic commits per task |
| V. Minimal Artifacts | ✅ Compliant | No summary docs; specs in `.specify/` |
| VI. Coverage ≥80% | ✅ Required | `internal/proxy/transform` must stay ≥80% |

## Project Structure

### Documentation (this feature)

```text
specs/018-proxy-transform-correctness/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── contracts/
│   └── transform-api.md # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (affected files)

```text
internal/proxy/
├── transform/
│   ├── transform.go         # Add FormatXxx constants; update NeedsTransform
│   ├── anthropic.go         # Remove init(), debugLogger, all debugLogger calls
│   ├── openai.go            # Update TransformRequest/Response for openai-chat vs openai-responses
│   ├── responses.go         # Update as needed for openai-responses client format
│   ├── stream.go            # Add scanner.Err() checks; protocol-native error events; route openai-chat vs openai-responses
│   ├── anthropic_test.go    # Add tool streaming delta tests
│   ├── openai_test.go       # Add openai-chat vs openai-responses shape tests
│   ├── stream_test.go       # Add scanner error propagation tests
│   └── responses_test.go    # Update for new format constants
├── server.go                # Add TransformError type; update forwardRequest; update tryProviders
├── profile_proxy.go         # Update detectClientFormat return values
└── server_test.go           # Add transform error classification tests
```

## Implementation Phases

### Phase A: Protocol Format Constants + detectClientFormat (P0-1 foundation)

**Goal**: Establish the three format identifiers and update format detection. All downstream changes depend on this.

**Tasks**:
1. Add `FormatAnthropicMessages`, `FormatOpenAIChat`, `FormatOpenAIResponses` constants to `transform.go`
2. Update `NeedsTransform` to handle new format strings
3. Update `detectClientFormat` in `profile_proxy.go` to return new constants
4. Update `StreamTransformer` routing in `stream.go` to branch on `openai-chat` vs `openai-responses`

**Tests first**: Update existing format detection tests; add table rows for new constants.

---

### Phase B: Remove Debug Logging (P1-5)

**Goal**: Remove all file I/O from transform hot path.

**Tasks**:
1. Delete `init()` function and `debugLogger` var from `anthropic.go`
2. Remove all `debugLogger.Printf(...)` call sites
3. Verify `go build ./...` passes

**Tests first**: Verify no test depends on debug log output.

---

### Phase C: SSE Scanner Error Handling (P0-3)

**Goal**: Check `scanner.Err()` after every loop; emit protocol-native error events.

**Tasks**:
1. Add `scanner.Err()` check after loop in `transformAnthropicToOpenAI`
2. Add `scanner.Err()` check after loop in `transformOpenAIToAnthropic`
3. Add `scanner.Err()` check after loop in `transformResponsesAPIToAnthropic`
4. Implement `writeStreamError(pw, clientFormat, err)` helper that emits protocol-native error event
5. Ensure completion events are only emitted when `scanner.Err() == nil`

**Tests first**: Add table-driven tests simulating truncated/errored readers for each streaming path.

---

### Phase D: Complete Tool Call Transformation (P0-2)

**Goal**: Correct bidirectional tool transformation including streaming deltas.

**Tasks**:
1. Verify/fix non-streaming tool request: OpenAI Chat `tools` → Anthropic `tools` with `input_schema`
2. Verify/fix non-streaming tool response: Anthropic `tool_use` → OpenAI Chat `tool_calls`
3. Fix streaming: Anthropic `content_block_start(tool_use)` → OpenAI Chat `tool_calls[].id` + `function.name` delta
4. Fix streaming: Anthropic `input_json_delta` → OpenAI Chat `tool_calls[].function.arguments` delta
5. Fix streaming: Anthropic `content_block_start(tool_use)` → OpenAI Responses `response.output_item.added`
6. Fix streaming: Anthropic `input_json_delta` → OpenAI Responses `response.function_call_arguments.delta`

**Tests first**: Add table-driven tests for each tool transformation direction and streaming scenario.

---

### Phase E: Transform Error Classification (P1-4)

**Goal**: Transform failures return 500 to client without marking provider unhealthy.

**Tasks**:
1. Define `TransformError` type in `internal/proxy/server.go`
2. Update `forwardRequest` to return `&TransformError{Op: "request", Err: err}` on transform failure
3. Update `tryProviders` to detect `TransformError` via `errors.As` and return 500 without health impact
4. Handle response transform errors similarly in `copyResponse`

**Tests first**: Add integration test verifying provider health unchanged after transform error.

---

## Complexity Tracking

No constitution violations. All changes are minimal modifications to existing files.
