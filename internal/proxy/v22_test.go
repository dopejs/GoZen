package proxy

import (
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

func TestUsageTracker_CalculateCost(t *testing.T) {
	tracker := &UsageTracker{
		pricing: map[string]*config.ModelPricing{
			"claude-3-5-sonnet-20241022": {InputPerMillion: 3.0, OutputPerMillion: 15.0},
			"claude-3-haiku-20240307":    {InputPerMillion: 0.25, OutputPerMillion: 1.25},
		},
	}

	tests := []struct {
		name         string
		model        string
		inputTokens  int
		outputTokens int
		wantCost     float64
	}{
		{
			name:         "sonnet model",
			model:        "claude-3-5-sonnet-20241022",
			inputTokens:  1000000,
			outputTokens: 100000,
			wantCost:     4.5, // 3.0 + 1.5
		},
		{
			name:         "haiku model",
			model:        "claude-3-haiku-20240307",
			inputTokens:  1000000,
			outputTokens: 1000000,
			wantCost:     1.5, // 0.25 + 1.25
		},
		{
			name:         "zero tokens",
			model:        "claude-3-5-sonnet-20241022",
			inputTokens:  0,
			outputTokens: 0,
			wantCost:     0,
		},
		{
			name:         "unknown model",
			model:        "unknown-model",
			inputTokens:  1000000,
			outputTokens: 1000000,
			wantCost:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tracker.CalculateCost(tt.model, tt.inputTokens, tt.outputTokens)
			if got != tt.wantCost {
				t.Errorf("CalculateCost() = %v, want %v", got, tt.wantCost)
			}
		})
	}
}

func TestBudgetChecker_checkLimit(t *testing.T) {
	checker := &BudgetChecker{}

	tests := []struct {
		name           string
		spent          float64
		limit          *config.BudgetLimit
		wantWarn       bool
		wantDowngrade  bool
		wantBlock      bool
	}{
		{
			name:  "under 70%",
			spent: 50,
			limit: &config.BudgetLimit{Amount: 100, Action: config.BudgetActionWarn},
		},
		{
			name:     "at 80% warning",
			spent:    80,
			limit:    &config.BudgetLimit{Amount: 100, Action: config.BudgetActionWarn},
			wantWarn: true,
		},
		{
			name:     "over limit with warn action",
			spent:    110,
			limit:    &config.BudgetLimit{Amount: 100, Action: config.BudgetActionWarn},
			wantWarn: true,
		},
		{
			name:          "over limit with downgrade action",
			spent:         110,
			limit:         &config.BudgetLimit{Amount: 100, Action: config.BudgetActionDowngrade},
			wantDowngrade: true,
		},
		{
			name:      "over limit with block action",
			spent:     110,
			limit:     &config.BudgetLimit{Amount: 100, Action: config.BudgetActionBlock},
			wantBlock: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := &BudgetStatus{}
			checker.checkLimit(status, tt.spent, tt.limit, "test")

			if status.ShouldWarn != tt.wantWarn {
				t.Errorf("ShouldWarn = %v, want %v", status.ShouldWarn, tt.wantWarn)
			}
			if status.ShouldDowngrade != tt.wantDowngrade {
				t.Errorf("ShouldDowngrade = %v, want %v", status.ShouldDowngrade, tt.wantDowngrade)
			}
			if status.ShouldBlock != tt.wantBlock {
				t.Errorf("ShouldBlock = %v, want %v", status.ShouldBlock, tt.wantBlock)
			}
		})
	}
}

func TestLoadBalancer_selectFailover(t *testing.T) {
	lb := &LoadBalancer{}

	healthy1 := &Provider{Name: "p1", Healthy: true}
	healthy2 := &Provider{Name: "p2", Healthy: true}
	unhealthy := &Provider{Name: "p3", Healthy: false}
	unhealthy.MarkFailed() // Set backoff to make it unhealthy

	providers := []*Provider{unhealthy, healthy1, healthy2}
	result := lb.selectFailover(providers)

	// Healthy providers should come first
	if len(result) != 3 {
		t.Fatalf("Expected 3 providers, got %d", len(result))
	}

	// Check that healthy providers are before unhealthy ones
	healthyCount := 0
	for i, p := range result {
		if p.IsHealthy() {
			healthyCount++
		} else {
			// Once we hit an unhealthy provider, all remaining should be unhealthy
			for j := i; j < len(result); j++ {
				if result[j].IsHealthy() {
					t.Errorf("Found healthy provider after unhealthy one at index %d", j)
				}
			}
			break
		}
	}

	if healthyCount != 2 {
		t.Errorf("Expected 2 healthy providers first, got %d", healthyCount)
	}
}

func TestHealthChecker_determineStatus(t *testing.T) {
	checker := &HealthChecker{}

	tests := []struct {
		name       string
		status     *ProviderHealthStatus
		wantStatus HealthStatus
	}{
		{
			name:       "no checks",
			status:     &ProviderHealthStatus{CheckCount: 0},
			wantStatus: HealthStatusUnknown,
		},
		{
			name:       "high success rate",
			status:     &ProviderHealthStatus{CheckCount: 100, SuccessRate: 98},
			wantStatus: HealthStatusHealthy,
		},
		{
			name:       "medium success rate",
			status:     &ProviderHealthStatus{CheckCount: 100, SuccessRate: 80},
			wantStatus: HealthStatusDegraded,
		},
		{
			name:       "low success rate",
			status:     &ProviderHealthStatus{CheckCount: 100, SuccessRate: 50},
			wantStatus: HealthStatusUnhealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checker.determineStatus(tt.status)
			if got != tt.wantStatus {
				t.Errorf("determineStatus() = %v, want %v", got, tt.wantStatus)
			}
		})
	}
}
