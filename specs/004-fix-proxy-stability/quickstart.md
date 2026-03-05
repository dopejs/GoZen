# Quickstart: Reverse Proxy Stability Fix

**Feature**: 004-fix-proxy-stability
**Date**: 2026-03-02

## Prerequisites

- Go 1.21+
- GoZen repository checked out on `004-fix-proxy-stability` branch

## Build & Test

```bash
# Build
go build ./...

# Run all tests
go test ./...

# Run only affected package tests
go test ./internal/proxy/... -v
go test ./cmd/... -v
go test ./internal/daemon/... -v

# Check coverage for affected packages
go test -cover ./internal/proxy/
```

## Manual Verification

### Test 1: ProxyURL propagation (US1)

1. Configure a provider with `proxy_url` in `~/.zen-dev/zen.json`:
   ```json
   {
     "providers": {
       "test-proxy": {
         "base_url": "https://api.example.com",
         "auth_token": "test-key",
         "proxy_url": "socks5://127.0.0.1:1080"
       }
     }
   }
   ```
2. Start dev daemon: `./scripts/dev.sh restart`
3. Check daemon logs for provider proxy client creation messages
4. Verify requests to this provider go through the SOCKS5 proxy

### Test 2: Daemon readiness (US2)

1. Stop any running daemon: `./scripts/dev.sh stop`
2. Start fresh: `./scripts/dev.sh`
3. Verify no ConnectionRefused errors on first request
4. Check logs confirm both ports verified before client launch

### Test 3: Error reporting (US3)

1. Configure a profile with only unreachable providers
2. Send a request through the daemon proxy
3. Verify the 502 response includes per-provider failure details
