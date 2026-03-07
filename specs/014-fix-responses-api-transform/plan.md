# Implementation Plan: Fix OpenAI Responses API Transform

**Branch**: `014-fix-responses-api-transform` | **Date**: 2026-03-07 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/014-fix-responses-api-transform/spec.md`

## Summary

The proxy's Anthropic→OpenAI transform produces Chat Completions format (`messages` field, `/v1/chat/completions`), but some providers only support the Responses API (`input` field, `/v1/responses`). When a Chat Completions request fails with `"input is required"`, the proxy should automatically retry using Responses API format. This requires: (1) detecting the error, (2) transforming the request body and path, (3) transforming the response (both streaming and non-streaming) back to Anthropic format.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: `net/http`, `encoding/json`, `bufio`, `io` (stdlib only)
**Storage**: N/A (no config schema changes, no version bump)
**Testing**: `go test ./internal/proxy/... -cover` (TDD per constitution)
**Target Platform**: Linux/macOS (CLI tool with reverse proxy)
**Project Type**: CLI / reverse proxy
**Performance Goals**: ≤1 additional round-trip per request (retry only on "input is required")
**Constraints**: No config schema changes; backward compatibility with Chat Completions providers
**Scale/Scope**: Single retry per provider attempt; no caching of provider format preference

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. TDD (NON-NEGOTIABLE) | ✅ PASS | Tests first for all new transform functions and retry logic |
| II. Simplicity & YAGNI | ✅ PASS | Minimal retry in existing flow; no format caching, no config options |
| III. Config Migration | ✅ PASS | No config schema changes needed |
| IV. Branch Protection | ✅ PASS | Feature branch with individual commits |
| V. Minimal Artifacts | ✅ PASS | No summary docs; plan + code only |
| VI. Test Coverage (NON-NEGOTIABLE) | ✅ PASS | Must maintain proxy ≥80%, transform ≥80% |

## Project Structure

### Documentation (this feature)

```text
specs/014-fix-responses-api-transform/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0: OpenAI Responses API research
├── data-model.md        # Phase 1: Request/response format mappings
├── checklists/
│   └── requirements.md  # Specification quality checklist
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
internal/proxy/
├── server.go                    # MODIFY: Add "input is required" detection + retry logic in tryProviders
├── server_test.go               # MODIFY: Add retry E2E tests
└── transform/
    ├── responses.go             # NEW: Chat Completions ↔ Responses API body + response transforms
    ├── responses_test.go        # NEW: Tests for Responses API transforms
    ├── stream.go                # MODIFY: Add Responses API SSE → Anthropic SSE stream path
    └── stream_test.go           # MODIFY: Add tests for Responses API streaming
```

**Structure Decision**: All changes fit within the existing `internal/proxy/` and `internal/proxy/transform/` packages. One new file (`responses.go`) for the Chat Completions→Responses API body transformation, which is logically distinct from the existing OpenAI Chat Completions transformer.

## Architecture

### Retry Flow

```
Anthropic Client
    │
    ▼
ProfileProxy.ServeHTTP
    │ detects format=anthropic
    ▼
ProxyServer.tryProviders
    │
    ├─ forwardRequest(provider, body, requestFormat)
    │   │ transforms: Anthropic → Chat Completions body
    │   │ path: /v1/messages → /v1/chat/completions
    │   │ sends to upstream
    │   ▼
    │  [500 "input is required"]
    │
    ├─ isResponsesAPIRequired(errBody) → true
    │
    ├─ retryWithResponsesAPI(r, provider, body, modelOverride, requestFormat)
    │   │ transforms: Chat Completions body → Responses API body
    │   │   - messages → input
    │   │   - max_completion_tokens → max_output_tokens
    │   │   - flatten tools (unwrap function wrapper)
    │   │ path: /v1/responses (with dedup)
    │   │ sends to upstream
    │   ▼
    │  [200 OK — Responses API format response]
    │
    ├─ copyResponseFromResponsesAPI(w, resp, provider, requestFormat)
    │   │ non-streaming: ResponsesAPI → Anthropic JSON
    │   │ streaming: ResponsesAPI SSE → Anthropic SSE
    │   ▼
    │  [Anthropic-format response to client]
    ▼
  SUCCESS
```

### Key Design Decisions

1. **Retry location**: Inside `tryProviders` at the 500 handler, before `isRequestRelatedError` check. The retry is internal to a single provider attempt — it does NOT consume a failover slot.

2. **Body transformation approach**: Two-step transform. `forwardRequest` already transformed Anthropic→Chat Completions. The retry takes the Chat Completions body and converts to Responses API. This avoids duplicating the Anthropic→OpenAI logic.

3. **Response transformation**: Handled in a new `copyResponseFromResponsesAPI` function that detects streaming vs non-streaming and applies the appropriate transform:
   - Non-streaming: Parse Responses API JSON `output` array, extract text/tool_use content, build Anthropic response
   - Streaming: Pipe through a new `transformResponsesAPIToAnthropic` stream transformer

4. **Error detection**: Substring match on `"input is required"` in the response body (any 4xx/5xx status). This matches the observed error format from production.

5. **No config changes**: The retry is automatic and transparent. No new config fields, no version bump.

### Field Mappings

#### Chat Completions → Responses API (Request)

| Chat Completions | Responses API | Notes |
|-----------------|---------------|-------|
| `messages` | `input` | Array format compatible |
| `model` | `model` | Unchanged |
| `stream` | `stream` | Unchanged |
| `max_completion_tokens` | `max_output_tokens` | Renamed |
| `temperature` | `temperature` | Unchanged |
| `top_p` | `top_p` | Unchanged |
| `tools[].function.{name,desc,params}` | `tools[].{name,desc,params}` | Flattened (remove `function` wrapper) |
| `tool_choice` | `tool_choice` | Unchanged |
| `stop` | `stop` | Keep as-is |
| `n` | — | Remove (not supported) |
| `stream_options` | — | Remove |
| `logprobs` | — | Remove |
| `presence_penalty` | — | Remove |
| `frequency_penalty` | — | Remove |
| `seed` | — | Remove |
| `response_format` | — | Remove |

#### Responses API → Anthropic (Non-streaming Response)

| Responses API | Anthropic | Notes |
|--------------|-----------|-------|
| `id` | `id` | Pass through |
| `model` | `model` | Pass through |
| `output[].type=="message"` → `.content[].text` | `content[].{type:"text", text}` | Extract text |
| `output[].type=="function_call"` | `content[].{type:"tool_use", id, name, input}` | Map tool calls |
| `status=="completed"` | `stop_reason="end_turn"` | Map status |
| `status=="incomplete"` | `stop_reason="max_tokens"` | Map status |
| `usage.input_tokens` | `usage.input_tokens` | Same field name |
| `usage.output_tokens` | `usage.output_tokens` | Same field name |

#### Responses API → Anthropic (Streaming SSE)

| Responses API Event | Anthropic Event | Notes |
|--------------------|-----------------|-------|
| `response.created` | `message_start` | Emit with message metadata |
| `response.output_item.added` (message) | `content_block_start` | Start text content block |
| `response.output_text.delta` | `content_block_delta` (text_delta) | Text token |
| `response.output_item.added` (function_call) | `content_block_start` | Start tool_use block |
| `response.function_call_arguments.delta` | `content_block_delta` (input_json_delta) | Tool args |
| `response.output_item.done` | `content_block_stop` | End content block |
| `response.completed` | `message_delta` + `message_stop` | Final usage + stop |

## Complexity Tracking

No constitution violations. All changes are minimal and focused on the retry mechanism.
