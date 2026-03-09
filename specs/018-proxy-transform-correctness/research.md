# Research: Proxy Transform Layer Correctness

**Feature**: 018-proxy-transform-correctness
**Date**: 2026-03-09

## Decision 1: Protocol Format Identifiers

**Decision**: Use three string constants for protocol formats: `anthropic-messages`, `openai-chat`, `openai-responses`

**Rationale**: The existing code uses `config.ProviderTypeOpenAI` ("openai") and `config.ProviderTypeAnthropic` ("anthropic") as provider type identifiers. The new client format identifiers must be distinct from provider types to avoid confusion. Using hyphenated names makes the distinction clear.

**Alternatives considered**:
- Reuse existing `ProviderTypeOpenAI`/`ProviderTypeAnthropic` constants — rejected because they conflate client format with provider type
- Use integer enum — rejected, string constants are idiomatic in this codebase

**Implementation note**: Define new constants in `internal/proxy/transform/transform.go`:
```go
const (
    FormatAnthropicMessages = "anthropic-messages"
    FormatOpenAIChat        = "openai-chat"
    FormatOpenAIResponses   = "openai-responses"
)
```

---

## Decision 2: Transformer Interface Extension

**Decision**: Extend `GetTransformer` to accept both client format and provider format, returning a transformer that knows the full context. Alternatively, keep the interface but pass `clientFormat` as a richer string.

**Rationale**: The current `Transformer` interface passes `clientFormat` as a string to `TransformRequest`/`TransformResponse`. The simplest change is to update `detectClientFormat` to return the new fine-grained format strings (`openai-chat`, `openai-responses`, `anthropic-messages`) and update the transformer methods to branch on these values. No interface change needed.

**Alternatives considered**:
- Create separate `OpenAIChatTransformer` and `OpenAIResponsesTransformer` structs — adds complexity; the existing `AnthropicTransformer` already handles both directions via `clientFormat` parameter
- Add a new `GetTransformer(clientFormat, providerFormat string)` — cleaner but requires more refactoring

**Implementation note**: Update `detectClientFormat` return values; update transformer `TransformRequest`/`TransformResponse`/`TransformSSEStream` to branch on `openai-chat` vs `openai-responses`.

---

## Decision 3: SSE Scanner Error Propagation

**Decision**: After each scanner loop, check `scanner.Err()`. If non-nil, emit a protocol-native error event then close the pipe writer with the error.

**Rationale**: The three streaming functions (`transformAnthropicToOpenAI`, `transformOpenAIToAnthropic`, `transformResponsesAPIToAnthropic`) all use `bufio.Scanner` loops with no `scanner.Err()` check. Adding the check after each loop is minimal and targeted.

**Protocol-native error event formats**:
- OpenAI Chat (`openai-chat` client): `event: error\ndata: {"error":{"type":"stream_error","message":"..."}}\n\n`
- OpenAI Responses (`openai-responses` client): `event: error\ndata: {"type":"error","error":{"type":"stream_error","message":"..."}}\n\n`
- Anthropic (`anthropic-messages` client): `event: error\ndata: {"type":"error","error":{"type":"stream_error","message":"..."}}\n\n`

**Alternatives considered**:
- Close stream silently — rejected per spec FR-010
- Return HTTP 500 — not possible mid-stream after headers already sent

---

## Decision 4: Transform Error Classification

**Decision**: In `server.go` `forwardRequest`, when `transformer.TransformRequest` returns an error, return an error immediately to the caller with a sentinel value that `tryProviders` can detect as a transform error (not a provider error).

**Rationale**: Currently transform errors are logged and silently ignored (body sent untransformed). The fix is to return early with an error. `tryProviders` must distinguish transform errors from provider errors to avoid marking the provider unhealthy.

**Implementation approach**:
- Define a `transformError` type in `internal/proxy/` (or use a sentinel variable)
- In `forwardRequest`, return `transformError` when transform fails
- In `tryProviders`, check `errors.As(err, &transformError{})` — if true, return 500 to client without marking provider unhealthy

**Alternatives considered**:
- Add a boolean flag to `providerFailure` struct — less idiomatic than typed errors
- Use error wrapping with `fmt.Errorf("transform: %w", err)` and string matching — fragile

---

## Decision 5: Debug Logging Removal

**Decision**: Remove the `init()` function and `debugLogger` from `anthropic.go`. Remove all `debugLogger.Printf(...)` call sites. The transform hot path will have zero file I/O by default.

**Rationale**: The `init()` function unconditionally opens/creates `~/.zen-dev/transform.log` on every process start. This is a dev artifact that should never have been in production code. Removing it entirely is simpler than gating it.

**Optional debug support**: If debug logging is needed in future, it can be added via `GOZEN_TRANSFORM_DEBUG` env var check at call site (lazy init, not `init()`). This is out of scope for this feature per spec.

**Alternatives considered**:
- Gate with `os.Getenv("GOZEN_TRANSFORM_DEBUG")` in `init()` — still runs `init()` on every start; lazy init is better
- Keep logger but write to stderr only — still adds overhead

---

## Existing Code Inventory

### Files to modify

| File | Change |
|------|--------|
| `internal/proxy/transform/transform.go` | Add format constants, update `GetTransformer` if needed |
| `internal/proxy/transform/anthropic.go` | Remove `init()`, `debugLogger`, all `debugLogger.Printf` calls |
| `internal/proxy/transform/stream.go` | Add `scanner.Err()` checks, add protocol-native error emission, branch on `openai-chat` vs `openai-responses` |
| `internal/proxy/transform/openai.go` | Update `TransformRequest`/`TransformResponse` to handle `openai-chat` vs `openai-responses` client formats |
| `internal/proxy/profile_proxy.go` | Update `detectClientFormat` to return `openai-chat`/`openai-responses`/`anthropic-messages` |
| `internal/proxy/server.go` | Add transform error type, update `forwardRequest` to return error on transform failure, update `tryProviders` to handle transform errors |

### Files to add tests to

| File | New test cases |
|------|---------------|
| `internal/proxy/transform/stream_test.go` | Scanner error propagation, protocol-native error events |
| `internal/proxy/transform/openai_test.go` | `openai-chat` vs `openai-responses` response shape |
| `internal/proxy/transform/anthropic_test.go` | Tool call streaming deltas |
| `internal/proxy/server_test.go` (or integration) | Transform error → no provider health impact |

---

## Constitution Compliance

- **Principle I (TDD)**: All changes will be test-driven. Tests written first for each sub-task.
- **Principle II (YAGNI)**: No new abstractions beyond what's needed. Minimal changes to existing interfaces.
- **Principle III (Config Migration)**: No config schema changes required.
- **Principle VI (Coverage)**: `internal/proxy/transform` must stay ≥ 80%. New test cases added for all changed paths.
