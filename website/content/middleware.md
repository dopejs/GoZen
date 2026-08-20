---
sidebar_position: 15
title: Middleware Pipeline (BETA)
---

# Middleware Pipeline (BETA)

:::warning BETA Feature
Middleware pipeline is currently in beta. It is disabled by default and requires explicit configuration to enable.
:::

Extend GoZen with pluggable middleware for request/response transformation, logging, rate limiting, and custom processing.

## Features

- **Pluggable architecture** — Add custom processing logic without modifying core code
- **Priority-based execution** — Control middleware execution order
- **Request/response hooks** — Process requests before sending, responses after receiving
- **Built-in middleware** — Context injection, logging, rate limiting, compression
- **Plugin loader** — Load middleware from local files or remote URLs
- **Error handling** — Graceful error handling with fallback behavior

## Architecture

```
Client Request
    ↓
[Middleware 1: Priority 100]
    ↓
[Middleware 2: Priority 200]
    ↓
[Middleware 3: Priority 300]
    ↓
Provider API
    ↓
[Middleware 3: Response]
    ↓
[Middleware 2: Response]
    ↓
[Middleware 1: Response]
    ↓
Client Response
```

## Configuration

### Enable Middleware Pipeline

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "context-injection",
        "enabled": true,
        "priority": 100,
        "config": {}
      },
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 200,
        "config": {
          "log_level": "info"
        }
      }
    ]
  }
}
```

**Options:**

| Option | Description |
|--------|-------------|
| `enabled` | Enable middleware pipeline |
| `pipeline` | Array of middleware configurations |
| `name` | Middleware identifier |
| `priority` | Execution order (lower = earlier) |
| `config` | Middleware-specific configuration |

## Built-in Middleware

### 1. Context Injection

Inject custom context into requests.

```json
{
  "name": "context-injection",
  "enabled": true,
  "priority": 100,
  "config": {
    "system_prompt": "You are a helpful coding assistant.",
    "metadata": {
      "session_id": "sess_123",
      "user_id": "user_456"
    }
  }
}
```

**Use cases:**
- Add system prompts
- Inject session metadata
- Add user context

### 2. Request Logger

Log all requests and responses.

```json
{
  "name": "request-logger",
  "enabled": true,
  "priority": 200,
  "config": {
    "log_level": "info",
    "log_body": false,
    "log_headers": true
  }
}
```

**Use cases:**
- Debugging
- Audit trails
- Performance monitoring

### 3. Rate Limiter

Limit request rate per provider or globally.

```json
{
  "name": "rate-limiter",
  "enabled": true,
  "priority": 300,
  "config": {
    "requests_per_minute": 60,
    "burst": 10,
    "per_provider": true
  }
}
```

**Use cases:**
- Prevent rate limit errors
- Control API usage
- Protect against abuse

### 4. Compression (BETA)

Compress context when token count exceeds threshold.

```json
{
  "name": "compression",
  "enabled": true,
  "priority": 400,
  "config": {
    "threshold_tokens": 50000,
    "target_tokens": 20000
  }
}
```

See [Context Compression](./compression.md) for details.

### 5. Session Memory (BETA)

Maintain conversation memory across sessions.

```json
{
  "name": "session-memory",
  "enabled": true,
  "priority": 150,
  "config": {
    "max_memories": 100,
    "ttl_hours": 24,
    "storage": "sqlite"
  }
}
```

**Use cases:**
- Remember user preferences
- Track conversation history
- Maintain context across sessions

### 6. Orchestration (BETA)

Route requests to multiple providers and aggregate responses.

```json
{
  "name": "orchestration",
  "enabled": true,
  "priority": 500,
  "config": {
    "strategy": "parallel",
    "providers": ["anthropic", "openai"],
    "consensus": "longest"
  }
}
```

**Use cases:**
- Compare model outputs
- Redundancy for critical requests
- Quality improvement through consensus

## Custom Middleware

### Middleware Interface

```go
type Middleware interface {
    Name() string
    Priority() int
    ProcessRequest(ctx *RequestContext) error
    ProcessResponse(ctx *ResponseContext) error
}

type RequestContext struct {
    Provider  string
    Model     string
    Messages  []Message
    Metadata  map[string]interface{}
}

type ResponseContext struct {
    Provider  string
    Model     string
    Response  *APIResponse
    Latency   time.Duration
    Metadata  map[string]interface{}
}
```

### Example: Custom Header Injection

```go
package main

import (
    "github.com/dopejs/gozen/internal/middleware"
)

type CustomHeaderMiddleware struct {
    headers map[string]string
}

func (m *CustomHeaderMiddleware) Name() string {
    return "custom-headers"
}

func (m *CustomHeaderMiddleware) Priority() int {
    return 250
}

func (m *CustomHeaderMiddleware) ProcessRequest(ctx *middleware.RequestContext) error {
    for k, v := range m.headers {
        ctx.Metadata[k] = v
    }
    return nil
}

func (m *CustomHeaderMiddleware) ProcessResponse(ctx *middleware.ResponseContext) error {
    // No response processing needed
    return nil
}

func init() {
    middleware.Register("custom-headers", func(config map[string]interface{}) middleware.Middleware {
        return &CustomHeaderMiddleware{
            headers: config["headers"].(map[string]string),
        }
    })
}
```

### Loading Custom Middleware

#### Local Plugin

```json
{
  "middleware": {
    "enabled": true,
    "plugins": [
      {
        "type": "local",
        "path": "/path/to/custom-middleware.so",
        "config": {
          "headers": {
            "X-Custom-Header": "value"
          }
        }
      }
    ]
  }
}
```

#### Remote Plugin

```json
{
  "middleware": {
    "enabled": true,
    "plugins": [
      {
        "type": "remote",
        "url": "https://example.com/middleware/custom-headers.so",
        "checksum": "sha256:abc123...",
        "config": {}
      }
    ]
  }
}
```

## Web UI

Access middleware settings at `http://localhost:19840/settings`:

1. Navigate to "Middleware" tab (marked with BETA badge)
2. Toggle "Enable Middleware Pipeline"
3. Add/remove middleware from pipeline
4. Adjust priority and configuration
5. Enable/disable individual middleware
6. Click "Save"

## API Endpoints

### List Middleware

```bash
GET /api/v1/middleware
```

Response:
```json
{
  "enabled": true,
  "pipeline": [
    {
      "name": "context-injection",
      "enabled": true,
      "priority": 100,
      "type": "builtin"
    },
    {
      "name": "request-logger",
      "enabled": true,
      "priority": 200,
      "type": "builtin"
    }
  ]
}
```

### Add Middleware

```bash
POST /api/v1/middleware
Content-Type: application/json

{
  "name": "rate-limiter",
  "enabled": true,
  "priority": 300,
  "config": {
    "requests_per_minute": 60
  }
}
```

### Update Middleware

```bash
PUT /api/v1/middleware/{name}
Content-Type: application/json

{
  "enabled": false
}
```

### Remove Middleware

```bash
DELETE /api/v1/middleware/{name}
```

### Reload Pipeline

```bash
POST /api/v1/middleware/reload
```

## Use Cases

### Development Environment

Add debug logging and request inspection:

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 100,
        "config": {
          "log_level": "debug",
          "log_body": true
        }
      }
    ]
  }
}
```

### Production Environment

Add rate limiting and monitoring:

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "rate-limiter",
        "enabled": true,
        "priority": 100,
        "config": {
          "requests_per_minute": 100,
          "burst": 20
        }
      },
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 200,
        "config": {
          "log_level": "info",
          "log_body": false
        }
      }
    ]
  }
}
```

### Multi-Provider Comparison

Use orchestration to compare outputs:

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "orchestration",
        "enabled": true,
        "priority": 500,
        "config": {
          "strategy": "parallel",
          "providers": ["anthropic", "openai", "google"],
          "consensus": "longest"
        }
      }
    ]
  }
}
```

## Best Practices

1. **Use appropriate priorities** — Lower numbers execute first
2. **Keep middleware focused** — Each middleware should do one thing well
3. **Handle errors gracefully** — Don't break the pipeline on errors
4. **Test thoroughly** — Verify middleware behavior before production
5. **Monitor performance** — Track middleware overhead
6. **Document configuration** — Clearly document config options

## Limitations

1. **Performance overhead** — Each middleware adds latency
2. **Complexity** — Too many middleware can make debugging difficult
3. **Plugin security** — Remote plugins require trust and verification
4. **Error propagation** — Middleware errors can affect all requests
5. **Configuration complexity** — Complex pipelines are harder to maintain

## Troubleshooting

### Middleware not executing

1. Verify `middleware.enabled` is `true`
2. Check middleware is enabled in pipeline
3. Verify priority is set correctly
4. Review daemon logs for middleware errors

### Unexpected behavior

1. Check middleware execution order (priority)
2. Verify configuration is correct
3. Test middleware in isolation
4. Review middleware logs

### Performance issues

1. Identify slow middleware (check logs)
2. Reduce middleware count
3. Optimize middleware implementation
4. Consider disabling non-essential middleware

### Plugin loading failures

1. Verify plugin path is correct
2. Check plugin is compiled for correct architecture
3. Verify checksum matches (for remote plugins)
4. Review plugin logs for errors

## Security Considerations

1. **Validate plugins** — Only load trusted plugins
2. **Verify checksums** — Always verify remote plugin checksums
3. **Sandbox plugins** — Consider running plugins in isolated environment
4. **Audit middleware** — Review middleware code before deployment
5. **Monitor behavior** — Watch for unexpected middleware behavior

## Future Enhancements

- WebAssembly plugin support for cross-platform compatibility
- Middleware marketplace for sharing community plugins
- Visual pipeline editor in Web UI
- Middleware performance profiling
- Hot-reload for plugin updates
- Middleware testing framework
