# Interface Contracts: Proxy Transform Layer

**Feature**: 018-proxy-transform-correctness
**Date**: 2026-03-09

## Transform Package Public API

### Constants

```go
// Protocol format identifiers for client-side format detection.
// These replace the ambiguous "openai" string used previously.
const (
    FormatAnthropicMessages = "anthropic-messages"
    FormatOpenAIChat        = "openai-chat"
    FormatOpenAIResponses   = "openai-responses"
)
```

### Transformer Interface (unchanged)

```go
type Transformer interface {
    Name() string
    TransformRequest(body []byte, clientFormat string) ([]byte, error)
    TransformResponse(body []byte, clientFormat string) ([]byte, error)
}
```

`clientFormat` now accepts `FormatAnthropicMessages`, `FormatOpenAIChat`, or `FormatOpenAIResponses`.

### GetTransformer (unchanged signature)

```go
func GetTransformer(providerType string) Transformer
```

Continues to accept `"openai"` or `"anthropic"` as provider type.

### StreamTransformer (ClientFormat field semantics updated)

```go
type StreamTransformer struct {
    ClientFormat   string // now: "anthropic-messages" | "openai-chat" | "openai-responses"
    ProviderFormat string // unchanged: "anthropic" | "openai"
    MessageID      string
    Model          string
}

func (st *StreamTransformer) TransformSSEStream(r io.Reader) io.Reader
```

### NeedsTransform (unchanged)

```go
func NeedsTransform(clientFormat, providerFormat string) bool
```

---

## Proxy Package: Transform Error Type

```go
// TransformError indicates a local transformation failure.
// It is distinct from provider errors and must NOT trigger provider health changes.
type TransformError struct {
    Op  string // "request" or "response"
    Err error
}

func (e *TransformError) Error() string
func (e *TransformError) Unwrap() error
```

**Contract**: When `tryProviders` receives a `TransformError`, it MUST:
1. Return HTTP 500 to the client with body `{"error":{"type":"transform_error","message":"..."}}`
2. NOT call `provider.MarkFailed()`, `provider.MarkAuthFailed()`, or any health-modifying method
3. NOT attempt failover to the next provider

---

## detectClientFormat Contract (updated)

```go
// detectClientFormat returns the client protocol format based on request path and client type.
// Return values are one of the FormatXxx constants from the transform package.
func detectClientFormat(path, clientType string) string
```

| Input path | Return value |
|---|---|
| ends with `/chat/completions` | `transform.FormatOpenAIChat` |
| ends with `/responses` or contains `/responses/` | `transform.FormatOpenAIResponses` |
| `clientType == "codex"` | `transform.FormatOpenAIChat` |
| anything else | `transform.FormatAnthropicMessages` |

---

## SSE Stream Error Contract

When `scanner.Err()` is non-nil after a streaming loop, the transformer MUST:

1. Emit a protocol-native error event (see data-model.md for shapes)
2. Close the pipe writer with the scanner error
3. NOT emit any completion event (`message_stop`, `response.completed`, `[DONE]`)

When `scanner.Err()` is nil (clean EOF), the transformer MUST:
1. Emit the appropriate completion event for the client format
2. Close the pipe writer with nil
