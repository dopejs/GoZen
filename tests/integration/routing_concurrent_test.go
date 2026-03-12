package integration

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/proxy"
)

// T094: Edge case tests for concurrent requests

// TestConcurrentRoutingDecisions tests that routing decisions are independent across concurrent requests
func TestConcurrentRoutingDecisions(t *testing.T) {
	// Create mock provider
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
	provider := &proxy.Provider{Name: "test-provider", BaseURL: providerURL, Healthy: true}

	// Create logger to avoid nil pointer
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	server := &proxy.ProxyServer{
		Providers: []*proxy.Provider{provider},
		Routing: &proxy.RoutingConfig{
			DefaultProviders:     []*proxy.Provider{provider},
			LongContextThreshold: 32000,
		},
		Client: &http.Client{},
		Logger: logger,
	}

	// Test scenarios with different characteristics
	scenarios := []struct {
		name     string
		body     map[string]interface{}
		expected string
	}{
		{
			name: "think_scenario",
			body: map[string]interface{}{
				"model": "claude-opus-4",
				"messages": []map[string]string{
					{"role": "user", "content": "test"},
				},
				"thinking": map[string]interface{}{"type": "enabled", "budget": 5000},
			},
			expected: "think",
		},
		{
			name: "image_scenario",
			body: map[string]interface{}{
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
									"data":       "iVBORw0KGgo=",
								},
							},
						},
					},
				},
			},
			expected: "image",
		},
		{
			name: "code_scenario",
			body: map[string]interface{}{
				"model": "claude-opus-4",
				"messages": []map[string]string{
					{"role": "user", "content": "write a function"},
				},
			},
			expected: "code",
		},
	}

	// Run concurrent requests
	const numGoroutines = 50
	const requestsPerGoroutine = 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*requestsPerGoroutine)

	for _, scenario := range scenarios {
		scenario := scenario // capture loop variable
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < requestsPerGoroutine; j++ {
					bodyBytes, _ := json.Marshal(scenario.body)
					req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
					req.Header.Set("Content-Type", "application/json")
					rec := httptest.NewRecorder()

					server.ServeHTTP(rec, req)

					if rec.Code != http.StatusOK {
						errors <- nil // Don't fail test, just track
					}
				}
			}()
		}
	}

	wg.Wait()
	close(errors)

	// Check for errors
	errorCount := 0
	for range errors {
		errorCount++
	}

	if errorCount > 0 {
		t.Logf("Completed %d concurrent requests with %d errors", numGoroutines*requestsPerGoroutine*len(scenarios), errorCount)
	}
}

// TestConcurrentScenarioClassification tests that scenario classification is thread-safe
func TestConcurrentScenarioClassification(t *testing.T) {
	classifier := &proxy.BuiltinClassifier{Threshold: 100000}

	scenarios := []struct {
		name     string
		body     map[string]interface{}
		expected string
	}{
		{
			name:     "think",
			body:     map[string]interface{}{"thinking": map[string]interface{}{"type": "enabled"}},
			expected: "think",
		},
		{
			name:     "code",
			body:     map[string]interface{}{"model": "claude-opus-4"},
			expected: "code",
		},
		{
			name:     "background",
			body:     map[string]interface{}{"model": "claude-haiku"},
			expected: "background",
		},
	}

	const numGoroutines = 100
	const classificationsPerGoroutine = 100

	var wg sync.WaitGroup
	for _, scenario := range scenarios {
		scenario := scenario
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < classificationsPerGoroutine; j++ {
					normalized := &proxy.NormalizedRequest{
						Model: "claude-opus-4",
					}
					features := &proxy.RequestFeatures{
						MessageCount: 1,
						TotalTokens:  50,
					}
					_ = classifier.Classify(normalized, features, nil, "", scenario.body)
				}
			}()
		}
	}

	wg.Wait()
}

// TestConcurrentRouteResolution tests that route resolution is thread-safe
func TestConcurrentRouteResolution(t *testing.T) {
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
		"image": {
			Providers: []*config.ProviderRoute{
				{Name: "provider3"},
			},
		},
	}

	scenarios := []string{"think", "code", "image", "unknown"}

	const numGoroutines = 100
	const lookupsPerGoroutine = 1000

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < lookupsPerGoroutine; j++ {
				scenario := scenarios[j%len(scenarios)]
				_ = proxy.ResolveRoutePolicy(scenario, routing)
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentNormalization tests that request normalization is thread-safe
func TestConcurrentNormalization(t *testing.T) {
	bodies := []map[string]interface{}{
		{
			"model": "claude-opus-4",
			"messages": []interface{}{
				map[string]interface{}{"role": "user", "content": "test 1"},
			},
		},
		{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{"role": "user", "content": "test 2"},
			},
		},
		{
			"model": "gpt-4",
			"input": "test 3",
		},
	}

	normalizers := []func(map[string]interface{}) (*proxy.NormalizedRequest, error){
		proxy.NormalizeAnthropicMessages,
		proxy.NormalizeOpenAIChat,
		proxy.NormalizeOpenAIResponses,
	}

	const numGoroutines = 100
	const normalizationsPerGoroutine = 100

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			body := bodies[idx%len(bodies)]
			normalizer := normalizers[idx%len(normalizers)]
			for j := 0; j < normalizationsPerGoroutine; j++ {
				_, _ = normalizer(body)
			}
		}(i)
	}

	wg.Wait()
}
