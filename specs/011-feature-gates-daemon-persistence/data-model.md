# Data Model: Feature Gates & Daemon Persistence

**Feature**: 011-feature-gates-daemon-persistence
**Date**: 2026-03-05

## Entities

### FeatureGates

Represents the state of experimental feature toggles.

**Fields**:
- `Bot` (bool): Enable bot gateway functionality (BETA)
- `Compression` (bool): Enable context compression (BETA)
- `Middleware` (bool): Enable middleware pipeline (BETA)
- `Agent` (bool): Enable agent infrastructure (BETA)

**Validation Rules**:
- All fields are boolean (true/false)
- No validation needed - any boolean value is valid
- Nil pointer means all features disabled (default state)

**State Transitions**:
- Initial state: nil (all features disabled)
- User enables feature: field changes from false → true
- User disables feature: field changes from true → false
- Config reload: daemon detects changes and logs them

**Storage**:
- Persisted in `~/.zen/zen.json` under `feature_gates` key
- JSON format: `{"feature_gates": {"bot": true, "compression": false, ...}}`
- Optional field (uses `omitempty` tag)

**Relationships**:
- Embedded in `OpenCCConfig` struct
- Referenced by daemon config reload logic
- Controls visibility/behavior of BETA features

**Example**:
```json
{
  "version": 13,
  "feature_gates": {
    "bot": true,
    "compression": false,
    "middleware": false,
    "agent": true
  },
  "providers": {...},
  "profiles": {...}
}
```

### OpenCCConfig (Modified)

Main configuration structure - enhanced to include feature gates.

**New Field**:
- `FeatureGates` (*FeatureGates): Experimental feature toggles (optional)

**Version Change**:
- `Version` field: 12 → 13

**Validation Rules**:
- `FeatureGates` can be nil (defaults to all disabled)
- When present, all boolean fields are valid
- No cross-field validation needed

**Migration**:
- Version 12 configs: `FeatureGates` field missing → treated as nil
- Version 13 configs: `FeatureGates` field present → parsed normally
- No migration logic needed (additive change)

### DaemonService (Conceptual)

Represents the daemon service configuration and state.

**Fields** (not stored, runtime state):
- `ServiceName` (string): "com.dopejs.zend" (macOS) or "zend.service" (Linux)
- `ExecutablePath` (string): Path to zen binary
- `WebPort` (int): 19840 (from config)
- `ProxyPort` (int): 19841 (from config)
- `PIDFile` (string): ~/.zen/zend.pid
- `LogFile` (string): ~/.zen/zend.log
- `AutoStart` (bool): true when installed as service
- `RestartPolicy` (string): "KeepAlive=true" (macOS) or "Restart=always" (Linux)

**State Transitions**:
- Not installed → Installed (via `zen daemon enable`)
- Installed → Running (via launchd/systemd)
- Running → Stopped (via `zen daemon stop`)
- Stopped → Running (auto-restart by init system)
- Running → Sleeping (system sleep)
- Sleeping → Running (system wake, <10 seconds)

**Platform-Specific**:

**macOS (launchd)**:
- Plist file: `~/Library/LaunchAgents/com.dopejs.zend.plist`
- `KeepAlive=true`: Auto-restart on any exit
- `RunAtLoad=true`: Start on login
- User agent (not system daemon)

**Linux (systemd)**:
- Unit file: `~/.config/systemd/user/zend.service`
- `Restart=always`: Auto-restart on any exit (enhanced from `on-failure`)
- `WantedBy=default.target`: Start on login
- User service (systemd --user)

## Relationships

```
OpenCCConfig
  ├── FeatureGates (1:1, optional)
  │   ├── Bot (bool)
  │   ├── Compression (bool)
  │   ├── Middleware (bool)
  │   └── Agent (bool)
  ├── Providers (1:many)
  ├── Profiles (1:many)
  └── ... (other fields)

DaemonService (runtime)
  ├── Reads: OpenCCConfig.WebPort
  ├── Reads: OpenCCConfig.ProxyPort
  ├── Watches: OpenCCConfig (file changes)
  └── Logs: FeatureGates changes
```

## Data Flow

### Feature Gate Enable/Disable

```
User: zen experience bot
  ↓
1. Load config from ~/.zen/zen.json
2. Parse OpenCCConfig (version 13)
3. Get FeatureGates (or create if nil)
4. Set FeatureGates.Bot = true
5. Write config back to ~/.zen/zen.json
6. Log audit entry to stderr
  ↓
Daemon: ConfigWatcher detects file change (2s poll)
  ↓
7. Reload config from ~/.zen/zen.json
8. Compare old vs new FeatureGates
9. Log changes to ~/.zen/zend.log
10. Continue serving (restart required for changes to apply)
```

### Daemon Persistence (Sleep/Wake)

```
System: User closes laptop
  ↓
1. macOS/Linux: System enters sleep mode
2. Daemon process: Suspended by OS
3. TCP listeners: Remain bound to ports 19840/19841
  ↓
System: User opens laptop
  ↓
4. macOS/Linux: System wakes from sleep
5. Daemon process: Resumed by OS (<10 seconds)
6. TCP listeners: Still bound to same ports
7. Daemon: Continues serving requests
```

### Daemon Persistence (Process Independence)

```
User: zen (launches Claude Code)
  ↓
1. CLI process: Starts Claude Code subprocess
2. Claude Code: Connects to daemon on localhost:19840
3. Daemon: Serves requests (independent process)
  ↓
User: kill <zen-cli-pid>
  ↓
4. CLI process: Terminated
5. Claude Code: May terminate (depends on signal)
6. Daemon: Continues running (separate session via Setsid)
7. Daemon: Still serving on ports 19840/19841
```

## Validation Rules Summary

### FeatureGates
- ✅ All boolean fields valid (no range checks)
- ✅ Nil pointer valid (means all disabled)
- ✅ Empty object `{}` valid (all false)

### OpenCCConfig (Version 13)
- ✅ `feature_gates` field optional
- ✅ Version 12 configs valid (field missing)
- ✅ Version 13 configs valid (field present)
- ✅ Unknown fields ignored (forward compatibility)

### DaemonService
- ✅ Ports 19840/19841 must not conflict with other services
- ✅ PID file must be writable
- ✅ Log file must be writable
- ✅ Executable path must exist and be executable

## Edge Cases

### Feature Gates
- **Nil FeatureGates**: Treated as all features disabled
- **Partial object**: Missing fields default to false
- **Invalid feature name**: CLI returns error with valid feature list
- **Concurrent modifications**: Last write wins (file-based locking not needed)

### Daemon Persistence
- **Port conflicts after wake**: Daemon logs error, continues on existing ports
- **Multiple daemon instances**: PID file prevents (atomic write with O_EXCL)
- **Config file deleted**: Daemon continues with in-memory config
- **Config file corrupted**: Daemon logs error, continues with old config

## Performance Characteristics

### Feature Gate Operations
- **Enable/Disable**: <100ms (file read + modify + write)
- **List**: <10ms (file read + parse)
- **Config reload**: <50ms (file read + parse + compare)

### Daemon Persistence
- **Sleep/Wake resume**: <10 seconds (OS-dependent)
- **Process restart**: <2 seconds (launchd/systemd)
- **Config reload detection**: 2 seconds (polling interval)

## Storage Format

### JSON Schema (Version 13)

```json
{
  "version": 13,
  "feature_gates": {
    "bot": false,
    "compression": false,
    "middleware": false,
    "agent": false
  },
  "default_profile": "default",
  "default_client": "claude",
  "proxy_port": 19841,
  "web_port": 19840,
  "providers": {...},
  "profiles": {...}
}
```

### File Location
- **Config**: `~/.zen/zen.json`
- **PID**: `~/.zen/zend.pid`
- **Log**: `~/.zen/zend.log`
- **macOS plist**: `~/Library/LaunchAgents/com.dopejs.zend.plist`
- **Linux unit**: `~/.config/systemd/user/zend.service`
