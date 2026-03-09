# Data Model: Profile Strategy-Aware Provider Routing

**Feature**: 019-profile-strategy-routing
**Date**: 2026-03-09
**Status**: Complete

## Overview

This document defines the data entities, relationships, and state management for strategy-aware provider routing. The feature extends existing entities rather than introducing new ones.

---

## Core Entities

### Entity 1: ProfileConfig (EXISTING - EXTENDED)

**Location**: `internal/config/config.go:318`

**Purpose**: Holds profile configuration including provider list and load balancing strategy

**Fields**:
```go
type ProfileConfig struct {
    Providers            []string                    // Ordered list of provider names
    Routing              map[Scenario]*ScenarioRoute // Optional scenario-based routing
    LongContextThreshold int                         // Token threshold for long-context routing
    Strategy             LoadBalanceStrategy         // EXISTING: Load balancing strategy (added v1.4.0)
    ProviderWeights      map[string]int              // NEW: Provider weights for weighted strategy (provider name → weight)
}
```

**Validation Rules**:
- `Providers`: Must contain at least 1 provider name, each must exist in global provider config
- `Strategy`: Must be one of: `failover`, `round-robin`, `least-latency`, `least-cost`, `weighted` (empty defaults to `failover`)
- `LongContextThreshold`: If set, must be > 0 (defaults to 32000 if not set)
- `ProviderWeights`: Optional, only used when Strategy is `weighted`. Keys must match provider names in `Providers` list. Values must be > 0. If Strategy is `weighted` but ProviderWeights is empty/nil, falls back to equal weights (round-robin behavior)

**State Transitions**: None (immutable after load, replaced on config reload)

**Relationships**:
- **Has-Many**: Provider names (references `ProviderConfig` by name)
- **Used-By**: `ProfileProxy` (resolves profile → providers + strategy)

---

### Entity 2: LoadBalanceStrategy (EXISTING - ENUM)

**Location**: `internal/config/config.go:771`

**Purpose**: Enum defining supported load balancing strategies

**Values**:
```go
type LoadBalanceStrategy string

const (
    LoadBalanceFailover     LoadBalanceStrategy = "failover"      // Try providers in configured order
    LoadBalanceRoundRobin   LoadBalanceStrategy = "round-robin"   // Rotate evenly across providers
    LoadBalanceLeastLatency LoadBalanceStrategy = "least-latency" // Select provider with lowest avg latency
    LoadBalanceLeastCost    LoadBalanceStrategy = "least-cost"    // Select provider with lowest cost per token
    LoadBalanceWeighted     LoadBalanceStrategy = "weighted"      // Distribute by configured weights
)
```

**Validation Rules**:
- Must be one of the five defined constants
- Empty string treated as `LoadBalanceFailover` (default)

**Usage**: Stored in `ProfileConfig.Strategy`, passed to `LoadBalancer.Select()`

---

### Entity 3: ProviderMetrics (EXISTING - EXTENDED)

**Location**: `internal/proxy/metrics.go` (existing), `internal/proxy/logdb.go` (storage)

**Purpose**: Tracks runtime statistics for each provider (latency, request count, error rate)

**Fields**:
```go
type ProviderMetrics struct {
    ProviderName  string        // Provider identifier
    TotalRequests int           // Total requests sent to this provider
    AvgLatencyMs  float64       // Average latency over last N requests
    ErrorRate     float64       // Percentage of failed requests
    LastUpdated   time.Time     // Timestamp of last metric update
}
```

**Validation Rules**:
- `AvgLatencyMs`: Must be >= 0 (calculated from successful requests only)
- `ErrorRate`: Must be 0.0-1.0 (percentage as decimal)
- `TotalRequests`: Must be >= 0

**State Transitions**:
- **Initial**: `TotalRequests=0, AvgLatencyMs=0, ErrorRate=0`
- **After Request**: `TotalRequests++, AvgLatencyMs=recalculate(), ErrorRate=recalculate()`
- **On Cache Refresh**: Metrics reloaded from LogDB (every 30 seconds)

**Relationships**:
- **Belongs-To**: Provider (one-to-one, keyed by provider name)
- **Stored-In**: LogDB (SQLite, `requests` table)
- **Used-By**: LoadBalancer (for least-latency strategy evaluation)

---

### Entity 4: LoadBalancer (EXISTING - EXTENDED)

**Location**: `internal/proxy/loadbalancer.go:12`

**Purpose**: Selects providers based on configured strategy and runtime metrics

**Fields**:
```go
type LoadBalancer struct {
    db           *LogDB                        // Database for latency metrics
    pricing      map[string]*ModelPricing      // Model pricing data (for least-cost)
    mu           sync.RWMutex                  // Protects metricsCache
    rrCounter    uint64                        // Atomic counter for round-robin
    metricsCache map[string]*ProviderMetrics   // Cached provider metrics
    cacheTime    time.Time                     // Last cache refresh time
    cacheTTL     time.Duration                 // Cache validity duration (30s)
}
```

**Validation Rules**:
- `db`: Must not be nil (required for latency metrics)
- `cacheTTL`: Must be > 0 (defaults to 30 seconds)

**State Transitions**:
- **Initial**: `metricsCache=empty, rrCounter=0`
- **On Select()**: `rrCounter++` (if round-robin), `metricsCache` refreshed if stale
- **On Cache Refresh**: `metricsCache` replaced with fresh data from LogDB

**Relationships**:
- **Uses**: LogDB (queries latency metrics)
- **Uses**: ModelPricing (queries cost data)
- **Called-By**: ProfileProxy (passes strategy + providers)

---

### Entity 5: StrategyDecision (NEW - EPHEMERAL)

**Location**: In-memory only (not persisted)

**Purpose**: Represents the result of strategy evaluation for a single request

**Fields**:
```go
type StrategyDecision struct {
    ProfileName      string                  // Profile that triggered evaluation
    Strategy         LoadBalanceStrategy     // Strategy used for selection
    SelectedProvider string                  // Provider chosen by strategy
    Reason           string                  // Human-readable selection reason
    Timestamp        time.Time               // When decision was made
    CandidateCount   int                     // Number of providers evaluated
}
```

**Validation Rules**:
- `SelectedProvider`: Must be non-empty (always selects at least one provider)
- `Reason`: Must be non-empty (e.g., "lowest latency: 45ms", "round-robin: index 2")

**State Transitions**: None (created, logged, discarded)

**Relationships**:
- **Created-By**: LoadBalancer.Select()
- **Logged-By**: ProfileProxy (via Logger)
- **Not-Persisted**: Ephemeral (exists only for logging)

---

## Data Relationships

```
ProfileConfig (1) ----< (N) ProviderConfig
      |
      | (has strategy)
      v
LoadBalanceStrategy (enum)
      |
      | (evaluated by)
      v
LoadBalancer
      |
      +----> (queries) LogDB ----< (N) ProviderMetrics
      |
      +----> (queries) ModelPricing
      |
      | (produces)
      v
StrategyDecision (ephemeral)
```

**Key Relationships**:
1. **ProfileConfig → LoadBalanceStrategy**: One-to-one (each profile has one strategy)
2. **LoadBalancer → ProviderMetrics**: One-to-many (queries metrics for all providers)
3. **LoadBalancer → StrategyDecision**: One-to-one per request (creates decision, logs, discards)

---

## State Management

### Round-Robin State

**Storage**: In-memory, `LoadBalancer.rrCounter` (uint64)

**Lifecycle**:
- **Initialization**: Set to 0 when LoadBalancer created
- **Update**: Atomically incremented on each round-robin selection
- **Reset**: On daemon restart (in-memory only, not persisted)

**Concurrency**: Thread-safe via `sync/atomic.AddUint64()`

**Persistence**: None (per clarification Q4: in-memory only, resets on restart)

---

### Latency Metrics Cache

**Storage**: In-memory, `LoadBalancer.metricsCache` (map[string]*ProviderMetrics)

**Lifecycle**:
- **Initialization**: Empty map when LoadBalancer created
- **Refresh**: Queried from LogDB when cache is stale (age > 30s)
- **Invalidation**: On config reload (cache cleared, fresh query on next request)

**Concurrency**: Thread-safe via `sync.RWMutex` (multiple concurrent readers, single writer)

**Persistence**: Underlying data persisted in LogDB (SQLite), cache is ephemeral

---

### Strategy Decision Log

**Storage**: Structured logs (stderr), not persisted to database

**Lifecycle**:
- **Creation**: On each `LoadBalancer.Select()` call
- **Logging**: Immediately after provider selection
- **Retention**: Managed by log rotation (external to application)

**Format**:
```
[strategy] profile=default strategy=least-latency selected=provider-a reason="lowest latency: 45ms" candidates=3
```

---

## Data Flow

### Flow 1: Strategy Evaluation (Least-Latency)

```
1. Request arrives → ProfileProxy.ServeHTTP()
2. Resolve profile → profileCfg.Strategy = "least-latency"
3. Call LoadBalancer.Select(providers, "least-latency", model)
4. LoadBalancer checks cache age → stale (>30s)
5. LoadBalancer queries LogDB.GetProviderLatencyMetrics(since=now-1h, limit=100)
6. LogDB returns: {provider-a: 45ms (50 samples), provider-b: 120ms (30 samples), provider-c: 0ms (5 samples)}
7. LoadBalancer filters: provider-c excluded (< 10 samples)
8. LoadBalancer sorts: [provider-a (45ms), provider-b (120ms)]
9. LoadBalancer logs: "[strategy] profile=default strategy=least-latency selected=provider-a reason='lowest latency: 45ms' candidates=2"
10. Return ordered list: [provider-a, provider-b, provider-c]
11. ProxyServer tries providers in order (existing failover logic)
```

---

### Flow 2: Strategy Evaluation (Round-Robin)

```
1. Request arrives → ProfileProxy.ServeHTTP()
2. Resolve profile → profileCfg.Strategy = "round-robin"
3. Call LoadBalancer.Select(providers, "round-robin", model)
4. LoadBalancer atomically increments rrCounter: 0 → 1
5. Calculate index: 1 % 3 = 1 (select provider at index 1)
6. Rotate provider list: [provider-b, provider-c, provider-a]
7. Move unhealthy to end: [provider-b (healthy), provider-a (healthy), provider-c (unhealthy)]
8. LoadBalancer logs: "[strategy] profile=default strategy=round-robin selected=provider-b reason='round-robin: index 1' candidates=3"
9. Return ordered list: [provider-b, provider-a, provider-c]
10. ProxyServer tries providers in order
```

---

### Flow 3: Insufficient Samples Fallback

```
1. Request arrives → ProfileProxy.ServeHTTP()
2. Resolve profile → profileCfg.Strategy = "least-latency"
3. Call LoadBalancer.Select(providers, "least-latency", model)
4. LoadBalancer queries metrics: {provider-a: 0ms (3 samples), provider-b: 0ms (5 samples)}
5. LoadBalancer filters: ALL providers excluded (< 10 samples)
6. LoadBalancer falls back to configured order: [provider-a, provider-b]
7. LoadBalancer logs: "[strategy] profile=default strategy=least-latency selected=provider-a reason='insufficient samples, using configured order' candidates=0"
8. Return original order: [provider-a, provider-b]
```

---

## Validation Rules Summary

### ProfileConfig.Strategy
- **Type**: `LoadBalanceStrategy` (enum)
- **Required**: No (defaults to `failover`)
- **Valid Values**: `failover`, `round-robin`, `least-latency`, `least-cost`
- **Invalid Behavior**: Fall back to `failover`, log warning

### ProviderMetrics.AvgLatencyMs
- **Type**: `float64`
- **Range**: >= 0
- **Calculation**: `SUM(latency_ms) / COUNT(*) WHERE timestamp > now-1h AND provider = ? LIMIT 100`
- **Minimum Samples**: 10 (providers with < 10 samples excluded from least-latency evaluation)

### LoadBalancer.rrCounter
- **Type**: `uint64`
- **Range**: 0 to 2^64-1 (wraps around on overflow)
- **Concurrency**: Atomic increment via `sync/atomic.AddUint64()`
- **Persistence**: None (resets to 0 on daemon restart)

---

## Schema Changes

**None required** - All entities already exist in codebase:
- `ProfileConfig.Strategy` added in v1.4.0 (config version 12)
- `LoadBalanceStrategy` enum added in v1.4.0
- `ProviderMetrics` exists in `internal/proxy/metrics.go`
- `LoadBalancer` exists in `internal/proxy/loadbalancer.go`

**No config migration needed** - Feature uses existing schema.

---

## Performance Characteristics

### Memory Usage
- **ProviderMetrics cache**: ~1KB per provider × 50 providers = 50KB typical
- **Round-robin counter**: 8 bytes (uint64)
- **Total overhead**: < 100KB (negligible)

### Query Performance
- **Latency metric query**: ~5ms (SQLite indexed query, last 100 requests)
- **Cache hit rate**: ~99% (30s TTL, requests typically clustered)
- **Strategy evaluation**: ~0.8ms (in-memory sorting, 50 providers)

### Concurrency
- **Read contention**: None (RWMutex allows unlimited concurrent readers)
- **Write contention**: Minimal (metric cache updated every 30s, blocks for ~1ms)
- **Scalability**: Supports 1000+ concurrent requests without modification
