# Research: Manual Provider Unavailability Marking

**Feature**: 015-mark-provider-unavailable
**Date**: 2026-03-07

## R-001: Data Storage Strategy for Unavailability Markings

**Decision**: Store unavailability markings as a top-level map in `OpenCCConfig` (the JSON config file at `~/.zen/zen.json`).

**Rationale**:
- Markings must persist across daemon restarts (FR-010)
- CLI must be able to set markings without the daemon running (spec edge case)
- The config store already has automatic file-change detection (`reloadIfModified`) so the daemon picks up CLI changes automatically
- Consistent with existing patterns (all persistent state lives in zen.json)
- Simple implementation: a `map[string]*UnavailableMarking` keyed by provider name

**Alternatives considered**:
- SQLite database (used for logs/metrics): Rejected — overkill for a small map; would require new read path for CLI and proxy
- Separate JSON file: Rejected — adds complexity for no benefit; config store already handles atomic writes
- In-memory only: Rejected — doesn't survive daemon restarts

## R-002: Config Version Bump Strategy

**Decision**: Bump `CurrentConfigVersion` from 13 to 14. Add `DisabledProviders map[string]*UnavailableMarking` field to `OpenCCConfig`.

**Rationale**:
- Adding a new top-level field to the JSON schema requires a version bump per Constitution Principle III
- The field is a map with `omitempty` — existing v13 configs parse correctly (empty map), and v14 configs with the new field are forward-compatible (older versions silently ignore unknown JSON keys)
- No complex migration logic needed — just bumping the version number triggers auto-save which writes the new field

**Alternatives considered**:
- Embed unavailability in ProviderConfig: Rejected — mixes static configuration (what the provider is) with dynamic runtime state (is it currently disabled)
- No version bump: Rejected — violates Constitution Principle III

## R-003: Expiration Evaluation Strategy

**Decision**: Evaluate expiration at request time using lazy evaluation (check `time.Now()` against stored expiration timestamp). No background timer needed.

**Rationale**:
- "Today" and "this month" markings are evaluated by comparing `time.Now()` to the pre-computed expiration timestamp stored in the marking
- Expiration timestamp is calculated at marking creation time: "today" → end of current day (23:59:59 local), "this month" → end of current month (last day 23:59:59 local), "permanent" → zero time (never expires)
- The `IsExpired()` method on the marking struct handles all evaluation
- SC-003 requires "cleared within 1 minute" — lazy evaluation at each request satisfies this naturally since requests are typically frequent
- No background goroutine needed to clean up expired markings; they simply have no effect once expired
- Config can optionally be cleaned up on next save (remove expired entries) but this is non-essential

**Alternatives considered**:
- Background timer/ticker to remove expired markings: Rejected — adds complexity; lazy evaluation is simpler and sufficient
- Cron-style scheduler: Rejected — overkill for local tool

## R-004: CLI Command Design

**Decision**: Add a `zen disable` / `zen enable` command pair for managing provider unavailability.

**Rationale**:
- `disable/enable` is intuitive and follows common CLI patterns (e.g., `systemctl enable/disable`)
- Concise and action-oriented, consistent with existing commands (`zen use`, `zen config`, `zen web`)
- Subcommands: `zen disable <provider> [--today|--month|--permanent]`, `zen enable <provider>`, `zen disable --list`

**Alternatives considered**:
- `zen provider disable/enable`: Rejected — too verbose; no existing `zen provider` command group
- `zen mark <provider> --unavailable`: Rejected — less intuitive verb
- `zen blacklist/whitelist`: Rejected — terminology issues

## R-005: Proxy Integration Point

**Decision**: Check `config.IsProviderDisabled(p.Name)` directly in `tryProviders()` for each provider at request time. No cached bool field on the runtime Provider struct. For all-unavailable case, pre-check with `filterDisabledProviders()` before entering the provider loop.

**Rationale**:
- Direct config check uses lazy evaluation — expiration is automatically handled by `IsActive()` → `IsExpired()` at each request
- No sync mechanism needed between config and runtime Provider structs
- The config store already has `reloadIfModified()` so CLI changes are picked up automatically
- Simpler than adding a `ManuallyDisabled` bool to Provider and maintaining sync (aligns with Constitution Principle II: Simplicity)
- The scenario fallback in `ServeHTTP` already handles "all scenario providers failed" → falls back to default providers. Adding unavailability filtering here integrates naturally.

**Alternatives considered**:
- `ManuallyDisabled` bool on runtime `Provider` struct: Rejected — requires sync mechanism between config and runtime; stale bool doesn't auto-clear on expiration; violates Principle II
- Filter in LoadBalancer.Select(): Rejected — LoadBalancer handles ordering, not filtering; mixing concerns
- Filter before building provider list: Possible but moves logic away from the request path where it's most visible

## R-006: Web API Endpoint Design

**Decision**: Add `POST /api/v1/providers/{name}/disable` and `POST /api/v1/providers/{name}/enable` endpoints.

**Rationale**:
- Action-oriented endpoints under the existing `/api/v1/providers/{name}` namespace
- Consistent with REST conventions for state-changing operations on a resource
- Response includes the updated unavailability status for immediate UI feedback
- Separate `GET /api/v1/providers/disabled` endpoint for listing all disabled providers

**Alternatives considered**:
- PATCH on provider with unavailability field: Rejected — mixes config mutation with runtime state
- New `/api/v1/unavailable` resource: Rejected — conceptually the state belongs to the provider
