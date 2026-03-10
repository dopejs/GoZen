# Refactoring Impact Analysis: Scenario Routing Architecture Redesign

**Date**: 2026-03-10
**Branch**: 020-scenario-routing-redesign
**Analysis Type**: Existing Code Impact Assessment

## Executive Summary

This is a **REFACTORING PROJECT**, not new development. The existing codebase already has:
- Scenario detection logic in `internal/proxy/scenario.go`
- Fixed enum-based routing with `config.Scenario` type
- Scenario routing configuration in `ProfileConfig.Routing`
- TUI and Web UI for managing scenario routes
- Integration throughout the proxy pipeline

**Key Challenge**: Migrate from fixed enum `Scenario` type to open string-based scenario keys while maintaining backward compatibility.

---

## Existing Code Structure

### 1. Core Scenario Detection (`internal/proxy/scenario.go`)

**Current Implementation**:
```go
func DetectScenario(body map[string]interface{}, threshold int, sessionID string) config.Scenario {
    if hasWebSearchTool(body) {
        return config.ScenarioWebSearch
    }
    if hasThinkingEnabled(body) {
        return config.ScenarioThink
    }
    if hasImageContent(body) {
        return config.ScenarioImage
    }
    if isLongContext(body, threshold, sessionID) {
        return config.ScenarioLongContext
    }
    if isBackgroundRequest(body) {
        return config.ScenarioBackground
    }
    return config.ScenarioCode
}
```

**Detection Functions**:
- `hasWebSearchTool()` - checks for `web_search` tool in request
- `hasThinkingEnabled()` - checks for `thinking` field in request
- `hasImageContent()` - checks for image content blocks
- `isLongContext()` - uses tiktoken for token counting with session history
- `isBackgroundRequest()` - checks for haiku model requests

**Protocol Support**: Currently **Anthropic-only** (checks Anthropic-specific fields)

**Priority Order**: webSearch > think > image > longContext > code > background > default

---

### 2. Config Types (`internal/config/config.go`)

**Current Scenario Enum**:
```go
type Scenario string

const (
    ScenarioThink       Scenario = "think"
    ScenarioImage       Scenario = "image"
    ScenarioLongContext Scenario = "longContext"
    ScenarioWebSearch   Scenario = "webSearch"
    ScenarioBackground  Scenario = "background"
    ScenarioCode        Scenario = "code"
    ScenarioDefault     Scenario = "default"
)
```

**Current Routing Config**:
```go
type ProfileConfig struct {
    Providers            []string                    `json:"providers"`
    Routing              map[Scenario]*ScenarioRoute `json:"routing,omitempty"`
    LongContextThreshold int                         `json:"long_context_threshold,omitempty"`
    Strategy             LoadBalanceStrategy         `json:"strategy,omitempty"`
    ProviderWeights      map[string]int              `json:"provider_weights,omitempty"`
}

type ScenarioRoute struct {
    Providers []*ProviderRoute `json:"providers"`
}

type ProviderRoute struct {
    Name  string `json:"name"`
    Model string `json:"model,omitempty"`
}
```

**Migration Support**: Already has `UnmarshalJSON` for backward compatibility with old format

---

### 3. Proxy Server Integration (`internal/proxy/server.go`)

**Current Usage** (line 360):
```go
detectedScenario, _ = DetectScenarioFromJSON(bodyBytes, threshold, sessionID)
if sp, ok := s.Routing.ScenarioRoutes[detectedScenario]; ok {
    // Use scenario-specific providers
}
```

**RoutingConfig Type** (line 115):
```go
type RoutingConfig struct {
    DefaultProviders     []*Provider
    ScenarioRoutes       map[config.Scenario]*ScenarioProviders
    LongContextThreshold int
}

type ScenarioProviders struct {
    Providers []*Provider
    Models    map[string]string // provider name → model override
}
```

**Middleware Pipeline**: Exists (lines 310-347) but **does NOT drive routing decisions** currently

---

### 4. ProfileProxy Integration (`internal/proxy/profile_proxy.go`)

**Current Flow** (lines 84-100):
```go
// Build routing config if scenario routing is configured
var routing *RoutingConfig
if profileCfg.routing != nil && len(profileCfg.routing) > 0 {
    scenarioRoutes := make(map[config.Scenario]*ScenarioProviders)
    for scenario, sr := range profileCfg.routing {
        scenarioProviders, err := pp.buildProviders(sr.ProviderNames(), profileCfg.providerWeights)
        // ... build ScenarioProviders
        scenarioRoutes[scenario] = &ScenarioProviders{
            Providers: scenarioProviders,
            Models:    models,
        }
    }
    routing = &RoutingConfig{
        DefaultProviders:     providers,
        ScenarioRoutes:       scenarioRoutes,
        LongContextThreshold: profileCfg.LongContextThreshold,
    }
}
```

---

### 5. TUI Integration (`tui/routing.go`)

**Current Implementation**:
- Fixed list of scenarios in `knownScenarios` (lines 55-65)
- Uses `config.Scenario` enum type throughout
- Scenario editor for configuring providers per scenario
- Reads/writes `ProfileConfig.Routing` as `map[config.Scenario]*config.ScenarioRoute`

**Known Scenarios**:
```go
var knownScenarios = []struct {
    scenario config.Scenario
    label    string
}{
    {config.ScenarioWebSearch, "webSearch   (requests with web_search tools)"},
    {config.ScenarioThink, "think       (thinking mode requests)"},
    {config.ScenarioImage, "image       (requests with images)"},
    {config.ScenarioLongContext, "longContext (exceeds threshold)"},
    {config.ScenarioCode, "code        (regular coding requests)"},
    {config.ScenarioBackground, "background  (haiku model requests)"},
}
```

---

### 6. Web UI Integration (`web/src/types/api.ts`)

**Current Types**:
```typescript
export type Scenario = 'think' | 'image' | 'longContext' | 'webSearch' | 'code' | 'background' | 'default'

export const SCENARIOS: Scenario[] = ['default', 'think', 'image', 'longContext', 'code', 'webSearch', 'background']

export const SCENARIO_LABELS: Record<Scenario, string> = {
  default: 'Default',
  think: 'Extended Thinking',
  image: 'Image Processing',
  longContext: 'Long Context',
  code: 'Code',
  webSearch: 'Web Search',
  background: 'Background Tasks',
}

export interface Profile {
  name: string
  providers: string[]
  routing?: Partial<Record<Scenario, ScenarioRoute>>
  long_context_threshold?: number
  strategy?: LoadBalanceStrategy
  is_default?: boolean
}
```

**Missing**: `weighted` strategy in `LOAD_BALANCE_STRATEGIES` (only has failover, round-robin, least-latency, least-cost)

---

## Files Requiring Modification

### High Impact (Core Refactoring)

1. **`internal/config/config.go`**
   - Change `Scenario` from enum to alias for `string`
   - Keep constants for backward compatibility
   - Update `ProfileConfig.Routing` type signature (already `map[Scenario]*ScenarioRoute`, so minimal change)
   - Add scenario key validation function
   - Add scenario alias mapping (think→reasoning, webSearch→search, etc.)

2. **`internal/proxy/scenario.go`**
   - Rename to `internal/proxy/routing_classifier.go`
   - Refactor `DetectScenario()` to return `string` instead of `config.Scenario`
   - Add protocol-agnostic detection (currently Anthropic-only)
   - Add normalization layer for OpenAI Chat and OpenAI Responses
   - Keep existing detection logic as builtin classifier

3. **`internal/proxy/server.go`**
   - Update `RoutingConfig.ScenarioRoutes` from `map[config.Scenario]*ScenarioProviders` to `map[string]*ScenarioProviders`
   - Add middleware routing decision integration
   - Add protocol detection and normalization
   - Update scenario detection call to use new classifier

4. **`internal/proxy/profile_proxy.go`**
   - Update routing config building to use string keys
   - No major logic changes needed

### Medium Impact (UI Updates)

5. **`tui/routing.go`**
   - Keep `knownScenarios` list for UI display
   - Allow custom scenario input (text field for scenario key)
   - Update type references from `config.Scenario` to `string`

6. **`web/src/types/api.ts`**
   - Change `Scenario` from union type to `string`
   - Keep `SCENARIOS` array for UI display (builtin scenarios)
   - Update `Profile.routing` to `Record<string, ScenarioRoute>` (remove `Partial`)
   - Add `weighted` to `LOAD_BALANCE_STRATEGIES`

7. **`web/src/pages/profiles/edit.tsx`**
   - Update scenario routing UI to allow custom scenario keys
   - Keep dropdown for builtin scenarios, add text input for custom

### Low Impact (Tests & Documentation)

8. **`internal/proxy/scenario_test.go`**
   - Update test expectations to use string scenario keys
   - Add tests for custom scenario keys

9. **`internal/config/config_test.go`**
   - Add tests for scenario key validation
   - Add tests for scenario alias mapping
   - Update existing routing tests

10. **`internal/proxy/server_test.go`**
    - Update routing tests to use string keys

11. **`tui/routing_test.go`** (if exists)
    - Update TUI tests

12. **`web/src/pages/profiles/edit.test.tsx`**
    - Update Web UI tests

---

## Backward Compatibility Strategy

### Config Migration (v14 → v15)

**Current Version**: 14 (from 019-profile-strategy-routing)

**New Version**: 15

**Migration Path**:
1. Keep `Scenario` type as `type Scenario = string` (not enum)
2. Keep scenario constants for backward compatibility
3. JSON unmarshaling already supports `map[Scenario]*ScenarioRoute` → `map[string]*ScenarioRoute` (no change needed)
4. Add scenario alias mapping in classifier:
   - `think` → `reasoning` (or keep `think` as canonical)
   - `webSearch` → `search` (or keep `webSearch` as canonical)
   - `longContext` → `long_context` (or keep `longContext` as canonical)
   - `code` → `coding` (or keep `code` as canonical)

**Decision Needed**: Should we normalize scenario keys to kebab-case (`long-context`, `web-search`) or keep camelCase for backward compatibility?

**Recommendation**: Keep existing keys as-is for backward compatibility, add aliases for new canonical names

---

## Protocol Normalization Strategy

### Current State
- Detection logic is **Anthropic-only**
- Checks Anthropic-specific fields: `thinking`, `system`, content blocks structure

### Target State
- Support 3 protocols: Anthropic Messages, OpenAI Chat, OpenAI Responses
- Normalize all protocols to common `NormalizedRequest` structure
- Extract features protocol-agnostically

### Implementation Approach

**Option 1: Refactor Existing Functions**
- Keep `scenario.go` structure
- Add protocol detection at the top
- Branch detection logic based on protocol
- Pros: Minimal file changes
- Cons: Complex branching logic, harder to test

**Option 2: New Normalization Layer (Recommended)**
- Create `routing_normalize.go` with protocol-agnostic normalization
- Create `routing_classifier.go` with builtin classifier (refactored from `scenario.go`)
- Keep `scenario.go` as deprecated wrapper for backward compatibility
- Pros: Clean separation, easier to test, follows plan
- Cons: More files, need to maintain wrapper

**Recommendation**: Use Option 2 (matches original plan)

---

## Middleware Integration Strategy

### Current State
- Middleware pipeline exists in `server.go` (lines 310-347)
- Middleware can modify request body but **cannot drive routing**
- No `RoutingDecision` or `RoutingHints` in `RequestContext`

### Target State
- Middleware can set `RoutingDecision` to explicitly choose scenario
- Middleware can set `RoutingHints` to influence builtin classifier
- Builtin classifier runs only if no `RoutingDecision` set

### Implementation Approach

1. Add fields to `RequestContext` in `internal/middleware/interface.go`:
   ```go
   type RequestContext struct {
       // ... existing fields
       RequestFormat      string
       NormalizedRequest  *NormalizedRequest
       RoutingDecision    *RoutingDecision
       RoutingHints       *RoutingHints
   }
   ```

2. Update `server.go` to check `RoutingDecision` after middleware:
   ```go
   // Run middleware pipeline
   reqCtx = pipeline.ProcessRequest(reqCtx)

   // Resolve routing decision
   var scenario string
   if reqCtx.RoutingDecision != nil {
       scenario = reqCtx.RoutingDecision.Scenario
   } else {
       scenario = classifier.Classify(reqCtx.NormalizedRequest, reqCtx.RoutingHints)
   }
   ```

---

## Plan & Tasks Revision Assessment

### What Needs Revision

1. **Phase 1: Setup**
   - ✅ Keep as-is (file structure still valid)

2. **Phase 2: Foundational**
   - ⚠️ **T004**: Config version already at 14, need to bump to 15
   - ⚠️ **T005**: `ProfileRoutingConfig` doesn't exist - should be `ProfileConfig.Routing`
   - ⚠️ **T006**: Scenario alias mapping - need to decide on canonical names
   - ✅ T007-T008: Keep as-is

3. **Phase 3: User Story 1 (Protocol-Agnostic)**
   - ⚠️ **T015-T016**: Types already exist in plan, but need to integrate with existing code
   - ⚠️ **T017**: Protocol detection - need to integrate with existing `DetectScenarioFromJSON`
   - ⚠️ **T018-T020**: Normalization - new code, but need to preserve existing detection logic
   - ⚠️ **T021**: Feature extraction - refactor from existing `hasImageContent()`, `isLongContext()`, etc.
   - ⚠️ **T022**: Token counting - already exists in `isLongContext()`, need to extract
   - ⚠️ **T023-T025**: Server integration - need to refactor existing code, not write from scratch

4. **Phase 4: User Story 2 (Middleware-Driven)**
   - ⚠️ **T030-T031**: Builtin classifier - refactor from existing `DetectScenario()`, not new
   - ⚠️ **T032**: Routing decision resolution - new logic, but integrate with existing routing
   - ⚠️ **T034-T036**: Server integration - refactor existing middleware integration

5. **Phase 5: User Story 3 (Open Namespace)**
   - ⚠️ **T041**: Scenario key normalization - need to decide on backward compatibility approach
   - ⚠️ **T042**: Route policy resolution - refactor existing routing lookup
   - ⚠️ **T044-T045**: Server integration - refactor existing code

6. **Phase 6: User Story 4 (Per-Scenario Policies)**
   - ⚠️ **T051-T052**: LoadBalancer already supports strategies, need to add route-specific overrides
   - ⚠️ **T053**: Model overrides already exist in `ScenarioProviders.Models`, need to refactor
   - ⚠️ **T054**: Threshold override - new feature
   - ⚠️ **T055**: Server integration - refactor existing code

7. **Phase 7-8: User Stories 5-6**
   - ✅ Keep as-is (validation and observability are new features)

8. **Phase 9: Config Migration**
   - ⚠️ **T082-T084**: Need to update for actual migration path (v14→v15, not v14→v15)
   - ⚠️ Need to add TUI and Web UI migration tasks

9. **Phase 10: Polish**
   - ✅ Keep as-is

### What Needs Addition

1. **TUI Refactoring Tasks**
   - Update `tui/routing.go` to support custom scenario keys
   - Update `tui/fallback.go` if it references scenarios
   - Update `tui/dashboard.go` if it displays scenario info
   - Update `tui/config_main.go` if it manages routing

2. **Web UI Refactoring Tasks**
   - Update `web/src/types/api.ts` to change `Scenario` type
   - Update `web/src/pages/profiles/edit.tsx` to support custom scenarios
   - Add `weighted` strategy to UI
   - Update tests

3. **Deprecation Tasks**
   - Add deprecation notice to `scenario.go` (keep as wrapper for backward compatibility)
   - Update documentation to reference new routing system

---

## Critical Decisions Needed

### 1. Scenario Key Naming Convention

**Options**:
- **A**: Keep existing camelCase keys (`think`, `webSearch`, `longContext`, `code`)
- **B**: Migrate to kebab-case keys (`reasoning`, `search`, `long-context`, `coding`)
- **C**: Support both via alias mapping

**Recommendation**: **Option C** - Support both for maximum backward compatibility
- Existing configs continue to work with camelCase keys
- New configs can use kebab-case keys
- Classifier normalizes all keys to canonical form
- UI displays both builtin and custom scenarios

### 2. Config Version Bump

**Current**: Version 14 (from 019-profile-strategy-routing)
**Target**: Version 15

**Changes**:
- `ProfileConfig.Routing` type signature (minimal - already `map[Scenario]*ScenarioRoute`)
- Add scenario alias support
- No breaking changes to JSON structure

**Migration**: Automatic (no manual intervention needed)

### 3. Backward Compatibility for `Scenario` Type

**Options**:
- **A**: Change `Scenario` from enum to `type Scenario = string`, keep constants
- **B**: Keep enum, add validation for custom keys
- **C**: Remove enum entirely, use plain `string`

**Recommendation**: **Option A** - Minimal breaking changes
- Go code using `config.ScenarioThink` continues to work
- New code can use string literals
- Type safety preserved for builtin scenarios

### 4. Protocol Detection Priority

**Current**: Anthropic-only
**Target**: Anthropic, OpenAI Chat, OpenAI Responses

**Detection Strategy**:
1. Check URL path (`/v1/messages`, `/v1/chat/completions`, `/v1/responses`)
2. Check request body structure (fallback)
3. Default to OpenAI Chat if ambiguous

### 5. Middleware Routing Decision Precedence

**Current**: No middleware routing
**Target**: Middleware can override builtin classifier

**Precedence**:
1. Middleware `RoutingDecision` (highest priority)
2. Builtin classifier with `RoutingHints`
3. Builtin classifier without hints
4. Default scenario (fallback)

---

## Recommended Revision to Plan & Tasks

### Revised Implementation Strategy

**Phase 0: Refactoring Preparation** (NEW)
- Analyze existing code structure
- Document current behavior
- Create refactoring test suite (preserve existing behavior)
- Decision: Scenario key naming convention
- Decision: Backward compatibility approach

**Phase 1: Setup** (KEEP)
- No changes needed

**Phase 2: Foundational** (REVISE)
- Bump config version 14 → 15
- Add scenario alias mapping (decision-dependent)
- Add scenario key validation
- Update `ProfileConfig` documentation

**Phase 3: Protocol Normalization** (REVISE)
- Extract existing detection logic to separate functions
- Add protocol detection (URL path + body structure)
- Create normalization layer for OpenAI Chat and OpenAI Responses
- Refactor existing Anthropic detection to use normalization
- **Preserve existing behavior** for Anthropic requests

**Phase 4: Middleware Integration** (REVISE)
- Add `RoutingDecision` and `RoutingHints` to `RequestContext`
- Refactor existing classifier to use normalized requests
- Add middleware decision precedence logic
- **Preserve existing behavior** when no middleware decision

**Phase 5: Open Namespace** (REVISE)
- Change `Scenario` type to `string` alias
- Update `RoutingConfig.ScenarioRoutes` to `map[string]*ScenarioProviders`
- Add custom scenario support in routing resolution
- **Preserve existing behavior** for builtin scenarios

**Phase 6: Per-Scenario Policies** (REVISE)
- Add route-specific strategy overrides
- Add route-specific threshold overrides
- Refactor existing model override logic
- **Preserve existing behavior** for default policies

**Phase 7: TUI Refactoring** (NEW)
- Update `tui/routing.go` to support custom scenarios
- Update other TUI files referencing scenarios
- Add tests

**Phase 8: Web UI Refactoring** (NEW)
- Update `web/src/types/api.ts`
- Update `web/src/pages/profiles/edit.tsx`
- Add `weighted` strategy to UI
- Add tests

**Phase 9: Config Validation** (KEEP)
- No changes needed

**Phase 10: Observability** (KEEP)
- No changes needed

**Phase 11: Config Migration** (REVISE)
- Update migration logic for v14→v15
- Add scenario alias migration
- Add tests

**Phase 12: Polish** (KEEP)
- No changes needed

---

## Risk Assessment

### High Risk

1. **Breaking existing routing behavior**
   - Mitigation: Comprehensive test suite before refactoring
   - Mitigation: Preserve existing detection logic as-is
   - Mitigation: Add integration tests for all existing scenarios

2. **Config migration failures**
   - Mitigation: Extensive migration testing with real configs
   - Mitigation: Fallback to default route on migration errors
   - Mitigation: Clear error messages for invalid configs

3. **TUI/Web UI breaking changes**
   - Mitigation: Update UI types carefully
   - Mitigation: Test with existing configs
   - Mitigation: Provide clear upgrade path in UI

### Medium Risk

4. **Performance regression from normalization**
   - Mitigation: Profile normalization overhead
   - Mitigation: Cache normalized requests per session
   - Mitigation: Lazy normalization (only when needed)

5. **Middleware integration complexity**
   - Mitigation: Clear precedence rules
   - Mitigation: Comprehensive logging
   - Mitigation: Fallback to builtin classifier on errors

### Low Risk

6. **Scenario key naming conflicts**
   - Mitigation: Scenario key validation
   - Mitigation: Reserved key list for builtins
   - Mitigation: Clear documentation

---

## Next Steps

1. **User Decision Required**:
   - Scenario key naming convention (camelCase vs kebab-case vs both)
   - Backward compatibility approach for `Scenario` type
   - Config version bump strategy

2. **Plan Revision**:
   - Update `plan.md` with refactoring context
   - Add Phase 0 (Refactoring Preparation)
   - Revise Phases 3-6 to focus on refactoring, not new development
   - Add Phases 7-8 for TUI/Web UI refactoring

3. **Tasks Revision**:
   - Update task descriptions to reflect refactoring nature
   - Add "Refactor from existing X" notes
   - Add "Preserve existing behavior" checkpoints
   - Add TUI/Web UI refactoring tasks
   - Add comprehensive test tasks for existing behavior

4. **Implementation**:
   - Start with Phase 0 (refactoring preparation)
   - Create comprehensive test suite for existing behavior
   - Proceed with refactoring only after tests pass

---

## Conclusion

This is a **significant refactoring project** that requires careful planning to avoid breaking existing functionality. The original plan and tasks were written for greenfield development and need substantial revision to account for:

1. Existing scenario detection logic
2. Existing routing configuration structure
3. Existing TUI and Web UI integration
4. Backward compatibility requirements
5. Config migration complexity

**Recommendation**: Revise plan and tasks before proceeding with implementation. Focus on refactoring existing code rather than writing new code from scratch.
