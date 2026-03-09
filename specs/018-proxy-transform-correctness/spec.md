# Feature Specification: Proxy Transform Layer Correctness

**Feature Branch**: `018-proxy-transform-correctness`
**Created**: 2026-03-09
**Status**: Draft
**Input**: User description: "完成该文档中对于 P0、P1"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Protocol-Aware Request Transformation (Priority: P0)

When a developer uses GoZen proxy to forward requests from OpenAI Chat Completions clients to Anthropic providers, or from OpenAI Responses API clients to Anthropic providers, the proxy must correctly identify which OpenAI protocol variant the client is using and transform requests/responses according to that specific protocol's schema.

**Why this priority**: Currently both OpenAI protocols are treated as a single "openai" format, causing response shape mismatches, incorrect SSE event types, and protocol violations. This is a P0 correctness issue that breaks basic proxy functionality.

**Independent Test**: Can be fully tested by sending Chat Completions requests and Responses API requests through the proxy to an Anthropic provider, then validating that responses match the expected protocol schema for each client type.

**Acceptance Scenarios**:

1. **Given** a client sends a request to `/chat/completions`, **When** the proxy forwards to an Anthropic provider, **Then** the response is transformed to OpenAI Chat Completions format with correct field names and structure
2. **Given** a client sends a request to `/responses`, **When** the proxy forwards to an Anthropic provider, **Then** the response is transformed to OpenAI Responses API format with correct event types and field names
3. **Given** a client sends a streaming request to `/chat/completions`, **When** the proxy forwards to an Anthropic provider, **Then** SSE events use Chat Completions event types (`chat.completion.chunk`)
4. **Given** a client sends a streaming request to `/responses`, **When** the proxy forwards to an Anthropic provider, **Then** SSE events use Responses API event types (`response.delta`, `response.completed`)

---

### User Story 2 - Complete Tool Call Transformation (Priority: P0)

When a developer uses GoZen proxy to forward tool-enabled requests between OpenAI and Anthropic formats, the proxy must correctly transform tool definitions, tool invocations, and tool results in both directions, including streaming scenarios.

**Why this priority**: Tool calls are a core feature of modern LLM APIs. Incomplete transformation breaks agent workflows and function calling use cases.

**Independent Test**: Can be fully tested by sending requests with tool definitions through the proxy, verifying tool invocations are correctly transformed, and checking that tool results flow back correctly in both streaming and non-streaming modes.

**Acceptance Scenarios**:

1. **Given** an OpenAI client sends a request with `tools` array, **When** the proxy forwards to Anthropic, **Then** tools are transformed to Anthropic `tools` schema with correct `input_schema` structure
2. **Given** an Anthropic provider returns a response with `tool_use` content blocks, **When** the proxy transforms to OpenAI format, **Then** response contains `tool_calls` array with correct `function` structure
3. **Given** an OpenAI streaming client receives tool call deltas, **When** the proxy transforms from Anthropic streaming, **Then** `tool_calls` deltas are correctly assembled with `function.arguments` incremental JSON
4. **Given** a Responses API client sends a request with function definitions, **When** the proxy forwards to Anthropic, **Then** functions are transformed to Anthropic tools format
5. **Given** an Anthropic provider streams `input_json_delta` events, **When** the proxy transforms to Responses API format, **Then** `function_call_arguments.delta` events are emitted correctly

---

### User Story 3 - Safe SSE Stream Error Handling (Priority: P0)

When a developer uses GoZen proxy for streaming requests and the upstream provider's SSE stream is truncated, malformed, or encounters a network error, the proxy must detect the error condition and propagate it to the client rather than fabricating a successful completion event.

**Why this priority**: Silently converting broken streams into fake successful completions causes clients to believe they received complete responses when they didn't, leading to data loss and incorrect application behavior.

**Independent Test**: Can be fully tested by simulating upstream stream failures (connection drops, malformed SSE, truncated responses) and verifying that the proxy propagates errors to clients instead of emitting completion events.

**Acceptance Scenarios**:

1. **Given** an upstream Anthropic stream is truncated mid-response, **When** the proxy's SSE scanner detects the error, **Then** the proxy emits an error event to the client instead of `message_stop`
2. **Given** an upstream stream contains malformed SSE data, **When** the proxy's scanner fails to parse, **Then** the proxy logs the error and closes the client stream with an error indication
3. **Given** an upstream stream connection drops during transformation, **When** the scanner returns an error, **Then** the proxy checks `scanner.Err()` and propagates the error downstream
4. **Given** a streaming response completes successfully, **When** `scanner.Err()` returns nil, **Then** the proxy emits the appropriate completion event (`message_stop`, `response.completed`, or `[DONE]`)

---

### User Story 4 - Transform Error Classification (Priority: P1)

When a developer uses GoZen proxy and a request transformation fails due to a local proxy bug (invalid JSON parsing, schema mismatch, etc.), the proxy must classify this as a transform/proxy error rather than marking the target provider as unhealthy.

**Why this priority**: Incorrectly marking providers as unhealthy due to local transform bugs poisons failover decisions and causes unnecessary provider rotation. This is a P1 stability issue that affects multi-provider reliability.

**Independent Test**: Can be fully tested by injecting transform failures (malformed input, unsupported schema) and verifying that providers are not marked unhealthy and failover behavior is correct.

**Acceptance Scenarios**:

1. **Given** a request body fails to parse during transformation, **When** the transform error occurs, **Then** the proxy returns an error to the client without marking the provider unhealthy
2. **Given** a response transformation fails due to unexpected schema, **When** the transform error occurs, **Then** the proxy logs the error as a transform failure and does not trigger provider backoff
3. **Given** a transform error occurs and multiple providers are configured, **When** failover is triggered, **Then** the same provider is retried (since it wasn't the provider's fault)
4. **Given** a provider returns a valid error response (401, 429, 5xx), **When** the proxy classifies the error, **Then** the provider is marked unhealthy according to existing error classification rules

---

### User Story 5 - Remove Transform Hot Path Logging (Priority: P1)

When a developer uses GoZen proxy in production, the proxy must not perform file I/O or log full request/response bodies on every request in the transform hot path, as this degrades performance and creates unnecessary disk usage.

**Why this priority**: Current debug logging writes full request/response bodies to disk on every transform, causing performance degradation and potential disk space issues in high-throughput scenarios.

**Independent Test**: Can be fully tested by running the proxy under load and verifying that no transform debug logs are written by default, and that performance is not impacted by logging overhead.

**Acceptance Scenarios**:

1. **Given** the proxy is running in production mode, **When** requests are transformed, **Then** no request/response bodies are logged to disk
2. **Given** the proxy is running with debug logging disabled (default), **When** transform operations occur, **Then** no file I/O is performed in the transform hot path
3. **Given** a developer enables debug logging via environment variable or config flag, **When** requests are transformed, **Then** debug logs are written to the configured location
4. **Given** the proxy handles 1000 requests/second, **When** debug logging is disabled, **Then** transform performance is not degraded by logging overhead

---

### Edge Cases

- **Mixed protocol fields**: When a client sends a request with both Chat Completions and Responses API fields mixed together, path-based detection takes precedence and incompatible fields from other protocols are ignored
- How does the system handle tool calls with extremely large `arguments` JSON (>100KB) in streaming mode?
- What happens when an upstream provider returns a valid SSE stream but with unexpected event types not in the transform mapping?
- How does the proxy handle partial tool call deltas that span multiple SSE events?
- What happens when a transform error occurs mid-stream after some events have already been sent to the client?
- How does the system handle requests to `/v1/messages` (native Anthropic) when the provider is also Anthropic (no transform needed)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST distinguish between three protocol formats: `anthropic-messages`, `openai-chat`, and `openai-responses`
- **FR-002**: System MUST detect client protocol format based on request path: `/chat/completions` → `openai-chat`, `/responses` → `openai-responses`, `/v1/messages` → `anthropic-messages`. Path-based detection takes precedence; incompatible fields from other protocols are ignored.
- **FR-003**: System MUST select the correct transformer implementation based on both client format and provider format
- **FR-004**: System MUST transform OpenAI Chat Completions `tools` array to Anthropic `tools` schema with correct `input_schema` structure
- **FR-005**: System MUST transform Anthropic `tool_use` content blocks to OpenAI `tool_calls` array with correct `function` structure
- **FR-006**: System MUST transform OpenAI Responses API function definitions to Anthropic tools format
- **FR-007**: System MUST transform Anthropic `input_json_delta` streaming events to OpenAI `function_call_arguments.delta` events
- **FR-008**: System MUST transform OpenAI streaming tool call deltas to Anthropic streaming tool use events
- **FR-009**: System MUST check `scanner.Err()` after every SSE stream scanning loop completes
- **FR-010**: System MUST propagate upstream stream errors to the client using protocol-native error events (OpenAI `error` event, Anthropic `error` event) with standardized error codes instead of emitting completion events when `scanner.Err()` is non-nil
- **FR-011**: System MUST classify request transformation failures as transform errors, not provider errors
- **FR-012**: System MUST classify response transformation failures as transform errors, not provider errors
- **FR-013**: System MUST NOT mark providers as unhealthy when transform errors occur
- **FR-014**: System MUST return HTTP 500 errors to clients when transform errors occur, with error messages indicating transform failure
- **FR-015**: System MUST NOT perform file I/O in the transform hot path by default
- **FR-016**: System MUST NOT log full request/response bodies in the transform hot path by default
- **FR-017**: System MUST support optional debug logging via environment variable or config flag
- **FR-018**: System MUST preserve existing transform test coverage while implementing changes

### Key Entities

- **Protocol Format**: Identifier for API protocol variant (`anthropic-messages`, `openai-chat`, `openai-responses`)
- **Transformer**: Component that converts requests/responses between two protocol formats
- **Transform Error**: Error that occurs during local request/response transformation, distinct from provider errors
- **SSE Scanner**: Component that reads and parses Server-Sent Events from upstream streams
- **Tool Definition**: Schema describing available functions/tools in either OpenAI or Anthropic format
- **Tool Invocation**: Request from LLM to call a specific tool with arguments
- **Tool Delta**: Incremental streaming update to tool invocation arguments

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Proxy correctly transforms 100% of Chat Completions requests to Anthropic format and responses back to Chat Completions format
- **SC-002**: Proxy correctly transforms 100% of Responses API requests to Anthropic format and responses back to Responses API format
- **SC-003**: Proxy correctly transforms tool definitions and invocations in both directions with 100% schema correctness
- **SC-004**: Proxy detects and propagates 100% of upstream stream errors instead of fabricating completion events
- **SC-005**: Transform errors do not cause providers to be marked unhealthy (0% false positive provider health failures due to transform bugs)
- **SC-006**: Proxy handles 1000 requests/second with debug logging disabled without performance degradation from logging overhead
- **SC-007**: All existing transform tests continue to pass after implementation
- **SC-008**: New tests cover all P0/P1 scenarios with 100% pass rate

## Assumptions

- The existing `internal/proxy/transform/` package structure will be preserved, with new transformer types added
- The existing `detectClientFormat()` function in `profile_proxy.go` will be extended to return the new protocol format identifiers
- The existing `GetTransformer()` function will be updated to return protocol-specific transformers
- Transform error classification will be implemented by adding error type checking in `server.go` error handling paths
- Debug logging will be controlled by a new environment variable `GOZEN_TRANSFORM_DEBUG` or config field
- The existing test suite structure will be preserved, with new test cases added for P0/P1 scenarios
- No changes to the config schema version are required (this is an internal implementation change)

## Dependencies

- Existing `internal/proxy/transform/` package
- Existing `internal/proxy/server.go` request/response handling
- Existing `internal/proxy/profile_proxy.go` format detection
- Existing test infrastructure in `*_test.go` files

## Clarifications

### Session 2026-03-09

- Q: When a stream error is detected, what specific error event format should be sent to the client for each protocol? → A: Use protocol-native error events with standardized error codes (e.g., OpenAI `error` event, Anthropic `error` event)
- Q: How should the proxy handle requests that contain fields from multiple OpenAI protocol variants? → A: Path-based detection takes precedence, ignore incompatible fields from other protocols

## Out of Scope

- P2 item: Connecting profile load balancing strategy to provider selection (deferred to future work)
- Changes to config schema or user-facing configuration
- Performance optimization beyond removing debug logging overhead
- Support for additional protocol formats beyond the three specified
- Automatic retry logic for transform errors (errors are returned to client)
