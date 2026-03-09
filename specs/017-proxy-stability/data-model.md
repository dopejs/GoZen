# Data Model

**Feature**: 017-proxy-stability
**Date**: 2026-03-08

## Overview

This document defines the data structures and entities for daemon proxy stability improvements. All entities are runtime-only (in-memory), with no persistent storage.

---

## Entity 1: Health Status

**Purpose**: Represents the overall daemon health state with supporting diagnostic metrics.

**Attributes**:
- `status` (string): Overall health state - "healthy", "degraded", or "unhealthy"
- `version` (string): Daemon version (e.g., "3.0.1")
- `uptime_seconds` (int64): Seconds since daemon started
- `goroutines` (int): Current goroutine count from `runtime.NumGoroutine()`
- `memory` (MemoryStats): Memory usage metrics
- `active_sessions` (int): Number of active client sessions
- `health_check_enabled` (bool): Whether provider health checking is enabled
- `health_check_running` (bool): Whether health checker is currently running
- `providers` ([]ProviderHealth): Health status of each configured provider

**Validation Rules**:
- `status` must be one of: "healthy", "degraded", "unhealthy"
- `uptime_seconds` must be >= 0
- `goroutines` must be > 0 (at least 1 goroutine exists)
- `active_sessions` must be >= 0

**State Transitions**:
```
healthy → degraded: When goroutines > 1000 OR memory > 500MB OR some providers fail
degraded → unhealthy: When all providers fail OR critical resource exhaustion
degraded → healthy: When metrics return to normal thresholds
unhealthy → degraded: When at least one provider recovers
```

**Relationships**:
- Contains multiple `ProviderHealth` entities (one per configured provider)
- Contains one `MemoryStats` entity

---

## Entity 2: Memory Stats

**Purpose**: Detailed memory usage metrics from Go runtime.

**Attributes**:
- `alloc_bytes` (uint64): Bytes of allocated heap objects
- `sys_bytes` (uint64): Total bytes obtained from OS
- `heap_alloc_bytes` (uint64): Bytes allocated on heap (same as alloc_bytes)
- `heap_objects` (uint64): Number of allocated heap objects
- `num_gc` (uint32): Number of completed GC cycles

**Validation Rules**:
- All byte values must be >= 0
- `num_gc` must be >= 0

**Source**: `runtime.MemStats` via `runtime.ReadMemStats(&m)`

---

## Entity 3: Provider Health

**Purpose**: Health status and performance metrics for a single provider.

**Attributes**:
- `name` (string): Provider name (e.g., "provider-a")
- `status` (HealthStatus): Health status enum - "healthy", "degraded", "unhealthy"
- `last_check` (*time.Time): Timestamp of last health check (nullable)
- `latency_ms` (int): Last check latency in milliseconds
- `success_rate` (float64): Success rate (0.0 to 1.0)
- `check_count` (int): Total number of health checks performed
- `fail_count` (int): Number of failed health checks

**Validation Rules**:
- `name` must not be empty
- `status` must be one of: "healthy", "degraded", "unhealthy"
- `latency_ms` must be >= 0
- `success_rate` must be between 0.0 and 1.0 (inclusive)
- `check_count` must be >= 0
- `fail_count` must be >= 0 and <= `check_count`

**Derived Values**:
- `success_rate = (check_count - fail_count) / check_count` (0.0 if check_count == 0)

**Relationships**:
- Belongs to one `Health Status` entity

---

## Entity 4: Request Metrics

**Purpose**: Aggregated statistics about request volume, success rate, and latency distribution.

**Attributes**:
- `total_requests` (int64): Total number of requests processed
- `success_requests` (int64): Number of successful requests (status < 400)
- `failed_requests` (int64): Number of failed requests (status >= 400)
- `success_rate` (float64): Success rate (0.0 to 1.0)
- `latency_p50_ms` (int): 50th percentile latency in milliseconds
- `latency_p95_ms` (int): 95th percentile latency in milliseconds
- `latency_p99_ms` (int): 99th percentile latency in milliseconds
- `errors_by_provider` (map[string]int64): Error counts grouped by provider name
- `errors_by_type` (map[string]int64): Error counts grouped by error type
- `peak_goroutines` (int): Maximum goroutine count observed
- `peak_memory_mb` (uint64): Maximum memory usage in MB observed

**Validation Rules**:
- All request counts must be >= 0
- `success_requests + failed_requests == total_requests`
- `success_rate` must be between 0.0 and 1.0 (inclusive)
- All latency values must be >= 0
- `peak_goroutines` must be > 0
- `peak_memory_mb` must be >= 0

**Derived Values**:
- `success_rate = success_requests / total_requests` (0.0 if total_requests == 0)
- Percentiles calculated from latency ring buffer (see implementation in research.md)

**Internal State** (not exposed in API):
- `latencies` ([]time.Duration): Ring buffer of recent request latencies (max 1000)
- `latency_index` (int): Current write position in ring buffer

**Relationships**:
- None (standalone aggregation)

---

## Entity 5: Log Event

**Purpose**: Structured log entry with timestamp, level, event type, and contextual fields.

**Attributes**:
- `timestamp` (string): ISO 8601 timestamp (RFC3339 format)
- `level` (string): Log level - "INFO", "WARN", "ERROR", "DEBUG"
- `event` (string): Event type identifier (e.g., "daemon_started", "request_received", "panic_recovered")
- `fields` (map[string]interface{}): Contextual key-value pairs specific to the event

**Validation Rules**:
- `timestamp` must be valid RFC3339 format
- `level` must be one of: "INFO", "WARN", "ERROR", "DEBUG"
- `event` must not be empty
- `fields` may be empty but must not be nil

**Common Event Types**:
- `daemon_started`: PID, proxy_port, web_port, version
- `daemon_shutdown`: uptime_seconds, reason
- `request_received`: session_id, method, path, provider, duration_ms (only logged if error or duration > 1s)
- `provider_failed`: session_id, provider, error, duration_ms
- `panic_recovered`: error, stack, method, path, session_id
- `goroutine_leak_detected`: baseline, current
- `daemon_crashed_restarting`: restart_count, backoff_sec, error

**Relationships**:
- None (immutable log entries)

---

## Entity 6: Concurrency Limiter

**Purpose**: Semaphore-based concurrency control to limit simultaneous requests.

**Attributes**:
- `max_concurrent` (int): Maximum number of concurrent requests allowed (100)
- `semaphore` (chan struct{}): Buffered channel used as semaphore

**Operations**:
- `Acquire()`: Blocks until a slot is available, then acquires it
- `Release()`: Releases a slot, allowing another request to proceed

**Validation Rules**:
- `max_concurrent` must be > 0
- `semaphore` buffer size must equal `max_concurrent`

**State**:
- Current concurrency = number of items in semaphore channel
- Available slots = `max_concurrent - current_concurrency`

**Relationships**:
- Used by `ProxyServer` to limit concurrent request handling

---

## Entity 7: HTTP Transport Pool

**Purpose**: Manages HTTP connection pools with configured limits to prevent resource exhaustion.

**Attributes**:
- `max_idle_conns` (int): Maximum idle connections across all hosts (100)
- `max_idle_conns_per_host` (int): Maximum idle connections per host (20)
- `max_conns_per_host` (int): Maximum total connections per host (50)
- `idle_conn_timeout` (time.Duration): How long idle connections are kept (90s)
- `tls_handshake_timeout` (time.Duration): TLS handshake timeout (10s)
- `response_header_timeout` (time.Duration): Response header read timeout (30s)

**Validation Rules**:
- All connection counts must be > 0
- `max_idle_conns_per_host` should be <= `max_idle_conns`
- `max_idle_conns_per_host` should be <= `max_conns_per_host`
- All timeouts must be > 0

**Lifecycle**:
- Created once per `http.Client` instance
- Reused across multiple requests
- Cleaned up via `CloseIdleConnections()` on cache invalidation or shutdown

**Relationships**:
- Used by `ProxyServer` (shared client) and `Provider` (per-provider clients)

---

## Data Flow

### Request Processing Flow

```
1. Client Request
   ↓
2. Concurrency Limiter.Acquire()
   ↓
3. Panic Recovery Middleware (defer recover)
   ↓
4. ProxyServer.ServeHTTP()
   ↓
5. Select Provider (existing failover logic)
   ↓
6. Forward Request (via HTTP Transport Pool)
   ↓
7. Record Metrics (latency, success/failure, provider)
   ↓
8. Log Event (if error or slow request)
   ↓
9. Concurrency Limiter.Release()
   ↓
10. Return Response to Client
```

### Health Check Flow

```
1. Periodic Timer (every N seconds)
   ↓
2. For each Provider:
   ↓
3. Send HEAD request to provider base URL
   ↓
4. Measure latency
   ↓
5. Update Provider Health (status, latency, success_rate)
   ↓
6. Aggregate into Health Status (overall status)
   ↓
7. Available via /api/v1/daemon/health
```

### Metrics Collection Flow

```
1. Request completes
   ↓
2. Record latency in ring buffer
   ↓
3. Increment total_requests counter
   ↓
4. Increment success_requests or failed_requests
   ↓
5. If error: increment errors_by_provider[provider]
   ↓
6. If error: increment errors_by_type[error_type]
   ↓
7. Update peak_goroutines if current > peak
   ↓
8. Update peak_memory_mb if current > peak
   ↓
9. Available via /api/v1/daemon/metrics
```

### Goroutine Leak Detection Flow

```
1. Background Monitor (every 1 minute)
   ↓
2. Read current goroutine count
   ↓
3. Compare to baseline (previous count)
   ↓
4. If current > baseline * 2 AND current > 100:
   ↓
5. Log Warning Event (goroutine_leak_detected)
   ↓
6. Dump all goroutine stacks (runtime.Stack)
   ↓
7. Update baseline to current count
```

---

## Concurrency Considerations

### Thread Safety

All entities that are accessed concurrently MUST use synchronization:

- **Request Metrics**: Protected by `sync.RWMutex` (read-heavy workload)
- **Concurrency Limiter**: Thread-safe via channel semantics
- **HTTP Transport Pool**: Thread-safe by design (Go stdlib)
- **Health Status**: Read-only after construction (no locking needed)
- **Provider Health**: Updated by single health checker goroutine (no locking needed)
- **Log Event**: Immutable after creation (no locking needed)

### Lock Ordering

To prevent deadlocks, always acquire locks in this order:
1. Metrics lock (if needed)
2. Session map lock (existing, in daemon)
3. Provider lock (existing, in proxy)

### Goroutine Lifecycle

Background goroutines MUST be cancellable via context:
- Health checker goroutine: Cancelled on daemon shutdown
- Goroutine monitor: Cancelled on daemon shutdown
- Session cleanup: Cancelled on daemon shutdown

All background goroutines MUST be tracked in `sync.WaitGroup` for graceful shutdown.

---

## Memory Management

### Ring Buffer Sizing

- Latency ring buffer: 1000 entries (sufficient for P99 calculation)
- Memory overhead: ~8KB (1000 * 8 bytes per duration)
- Eviction: Circular overwrite (oldest entry replaced)

### Map Growth

- `errors_by_provider`: Bounded by number of providers (typically < 10)
- `errors_by_type`: Bounded by error types (typically < 20)
- No unbounded growth risk

### Cleanup

- HTTP connections: Closed on cache invalidation and shutdown
- Goroutines: Cancelled via context on shutdown
- Metrics: Reset on daemon restart (no persistence)

---

## Summary

All entities are designed for:
- **In-memory operation**: No database, no file I/O
- **Thread safety**: Appropriate locking for concurrent access
- **Bounded growth**: No unbounded maps or slices
- **Graceful cleanup**: Proper resource release on shutdown
- **Testability**: Clear interfaces, mockable dependencies
