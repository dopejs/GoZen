package proxy

import (
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

	result := lb.Select(nil, config.LoadBalanceFailover, "")
	if result != nil {
		t.Error("Expected nil for nil input")
	}

	result = lb.Select([]*Provider{}, config.LoadBalanceFailover, "")
	if len(result) != 0 {
		t.Error("Expected empty slice for empty input")
	}
}

func TestLoadBalancer_Select_Single(t *testing.T) {
	lb := &LoadBalancer{}
	provider := &Provider{Name: "test", Healthy: true}

	result := lb.Select([]*Provider{provider}, config.LoadBalanceFailover, "")
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

	result := lb.Select([]*Provider{unhealthy, healthy}, config.LoadBalanceFailover, "")
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
	result1 := lb.Select(providers, config.LoadBalanceRoundRobin, "")
	// Second call should rotate
	result2 := lb.Select(providers, config.LoadBalanceRoundRobin, "")

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

	result := lb.Select([]*Provider{slow, medium, fast}, config.LoadBalanceLeastLatency, "")
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

	result := lb.Select([]*Provider{opus, haiku}, config.LoadBalanceLeastCost, "")
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
	result := lb.Select([]*Provider{p1, p2}, config.LoadBalanceLeastLatency, "")
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

	result := lb.Select([]*Provider{p1, p2}, config.LoadBalanceLeastCost, "")
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
	result := lb.Select([]*Provider{p1, p2}, "unknown-strategy", "")
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
		result := lb.Select(providers, config.LoadBalanceRoundRobin, "")
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
