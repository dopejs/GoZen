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

// TestDisableEnableProviderE2E verifies the full disable/enable flow
// through the real daemon, testing CLI commands and HTTP API endpoints.
func TestDisableEnableProviderE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup isolated test environment
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	configDir := filepath.Join(tmpDir, ".zen")
	t.Setenv("GOZEN_CONFIG_DIR", configDir)
	os.MkdirAll(configDir, 0755)

	// Cleanup go module cache permissions
	t.Cleanup(func() {
		modCache := filepath.Join(tmpDir, "go", "pkg", "mod")
		filepath.Walk(modCache, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			os.Chmod(path, 0755)
			return nil
		})
	})

	// Write test config with two providers on ephemeral ports
	webPort := 29850
	proxyPort := 29851
	cfg := fmt.Sprintf(`{
		"version": 14,
		"web_port": %d,
		"proxy_port": %d,
		"providers": {
			"alpha": {
				"base_url": "https://api.alpha.example.com",
				"auth_token": "tok-alpha",
				"model": "claude-sonnet-4-5"
			},
			"beta": {
				"base_url": "https://api.beta.example.com",
				"auth_token": "tok-beta",
				"model": "claude-sonnet-4-5"
			}
		},
		"profiles": {
			"default": {"providers": ["alpha", "beta"]}
		}
	}`, webPort, proxyPort)
	if err := os.WriteFile(filepath.Join(configDir, "zen.json"), []byte(cfg), 0600); err != nil {
		t.Fatal(err)
	}

	// Build zen binary
	zenBinary := filepath.Join(tmpDir, "zen")
	buildCmd := exec.Command("go", "build", "-o", zenBinary, "../../")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build: %v\n%s", err, output)
	}

	// Start daemon
	startCmd := exec.Command(zenBinary, "daemon", "start")
	if output, err := startCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to start daemon: %v\n%s", err, output)
	}

	// Wait for daemon to be ready
	apiBase := fmt.Sprintf("http://127.0.0.1:%d/api/v1", webPort)
	if !waitForReady(apiBase, 5*time.Second) {
		t.Fatal("daemon did not become ready in time")
	}

	// Cleanup: stop daemon
	t.Cleanup(func() {
		exec.Command(zenBinary, "daemon", "stop").CombinedOutput()
	})

	client := &http.Client{Timeout: 5 * time.Second}

	// ================================================================
	// Test 1: Initially no disabled providers (API)
	// ================================================================
	t.Run("api_initially_no_disabled", func(t *testing.T) {
		resp := apiGet(t, client, apiBase+"/providers/disabled")
		var result struct {
			DisabledProviders []interface{} `json:"disabled_providers"`
		}
		json.Unmarshal(resp, &result)
		if len(result.DisabledProviders) != 0 {
			t.Errorf("initially %d disabled, want 0", len(result.DisabledProviders))
		}
	})

	// ================================================================
	// Test 2: Disable provider via API
	// ================================================================
	t.Run("api_disable_provider", func(t *testing.T) {
		resp := apiPost(t, client, apiBase+"/providers/alpha/disable", `{"type":"today"}`)
		var result map[string]interface{}
		json.Unmarshal(resp, &result)
		if result["provider"] != "alpha" {
			t.Errorf("provider = %v, want alpha", result["provider"])
		}
		if result["disabled"] != true {
			t.Errorf("disabled = %v, want true", result["disabled"])
		}
		if result["type"] != "today" {
			t.Errorf("type = %v, want today", result["type"])
		}
	})

	// ================================================================
	// Test 3: Provider list shows disabled field
	// ================================================================
	t.Run("api_providers_list_shows_disabled", func(t *testing.T) {
		resp := apiGet(t, client, apiBase+"/providers")
		var providers []struct {
			Name     string      `json:"name"`
			Disabled interface{} `json:"disabled"`
		}
		json.Unmarshal(resp, &providers)

		for _, p := range providers {
			if p.Name == "alpha" && p.Disabled == nil {
				t.Error("alpha should have disabled field")
			}
			if p.Name == "beta" && p.Disabled != nil {
				t.Error("beta should not have disabled field")
			}
		}
	})

	// ================================================================
	// Test 4: Disabled list shows alpha
	// ================================================================
	t.Run("api_disabled_list", func(t *testing.T) {
		resp := apiGet(t, client, apiBase+"/providers/disabled")
		var result struct {
			DisabledProviders []struct {
				Provider string `json:"provider"`
				Type     string `json:"type"`
			} `json:"disabled_providers"`
		}
		json.Unmarshal(resp, &result)
		if len(result.DisabledProviders) != 1 {
			t.Fatalf("disabled count = %d, want 1", len(result.DisabledProviders))
		}
		if result.DisabledProviders[0].Provider != "alpha" {
			t.Errorf("disabled provider = %q, want alpha", result.DisabledProviders[0].Provider)
		}
	})

	// ================================================================
	// Test 5: Enable via API
	// ================================================================
	t.Run("api_enable_provider", func(t *testing.T) {
		resp := apiPost(t, client, apiBase+"/providers/alpha/enable", "")
		var result map[string]interface{}
		json.Unmarshal(resp, &result)
		if result["disabled"] != false {
			t.Errorf("disabled = %v, want false", result["disabled"])
		}
	})

	// ================================================================
	// Test 6: CLI disable
	// ================================================================
	t.Run("cli_disable", func(t *testing.T) {
		out, err := exec.Command(zenBinary, "disable", "beta", "--permanent").CombinedOutput()
		if err != nil {
			t.Fatalf("zen disable failed: %v\n%s", err, out)
		}
		output := string(out)
		if !strings.Contains(output, "unavailable") {
			t.Errorf("expected 'unavailable' in output, got: %s", output)
		}
		if !strings.Contains(output, "permanent") {
			t.Errorf("expected 'permanent' in output, got: %s", output)
		}
	})

	// ================================================================
	// Test 7: CLI disable --list
	// ================================================================
	t.Run("cli_disable_list", func(t *testing.T) {
		out, err := exec.Command(zenBinary, "disable", "--list").CombinedOutput()
		if err != nil {
			t.Fatalf("zen disable --list failed: %v\n%s", err, out)
		}
		output := string(out)
		if !strings.Contains(output, "beta") {
			t.Errorf("expected 'beta' in list, got: %s", output)
		}
		if !strings.Contains(output, "permanent") {
			t.Errorf("expected 'permanent' in list, got: %s", output)
		}
	})

	// ================================================================
	// Test 8: CLI enable
	// ================================================================
	t.Run("cli_enable", func(t *testing.T) {
		out, err := exec.Command(zenBinary, "enable", "beta").CombinedOutput()
		if err != nil {
			t.Fatalf("zen enable failed: %v\n%s", err, out)
		}
		output := string(out)
		if !strings.Contains(output, "enabled") {
			t.Errorf("expected 'enabled' in output, got: %s", output)
		}
	})

	// ================================================================
	// Test 9: CLI disable nonexistent provider
	// ================================================================
	t.Run("cli_disable_nonexistent", func(t *testing.T) {
		out, err := exec.Command(zenBinary, "disable", "nonexistent").CombinedOutput()
		if err == nil {
			t.Errorf("expected error for nonexistent provider, got: %s", out)
		}
		if !strings.Contains(string(out), "not found") {
			t.Errorf("expected 'not found' in error, got: %s", out)
		}
	})

	// ================================================================
	// Test 10: zen list shows disabled indicator
	// ================================================================
	t.Run("cli_list_shows_disabled", func(t *testing.T) {
		// Disable alpha via CLI first
		exec.Command(zenBinary, "disable", "alpha", "--month").CombinedOutput()

		out, err := exec.Command(zenBinary, "list").CombinedOutput()
		if err != nil {
			t.Fatalf("zen list failed: %v\n%s", err, out)
		}
		output := string(out)
		if !strings.Contains(output, "[disabled: month]") {
			t.Errorf("expected '[disabled: month]' in list output, got: %s", output)
		}

		// Clean up
		exec.Command(zenBinary, "enable", "alpha").CombinedOutput()
	})

	// ================================================================
	// Test 11: API and CLI state are synchronized
	// ================================================================
	t.Run("api_cli_state_sync", func(t *testing.T) {
		// Disable via CLI
		exec.Command(zenBinary, "disable", "alpha", "--today").CombinedOutput()

		// Verify via API
		resp := apiGet(t, client, apiBase+"/providers/disabled")
		if !strings.Contains(string(resp), "alpha") {
			t.Error("API should show alpha as disabled after CLI disable")
		}

		// Enable via API
		apiPost(t, client, apiBase+"/providers/alpha/enable", "")

		// Verify via CLI
		out, _ := exec.Command(zenBinary, "disable", "--list").CombinedOutput()
		if strings.Contains(string(out), "alpha") {
			t.Error("CLI list should not show alpha after API enable")
		}
	})
}

// waitForReady polls the health endpoint until the daemon is ready.
func waitForReady(apiBase string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 1 * time.Second}
	for time.Now().Before(deadline) {
		resp, err := client.Get(apiBase + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return true
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}

func apiGet(t *testing.T, client *http.Client, url string) []byte {
	t.Helper()
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("GET %s status = %d, body: %s", url, resp.StatusCode, body)
	}
	return body
}

func apiPost(t *testing.T, client *http.Client, url, jsonBody string) []byte {
	t.Helper()
	var bodyReader io.Reader
	if jsonBody != "" {
		bodyReader = strings.NewReader(jsonBody)
	}
	resp, err := client.Post(url, "application/json", bodyReader)
	if err != nil {
		t.Fatalf("POST %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("POST %s status = %d, body: %s", url, resp.StatusCode, body)
	}
	return body
}
