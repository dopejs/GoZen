package proxy

import (
	"strings"
	"unicode"

	"github.com/dopejs/gozen/internal/config"
)

// BuiltinClassifier classifies requests into scenarios using heuristics.
// It analyzes request features and optionally considers routing hints.
type BuiltinClassifier struct {
	// Threshold is the token count threshold for long-context detection.
	// If 0, uses defaultLongContextThreshold (32000).
	Threshold int

	// ScenarioPriority defines the priority order for scenario selection.
	// If empty, uses default priority: webSearch > think > image > longContext > code > background > default
	ScenarioPriority []string
}

// Classify analyzes the normalized request and returns a routing decision.
// It uses feature-based heuristics with priority ordering:
// webSearch > think > image > longContext > code > background > default
//
// If routing hints provide scenario candidates, they are considered with
// a confidence boost when the builtin analysis is ambiguous.
func (c *BuiltinClassifier) Classify(
	normalized *NormalizedRequest,
	features *RequestFeatures,
	hints *RoutingHints,
	sessionID string,
	body map[string]interface{},
) *RoutingDecision {
	threshold := c.Threshold
	if threshold <= 0 {
		threshold = defaultLongContextThreshold
	}

	// If we have features from normalization, use them
	if features != nil {
		// Check hints first: if hints strongly suggest a scenario and
		// the features don't contradict it, prefer hints
		if hints != nil && len(hints.ScenarioCandidates) > 0 {
			topCandidate := NormalizeScenarioKey(hints.ScenarioCandidates[0])
			hintConfidence := 0.5
			if c, ok := hints.Confidence[topCandidate]; ok {
				hintConfidence = c
			}

			// If hint confidence is high (>= 0.8) and doesn't conflict
			// with a clear feature signal, use the hint
			if hintConfidence >= 0.8 {
				return &RoutingDecision{
					Scenario:   topCandidate,
					Source:     "builtin:classifier+hints",
					Reason:     "routing hint with high confidence",
					Confidence: hintConfidence,
				}
			}
		}

		return c.classifyFromFeatures(features, threshold, sessionID, body)
	}

	// No features available, fall back to raw body analysis
	if body != nil {
		scenario := DetectScenario(body, threshold, sessionID)
		return &RoutingDecision{
			Scenario:   scenario,
			Source:     "builtin:classifier",
			Reason:     "raw body analysis (no normalization available)",
			Confidence: confidenceForScenario(scenario),
		}
	}

	// No information available at all, return default
	return &RoutingDecision{
		Scenario:   string(config.ScenarioDefault),
		Source:     "builtin:classifier",
		Reason:     "no request data available",
		Confidence: 0.3,
	}
}

// classifyFromFeatures uses extracted features to determine the scenario.
// Uses configurable scenario priority order (FR-005).
// Default priority: webSearch > think > image > longContext > code > background > default
func (c *BuiltinClassifier) classifyFromFeatures(
	features *RequestFeatures,
	threshold int,
	sessionID string,
	body map[string]interface{},
) *RoutingDecision {
	// Build a map of scenario → decision for all matching scenarios
	candidates := make(map[string]*RoutingDecision)

	// Check for web search tools
	if features.HasWebSearch {
		candidates[string(config.ScenarioWebSearch)] = &RoutingDecision{
			Scenario:   string(config.ScenarioWebSearch),
			Source:     "builtin:classifier",
			Reason:     "web_search tool detected",
			Confidence: 0.9,
		}
	}

	// Check for thinking/reasoning mode
	if features.HasThinking {
		candidates[string(config.ScenarioThink)] = &RoutingDecision{
			Scenario:   string(config.ScenarioThink),
			Source:     "builtin:classifier",
			Reason:     "thinking mode enabled",
			Confidence: 0.9,
		}
	}

	// Check for image content
	if features.HasImage {
		candidates[string(config.ScenarioImage)] = &RoutingDecision{
			Scenario:   string(config.ScenarioImage),
			Source:     "builtin:classifier",
			Reason:     "image content detected",
			Confidence: 0.9,
		}
	}

	// Check for long context
	// FR-002: Without session history, use 80% of threshold (0.8 × threshold)
	// With session history, use full threshold
	effectiveThreshold := threshold
	hasSessionHistory := sessionID != "" && GetSessionUsage(sessionID) != nil
	if !hasSessionHistory {
		// No session history: use 80% threshold for current request only
		effectiveThreshold = int(float64(threshold) * 0.8)
	}

	if features.TotalTokens > effectiveThreshold {
		reason := "token count exceeds threshold"
		if !hasSessionHistory {
			reason = "token count exceeds 80% threshold (no session history)"
		}
		candidates[string(config.ScenarioLongContext)] = &RoutingDecision{
			Scenario:   string(config.ScenarioLongContext),
			Source:     "builtin:classifier",
			Reason:     reason,
			Confidence: 0.9,
		}
	}

	// Check session history for long context continuation
	// This path handles cases where current request is below threshold but
	// session history indicates we're in a long context conversation
	if sessionID != "" && body != nil && isLongContext(body, threshold, sessionID) {
		if _, exists := candidates[string(config.ScenarioLongContext)]; !exists {
			candidates[string(config.ScenarioLongContext)] = &RoutingDecision{
				Scenario:   string(config.ScenarioLongContext),
				Source:     "builtin:classifier",
				Reason:     "session history indicates long context",
				Confidence: 0.7,
			}
		}
	}

	// Check for background (Haiku model)
	modelLower := strings.ToLower(features.Model)
	if strings.Contains(modelLower, "claude") && strings.Contains(modelLower, "haiku") {
		candidates[string(config.ScenarioBackground)] = &RoutingDecision{
			Scenario:   string(config.ScenarioBackground),
			Source:     "builtin:classifier",
			Reason:     "haiku model detected",
			Confidence: 0.9,
		}
	}

	// Check for code scenario (non-haiku models)
	if features.Model != "" && !strings.Contains(modelLower, "haiku") {
		candidates[string(config.ScenarioCode)] = &RoutingDecision{
			Scenario:   string(config.ScenarioCode),
			Source:     "builtin:classifier",
			Reason:     "non-haiku model (default coding scenario)",
			Confidence: 0.5,
		}
	}

	// If no candidates match, return default
	if len(candidates) == 0 {
		return &RoutingDecision{
			Scenario:   string(config.ScenarioDefault),
			Source:     "builtin:classifier",
			Reason:     "no distinctive features detected",
			Confidence: 0.3,
		}
	}

	// Select scenario based on priority order
	priority := c.ScenarioPriority
	if len(priority) == 0 {
		// Use default priority order
		priority = []string{
			string(config.ScenarioWebSearch),
			string(config.ScenarioThink),
			string(config.ScenarioImage),
			string(config.ScenarioLongContext),
			string(config.ScenarioCode),
			string(config.ScenarioBackground),
			string(config.ScenarioDefault),
		}
	}

	// Find first matching scenario in priority order
	for _, scenario := range priority {
		if decision, ok := candidates[scenario]; ok {
			return decision
		}
	}

	// Fallback: return first candidate (shouldn't happen if priority list is complete)
	for _, decision := range candidates {
		return decision
	}

	return &RoutingDecision{
		Scenario:   string(config.ScenarioDefault),
		Source:     "builtin:classifier",
		Reason:     "no distinctive features detected",
		Confidence: 0.3,
	}
}

// confidenceForScenario returns a confidence score for a given scenario.
func confidenceForScenario(scenario string) float64 {
	switch scenario {
	case string(config.ScenarioWebSearch), string(config.ScenarioThink),
		string(config.ScenarioImage), string(config.ScenarioBackground):
		return 0.9
	case string(config.ScenarioLongContext):
		return 0.9
	case string(config.ScenarioCode):
		return 0.5
	default:
		return 0.3
	}
}

// NormalizeScenarioKey converts scenario keys to canonical camelCase format.
// Supports kebab-case, snake_case, and camelCase inputs.
// Examples:
//   - "web-search" → "webSearch"
//   - "long_context" → "longContext"
//   - "webSearch" → "webSearch" (unchanged)
//   - "think" → "think" (unchanged)
func NormalizeScenarioKey(key string) string {
	if key == "" {
		return ""
	}

	// Check if key contains delimiters (hyphens or underscores)
	hasDelimiters := strings.ContainsAny(key, "-_")
	if !hasDelimiters {
		// No delimiters - return as-is (already camelCase or single word)
		return key
	}

	// Split on hyphens and underscores
	parts := splitOnDelimiters(key)
	if len(parts) == 0 {
		return key
	}

	// First part stays lowercase, rest are title-cased
	result := strings.ToLower(parts[0])
	for i := 1; i < len(parts); i++ {
		if parts[i] != "" {
			result += titleCase(parts[i])
		}
	}

	return result
}

// splitOnDelimiters splits a string on hyphens and underscores
func splitOnDelimiters(s string) []string {
	var parts []string
	var current strings.Builder

	for _, r := range s {
		if r == '-' || r == '_' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// titleCase converts the first character to uppercase, rest to lowercase
func titleCase(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	for i := 1; i < len(runes); i++ {
		runes[i] = unicode.ToLower(runes[i])
	}
	return string(runes)
}
