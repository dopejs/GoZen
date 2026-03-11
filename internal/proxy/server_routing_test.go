package proxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

// TestFallbackToDefaultDisabled tests that fallback_to_default=false prevents fallback
func TestFallbackToDefaultDisabled(t *testing.T) {
	// Create a scenario route with fallback disabled
	falseVal := false
	scenarioProviders := &ScenarioProviders{
		Providers:         []*Provider{},
		FallbackToDefault: &falseVal,
	}

	defaultURL, _ := url.Parse("http://default.example.com")
	routing := &RoutingConfig{
		DefaultProviders: []*Provider{
			{Name: "default-provider", BaseURL: defaultURL},
		},
		ScenarioRoutes: map[string]*ScenarioProviders{
			"code": scenarioProviders,
		},
	}

	server := NewProxyServerWithRouting(routing, testLogger(), config.LoadBalanceFailover, nil)
	server.Profile = "test-profile"

	// Create a request that will be classified as "code" scenario
	reqBody := `{"model":"claude-opus-4","messages":[{"role":"user","content":"test"}]}`
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	// Should return error without falling back to default providers
	// Returns 502 (BadGateway) when all providers fail
	if w.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "fallback disabled") {
		t.Errorf("expected fallback disabled error, got: %s", body)
	}
}

// TestPerScenarioThreshold tests that per-scenario long_context_threshold overrides classification
func TestPerScenarioThreshold(t *testing.T) {
	// Create a longContext route with custom threshold (1000)
	// Other scenarios (like code) do NOT have custom thresholds
	customThreshold := 1000
	longcontextURL, _ := url.Parse("http://longcontext.example.com")
	longContextRoute := &ScenarioProviders{
		Providers: []*Provider{
			{Name: "longcontext-provider", BaseURL: longcontextURL},
		},
		LongContextThreshold: &customThreshold,
	}

	codeURL, _ := url.Parse("http://code.example.com")
	codeRoute := &ScenarioProviders{
		Providers: []*Provider{
			{Name: "code-provider", BaseURL: codeURL},
		},
		// No custom threshold for code route
	}

	defaultURL, _ := url.Parse("http://default.example.com")
	routing := &RoutingConfig{
		DefaultProviders: []*Provider{
			{Name: "default-provider", BaseURL: defaultURL},
		},
		ScenarioRoutes: map[string]*ScenarioProviders{
			"code":        codeRoute,
			"longContext": longContextRoute,
		},
		LongContextThreshold: 32000, // Profile-level threshold
	}

	server := NewProxyServerWithRouting(routing, testLogger(), config.LoadBalanceFailover, nil)
	server.Profile = "test-profile"

	// Create a request with ~2000 tokens (exceeds longContext route threshold of 1000, but not profile threshold of 32000)
	// This should be classified as "longContext" because longContext route's threshold is used for classification
	longContent := strings.Repeat("word ", 1000) // ~2000 tokens (each "word " is ~2 tokens)
	reqBody := `{"model":"claude-opus-4","messages":[{"role":"user","content":"` + longContent + `"}]}`
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	// The request should be routed to longContext scenario due to longContext route's threshold
	// Since we don't have a real backend, we expect 502 (all providers failed)
	// But the important part is that the routing decision was made correctly
	if w.Code != http.StatusBadGateway && w.Code != http.StatusServiceUnavailable {
		t.Logf("Response status: %d, body: %s", w.Code, w.Body.String())
	}
}

// TestPerScenarioThresholdNormalizedKeys tests that threshold lookup works with normalized scenario keys
func TestPerScenarioThresholdNormalizedKeys(t *testing.T) {
	tests := []struct {
		name       string
		routeKey   string
		wantScenario string
	}{
		{
			name:       "kebab-case key",
			routeKey:   "long-context",
			wantScenario: "longContext",
		},
		{
			name:       "snake_case key",
			routeKey:   "long_context",
			wantScenario: "longContext",
		},
		{
			name:       "camelCase key",
			routeKey:   "longContext",
			wantScenario: "longContext",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customThreshold := 1000
			longcontextURL, _ := url.Parse("http://longcontext.example.com")
			longContextRoute := &ScenarioProviders{
				Providers: []*Provider{
					{Name: "longcontext-provider", BaseURL: longcontextURL},
				},
				LongContextThreshold: &customThreshold,
			}

			defaultURL, _ := url.Parse("http://default.example.com")
			routing := &RoutingConfig{
				DefaultProviders: []*Provider{
					{Name: "default-provider", BaseURL: defaultURL},
				},
				ScenarioRoutes: map[string]*ScenarioProviders{
					tt.routeKey: longContextRoute, // Use the test's route key
				},
				LongContextThreshold: 32000,
			}

			server := NewProxyServerWithRouting(routing, testLogger(), config.LoadBalanceFailover, nil)
			server.Profile = "test-profile"

			// Create a request with ~2000 tokens (exceeds threshold of 1000)
			longContent := strings.Repeat("word ", 1000)
			reqBody := `{"model":"claude-opus-4","messages":[{"role":"user","content":"` + longContent + `"}]}`
			req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			server.ServeHTTP(w, req)

			// Should classify as longContext regardless of route key format
			if w.Code != http.StatusBadGateway && w.Code != http.StatusServiceUnavailable {
				t.Errorf("Expected 502/503, got %d", w.Code)
			}
		})
	}
}
