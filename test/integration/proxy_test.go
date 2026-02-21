// Package integration contains integration tests for the proxy module.
// These tests verify the proxy's behavior in real-world scenarios.
//
// Run with: go test -tags=integration ./test/integration/...
//
//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ProxyTestConfig holds test configuration for proxy tests
type ProxyTestConfig struct {
	BinaryPath string
	ConfigDir  string
	ProxyPort  int
	WebPort    int
	MockServer *httptest.Server
}

func setupProxyTest(t *testing.T) *ProxyTestConfig {
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

	return &ProxyTestConfig{
		BinaryPath: binaryPath,
		ConfigDir:  configDir,
		ProxyPort:  proxyPort,
		WebPort:    webPort,
	}
}

func (tc *ProxyTestConfig) writeConfig(t *testing.T, providers map[string]interface{}, profiles map[string]interface{}) {
	t.Helper()
	config := map[string]interface{}{
		"version":    6,
		"proxy_port": tc.ProxyPort,
		"web_port":   tc.WebPort,
		"providers":  providers,
		"profiles":   profiles,
	}
	data, _ := json.Marshal(config)
	configPath := filepath.Join(tc.ConfigDir, "zen.json")
	os.WriteFile(configPath, data, 0644)
}

func (tc *ProxyTestConfig) startDaemon(t *testing.T) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(tc.BinaryPath, "daemon", "start", "--foreground")
	cmd.Env = append(os.Environ(),
		"HOME="+filepath.Dir(tc.ConfigDir),
		"GOZEN_DAEMON=1",
	)
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start daemon: %v", err)
	}

	// Wait for daemon to be ready
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", tc.ProxyPort), 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return cmd
		}
		time.Sleep(100 * time.Millisecond)
	}
	cmd.Process.Kill()
	t.Fatalf("daemon not ready after 5s")
	return nil
}

// =============================================================================
// Test: Basic Proxy Routing
// =============================================================================

// TestProxy_ShouldRouteToProvider verifies that the proxy correctly routes
// requests to the configured provider.
func TestProxy_ShouldRouteToProvider(t *testing.T) {
	tc := setupProxyTest(t)

	// Create mock Anthropic server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/v1/messages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		apiKey := r.Header.Get("x-api-key")
		if apiKey != "test-api-key" {
			t.Errorf("missing or wrong API key: got %q, want %q", apiKey, "test-api-key")
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    "msg_123",
			"type":  "message",
			"role":  "assistant",
			"model": "claude-sonnet-4-20250514",
			"content": []map[string]interface{}{
				{"type": "text", "text": "Hello from mock!"},
			},
		})
	}))
	defer mockServer.Close()

	// Configure provider pointing to mock server
	tc.writeConfig(t,
		map[string]interface{}{
			"test-provider": map[string]interface{}{
				"auth_token": "test-api-key",
				"base_url":   mockServer.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"test-provider"},
			},
		},
	)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	// Send request through proxy
	reqBody := map[string]interface{}{
		"model": "claude-sonnet-4-20250514",
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
		"max_tokens": 100,
	}
	body, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("http://127.0.0.1:%d/default/test-session/v1/messages", tc.ProxyPort)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["id"] != "msg_123" {
		t.Errorf("unexpected response id: %v", result["id"])
	}
}

// =============================================================================
// Test: Provider Failover
// =============================================================================

// TestProxy_ShouldFailoverToBackupProvider verifies that the proxy
// automatically fails over to a backup provider when the primary fails.
func TestProxy_ShouldFailoverToBackupProvider(t *testing.T) {
	tc := setupProxyTest(t)

	// Create failing primary server
	failingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"type":    "overloaded_error",
				"message": "Server overloaded",
			},
		})
	}))
	defer failingServer.Close()

	// Create working backup server
	backupServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    "msg_backup",
			"type":  "message",
			"role":  "assistant",
			"model": "claude-sonnet-4-20250514",
			"content": []map[string]interface{}{
				{"type": "text", "text": "Hello from backup!"},
			},
		})
	}))
	defer backupServer.Close()

	// Configure with primary and backup providers
	tc.writeConfig(t,
		map[string]interface{}{
			"primary": map[string]interface{}{
				"auth_token": "primary-key",
				"base_url": failingServer.URL,
			},
			"backup": map[string]interface{}{
				"auth_token": "backup-key",
				"base_url": backupServer.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"primary", "backup"},
			},
		},
	)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	// Send request - should failover to backup
	reqBody := map[string]interface{}{
		"model": "claude-sonnet-4-20250514",
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
		"max_tokens": 100,
	}
	body, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("http://127.0.0.1:%d/default/test-session/v1/messages", tc.ProxyPort)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["id"] != "msg_backup" {
		t.Errorf("expected backup response, got: %v", result["id"])
	}
}

// =============================================================================
// Test: All Providers Fail
// =============================================================================

// TestProxy_ShouldReturnErrorWhenAllProvidersFail verifies that the proxy
// returns an appropriate error when all providers fail.
func TestProxy_ShouldReturnErrorWhenAllProvidersFail(t *testing.T) {
	tc := setupProxyTest(t)

	// Create failing servers
	failingServer1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer failingServer1.Close()

	failingServer2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer failingServer2.Close()

	tc.writeConfig(t,
		map[string]interface{}{
			"provider1": map[string]interface{}{
				"auth_token": "key1",
				"base_url": failingServer1.URL,
			},
			"provider2": map[string]interface{}{
				"auth_token": "key2",
				"base_url": failingServer2.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"provider1", "provider2"},
			},
		},
	)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	reqBody := map[string]interface{}{
		"model": "claude-sonnet-4-20250514",
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
		"max_tokens": 100,
	}
	body, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("http://127.0.0.1:%d/default/test-session/v1/messages", tc.ProxyPort)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Should return error status
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected error status, got 200")
	}
}

// =============================================================================
// Test: Session Persistence
// =============================================================================

// TestProxy_ShouldMaintainSessionAcrossRequests verifies that the proxy
// maintains session state across multiple requests.
func TestProxy_ShouldMaintainSessionAcrossRequests(t *testing.T) {
	tc := setupProxyTest(t)

	requestCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    fmt.Sprintf("msg_%d", requestCount),
			"type":  "message",
			"role":  "assistant",
			"model": "claude-sonnet-4-20250514",
			"content": []map[string]interface{}{
				{"type": "text", "text": fmt.Sprintf("Response %d", requestCount)},
			},
		})
	}))
	defer mockServer.Close()

	tc.writeConfig(t,
		map[string]interface{}{
			"test-provider": map[string]interface{}{
				"auth_token": "test-key",
				"base_url": mockServer.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"test-provider"},
			},
		},
	)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	sessionID := "persistent-session"
	url := fmt.Sprintf("http://127.0.0.1:%d/default/%s/v1/messages", tc.ProxyPort, sessionID)

	// Send multiple requests with same session
	for i := 1; i <= 3; i++ {
		reqBody := map[string]interface{}{
			"model": "claude-sonnet-4-20250514",
			"messages": []map[string]interface{}{
				{"role": "user", "content": fmt.Sprintf("Message %d", i)},
			},
			"max_tokens": 100,
		}
		body, _ := json.Marshal(reqBody)

		resp, err := http.Post(url, "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("request %d returned status %d", i, resp.StatusCode)
		}
		resp.Body.Close()
	}

	if requestCount != 3 {
		t.Errorf("expected 3 requests, got %d", requestCount)
	}
}

// =============================================================================
// Test: Profile Routing
// =============================================================================

// TestProxy_ShouldRouteToCorrectProfile verifies that the proxy routes
// requests to the correct profile based on the URL.
func TestProxy_ShouldRouteToCorrectProfile(t *testing.T) {
	tc := setupProxyTest(t)

	profile1Requests := 0
	profile2Requests := 0

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		profile1Requests++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_profile1",
			"type":    "message",
			"role":    "assistant",
			"model":   "claude-sonnet-4-20250514",
			"content": []map[string]interface{}{{"type": "text", "text": "Profile 1"}},
		})
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		profile2Requests++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_profile2",
			"type":    "message",
			"role":    "assistant",
			"model":   "claude-sonnet-4-20250514",
			"content": []map[string]interface{}{{"type": "text", "text": "Profile 2"}},
		})
	}))
	defer server2.Close()

	tc.writeConfig(t,
		map[string]interface{}{
			"provider1": map[string]interface{}{
				"auth_token": "key1",
				"base_url": server1.URL,
			},
			"provider2": map[string]interface{}{
				"auth_token": "key2",
				"base_url": server2.URL,
			},
		},
		map[string]interface{}{
			"profile1": map[string]interface{}{
				"providers": []string{"provider1"},
			},
			"profile2": map[string]interface{}{
				"providers": []string{"provider2"},
			},
		},
	)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	reqBody := map[string]interface{}{
		"model": "claude-sonnet-4-20250514",
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
		"max_tokens": 100,
	}
	body, _ := json.Marshal(reqBody)

	// Request to profile1
	url1 := fmt.Sprintf("http://127.0.0.1:%d/profile1/session1/v1/messages", tc.ProxyPort)
	resp1, _ := http.Post(url1, "application/json", bytes.NewReader(body))
	resp1.Body.Close()

	// Request to profile2
	url2 := fmt.Sprintf("http://127.0.0.1:%d/profile2/session2/v1/messages", tc.ProxyPort)
	resp2, _ := http.Post(url2, "application/json", bytes.NewReader(body))
	resp2.Body.Close()

	if profile1Requests != 1 {
		t.Errorf("expected 1 request to profile1, got %d", profile1Requests)
	}
	if profile2Requests != 1 {
		t.Errorf("expected 1 request to profile2, got %d", profile2Requests)
	}
}

// =============================================================================
// Test: Streaming Response
// =============================================================================

// TestProxy_ShouldHandleStreamingResponse verifies that the proxy correctly
// handles streaming responses from the provider.
func TestProxy_ShouldHandleStreamingResponse(t *testing.T) {
	tc := setupProxyTest(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if streaming is requested
		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)

		if reqBody["stream"] != true {
			t.Errorf("expected stream=true in request")
		}

		// Send streaming response
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected http.Flusher")
		}

		events := []string{
			`{"type":"message_start","message":{"id":"msg_stream","type":"message","role":"assistant","model":"claude-sonnet-4-20250514","content":[]}}`,
			`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
			`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}`,
			`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" World"}}`,
			`{"type":"content_block_stop","index":0}`,
			`{"type":"message_stop"}`,
		}

		for _, event := range events {
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", event)
			flusher.Flush()
		}
	}))
	defer mockServer.Close()

	tc.writeConfig(t,
		map[string]interface{}{
			"test-provider": map[string]interface{}{
				"auth_token": "test-key",
				"base_url": mockServer.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"test-provider"},
			},
		},
	)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	reqBody := map[string]interface{}{
		"model": "claude-sonnet-4-20250514",
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
		"max_tokens": 100,
		"stream":     true,
	}
	body, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("http://127.0.0.1:%d/default/test-session/v1/messages", tc.ProxyPort)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d: %s", resp.StatusCode, body)
	}

	// Read streaming response
	respBody, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(respBody), "Hello") || !strings.Contains(string(respBody), "World") {
		t.Errorf("streaming response missing expected content: %s", respBody)
	}
}

// =============================================================================
// Test: Request Timeout
// =============================================================================

// TestProxy_ShouldHandleSlowProvider verifies that the proxy handles
// slow providers appropriately.
func TestProxy_ShouldHandleSlowProvider(t *testing.T) {
	tc := setupProxyTest(t)

	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_slow",
			"type":    "message",
			"role":    "assistant",
			"model":   "claude-sonnet-4-20250514",
			"content": []map[string]interface{}{{"type": "text", "text": "Slow response"}},
		})
	}))
	defer slowServer.Close()

	tc.writeConfig(t,
		map[string]interface{}{
			"slow-provider": map[string]interface{}{
				"auth_token": "test-key",
				"base_url": slowServer.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"slow-provider"},
			},
		},
	)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	reqBody := map[string]interface{}{
		"model": "claude-sonnet-4-20250514",
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
		"max_tokens": 100,
	}
	body, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("http://127.0.0.1:%d/default/test-session/v1/messages", tc.ProxyPort)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// =============================================================================
// Test: Invalid Profile
// =============================================================================

// TestProxy_ShouldRejectInvalidProfile verifies that the proxy returns
// an error for requests to non-existent profiles.
func TestProxy_ShouldRejectInvalidProfile(t *testing.T) {
	tc := setupProxyTest(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not reach provider")
	}))
	defer mockServer.Close()

	tc.writeConfig(t,
		map[string]interface{}{
			"test-provider": map[string]interface{}{
				"auth_token": "test-key",
				"base_url": mockServer.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"test-provider"},
			},
		},
	)

	cmd := tc.startDaemon(t)
	defer cmd.Process.Kill()

	reqBody := map[string]interface{}{
		"model": "claude-sonnet-4-20250514",
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
		"max_tokens": 100,
	}
	body, _ := json.Marshal(reqBody)

	// Request to non-existent profile
	url := fmt.Sprintf("http://127.0.0.1:%d/nonexistent-profile/session/v1/messages", tc.ProxyPort)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected error for invalid profile, got 200")
	}
}
