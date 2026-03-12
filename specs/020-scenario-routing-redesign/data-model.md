# Data Model: Scenario Routing Architecture Redesign

**Feature**: 020-scenario-routing-redesign
**Date**: 2026-03-10
**Purpose**: Define data structures and relationships for protocol-agnostic, middleware-extensible routing

## Core Entities

### 1. NormalizedRequest

**Purpose**: Protocol-agnostic representation of API requests from Anthropic Messages, OpenAI Chat, and OpenAI Responses

**Fields**:
- `Model` (string): Model identifier (e.g., "claude-opus-4", "gpt-4")
- `MaxTokens` (int): Maximum tokens to generate
- `Temperature` (*float64): Sampling temperature (nil = not set)
- `Stream` (bool): Whether response should be streamed
- `System` (string): Normalized system prompt/instructions
- `Messages` ([]NormalizedMessage): Conversation history
- `Tools` ([]NormalizedTool): Available tools/functions
- `ToolChoice` (string): Tool selection mode ("auto", "any", "none", or tool name)
- `Thinking` (*ThinkingConfig): Reasoning/thinking configuration (nil = disabled)
- `Features` (RequestFeatures): Extracted semantic features for routing
- `OriginalBody` (map[string]interface{}): Preserved original request for passthrough

**Relationships**:
- Contains multiple `NormalizedMessage` (1:N)
- Contains multiple `NormalizedTool` (1:N)
- Contains one `RequestFeatures` (1:1)
- Contains optional `ThinkingConfig` (1:0..1)

**Validation Rules**:
- `Model` must not be empty
- `MaxTokens` must be positive if set
- `Temperature` must be in range [0.0, 2.0] if set
- `Messages` must not be empty

**State Transitions**: Immutable after creation (read-only)

---

### 2. NormalizedMessage

**Purpose**: Protocol-agnostic message representation

**Fields**:
- `Role` (string): Message role ("user", "assistant")
- `Content` ([]ContentBlock): Message content blocks

**Relationships**:
- Belongs to `NormalizedRequest` (N:1)
- Contains multiple `ContentBlock` (1:N)

**Validation Rules**:
- `Role` must be "user" or "assistant"
- `Content` must not be empty

---

### 3. ContentBlock

**Purpose**: Unified content representation across protocols

**Fields**:
- `Type` (string): Block type ("text", "image", "tool_use", "tool_result", "thinking")
- `Text` (string): Text content (for type="text")
- `ImageSource` (*ImageSource): Image data (for type="image")
- `ToolUseID` (string): Tool invocation ID (for type="tool_use")
- `ToolName` (string): Tool name (for type="tool_use")
- `ToolInput` (map[string]interface{}): Tool parameters (for type="tool_use")
- `ToolResultID` (string): Tool result ID (for type="tool_result")
- `ToolContent` (interface{}): Tool output (for type="tool_result")
- `ThinkingText` (string): Reasoning content (for type="thinking")
- `Signature` (string): Thinking signature (for type="thinking")

**Relationships**:
- Belongs to `NormalizedMessage` (N:1)
- Contains optional `ImageSource` (1:0..1)

**Validation Rules**:
- `Type` must be one of: "text", "image", "tool_use", "tool_result", "thinking"
- Type-specific fields must be populated based on `Type`

---

### 4. RequestFeatures

**Purpose**: Extracted semantic features for scenario classification

**Fields**:
- `HasReasoning` (bool): Request includes thinking/reasoning mode
- `HasImages` (bool): Request contains image content
- `HasWebSearch` (bool): Request uses web search tools
- `HasToolLoop` (bool): Request involves tool use
- `IsLongContext` (bool): Request exceeds long-context threshold
- `TokenCount` (int): Estimated token count
- `ToolCount` (int): Number of tools available

**Relationships**:
- Belongs to `NormalizedRequest` (N:1)

**Validation Rules**:
- `TokenCount` must be non-negative
- `ToolCount` must be non-negative

---

### 5. RoutingDecision

**Purpose**: Explicit routing choice (binding, overrides builtin classifier)

**Fields**:
- `Scenario` (string): Scenario key (e.g., "plan", "coding", "reasoning")
- `Source` (string): Decision source (e.g., "middleware:spec-kit", "builtin:classifier")
- `Reason` (string): Human-readable explanation
- `Confidence` (float64): Confidence score [0.0, 1.0]
- `ModelHint` (*string): Suggested model override (nil = not set)
- `StrategyOverride` (*LoadBalanceStrategy): Strategy override (nil = use route default)
- `ThresholdOverride` (*int): Long-context threshold override (nil = use route default)
- `ProviderAllowlist` ([]string): Only consider these providers (empty = no filter)
- `ProviderDenylist` ([]string): Exclude these providers (empty = no filter)
- `Metadata` (map[string]interface{}): Extensibility for custom fields

**Relationships**:
- Belongs to `RequestContext` (N:0..1)
- References `RoutePolicy` by scenario key (N:0..1)

**Validation Rules**:
- `Scenario` must not be empty
- `Source` must not be empty
- `Confidence` must be in range [0.0, 1.0]
- `ProviderAllowlist` and `ProviderDenylist` must not overlap

**State Transitions**:
- Created by middleware or builtin classifier
- Immutable after creation
- Can be replaced by later middleware (last-wins precedence)

---

### 6. RoutingHints

**Purpose**: Non-binding routing suggestions (influences builtin classifier)

**Fields**:
- `ScenarioCandidates` ([]string): Possible scenarios in priority order
- `Tags` ([]string): Semantic tags (e.g., "high-quality", "fast")
- `CostClass` (string): Cost preference ("low", "medium", "high")
- `CapabilityNeeds` ([]string): Required capabilities (e.g., "vision", "tools")
- `Confidence` (map[string]float64): Per-scenario confidence scores
- `Metadata` (map[string]interface{}): Extensibility

**Relationships**:
- Belongs to `RequestContext` (N:0..1)

**Validation Rules**:
- `CostClass` must be one of: "low", "medium", "high", or empty
- `Confidence` values must be in range [0.0, 1.0]

---

### 7. RoutePolicy

**Purpose**: Per-scenario routing configuration (replaces legacy `ScenarioRoute`)

**Fields**:
- `Providers` ([]*ProviderRoute): Ordered provider list with optional model overrides
- `Strategy` (LoadBalanceStrategy): Load balancing strategy (empty = use profile default)
- `ProviderWeights` (map[string]int): Per-provider weights for weighted strategy
- `LongContextThreshold` (*int): Token threshold for long-context detection (nil = use profile default)
- `FallbackToDefault` (*bool): Whether to fall back to default route on failure (nil = true)

**Relationships**:
- Belongs to `ProfileConfig.Routing` (N:1)
- Contains multiple `ProviderRoute` (1:N)
- References providers by name (N:N)

**Validation Rules**:
- `Providers` must not be empty
- `Strategy` must be valid enum value if set (failover, round-robin, least-latency, least-cost, weighted)
- `ProviderWeights` keys must match provider names in `Providers`
- `ProviderWeights` values must be non-negative
- `LongContextThreshold` must be positive if set

**State Transitions**: Loaded from config, immutable during request processing

**Migration from v14**: Legacy `ScenarioRoute` (only `Providers` field) automatically converted to `RoutePolicy` with default values for new fields

---

### 8. ProfileConfig (Extended)

**Purpose**: Complete profile configuration including routing

**Fields** (routing-related):
- `Providers` ([]string): Default provider list
- `Routing` (map[string]*RoutePolicy): Scenario-specific route policies (key = scenario name)
- `LongContextThreshold` (int): Default token threshold for long-context detection
- `Strategy` (LoadBalanceStrategy): Default load balancing strategy
- `ProviderWeights` (map[string]int): Default per-provider weights

**Relationships**:
- Contains multiple scenario-specific `RoutePolicy` (1:N)

**Validation Rules**:
- All scenario keys must be valid format (alphanumeric + `-` or `_`, max 64 chars)
- All routes must pass `RoutePolicy` validation
- Scenario keys are case-insensitive, normalized to camelCase internally

**Config Version**: v15 (migrated from v14)

---

### 9. RequestContext (Extended)

**Purpose**: Middleware request context with routing fields

**New Fields** (added to existing context):
- `RequestFormat` (string): Detected protocol ("anthropic", "openai_chat", "openai_responses")
- `NormalizedRequest` (*NormalizedRequest): Protocol-agnostic request view
- `RoutingDecision` (*RoutingDecision): Explicit routing decision (binding)
- `RoutingHints` (*RoutingHints): Routing suggestions (non-binding)

**Relationships**:
- Contains one `NormalizedRequest` (1:0..1)
- Contains one `RoutingDecision` (1:0..1)
- Contains one `RoutingHints` (1:0..1)

---

## Entity Relationships Diagram

```
ProfileConfig
├── Providers: []string
├── Strategy: LoadBalanceStrategy
├── ProviderWeights: map[string]int
├── LongContextThreshold: int
└── Routing: map[string]RoutePolicy (1..N)
    └── RoutePolicy
        ├── Providers: []ProviderRoute (1..N)
        ├── Strategy: LoadBalanceStrategy (optional)
        ├── ProviderWeights: map[string]int (optional)
        ├── LongContextThreshold: *int (optional)
        └── FallbackToDefault: *bool (optional)

RequestContext
├── NormalizedRequest (0..1)
│   ├── Messages: []NormalizedMessage (1..N)
│   │   └── Content: []ContentBlock (1..N)
│   │       └── ImageSource (0..1)
│   ├── Tools: []NormalizedTool (0..N)
│   ├── Thinking: ThinkingConfig (0..1)
│   └── Features: RequestFeatures (1)
├── RoutingDecision (0..1)
│   └── references RoutePolicy by scenario key
└── RoutingHints (0..1)
```

---

## Data Flow

### 1. Request Normalization
```
Raw Request (Anthropic/OpenAI Chat/OpenAI Responses)
  → Protocol Detection (URL path, headers, body structure)
  → Normalize() function
  → NormalizedRequest with RequestFeatures
```

### 2. Routing Decision
```
NormalizedRequest
  → Middleware Pipeline (may set RoutingDecision or RoutingHints)
  → Builtin Classifier (if no RoutingDecision)
  → RoutingDecision with scenario, source, reason, confidence
```

### 3. Route Resolution
```
RoutingDecision.Scenario
  → Lookup in ProfileConfig.Routing (map[string]*RoutePolicy)
  → Apply RoutePolicy (providers, strategy, weights, thresholds)
  → Fallback to default providers if scenario not found
```

### 4. Provider Selection
```
RoutePolicy.Providers
  → Filter disabled/unhealthy providers
  → Apply LoadBalanceStrategy (failover, round-robin, least-latency, least-cost, weighted)
  → Select provider and model
```

---

## Config Schema Changes

### Version 14 (Legacy)
```json
{
  "version": 14,
  "profiles": {
    "default": {
      "providers": ["p1", "p2"],
      "routing": {
        "think": {"providers": [{"name": "p1", "model": "claude-opus-4"}]},
        "code": {"providers": [{"name": "p2"}]}
      }
    }
  }
}
```

### Version 15 (New)
```json
{
  "version": 15,
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
        },
        "my-custom-scenario": {
          "providers": [{"name": "p1"}],
          "strategy": "failover",
          "fallback_to_default": true
        }
      }
    }
  }
}
```

**Migration**: v14 `routing` map values converted from `ScenarioRoute` (only `providers` field) to `RoutePolicy` (adds `strategy`, `provider_weights`, `long_context_threshold`, `fallback_to_default` fields with default values)

**Key Changes**:
1. `ScenarioRoute` → `RoutePolicy` (new fields added)
2. Scenario keys remain as strings (no enum constraint)
3. Custom scenario keys supported (e.g., "my-custom-scenario")
4. Per-scenario strategy/weights/threshold now supported

---

## Scenario Aliases

**Mapping** (for backward compatibility):
- `web-search` → `webSearch`
- `long-context` → `longContext`
- `long_context` → `longContext`

**Normalization**: All scenario keys normalized to camelCase internally
- Input: `web-search`, `web_search`, `webSearch` → Normalized: `webSearch`
- Input: `long-context`, `long_context`, `longContext` → Normalized: `longContext`

**Builtin Scenarios** (preserved from v14):
- `think` - Extended thinking mode requests
- `image` - Requests with image content
- `longContext` - Requests exceeding token threshold
- `webSearch` - Requests with web_search tools
- `code` - Regular coding requests
- `background` - Haiku model requests
- `default` - Fallback scenario

---

## Confidence Scoring

**Ranges**:
- `1.0` - Explicit (middleware set scenario)
- `0.9` - High (strong signal, e.g., `thinking=true`)
- `0.7` - Medium (multiple weak signals)
- `0.5` - Low (single weak signal or heuristic)
- `0.3` - Guess (fallback/default)

**Usage**: Logged for observability, not used for routing decisions (decision is binding regardless of confidence)

---

## Observability Fields

**Logged for each routed request**:
- `profile`: Profile name
- `session_id`: Session identifier
- `request_format`: Detected protocol
- `scenario`: Selected scenario
- `decision_source`: Decision source (middleware vs builtin)
- `decision_reason`: Human-readable explanation
- `confidence`: Confidence score
- `provider_selected`: Final provider name
- `model_selected`: Final model name
- `has_reasoning`, `has_image`, `has_web_search`, `token_estimate`: Request features

---

## Performance Characteristics

**Normalization**:
- Time complexity: O(n) where n = number of messages
- Space complexity: O(n) for normalized representation
- Target latency: < 10ms per request

**Route Resolution**:
- Time complexity: O(1) for scenario lookup in map
- Space complexity: O(1) for decision
- Target latency: < 5ms per request

**Config Validation**:
- Time complexity: O(r × p) where r = routes, p = providers per route
- Performed once at config load, not per request
