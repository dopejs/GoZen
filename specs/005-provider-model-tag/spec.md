# Feature Specification: Provider & Model Tag in Proxy Responses

**Feature Branch**: `005-provider-model-tag`
**Created**: 2026-03-02
**Status**: Draft
**Input**: User description: "在反向代理模式下，让用户知道实际使用的 provider 和 model。在 response 的 content 前注入 `[provider: xxx, model: xxx]\n`。WebUI 开关控制，默认关闭。支持 Anthropic Messages 和 OpenAI Completions 两种结构。"

## User Scenarios & Testing

### User Story 1 — Non-Streaming Response Tag Injection (Priority: P1)

A developer using `zen` with the daemon proxy wants to know which provider and model actually served their request. When the tag feature is enabled, the first text content block in the response has a provider/model tag prepended, so the developer can see at a glance which provider handled the request.

**Why this priority**: This is the core value — giving users visibility into which provider is being used. Non-streaming responses are simpler to implement and validate, making this the ideal MVP.

**Independent Test**: Enable the tag in settings, send a non-streaming request through the proxy, verify the first text content block starts with `[provider: xxx, model: xxx]\n`.

**Acceptance Scenarios**:

1. **Given** tag is enabled and request format is Anthropic Messages, **When** a non-streaming response is returned with a `content` array containing text blocks, **Then** the first text block's `text` field is prepended with `[provider: <name>, model: <model>]\n`
2. **Given** tag is enabled and request format is OpenAI Chat Completions, **When** a non-streaming response is returned with `choices[0].message.content`, **Then** the content string is prepended with `[provider: <name>, model: <model>]\n`
3. **Given** tag is disabled (default), **When** any response is returned, **Then** the response body is unmodified

---

### User Story 2 — SSE Streaming Response Tag Injection (Priority: P1)

A developer using `zen` with streaming enabled wants the same provider/model visibility. When the tag feature is enabled and the response is SSE-streamed, the tag is injected as the first content delta in the stream, so it appears naturally at the beginning of the output.

**Why this priority**: Most real-world usage (Claude Code, Codex) uses streaming. Without streaming support, the feature would be unusable in the primary use case.

**Independent Test**: Enable the tag, send a streaming request, verify the first content delta event contains the tag text before the actual content begins.

**Acceptance Scenarios**:

1. **Given** tag is enabled and streaming Anthropic response, **When** the first `content_block_delta` event with `type: text_delta` arrives, **Then** the `text` field is prepended with `[provider: <name>, model: <model>]\n`
2. **Given** tag is enabled and streaming OpenAI response, **When** the first `delta.content` chunk arrives in the SSE stream, **Then** the content is prepended with `[provider: <name>, model: <model>]\n`
3. **Given** tag is enabled but the response contains only tool-use content (no text), **Then** no tag is injected (the response is unmodified)

---

### User Story 3 — Web UI Toggle for Tag Feature (Priority: P2)

An administrator wants to enable or disable the provider/model tag from the Web UI settings page, without manually editing the config file.

**Why this priority**: Provides a user-friendly way to control the feature. Lower priority because the config file can be edited manually as a workaround.

**Independent Test**: Open Web UI settings, toggle the "Show provider info in responses" switch, save, verify the setting persists and takes effect on the next proxy request.

**Acceptance Scenarios**:

1. **Given** a user opens the Web UI General Settings page, **When** the page loads, **Then** a toggle for "Show provider info in responses" is visible, defaulting to OFF
2. **Given** a user toggles the setting to ON and saves, **When** a proxy request is made, **Then** responses include the provider/model tag
3. **Given** a user toggles the setting to OFF and saves, **When** a proxy request is made, **Then** responses are unmodified

---

### Edge Cases

- What happens when the response has no text content (e.g., pure tool_use / function_call)? → Tag is NOT injected; response is unmodified
- What happens when the response content array is empty? → No injection; response passed through as-is
- What happens when the provider name or model contains special characters? → Values are included as-is (they come from the config, which the user controls)
- What happens during failover (provider A fails, provider B succeeds)? → The tag shows provider B (the one that actually served the response)
- What happens when the response is already transformed between formats (OpenAI↔Anthropic)? → Tag injection happens AFTER format transformation, using the client's expected format
- What happens with non-200 error responses? → Tag is NOT injected on error responses (only on successful 2xx responses)

## Requirements

### Functional Requirements

- **FR-001**: The system MUST support a per-config boolean setting to enable/disable provider/model tag injection (default: disabled)
- **FR-002**: When enabled, the system MUST prepend `[provider: <name>, model: <model>]\n` to the first text content in successful (2xx) responses
- **FR-003**: For Anthropic Messages format, the tag MUST be prepended to the `text` field of the first `text` type block in the `content` array
- **FR-004**: For OpenAI Chat Completions format, the tag MUST be prepended to `choices[0].message.content`
- **FR-005**: For Anthropic SSE streaming, the tag MUST be prepended to the `text` field of the first `content_block_delta` event with `type: text_delta`
- **FR-006**: For OpenAI SSE streaming, the tag MUST be prepended to the first `delta.content` chunk
- **FR-007**: The tag MUST NOT be injected on error responses (non-2xx), tool-use-only responses, or empty content
- **FR-008**: The tag MUST reflect the actual provider and model that served the response (after failover, after model mapping)
- **FR-009**: The Web UI General Settings page MUST include a toggle for this setting
- **FR-010**: Tag injection MUST occur after any format transformation (OpenAI↔Anthropic) is applied
- **FR-011**: The setting MUST be readable and writable via the existing `/api/v1/settings` REST endpoint

### Key Entities

- **Provider Tag Setting**: A boolean flag in the global config that controls whether provider/model information is injected into proxy responses. Stored alongside other settings (web_port, default_profile, etc.)

## Success Criteria

### Measurable Outcomes

- **SC-001**: When enabled, 100% of successful text-bearing proxy responses contain the provider/model tag at the start of the first text content
- **SC-002**: When disabled (default), 0% of proxy responses are modified by this feature
- **SC-003**: Tag injection adds less than 5ms latency to response processing
- **SC-004**: The toggle in Web UI is discoverable and changes take effect on the next request without daemon restart
- **SC-005**: Feature works correctly for both Anthropic Messages and OpenAI Chat Completions formats, in both streaming and non-streaming modes

## Assumptions

- The tag format `[provider: xxx, model: xxx]\n` is fixed and not user-customizable (simplicity over configurability)
- The "model" shown in the tag is the actual model sent to the upstream provider (after any model mapping/override), not the original model from the client request
- The config change (adding a boolean field) uses the existing config migration pattern (version bump with backward-compatible default of `false`)
- The tag is plain text, not markdown or any structured format, to avoid interfering with client-side rendering
