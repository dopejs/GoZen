# Quickstart: Manual Provider Unavailability Marking

**Feature**: 015-mark-provider-unavailable
**Date**: 2026-03-07

## Overview

This feature adds the ability to manually mark API providers as unavailable, preventing the proxy from routing requests to them. Markings can be set via Web UI or CLI with three duration types: today, this month, or permanent.

## Key Integration Points

### Config Layer (`internal/config/`)
1. Add `UnavailableMarking` struct and `DisabledProviders` map to `OpenCCConfig`
2. Bump config version 13 → 14
3. Add Store methods: `DisableProvider()`, `EnableProvider()`, `GetDisabledProviders()`, `IsProviderDisabled()`
4. Update `DeleteProvider()` to clean up disabled entries

### Proxy Layer (`internal/proxy/`)
1. Add `isProviderDisabled(name)` helper that calls `config.IsProviderDisabled()` (lazy, no sync)
2. Modify `tryProviders()` in `server.go` to skip disabled providers via config check
3. Add pre-check in `ServeHTTP()` for "all providers unavailable" → return 503 error
4. Ensure scenario fallback path respects disabled markings

### CLI Layer (`cmd/`)
1. Add `cmd/disable.go` — `zen disable <provider> [--today|--month|--permanent]` + `--list`
2. Add `cmd/enable.go` — `zen enable <provider>`
3. Register both in `cmd/root.go`

### Web API Layer (`internal/web/`)
1. Add `POST /api/v1/providers/{name}/disable` endpoint
2. Add `POST /api/v1/providers/{name}/enable` endpoint
3. Add `GET /api/v1/providers/disabled` endpoint
4. Extend existing provider list response with `disabled` field

### Web Frontend (`web/src/`)
1. Add disable/enable toggle to provider list or detail view
2. Add disabled status badge/indicator
3. Show disabled providers in health monitoring view

## Development Flow

```bash
# 1. Start dev environment
./scripts/dev.sh restart

# 2. Run tests
go test ./internal/config/... ./internal/proxy/... ./internal/web/...

# 3. Test CLI commands
go run . disable my-provider --today
go run . disable --list
go run . enable my-provider

# 4. Test Web UI
./scripts/dev.sh web    # Start Vite dev server
# Open http://localhost:29840

# 5. Test proxy behavior
# Mark a provider disabled, send a request through proxy, verify it's skipped
```

## TDD Approach

Per Constitution Principle I, write tests first:

1. **Config tests**: `internal/config/config_test.go`
   - Test UnavailableMarking.IsExpired() for all duration types
   - Test Store.DisableProvider/EnableProvider round-trip
   - Test config migration v13 → v14 (old format parse, round-trip, field preservation)
   - Test cleanup on DeleteProvider

2. **Proxy tests**: `internal/proxy/server_test.go`
   - Test tryProviders skips disabled providers (via config.IsProviderDisabled check)
   - Test all-disabled returns 503 error
   - Test scenario fallback with disabled providers
   - Test all providers disabled → error (no silent attempt)

3. **Web API tests**: `internal/web/api_providers_test.go`
   - Test disable/enable endpoints
   - Test provider list includes disabled status
   - Test disabled list endpoint

4. **CLI tests**: `cmd/cmd_test.go`
   - Test disable/enable command parsing
   - Test --list flag
   - Test error on nonexistent provider
