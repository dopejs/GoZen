# Research: Scenario Routing Architecture Redesign

**Feature**: 020-scenario-routing-redesign
**Date**: 2026-03-10
**Purpose**: Resolve technical unknowns and establish implementation patterns for protocol-agnostic, middleware-extensible routing

## Research Areas

### 1. Config Migration Strategy

**Decision**: Use custom `UnmarshalJSON` with new-format-first detection, fall back to `json.RawMessage` for legacy format conversion

**Rationale**:
- GoZen already uses this pattern successfully for previous config migrations
- Allows automatic, lossless migration from v14 (fixed scenario enums) to v15 (open string keys)
- Preserves backward compatibility while enabling new features
- Fail-fast validation catches configuration errors at load time

**Implementation Pattern**:
```go
func (pc *ProfileConfig) UnmarshalJSON(data []byte) error {
    // Try new format first (v15+)
    type newFormat struct {
        Routing map[string]*RoutePolicy `json:"routing,omitempty"`
    }
    var nf newFormat
    if err := json.Unmarshal(data, &nf); err == nil {
        if nf.Routing != nil && len(nf.Routing) > 0 {
            // Validate it's new format by checking RoutePolicy structure
            for _, policy := range nf.Routing {
                if policy != nil {
                    // New format confirmed
                    pc.Routing = nf.Routing
                    return nil
                }
            }
        }
    }

    // Fall back to legacy format with json.RawMessage
    type legacyFormat struct {
        Routing map[string]json.RawMessage `json:"routing,omitempty"`
    }
    var lf legacyFormat
    if err := json.Unmarshal(data, &lf); err != nil {
        return err
    }

    // Convert legacy ScenarioRoute to new RoutePolicy
    pc.Routing = make(map[string]*RoutePolicy, len(lf.Routing))
    for key, rawMsg := range lf.Routing {
        var legacyRoute ScenarioRoute
        if err := json.Unmarshal(rawMsg, &legacyRoute); err != nil {
            continue
        }
        pc.Routing[key] = &RoutePolicy{
            Providers: legacyRoute.Providers,
        }
    }
    return nil
}
```

**Validation Strategy**:
- Validate at save time with `ValidateRoutingConfig()`
- Check scenario key format (alphanumeric + `-` or `_`, max 64 chars)
- Verify all referenced providers exist
- Validate strategy values against enum
- Validate weights for weighted strategy
- Return structured errors with clear messages

**Scenario Aliases**:
```go
var ScenarioAliases = map[string]string{
    "think":       "reasoning",
    "webSearch":   "search",
    "longContext": "long_context",
}
```

**Alternatives Considered**:
- Database-style migrations: Too heavyweight for JSON config file
- Breaking change without migration: Unacceptable, violates Constitution III
- Dual config format support: Increases complexity, harder to maintain

---

### 2. Protocol Normalization

**Decision**: Create `NormalizedRequest` struct that captures protocol-agnostic semantics from Anthropic Messages, OpenAI Chat, and OpenAI Responses

**Rationale**:
- Three API formats share common semantic elements (model, messages, tools, system prompts)
- Normalization enables protocol-agnostic scenario detection
- Preserves original request for passthrough to providers
- Allows middleware to work with unified request representation

**Struct Design**:
```go
type NormalizedRequest struct {
    // Core fields
    Model       string
    MaxTokens   int
    Temperature *float64
    Stream      bool

    // Conversation
    System   string
    Messages []NormalizedMessage

    // Tools
    Tools      []NormalizedTool
    ToolChoice string

    // Advanced features
    Thinking *ThinkingConfig

    // Metadata
    Features     RequestFeatures
    OriginalBody map[string]interface{}
}

type RequestFeatures struct {
    HasReasoning   bool
    HasImages      bool
    HasWebSearch   bool
    HasToolLoop    bool
    IsLongContext  bool
    TokenCount     int
    ToolCount      int
}
```

**Protocol Detection**:
1. **Primary**: URL path patterns (`/messages`, `/chat/completions`, `/responses`)
2. **Fallback**: Request body structure (presence of `input`, `system`, `thinking` fields)
3. **Supplementary**: Headers (`anthropic-version`, `OpenAI-Beta`)

**Key Differences**:
- **Anthropic**: `system` field, `thinking` object, typed content blocks
- **OpenAI Chat**: System role in messages, `max_completion_tokens`
- **OpenAI Responses**: `input` field, `instructions`, `previous_response_id`

**Edge Cases**:
- **Malformed requests**: Route to default route per FR-001 clarification
- **Protocol-specific features**: Store in `OriginalBody`, preserve during denormalization
- **Tool format mismatches**: Bidirectional mapping (Anthropic `input_schema` ↔ OpenAI `parameters`)
- **System prompt placement**: Extract to normalized `System` field, reconstruct based on target protocol

**Alternatives Considered**:
- Protocol-specific routing: Doesn't solve the core problem, duplicates logic
- Runtime protocol conversion: Too complex, increases latency
- Middleware-based normalization: Couples normalization to middleware, not reusable

---

### 3. Routing Decision Precedence

**Decision**: Use last-middleware-wins precedence with separate binding decisions (`RoutingDecision`) and non-binding hints (`RoutingHints`)

**Rationale**:
- Consistent with Go HTTP middleware patterns (sequential execution, last writer wins)
- Clear separation between explicit decisions (override builtin) and suggestions (influence builtin)
- Middleware pipeline order determines precedence (configurable by user)
- Enables debugging through decision source tracking

**Type Design**:
```go
type RoutingDecision struct {
    Scenario   string  // Required: scenario key
    Source     string  // Required: decision source (e.g., "middleware:spec-kit")
    Reason     string  // Required: human-readable explanation
    Confidence float64 // 0.0-1.0, where 1.0 = certain

    // Optional overrides (nil = not set)
    ModelHint         *string
    StrategyOverride  *config.LoadBalanceStrategy
    ThresholdOverride *int

    // Optional filters
    ProviderAllowlist []string
    ProviderDenylist  []string

    Metadata map[string]interface{}
}

type RoutingHints struct {
    ScenarioCandidates []string
    Tags               []string
    CostClass          string
    CapabilityNeeds    []string
    Confidence         map[string]float64
    Metadata           map[string]interface{}
}
```

**Precedence Algorithm**:
```
1. If middleware set RoutingDecision → use it
2. Else run builtin classifier with RoutingHints → use result
3. Else use default route
```

**Confidence Scoring**:
- `1.0` - Explicit (middleware set)
- `0.9` - High (strong signal like `thinking=true`)
- `0.7` - Medium (multiple weak signals)
- `0.5` - Low (single weak signal or heuristic)
- `0.3` - Guess (fallback/default)

**Pointer Fields Rationale**:
- Using `*string`, `*LoadBalanceStrategy`, `*int` for optional overrides
- Distinguishes "not set" (nil) from "set to zero value"
- Critical for overrides where zero values might be valid

**Observability**:
```go
func LogRoutingDecision(logger *StructuredLogger, decision *RoutingDecision, ctx *RequestContext, selectedProvider string) {
    fields := map[string]interface{}{
        "scenario":          decision.Scenario,
        "decision_source":   decision.Source,
        "decision_reason":   decision.Reason,
        "confidence":        decision.Confidence,
        "provider_selected": selectedProvider,
    }
    logger.Info("routing_decision", fields)
}
```

**Alternatives Considered**:
- First-middleware-wins: Less intuitive, harder to override earlier decisions
- Voting/consensus: Too complex, unclear semantics when middleware disagree
- Priority-based: Requires explicit priority configuration, less flexible

---

## Implementation Recommendations

### Phase 1: Normalization Layer
1. Create `internal/proxy/routing_normalize.go`
2. Implement `Normalize(body []byte, protocol string) (*NormalizedRequest, error)`
3. Add protocol detection functions
4. Implement feature extraction from normalized request
5. Add comprehensive tests for all three protocols

### Phase 2: Config Migration
1. Bump `CurrentConfigVersion` to 15 in `internal/config/config.go`
2. Implement `ProfileConfig.UnmarshalJSON` with new-format-first detection
3. Add `ValidateRoutingConfig()` with fail-fast validation
4. Implement scenario alias mapping
5. Add migration tests (v14→v15, mixed formats, validation edge cases)

### Phase 3: Routing Decision Types
1. Add `RoutingDecision` and `RoutingHints` types to `internal/proxy/routing_decision.go`
2. Update `RequestContext` in `internal/middleware/interface.go`
3. Implement `ResolveRoutingDecision()` precedence algorithm
4. Add validation and sanitization for invalid decisions
5. Implement structured logging for routing decisions

### Phase 4: Builtin Classifier Refactor
1. Create `internal/proxy/routing_classifier.go`
2. Refactor `DetectScenario()` to accept `*NormalizedRequest`
3. Implement protocol-agnostic feature detection
4. Add confidence scoring to classifier
5. Support `RoutingHints` in classification logic

### Phase 5: Integration
1. Update `ProxyServer.ServeHTTP()` to populate `RequestContext` with routing fields
2. Integrate normalization before middleware pipeline
3. Integrate decision resolution after middleware pipeline
4. Update `ProfileProxy` to use new routing flow
5. Update `LoadBalancer` to accept route-specific overrides

### Phase 6: Testing
1. Unit tests for normalization (all protocols)
2. Unit tests for config migration (v14→v15)
3. Unit tests for decision precedence
4. Integration tests for protocol-agnostic routing
5. Integration tests for middleware-driven routing
6. Integration tests for per-scenario policies

---

## Performance Considerations

**Normalization Overhead**:
- Target: < 10ms per request
- Approach: Lazy parsing (only parse fields needed for routing)
- Optimization: Cache protocol detection result in request context

**Config Validation**:
- Validate once at load time, not per request
- Cache validation results for hot path

**Decision Resolution**:
- Target: < 5ms overhead
- Approach: Early exit when middleware provides decision
- Optimization: Avoid unnecessary classifier execution

---

## Testing Strategy

**Unit Tests**:
- Normalization: All three protocols, edge cases, malformed requests
- Config migration: v14→v15, mixed formats, validation failures
- Decision precedence: Middleware override, builtin fallback, default fallback
- Classifier: Feature detection, confidence scoring, hint integration

**Integration Tests**:
- End-to-end routing flow with real requests
- Protocol-agnostic routing (same semantic content, different protocols)
- Middleware-driven routing (custom scenarios)
- Per-scenario policies (different strategies per scenario)

**Coverage Targets**:
- `internal/proxy`: 80% (per Constitution VI)
- `internal/config`: 80% (per Constitution VI)
- New routing files: 80%+ (critical path code)

---

## References

- GoZen existing config migration pattern in `internal/config/config.go`
- Go middleware chaining patterns
- Anthropic Messages API documentation
- OpenAI Chat Completions API documentation
- OpenAI Responses API specification
- Go struct optional fields patterns (pointer vs value)
- AI confidence scoring best practices
