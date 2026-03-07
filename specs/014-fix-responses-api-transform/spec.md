# Feature Specification: Fix OpenAI Responses API Transform

**Feature Branch**: `014-fix-responses-api-transform`
**Created**: 2026-03-06
**Status**: Draft
**Input**: User description: "Fix proxy transform for OpenAI providers that expect Responses API format (input field) instead of Chat Completions (messages field)"

## Context

The proxy's Anthropic→OpenAI transform currently only generates **Chat Completions** format (`POST /v1/chat/completions` with `messages` field). Some OpenAI-compatible providers (e.g., `code.b886.top`) only support the newer **Responses API** format (`POST /v1/responses` with `input` field), causing `"input is required"` errors.

**Evidence from production logs**:
```
[cctq-codex] model mapping: claude-sonnet-4-6 → gpt-5.4          ✅
[cctq-codex] transformed request: anthropic → openai              ✅
[cctq-codex] path transform: /v1/messages → /v1/chat/completions  ✅
[cctq-codex] path dedup: /v1/chat/completions → /chat/completions ✅
[cctq-codex] 500 "input is required"                              ❌
```

The path and model transforms work (fixed in #15), but the upstream rejects the Chat Completions body format because it expects the Responses API `input` field.

## User Scenarios & Testing

### User Story 1 - Responses API Provider Works with Claude Code (Priority: P1)

A user configures an OpenAI-compatible provider that only supports the Responses API format. When they use Claude Code (which sends Anthropic-format requests), the proxy should detect that Chat Completions failed and automatically retry using the Responses API format, so the request succeeds without user intervention.

**Why this priority**: This is the core bug — users cannot use Claude Code with Responses API providers at all. No workaround exists.

**Independent Test**: Send an Anthropic-format request through a provider whose Chat Completions endpoint returns "input is required", verify the proxy retries with Responses API format and succeeds.

**Acceptance Scenarios**:

1. **Given** an OpenAI provider whose `/v1/chat/completions` returns `"input is required"`, **When** Claude Code sends an Anthropic-format request, **Then** the proxy retries using `/v1/responses` with `input` field and the request succeeds
2. **Given** an OpenAI provider that supports Chat Completions normally, **When** Claude Code sends an Anthropic-format request, **Then** the proxy uses Chat Completions format as before (no change in behavior)
3. **Given** an OpenAI provider returning "input is required" on Chat Completions, **When** the retry with Responses API also fails (e.g., auth error), **Then** the proxy reports the Responses API error (not the original Chat Completions error)

---

### User Story 2 - Streaming Responses via Responses API (Priority: P1)

When a streaming request is sent through a Responses API provider, the proxy should correctly transform SSE events between the Responses API streaming format and the Anthropic streaming format.

**Why this priority**: Streaming is essential for Claude Code usability — without it, users see no output until the entire response completes.

**Independent Test**: Send a streaming Anthropic request through a Responses API provider, verify SSE events are correctly transformed back to Anthropic format.

**Acceptance Scenarios**:

1. **Given** a Responses API provider with streaming enabled, **When** Claude Code sends a streaming request, **Then** the proxy transforms Responses API SSE events to Anthropic SSE format and streams them back
2. **Given** a Chat Completions provider with streaming enabled, **When** Claude Code sends a streaming request, **Then** existing streaming behavior is unchanged

---

### User Story 3 - Response Transformation Back to Anthropic Format (Priority: P2)

When a non-streaming response is returned from a Responses API provider, the proxy should correctly transform the Responses API response format back to Anthropic format so Claude Code can parse it.

**Why this priority**: Non-streaming responses must also work correctly, but streaming is the more common code path for Claude Code.

**Independent Test**: Send a non-streaming Anthropic request through a Responses API provider, verify the response is correctly transformed to Anthropic format with all expected fields.

**Acceptance Scenarios**:

1. **Given** a non-streaming Responses API response with `output` field, **When** the proxy transforms it, **Then** it returns a valid Anthropic response with `id`, `type`, `role`, `content`, `model`, `stop_reason`, `usage` fields
2. **Given** a Responses API response with tool call output, **When** the proxy transforms it, **Then** tool use content blocks are correctly mapped to Anthropic tool_use format

---

### Edge Cases

- What happens when the provider returns a non-"input is required" error on Chat Completions (e.g., auth failure)? The proxy should NOT retry with Responses API — it should failover to the next provider as usual.
- What happens when both Chat Completions and Responses API endpoints fail? The proxy should report the last meaningful error.
- What happens when the Responses API response format is missing expected fields (e.g., no `usage`)? The proxy should handle gracefully with zero/empty defaults.
- What happens with tool use requests? The proxy must transform Anthropic tool definitions to OpenAI function format for the Responses API.

## Requirements

### Functional Requirements

- **FR-001**: The proxy MUST detect when a Chat Completions request fails with "input is required" and automatically retry using the Responses API format
- **FR-002**: The Responses API retry MUST transform the request body from Chat Completions format (`messages`) to Responses API format (`input`)
- **FR-003**: The Responses API retry MUST use the path `/v1/responses` instead of `/v1/chat/completions`
- **FR-004**: The proxy MUST transform non-streaming Responses API responses back to Anthropic format
- **FR-005**: The proxy MUST transform streaming Responses API SSE events back to Anthropic SSE format
- **FR-006**: The proxy MUST NOT retry with Responses API for errors other than "input is required" (e.g., auth failures, rate limits, server errors)
- **FR-007**: The proxy MUST preserve backward compatibility — Chat Completions providers MUST continue to work unchanged
- **FR-008**: The proxy MUST log the Responses API retry at debug level for troubleshooting

### Key Entities

- **Request Format**: The body format sent to the provider — either Chat Completions (`messages` field) or Responses API (`input` field)
- **Response Format**: The body format returned by the provider — either Chat Completions (`choices` array) or Responses API (`output` array)
- **Retry Decision**: The logic that determines whether to retry with Responses API based on the error response from Chat Completions

## Success Criteria

### Measurable Outcomes

- **SC-001**: Claude Code users can successfully use providers that only support the Responses API format (zero "input is required" errors reaching the user)
- **SC-002**: Existing Chat Completions providers continue to work with identical behavior (zero regressions)
- **SC-003**: Both streaming and non-streaming requests work through Responses API providers
- **SC-004**: The retry adds no more than one additional round-trip per request (only triggered on "input is required" error)

## Assumptions

- The "input is required" error message is a reliable indicator that the provider expects Responses API format. This is consistent with observed behavior from `code.b886.top` and similar Chinese OpenAI-compatible proxy services.
- The Responses API format follows the OpenAI Responses API specification: `input` field for messages, `output` array for response content, and OpenAI-style SSE events for streaming.
- The retry mechanism is per-request, not cached — each request to a Chat Completions endpoint that fails will trigger a Responses API retry. Future optimization could cache the provider's preferred format.
