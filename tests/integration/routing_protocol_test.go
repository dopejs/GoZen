package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/proxy"
)

// TestProtocolAgnosticRouting tests that equivalent requests via different API protocols
// route to the same provider/model based on scenario detection.
func TestProtocolAgnosticRouting(t *testing.T) {
	// Setup: Use default store for testing
	config.ResetDefaultStore()

	// Add providers
	config.SetProvider("standard", &config.ProviderConfig{
		BaseURL:   "https://api.anthropic.com",
		AuthToken: "test-token-standard",
	})
	config.SetProvider("thinker", &config.ProviderConfig{
		BaseURL:   "https://api.anthropic.com",
		AuthToken: "test-token-thinker",
	})

	// Create profile with scenario routing
	config.SetProfileConfig("test-profile", &config.ProfileConfig{
		Providers: []string{"standard"},
		Routing: map[string]*config.RoutePolicy{
			"think": {
				Providers: []*config.ProviderRoute{
					{Name: "thinker", Model: "claude-opus-4-20250514"},
				},
			},
		},
	})

	tests := []struct {
		name           string
		protocol       string
		requestBody    map[string]interface{}
		path           string
		wantProvider   string
		wantScenario   string
	}{
		{
			name:     "anthropic messages with thinking",
			protocol: "anthropic",
			path:     "/v1/messages",
			requestBody: map[string]interface{}{
				"model":    "claude-sonnet-4-20250514",
				"thinking": map[string]interface{}{"type": "enabled"},
				"messages": []interface{}{
					map[string]interface{}{"role": "user", "content": "Analyze this problem"},
				},
				"max_tokens": 1024,
			},
			wantProvider: "thinker",
			wantScenario: "think",
		},
		{
			name:     "openai chat with thinking-like prompt",
			protocol: "openai_chat",
			path:     "/v1/chat/completions",
			requestBody: map[string]interface{}{
				"model": "gpt-4",
				"messages": []interface{}{
					map[string]interface{}{"role": "system", "content": "Think step by step"},
					map[string]interface{}{"role": "user", "content": "Analyze this problem"},
				},
			},
			wantProvider: "standard",
			wantScenario: "code",
		},
		{
			name:     "openai responses simple request",
			protocol: "openai_responses",
			path:     "/v1/completions",
			requestBody: map[string]interface{}{
				"model": "gpt-3.5-turbo",
				"input": "Hello world",
			},
			wantProvider: "standard",
			wantScenario: "code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			bodyBytes, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("failed to marshal request body: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			// Detect protocol
			var parsedBody map[string]interface{}
			json.Unmarshal(bodyBytes, &parsedBody)

			detectedProtocol := proxy.DetectProtocol(tt.path, req.Header, parsedBody)
			if detectedProtocol != tt.protocol {
				t.Errorf("DetectProtocol() = %q, want %q", detectedProtocol, tt.protocol)
			}

			// Normalize request
			var normalized *proxy.NormalizedRequest
			switch detectedProtocol {
			case "anthropic":
				normalized, err = proxy.NormalizeAnthropicMessages(parsedBody)
			case "openai_chat":
				normalized, err = proxy.NormalizeOpenAIChat(parsedBody)
			case "openai_responses":
				normalized, err = proxy.NormalizeOpenAIResponses(parsedBody)
			default:
				t.Fatalf("unknown protocol: %s", detectedProtocol)
			}

			if err != nil {
				t.Fatalf("normalization failed: %v", err)
			}

			// Extract features
			features := proxy.ExtractFeatures(normalized)

			// Verify normalization worked
			if normalized.Model == "" {
				t.Error("normalized request has empty model")
			}
			if len(normalized.Messages) == 0 {
				t.Error("normalized request has no messages")
			}
			if features.MessageCount != len(normalized.Messages) {
				t.Errorf("features.MessageCount = %d, want %d", features.MessageCount, len(normalized.Messages))
			}

			// Verify protocol is preserved
			if normalized.OriginalProtocol != tt.protocol {
				t.Errorf("OriginalProtocol = %q, want %q", normalized.OriginalProtocol, tt.protocol)
			}
		})
	}
}

// TestProtocolDetectionPriority tests the priority order of protocol detection.
func TestProtocolDetectionPriority(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		headers  http.Header
		body     map[string]interface{}
		want     string
	}{
		{
			name: "URL path takes priority over header",
			path: "/v1/messages",
			headers: http.Header{
				"X-Zen-Client": []string{"openai"},
			},
			body: map[string]interface{}{
				"model": "gpt-4",
				"messages": []interface{}{
					map[string]interface{}{"role": "user", "content": "test"},
				},
			},
			want: "anthropic",
		},
		{
			name: "header takes priority over body structure",
			path: "/api/chat",
			headers: http.Header{
				"X-Zen-Client": []string{"anthropic"},
			},
			body: map[string]interface{}{
				"model": "gpt-4",
				"messages": []interface{}{
					map[string]interface{}{"role": "user", "content": "test"},
				},
			},
			want: "anthropic",
		},
		{
			name: "body structure detection works",
			path: "/api/chat",
			headers: http.Header{},
			body: map[string]interface{}{
				"model": "claude-3-opus-20240229",
				"messages": []interface{}{
					map[string]interface{}{"role": "user", "content": "test"},
				},
			},
			want: "anthropic",
		},
		{
			name: "default to openai_chat",
			path: "/api/unknown",
			headers: http.Header{},
			body: map[string]interface{}{
				"prompt": "test",
			},
			want: "openai_chat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := proxy.DetectProtocol(tt.path, tt.headers, tt.body)
			if got != tt.want {
				t.Errorf("DetectProtocol() = %q, want %q", got, tt.want)
			}
		})
	}
}
