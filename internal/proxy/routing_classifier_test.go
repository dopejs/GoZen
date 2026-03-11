package proxy

import (
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

// T027: Test builtin classifier fallback when no middleware decision
func TestBuiltinClassifier_BasicScenarios(t *testing.T) {
	classifier := &BuiltinClassifier{Threshold: 32000}

	tests := []struct {
		name           string
		features       *RequestFeatures
		body           map[string]interface{}
		expectedScenario string
		minConfidence  float64
	}{
		{
			name: "image detection",
			features: &RequestFeatures{
				Model:        "claude-opus-4",
				HasImage:     true,
				TotalTokens:  100,
				MessageCount: 1,
			},
			expectedScenario: string(config.ScenarioImage),
			minConfidence:    0.9,
		},
		{
			name: "long context detection",
			features: &RequestFeatures{
				Model:        "claude-opus-4",
				TotalTokens:  50000,
				MessageCount: 10,
			},
			expectedScenario: string(config.ScenarioLongContext),
			minConfidence:    0.9,
		},
		{
			name: "haiku model detection",
			features: &RequestFeatures{
				Model:        "claude-haiku-4",
				TotalTokens:  100,
				MessageCount: 1,
			},
			expectedScenario: string(config.ScenarioBackground),
			minConfidence:    0.9,
		},
		{
			name: "code scenario default",
			features: &RequestFeatures{
				Model:        "claude-opus-4",
				TotalTokens:  100,
				MessageCount: 1,
			},
			expectedScenario: string(config.ScenarioCode),
			minConfidence:    0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.Classify(nil, tt.features, nil, "", tt.body)

			if result.Scenario != tt.expectedScenario {
				t.Errorf("expected scenario '%s', got '%s'", tt.expectedScenario, result.Scenario)
			}
			if result.Confidence < tt.minConfidence {
				t.Errorf("expected confidence >= %f, got %f", tt.minConfidence, result.Confidence)
			}
			if result.Source != "builtin:classifier" {
				t.Errorf("expected source 'builtin:classifier', got '%s'", result.Source)
			}
		})
	}
}

// T027: Test thinking mode detection
func TestBuiltinClassifier_ThinkingMode(t *testing.T) {
	classifier := &BuiltinClassifier{Threshold: 32000}

	body := map[string]interface{}{
		"model":    "claude-opus-4",
		"thinking": map[string]interface{}{"type": "enabled"},
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "test"},
		},
	}

	features := &RequestFeatures{
		Model:        "claude-opus-4",
		HasThinking:  true,
		TotalTokens:  100,
		MessageCount: 1,
	}

	result := classifier.Classify(nil, features, nil, "", body)

	if result.Scenario != string(config.ScenarioThink) {
		t.Errorf("expected scenario 'think', got '%s'", result.Scenario)
	}
	if result.Confidence < 0.9 {
		t.Errorf("expected confidence >= 0.9, got %f", result.Confidence)
	}
}

// T027: Test web search tool detection
func TestBuiltinClassifier_WebSearchTool(t *testing.T) {
	classifier := &BuiltinClassifier{Threshold: 32000}

	body := map[string]interface{}{
		"model": "claude-opus-4",
		"tools": []interface{}{
			map[string]interface{}{"type": "web_search_google"},
		},
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "test"},
		},
	}

	features := &RequestFeatures{
		Model:        "claude-opus-4",
		HasTools:     true,
		HasWebSearch: true,
		TotalTokens:  100,
		MessageCount: 1,
	}

	result := classifier.Classify(nil, features, nil, "", body)

	if result.Scenario != string(config.ScenarioWebSearch) {
		t.Errorf("expected scenario 'webSearch', got '%s'", result.Scenario)
	}
	if result.Confidence < 0.9 {
		t.Errorf("expected confidence >= 0.9, got %f", result.Confidence)
	}
}

// T028: Test routing hints integration
func TestBuiltinClassifier_RoutingHints(t *testing.T) {
	classifier := &BuiltinClassifier{Threshold: 32000}

	hints := &RoutingHints{
		ScenarioCandidates: []string{"custom-plan"},
		Confidence: map[string]float64{
			"customPlan": 0.85, // High confidence hint
		},
	}

	features := &RequestFeatures{
		Model:        "claude-opus-4",
		TotalTokens:  100,
		MessageCount: 1,
	}

	result := classifier.Classify(nil, features, hints, "", nil)

	// High confidence hint should be used
	if result.Scenario != "customPlan" {
		t.Errorf("expected scenario 'customPlan' from hints, got '%s'", result.Scenario)
	}
	if result.Source != "builtin:classifier+hints" {
		t.Errorf("expected source 'builtin:classifier+hints', got '%s'", result.Source)
	}
	if result.Confidence < 0.8 {
		t.Errorf("expected confidence >= 0.8, got %f", result.Confidence)
	}
}

// T028: Test low confidence hints are ignored
func TestBuiltinClassifier_LowConfidenceHintsIgnored(t *testing.T) {
	classifier := &BuiltinClassifier{Threshold: 32000}

	hints := &RoutingHints{
		ScenarioCandidates: []string{"custom-plan"},
		Confidence: map[string]float64{
			"customPlan": 0.5, // Low confidence hint
		},
	}

	features := &RequestFeatures{
		Model:        "claude-haiku-4",
		TotalTokens:  100,
		MessageCount: 1,
	}

	result := classifier.Classify(nil, features, hints, "", nil)

	// Low confidence hint should be ignored, use builtin classification
	if result.Scenario != string(config.ScenarioBackground) {
		t.Errorf("expected scenario 'background' (haiku), got '%s'", result.Scenario)
	}
	if result.Source != "builtin:classifier" {
		t.Errorf("expected source 'builtin:classifier', got '%s'", result.Source)
	}
}

// T031: Test confidence scoring
func TestBuiltinClassifier_ConfidenceScoring(t *testing.T) {
	tests := []struct {
		name          string
		scenario      string
		minConfidence float64
		maxConfidence float64
	}{
		{"webSearch high confidence", string(config.ScenarioWebSearch), 0.9, 1.0},
		{"think high confidence", string(config.ScenarioThink), 0.9, 1.0},
		{"image high confidence", string(config.ScenarioImage), 0.9, 1.0},
		{"longContext high confidence", string(config.ScenarioLongContext), 0.9, 1.0},
		{"background high confidence", string(config.ScenarioBackground), 0.9, 1.0},
		{"code medium confidence", string(config.ScenarioCode), 0.5, 0.6},
		{"default low confidence", string(config.ScenarioDefault), 0.3, 0.4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := confidenceForScenario(tt.scenario)
			if confidence < tt.minConfidence || confidence > tt.maxConfidence {
				t.Errorf("expected confidence in range [%f, %f], got %f",
					tt.minConfidence, tt.maxConfidence, confidence)
			}
		})
	}
}

// T038: Test scenario key normalization
func TestNormalizeScenarioKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"web-search", "webSearch"},
		{"web_search", "webSearch"},
		{"webSearch", "webSearch"},
		{"long-context", "longContext"},
		{"long_context", "longContext"},
		{"longContext", "longContext"},
		{"think", "think"},
		{"custom-plan", "customPlan"},
		{"custom_plan", "customPlan"},
		{"customPlan", "customPlan"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeScenarioKey(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeScenarioKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// T049: Test per-scenario threshold override
func TestBuiltinClassifier_PerScenarioThreshold(t *testing.T) {
	tests := []struct {
		name              string
		threshold         int
		tokenCount        int
		expectedScenario  string
		expectedConfidence float64
	}{
		{
			name:              "below default threshold",
			threshold:         32000,
			tokenCount:        20000,
			expectedScenario:  string(config.ScenarioCode),
			expectedConfidence: 0.5,
		},
		{
			name:              "above default threshold",
			threshold:         32000,
			tokenCount:        50000,
			expectedScenario:  string(config.ScenarioLongContext),
			expectedConfidence: 0.9,
		},
		{
			name:              "below custom threshold",
			threshold:         100000,
			tokenCount:        50000,
			expectedScenario:  string(config.ScenarioCode),
			expectedConfidence: 0.5,
		},
		{
			name:              "above custom threshold",
			threshold:         10000,
			tokenCount:        20000,
			expectedScenario:  string(config.ScenarioLongContext),
			expectedConfidence: 0.9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifier := &BuiltinClassifier{Threshold: tt.threshold}

			features := &RequestFeatures{
				Model:        "claude-opus-4",
				TotalTokens:  tt.tokenCount,
				MessageCount: 1,
			}

			decision := classifier.Classify(nil, features, nil, "", nil)

			if decision.Scenario != tt.expectedScenario {
				t.Errorf("expected scenario %q, got %q", tt.expectedScenario, decision.Scenario)
			}
			if decision.Confidence < tt.expectedConfidence-0.1 || decision.Confidence > tt.expectedConfidence+0.1 {
				t.Errorf("expected confidence ~%.1f, got %.2f", tt.expectedConfidence, decision.Confidence)
			}
		})
	}
}

// Test 80% threshold rule for long context without session history (FR-002)
func TestBuiltinClassifier_LongContextThresholdWithoutSession(t *testing.T) {
	classifier := &BuiltinClassifier{Threshold: 32000}

	tests := []struct {
		name             string
		tokenCount       int
		sessionID        string
		expectedScenario string
		reason           string
	}{
		{
			name:             "below 80% threshold without session",
			tokenCount:       25000, // 25000 < 25600 (0.8 × 32000)
			sessionID:        "",
			expectedScenario: string(config.ScenarioCode),
			reason:           "should not trigger longContext",
		},
		{
			name:             "at 80% threshold without session",
			tokenCount:       25600, // exactly 0.8 × 32000
			sessionID:        "",
			expectedScenario: string(config.ScenarioCode),
			reason:           "should not trigger longContext (not exceeding)",
		},
		{
			name:             "above 80% threshold without session",
			tokenCount:       26000, // 26000 > 25600 (0.8 × 32000)
			sessionID:        "",
			expectedScenario: string(config.ScenarioLongContext),
			reason:           "should trigger longContext with 80% threshold",
		},
		{
			name:             "between 80% and 100% threshold without session",
			tokenCount:       30000, // 25600 < 30000 < 32000
			sessionID:        "",
			expectedScenario: string(config.ScenarioLongContext),
			reason:           "should trigger longContext (in 80%-100% range)",
		},
		{
			name:             "above 100% threshold without session",
			tokenCount:       35000, // 35000 > 32000
			sessionID:        "",
			expectedScenario: string(config.ScenarioLongContext),
			reason:           "should trigger longContext (exceeds full threshold)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features := &RequestFeatures{
				Model:        "claude-opus-4",
				TotalTokens:  tt.tokenCount,
				MessageCount: 1,
			}

			decision := classifier.Classify(nil, features, nil, tt.sessionID, nil)

			if decision.Scenario != tt.expectedScenario {
				t.Errorf("%s: got scenario %s, want %s", tt.reason, decision.Scenario, tt.expectedScenario)
			}

			// Verify reason mentions 80% threshold when no session
			if tt.sessionID == "" && tt.expectedScenario == string(config.ScenarioLongContext) {
				if decision.Reason != "token count exceeds 80% threshold (no session history)" {
					t.Errorf("expected reason to mention 80%% threshold, got: %s", decision.Reason)
				}
			}
		})
	}
}

// Test full threshold with session history
func TestBuiltinClassifier_LongContextThresholdWithSession(t *testing.T) {
	classifier := &BuiltinClassifier{Threshold: 32000}

	// Set up session with previous usage that exceeded threshold
	sessionID := "test-session-with-history"
	UpdateSessionUsage(sessionID, &SessionUsage{
		InputTokens:  35000, // Previous request exceeded threshold
		OutputTokens: 1000,
	})
	defer ClearSessionUsage(sessionID)

	tests := []struct {
		name             string
		tokenCount       int
		expectedScenario string
		reason           string
	}{
		{
			name:             "below 80% threshold with session",
			tokenCount:       25000, // 25000 < 25600 (0.8 × 32000)
			expectedScenario: string(config.ScenarioCode),
			reason:           "should not trigger (below full threshold, current request uses full threshold with session)",
		},
		{
			name:             "between 80% and 100% threshold with session",
			tokenCount:       30000, // 25600 < 30000 < 32000
			expectedScenario: string(config.ScenarioCode),
			reason:           "should not trigger via current request check (uses full threshold=32000 with session)",
		},
		{
			name:             "above 100% threshold with session",
			tokenCount:       35000, // 35000 > 32000
			expectedScenario: string(config.ScenarioLongContext),
			reason:           "should trigger (exceeds full threshold)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features := &RequestFeatures{
				Model:        "claude-opus-4",
				TotalTokens:  tt.tokenCount,
				MessageCount: 1,
			}

			decision := classifier.Classify(nil, features, nil, sessionID, nil)

			if decision.Scenario != tt.expectedScenario {
				t.Errorf("%s: got scenario %s, want %s", tt.reason, decision.Scenario, tt.expectedScenario)
			}
		})
	}
}

// Test configurable scenario priority (FR-005)
func TestBuiltinClassifier_ConfigurableScenarioPriority(t *testing.T) {
	// Create a request that matches multiple scenarios
	features := &RequestFeatures{
		Model:        "claude-opus-4",
		HasImage:     true,  // matches image scenario
		HasTools:     true,  // matches code scenario (tools are common in code)
		TotalTokens:  50000, // matches longContext scenario
		MessageCount: 10,
	}

	tests := []struct {
		name             string
		priority         []string
		expectedScenario string
		reason           string
	}{
		{
			name:             "default priority (image > longContext)",
			priority:         nil, // use default
			expectedScenario: string(config.ScenarioImage),
			reason:           "default priority puts image before longContext",
		},
		{
			name:             "custom priority (longContext first)",
			priority:         []string{"longContext", "image", "code"},
			expectedScenario: string(config.ScenarioLongContext),
			reason:           "custom priority puts longContext first",
		},
		{
			name:             "custom priority (code first)",
			priority:         []string{"code", "longContext", "image"},
			expectedScenario: string(config.ScenarioCode),
			reason:           "custom priority puts code first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifier := &BuiltinClassifier{
				Threshold:        32000,
				ScenarioPriority: tt.priority,
			}

			decision := classifier.Classify(nil, features, nil, "", nil)

			if decision.Scenario != tt.expectedScenario {
				t.Errorf("%s: got scenario %s, want %s", tt.reason, decision.Scenario, tt.expectedScenario)
			}
		})
	}
}

// Test scenario priority with single matching scenario
func TestBuiltinClassifier_PrioritySingleMatch(t *testing.T) {
	// Request that only matches one scenario (think)
	features := &RequestFeatures{
		Model:        "claude-opus-4",
		HasThinking:  true, // only matches think scenario
		TotalTokens:  1000,
		MessageCount: 1,
	}

	// Even with custom priority that puts think last, should still match it
	// Note: priority list must include all scenarios that might match
	classifier := &BuiltinClassifier{
		Threshold: 32000,
		ScenarioPriority: []string{
			"code",        // code will also match (has model)
			"longContext", // won't match
			"image",       // won't match
			"think",       // will match
		},
	}

	decision := classifier.Classify(nil, features, nil, "", nil)

	// Should match code first (higher priority than think in this custom order)
	if decision.Scenario != string(config.ScenarioCode) {
		t.Errorf("expected code scenario (higher priority), got %s", decision.Scenario)
	}

	// Now test with think having higher priority
	classifier.ScenarioPriority = []string{"think", "code", "longContext", "image"}
	decision = classifier.Classify(nil, features, nil, "", nil)

	if decision.Scenario != string(config.ScenarioThink) {
		t.Errorf("expected think scenario (higher priority), got %s", decision.Scenario)
	}
}
