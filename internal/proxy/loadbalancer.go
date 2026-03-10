package proxy

import (
	"fmt"
	"log"
	"math/rand"
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
	rrCounter    uint64 // global fallback for round-robin
	rrCounters   map[string]*uint64 // per-profile round-robin counters
	metricsCache map[string]*ProviderMetrics
	cacheTime    time.Time
	cacheTTL     time.Duration
}

// NewLoadBalancer creates a new load balancer.
func NewLoadBalancer(db *LogDB) *LoadBalancer {
	return &LoadBalancer{
		db:           db,
		pricing:      config.GetPricing(),
		rrCounters:   make(map[string]*uint64),
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
// profile is used for per-profile state isolation (e.g. round-robin counters).
// modelOverrides maps provider name → override model for scenario routes (used by least-cost).
func (lb *LoadBalancer) Select(providers []*Provider, strategy config.LoadBalanceStrategy, model string, profile string, modelOverrides map[string]string) []*Provider {
	if len(providers) <= 1 {
		return providers
	}

	var result []*Provider
	var strategyName string
	var reason string

	switch strategy {
	case config.LoadBalanceRoundRobin:
		strategyName = "round-robin"
		result = lb.selectRoundRobin(providers, profile)
		if len(result) > 0 {
			reason = "round-robin rotation"
		}
	case config.LoadBalanceLeastLatency:
		strategyName = "least-latency"
		result = lb.selectLeastLatency(providers)
		if len(result) > 0 {
			metrics := lb.getMetricsCache()
			if m, ok := metrics[result[0].Name]; ok {
				reason = fmt.Sprintf("lowest latency: %.2fms", m.AvgLatencyMs)
			} else {
				reason = "insufficient samples, using configured order"
			}
		}
	case config.LoadBalanceLeastCost:
		strategyName = "least-cost"
		result = lb.selectLeastCost(providers, model, modelOverrides)
		if len(result) > 0 {
			// Get pricing info for the selected provider
			lb.mu.RLock()
			pricing := lb.pricing
			lb.mu.RUnlock()

			providerModel := model
			if result[0].Model != "" {
				providerModel = result[0].Model
			}
			if modelOverrides != nil {
				if override, ok := modelOverrides[result[0].Name]; ok && override != "" {
					providerModel = override
				}
			}

			if price := findModelPricing(providerModel, pricing); price != nil {
				totalCost := price.InputPerMillion + price.OutputPerMillion
				reason = fmt.Sprintf("lowest cost: $%.3f/1M tokens", totalCost)
			} else {
				reason = "lowest cost"
			}
		}
	case config.LoadBalanceWeighted:
		strategyName = "weighted"
		result = lb.selectWeighted(providers)
		if len(result) > 0 {
			// Calculate percentage for selected provider
			totalWeight := 0
			selectedWeight := result[0].Weight
			for _, p := range providers {
				if p.IsHealthy() {
					totalWeight += p.Weight
				}
			}
			// If no weights configured, use equal weights for percentage calculation
			if totalWeight == 0 {
				healthyCount := 0
				for _, p := range providers {
					if p.IsHealthy() {
						healthyCount++
					}
				}
				if healthyCount > 0 {
					percentage := 100.0 / float64(healthyCount)
					reason = fmt.Sprintf("weighted: %.1f%%", percentage)
				} else {
					reason = "weighted: equal distribution"
				}
			} else {
				percentage := float64(selectedWeight) / float64(totalWeight) * 100
				reason = fmt.Sprintf("weighted: %.1f%%", percentage)
			}
		}
	default:
		strategyName = "failover"
		result = lb.selectFailover(providers)
		if len(result) > 0 {
			reason = "first healthy provider"
		}
	}

	// Log strategy decision
	if len(result) > 0 {
		log.Printf("[strategy] strategy=%s selected=%s reason=%q candidates=%d",
			strategyName, result[0].Name, reason, len(providers))
	}

	return result
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

// selectRoundRobin rotates evenly across healthy providers only.
// Unhealthy providers are appended at the end as fallbacks.
// Uses a per-profile counter so different profiles have independent rotation.
func (lb *LoadBalancer) selectRoundRobin(providers []*Provider, profile string) []*Provider {
	if len(providers) == 0 {
		return providers
	}

	// Separate healthy and unhealthy first so counter only rotates among healthy
	healthy := make([]*Provider, 0, len(providers))
	unhealthy := make([]*Provider, 0)
	for _, p := range providers {
		if p.IsHealthy() {
			healthy = append(healthy, p)
		} else {
			unhealthy = append(unhealthy, p)
		}
	}

	if len(healthy) == 0 {
		// All unhealthy — rotate through all as last resort
		return providers
	}

	// Rotate only among healthy providers
	counter := lb.getProfileRRCounter(profile)
	idx := atomic.AddUint64(counter, 1) % uint64(len(healthy))

	result := make([]*Provider, 0, len(providers))
	for i := 0; i < len(healthy); i++ {
		result = append(result, healthy[(int(idx)+i)%len(healthy)])
	}

	// Append unhealthy as fallbacks
	return append(result, unhealthy...)
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
			} else if items[i].healthy && !items[j].healthy {
				// i healthy, j unhealthy: keep order (don't swap)
				swap = false
			} else if !items[i].healthy && items[j].healthy {
				// i unhealthy, j healthy: swap to put healthy first
				swap = true
			} else {
				// Both unhealthy: sort by latency
				swap = items[i].latency > items[j].latency
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
// modelOverrides maps provider name → override model (from scenario routes).
func (lb *LoadBalancer) selectLeastCost(providers []*Provider, model string, modelOverrides map[string]string) []*Provider {
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

		// Determine which model this provider would actually use:
		// 1. Scenario model override (highest precedence)
		// 2. Provider's own model
		// 3. Request body model (fallback)
		providerModel := model
		if p.Model != "" {
			providerModel = p.Model
		}
		if modelOverrides != nil {
			if override, ok := modelOverrides[p.Name]; ok && override != "" {
				providerModel = override
			}
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

// getProfileRRCounter returns the round-robin counter for a given profile.
// Creates a new counter if one doesn't exist. Falls back to global counter if profile is empty.
func (lb *LoadBalancer) getProfileRRCounter(profile string) *uint64 {
	if profile == "" {
		return &lb.rrCounter
	}

	lb.mu.RLock()
	if c, ok := lb.rrCounters[profile]; ok {
		lb.mu.RUnlock()
		return c
	}
	lb.mu.RUnlock()

	lb.mu.Lock()
	defer lb.mu.Unlock()
	// Double-check
	if c, ok := lb.rrCounters[profile]; ok {
		return c
	}
	c := new(uint64)
	lb.rrCounters[profile] = c
	return c
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
		since := time.Now().UTC().Add(-1 * time.Hour)
		if metrics, err := lb.db.GetProviderLatencyMetrics(since, 100); err == nil {
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

// selectWeighted performs weighted random selection among healthy providers.
// Weights are recalculated to exclude unhealthy providers.
// If no weights are configured (all weights are 0), uses equal weights.
func (lb *LoadBalancer) selectWeighted(providers []*Provider) []*Provider {
	if len(providers) == 0 {
		return providers
	}

	// Separate healthy and unhealthy providers
	healthy := make([]*Provider, 0, len(providers))
	unhealthy := make([]*Provider, 0)

	for _, p := range providers {
		if p.IsHealthy() {
			healthy = append(healthy, p)
		} else {
			unhealthy = append(unhealthy, p)
		}
	}

	if len(healthy) == 0 {
		// No healthy providers, return all in original order
		return providers
	}

	// Calculate total weight of healthy providers
	totalWeight := 0
	weights := make([]int, len(healthy))
	for i, p := range healthy {
		weights[i] = p.Weight
		totalWeight += p.Weight
	}

	// If no weights configured (all 0), use equal weights
	if totalWeight == 0 {
		totalWeight = len(healthy)
		for i := range weights {
			weights[i] = 1
		}
	}

	// Weighted random selection
	randVal := lb.weightedRand(totalWeight)
	cumulative := 0
	selectedIdx := 0

	for i := range healthy {
		cumulative += weights[i]
		if randVal < cumulative {
			selectedIdx = i
			break
		}
	}

	// Rotate to put selected provider first
	result := make([]*Provider, 0, len(providers))
	result = append(result, healthy[selectedIdx])
	for i, p := range healthy {
		if i != selectedIdx {
			result = append(result, p)
		}
	}
	result = append(result, unhealthy...)

	return result
}

// weightedRand returns a random number in [0, max)
func (lb *LoadBalancer) weightedRand(max int) int {
	if max <= 0 {
		return 0
	}
	// Use atomic counter as seed for better distribution
	seed := atomic.AddUint64(&lb.rrCounter, 1)
	// Create a new random source with the seed
	src := rand.NewSource(int64(seed))
	r := rand.New(src)
	return r.Intn(max)
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
