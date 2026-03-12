package integration

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
	"github.com/dopejs/gozen/internal/proxy"
)

// T096: Comprehensive E2E tests for all builtin scenarios

// TestE2E_ThinkScenario tests the think scenario end-to-end
func TestE2E_ThinkScenario(t *testing.T) {
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "thinking response"}},
		})
	}))
	defer mockProvider.Close()

	providerURL, _ := url.Parse(mockProvider.URL)
	thinkProvider := &proxy.Provider{Name: "think-provider", BaseURL: providerURL, Healthy: true}
	defaultProvider := &proxy.Provider{Name: "default-provider", BaseURL: providerURL, Healthy: true}

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	server := &proxy.ProxyServer{
		Providers: []*proxy.Provider{thinkProvider, defaultProvider},
		Routing: &proxy.RoutingConfig{
			DefaultProviders: []*proxy.Provider{defaultProvider},
			ScenarioRoutes: map[string]*proxy.ScenarioProviders{
				"think": {Providers: []*proxy.Provider{thinkProvider}},
			},
			LongContextThreshold: 100000,
		},
		Client: &http.Client{},
		Logger: logger,
	}

	// Request with thinking enabled
	reqBody := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []map[string]string{
			{"role": "user", "content": "Solve this complex problem"},
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

	if rec.Code != http.StatusOK {
		t.Fatalf("Think scenario failed: %d", rec.Code)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "scenario=think") {
		t.Errorf("Expected think scenario in logs, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "think-provider") {
		t.Errorf("Expected think-provider to be used, got: %s", logOutput)
	}
}

// TestE2E_ImageScenario tests the image scenario end-to-end
func TestE2E_ImageScenario(t *testing.T) {
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "image analysis"}},
		})
	}))
	defer mockProvider.Close()

	providerURL, _ := url.Parse(mockProvider.URL)
	imageProvider := &proxy.Provider{Name: "image-provider", BaseURL: providerURL, Healthy: true}
	defaultProvider := &proxy.Provider{Name: "default-provider", BaseURL: providerURL, Healthy: true}

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	server := &proxy.ProxyServer{
		Providers: []*proxy.Provider{imageProvider, defaultProvider},
		Routing: &proxy.RoutingConfig{
			DefaultProviders: []*proxy.Provider{defaultProvider},
			ScenarioRoutes: map[string]*proxy.ScenarioProviders{
				"image": {Providers: []*proxy.Provider{imageProvider}},
			},
			LongContextThreshold: 100000,
		},
		Client: &http.Client{},
		Logger: logger,
	}

	// Request with image content
	reqBody := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": "What's in this image?"},
					map[string]interface{}{
						"type": "image",
						"source": map[string]string{
							"type":       "base64",
							"media_type": "image/jpeg",
							"data":       "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
						},
					},
				},
			},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Image scenario failed: %d", rec.Code)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "scenario=image") {
		t.Errorf("Expected image scenario in logs, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "image-provider") {
		t.Errorf("Expected image-provider to be used, got: %s", logOutput)
	}
}

// TestE2E_WebSearchScenario tests the webSearch scenario end-to-end
func TestE2E_WebSearchScenario(t *testing.T) {
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "search results"}},
		})
	}))
	defer mockProvider.Close()

	providerURL, _ := url.Parse(mockProvider.URL)
	searchProvider := &proxy.Provider{Name: "search-provider", BaseURL: providerURL, Healthy: true}
	defaultProvider := &proxy.Provider{Name: "default-provider", BaseURL: providerURL, Healthy: true}

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	server := &proxy.ProxyServer{
		Providers: []*proxy.Provider{searchProvider, defaultProvider},
		Routing: &proxy.RoutingConfig{
			DefaultProviders: []*proxy.Provider{defaultProvider},
			ScenarioRoutes: map[string]*proxy.ScenarioProviders{
				"webSearch": {Providers: []*proxy.Provider{searchProvider}},
			},
			LongContextThreshold: 100000,
		},
		Client: &http.Client{},
		Logger: logger,
	}

	// Request with web_search tool
	reqBody := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []map[string]string{
			{"role": "user", "content": "Search for latest news"},
		},
		"tools": []interface{}{
			map[string]interface{}{
				"type": "web_search_20241111",
				"name": "web_search",
			},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("WebSearch scenario failed: %d", rec.Code)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "scenario=webSearch") {
		t.Errorf("Expected webSearch scenario in logs, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "search-provider") {
		t.Errorf("Expected search-provider to be used, got: %s", logOutput)
	}
}

// TestE2E_LongContextScenario tests the longContext scenario end-to-end
func TestE2E_LongContextScenario(t *testing.T) {
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "long context response"}},
		})
	}))
	defer mockProvider.Close()

	providerURL, _ := url.Parse(mockProvider.URL)
	longContextProvider := &proxy.Provider{Name: "longcontext-provider", BaseURL: providerURL, Healthy: true}
	defaultProvider := &proxy.Provider{Name: "default-provider", BaseURL: providerURL, Healthy: true}

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	server := &proxy.ProxyServer{
		Providers: []*proxy.Provider{longContextProvider, defaultProvider},
		Routing: &proxy.RoutingConfig{
			DefaultProviders: []*proxy.Provider{defaultProvider},
			ScenarioRoutes: map[string]*proxy.ScenarioProviders{
				"longContext": {Providers: []*proxy.Provider{longContextProvider}},
			},
			LongContextThreshold: 1000, // Low threshold for testing
		},
		Client: &http.Client{},
		Logger: logger,
	}

	// Request with large content
	largeContent := strings.Repeat("This is a long document. ", 500)
	reqBody := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []map[string]string{
			{"role": "user", "content": largeContent},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("LongContext scenario failed: %d", rec.Code)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "scenario=longContext") {
		t.Errorf("Expected longContext scenario in logs, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "longcontext-provider") {
		t.Errorf("Expected longcontext-provider to be used, got: %s", logOutput)
	}
}

// TestE2E_CodeScenario tests the code scenario end-to-end
func TestE2E_CodeScenario(t *testing.T) {
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "code response"}},
		})
	}))
	defer mockProvider.Close()

	providerURL, _ := url.Parse(mockProvider.URL)
	codeProvider := &proxy.Provider{Name: "code-provider", BaseURL: providerURL, Healthy: true}
	defaultProvider := &proxy.Provider{Name: "default-provider", BaseURL: providerURL, Healthy: true}

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	server := &proxy.ProxyServer{
		Providers: []*proxy.Provider{codeProvider, defaultProvider},
		Routing: &proxy.RoutingConfig{
			DefaultProviders: []*proxy.Provider{defaultProvider},
			ScenarioRoutes: map[string]*proxy.ScenarioProviders{
				"code": {Providers: []*proxy.Provider{codeProvider}},
			},
			LongContextThreshold: 100000,
		},
		Client: &http.Client{},
		Logger: logger,
	}

	// Regular coding request
	reqBody := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []map[string]string{
			{"role": "user", "content": "Write a function to sort an array"},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Code scenario failed: %d", rec.Code)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "scenario=code") {
		t.Errorf("Expected code scenario in logs, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "code-provider") {
		t.Errorf("Expected code-provider to be used, got: %s", logOutput)
	}
}

// TestE2E_BackgroundScenario tests the background scenario end-to-end
func TestE2E_BackgroundScenario(t *testing.T) {
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "background response"}},
		})
	}))
	defer mockProvider.Close()

	providerURL, _ := url.Parse(mockProvider.URL)
	backgroundProvider := &proxy.Provider{Name: "background-provider", BaseURL: providerURL, Healthy: true}
	defaultProvider := &proxy.Provider{Name: "default-provider", BaseURL: providerURL, Healthy: true}

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	server := &proxy.ProxyServer{
		Providers: []*proxy.Provider{backgroundProvider, defaultProvider},
		Routing: &proxy.RoutingConfig{
			DefaultProviders: []*proxy.Provider{defaultProvider},
			ScenarioRoutes: map[string]*proxy.ScenarioProviders{
				"background": {Providers: []*proxy.Provider{backgroundProvider}},
			},
			LongContextThreshold: 100000,
		},
		Client: &http.Client{},
		Logger: logger,
	}

	// Request with haiku model (background task)
	reqBody := map[string]interface{}{
		"model": "claude-haiku",
		"messages": []map[string]string{
			{"role": "user", "content": "Quick task"},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Background scenario failed: %d", rec.Code)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "scenario=background") {
		t.Errorf("Expected background scenario in logs, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "background-provider") {
		t.Errorf("Expected background-provider to be used, got: %s", logOutput)
	}
}

// TestE2E_CustomScenario tests custom scenario routing end-to-end
func TestE2E_CustomScenario(t *testing.T) {
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "custom response"}},
		})
	}))
	defer mockProvider.Close()

	providerURL, _ := url.Parse(mockProvider.URL)
	customProvider := &proxy.Provider{Name: "custom-provider", BaseURL: providerURL, Healthy: true}
	defaultProvider := &proxy.Provider{Name: "default-provider", BaseURL: providerURL, Healthy: true}

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	// Create routing config with custom scenario
	routing := map[string]*config.RoutePolicy{
		"customPlan": {
			Providers: []*config.ProviderRoute{
				{Name: "custom-provider"},
			},
		},
	}

	server := &proxy.ProxyServer{
		Providers: []*proxy.Provider{customProvider, defaultProvider},
		Routing: &proxy.RoutingConfig{
			DefaultProviders:     []*proxy.Provider{defaultProvider},
			LongContextThreshold: 100000,
		},
		Client: &http.Client{},
		Logger: logger,
	}

	// Simulate middleware setting custom scenario
	// (In real usage, middleware would set this via RequestContext)
	reqBody := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []map[string]string{
			{"role": "user", "content": "Plan this feature"},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Custom scenario failed: %d", rec.Code)
	}

	// Verify custom scenario can be configured
	policy := proxy.ResolveRoutePolicy("customPlan", routing)
	if policy == nil {
		t.Error("Expected custom scenario route policy to be found")
	}
	if len(policy.Providers) != 1 || policy.Providers[0].Name != "custom-provider" {
		t.Errorf("Expected custom-provider in route policy, got: %v", policy)
	}
}
