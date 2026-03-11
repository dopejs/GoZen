package web

import (
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

// TestRoutePolicyRoundTrip verifies that all RoutePolicy fields are preserved
// through serialization and deserialization.
func TestRoutePolicyRoundTrip(t *testing.T) {
	// Create a RoutePolicy with all fields set
	threshold := 50000
	fallback := true
	strategy := config.LoadBalanceWeighted

	original := &config.RoutePolicy{
		Providers: []*config.ProviderRoute{
			{Name: "provider1", Model: "claude-opus-4"},
			{Name: "provider2", Model: "claude-sonnet-4"},
		},
		Strategy: strategy,
		ProviderWeights: map[string]int{
			"provider1": 70,
			"provider2": 30,
		},
		LongContextThreshold: &threshold,
		FallbackToDefault:    &fallback,
	}

	// Create a profile config with routing
	pc := &config.ProfileConfig{
		Providers: []string{"provider1", "provider2"},
		Routing: map[string]*config.RoutePolicy{
			"customScenario": original,
		},
	}

	// Convert to response (serialize)
	resp := profileConfigToResponse("test-profile", pc)

	// Verify response has routing
	if resp.Routing == nil {
		t.Fatal("Expected routing in response")
	}

	scenarioResp, ok := resp.Routing["customScenario"]
	if !ok {
		t.Fatal("Expected customScenario in routing")
	}

	// Verify all fields are serialized
	if len(scenarioResp.Providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(scenarioResp.Providers))
	}

	if scenarioResp.Strategy == nil || *scenarioResp.Strategy != "weighted" {
		t.Errorf("Expected strategy 'weighted', got %v", scenarioResp.Strategy)
	}

	if len(scenarioResp.ProviderWeights) != 2 {
		t.Errorf("Expected 2 provider weights, got %d", len(scenarioResp.ProviderWeights))
	}

	if scenarioResp.ProviderWeights["provider1"] != 70 {
		t.Errorf("Expected provider1 weight 70, got %d", scenarioResp.ProviderWeights["provider1"])
	}

	if scenarioResp.LongContextThreshold == nil || *scenarioResp.LongContextThreshold != 50000 {
		t.Errorf("Expected threshold 50000, got %v", scenarioResp.LongContextThreshold)
	}

	if scenarioResp.FallbackToDefault == nil || *scenarioResp.FallbackToDefault != true {
		t.Errorf("Expected fallback true, got %v", scenarioResp.FallbackToDefault)
	}

	// Convert back to config (deserialize)
	routing := routingResponseToConfig(resp.Routing)

	if routing == nil {
		t.Fatal("Expected routing after deserialization")
	}

	restored, ok := routing["customScenario"]
	if !ok {
		t.Fatal("Expected customScenario after deserialization")
	}

	// Verify all fields are restored
	if len(restored.Providers) != 2 {
		t.Errorf("Expected 2 providers after restore, got %d", len(restored.Providers))
	}

	if restored.Strategy != config.LoadBalanceWeighted {
		t.Errorf("Expected strategy weighted after restore, got %s", restored.Strategy)
	}

	if len(restored.ProviderWeights) != 2 {
		t.Errorf("Expected 2 provider weights after restore, got %d", len(restored.ProviderWeights))
	}

	if restored.ProviderWeights["provider1"] != 70 {
		t.Errorf("Expected provider1 weight 70 after restore, got %d", restored.ProviderWeights["provider1"])
	}

	if restored.LongContextThreshold == nil || *restored.LongContextThreshold != 50000 {
		t.Errorf("Expected threshold 50000 after restore, got %v", restored.LongContextThreshold)
	}

	if restored.FallbackToDefault == nil || *restored.FallbackToDefault != true {
		t.Errorf("Expected fallback true after restore, got %v", restored.FallbackToDefault)
	}
}
