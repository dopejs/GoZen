# Data Model: Code Scenario Routing

**Feature**: 009-code-scenario-routing
**Date**: 2026-03-04

## Entity Changes

### Scenario (extended)

**File**: `internal/config/config.go`

| Constant | Value | New? | Description |
|----------|-------|------|-------------|
| ScenarioThink | `"think"` | No | Requests with `thinking.type=enabled` |
| ScenarioImage | `"image"` | No | Requests with image content blocks |
| ScenarioLongContext | `"longContext"` | No | Requests exceeding token threshold |
| ScenarioWebSearch | `"webSearch"` | No | Requests with `web_search` tools |
| ScenarioBackground | `"background"` | No | Requests for Haiku models |
| **ScenarioCode** | **`"code"`** | **Yes** | **Non-specialized, non-background requests (catch-all when configured)** |
| ScenarioDefault | `"default"` | No | Final fallback |

**Detection priority order**: webSearch > think > image > longContext > **code** > background > default

### Detection Logic (new function)

**File**: `internal/proxy/scenario.go`

```
isCodeRequest(body) → bool
  Returns true when:
    - Request is NOT a background request (not Haiku model)
  Note: All higher-priority checks (webSearch, think, image, longContext)
  are already evaluated before this function is called in DetectScenario().
```

### No Changes to Existing Entities

- **ScenarioRoute**: Unchanged — already supports arbitrary scenario keys via `map[Scenario]*ScenarioRoute`
- **ProviderRoute**: Unchanged — `{Name, Model}` structure works for all scenarios
- **ProfileConfig**: Unchanged — `Routing map[Scenario]*ScenarioRoute` accepts `"code"` key without modification
- **OpenCCConfig**: Unchanged — no config version bump needed

## State Transitions

None. Scenario detection is stateless per-request. The `code` scenario does not introduce any new state or lifecycle.

## Validation Rules

- `code` scenario route in `zen.json` is optional. When absent, behavior is identical to current (backward compatible).
- When present, `code` route must follow the same validation as other scenario routes: at least one provider with a valid name.
- The `code` detection explicitly excludes Haiku models to prevent overlap with `background`.

## Data Volume / Scale

No impact. The `code` check adds one `isBackgroundRequest()` call (string contains check) to the detection path — negligible cost.
