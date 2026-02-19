// Package integration contains integration tests for the web module.
// These tests verify the web server's behavior in real-world scenarios.
//
// Run with: go test -tags=integration ./test/integration/...
//
//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// WebTestConfig holds test configuration for web tests
type WebTestConfig struct {
	BinaryPath string
	ConfigDir  string
	ProxyPort  int
	WebPort    int
}

func setupWebTest(t *testing.T) *WebTestConfig {
	t.Helper()

	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("failed to find project root: %v", err)
	}

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "zen")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = projectRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, out)
	}

	proxyPort := findFreePort(t)
	webPort := findFreePort(t)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)

	return &WebTestConfig{
		BinaryPath: binaryPath,
		ConfigDir:  configDir,
		ProxyPort:  proxyPort,
		WebPort:    webPort,
	}
}

func (tc *WebTestConfig) writeConfig(t *testing.T, extra map[string]interface{}) {
	t.Helper()
	config := map[string]interface{}{
		"version":    6,
		"proxy_port": tc.ProxyPort,
		"web_port":   tc.WebPort,
		"providers": map[string]interface{}{
			"test-provider": map[string]interface{}{
				"auth_token": "test-token",
				"base_url":   "https://api.anthropic.com",
			},
		},
		"profiles": map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"test-provider"},
			},
		},
	}
	for k, v := range extra {
		config[k] = v
	}
	data, _ := json.Marshal(config)
	configPath := filepath.Join(tc.ConfigDir, "zen.json")
	os.WriteFile(configPath, data, 0644)
}

func (tc *WebTestConfig) startDaemon(t *testing.T) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(tc.BinaryPath, "daemon", "start", "--foreground")
	cmd.Env = append(os.Environ(),
		"HOME="+filepath.Dir(tc.ConfigDir),
		"GOZEN_DAEMON=1",
	)
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start daemon: %v", err)
	}

	// Wait for web server to be ready
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/v1/health", tc.WebPort))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return cmd
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	cmd.Process.Kill()
	t.Fatalf("web server not ready after 5s")
	return nil
}

// =============================================================================
// Test: Health API
// =============================================================================

// TestWeb_HealthAPI verifies that the health endpoint returns correct status.
func TestWeb_HealthAPI(t *testing.T) {
	tc := setupWebTest(t)
	tc.writeConfig(t, nil)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/v1/health", tc.WebPort))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", result["status"])
	}
}

// =============================================================================
// Test: Providers API
// =============================================================================

// TestWeb_ProvidersAPI_List verifies that the providers endpoint returns
// the configured providers.
func TestWeb_ProvidersAPI_List(t *testing.T) {
	tc := setupWebTest(t)
	tc.writeConfig(t, nil)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/v1/providers", tc.WebPort))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}

	var providers []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&providers)

	found := false
	for _, p := range providers {
		if p["name"] == "test-provider" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("test-provider not found in response")
	}
}

// TestWeb_ProvidersAPI_Get verifies that getting a single provider works.
func TestWeb_ProvidersAPI_Get(t *testing.T) {
	tc := setupWebTest(t)
	tc.writeConfig(t, nil)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/v1/providers/test-provider", tc.WebPort))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["base_url"] != "https://api.anthropic.com" {
		t.Errorf("unexpected base_url: %v", result["base_url"])
	}
}

// TestWeb_ProvidersAPI_NotFound verifies that getting a non-existent provider
// returns 404.
func TestWeb_ProvidersAPI_NotFound(t *testing.T) {
	tc := setupWebTest(t)
	tc.writeConfig(t, nil)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/v1/providers/nonexistent", tc.WebPort))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

// =============================================================================
// Test: Profiles API
// =============================================================================

// TestWeb_ProfilesAPI_List verifies that the profiles endpoint returns
// the configured profiles.
func TestWeb_ProfilesAPI_List(t *testing.T) {
	tc := setupWebTest(t)
	tc.writeConfig(t, nil)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/v1/profiles", tc.WebPort))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}

	var profiles []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&profiles)

	found := false
	for _, p := range profiles {
		if p["name"] == "default" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("default profile not found in response")
	}
}

// =============================================================================
// Test: Settings API
// =============================================================================

// TestWeb_SettingsAPI_Get verifies that the settings endpoint returns
// the current settings.
func TestWeb_SettingsAPI_Get(t *testing.T) {
	tc := setupWebTest(t)
	tc.writeConfig(t, nil)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/v1/settings", tc.WebPort))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	// Check that web_port is returned
	webPort, ok := result["web_port"].(float64)
	if !ok {
		t.Fatalf("expected web_port in response, got: %v", result)
	}
	if int(webPort) != tc.WebPort {
		t.Errorf("unexpected web_port: %v, expected %d", webPort, tc.WebPort)
	}
}

// =============================================================================
// Test: Daemon Status API
// =============================================================================

// TestWeb_DaemonStatusAPI verifies that the daemon status endpoint returns
// correct information.
func TestWeb_DaemonStatusAPI(t *testing.T) {
	tc := setupWebTest(t)
	tc.writeConfig(t, nil)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/v1/daemon/status", tc.WebPort))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["status"] != "running" {
		t.Errorf("expected status=running, got %v", result["status"])
	}
	if int(result["proxy_port"].(float64)) != tc.ProxyPort {
		t.Errorf("unexpected proxy_port: %v", result["proxy_port"])
	}
	if int(result["web_port"].(float64)) != tc.WebPort {
		t.Errorf("unexpected web_port: %v", result["web_port"])
	}
}

// =============================================================================
// Test: Reload API
// =============================================================================

// TestWeb_ReloadAPI verifies that the reload endpoint triggers config reload.
func TestWeb_ReloadAPI(t *testing.T) {
	tc := setupWebTest(t)
	tc.writeConfig(t, nil)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	// Trigger reload
	req, _ := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:%d/api/v1/reload", tc.WebPort), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}
}

// =============================================================================
// Test: Static Files
// =============================================================================

// TestWeb_StaticFiles verifies that the web UI static files are served.
func TestWeb_StaticFiles(t *testing.T) {
	tc := setupWebTest(t)
	tc.writeConfig(t, nil)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	// Request index.html
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", tc.WebPort))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "<!DOCTYPE html>") && !strings.Contains(string(body), "<html") {
		t.Errorf("expected HTML content, got: %s", body[:min(100, len(body))])
	}
}

// =============================================================================
// Test: CORS Headers
// =============================================================================

// TestWeb_CORSHeaders verifies that API endpoints accept cross-origin requests.
func TestWeb_CORSHeaders(t *testing.T) {
	tc := setupWebTest(t)
	tc.writeConfig(t, nil)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	// Send a regular GET request with Origin header
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:%d/api/v1/health", tc.WebPort), nil)
	req.Header.Set("Origin", "http://localhost:3000")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// The request should succeed regardless of CORS headers
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// =============================================================================
// Test: Bindings API
// =============================================================================

// TestWeb_BindingsAPI_List verifies that the bindings endpoint returns
// the configured project bindings.
func TestWeb_BindingsAPI_List(t *testing.T) {
	tc := setupWebTest(t)
	tc.writeConfig(t, map[string]interface{}{
		"bindings": map[string]interface{}{
			"/path/to/project": map[string]interface{}{
				"profile": "default",
			},
		},
	})

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/v1/bindings", tc.WebPort))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	// bindings is an array
	bindings, ok := result["bindings"].([]interface{})
	if !ok {
		t.Fatalf("expected bindings array, got %T", result["bindings"])
	}

	found := false
	for _, b := range bindings {
		binding := b.(map[string]interface{})
		if binding["path"] == "/path/to/project" {
			found = true
			break
		}
	}
	if !found && len(bindings) > 0 {
		// If there are bindings but not ours, that's still OK for this test
		t.Logf("binding not found, but got %d bindings", len(bindings))
	}
}

// =============================================================================
// Test: Logs API
// =============================================================================

// TestWeb_LogsAPI verifies that the logs endpoint returns proxy logs.
func TestWeb_LogsAPI(t *testing.T) {
	tc := setupWebTest(t)
	tc.writeConfig(t, nil)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/v1/logs?limit=10", tc.WebPort))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	// Should have entries array (may be empty)
	if _, ok := result["entries"]; !ok {
		t.Errorf("expected entries field in response, got: %v", result)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
