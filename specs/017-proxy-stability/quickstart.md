# Quick Start Guide

**Feature**: 017-proxy-stability
**Date**: 2026-03-08

## Overview

This guide helps developers quickly understand and work with the daemon proxy stability improvements. It covers the key components, how to test them, and common workflows.

---

## What's New

This feature adds comprehensive stability improvements to the daemon proxy:

1. **Panic Recovery**: Prevents single request panics from crashing the daemon
2. **Health Monitoring**: Real-time health status via `/api/v1/daemon/health`
3. **Metrics Collection**: Request statistics via `/api/v1/daemon/metrics`
4. **Structured Logging**: JSON-formatted logs for all critical events
5. **Auto-Restart**: Automatic recovery from unrecoverable errors
6. **Concurrency Limiting**: Prevents resource exhaustion (100 concurrent limit)
7. **Connection Pool Management**: Proper cleanup to prevent leaks
8. **Goroutine Leak Detection**: Monitors and alerts on goroutine growth

---

## Quick Test

### 1. Start Dev Daemon

```bash
cd /Users/John/Code/GoZen
./scripts/dev.sh restart
```

Dev daemon runs on:
- Proxy: `http://localhost:29841`
- Web UI: `http://localhost:29840`
- Config: `~/.zen-dev/zen.json`

### 2. Check Health

```bash
curl http://localhost:29841/api/v1/daemon/health | jq .
```

Expected output:
```json
{
  "status": "healthy",
  "version": "3.0.1",
  "uptime_seconds": 10,
  "goroutines": 42,
  "memory": {
    "alloc_bytes": 52428800,
    "sys_bytes": 104857600,
    ...
  },
  "active_sessions": 0,
  "providers": [...]
}
```

### 3. Check Metrics

```bash
curl http://localhost:29841/api/v1/daemon/metrics | jq .
```

Expected output (before any requests):
```json
{
  "total_requests": 0,
  "success_requests": 0,
  "failed_requests": 0,
  "success_rate": 0.0,
  "latency_p50_ms": 0,
  "latency_p95_ms": 0,
  "latency_p99_ms": 0,
  "errors_by_provider": {},
  "errors_by_type": {},
  "peak_goroutines": 42,
  "peak_memory_mb": 50
}
```

### 4. Send Test Request

```bash
# Send a request through the proxy
curl -X POST http://localhost:29841/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-key" \
  -d '{"model":"claude-3-5-sonnet-20241022","max_tokens":100,"messages":[{"role":"user","content":"Hello"}]}'
```

### 5. Check Metrics Again

```bash
curl http://localhost:29841/api/v1/daemon/metrics | jq .
```

Now you should see:
```json
{
  "total_requests": 1,
  "success_requests": 1,
  "failed_requests": 0,
  "success_rate": 1.0,
  "latency_p50_ms": 450,
  "latency_p95_ms": 450,
  "latency_p99_ms": 450,
  ...
}
```

---

## Key Components

### 1. Panic Recovery Middleware (`internal/httpx/recovery.go`)

**Purpose**: Catches panics in HTTP handlers and prevents daemon crashes.

**Usage**:
```go
// Wrap any http.Handler
handler := httpx.Recover(logger, "proxy", yourHandler)
```

**Test**:
```go
func TestPanicRecovery(t *testing.T) {
    handler := httpx.Recover(logger, "test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        panic("boom")
    }))

    req := httptest.NewRequest("GET", "/", nil)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)

    if rec.Code != 500 {
        t.Errorf("expected 500, got %d", rec.Code)
    }
}
```

### 2. Health Check API (`internal/daemon/api.go`)

**Purpose**: Provides real-time daemon health status.

**Endpoint**: `GET /api/v1/daemon/health`

**Implementation**:
```go
func (d *Daemon) handleDaemonHealth(w http.ResponseWriter, r *http.Request) {
    var mem runtime.MemStats
    runtime.ReadMemStats(&mem)

    status := "healthy"
    if runtime.NumGoroutine() > 1000 || mem.Alloc > 500*1024*1024 {
        status = "degraded"
    }

    writeJSON(w, http.StatusOK, daemonHealthResponse{
        Status:     status,
        Goroutines: runtime.NumGoroutine(),
        Memory:     convertMemStats(mem),
        // ...
    })
}
```

**Test**:
```go
func TestHealthEndpoint(t *testing.T) {
    d := NewDaemon("test", logger)
    req := httptest.NewRequest("GET", "/api/v1/daemon/health", nil)
    rec := httptest.NewRecorder()

    d.handleDaemonHealth(rec, req)

    if rec.Code != 200 {
        t.Fatalf("expected 200, got %d", rec.Code)
    }

    var resp daemonHealthResponse
    json.NewDecoder(rec.Body).Decode(&resp)

    if resp.Status == "" {
        t.Error("status should not be empty")
    }
}
```

### 3. Metrics Collection (`internal/daemon/metrics.go`)

**Purpose**: Tracks request statistics, latency percentiles, and errors.

**Usage**:
```go
// Create metrics collector
metrics := NewMetrics()

// Record request
start := time.Now()
// ... handle request ...
duration := time.Since(start)

metrics.RecordRequest(duration, success, provider, errorType)

// Get metrics
stats := metrics.GetStats()
```

**Test**:
```go
func TestMetricsCollection(t *testing.T) {
    m := NewMetrics()

    // Record some requests
    m.RecordRequest(50*time.Millisecond, true, "provider-a", "")
    m.RecordRequest(100*time.Millisecond, true, "provider-a", "")
    m.RecordRequest(200*time.Millisecond, false, "provider-b", "timeout")

    stats := m.GetStats()

    if stats.TotalRequests != 3 {
        t.Errorf("expected 3 requests, got %d", stats.TotalRequests)
    }
    if stats.SuccessRate < 0.66 || stats.SuccessRate > 0.67 {
        t.Errorf("expected ~0.67 success rate, got %f", stats.SuccessRate)
    }
}
```

### 4. Structured Logger (`internal/daemon/logger.go`)

**Purpose**: JSON-formatted logging for all critical events.

**Usage**:
```go
logger := NewStructuredLogger(stdLogger)

logger.Info("daemon_started", map[string]interface{}{
    "pid":        os.Getpid(),
    "proxy_port": 19841,
    "web_port":   19840,
})

logger.Error("provider_failed", map[string]interface{}{
    "provider": "provider-a",
    "error":    err.Error(),
    "duration": duration.Milliseconds(),
})
```

**Output**:
```json
{"timestamp":"2026-03-08T10:30:00Z","level":"INFO","event":"daemon_started","pid":12345,"proxy_port":19841,"web_port":19840}
{"timestamp":"2026-03-08T10:30:15Z","level":"ERROR","event":"provider_failed","provider":"provider-a","error":"connection refused","duration":5000}
```

### 5. Concurrency Limiter (`internal/proxy/limiter.go`)

**Purpose**: Limits concurrent requests to prevent resource exhaustion.

**Usage**:
```go
limiter := NewLimiter(100) // Max 100 concurrent

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    limiter.Acquire()
    defer limiter.Release()

    // Handle request
}
```

**Test**:
```go
func TestConcurrencyLimit(t *testing.T) {
    limiter := NewLimiter(2)

    // Acquire 2 slots
    limiter.Acquire()
    limiter.Acquire()

    // Third acquire should block
    done := make(chan bool)
    go func() {
        limiter.Acquire()
        done <- true
    }()

    select {
    case <-done:
        t.Error("should have blocked")
    case <-time.After(100 * time.Millisecond):
        // Expected: blocked
    }

    // Release one slot
    limiter.Release()

    // Now third acquire should succeed
    select {
    case <-done:
        // Expected: unblocked
    case <-time.After(100 * time.Millisecond):
        t.Error("should have unblocked")
    }
}
```

---

## Development Workflow

### 1. TDD Cycle (Red-Green-Refactor)

```bash
# 1. Write failing test
vim internal/daemon/metrics_test.go

# 2. Run test (should fail)
go test ./internal/daemon -run TestMetricsCollection

# 3. Implement feature
vim internal/daemon/metrics.go

# 4. Run test (should pass)
go test ./internal/daemon -run TestMetricsCollection

# 5. Refactor if needed
# 6. Run all tests
go test ./...
```

### 2. Manual Testing

```bash
# Start dev daemon
./scripts/dev.sh restart

# Watch logs (structured JSON)
tail -f ~/.zen-dev/daemon.log | jq .

# Monitor health in another terminal
watch -n 5 'curl -s http://localhost:29841/api/v1/daemon/health | jq ".status, .goroutines, .memory.alloc_bytes"'

# Send test requests
for i in {1..100}; do
  curl -X POST http://localhost:29841/v1/messages \
    -H "Content-Type: application/json" \
    -H "x-api-key: your-key" \
    -d '{"model":"claude-3-5-sonnet-20241022","max_tokens":10,"messages":[{"role":"user","content":"Hi"}]}' &
done
wait

# Check metrics
curl http://localhost:29841/api/v1/daemon/metrics | jq .
```

### 3. Load Testing

```bash
# Install vegeta (if not already installed)
go install github.com/tsenart/vegeta@latest

# Create targets file
cat > targets.txt <<EOF
POST http://localhost:29841/v1/messages
Content-Type: application/json
x-api-key: your-key

{"model":"claude-3-5-sonnet-20241022","max_tokens":10,"messages":[{"role":"user","content":"Hi"}]}
EOF

# Run load test (100 req/s for 1 minute)
vegeta attack -rate=100 -duration=60s -targets=targets.txt | vegeta report

# Monitor health during load test
watch -n 1 'curl -s http://localhost:29841/api/v1/daemon/health | jq ".status, .goroutines"'
```

### 4. Stability Testing

```bash
# 24-hour stability test
./scripts/stability-test.sh

# Or manually:
start_time=$(date +%s)
while true; do
  # Send request
  curl -s http://localhost:29841/api/v1/daemon/health > /dev/null

  # Check if daemon is still running
  if ! pgrep -f "zen.*daemon" > /dev/null; then
    echo "Daemon crashed after $(($(date +%s) - start_time)) seconds"
    exit 1
  fi

  # Check memory growth
  mem=$(curl -s http://localhost:29841/api/v1/daemon/health | jq .memory.alloc_bytes)
  echo "$(date): Memory=$mem bytes"

  sleep 60
done
```

---

## Common Tasks

### Add New Metric

1. Update `Metrics` struct in `internal/daemon/metrics.go`:
```go
type Metrics struct {
    // ... existing fields
    NewMetric int64 // Add new field
}
```

2. Update `RecordRequest` to track new metric:
```go
func (m *Metrics) RecordRequest(...) {
    // ... existing logic
    m.NewMetric++ // Update new metric
}
```

3. Update `GetStats` to include new metric:
```go
func (m *Metrics) GetStats() MetricsStats {
    return MetricsStats{
        // ... existing fields
        NewMetric: m.NewMetric,
    }
}
```

4. Update API contract in `contracts/api.md`

5. Write test in `internal/daemon/metrics_test.go`

### Add New Log Event

1. Define event in `internal/daemon/logger.go`:
```go
func (l *StructuredLogger) LogNewEvent(field1 string, field2 int) {
    l.Info("new_event", map[string]interface{}{
        "field1": field1,
        "field2": field2,
    })
}
```

2. Call from appropriate location:
```go
logger.LogNewEvent("value", 42)
```

3. Update `data-model.md` with new event type

4. Write test to verify log output

### Adjust Thresholds

Current thresholds (in `internal/daemon/api.go`):
- Goroutines: 1000 (degraded threshold)
- Memory: 500MB (degraded threshold)
- Concurrency: 100 (max concurrent requests)
- Request timeout: 120 seconds

To change:
1. Update constant/variable in code
2. Update assumption in `spec.md`
3. Update tests to match new threshold
4. Document in commit message

---

## Debugging

### Check Daemon Status

```bash
# Is daemon running?
pgrep -f "zen.*daemon"

# Check PID file
cat ~/.zen-dev/daemon.pid

# Check logs
tail -f ~/.zen-dev/daemon.log | jq .
```

### Trigger Panic (for testing recovery)

```go
// Add temporary panic trigger endpoint (dev only)
func (d *Daemon) handleTestPanic(w http.ResponseWriter, r *http.Request) {
    panic("test panic")
}

// Register in dev mode
if os.Getenv("GOZEN_DEV") == "1" {
    d.mux.HandleFunc("/api/v1/test/panic", d.handleTestPanic)
}
```

```bash
# Trigger panic
curl http://localhost:29841/api/v1/test/panic

# Daemon should recover and log panic
tail -f ~/.zen-dev/daemon.log | jq 'select(.event == "panic_recovered")'
```

### Simulate Goroutine Leak

```go
// Add temporary leak trigger (dev only)
func (d *Daemon) handleTestLeak(w http.ResponseWriter, r *http.Request) {
    for i := 0; i < 1000; i++ {
        go func() {
            time.Sleep(1 * time.Hour) // Leak: never exits
        }()
    }
    w.WriteHeader(200)
}
```

```bash
# Trigger leak
curl http://localhost:29841/api/v1/test/leak

# Wait 1 minute for monitor to detect
sleep 60

# Check logs for leak detection
tail -f ~/.zen-dev/daemon.log | jq 'select(.event == "goroutine_leak_detected")'
```

---

## Testing Checklist

Before opening PR:

- [ ] All tests pass: `go test ./...`
- [ ] Coverage maintained: `go test -cover ./internal/{daemon,proxy,web,httpx}`
- [ ] Health endpoint responds <100ms: `time curl http://localhost:29841/api/v1/daemon/health`
- [ ] Metrics endpoint responds <100ms: `time curl http://localhost:29841/api/v1/daemon/metrics`
- [ ] Panic recovery works: Trigger panic, verify daemon continues
- [ ] Concurrency limit works: Send 150 concurrent requests, verify 100 processed + 50 queued
- [ ] 24-hour stability: Run daemon for 24 hours, verify <10% memory growth
- [ ] Logs are valid JSON: `tail -f ~/.zen-dev/daemon.log | jq .`
- [ ] Dev daemon restarts cleanly: `./scripts/dev.sh restart`

---

## Resources

- **Spec**: `specs/017-proxy-stability/spec.md`
- **Plan**: `specs/017-proxy-stability/plan.md`
- **Research**: `specs/017-proxy-stability/research.md`
- **Data Model**: `specs/017-proxy-stability/data-model.md`
- **API Contracts**: `specs/017-proxy-stability/contracts/api.md`
- **Design Doc**: `~/Work/docs/gozen-dynamic-switching/06-proxy-stability.md`

---

## Next Steps

After completing this feature:
1. Run full test suite: `go test ./...`
2. Run integration tests: `go test ./tests/integration/...`
3. Verify coverage thresholds met
4. Open PR to `main` branch
5. After merge: Tag release `v3.0.1`
6. Proceed to v3.1.0 dynamic switching features
