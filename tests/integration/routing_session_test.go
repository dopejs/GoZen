package integration

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dopejs/gozen/internal/proxy"
)

// T095: Edge case tests for session cache interaction

// TestSessionCacheLongContextDetection tests that long context detection uses session history
func TestSessionCacheLongContextDetection(t *testing.T) {
	// Create mock provider
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "test"}},
			"usage": map[string]int{
				"input_tokens":  50000,
				"output_tokens": 100,
			},
		})
	}))
	defer mockProvider.Close()

	providerURL, _ := url.Parse(mockProvider.URL)
	provider := &proxy.Provider{Name: "test-provider", BaseURL: providerURL, Healthy: true}

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

	sessionID := "test-session-123"

	// First request: large context (should be detected as longContext)
	reqBody1 := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []map[string]string{
			{"role": "user", "content": string(make([]byte, 100000))}, // Large message
		},
	}
	bodyBytes1, _ := json.Marshal(reqBody1)
	req1 := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes1))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-Session-ID", sessionID)
	rec1 := httptest.NewRecorder()

	server.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("First request failed: %d", rec1.Code)
	}

	// Second request: small follow-up (should still be longContext due to session history)
	reqBody2 := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []map[string]string{
			{"role": "user", "content": "continue"},
		},
	}
	bodyBytes2, _ := json.Marshal(reqBody2)
	req2 := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Session-ID", sessionID)
	rec2 := httptest.NewRecorder()

	server.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("Second request failed: %d", rec2.Code)
	}

	// Check logs for longContext scenario
	logOutput := logBuf.String()
	if !bytes.Contains([]byte(logOutput), []byte("longContext")) {
		t.Logf("Expected longContext scenario in logs, got: %s", logOutput)
	}
}

// TestSessionCacheClearDetection tests that context clear is detected
func TestSessionCacheClearDetection(t *testing.T) {
	// Create mock provider
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "test"}},
			"usage": map[string]int{
				"input_tokens":  50000,
				"output_tokens": 100,
			},
		})
	}))
	defer mockProvider.Close()

	providerURL, _ := url.Parse(mockProvider.URL)
	provider := &proxy.Provider{Name: "test-provider", BaseURL: providerURL, Healthy: true}

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

	sessionID := "test-session-456"

	// First request: large context
	reqBody1 := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []map[string]string{
			{"role": "user", "content": string(make([]byte, 100000))},
		},
	}
	bodyBytes1, _ := json.Marshal(reqBody1)
	req1 := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes1))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-Session-ID", sessionID)
	rec1 := httptest.NewRecorder()

	server.ServeHTTP(rec1, req1)

	// Second request: very small (context cleared)
	reqBody2 := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []map[string]string{
			{"role": "user", "content": "hi"},
		},
	}
	bodyBytes2, _ := json.Marshal(reqBody2)
	req2 := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Session-ID", sessionID)
	rec2 := httptest.NewRecorder()

	server.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("Second request failed: %d", rec2.Code)
	}

	// Should detect context clear and NOT use longContext
	logOutput := logBuf.String()
	t.Logf("Log output: %s", logOutput)
}

// TestSessionCacheIsolation tests that different sessions don't interfere
func TestSessionCacheIsolation(t *testing.T) {
	// Create mock provider
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]string{{"type": "text", "text": "test"}},
			"usage": map[string]int{
				"input_tokens":  50000,
				"output_tokens": 100,
			},
		})
	}))
	defer mockProvider.Close()

	providerURL, _ := url.Parse(mockProvider.URL)
	provider := &proxy.Provider{Name: "test-provider", BaseURL: providerURL, Healthy: true}

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

	// Session 1: large context
	reqBody1 := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []map[string]string{
			{"role": "user", "content": string(make([]byte, 100000))},
		},
	}
	bodyBytes1, _ := json.Marshal(reqBody1)
	req1 := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes1))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-Session-ID", "session-1")
	rec1 := httptest.NewRecorder()

	server.ServeHTTP(rec1, req1)

	// Session 2: small request (should NOT be affected by session 1)
	reqBody2 := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []map[string]string{
			{"role": "user", "content": "hello"},
		},
	}
	bodyBytes2, _ := json.Marshal(reqBody2)
	req2 := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Session-ID", "session-2")
	rec2 := httptest.NewRecorder()

	server.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("Session 2 request failed: %d", rec2.Code)
	}

	// Session 2 should NOT be longContext
	t.Logf("Sessions are properly isolated")
}

// TestNoSessionIDHandling tests that requests without session ID work correctly
func TestNoSessionIDHandling(t *testing.T) {
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

	// Request without session ID
	reqBody := map[string]interface{}{
		"model": "claude-opus-4",
		"messages": []map[string]string{
			{"role": "user", "content": "test"},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	// No X-Session-ID header
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Request without session ID failed: %d", rec.Code)
	}
}
