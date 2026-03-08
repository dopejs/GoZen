# API Contracts

**Feature**: 017-proxy-stability
**Date**: 2026-03-08

## Overview

This document defines the HTTP API contracts for daemon stability monitoring endpoints. These are internal APIs exposed by the daemon for monitoring and diagnostics.

---

## Endpoint 1: Health Check

### Request

**Method**: `GET`
**Path**: `/api/v1/daemon/health`
**Headers**: None required
**Query Parameters**: None
**Body**: None

### Response

**Status Codes**:
- `200 OK`: Health check completed successfully (status may be healthy, degraded, or unhealthy)
- `500 Internal Server Error`: Health check endpoint itself failed

**Headers**:
- `Content-Type: application/json`

**Body Schema**:
```json
{
  "status": "healthy|degraded|unhealthy",
  "version": "string",
  "uptime_seconds": "integer",
  "goroutines": "integer",
  "memory": {
    "alloc_bytes": "uint64",
    "sys_bytes": "uint64",
    "heap_alloc_bytes": "uint64",
    "heap_objects": "uint64",
    "num_gc": "uint32"
  },
  "active_sessions": "integer",
  "health_check_enabled": "boolean",
  "health_check_running": "boolean",
  "providers": [
    {
      "name": "string",
      "status": "healthy|degraded|unhealthy",
      "last_check": "string (ISO 8601) or null",
      "latency_ms": "integer",
      "success_rate": "float (0.0-1.0)",
      "check_count": "integer",
      "fail_count": "integer"
    }
  ]
}
```

**Example Response** (Healthy):
```json
{
  "status": "healthy",
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
    },
    {
      "name": "provider-b",
      "status": "healthy",
      "last_check": "2026-03-08T10:30:05Z",
      "latency_ms": 200,
      "success_rate": 1.0,
      "check_count": 100,
      "fail_count": 0
    }
  ]
}
```

**Example Response** (Degraded):
```json
{
  "status": "degraded",
  "version": "3.0.1",
  "uptime_seconds": 43200,
  "goroutines": 1200,
  "memory": {
    "alloc_bytes": 524288000,
    "sys_bytes": 629145600,
    "heap_alloc_bytes": 524288000,
    "heap_objects": 500000,
    "num_gc": 50
  },
  "active_sessions": 10,
  "health_check_enabled": true,
  "health_check_running": true,
  "providers": [
    {
      "name": "provider-a",
      "status": "healthy",
      "last_check": "2026-03-08T10:30:00Z",
      "latency_ms": 150,
      "success_rate": 0.998,
      "check_count": 200,
      "fail_count": 4
    },
    {
      "name": "provider-b",
      "status": "degraded",
      "last_check": "2026-03-08T10:30:05Z",
      "latency_ms": 5000,
      "success_rate": 0.85,
      "check_count": 200,
      "fail_count": 30
    }
  ]
}
```

**Example Response** (Unhealthy):
```json
{
  "status": "unhealthy",
  "version": "3.0.1",
  "uptime_seconds": 3600,
  "goroutines": 50,
  "memory": {
    "alloc_bytes": 104857600,
    "sys_bytes": 209715200,
    "heap_alloc_bytes": 104857600,
    "heap_objects": 200000,
    "num_gc": 5
  },
  "active_sessions": 3,
  "health_check_enabled": true,
  "health_check_running": true,
  "providers": [
    {
      "name": "provider-a",
      "status": "unhealthy",
      "last_check": "2026-03-08T10:29:50Z",
      "latency_ms": 0,
      "success_rate": 0.0,
      "check_count": 50,
      "fail_count": 50
    },
    {
      "name": "provider-b",
      "status": "unhealthy",
      "last_check": "2026-03-08T10:29:55Z",
      "latency_ms": 0,
      "success_rate": 0.0,
      "check_count": 50,
      "fail_count": 50
    }
  ]
}
```

**Field Descriptions**:
- `status`: Overall daemon health. "healthy" = all good, "degraded" = warnings, "unhealthy" = critical issues
- `version`: Daemon version string (e.g., "3.0.1")
- `uptime_seconds`: Seconds since daemon started
- `goroutines`: Current number of goroutines (from `runtime.NumGoroutine()`)
- `memory.alloc_bytes`: Bytes of allocated heap objects
- `memory.sys_bytes`: Total bytes obtained from OS
- `memory.heap_alloc_bytes`: Bytes allocated on heap (same as alloc_bytes)
- `memory.heap_objects`: Number of allocated heap objects
- `memory.num_gc`: Number of completed GC cycles
- `active_sessions`: Number of active client sessions
- `health_check_enabled`: Whether provider health checking is enabled in config
- `health_check_running`: Whether health checker goroutine is currently running
- `providers[].name`: Provider name from config
- `providers[].status`: Provider health status
- `providers[].last_check`: ISO 8601 timestamp of last health check (null if never checked)
- `providers[].latency_ms`: Last check latency in milliseconds
- `providers[].success_rate`: Success rate (0.0 to 1.0)
- `providers[].check_count`: Total number of health checks performed
- `providers[].fail_count`: Number of failed health checks

**Status Determination Logic**:
- **Healthy**: `goroutines <= 1000` AND `memory.alloc_bytes <= 500MB` AND all providers healthy/degraded
- **Degraded**: `goroutines > 1000` OR `memory.alloc_bytes > 500MB` OR some providers degraded/unhealthy
- **Unhealthy**: All providers unhealthy

**Performance Requirements**:
- Response time: <100ms (SC-008)
- Must not block on slow operations
- Must be callable even under high load

---

## Endpoint 2: Metrics

### Request

**Method**: `GET`
**Path**: `/api/v1/daemon/metrics`
**Headers**: None required
**Query Parameters**: None
**Body**: None

### Response

**Status Codes**:
- `200 OK`: Metrics retrieved successfully
- `500 Internal Server Error`: Metrics endpoint failed

**Headers**:
- `Content-Type: application/json`

**Body Schema**:
```json
{
  "total_requests": "integer",
  "success_requests": "integer",
  "failed_requests": "integer",
  "success_rate": "float (0.0-1.0)",
  "latency_p50_ms": "integer",
  "latency_p95_ms": "integer",
  "latency_p99_ms": "integer",
  "errors_by_provider": {
    "provider_name": "integer"
  },
  "errors_by_type": {
    "error_type": "integer"
  },
  "peak_goroutines": "integer",
  "peak_memory_mb": "integer"
}
```

**Example Response**:
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

**Example Response** (No Requests Yet):
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

**Field Descriptions**:
- `total_requests`: Total number of requests processed since daemon start
- `success_requests`: Number of successful requests (HTTP status < 400)
- `failed_requests`: Number of failed requests (HTTP status >= 400)
- `success_rate`: Success rate (success_requests / total_requests), 0.0 if no requests
- `latency_p50_ms`: 50th percentile (median) latency in milliseconds
- `latency_p95_ms`: 95th percentile latency in milliseconds
- `latency_p99_ms`: 99th percentile latency in milliseconds
- `errors_by_provider`: Map of provider name to error count
- `errors_by_type`: Map of error type to error count
- `peak_goroutines`: Maximum goroutine count observed since daemon start
- `peak_memory_mb`: Maximum memory usage in MB observed since daemon start

**Latency Calculation**:
- Percentiles calculated from ring buffer of last 1000 request latencies
- If fewer than 1000 requests, percentiles calculated from available data
- Latencies measured from request start to response completion

**Error Types**:
- `timeout`: Request exceeded timeout threshold
- `connection_refused`: Provider connection refused
- `connection_reset`: Connection reset by peer
- `rate_limit`: Provider rate limit exceeded (429 status)
- `server_error`: Provider returned 5xx status
- `client_error`: Client sent invalid request (4xx status)
- `unknown`: Other errors

**Performance Requirements**:
- Response time: <100ms
- Percentile calculation: O(n log n) where n = ring buffer size (1000)

---

## Endpoint 3: Daemon Status (Existing)

**Note**: This endpoint already exists. Documented here for completeness.

### Request

**Method**: `GET`
**Path**: `/api/v1/daemon/status`
**Headers**: None required
**Query Parameters**: None
**Body**: None

### Response

**Status Codes**:
- `200 OK`: Status retrieved successfully

**Headers**:
- `Content-Type: application/json`

**Body Schema**:
```json
{
  "status": "running",
  "version": "string",
  "uptime": "string (human-readable)",
  "uptime_seconds": "integer",
  "proxy_port": "integer",
  "web_port": "integer",
  "active_sessions": "integer",
  "feature_gates": {
    "gate_name": "boolean"
  }
}
```

**Differences from /health**:
- `/status`: Simple uptime and port info (existing endpoint)
- `/health`: Detailed health metrics with provider status (new endpoint)

---

## Error Responses

All endpoints may return error responses in this format:

**Status Codes**:
- `405 Method Not Allowed`: Wrong HTTP method used
- `500 Internal Server Error`: Endpoint failed internally

**Body Schema**:
```json
{
  "error": "string (error message)"
}
```

**Example**:
```json
{
  "error": "method not allowed"
}
```

---

## Backward Compatibility

- New endpoints (`/health`, `/metrics`) do not affect existing endpoints
- Existing `/status` endpoint remains unchanged
- No breaking changes to existing APIs

---

## Security Considerations

- All endpoints are localhost-only (daemon binds to 127.0.0.1)
- No authentication required (local access only)
- No sensitive data exposed (metrics are operational, not user data)
- Rate limiting not required (local access, low volume)

---

## Testing Contracts

### Health Endpoint Tests

1. **Normal operation**: Returns 200 with "healthy" status
2. **High goroutines**: Returns 200 with "degraded" status when goroutines > 1000
3. **High memory**: Returns 200 with "degraded" status when memory > 500MB
4. **All providers down**: Returns 200 with "unhealthy" status
5. **Method not allowed**: Returns 405 for POST/PUT/DELETE

### Metrics Endpoint Tests

1. **No requests**: Returns 200 with zero counters and empty maps
2. **After requests**: Returns 200 with accurate counts and percentiles
3. **Error tracking**: Correctly groups errors by provider and type
4. **Peak tracking**: Correctly tracks peak goroutines and memory
5. **Method not allowed**: Returns 405 for POST/PUT/DELETE

### Performance Tests

1. **Health endpoint latency**: <100ms under normal load
2. **Health endpoint under load**: <100ms even with 100 concurrent requests
3. **Metrics endpoint latency**: <100ms with full ring buffer (1000 entries)

---

## Client Usage Examples

### cURL

```bash
# Check daemon health
curl http://localhost:19841/api/v1/daemon/health

# Get metrics
curl http://localhost:19841/api/v1/daemon/metrics

# Pretty-print with jq
curl -s http://localhost:19841/api/v1/daemon/health | jq .
```

### Go Client

```go
resp, err := http.Get("http://localhost:19841/api/v1/daemon/health")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

var health struct {
    Status    string `json:"status"`
    Goroutines int   `json:"goroutines"`
    // ... other fields
}

if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
    log.Fatal(err)
}

fmt.Printf("Daemon status: %s, goroutines: %d\n", health.Status, health.Goroutines)
```

### Monitoring Script

```bash
#!/bin/bash
# Monitor daemon health every 60 seconds

while true; do
    status=$(curl -s http://localhost:19841/api/v1/daemon/health | jq -r .status)
    goroutines=$(curl -s http://localhost:19841/api/v1/daemon/health | jq .goroutines)

    echo "$(date): Status=$status, Goroutines=$goroutines"

    if [ "$status" = "unhealthy" ]; then
        echo "ALERT: Daemon is unhealthy!"
        # Send notification
    fi

    sleep 60
done
```

---

## Summary

Two new monitoring endpoints:
1. **GET /api/v1/daemon/health**: Comprehensive health check with provider status
2. **GET /api/v1/daemon/metrics**: Request statistics and performance metrics

Both endpoints:
- Return JSON responses
- Are localhost-only
- Respond in <100ms
- Are backward compatible with existing APIs
