# Research & Technical Decisions

**Feature**: 017-proxy-stability
**Date**: 2026-03-08

## Overview

This document records technical decisions, research findings, and rationale for implementation choices in the daemon proxy stability improvements.

## Decision 1: Panic Recovery Strategy

**Decision**: Use middleware-based panic recovery with `defer recover()` pattern

**Rationale**:
- Go's standard pattern for panic recovery in HTTP handlers
- Isolates panics to individual requests without terminating the daemon
- Allows logging of stack traces via `debug.Stack()` for diagnostics
- Minimal performance overhead (defer is cheap in Go 1.14+)

**Alternatives Considered**:
- Process-level supervision (systemd, launchd): Requires external tooling, doesn't prevent user-facing errors
- Per-goroutine recovery: Too granular, misses handler-level panics
- No recovery: Unacceptable - single panic crashes entire daemon

**Implementation Pattern**:
```go
func RecoverMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                stack := debug.Stack()
                logger.Error("panic_recovered", map[string]interface{}{
                    "error": err,
                    "stack": string(stack),
                    "path":  r.URL.Path,
                })
                http.Error(w, "Internal Server Error", 500)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

**References**:
- Go blog: Error handling and Go (https://go.dev/blog/error-handling-and-go)
- Effective Go: Recover (https://go.dev/doc/effective_go#recover)

---

## Decision 2: Metrics Storage

**Decision**: In-memory metrics with ring buffer for request history, no persistence

**Rationale**:
- Spec explicitly excludes persistent metrics storage (Out of Scope)
- In-memory is sufficient for real-time monitoring and recent history
- Avoids complexity of database, file I/O, or external systems
- Metrics reset on daemon restart is acceptable for stability monitoring
- Ring buffer (e.g., 1000 recent requests) provides enough data for P50/P95/P99 calculation

**Alternatives Considered**:
- SQLite persistence: Out of scope, adds complexity
- Prometheus integration: Out of scope, requires external dependency
- File-based logs: Inefficient for real-time queries, log rotation complexity

**Implementation Pattern**:
```go
type Metrics struct {
    mu sync.RWMutex

    // Counters
    TotalRequests   int64
    SuccessRequests int64
    FailedRequests  int64

    // Latency tracking (ring buffer)
    latencies       []time.Duration
    latencyIndex    int

    // Error tracking
    ErrorsByProvider map[string]int64
    ErrorsByType     map[string]int64

    // Resource peaks
    PeakGoroutines int
    PeakMemoryMB   uint64
}
```

**References**:
- Go sync package: https://pkg.go.dev/sync
- Percentile calculation: Use sort on ring buffer snapshot

---

## Decision 3: Structured Logging Format

**Decision**: JSON-formatted logs with stdlib `log` package, no external logging framework

**Rationale**:
- Spec requires JSON format with timestamp, level, event type, context fields
- Stdlib `log` package is sufficient with custom formatting
- No need for log levels beyond Info/Warn/Error (Debug can be added later)
- Avoids external dependencies (zerolog, zap, logrus)
- Easy to parse with `jq` or log aggregation tools

**Alternatives Considered**:
- Structured logging libraries (zap, zerolog): Adds dependency, overkill for current needs
- Plain text logs: Not machine-parseable, harder to query
- syslog integration: Platform-specific, adds complexity

**Implementation Pattern**:
```go
type StructuredLogger struct {
    logger *log.Logger
}

func (l *StructuredLogger) Info(event string, fields map[string]interface{}) {
    entry := map[string]interface{}{
        "timestamp": time.Now().UTC().Format(time.RFC3339),
        "level":     "INFO",
        "event":     event,
    }
    for k, v := range fields {
        entry[k] = v
    }
    data, _ := json.Marshal(entry)
    l.logger.Println(string(data))
}
```

**Selective Logging** (from clarification):
- Log ALL errors (status >= 400)
- Log requests exceeding 1 second latency
- Skip logging for successful fast requests (reduces log volume)

**References**:
- JSON logging best practices: https://www.loggly.com/ultimate-guide/json-logging-best-practices/

---

## Decision 4: Concurrency Limiting

**Decision**: Semaphore-based limiting with 100 concurrent request limit (from clarification)

**Rationale**:
- Simple, predictable behavior: requests block when limit reached
- Go channels provide natural semaphore implementation
- 100 concurrent limit matches high-load test scenario (SC-004)
- No need for complex rate limiting algorithms (token bucket, leaky bucket)
- Backpressure naturally propagates to clients

**Alternatives Considered**:
- Adaptive throttling: Too complex, hard to test, unpredictable behavior
- No limit with resource monitoring: Reactive rather than proactive, can still crash
- Per-provider limits: Unnecessary complexity, global limit is sufficient

**Implementation Pattern**:
```go
type Limiter struct {
    semaphore chan struct{}
}

func NewLimiter(max int) *Limiter {
    return &Limiter{semaphore: make(chan struct{}, max)}
}

func (l *Limiter) Acquire() { l.semaphore <- struct{}{} }
func (l *Limiter) Release() { <-l.semaphore }

// Usage in handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    s.limiter.Acquire()
    defer s.limiter.Release()
    // ... handle request
}
```

**References**:
- Go concurrency patterns: https://go.dev/blog/pipelines
- Semaphore pattern: https://en.wikipedia.org/wiki/Semaphore_(programming)

---

## Decision 5: HTTP Connection Pool Configuration

**Decision**: Unified transport with tuned connection pool limits

**Rationale**:
- Default `http.Transport` has unlimited connections, can exhaust file descriptors
- Tuned limits prevent resource exhaustion while allowing reasonable concurrency
- `MaxIdleConns: 100`, `MaxIdleConnsPerHost: 20`, `MaxConnsPerHost: 50` balance reuse and limits
- `IdleConnTimeout: 90s` matches typical keep-alive timeouts
- `ForceAttemptHTTP2: true` improves performance with Anthropic API

**Configuration**:
```go
&http.Transport{
    Proxy:                 http.ProxyFromEnvironment,
    DialContext:           (&net.Dialer{Timeout: 10*time.Second, KeepAlive: 30*time.Second}).DialContext,
    ForceAttemptHTTP2:     true,
    MaxIdleConns:          100,
    MaxIdleConnsPerHost:   20,
    MaxConnsPerHost:       50,
    IdleConnTimeout:       90 * time.Second,
    TLSHandshakeTimeout:   10 * time.Second,
    ResponseHeaderTimeout: 30 * time.Second,
    ExpectContinueTimeout: 1 * time.Second,
}
```

**Cleanup Strategy**:
- Call `client.CloseIdleConnections()` when cache is invalidated
- Call on daemon shutdown to release resources
- Track clients in `ProxyServer.Close()` to avoid double-close

**References**:
- Go http.Transport docs: https://pkg.go.dev/net/http#Transport
- Connection pool tuning: https://www.loginradius.com/blog/engineering/tune-the-go-http-client-for-high-performance/

---

## Decision 6: Auto-Restart Mechanism

**Decision**: Exponential backoff with max 5 restarts, implemented in daemon start wrapper

**Rationale**:
- Prevents crash loops while allowing recovery from transient issues
- Exponential backoff (1s, 2s, 4s, 8s, 16s) gives system time to stabilize
- Max 5 restarts prevents infinite loops, forces manual intervention for persistent issues
- Implemented at process level (not systemd/launchd) for cross-platform consistency

**Implementation Pattern**:
```go
func runDaemonWithRestart() error {
    maxRestarts := 5
    restartCount := 0

    for {
        err := runDaemon()
        if err == nil {
            return nil // Normal exit
        }

        restartCount++
        if restartCount >= maxRestarts {
            return fmt.Errorf("daemon failed after %d restarts: %w", maxRestarts, err)
        }

        backoff := time.Duration(restartCount) * time.Second
        logger.Warn("daemon_crashed_restarting", map[string]interface{}{
            "restart_count": restartCount,
            "backoff_sec":   backoff.Seconds(),
            "error":         err.Error(),
        })
        time.Sleep(backoff)
    }
}
```

**Alternatives Considered**:
- systemd/launchd supervision: Platform-specific, requires user configuration
- Unlimited restarts: Risk of crash loop consuming resources
- No auto-restart: Forces manual intervention, poor UX

**References**:
- Exponential backoff: https://en.wikipedia.org/wiki/Exponential_backoff

---

## Decision 7: Goroutine Leak Detection

**Decision**: Periodic monitoring with baseline comparison, stack dump on anomaly

**Rationale**:
- `runtime.NumGoroutine()` provides current count
- Compare against baseline (initial count) to detect growth
- Threshold: 2x baseline AND >100 goroutines indicates leak
- `runtime.Stack(buf, true)` dumps all goroutine stacks for diagnosis
- Check every 1 minute (balance between detection speed and overhead)

**Implementation Pattern**:
```go
func (d *Daemon) StartGoroutineMonitor(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Done()

    baseline := runtime.NumGoroutine()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            current := runtime.NumGoroutine()
            if current > baseline*2 && current > 100 {
                logger.Warn("goroutine_leak_detected", map[string]interface{}{
                    "baseline": baseline,
                    "current":  current,
                })

                buf := make([]byte, 1<<20) // 1MB buffer
                stackSize := runtime.Stack(buf, true)
                logger.Debug("goroutine_stacks", map[string]interface{}{
                    "stacks": string(buf[:stackSize]),
                })
            }
            baseline = current // Update baseline to current
        }
    }
}
```

**References**:
- Go runtime package: https://pkg.go.dev/runtime
- Goroutine leak detection: https://github.com/uber-go/goleak

---

## Decision 8: Request Timeout Strategy

**Decision**: 120-second timeout at proxy server level, context cancellation for in-flight requests

**Rationale**:
- Spec assumption: 120 seconds sufficient for long-running Claude API calls with extended thinking
- Server-level timeout (`WriteTimeout: 10*time.Minute`) prevents indefinite hangs
- Per-request context timeout allows earlier cancellation if needed
- Context cancellation propagates to upstream HTTP client

**Implementation Pattern**:
```go
func (s *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
    defer cancel()

    r = r.WithContext(ctx)

    resp, err := s.forwardRequest(r)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            http.Error(w, "Gateway Timeout", http.StatusGatewayTimeout)
            return
        }
        // ... other error handling
    }
    // ... copy response
}
```

**Server Timeouts** (already implemented in buffer):
- `ReadHeaderTimeout: 15s` - Prevents Slowloris attacks
- `ReadTimeout: 2min` - Header + body read timeout
- `WriteTimeout: 10min` - Aligned with client timeout, allows long streaming
- `IdleTimeout: 90s` - Keep-alive timeout

**References**:
- Go context package: https://pkg.go.dev/context
- HTTP timeouts: https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/

---

## Decision 9: Health Check Endpoint Design

**Decision**: `/api/v1/daemon/health` with three-tier status (healthy/degraded/unhealthy)

**Rationale**:
- Three-tier status provides nuance: healthy (all good), degraded (warning), unhealthy (critical)
- Include runtime metrics (goroutines, memory) for diagnostics
- Include provider health status for upstream visibility
- Fast response (<100ms) even under load (SC-008)

**Status Logic**:
- **Healthy**: All metrics within thresholds, all providers available
- **Degraded**: Goroutines >1000 OR memory >500MB OR some providers failing
- **Unhealthy**: All providers failing OR critical resource exhaustion

**Response Schema**:
```json
{
  "status": "healthy|degraded|unhealthy",
  "version": "3.0.1",
  "uptime_seconds": 86400,
  "goroutines": 42,
  "memory": {
    "alloc_bytes": 52428800,
    "sys_bytes": 104857600,
    "heap_alloc_bytes": 52428800,
    "heap_objects": 123456,
    "num_gc": 10
  },
  "active_sessions": 5,
  "health_check_enabled": true,
  "health_check_running": true,
  "providers": [
    {
      "name": "provider-a",
      "status": "healthy",
      "last_check": "2026-03-08T10:30:00Z",
      "latency_ms": 150,
      "success_rate": 0.998,
      "check_count": 100,
      "fail_count": 2
    }
  ]
}
```

**References**:
- Health check patterns: https://microservices.io/patterns/observability/health-check-api.html

---

## Decision 10: Metrics Endpoint Design

**Decision**: `/api/v1/daemon/metrics` with request stats, latency percentiles, error breakdowns, resource peaks

**Rationale**:
- Separate from health check (different use case: trending vs status)
- Percentiles calculated from ring buffer snapshot (sort + index)
- Error grouping by provider and type enables root cause analysis
- Peak tracking shows historical highs (useful for capacity planning)

**Response Schema**:
```json
{
  "total_requests": 10000,
  "success_requests": 9990,
  "failed_requests": 10,
  "success_rate": 0.999,
  "latency_p50_ms": 45,
  "latency_p95_ms": 120,
  "latency_p99_ms": 250,
  "errors_by_provider": {
    "provider-a": 5,
    "provider-b": 5
  },
  "errors_by_type": {
    "timeout": 3,
    "connection_refused": 2,
    "rate_limit": 5
  },
  "peak_goroutines": 150,
  "peak_memory_mb": 320
}
```

**Percentile Calculation**:
```go
func (m *Metrics) GetPercentile(p float64) time.Duration {
    m.mu.RLock()
    defer m.mu.RUnlock()

    if len(m.latencies) == 0 {
        return 0
    }

    sorted := make([]time.Duration, len(m.latencies))
    copy(sorted, m.latencies)
    sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

    index := int(float64(len(sorted)) * p)
    if index >= len(sorted) {
        index = len(sorted) - 1
    }
    return sorted[index]
}
```

**References**:
- Percentile calculation: https://en.wikipedia.org/wiki/Percentile

---

## Summary

All technical decisions are based on:
1. **Simplicity**: Stdlib-first, no external dependencies
2. **Testability**: Clear interfaces, table-driven tests
3. **Performance**: In-memory metrics, efficient data structures
4. **Observability**: Structured logs, comprehensive metrics, health checks
5. **Reliability**: Panic recovery, auto-restart, resource limits

No unresolved clarifications remain. Ready for Phase 1 (Design & Contracts).
