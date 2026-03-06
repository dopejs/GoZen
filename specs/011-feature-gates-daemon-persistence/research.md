# Research: Feature Gates & Daemon Persistence

**Feature**: 011-feature-gates-daemon-persistence
**Date**: 2026-03-05

## Overview

This document consolidates research findings for implementing feature gates and verifying daemon persistence mechanisms.

## Feature Gates Implementation

### Decision: Flat Boolean Struct in Config

**Chosen Approach**: Add `FeatureGates` struct to `OpenCCConfig` with explicit boolean fields for each experimental feature.

```go
type FeatureGates struct {
    Bot          bool `json:"bot"`          // Bot gateway (BETA)
    Compression  bool `json:"compression"`  // Context compression (BETA)
    Middleware   bool `json:"middleware"`   // Middleware pipeline (BETA)
    Agent        bool `json:"agent"`        // Agent infrastructure (BETA)
}
```

**Rationale**:
- Simple and explicit - no dynamic feature registry needed
- Type-safe - compiler catches typos
- Easy to serialize/deserialize with standard JSON
- Aligns with existing config patterns in GoZen
- Supports only 4 features - no need for complex abstraction

**Alternatives Considered**:
- **Dynamic map[string]bool**: More flexible but loses type safety, harder to validate
- **Separate config file**: Adds complexity, violates single config file principle
- **Environment variables**: Not persistent across sessions, harder to manage

### Decision: Hidden Cobra Command

**Chosen Approach**: Use `Hidden: true` flag on Cobra command to hide from help text.

```go
var experienceCmd = &cobra.Command{
    Use:    "experience [feature]",
    Short:  "Manage experimental feature gates",
    Hidden: true, // Not shown in zen --help
    RunE:   runExperience,
}
```

**Rationale**:
- Native Cobra feature - no custom help text manipulation
- Command still accessible via `zen experience` for power users
- Follows existing patterns in GoZen codebase
- Simple to implement and maintain

**Alternatives Considered**:
- **Custom help function**: More complex, requires maintaining custom help text
- **Separate binary**: Overkill for simple feature toggle
- **TUI-only interface**: Less scriptable, harder to automate

### Decision: File-Watch Based Config Reload

**Chosen Approach**: Leverage existing `ConfigWatcher` (2-second polling) with feature gate change detection.

**Rationale**:
- GoZen already has working file-watch implementation
- Polling every 2 seconds is sufficient for config changes
- No external dependencies (avoids fsnotify)
- Portable across platforms (Windows compatible)
- Consistent with existing codebase patterns

**Alternatives Considered**:
- **SIGHUP signal**: Requires signal handling complexity, less portable to Windows
- **fsnotify library**: External dependency, GoZen deliberately avoids this
- **Manual restart**: Poor UX, violates FR-017 (auto-reload requirement)

### Decision: Structured Audit Logging

**Chosen Approach**: Log feature gate changes to `~/.zen/zend.log` with structured format.

```
[zend] 2026-03-05 10:30:45 [AUDIT] action=enable_feature_gate resource=bot user=john
```

**Rationale**:
- Uses existing daemon log infrastructure
- Structured key-value format enables log analysis
- Includes timestamp, action, resource, user context
- Follows Go structured logging best practices
- No additional dependencies

**Alternatives Considered**:
- **Separate audit log file**: Adds complexity, splits logs across files
- **Database logging**: Overkill for simple audit trail
- **Syslog integration**: Platform-specific, adds complexity

## Daemon Persistence Mechanisms

### Decision: Keep macOS launchd Configuration Unchanged

**Current Configuration**: `KeepAlive=true` in launchd plist

**Verification Result**: ✅ Sufficient - no changes needed

**Rationale**:
- `KeepAlive=true` provides automatic restart after any exit
- launchd natively handles sleep/wake cycles
- Daemon automatically resumes within seconds of wake
- Tested and proven in production use

**Research Sources**:
- Apple launchd.plist(5) man page
- launchd KeepAlive behavior documentation
- GoZen existing implementation in `daemon_darwin.go`

### Decision: Enhance Linux systemd Configuration

**Current Configuration**: `Restart=on-failure`

**Recommended Change**: `Restart=always`

**Rationale**:
- `Restart=on-failure` only restarts on non-zero exit codes
- `Restart=always` ensures restart after any termination
- Aligns Linux behavior with macOS `KeepAlive=true`
- Provides better resilience during suspend/resume cycles
- Recommended by systemd documentation for daemon services

**Impact**: One-line change in `daemon_linux.go` line 23

**Research Sources**:
- systemd.service(5) man page
- systemd Restart= directive documentation
- Best practices for user services (systemd --user)

### Decision: Process Independence Already Correct

**Current Implementation**: `Setsid: true` in `syscall.SysProcAttr`

**Verification Result**: ✅ Correct - no changes needed

**Rationale**:
- `Setsid: true` creates new session, detaches from terminal
- Double-fork pattern NOT needed when managed by init systems
- systemd documentation explicitly warns against double-fork
- launchd handles process lifecycle correctly with single fork
- PID file management is atomic and correct

**Research Sources**:
- Unix daemon best practices
- systemd Type=forking vs Type=simple documentation
- Advanced Programming in the UNIX Environment (Stevens)

### Decision: Port Persistence Already Correct

**Current Implementation**: Ports stored in config, protected during reload

**Verification Result**: ✅ Correct - no changes needed

**Rationale**:
- Ports 19840 (web) and 19841 (proxy) stored in `~/.zen/zen.json`
- Config reload protects running ports from changes (see `server.go:361-386`)
- TCP listeners remain bound during sleep/wake cycles
- No port conflicts observed in production use

## Config Schema Migration

### Decision: Version 12 → 13 (Additive Change)

**Migration Strategy**: No migration logic needed

**Rationale**:
- Adding optional `feature_gates` field (uses `omitempty`)
- Nil/missing field defaults to all features disabled
- Backward compatible - old configs work without changes
- Forward compatible - new configs readable by old versions (field ignored)

**Version Bump**:
```go
const CurrentConfigVersion = 13  // was 12
```

## Testing Strategy

### Unit Tests

**Config Package** (`internal/config/config_test.go`):
- FeatureGates struct serialization/deserialization
- GetFeatureGates/SetFeatureGates helpers
- Config version 13 parsing
- Backward compatibility with version 12

**Daemon Package** (`internal/daemon/server_test.go`):
- Config reload with feature gate changes
- Feature gate change detection logic
- Audit log output format

### Integration Tests

**Daemon Persistence** (`tests/integration/daemon_persistence_test.go`):
- Daemon survives CLI process termination
- Daemon maintains port numbers across restarts
- Config reload triggers on file changes
- Feature gate changes logged correctly

**Platform-Specific** (manual testing required):
- macOS: Sleep/wake cycle survival (5+ minute sleep)
- Linux: Suspend/resume cycle survival (systemctl suspend)
- Both: Port consistency after system reboot

## Implementation Notes

### Config Hot-Reload Behavior

Feature gate changes detected during config reload will:
1. Log warning: "feature gates changed - restart daemon to apply"
2. Log specific changes: "bot: false → true"
3. Continue serving with old feature gate values
4. Require manual daemon restart for changes to take effect

**Rationale**: Infrastructure-level features (bot, middleware, agent) require daemon restart to initialize/teardown properly. Hot-reload would add significant complexity for minimal benefit.

### Audit Log Format

```
[zend] TIMESTAMP [AUDIT] action=ACTION resource=RESOURCE user=USER
```

**Fields**:
- `action`: enable_feature_gate, disable_feature_gate
- `resource`: bot, compression, middleware, agent
- `user`: $USER environment variable

**Example**:
```
[zend] 2026-03-05 10:30:45 [AUDIT] action=enable_feature_gate resource=bot user=john
[zend] 2026-03-05 10:31:12 [AUDIT] action=disable_feature_gate resource=compression user=john
```

## Summary

All research complete. Key decisions:
1. Flat boolean struct for feature gates (simple, type-safe)
2. Hidden Cobra command with `Hidden: true` flag
3. File-watch based config reload (existing pattern)
4. Structured audit logging to daemon log
5. macOS launchd configuration unchanged (already correct)
6. Linux systemd: change `Restart=on-failure` to `Restart=always`
7. Process independence already correct (no changes)
8. Port persistence already correct (no changes)
9. Config version 12→13, no migration logic needed

Ready to proceed to Phase 1 (Design & Contracts).
