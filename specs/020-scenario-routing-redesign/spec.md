# Feature Specification: Scenario Routing Architecture Redesign

**Feature Branch**: `020-scenario-routing-redesign`
**Created**: 2026-03-10
**Status**: Draft
**Input**: User description: "Scenario routing architecture redesign for protocol-agnostic, middleware-extensible routing"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Protocol-Agnostic Scenario Detection (Priority: P1)

As a GoZen user, I want scenario routing to work consistently regardless of which API protocol my client uses (Anthropic Messages, OpenAI Chat, or OpenAI Responses), so that I get the same cost optimization and provider selection benefits across all my tools.

**Why this priority**: This is the foundation for all other routing improvements. Without protocol-agnostic detection, the routing system remains limited to Anthropic-native clients and cannot serve as a general proxy capability.

**Independent Test**: Can be fully tested by sending equivalent requests (same semantic content) via different API protocols and verifying they route to the same provider/model. Delivers immediate value by making routing work for OpenAI-compatible clients.

**Acceptance Scenarios**:

1. **Given** a request with reasoning features sent via Anthropic Messages API, **When** the proxy processes it, **Then** it routes to the `reasoning` scenario
2. **Given** an equivalent request with reasoning features sent via OpenAI Chat API, **When** the proxy processes it, **Then** it routes to the same `reasoning` scenario
3. **Given** a request with image content sent via OpenAI Responses API, **When** the proxy processes it, **Then** it routes to the `image` scenario
4. **Given** a long-context request (>32K tokens) sent via any supported protocol, **When** the proxy processes it, **Then** it routes to the `long_context` scenario

---

### User Story 2 - Middleware-Driven Custom Routing (Priority: P1)

As a middleware developer, I want to explicitly set routing decisions from my middleware plugin without manipulating request body shapes, so that I can implement custom routing logic (like spec-kit workflow stages) that the builtin classifier doesn't understand.

**Why this priority**: This enables the core extensibility promise. Without this, middleware cannot truly control routing, making the system closed and requiring core code changes for every new routing scenario.

**Independent Test**: Can be fully tested by creating a test middleware that sets a custom scenario (e.g., "plan") and verifying the request routes to the configured provider for that scenario. Delivers value by enabling spec-kit and other workflow-aware routing.

**Acceptance Scenarios**:

1. **Given** a middleware that sets `RoutingDecision.Scenario = "plan"`, **When** the request is processed, **Then** the proxy uses the `plan` route from config
2. **Given** a middleware that sets `RoutingDecision.Scenario = "implement"`, **When** the request is processed, **Then** the proxy uses the `implement` route from config
3. **Given** a middleware decision and a builtin classifier result, **When** both are present, **Then** the middleware decision takes precedence
4. **Given** no middleware decision, **When** the request is processed, **Then** the builtin classifier runs and provides a scenario
5. **Given** a middleware that sets routing hints but no explicit decision, **When** the builtin classifier runs, **Then** it can use the hints to improve classification

---

### User Story 3 - Open Scenario Namespace (Priority: P2)

As a GoZen administrator, I want to define custom scenario routes in my config (like "specify", "clarify", "plan", "tasks") without modifying GoZen's source code, so that I can optimize routing for my specific workflows.

**Why this priority**: This makes the routing system truly extensible at the configuration level. Users can add new scenarios as their needs evolve without waiting for core updates.

**Independent Test**: Can be fully tested by adding a custom scenario route to the config, having middleware emit that scenario, and verifying the request routes correctly. Delivers value by enabling user-specific workflow optimization.

**Acceptance Scenarios**:

1. **Given** a config with a custom route key "specify", **When** a request is classified as "specify", **Then** the proxy uses the providers and settings from that route
2. **Given** a config with multiple custom routes ("plan", "tasks", "implement"), **When** requests are classified with those scenarios, **Then** each routes to its configured providers
3. **Given** a custom route that doesn't exist in the builtin classifier, **When** middleware emits that scenario, **Then** the routing system accepts and uses it
4. **Given** a request classified with an unknown scenario (no route defined), **When** routing is resolved, **Then** the system falls back to the default route

---

### User Story 4 - Per-Scenario Routing Policies (Priority: P2)

As a GoZen administrator, I want each scenario route to have its own strategy, weights, and model overrides, so that I can fine-tune cost and performance for different task types (e.g., weighted selection for planning, least-cost for coding).

**Why this priority**: This enables sophisticated cost optimization strategies. Different scenarios have different cost/quality tradeoffs, and the routing system should support expressing those differences.

**Independent Test**: Can be fully tested by configuring different strategies for different scenarios and verifying each scenario uses its own policy. Delivers value by enabling per-scenario cost optimization.

**Acceptance Scenarios**:

1. **Given** a "plan" route with strategy "weighted" and custom weights, **When** a planning request is processed, **Then** providers are selected using weighted random distribution
2. **Given** a "coding" route with strategy "least-cost", **When** a coding request is processed, **Then** the cheapest provider is selected
3. **Given** a "reasoning" route with per-provider model overrides, **When** a reasoning request is processed, **Then** the specified models are used for each provider
4. **Given** a "long_context" route with a custom threshold override, **When** token counting is performed, **Then** the route-specific threshold is used instead of the profile default

---

### User Story 5 - Strong Config Validation (Priority: P3)

As a GoZen administrator, I want the system to reject invalid routing configurations at load time with clear error messages, so that I don't discover routing problems during production traffic.

**Why this priority**: This prevents silent failures and configuration mistakes. While less critical than core routing functionality, it significantly improves operational reliability.

**Independent Test**: Can be fully tested by attempting to load various invalid configs and verifying each fails with a specific error message. Delivers value by catching configuration errors early.

**Acceptance Scenarios**:

1. **Given** a route that references a non-existent provider, **When** the config is loaded, **Then** loading fails with an error identifying the missing provider
2. **Given** a route with an empty provider list, **When** the config is loaded, **Then** loading fails with an error about the empty route
3. **Given** a route with invalid weights (negative values), **When** the config is loaded, **Then** loading fails with an error about invalid weights
4. **Given** a route with an invalid strategy name, **When** the config is loaded, **Then** loading fails with an error about the unknown strategy

---

### User Story 6 - Routing Observability (Priority: P3)

As a GoZen administrator, I want structured logs that explain why each request was routed to a specific provider and model, so that I can debug routing issues and verify my cost optimization strategies are working.

**Why this priority**: This enables operational visibility and debugging. While the routing must work correctly first, observability is essential for maintaining and tuning the system.

**Independent Test**: Can be fully tested by processing requests and verifying the expected log entries are emitted with correct fields. Delivers value by making routing decisions transparent.

**Acceptance Scenarios**:

1. **Given** a request that is routed by middleware decision, **When** the request is processed, **Then** logs include the scenario, decision source, and reason
2. **Given** a request that is routed by builtin classifier, **When** the request is processed, **Then** logs include the detected features and classification logic
3. **Given** a request that falls back to the default route, **When** the request is processed, **Then** logs indicate fallback was used and why
4. **Given** a request that tries multiple providers, **When** failover occurs, **Then** logs show the provider chain and failure reasons

---

### Edge Cases

- What happens when middleware sets an invalid scenario name that has no configured route?
- How does the system handle requests that match multiple scenario patterns simultaneously?
- What happens when a scenario route's providers are all disabled or unhealthy?
- How does long-context detection work when session history is unavailable?
- What happens when a middleware sets conflicting routing hints?
- How does the system handle protocol normalization for malformed or non-standard requests?
- What happens when a scenario route has fallback disabled and all providers fail?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST normalize Anthropic Messages, OpenAI Chat, and OpenAI Responses requests into a common semantic representation
- **FR-002**: System MUST extract request features (reasoning, image, search, tool loop, long context) from normalized requests regardless of protocol
- **FR-003**: System MUST allow middleware to set explicit routing decisions via `RoutingDecision` field in `RequestContext`
- **FR-004**: System MUST prioritize middleware routing decisions over builtin classifier results
- **FR-005**: System MUST run builtin classifier only when middleware does not provide a routing decision
- **FR-006**: System MUST support custom scenario route keys defined in configuration without code changes
- **FR-007**: System MUST support builtin scenario aliases for backward compatibility (think→reasoning, webSearch→search, etc.)
- **FR-008**: System MUST allow each scenario route to define its own provider list, strategy, weights, and model overrides
- **FR-009**: System MUST allow each scenario route to define its own long-context threshold override
- **FR-010**: System MUST allow each scenario route to define whether it falls back to default route on failure
- **FR-011**: System MUST validate routing configuration at load time and fail fast on invalid config
- **FR-012**: System MUST reject routes that reference non-existent providers
- **FR-013**: System MUST reject routes with empty provider lists
- **FR-014**: System MUST reject routes with invalid weights or strategies
- **FR-015**: System MUST emit structured logs for routing normalization, decision, policy selection, and provider selection
- **FR-016**: System MUST log decision source (middleware vs builtin), scenario, reason, and confidence for each routed request
- **FR-017**: System MUST preserve existing failover behavior when scenario routes are not configured
- **FR-018**: System MUST migrate legacy routing config (top-level providers, old scenario names) to new route-policy model
- **FR-019**: System MUST populate `RequestContext` with profile, request format, normalized request, and routing fields for middleware
- **FR-020**: System MUST allow middleware to provide routing hints (scenario candidates, tags, cost class, capability needs) even without explicit decision

### Key Entities

- **NormalizedRequest**: Represents a protocol-agnostic view of an API request with extracted semantic features (model, messages, tools, reasoning, image, search, long-context indicators)
- **RoutingDecision**: Represents an explicit routing choice with scenario, source, reason, confidence, and optional overrides (model hint, strategy, threshold, provider filters)
- **RoutingHints**: Represents non-binding routing suggestions from middleware (scenario candidates, tags, cost class, capability needs)
- **RoutePolicy**: Represents the routing configuration for a specific scenario (providers, strategy, weights, threshold, fallback behavior)
- **ProfileRoutingConfig**: Represents the complete routing configuration for a profile (default route, scenario-specific routes)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Requests with identical semantic content route to the same scenario regardless of API protocol (Anthropic, OpenAI Chat, OpenAI Responses)
- **SC-002**: Middleware can successfully route requests to custom scenarios (e.g., "plan", "implement") without modifying request body structure
- **SC-003**: Users can add new scenario routes to configuration and have them work immediately without code changes
- **SC-004**: Each scenario route independently applies its configured strategy (e.g., "plan" uses weighted, "coding" uses least-cost)
- **SC-005**: Invalid routing configurations are rejected at daemon startup with clear error messages identifying the specific problem
- **SC-006**: Every routed request produces structured logs showing scenario, decision source, reason, and selected provider/model
- **SC-007**: Legacy routing configurations continue to work after upgrade with automatic migration to new format
- **SC-008**: Spec-kit middleware can route all six workflow stages (specify, clarify, plan, tasks, analyse, implement) to different providers based on configuration

## Assumptions

- The existing middleware pipeline infrastructure (`internal/middleware/interface.go`) is stable and will not require breaking changes
- The existing load balancing strategies (failover, round-robin, least-latency, least-cost, weighted) will continue to work with the new routing system
- The existing session cache and token counting logic can be reused for long-context detection in normalized requests
- Backward compatibility with existing routing configurations is required for at least one major version
- The three supported protocols (Anthropic Messages, OpenAI Chat, OpenAI Responses) cover the majority of client use cases
- Middleware authors are willing to adopt the new `RoutingDecision` API instead of relying on body manipulation
- Configuration validation errors at startup are acceptable (fail-fast approach)
- Structured JSON logging is the preferred observability mechanism for routing decisions

## Dependencies

- Existing middleware pipeline must be functional and integrated into request processing flow
- Existing load balancer must support per-request provider reordering based on strategy
- Existing config store must support schema versioning and migration
- Existing token counting logic (tiktoken or character-based fallback) must be available for long-context detection
- Existing SQLite LogDB must be available for latency metrics (used by least-latency strategy)

## Out of Scope

- Adding support for additional API protocols beyond Anthropic Messages, OpenAI Chat, and OpenAI Responses
- Implementing new load balancing strategies beyond the existing five
- Building a UI for visualizing or editing routing configurations
- Implementing automatic scenario detection based on machine learning or LLM classification
- Adding support for conditional routing based on user identity, time of day, or other external factors
- Implementing routing analytics or cost tracking dashboards
- Adding support for A/B testing or gradual rollout of routing changes
- Implementing routing rules based on response quality or user feedback
