# Quickstart: Scenario Routing Architecture Redesign

**Feature**: 020-scenario-routing-redesign
**Date**: 2026-03-10
**Purpose**: Quick reference for implementing protocol-agnostic, middleware-extensible routing

## Overview

This feature redesigns GoZen's scenario routing to:
1. Support multiple API protocols (Anthropic, OpenAI Chat, OpenAI Responses)
2. Allow middleware to drive routing decisions
3. Support custom scenario routes without code changes
4. Enable per-scenario routing policies (strategy, weights, thresholds)

## Implementation Phases

### Phase 1: Normalization Layer (Days 1-2)

**Files to Create**:
- `internal/proxy/routing_normalize.go`
- `internal/proxy/routing_normalize_test.go`

**Key Functions**:
```go
func Normalize(body []byte, protocol string, sessionID string, threshold int) (*NormalizedRequest, error)
func DetectProtocol(path string, headers http.Header, body map[string]interface{}) string
func ExtractFeatures(req *NormalizedRequest) RequestFeatures
```

**Tests to Write**:
- Anthropic Messages normalization
- OpenAI Chat normalization
- OpenAI Responses normalization
- Malformed request handling
- Feature extraction accuracy

**Success Criteria**:
- All three protocols normalize correctly
- Token counting works for long-context detection
- Test coverage ≥ 80%

---

### Phase 2: Config Migration (Days 2-3)

**Files to Modify**:
- `internal/config/config.go` (bump version to 15, add new types)
- `internal/config/store.go` (add validation)
- `internal/config/config_test.go` (add migration tests)

**Key Changes**:
```go
const CurrentConfigVersion = 15

type ProfileRoutingConfig struct {
    Default *RoutePolicy `json:"default,omitempty"`
    Routes  map[string]*RoutePolicy `json:"routes,omitempty"`
}

type RoutePolicy struct {
    Providers            []*ProviderRoute `json:"providers"`
    Strategy             LoadBalanceStrategy `json:"strategy,omitempty"`
    ProviderWeights      map[string]int `json:"provider_weights,omitempty"`
    LongContextThreshold *int `json:"long_context_threshold,omitempty"`
    FallbackToDefault    *bool `json:"fallback_to_default,omitempty"`
}
```

**Tests to Write**:
- v14→v15 migration
- Mixed legacy/custom scenario keys
- Config validation (invalid providers, empty routes, bad weights)
- Scenario alias mapping

**Success Criteria**:
- Legacy configs migrate automatically
- Invalid configs fail fast with clear errors
- Test coverage ≥ 80%

---

### Phase 3: Routing Decision Types (Day 3)

**Files to Create**:
- `internal/proxy/routing_decision.go`

**Files to Modify**:
- `internal/middleware/interface.go` (extend RequestContext)

**Key Types**:
```go
type RoutingDecision struct {
    Scenario   string
    Source     string
    Reason     string
    Confidence float64
    ModelHint         *string
    StrategyOverride  *LoadBalanceStrategy
    ThresholdOverride *int
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

**Tests to Write**:
- Decision validation
- Confidence scoring
- Pointer field handling (nil vs zero value)

**Success Criteria**:
- Types compile and serialize correctly
- Validation catches invalid decisions
- Test coverage ≥ 80%

---

### Phase 4: Builtin Classifier Refactor (Days 4-5)

**Files to Create**:
- `internal/proxy/routing_classifier.go`
- `internal/proxy/routing_classifier_test.go`
- `internal/proxy/routing_resolver.go`
- `internal/proxy/routing_resolver_test.go`

**Files to Modify**:
- `internal/proxy/scenario.go` (refactor to use new classifier)
- `internal/proxy/scenario_test.go`

**Key Functions**:
```go
func (c *BuiltinClassifier) Classify(req *NormalizedRequest, hints *RoutingHints) *RoutingDecision
func ResolveRoutePolicy(scenario string, config *ProfileRoutingConfig) *RoutePolicy
func NormalizeScenarioKey(key string) string
```

**Tests to Write**:
- Protocol-agnostic feature detection
- Confidence scoring for different signals
- Hint integration
- Scenario alias mapping
- Route policy resolution with fallback

**Success Criteria**:
- Same semantic content routes to same scenario across protocols
- Hints influence classification when no strong signal
- Test coverage ≥ 80%

---

### Phase 5: Integration (Days 5-6)

**Files to Modify**:
- `internal/proxy/server.go` (populate RequestContext, integrate normalization)
- `internal/proxy/profile_proxy.go` (use new routing flow)
- `internal/proxy/loadbalancer.go` (accept route-specific overrides)
- `internal/proxy/server_test.go`
- `internal/proxy/profile_proxy_test.go`
- `internal/proxy/loadbalancer_test.go`

**Key Changes**:
```go
// In ProxyServer.ServeHTTP()
protocol := DetectProtocol(r.URL.Path, r.Header, bodyMap)
normalized, err := Normalize(bodyBytes, protocol, sessionID, threshold)
if err != nil {
    // Route to default
}

reqCtx.RequestFormat = protocol
reqCtx.NormalizedRequest = normalized

// Run middleware pipeline
reqCtx = pipeline.ProcessRequest(reqCtx)

// Resolve routing decision
decision := ResolveRoutingDecision(reqCtx, builtinClassifier, "coding")
policy := ResolveRoutePolicy(decision.Scenario, profileConfig)

// Apply policy
providers := applyRoutePolicy(policy, profileProviders)
providers = loadBalancer.Select(providers, policy.Strategy, model, profile, policy.ProviderWeights, policy.ModelOverrides)
```

**Tests to Write**:
- End-to-end routing flow
- Middleware decision precedence
- Builtin classifier fallback
- Default route fallback
- Route policy application

**Success Criteria**:
- Requests route correctly through full pipeline
- Middleware can override builtin classifier
- Test coverage ≥ 80%

---

### Phase 6: Integration Tests (Days 6-7)

**Files to Create**:
- `tests/integration/routing_protocol_test.go`
- `tests/integration/routing_middleware_test.go`
- `tests/integration/routing_policy_test.go`

**Test Scenarios**:
1. **Protocol-agnostic routing**: Same semantic request via Anthropic/OpenAI Chat/OpenAI Responses routes to same scenario
2. **Middleware-driven routing**: Test middleware sets custom scenario, request routes correctly
3. **Per-scenario policies**: Different scenarios use different strategies (weighted, least-cost, etc.)
4. **Config validation**: Invalid configs fail at daemon startup
5. **Fallback behavior**: Scenario route failure falls back to default
6. **Observability**: Routing decisions logged with correct fields

**Success Criteria**:
- All integration tests pass
- Test coverage ≥ 80% for new code
- No regressions in existing tests

---

## Quick Reference

### Adding a Custom Scenario Route

**Config** (`~/.zen/zen.json`):
```json
{
  "profiles": {
    "default": {
      "routing": {
        "my-custom-scenario": {
          "providers": [{"name": "provider1", "model": "claude-opus-4"}],
          "strategy": "weighted",
          "provider_weights": {"provider1": 100},
          "fallback_to_default": true
        }
      }
    }
  }
}
```

**Middleware**:
```go
func (m *MyMiddleware) ProcessRequest(ctx *RequestContext) (*RequestContext, error) {
    ctx.RoutingDecision = &RoutingDecision{
        Scenario:   "my-custom-scenario",
        Source:     "middleware:my-middleware",
        Reason:     "detected custom workflow",
        Confidence: 1.0,
    }
    return ctx, nil
}
```

---

### Debugging Routing Decisions

**Check logs** for `routing_decision` events:
```json
{
  "event": "routing_decision",
  "scenario": "reasoning",
  "decision_source": "middleware:spec-kit",
  "decision_reason": "detected planning task",
  "confidence": 1.0,
  "provider_selected": "p1",
  "model_selected": "claude-opus-4"
}
```

**Common Issues**:
- Scenario not found → Check config has route for scenario key
- Wrong provider selected → Check route policy strategy and weights
- Middleware decision ignored → Check middleware order (last wins)
- Normalization failed → Check request format matches protocol

---

### Testing Checklist

Before opening PR:
- [ ] All unit tests pass: `go test ./internal/proxy ./internal/config`
- [ ] Integration tests pass: `go test ./tests/integration`
- [ ] Coverage ≥ 80%: `go test -cover ./internal/proxy ./internal/config`
- [ ] No regressions: `go test ./...`
- [ ] Config migration tested with real v14 config
- [ ] All three protocols tested (Anthropic, OpenAI Chat, OpenAI Responses)
- [ ] Middleware precedence tested
- [ ] Invalid config validation tested
- [ ] Observability logs verified

---

## Common Patterns

### Pattern 1: Protocol Detection

```go
func DetectProtocol(path string, headers http.Header, body map[string]interface{}) string {
    // Primary: URL path
    if strings.HasSuffix(path, "/messages") {
        return "anthropic"
    }
    if strings.HasSuffix(path, "/chat/completions") {
        return "openai_chat"
    }
    if strings.HasSuffix(path, "/responses") {
        return "openai_responses"
    }

    // Fallback: body structure
    if _, hasInput := body["input"]; hasInput {
        return "openai_responses"
    }
    if _, hasSystem := body["system"]; hasSystem {
        return "anthropic"
    }

    return "openai_chat" // default
}
```

### Pattern 2: Middleware Decision

```go
func (m *SpecKitMiddleware) ProcessRequest(ctx *RequestContext) (*RequestContext, error) {
    stage := detectSpecKitStage(ctx)
    if stage != "" {
        ctx.RoutingDecision = &RoutingDecision{
            Scenario:   stage, // "plan", "implement", etc.
            Source:     "middleware:spec-kit",
            Reason:     fmt.Sprintf("detected spec-kit stage: %s", stage),
            Confidence: 1.0,
        }
    }
    return ctx, nil
}
```

### Pattern 3: Config Validation

```go
func ValidateRoutingConfig(pc *ProfileConfig, providers map[string]*ProviderConfig) error {
    var errs []string
    for scenarioKey, policy := range pc.Routing {
        if !isValidScenarioKey(scenarioKey) {
            errs = append(errs, fmt.Sprintf("routing[%q]: invalid key format", scenarioKey))
        }
        for _, pr := range policy.Providers {
            if _, exists := providers[pr.Name]; !exists {
                errs = append(errs, fmt.Sprintf("routing[%q]: provider %q does not exist", scenarioKey, pr.Name))
            }
        }
    }
    if len(errs) > 0 {
        return fmt.Errorf("routing validation failed:\n  - %s", strings.Join(errs, "\n  - "))
    }
    return nil
}
```

---

## Performance Tips

1. **Lazy parsing**: Only parse fields needed for routing, defer full parsing
2. **Cache protocol detection**: Store result in RequestContext
3. **Avoid unnecessary normalization**: Skip if middleware provides explicit decision
4. **Reuse buffers**: Pool byte buffers for JSON parsing
5. **Profile hot path**: Use `go test -bench` to identify bottlenecks

---

## Rollback Plan

If issues arise after deployment:

1. **Config rollback**: Revert to v14 config format (automatic migration on next load)
2. **Feature flag**: Add `GOZEN_DISABLE_NEW_ROUTING=1` env var to use legacy routing
3. **Gradual rollout**: Deploy to dev environment first, monitor for 24 hours
4. **Monitoring**: Watch for increased latency, routing errors, config validation failures

---

## Next Steps

After implementation:
1. Run `/speckit.tasks` to generate detailed task breakdown
2. Implement tasks in order (normalization → config → types → classifier → integration)
3. Write tests first (TDD per Constitution I)
4. Commit each logical unit separately (per Constitution IV)
5. Verify coverage before opening PR (per Constitution VI)
6. Update CLAUDE.md with new routing patterns
