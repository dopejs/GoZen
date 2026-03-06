# Feature Specification: Feature Gates & Daemon Persistence

**Feature Branch**: `011-feature-gates-daemon-persistence`
**Created**: 2026-03-05
**Status**: Draft
**Input**: User description: "在发布之前还有几件事情要做：
1. bot目前还没有实现出想要的效果，我希望建立一个FG的机制，默认关闭一些功能，允许通过 zen experience xxx 开启，或者 zen experience xxx -c 关闭。zen experience 这个名字不在 usage 中展示，仅内部知道就好
2. 我想知道现在是如何保障 daemon 始终运行的，并确保端口号不变的。我说的始终运行是用户进入休眠后依然保证运行（至少要在wake时及时唤醒），同时杀掉 zen 进程时不会终止掉 zen daemon。"

## Clarifications

### Session 2026-03-05

- Q: When a user upgrades GoZen (via `zen upgrade`), should previously enabled experimental features remain enabled? → A: Yes, persist enabled features across upgrades (user's explicit choices are preserved)
- Q: Should the daemon automatically restart if it crashes unexpectedly? → A: Yes, auto-restart on crash (maximize uptime and reliability)
- Q: When a user enables/disables an experimental feature via `zen experience`, should the daemon automatically reload its configuration? → A: Yes, auto-reload daemon config (changes take effect immediately without restart)
- Q: Should the system log when experimental features are enabled or disabled? → A: Yes, log to daemon log file (record timestamp, feature name, action)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Enable Experimental Features via Hidden Command (Priority: P1)

Power users and early adopters want to test experimental features (like bot gateway) before they are production-ready, without exposing these features to general users who might be confused by incomplete functionality.

**Why this priority**: This is the core feature gate mechanism that enables controlled rollout of experimental features. Without this, incomplete features like bot gateway cannot be safely included in the release.

**Independent Test**: Can be fully tested by running `zen experience bot` to enable the bot feature, verifying it appears in config, then running `zen experience bot -c` to disable it and confirming it's removed from config.

**Acceptance Scenarios**:

1. **Given** bot feature is disabled (default state), **When** user runs `zen experience bot`, **Then** bot feature is enabled in config, daemon detects change and logs warning, and user sees confirmation message instructing to restart daemon
2. **Given** bot feature is enabled, **When** user runs `zen experience bot -c`, **Then** bot feature is disabled in config, daemon detects change and logs warning, and user sees confirmation message instructing to restart daemon
3. **Given** user runs `zen experience` without arguments, **When** command executes, **Then** system displays list of available experimental features with their current status (enabled/disabled)
4. **Given** user runs `zen experience invalid-feature`, **When** command executes, **Then** system displays error message listing valid feature names
5. **Given** user runs `zen --help` or `zen help`, **When** help text is displayed, **Then** `experience` command is NOT listed in the output (hidden command)
6. **Given** multiple features are available, **When** user enables/disables features, **Then** each feature's state is persisted independently in config

---

### User Story 2 - Daemon Survives System Sleep/Wake Cycles (Priority: P1)

Users who put their computers to sleep expect the daemon to automatically resume when the system wakes up, ensuring uninterrupted service without manual intervention.

**Why this priority**: This is critical for user experience. If the daemon doesn't survive sleep/wake cycles, users will experience service interruptions and need to manually restart the daemon, which is frustrating and error-prone.

**Independent Test**: Can be tested by enabling the daemon as a system service (`zen daemon enable`), putting the computer to sleep for 5+ minutes, waking it up, and verifying the daemon is still running and responsive on the same ports.

**Acceptance Scenarios**:

1. **Given** daemon is installed as system service (launchd on macOS, systemd on Linux), **When** system goes to sleep and wakes up, **Then** daemon automatically resumes within 10 seconds of wake
2. **Given** daemon is running on ports 19840 (web) and 19841 (proxy), **When** system wakes from sleep, **Then** daemon continues using the same port numbers
3. **Given** daemon was processing requests before sleep, **When** system wakes up, **Then** daemon can immediately accept new requests without manual restart
4. **Given** daemon service is enabled, **When** system reboots, **Then** daemon automatically starts on boot with correct port configuration

---

### User Story 3 - Daemon Independence from CLI Process (Priority: P2)

Users expect the daemon to continue running even if they terminate the `zen` CLI process, ensuring background services remain available for other tools and sessions.

**Why this priority**: Important for reliability but less critical than sleep/wake survival. Users can work around this by not killing the CLI process, but proper process isolation improves system robustness.

**Independent Test**: Can be tested by starting the daemon via `zen daemon start`, finding the `zen` CLI process PID, killing it with `kill <pid>`, and verifying the daemon process (zend) continues running and responding to requests.

**Acceptance Scenarios**:

1. **Given** daemon is started via `zen daemon start`, **When** user kills the `zen` CLI process, **Then** daemon process continues running independently
2. **Given** daemon is running, **When** user runs `zen` command to launch Claude Code, **Then** killing the Claude Code process does not affect the daemon
3. **Given** daemon is running as background process, **When** user closes terminal window, **Then** daemon continues running (not tied to terminal session)
4. **Given** daemon is installed as system service, **When** service manager restarts the daemon, **Then** daemon uses the same port configuration as before

---

### Edge Cases

**Handled by Implementation**:
- Feature name case sensitivity: Use case-insensitive matching for feature names
- Config file corruption: Validate config before writing, backup before modifications
- Port conflicts after wake: Daemon should detect port conflicts and log errors
- Multiple daemon instances: PID file prevents multiple instances from starting
- Service installation without sudo: Use user-level services (launchd user agents, systemd --user)
- Daemon crash recovery: Auto-restart on crash to maximize uptime
- Feature gate audit trail: Log all enable/disable actions to daemon log file

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a hidden `zen experience` command not shown in help text or usage output
- **FR-002**: System MUST support `zen experience <feature-name>` to enable an experimental feature
- **FR-003**: System MUST support `zen experience <feature-name> -c` to disable an experimental feature
- **FR-004**: System MUST support `zen experience` (no args) to list all available features with their status
- **FR-005**: System MUST persist feature gate state in config file (`~/.zen/zen.json`)
- **FR-006**: System MUST validate feature names and display error for invalid names
- **FR-007**: System MUST support at minimum these experimental features: `bot`, `compression`, `middleware`, `agent`
- **FR-016**: System MUST preserve enabled feature gates across upgrades (user's explicit choices persist)
- **FR-017**: System MUST detect feature gate changes and log warning when configuration is reloaded; daemon restart required to apply feature gate changes
- **FR-018**: System MUST log feature gate changes (enable/disable actions) to daemon log file with timestamp and feature name
- **FR-008**: Daemon MUST automatically resume after system sleep/wake cycles when installed as system service
- **FR-009**: Daemon MUST use consistent port numbers (19840 web, 19841 proxy) across restarts and wake cycles
- **FR-010**: Daemon MUST run as independent background process, not tied to CLI process lifecycle
- **FR-011**: System MUST use launchd (macOS) or systemd (Linux) for daemon persistence
- **FR-012**: Daemon service MUST have `KeepAlive=true` (launchd) or `Restart=always` (systemd) configuration to auto-restart on crash
- **FR-013**: Daemon MUST write PID file to prevent multiple instances
- **FR-014**: System MUST provide `zen daemon enable` to install daemon as system service
- **FR-015**: System MUST provide `zen daemon disable` to uninstall daemon system service

### Key Entities

- **Feature Gate**: Represents an experimental feature that can be enabled/disabled
  - Name: Unique identifier (e.g., "bot", "compression")
  - Enabled: Boolean state (true/false)
  - Description: Human-readable description of the feature
  - Default: Default state (typically false for experimental features)

- **Daemon Service**: System-level service configuration
  - Service Name: `com.dopejs.zend` (macOS) or `zend.service` (Linux)
  - Executable Path: Path to zen binary
  - Ports: Web (19840), Proxy (19841)
  - Auto-start: Enabled on boot and wake
  - Restart Policy: Automatic restart on failure

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can enable/disable experimental features in under 5 seconds via single command
- **SC-002**: Daemon survives 100% of system sleep/wake cycles when installed as system service
- **SC-003**: Daemon maintains same port numbers across all restarts, reboots, and wake cycles
- **SC-004**: Daemon continues running when CLI process is terminated (100% independence)
- **SC-005**: Feature gate state persists across daemon restarts and system reboots
- **SC-006**: Daemon automatically starts within 30 seconds of system boot when service is enabled
- **SC-007**: Zero user-visible errors related to daemon port conflicts after wake from sleep

## Assumptions

- Users have permission to install user-level services (launchd user agents, systemd --user)
- System has launchd (macOS) or systemd (Linux) available for service management
- Port numbers 19840 and 19841 are not used by other system services
- Config file `~/.zen/zen.json` is writable by the user
- Feature gates apply globally (not per-profile or per-project)
- Experimental features are disabled by default in fresh installations
- The `zen experience` command is intentionally undocumented for internal/power user use only
- Current daemon implementation already uses launchd/systemd but may need configuration adjustments for sleep/wake reliability
