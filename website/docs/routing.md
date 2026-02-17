---
sidebar_position: 4
title: Scenario Routing
---

# Scenario Routing

Automatically route requests to different providers based on request characteristics.

## Supported Scenarios

| Scenario | Description |
|----------|-------------|
| `think` | Thinking mode enabled |
| `image` | Contains image content |
| `longContext` | Content exceeds threshold |
| `webSearch` | Uses web_search tool |
| `background` | Uses Haiku model |

## Fallback Mechanism

If all providers for a scenario fail, it automatically falls back to the profile's default providers.

## Configuration Example

```json
{
  "profiles": {
    "smart": {
      "providers": ["main-api"],
      "long_context_threshold": 60000,
      "routing": {
        "think": {
          "providers": [{"name": "thinking-api", "model": "claude-opus-4-5"}]
        },
        "longContext": {
          "providers": [{"name": "long-context-api"}]
        }
      }
    }
  }
}
```
