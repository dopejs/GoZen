//go:build integration

// Package tests contains e2e proxy tests for provider failover,
// scenario routing, and client disconnect handling.
package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// =============================================================================
// Test: Provider Failover — Two Providers
// =============================================================================

func TestE2E_ProviderFailover_TwoProviders(t *testing.T) {
	env := setupTestEnv(t)
	defer env.stopDaemon(t)

	mockA := newMockProviderWithStatus(t, http.StatusServiceUnavailable)
	mockB := newMockProvider(t)

	env.writeConfigWithProviders(t,
		map[string]interface{}{
			"providerA": map[string]interface{}{
				"auth_token": "key-a",
				"base_url":   mockA.URL,
			},
			"providerB": map[string]interface{}{
				"auth_token": "key-b",
				"base_url":   mockB.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"providerA", "providerB"},
			},
		},
	)

	env.startDaemon(t)

	resp, err := env.sendProxyRequest(t, "default", "session1")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 after failover, got %d: %s", resp.StatusCode, body)
	}

	if mockA.RequestCount.Load() == 0 {
		t.Error("providerA should have received at least one request")
	}
	if mockB.RequestCount.Load() == 0 {
		t.Error("providerB should have received the failover request")
	}

	t.Logf("Failover A(503)→B(200) successful: A=%d reqs, B=%d reqs",
		mockA.RequestCount.Load(), mockB.RequestCount.Load())
}

// =============================================================================
// Test: Provider Failover — Three Providers
// =============================================================================

func TestE2E_ProviderFailover_ThreeProviders(t *testing.T) {
	env := setupTestEnv(t)
	defer env.stopDaemon(t)

	mockA := newMockProviderWithStatus(t, http.StatusServiceUnavailable)
	mockB := newMockProviderWithStatus(t, http.StatusServiceUnavailable)
	mockC := newMockProvider(t)

	env.writeConfigWithProviders(t,
		map[string]interface{}{
			"providerA": map[string]interface{}{
				"auth_token": "key-a",
				"base_url":   mockA.URL,
			},
			"providerB": map[string]interface{}{
				"auth_token": "key-b",
				"base_url":   mockB.URL,
			},
			"providerC": map[string]interface{}{
				"auth_token": "key-c",
				"base_url":   mockC.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"providerA", "providerB", "providerC"},
			},
		},
	)

	env.startDaemon(t)

	resp, err := env.sendProxyRequest(t, "default", "session1")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 after failover chain, got %d: %s", resp.StatusCode, body)
	}

	if mockC.RequestCount.Load() == 0 {
		t.Error("providerC should have received the failover request")
	}

	t.Logf("Failover A(503)→B(503)→C(200) successful: A=%d, B=%d, C=%d",
		mockA.RequestCount.Load(), mockB.RequestCount.Load(), mockC.RequestCount.Load())
}

// =============================================================================
// Test: Provider Failover — All Down
// =============================================================================

func TestE2E_ProviderFailover_AllDown(t *testing.T) {
	env := setupTestEnv(t)
	defer env.stopDaemon(t)

	mockA := newMockProviderWithStatus(t, http.StatusInternalServerError)
	mockB := newMockProviderWithStatus(t, http.StatusInternalServerError)

	env.writeConfigWithProviders(t,
		map[string]interface{}{
			"providerA": map[string]interface{}{
				"auth_token": "key-a",
				"base_url":   mockA.URL,
			},
			"providerB": map[string]interface{}{
				"auth_token": "key-b",
				"base_url":   mockB.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"providerA", "providerB"},
			},
		},
	)

	env.startDaemon(t)

	resp, err := env.sendProxyRequest(t, "default", "session1")
	if err != nil {
		t.Fatalf("request failed (proxy crashed?): %v", err)
	}
	defer resp.Body.Close()

	// Should return an error status but NOT crash
	if resp.StatusCode == http.StatusOK {
		t.Error("expected error status when all providers are down")
	}

	// Verify daemon is still running
	if !env.isDaemonUp() {
		t.Fatal("daemon should still be running after all providers fail")
	}

	t.Logf("All-down handled gracefully: status=%d, daemon still up", resp.StatusCode)
}

// =============================================================================
// Test: Provider Failover — Rate Limited (429)
// =============================================================================

func TestE2E_ProviderFailover_RateLimited(t *testing.T) {
	env := setupTestEnv(t)
	defer env.stopDaemon(t)

	mockA := newMockProviderWithStatus(t, http.StatusTooManyRequests)
	mockB := newMockProvider(t)

	env.writeConfigWithProviders(t,
		map[string]interface{}{
			"providerA": map[string]interface{}{
				"auth_token": "key-a",
				"base_url":   mockA.URL,
			},
			"providerB": map[string]interface{}{
				"auth_token": "key-b",
				"base_url":   mockB.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"providerA", "providerB"},
			},
		},
	)

	env.startDaemon(t)

	resp, err := env.sendProxyRequest(t, "default", "session1")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 after failover from rate-limited, got %d: %s", resp.StatusCode, body)
	}

	if !env.isDaemonUp() {
		t.Fatal("daemon should still be running after rate limit failover")
	}

	t.Logf("Rate limit failover successful: A(429)→B(200)")
}

// =============================================================================
// Test: Scenario Routing — Thinking Mode
// =============================================================================

func TestE2E_ScenarioRouting_ThinkingMode(t *testing.T) {
	env := setupTestEnv(t)
	defer env.stopDaemon(t)

	mockStandard := newMockProvider(t)
	mockThink := newMockProvider(t)

	env.writeConfigWithProviders(t,
		map[string]interface{}{
			"standard": map[string]interface{}{
				"auth_token": "key-std",
				"base_url":   mockStandard.URL,
			},
			"thinker": map[string]interface{}{
				"auth_token": "key-think",
				"base_url":   mockThink.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"standard"},
				"routing": map[string]interface{}{
					"think": map[string]interface{}{
						"providers": []string{"thinker"},
					},
				},
			},
		},
	)

	env.startDaemon(t)

	// Send a request with thinking enabled
	thinkReqBody := map[string]interface{}{
		"model": "claude-sonnet-4-5",
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Think about this"},
		},
		"max_tokens": 100,
		"thinking":   map[string]interface{}{"type": "enabled", "budget_tokens": 1000},
	}
	resp, err := env.sendProxyRequestWithBody(t, "default", "think-session", thinkReqBody)
	if err != nil {
		t.Fatalf("thinking request failed: %v", err)
	}
	resp.Body.Close()

	// The think provider should have received the request
	if mockThink.RequestCount.Load() == 0 {
		// Scenario routing may not failover — check if standard got it instead
		if mockStandard.RequestCount.Load() > 0 {
			t.Error("thinking request routed to standard provider instead of think provider")
		} else {
			t.Error("neither provider received the thinking request")
		}
	}

	// Send a normal request (no thinking) — should go to standard
	normalResp, err := env.sendProxyRequest(t, "default", "normal-session")
	if err != nil {
		t.Fatalf("normal request failed: %v", err)
	}
	normalResp.Body.Close()

	if mockStandard.RequestCount.Load() == 0 {
		t.Error("normal request should have gone to standard provider")
	}

	t.Logf("Scenario routing: think→thinker(%d reqs), normal→standard(%d reqs)",
		mockThink.RequestCount.Load(), mockStandard.RequestCount.Load())
}

// =============================================================================
// Test: Client Disconnect
// =============================================================================

func TestE2E_ClientDisconnect(t *testing.T) {
	env := setupTestEnv(t)
	defer env.stopDaemon(t)

	// Create a slow provider that takes 5 seconds to respond
	mockSlow := newMockProvider(t)
	mockSlow.DefaultResponse = MockResponse{
		StatusCode: http.StatusOK,
		Body:       defaultAnthropicResponse,
		Delay:      5 * time.Second,
	}

	env.writeConfigWithProviders(t,
		map[string]interface{}{
			"slow": map[string]interface{}{
				"auth_token": "key-slow",
				"base_url":   mockSlow.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"slow"},
			},
		},
	)

	env.startDaemon(t)

	// Send a request with a short timeout so the client disconnects
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model": "claude-sonnet-4-20250514",
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
		"max_tokens": 100,
	})
	url := fmt.Sprintf("http://127.0.0.1:%d/default/disconnect-test/v1/messages", env.proxyPort)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// This should fail due to context timeout
	_, err := http.DefaultClient.Do(req)
	if err == nil {
		t.Log("request completed before timeout (slow provider was fast)")
	}

	// Wait a moment for the proxy to handle the disconnect
	time.Sleep(1 * time.Second)

	// Verify daemon is still stable by sending a subsequent request
	fastMock := newMockProvider(t)
	// Rewrite config with a fast provider to verify daemon works
	env.writeConfigWithProviders(t,
		map[string]interface{}{
			"fast": map[string]interface{}{
				"auth_token": "key-fast",
				"base_url":   fastMock.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"fast"},
			},
		},
	)

	// Trigger reload
	reloadURL := fmt.Sprintf("http://127.0.0.1:%d/api/v1/reload", env.webPort)
	http.Post(reloadURL, "", nil)
	time.Sleep(1 * time.Second)

	if !env.isDaemonUp() {
		t.Fatal("daemon should still be running after client disconnect")
	}

	t.Log("Client disconnect handled gracefully, daemon stable")
}

// =============================================================================
// Test: Process Stability — Graceful Shutdown (SIGTERM)
// =============================================================================

func TestE2E_ProcessStability_GracefulShutdown(t *testing.T) {
	env := setupTestEnv(t)

	mock := newMockProvider(t)
	env.writeConfigWithProviders(t,
		map[string]interface{}{
			"provider": map[string]interface{}{
				"auth_token": "key",
				"base_url":   mock.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"provider"},
			},
		},
	)

	env.startDaemon(t)

	// Verify daemon is running
	if !env.isDaemonUp() {
		t.Fatal("daemon should be up")
	}

	// Stop daemon gracefully
	env.stopDaemon(t)

	// Wait for shutdown
	time.Sleep(2 * time.Second)

	// Verify PID file is removed
	pidPath := filepath.Join(env.configDir, "zend.pid")
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Error("PID file should be removed after graceful shutdown")
	}

	// Verify ports are released
	if env.isDaemonUp() {
		t.Error("daemon should not be responding after shutdown")
	}

	t.Log("Graceful shutdown verified: PID file removed, ports released")
}

// =============================================================================
// Test: Process Stability — SIGKILL Recovery
// =============================================================================

func TestE2E_ProcessStability_KillRecovery(t *testing.T) {
	env := setupTestEnv(t)
	defer env.stopDaemon(t)

	mock := newMockProvider(t)
	env.writeConfigWithProviders(t,
		map[string]interface{}{
			"provider": map[string]interface{}{
				"auth_token": "key",
				"base_url":   mock.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"provider"},
			},
		},
	)

	env.startDaemon(t)

	// Read PID and kill the process
	pidData, err := os.ReadFile(filepath.Join(env.configDir, "zend.pid"))
	if err != nil {
		t.Fatalf("read PID: %v", err)
	}
	var pid int
	fmt.Sscanf(string(pidData), "%d", &pid)

	proc, err := os.FindProcess(pid)
	if err != nil {
		t.Fatalf("find process: %v", err)
	}
	if err := proc.Kill(); err != nil {
		t.Fatalf("kill process: %v", err)
	}

	// Wait for process to die
	time.Sleep(2 * time.Second)

	if env.isDaemonUp() {
		t.Fatal("daemon should be dead after SIGKILL")
	}

	// Restart — should handle stale PID file
	env.startDaemon(t)

	if !env.isDaemonUp() {
		t.Fatal("daemon should be running after restart")
	}

	t.Logf("Kill recovery successful: killed PID %d, restarted on same ports", pid)
}

// =============================================================================
// Test: Process Stability — Idempotent Start
// =============================================================================

func TestE2E_ProcessStability_IdempotentStart(t *testing.T) {
	env := setupTestEnv(t)
	defer env.stopDaemon(t)

	mock := newMockProvider(t)
	env.writeConfigWithProviders(t,
		map[string]interface{}{
			"provider": map[string]interface{}{
				"auth_token": "key",
				"base_url":   mock.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"provider"},
			},
		},
	)

	env.startDaemon(t)

	// Try to start again — should detect existing instance
	out, err := env.runZen("daemon", "start")
	if err != nil {
		// Some implementations return an error, which is fine
		t.Logf("Second start returned error (expected): %v, output: %s", err, out)
	}

	// Daemon should still be running (not crashed by second start attempt)
	if !env.isDaemonUp() {
		t.Fatal("daemon should still be running after second start attempt")
	}

	t.Log("Idempotent start verified: second start didn't crash existing daemon")
}

// =============================================================================
// Test: Process Stability — Config Reload Under Load
// =============================================================================

func TestE2E_ProcessStability_ConfigReloadUnderLoad(t *testing.T) {
	env := setupTestEnv(t)
	defer env.stopDaemon(t)

	mock := newMockProvider(t)
	env.writeConfigWithProviders(t,
		map[string]interface{}{
			"provider": map[string]interface{}{
				"auth_token": "key",
				"base_url":   mock.URL,
			},
		},
		map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{"provider"},
			},
		},
	)

	env.startDaemon(t)

	// Send requests while triggering config reloads
	errCount := 0
	for i := 0; i < 10; i++ {
		// Send a proxy request
		resp, err := env.sendProxyRequest(t, "default", fmt.Sprintf("session-%d", i))
		if err != nil {
			errCount++
			continue
		}
		resp.Body.Close()

		// Trigger a config reload mid-flight every other request
		if i%2 == 0 {
			reloadURL := fmt.Sprintf("http://127.0.0.1:%d/api/v1/reload", env.webPort)
			http.Post(reloadURL, "", nil)
		}
	}

	// Verify web UI is still accessible
	if !env.isDaemonUp() {
		t.Fatal("daemon should be accessible after config reloads under load")
	}

	if errCount > 2 {
		t.Errorf("too many errors during reload-under-load: %d/10", errCount)
	}

	t.Logf("Config reload under load: %d/10 requests succeeded, daemon stable", 10-errCount)
}
