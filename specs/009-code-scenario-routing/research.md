# Research: Code Scenario Routing

**Feature**: 009-code-scenario-routing
**Date**: 2026-03-04

## Research Summary

No NEEDS CLARIFICATION items existed in the Technical Context. This feature is a small addition to an existing, well-understood system. Research focused on confirming implementation patterns and validating the detection approach.

## R1: Detection Strategy for `code` Scenario

**Decision**: Negative detection — `code` matches when no other specialized scenario matches AND the request is not a background (Haiku) request.

**Rationale**: The `code` scenario is defined as a catch-all for non-specialized requests. There is no positive signal that distinguishes a "coding" request from any other regular request. The only distinguishing factor vs. `default` is that `code` is explicitly configured.

**Implementation**: In `DetectScenario()`, after the `longContext` check and before the `background` check, add:

```go
if !isBackgroundRequest(body) {
    return config.ScenarioCode
}
```

This ensures:
- All specialized scenarios (webSearch, think, image, longContext) are checked first
- Haiku/background requests are explicitly excluded from `code`
- `code` catches everything else before `background` and `default`

**Alternatives considered**:
1. **Positive detection (tool_use presence)**: Rejected — not all coding requests use tools, and non-coding requests can also use tools.
2. **Check background before code**: Rejected per user decision in clarify session — `code` appears before `background` in priority order, with explicit haiku exclusion in the `code` detection itself.

## R2: Config Schema Impact

**Decision**: No config version bump or migration needed.

**Rationale**: The `ProfileConfig.Routing` field is `map[Scenario]*ScenarioRoute` where `Scenario` is a `string` type. Adding a new string value `"code"` requires no structural change. Existing configs without a `"code"` key will simply not trigger the code scenario — backward compatible by design.

**Verification**: Confirmed in `internal/config/config.go` lines 235-245 and `ScenarioRoute.UnmarshalJSON` lines 258-269.

**Alternatives considered**: None needed — the existing design explicitly supports new scenario keys without migration.

## R3: Web UI Rendering Approach

**Decision**: Add `'code'` to the `SCENARIOS` array and `SCENARIO_LABELS` map in `web/src/types/api.ts`. No changes needed to `edit.tsx`.

**Rationale**: The profile editor (`edit.tsx`) iterates over `SCENARIOS` to render the routing tab. Adding a new entry to the array and labels map is sufficient for the UI to display it.

**Verification**: Confirmed in `web/src/pages/profiles/edit.tsx` line 21 (imports `SCENARIOS`, `SCENARIO_LABELS`) and the rendering loop that maps over `SCENARIOS`.

**Alternatives considered**: None — the existing architecture was designed for extensibility.

## R4: TUI Routing Editor

**Decision**: Add a new entry to the `knownScenarios` slice in `tui/routing.go`.

**Rationale**: The TUI routing editor iterates over `knownScenarios` to display scenario entries. Adding a new entry follows the exact same pattern as existing scenarios.

**Verification**: Confirmed in `tui/routing.go` lines 55-64.

**Alternatives considered**: None — follows existing pattern exactly.
