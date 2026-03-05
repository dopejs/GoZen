# Quickstart: Fix Daemon Proxy Stability

**Feature**: 007-fix-daemon-stability

## Prerequisites

- Go 1.21+
- Node.js + pnpm (for web UI changes)
- Dev environment: `./scripts/dev.sh` for dev daemon (ports 29840/29841)

## Build & Test

```sh
# Build all
go build ./...

# Run all tests
go test ./...

# Run specific package tests
go test ./internal/daemon/...
go test ./internal/proxy/...
go test ./internal/config/...
go test ./internal/bot/...

# Check coverage for affected packages
go test -cover ./internal/daemon/
go test -cover ./internal/proxy/
go test -cover ./internal/config/
go test -cover ./internal/bot/

# Web UI (for settings page changes)
cd web && pnpm install && pnpm test
```

## Dev Workflow

```sh
# Start dev daemon (separate ports, won't interfere with production)
./scripts/dev.sh

# After Go changes, rebuild and restart
./scripts/dev.sh restart

# For web UI development
./scripts/dev.sh web

# Stop dev daemon
./scripts/dev.sh stop
```

## Manual Testing

### Test 1: Port Stability
```sh
# Start daemon, note port
zen daemon start && curl -s http://127.0.0.1:19840/api/v1/daemon/status | jq .

# Stop and restart, verify same port
zen daemon stop && zen daemon start
curl -s http://127.0.0.1:19840/api/v1/daemon/status | jq .
# Port should be identical
```

### Test 2: Port Conflict Detection
```sh
# Occupy port 19841 with a non-zen process
python3 -c "import socket; s=socket.socket(); s.bind(('127.0.0.1', 19841)); s.listen(); input()" &

# Try to start daemon — should fail with error naming the process
zen daemon start
# Expected: Error message identifying python3 as the conflicting process
kill %1
```

### Test 3: Daemon Recovery
```sh
# Start a client session
zen claude

# In another terminal, kill the daemon
kill $(cat ~/.zen/zend.pid)

# The client should exit, zen should restart daemon and re-launch client
```

### Test 4: Duration Fix
```sh
# Make a request through proxy, check monitoring
curl -s http://127.0.0.1:19840/api/v1/monitoring/requests | jq '.requests[0].duration_ms'
# Should be a reasonable number like 2500 (ms), not 2500000000 (ns)
```

### Test 5: Config Set
```sh
# Set custom proxy port
zen config set proxy_port 29841
# Should restart daemon and warn about client processes

# Verify
curl -s http://127.0.0.1:19840/api/v1/settings | jq .proxy_port
# Should be 29841
```

## Key Files to Modify

| Priority | File | Change |
|----------|------|--------|
| 1 | `internal/proxy/request_monitor.go` | `Duration time.Duration` → `DurationMs int64` |
| 1 | `internal/bot/matcher.go` | Same duration fix |
| 1 | `internal/proxy/server.go` | `.Milliseconds()` at assignment |
| 2 | `internal/config/store.go` | Add `EnsureProxyPort()` |
| 2 | `internal/daemon/server.go` | Call `EnsureProxyPort()` on start |
| 2 | `internal/daemon/daemon.go` | Add file lock + port conflict detection |
| 2 | `internal/daemon/process_unix.go` | Add `getProcessOnPort()` |
| 3 | `cmd/root.go` | Wrapper retry loop in `startViaDaemon()` |
| 3 | `cmd/config.go` | Add `zen config set` subcommand |
| 4 | `internal/web/api_settings.go` | Add `proxy_port` to settings response |
| 4 | `web/src/pages/settings/tabs/GeneralSettings.tsx` | Read-only proxy port display |

## Coverage Thresholds

| Package | Threshold |
|---------|-----------|
| `internal/daemon` | ≥ 50% |
| `internal/proxy` | ≥ 80% |
| `internal/config` | ≥ 80% |
| `internal/bot` | ≥ 80% |
| `internal/web` | ≥ 80% |
| `web/` (Vitest) | ≥ 70% branch |
