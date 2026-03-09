# Daemon Proxy Stability Plan

This document captures the next stability-focused follow-up work for the daemon proxy, with emphasis on the `transform` layer and provider failover/fallback behavior.

## Scope

This plan is intended for the stability-improvement track around the `3.0.1` release line.

Priority is based on one core rule:

- The daemon proxy is a P0 component.
- Protocol correctness and self-healing behavior take precedence over feature completeness.
- Transform-layer mistakes must not be misclassified as provider instability.

## Current Assessment

### Transform

The transform layer currently has a structural correctness problem:

- OpenAI Chat Completions and OpenAI Responses API are both treated as a single `openai` format.
- Request path mapping, request transformation, response transformation, and SSE transformation do not consistently preserve which protocol the client actually used.
- Tool-call / tool-use conversion is only partially implemented.
- SSE transformers do not always surface upstream stream errors correctly.

This means the transform layer can appear to "mostly work" while still failing on real protocol boundaries, especially for:

- `/responses` vs `/chat/completions`
- streaming responses
- tool calls / tool use
- partial or interrupted streams

### Fallback / Failover

Provider failover currently behaves more like ordered failover than a full strategy-aware fallback system:

- fixed provider order remains dominant
- unhealthy/backoff skipping works
- scenario route to default route fallback works
- profile-level load-balancing strategy is not yet fully connected to the main request path

This is acceptable for baseline availability, but it is not yet the final form of an optimized fallback system.

## Priority Plan

## P0

### 1. Split OpenAI protocol handling into explicit formats

Introduce distinct protocol identities instead of treating all OpenAI traffic as one format.

Recommended minimum set:

- `anthropic-messages`
- `openai-chat`
- `openai-responses`

Required follow-up:

- detect client format using the actual endpoint semantics
- transform request path using the correct protocol pair
- return response bodies in the same protocol family the client used
- transform SSE streams according to the real client protocol, not a generic `openai` bucket

Expected outcome:

- `/responses` clients never receive Chat Completions payloads
- `/chat/completions` clients never receive Responses API event streams

### 2. Complete tool-call / tool-use transformation

Complete both non-streaming and streaming conversions for tool-related payloads.

Required directions:

- OpenAI -> Anthropic
  - `tool_calls` -> `tool_use`
- Anthropic -> OpenAI
  - `tool_use` -> `tool_calls`
- Responses API -> Anthropic SSE
  - function-call arguments -> `input_json_delta`
- Anthropic SSE -> OpenAI SSE
  - tool-related block events -> protocol-correct OpenAI events

Expected outcome:

- tool-enabled clients continue working across provider failover
- transform correctness is no longer limited to text-only requests

### 3. Make SSE transformation fail safely

SSE transformers must not fabricate successful completion when the upstream stream is malformed or truncated.

Required behavior:

- check and handle `scanner.Err()` explicitly
- stop emitting completion events when the upstream stream ended in error
- preserve interruption semantics instead of converting them into a normal completion

Expected outcome:

- broken upstream streams remain visibly broken
- the proxy does not convert transport/protocol corruption into fake success

## P1

### 4. Treat local transform failures as proxy/transform errors, not provider failures

When request/response transformation fails locally, the system should not continue as if the provider failed.

Recommended policy:

- request transform failure:
  - return a clear proxy/transform error to the client
  - do not send malformed payloads to the provider
  - do not mark the provider unhealthy
- response transform failure:
  - return a clear proxy/transform error
  - do not silently pass through an incompatible response shape

Expected outcome:

- provider health signals remain trustworthy
- local protocol bugs do not poison fallback decisions

### 5. Remove default body-level debug logging from the transform hot path

The transform package should not perform default file I/O or log full request bodies during normal execution.

Recommended policy:

- no package-init file creation for transform logging
- no default request/response body dumps
- if debugging is required, gate it behind an explicit debug flag

Expected outcome:

- lower runtime noise
- reduced risk of oversized logs
- cleaner daemon proxy hot path

## P2

### 6. Connect profile strategy to real provider selection

The repository already has:

- profile-level strategy configuration
- load-balancer implementation

But the main request path still behaves primarily like ordered failover.

Recommended follow-up:

- pass profile strategy into the active proxy path
- let provider selection honor configured strategy before failover execution
- preserve unhealthy/backoff handling after strategy ordering

Expected outcome:

- fallback evolves from static failover into strategy-aware selection
- least-latency / least-cost / round-robin can influence real runtime routing

## Recommended Execution Order

1. Split protocol identities: `openai-chat` vs `openai-responses`
2. Fix response/SSE shape correctness for all protocol pairs
3. Complete tool-call and tool-use transformations
4. Tighten transform failure semantics
5. Remove transform hot-path debug logging
6. Connect profile strategy to provider selection

## Release Guidance

For the `3.0.1` stability track, the most valuable work is:

1. protocol-boundary correctness in `transform`
2. stream correctness under failure
3. tool-call correctness across protocol conversion

Fallback optimization should come after transform correctness is trustworthy.

In short:

- Fix `transform` correctness first.
- Improve fallback strategy second.

