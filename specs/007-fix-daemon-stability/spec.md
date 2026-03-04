# Feature Specification: Fix Daemon Proxy Stability

**Feature Branch**: `007-fix-daemon-stability`
**Created**: 2026-03-04
**Status**: Draft
**Input**: User description: "根据上面的分析结果，解决 daemon proxy 的稳定性问题。另外上面的结果中的P2我不认可，这个应该是P0，我们要保证Proxy port一定不能变化。"

## Problem Statement

GoZen's daemon proxy suffers from three critical stability issues that make the tool unreliable for daily use:

1. **Proxy port changes on every restart**: The proxy port is not pinned to a stable value. Each daemon restart may allocate a different port, breaking client sessions that were configured with the previous port.

2. **Daemon dies after system sleep/wake cycles**: macOS sends SIGTERM to the background daemon process during sleep/wake transitions. When the daemon stops, all connected clients (e.g., Claude Code) immediately lose connectivity with "Connection error", requiring manual intervention to recover.

3. **Monitoring data shows incorrect durations**: The request monitoring displays duration values in nanoseconds but labels them as milliseconds, making the monitoring dashboard unusable for performance analysis.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Stable Proxy Port Across Restarts (Priority: P0)

As a user, when the daemon restarts (whether manually or automatically), the proxy server must always bind to the same port so that connected clients are not broken by port changes.

**Why this priority**: If the proxy port changes, all existing client sessions break immediately. Even if the daemon survives sleep/wake, a port change makes recovery impossible without restarting the client. The proxy port is embedded in the client's `ANTHROPIC_BASE_URL` — any change to it is catastrophic.

**Independent Test**: Start the daemon, note the proxy port, stop and restart the daemon, verify the proxy port is identical.

**Acceptance Scenarios**:

1. **Given** the daemon is running on port P, **When** the daemon is stopped and restarted, **Then** it binds to the same port P.
2. **Given** the proxy port is occupied by a stale zen daemon process (e.g., from a previous crash or orphaned instance), **When** the daemon tries to start, **Then** it detects the occupying process is a zen daemon, kills it, and starts the new daemon on that port successfully.
7. **Given** the proxy port is occupied by a non-zen process (e.g., another application), **When** the daemon tries to start, **Then** it reports a clear error message identifying the conflicting process name and fails to start — it MUST NOT silently choose a different port.
3. **Given** a fresh installation with no prior configuration, **When** the daemon starts for the first time, **Then** the proxy port is set to a well-known default value (19841) and persisted in configuration.
4. **Given** the user has explicitly configured a custom proxy port, **When** the daemon starts, **Then** it uses the user-configured port, not a random one.
5. **Given** a user whose default port 19841 is occupied by another important application, **When** the user sets a custom proxy port via CLI (`zen config`), **Then** the daemon is automatically restarted on the new port and the user is told to restart all running `zen` client processes.
6. **Given** a user opens the Web UI settings page, **When** they view the proxy port setting, **Then** the port is displayed as read-only with a message: "To change the proxy port, use `zen config` in the terminal."

---

### User Story 2 - Automatic Daemon Recovery After System Sleep (Priority: P0)

As a user, when my computer wakes from sleep and the daemon has been killed by the operating system, any subsequent request from the client should automatically detect the dead daemon and restart it, so my workflow continues without manual intervention.

**Why this priority**: This is the primary cause of "Connection error" disruptions. Users should never need to know that the daemon was killed — recovery must be automatic and transparent.

**Independent Test**: Start a client session through the daemon via `zen`, kill the daemon process, observe that the client exits, and verify that `zen` automatically restarts the daemon and re-launches the client.

**Acceptance Scenarios**:

1. **Given** the daemon has been killed (e.g., by SIGTERM during sleep), **When** the client process exits with a connection error, **Then** the `zen` wrapper detects the exit, restarts the daemon, and re-launches the client process so the user can continue their session.
2. **Given** the daemon dies while the client is idle, **When** the client later sends a request and exits with a connection error, **Then** the `zen` wrapper restarts the daemon and re-launches the client — the user may need to re-send their last message but does not need to manually restart the daemon.
3. **Given** the daemon cannot be restarted (e.g., port permanently blocked), **When** recovery is attempted, **Then** the user sees a clear diagnostic message explaining why recovery failed and what to do.
4. **Given** the daemon is auto-restarted after being killed, **When** the restart completes, **Then** the daemon is fully functional within 5 seconds.
5. **Given** multiple `zen` wrapper processes detect the dead daemon simultaneously, **When** they all attempt to restart the daemon, **Then** only one restart proceeds (coordinated via file lock) and the others wait and verify the daemon is alive.

---

### User Story 3 - Accurate Request Duration in Monitoring (Priority: P1)

As a user viewing the monitoring dashboard, I should see correct request durations (in milliseconds) so I can assess provider performance and diagnose slow requests.

**Why this priority**: Incorrect duration data (showing hundreds of hours instead of seconds) makes the entire monitoring feature useless, but it does not block active usage of the proxy.

**Independent Test**: Make a request through the proxy, query the monitoring endpoint, verify the duration value is a reasonable number in milliseconds (e.g., 500-5000ms for a typical request, not billions).

**Acceptance Scenarios**:

1. **Given** a request that took approximately 2 seconds, **When** the monitoring data is queried, **Then** the `duration_ms` field shows a value between 1500 and 3000 (milliseconds), not nanoseconds.
2. **Given** multiple requests with varying durations, **When** the monitoring dashboard displays them, **Then** all durations are human-readable and correctly represent actual elapsed time.
3. **Given** a request that involved failover (multiple provider attempts), **When** the monitoring data is queried, **Then** each provider attempt in the failover chain also shows correct duration in milliseconds.

---

### Edge Cases

- What happens when the daemon is killed mid-request (while streaming a response)?
- What happens when multiple `zen` processes try to auto-restart the daemon simultaneously? (Coordinated via file lock — only one proceeds, others wait and verify.)
- What happens when the system wakes from sleep but network connectivity is not yet restored?
- What happens when the configured proxy port is in the ephemeral range and the OS has allocated it to another process?
- What happens when the daemon is auto-restarted but the config file has been modified during sleep?
- What happens when the user sets an invalid port number (e.g., 0, negative, above 65535, or a privileged port below 1024)?
- What happens when the user changes the proxy port while active client sessions are connected?
- What happens when the stale zen daemon on the port cannot be killed (e.g., permission denied)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST use a fixed, deterministic proxy port that does not change across daemon restarts. The default port MUST be 19841 unless explicitly overridden by the user in configuration.
- **FR-002**: The system MUST persist the proxy port in configuration on first startup so it remains stable across all future restarts.
- **FR-003**: When the configured proxy port is unavailable, the system MUST identify the process occupying it. If the occupying process is a zen daemon (stale/orphaned), the system MUST kill it and proceed to start. If the occupying process is NOT a zen daemon, the system MUST report the conflicting process name and fail to start — it MUST NOT silently fall back to a random port.
- **FR-011**: The proxy port MUST be configurable through the CLI via a generic `zen config set <key> <value>` subcommand (e.g., `zen config set proxy_port 29841`). This subcommand should be extensible for other settings in the future. The Web UI settings page MUST display the current proxy port as read-only, with a message directing the user to use the CLI to change it.
- **FR-012**: When the user changes the proxy port via CLI, the system MUST validate that the port number is within the valid range (1024-65535) before saving.
- **FR-013**: After a proxy port change via CLI, the system MUST automatically restart the daemon to apply the new port, and MUST display a message telling the user to restart all running `zen` client processes (since their `ANTHROPIC_BASE_URL` still points to the old port).
- **FR-004**: When the client process (e.g., claude) exits with a connection error, the `zen` CLI wrapper MUST detect this, check if the daemon is still running, and attempt to restart the daemon automatically. The recovery mechanism operates at the wrapper level — `zen` monitors the child process exit, not individual HTTP requests.
- **FR-005**: After automatic daemon restart, the `zen` wrapper MUST re-launch the client process so the user's session continues. The user may need to re-send their last message, but they should not need to manually restart the daemon or reconfigure anything.
- **FR-006**: Automatic daemon restart MUST complete within 5 seconds, including port binding and readiness verification.
- **FR-007**: When multiple `zen` wrapper processes detect a dead daemon simultaneously, only one restart attempt MUST proceed. Coordination MUST use a file lock mechanism (e.g., `~/.zen/zend.lock`) so that competing processes wait for the lock, then verify the daemon is already alive before attempting their own start.
- **FR-008**: The monitoring data MUST report request duration in milliseconds as indicated by the field name `duration_ms`, not in nanoseconds or any other unit. This fix MUST apply to all modules that serialize `duration_ms`, including both the proxy request monitor and the bot matcher.
- **FR-009**: The duration correction MUST apply to both the main request record and individual provider attempts in the failover chain.
- **FR-010**: The system SHOULD log daemon restart events so users can audit recovery behavior.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The proxy port remains identical across 10 consecutive daemon stop/start cycles.
- **SC-002**: After a simulated daemon kill, the `zen` wrapper detects the client exit, restarts the daemon, and re-launches the client process automatically within 10 seconds, with zero manual intervention required.
- **SC-003**: 100% of `duration_ms` values in monitoring data are within 2x of the actual wall-clock request duration (i.e., no nanosecond-to-millisecond confusion).
- **SC-004**: Users experience zero "Connection error" disruptions due to daemon death during a full workday that includes at least one system sleep/wake cycle.
- **SC-005**: When the proxy port is occupied by a non-zen process, the daemon startup fails with a diagnostic message naming the conflicting process within 3 seconds (no silent fallback to random port).
- **SC-007**: When the proxy port is occupied by a stale zen daemon, the system automatically kills it and starts successfully within 5 seconds.
- **SC-006**: User can set a custom proxy port via CLI, the daemon auto-restarts on the new port, and the user is informed to restart running client processes.

## Assumptions

- The daemon will continue to run as a background process (not a system service via launchd) by default. System service integration (launchd KeepAlive) is a future enhancement, not part of this scope.
- The auto-restart mechanism operates at the `zen` CLI wrapper level. When the client process (e.g., claude) exits with a connection error, the wrapper detects the exit, restarts the daemon, and re-launches the client. Individual HTTP request retry is not in scope — the user may need to re-send their last message after the session is re-established.
- Concurrent daemon restart attempts by multiple `zen` wrapper processes are coordinated via a file lock (`~/.zen/zend.lock`).
- The `zen config set` subcommand is a new generic mechanism for setting config values from the CLI, starting with `proxy_port` but extensible to other keys.
- The default proxy port 19841 is outside the ephemeral port range on macOS (49152-65535) and is unlikely to conflict with common services.
- The monitoring duration fix is a display/serialization concern and does not require changes to the underlying data model or storage.

## Scope

### In Scope

- Fixed proxy port enforcement
- User-configurable proxy port via `zen config set` CLI (read-only display in Web UI)
- Wrapper-level daemon liveness detection and auto-restart (client process re-launch)
- Monitoring duration unit correction (nanoseconds to milliseconds) across proxy and bot modules
- Race condition prevention for concurrent restart attempts via file lock

### Out of Scope

- launchd/systemd service integration (future enhancement)
- Daemon self-monitoring/watchdog process
- Automatic recovery of in-flight streaming requests interrupted by daemon death
- Changes to the Web UI dashboard rendering (only the data source is fixed)
