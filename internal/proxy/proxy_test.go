package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestNewHTTPClientWithProxy(t *testing.T) {
	t.Run("http scheme sets Transport.Proxy", func(t *testing.T) {
		client, err := NewHTTPClientWithProxy("http://proxy:8080", 10*time.Second)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		transport, ok := client.Transport.(*http.Transport)
		if !ok {
			t.Fatal("expected *http.Transport")
		}
		if transport.Proxy == nil {
			t.Fatal("expected Proxy function to be set")
		}
		// Verify the proxy function returns the correct URL
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		proxyURL, err := transport.Proxy(req)
		if err != nil {
			t.Fatalf("proxy function error: %v", err)
		}
		if proxyURL.Host != "proxy:8080" {
			t.Errorf("proxy host = %q, want %q", proxyURL.Host, "proxy:8080")
		}
	})

	t.Run("https scheme sets Transport.Proxy", func(t *testing.T) {
		client, err := NewHTTPClientWithProxy("https://proxy:8443", 10*time.Second)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		transport, ok := client.Transport.(*http.Transport)
		if !ok {
			t.Fatal("expected *http.Transport")
		}
		if transport.Proxy == nil {
			t.Fatal("expected Proxy function to be set")
		}
	})

	t.Run("socks5 scheme sets DialContext", func(t *testing.T) {
		client, err := NewHTTPClientWithProxy("socks5://proxy:1080", 10*time.Second)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		transport, ok := client.Transport.(*http.Transport)
		if !ok {
			t.Fatal("expected *http.Transport")
		}
		if transport.DialContext == nil {
			t.Fatal("expected DialContext to be set for SOCKS5")
		}
		// For SOCKS5, Proxy should NOT be set
		if transport.Proxy != nil {
			t.Fatal("expected Proxy to be nil for SOCKS5")
		}
	})

	t.Run("empty URL returns error", func(t *testing.T) {
		_, err := NewHTTPClientWithProxy("", 10*time.Second)
		if err == nil {
			t.Fatal("expected error for empty URL")
		}
	})

	t.Run("unsupported scheme returns error", func(t *testing.T) {
		_, err := NewHTTPClientWithProxy("ftp://proxy:21", 10*time.Second)
		if err == nil {
			t.Fatal("expected error for unsupported scheme")
		}
	})

	t.Run("timeout is set", func(t *testing.T) {
		client, err := NewHTTPClientWithProxy("http://proxy:8080", 30*time.Second)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client.Timeout != 30*time.Second {
			t.Errorf("timeout = %v, want %v", client.Timeout, 30*time.Second)
		}
	})
}

func TestForwardRequestUsesProviderClient(t *testing.T) {
	// Create a backend that echoes a custom header to verify which client was used
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Received", "true")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend.Close()

	backendURL, _ := url.Parse(backend.URL)

	// Track which client was used
	providerClientUsed := false
	sharedClientUsed := false

	providerClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			providerClientUsed = true
			return http.DefaultTransport.RoundTrip(req)
		}),
	}

	sharedClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			sharedClientUsed = true
			return http.DefaultTransport.RoundTrip(req)
		}),
	}

	s := &ProxyServer{
		Client: sharedClient,
		Logger: discardLogger(),
	}

	t.Run("uses provider client when set", func(t *testing.T) {
		providerClientUsed = false
		sharedClientUsed = false

		p := &Provider{
			Name:    "test",
			BaseURL: backendURL,
			Token:   "tok",
			Client:  providerClient,
			Healthy: true,
		}

		req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-4-5"}`))
		resp, err := s.forwardRequest(req, p, []byte(`{"model":"claude-sonnet-4-5"}`), "", "anthropic")
		if err != nil {
			t.Fatalf("forwardRequest error: %v", err)
		}
		defer resp.Body.Close()
		io.ReadAll(resp.Body)

		if !providerClientUsed {
			t.Error("expected provider client to be used")
		}
		if sharedClientUsed {
			t.Error("expected shared client NOT to be used")
		}
	})

	t.Run("falls back to shared client when provider client is nil", func(t *testing.T) {
		providerClientUsed = false
		sharedClientUsed = false

		p := &Provider{
			Name:    "test",
			BaseURL: backendURL,
			Token:   "tok",
			Client:  nil, // no per-provider client
			Healthy: true,
		}

		req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-4-5"}`))
		resp, err := s.forwardRequest(req, p, []byte(`{"model":"claude-sonnet-4-5"}`), "", "anthropic")
		if err != nil {
			t.Fatalf("forwardRequest error: %v", err)
		}
		defer resp.Body.Close()
		io.ReadAll(resp.Body)

		if providerClientUsed {
			t.Error("expected provider client NOT to be used")
		}
		if !sharedClientUsed {
			t.Error("expected shared client to be used")
		}
	})
}

// roundTripFunc is an adapter to allow the use of ordinary functions as http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
