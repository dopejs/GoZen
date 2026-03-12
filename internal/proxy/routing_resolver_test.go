package proxy

import (
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

// T026: Test middleware decision precedence
func TestResolveRoutingDecision_MiddlewarePrecedence(t *testing.T) {
	// Middleware decision should take precedence over builtin classifier
	middlewareDecision := &RoutingDecision{
		Scenario:   "custom-plan",
		Source:     "middleware:spec-kit",
		Reason:     "explicit plan scenario",
		Confidence: 1.0,
	}

	normalized := &NormalizedRequest{
		Model: "claude-opus-4",
		Messages: []NormalizedMessage{
			{Role: "user", Content: "test", TokenCount: 10},
		},
	}

	features := &RequestFeatures{
		Model:        "claude-opus-4",
		TotalTokens:  10,
		MessageCount: 1,
	}

	result := ResolveRoutingDecision(middlewareDecision, normalized, features, nil, 32000, nil, "", nil)

	if result.Scenario != "custom-plan" {
		t.Errorf("expected scenario 'custom-plan', got '%s'", result.Scenario)
	}
	if result.Source != "middleware:spec-kit" {
		t.Errorf("expected source 'middleware:spec-kit', got '%s'", result.Source)
	}
	if result.Confidence != 1.0 {
		t.Errorf("expected confidence 1.0, got %f", result.Confidence)
	}
}

// T026: Test builtin classifier fallback when no middleware decision
func TestResolveRoutingDecision_BuiltinFallback(t *testing.T) {
	normalized := &NormalizedRequest{
		Model:    "claude-opus-4",
		HasTools: false,
		Messages: []NormalizedMessage{
			{Role: "user", Content: "test", HasImage: false, TokenCount: 10},
		},
	}

	features := &RequestFeatures{
		Model:        "claude-opus-4",
		TotalTokens:  10,
		MessageCount: 1,
		HasImage:     false,
		HasTools:     false,
	}

	// No middleware decision - should use builtin classifier
	result := ResolveRoutingDecision(nil, normalized, features, nil, 32000, nil, "", nil)

	if result.Source != "builtin:classifier" {
		t.Errorf("expected source 'builtin:classifier', got '%s'", result.Source)
	}
	// Should classify as "code" for non-haiku model
	if result.Scenario != string(config.ScenarioCode) {
		t.Errorf("expected scenario 'code', got '%s'", result.Scenario)
	}
}

// T026: Test empty middleware decision is ignored
func TestResolveRoutingDecision_EmptyMiddlewareIgnored(t *testing.T) {
	// Empty scenario in middleware decision should be ignored
	emptyDecision := &RoutingDecision{
		Scenario: "", // Empty scenario
		Source:   "middleware:test",
	}

	normalized := &NormalizedRequest{
		Model: "claude-haiku-4",
		Messages: []NormalizedMessage{
			{Role: "user", Content: "test", TokenCount: 10},
		},
	}

	features := &RequestFeatures{
		Model:        "claude-haiku-4",
		TotalTokens:  10,
		MessageCount: 1,
	}

	result := ResolveRoutingDecision(emptyDecision, normalized, features, nil, 32000, nil, "", nil)

	// Should fall back to builtin classifier
	if result.Source != "builtin:classifier" {
		t.Errorf("expected source 'builtin:classifier', got '%s'", result.Source)
	}
	// Should classify as "background" for haiku model
	if result.Scenario != string(config.ScenarioBackground) {
		t.Errorf("expected scenario 'background', got '%s'", result.Scenario)
	}
}

// T037: Test custom scenario route lookup
func TestResolveRoutePolicy_CustomScenario(t *testing.T) {
	routing := map[string]*config.RoutePolicy{
		"customPlan": {
			Providers: []*config.ProviderRoute{
				{Name: "provider1", Model: "claude-opus-4"},
			},
		},
		"webSearch": {
			Providers: []*config.ProviderRoute{
				{Name: "provider2"},
			},
		},
	}

	// Test exact match
	route := ResolveRoutePolicy("customPlan", routing)
	if route == nil {
		t.Fatal("expected route for 'customPlan', got nil")
	}
	if len(route.Providers) != 1 || route.Providers[0].Name != "provider1" {
		t.Errorf("unexpected route providers: %v", route.Providers)
	}

	// Test normalized key match (kebab-case → camelCase)
	route = ResolveRoutePolicy("custom-plan", routing)
	if route == nil {
		t.Fatal("expected route for 'custom-plan' (normalized to 'customPlan'), got nil")
	}
	if len(route.Providers) != 1 || route.Providers[0].Name != "provider1" {
		t.Errorf("unexpected route providers: %v", route.Providers)
	}
}

// T039: Test unknown scenario fallback
func TestResolveRoutePolicy_UnknownScenario(t *testing.T) {
	routing := map[string]*config.RoutePolicy{
		"think": {
			Providers: []*config.ProviderRoute{
				{Name: "provider1"},
			},
		},
	}

	// Unknown scenario should return nil
	route := ResolveRoutePolicy("unknown-scenario", routing)
	if route != nil {
		t.Errorf("expected nil for unknown scenario, got %v", route)
	}
}

// T039: Test nil routing map
func TestResolveRoutePolicy_NilRouting(t *testing.T) {
	route := ResolveRoutePolicy("any-scenario", nil)
	if route != nil {
		t.Errorf("expected nil for nil routing map, got %v", route)
	}
}
