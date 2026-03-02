# Research: Provider & Model Tag in Proxy Responses

**Feature**: 005-provider-model-tag
**Date**: 2026-03-02

## Research Questions

### RQ1: Where should tag injection happen in the response pipeline?

**Decision**: Tag injection in `copyResponse()` in `internal/proxy/server.go`, AFTER format transformation.

**Rationale**: `copyResponse()` (line 502) is the single point where all successful responses are written to the client. It already handles two paths:
- **Non-streaming** (line 547-577): reads full body, applies transform, writes. Tag injection parses the JSON body and modifies the content field before writing.
- **SSE streaming** (line 510-544): copies raw bytes through a reader. Tag injection wraps the reader to intercept the first text delta event.

Tag injection AFTER transformation ensures FR-010 compliance (tag uses client's expected format).

**Alternatives considered**:
- Inject in `forwardRequest()` → rejected: too early, before transformation
- Inject in `tryProviders()` → rejected: wrong abstraction level, would couple tag logic to failover
- New middleware → rejected: YAGNI, existing pipeline is sufficient

### RQ2: How to access the actual provider name and model for the tag?

**Decision**: Provider name from `p.Name` (already available in `copyResponse`). Model from the modified request body parsed in `forwardRequest()`, passed through to `copyResponse()` via a new parameter or extracted from the response body.

**Rationale**: The Provider `p` passed to `copyResponse()` already has `p.Name`. For the model, two approaches:
1. Extract from the upstream response body (Anthropic responses include `model` field; OpenAI responses include `model` field in non-streaming, and in the first SSE chunk)
2. Parse the modified request body to get the model after mapping

Option 1 is preferred because:
- Non-streaming: the response JSON already contains the actual model used
- Streaming: the first SSE event (Anthropic `message_start` or OpenAI first chunk) contains the model
- No need to change `forwardRequest()` signature
- Shows the model the upstream provider actually reports (most accurate after failover)

For non-streaming responses, parse the response body JSON to extract the model. For streaming, parse the first SSE event to extract the model before prepending the tag.

**Alternatives considered**:
- Pass model from `forwardRequest()` → rejected: requires signature change, response body already has it
- Use `p.Model` field → rejected: doesn't reflect model mapping/override results

### RQ3: How should the config field be added?

**Decision**: Add `ShowProviderTag bool` field to `OpenCCConfig` struct with `json:"show_provider_tag,omitempty"`. Bump `CurrentConfigVersion` from 10 to 11.

**Rationale**: Per Constitution Principle III, any config schema change requires a version bump. The new boolean field defaults to `false` (Go zero value + `omitempty`), so no migration logic is needed — old configs simply won't have the field, which is the desired default-off behavior. The version bump is still required per constitution but migration is a no-op.

Pattern follows existing fields like `ProxyPort`, `WebPort` (omitempty, zero-value safe).

**Alternatives considered**:
- Nested config struct → rejected: YAGNI, single boolean doesn't need nesting
- Per-profile setting → rejected: spec says "per-config" (global) setting

### RQ4: How to modify non-streaming responses for both formats?

**Decision**: Parse the response body as JSON, modify the content field based on the client's requested format (after transformation), re-marshal.

**Implementation approach**:
- **Anthropic format**: Parse `content` array, find first block with `type: "text"`, prepend tag to its `text` field
- **OpenAI format**: Parse `choices[0].message.content`, prepend tag string

This happens after transformation in `copyResponse()`, so `requestFormat` tells us which format the client expects.

### RQ5: How to modify SSE streaming responses for both formats?

**Decision**: Create a tag-injecting reader wrapper that intercepts the first text delta event and prepends the tag.

**Implementation approach**:
- For **Anthropic SSE** (client expects Anthropic format): intercept first `content_block_delta` event with `type: text_delta`, prepend tag to its `text` field
- For **OpenAI SSE** (client expects OpenAI chat completions): intercept first `data:` line with `delta.content`, prepend tag
- For **OpenAI Responses API SSE** (transformed from Anthropic): intercept first `response.output_text.delta` event, prepend tag to its `delta` field

The wrapper reader sits in the pipeline after any `StreamTransformer` but before the raw copy loop. It buffers SSE events, parses the first text delta, injects the tag, then passes through all remaining data unmodified.

### RQ6: What changes are needed in the Web UI?

**Decision**: Add `show_provider_tag` field to Settings type, settingsRequest, settingsResponse. Add a toggle Switch in GeneralSettings tab.

**Files affected**:
- `web/src/types/api.ts`: Add `show_provider_tag?: boolean` to `Settings` interface
- `web/src/pages/settings/tabs/GeneralSettings.tsx`: Add Switch component for toggle
- `web/src/i18n/locales/*.json`: Add translation keys for the toggle label and description
- `internal/web/api_settings.go`: Add `ShowProviderTag` to settingsResponse/settingsRequest, wire getter/setter

### RQ7: How to extract model from streaming responses for the tag?

**Decision**: For SSE streams, the tag injection reader must buffer and parse early events to extract the model, then prepend the tag to the first text delta.

**Details**:
- **Anthropic SSE**: `message_start` event contains `message.model` field. This arrives before any `content_block_delta`.
- **OpenAI SSE**: First chunk contains `model` field. This arrives before any `delta.content`.
- **OpenAI Responses API SSE**: `response.created` event contains `response.model`. Arrives before `response.output_text.delta`.

The reader can extract the model from these early events while passing them through to the client, then use it when the first text delta arrives.

## Technology Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Injection point | `copyResponse()` in server.go | Single point for all response writing |
| Model source | Response body (non-streaming) / first SSE event (streaming) | Most accurate, no signature changes |
| Config field | `ShowProviderTag bool` in OpenCCConfig | Simplest, zero-value = disabled |
| Config version | 10 → 11 | Constitution Principle III requirement |
| Streaming approach | Reader wrapper that intercepts first text delta | Composable with existing StreamTransformer pipeline |
| Web UI | Switch toggle in GeneralSettings tab | Consistent with existing settings UX pattern |
