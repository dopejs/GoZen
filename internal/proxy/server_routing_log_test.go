package proxy

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

// T068: Test for middleware decision logging
func TestMiddlewareDecisionLogging(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	// Create mock provider
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{" type": "text", "text": "test"}},
		})
	}))
	defer mockProvider.Close()

	providerURL, _ := url.Parse(mockProvider.URL)
	provider := &Provider{Name: "test-provider", BaseURL: providerURL, Healthy: true}

	server := &ProxyServer{
		Providers: []*Provider{provider},
		Routing: &RoutingConfig{
			DefaultProviders:     []*Provider{provider},
			LongContextThreshold: 32000,
		},
		Logger: logger,
		Client: &http.Client{},
	}

	// Create request
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

	// Verify logging
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "[routing]") {
		t.Errorf("expected routing log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "scenario=") {
		t.Errorf("expected scenario field in log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "source=") {
		t.Errorf("expected source field in log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "confidence=") {
		t.Errorf("expected confidence field in log, got: %s", logOutput)
	}
}

// T069: Test for builtin classifier logging
func TestBuiltinClassifierLogging(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "test"}},
		})
	}))
	defer mockProvider.Close()

	providerURL, _ := url.Parse(mockProvider.URL)
	provider := &Provider{Name: "test-provider", BaseURL: providerURL, Healthy: true}

	server := &ProxyServer{
		Providers: []*Provider{provider},
		Routing: &RoutingConfig{
			DefaultProviders:     []*Provider{provider},
			LongContextThreshold: 32000,
		},
		Logger: logger,
		Client: &http.Client{},
	}

	// Create request with thinking mode (should trigger builtin classifier)
	reqBody := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []map[string]string{
			{"role": "user", "content": "test message"},
		},
		"thinking": map[string]interface{}{
			"type":   "enabled",
			"budget": 10000,
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	// Verify builtin classifier logging
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "[routing]") {
		t.Errorf("expected routing log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "scenario=think") {
		t.Errorf("expected think scenario in log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "source=builtin") {
		t.Errorf("expected builtin source in log, got: %s", logOutput)
	}
}

// T070: Test for fallback logging
func TestFallbackLogging(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	// Mock provider that returns success
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "test"}},
		})
	}))
	defer mockProvider.Close()

	// Mock provider that returns error
	failingProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("service unavailable"))
	}))
	defer failingProvider.Close()

	providerURL, _ := url.Parse(mockProvider.URL)
	failingURL, _ := url.Parse(failingProvider.URL)
	defaultProvider := &Provider{Name: "default-provider", BaseURL: providerURL, Healthy: true}
	scenarioProvider := &Provider{Name: "scenario-provider", BaseURL: failingURL, Healthy: true}

	server := &ProxyServer{
		Providers: []*Provider{defaultProvider},
		Routing: &RoutingConfig{
			DefaultProviders: []*Provider{defaultProvider},
			ScenarioRoutes: map[string]*ScenarioProviders{
				"code": {
					Providers: []*Provider{scenarioProvider},
				},
			},
			LongContextThreshold: 32000,
		},
		Logger: logger,
		Client: &http.Client{},
	}

	// Create request that triggers code scenario
	reqBody := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []map[string]string{
			{"role": "user", "content": "write a function"},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	// Verify fallback logging
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "[routing]") {
		t.Errorf("expected routing log, got: %s", logOutput)
	}
	// Should log fallback when scenario provider fails
	if !strings.Contains(logOutput, "falling back") && !strings.Contains(logOutput, "using default") {
		t.Errorf("expected fallback log, got: %s", logOutput)
	}
}

// T071: Test for provider selection logging
func TestProviderSelectionLogging(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	mockProvider1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "test"}},
		})
	}))
	defer mockProvider1.Close()

	mockProvider2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "test"}},
		})
	}))
	defer mockProvider2.Close()

	providerURL1, _ := url.Parse(mockProvider1.URL)
	providerURL2, _ := url.Parse(mockProvider2.URL)
	provider1 := &Provider{Name: "provider1", BaseURL: providerURL1, Healthy: true}
	provider2 := &Provider{Name: "provider2", BaseURL: providerURL2, Healthy: true}

	lb := NewLoadBalancer(nil)

	server := &ProxyServer{
		Providers: []*Provider{provider1, provider2},
		Routing: &RoutingConfig{
			DefaultProviders:     []*Provider{provider1, provider2},
			LongContextThreshold: 32000,
		},
		Logger:       logger,
		Client:       &http.Client{},
		LoadBalancer: lb,
		Strategy:     config.LoadBalanceRoundRobin,
	}

	// Create request
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

	// Verify provider selection logging
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "[routing]") {
		t.Errorf("expected routing log, got: %s", logOutput)
	}
	// Should log which provider was selected
	if !strings.Contains(logOutput, "provider1") && !strings.Contains(logOutput, "provider2") {
		t.Errorf("expected provider name in log, got: %s", logOutput)
	}
}

// T077: Test for request features logging
func TestRequestFeaturesLogging(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "test"}},
		})
	}))
	defer mockProvider.Close()

	providerURL, _ := url.Parse(mockProvider.URL)
	provider := &Provider{Name: "test-provider", BaseURL: providerURL, Healthy: true}

	server := &ProxyServer{
		Providers: []*Provider{provider},
		Routing: &RoutingConfig{
			DefaultProviders:     []*Provider{provider},
			LongContextThreshold: 32000,
		},
		Logger: logger,
		Client: &http.Client{},
	}

	// Create request with various features
	reqBody := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{"type": "text", "text": "What's in this image?"},
					{
						"type": "image",
						"source": map[string]string{
							"type":       "base64",
							"media_type": "image/jpeg",
							"data":       "iVBORw0KGgo=",
						},
					},
				},
			},
		},
		"thinking": map[string]interface{}{
			"type":   "enabled",
			"budget": 5000,
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	// Verify request features logging
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "[routing] features:") {
		t.Errorf("expected features log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "has_image=") {
		t.Errorf("expected has_image field in log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "has_tools=") {
		t.Errorf("expected has_tools field in log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "total_tokens=") {
		t.Errorf("expected total_tokens field in log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "message_count=") {
		t.Errorf("expected message_count field in log, got: %s", logOutput)
	}
}
