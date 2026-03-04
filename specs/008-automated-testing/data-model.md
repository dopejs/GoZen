# Data Model: Automated Testing Infrastructure

**Branch**: `008-automated-testing` | **Date**: 2026-03-04

## Entities

### 1. BaseTestConfig (Go, shared by all integration tests)

Consolidation of the three existing test config structs.

| Field | Type | Description |
|-------|------|-------------|
| BinaryPath | string | Path to built `zen` binary in temp dir |
| ConfigDir | string | Isolated config directory (`t.TempDir()/.zen`) |
| ProxyPort | int | Ephemeral TCP port for proxy |
| WebPort | int | Ephemeral TCP port for web UI |

**Lifecycle**: Created by `setupBaseTest(t)` → binary built → config dir created → ports allocated → used by test → cleaned up by `t.Cleanup()`.

### 2. MockProvider (Go, configurable mock server)

| Field | Type | Description |
|-------|------|-------------|
| Server | *httptest.Server | Underlying test HTTP server |
| URL | string | Server URL (from `Server.URL`) |
| RequestCount | atomic.Int64 | Number of requests received |
| Responses | []MockResponse | Queue of responses to return (FIFO) |
| DefaultResponse | MockResponse | Response when queue is empty |
| mu | sync.Mutex | Protects Responses slice |

### 3. MockResponse (Go, response configuration)

| Field | Type | Description |
|-------|------|-------------|
| StatusCode | int | HTTP status code to return |
| Body | string | JSON response body |
| Delay | time.Duration | Artificial latency before responding |
| Headers | map[string]string | Additional response headers |

**State transitions**:
- Queue has entries → dequeue first entry, return it
- Queue empty → return DefaultResponse
- Request counting always increments regardless

### 4. TestDaemon (Go, daemon process wrapper)

Extends BaseTestConfig with daemon lifecycle methods.

| Method | Signature | Description |
|--------|-----------|-------------|
| Start | `(t *testing.T)` | Start daemon in foreground mode, wait for readiness |
| Stop | `(t *testing.T)` | Send SIGTERM, wait for exit |
| Kill | `(t *testing.T)` | Send SIGKILL (simulate crash) |
| IsUp | `() bool` | Check if web port responds |
| PID | `() int` | Read PID from `zend.pid` file |
| WriteConfig | `(t, providers, profiles)` | Write `zen.json` with mock provider URLs |
| SendProxyRequest | `(t, body) (*http.Response, error)` | Send request through proxy port |
| GetJSON | `(t, path, result)` | GET web API endpoint, decode JSON |
| PostJSON | `(t, path, body, result)` | POST/PUT web API endpoint, decode JSON |

### 5. Testing Skill (Markdown, Claude Code command)

| Field | Location | Description |
|-------|----------|-------------|
| description | YAML frontmatter | One-line skill description |
| handoffs | YAML frontmatter (optional) | Suggested next commands |
| body | Markdown | Instructions for Claude Code to execute |
| $ARGUMENTS | Body placeholder | User-provided arguments (e.g., package name) |

## Relationships

```
BaseTestConfig
├── TestDaemon (embeds BaseTestConfig, adds lifecycle methods)
├── ProxyTestConfig (existing, will embed BaseTestConfig)
└── WebTestConfig (existing, will embed BaseTestConfig)

MockProvider
├── MockResponse (1:N, FIFO queue + default)
└── httptest.Server (1:1, wraps)

TestDaemon → MockProvider (N:M, config references mock URLs)
```

## Validation Rules

- Ports must be ephemeral (allocated via `net.Listen("tcp", "127.0.0.1:0")`)
- Config version must be set to current `CurrentConfigVersion` (currently 6)
- MockResponse.StatusCode must be valid HTTP status (100-599)
- MockResponse.Delay must be non-negative
- BinaryPath must exist and be executable after build
