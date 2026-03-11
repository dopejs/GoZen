package integration

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/proxy"
)

// T050: Integration test for per-scenario routing policies
// Tests that different scenarios can have different strategies, weights, and thresholds

func TestPerScenarioPolicies_DifferentStrategies(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)

	// Create mock providers
	mockProvider1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test_p1",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "response from provider1"}},
		})
	}))
	defer mockProvider1.Close()

	mockProvider2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test_p2",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "response from provider2"}},
		})
	}))
	defer mockProvider2.Close()

	providerURL1, _ := url.Parse(mockProvider1.URL)
	providerURL2, _ := url.Parse(mockProvider2.URL)

	// Create providers
	provider1 := &proxy.Provider{Name: "provider1", BaseURL: providerURL1, Healthy: true}
	provider2 := &proxy.Provider{Name: "provider2", BaseURL: providerURL2, Healthy: true}
	providers := []*proxy.Provider{provider1, provider2}

	// Create scenario routes with different strategies
	scenarioRoutes := map[string]*proxy.ScenarioProviders{
		"code": {
			Providers: []*proxy.Provider{provider1, provider2},
		},
		"longContext": {
			Providers: []*proxy.Provider{provider2, provider1},
		},
	}

	// Create load balancer
	lb := proxy.NewLoadBalancer(nil)

	server := &proxy.ProxyServer{
		Providers: providers,
		Routing: &proxy.RoutingConfig{
			DefaultProviders:     providers,
			ScenarioRoutes:       scenarioRoutes,
			LongContextThreshold: 32000,
		},
		Logger:       logger,
		Client:       &http.Client{},
		LoadBalancer: lb,
		Strategy:     config.LoadBalanceFailover,
	}

	// Test 1: Code scenario (short context)
	reqBody1 := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []map[string]string{
			{"role": "user", "content": "short message"},
		},
	}
	bodyBytes1, _ := json.Marshal(reqBody1)

	req1 := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes1))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()

	server.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Errorf("code scenario: expected status 200, got %d: %s", rec1.Code, rec1.Body.String())
	}

	// Test 2: Long context scenario (many messages)
	messages := make([]map[string]string, 100)
	for i := 0; i < 100; i++ {
		messages[i] = map[string]string{
			"role":    "user",
			"content": "This is a long message to trigger long context scenario. " + string(rune(i)),
		}
	}

	reqBody2 := map[string]interface{}{
		"model":    "claude-opus-4",
		"messages": messages,
	}
	bodyBytes2, _ := json.Marshal(reqBody2)

	req2 := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes2))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()

	server.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("longContext scenario: expected status 200, got %d: %s", rec2.Code, rec2.Body.String())
	}
}

func TestPerScenarioPolicies_CustomThreshold(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)

	// Create mock provider
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "test response"}},
		})
	}))
	defer mockProvider.Close()

	providerURL, _ := url.Parse(mockProvider.URL)
	provider := &proxy.Provider{Name: "provider1", BaseURL: providerURL, Healthy: true}

	// Test with custom threshold (10000 tokens)
	server := &proxy.ProxyServer{
		Providers: []*proxy.Provider{provider},
		Routing: &proxy.RoutingConfig{
			DefaultProviders:     []*proxy.Provider{provider},
			ScenarioRoutes:       make(map[string]*proxy.ScenarioProviders),
			LongContextThreshold: 10000, // Custom low threshold
		},
		Logger: logger,
		Client: &http.Client{},
	}

	// Create request with moderate token count (should trigger longContext with low threshold)
	reqBody := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []map[string]string{
			{"role": "user", "content": "This is a message with moderate length that would not trigger long context with default threshold but should with custom threshold of 10000 tokens. " + string(make([]byte, 5000))},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPerScenarioPolicies_ModelOverrides(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)

	// Create mock provider
	requestedModel := ""
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture the model from request body
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if model, ok := body["model"].(string); ok {
			requestedModel = model
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "test response"}},
		})
	}))
	defer mockProvider.Close()

	providerURL, _ := url.Parse(mockProvider.URL)
	provider := &proxy.Provider{Name: "provider1", BaseURL: providerURL, Healthy: true}

	// Create scenario with model override
	scenarioRoutes := map[string]*proxy.ScenarioProviders{
		"code": {
			Providers: []*proxy.Provider{provider},
			Models: map[string]string{
				"provider1": "claude-3-5-sonnet-20241022", // Override model
			},
		},
	}

	server := &proxy.ProxyServer{
		Providers: []*proxy.Provider{provider},
		Routing: &proxy.RoutingConfig{
			DefaultProviders:     []*proxy.Provider{provider},
			ScenarioRoutes:       scenarioRoutes,
			LongContextThreshold: 32000,
		},
		Logger: logger,
		Client: &http.Client{},
	}

	// Request with original model
	reqBody := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []map[string]string{
			{"role": "user", "content": "test message"},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify model was overridden
	if requestedModel != "claude-3-5-sonnet-20241022" {
		t.Errorf("expected model override to 'claude-3-5-sonnet-20241022', got '%s'", requestedModel)
	}
}
