package proxy

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

// LoadBalancer selects providers based on configured strategy.
type LoadBalancer struct {
	db           *LogDB
	pricing      map[string]*config.ModelPricing
	mu           sync.RWMutex
	rrCounter    uint64 // for round-robin
	metricsCache map[string]*ProviderMetrics
	cacheTime    time.Time
	cacheTTL     time.Duration
}

// NewLoadBalancer creates a new load balancer.
func NewLoadBalancer(db *LogDB) *LoadBalancer {
	return &LoadBalancer{
		db:           db,
		pricing:      config.GetPricing(),
		metricsCache: make(map[string]*ProviderMetrics),
		cacheTTL:     30 * time.Second,
	}
}

// ReloadPricing refreshes the pricing data from config.
func (lb *LoadBalancer) ReloadPricing() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.pricing = config.GetPricing()
}

// Select chooses providers in order based on the strategy.
// Returns a reordered slice of providers (does not modify original).
func (lb *LoadBalancer) Select(providers []*Provider, strategy config.LoadBalanceStrategy, model string) []*Provider {
	if len(providers) <= 1 {
		return providers
	}

	switch strategy {
	case config.LoadBalanceRoundRobin:
		return lb.selectRoundRobin(providers)
	case config.LoadBalanceLeastLatency:
		return lb.selectLeastLatency(providers)
	case config.LoadBalanceLeastCost:
		return lb.selectLeastCost(providers, model)
	default:
		// Failover: return as-is (first healthy provider wins)
		return lb.selectFailover(providers)
	}
}

// selectFailover returns providers in original order, with unhealthy ones moved to the end.
func (lb *LoadBalancer) selectFailover(providers []*Provider) []*Provider {
	result := make([]*Provider, 0, len(providers))
	unhealthy := make([]*Provider, 0)

	for _, p := range providers {
		if p.IsHealthy() {
			result = append(result, p)
		} else {
			unhealthy = append(unhealthy, p)
		}
	}

	return append(result, unhealthy...)
}

// selectRoundRobin rotates through providers evenly.
func (lb *LoadBalancer) selectRoundRobin(providers []*Provider) []*Provider {
	n := len(providers)
	if n == 0 {
		return providers
	}

	// Get next index atomically
	idx := atomic.AddUint64(&lb.rrCounter, 1) % uint64(n)

	// Rotate the slice starting from idx
	result := make([]*Provider, n)
	for i := 0; i < n; i++ {
		result[i] = providers[(int(idx)+i)%n]
	}

	// Move unhealthy to end while preserving rotation order
	return lb.moveUnhealthyToEnd(result)
}

// selectLeastLatency orders providers by average latency (lowest first).
func (lb *LoadBalancer) selectLeastLatency(providers []*Provider) []*Provider {
	metrics := lb.getMetricsCache()

	// Create a copy with latency info
	type providerLatency struct {
		provider *Provider
		latency  float64
		healthy  bool
	}

	items := make([]providerLatency, len(providers))
	for i, p := range providers {
		items[i] = providerLatency{
			provider: p,
			latency:  float64(^uint(0) >> 1), // max value as default
			healthy:  p.IsHealthy(),
		}

		if m, ok := metrics[p.Name]; ok && m.TotalRequests > 0 {
			items[i].latency = m.AvgLatencyMs
		}
	}

	// Sort by: healthy first, then by latency
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			swap := false
			if items[i].healthy && items[j].healthy {
				// Both healthy: sort by latency
				swap = items[i].latency > items[j].latency
			} else if !items[i].healthy && items[j].healthy {
				// Unhealthy before healthy: swap
				swap = true
			}
			if swap {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	result := make([]*Provider, len(items))
	for i, item := range items {
		result[i] = item.provider
	}

	return result
}

// selectLeastCost orders providers by cost for the given model (lowest first).
func (lb *LoadBalancer) selectLeastCost(providers []*Provider, model string) []*Provider {
	lb.mu.RLock()
	pricing := lb.pricing
	lb.mu.RUnlock()

	// Create a copy with cost info
	type providerCost struct {
		provider *Provider
		cost     float64
		healthy  bool
	}

	items := make([]providerCost, len(providers))
	for i, p := range providers {
		items[i] = providerCost{
			provider: p,
			cost:     float64(^uint(0) >> 1), // max value as default
			healthy:  p.IsHealthy(),
		}

		// Determine which model this provider would use
		providerModel := model
		if p.Model != "" {
			providerModel = p.Model
		}

		// Look up pricing
		if price := findModelPricing(providerModel, pricing); price != nil {
			// Use combined input+output cost as a simple metric
			items[i].cost = price.InputPerMillion + price.OutputPerMillion
		}
	}

	// Sort by: healthy first, then by cost
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			swap := false
			if items[i].healthy && items[j].healthy {
				// Both healthy: sort by cost
				swap = items[i].cost > items[j].cost
			} else if !items[i].healthy && items[j].healthy {
				// Unhealthy before healthy: swap
				swap = true
			}
			if swap {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	result := make([]*Provider, len(items))
	for i, item := range items {
		result[i] = item.provider
	}

	return result
}

// moveUnhealthyToEnd moves unhealthy providers to the end while preserving order.
func (lb *LoadBalancer) moveUnhealthyToEnd(providers []*Provider) []*Provider {
	healthy := make([]*Provider, 0, len(providers))
	unhealthy := make([]*Provider, 0)

	for _, p := range providers {
		if p.IsHealthy() {
			healthy = append(healthy, p)
		} else {
			unhealthy = append(unhealthy, p)
		}
	}

	return append(healthy, unhealthy...)
}

// getMetricsCache returns cached metrics or fetches fresh ones.
func (lb *LoadBalancer) getMetricsCache() map[string]*ProviderMetrics {
	lb.mu.RLock()
	if time.Since(lb.cacheTime) < lb.cacheTTL && len(lb.metricsCache) > 0 {
		cache := lb.metricsCache
		lb.mu.RUnlock()
		return cache
	}
	lb.mu.RUnlock()

	// Fetch fresh metrics
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Double-check after acquiring write lock
	if time.Since(lb.cacheTime) < lb.cacheTTL && len(lb.metricsCache) > 0 {
		return lb.metricsCache
	}

	if lb.db != nil {
		since := time.Now().Add(-1 * time.Hour)
		if metrics, err := lb.db.GetAllProviderMetrics(since); err == nil {
			lb.metricsCache = metrics
			lb.cacheTime = time.Now()
			return metrics
		}
	}

	return lb.metricsCache
}

// findModelPricing finds pricing for a model with partial matching.
func findModelPricing(model string, pricing map[string]*config.ModelPricing) *config.ModelPricing {
	if pricing == nil {
		return nil
	}

	// Exact match
	if p, ok := pricing[model]; ok {
		return p
	}

	// Try to match by model family
	for key, p := range pricing {
		if len(key) > 0 && len(model) > 0 {
			// Check if model contains the key or vice versa
			if contains(model, key) || contains(key, model) {
				return p
			}
		}
	}

	return nil
}

func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// --- Global load balancer ---

var globalLoadBalancer *LoadBalancer

// InitGlobalLoadBalancer initializes the global load balancer.
func InitGlobalLoadBalancer(db *LogDB) {
	globalLoadBalancer = NewLoadBalancer(db)
}

// GetGlobalLoadBalancer returns the global load balancer.
func GetGlobalLoadBalancer() *LoadBalancer {
	return globalLoadBalancer
}
