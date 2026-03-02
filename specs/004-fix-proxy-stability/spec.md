# Feature Specification: Reverse Proxy Stability Fix

**Feature Branch**: `004-fix-proxy-stability`
**Created**: 2026-03-02
**Status**: Draft
**Input**: User description: "反向代理在使用 zen 启动时报 ConnectionRefused，即使 profile 中有阿里和火山兜底 provider。使用 zen use <provider> 直连则正常。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Provider Network Proxy Support in Daemon Path (Priority: P1) 🎯 MVP

When a user runs `zen` to start a session through the daemon proxy, providers configured with a network proxy (ProxyURL for SOCKS5/HTTP) must use that proxy for upstream connections. Currently, the daemon path ignores per-provider ProxyURL settings, causing providers that require network proxies (common for accessing APIs from restricted networks) to fail with connection errors. This directly prevents failover from working — all providers fail, so the user sees errors even though the providers themselves have available capacity.

**Why this priority**: This is a confirmed code-level bug — the root cause of the reported ConnectionRefused. The `ProfileProxy.buildProviders()` function does not set `ProxyURL` or per-provider HTTP clients, unlike the `zen use` direct path which works correctly.

**Independent Test**: Configure a provider with `proxy_url: "socks5://..."`. Send a request through `zen` (daemon path). Verify the upstream connection goes through the configured network proxy.

**Acceptance Scenarios**:

1. **Given** a provider configured with a SOCKS5 ProxyURL, **When** a request is routed to this provider via the daemon, **Then** the upstream connection uses the SOCKS5 proxy.
2. **Given** a profile with mixed providers (some with ProxyURL, some without), **When** requests fail over between them, **Then** each provider uses its own configured proxy (or direct connection) independently.
3. **Given** a provider with an invalid ProxyURL, **When** the daemon builds the provider chain, **Then** the provider is still included but logs a warning about the proxy configuration failure.

---

### User Story 2 - Daemon Readiness Verification (Priority: P1)

The daemon startup flow must verify that BOTH the proxy port and web port are accepting connections before launching the client application. Currently, the readiness check only polls the web server port, creating a gap where the client could be launched before the proxy is ready.

**Why this priority**: Even after fixing the ProxyURL bug, a race condition in daemon readiness can still cause ConnectionRefused at the local proxy level.

**Independent Test**: Start the daemon from a cold state (no existing daemon). Measure the time between `zen` invocation and the client receiving its first successful proxy response. It should never get ConnectionRefused.

**Acceptance Scenarios**:

1. **Given** no daemon is running, **When** a user runs `zen`, **Then** the system verifies both proxy and web ports are accepting connections before launching the client.
2. **Given** a daemon that takes longer than usual to initialize, **When** the readiness check runs, **Then** it waits for both ports (up to the timeout) rather than launching early.
3. **Given** a stale daemon (PID file exists but process is dead), **When** a user runs `zen`, **Then** the system cleans up the stale state and starts a fresh daemon.

---

### User Story 3 - Improved Failover Error Reporting (Priority: P2)

When all providers in a profile fail, the error returned to the user should clearly indicate what went wrong: which providers were tried, why each failed, and what the user can do to fix it. Currently, connection errors can appear as raw "ConnectionRefused" without context.

**Why this priority**: Even after fixing the root causes, users will occasionally encounter all-providers-failed scenarios. Clear error messages reduce confusion and help users self-diagnose configuration issues.

**Independent Test**: Configure a profile where all providers are unreachable. Send a request through `zen`. Verify the error message lists each provider and its failure reason in a human-readable format.

**Acceptance Scenarios**:

1. **Given** all providers in a profile fail, **When** the proxy returns an error, **Then** the response includes each provider name and specific failure reason (connection refused, auth error, rate limited, timeout).
2. **Given** the daemon is not running or has crashed, **When** the client cannot connect, **Then** the CLI wrapper detects this and suggests running `zen daemon status` for diagnostics.
3. **Given** a provider fails with a connection error, **When** the proxy logs the failover, **Then** the log entry includes provider name, error details, time elapsed, and which provider was tried next.

---

### Edge Cases

- What happens when the daemon starts but the proxy port fails to bind (port already in use by another process)?
- How does the system handle a provider whose network proxy is down (ProxyURL points to a dead SOCKS5 server)?
- What happens when ALL providers are in health-check backoff — does the last-provider-always-tried guarantee still work?
- What happens during a config hot-reload — are active sessions disrupted when providers are rebuilt?
- What happens when a provider hangs indefinitely (neither succeeds nor fails within the 10-minute timeout)?

### Test Coverage Gaps (identified via code review)

The existing test suite has 150+ proxy-related tests with strong failover coverage (15+ failover-specific tests). The following gaps are in scope for this fix:

1. **ProxyURL propagation in daemon path** — `ProfileProxy.buildProviders()` never tested with ProxyURL config (this is the root cause bug)
2. **`buildProviders` with ProxyURL in direct path** — `cmd/root.go` reference path not tested with ProxyURL
3. **`waitForDaemonReady` proxy port check** — no test for the readiness function itself
4. **502 response body format** — status code tested but response body content not validated

Out of scope (real gaps but require significant infrastructure and don't address the reported issue):
- Concurrent failover under load (provider health mutex already handles this)
- SSE mid-stream failover (edge case not related to ConnectionRefused)
- Half-open recovery under load (isolated tests already cover the logic)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The daemon proxy MUST apply per-provider network proxy settings (ProxyURL) when building provider chains, creating dedicated HTTP clients for providers that need them.
- **FR-002**: The daemon readiness check MUST verify both the web server port AND the proxy server port are accepting TCP connections before reporting ready.
- **FR-003**: The proxy failover loop MUST attempt all configured providers before returning an error, regardless of individual provider failures (existing behavior — must be preserved).
- **FR-004**: The proxy MUST always attempt the last provider in the chain even if it is marked unhealthy (existing behavior — must be preserved).
- **FR-005**: Error responses from the proxy (when all providers fail) MUST include per-provider failure details (provider name, failure type).
- **FR-006**: The system MUST detect and recover from stale daemon state (dead process with leftover PID file) on startup.
- **FR-007**: The daemon MUST log all provider failover events with provider name, error details, and elapsed time.

### Key Entities

- **Provider**: An upstream API endpoint with optional network proxy (ProxyURL), authentication key, and health state tracking.
- **Profile**: An ordered list of providers defining the failover chain for a session.
- **Daemon (zend)**: Long-running background process hosting the proxy and web server on separate ports.
- **Session**: A unique identifier tying a client connection to a specific profile and provider chain.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A profile with 3 providers where the first 2 are unreachable successfully serves requests via the 3rd provider within 5 seconds.
- **SC-002**: Providers configured with ProxyURL in the daemon path behave identically to the `zen use` direct path — 100% feature parity for network proxy support.
- **SC-003**: Daemon startup-to-ready time (both ports verified) is under 5 seconds on a standard machine.
- **SC-004**: When all providers fail, the error response contains failure details for 100% of attempted providers.
- **SC-005**: Zero occurrences of ConnectionRefused caused by daemon lifecycle issues (stale PID, readiness race) in normal usage flow.

## Assumptions

- The user's network environment may require SOCKS5 or HTTP proxies to reach certain API endpoints (common in China for accessing foreign APIs).
- The daemon is expected to be a long-running process but may crash due to external factors (OOM, system restart).
- The 10-minute HTTP client timeout for upstream requests is appropriate and does not need adjustment.
- The existing health check backoff strategy (60s initial / 5min max for connection errors; 30min initial / 2h max for auth errors) is appropriate.
- The `zen use <provider>` direct path is correctly implemented and serves as the reference for expected provider behavior.
