# Contract: Routing API

**Feature**: 020-scenario-routing-redesign
**Date**: 2026-03-10
**Purpose**: Define the public API contract for routing normalization, classification, and resolution

## 1. Normalization API

### Function: `Normalize`

**Purpose**: Convert protocol-specific requests into normalized representation

**Signature**:
```go
func Normalize(body []byte, protocol string, sessionID string, threshold int) (*NormalizedRequest, error)
```

**Parameters**:
- `body` ([]byte): Raw request body (JSON)
- `protocol` (string): Detected protocol ("anthropic", "openai_chat", "openai_responses")
- `sessionID` (string): Session identifier for long-context detection
- `threshold` (int): Long-context token threshold

**Returns**:
- `*NormalizedRequest`: Protocol-agnostic request representation
- `error`: Normalization error (malformed request, unsupported protocol)

**Behavior**:
- Parse request body based on protocol
- Extract model, messages, tools, system prompt
- Normalize content blocks (text, image, tool_use, tool_result, thinking)
- Calculate token count for long-context detection
- Extract request features (reasoning, image, search, tool loop)
- Preserve original body in `OriginalBody` field

**Error Handling**:
- Malformed JSON → return error
- Missing required fields → return error with specific field name
- Unsupported protocol → return error
- Partial normalization failure → return best-effort normalized request (per FR-001 clarification: route to default)

**Example**:
```go
normalized, err := Normalize(requestBody, "anthropic", "session-123", 32000)
if err != nil {
    // Route to default route
    return handleDefaultRoute(requestBody)
}
// Use normalized.Features for routing
```

---

## 2. Classification API

### Function: `Classify`

**Purpose**: Determine scenario from normalized request

**Signature**:
```go
func (c *BuiltinClassifier) Classify(req *NormalizedRequest, hints *RoutingHints) *RoutingDecision
```

**Parameters**:
- `req` (*NormalizedRequest): Normalized request
- `hints` (*RoutingHints): Optional routing hints from middleware (nil = no hints)

**Returns**:
- `*RoutingDecision`: Routing decision with scenario, source, reason, confidence

**Behavior**:
- Check features in priority order (configurable via `ScenarioPriority`):
  1. `HasWebSearch` → "search"
  2. `HasReasoning` → "reasoning"
  3. `HasImages` → "image"
  4. `IsLongContext` → "long_context"
  5. Model heuristics → "background" or "coding"
- Apply scenario aliases (think→reasoning, webSearch→search)
- Use hints if no strong signal detected
- Set confidence based on signal strength
- Set source to "builtin:classifier"

**Confidence Scoring**:
- `1.0` - Explicit middleware decision (not used by builtin)
- `0.9` - Strong signal (e.g., `HasReasoning=true`)
- `0.7` - Multiple weak signals
- `0.5` - Single weak signal or heuristic
- `0.3` - Fallback/default

**Example**:
```go
decision := classifier.Classify(normalized, ctx.RoutingHints)
// decision.Scenario = "reasoning"
// decision.Source = "builtin:classifier"
// decision.Reason = "thinking mode enabled"
// decision.Confidence = 0.9
```

---

## 3. Resolution API

### Function: `ResolveRoutePolicy`

**Purpose**: Resolve route policy for a scenario

**Signature**:
```go
func ResolveRoutePolicy(scenario string, config *ProfileRoutingConfig) *RoutePolicy
```

**Parameters**:
- `scenario` (string): Scenario key from routing decision
- `config` (*ProfileRoutingConfig): Profile routing configuration

**Returns**:
- `*RoutePolicy`: Route policy for the scenario (never nil)

**Behavior**:
- Normalize scenario key (apply aliases, lowercase kebab-case)
- Lookup in `config.Routes[scenario]`
- If not found, return `config.Default`
- If default not configured, return failover policy with profile providers
- Apply profile-level defaults (strategy, threshold, weights)

**Example**:
```go
policy := ResolveRoutePolicy("plan", profileConfig)
// policy.Providers = [{"name": "p1"}]
// policy.Strategy = "weighted"
// policy.FallbackToDefault = true
```

---

## 4. Middleware Context API

### Type: `RequestContext`

**Purpose**: Middleware request context with routing fields

**New Fields**:
```go
type RequestContext struct {
    // ... existing fields ...

    // NEW: Routing fields
    RequestFormat     string             // Detected protocol
    NormalizedRequest *NormalizedRequest // Protocol-agnostic view
    RoutingDecision   *RoutingDecision   // Explicit decision (binding)
    RoutingHints      *RoutingHints      // Suggestions (non-binding)
}
```

**Contract**:
- `RequestFormat` populated before middleware pipeline
- `NormalizedRequest` populated before middleware pipeline (nil if normalization failed)
- `RoutingDecision` may be set by middleware (last middleware wins)
- `RoutingHints` may be set by middleware (accumulated, not overwritten)
- Middleware MUST NOT modify `NormalizedRequest` (read-only)

**Middleware Behavior**:
```go
func (m *MyMiddleware) ProcessRequest(ctx *RequestContext) (*RequestContext, error) {
    // Read normalized request
    if ctx.NormalizedRequest != nil {
        // Analyze request features
        if ctx.NormalizedRequest.Features.HasReasoning {
            // Set explicit decision
            ctx.RoutingDecision = &RoutingDecision{
                Scenario:   "plan",
                Source:     "middleware:my-middleware",
                Reason:     "detected planning task",
                Confidence: 1.0,
            }
        }
    }

    // Or provide hints
    ctx.RoutingHints = &RoutingHints{
        ScenarioCandidates: []string{"plan", "coding"},
        CostClass:          "high",
    }

    return ctx, nil
}
```

---

## 5. Config Validation API

### Function: `ValidateRoutingConfig`

**Purpose**: Validate routing configuration at load time

**Signature**:
```go
func ValidateRoutingConfig(pc *ProfileConfig, providers map[string]*ProviderConfig) error
```

**Parameters**:
- `pc` (*ProfileConfig): Profile configuration to validate
- `providers` (map[string]*ProviderConfig): Available providers

**Returns**:
- `error`: Validation error with structured message (nil = valid)

**Validation Rules**:
1. Scenario keys must be valid format (alphanumeric + `-` or `_`, max 64 chars)
2. Route policies must not be nil
3. Provider names in routes must exist in `providers` map
4. Strategies must be valid enum values
5. Weights must be non-negative
6. Weighted strategy requires `provider_weights`
7. Thresholds must be positive if set

**Error Format**:
```
routing validation failed:
  - routing["plan"]: provider "nonexistent" does not exist
  - routing["coding"]: weighted strategy requires provider_weights
  - routing["invalid-key!"]: invalid key format (must be alphanumeric with - or _)
```

**Example**:
```go
if err := ValidateRoutingConfig(profileConfig, allProviders); err != nil {
    return fmt.Errorf("config load failed: %w", err)
}
```

---

## 6. Observability API

### Function: `LogRoutingDecision`

**Purpose**: Emit structured log for routing decision

**Signature**:
```go
func LogRoutingDecision(logger *StructuredLogger, decision *RoutingDecision, ctx *RequestContext, selectedProvider string, selectedModel string)
```

**Parameters**:
- `logger` (*StructuredLogger): Structured logger instance
- `decision` (*RoutingDecision): Routing decision
- `ctx` (*RequestContext): Request context
- `selectedProvider` (string): Final provider name
- `selectedModel` (string): Final model name

**Log Fields**:
```json
{
  "level": "info",
  "event": "routing_decision",
  "profile": "default",
  "session_id": "session-123",
  "request_format": "anthropic",
  "scenario": "reasoning",
  "decision_source": "middleware:spec-kit",
  "decision_reason": "detected planning task",
  "confidence": 1.0,
  "provider_selected": "p1",
  "model_selected": "claude-opus-4",
  "has_reasoning": true,
  "has_image": false,
  "has_web_search": false,
  "token_estimate": 5000
}
```

---

## 7. Backward Compatibility

### Scenario Aliases

**Mapping**:
```go
var ScenarioAliases = map[string]string{
    "think":       "reasoning",
    "webSearch":   "search",
    "longContext": "long_context",
    "code":        "coding",
}
```

**Behavior**:
- Old scenario keys automatically mapped to new keys
- Both old and new keys accepted in config
- Logs use new canonical keys

### Config Migration

**Version 14 → 15**:
- `routing` map keys preserved
- `ScenarioRoute` values converted to `RoutePolicy`
- Profile-level `strategy` inherited by routes
- Top-level `providers` migrated to `routing.default.providers`

**Migration Function**:
```go
func MigrateV14ToV15(v14Config *OpenCCConfig) *OpenCCConfig {
    // Bump version
    v14Config.Version = 15

    // Migrate each profile
    for _, profile := range v14Config.Profiles {
        if profile.Routing != nil {
            // Convert ScenarioRoute to RoutePolicy
            for key, route := range profile.Routing {
                profile.Routing[key] = &RoutePolicy{
                    Providers: route.Providers,
                    Strategy:  profile.Strategy, // inherit
                }
            }
        }
    }

    return v14Config
}
```

---

## 8. Error Handling Contract

### Normalization Errors

**Behavior**: Route to default route (per FR-001 clarification)

**Example**:
```go
normalized, err := Normalize(body, protocol, sessionID, threshold)
if err != nil {
    // Don't fail request, route to default
    return routeToDefault(body, profile)
}
```

### Invalid Scenario

**Behavior**: Fall back to default route

**Example**:
```go
policy := ResolveRoutePolicy(decision.Scenario, config)
if policy == nil {
    // Should never happen (ResolveRoutePolicy always returns non-nil)
    policy = config.Default
}
```

### All Providers Failed

**Behavior**: Override `FallbackToDefault=false` and force attempt default route (per FR-010 clarification)

**Example**:
```go
success := tryProviders(policy.Providers)
if !success {
    if policy.ShouldFallback() || true { // Always fallback on total failure
        tryDefaultRoute()
    }
}
```

---

## 9. Performance Contract

**Normalization**:
- Target latency: < 10ms per request
- Memory allocation: O(n) where n = message count
- No blocking I/O

**Classification**:
- Target latency: < 5ms per request
- Memory allocation: O(1) for decision
- No blocking I/O

**Resolution**:
- Target latency: < 1ms per request
- Memory allocation: O(1) for policy lookup
- No blocking I/O

**Total Routing Overhead**:
- Target: < 20ms per request (normalization + classification + resolution)
- Measured at p95 latency

---

## 10. Thread Safety

**Immutable Types**:
- `NormalizedRequest` - read-only after creation
- `RoutingDecision` - read-only after creation
- `RoutePolicy` - read-only during request processing

**Mutable Types**:
- `RequestContext` - modified by middleware pipeline (sequential, no concurrent access)
- `ProfileRoutingConfig` - loaded once, read-only during requests

**Concurrency**:
- Normalization: thread-safe (no shared state)
- Classification: thread-safe (no shared state)
- Resolution: thread-safe (read-only config)
- Middleware pipeline: sequential execution (no concurrent modification of RequestContext)
