package proxy

import (
	"encoding/json"
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

// T093: Performance benchmarks for normalization and classification

// BenchmarkNormalizeAnthropicMessages benchmarks Anthropic Messages normalization
func BenchmarkNormalizeAnthropicMessages(b *testing.B) {
	body := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "Hello, how are you?"},
		},
		"max_tokens": 1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NormalizeAnthropicMessages(body)
	}
}

// BenchmarkNormalizeOpenAIChat benchmarks OpenAI Chat normalization
func BenchmarkNormalizeOpenAIChat(b *testing.B) {
	body := map[string]interface{}{
		"model": "gpt-4",
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "Hello, how are you?"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NormalizeOpenAIChat(body)
	}
}

// BenchmarkNormalizeOpenAIResponses benchmarks OpenAI Responses normalization
func BenchmarkNormalizeOpenAIResponses(b *testing.B) {
	body := map[string]interface{}{
		"model": "gpt-4",
		"input": "Hello, how are you?",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NormalizeOpenAIResponses(body)
	}
}

// BenchmarkExtractFeatures benchmarks feature extraction
func BenchmarkExtractFeatures(b *testing.B) {
	normalized := &NormalizedRequest{
		OriginalProtocol: "anthropic_messages",
		Model:            "claude-opus-4",
		Messages: []NormalizedMessage{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ExtractFeatures(normalized)
	}
}

// BenchmarkBuiltinClassifier benchmarks builtin classification
func BenchmarkBuiltinClassifier(b *testing.B) {
	normalized := &NormalizedRequest{
		OriginalProtocol: "anthropic_messages",
		Model:            "claude-opus-4",
		Messages: []NormalizedMessage{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
	}
	features := &RequestFeatures{
		HasImage:      false,
		HasTools:      false,
		IsLongContext: false,
		TotalTokens:   50,
		MessageCount:  1,
	}

	classifier := &BuiltinClassifier{Threshold: 100000}

	var body map[string]interface{}
	json.Unmarshal([]byte(`{"model": "claude-opus-4"}`), &body)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = classifier.Classify(normalized, features, nil, "", body)
	}
}

// BenchmarkResolveRoutingDecision benchmarks decision resolution
func BenchmarkResolveRoutingDecision(b *testing.B) {
	normalized := &NormalizedRequest{
		OriginalProtocol: "anthropic_messages",
		Model:            "claude-opus-4",
	}
	features := &RequestFeatures{
		TotalTokens:  50,
		MessageCount: 1,
	}

	var body map[string]interface{}
	json.Unmarshal([]byte(`{"model": "claude-opus-4"}`), &body)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ResolveRoutingDecision(nil, normalized, features, nil, 100000, "", body)
	}
}

// BenchmarkResolveRoutePolicy benchmarks route policy lookup
func BenchmarkResolveRoutePolicy(b *testing.B) {
	routing := map[string]*config.RoutePolicy{
		"think": {
			Providers: []*config.ProviderRoute{
				{Name: "provider1", Model: "claude-opus-4"},
			},
		},
		"code": {
			Providers: []*config.ProviderRoute{
				{Name: "provider2"},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ResolveRoutePolicy("code", routing)
	}
}

// BenchmarkNormalizeScenarioKey benchmarks scenario key normalization
func BenchmarkNormalizeScenarioKey(b *testing.B) {
	keys := []string{"web-search", "long_context", "customPlan", "think"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, key := range keys {
			_ = NormalizeScenarioKey(key)
		}
	}
}

// BenchmarkFullRoutingPipeline benchmarks the complete routing pipeline
func BenchmarkFullRoutingPipeline(b *testing.B) {
	body := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "Write a function to calculate fibonacci numbers"},
		},
		"max_tokens": 1024,
	}

	routing := map[string]*config.RoutePolicy{
		"code": {
			Providers: []*config.ProviderRoute{
				{Name: "provider1"},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Normalize
		normalized, _ := NormalizeAnthropicMessages(body)
		if normalized == nil {
			continue
		}

		// Extract features
		features := ExtractFeatures(normalized)

		// Classify
		decision := ResolveRoutingDecision(nil, normalized, features, nil, 100000, "", body)

		// Resolve route
		_ = ResolveRoutePolicy(decision.Scenario, routing)
	}
}
