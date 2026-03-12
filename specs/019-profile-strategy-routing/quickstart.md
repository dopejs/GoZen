# Quickstart: Profile Strategy-Aware Provider Routing

**Feature**: 019-profile-strategy-routing
**Date**: 2026-03-09
**Audience**: Developers implementing this feature

## Overview

This quickstart guide provides a step-by-step walkthrough for implementing strategy-aware provider routing. Follow these steps in order to ensure correct implementation.

---

## Prerequisites

Before starting implementation:

1. **Read the specification**: [spec.md](./spec.md)
2. **Review the research**: [research.md](./research.md)
3. **Understand the data model**: [data-model.md](./data-model.md)
4. **Review the API contract**: [contracts/strategy-api.md](./contracts/strategy-api.md)
5. **Verify constitution compliance**: [plan.md](./plan.md#constitution-check)

---

## Implementation Phases

### Phase 1: Extend LogDB for Latency Metrics (TDD)

**Goal**: Add method to query provider latency metrics from SQLite database

**Steps**:

1. **Write test first** (`internal/proxy/logdb_test.go`):
   ```go
   func TestLogDB_GetProviderLatencyMetrics(t *testing.T) {
       // Setup: Create temp DB, insert test requests
       // Test: Query metrics for last 100 requests
       // Assert: Correct average latency, minimum 10 samples enforced
   }
   ```

2. **Implement method** (`internal/proxy/logdb.go`):
   ```go
   func (db *LogDB) GetProviderLatencyMetrics(since time.Time, limit int) (map[string]*ProviderMetrics, error) {
       // SQL query with GROUP BY provider, HAVING COUNT(*) >= 10
       // Return map[providerName]*ProviderMetrics
   }
   ```

3. **Run test**: `go test -v ./internal/proxy -run TestLogDB_GetProviderLatencyMetrics`

4. **Verify coverage**: `go test -cover ./internal/proxy` (target: ≥80%)

**Acceptance Criteria**:
- ✅ Test passes with correct latency averages
- ✅ Providers with < 10 samples excluded from result
- ✅ Query handles empty database gracefully
- ✅ Coverage ≥ 80%

---

### Phase 2: Extend LoadBalancer for Profile-Aware Selection (TDD)

**Goal**: Modify `LoadBalancer.Select()` to use profile strategy instead of global strategy

**Steps**:

1. **Write test first** (`internal/proxy/loadbalancer_test.go`):
   ```go
   func TestLoadBalancer_SelectWithStrategy(t *testing.T) {
       tests := []struct {
           name      string
           strategy  config.LoadBalanceStrategy
           providers []*Provider
           metrics   map[string]*ProviderMetrics
           want      []string // Expected provider order
       }{
           {"least-latency", config.LoadBalanceLeastLatency, ...},
           {"round-robin", config.LoadBalanceRoundRobin, ...},
           {"least-cost", config.LoadBalanceLeastCost, ...},
           {"failover", config.LoadBalanceFailover, ...},
       }
       // Run table-driven tests
   }
   ```

2. **Modify signature** (`internal/proxy/loadbalancer.go`):
   ```go
   // BEFORE: func (lb *LoadBalancer) Select(providers []*Provider, strategy config.LoadBalanceStrategy, model string)
   // AFTER: Same signature, but use passed strategy instead of global config
   ```

3. **Add logging** (`internal/proxy/loadbalancer.go`):
   ```go
   func (lb *LoadBalancer) Select(...) []*Provider {
       // ... existing logic ...
       lb.logger.Printf("[strategy] profile=%s strategy=%s selected=%s reason=%q candidates=%d",
           profileName, strategy, selected.Name, reason, candidateCount)
       return result
   }
   ```

4. **Run tests**: `go test -v ./internal/proxy -run TestLoadBalancer`

5. **Run race detector**: `go test -race ./internal/proxy`

**Acceptance Criteria**:
- ✅ All 4 strategies tested (failover, round-robin, least-latency, least-cost)
- ✅ Insufficient samples handled correctly (< 10 samples → excluded)
- ✅ Logging includes profile, strategy, selected provider, reason
- ✅ No race conditions detected
- ✅ Coverage ≥ 80%

---

### Phase 3: Connect ProfileProxy to LoadBalancer (TDD)

**Goal**: Pass profile strategy from ProfileProxy to LoadBalancer

**Steps**:

1. **Write test first** (`internal/proxy/profile_proxy_test.go`):
   ```go
   func TestProfileProxy_StrategyRouting(t *testing.T) {
       // Setup: Create profile with least-latency strategy
       // Test: Send request, verify LoadBalancer.Select() called with correct strategy
       // Assert: Provider with lowest latency selected
   }
   ```

2. **Modify ProfileProxy** (`internal/proxy/profile_proxy.go`):
   ```go
   func (pp *ProfileProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
       // ... existing code ...
       profileCfg, err := pp.resolveProfileConfig(route)
       strategy := profileCfg.Strategy // Extract strategy from profile
       if strategy == "" {
           strategy = config.LoadBalanceFailover // Default
       }
       // Pass strategy to LoadBalancer
       lb := GetGlobalLoadBalancer()
       orderedProviders := lb.Select(providers, strategy, model)
       // ... rest of existing code ...
   }
   ```

3. **Run tests**: `go test -v ./internal/proxy -run TestProfileProxy`

4. **Integration test**: `go test -v ./tests/integration -run TestStrategyRouting`

**Acceptance Criteria**:
- ✅ Profile strategy passed to LoadBalancer correctly
- ✅ Default strategy (failover) used when profile.Strategy is empty
- ✅ Integration test verifies end-to-end flow
- ✅ Coverage ≥ 80%

---

### Phase 4: Add Insufficient Sample Handling (TDD)

**Goal**: Exclude providers with < 10 latency samples from least-latency evaluation

**Steps**:

1. **Write test first** (`internal/proxy/loadbalancer_test.go`):
   ```go
   func TestLoadBalancer_InsufficientSamples(t *testing.T) {
       // Setup: Providers with 3, 5, 15 samples
       // Test: Select with least-latency strategy
       // Assert: Only provider with 15 samples included, others appended to end
   }
   ```

2. **Modify selectLeastLatency** (`internal/proxy/loadbalancer.go`):
   ```go
   func (lb *LoadBalancer) selectLeastLatency(providers []*Provider) []*Provider {
       metrics := lb.getMetricsCache()
       validProviders := []providerLatency{}
       insufficientProviders := []*Provider{}

       for _, p := range providers {
           if m, ok := metrics[p.Name]; ok && m.TotalRequests >= 10 {
               validProviders = append(validProviders, providerLatency{...})
           } else {
               insufficientProviders = append(insufficientProviders, p)
           }
       }

       // Sort validProviders by latency
       // Append insufficientProviders to end
       return result
   }
   ```

3. **Run tests**: `go test -v ./internal/proxy -run TestLoadBalancer_InsufficientSamples`

**Acceptance Criteria**:
- ✅ Providers with < 10 samples excluded from sorting
- ✅ Excluded providers appended to end (preserve configured order)
- ✅ Log warning when all providers have insufficient samples
- ✅ Coverage ≥ 80%

---

### Phase 5: Add Concurrency Safety Tests (TDD)

**Goal**: Verify thread-safety of strategy evaluation under concurrent load

**Steps**:

1. **Write test first** (`internal/proxy/loadbalancer_test.go`):
   ```go
   func TestLoadBalancer_ConcurrentAccess(t *testing.T) {
       lb := NewLoadBalancer(db)
       var wg sync.WaitGroup
       for i := 0; i < 100; i++ {
           wg.Add(1)
           go func() {
               defer wg.Done()
               lb.Select(providers, config.LoadBalanceLeastLatency, "claude-sonnet-4-5")
           }()
       }
       wg.Wait()
       // Assert: No race conditions, consistent results
   }
   ```

2. **Run with race detector**: `go test -race -v ./internal/proxy -run TestLoadBalancer_ConcurrentAccess`

3. **Verify RWMutex usage** (`internal/proxy/loadbalancer.go`):
   ```go
   func (lb *LoadBalancer) getMetricsCache() map[string]*ProviderMetrics {
       lb.mu.RLock()
       if time.Since(lb.cacheTime) < lb.cacheTTL {
           cache := lb.metricsCache
           lb.mu.RUnlock()
           return cache // Return snapshot, not reference
       }
       lb.mu.RUnlock()
       // ... refresh logic ...
   }
   ```

**Acceptance Criteria**:
- ✅ 100 concurrent calls complete without race conditions
- ✅ Race detector reports no issues
- ✅ Metric snapshots are read-only (no shared mutable state)
- ✅ Coverage ≥ 80%

---

### Phase 6: Integration Testing

**Goal**: Verify end-to-end strategy routing with real daemon

**Steps**:

1. **Write integration test** (`tests/integration/strategy_routing_test.go`):
   ```go
   func TestIntegration_StrategyRouting(t *testing.T) {
       // Setup: Start dev daemon, configure profile with least-latency strategy
       // Action: Send 10 requests to each provider (build latency history)
       // Action: Send test request, observe which provider is selected
       // Assert: Provider with lowest latency selected first
   }
   ```

2. **Run integration test**: `./scripts/dev.sh && go test -v ./tests/integration -run TestIntegration_StrategyRouting`

3. **Verify logs**: Check daemon logs for `[strategy]` entries

**Acceptance Criteria**:
- ✅ Integration test passes with real daemon
- ✅ Strategy decisions logged correctly
- ✅ Provider with lowest latency selected first
- ✅ Failover works if selected provider fails

---

## Testing Checklist

Before marking implementation complete, verify:

- [ ] **Unit Tests**: All unit tests pass (`go test ./internal/proxy`)
- [ ] **Race Detector**: No race conditions (`go test -race ./internal/proxy`)
- [ ] **Coverage**: ≥80% coverage (`go test -cover ./internal/proxy`)
- [ ] **Integration Tests**: End-to-end tests pass (`go test ./tests/integration`)
- [ ] **Manual Testing**: Test with dev daemon (`./scripts/dev.sh`)
- [ ] **Logging**: Strategy decisions logged with correct format
- [ ] **Backward Compatibility**: Existing configs work without modification
- [ ] **Performance**: Strategy evaluation < 5ms (measure with benchmarks)

---

## Common Pitfalls

### Pitfall 1: Modifying Input Slice

**Problem**: Modifying the input `providers` slice instead of creating a new one

**Solution**: Always create a new slice:
```go
result := make([]*Provider, len(providers))
copy(result, providers)
// Now modify result, not providers
```

---

### Pitfall 2: Race Conditions in Metric Cache

**Problem**: Returning reference to shared `metricsCache` map instead of snapshot

**Solution**: Return copy, not reference:
```go
lb.mu.RLock()
cache := lb.metricsCache // This is a reference, not a copy!
lb.mu.RUnlock()
return cache // WRONG: Caller can mutate shared state

// CORRECT:
lb.mu.RLock()
snapshot := make(map[string]*ProviderMetrics, len(lb.metricsCache))
for k, v := range lb.metricsCache {
    snapshot[k] = v // Shallow copy is sufficient (ProviderMetrics is immutable)
}
lb.mu.RUnlock()
return snapshot
```

---

### Pitfall 3: Forgetting Minimum Sample Size

**Problem**: Including providers with < 10 samples in least-latency evaluation

**Solution**: Always check sample count:
```go
if m, ok := metrics[p.Name]; ok && m.TotalRequests >= 10 {
    // Include in evaluation
} else {
    // Exclude, append to end
}
```

---

### Pitfall 4: Not Logging Strategy Decisions

**Problem**: Forgetting to log which provider was selected and why

**Solution**: Always log after selection:
```go
lb.logger.Printf("[strategy] profile=%s strategy=%s selected=%s reason=%q candidates=%d",
    profileName, strategy, selected.Name, reason, candidateCount)
```

---

## Debugging Tips

### Tip 1: Enable Verbose Logging

```bash
# Set log level to DEBUG
export GOZEN_LOG_LEVEL=debug
./scripts/dev.sh
```

### Tip 2: Check Metric Cache

```go
// Add temporary debug logging
metrics := lb.getMetricsCache()
for name, m := range metrics {
    log.Printf("[debug] provider=%s latency=%.2fms samples=%d", name, m.AvgLatencyMs, m.TotalRequests)
}
```

### Tip 3: Verify SQL Query

```bash
# Query LogDB directly
sqlite3 ~/.zen/logs.db "SELECT provider, COUNT(*), AVG(latency_ms) FROM requests WHERE timestamp > datetime('now', '-1 hour') GROUP BY provider HAVING COUNT(*) >= 10;"
```

---

## Performance Benchmarks

### Benchmark: Strategy Evaluation

```go
func BenchmarkLoadBalancer_Select(b *testing.B) {
    lb := NewLoadBalancer(db)
    providers := []*Provider{...} // 50 providers

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        lb.Select(providers, config.LoadBalanceLeastLatency, "claude-sonnet-4-5")
    }
}
```

**Target**: < 5ms per operation (99th percentile)

---

## Next Steps

After completing implementation:

1. **Run full test suite**: `go test ./...`
2. **Update CLAUDE.md**: Add feature to "Active Technologies" section
3. **Create PR**: Use `/cpm` skill to commit, push, and create PR
4. **Request review**: Tag reviewers in PR description
5. **Merge**: After approval, merge to main
6. **Tag release**: Follow release process in CLAUDE.md

---

## References

- **Specification**: [spec.md](./spec.md)
- **Research**: [research.md](./research.md)
- **Data Model**: [data-model.md](./data-model.md)
- **API Contract**: [contracts/strategy-api.md](./contracts/strategy-api.md)
- **Implementation Plan**: [plan.md](./plan.md)
- **Go Testing**: https://go.dev/doc/tutorial/add-a-test
- **Go Concurrency**: https://go.dev/blog/pipelines
