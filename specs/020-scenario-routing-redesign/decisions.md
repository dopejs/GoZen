# Key Design Decisions: Scenario Routing Architecture Redesign

**Date**: 2026-03-10
**Feature**: 020-scenario-routing-redesign
**Status**: Finalized

## Decision Summary

This document records all key design decisions made during the planning phase. These decisions are **final** and should not be changed without revisiting the entire plan.

---

## Decision 1: Scenario Key Naming Convention

**Question**: How should scenario keys be named and normalized?

**Options Considered**:
- A: Keep existing camelCase only
- B: Migrate to kebab-case only
- C: Support both, normalize internally

**Decision**: **C - Support camelCase, kebab-case, and snake_case; normalize internally to camelCase**

**Rationale**:
1. Backward compatibility with existing configs (all use camelCase)
2. Flexibility for users to use any naming style
3. Internal normalization ensures consistency

**Implementation**:
```go
func NormalizeScenarioKey(key string) string {
    // Convert kebab-case and snake_case to camelCase
    // "web-search" → "webSearch"
    // "long_context" → "longContext"
    return normalized
}
```

**Examples**:
- Input: `web-search`, `web_search`, `webSearch` → Output: `webSearch`
- Input: `long-context`, `long_context`, `longContext` → Output: `longContext`

---

## Decision 2: Scenario Type Definition

**Question**: How should the `Scenario` type be defined to support open namespace?

**Options Considered**:
- A: Type alias + constants (backward compatible)
- B: Keep enum, add validation
- C: Remove enum entirely, use plain string

**Decision**: **A - Type alias with constants for builtin scenarios**

**Rationale**:
1. Minimal breaking changes (existing code using constants continues to work)
2. Type safety for builtin scenarios
3. Flexibility for custom scenario strings
4. Go idiomatic pattern

**Implementation**:
```go
// config.go
type Scenario = string  // Type alias, not new type

// Constants for builtin scenarios (backward compatibility)
const (
    ScenarioThink       = "think"
    ScenarioImage       = "image"
    ScenarioLongContext = "longContext"
    ScenarioWebSearch   = "webSearch"
    ScenarioBackground  = "background"
    ScenarioCode        = "code"
    ScenarioDefault     = "default"
)
```

**Impact**:
- `ProfileConfig.Routing` type signature unchanged: `map[Scenario]*RoutePolicy`
- Now accepts any string as key, not just enum values
- JSON serialization/deserialization unchanged

---

## Decision 3: Config Version and Structure

**Question**: How should config be migrated to support new routing features?

**Options Considered**:
- A: Simple version bump, no structure changes
- B: Add migration logic, normalize keys
- C: New RoutePolicy structure, v14 → v15 migration

**Decision**: **C - New RoutePolicy structure with v14 → v15 migration**

**Rationale**:
1. Enables per-scenario strategies, weights, thresholds
2. Clean separation of concerns
3. Automatic migration preserves user configs
4. Aligns with original design goals

**Old Structure (v14)**:
```go
type ScenarioRoute struct {
    Providers []*ProviderRoute `json:"providers"`
}
```

**New Structure (v15)**:
```go
type RoutePolicy struct {
    Providers            []*ProviderRoute        `json:"providers"`
    Strategy             LoadBalanceStrategy     `json:"strategy,omitempty"`
    ProviderWeights      map[string]int          `json:"provider_weights,omitempty"`
    LongContextThreshold *int                    `json:"long_context_threshold,omitempty"`
    FallbackToDefault    *bool                   `json:"fallback_to_default,omitempty"`
}
```

**Migration Logic**:
```go
func (rp *RoutePolicy) UnmarshalJSON(data []byte) error {
    // Detect v14 format (only has "providers" field)
    // Convert to v15 format (add default values for new fields)
}
```

---

## Decision 4: Per-Scenario Routing Policies

**Question**: Should each scenario support independent routing policies?

**Options Considered**:
- A: No per-scenario policies (use profile defaults)
- B: Extend ScenarioRoute with policy fields
- C: Create new RoutePolicy type

**Decision**: **C - New RoutePolicy type** (already decided in Decision 3)

**Rationale**:
1. Enables sophisticated cost optimization per scenario
2. Different scenarios have different cost/quality tradeoffs
3. Clean type definition
4. Aligns with original design

**Supported Per-Scenario Policies**:
- `Strategy`: Load balancing strategy (failover, round-robin, least-latency, least-cost, weighted)
- `ProviderWeights`: Custom weights for weighted strategy
- `LongContextThreshold`: Custom token threshold
- `FallbackToDefault`: Whether to fall back to default providers on failure

**Example Config**:
```json
{
  "profiles": {
    "default": {
      "providers": ["p1", "p2"],
      "strategy": "failover",
      "routing": {
        "think": {
          "providers": [{"name": "p1", "model": "claude-opus-4"}],
          "strategy": "weighted",
          "provider_weights": {"p1": 100}
        },
        "code": {
          "providers": [{"name": "p2"}],
          "strategy": "least-cost"
        }
      }
    }
  }
}
```

---

## Decision 5: Protocol Detection Strategy

**Question**: How should the system detect which API protocol a request uses?

**Options Considered**:
- A: URL path priority, default to Anthropic
- B: URL path priority, default to OpenAI Chat
- C: Only URL path, no fallback

**Decision**: **Modified B - URL path → X-Zen-Client header → body structure → default to OpenAI Chat**

**Rationale**:
1. URL path is most reliable indicator
2. X-Zen-Client header provides context when path is ambiguous
3. Body structure as last resort
4. OpenAI Chat is most universal format

**Detection Priority**:
```go
func DetectProtocol(path string, headers http.Header, body map[string]interface{}) string {
    // 1. URL path (highest priority)
    if strings.HasSuffix(path, "/messages") {
        return "anthropic"
    }
    if strings.HasSuffix(path, "/chat/completions") {
        return "openai_chat"
    }
    if strings.HasSuffix(path, "/responses") {
        return "openai_responses"
    }

    // 2. Client header (next priority)
    clientType := headers.Get("X-Zen-Client")
    switch clientType {
    case "claude":
        return "anthropic"
    case "codex", "opencode":
        return "openai_chat"
    }

    // 3. Body structure (fallback)
    if _, hasInput := body["input"]; hasInput {
        return "openai_responses"
    }
    if _, hasSystem := body["system"]; hasSystem {
        return "anthropic"
    }

    // 4. Default
    return "openai_chat"
}
```

**Examples**:
- Claude Code → `/v1/messages` → `anthropic`
- Unknown client → `/v1/chat/completions` → `openai_chat`
- Unknown path + `X-Zen-Client: claude` → `anthropic`
- Completely unknown → `openai_chat`

---

## Decision 6: Implementation Strategy

**Question**: Should we refactor existing code or rewrite from scratch?

**Options Considered**:
- A: Refactor existing scenario.go (preserve and modify)
- B: Complete rewrite (replace scenario.go)
- C: Hybrid (new core, keep wrappers)

**Decision**: **B - Complete rewrite (replace existing implementation)**

**Rationale**:
1. Existing code has limited test coverage (only 1 E2E test)
2. Existing architecture doesn't support open scenario namespace
3. Existing code is Anthropic-only, requires major changes for multi-protocol
4. Clean slate enables better architecture
5. Original tasks.md was written for new development

**Approach**:
1. Create new files: `routing_normalize.go`, `routing_classifier.go`, `routing_resolver.go`
2. Deprecate old file: `scenario.go` (mark as deprecated, remove in future version)
3. Update all integration points: `server.go`, `profile_proxy.go`, `loadbalancer.go`
4. Build comprehensive test suite (TDD approach)
5. Preserve config compatibility (v14 → v15 migration)

**Files to Create**:
- `internal/proxy/routing_normalize.go` - Protocol normalization
- `internal/proxy/routing_classifier.go` - Builtin scenario classifier
- `internal/proxy/routing_decision.go` - RoutingDecision types
- `internal/proxy/routing_resolver.go` - Route policy resolution

**Files to Deprecate**:
- `internal/proxy/scenario.go` - Old scenario detection (will be removed)
- `internal/proxy/scenario_test.go` - Old tests (will be replaced)

**Files to Modify**:
- `internal/config/config.go` - Add RoutePolicy, change Scenario to string alias
- `internal/proxy/server.go` - Integrate new routing flow
- `internal/proxy/profile_proxy.go` - Use new routing types
- `internal/middleware/interface.go` - Add routing fields to RequestContext
- `tui/routing.go` - Support custom scenario keys
- `web/src/types/api.ts` - Update Scenario type

---

## Decision Impact Summary

| Decision | Impact | Risk | Mitigation |
|----------|--------|------|------------|
| 1. Scenario naming | Low | Low | Normalization function handles all cases |
| 2. Scenario type | Medium | Low | Type alias preserves backward compatibility |
| 3. Config structure | High | Medium | Automatic migration, comprehensive tests |
| 4. Per-scenario policies | Medium | Low | Optional fields, defaults to profile settings |
| 5. Protocol detection | Medium | Low | Clear priority order, well-tested |
| 6. Complete rewrite | High | High | TDD approach, comprehensive test coverage |

---

## Implementation Checklist

Before starting implementation, verify:

- [x] All 6 decisions finalized
- [x] plan.md updated with decisions
- [x] data-model.md updated with new structures
- [x] tasks.md updated with correct task descriptions
- [ ] Team alignment on complete rewrite approach
- [ ] Test strategy defined (TDD)
- [ ] Migration strategy validated

---

## Change Log

- 2026-03-10: Initial decisions finalized
- 2026-03-10: Updated plan.md, data-model.md, tasks.md with decisions
