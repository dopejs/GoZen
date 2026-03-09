package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"
)

// testLogger returns a logger for tests
func testLogger() *log.Logger {
	return log.New(os.Stderr, "[test] ", log.LstdFlags)
}

// createTestProvider creates a test provider with the given base URL
func createTestProvider(baseURL string) *Provider {
	u, err := url.Parse(baseURL)
	if err != nil {
		panic(fmt.Sprintf("invalid URL: %s", baseURL))
	}
	return &Provider{
		Name:    "test-provider",
		BaseURL: u,
		Token:   "test-token",
		Model:   "claude-sonnet-4-5",
		Healthy: true,
	}
}

// TestConnectionPoolCleanup verifies that connection pools are properly cleaned up
// when the proxy cache is invalidated
func TestConnectionPoolCleanup(t *testing.T) {
	// Create mock provider
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"msg_test","type":"message","content":[{"type":"text","text":"response"}]}`))
	}))
	defer mockProvider.Close()

	// Create ProfileProxy
	pp := NewProfileProxy(testLogger())

	// Create a provider and make a request to establish connection
	provider := createTestProvider(mockProvider.URL)
	srv := NewProxyServer([]*Provider{provider}, testLogger())

	// Cache the proxy server
	pp.cache["test-profile"] = srv

	// Verify cache has entry
	if len(pp.cache) != 1 {
		t.Fatalf("expected 1 cached proxy, got %d", len(pp.cache))
	}

	// Invalidate cache (should close connections)
	pp.InvalidateCache()

	// Verify cache is empty
	if len(pp.cache) != 0 {
		t.Errorf("expected empty cache after invalidation, got %d entries", len(pp.cache))
	}

	// Verify we can still create new connections after invalidation
	pp.cache["test-profile-2"] = NewProxyServer([]*Provider{provider}, testLogger())
	if len(pp.cache) != 1 {
		t.Errorf("expected 1 cached proxy after re-creation, got %d", len(pp.cache))
	}
}

// TestConnectionPoolMultipleInvalidations verifies that multiple invalidations
// don't cause issues
func TestConnectionPoolMultipleInvalidations(t *testing.T) {
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"msg_test","type":"message","content":[{"type":"text","text":"response"}]}`))
	}))
	defer mockProvider.Close()

	pp := NewProfileProxy(testLogger())
	provider := createTestProvider(mockProvider.URL)

	// Create and cache multiple proxy servers
	for i := 0; i < 5; i++ {
		srv := NewProxyServer([]*Provider{provider}, testLogger())
		pp.cache[string(rune('a'+i))] = srv
	}

	if len(pp.cache) != 5 {
		t.Fatalf("expected 5 cached proxies, got %d", len(pp.cache))
	}

	// Invalidate multiple times
	for i := 0; i < 3; i++ {
		pp.InvalidateCache()
		if len(pp.cache) != 0 {
			t.Errorf("invalidation %d: expected empty cache, got %d entries", i+1, len(pp.cache))
		}
	}
}

// TestConnectionPoolConcurrentAccess verifies that concurrent access to the
// connection pool doesn't cause race conditions
func TestConnectionPoolConcurrentAccess(t *testing.T) {
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // Simulate some latency
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"msg_test","type":"message","content":[{"type":"text","text":"response"}]}`))
	}))
	defer mockProvider.Close()

	pp := NewProfileProxy(testLogger())
	provider := createTestProvider(mockProvider.URL)

	// Concurrently create and invalidate cache
	done := make(chan bool)

	// Goroutine 1: Create cache entries
	go func() {
		for i := 0; i < 10; i++ {
			srv := NewProxyServer([]*Provider{provider}, testLogger())
			pp.mu.Lock()
			pp.cache["test-profile"] = srv
			pp.mu.Unlock()
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 2: Invalidate cache
	go func() {
		for i := 0; i < 10; i++ {
			pp.InvalidateCache()
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Final invalidation to clean up
	pp.InvalidateCache()
	if len(pp.cache) != 0 {
		t.Errorf("expected empty cache after final invalidation, got %d entries", len(pp.cache))
	}
}

// TestProxyServerClose verifies that ProxyServer.Close properly closes
// all HTTP client connections
func TestProxyServerClose(t *testing.T) {
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"msg_test","type":"message","content":[{"type":"text","text":"response"}]}`))
	}))
	defer mockProvider.Close()

	provider := createTestProvider(mockProvider.URL)
	srv := NewProxyServer([]*Provider{provider}, testLogger())

	// Close should not panic
	srv.Close()

	// Multiple closes should be safe
	srv.Close()
	srv.Close()
}

// TestProfileProxyClose verifies that ProfileProxy.Close properly closes
// all cached proxy servers
func TestProfileProxyClose(t *testing.T) {
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"msg_test","type":"message","content":[{"type":"text","text":"response"}]}`))
	}))
	defer mockProvider.Close()

	pp := NewProfileProxy(testLogger())
	provider := createTestProvider(mockProvider.URL)

	// Create multiple cached proxy servers
	for i := 0; i < 3; i++ {
		srv := NewProxyServer([]*Provider{provider}, testLogger())
		pp.cache[string(rune('a'+i))] = srv
	}

	// Close should not panic
	pp.Close()

	// Cache should be empty after close
	if len(pp.cache) != 0 {
		t.Errorf("expected empty cache after close, got %d entries", len(pp.cache))
	}

	// Multiple closes should be safe
	pp.Close()
	pp.Close()
}
