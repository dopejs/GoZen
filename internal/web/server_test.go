package web

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

func setupTestServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	config.ResetDefaultStore()
	t.Cleanup(func() { config.ResetDefaultStore() })

	// Create config dir and file
	configDir := filepath.Join(dir, config.ConfigDir)
	os.MkdirAll(configDir, 0755)
	cfg := &config.OpenCCConfig{
		Providers: map[string]*config.ProviderConfig{
			"test-provider": {
				BaseURL:   "https://api.test.com",
				AuthToken: "sk-test-secret-token-1234",
				Model:     "claude-sonnet-4-5",
			},
			"backup": {
				BaseURL:   "https://api.backup.com",
				AuthToken: "sk-backup-token-5678",
			},
		},
		Profiles: map[string]*config.ProfileConfig{
			"default": {Providers: []string{"test-provider", "backup"}},
			"work":    {Providers: []string{"test-provider"}},
		},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(filepath.Join(configDir, config.ConfigFile), data, 0600)

	// Force reload
	config.DefaultStore()

	logger := log.New(io.Discard, "", 0)
	return NewServer("1.0.0-test", logger, 0)
}

func doRequest(s *Server, method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewReader(data)
	}
	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(w, req)
	return w
}

func doRequestRaw(s *Server, method, path string, body []byte) *httptest.ResponseRecorder {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(w, req)
	return w
}

func decodeJSON(t *testing.T, r *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}
}

// --- Health ---

func TestHealthEndpoint(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/health", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	decodeJSON(t, w, &resp)
	if resp["status"] != "ok" {
		t.Errorf("status = %q", resp["status"])
	}
	if resp["version"] != "1.0.0-test" {
		t.Errorf("version = %q", resp["version"])
	}
}

func TestHealthMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/health", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// --- Security Headers ---

func TestSecurityHeaders(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/health", nil)

	if v := w.Header().Get("X-Content-Type-Options"); v != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q", v)
	}
	if v := w.Header().Get("X-Frame-Options"); v != "DENY" {
		t.Errorf("X-Frame-Options = %q", v)
	}
}

// --- Providers ---

func TestListProviders(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/providers", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var providers []providerResponse
	decodeJSON(t, w, &providers)
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}

	// Tokens should be masked in list responses
	for _, p := range providers {
		if p.AuthToken == "sk-test-secret-token-1234" || p.AuthToken == "sk-backup-token-5678" {
			t.Errorf("token for %s should be masked, got %q", p.Name, p.AuthToken)
		}
	}
}

func TestGetProvider(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/providers/test-provider", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var p providerResponse
	decodeJSON(t, w, &p)
	if p.Name != "test-provider" {
		t.Errorf("name = %q", p.Name)
	}
	if p.BaseURL != "https://api.test.com" {
		t.Errorf("base_url = %q", p.BaseURL)
	}
	// Token should be masked in get response
	if p.AuthToken == "sk-test-secret-token-1234" {
		t.Errorf("token should be masked in get response, got %q", p.AuthToken)
	}
}

func TestGetProviderNotFound(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/providers/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestCreateProvider(t *testing.T) {
	s := setupTestServer(t)

	body := createProviderRequest{
		Name: "new-provider",
		Config: config.ProviderConfig{
			BaseURL:   "https://api.new.com",
			AuthToken: "sk-new-token",
			Model:     "claude-opus-4-5",
		},
	}
	w := doRequest(s, "POST", "/api/v1/providers", body)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Verify it's persisted
	w2 := doRequest(s, "GET", "/api/v1/providers/new-provider", nil)
	if w2.Code != http.StatusOK {
		t.Fatalf("created provider not found")
	}
}

func TestCreateProviderConflict(t *testing.T) {
	s := setupTestServer(t)

	body := createProviderRequest{
		Name: "test-provider",
		Config: config.ProviderConfig{
			BaseURL:   "https://dup.com",
			AuthToken: "tok",
		},
	}
	w := doRequest(s, "POST", "/api/v1/providers", body)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestCreateProviderNoName(t *testing.T) {
	s := setupTestServer(t)
	body := createProviderRequest{
		Config: config.ProviderConfig{BaseURL: "https://x.com", AuthToken: "tok"},
	}
	w := doRequest(s, "POST", "/api/v1/providers", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateProvider(t *testing.T) {
	s := setupTestServer(t)

	update := config.ProviderConfig{
		BaseURL: "https://api.updated.com",
		Model:   "claude-opus-4-5",
	}
	w := doRequest(s, "PUT", "/api/v1/providers/test-provider", update)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp providerResponse
	decodeJSON(t, w, &resp)
	if resp.BaseURL != "https://api.updated.com" {
		t.Errorf("base_url = %q", resp.BaseURL)
	}
	if resp.Model != "claude-opus-4-5" {
		t.Errorf("model = %q", resp.Model)
	}
}

func TestUpdateProviderKeepsToken(t *testing.T) {
	s := setupTestServer(t)

	// Send empty token - should keep original
	update := config.ProviderConfig{
		BaseURL: "https://api.updated.com",
	}
	w := doRequest(s, "PUT", "/api/v1/providers/test-provider", update)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verify token is still there by checking the store directly
	p := config.DefaultStore().GetProvider("test-provider")
	if p.AuthToken != "sk-test-secret-token-1234" {
		t.Errorf("token was changed, got %q", p.AuthToken)
	}
}

func TestUpdateProviderNotFound(t *testing.T) {
	s := setupTestServer(t)
	update := config.ProviderConfig{BaseURL: "https://x.com"}
	w := doRequest(s, "PUT", "/api/v1/providers/nonexistent", update)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDeleteProvider(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/providers/backup", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verify deleted
	w2 := doRequest(s, "GET", "/api/v1/providers/backup", nil)
	if w2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", w2.Code)
	}

	// Verify cascade: backup should be removed from default profile
	order := config.DefaultStore().GetProfileOrder("default")
	for _, n := range order {
		if n == "backup" {
			t.Error("backup should have been removed from default profile")
		}
	}
}

func TestDeleteProviderNotFound(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/providers/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- Profiles ---

func TestListProfiles(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/profiles", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var profiles []profileResponse
	decodeJSON(t, w, &profiles)
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}
}

func TestGetProfile(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/profiles/default", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var p profileResponse
	decodeJSON(t, w, &p)
	if p.Name != "default" {
		t.Errorf("name = %q", p.Name)
	}
	if len(p.Providers) != 2 {
		t.Errorf("expected 2 providers in default profile, got %d", len(p.Providers))
	}
}

func TestGetProfileNotFound(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/profiles/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestCreateProfile(t *testing.T) {
	s := setupTestServer(t)

	body := createProfileRequest{
		Name:      "staging",
		Providers: []string{"backup", "test-provider"},
	}
	w := doRequest(s, "POST", "/api/v1/profiles", body)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Verify
	w2 := doRequest(s, "GET", "/api/v1/profiles/staging", nil)
	if w2.Code != http.StatusOK {
		t.Fatalf("created profile not found")
	}
	var p profileResponse
	decodeJSON(t, w2, &p)
	if len(p.Providers) != 2 || p.Providers[0] != "backup" {
		t.Errorf("providers = %v", p.Providers)
	}
}

func TestCreateProfileConflict(t *testing.T) {
	s := setupTestServer(t)
	body := createProfileRequest{
		Name:      "default",
		Providers: []string{"test-provider"},
	}
	w := doRequest(s, "POST", "/api/v1/profiles", body)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestCreateProfileNoName(t *testing.T) {
	s := setupTestServer(t)
	body := createProfileRequest{Providers: []string{"test-provider"}}
	w := doRequest(s, "POST", "/api/v1/profiles", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateProfile(t *testing.T) {
	s := setupTestServer(t)

	body := updateProfileRequest{Providers: []string{"backup"}}
	w := doRequest(s, "PUT", "/api/v1/profiles/work", body)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var p profileResponse
	decodeJSON(t, w, &p)
	if len(p.Providers) != 1 || p.Providers[0] != "backup" {
		t.Errorf("providers = %v", p.Providers)
	}
}

func TestUpdateProfileNotFound(t *testing.T) {
	s := setupTestServer(t)
	body := updateProfileRequest{Providers: []string{"test-provider"}}
	w := doRequest(s, "PUT", "/api/v1/profiles/nonexistent", body)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDeleteProfile(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/profiles/work", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verify deleted
	w2 := doRequest(s, "GET", "/api/v1/profiles/work", nil)
	if w2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", w2.Code)
	}
}

func TestDeleteProfileDefault(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/profiles/default", nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestDeleteProfileNotFound(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/profiles/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- Reload ---

func TestReload(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/reload", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	decodeJSON(t, w, &resp)
	if resp["status"] != "reloaded" {
		t.Errorf("status = %q", resp["status"])
	}
}

func TestReloadMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/reload", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// --- Token masking ---

func TestMaskToken(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"sk-test-secret-token-1234", "sk-te...1234"},
		{"short", "****"},
		{"12345678", "****"},
		{"123456789", "12345...6789"},
	}
	for _, tt := range tests {
		got := maskToken(tt.input)
		if got != tt.want {
			t.Errorf("maskToken(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- Profile Routing ---

func TestCreateProfileWithRouting(t *testing.T) {
	s := setupTestServer(t)

	body := createProfileRequest{
		Name:      "routed",
		Providers: []string{"test-provider", "backup"},
		Routing: map[config.Scenario]*scenarioRouteResponse{
			config.ScenarioThink: {
				Providers: []*providerRouteResponse{
					{Name: "backup", Model: "claude-opus-4-5"},
					{Name: "test-provider"},
				},
			},
			config.ScenarioImage: {
				Providers: []*providerRouteResponse{
					{Name: "test-provider"},
				},
			},
		},
	}
	w := doRequest(s, "POST", "/api/v1/profiles", body)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Verify routing is returned
	var resp profileResponse
	decodeJSON(t, w, &resp)
	if resp.Routing == nil {
		t.Fatal("routing should not be nil in response")
	}
	if len(resp.Routing) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(resp.Routing))
	}

	thinkRoute := resp.Routing[config.ScenarioThink]
	if thinkRoute == nil {
		t.Fatal("think route should exist")
	}
	if len(thinkRoute.Providers) != 2 || thinkRoute.Providers[0].Name != "backup" {
		t.Errorf("think providers = %v", thinkRoute.Providers)
	}
	if thinkRoute.Providers[0].Model != "claude-opus-4-5" {
		t.Errorf("think model = %q", thinkRoute.Providers[0].Model)
	}

	// Verify persisted via GET
	w2 := doRequest(s, "GET", "/api/v1/profiles/routed", nil)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
	var got profileResponse
	decodeJSON(t, w2, &got)
	if got.Routing == nil || len(got.Routing) != 2 {
		t.Errorf("routing not persisted: %v", got.Routing)
	}
}

func TestUpdateProfileWithRouting(t *testing.T) {
	s := setupTestServer(t)

	// Update work profile to add routing
	body := updateProfileRequest{
		Providers: []string{"test-provider"},
		Routing: map[config.Scenario]*scenarioRouteResponse{
			config.ScenarioLongContext: {
				Providers: []*providerRouteResponse{
					{Name: "backup", Model: "claude-haiku-4-5"},
				},
			},
		},
	}
	w := doRequest(s, "PUT", "/api/v1/profiles/work", body)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp profileResponse
	decodeJSON(t, w, &resp)
	if resp.Routing == nil {
		t.Fatal("routing should not be nil")
	}
	lcRoute := resp.Routing[config.ScenarioLongContext]
	if lcRoute == nil {
		t.Fatal("longContext route should exist")
	}
	if len(lcRoute.Providers) != 1 || lcRoute.Providers[0].Name != "backup" {
		t.Errorf("providers = %v", lcRoute.Providers)
	}
	if lcRoute.Providers[0].Model != "claude-haiku-4-5" {
		t.Errorf("model = %q", lcRoute.Providers[0].Model)
	}
}

func TestUpdateProfileClearRouting(t *testing.T) {
	s := setupTestServer(t)

	// First add routing
	body1 := updateProfileRequest{
		Providers: []string{"test-provider"},
		Routing: map[config.Scenario]*scenarioRouteResponse{
			config.ScenarioThink: {Providers: []*providerRouteResponse{{Name: "backup"}}},
		},
	}
	doRequest(s, "PUT", "/api/v1/profiles/work", body1)

	// Then update without routing — should clear it
	body2 := updateProfileRequest{
		Providers: []string{"test-provider", "backup"},
	}
	w := doRequest(s, "PUT", "/api/v1/profiles/work", body2)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp profileResponse
	decodeJSON(t, w, &resp)
	if resp.Routing != nil {
		t.Errorf("routing should be nil after clearing, got %v", resp.Routing)
	}
}

func TestListProfilesWithRouting(t *testing.T) {
	s := setupTestServer(t)

	// Add routing to default
	body := updateProfileRequest{
		Providers: []string{"test-provider", "backup"},
		Routing: map[config.Scenario]*scenarioRouteResponse{
			config.ScenarioThink: {Providers: []*providerRouteResponse{{Name: "backup", Model: "opus"}}},
		},
	}
	doRequest(s, "PUT", "/api/v1/profiles/default", body)

	// List profiles
	w := doRequest(s, "GET", "/api/v1/profiles", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var profiles []profileResponse
	decodeJSON(t, w, &profiles)

	found := false
	for _, p := range profiles {
		if p.Name == "default" {
			found = true
			if p.Routing == nil || len(p.Routing) != 1 {
				t.Errorf("default profile routing not returned in list: %v", p.Routing)
			}
		}
	}
	if !found {
		t.Error("default profile not found in list")
	}
}

func TestCreateProfileWithEmptyRouting(t *testing.T) {
	s := setupTestServer(t)

	// Empty routing providers should be ignored
	body := createProfileRequest{
		Name:      "empty-routes",
		Providers: []string{"test-provider"},
		Routing: map[config.Scenario]*scenarioRouteResponse{
			config.ScenarioThink: {Providers: []*providerRouteResponse{}},
		},
	}
	w := doRequest(s, "POST", "/api/v1/profiles", body)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp profileResponse
	decodeJSON(t, w, &resp)
	// Empty providers route should be filtered out
	if resp.Routing != nil {
		t.Errorf("routing with empty providers should be nil, got %v", resp.Routing)
	}
}

func TestCreateProviderWithEnvVars(t *testing.T) {
	s := setupTestServer(t)

	body := createProviderRequest{
		Name: "env-provider",
		Config: config.ProviderConfig{
			BaseURL:   "https://api.example.com",
			AuthToken: "sk-test-token",
			Model:     "claude-sonnet-4-5",
			EnvVars: map[string]string{
				"CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000",
				"MAX_THINKING_TOKENS":            "50000",
				"MY_CUSTOM_VAR":                  "custom_value",
			},
		},
	}
	w := doRequest(s, "POST", "/api/v1/providers", body)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp providerResponse
	decodeJSON(t, w, &resp)
	if resp.Name != "env-provider" {
		t.Errorf("name = %q, want env-provider", resp.Name)
	}
	if len(resp.EnvVars) != 3 {
		t.Errorf("expected 3 env vars, got %d", len(resp.EnvVars))
	}
	if resp.EnvVars["CLAUDE_CODE_MAX_OUTPUT_TOKENS"] != "64000" {
		t.Errorf("CLAUDE_CODE_MAX_OUTPUT_TOKENS = %q", resp.EnvVars["CLAUDE_CODE_MAX_OUTPUT_TOKENS"])
	}
	if resp.EnvVars["MY_CUSTOM_VAR"] != "custom_value" {
		t.Errorf("MY_CUSTOM_VAR = %q", resp.EnvVars["MY_CUSTOM_VAR"])
	}
}

func TestUpdateProviderWithEnvVars(t *testing.T) {
	s := setupTestServer(t)

	// First create a provider
	createBody := createProviderRequest{
		Name: "update-env-provider",
		Config: config.ProviderConfig{
			BaseURL:   "https://api.example.com",
			AuthToken: "sk-test-token",
			EnvVars: map[string]string{
				"VAR1": "value1",
			},
		},
	}
	doRequest(s, "POST", "/api/v1/providers", createBody)

	// Update with new env vars
	updateBody := config.ProviderConfig{
		BaseURL:   "https://api.example.com",
		AuthToken: "sk-test-token",
		EnvVars: map[string]string{
			"VAR1": "updated_value1",
			"VAR2": "value2",
		},
	}
	w := doRequest(s, "PUT", "/api/v1/providers/update-env-provider", updateBody)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp providerResponse
	decodeJSON(t, w, &resp)
	if len(resp.EnvVars) != 2 {
		t.Errorf("expected 2 env vars, got %d", len(resp.EnvVars))
	}
	if resp.EnvVars["VAR1"] != "updated_value1" {
		t.Errorf("VAR1 = %q, want updated_value1", resp.EnvVars["VAR1"])
	}
	if resp.EnvVars["VAR2"] != "value2" {
		t.Errorf("VAR2 = %q, want value2", resp.EnvVars["VAR2"])
	}
}

// --- Bindings ---

func TestListBindingsEmpty(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/bindings", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp bindingsResponse
	decodeJSON(t, w, &resp)
	if len(resp.Bindings) != 0 {
		t.Errorf("expected 0 bindings, got %d", len(resp.Bindings))
	}
	if len(resp.Profiles) == 0 {
		t.Error("expected profiles list")
	}
}

func TestCreateAndGetBinding(t *testing.T) {
	s := setupTestServer(t)

	body := bindingRequest{Path: "/tmp/test-project", Profile: "default"}
	w := doRequest(s, "POST", "/api/v1/bindings", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// List should have 1
	w2 := doRequest(s, "GET", "/api/v1/bindings", nil)
	var resp bindingsResponse
	decodeJSON(t, w2, &resp)
	if len(resp.Bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(resp.Bindings))
	}

	// Get specific
	w3 := doRequest(s, "GET", "/api/v1/bindings/%2Ftmp%2Ftest-project", nil)
	if w3.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w3.Code)
	}
	var br bindingResponse
	decodeJSON(t, w3, &br)
	if br.Profile != "default" {
		t.Errorf("profile = %q", br.Profile)
	}
}

func TestCreateBindingMissingPath(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/bindings", bindingRequest{Profile: "default"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateBindingInvalidProfile(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/bindings", bindingRequest{Path: "/tmp/x", Profile: "nonexistent"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateBindingInvalidCLI(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/bindings", bindingRequest{Path: "/tmp/x", Client: "invalid-cli"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateBinding(t *testing.T) {
	s := setupTestServer(t)

	// Create
	doRequest(s, "POST", "/api/v1/bindings", bindingRequest{Path: "/tmp/proj", Profile: "default"})

	// Update
	w := doRequest(s, "PUT", "/api/v1/bindings/%2Ftmp%2Fproj", bindingRequest{Profile: "work"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var br bindingResponse
	decodeJSON(t, w, &br)
	if br.Profile != "work" {
		t.Errorf("profile = %q, want work", br.Profile)
	}
}

func TestDeleteBinding(t *testing.T) {
	s := setupTestServer(t)

	doRequest(s, "POST", "/api/v1/bindings", bindingRequest{Path: "/tmp/del", Profile: "default"})

	w := doRequest(s, "DELETE", "/api/v1/bindings/%2Ftmp%2Fdel", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Should be gone
	w2 := doRequest(s, "GET", "/api/v1/bindings/%2Ftmp%2Fdel", nil)
	if w2.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w2.Code)
	}
}

func TestDeleteBindingNotFound(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/bindings/%2Ftmp%2Fno", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetBindingNotFound(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/bindings/%2Ftmp%2Fno", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- Settings ---

func TestGetSettings(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/settings", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp settingsResponse
	decodeJSON(t, w, &resp)
	if len(resp.Clients) == 0 {
		t.Error("expected CLIs list")
	}
}

func TestUpdateSettings(t *testing.T) {
	s := setupTestServer(t)

	body := settingsRequest{DefaultProfile: "work", DefaultClient: "claude", WebPort: 8080}
	w := doRequest(s, "PUT", "/api/v1/settings", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp settingsResponse
	decodeJSON(t, w, &resp)
	if resp.DefaultProfile != "work" {
		t.Errorf("default_profile = %q", resp.DefaultProfile)
	}
	if resp.DefaultClient != "claude" {
		t.Errorf("default_cli = %q", resp.DefaultClient)
	}
	if resp.WebPort != 8080 {
		t.Errorf("web_port = %d", resp.WebPort)
	}
}

func TestUpdateSettingsInvalidProfile(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "PUT", "/api/v1/settings", settingsRequest{DefaultProfile: "nonexistent"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateSettingsInvalidCLI(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "PUT", "/api/v1/settings", settingsRequest{DefaultClient: "bad-cli"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateSettingsInvalidPort(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "PUT", "/api/v1/settings", settingsRequest{WebPort: 80})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSettingsMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/settings", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestBindingsMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "PATCH", "/api/v1/bindings", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestBindingMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "PATCH", "/api/v1/bindings/%2Ftmp%2Fx", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestUpdateBindingInvalidProfile(t *testing.T) {
	s := setupTestServer(t)
	doRequest(s, "POST", "/api/v1/bindings", bindingRequest{Path: "/tmp/up", Profile: "default"})

	w := doRequest(s, "PUT", "/api/v1/bindings/%2Ftmp%2Fup", bindingRequest{Profile: "nonexistent"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateBindingInvalidCLI(t *testing.T) {
	s := setupTestServer(t)
	doRequest(s, "POST", "/api/v1/bindings", bindingRequest{Path: "/tmp/up2", Profile: "default"})

	w := doRequest(s, "PUT", "/api/v1/bindings/%2Ftmp%2Fup2", bindingRequest{Client: "bad"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProviderMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "PATCH", "/api/v1/providers/test-provider", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestProvidersMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "PATCH", "/api/v1/providers", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestProfileMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "PATCH", "/api/v1/profiles/default", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestProfilesMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "PATCH", "/api/v1/profiles", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestReloadEndpoint(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/reload", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestLogsEndpoint(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/logs", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestLogsEndpointWithFilters(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/logs?provider=test&errors_only=true&level=error", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestLogsMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/logs", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestLogsEndpointAllFilters(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/logs?provider=test&errors_only=true&level=error&status_code=500&status_min=400&status_max=599&limit=50", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestLogsEndpointInvalidFilterValues(t *testing.T) {
	s := setupTestServer(t)
	// Invalid numeric values should be ignored, not cause errors
	w := doRequest(s, "GET", "/api/v1/logs?status_code=abc&status_min=xyz&status_max=&limit=-1", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCreateProviderWithAddToProfiles(t *testing.T) {
	s := setupTestServer(t)

	body := createProviderRequest{
		Name: "profile-provider",
		Config: config.ProviderConfig{
			BaseURL:   "https://api.profile.com",
			AuthToken: "sk-profile-token",
		},
		AddToProfiles: []string{"default", "work"},
	}
	w := doRequest(s, "POST", "/api/v1/providers", body)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Verify provider was added to profiles
	order := config.DefaultStore().GetProfileOrder("default")
	found := false
	for _, n := range order {
		if n == "profile-provider" {
			found = true
			break
		}
	}
	if !found {
		t.Error("profile-provider should have been added to default profile")
	}
}

func TestCreateProviderInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest("POST", "/api/v1/providers", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateProviderInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest("PUT", "/api/v1/providers/test-provider", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateProfileInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest("POST", "/api/v1/profiles", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateProfileInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest("PUT", "/api/v1/profiles/work", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateBindingInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest("POST", "/api/v1/bindings", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateSettingsInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest("PUT", "/api/v1/settings", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDeleteProfileNotFoundName(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/profiles/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUpdateBindingInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	doRequest(s, "POST", "/api/v1/bindings", bindingRequest{Path: "/tmp/upj", Profile: "default"})

	req := httptest.NewRequest("PUT", "/api/v1/bindings/%2Ftmp%2Fupj", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- Sync API ---

func TestSyncConfigGetNotConfigured(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/sync/config", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	decodeJSON(t, w, &resp)
	if resp["configured"] != false {
		t.Errorf("expected configured=false, got %v", resp["configured"])
	}
}

func TestSyncConfigPutAndGet(t *testing.T) {
	s := setupTestServer(t)

	body := config.SyncConfig{
		Backend:    "webdav",
		Endpoint:   "https://dav.example.com/zen-sync.json",
		Username:   "user",
		Token:      "pass123456789",
		Passphrase: "my-secret",
		AutoPull:   true,
		PullInterval: 300,
	}
	w := doRequest(s, "PUT", "/api/v1/sync/config", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// GET should return masked tokens
	w2 := doRequest(s, "GET", "/api/v1/sync/config", nil)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
	var resp map[string]interface{}
	decodeJSON(t, w2, &resp)
	if resp["configured"] != true {
		t.Error("expected configured=true")
	}
	cfg := resp["config"].(map[string]interface{})
	if cfg["backend"] != "webdav" {
		t.Errorf("backend = %v", cfg["backend"])
	}
	// Token should be masked
	if cfg["token"] == "pass123456789" {
		t.Error("token should be masked")
	}
	if cfg["passphrase"] != "********" {
		t.Errorf("passphrase should be masked, got %v", cfg["passphrase"])
	}
}

func TestSyncConfigMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "DELETE", "/api/v1/sync/config", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestSyncStatusNotConfigured(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/sync/status", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	decodeJSON(t, w, &resp)
	if resp["configured"] != false {
		t.Errorf("expected configured=false, got %v", resp["configured"])
	}
}

func TestSyncStatusMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/sync/status", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestSyncPullNotConfigured(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/sync/pull", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSyncPullMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/sync/pull", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestSyncPushNotConfigured(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/sync/push", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSyncPushMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/sync/push", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestSyncTestNotConfigured(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/sync/test", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSyncTestMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/sync/test", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestSyncCreateGistNoToken(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "POST", "/api/v1/sync/create-gist", map[string]string{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSyncCreateGistMethodNotAllowed(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/sync/create-gist", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestSyncConfigPutPreservesSecrets(t *testing.T) {
	s := setupTestServer(t)

	// First save with real token
	body1 := config.SyncConfig{
		Backend:    "gist",
		Token:      "ghp_realtoken12345",
		GistID:     "abc123",
		Passphrase: "secret-pass",
	}
	doRequest(s, "PUT", "/api/v1/sync/config", body1)

	// Now save with masked token (simulating UI sending back masked values)
	body2 := config.SyncConfig{
		Backend:    "gist",
		Token:      maskToken("ghp_realtoken12345"),
		GistID:     "abc123",
		Passphrase: "********",
	}
	w := doRequest(s, "PUT", "/api/v1/sync/config", body2)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verify the real token was preserved
	stored := config.GetSyncConfig()
	if stored.Token != "ghp_realtoken12345" {
		t.Errorf("token should be preserved, got %q", stored.Token)
	}
	if stored.Passphrase != "secret-pass" {
		t.Errorf("passphrase should be preserved, got %q", stored.Passphrase)
	}
}

func TestSyncConfigPutInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest("PUT", "/api/v1/sync/config", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func setupSyncTestServer(t *testing.T) *Server {
	t.Helper()
	s := setupTestServer(t)
	// Configure sync with webdav backend
	cfg := config.SyncConfig{
		Backend:  "webdav",
		Endpoint: "http://127.0.0.1:1/nonexistent",
		Username: "user",
		Token:    "pass",
	}
	config.SetSyncConfig(&cfg)
	return s
}

func TestSyncStatusConfigured(t *testing.T) {
	s := setupSyncTestServer(t)
	w := doRequest(s, "GET", "/api/v1/sync/status", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	decodeJSON(t, w, &resp)
	if resp["configured"] != true {
		t.Errorf("expected configured=true, got %v", resp["configured"])
	}
	if resp["backend"] != "webdav" {
		t.Errorf("expected webdav, got %v", resp["backend"])
	}
}

func TestSyncPullConfiguredFails(t *testing.T) {
	s := setupSyncTestServer(t)
	w := doRequest(s, "POST", "/api/v1/sync/pull", nil)
	// Will fail because webdav endpoint is unreachable, but exercises the code path
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestSyncPushConfiguredFails(t *testing.T) {
	s := setupSyncTestServer(t)
	w := doRequest(s, "POST", "/api/v1/sync/push", nil)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestSyncTestWithBody(t *testing.T) {
	s := setupTestServer(t)
	body := config.SyncConfig{
		Backend:  "webdav",
		Endpoint: "http://127.0.0.1:1/nonexistent",
	}
	w := doRequest(s, "POST", "/api/v1/sync/test", body)
	// Connection will fail but exercises the with-body code path
	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", w.Code)
	}
}

func TestSyncTestConfiguredFails(t *testing.T) {
	s := setupSyncTestServer(t)
	w := doRequest(s, "POST", "/api/v1/sync/test", nil)
	// Falls through to getSyncManager path, connection fails
	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", w.Code)
	}
}

func TestSyncCreateGistInvalidJSON(t *testing.T) {
	s := setupTestServer(t)
	req := httptest.NewRequest("POST", "/api/v1/sync/create-gist", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSyncCreateGistWithToken(t *testing.T) {
	s := setupTestServer(t)
	// Token is provided but GitHub API will reject it — exercises the code path
	w := doRequest(s, "POST", "/api/v1/sync/create-gist", map[string]string{"token": "ghp_fake_token"})
	// Will get 502 because GitHub rejects the token
	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", w.Code, w.Body.String())
	}
}

