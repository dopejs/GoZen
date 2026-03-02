# Data Model: Provider & Model Tag in Proxy Responses

**Feature**: 005-provider-model-tag
**Date**: 2026-03-02

## Entity Changes

### Modified Entity: OpenCCConfig

**File**: `internal/config/config.go`

| Field | Type | JSON Key | Default | Description |
|-------|------|----------|---------|-------------|
| ShowProviderTag | `bool` | `show_provider_tag,omitempty` | `false` | Enable/disable provider/model tag injection in proxy responses |

**Notes**:
- Added alongside existing global settings (WebPort, ProxyPort, DefaultProfile, etc.)
- `omitempty` + Go zero value (`false`) means field is absent in JSON when disabled — backward compatible
- No migration logic needed beyond version bump (absent field = `false` = feature disabled)

### Modified Entity: settingsResponse (API)

**File**: `internal/web/api_settings.go`

| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| ShowProviderTag | `bool` | `show_provider_tag` | Current state of provider tag setting |

### Modified Entity: settingsRequest (API)

**File**: `internal/web/api_settings.go`

| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| ShowProviderTag | `*bool` | `show_provider_tag,omitempty` | Pointer to distinguish "not sent" from "set to false" |

**Note**: Using `*bool` for the request allows the API to distinguish between "field not included" (nil) and "explicitly set to false". This is necessary because `omitempty` on a plain `bool` would prevent the user from disabling the feature via the API.

### Modified Entity: Settings (Web UI TypeScript)

**File**: `web/src/types/api.ts`

| Field | Type | Description |
|-------|------|-------------|
| show_provider_tag? | `boolean` | Optional, matches API response |

## Config Version

| Version | Change |
|---------|--------|
| 10 (current) | — |
| 11 (new) | Added `show_provider_tag` boolean field |

## State Transitions

N/A — The setting is a simple boolean toggle with no intermediate states.

## Relationships

```
OpenCCConfig.ShowProviderTag
  → read by ProxyServer.copyResponse() to decide tag injection
  → read/written by Web API /api/v1/settings endpoint
  → displayed/toggled in Web UI GeneralSettings tab
```

## Validation Rules

- `ShowProviderTag` accepts only `true` or `false` (enforced by Go type system)
- No cross-field dependencies
- No required combinations with other settings
