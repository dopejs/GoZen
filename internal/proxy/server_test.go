package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func discardLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}

func TestSingleJoiningSlash(t *testing.T) {
	tests := []struct {
		a, b, want string
	}{
		{"http://host", "/path", "http://host/path"},
		{"http://host/", "/path", "http://host/path"},
		{"http://host/", "path", "http://host/path"},
		{"http://host", "path", "http://host/path"},
		{"http://host/api", "/v1/messages", "http://host/api/v1/messages"},
		{"http://host/api/", "/v1/messages", "http://host/api/v1/messages"},
	}
	for _, tt := range tests {
		got := singleJoiningSlash(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("singleJoiningSlash(%q, %q) = %q, want %q", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestInjectModel(t *testing.T) {
	body := []byte(`{"prompt":"hello","max_tokens":100}`)
	result := injectModel(body, "claude-opus")

	var data map[string]interface{}
	if err := json.Unmarshal(result, &data); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if data["model"] != "claude-opus" {
		t.Errorf("model = %v, want %q", data["model"], "claude-opus")
	}
	if data["prompt"] != "hello" {
		t.Errorf("prompt should be preserved, got %v", data["prompt"])
	}
	if data["max_tokens"] != float64(100) {
		t.Errorf("max_tokens should be preserved, got %v", data["max_tokens"])
	}
}

func TestInjectModelOverridesExisting(t *testing.T) {
	body := []byte(`{"model":"old-model","prompt":"hi"}`)
	result := injectModel(body, "new-model")

	var data map[string]interface{}
	json.Unmarshal(result, &data)
	if data["model"] != "new-model" {
		t.Errorf("model = %v, want %q", data["model"], "new-model")
	}
}

func TestInjectModelInvalidJSON(t *testing.T) {
	body := []byte(`not json`)
	result := injectModel(body, "model")
	if string(result) != "not json" {
		t.Errorf("should return original body for invalid JSON")
	}
}

func TestInjectModelEmptyModel(t *testing.T) {
	body := []byte(`{"prompt":"hi"}`)
	// When model is empty, forwardRequest skips injection,
	// but injectModel itself still works
	result := injectModel(body, "")
	var data map[string]interface{}
	json.Unmarshal(result, &data)
	if data["model"] != "" {
		t.Errorf("model = %v, want empty", data["model"])
	}
}

// TestServeHTTPSuccess tests a successful proxy request.
func TestServeHTTPSuccess(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth headers
		if r.Header.Get("x-api-key") != "test-token" {
			t.Errorf("x-api-key = %q", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}

		// Verify model injection
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if data["model"] != "test-model" {
			t.Errorf("model = %v, want %q", data["model"], "test-model")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{{
		Name: "test", BaseURL: u, Token: "test-token", Model: "test-model", Healthy: true,
	}}

	srv := NewProxyServer(providers, discardLogger())

	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"ok":true`) {
		t.Errorf("body = %q", w.Body.String())
	}
}

// TestServeHTTPFailoverOn500 tests that 500 triggers failover to next provider.
func TestServeHTTPFailoverOn500(t *testing.T) {
	callCount := 0
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(500)
		w.Write([]byte("error"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "t1", Model: "m", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "t2", Model: "m", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())

	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200 (failover)", w.Code)
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
}

// TestServeHTTPFailoverOn429 tests that 429 triggers failover.
func TestServeHTTPFailoverOn429(t *testing.T) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "t1", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "t2", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestServeHTTPAllProvidersFail tests 502 when all providers fail.
func TestServeHTTPAllProvidersFail(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadGateway)
	}
}

// TestServeHTTPSkipsUnhealthyProvider tests that unhealthy providers are skipped.
func TestServeHTTPSkipsUnhealthyProvider(t *testing.T) {
	called := make(map[string]bool)

	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called["p1"] = true
		w.WriteHeader(200)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called["p2"] = true
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	p1 := &Provider{Name: "p1", BaseURL: u1, Token: "t1", Healthy: true}
	p2 := &Provider{Name: "p2", BaseURL: u2, Token: "t2", Healthy: true}

	// Mark p1 as unhealthy
	p1.MarkFailed()

	srv := NewProxyServer([]*Provider{p1, p2}, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if called["p1"] {
		t.Error("p1 should have been skipped (unhealthy)")
	}
	if !called["p2"] {
		t.Error("p2 should have been called")
	}
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestServeHTTPNoModelInjectionWhenEmpty tests that empty model skips injection.
func TestServeHTTPNoModelInjectionWhenEmpty(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		if _, ok := data["model"]; ok {
			t.Error("model should not be injected when provider model is empty")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Model: "", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"prompt":"hi"}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
}

// TestServeHTTPPreservesQueryString tests that query params are forwarded.
func TestServeHTTPPreservesQueryString(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "beta=true" {
			t.Errorf("query = %q, want %q", r.URL.RawQuery, "beta=true")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages?beta=true", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
}

// TestServeHTTPSSEStreaming tests SSE response streaming.
func TestServeHTTPSSEStreaming(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		w.Write([]byte("data: hello\n\n"))
		w.Write([]byte("data: world\n\n"))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "data: hello") || !strings.Contains(body, "data: world") {
		t.Errorf("SSE body = %q", body)
	}
}

// TestStartProxy tests that StartProxy returns a valid port.
func TestStartProxy(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	port, err := StartProxy(providers, "127.0.0.1:0", discardLogger())
	if err != nil {
		t.Fatalf("StartProxy() error: %v", err)
	}
	if port <= 0 {
		t.Errorf("port = %d, want > 0", port)
	}

	// Verify the server is actually listening
	resp, err := http.Post(
		fmt.Sprintf("http://127.0.0.1:%d/v1/messages", port),
		"application/json",
		strings.NewReader(`{}`),
	)
	if err != nil {
		t.Fatalf("request to proxy error: %v", err)
	}
	resp.Body.Close()
	// Should get 502 since the backend URL is fake
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadGateway)
	}
}

func TestNewProxyServer(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}
	srv := NewProxyServer(providers, discardLogger())
	if srv == nil {
		t.Fatal("NewProxyServer returned nil")
	}
	if len(srv.Providers) != 1 {
		t.Errorf("providers count = %d, want 1", len(srv.Providers))
	}
	if srv.Client == nil {
		t.Error("Client should not be nil")
	}
}

// TestServeHTTPCopiesResponseHeaders tests that response headers are forwarded.
func TestServeHTTPCopiesResponseHeaders(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "custom-value")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Header().Get("X-Custom-Header") != "custom-value" {
		t.Errorf("X-Custom-Header = %q, want %q", w.Header().Get("X-Custom-Header"), "custom-value")
	}
}

// TestStartProxyListenError tests that StartProxy returns error for invalid address.
func TestStartProxyListenError(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	// Use an invalid listen address
	_, err := StartProxy(providers, "999.999.999.999:0", discardLogger())
	if err == nil {
		t.Error("expected error for invalid listen address")
	}
}

// TestServeHTTPConnectionError tests failover when backend is unreachable.
func TestServeHTTPConnectionError(t *testing.T) {
	// Use a URL that will refuse connections
	u1, _ := url.Parse("http://127.0.0.1:1") // port 1 should refuse
	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer backend2.Close()
	u2, _ := url.Parse(backend2.URL)

	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "t1", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "t2", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200 (failover from connection error)", w.Code)
	}
}

// TestServeHTTPBadBodyRead tests handling of body read error.
func TestServeHTTPBadBodyRead(t *testing.T) {
	u, _ := url.Parse("https://api.example.com")
	providers := []*Provider{
		{Name: "p1", BaseURL: u, Token: "t1", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", &errorReader{})
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadGateway)
	}
}

// errorReader always returns an error on Read.
type errorReader struct{}

func (e *errorReader) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("read error")
}

// TestServeHTTP200DoesNotFailover tests that 2xx/3xx/4xx (non-429) don't trigger failover.
func TestServeHTTP4xxNoFailover(t *testing.T) {
	callCount := 0
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(400)
		w.Write([]byte("bad request"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
	}))
	defer backend2.Close()

	u1, _ := url.Parse(backend1.URL)
	u2, _ := url.Parse(backend2.URL)
	providers := []*Provider{
		{Name: "p1", BaseURL: u1, Token: "t1", Healthy: true},
		{Name: "p2", BaseURL: u2, Token: "t2", Healthy: true},
	}

	srv := NewProxyServer(providers, discardLogger())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// 400 should NOT trigger failover — only 429 and 5xx do
	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1 (no failover for 400)", callCount)
	}
}
