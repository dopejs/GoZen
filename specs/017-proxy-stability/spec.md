# Feature Specification: Daemon Proxy Stability Improvements

**Feature Branch**: `017-proxy-stability`
**Created**: 2026-03-08
**Status**: Draft
**Input**: User description: "阅读 @~/Work/docs/gozen-dynamic-switching/ 中的文档 特别是 @~/Work/docs/gozen-dynamic-switching/06-proxy-stability.md，他是我们进行后续一切工作的前提。我们需要在进行 3.1.0 的工作前，先保证 daemon proxy 的稳定，这是 3.0.1 的重要工作。务必想尽一切方法保证 daemon proxy 的可用性和可靠性。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Continuous Reliable Service (Priority: P1)

Users rely on the daemon proxy to route all their Claude API requests without interruption. When the daemon crashes or becomes unresponsive, users lose access to Claude and must manually restart the service or switch to direct mode as an "escape hatch."

**Why this priority**: This is the foundation of user trust. If the proxy cannot stay running reliably, no other features matter. Users currently experience unexpected failures that force them to abandon proxy mode entirely.

**Independent Test**: Run the daemon for 24 hours under normal load (10-50 requests/hour). Verify the daemon remains responsive throughout, with no crashes, hangs, or memory leaks. Success means users can "set it and forget it."

**Acceptance Scenarios**:

1. **Given** daemon is running normally, **When** a single request causes a panic, **Then** the daemon recovers gracefully, logs the error, returns an error response to that request, and continues serving other requests without restart
2. **Given** daemon has been running for 24 hours, **When** checking memory usage, **Then** memory growth is less than 10% from startup baseline
3. **Given** daemon is under load, **When** checking goroutine count, **Then** goroutine count remains stable (no continuous growth indicating leaks)
4. **Given** daemon encounters an unrecoverable error, **When** the process exits, **Then** the daemon automatically restarts within 5 seconds and resumes service

---

### User Story 2 - Transparent Health Visibility (Priority: P2)

Users and system administrators need to know whether the daemon is healthy, degraded, or failing before problems impact their work. Currently, users only discover issues when requests fail, with no proactive visibility.

**Why this priority**: Proactive monitoring prevents user-facing failures. Users can check health status before starting critical work, and automated monitoring can alert administrators before users are impacted.

**Independent Test**: Query the health endpoint while the daemon is under various conditions (normal, high load, provider failures). Verify the health status accurately reflects the daemon's actual state and provides actionable diagnostic information.

**Acceptance Scenarios**:

1. **Given** daemon is running normally, **When** user queries health endpoint, **Then** status is "healthy" with current uptime, memory usage, goroutine count, and active session count
2. **Given** daemon has excessive goroutines (>1000) or high memory (>500MB), **When** user queries health endpoint, **Then** status is "degraded" with specific metrics showing the issue
3. **Given** all configured providers are failing, **When** user queries health endpoint, **Then** status is "unhealthy" with provider-specific error details
4. **Given** daemon is healthy, **When** user checks provider health details, **Then** each provider shows availability status, last check time, response time, and recent error rate

---

### User Story 3 - Observable Request Performance (Priority: P2)

Users need to understand how the daemon is performing over time - request success rates, latency percentiles, error patterns, and resource utilization trends. This helps diagnose performance issues and validate that stability improvements are working.

**Why this priority**: Metrics enable data-driven decisions. Users can identify patterns (e.g., "requests slow down after 12 hours") and validate fixes. Without metrics, stability is subjective.

**Independent Test**: Generate 1000 requests over 10 minutes with varying patterns (success, failures, different providers). Query the metrics endpoint and verify it accurately reports request counts, latency percentiles, error breakdowns, and resource peaks.

**Acceptance Scenarios**:

1. **Given** daemon has processed requests, **When** user queries metrics endpoint, **Then** response includes total requests, success count, failure count, and success rate percentage
2. **Given** daemon has processed requests with varying latencies, **When** user queries metrics endpoint, **Then** response includes P50, P95, and P99 latency values
3. **Given** requests have failed across different providers, **When** user queries metrics endpoint, **Then** response includes error counts grouped by provider and by error type
4. **Given** daemon has been running under load, **When** user queries metrics endpoint, **Then** response includes peak goroutine count and peak memory usage since startup

---

### User Story 4 - Resilient Under Load (Priority: P1)

Users may send bursts of concurrent requests or sustained high load. The daemon must handle this gracefully without crashing, hanging, or degrading to the point of unusability. Resource exhaustion should be prevented through proper limits.

**Why this priority**: Production workloads are unpredictable. A daemon that works under light load but fails under stress is unreliable. Users need confidence that the daemon won't become a bottleneck.

**Independent Test**: Send 100 concurrent requests continuously for 5 minutes. Verify all requests complete successfully (or fail gracefully with proper errors), the daemon remains responsive, and resource usage stays within acceptable bounds.

**Acceptance Scenarios**:

1. **Given** daemon is configured with 100 concurrent request limit, **When** 100 concurrent requests arrive, **Then** requests are processed up to the limit, excess requests wait in queue, and all eventually complete without daemon crash
2. **Given** daemon is processing a long-running request, **When** request exceeds timeout threshold, **Then** request is cancelled, client receives timeout error, and daemon resources are released
3. **Given** daemon is under sustained load, **When** monitoring connection pool usage, **Then** connections are properly reused and released, with no connection leaks
4. **Given** upstream provider is slow or unresponsive, **When** requests are routed to that provider, **Then** daemon enforces timeouts, fails over to healthy providers, and does not accumulate blocked goroutines

---

### User Story 5 - Structured Diagnostic Logging (Priority: P3)

When issues occur, users and developers need detailed, structured logs to diagnose root causes. Current logs may be unstructured or missing critical context, making troubleshooting difficult.

**Why this priority**: Good logging is essential for post-incident analysis and ongoing debugging. While not user-facing, it directly impacts time-to-resolution for stability issues.

**Independent Test**: Trigger various scenarios (normal requests, errors, panics, resource warnings). Review logs and verify they contain structured JSON entries with timestamps, log levels, event types, and relevant context fields.

**Acceptance Scenarios**:

1. **Given** daemon starts up, **When** reviewing logs, **Then** startup event is logged with PID, proxy port, web port, and version
2. **Given** daemon processes a request, **When** reviewing logs, **Then** request event is logged with session ID, method, path, provider used, and duration
3. **Given** a provider fails, **When** reviewing logs, **Then** error event is logged with session ID, provider name, error message, and duration
4. **Given** daemon detects resource anomalies (goroutine spike, memory growth), **When** reviewing logs, **Then** warning event is logged with specific metrics and thresholds

---

### Edge Cases

- What happens when daemon receives a malformed request that causes a panic in the request handler?
- How does the daemon behave when all providers are simultaneously unavailable?
- What happens when the daemon runs out of file descriptors or hits OS resource limits?
- How does the daemon handle rapid restarts (e.g., crash loop) without overwhelming the system?
- What happens when a client disconnects mid-stream during a long SSE response?
- How does the daemon behave when memory pressure is high but not yet at the degraded threshold?
- What happens when the health check endpoint itself becomes slow or unresponsive?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST recover from panics in request handlers without terminating the daemon process
- **FR-002**: System MUST provide a health check endpoint that reports daemon status as "healthy", "degraded", or "unhealthy"
- **FR-003**: System MUST include runtime metrics in health check: uptime, goroutine count, memory usage, active sessions
- **FR-004**: System MUST include provider health status in health check: availability, last check time, response time, error rate
- **FR-005**: System MUST provide a metrics endpoint that reports request statistics: total, success, failure, success rate
- **FR-006**: System MUST provide a metrics endpoint that reports latency percentiles: P50, P95, P99
- **FR-007**: System MUST provide a metrics endpoint that reports errors grouped by provider and by error type
- **FR-008**: System MUST provide a metrics endpoint that reports peak resource usage: goroutines, memory
- **FR-009**: System MUST log structured events in JSON format with timestamp, level, event type, and context fields
- **FR-010**: System MUST log daemon lifecycle events: startup, shutdown, restart
- **FR-011**: System MUST log request events for errors and requests exceeding 1 second latency: session ID, method, path, provider, duration
- **FR-012**: System MUST log error events: session ID, provider, error message, duration
- **FR-013**: System MUST log resource warning events when goroutine count exceeds 1000
- **FR-014**: System MUST log resource warning events when memory usage exceeds 500MB
- **FR-015**: System MUST automatically restart after unrecoverable errors with exponential backoff
- **FR-016**: System MUST limit automatic restarts to prevent crash loops (max 5 restarts)
- **FR-017**: System MUST enforce request timeouts to prevent indefinite hangs
- **FR-018**: System MUST cancel in-flight requests when timeout is exceeded
- **FR-019**: System MUST limit concurrent requests to 100 to prevent resource exhaustion
- **FR-020**: System MUST properly manage HTTP connection pools with configured limits
- **FR-021**: System MUST close idle connections when cache is invalidated or daemon shuts down
- **FR-022**: System MUST detect goroutine leaks by monitoring baseline vs current goroutine count
- **FR-023**: System MUST dump goroutine stacks when leak is detected for diagnostic purposes
- **FR-024**: System MUST handle client disconnections during streaming responses without leaking resources

### Key Entities

- **Health Status**: Represents the overall daemon health state (healthy/degraded/unhealthy) with supporting metrics
- **Provider Health**: Represents individual provider availability, response time, and error rate
- **Request Metrics**: Aggregated statistics about request volume, success rate, and latency distribution
- **Error Metrics**: Categorized error counts by provider and error type
- **Resource Metrics**: Runtime resource usage including goroutines, memory, and connection pools
- **Log Event**: Structured log entry with timestamp, level, event type, and contextual fields

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Daemon runs continuously for 24 hours without crashes or restarts under normal load (10-50 requests/hour)
- **SC-002**: Memory usage growth is less than 10% over 24 hours of continuous operation
- **SC-003**: Goroutine count remains stable (no continuous growth) over 24 hours of continuous operation
- **SC-004**: Daemon handles 100 concurrent requests for 5 minutes without crashes or hangs
- **SC-005**: 99th percentile request latency is under 100ms for successful requests
- **SC-006**: Request success rate is above 99.9% when all providers are healthy
- **SC-007**: Single request panic does not impact other concurrent requests (isolation verified)
- **SC-008**: Health check endpoint responds within 100ms even under high load
- **SC-009**: Daemon automatically recovers within 5 seconds after unrecoverable crash
- **SC-010**: All critical events (startup, errors, panics, resource warnings) appear in structured logs with complete context

## Assumptions

- Normal load is defined as 10-50 requests per hour based on typical user patterns
- High load is defined as 100 concurrent requests based on expected peak usage
- Memory degradation threshold of 500MB is appropriate for typical deployment environments
- Goroutine threshold of 1000 indicates a potential leak based on expected concurrency patterns
- Request timeout of 120 seconds is sufficient for long-running Claude API calls including extended thinking
- Connection pool limits (100 max idle, 10 per host) are appropriate for typical provider configurations
- Exponential backoff with max 5 restarts prevents crash loops while allowing recovery from transient issues

## Dependencies

- Existing daemon architecture and proxy implementation
- Existing provider configuration and failover logic
- Existing session management system
- Go runtime metrics APIs (runtime.MemStats, runtime.NumGoroutine, debug.Stack)
- HTTP server timeout configuration capabilities
- Context cancellation support for request lifecycle management

## Out of Scope

- Dynamic provider switching based on health status (deferred to v3.1.0)
- Persistent metrics storage or historical trend analysis
- External monitoring system integration (Prometheus, Grafana, etc.)
- Automated alerting or notification systems
- Performance optimization beyond stability requirements
- Load balancing algorithms or intelligent routing strategies
- Provider health check customization or configuration

## Clarifications

### Session 2026-03-08

- Q: What should be the maximum concurrent request limit? → A: 100 concurrent requests (matches high-load test scenario in SC-004)
- Q: Should the daemon log every single request, or use selective logging? → A: Log errors + requests exceeding latency threshold (e.g., >1s)
