package proxy

import "github.com/dopejs/gozen/internal/config"

// ResolveRoutingDecision determines the final routing decision for a request.
// Priority: middleware RoutingDecision > builtin classifier.
// If middleware set a RoutingDecision, it takes precedence regardless of confidence.
// Otherwise, the builtin classifier analyzes the normalized request.
func ResolveRoutingDecision(
	middlewareDecision *RoutingDecision,
	normalized *NormalizedRequest,
	features *RequestFeatures,
	hints *RoutingHints,
	threshold int,
	sessionID string,
	body map[string]interface{},
) *RoutingDecision {
	// If middleware explicitly set a routing decision, use it (highest priority)
	if middlewareDecision != nil && middlewareDecision.Scenario != "" {
		return middlewareDecision
	}

	// Fall back to builtin classifier
	classifier := &BuiltinClassifier{Threshold: threshold}
	return classifier.Classify(normalized, features, hints, sessionID, body)
}

// ResolveRoutePolicy looks up the RoutePolicy for a given scenario in the profile config.
// Returns nil if no route is configured for that scenario.
// Falls back to default providers if scenario not found and fallback is enabled.
func ResolveRoutePolicy(scenario string, routing map[string]*config.RoutePolicy) *config.RoutePolicy {
	if routing == nil {
		return nil
	}

	// Direct lookup with normalized key
	normalized := NormalizeScenarioKey(scenario)
	if route, ok := routing[normalized]; ok {
		return route
	}

	// Try original key as-is (in case config uses non-normalized key)
	if route, ok := routing[scenario]; ok {
		return route
	}

	return nil
}
