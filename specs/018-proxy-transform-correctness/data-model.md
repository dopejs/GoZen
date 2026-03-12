# Data Model: Proxy Transform Layer Correctness

**Feature**: 018-proxy-transform-correctness
**Date**: 2026-03-09

## Protocol Format Constants

New string constants replacing the ambiguous "openai" client format identifier.

| Constant | Value | Usage |
|----------|-------|-------|
| `FormatAnthropicMessages` | `"anthropic-messages"` | Client using Anthropic Messages API (`/v1/messages`) |
| `FormatOpenAIChat` | `"openai-chat"` | Client using OpenAI Chat Completions API (`/chat/completions`) |
| `FormatOpenAIResponses` | `"openai-responses"` | Client using OpenAI Responses API (`/responses`) |

**Location**: `internal/proxy/transform/transform.go`

**Relationships**:
- `detectClientFormat()` returns one of these three values (previously returned `"openai"` or `"anthropic"`)
- `StreamTransformer.ClientFormat` holds one of these values
- `Transformer.TransformRequest(body, clientFormat)` receives one of these values

---

## Transform Error Type

New sentinel error type to distinguish local transform failures from provider failures.

```
TransformError
  - Op  string   // "request" or "response"
  - Err error    // underlying error
```

**Location**: `internal/proxy/server.go` (or `internal/proxy/transform/transform.go`)

**Lifecycle**:
1. Created in `forwardRequest` when `transformer.TransformRequest` or `transformer.TransformResponse` returns an error
2. Returned to `tryProviders`
3. Detected via `errors.As` in `tryProviders` → returns HTTP 500 to client, does NOT mark provider unhealthy

---

## StreamTransformer State

Existing struct — no new fields needed. The `ClientFormat` field will now carry fine-grained values.

```
StreamTransformer
  - ClientFormat   string   // "anthropic-messages" | "openai-chat" | "openai-responses"
  - ProviderFormat string   // "anthropic" | "openai"
  - MessageID      string
  - Model          string
```

**Routing logic** (updated):

| ClientFormat | ProviderFormat | Streaming Path |
|---|---|---|
| `openai-chat` | `anthropic` | `transformAnthropicToOpenAIChat` |
| `openai-responses` | `anthropic` | `transformAnthropicToOpenAIResponses` |
| `anthropic-messages` | `openai` | `transformOpenAIToAnthropic` |
| `anthropic-messages` | `openai-responses` | `transformResponsesAPIToAnthropic` |
| same format | same format | passthrough (no transform) |

---

## SSE Error Event Shapes

Protocol-native error events emitted when `scanner.Err()` is non-nil.

### OpenAI Chat (`openai-chat` client)
```
event: error
data: {"error":{"type":"stream_error","message":"<err>"}}

```

### OpenAI Responses (`openai-responses` client)
```
event: error
data: {"type":"error","error":{"type":"stream_error","message":"<err>"}}

```

### Anthropic Messages (`anthropic-messages` client)
```
event: error
data: {"type":"error","error":{"type":"stream_error","message":"<err>"}}

```

---

## Tool Call Transformation Mapping

Bidirectional mapping between OpenAI and Anthropic tool schemas.

### Request: OpenAI Chat → Anthropic

| OpenAI Chat field | Anthropic field |
|---|---|
| `tools[].type` = `"function"` | (implicit) |
| `tools[].function.name` | `tools[].name` |
| `tools[].function.description` | `tools[].description` |
| `tools[].function.parameters` | `tools[].input_schema` |
| `tool_choice` = `"auto"` | `tool_choice.type` = `"auto"` |
| `tool_choice` = `"none"` | (omit tool_choice) |
| `tool_choice.function.name` | `tool_choice.type` = `"tool"`, `tool_choice.name` |

### Response: Anthropic → OpenAI Chat

| Anthropic field | OpenAI Chat field |
|---|---|
| `content[].type` = `"tool_use"` | `choices[].message.tool_calls[]` |
| `content[].id` | `tool_calls[].id` |
| `content[].name` | `tool_calls[].function.name` |
| `content[].input` (object) | `tool_calls[].function.arguments` (JSON string) |
| `stop_reason` = `"tool_use"` | `choices[].finish_reason` = `"tool_calls"` |

### Streaming: Anthropic → OpenAI Chat tool deltas

| Anthropic SSE event | OpenAI Chat SSE event |
|---|---|
| `content_block_start` with `type=tool_use` | `chat.completion.chunk` with `tool_calls[].id`, `tool_calls[].function.name` |
| `content_block_delta` with `type=input_json_delta` | `chat.completion.chunk` with `tool_calls[].function.arguments` delta |
| `content_block_stop` | (no direct equivalent, index advances) |

### Streaming: Anthropic → OpenAI Responses tool deltas

| Anthropic SSE event | OpenAI Responses SSE event |
|---|---|
| `content_block_start` with `type=tool_use` | `response.output_item.added` with `type=function_call` |
| `content_block_delta` with `type=input_json_delta` | `response.function_call_arguments.delta` |
| `content_block_stop` | `response.output_item.done` |
