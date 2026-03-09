# Strategy API Contract

**Feature**: 019-profile-strategy-routing
**Date**: 2026-03-09
**Version**: 1.0

## Overview

This document defines the interface contract for strategy-aware provider selection. The contract specifies how components interact to evaluate load balancing strategies and select providers.

---

## Interface: LoadBalancer.Select()

### Purpose
Selects and orders providers based on configured strategy and runtime metrics.

### Signature
```go
func (lb *LoadBalancer) Select(
    providers []*Provider,
    strategy config.LoadBalanceStrategy,
    model string,
) []*Provider
```

### Parameters

#### `providers` []*Provider
- **Type**: Slice of Provider pointers
- **Required**: Yes
- **Constraints**:
  - Must not be nil
  - May be empty (returns empty slice)
  - Order represents user-configured preference (used as tiebreaker)
- **Example**:
  ```go
  []*Provider{
      {Name: "provider-a", BaseURL: "https://api.anthropic.com", Healthy: true},
      {Name: "provider-b", BaseURL: "https://api.openai.com", Healthy: true},
  }
  ```

#### `strategy` config.LoadBalanceStrategy
- **Type**: LoadBalanceStrategy enum
- **Required**: Yes (empty string treated as `failover`)
- **Valid Values**:
  - `"failover"`: Return providers in original order (healthy first)
  - `"round-robin"`: Rotate evenly across providers
  - `"least-latency"`: Sort by average latency (lowest first)
  - `"least-cost"`: Sort by cost per token (lowest first)
- **Invalid Values**: Fall back to `failover`, log warning
- **Example**: `config.LoadBalanceLeastLatency`

#### `model` string
- **Type**: String
- **Required**: Yes (may be empty)
- **Purpose**: Used for cost calculation in `least-cost` strategy
- **Constraints**: Must match model name in pricing data (partial match supported)
- **Example**: `"claude-sonnet-4-5"`, `"gpt-4"`

### Return Value

#### []*Provider
- **Type**: Slice of Provider pointers (reordered copy)
- **Guarantees**:
  - Same providers as input (no additions/removals)
  - Healthy providers always before unhealthy providers
  - Order determined by strategy evaluation
  - Original slice not modified (returns new slice)
- **Empty Input**: Returns empty slice
- **Single Provider**: Returns single-element slice (no reordering)

### Behavior Specification

#### Strategy: `failover`
```
Input:  [A (healthy), B (unhealthy), C (healthy)]
Output: [A, C, B]  // Healthy first, preserve original order
```

#### Strategy: `round-robin`
```
Call 1: [A, B, C] → [B, C, A]  // Start at index 1
Call 2: [A, B, C] → [C, A, B]  // Start at index 2
Call 3: [A, B, C] → [A, B, C]  // Start at index 0 (wraps)
```

#### Strategy: `least-latency`
```
Metrics: {A: 45ms (50 samples), B: 120ms (30 samples), C: 0ms (5 samples)}
Output:  [A, B, C]  // A lowest latency, C excluded (< 10 samples), appended to end
```

#### Strategy: `least-cost`
```
Pricing: {A: $0.01/1K tokens, B: $0.005/1K tokens, C: $0.02/1K tokens}
Output:  [B, A, C]  // B cheapest, A second, C most expensive
```

#### Strategy: `weighted`
```
Weights: {A: 70, B: 20, C: 10}  // From ProfileConfig.ProviderWeights
Random:  42 (generated 0-100)
Ranges:  A=[0-70), B=[70-90), C=[90-100)
Output:  [A, B, C]  // 42 falls in A's range, A selected first

If A becomes unhealthy:
  Weights: {B: 20, C: 10}  // A excluded
  Recalculated: B=20/(20+10)=66.7%, C=10/(20+10)=33.3%
  Ranges:  B=[0-67), C=[67-100)
  Output:  [B, C, A]  // B selected first (proportional redistribution)

If no weights configured (ProviderWeights is nil/empty):
  Fallback: Equal weights (round-robin behavior)
  Output:  [A, B, C] rotated by round-robin counter
```

### Error Handling

#### Metric Query Failure
- **Condition**: LogDB query fails during `least-latency` evaluation
- **Behavior**: Fall back to `failover` strategy, log error
- **Example**:
  ```go
  if metrics, err := lb.db.GetProviderLatencyMetrics(...); err != nil {
      lb.logger.Printf("[strategy] error: failed to query metrics: %v, falling back to failover", err)
      return lb.selectFailover(providers)
  }
  ```

#### Insufficient Samples
- **Condition**: All providers have < 10 latency samples
- **Behavior**: Return providers in original order, log warning
- **Example**:
  ```go
  if len(validProviders) == 0 {
      lb.logger.Printf("[strategy] warning: no providers with sufficient samples (minimum 10), using configured order")
      return providers
  }
  ```

#### Invalid Strategy
- **Condition**: Strategy value not recognized
- **Behavior**: Fall back to `failover`, log warning
- **Example**:
  ```go
  default:
      lb.logger.Printf("[strategy] warning: unknown strategy %q, falling back to failover", strategy)
      return lb.selectFailover(providers)
  ```

#### No Weights Configured (Weighted Strategy)
- **Condition**: Strategy is `weighted` but `ProfileConfig.ProviderWeights` is nil or empty
- **Behavior**: Fall back to equal weights (round-robin behavior), log info
- **Example**:
  ```go
  if strategy == config.LoadBalanceWeighted && len(profileCfg.ProviderWeights) == 0 {
      lb.logger.Printf("[strategy] info: weighted strategy with no weights configured, using equal weights")
      return lb.selectRoundRobin(providers)
  }
  ```

#### Invalid Weights (Weighted Strategy)
- **Condition**: Weight value is <= 0 or provider name doesn't match any provider in list
- **Behavior**: Skip invalid weights, use valid weights only, log warning
- **Example**:
  ```go
  if weight <= 0 {
      lb.logger.Printf("[strategy] warning: invalid weight %d for provider %s, skipping", weight, providerName)
      continue
  }
  ```
  default:
      lb.logger.Printf("[strategy] warning: unknown strategy %q, falling back to failover", strategy)
      return lb.selectFailover(providers)
  ```

### Concurrency Guarantees

#### Thread Safety
- **Read Operations**: Multiple concurrent calls to `Select()` are safe
- **Write Operations**: Metric cache updates are serialized via `sync.RWMutex`
- **Atomicity**: Round-robin counter increments are atomic via `sync/atomic`

#### Consistency
- **Metric Snapshot**: Each `Select()` call uses consistent metric snapshot (no partial updates)
- **Cache Staleness**: Metrics may be up to 30 seconds stale (acceptable for load balancing)
- **Provider List**: Input slice not modified (returns new slice)

### Performance Guarantees

#### Latency
- **Target**: < 5ms per call (99th percentile)
- **Typical**: ~0.8ms (in-memory operations only)
- **Worst Case**: ~10ms (cache miss + DB query)

#### Throughput
- **Concurrent Calls**: Supports 1000+ concurrent calls (read-heavy workload)
- **Bottleneck**: None identified (RWMutex allows unlimited concurrent readers)

---

## Interface: LogDB.GetProviderLatencyMetrics()

### Purpose
Queries average latency for each provider over a time window.

### Signature
```go
func (db *LogDB) GetProviderLatencyMetrics(
    since time.Time,
    limit int,
) (map[string]*ProviderMetrics, error)
```

### Parameters

#### `since` time.Time
- **Type**: Time
- **Required**: Yes
- **Purpose**: Start of time window for metric calculation
- **Typical Value**: `time.Now().Add(-1 * time.Hour)` (last 1 hour)
- **Constraints**: Must be in the past (future times return empty result)

#### `limit` int
- **Type**: Integer
- **Required**: Yes
- **Purpose**: Maximum number of requests per provider to include in average
- **Typical Value**: `100` (last 100 requests per provider)
- **Constraints**: Must be > 0 (0 or negative returns empty result)

### Return Value

#### map[string]*ProviderMetrics
- **Type**: Map of provider name to metrics
- **Key**: Provider name (string)
- **Value**: ProviderMetrics struct
  ```go
  type ProviderMetrics struct {
      ProviderName  string
      TotalRequests int     // Number of requests in time window
      AvgLatencyMs  float64 // Average latency in milliseconds
      ErrorRate     float64 // Percentage of failed requests (0.0-1.0)
      LastUpdated   time.Time
  }
  ```
- **Empty Result**: Empty map (not nil) if no requests in time window
- **Minimum Samples**: Providers with < 10 requests excluded from result

#### error
- **Type**: Error
- **Nil**: Success
- **Non-Nil**: Database query failed (e.g., connection error, SQL syntax error)

### SQL Query
```sql
SELECT
    provider,
    COUNT(*) as total_requests,
    AVG(latency_ms) as avg_latency_ms,
    SUM(CASE WHEN error != '' THEN 1 ELSE 0 END) * 1.0 / COUNT(*) as error_rate
FROM requests
WHERE timestamp > ?
GROUP BY provider
HAVING COUNT(*) >= 10
ORDER BY timestamp DESC
LIMIT ?
```

### Error Handling

#### Database Connection Error
- **Condition**: SQLite database file not accessible
- **Return**: `nil, fmt.Errorf("failed to open database: %w", err)`

#### Query Execution Error
- **Condition**: SQL syntax error or constraint violation
- **Return**: `nil, fmt.Errorf("failed to query metrics: %w", err)`

#### No Results
- **Condition**: No requests in time window or all providers have < 10 samples
- **Return**: `map[string]*ProviderMetrics{}, nil` (empty map, no error)

---

## Logging Contract

### Log Format
```
[strategy] profile=<profile> strategy=<strategy> selected=<provider> reason=<reason> candidates=<count>
```

### Log Fields

#### `profile` string
- **Purpose**: Identifies which profile triggered strategy evaluation
- **Example**: `default`, `work`, `_tmp_abc123`

#### `strategy` string
- **Purpose**: Strategy used for selection
- **Values**: `failover`, `round-robin`, `least-latency`, `least-cost`

#### `selected` string
- **Purpose**: Provider chosen by strategy
- **Example**: `provider-a`, `anthropic-official`

#### `reason` string
- **Purpose**: Human-readable explanation of selection
- **Examples**:
  - `"lowest latency: 45ms"`
  - `"round-robin: index 2"`
  - `"lowest cost: $0.005/1K tokens"`
  - `"insufficient samples, using configured order"`

#### `candidates` int
- **Purpose**: Number of providers evaluated (excludes insufficient samples)
- **Example**: `3` (3 providers had sufficient metrics)

### Log Examples

#### Least-Latency Selection
```
[strategy] profile=default strategy=least-latency selected=provider-a reason="lowest latency: 45ms" candidates=2
```

#### Round-Robin Selection
```
[strategy] profile=work strategy=round-robin selected=provider-b reason="round-robin: index 1" candidates=3
```

#### Fallback to Configured Order
```
[strategy] profile=default strategy=least-latency selected=provider-a reason="insufficient samples, using configured order" candidates=0
```

#### Error Fallback
```
[strategy] error: failed to query metrics: database locked, falling back to failover
[strategy] profile=default strategy=failover selected=provider-a reason="failover: first healthy provider" candidates=3
```

---

## Backward Compatibility

### Config Schema
- **No Changes**: `ProfileConfig.Strategy` already exists (added v1.4.0)
- **Default Value**: Empty string treated as `failover` (preserves existing behavior)
- **Migration**: None required (existing configs work without modification)

### API Compatibility
- **LoadBalancer.Select()**: Existing signature unchanged, strategy parameter already present
- **Existing Callers**: Continue to work (pass empty string for strategy → defaults to failover)

### Behavior Compatibility
- **Failover**: Identical to existing behavior (healthy providers first, original order preserved)
- **Health Checks**: Unchanged (unhealthy providers always moved to end, regardless of strategy)
- **Failover Logic**: Unchanged (ProxyServer tries providers in returned order, existing retry logic preserved)

---

## Testing Contract

### Unit Test Requirements

#### Test: Strategy Evaluation
- **Input**: Providers with known metrics, specific strategy
- **Expected**: Providers ordered according to strategy rules
- **Coverage**: All 4 strategies (failover, round-robin, least-latency, least-cost)

#### Test: Insufficient Samples
- **Input**: Providers with < 10 latency samples
- **Expected**: Providers excluded from least-latency evaluation, appended to end
- **Coverage**: Edge case handling

#### Test: Concurrent Access
- **Input**: 100 concurrent calls to `Select()` with same providers
- **Expected**: No race conditions, consistent results
- **Coverage**: Concurrency safety (run with `-race` flag)

### Integration Test Requirements

#### Test: End-to-End Strategy Routing
- **Setup**: Configure profile with `least-latency` strategy, send 10 requests to each provider
- **Action**: Send request, observe which provider is selected
- **Expected**: Provider with lowest latency selected first
- **Coverage**: Full request flow (ProfileProxy → LoadBalancer → ProxyServer)

#### Test: Metric Collection
- **Setup**: Send requests to providers, wait for metric cache refresh
- **Action**: Query `GetProviderLatencyMetrics()`
- **Expected**: Metrics reflect actual request latencies
- **Coverage**: Metric persistence and query accuracy

---

## Version History

### v1.0 (2026-03-09)
- Initial contract definition
- Supports 4 strategies: failover, round-robin, least-latency, least-cost
- Metric-based selection with 100-request rolling window
- Concurrency-safe via RWMutex and atomic operations
