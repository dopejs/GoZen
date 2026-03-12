package proxy

import "github.com/dopejs/gozen/internal/config"

// RoutingDecision represents an explicit routing choice made by middleware or builtin classifier.
// It is binding and overrides any default routing behavior.
type RoutingDecision struct {
	// Scenario is the scenario key (e.g., "plan", "code", "think")
	Scenario string

	// Source identifies who made this decision (e.g., "middleware:spec-kit", "builtin:classifier")
	Source string

	// Reason is a human-readable explanation for this routing decision
	Reason string

	// Confidence is a score from 0.0 to 1.0 indicating decision confidence
	Confidence float64

	// ModelHint suggests a specific model override (nil = not set)
	ModelHint *string

	// StrategyOverride overrides the route's load balancing strategy (nil = use route default)
	StrategyOverride *config.LoadBalanceStrategy

	// ThresholdOverride overrides the long-context threshold (nil = use route default)
	ThresholdOverride *int

	// ProviderAllowlist restricts routing to only these providers (empty = no filter)
	ProviderAllowlist []string

	// ProviderDenylist excludes these providers from routing (empty = no filter)
	ProviderDenylist []string

	// Metadata allows custom fields for extensibility
	Metadata map[string]interface{}
}

// RoutingHints provides non-binding suggestions that influence the builtin classifier.
// Unlike RoutingDecision, hints do not force a specific routing choice.
type RoutingHints struct {
	// ScenarioCandidates lists possible scenarios in priority order
	ScenarioCandidates []string

	// Tags are semantic labels (e.g., "high-quality", "fast")
	Tags []string

	// CostClass indicates cost preference: "low", "medium", "high", or empty
	CostClass string

	// CapabilityNeeds lists required capabilities (e.g., "vision", "tools")
	CapabilityNeeds []string

	// Confidence provides per-scenario confidence scores (0.0 to 1.0)
	Confidence map[string]float64

	// Metadata allows custom fields for extensibility
	Metadata map[string]interface{}
}
