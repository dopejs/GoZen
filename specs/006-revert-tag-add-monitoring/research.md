# Request Monitoring Feature - Technical Research

## 1. In-memory Ring Buffer Implementation

### Decision
Use **slice-based ring buffer with sync.RWMutex** for thread-safe request record storage.

### Rationale
- The codebase already uses this pattern successfully in `internal/proxy/logger.go` (lines 43-122)
- `StructuredLogger` implements a proven slice-based ring buffer with mutex protection:
  - `entries []LogEntry` slice with `maxEntries` capacity
  - `sync.Mutex` for write protection
  - LRU eviction: when full, keeps last 80% and removes oldest 20%
  - Simple, efficient, and battle-tested in production
- RWMutex allows concurrent reads while protecting writes
- No channel overhead or goroutine management complexity

### Alternatives Considered
- **Channel-based approach**: More complex, requires dedicated goroutine, harder to query historical data
- **Circular buffer library**: External dependency, overkill for simple use case
- **sync.Map**: No ordering guarantees, harder to implement LRU eviction

### Implementation Pattern (from logger.go)
```go
type RequestMonitor struct {
    mu         sync.RWMutex
    records    []RequestRecord
    maxRecords int
}

func (m *RequestMonitor) Add(record RequestRecord) {
    m.mu.Lock()
    defer m.mu.Unlock()

    if len(m.records) >= m.maxRecords {
        keep := m.maxRecords * 8 / 10
        copy(m.records, m.records[len(m.records)-keep:])
        m.records = m.records[:keep]
    }
    m.records = append(m.records, record)
}

func (m *RequestMonitor) GetRecent(limit int) []RequestRecord {
    m.mu.RLock()
    defer m.mu.RUnlock()
    // Return newest first
}
```

---

## 2. Request Metadata Capture Timing

### Decision
Capture request metadata in **two phases**:
1. **Initial capture in `ServeHTTP()`** (line 120): Extract request body, session ID, client type, model
2. **Provider attempt tracking in `tryProviders()`** (line 270): Record each provider attempt with timing
3. **Final recording after `copyResponse()`** (line 503): Capture final response status, tokens, cost

### Rationale
- `ServeHTTP()` has access to the original request before any modifications
- `tryProviders()` loop (lines 271-377) already tracks:
  - Provider selection order
  - Health check skips
  - Failure reasons (auth, rate limit, server error)
  - Elapsed time per provider
  - The `failures []providerFailure` slice (lines 231, 311, 323, 335, 351, 359) already captures failover history
- `copyResponse()` is called only on success, has access to final response
- `recordUsageAndMetrics()` (line 1015) already extracts tokens and calculates cost

### Alternatives Considered
- **Only capture in tryProviders()**: Misses request body details, harder to correlate
- **Only capture after copyResponse()**: Misses failover attempts and timing per provider
- **Middleware approach**: More invasive, requires restructuring existing flow

### Failover History Capture
The existing `providerFailure` struct (lines 81-86) already tracks:
```go
type providerFailure struct {
    Name       string
    StatusCode int
    Body       string
    Elapsed    time.Duration
}
```
We can reuse this pattern for the monitoring feature.

---

## 3. Existing Request Tracking Infrastructure

### Decision
**Leverage and extend existing infrastructure** rather than building from scratch.

### Existing Components

#### A. StructuredLogger (`internal/proxy/logger.go`)
- **In-memory ring buffer**: Lines 43-44 (`entries []LogEntry`, `maxEntries int`)
- **Thread-safe**: Uses `sync.Mutex` (line 38)
- **LRU eviction**: Lines 116-121
- **SQLite persistence**: Lines 41, 130-133 via `LogDB`
- **Filtering**: Lines 261-283 (`GetEntries(filter LogFilter)`)
- **Already captures**: Provider, method, path, status code, session ID, client type (lines 22-34)

#### B. UsageTracker (`internal/proxy/usage.go`)
- **Cost calculation**: Lines 63-72 (`CalculateCost()`)
- **Model pricing**: Lines 74-107 (`findPricing()`)
- **SQLite storage**: Lines 110-131 (`Record()`)
- **Already tracks**: Timestamp, session ID, provider, model, tokens, cost, latency, project path, client type (lines 11-22)

#### C. LogDB (`internal/proxy/log_db.go`)
- **SQLite persistence**: Already has `logs` and `usage` tables
- **Query support**: Filter by provider, level, status code, session ID (lines in logger.go 301-355)
- **Cross-process access**: Web UI can query logs from daemon process

#### D. Provider Metrics (`internal/proxy/metrics.go`)
- **Health tracking**: Lines 32-50 (`RecordMetric()`)
- **Aggregated stats**: Lines 52-116 (`GetProviderMetrics()`)
- **Latency history**: Lines 168-249 (`GetLatencyHistory()`)

### What's Missing for Request Monitoring
1. **Request-level details**: Model, input/output tokens, cost per request (not just session totals)
2. **Failover chain**: Which providers were tried, in what order, why each failed
3. **Request body size**: For correlation with errors
4. **Unified view**: Combining logs, usage, and metrics into a single request record

### Integration Strategy
- Extend `LogEntry` to include token/cost data OR create new `RequestRecord` type
- Add new SQLite table `request_records` for detailed monitoring
- Reuse existing `LogDB` connection and query patterns
- Web API can aggregate data from `logs`, `usage`, and `request_records` tables

---

## 4. Web UI State Management

### Decision
Use **vanilla JS with fetch polling** (every 5-10 seconds), following existing Web UI patterns.

### Rationale
- The codebase uses **no frontend framework** - all vanilla JS (see `internal/web/dist/app.js`)
- Existing pages use simple patterns:
  - Fetch API for HTTP requests
  - `setInterval()` for polling
  - Direct DOM manipulation
  - No state management library
- Example from `app.js` (lines 18-32):
  ```javascript
  async function loadSkills() {
    const list = document.getElementById('skill-list');
    list.innerHTML = '<p>Loading...</p>';
    try {
      const res = await fetch(API);
      const skills = await res.json();
      list.innerHTML = skills.map(s => renderSkillCard(s)).join('');
    } catch (e) {
      list.innerHTML = '<p>Failed to load: ' + e.message + '</p>';
    }
  }
  ```

### Existing Web UI Patterns
- **Tab navigation**: Lines 8-15 (click handlers, classList toggle)
- **Modal dialogs**: Lines 64-78 (native `<dialog>` element)
- **Form handling**: Lines 100+ (async submit, fetch POST/PUT/DELETE)
- **Error handling**: Try-catch with inline error display
- **No build step**: Pure HTML/CSS/JS served from `internal/web/dist/`

### Monitoring Page Implementation
```javascript
// Polling pattern
let pollInterval;

async function loadRecentRequests() {
  try {
    const res = await fetch('/api/v1/monitoring/requests?limit=50');
    const data = await res.json();
    renderRequestTable(data.requests);
  } catch (e) {
    console.error('Failed to load requests:', e);
  }
}

function startPolling() {
  loadRecentRequests();
  pollInterval = setInterval(loadRecentRequests, 5000);
}

function stopPolling() {
  if (pollInterval) clearInterval(pollInterval);
}
```

### Alternatives Considered
- **WebSocket/SSE**: Overkill for 5-second polling, adds complexity
- **React/Vue**: Not used in codebase, would require build tooling
- **State management library**: Unnecessary for simple list rendering

---

## 5. Cost Calculation Integration

### Decision
**Reuse existing `UsageTracker.CalculateCost()`** method and pricing infrastructure.

### Existing Implementation (`internal/proxy/usage.go`)

#### Cost Calculation (lines 63-72)
```go
func (t *UsageTracker) CalculateCost(model string, inputTokens, outputTokens int) float64 {
    pricing := t.findPricing(model)
    if pricing == nil {
        return 0
    }

    inputCost := float64(inputTokens) / 1_000_000 * pricing.InputPerMillion
    outputCost := float64(outputTokens) / 1_000_000 * pricing.OutputPerMillion
    return inputCost + outputCost
}
```

#### Pricing Data Structure (`internal/config/config.go`, lines 404-449)
```go
type ModelPricing struct {
    InputPerMillion  float64 `json:"input_per_million"`
    OutputPerMillion float64 `json:"output_per_million"`
}

var DefaultModelPricing = map[string]*ModelPricing{
    "claude-opus-4-20250514":     {InputPerMillion: 15.0, OutputPerMillion: 75.0},
    "claude-sonnet-4-20250514":   {InputPerMillion: 3.0, OutputPerMillion: 15.0},
    "claude-haiku-3-5-20241022":  {InputPerMillion: 0.80, OutputPerMillion: 4.0},
    "gpt-4o":                     {InputPerMillion: 2.5, OutputPerMillion: 10.0},
    "gpt-4o-mini":                {InputPerMillion: 0.15, OutputPerMillion: 0.6},
    // ... 40+ models with pricing
}
```

#### Smart Model Matching (lines 74-107)
- Exact match first
- Partial match (case-insensitive substring)
- Family fallback (opus/sonnet/haiku)

#### Already Used In Production
- `recordUsageAndMetrics()` in `server.go` (line 1047): `cost := tracker.CalculateCost(model, usage.InputTokens, usage.OutputTokens)`
- Web API `/api/v1/usage/summary` (lines 41-109 in `api_usage.go`)
- Budget checking system

### Integration for Request Monitoring
```go
// In request monitoring code
tracker := proxy.GetGlobalUsageTracker()
if tracker != nil {
    cost := tracker.CalculateCost(record.Model, record.InputTokens, record.OutputTokens)
    record.Cost = cost
}
```

### Pricing Data Management
- **Default pricing**: Built-in for 40+ models (Anthropic, OpenAI, DeepSeek, MiniMax)
- **Custom overrides**: Users can set via config (lines 784-806 in `config/store.go`)
- **Web API**: `/api/v1/pricing` for viewing/editing (see `internal/web/api_pricing.go`)
- **Reload support**: `tracker.ReloadPricing()` after config changes

### Alternatives Considered
- **Duplicate cost calculation logic**: Unnecessary, existing code is robust
- **External pricing API**: Adds latency and dependency
- **Hardcoded pricing**: Already have comprehensive defaults with override support

---

## Summary

### Recommended Architecture

1. **Data Storage**
   - New `RequestRecord` struct with fields: timestamp, session_id, provider, model, input_tokens, output_tokens, cost, latency_ms, status_code, failover_chain, request_size
   - In-memory ring buffer (1000 records) using slice + RWMutex pattern from `logger.go`
   - SQLite persistence via new `request_records` table in existing `LogDB`

2. **Capture Points**
   - `ServeHTTP()`: Extract request metadata (model, session, client type, body size)
   - `tryProviders()`: Track provider attempts and failures
   - `recordUsageAndMetrics()`: Calculate cost and finalize record

3. **Web API**
   - `GET /api/v1/monitoring/requests?limit=50&provider=X&session=Y`
   - Returns: Recent requests with full details (tokens, cost, failover chain, timing)
   - Reuse existing `LogDB` query patterns and JSON response helpers

4. **Web UI**
   - New monitoring page with vanilla JS + fetch polling (5s interval)
   - Table view: timestamp, session, provider(s), model, tokens, cost, latency, status
   - Filters: provider, session, time range, status code
   - No framework, follows existing `app.js` patterns

5. **Cost Calculation**
   - Reuse `UsageTracker.CalculateCost()` with existing pricing data
   - No changes needed to pricing infrastructure

### Files to Create/Modify
- **New**: `internal/proxy/request_monitor.go` (ring buffer + record type)
- **New**: `internal/web/api_monitoring.go` (Web API handlers)
- **New**: `internal/web/dist/monitoring.html` + `monitoring.js` (UI)
- **Modify**: `internal/proxy/server.go` (add monitoring calls)
- **Modify**: `internal/proxy/logdb.go` (add request_records table + queries)
- **Modify**: `internal/web/server.go` (register monitoring routes)

### Performance Considerations
- Ring buffer: O(1) append, O(n) query (acceptable for 1000 records)
- SQLite writes: Async, non-blocking (existing pattern)
- Web UI polling: 5s interval, minimal overhead
- Memory: ~100KB for 1000 records (negligible)
