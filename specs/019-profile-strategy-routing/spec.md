# Feature Specification: Profile Strategy-Aware Provider Routing

**Feature Branch**: `019-profile-strategy-routing`
**Created**: 2026-03-09
**Status**: Draft
**Input**: User description: "Connect profile strategy to real provider selection"

## Clarifications

### Session 2026-03-09

- Q: Over what time period should the system calculate average latency for strategy decisions? → A: Last 100 requests (rolling count)
- Q: Should the system log or expose which provider was selected and why (strategy reason)? → A: Log each strategy decision with provider and reason
- Q: How should the system handle providers with insufficient latency samples? → A: Use available samples, minimum 10 required
- Q: Should round-robin state persist across daemon restarts? → A: In-memory only, reset on restart
- Q: How should the system handle concurrent access to provider metrics during strategy evaluation? → A: Read-only snapshots per request

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Least-Latency Strategy Routing (Priority: P1)

As a user with a profile configured for least-latency strategy, when I make a request, the system should automatically route to the provider with the lowest average latency, ensuring I get the fastest possible response.

**Why this priority**: This is the most common optimization users want - minimizing response time. It directly impacts user experience and is the primary value proposition of strategy-aware routing.

**Independent Test**: Can be fully tested by configuring a profile with least-latency strategy, making requests, and verifying that the provider with lowest latency is consistently selected first (before any failover). Delivers immediate value by reducing response times.

**Acceptance Scenarios**:

1. **Given** a profile with strategy "least-latency" and three providers (A: 100ms avg, B: 50ms avg, C: 200ms avg), **When** a request is made, **Then** provider B is selected first
2. **Given** a profile with least-latency strategy and provider B becomes unhealthy, **When** a request is made, **Then** provider A (next lowest latency) is selected
3. **Given** a profile with least-latency strategy and all providers have similar latency, **When** multiple requests are made, **Then** the provider with consistently lowest latency is preferred

---

### User Story 2 - Least-Cost Strategy Routing (Priority: P2)

As a cost-conscious user with a profile configured for least-cost strategy, when I make a request, the system should automatically route to the provider with the lowest cost per token, helping me minimize API expenses.

**Why this priority**: Cost optimization is important for high-volume users but secondary to performance. Users typically want fast responses first, then cost optimization.

**Independent Test**: Can be fully tested by configuring a profile with least-cost strategy, making requests, and verifying that the cheapest provider is selected first. Delivers value by reducing operational costs.

**Acceptance Scenarios**:

1. **Given** a profile with strategy "least-cost" and three providers (A: $0.01/1K tokens, B: $0.005/1K tokens, C: $0.02/1K tokens), **When** a request is made, **Then** provider B is selected first
2. **Given** a profile with least-cost strategy and cheapest provider is unhealthy, **When** a request is made, **Then** the next cheapest healthy provider is selected
3. **Given** a profile with least-cost strategy and multiple providers have identical costs, **When** a request is made, **Then** selection falls back to configured order

---

### User Story 3 - Round-Robin Strategy Routing (Priority: P3)

As a user with a profile configured for round-robin strategy, when I make multiple requests, the system should distribute them evenly across all healthy providers, ensuring balanced load distribution.

**Why this priority**: Load balancing is useful for distributing quota usage but less critical than performance or cost optimization. Most users prefer optimized routing over even distribution.

**Independent Test**: Can be fully tested by configuring a profile with round-robin strategy, making multiple requests, and verifying that each healthy provider receives approximately equal request counts. Delivers value by preventing quota exhaustion on any single provider.

**Acceptance Scenarios**:

1. **Given** a profile with strategy "round-robin" and three healthy providers, **When** 9 requests are made, **Then** each provider receives exactly 3 requests
2. **Given** a profile with round-robin strategy and one provider becomes unhealthy, **When** 6 requests are made, **Then** the two healthy providers each receive 3 requests
3. **Given** a profile with round-robin strategy and a provider recovers from unhealthy state, **When** subsequent requests are made, **Then** the recovered provider is included in rotation

---

### User Story 4 - Weighted Strategy Routing (Priority: P3)

As a user with a profile configured for weighted strategy, when I make requests, the system should distribute them according to configured weights, allowing me to prefer certain providers while still using others.

**Why this priority**: Weighted distribution is an advanced feature for users who want fine-grained control. It's less commonly needed than the basic strategies.

**Independent Test**: Can be fully tested by configuring a profile with weighted strategy (e.g., A:70%, B:20%, C:10%), making 100 requests, and verifying distribution matches weights within acceptable variance. Delivers value by enabling custom load distribution patterns.

**Acceptance Scenarios**:

1. **Given** a profile with strategy "weighted" and weights (A:70, B:20, C:10), **When** 100 requests are made, **Then** provider A receives ~70 requests, B receives ~20, C receives ~10
2. **Given** a profile with weighted strategy and the highest-weighted provider is unhealthy, **When** requests are made, **Then** weights are recalculated among healthy providers proportionally
3. **Given** a profile with weighted strategy and no weights configured, **When** a request is made, **Then** system falls back to equal weights (uniform random distribution across healthy providers)

---

### Edge Cases

- What happens when all providers have identical strategy metrics (same latency/cost)? → System uses configured provider order as tiebreaker (FR-010)
- How does the system handle strategy selection when a provider's metrics are temporarily unavailable? → Provider is skipped if metrics unavailable, falls back to next available provider
- What happens when strategy configuration is invalid or missing? → System falls back to ordered failover (FR-011)
- How does failover work after strategy-based selection fails? → System preserves existing failover behavior (FR-009)
- What happens when a provider's health changes during strategy evaluation? → Provider is skipped if unhealthy/backoff (FR-006, FR-007)
- How does the system handle concurrent requests with different strategies? → Each request evaluates strategy independently using current metrics
- What happens when a provider has fewer than 10 latency samples? → Provider is excluded from least-latency strategy evaluation until minimum sample size reached, falls back to configured order

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST evaluate profile strategy before selecting a provider for each request
- **FR-002**: System MUST support least-latency strategy by selecting the provider with lowest average response time
- **FR-003**: System MUST support least-cost strategy by selecting the provider with lowest cost per token
- **FR-004**: System MUST support round-robin strategy by distributing requests evenly across healthy providers
- **FR-005**: System MUST support weighted strategy by distributing requests according to configured provider weights
- **FR-006**: System MUST skip unhealthy providers during strategy-based selection
- **FR-007**: System MUST skip providers in backoff state during strategy-based selection
- **FR-008**: System MUST fall back to ordered failover if strategy-based selection fails
- **FR-009**: System MUST preserve existing failover behavior after initial strategy-based selection
- **FR-010**: System MUST use configured provider order as tiebreaker when strategy metrics are equal
- **FR-011**: System MUST handle missing or invalid strategy configuration by falling back to ordered failover
- **FR-012**: System MUST track provider latency metrics for least-latency strategy evaluation (calculated as average over last 100 requests per provider, minimum 10 requests required for consideration)
- **FR-013**: System MUST track provider cost metrics for least-cost strategy evaluation
- **FR-014**: System MUST maintain round-robin state per profile to ensure even distribution (in-memory only, resets on daemon restart)
- **FR-015**: System MUST recalculate weighted distribution when provider health changes
- **FR-016**: System MUST log each strategy-based provider selection including selected provider, strategy type, and selection reason
- **FR-017**: System MUST use read-only metric snapshots for each request's strategy evaluation to ensure consistent view under concurrent access

### Key Entities

- **Profile Strategy**: Configuration that determines how providers are selected (least-latency, least-cost, round-robin, weighted, ordered)
- **Provider Metrics**: Runtime statistics including average latency (calculated over last 100 requests), cost per token, health status, and backoff state
- **Selection Context**: Per-request state including profile strategy, available providers, and current metrics snapshot (read-only view for consistent evaluation)
- **Round-Robin State**: Per-profile counter tracking the next provider to use in rotation (in-memory, resets on daemon restart)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Requests with least-latency strategy are routed to the provider with lowest average latency 95% of the time (excluding failover scenarios)
- **SC-002**: Requests with least-cost strategy are routed to the provider with lowest cost 95% of the time (excluding failover scenarios)
- **SC-003**: Requests with round-robin strategy achieve even distribution within 10% variance across healthy providers over 100 requests
- **SC-004**: Requests with weighted strategy achieve distribution within 15% of configured weights over 100 requests
- **SC-005**: Strategy-based selection completes in under 5ms to avoid adding latency to request path
- **SC-006**: System maintains existing failover success rate (no regression in reliability)
- **SC-007**: Users with least-latency strategy experience average response time reduction of at least 20% compared to ordered failover
- **SC-008**: Users with least-cost strategy experience cost reduction of at least 15% compared to ordered failover

## Assumptions

- Provider latency metrics are already being tracked by the existing monitoring system
- Provider cost information is available in configuration or can be calculated from usage data
- The existing load balancer implementation provides the foundation for strategy evaluation
- Profile strategy configuration already exists in the config schema
- Unhealthy and backoff provider filtering is already implemented and working correctly
- The main request path has a clear injection point for strategy-based provider selection
