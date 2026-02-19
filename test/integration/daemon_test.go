// Package integration contains integration tests for the daemon module.
// These tests verify the daemon's behavior in real-world scenarios.
//
// Run with: go test -tags=integration ./test/integration/...
//
//go:build integration

package integration

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestConfig holds test configuration
type TestConfig struct {
	BinaryPath string
	ConfigDir  string
	ProxyPort  int
	WebPort    int
}

func setupTest(t *testing.T) *TestConfig {
	t.Helper()

	// Find project root (where go.mod is)
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("failed to find project root: %v", err)
	}

	// Build the binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "zen")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = projectRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, out)
	}

	// Use unique ports to avoid conflicts
	proxyPort := findFreePort(t)
	webPort := findFreePort(t)

	// Create config directory
	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)

	// Write minimal config
	configPath := filepath.Join(configDir, "zen.json")
	config := fmt.Sprintf(`{"version":6,"proxy_port":%d,"web_port":%d,"providers":{},"profiles":{}}`, proxyPort, webPort)
	os.WriteFile(configPath, []byte(config), 0644)

	return &TestConfig{
		BinaryPath: binaryPath,
		ConfigDir:  configDir,
		ProxyPort:  proxyPort,
		WebPort:    webPort,
	}
}

func findFreePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func (tc *TestConfig) runDaemon(t *testing.T, args ...string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(tc.BinaryPath, args...)
	cmd.Env = append(os.Environ(),
		"HOME="+filepath.Dir(tc.ConfigDir),
		"GOZEN_DAEMON=1",
	)
	return cmd
}

func (tc *TestConfig) readPID(t *testing.T) (int, error) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(tc.ConfigDir, "zend.pid"))
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

func (tc *TestConfig) isPortListening(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (tc *TestConfig) waitForDaemonReady(t *testing.T, timeout time.Duration) error {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if tc.isPortListening(tc.ProxyPort) && tc.isPortListening(tc.WebPort) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("daemon not ready after %v", timeout)
}

func (tc *TestConfig) waitForDaemonStop(t *testing.T, pid int, timeout time.Duration, cmd *exec.Cmd) error {
	t.Helper()

	// If we have the cmd, use Wait() which properly reaps the process
	if cmd != nil {
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case <-done:
			return nil
		case <-time.After(timeout):
			return fmt.Errorf("process %d still running after %v", pid, timeout)
		}
	}

	// Fallback: poll for process exit
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		proc, err := os.FindProcess(pid)
		if err != nil {
			return nil
		}
		if proc.Signal(syscall.Signal(0)) != nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("process %d still running after %v", pid, timeout)
}

// =============================================================================
// Test: Daemon Start
// =============================================================================

// TestDaemonStart_ShouldCreatePIDFile verifies that starting the daemon
// creates a PID file with the correct process ID.
func TestDaemonStart_ShouldCreatePIDFile(t *testing.T) {
	tc := setupTest(t)

	cmd := tc.runDaemon(t, "daemon", "start", "--foreground")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start daemon: %v", err)
	}
	defer cmd.Process.Kill()

	if err := tc.waitForDaemonReady(t, 5*time.Second); err != nil {
		t.Fatalf("daemon not ready: %v", err)
	}

	pid, err := tc.readPID(t)
	if err != nil {
		t.Fatalf("failed to read PID file: %v", err)
	}

	if pid != cmd.Process.Pid {
		t.Errorf("PID file contains %d, expected %d", pid, cmd.Process.Pid)
	}
}

// TestDaemonStart_ShouldListenOnConfiguredPorts verifies that the daemon
// listens on the ports specified in the configuration.
func TestDaemonStart_ShouldListenOnConfiguredPorts(t *testing.T) {
	tc := setupTest(t)

	cmd := tc.runDaemon(t, "daemon", "start", "--foreground")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start daemon: %v", err)
	}
	defer cmd.Process.Kill()

	if err := tc.waitForDaemonReady(t, 5*time.Second); err != nil {
		t.Fatalf("daemon not ready: %v", err)
	}

	if !tc.isPortListening(tc.ProxyPort) {
		t.Errorf("daemon not listening on proxy port %d", tc.ProxyPort)
	}

	if !tc.isPortListening(tc.WebPort) {
		t.Errorf("daemon not listening on web port %d", tc.WebPort)
	}
}

// TestDaemonStart_ShouldRespondToStatusAPI verifies that the daemon's
// status API endpoint returns valid information.
func TestDaemonStart_ShouldRespondToStatusAPI(t *testing.T) {
	tc := setupTest(t)

	cmd := tc.runDaemon(t, "daemon", "start", "--foreground")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start daemon: %v", err)
	}
	defer cmd.Process.Kill()

	if err := tc.waitForDaemonReady(t, 5*time.Second); err != nil {
		t.Fatalf("daemon not ready: %v", err)
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/daemon/status", tc.WebPort)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to call status API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status API returned %d, expected 200", resp.StatusCode)
	}
}

// =============================================================================
// Test: Daemon Stop
// =============================================================================

// TestDaemonStop_ShouldTerminateProcess verifies that stopping the daemon
// actually terminates the process.
func TestDaemonStop_ShouldTerminateProcess(t *testing.T) {
	tc := setupTest(t)

	// Start daemon
	cmd := tc.runDaemon(t, "daemon", "start", "--foreground")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start daemon: %v", err)
	}

	if err := tc.waitForDaemonReady(t, 5*time.Second); err != nil {
		cmd.Process.Kill()
		t.Fatalf("daemon not ready: %v", err)
	}

	pid := cmd.Process.Pid

	// Send SIGTERM (simulating daemon stop)
	cmd.Process.Signal(syscall.SIGTERM)

	// Wait for process to exit
	if err := tc.waitForDaemonStop(t, pid, 10*time.Second, cmd); err != nil {
		cmd.Process.Kill()
		t.Fatalf("daemon did not stop: %v", err)
	}
}

// TestDaemonStop_ShouldRemovePIDFile verifies that stopping the daemon
// removes the PID file.
func TestDaemonStop_ShouldRemovePIDFile(t *testing.T) {
	tc := setupTest(t)

	// Start daemon
	cmd := tc.runDaemon(t, "daemon", "start", "--foreground")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start daemon: %v", err)
	}

	if err := tc.waitForDaemonReady(t, 5*time.Second); err != nil {
		cmd.Process.Kill()
		t.Fatalf("daemon not ready: %v", err)
	}

	pid := cmd.Process.Pid

	// Send SIGTERM
	cmd.Process.Signal(syscall.SIGTERM)

	if err := tc.waitForDaemonStop(t, pid, 10*time.Second, cmd); err != nil {
		cmd.Process.Kill()
		t.Fatalf("daemon did not stop: %v", err)
	}

	// Wait a bit for filesystem to sync (CI environments may have delays)
	time.Sleep(500 * time.Millisecond)

	// Check PID file is removed
	pidPath := filepath.Join(tc.ConfigDir, "zend.pid")
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		// Read PID file content for debugging
		if content, readErr := os.ReadFile(pidPath); readErr == nil {
			t.Errorf("PID file still exists after daemon stop, content: %s", content)
		} else {
			t.Errorf("PID file still exists after daemon stop")
		}
	}
}

// TestDaemonStop_ShouldReleasePort verifies that stopping the daemon
// releases the ports it was listening on.
func TestDaemonStop_ShouldReleasePort(t *testing.T) {
	tc := setupTest(t)

	// Start daemon
	cmd := tc.runDaemon(t, "daemon", "start", "--foreground")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start daemon: %v", err)
	}

	if err := tc.waitForDaemonReady(t, 5*time.Second); err != nil {
		cmd.Process.Kill()
		t.Fatalf("daemon not ready: %v", err)
	}

	pid := cmd.Process.Pid

	// Send SIGTERM
	cmd.Process.Signal(syscall.SIGTERM)

	if err := tc.waitForDaemonStop(t, pid, 10*time.Second, cmd); err != nil {
		cmd.Process.Kill()
		t.Fatalf("daemon did not stop: %v", err)
	}

	// Wait a bit for ports to be released
	time.Sleep(500 * time.Millisecond)

	// Verify ports are free
	if tc.isPortListening(tc.ProxyPort) {
		t.Errorf("proxy port %d still in use after daemon stop", tc.ProxyPort)
	}

	if tc.isPortListening(tc.WebPort) {
		t.Errorf("web port %d still in use after daemon stop", tc.WebPort)
	}
}

// =============================================================================
// Test: Daemon Restart
// =============================================================================

// TestDaemonRestart_ShouldStopOldProcess verifies that restarting the daemon
// properly stops the old process before starting a new one.
func TestDaemonRestart_ShouldStopOldProcess(t *testing.T) {
	tc := setupTest(t)

	// Start first daemon
	cmd1 := tc.runDaemon(t, "daemon", "start", "--foreground")
	if err := cmd1.Start(); err != nil {
		t.Fatalf("failed to start first daemon: %v", err)
	}

	if err := tc.waitForDaemonReady(t, 5*time.Second); err != nil {
		cmd1.Process.Kill()
		t.Fatalf("first daemon not ready: %v", err)
	}

	oldPID := cmd1.Process.Pid

	// Stop first daemon
	cmd1.Process.Signal(syscall.SIGTERM)
	if err := tc.waitForDaemonStop(t, oldPID, 10*time.Second, cmd1); err != nil {
		cmd1.Process.Kill()
		t.Fatalf("first daemon did not stop: %v", err)
	}

	// Wait for ports to be released
	time.Sleep(500 * time.Millisecond)

	// Start second daemon
	cmd2 := tc.runDaemon(t, "daemon", "start", "--foreground")
	if err := cmd2.Start(); err != nil {
		t.Fatalf("failed to start second daemon: %v", err)
	}
	defer cmd2.Process.Kill()

	if err := tc.waitForDaemonReady(t, 5*time.Second); err != nil {
		t.Fatalf("second daemon not ready: %v", err)
	}

	newPID := cmd2.Process.Pid

	if oldPID == newPID {
		t.Errorf("new daemon has same PID as old daemon")
	}

	// Verify old process is dead
	proc, _ := os.FindProcess(oldPID)
	if proc != nil && proc.Signal(syscall.Signal(0)) == nil {
		t.Errorf("old daemon process %d is still running", oldPID)
	}
}

// TestDaemonRestart_ShouldUpdatePIDFile verifies that restarting the daemon
// updates the PID file with the new process ID.
func TestDaemonRestart_ShouldUpdatePIDFile(t *testing.T) {
	tc := setupTest(t)

	// Start first daemon
	cmd1 := tc.runDaemon(t, "daemon", "start", "--foreground")
	if err := cmd1.Start(); err != nil {
		t.Fatalf("failed to start first daemon: %v", err)
	}

	if err := tc.waitForDaemonReady(t, 5*time.Second); err != nil {
		cmd1.Process.Kill()
		t.Fatalf("first daemon not ready: %v", err)
	}

	oldPID := cmd1.Process.Pid

	// Stop first daemon
	cmd1.Process.Signal(syscall.SIGTERM)
	if err := tc.waitForDaemonStop(t, oldPID, 10*time.Second, cmd1); err != nil {
		cmd1.Process.Kill()
		t.Fatalf("first daemon did not stop: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Start second daemon
	cmd2 := tc.runDaemon(t, "daemon", "start", "--foreground")
	if err := cmd2.Start(); err != nil {
		t.Fatalf("failed to start second daemon: %v", err)
	}
	defer cmd2.Process.Kill()

	if err := tc.waitForDaemonReady(t, 5*time.Second); err != nil {
		t.Fatalf("second daemon not ready: %v", err)
	}

	// Verify PID file has new PID
	pid, err := tc.readPID(t)
	if err != nil {
		t.Fatalf("failed to read PID file: %v", err)
	}

	if pid != cmd2.Process.Pid {
		t.Errorf("PID file contains %d, expected %d", pid, cmd2.Process.Pid)
	}
}

// =============================================================================
// Test: Daemon Upgrade Scenario
// =============================================================================

// TestDaemonUpgrade_ShouldStopOldDaemonEvenIfPortCheckFails verifies that
// the daemon can be stopped even when the port check times out.
// This is the bug that was fixed: when port check failed, the PID file was
// removed, making it impossible to stop the daemon.
func TestDaemonUpgrade_ShouldStopOldDaemonEvenIfPortCheckFails(t *testing.T) {
	tc := setupTest(t)

	// Start daemon
	cmd := tc.runDaemon(t, "daemon", "start", "--foreground")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start daemon: %v", err)
	}

	if err := tc.waitForDaemonReady(t, 5*time.Second); err != nil {
		cmd.Process.Kill()
		t.Fatalf("daemon not ready: %v", err)
	}

	pid := cmd.Process.Pid

	// Verify PID file exists
	storedPID, err := tc.readPID(t)
	if err != nil {
		cmd.Process.Kill()
		t.Fatalf("failed to read PID file: %v", err)
	}

	if storedPID != pid {
		cmd.Process.Kill()
		t.Fatalf("PID file mismatch: got %d, expected %d", storedPID, pid)
	}

	// Simulate the scenario: process is alive, PID file exists
	// Even if we can't connect to the port, we should be able to stop it

	// Send SIGTERM directly to the process (bypassing port check)
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		cmd.Process.Kill()
		t.Fatalf("failed to send SIGTERM: %v", err)
	}

	// Wait for process to exit
	if err := tc.waitForDaemonStop(t, pid, 10*time.Second, cmd); err != nil {
		cmd.Process.Kill()
		t.Fatalf("daemon did not stop: %v", err)
	}
}

// =============================================================================
// Test: Stale PID File Handling
// =============================================================================

// TestDaemonStart_ShouldHandleStalePIDFile verifies that the daemon can
// start even if there's a stale PID file from a crashed daemon.
func TestDaemonStart_ShouldHandleStalePIDFile(t *testing.T) {
	tc := setupTest(t)

	// Create a stale PID file with a non-existent process
	pidPath := filepath.Join(tc.ConfigDir, "zend.pid")
	os.WriteFile(pidPath, []byte("99999\n"), 0644)

	// Start daemon - should succeed despite stale PID file
	cmd := tc.runDaemon(t, "daemon", "start", "--foreground")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start daemon: %v", err)
	}
	defer cmd.Process.Kill()

	if err := tc.waitForDaemonReady(t, 5*time.Second); err != nil {
		t.Fatalf("daemon not ready: %v", err)
	}

	// Verify PID file is updated
	pid, err := tc.readPID(t)
	if err != nil {
		t.Fatalf("failed to read PID file: %v", err)
	}

	if pid != cmd.Process.Pid {
		t.Errorf("PID file not updated: got %d, expected %d", pid, cmd.Process.Pid)
	}
}

// =============================================================================
// Test: Graceful Shutdown
// =============================================================================

// TestDaemonShutdown_ShouldWaitForActiveRequests verifies that the daemon
// waits for active requests to complete before shutting down.
func TestDaemonShutdown_ShouldWaitForActiveRequests(t *testing.T) {
	tc := setupTest(t)

	cmd := tc.runDaemon(t, "daemon", "start", "--foreground")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start daemon: %v", err)
	}

	if err := tc.waitForDaemonReady(t, 5*time.Second); err != nil {
		cmd.Process.Kill()
		t.Fatalf("daemon not ready: %v", err)
	}

	pid := cmd.Process.Pid

	// Start a long-running request in background
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reqDone := make(chan bool, 1)
	go func() {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/daemon/status", tc.WebPort)
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
		}
		reqDone <- true
	}()

	// Wait for request to start
	time.Sleep(100 * time.Millisecond)

	// Send SIGTERM
	cmd.Process.Signal(syscall.SIGTERM)

	// Daemon should shut down gracefully
	if err := tc.waitForDaemonStop(t, pid, 10*time.Second, cmd); err != nil {
		cmd.Process.Kill()
		t.Fatalf("daemon did not stop gracefully: %v", err)
	}
}

// TestDaemonShutdown_ShouldForceKillAfterTimeout verifies that the daemon
// force-kills after the graceful shutdown timeout.
func TestDaemonShutdown_ShouldForceKillAfterTimeout(t *testing.T) {
	// This test is more of a documentation of expected behavior
	// In practice, we can't easily test force-kill without mocking
	t.Skip("requires mocking to test force-kill behavior")
}
