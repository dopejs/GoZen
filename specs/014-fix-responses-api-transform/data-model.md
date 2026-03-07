# Data Model: Responses API Transform

**Feature**: 014-fix-responses-api-transform
**Date**: 2026-03-07

## Entities

This feature operates on JSON request/response bodies. No persistent data model changes.

### Chat Completions Request Body (existing)

```json
{
  "model": "gpt-5.4",
  "messages": [
    {"role": "system", "content": "..."},
    {"role": "user", "content": "..."}
  ],
  "stream": true,
  "max_completion_tokens": 16384,
  "temperature": 1.0,
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "tool_name",
        "description": "...",
        "parameters": {"type": "object", "properties": {...}}
      }
    }
  ]
}
```

### Responses API Request Body (new target)

```json
{
  "model": "gpt-5.4",
  "input": [
    {"role": "system", "content": "..."},
    {"role": "user", "content": "..."}
  ],
  "stream": true,
  "max_output_tokens": 16384,
  "temperature": 1.0,
  "tools": [
    {
      "type": "function",
      "name": "tool_name",
      "description": "...",
      "parameters": {"type": "object", "properties": {...}}
    }
  ]
}
```

### Responses API Response Body (non-streaming)

```json
{
  "id": "resp_abc123",
  "object": "response",
  "status": "completed",
  "model": "gpt-5.4",
  "output": [
    {
      "id": "msg_xyz",
      "type": "message",
      "role": "assistant",
      "content": [
        {"type": "output_text", "text": "Hello!"}
      ]
    }
  ],
  "usage": {
    "input_tokens": 100,
    "output_tokens": 50,
    "total_tokens": 150
  }
}
```

### Responses API Function Call Output

```json
{
  "output": [
    {
      "id": "fc_123",
      "type": "function_call",
      "call_id": "call_123",
      "name": "get_weather",
      "arguments": "{\"location\":\"Paris\"}",
      "status": "completed"
    }
  ]
}
```

### Anthropic Response Body (transform target)

```json
{
  "id": "resp_abc123",
  "type": "message",
  "role": "assistant",
  "model": "gpt-5.4",
  "content": [
    {"type": "text", "text": "Hello!"}
  ],
  "stop_reason": "end_turn",
  "usage": {
    "input_tokens": 100,
    "output_tokens": 50
  }
}
```

## Transformation Rules

### Status → Stop Reason Mapping

| Responses API `status` | Anthropic `stop_reason` |
|------------------------|------------------------|
| `completed` | `end_turn` |
| `incomplete` | `max_tokens` |
| `failed` | `end_turn` (with empty content) |

### Output Type → Content Type Mapping

| Responses API output `type` | Anthropic content `type` |
|----------------------------|-------------------------|
| `message` → `content[].output_text` | `text` |
| `function_call` | `tool_use` |

### Tool Call Field Mapping

| Responses API | Anthropic |
|--------------|-----------|
| `call_id` | `id` |
| `name` | `name` |
| `arguments` (JSON string) | `input` (parsed JSON object) |
