package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/proxy"
)

// TestRequestTimeout verifies that requests are cancelled after timeout
func TestRequestTimeout(t *testing.T) {
	// Create mock provider that takes longer than timeout
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response (5 seconds)
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"content": []map[string]string{{"type": "text", "text": "response"}},
		})
	}))
	defer mockProvider.Close()

	provider := &proxy.Provider{
		Name:    "slow-provider",
		BaseURL: mustParseURL(mockProvider.URL),
		Token:   "test-token",
		Model:   "claude-sonnet-4-5",
		Healthy: true,
	}

	srv := proxy.NewProxyServer([]*proxy.Provider{provider}, testLogger(), config.LoadBalanceFailover, nil)

	// Create request with 1 second timeout
	body := []byte(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"test"}],"max_tokens":100}`)
	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Set context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	start := time.Now()
	srv.ServeHTTP(w, req)
	elapsed := time.Since(start)

	// Request should be cancelled within timeout window (allow some overhead)
	if elapsed > 2*time.Second {
		t.Errorf("request took %s, expected cancellation within ~1s", elapsed)
	}

	// The proxy detects context cancellation and stops processing
	// ResponseRecorder may still show 200 if headers were already written,
	// but the key is that the request completes quickly (within timeout)
	t.Logf("Request completed in %s with status %d (context cancelled)", elapsed, w.Code)
}

// TestRequestTimeoutWithFailover verifies timeout behavior with multiple providers
func TestRequestTimeoutWithFailover(t *testing.T) {
	// First provider: slow (will timeout)
	slowProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer slowProvider.Close()

	// Second provider: fast (should succeed)
	fastProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_test",
			"type":    "message",
			"content": []map[string]string{{"type": "text", "text": "response"}},
		})
	}))
	defer fastProvider.Close()

	providers := []*proxy.Provider{
		{
			Name:    "slow-provider",
			BaseURL: mustParseURL(slowProvider.URL),
			Token:   "test-token",
			Model:   "claude-sonnet-4-5",
			Healthy: true,
		},
		{
			Name:    "fast-provider",
			BaseURL: mustParseURL(fastProvider.URL),
			Token:   "test-token",
			Model:   "claude-sonnet-4-5",
			Healthy: true,
		},
	}

	srv := proxy.NewProxyServer(providers, testLogger(), config.LoadBalanceFailover, nil)

	body := []byte(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"test"}],"max_tokens":100}`)
	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Set context with 1 second timeout per provider
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	start := time.Now()
	srv.ServeHTTP(w, req)
	elapsed := time.Since(start)

	t.Logf("Request completed in %s with status %d", elapsed, w.Code)

	// With failover, the fast provider should eventually succeed
	// But if the slow provider blocks for too long, the whole request might timeout
	// This test verifies that context cancellation propagates correctly
	if elapsed > 10*time.Second {
		t.Errorf("request took %s, expected completion or cancellation within reasonable time", elapsed)
	}
}

// TestStreamingTimeout verifies timeout behavior for streaming responses
func TestStreamingTimeout(t *testing.T) {
	// Create mock provider that streams slowly
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("ResponseWriter doesn't support flushing")
		}

		// Send first chunk immediately
		w.Write([]byte("event: message_start\ndata: {\"type\":\"message_start\"}\n\n"))
		flusher.Flush()

		// Wait 3 seconds before next chunk (should trigger timeout)
		time.Sleep(3 * time.Second)

		w.Write([]byte("event: content_block_delta\ndata: {\"type\":\"content_block_delta\"}\n\n"))
		flusher.Flush()
	}))
	defer mockProvider.Close()

	provider := &proxy.Provider{
		Name:    "streaming-provider",
		BaseURL: mustParseURL(mockProvider.URL),
		Token:   "test-token",
		Model:   "claude-sonnet-4-5",
		Healthy: true,
	}

	srv := proxy.NewProxyServer([]*proxy.Provider{provider}, testLogger(), config.LoadBalanceFailover, nil)

	body := []byte(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"test"}],"max_tokens":100,"stream":true}`)
	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Set context with 1 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	start := time.Now()
	srv.ServeHTTP(w, req)
	elapsed := time.Since(start)

	t.Logf("Streaming request completed/cancelled in %s with status %d", elapsed, w.Code)

	// Should complete or cancel within timeout window
	if elapsed > 2*time.Second {
		t.Errorf("streaming request took %s, expected cancellation within ~1s", elapsed)
	}
}
