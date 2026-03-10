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
