# Feature Specification: Fix Proxy Port Stability

**Feature Branch**: `002-fix-proxy-port`
**Created**: 2026-02-28
**Status**: Draft
**Input**: User description: "反向代理服务的端口号一直变化，应该严格遵守用户的设置，端口号不要变化"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Stable Proxy Port on Legacy Path (Priority: P1)

As a user who has configured a specific proxy port, I expect the
reverse proxy to always start on that exact port. Currently, the
legacy (non-daemon) startup path ignores my configured port and
binds to a random OS-assigned port every time, meaning clients
cannot reliably connect.

**Why this priority**: This is the core bug. Every user on the
legacy code path is affected — the proxy port changes on every
restart, breaking client connections and automation.

**Independent Test**: Start GoZen via the legacy proxy path,
verify the proxy listens on the configured port (default 19841
or custom value). Restart and confirm the same port is used.

**Acceptance Scenarios**:

1. **Given** a user has not customized the proxy port,
   **When** the legacy proxy starts,
   **Then** it MUST listen on the default port (19841).

2. **Given** a user has set a custom proxy port in config,
   **When** the legacy proxy starts,
   **Then** it MUST listen on that exact custom port.

3. **Given** the configured port is already in use by another
   process, **When** the legacy proxy attempts to start,
   **Then** it MUST fail with a clear error message indicating
   the port conflict — not silently fall back to a random port.

---

### User Story 2 - Consistent Port Across Startup Modes (Priority: P2)

As a user, I expect the proxy port to be the same whether I start
GoZen via the daemon or the legacy (foreground) path. The two
startup modes MUST NOT produce different port behavior.

**Why this priority**: Behavioral consistency between startup modes
prevents user confusion and ensures documentation accuracy.

**Independent Test**: Start GoZen via daemon mode, note the proxy
port. Stop the daemon. Start GoZen via legacy mode, confirm the
proxy port is identical.

**Acceptance Scenarios**:

1. **Given** a default configuration with no custom proxy port,
   **When** the user starts GoZen via daemon mode and then via
   legacy mode, **Then** both MUST use port 19841.

2. **Given** a custom proxy port configured,
   **When** the user starts GoZen via either mode,
   **Then** both MUST use the same custom port.

---

### Edge Cases

- What happens when the configured port is 0? The system MUST
  treat port 0 as "use default" (19841), not as "assign random."
- What happens when the configured port is out of valid range
  (e.g., negative, or > 65535)? The system MUST reject it with
  a clear error at startup.
- What happens when the user changes the port in config while
  the proxy is running? The running proxy keeps its current port;
  the new port takes effect on next restart.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The legacy proxy startup path MUST read the proxy
  port from the user's configuration (default: 19841).
- **FR-002**: The legacy proxy MUST NOT use OS-assigned random
  ports (i.e., binding to port 0 is prohibited).
- **FR-003**: When the configured port is unavailable, the system
  MUST exit with a descriptive error message including the port
  number and the underlying OS error.
- **FR-004**: Both daemon and legacy startup paths MUST use the
  same port resolution logic: read from config, fall back to
  default (19841) if unset or zero.
- **FR-005**: The proxy port value displayed in logs and client
  environment variables MUST match the actual listening port.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The proxy port remains identical across 10
  consecutive restarts using the legacy startup path.
- **SC-002**: A user-configured custom port is respected 100% of
  the time — no random port assignment occurs.
- **SC-003**: When the port is occupied, the user sees an error
  within 1 second that names the conflicting port.
- **SC-004**: Daemon and legacy startup modes produce the same
  proxy port for the same configuration.

### Assumptions

- The default proxy port (19841) is documented and well-known to
  users. Changing this default is out of scope.
- The daemon startup path already works correctly; only the legacy
  path needs fixing.
- Port validation (range check) is a minor hardening improvement,
  not the primary fix.
