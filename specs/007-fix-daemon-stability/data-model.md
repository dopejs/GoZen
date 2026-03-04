# Data Model: Fix Daemon Proxy Stability

**Feature**: 007-fix-daemon-stability
**Date**: 2026-03-04

## Entities

### 1. OpenCCConfig (existing — no schema change)

**File**: `internal/config/config.go:788-807`

| Field | Type | JSON | Change |
|-------|------|------|--------|
| ProxyPort | `int` | `proxy_port,omitempty` | No schema change. Behavioral change: `EnsureProxyPort()` persists `DefaultProxyPort` (19841) if value is 0 on first daemon start. |

**Note**: No `CurrentConfigVersion` bump required — the JSON schema is unchanged. The fix is behavioral: ensure the port is always written to config so `omitempty` doesn't cause it to be omitted.

### 2. RequestRecord (modified field type)

**File**: `internal/proxy/request_monitor.go:10-26`

| Field | Old Type | New Type | JSON | Change |
|-------|----------|----------|------|--------|
| DurationMs | `time.Duration` | `int64` | `duration_ms` | Field renamed from `Duration` to `DurationMs`, type changed from `time.Duration` to `int64`. Value is in milliseconds. |

**Assignment points** (all in `internal/proxy/server.go`):
- Line 789: streaming response record
- Line 810: non-streaming without usage
- Line 869: non-streaming with usage

All change from `Duration: duration` to `DurationMs: duration.Milliseconds()`.

### 3. ProviderAttempt (modified field type)

**File**: `internal/proxy/request_monitor.go:30-37`

| Field | Old Type | New Type | JSON | Change |
|-------|----------|----------|------|--------|
| DurationMs | `time.Duration` | `int64` | `duration_ms` | Same change as RequestRecord. |

**Assignment point** (`internal/proxy/server.go`):
- `buildFailoverChain()` line 907: `Duration: f.Elapsed` → `DurationMs: f.Elapsed.Milliseconds()`

### 4. MatchLog (modified field type)

**File**: `internal/bot/matcher.go:33-42`

| Field | Old Type | New Type | JSON | Change |
|-------|----------|----------|------|--------|
| DurationMs | `time.Duration` | `int64` | `duration_ms` | Same change pattern. |

**Assignment points** (all in `internal/bot/matcher.go`):
- Line 296 in `recordMatchLog()`: `log.Duration = duration` → `log.DurationMs = duration.Milliseconds()`

### 5. Lock File (new — filesystem entity, not config)

**Path**: `~/.zen/zend.lock`

| Attribute | Value |
|-----------|-------|
| Location | `filepath.Join(config.ConfigDirPath(), "zend.lock")` |
| Type | Advisory file lock via `syscall.Flock()` |
| Lifecycle | Created on first daemon start attempt. Lock acquired during startup, released on process exit (kernel auto-release). File persists on disk but lock state is kernel-managed. |
| Permissions | 0600 |

### 6. Settings API Response (extended)

**File**: `internal/web/api_settings.go:11-17`

| Field | Type | JSON | Change |
|-------|------|------|--------|
| ProxyPort | `int` | `proxy_port` | New field added to `settingsResponse`. Read-only — not in `settingsRequest`. |

## State Transitions

### Daemon Startup (port binding)

```
[Start] → GetProxyPort() → EnsureProxyPort() → [Port Persisted]
                                                      ↓
                                              TryBindPort(port)
                                              ↓              ↓
                                          [Success]    [Port Busy]
                                              ↓              ↓
                                        [Running]    IdentifyProcess(port)
                                                     ↓              ↓
                                               [Is Zen]      [Not Zen]
                                                  ↓              ↓
                                            KillStale()    [Error: "port X occupied by <name>"]
                                                  ↓
                                           RetryBind(port)
                                           ↓          ↓
                                       [Success]  [Still Busy]
                                          ↓           ↓
                                      [Running]   [Fatal Error]
```

### Wrapper Recovery (client re-launch)

```
[zen startViaDaemon] → ensureDaemonRunning() → [Daemon Alive]
                                                      ↓
                                               runClient(claude)
                                               ↓              ↓
                                          [Exit 0]    [Non-zero Exit]
                                              ↓              ↓
                                          [Done]    isConnectionError(stderr)?
                                                    ↓              ↓
                                                [Yes]          [No]
                                                  ↓              ↓
                                          isDaemonAlive()?   [Forward Exit Code]
                                          ↓          ↓
                                      [Alive]    [Dead]
                                        ↓           ↓
                                   [Forward    AcquireLock() → ensureDaemonRunning()
                                    Exit Code]      ↓
                                              ReleaseLock() → runClient(claude) [retry]
                                                              ↓
                                                        [Forward Exit Code]
```

## Validation Rules

| Entity | Rule | Error |
|--------|------|-------|
| ProxyPort | Must be 1024-65535 | "port must be between 1024 and 65535" |
| ProxyPort | Must not be 0 (unset) at runtime | Auto-set to DefaultProxyPort (19841) |
| `zen config set` key | Must be in whitelist (`proxy_port`) | "unknown config key: <key>" |
| `zen config set` value | Must be parseable for the key type | "invalid value for <key>: <value>" |
