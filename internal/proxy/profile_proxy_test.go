package proxy

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

	// Manually populate cache
	pp.cache["test"] = &ProxyServer{}
	if len(pp.cache) != 1 {
		t.Fatal("cache should have 1 entry")
	}

	pp.InvalidateCache()
	if len(pp.cache) != 0 {
		t.Error("cache should be empty after InvalidateCache")
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
	srv1 := pp.getOrCreateProxy("prof1", providers, nil)
	if srv1 == nil {
		t.Fatal("expected non-nil proxy server")
	}

	// Second call returns cached
	srv2 := pp.getOrCreateProxy("prof1", providers, nil)
	if srv1 != srv2 {
		t.Error("expected same cached proxy server")
	}

	// Different profile creates new
	srv3 := pp.getOrCreateProxy("prof2", providers, nil)
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

	srv := pp.getOrCreateProxy("routed", defaultProviders, routing)
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
