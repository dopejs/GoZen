# Data Model: Manual Provider Unavailability Marking

**Feature**: 015-mark-provider-unavailable
**Date**: 2026-03-07

## New Entity: UnavailableMarking

Represents a user's manual decision to mark a provider as temporarily or permanently unavailable.

### Fields

| Field       | Type   | Description                                                                                             |
|-------------|--------|---------------------------------------------------------------------------------------------------------|
| Type        | string | Expiration type: `"today"`, `"month"`, or `"permanent"`                                                 |
| CreatedAt   | time   | Timestamp when the marking was created (local time)                                                     |
| ExpiresAt   | time   | Pre-computed expiration timestamp. Zero value for permanent markings. Calculated at creation time.       |

### JSON Representation

```json
{
  "type": "today",
  "created_at": "2026-03-07T14:30:00+08:00",
  "expires_at": "2026-03-07T23:59:59+08:00"
}
```

### Behaviors

- `IsExpired()`: Returns true if `ExpiresAt` is non-zero and `time.Now()` is after `ExpiresAt`. Permanent markings (zero `ExpiresAt`) never expire.
- `IsActive()`: Returns true if the marking exists and is not expired.

### Constraints

- One marking per provider (keyed by provider name)
- Setting a new marking replaces any existing one
- Marking is removed when provider is deleted from config

## Modified Entity: OpenCCConfig

### New Field

| Field              | Type                              | JSON Key              | Description                                  |
|--------------------|-----------------------------------|-----------------------|----------------------------------------------|
| DisabledProviders  | `map[string]*UnavailableMarking`  | `disabled_providers`  | Map of provider name → unavailability marking |

### Version Change

- Config version: 13 → 14
- Migration: No special migration needed. Old configs without `disabled_providers` parse as empty map.

### JSON Example (within zen.json)

```json
{
  "version": 14,
  "providers": {
    "anthropic-main": { "base_url": "...", "auth_token": "..." },
    "openai-backup": { "base_url": "...", "auth_token": "..." }
  },
  "disabled_providers": {
    "openai-backup": {
      "type": "today",
      "created_at": "2026-03-07T14:30:00+08:00",
      "expires_at": "2026-03-07T23:59:59+08:00"
    }
  }
}
```

## State Transitions

```
Available ──[user marks disable]──→ Unavailable (today/month/permanent)
    ↑                                      │
    │                                      │
    ├──[user clears marking]───────────────┘
    │                                      │
    └──[expiration time reached]───────────┘
```

## Interaction with Existing Health System

The provider has two independent availability dimensions:

1. **Auto health** (existing): `Provider.IsHealthy()` — backoff-based, auto-recovers
2. **Manual unavailability** (new): `config.IsProviderDisabled(name)` → calls `UnavailableMarking.IsActive()` — user-controlled, checked directly from config at request time (lazy evaluation, no sync needed)

During proxy routing (in `tryProviders()`):
- Check `config.IsProviderDisabled(p.Name)` for each provider before attempting it
- If manually marked unavailable AND active → skip (unless all providers unavailable → return 503 error)
- If auto-unhealthy → existing backoff logic applies (skip unless last provider)
- Both dimensions are independent: clearing a manual mark doesn't affect auto health, and vice versa
- Expiration is evaluated lazily via `IsActive()` — no background timer or sync mechanism needed
- `zen use <provider>` bypasses both manual unavailability (by design; auto health doesn't apply to direct use)

## Cleanup Rules

- When a provider is deleted from config (`DeleteProvider`), its entry in `disabled_providers` is also removed
- Expired markings are lazily ignored (treated as available) and optionally cleaned up on next config save
