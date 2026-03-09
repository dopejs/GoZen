package proxy

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

type mockTempProfileProvider struct {
	profiles map[string][]string
}

func (m *mockTempProfileProvider) GetTempProfileProviders(id string) []string {
	return m.profiles[id]
}

func TestNewProfileProxy(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)
	if pp == nil {
		t.Fatal("NewProfileProxy returned nil")
	}
	if pp.cache == nil {
		t.Error("cache should be initialized")
	}
}

func TestProfileProxyInvalidPath(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)
	pp.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid path, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	errObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["type"] != "invalid_path" {
		t.Errorf("error type = %q, want invalid_path", errObj["type"])
	}
}

func TestProfileProxySingleSegmentPath(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/onlyone", nil)
	pp.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProfileProxyTempProfileNotSupported(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)
	// TempProfiles is nil

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/_tmp_abc/sess1/v1/messages", nil)
	pp.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestProfileProxyTempProfileNotFound(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)
	pp.TempProfiles = &mockTempProfileProvider{profiles: map[string][]string{}}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/_tmp_abc/sess1/v1/messages", nil)
	pp.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestProfileProxyProfileNotFound(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	// Use a temp HOME so config store is empty
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/nonexistent/sess1/v1/messages", nil)
	pp.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestProfileProxyInvalidateCache(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	sharedTransport := &trackingTransport{}
	providerTransport := &trackingTransport{}

	pp.cache["test"] = &ProxyServer{
		Client: &http.Client{Transport: sharedTransport},
		Providers: []*Provider{{
			Name:   "provider-a",
			Client: &http.Client{Transport: providerTransport},
		}},
	}
	if len(pp.cache) != 1 {
		t.Fatal("cache should have 1 entry")
	}

	pp.InvalidateCache()
	if len(pp.cache) != 0 {
		t.Error("cache should be empty after InvalidateCache")
	}
	if !sharedTransport.closed {
		t.Error("expected shared proxy client idle connections to be closed")
	}
	if !providerTransport.closed {
		t.Error("expected provider proxy client idle connections to be closed")
	}
}

func TestProfileProxyWriteError(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	w := httptest.NewRecorder()
	pp.writeError(w, http.StatusTeapot, "test_error", "test message")

	if w.Code != http.StatusTeapot {
		t.Fatalf("expected 418, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	errObj := resp["error"].(map[string]interface{})
	if errObj["type"] != "test_error" {
		t.Errorf("error type = %q, want test_error", errObj["type"])
	}
	if errObj["message"] != "test message" {
		t.Errorf("error message = %q, want 'test message'", errObj["message"])
	}
}

func TestProfileProxyGetOrCreateProxy(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	providers := []*Provider{
		{Name: "test", Healthy: true},
	}

	// First call creates
	srv1 := pp.getOrCreateProxy("prof1", providers, nil, config.LoadBalanceFailover)
	if srv1 == nil {
		t.Fatal("expected non-nil proxy server")
	}

	// Second call returns cached
	srv2 := pp.getOrCreateProxy("prof1", providers, nil, config.LoadBalanceFailover)
	if srv1 != srv2 {
		t.Error("expected same cached proxy server")
	}

	// Different profile creates new
	srv3 := pp.getOrCreateProxy("prof2", providers, nil, config.LoadBalanceFailover)
	if srv3 == srv1 {
		t.Error("expected different proxy server for different profile")
	}
}

func TestProfileProxyGetOrCreateProxyWithRouting(t *testing.T) {
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	defaultProviders := []*Provider{
		{Name: "standard", Healthy: true},
	}
	thinkProviders := []*Provider{
		{Name: "thinker", Healthy: true},
	}

	routing := &RoutingConfig{
		DefaultProviders: defaultProviders,
		ScenarioRoutes: map[config.Scenario]*ScenarioProviders{
			config.ScenarioThink: {
				Providers: thinkProviders,
			},
		},
	}

	srv := pp.getOrCreateProxy("routed", defaultProviders, routing, config.LoadBalanceFailover)
	if srv == nil {
		t.Fatal("expected non-nil proxy server")
	}
	if srv.Routing == nil {
		t.Fatal("expected proxy server to have routing config")
	}
	if len(srv.Routing.ScenarioRoutes) != 1 {
		t.Errorf("expected 1 scenario route, got %d", len(srv.Routing.ScenarioRoutes))
	}
	if sp, ok := srv.Routing.ScenarioRoutes[config.ScenarioThink]; !ok {
		t.Error("expected think scenario route")
	} else if len(sp.Providers) != 1 || sp.Providers[0].Name != "thinker" {
		t.Error("think scenario should route to thinker provider")
	}
}

// setupTestConfig creates a temp HOME with config dir and resets the store.
func setupTestConfig(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	os.MkdirAll(filepath.Join(tmpDir, ".zen"), 0755)
	config.ResetDefaultStore()
	t.Cleanup(func() { config.ResetDefaultStore() })
}

func TestResolveProfileConfigWithRouting(t *testing.T) {
	setupTestConfig(t)

	// Set up providers
	config.SetProvider("standard", &config.ProviderConfig{
		BaseURL:   "https://api.standard.com",
		AuthToken: "tok-std",
	})
	config.SetProvider("thinker", &config.ProviderConfig{
		BaseURL:   "https://api.thinker.com",
		AuthToken: "tok-think",
	})

	// Set up profile with routing
	config.SetProfileConfig("routed", &config.ProfileConfig{
		Providers: []string{"standard"},
		Routing: map[config.Scenario]*config.ScenarioRoute{
			config.ScenarioThink: {
				Providers: []*config.ProviderRoute{
					{Name: "thinker", Model: "custom-think-model"},
				},
			},
		},
		LongContextThreshold: 50000,
	})

	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	route := &RouteInfo{Profile: "routed", SessionID: "s1", Remainder: "/v1/messages"}
	info, err := pp.resolveProfileConfig(route)
	if err != nil {
		t.Fatalf("resolveProfileConfig() error: %v", err)
	}
	if len(info.providers) != 1 || info.providers[0] != "standard" {
		t.Errorf("default providers = %v, want [standard]", info.providers)
	}
	if info.routing == nil {
		t.Fatal("expected routing config")
	}
	thinkRoute, ok := info.routing[config.ScenarioThink]
	if !ok {
		t.Fatal("expected think scenario route")
	}
	if len(thinkRoute.Providers) != 1 || thinkRoute.Providers[0].Name != "thinker" {
		t.Errorf("think route providers = %v", thinkRoute.Providers)
	}
	if thinkRoute.Providers[0].Model != "custom-think-model" {
		t.Errorf("think route model = %q, want custom-think-model", thinkRoute.Providers[0].Model)
	}
	if info.longContextThreshold != 50000 {
		t.Errorf("longContextThreshold = %d, want 50000", info.longContextThreshold)
	}
}

// T003: Test that buildProviders propagates ProxyURL and creates per-provider Client.
func TestBuildProvidersProxyURL(t *testing.T) {
	setupTestConfig(t)

	// Provider with a SOCKS5 proxy
	config.SetProvider("with-proxy", &config.ProviderConfig{
		BaseURL:   "https://api.example.com",
		AuthToken: "tok1",
		Model:     "claude-sonnet-4-5",
		ProxyURL:  "socks5://proxy.example.com:1080",
	})

	// Provider without proxy
	config.SetProvider("no-proxy", &config.ProviderConfig{
		BaseURL:   "https://api2.example.com",
		AuthToken: "tok2",
		Model:     "claude-sonnet-4-5",
	})

	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)
	providers, err := pp.buildProviders([]string{"with-proxy", "no-proxy"})
	if err != nil {
		t.Fatalf("buildProviders() error: %v", err)
	}
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}

	// Provider with proxy should have ProxyURL set and non-nil Client
	if providers[0].ProxyURL != "socks5://proxy.example.com:1080" {
		t.Errorf("providers[0].ProxyURL = %q, want socks5://proxy.example.com:1080", providers[0].ProxyURL)
	}
	if providers[0].Client == nil {
		t.Error("providers[0].Client should be non-nil for provider with ProxyURL")
	}

	// Provider without proxy should have empty ProxyURL and nil Client
	if providers[1].ProxyURL != "" {
		t.Errorf("providers[1].ProxyURL = %q, want empty", providers[1].ProxyURL)
	}
	if providers[1].Client != nil {
		t.Error("providers[1].Client should be nil for provider without ProxyURL")
	}
}

// T004: Test that buildProviders populates model default fallbacks.
func TestBuildProvidersModelDefaults(t *testing.T) {
	setupTestConfig(t)

	// Provider with Model set but no specific model overrides
	config.SetProvider("defaults", &config.ProviderConfig{
		BaseURL:   "https://api.example.com",
		AuthToken: "tok1",
		Model:     "claude-sonnet-4-5",
	})

	// Provider with all model fields explicitly set
	config.SetProvider("explicit", &config.ProviderConfig{
		BaseURL:        "https://api2.example.com",
		AuthToken:      "tok2",
		Model:          "custom-model",
		ReasoningModel: "custom-reasoning",
		HaikuModel:     "custom-haiku",
		OpusModel:      "custom-opus",
		SonnetModel:    "custom-sonnet",
	})

	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)
	providers, err := pp.buildProviders([]string{"defaults", "explicit"})
	if err != nil {
		t.Fatalf("buildProviders() error: %v", err)
	}

	// Provider with empty model fields should get defaults
	p := providers[0]
	if p.ReasoningModel != "claude-sonnet-4-5-thinking" {
		t.Errorf("ReasoningModel = %q, want claude-sonnet-4-5-thinking", p.ReasoningModel)
	}
	if p.HaikuModel != "claude-haiku-4-5" {
		t.Errorf("HaikuModel = %q, want claude-haiku-4-5", p.HaikuModel)
	}
	if p.OpusModel != "claude-opus-4-5" {
		t.Errorf("OpusModel = %q, want claude-opus-4-5", p.OpusModel)
	}
	if p.SonnetModel != "claude-sonnet-4-5" {
		t.Errorf("SonnetModel = %q, want claude-sonnet-4-5", p.SonnetModel)
	}

	// Provider with explicit model fields should keep them
	p2 := providers[1]
	if p2.ReasoningModel != "custom-reasoning" {
		t.Errorf("ReasoningModel = %q, want custom-reasoning", p2.ReasoningModel)
	}
	if p2.HaikuModel != "custom-haiku" {
		t.Errorf("HaikuModel = %q, want custom-haiku", p2.HaikuModel)
	}
	if p2.OpusModel != "custom-opus" {
		t.Errorf("OpusModel = %q, want custom-opus", p2.OpusModel)
	}
	if p2.SonnetModel != "custom-sonnet" {
		t.Errorf("SonnetModel = %q, want custom-sonnet", p2.SonnetModel)
	}
}

func TestDetectClientFormat(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		clientType string
		want       string
	}{
		{
			name:       "chat completions path",
			path:       "/v1/chat/completions",
			clientType: "",
			want:       "openai-chat",
		},
		{
			name:       "responses api path",
			path:       "/responses",
			clientType: "",
			want:       "openai-responses",
		},
		{
			name:       "responses api with prefix",
			path:       "/v1/responses",
			clientType: "",
			want:       "openai-responses",
		},
		{
			name:       "anthropic messages path",
			path:       "/v1/messages",
			clientType: "",
			want:       "anthropic-messages",
		},
		{
			name:       "codex client type",
			path:       "/v1/messages",
			clientType: "codex",
			want:       "openai-chat",
		},
		{
			name:       "unknown path defaults to anthropic",
			path:       "/unknown",
			clientType: "",
			want:       "anthropic-messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectClientFormat(tt.path, tt.clientType)
			if got != tt.want {
				t.Errorf("detectClientFormat(%q, %q) = %q, want %q", tt.path, tt.clientType, got, tt.want)
			}
		})
	}
}

// Test detectClientFormat with Codex client type
func TestDetectClientFormat_Codex(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		clientType string
		expected   string
	}{
		{
			name:       "codex with responses path",
			path:       "/v1/responses",
			clientType: "codex",
			expected:   "openai-responses",
		},
		{
			name:       "codex with chat completions path",
			path:       "/v1/chat/completions",
			clientType: "codex",
			expected:   "openai-chat",
		},
		{
			name:       "codex with unknown path defaults to chat",
			path:       "/v1/unknown",
			clientType: "codex",
			expected:   "openai-chat",
		},
		{
			name:       "codex with messages path defaults to chat",
			path:       "/v1/messages",
			clientType: "codex",
			expected:   "openai-chat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectClientFormat(tt.path, tt.clientType)
			if result != tt.expected {
				t.Errorf("detectClientFormat(%q, %q) = %q, want %q",
					tt.path, tt.clientType, result, tt.expected)
			}
		})
	}
}

// TestProfileProxyLeastLatencyRouting tests end-to-end profile → strategy → provider selection
// for least-latency strategy (T012 - User Story 1 integration test)
func TestProfileProxyLeastLatencyRouting(t *testing.T) {
	// Setup: Create temp config directory
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	// Setup: Create LogDB with latency metrics
	db, err := OpenLogDB(configDir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	defer db.Close()

	// Insert metrics: provider-a=100ms, provider-b=50ms, provider-c=200ms (all 15 samples)
	for i := 0; i < 15; i++ {
		db.RecordMetric("provider-a", 100, 200, false, false)
		db.RecordMetric("provider-b", 50, 200, false, false)
		db.RecordMetric("provider-c", 200, 200, false, false)
	}

	// Initialize global LoadBalancer with the test DB
	InitGlobalLoadBalancer(db)

	// Setup: Create mock providers
	providerA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"msg_a","type":"message","role":"assistant","content":[{"type":"text","text":"response from A"}],"model":"claude-sonnet-4-5","usage":{"input_tokens":10,"output_tokens":20}}`))
	}))
	defer providerA.Close()

	providerB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"msg_b","type":"message","role":"assistant","content":[{"type":"text","text":"response from B"}],"model":"claude-sonnet-4-5","usage":{"input_tokens":10,"output_tokens":20}}`))
	}))
	defer providerB.Close()

	providerC := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"msg_c","type":"message","role":"assistant","content":[{"type":"text","text":"response from C"}],"model":"claude-sonnet-4-5","usage":{"input_tokens":10,"output_tokens":20}}`))
	}))
	defer providerC.Close()

	// Setup: Configure providers in config store
	config.SetProvider("provider-a", &config.ProviderConfig{
		BaseURL:   providerA.URL,
		AuthToken: "token-a",
		Model:     "claude-sonnet-4-5",
	})
	config.SetProvider("provider-b", &config.ProviderConfig{
		BaseURL:   providerB.URL,
		AuthToken: "token-b",
		Model:     "claude-sonnet-4-5",
	})
	config.SetProvider("provider-c", &config.ProviderConfig{
		BaseURL:   providerC.URL,
		AuthToken: "token-c",
		Model:     "claude-sonnet-4-5",
	})

	// Setup: Create profile with least-latency strategy
	config.SetProfileConfig("test-profile", &config.ProfileConfig{
		Providers: []string{"provider-a", "provider-b", "provider-c"},
		Strategy:  config.LoadBalanceLeastLatency,
	})

	// Create ProfileProxy
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	// Create test request
	reqBody := `{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"test"}],"max_tokens":100}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test-profile/session123/v1/messages", strings.NewReader(reqBody))
	r.Header.Set("Content-Type", "application/json")

	// Execute request
	pp.ServeHTTP(w, r)

	// Verify: Response should be from provider-b (lowest latency: 50ms)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Check that response is from provider-b
	if id, ok := resp["id"].(string); !ok || id != "msg_b" {
		t.Errorf("expected response from provider-b (msg_b), got id=%v", resp["id"])
	}

	content := resp["content"].([]interface{})[0].(map[string]interface{})
	text := content["text"].(string)
	if text != "response from B" {
		t.Errorf("expected 'response from B', got %q", text)
	}
}

// TestProfileProxyLeastCostRouting tests end-to-end profile → strategy → provider selection
// for least-cost strategy (T021 - User Story 2 integration test)
func TestProfileProxyLeastCostRouting(t *testing.T) {
	// Setup: Create temp config directory
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	// Setup: Create LogDB (not needed for cost routing, but required for LoadBalancer)
	db, err := OpenLogDB(configDir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	defer db.Close()

	// Initialize global LoadBalancer
	InitGlobalLoadBalancer(db)

	// Setup: Create mock providers
	// Provider A: Opus (most expensive)
	providerA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"msg_a","type":"message","role":"assistant","content":[{"type":"text","text":"response from A (Opus)"}],"model":"claude-3-opus-20240229","usage":{"input_tokens":10,"output_tokens":20}}`))
	}))
	defer providerA.Close()

	// Provider B: Haiku (cheapest)
	providerB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"msg_b","type":"message","role":"assistant","content":[{"type":"text","text":"response from B (Haiku)"}],"model":"claude-3-5-haiku-20241022","usage":{"input_tokens":10,"output_tokens":20}}`))
	}))
	defer providerB.Close()

	// Provider C: Sonnet (mid-range)
	providerC := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"msg_c","type":"message","role":"assistant","content":[{"type":"text","text":"response from C (Sonnet)"}],"model":"claude-3-5-sonnet-20241022","usage":{"input_tokens":10,"output_tokens":20}}`))
	}))
	defer providerC.Close()

	// Setup: Configure providers with different models (different costs)
	config.SetProvider("provider-opus", &config.ProviderConfig{
		BaseURL:   providerA.URL,
		AuthToken: "token-a",
		Model:     "claude-3-opus-20240229", // Most expensive
	})
	config.SetProvider("provider-haiku", &config.ProviderConfig{
		BaseURL:   providerB.URL,
		AuthToken: "token-b",
		Model:     "claude-3-5-haiku-20241022", // Cheapest
	})
	config.SetProvider("provider-sonnet", &config.ProviderConfig{
		BaseURL:   providerC.URL,
		AuthToken: "token-c",
		Model:     "claude-3-5-sonnet-20241022", // Mid-range
	})

	// Setup: Create profile with least-cost strategy
	config.SetProfileConfig("cost-profile", &config.ProfileConfig{
		Providers: []string{"provider-opus", "provider-haiku", "provider-sonnet"},
		Strategy:  config.LoadBalanceLeastCost,
	})

	// Create ProfileProxy
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	// Create test request
	reqBody := `{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"test"}],"max_tokens":100}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cost-profile/session456/v1/messages", strings.NewReader(reqBody))
	r.Header.Set("Content-Type", "application/json")

	// Execute request
	pp.ServeHTTP(w, r)

	// Verify: Response should be from provider-haiku (lowest cost)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Check that response is from provider-haiku
	if id, ok := resp["id"].(string); !ok || id != "msg_b" {
		t.Errorf("expected response from provider-haiku (msg_b), got id=%v", resp["id"])
	}

	content := resp["content"].([]interface{})[0].(map[string]interface{})
	text := content["text"].(string)
	if text != "response from B (Haiku)" {
		t.Errorf("expected 'response from B (Haiku)', got %q", text)
	}
}

// TestProfileProxyRoundRobinRouting tests end-to-end profile → strategy → provider selection
// for round-robin strategy (T029 - User Story 3 integration test)
func TestProfileProxyRoundRobinRouting(t *testing.T) {
	// Setup: Create temp config directory
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	// Setup: Create LogDB (required for LoadBalancer)
	db, err := OpenLogDB(configDir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	defer db.Close()

	// Initialize global LoadBalancer
	InitGlobalLoadBalancer(db)

	// Setup: Create mock providers that return their name in response
	createMockProvider := func(name string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := fmt.Sprintf(`{"id":"msg_%s","type":"message","role":"assistant","content":[{"type":"text","text":"response from %s"}],"model":"claude-sonnet-4-5","usage":{"input_tokens":10,"output_tokens":20}}`, name, name)
			w.Write([]byte(response))
		}))
	}

	providerA := createMockProvider("A")
	defer providerA.Close()
	providerB := createMockProvider("B")
	defer providerB.Close()
	providerC := createMockProvider("C")
	defer providerC.Close()

	// Setup: Configure providers
	config.SetProvider("provider-a", &config.ProviderConfig{
		BaseURL:   providerA.URL,
		AuthToken: "token-a",
		Model:     "claude-sonnet-4-5",
	})
	config.SetProvider("provider-b", &config.ProviderConfig{
		BaseURL:   providerB.URL,
		AuthToken: "token-b",
		Model:     "claude-sonnet-4-5",
	})
	config.SetProvider("provider-c", &config.ProviderConfig{
		BaseURL:   providerC.URL,
		AuthToken: "token-c",
		Model:     "claude-sonnet-4-5",
	})

	// Setup: Create profile with round-robin strategy
	config.SetProfileConfig("rr-profile", &config.ProfileConfig{
		Providers: []string{"provider-a", "provider-b", "provider-c"},
		Strategy:  config.LoadBalanceRoundRobin,
	})

	// Create ProfileProxy
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	// Make 9 requests and track which provider responds
	reqBody := `{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"test"}],"max_tokens":100}`
	selections := make([]string, 9)

	for i := 0; i < 9; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", fmt.Sprintf("/rr-profile/session%d/v1/messages", i), strings.NewReader(reqBody))
		r.Header.Set("Content-Type", "application/json")

		pp.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d: %s", i, w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("request %d: failed to decode response: %v", i, err)
		}

		// Extract provider name from response ID
		id := resp["id"].(string)
		if strings.HasPrefix(id, "msg_A") {
			selections[i] = "provider-a"
		} else if strings.HasPrefix(id, "msg_B") {
			selections[i] = "provider-b"
		} else if strings.HasPrefix(id, "msg_C") {
			selections[i] = "provider-c"
		} else {
			t.Fatalf("request %d: unexpected response id: %s", i, id)
		}
	}

	// Count selections
	counts := make(map[string]int)
	for _, name := range selections {
		counts[name]++
	}

	// Verify even distribution: each provider should be selected exactly 3 times
	for _, provider := range []string{"provider-a", "provider-b", "provider-c"} {
		if counts[provider] != 3 {
			t.Errorf("provider %s selected %d times, want 3 (selections: %v)", provider, counts[provider], selections)
		}
	}
}

// TestProfileProxyWeightedRouting tests end-to-end profile → strategy → provider selection
// for weighted strategy (T037 - User Story 4 integration test)
func TestProfileProxyWeightedRouting(t *testing.T) {
	// Setup: Create temp config directory
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	// Setup: Create LogDB (required for LoadBalancer)
	db, err := OpenLogDB(configDir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	defer db.Close()

	// Initialize global LoadBalancer
	InitGlobalLoadBalancer(db)

	// Setup: Create mock providers that return their name in response
	createMockProvider := func(name string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := fmt.Sprintf(`{"id":"msg_%s","type":"message","role":"assistant","content":[{"type":"text","text":"response from %s"}],"model":"claude-sonnet-4-5","usage":{"input_tokens":10,"output_tokens":20}}`, name, name)
			w.Write([]byte(response))
		}))
	}

	providerA := createMockProvider("A")
	defer providerA.Close()
	providerB := createMockProvider("B")
	defer providerB.Close()
	providerC := createMockProvider("C")
	defer providerC.Close()

	// Setup: Configure providers with weights (A:70, B:20, C:10)
	config.SetProvider("provider-a", &config.ProviderConfig{
		BaseURL:   providerA.URL,
		AuthToken: "token-a",
		Model:     "claude-sonnet-4-5",
		Weight:    70,
	})
	config.SetProvider("provider-b", &config.ProviderConfig{
		BaseURL:   providerB.URL,
		AuthToken: "token-b",
		Model:     "claude-sonnet-4-5",
		Weight:    20,
	})
	config.SetProvider("provider-c", &config.ProviderConfig{
		BaseURL:   providerC.URL,
		AuthToken: "token-c",
		Model:     "claude-sonnet-4-5",
		Weight:    10,
	})

	// Setup: Create profile with weighted strategy
	config.SetProfileConfig("weighted-profile", &config.ProfileConfig{
		Providers: []string{"provider-a", "provider-b", "provider-c"},
		Strategy:  config.LoadBalanceWeighted,
	})

	// Create ProfileProxy
	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	// Make 100 requests and track which provider responds
	reqBody := `{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"test"}],"max_tokens":100}`
	const numRequests = 100
	selections := make([]string, numRequests)

	for i := 0; i < numRequests; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", fmt.Sprintf("/weighted-profile/session%d/v1/messages", i), strings.NewReader(reqBody))
		r.Header.Set("Content-Type", "application/json")

		pp.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d: %s", i, w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("request %d: failed to decode response: %v", i, err)
		}

		// Extract provider name from response ID
		id := resp["id"].(string)
		if strings.HasPrefix(id, "msg_A") {
			selections[i] = "provider-a"
		} else if strings.HasPrefix(id, "msg_B") {
			selections[i] = "provider-b"
		} else if strings.HasPrefix(id, "msg_C") {
			selections[i] = "provider-c"
		} else {
			t.Fatalf("request %d: unexpected response id: %s", i, id)
		}
	}

	// Count selections
	counts := make(map[string]int)
	for _, name := range selections {
		counts[name]++
	}

	// Verify distribution matches weights within 15% variance
	// Expected: A=70, B=20, C=10
	expectedA := 70
	expectedB := 20
	expectedC := 10
	tolerance := 15 // 15%

	if counts["provider-a"] < expectedA-tolerance || counts["provider-a"] > expectedA+tolerance {
		t.Errorf("provider-a selected %d times, want %d±%d (70%%) (distribution: %v)", counts["provider-a"], expectedA, tolerance, counts)
	}
	if counts["provider-b"] < expectedB-tolerance || counts["provider-b"] > expectedB+tolerance {
		t.Errorf("provider-b selected %d times, want %d±%d (20%%) (distribution: %v)", counts["provider-b"], expectedB, tolerance, counts)
	}
	if counts["provider-c"] < expectedC-tolerance || counts["provider-c"] > expectedC+tolerance {
		t.Errorf("provider-c selected %d times, want %d±%d (10%%) (distribution: %v)", counts["provider-c"], expectedC, tolerance, counts)
	}
}

// T048: Backward compatibility - empty strategy defaults to failover
func TestProfileProxyDefaultFailoverStrategy(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)
	config.ResetDefaultStore()

	db, err := OpenLogDB(configDir)
	if err != nil {
		t.Fatalf("OpenLogDB: %v", err)
	}
	defer db.Close()
	InitGlobalLoadBalancer(db)

	// Provider A returns 200, Provider B returns 200
	providerA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"msg_a","type":"message","role":"assistant","content":[{"type":"text","text":"A"}],"model":"claude-sonnet-4-5","usage":{"input_tokens":10,"output_tokens":20}}`))
	}))
	defer providerA.Close()

	providerB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"msg_b","type":"message","role":"assistant","content":[{"type":"text","text":"B"}],"model":"claude-sonnet-4-5","usage":{"input_tokens":10,"output_tokens":20}}`))
	}))
	defer providerB.Close()

	config.SetProvider("pa", &config.ProviderConfig{
		BaseURL: providerA.URL, AuthToken: "t", Model: "claude-sonnet-4-5",
	})
	config.SetProvider("pb", &config.ProviderConfig{
		BaseURL: providerB.URL, AuthToken: "t", Model: "claude-sonnet-4-5",
	})

	// Profile with NO strategy set (empty string = default failover)
	config.SetProfileConfig("default-profile", &config.ProfileConfig{
		Providers: []string{"pa", "pb"},
		// Strategy intentionally omitted
	})

	logger := log.New(os.Stderr, "[test] ", 0)
	pp := NewProfileProxy(logger)

	// All requests should go to provider A (failover = first healthy)
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST",
			fmt.Sprintf("/default-profile/s%d/v1/messages", i),
			strings.NewReader(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"hi"}],"max_tokens":10}`))
		r.Header.Set("Content-Type", "application/json")
		pp.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("req %d: got %d", i, w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)
		if resp["id"] != "msg_a" {
			t.Errorf("req %d: expected failover to provider-a (msg_a), got %v", i, resp["id"])
		}
	}
}
