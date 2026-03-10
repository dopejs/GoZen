# Scenario Routing Architecture Review and Redesign

## Purpose

This document captures:

- the current review conclusion for scenario routing
- the concrete bugs and architectural gaps found in the current implementation
- the target design for turning scenario routing into a general daemon proxy capability
- the required middleware extension model
- a complete implementation target suitable for handing off to Claude for one-shot development

This is not a phased plan. The intended use is full implementation in one development pass.

## Target Product Goal

Scenario routing is not just a convenience feature for Claude-native requests.

It is intended to become a core daemon proxy capability that:

- routes different task types to different providers and models
- reduces token cost by matching the right model to the right scenario
- works across different client protocols
- remains extensible through middleware

Examples:

- planning → Opus / high-quality model
- coding → low-cost coding-capable provider
- image → image-capable provider
- long-context → cheaper long-context model
- spec-kit flow:
  - `specify`
  - `clarify`
  - `plan`
  - `tasks`
  - `analyse`
  - `implement`
  each routed to different models via a middleware plugin

Under this goal, the current implementation is not complete enough.

## Current Review Conclusion

Current conclusion: **Changes requested**

Reason:

- current scenario routing is useful for Anthropic-style request splitting
- but it is not yet a general routing layer
- and it does not yet provide a first-class extension mechanism for middleware-driven scenarios

## Current Implementation Review

### What already works

- profile-level routing config can select a scenario-specific provider chain
- per-provider model override works within a scenario route
- scenario route can fail over to default providers
- disabled providers are filtered before strategy selection
- load-balancing strategy can be applied after scenario route selection

Relevant files:

- `internal/proxy/profile_proxy.go`
- `internal/proxy/server.go`
- `internal/proxy/scenario.go`
- `internal/config/config.go`

### Core problems

#### 1. Scenario detection is protocol-specific, not semantic

Current routing is based on parsing the raw request body and applying hardcoded checks:

- `thinking`
- Anthropic image content blocks
- `tools[].type` with `web_search*`
- `model` containing `claude` + `haiku`

This means routing is coupled to one request shape instead of to normalized task semantics.

Current consequence:

- OpenAI Chat requests are not classified equivalently
- OpenAI Responses requests are not classified equivalently
- image/search/reasoning/background signals from non-Anthropic clients are under-detected or not detected at all

This is the largest functional gap.

#### 2. Middleware cannot actually drive routing

Middleware runs before routing, but routing ignores middleware output.

Current behavior:

- middleware receives `RequestContext`
- middleware can mutate `ctx.Body` and `ctx.Metadata`
- router then ignores routing metadata and re-detects scenario directly from `bodyBytes`

Current consequence:

- middleware cannot explicitly say: `scenario = plan`
- middleware cannot supply confidence, reason, or route hints
- middleware cannot install new scenario classes
- middleware can only indirectly manipulate body shape and hope the builtin detector picks it up

That is not a real extension API.

#### 3. Scenario space is closed and hardcoded

Current scenario identifiers are fixed enum-like constants:

- `think`
- `image`
- `longContext`
- `webSearch`
- `background`
- `code`
- `default`

Current consequence:

- custom scenarios such as `specify`, `clarify`, `plan`, `tasks`, `analyse`, `implement` cannot be expressed naturally
- adding new scenarios requires code changes in core routing logic
- third-party middleware cannot register new route keys as data

#### 4. Routing model is too weak for the intended product use

Current `ScenarioRoute` only expresses:

- ordered providers
- optional per-provider model override

This is insufficient for long-term routing goals.

Missing capabilities:

- per-scenario strategy
- per-scenario weights
- per-scenario threshold overrides
- per-scenario fallback policy
- middleware-provided route hints
- route-specific provider filtering

#### 5. `default` semantics are ambiguous

Current code keeps both:

- top-level default providers
- `ScenarioDefault`

But normal requests usually resolve to `code` or `background`, which makes `default` effectively rare or unreachable for valid traffic.

Current consequence:

- config semantics are unclear
- users may incorrectly assume `routing.default` is a normal scenario route
- runtime behavior and configuration vocabulary are misaligned

#### 6. Config validation is too weak

The routing config is structurally permissive.

Missing hard validation:

- unknown scenario keys
- empty provider list for a route
- provider referenced by route but absent from profile
- illegal `default` route usage
- invalid per-scenario weights
- invalid per-scenario strategy

For a cost-optimization routing system, silent invalid config is not acceptable.

#### 7. Middleware context is not populated enough for routing plugins

`RequestContext` has fields such as:

- `Profile`
- `Provider`
- `ProjectPath`

But the proxy path that runs middleware currently does not populate enough routing-relevant context for decision plugins.

At minimum, a routing middleware needs:

- `Profile`
- request protocol / request format
- normalized request semantics
- original path
- session id
- client type

Without those fields, middleware authors cannot make reliable routing decisions.

## Current Bugs and Product Risks

These are not hypothetical design concerns. They directly affect the target product.

### Bug 1: Non-Anthropic requests cannot be routed consistently

A request may be semantically:

- planning
- image
- search
- reasoning

but if it arrives in OpenAI Chat or OpenAI Responses shape, current detection may miss it.

Impact:

- wrong provider selected
- wrong model selected
- expected cost optimization does not happen

### Bug 2: Middleware-based scenario routing is effectively impossible

A spec-kit middleware cannot reliably assign:

- `specify`
- `clarify`
- `plan`
- `tasks`
- `analyse`
- `implement`

because there is no first-class output channel from middleware to router.

Impact:

- extension promise is not real yet
- plugin authors must depend on brittle request rewrites

### Bug 3: `default` route semantics are confusing

Current naming suggests `default` is part of scenario routing, but in practice top-level providers already serve as the default route.

Impact:

- configuration confusion
- maintenance complexity
- harder long-term API design

### Bug 4: Per-scenario routing policy is underpowered

You want to optimize cost and capability by scenario.

Current model cannot express:

- `plan` → weighted between two expensive models
- `implement` → least-cost among coding providers
- `analyse` → reasoning model with dedicated fallback

Impact:

- capability ceiling is low
- many future routing strategies require another config redesign

## Required Target State

The target system should satisfy all of the following:

### 1. Protocol-agnostic routing

Scenario routing must work from a normalized semantic request model, not raw protocol-specific JSON.

Supported client protocol families should include at least:

- Anthropic Messages
- OpenAI Chat Completions
- OpenAI Responses

### 2. First-class middleware extensibility

Middleware must be able to:

- emit a scenario key
- emit routing hints
- override builtin classification
- attach explanation and confidence

Builtin routing should become fallback behavior, not the only decision source.

### 3. Open scenario namespace

Scenario keys must support custom names.

Builtin scenarios remain supported, but the system must not require compile-time registration for every new route key.

### 4. Per-scenario policy richness

Each scenario route must be able to define:

- providers
- model overrides
- strategy
- weights
- threshold override
- fallback policy

### 5. Strong config validation

Invalid routing configuration must fail early at load time.

### 6. Good observability

Every routed request should log:

- normalized request features
- decision source
- selected scenario
- fallback behavior
- final provider/model chosen

## Proposed Architecture

## A. Introduce a normalized request layer

Add a new internal type:

```go
type NormalizedRequest struct {
    RequestFormat string
    EndpointKind  string
    Stream        bool

    Model         string
    System        []string
    Messages      []NormalizedMessage
    Tools         []NormalizedTool

    Features      RequestFeatures
    RawBody       []byte
}

type NormalizedMessage struct {
    Role    string
    Blocks  []NormalizedBlock
}

type NormalizedBlock struct {
    Type      string
    Text      string
    ImageURL  string
    MediaType string
    ToolName  string
    ToolID    string
    Input     map[string]interface{}
    Output    string
}

type RequestFeatures struct {
    HasReasoning     bool
    HasImage         bool
    HasWebSearch     bool
    HasToolLoop      bool
    IsBackgroundLike bool
    IsLongContext    bool
    TokenEstimate    int
}
```

Rules:

- normalize Anthropic / OpenAI Chat / OpenAI Responses into one semantic view
- long-context detection runs on normalized content
- image/search/reasoning detection runs on normalized features
- routing no longer depends on provider-specific field names

Recommended file additions:

- `internal/proxy/routing_normalize.go`
- `internal/proxy/routing_normalize_test.go`

## B. Split routing decision from builtin classification

Introduce:

```go
type RoutingDecision struct {
    Scenario   string
    Source     string
    Reason     string
    Confidence float64

    ModelHint         string
    StrategyOverride  config.LoadBalanceStrategy
    ThresholdOverride int

    ProviderAllowlist []string
    ProviderDenylist  []string
    Metadata          map[string]interface{}
}
```

Decision precedence:

1. explicit middleware decision
2. builtin classifier on `NormalizedRequest`
3. default route

Builtin classifier should return decisions such as:

- `reasoning`
- `image`
- `search`
- `long_context`
- `background`
- `coding`

Do not keep the current Anthropic-product-centric names as the only semantic layer.

Backward compatibility can map:

- `think` → `reasoning`
- `webSearch` → `search`
- `longContext` → `long_context`
- `code` → `coding`

## C. Add first-class middleware routing hooks

Do not rely on `Metadata["..."]` as the only contract.

Extend middleware request context with explicit routing fields:

```go
type RequestContext struct {
    SessionID     string
    Profile       string
    ClientType    string
    RequestFormat string
    Method        string
    Path          string
    Headers       http.Header
    Body          []byte

    Model         string
    Messages      []Message

    NormalizedRequest *NormalizedRequest
    RoutingDecision   *RoutingDecision
    RoutingHints      *RoutingHints

    Metadata map[string]interface{}
}

type RoutingHints struct {
    ScenarioCandidates []string
    Tags               []string
    CostClass          string
    CapabilityNeeds    []string
}
```

Rules:

- middleware may set `RoutingDecision`
- middleware may add `RoutingHints`
- router must consume these fields directly
- builtin detector runs only if `RoutingDecision == nil`

This is the key change needed for spec-kit middleware support.

### Example: spec-kit middleware behavior

A `spec-kit-routing` middleware should be able to detect:

- `specify`
- `clarify`
- `plan`
- `tasks`
- `analyse`
- `implement`

and set:

```go
ctx.RoutingDecision = &RoutingDecision{
    Scenario:   "plan",
    Source:     "middleware:spec-kit-routing",
    Reason:     "detected spec-kit planning stage",
    Confidence: 0.95,
}
```

Then router resolves the `plan` route directly from config.

## D. Redesign routing config as open scenario keys

Replace the current fixed `map[Scenario]*ScenarioRoute` design with an open-key route map.

Recommended model:

```go
type ProfileRoutingConfig struct {
    Default *RoutePolicy `json:"default,omitempty"`
    Routes  map[string]*RoutePolicy `json:"routes,omitempty"`
}

type RoutePolicy struct {
    Providers            []*ProviderRoute         `json:"providers"`
    Strategy             LoadBalanceStrategy      `json:"strategy,omitempty"`
    ProviderWeights      map[string]int           `json:"provider_weights,omitempty"`
    LongContextThreshold int                      `json:"long_context_threshold,omitempty"`
    FallbackToDefault    *bool                    `json:"fallback_to_default,omitempty"`
}

type ProviderRoute struct {
    Name  string `json:"name"`
    Model string `json:"model,omitempty"`
}
```

Key points:

- route keys are strings
- builtin routes and custom middleware routes use the same namespace
- each route can define its own strategy and weights
- top-level profile default and route map semantics become unambiguous

### Recommended semantics

- `routing.default` is the only default route
- top-level `providers` remains supported only as legacy config
- legacy config should be migrated into `routing.default`

If full config migration is too large for one change, keep current top-level `providers` as the default route internally, but do not keep `ScenarioDefault` as a runtime routing class.

## E. Route resolution algorithm

The routing flow should become:

1. parse request path and request format
2. build normalized request
3. run middleware pipeline
4. read `RoutingDecision` from middleware if present
5. else run builtin classifier on normalized request
6. resolve route policy by scenario key
7. if no route policy exists, use default route
8. apply scenario-level strategy and weights
9. try scenario providers
10. if route policy allows fallback, try default route
11. log final routing decision and provider/model result

Pseudo-code:

```go
normalized := NormalizeRequest(bodyBytes, requestFormat, sessionID, threshold)
reqCtx.NormalizedRequest = normalized

reqCtx = pipeline.ProcessRequest(reqCtx)

decision := reqCtx.RoutingDecision
if decision == nil {
    decision = builtinClassifier.Classify(normalized)
}

policy := resolveRoutePolicy(profileConfig, decision)
providers := applyRoutePolicy(policy, profileProviders)
providers = filterDisabledProviders(providers)
providers = loadBalancer.Select(providers, policy.Strategy, normalized.Model, profileName, policy.ProviderWeights, policy.ModelOverrides)

success := tryProviders(providers)
if !success && policy.FallbackToDefault {
    tryDefaultRoute()
}
```

## F. Validation rules

Configuration validation must enforce:

- route key cannot be empty
- default route cannot be empty
- all referenced providers must exist
- route providers should normally be a subset of profile-known providers unless explicitly allowed
- provider weights only valid for providers in the route
- weight must be non-negative
- strategy must be valid
- deprecated legacy keys should warn loudly

If a route is invalid, config load should fail instead of silently continuing.

## G. Observability requirements

Add structured logs for:

- `routing_normalized`
- `routing_decision`
- `routing_fallback`
- `routing_policy_selected`
- `routing_provider_selected`

Example fields:

- `profile`
- `client_type`
- `request_format`
- `scenario`
- `decision_source`
- `decision_reason`
- `fallback_used`
- `providers_considered`
- `provider_selected`
- `model_selected`

## H. Backward compatibility

Backward compatibility is required, but only as migration support.

Support old config:

- top-level `providers`
- existing `routing` map keyed by builtin scenarios
- old scenario names

Internally convert old config into the new route-policy model.

Recommended builtin alias map:

- `think` → `reasoning`
- `image` → `image`
- `longContext` → `long_context`
- `webSearch` → `search`
- `background` → `background`
- `code` → `coding`

Do not keep `default` as a normal classifier output.

## Required Code Changes

This is the recommended one-shot implementation scope.

### 1. Routing normalization

Add:

- request normalization for Anthropic / OpenAI Chat / OpenAI Responses
- normalized feature extraction
- tests for all supported request families

### 2. Middleware routing API

Change:

- `internal/middleware/interface.go`
- `internal/proxy/server.go`

Required:

- populate `Profile`
- populate `RequestFormat`
- attach `NormalizedRequest`
- allow middleware to emit `RoutingDecision`

### 3. Builtin classifier refactor

Replace current hardcoded `DetectScenario()` behavior with:

- builtin classifier over normalized request
- no direct dependence on Anthropic-only field names

Recommended file split:

- `internal/proxy/routing_classifier.go`
- `internal/proxy/routing_classifier_test.go`

### 4. Config model upgrade

Change:

- `internal/config/config.go`
- `internal/config/store.go`
- config tests

Required:

- open route keys
- per-route strategy
- per-route weights
- per-route threshold override
- fallback policy
- migration from legacy config

### 5. Runtime route resolution

Change:

- `internal/proxy/profile_proxy.go`
- `internal/proxy/server.go`
- `internal/proxy/loadbalancer.go`

Required:

- resolve route policy by string route key
- apply route-specific strategy and weights
- preserve model overrides
- preserve default fallback behavior

### 6. Validation

Add strict config validation for routing.

### 7. Observability

Add structured logs and tests for routing decisions.

## Required Test Matrix

Claude should implement tests for all of the following.

### Protocol coverage

- Anthropic Messages → builtin reasoning/image/search/background/long-context detection
- OpenAI Chat → equivalent detection
- OpenAI Responses → equivalent detection

### Middleware-driven routing

- middleware sets `scenario = plan` → `plan` route used
- middleware sets `scenario = implement` → `implement` route used
- middleware output overrides builtin classifier
- middleware absent → builtin classifier still works

### Config behavior

- custom route key config loads successfully
- invalid route key config fails clearly
- invalid provider in route fails clearly
- invalid per-route weights fail clearly
- legacy routing config still migrates correctly

### Runtime behavior

- scenario route uses its own strategy
- scenario route uses its own weights
- scenario route uses its own model override
- scenario route falls back to default when configured
- scenario route does not fall back when disabled

### Product scenarios

- planning route goes to high-quality provider
- coding route goes to low-cost provider
- long-context route goes to cheaper long-context model
- spec-kit:
  - `specify`
  - `clarify`
  - `plan`
  - `tasks`
  - `analyse`
  - `implement`
  all route correctly

## Acceptance Criteria

The redesign is complete only if all of the following are true:

- scenario routing works the same for Anthropic, OpenAI Chat, and OpenAI Responses clients
- middleware can explicitly choose a route without body-shape hacks
- custom route keys can be introduced without modifying core classifier enums
- each route can independently define providers, model overrides, strategy, and weights
- config validation fails fast on invalid routing config
- structured logs explain why a request was routed the way it was
- legacy config remains readable and migratable

## Recommended Naming

Use stable semantic names instead of provider-specific names.

Recommended builtin route keys:

- `reasoning`
- `image`
- `search`
- `long_context`
- `background`
- `coding`

Recommended custom route keys for middleware:

- `specify`
- `clarify`
- `plan`
- `tasks`
- `analyse`
- `implement`

## Final Assessment

The current implementation is a useful foundation, but it is still a builtin rule-based scenario splitter.

It is not yet:

- protocol-agnostic
- middleware-extensible
- semantically open
- strong enough for cost-optimized multi-model routing

For the product direction described here, the correct next step is not incremental patching of the current detector.

The correct next step is a full routing-layer redesign around:

- normalized requests
- explicit routing decisions
- open route keys
- route-policy config
- middleware-driven extensibility

