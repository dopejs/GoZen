# Research: Fix Daemon Proxy Stability

**Feature**: 007-fix-daemon-stability
**Date**: 2026-03-04

## R1: File Locking for Daemon Startup Coordination

**Decision**: Use `syscall.Flock()` with `LOCK_EX|LOCK_NB` on `~/.zen/zend.lock`

**Rationale**:
- Standard Go `syscall` package — no external dependencies needed
- Advisory file locks are automatically released by the kernel when the holding process dies (critical for crash recovery)
- Non-blocking try-lock (`LOCK_NB`) allows detecting contention immediately, then switching to blocking `LOCK_EX` to wait
- Works on both macOS and Linux with identical API

**Implementation pattern**:
```go
f, _ := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
if err == syscall.EWOULDBLOCK {
    // Another process is starting the daemon, wait for them
    syscall.Flock(int(f.Fd()), syscall.LOCK_EX) // blocking
    // After acquiring, check if daemon is already alive
}
```

**Alternatives considered**:
- PID file atomic rename: fragile, no automatic cleanup on crash
- Port binding race: narrow window, unreliable coordination
- `golang.org/x/sys/unix.Flock`: unnecessary dependency, `syscall.Flock` suffices

## R2: Port Conflict Detection — Identifying Process on Port

**Decision**: Shell out to `lsof -i :PORT -sTCP:LISTEN -t` on macOS, fall back to `/proc/net/tcp` on Linux

**Rationale**:
- `lsof` is pre-installed on macOS, reliable and well-understood
- Returns PID of the process listening on the port
- Combined with `ps -p <PID> -o comm=` to get the process name for error messages
- No Go-native cross-platform way to map port → PID without external dependencies

**Implementation pattern**:
```go
func getProcessOnPort(port int) (pid int, processName string, err error) {
    out, err := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-sTCP:LISTEN", "-t").Output()
    // Parse PID, then:
    out2, _ := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=").Output()
    processName = strings.TrimSpace(string(out2))
}
```

**Zen daemon detection**: Check if the process name contains "zen" or "gozen". Combined with flock — if the lock file is held by a live process but the port is occupied, it's likely a zen daemon. If the lock is NOT held, and the process name matches zen, it's a stale/orphaned daemon that can be killed.

**Alternatives considered**:
- Go-native `/proc/net/tcp` parsing: Linux-only, doesn't work on macOS
- `net.Listen` attempt: only tells us the port is busy, not who owns it

## R3: Duration Serialization Fix

**Decision**: Change `Duration time.Duration` field to `DurationMs int64` with `.Milliseconds()` at assignment time

**Rationale**:
- Simplest possible fix — no custom types, no `MarshalJSON`, no wrapper structs
- `time.Duration.Milliseconds()` returns `int64` — exact type match
- The field is set in exactly 3 places in `server.go` and 3 places in `matcher.go` — minimal change surface
- JSON tag stays `"duration_ms"` — no API change for consumers
- Breaking change to Go struct API (field name and type change) but this is internal — no external consumers

**Alternatives considered**:
- Custom `DurationMs` type with `MarshalJSON`: over-engineered for this use case, adds cognitive load
- Keep `time.Duration` and add `MarshalJSON` to parent structs: fragile, requires remembering to update if struct gains new fields
- `json.Number`: doesn't solve the underlying unit mismatch

## R4: Wrapper Retry / Client Re-launch Strategy

**Decision**: Wrap child process execution in a retry loop in `startViaDaemon()`. On non-zero exit from child, check if daemon is alive. If dead, restart daemon and re-launch child. Limit to 1 retry.

**Rationale**:
- Minimal change to existing flow — `startViaDaemon()` already runs the child and checks exit code
- Single retry avoids infinite loop if the client has a genuine error
- User experience: client restarts, session state may be partially lost (user re-sends last message), but no manual daemon management
- Exit code detection: check both exit code and stderr for connection-related patterns ("connection refused", "connection reset", etc.)

**Implementation pattern**:
```go
func startViaDaemon(...) error {
    for attempt := 0; attempt <= 1; attempt++ {
        ensureDaemonRunning()
        exitCode, connErr := runClient(...)
        if exitCode == 0 || !connErr {
            return nil // success or non-connection error
        }
        if attempt == 0 {
            log("Daemon appears dead, restarting...")
            // daemon restart + retry
        }
    }
}
```

**Alternatives considered**:
- Transparent proxy layer: too complex, requires running another local server
- Client-side env hook: fragile, requires separate binary
- No retry (just restart daemon, let user re-run): worse UX, user must manually re-run command

## R5: `zen config set` Subcommand Design

**Decision**: Add a `zen config set <key> <value>` Cobra subcommand with a whitelist of supported keys, starting with `proxy_port`

**Rationale**:
- Follows existing Cobra pattern in `cmd/config.go`
- Whitelist prevents arbitrary config mutation (security, validation)
- Each key has its own validation logic (e.g., port range for `proxy_port`)
- Extensible: adding a new key means adding a case to the switch statement

**Supported keys (initial)**:
- `proxy_port`: validates 1024-65535, saves via `config.SetProxyPort()`, auto-restarts daemon, warns about client processes

**Alternatives considered**:
- Generic JSON path mutation: too dangerous, no validation
- Interactive TUI editor: doesn't support scripting/automation
- Dedicated `zen config proxy-port` command: less extensible, single-purpose
