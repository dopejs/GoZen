# GoZen Middleware Development Guide

This guide explains how to develop custom middleware for GoZen's programmable middleware pipeline.

## Overview

GoZen's middleware pipeline allows you to intercept, modify, and enhance API requests and responses flowing through the proxy. Middleware can be used for:

- Auto-injecting project context
- Logging and monitoring
- Request/response transformation
- Caching
- Rate limiting
- Custom routing logic
- And much more

## Middleware Interface

All middleware must implement the `Middleware` interface:

```go
type Middleware interface {
    // Name returns the unique identifier for this middleware
    Name() string

    // Version returns the middleware version (semver)
    Version() string

    // Description returns a human-readable description
    Description() string

    // Init is called once when the middleware is loaded
    // config contains middleware-specific configuration from zen.json
    Init(config json.RawMessage) error

    // ProcessRequest is called before the request is sent to the provider
    // Return modified context, or error to abort
    ProcessRequest(ctx *RequestContext) (*RequestContext, error)

    // ProcessResponse is called after receiving the response
    // Return modified context, or error to abort
    ProcessResponse(ctx *ResponseContext) (*ResponseContext, error)

    // Priority returns the execution order (lower = earlier)
    // Built-in middleware: 0-99, User middleware: 100+
    Priority() int

    // Close is called when the middleware is unloaded
    Close() error
}
```

## Context Types

### RequestContext

```go
type RequestContext struct {
    // Request metadata
    SessionID   string            // Unique session identifier
    Profile     string            // GoZen profile name
    Provider    string            // Target provider name
    ClientType  string            // Client type (claude, codex, opencode)
    ProjectPath string            // Bound project directory path

    // Request data
    Method      string            // HTTP method
    Path        string            // Request path
    Headers     http.Header       // HTTP headers
    Body        []byte            // Raw request body

    // Parsed body (for convenience)
    Model       string            // Model name from request
    Messages    []Message         // Parsed messages array

    // Middleware can store data here for use in ProcessResponse
    Metadata    map[string]interface{}
}
```

### ResponseContext

```go
type ResponseContext struct {
    // Original request context
    Request      *RequestContext

    // Response data
    StatusCode   int
    Headers      http.Header
    Body         []byte

    // Parsed usage (if available)
    InputTokens  int
    OutputTokens int
}
```

## Creating Your First Middleware

### Step 1: Create the Go file

Create a new file `my_middleware.go`:

```go
package main

import (
    "encoding/json"
    "log"
)

// MyConfig holds configuration for this middleware
type MyConfig struct {
    Option1 string `json:"option1"`
    Option2 int    `json:"option2"`
}

// MyMiddleware is an example middleware
type MyMiddleware struct {
    config MyConfig
    logger *log.Logger
}

func (m *MyMiddleware) Name() string {
    return "my-middleware"
}

func (m *MyMiddleware) Version() string {
    return "1.0.0"
}

func (m *MyMiddleware) Description() string {
    return "My custom middleware"
}

func (m *MyMiddleware) Priority() int {
    return 100 // User middleware should use 100+
}

func (m *MyMiddleware) Init(config json.RawMessage) error {
    if len(config) > 0 {
        if err := json.Unmarshal(config, &m.config); err != nil {
            return err
        }
    }
    m.logger = log.Default()
    return nil
}

func (m *MyMiddleware) ProcessRequest(ctx *RequestContext) (*RequestContext, error) {
    m.logger.Printf("[my-middleware] Processing request for session %s", ctx.SessionID)

    // Modify request here
    // Example: Add a custom header
    ctx.Headers.Set("X-Custom-Header", "value")

    return ctx, nil
}

func (m *MyMiddleware) ProcessResponse(ctx *ResponseContext) (*ResponseContext, error) {
    m.logger.Printf("[my-middleware] Processing response with status %d", ctx.StatusCode)

    // Modify response here

    return ctx, nil
}

func (m *MyMiddleware) Close() error {
    return nil
}

// Export for Go plugin system
var Middleware MyMiddleware
```

### Step 2: Build as a Go Plugin

```bash
go build -buildmode=plugin -o my-middleware.so my_middleware.go
```

> **Note**: Go plugins are only supported on Linux and macOS. Windows is not supported.

### Step 3: Configure in zen.json

Add your middleware to the configuration:

```json
{
  "middleware": {
    "enabled": true,
    "middlewares": [
      {
        "name": "my-middleware",
        "enabled": true,
        "source": "local",
        "path": "/path/to/my-middleware.so",
        "config": {
          "option1": "value1",
          "option2": 42
        }
      }
    ]
  }
}
```

## Built-in Middleware

GoZen includes several built-in middleware:

| Name | Priority | Description |
|------|----------|-------------|
| `context-injection` | 10 | Auto-injects .cursorrules, CLAUDE.md into requests |
| `session-memory` | 15 | Cross-session intelligence and memory |
| `request-logger` | 20 | Logs all requests and responses |
| `orchestration` | 50 | Multi-model orchestration (voting, chain, review) |

### Enabling Built-in Middleware

```json
{
  "middleware": {
    "enabled": true,
    "middlewares": [
      {
        "name": "context-injection",
        "enabled": true,
        "source": "builtin"
      },
      {
        "name": "request-logger",
        "enabled": true,
        "source": "builtin",
        "config": {
          "log_body": true,
          "max_body_size": 1000
        }
      }
    ]
  }
}
```

## Remote Plugins

GoZen supports loading plugins from remote URLs. The URL should point to a JSON manifest:

```json
{
  "name": "my-middleware",
  "version": "1.0.0",
  "description": "My awesome middleware",
  "author": "Your Name",
  "downloads": {
    "linux-amd64": "https://example.com/my-middleware-linux-amd64.so",
    "darwin-amd64": "https://example.com/my-middleware-darwin-amd64.so",
    "darwin-arm64": "https://example.com/my-middleware-darwin-arm64.so"
  },
  "checksums": {
    "linux-amd64": "sha256-hash-here",
    "darwin-amd64": "sha256-hash-here",
    "darwin-arm64": "sha256-hash-here"
  }
}
```

Configure in zen.json:

```json
{
  "middleware": {
    "enabled": true,
    "middlewares": [
      {
        "name": "my-middleware",
        "enabled": true,
        "source": "remote",
        "url": "https://example.com/my-middleware/manifest.json"
      }
    ]
  }
}
```

## Best Practices

### 1. Priority Selection

- **0-49**: Reserved for system-level middleware
- **50-99**: Reserved for built-in middleware
- **100-199**: Recommended for user middleware
- **200+**: Low-priority middleware (runs last)

### 2. Error Handling

```go
func (m *MyMiddleware) ProcessRequest(ctx *RequestContext) (*RequestContext, error) {
    // Return error to abort the request
    if someCondition {
        return nil, fmt.Errorf("request blocked: reason")
    }

    // Return context to continue
    return ctx, nil
}
```

### 3. Modifying Request Body

```go
func (m *MyMiddleware) ProcessRequest(ctx *RequestContext) (*RequestContext, error) {
    // Clone context to avoid modifying original
    newCtx := ctx.Clone()

    // Parse body
    var bodyMap map[string]interface{}
    if err := json.Unmarshal(newCtx.Body, &bodyMap); err != nil {
        return ctx, nil // Return original on error
    }

    // Modify
    bodyMap["custom_field"] = "value"

    // Re-marshal
    newBody, err := json.Marshal(bodyMap)
    if err != nil {
        return ctx, nil
    }
    newCtx.Body = newBody

    return newCtx, nil
}
```

### 4. Storing Data Between Request and Response

```go
func (m *MyMiddleware) ProcessRequest(ctx *RequestContext) (*RequestContext, error) {
    // Store data in metadata
    ctx.Metadata["start_time"] = time.Now()
    return ctx, nil
}

func (m *MyMiddleware) ProcessResponse(ctx *ResponseContext) (*ResponseContext, error) {
    // Retrieve data from request metadata
    if startTime, ok := ctx.Request.Metadata["start_time"].(time.Time); ok {
        duration := time.Since(startTime)
        log.Printf("Request took %v", duration)
    }
    return ctx, nil
}
```

### 5. Security Considerations

- **Validate all input**: Never trust data from the request
- **Avoid storing secrets**: Don't log or store API keys, tokens, etc.
- **Use checksums**: Always verify remote plugin checksums
- **Limit permissions**: Only request necessary access

## API Reference

### Web API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/middleware` | GET | List all middleware |
| `/api/v1/middleware` | PUT | Update middleware config |
| `/api/v1/middleware/{name}` | GET | Get middleware details |
| `/api/v1/middleware/{name}/enable` | POST | Enable middleware |
| `/api/v1/middleware/{name}/disable` | POST | Disable middleware |
| `/api/v1/middleware/reload` | POST | Reload all middleware |

## Troubleshooting

### Plugin won't load

1. Check that the plugin was built with the same Go version as GoZen
2. Verify the plugin exports a `Middleware` variable
3. Check file permissions (must be executable)

### Middleware not executing

1. Verify `middleware.enabled` is `true` in config
2. Check that the specific middleware entry has `enabled: true`
3. Check the logs for initialization errors

### Changes not taking effect

Run `POST /api/v1/middleware/reload` or restart the daemon.

## Example Middleware

### Request Rate Limiter

```go
type RateLimiter struct {
    requests map[string][]time.Time
    limit    int
    window   time.Duration
    mu       sync.Mutex
}

func (m *RateLimiter) ProcessRequest(ctx *RequestContext) (*RequestContext, error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    key := ctx.SessionID
    now := time.Now()

    // Clean old requests
    var recent []time.Time
    for _, t := range m.requests[key] {
        if now.Sub(t) < m.window {
            recent = append(recent, t)
        }
    }

    if len(recent) >= m.limit {
        return nil, fmt.Errorf("rate limit exceeded")
    }

    m.requests[key] = append(recent, now)
    return ctx, nil
}
```

### Response Cache

```go
type ResponseCache struct {
    cache map[string][]byte
    ttl   time.Duration
    mu    sync.RWMutex
}

func (m *ResponseCache) ProcessRequest(ctx *RequestContext) (*RequestContext, error) {
    key := m.cacheKey(ctx)

    m.mu.RLock()
    if cached, ok := m.cache[key]; ok {
        m.mu.RUnlock()
        ctx.Metadata["cached_response"] = cached
        return ctx, nil
    }
    m.mu.RUnlock()

    return ctx, nil
}

func (m *ResponseCache) ProcessResponse(ctx *ResponseContext) (*ResponseContext, error) {
    if cached, ok := ctx.Request.Metadata["cached_response"].([]byte); ok {
        ctx.Body = cached
        return ctx, nil
    }

    // Store in cache
    key := m.cacheKey(ctx.Request)
    m.mu.Lock()
    m.cache[key] = ctx.Body
    m.mu.Unlock()

    return ctx, nil
}
```

## Contributing

We welcome community middleware contributions! To submit your middleware:

1. Create a GitHub repository for your middleware
2. Include a `manifest.json` for remote loading
3. Add documentation and examples
4. Open an issue on the GoZen repository to be listed in the community registry

## Support

- GitHub Issues: https://github.com/anthropics/gozen/issues
- Documentation: https://gozen.dev/docs/middleware
