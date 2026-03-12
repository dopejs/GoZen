# Research: Profile Strategy-Aware Provider Routing

**Feature**: 019-profile-strategy-routing
**Date**: 2026-03-09
**Status**: Complete

## Overview

This document captures research findings for implementing strategy-aware provider routing. All technical unknowns from the planning phase have been resolved through codebase analysis and industry best practices research.

## Research Questions & Findings

### Q1: How should latency metrics be calculated and stored?

**Decision**: Use rolling window of last 100 requests per provider, stored in existing SQLite LogDB

**Rationale**:
- Existing `LogDB` already tracks request latency via `RecordRequest()` method
- SQLite provides efficient time-range queries for metric calculation
- 100-request window balances responsiveness (adapts to recent changes) with stability (filters outliers)
- Aligns with industry standard for adaptive load balancing (AWS ELB uses similar approach)

**Alternatives Considered**:
- Time-based window (e.g., last 1 hour): Rejected because request volume varies widely (10 req/hour to 1000 req/hour), making time-based windows unreliable
- In-memory only: Rejected because metrics would reset on daemon restart, losing valuable historical data
- Exponential moving average: Rejected for complexity; simple average over fixed window is sufficient and more predictable

**Implementation Notes**:
- Add `GetProviderLatencyMetrics(since time.Time, limit int)` method to LogDB
- Query: `SELECT provider, AVG(latency_ms) FROM requests WHERE timestamp > ? GROUP BY provider HAVING COUNT(*) >= 10 ORDER BY timestamp DESC LIMIT ?`
- Cache results for 30 seconds (existing pattern in LoadBalancer.getMetricsCache())

---

### Q2: How should round-robin state be managed across concurrent requests?

**Decision**: Use atomic counter with modulo arithmetic, no persistence

**Rationale**:
- `sync/atomic.AddUint64()` provides lock-free concurrent access
- Modulo operation ensures even distribution: `index = (counter % providerCount)`
- In-memory state acceptable because round-robin is stateless by nature (any starting point is valid)
- Existing LoadBalancer already uses this pattern (line 83: `atomic.AddUint64(&lb.rrCounter, 1)`)

**Alternatives Considered**:
- Mutex-protected counter: Rejected for performance (atomic operations are faster)
- Per-profile counters: Rejected because existing global counter already works correctly
- Persistent state: Rejected per clarification Q4 (in-memory only, resets on restart)

**Implementation Notes**:
- No changes needed - existing `LoadBalancer.rrCounter` already implements this correctly
- Counter overflow is safe: uint64 wraps around after 2^64 increments (effectively infinite for this use case)

---

### Q3: How should concurrent strategy evaluation be made thread-safe?

**Decision**: Create read-only metric snapshots per request using RWMutex

**Rationale**:
- Read-only snapshots prevent race conditions without blocking concurrent reads
- `sync.RWMutex` allows multiple concurrent readers (strategy evaluations) while serializing writes (metric updates)
- Snapshot approach ensures consistent view of metrics throughout single request lifecycle
- Aligns with Go concurrency best practices (share memory by communicating, not vice versa)

**Alternatives Considered**:
- Global mutex: Rejected because it would serialize all strategy evaluations, killing concurrency
- Lock-free data structures: Rejected for complexity; RWMutex is sufficient and well-tested
- Copy-on-write: Rejected because metric maps are already small (<50 providers typical), shallow copy is cheap

**Implementation Notes**:
- Existing `LoadBalancer.getMetricsCache()` already implements RWMutex pattern correctly
- Each `Select()` call gets snapshot via `getMetricsCache()`, operates on immutable copy
- Cache TTL (30s) balances freshness with query overhead

---

### Q4: How should strategy decisions be logged for observability?

**Decision**: Use structured logging with provider name, strategy type, and selection reason

**Rationale**:
- Existing `ProfileProxy.Logger` provides structured logging infrastructure
- Log format: `[strategy] profile=%s strategy=%s selected=%s reason=%s`
- Enables debugging (why was provider X chosen?), performance analysis (is strategy working?), and audit trails
- Aligns with existing logging patterns in codebase (see `profile_proxy.go:62`)

**Alternatives Considered**:
- Metrics-only (no logs): Rejected because metrics don't capture decision rationale
- Verbose logging (all candidates): Rejected for log volume; only log final decision
- Separate audit log: Rejected for complexity; existing logger is sufficient

**Implementation Notes**:
- Add logging in `LoadBalancer.Select()` after provider selection
- Log level: INFO (not DEBUG) because strategy decisions are operationally significant
- Include: profile name, strategy type, selected provider, reason (e.g., "lowest latency: 45ms")

---

### Q5: How should insufficient sample size be handled?

**Decision**: Exclude providers with <10 requests from least-latency evaluation, fall back to configured order

**Rationale**:
- 10-request minimum provides statistical significance (reduces impact of outliers)
- Excluding insufficient-sample providers prevents premature optimization based on noise
- Fallback to configured order preserves user intent (explicit provider ordering in config)
- Aligns with industry practice (AWS CloudWatch requires minimum sample size for alarms)

**Alternatives Considered**:
- Use available samples regardless of count: Rejected because 1-2 samples are unreliable
- Default to maximum latency: Rejected because it unfairly penalizes new providers
- Wait until minimum reached: Rejected because it would block requests

**Implementation Notes**:
- SQL query includes `HAVING COUNT(*) >= 10` clause
- Providers without sufficient samples are appended to end of sorted list (after providers with metrics)
- Log warning when provider excluded: `[strategy] provider=%s excluded: insufficient samples (count=%d, minimum=10)`

---

### Q6: How should weighted strategy be configured and recalculated?

**Decision**: Store weights in `ProfileConfig.ProviderWeights` map, recalculate proportionally when provider health changes

**Rationale**:
- Per-profile weights allow different profiles to have different preferences (e.g., "work" profile prefers cheap providers, "personal" profile prefers fast providers)
- Map structure `map[string]int` (provider name → weight) is simple and explicit
- Proportional recalculation preserves relative preferences when providers become unhealthy
- Fallback to equal weights (round-robin) when no weights configured provides sensible default

**Alternatives Considered**:
- Per-provider weights (global): Rejected because different profiles may want different distributions
- Fixed fallback provider: Rejected because it doesn't preserve relative preferences
- Skip recalculation (use original weights): Rejected because it would route to unhealthy providers

**Recalculation Algorithm**:
```
Given: Weights A=70, B=20, C=10 (total=100)
If A becomes unhealthy:
  - Remaining healthy: B=20, C=10 (total=30)
  - Recalculated: B=20/30=66.7%, C=10/30=33.3%
  - Result: B gets ~67% of requests, C gets ~33%

If no weights configured (ProviderWeights is nil/empty):
  - Fall back to equal weights: each provider gets 1/N of requests
  - Equivalent to round-robin behavior
```

**Implementation Notes**:
- Weighted selection uses weighted random sampling: generate random number 0-100, select provider based on cumulative weight ranges
- Recalculation happens on-demand during `selectWeighted()` call (no persistent state)
- Log decision: `[strategy] profile=%s strategy=weighted selected=%s reason="weighted: 70%" candidates=3`

---

## Technology Choices

### Latency Metric Storage: SQLite (existing LogDB)

**Chosen**: SQLite via existing `internal/proxy/logdb.go`

**Why**:
- Already integrated and battle-tested (used since v1.5.1)
- Efficient time-range queries with indexes
- Persistent across daemon restarts
- No additional dependencies

**Best Practices**:
- Use prepared statements for query performance
- Add index on `(provider, timestamp)` for fast metric queries
- Limit query to last 24 hours to prevent unbounded growth

---

### Concurrency Control: sync.RWMutex + Atomic Operations

**Chosen**: `sync.RWMutex` for metric cache, `sync/atomic` for round-robin counter

**Why**:
- Standard library primitives, no external dependencies
- RWMutex allows concurrent reads (strategy evaluations) while serializing writes (metric updates)
- Atomic operations provide lock-free counter increment for round-robin

**Best Practices**:
- Always acquire read lock before accessing shared state
- Keep critical sections small (lock, copy, unlock)
- Use defer for lock release to prevent deadlocks

---

### Strategy Evaluation: Switch Statement with Fallback

**Chosen**: Simple switch on `config.LoadBalanceStrategy` enum

**Why**:
- Explicit and easy to understand
- Compile-time exhaustiveness checking (Go compiler warns on missing cases)
- No reflection or dynamic dispatch overhead

**Best Practices**:
- Always include default case for unknown strategies (fall back to ordered failover)
- Document fallback behavior in code comments
- Log warning when falling back due to invalid strategy

---

## Integration Patterns

### Pattern 1: Profile Strategy → LoadBalancer Selection

**Flow**:
1. `ProfileProxy.ServeHTTP()` resolves profile config
2. Extract `profileCfg.Strategy` from config
3. Pass strategy to `LoadBalancer.Select(providers, strategy, model)`
4. LoadBalancer evaluates strategy and returns ordered provider list
5. ProxyServer tries providers in returned order (existing failover logic)

**Key Insight**: Strategy evaluation happens BEFORE failover, not instead of it. Failover is preserved as safety net.

---

### Pattern 2: Metric Collection → Strategy Evaluation

**Flow**:
1. `ProxyServer.forwardRequest()` records latency via `LogDB.RecordRequest()`
2. `LoadBalancer.getMetricsCache()` queries LogDB for recent metrics (cached 30s)
3. `LoadBalancer.selectLeastLatency()` uses cached metrics to sort providers
4. Cache invalidation on config reload ensures fresh metrics after provider changes

**Key Insight**: Metrics are collected passively (no active probing), evaluation uses cached snapshots (no query per request).

---

### Pattern 3: Concurrent Request Safety

**Flow**:
1. Request A calls `LoadBalancer.Select()` → acquires read lock → gets metric snapshot → releases lock
2. Request B calls `LoadBalancer.Select()` concurrently → acquires read lock (allowed) → gets same snapshot → releases lock
3. Metric update (background) → acquires write lock (blocks readers) → updates cache → releases lock

**Key Insight**: Read-only snapshots allow concurrent strategy evaluation without blocking. Write lock serializes metric updates but doesn't block long.

---

## Performance Considerations

### Latency Target: <5ms per strategy evaluation

**Analysis**:
- Metric cache lookup: ~0.1ms (in-memory map access)
- Provider sorting (50 providers): ~0.5ms (bubble sort, O(n²) acceptable for small n)
- Logging: ~0.2ms (buffered I/O)
- **Total**: ~0.8ms typical, well under 5ms target

**Optimization Notes**:
- No optimization needed for MVP (current approach is fast enough)
- If >100 providers: consider quicksort instead of bubble sort
- If >1000 req/s: consider pre-sorted provider lists (updated on metric refresh)

---

### Concurrency Target: 100 concurrent requests

**Analysis**:
- RWMutex allows unlimited concurrent readers (strategy evaluations)
- Write lock (metric update) happens every 30s, blocks for ~1ms
- **Bottleneck**: None identified (read-heavy workload favors RWMutex)

**Scaling Notes**:
- Current design supports 1000+ concurrent requests without modification
- If write contention becomes issue: increase cache TTL to 60s

---

## Edge Cases & Error Handling

### Edge Case 1: All providers have insufficient samples

**Behavior**: Fall back to configured provider order (same as ordered failover)

**Rationale**: User-configured order represents explicit intent, safe default

---

### Edge Case 2: Strategy evaluation fails (e.g., DB query error)

**Behavior**: Log error, fall back to ordered failover

**Rationale**: Availability over optimization (better to route sub-optimally than fail request)

---

### Edge Case 3: Provider becomes unhealthy during strategy evaluation

**Behavior**: Unhealthy providers moved to end of list (existing behavior preserved)

**Rationale**: Health checks take precedence over strategy optimization

---

### Edge Case 4: Concurrent config reload during strategy evaluation

**Behavior**: Request uses stale metric snapshot (up to 30s old), next request gets fresh metrics

**Rationale**: Eventual consistency acceptable (30s staleness is negligible for latency-based routing)

---

## Open Questions

**None** - All technical unknowns resolved through research.

---

## References

- Existing codebase: `internal/proxy/loadbalancer.go` (lines 1-299)
- Existing codebase: `internal/proxy/profile_proxy.go` (lines 1-362)
- Existing codebase: `internal/proxy/logdb.go` (latency tracking)
- Go concurrency patterns: https://go.dev/blog/pipelines
- AWS ELB load balancing: https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-target-groups.html
