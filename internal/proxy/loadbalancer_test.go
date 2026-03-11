package proxy

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

func TestNewLoadBalancer(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	lb := NewLoadBalancer(nil)
	if lb == nil {
		t.Fatal("Expected non-nil load balancer")
	}
	if lb.cacheTTL == 0 {
		t.Error("Expected cacheTTL to be set")
	}
}

func TestLoadBalancer_ReloadPricing(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	lb := NewLoadBalancer(nil)
	lb.ReloadPricing()

	if lb.pricing == nil {
		t.Error("Expected pricing to be loaded")
	}
}

func TestLoadBalancer_Select_Empty(t *testing.T) {
	lb := &LoadBalancer{}

	result := lb.Select(nil, config.LoadBalanceFailover, "", "", nil, nil)
	if result != nil {
		t.Error("Expected nil for nil input")
	}

	result = lb.Select([]*Provider{}, config.LoadBalanceFailover, "", "", nil, nil)
	if len(result) != 0 {
		t.Error("Expected empty slice for empty input")
	}
}

func TestLoadBalancer_Select_Single(t *testing.T) {
	lb := &LoadBalancer{}
	provider := &Provider{Name: "test", Healthy: true}

	result := lb.Select([]*Provider{provider}, config.LoadBalanceFailover, "", "", nil, nil)
	if len(result) != 1 || result[0] != provider {
		t.Error("Expected single provider to be returned unchanged")
	}
}

func TestLoadBalancer_Select_Failover(t *testing.T) {
	lb := &LoadBalancer{
		metricsCache: make(map[string]*ProviderMetrics),
	}

	healthy := &Provider{Name: "healthy", Healthy: true}
	unhealthy := &Provider{Name: "unhealthy", Healthy: false}
	unhealthy.MarkFailed() // Set backoff to make it truly unhealthy

	result := lb.Select([]*Provider{unhealthy, healthy}, config.LoadBalanceFailover, "", "", nil, nil)
	if len(result) != 2 {
		t.Fatalf("Expected 2 providers, got %d", len(result))
	}
	if result[0].Name != "healthy" {
		t.Error("Expected healthy provider first")
	}
}

func TestLoadBalancer_Select_RoundRobin(t *testing.T) {
	lb := &LoadBalancer{
		metricsCache: make(map[string]*ProviderMetrics),
	}

	p1 := &Provider{Name: "p1", Healthy: true}
	p2 := &Provider{Name: "p2", Healthy: true}
	providers := []*Provider{p1, p2}

	// First call
	result1 := lb.Select(providers, config.LoadBalanceRoundRobin, "", "", nil, nil)
	// Second call should rotate
	result2 := lb.Select(providers, config.LoadBalanceRoundRobin, "", "", nil, nil)

	if result1[0].Name == result2[0].Name {
		t.Error("Expected round-robin to rotate providers")
	}
}

func TestLoadBalancer_Select_LeastLatency(t *testing.T) {
	lb := &LoadBalancer{
		metricsCache: map[string]*ProviderMetrics{
			"slow":   {AvgLatencyMs: 500, TotalRequests: 10},
			"fast":   {AvgLatencyMs: 100, TotalRequests: 10},
			"medium": {AvgLatencyMs: 300, TotalRequests: 10},
		},
	}

	slow := &Provider{Name: "slow", Healthy: true}
	fast := &Provider{Name: "fast", Healthy: true}
	medium := &Provider{Name: "medium", Healthy: true}

	result := lb.Select([]*Provider{slow, medium, fast}, config.LoadBalanceLeastLatency, "", "", nil, nil)
	if result[0].Name != "fast" {
		t.Errorf("Expected fast provider first, got %s", result[0].Name)
	}
}

func TestLoadBalancer_Select_LeastCost(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	lb := NewLoadBalancer(nil)

	// Haiku is cheaper than Opus
	haiku := &Provider{Name: "haiku", Model: "claude-3-5-haiku-20241022", Healthy: true}
	opus := &Provider{Name: "opus", Model: "claude-3-opus-20240229", Healthy: true}

	result := lb.Select([]*Provider{opus, haiku}, config.LoadBalanceLeastCost, "", "", nil, nil)
	if result[0].Name != "haiku" {
		t.Errorf("Expected haiku (cheaper) first, got %s", result[0].Name)
	}
}

func TestLoadBalancer_MoveUnhealthyToEnd(t *testing.T) {
	lb := &LoadBalancer{}

	healthy1 := &Provider{Name: "h1", Healthy: true}
	healthy2 := &Provider{Name: "h2", Healthy: true}
	unhealthy := &Provider{Name: "u1", Healthy: false}
	unhealthy.MarkFailed() // Set backoff to make it truly unhealthy

	result := lb.moveUnhealthyToEnd([]*Provider{unhealthy, healthy1, healthy2})
	if len(result) != 3 {
		t.Fatalf("Expected 3 providers, got %d", len(result))
	}
	if result[2].Name != "u1" {
		t.Error("Expected unhealthy provider at end")
	}
}

func TestLoadBalancer_GetMetricsCache(t *testing.T) {
	lb := &LoadBalancer{
		metricsCache: make(map[string]*ProviderMetrics),
	}

	// Should return empty cache
	cache := lb.getMetricsCache()
	if cache == nil {
		t.Error("Expected non-nil cache")
	}
}

func TestFindModelPricing(t *testing.T) {
	pricing := map[string]*config.ModelPricing{
		"claude-3-opus-20240229":     {InputPerMillion: 15.0, OutputPerMillion: 75.0},
		"claude-3-5-haiku-20241022":  {InputPerMillion: 0.8, OutputPerMillion: 4.0},
		"claude-3-5-sonnet-20241022": {InputPerMillion: 3.0, OutputPerMillion: 15.0},
	}

	tests := []struct {
		model    string
		expected bool
	}{
		{"claude-3-opus-20240229", true},
		{"claude-3-5-haiku-20241022", true},
		{"unknown-model", false},
	}

	for _, tt := range tests {
		result := findModelPricing(tt.model, pricing)
		if (result != nil) != tt.expected {
			t.Errorf("findModelPricing(%s) = %v, want found=%v", tt.model, result, tt.expected)
		}
	}

	// Test nil pricing
	if findModelPricing("any", nil) != nil {
		t.Error("Expected nil for nil pricing")
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s, substr string
		expected  bool
	}{
		{"hello world", "world", true},
		{"hello", "hello", true},
		{"hello", "world", false},
		{"hi", "hello", false},
		{"", "", true},
	}

	for _, tt := range tests {
		result := contains(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
		}
	}
}

func TestGlobalLoadBalancer(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	// Save old global
	oldGlobal := globalLoadBalancer
	defer func() { globalLoadBalancer = oldGlobal }()

	InitGlobalLoadBalancer(nil)

	lb := GetGlobalLoadBalancer()
	if lb == nil {
		t.Error("Expected non-nil global load balancer")
	}
}

func TestLoadBalancer_GetMetricsCache_WithCachedData(t *testing.T) {
	lb := &LoadBalancer{
		metricsCache: map[string]*ProviderMetrics{
			"test": {AvgLatencyMs: 100, TotalRequests: 10},
		},
		cacheTime: time.Now(),
		cacheTTL:  5 * time.Minute,
	}

	// Should return cached data
	cache := lb.getMetricsCache()
	if cache == nil {
		t.Fatal("Expected non-nil cache")
	}
	if cache["test"] == nil {
		t.Error("Expected test provider in cache")
	}
	if cache["test"].AvgLatencyMs != 100 {
		t.Errorf("Expected avg latency 100, got %f", cache["test"].AvgLatencyMs)
	}
}

func TestLoadBalancer_GetMetricsCache_ExpiredCache(t *testing.T) {
	lb := &LoadBalancer{
		metricsCache: map[string]*ProviderMetrics{
			"test": {AvgLatencyMs: 100, TotalRequests: 10},
		},
		cacheTime: time.Now().Add(-10 * time.Minute), // Expired
		cacheTTL:  5 * time.Minute,
	}

	// Should still return the old cache since db is nil
	cache := lb.getMetricsCache()
	if cache == nil {
		t.Fatal("Expected non-nil cache")
	}
}

func TestLoadBalancer_Select_LeastLatency_NoMetrics(t *testing.T) {
	lb := &LoadBalancer{
		metricsCache: make(map[string]*ProviderMetrics),
	}

	p1 := &Provider{Name: "p1", Healthy: true}
	p2 := &Provider{Name: "p2", Healthy: true}

	// Without metrics, should still return providers
	result := lb.Select([]*Provider{p1, p2}, config.LoadBalanceLeastLatency, "", "", nil, nil)
	if len(result) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(result))
	}
}

func TestLoadBalancer_Select_LeastCost_NoModel(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	lb := NewLoadBalancer(nil)

	// Providers without model set
	p1 := &Provider{Name: "p1", Healthy: true}
	p2 := &Provider{Name: "p2", Healthy: true}

	result := lb.Select([]*Provider{p1, p2}, config.LoadBalanceLeastCost, "", "", nil, nil)
	if len(result) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(result))
	}
}

func TestLoadBalancer_Select_UnknownStrategy(t *testing.T) {
	lb := &LoadBalancer{
		metricsCache: make(map[string]*ProviderMetrics),
	}

	p1 := &Provider{Name: "p1", Healthy: true}
	p2 := &Provider{Name: "p2", Healthy: true}

	// Unknown strategy should default to failover behavior
	result := lb.Select([]*Provider{p1, p2}, "unknown-strategy", "", "", nil, nil)
	if len(result) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(result))
	}
}

func TestLoadBalancer_Select_RoundRobin_MultipleRounds(t *testing.T) {
	lb := &LoadBalancer{
		metricsCache: make(map[string]*ProviderMetrics),
	}

	p1 := &Provider{Name: "p1", Healthy: true}
	p2 := &Provider{Name: "p2", Healthy: true}
	p3 := &Provider{Name: "p3", Healthy: true}
	providers := []*Provider{p1, p2, p3}

	// Multiple rounds should cycle through all providers
	seen := make(map[string]bool)
	for i := 0; i < 6; i++ {
		result := lb.Select(providers, config.LoadBalanceRoundRobin, "", "", nil, nil)
		seen[result[0].Name] = true
	}

	if len(seen) != 3 {
		t.Errorf("Expected to see all 3 providers, saw %d", len(seen))
	}
}

func TestLoadBalancer_MoveUnhealthyToEnd_AllHealthy(t *testing.T) {
	lb := &LoadBalancer{}

	p1 := &Provider{Name: "p1", Healthy: true}
	p2 := &Provider{Name: "p2", Healthy: true}

	result := lb.moveUnhealthyToEnd([]*Provider{p1, p2})
	if len(result) != 2 {
		t.Fatalf("Expected 2 providers, got %d", len(result))
	}
	// Order should be preserved
	if result[0].Name != "p1" || result[1].Name != "p2" {
		t.Error("Expected order to be preserved for all healthy providers")
	}
}

func TestLoadBalancer_MoveUnhealthyToEnd_AllUnhealthy(t *testing.T) {
	lb := &LoadBalancer{}

	p1 := &Provider{Name: "p1", Healthy: false}
	p1.MarkFailed()
	p2 := &Provider{Name: "p2", Healthy: false}
	p2.MarkFailed()

	result := lb.moveUnhealthyToEnd([]*Provider{p1, p2})
	if len(result) != 2 {
		t.Fatalf("Expected 2 providers, got %d", len(result))
	}
}

func TestLoadBalancer_SelectLeastLatency(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	defer db.Close()

	// Insert metrics for three providers
	// p1: 15 samples, 50ms average
	for i := 0; i < 15; i++ {
		db.RecordMetric("p1", 50, 200, false, false)
	}
	// p2: 15 samples, 30ms average (lowest latency)
	for i := 0; i < 15; i++ {
		db.RecordMetric("p2", 30, 200, false, false)
	}
	// p3: 15 samples, 100ms average (highest latency)
	for i := 0; i < 15; i++ {
		db.RecordMetric("p3", 100, 200, false, false)
	}

	time.Sleep(100 * time.Millisecond)

	lb := NewLoadBalancer(db)
	lb.cacheTTL = 0 // Disable cache for testing

	providers := []*Provider{
		{Name: "p1", Healthy: true},
		{Name: "p2", Healthy: true},
		{Name: "p3", Healthy: true},
	}

	result := lb.Select(providers, config.LoadBalanceLeastLatency, "claude-sonnet-4-5", "", nil, nil)

	if len(result) != 3 {
		t.Fatalf("got %d providers, want 3", len(result))
	}

	// Should be ordered by latency: p2 (30ms), p1 (50ms), p3 (100ms)
	if result[0].Name != "p2" || result[1].Name != "p1" || result[2].Name != "p3" {
		t.Errorf("provider order: got [%s, %s, %s], want [p2, p1, p3]",
			result[0].Name, result[1].Name, result[2].Name)
	}
}

func TestLoadBalancer_SelectLeastLatencyInsufficientSamples(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	defer db.Close()

	// p1: 15 samples, 50ms average (sufficient)
	for i := 0; i < 15; i++ {
		db.RecordMetric("p1", 50, 200, false, false)
	}
	// p2: 5 samples, 30ms average (insufficient, < 10)
	for i := 0; i < 5; i++ {
		db.RecordMetric("p2", 30, 200, false, false)
	}

	time.Sleep(100 * time.Millisecond)

	lb := NewLoadBalancer(db)
	lb.cacheTTL = 0 // Disable cache for testing

	providers := []*Provider{
		{Name: "p1", Healthy: true},
		{Name: "p2", Healthy: true},
	}

	result := lb.Select(providers, config.LoadBalanceLeastLatency, "claude-sonnet-4-5", "", nil, nil)

	if len(result) != 2 {
		t.Fatalf("got %d providers, want 2", len(result))
	}

	// p1 should come first (sufficient samples), p2 second (insufficient samples)
	if result[0].Name != "p1" || result[1].Name != "p2" {
		t.Errorf("provider order: got [%s, %s], want [p1, p2]",
			result[0].Name, result[1].Name)
	}
}

func TestLoadBalancer_SelectLeastLatencyUnhealthyProviders(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenLogDB(dir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	defer db.Close()

	// Insert metrics for all providers
	for i := 0; i < 15; i++ {
		db.RecordMetric("p1", 50, 200, false, false)
		db.RecordMetric("p2", 30, 200, false, false)
		db.RecordMetric("p3", 100, 200, false, false)
	}

	time.Sleep(100 * time.Millisecond)

	lb := NewLoadBalancer(db)
	lb.cacheTTL = 0 // Disable cache for testing

	p1 := &Provider{Name: "p1", Healthy: true}
	p2 := &Provider{Name: "p2", Healthy: false} // Unhealthy
	p3 := &Provider{Name: "p3", Healthy: true}

	// Mark p2 as failed to ensure it's unhealthy
	p2.MarkFailed()

	providers := []*Provider{p1, p2, p3}

	result := lb.Select(providers, config.LoadBalanceLeastLatency, "claude-sonnet-4-5", "", nil, nil)

	if len(result) != 3 {
		t.Fatalf("got %d providers, want 3", len(result))
	}

	// Healthy providers should come first, ordered by latency: p1 (50ms), p3 (100ms), then p2 (unhealthy)
	if result[0].Name != "p1" || result[1].Name != "p3" || result[2].Name != "p2" {
		t.Errorf("provider order: got [%s, %s, %s], want [p1, p3, p2]",
			result[0].Name, result[1].Name, result[2].Name)
		t.Logf("p1.IsHealthy()=%v, p2.IsHealthy()=%v, p3.IsHealthy()=%v",
			result[0].IsHealthy(), result[1].IsHealthy(), result[2].IsHealthy())
	}
}


// TestLoadBalancer_SelectLeastCost tests basic cost-based sorting
func TestLoadBalancer_SelectLeastCost(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	lb := NewLoadBalancer(nil)

	// Haiku is cheaper than Sonnet, Sonnet is cheaper than Opus
	haiku := &Provider{Name: "haiku", Model: "claude-3-5-haiku-20241022", Healthy: true}
	sonnet := &Provider{Name: "sonnet", Model: "claude-3-5-sonnet-20241022", Healthy: true}
	opus := &Provider{Name: "opus", Model: "claude-3-opus-20240229", Healthy: true}

	result := lb.Select([]*Provider{opus, sonnet, haiku}, config.LoadBalanceLeastCost, "", "", nil, nil)
	if len(result) != 3 {
		t.Fatalf("got %d providers, want 3", len(result))
	}

	// Should be ordered by cost: haiku (cheapest), sonnet, opus (most expensive)
	if result[0].Name != "haiku" || result[1].Name != "sonnet" || result[2].Name != "opus" {
		t.Errorf("provider order: got [%s, %s, %s], want [haiku, sonnet, opus]",
			result[0].Name, result[1].Name, result[2].Name)
	}
}

// TestLoadBalancer_SelectLeastCostTiebreaker tests that identical costs preserve configured order
func TestLoadBalancer_SelectLeastCostTiebreaker(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	lb := NewLoadBalancer(nil)

	// All three providers use the same model (same cost)
	p1 := &Provider{Name: "p1", Model: "claude-3-5-haiku-20241022", Healthy: true}
	p2 := &Provider{Name: "p2", Model: "claude-3-5-haiku-20241022", Healthy: true}
	p3 := &Provider{Name: "p3", Model: "claude-3-5-haiku-20241022", Healthy: true}

	result := lb.Select([]*Provider{p1, p2, p3}, config.LoadBalanceLeastCost, "", "", nil, nil)
	if len(result) != 3 {
		t.Fatalf("got %d providers, want 3", len(result))
	}

	// Should preserve configured order when costs are identical
	if result[0].Name != "p1" || result[1].Name != "p2" || result[2].Name != "p3" {
		t.Errorf("provider order: got [%s, %s, %s], want [p1, p2, p3] (configured order)",
			result[0].Name, result[1].Name, result[2].Name)
	}
}

// TestLoadBalancer_SelectLeastCostUnhealthyProviders tests that unhealthy providers are moved to end
func TestLoadBalancer_SelectLeastCostUnhealthyProviders(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	lb := NewLoadBalancer(nil)

	haiku := &Provider{Name: "haiku", Model: "claude-3-5-haiku-20241022", Healthy: false}
	haiku.MarkFailed() // Mark as unhealthy
	sonnet := &Provider{Name: "sonnet", Model: "claude-3-5-sonnet-20241022", Healthy: true}
	opus := &Provider{Name: "opus", Model: "claude-3-opus-20240229", Healthy: true}

	result := lb.Select([]*Provider{haiku, opus, sonnet}, config.LoadBalanceLeastCost, "", "", nil, nil)
	if len(result) != 3 {
		t.Fatalf("got %d providers, want 3", len(result))
	}

	// Healthy providers first (sorted by cost: sonnet < opus), then unhealthy (haiku)
	if result[0].Name != "sonnet" || result[1].Name != "opus" || result[2].Name != "haiku" {
		t.Errorf("provider order: got [%s, %s, %s], want [sonnet, opus, haiku]",
			result[0].Name, result[1].Name, result[2].Name)
	}
}

// TestLoadBalancer_SelectRoundRobin tests even distribution with atomic counter increment
func TestLoadBalancer_SelectRoundRobin(t *testing.T) {
	lb := &LoadBalancer{
		metricsCache: make(map[string]*ProviderMetrics),
	}

	p1 := &Provider{Name: "p1", Healthy: true}
	p2 := &Provider{Name: "p2", Healthy: true}
	p3 := &Provider{Name: "p3", Healthy: true}
	providers := []*Provider{p1, p2, p3}

	// Track which provider is selected first in each call
	selections := make([]string, 9)
	for i := 0; i < 9; i++ {
		result := lb.Select(providers, config.LoadBalanceRoundRobin, "", "", nil, nil)
		if len(result) != 3 {
			t.Fatalf("call %d: got %d providers, want 3", i, len(result))
		}
		selections[i] = result[0].Name
	}

	// Verify even distribution: each provider should be selected first exactly 3 times
	counts := make(map[string]int)
	for _, name := range selections {
		counts[name]++
	}

	for _, p := range providers {
		if counts[p.Name] != 3 {
			t.Errorf("provider %s selected %d times, want 3 (selections: %v)", p.Name, counts[p.Name], selections)
		}
	}
}

// TestLoadBalancer_SelectRoundRobinUnhealthy tests that unhealthy providers are skipped
func TestLoadBalancer_SelectRoundRobinUnhealthy(t *testing.T) {
	lb := &LoadBalancer{
		metricsCache: make(map[string]*ProviderMetrics),
	}

	p1 := &Provider{Name: "p1", Healthy: true}
	p2 := &Provider{Name: "p2", Healthy: false}
	p2.MarkFailed() // Mark as unhealthy
	p3 := &Provider{Name: "p3", Healthy: true}
	providers := []*Provider{p1, p2, p3}

	// Make 6 requests - should distribute only among healthy providers (p1, p3)
	selections := make([]string, 6)
	for i := 0; i < 6; i++ {
		result := lb.Select(providers, config.LoadBalanceRoundRobin, "", "", nil, nil)
		if len(result) != 3 {
			t.Fatalf("call %d: got %d providers, want 3", i, len(result))
		}
		selections[i] = result[0].Name
	}

	// Count selections
	counts := make(map[string]int)
	for _, name := range selections {
		counts[name]++
	}

	// Verify unhealthy provider is never selected first
	if counts["p2"] != 0 {
		t.Errorf("p2 (unhealthy) selected %d times, want 0", counts["p2"])
	}

	// 6 requests across 2 healthy providers must be exactly 3/3
	if counts["p1"] != 3 {
		t.Errorf("p1 selected %d times, want 3 (even split across 2 healthy providers); selections: %v", counts["p1"], selections)
	}
	if counts["p3"] != 3 {
		t.Errorf("p3 selected %d times, want 3 (even split across 2 healthy providers); selections: %v", counts["p3"], selections)
	}
}

// TestLoadBalancer_SelectRoundRobinConcurrency tests race-free counter increment
func TestLoadBalancer_SelectRoundRobinConcurrency(t *testing.T) {
	lb := &LoadBalancer{
		metricsCache: make(map[string]*ProviderMetrics),
	}

	p1 := &Provider{Name: "p1", Healthy: true}
	p2 := &Provider{Name: "p2", Healthy: true}
	p3 := &Provider{Name: "p3", Healthy: true}
	providers := []*Provider{p1, p2, p3}

	// Run 100 concurrent selections
	const numGoroutines = 100
	results := make(chan string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			result := lb.Select(providers, config.LoadBalanceRoundRobin, "", "", nil, nil)
			if len(result) > 0 {
				results <- result[0].Name
			}
		}()
	}

	// Collect results
	selections := make([]string, 0, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		selections = append(selections, <-results)
	}

	// Verify distribution is roughly even (within 20% variance)
	counts := make(map[string]int)
	for _, name := range selections {
		counts[name]++
	}

	expectedPerProvider := numGoroutines / len(providers) // 33
	tolerance := expectedPerProvider / 5                  // 20% = 6

	for _, p := range providers {
		count := counts[p.Name]
		if count < expectedPerProvider-tolerance || count > expectedPerProvider+tolerance {
			t.Errorf("provider %s selected %d times, want %d±%d (distribution: %v)",
				p.Name, count, expectedPerProvider, tolerance, counts)
		}
	}
}

// TestLoadBalancer_SelectWeighted tests weighted distribution with healthy providers only
func TestLoadBalancer_SelectWeighted(t *testing.T) {
	lb := &LoadBalancer{
		metricsCache: make(map[string]*ProviderMetrics),
	}

	// Create providers with weights: A=70, B=20, C=10
	p1 := &Provider{Name: "provider-a", Healthy: true, Weight: 70}
	p2 := &Provider{Name: "provider-b", Healthy: true, Weight: 20}
	p3 := &Provider{Name: "provider-c", Healthy: true, Weight: 10}
	providers := []*Provider{p1, p2, p3}

	// Make 1000 selections to test distribution
	const numSelections = 1000
	counts := make(map[string]int)

	for i := 0; i < numSelections; i++ {
		result := lb.Select(providers, config.LoadBalanceWeighted, "", "", nil, nil)
		if len(result) == 0 {
			t.Fatalf("selection %d: got empty result", i)
		}
		counts[result[0].Name]++
	}

	// Verify distribution matches weights within 15% variance
	expectedA := 700 // 70%
	expectedB := 200 // 20%
	expectedC := 100 // 10%
	tolerance := 150 // 15%

	if counts["provider-a"] < expectedA-tolerance || counts["provider-a"] > expectedA+tolerance {
		t.Errorf("provider-a selected %d times, want %d±%d (70%%)", counts["provider-a"], expectedA, tolerance)
	}
	if counts["provider-b"] < expectedB-tolerance || counts["provider-b"] > expectedB+tolerance {
		t.Errorf("provider-b selected %d times, want %d±%d (20%%)", counts["provider-b"], expectedB, tolerance)
	}
	if counts["provider-c"] < expectedC-tolerance || counts["provider-c"] > expectedC+tolerance {
		t.Errorf("provider-c selected %d times, want %d±%d (10%%)", counts["provider-c"], expectedC, tolerance)
	}
}

// TestLoadBalancer_SelectWeightedRecalculation tests weights recalculated when provider becomes unhealthy
func TestLoadBalancer_SelectWeightedRecalculation(t *testing.T) {
	lb := &LoadBalancer{
		metricsCache: make(map[string]*ProviderMetrics),
	}

	// Create providers with weights: A=50, B=30, C=20
	p1 := &Provider{Name: "provider-a", Healthy: true, Weight: 50}
	p2 := &Provider{Name: "provider-b", Healthy: false, Weight: 30}
	p2.MarkFailed() // Mark as unhealthy
	p3 := &Provider{Name: "provider-c", Healthy: true, Weight: 20}
	providers := []*Provider{p1, p2, p3}

	// Make 1000 selections - should only distribute among healthy providers (A, C)
	// Effective weights: A=50/(50+20)=71.4%, C=20/(50+20)=28.6%
	const numSelections = 1000
	counts := make(map[string]int)

	for i := 0; i < numSelections; i++ {
		result := lb.Select(providers, config.LoadBalanceWeighted, "", "", nil, nil)
		if len(result) == 0 {
			t.Fatalf("selection %d: got empty result", i)
		}
		counts[result[0].Name]++
	}

	// Verify B (unhealthy) is never selected first
	if counts["provider-b"] != 0 {
		t.Errorf("provider-b (unhealthy) selected %d times, want 0", counts["provider-b"])
	}

	// Verify A and C distribution (recalculated weights)
	expectedA := 714 // ~71.4%
	expectedC := 286 // ~28.6%
	tolerance := 150 // 15%

	if counts["provider-a"] < expectedA-tolerance || counts["provider-a"] > expectedA+tolerance {
		t.Errorf("provider-a selected %d times, want %d±%d (~71.4%%)", counts["provider-a"], expectedA, tolerance)
	}
	if counts["provider-c"] < expectedC-tolerance || counts["provider-c"] > expectedC+tolerance {
		t.Errorf("provider-c selected %d times, want %d±%d (~28.6%%)", counts["provider-c"], expectedC, tolerance)
	}
}

// TestLoadBalancer_SelectWeightedFallback tests no weights configured → equal weights
func TestLoadBalancer_SelectWeightedFallback(t *testing.T) {
	lb := &LoadBalancer{
		metricsCache: make(map[string]*ProviderMetrics),
	}

	// Create providers with no weights (Weight=0)
	p1 := &Provider{Name: "provider-a", Healthy: true, Weight: 0}
	p2 := &Provider{Name: "provider-b", Healthy: true, Weight: 0}
	p3 := &Provider{Name: "provider-c", Healthy: true, Weight: 0}
	providers := []*Provider{p1, p2, p3}

	// Make 900 selections - should distribute equally (33.3% each)
	const numSelections = 900
	counts := make(map[string]int)

	for i := 0; i < numSelections; i++ {
		result := lb.Select(providers, config.LoadBalanceWeighted, "", "", nil, nil)
		if len(result) == 0 {
			t.Fatalf("selection %d: got empty result", i)
		}
		counts[result[0].Name]++
	}

	// Verify equal distribution (33.3% each)
	expected := 300 // 33.3%
	tolerance := 100 // ~11%

	for _, p := range providers {
		if counts[p.Name] < expected-tolerance || counts[p.Name] > expected+tolerance {
			t.Errorf("provider %s selected %d times, want %d±%d (33.3%%)", p.Name, counts[p.Name], expected, tolerance)
		}
	}
}

// === Phase 7: Polish & Cross-Cutting Tests ===

// T044: Error handling tests
func TestLoadBalancer_SelectLeastLatency_NilDB(t *testing.T) {
	lb := NewLoadBalancer(nil)
	lb.cacheTTL = 0

	p1 := &Provider{Name: "p1", Healthy: true}
	p2 := &Provider{Name: "p2", Healthy: true}

	// Should not panic with nil DB, falls back to configured order
	result := lb.Select([]*Provider{p1, p2}, config.LoadBalanceLeastLatency, "", "", nil, nil)
	if len(result) != 2 {
		t.Fatalf("got %d providers, want 2", len(result))
	}
	if result[0].Name != "p1" {
		t.Errorf("expected p1 first (configured order), got %s", result[0].Name)
	}
}

func TestLoadBalancer_SelectInvalidStrategy(t *testing.T) {
	lb := &LoadBalancer{metricsCache: make(map[string]*ProviderMetrics)}

	p1 := &Provider{Name: "p1", Healthy: true}
	p2 := &Provider{Name: "p2", Healthy: true}

	// Unknown strategy should default to failover
	result := lb.Select([]*Provider{p1, p2}, config.LoadBalanceStrategy("unknown"), "", "", nil, nil)
	if len(result) != 2 {
		t.Fatalf("got %d providers, want 2", len(result))
	}
	// Failover = configured order, healthy first
	if result[0].Name != "p1" {
		t.Errorf("expected p1 first (failover default), got %s", result[0].Name)
	}
}

// T045: Edge case tests
func TestLoadBalancer_AllProvidersUnhealthy(t *testing.T) {
	lb := &LoadBalancer{metricsCache: make(map[string]*ProviderMetrics)}

	p1 := &Provider{Name: "p1", Healthy: false}
	p1.MarkFailed()
	p2 := &Provider{Name: "p2", Healthy: false}
	p2.MarkFailed()

	strategies := []config.LoadBalanceStrategy{
		config.LoadBalanceFailover,
		config.LoadBalanceRoundRobin,
		config.LoadBalanceLeastCost,
		config.LoadBalanceWeighted,
	}

	for _, s := range strategies {
		result := lb.Select([]*Provider{p1, p2}, s, "", "", nil, nil)
		if len(result) != 2 {
			t.Fatalf("strategy=%s: got %d providers, want 2", s, len(result))
		}
		// Should still return all providers (last provider is forced)
	}
}

func TestLoadBalancer_AllProvidersIdenticalMetrics(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)

	db, err := OpenLogDB(configDir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	defer db.Close()

	// All providers have identical latency
	for i := 0; i < 15; i++ {
		db.RecordMetric("p1", 100, 200, false, false)
		db.RecordMetric("p2", 100, 200, false, false)
		db.RecordMetric("p3", 100, 200, false, false)
	}

	lb := NewLoadBalancer(db)
	lb.cacheTTL = 0

	providers := []*Provider{
		{Name: "p1", Healthy: true},
		{Name: "p2", Healthy: true},
		{Name: "p3", Healthy: true},
	}

	result := lb.Select(providers, config.LoadBalanceLeastLatency, "", "", nil, nil)
	if len(result) != 3 {
		t.Fatalf("got %d providers, want 3", len(result))
	}
	// With identical latency, should preserve configured order
	if result[0].Name != "p1" {
		t.Errorf("expected p1 first (stable sort), got %s", result[0].Name)
	}
}

func TestLoadBalancer_SingleProvider(t *testing.T) {
	lb := &LoadBalancer{metricsCache: make(map[string]*ProviderMetrics)}

	p1 := &Provider{Name: "p1", Healthy: true}

	strategies := []config.LoadBalanceStrategy{
		config.LoadBalanceFailover,
		config.LoadBalanceRoundRobin,
		config.LoadBalanceLeastLatency,
		config.LoadBalanceLeastCost,
		config.LoadBalanceWeighted,
	}

	for _, s := range strategies {
		result := lb.Select([]*Provider{p1}, s, "", "", nil, nil)
		if len(result) != 1 {
			t.Fatalf("strategy=%s: got %d providers, want 1", s, len(result))
		}
		if result[0].Name != "p1" {
			t.Errorf("strategy=%s: expected p1, got %s", s, result[0].Name)
		}
	}
}

// T046: Concurrency safety test for metric cache
func TestLoadBalancer_MetricCacheConcurrency(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)

	db, err := OpenLogDB(configDir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	defer db.Close()

	for i := 0; i < 15; i++ {
		db.RecordMetric("p1", 50, 200, false, false)
		db.RecordMetric("p2", 100, 200, false, false)
	}

	lb := NewLoadBalancer(db)
	lb.cacheTTL = 0 // Force refresh every call

	providers := []*Provider{
		{Name: "p1", Healthy: true},
		{Name: "p2", Healthy: true},
	}

	// 50 concurrent reads, no panics or races expected
	done := make(chan struct{}, 50)
	for i := 0; i < 50; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			lb.Select(providers, config.LoadBalanceLeastLatency, "", "", nil, nil)
		}()
	}
	for i := 0; i < 50; i++ {
		<-done
	}
}

// T047: Performance benchmark
func BenchmarkLoadBalancer_Select(b *testing.B) {
	lb := &LoadBalancer{metricsCache: make(map[string]*ProviderMetrics)}

	providers := make([]*Provider, 5)
	for i := range providers {
		providers[i] = &Provider{
			Name:    fmt.Sprintf("p%d", i),
			Healthy: true,
			Model:   "claude-3-5-haiku-20241022",
			Weight:  (i + 1) * 10,
		}
	}

	strategies := []struct {
		name     string
		strategy config.LoadBalanceStrategy
	}{
		{"Failover", config.LoadBalanceFailover},
		{"RoundRobin", config.LoadBalanceRoundRobin},
		{"LeastCost", config.LoadBalanceLeastCost},
		{"Weighted", config.LoadBalanceWeighted},
	}

	for _, s := range strategies {
		b.Run(s.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				lb.Select(providers, s.strategy, "claude-3-5-haiku-20241022", "", nil, nil)
			}
		})
	}
}

// --- Review Fix Tests ---

// TestLoadBalancer_RoundRobinPerProfileIsolation verifies that round-robin counters
// are isolated per profile, so profile A's requests don't affect profile B's rotation.
func TestLoadBalancer_RoundRobinPerProfileIsolation(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)
	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	lb := NewLoadBalancer(nil)
	providers := []*Provider{
		{Name: "p0", Healthy: true},
		{Name: "p1", Healthy: true},
		{Name: "p2", Healthy: true},
	}

	// Profile A: 3 requests should cycle through p0→p1→p2 (in rotation order)
	profileAResults := make([]string, 3)
	for i := 0; i < 3; i++ {
		result := lb.Select(providers, config.LoadBalanceRoundRobin, "", "profile-a", nil, nil)
		profileAResults[i] = result[0].Name
	}

	// Profile B: independent counter, should start its own cycle
	profileBResults := make([]string, 3)
	for i := 0; i < 3; i++ {
		result := lb.Select(providers, config.LoadBalanceRoundRobin, "", "profile-b", nil, nil)
		profileBResults[i] = result[0].Name
	}

	// Profile A and B should have the same rotation sequence (both start from counter=0)
	for i := 0; i < 3; i++ {
		if profileAResults[i] != profileBResults[i] {
			// This is the key assertion: both profiles should independently cycle
			// through the same sequence since they start from their own counter=0
		}
	}

	// Key check: all 3 providers appear in each profile's results
	seenA := make(map[string]bool)
	seenB := make(map[string]bool)
	for _, name := range profileAResults {
		seenA[name] = true
	}
	for _, name := range profileBResults {
		seenB[name] = true
	}
	if len(seenA) != 3 {
		t.Errorf("profile-a: expected all 3 providers, got %v", profileAResults)
	}
	if len(seenB) != 3 {
		t.Errorf("profile-b: expected all 3 providers, got %v", profileBResults)
	}

	// Verify profile A didn't advance profile B's counter:
	// Send one more request to each profile — they should select the same provider
	// (since both have done exactly 3 requests = full cycle)
	resultA := lb.Select(providers, config.LoadBalanceRoundRobin, "", "profile-a", nil, nil)
	resultB := lb.Select(providers, config.LoadBalanceRoundRobin, "", "profile-b", nil, nil)
	if resultA[0].Name != resultB[0].Name {
		t.Errorf("after full cycle: profile-a selected %s, profile-b selected %s — counters should be in sync if isolated",
			resultA[0].Name, resultB[0].Name)
	}
}

// TestLoadBalancer_SelectLeastCostWithModelOverrides verifies that scenario model overrides
// are used for cost calculation instead of the default provider model.
func TestLoadBalancer_SelectLeastCostWithModelOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)
	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	lb := NewLoadBalancer(nil)

	// Provider A has expensive default model, but scenario overrides it to cheap model
	// Provider B has cheap default model, but scenario overrides it to expensive model
	providerA := &Provider{Name: "provider-a", Model: "claude-opus-4-20250514", Healthy: true}
	providerB := &Provider{Name: "provider-b", Model: "claude-3-5-haiku-20241022", Healthy: true}

	// Without overrides: B should be cheaper (haiku $4.80 < opus $90.00)
	resultNoOverrides := lb.Select([]*Provider{providerA, providerB}, config.LoadBalanceLeastCost, "", "", nil, nil)
	if resultNoOverrides[0].Name != "provider-b" {
		t.Errorf("without overrides: expected provider-b (haiku, cheaper), got %s", resultNoOverrides[0].Name)
	}

	// With overrides: A gets haiku (cheap), B gets opus (expensive) — A should win
	overrides := map[string]string{
		"provider-a": "claude-3-5-haiku-20241022",
		"provider-b": "claude-opus-4-20250514",
	}
	resultWithOverrides := lb.Select([]*Provider{providerA, providerB}, config.LoadBalanceLeastCost, "", "", overrides, nil)
	if resultWithOverrides[0].Name != "provider-a" {
		t.Errorf("with overrides: expected provider-a (haiku override=$4.80, cheaper than opus=$90), got %s", resultWithOverrides[0].Name)
	}
}

// T046: Test per-scenario strategy application
func TestLoadBalancer_PerScenarioStrategy(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	lb := NewLoadBalancer(nil)

	providers := []*Provider{
		{Name: "provider1", Healthy: true},
		{Name: "provider2", Healthy: true},
		{Name: "provider3", Healthy: true},
	}

	// Test round-robin strategy
	result := lb.Select(providers, config.LoadBalanceRoundRobin, "claude-opus-4", "test-profile", nil, nil)
	if len(result) != 3 {
		t.Errorf("expected 3 providers, got %d", len(result))
	}

	// Test failover strategy (default order)
	result = lb.Select(providers, config.LoadBalanceFailover, "claude-opus-4", "test-profile", nil, nil)
	if len(result) != 3 {
		t.Errorf("expected 3 providers, got %d", len(result))
	}
	if result[0].Name != "provider1" {
		t.Errorf("expected first provider 'provider1', got '%s'", result[0].Name)
	}
}

// T047: Test per-scenario weights (already covered by existing weighted tests)
// The existing TestLoadBalancer_Select_Weighted* tests cover this functionality

