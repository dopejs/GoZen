# Data Model: Fix Proxy Model Transform

## Entities (no changes)

This is a bug fix — no new entities, fields, or schema changes are introduced.

## Behavioral Changes

### Provider (internal/proxy/Provider struct)

| Field | Before | After |
|-------|--------|-------|
| `ReasoningModel` | Filled with `"claude-sonnet-4-5-thinking"` when empty (all providers) | Filled only when provider `Type == "anthropic"` |
| `HaikuModel` | Filled with `"claude-haiku-4-5"` when empty (all providers) | Filled only when provider `Type == "anthropic"` |
| `OpusModel` | Filled with `"claude-opus-4-5"` when empty (all providers) | Filled only when provider `Type == "anthropic"` |
| `SonnetModel` | Filled with `"claude-sonnet-4-5"` when empty (all providers) | Filled only when provider `Type == "anthropic"` |

### Model Mapping (mapModel behavior)

| Scenario | Before | After |
|----------|--------|-------|
| OpenAI provider, no sonnet_model, request has "sonnet" | Returns `"claude-sonnet-4-5"` (wrong default) | Falls through to `p.Model` (provider's default) |
| OpenAI provider, explicit sonnet_model | Returns the explicit value | No change |
| Anthropic provider, no sonnet_model | Returns `"claude-sonnet-4-5"` (correct default) | No change |

### Path Construction (forwardRequest behavior)

| base_url | targetPath | Before | After |
|----------|------------|--------|-------|
| `https://host/v1` | `/v1/chat/completions` | `https://host/v1/v1/chat/completions` (404) | `https://host/v1/chat/completions` |
| `https://host` | `/v1/chat/completions` | `https://host/v1/chat/completions` | No change |
| `https://host/v1/` | `/v1/chat/completions` | `https://host/v1//v1/chat/completions` | `https://host/v1/chat/completions` |
| `https://host/api` | `/v1/messages` | `https://host/api/v1/messages` | No change |

## State Transitions

No state transitions affected. No config migration needed.
