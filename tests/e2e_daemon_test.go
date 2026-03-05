//go:build integration

// Package tests contains end-to-end integration tests for daemon stability features.
//
// These tests build and run the actual zen binary against an isolated config directory
// with unique ports, so they do not interfere with production or dev instances.
//
// Run with:
//
//	go test -tags integration -v -timeout 120s ./tests/
//
// Or via the helper script:
//
//	./scripts/run_e2e.sh
package tests

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// testEnv holds isolated ports and config dir for a test run.
type testEnv struct {
	configDir string
	webPort   int
	proxyPort int
	binPath   string
}

// findFreePorts returns two consecutive free TCP ports.
func findFreePorts(t *testing.T) (int, int) {
	t.Helper()
	ln1, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("cannot find free port: %v", err)
	}
	port1 := ln1.Addr().(*net.TCPAddr).Port
	ln1.Close()

	ln2, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("cannot find free port: %v", err)
	}
	port2 := ln2.Addr().(*net.TCPAddr).Port
	ln2.Close()

	return port1, port2
}

// setupTestEnv builds the binary and creates an isolated config dir.
func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	// Find project root (this file is in tests/)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get working directory: %v", err)
	}
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err != nil {
		// Maybe we're running from project root
		projectRoot = wd
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err != nil {
			t.Fatalf("cannot find project root (go.mod): tried %s and %s", filepath.Dir(wd), wd)
		}
	}

	// Build binary
	binPath := filepath.Join(t.TempDir(), "zen-e2e-test")
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = projectRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}

	// Create isolated config dir
	configDir := t.TempDir()
	webPort, proxyPort := findFreePorts(t)

	// Write initial config
	cfg := map[string]interface{}{
		"version":    6,
		"web_port":   webPort,
		"proxy_port": proxyPort,
		"providers":  map[string]interface{}{},
		"profiles": map[string]interface{}{
			"default": map[string]interface{}{
				"providers": []string{},
			},
		},
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "zen.json"), data, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	return &testEnv{
		configDir: configDir,
		webPort:   webPort,
		proxyPort: proxyPort,
		binPath:   binPath,
	}
}

// runZen executes the zen binary with the test environment.
func (e *testEnv) runZen(args ...string) (string, error) {
	cmd := exec.Command(e.binPath, args...)
	cmd.Env = append(os.Environ(),
		"GOZEN_CONFIG_DIR="+e.configDir,
		"HOME="+e.configDir, // fallback in case GOZEN_CONFIG_DIR isn't checked
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// startDaemon starts the daemon in the background and waits for it to be ready.
func (e *testEnv) startDaemon(t *testing.T) {
	t.Helper()
	out, err := e.runZen("daemon", "start")
	if err != nil {
		t.Fatalf("daemon start failed: %v\n%s", err, out)
	}
	// Wait for daemon to be ready
	if err := e.waitForDaemon(t, 10*time.Second); err != nil {
		t.Fatalf("daemon did not become ready: %v", err)
	}
}

// stopDaemon stops the daemon.
func (e *testEnv) stopDaemon(t *testing.T) {
	t.Helper()
	// Use stdin "y" in case there are active sessions
	cmd := exec.Command(e.binPath, "daemon", "stop")
	cmd.Env = append(os.Environ(), "GOZEN_CONFIG_DIR="+e.configDir)
	cmd.Stdin = strings.NewReader("y\n")
	cmd.CombinedOutput() // ignore errors — daemon might not be running
}

// waitForDaemon waits until the web port responds.
func (e *testEnv) waitForDaemon(t *testing.T, timeout time.Duration) error {
	t.Helper()
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 500 * time.Millisecond}
	url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/daemon/status", e.webPort)

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("daemon not ready after %v", timeout)
}

// isDaemonUp checks if the daemon web port is responding.
func (e *testEnv) isDaemonUp() bool {
	client := &http.Client{Timeout: 1 * time.Second}
	url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/daemon/status", e.webPort)
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// getJSON fetches a JSON endpoint and decodes into result.
func (e *testEnv) getJSON(t *testing.T, path string, result interface{}) {
	t.Helper()
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://127.0.0.1:%d%s", e.webPort, path)
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET %s failed: %v", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s returned status %d", path, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		t.Fatalf("decode response from %s: %v", path, err)
	}
}

// readConfig reads the config file and returns it as a map.
func (e *testEnv) readConfig(t *testing.T) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(e.configDir, "zen.json"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	return cfg
}

// ============================================================
// Test 1: Port Stability — port stays the same across restarts
// ============================================================
func TestE2E_PortStability(t *testing.T) {
	env := setupTestEnv(t)
	defer env.stopDaemon(t)

	// Start daemon
	env.startDaemon(t)

	// Query daemon status for proxy port
	var status1 struct {
		ProxyPort int `json:"proxy_port"`
		WebPort   int `json:"web_port"`
	}
	env.getJSON(t, "/api/v1/daemon/status", &status1)

	if status1.ProxyPort != env.proxyPort {
		t.Errorf("first start: proxy_port = %d, want %d", status1.ProxyPort, env.proxyPort)
	}

	// Stop daemon
	env.stopDaemon(t)

	// Wait for port to be released
	time.Sleep(1 * time.Second)

	// Restart daemon
	env.startDaemon(t)

	// Query daemon status again
	var status2 struct {
		ProxyPort int `json:"proxy_port"`
		WebPort   int `json:"web_port"`
	}
	env.getJSON(t, "/api/v1/daemon/status", &status2)

	if status2.ProxyPort != status1.ProxyPort {
		t.Errorf("after restart: proxy_port = %d, want %d (same as before)", status2.ProxyPort, status1.ProxyPort)
	}

	// Verify config file also has the correct port persisted
	cfg := env.readConfig(t)
	if cfgPort, ok := cfg["proxy_port"].(float64); ok {
		if int(cfgPort) != env.proxyPort {
			t.Errorf("config proxy_port = %d, want %d", int(cfgPort), env.proxyPort)
		}
	}

	t.Logf("Port stability verified: proxy_port=%d across restart", status1.ProxyPort)
}

// ============================================================
// Test 2: Port Conflict Detection — error when port is occupied
// ============================================================
func TestE2E_PortConflictDetection(t *testing.T) {
	env := setupTestEnv(t)

	// Occupy the proxy port with a non-zen process (a test TCP listener)
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", env.proxyPort))
	if err != nil {
		t.Fatalf("cannot occupy port %d: %v", env.proxyPort, err)
	}
	defer ln.Close()

	// Run daemon in foreground mode with a short timeout.
	// Foreground mode is more reliable for detecting port conflicts because
	// it doesn't fork — the error propagates directly to the exit code.
	cmd := exec.Command(env.binPath, "daemon", "start", "--foreground")
	cmd.Env = append(os.Environ(),
		"GOZEN_CONFIG_DIR="+env.configDir,
		"HOME="+env.configDir,
	)

	// Start process and wait for it to exit (it should fail quickly)
	done := make(chan struct{})
	var out []byte
	var cmdErr error
	go func() {
		out, cmdErr = cmd.CombinedOutput()
		close(done)
	}()

	select {
	case <-done:
		// Process exited
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		t.Fatal("daemon foreground did not exit within timeout")
	}

	if cmdErr == nil {
		t.Fatalf("daemon start should have failed due to port conflict, output: %s", out)
	}

	// Check the daemon log for port conflict details
	logData, _ := os.ReadFile(filepath.Join(env.configDir, "zend.log"))
	combined := strings.ToLower(string(out)) + " " + strings.ToLower(string(logData))

	// The error should mention the port being in use or occupied
	if !strings.Contains(combined, "occupied") && !strings.Contains(combined, "in use") && !strings.Contains(combined, "not a zen process") {
		t.Errorf("expected port conflict error in output or log, got output: %s, log: %s", string(out), string(logData))
	}

	t.Logf("Port conflict detected correctly")
}

// ============================================================
// Test 3: Daemon Recovery — isConnectionError + restart logic
// ============================================================
// NOTE: Full daemon recovery requires a real client (claude/codex) process,
// which is not available in CI. Instead, we test the building blocks:
// 1. Daemon can be started after being killed (the restart part)
// 2. isConnectionError detection is covered by unit tests in cmd/root_test.go
func TestE2E_DaemonRestartAfterKill(t *testing.T) {
	env := setupTestEnv(t)
	defer env.stopDaemon(t)

	// Start daemon
	env.startDaemon(t)

	// Read PID
	pidData, err := os.ReadFile(filepath.Join(env.configDir, "zend.pid"))
	if err != nil {
		t.Fatalf("read PID file: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidData)))
	if err != nil {
		t.Fatalf("parse PID: %v", err)
	}

	// Kill the daemon process
	proc, err := os.FindProcess(pid)
	if err != nil {
		t.Fatalf("find process %d: %v", pid, err)
	}
	if err := proc.Kill(); err != nil {
		t.Fatalf("kill process %d: %v", pid, err)
	}

	// Wait for process to die and port to be released
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !env.isDaemonUp() {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if env.isDaemonUp() {
		t.Fatal("daemon should be dead after kill")
	}

	// Wait a bit for the port to be released
	time.Sleep(1 * time.Second)

	// Restart daemon — should succeed on the same port
	env.startDaemon(t)

	// Verify it's running on the same port
	var status struct {
		ProxyPort int `json:"proxy_port"`
	}
	env.getJSON(t, "/api/v1/daemon/status", &status)

	if status.ProxyPort != env.proxyPort {
		t.Errorf("restarted daemon proxy_port = %d, want %d", status.ProxyPort, env.proxyPort)
	}

	t.Logf("Daemon recovered after kill: PID %d killed, restarted on port %d", pid, status.ProxyPort)
}

// ============================================================
// Test 4: Duration Fix — duration_ms should be milliseconds
// ============================================================
// NOTE: We cannot generate real proxy traffic in e2e without a provider,
// but we verify that the monitoring endpoint returns the correct field type.
// The actual millisecond conversion is covered by unit tests in
// internal/proxy/request_monitor_test.go and internal/bot/matcher_test.go.
func TestE2E_DurationFieldExists(t *testing.T) {
	env := setupTestEnv(t)
	defer env.stopDaemon(t)

	// Start daemon
	env.startDaemon(t)

	// Query monitoring endpoint — should return empty list but with correct schema
	var resp struct {
		Requests []struct {
			DurationMs json.Number `json:"duration_ms"`
		} `json:"requests"`
		Total int `json:"total"`
	}
	env.getJSON(t, "/api/v1/monitoring/requests", &resp)

	// No requests yet, but the endpoint should work
	if resp.Total != 0 {
		// If there are requests, verify duration_ms is a reasonable value
		for i, req := range resp.Requests {
			ms, err := req.DurationMs.Int64()
			if err != nil {
				t.Errorf("request[%d].duration_ms is not a valid integer: %v", i, err)
				continue
			}
			// If it were nanoseconds, it would be >1,000,000 for even 1ms
			if ms > 1_000_000 {
				t.Errorf("request[%d].duration_ms = %d — looks like nanoseconds, not milliseconds", i, ms)
			}
		}
	}

	t.Logf("Monitoring endpoint responds correctly, total=%d requests", resp.Total)
}

// ============================================================
// Test 5: Config Set — set proxy_port via CLI, verify via API
// ============================================================
func TestE2E_ConfigSetProxyPort(t *testing.T) {
	env := setupTestEnv(t)
	defer env.stopDaemon(t)

	// Find a new free port for the new proxy port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("cannot find free port: %v", err)
	}
	newPort := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	// Step 1: Set port while daemon is NOT running — config change persists
	out, err := env.runZen("config", "set", "proxy_port", strconv.Itoa(newPort))
	if err != nil {
		t.Fatalf("config set failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, fmt.Sprintf("proxy_port set to %d", newPort)) {
		t.Errorf("unexpected output: %s", out)
	}

	// Verify the config file was updated
	cfg := env.readConfig(t)
	if cfgPort, ok := cfg["proxy_port"].(float64); ok {
		if int(cfgPort) != newPort {
			t.Errorf("config proxy_port = %d, want %d", int(cfgPort), newPort)
		}
	} else {
		t.Error("proxy_port not found in config")
	}
	t.Logf("Config set verified: proxy_port set to %d in config file", newPort)

	// Step 2: Start daemon with the new port
	env.proxyPort = newPort
	env.startDaemon(t)

	var status struct {
		ProxyPort int `json:"proxy_port"`
	}
	env.getJSON(t, "/api/v1/daemon/status", &status)

	if status.ProxyPort != newPort {
		t.Errorf("daemon proxy_port = %d, want %d", status.ProxyPort, newPort)
	}

	// Step 3: Verify settings API also returns the new port
	var settings struct {
		ProxyPort int `json:"proxy_port"`
	}
	env.getJSON(t, "/api/v1/settings", &settings)

	if settings.ProxyPort != newPort {
		t.Errorf("settings proxy_port = %d, want %d", settings.ProxyPort, newPort)
	}

	t.Logf("Config set + start verified: daemon running on port %d", status.ProxyPort)
}

// ============================================================
// Test: Config Set Validation — invalid values rejected
// ============================================================
func TestE2E_ConfigSetValidation(t *testing.T) {
	env := setupTestEnv(t)

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"port too low", []string{"config", "set", "proxy_port", "80"}, "must be between 1024 and 65535"},
		{"port too high", []string{"config", "set", "proxy_port", "99999"}, "must be between 1024 and 65535"},
		{"port not a number", []string{"config", "set", "proxy_port", "abc"}, "must be a number"},
		{"unknown key", []string{"config", "set", "unknown_key", "value"}, "unknown configuration key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := env.runZen(tt.args...)
			if err == nil {
				t.Errorf("expected error for %v, got success: %s", tt.args, out)
				return
			}
			if !strings.Contains(strings.ToLower(out), strings.ToLower(tt.wantErr)) {
				t.Errorf("expected error containing %q, got: %s", tt.wantErr, out)
			}
		})
	}
}
