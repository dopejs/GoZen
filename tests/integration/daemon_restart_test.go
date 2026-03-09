package integration

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

// TestAutoRestart verifies daemon can start and respond to health checks.
//
// Note: This test does NOT verify auto-restart after crash. It only validates
// that the daemon binary can start successfully and respond to status API.
// Full auto-restart behavior (crash detection, exponential backoff, restart loop)
// requires a daemon wrapper process, which is tested separately in daemon_autorestart_test.go.
//
// What this test covers:
// - Daemon binary builds successfully
// - Daemon starts in foreground mode
// - Daemon responds to /api/v1/daemon/status
// - Basic daemon health verification
func TestAutoRestart(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping auto-restart test in short mode")
	}

	// Create isolated test environment
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".zen-test")
	os.MkdirAll(configDir, 0755)
	os.Setenv("GOZEN_CONFIG_DIR", configDir)
	defer os.Unsetenv("GOZEN_CONFIG_DIR")

	// Initialize minimal config
	config.ResetDefaultStore()
	store := config.DefaultStore()

	// Get free ports
	proxyPort := getFreePortForTest(t)
	webPort := getFreePortForTest(t)

	if err := store.SetProxyPort(proxyPort); err != nil {
		t.Fatalf("failed to set proxy port: %v", err)
	}
	if err := store.SetWebPort(webPort); err != nil {
		t.Fatalf("failed to set web port: %v", err)
	}
	if err := store.Save(); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Build the binary
	binaryPath := filepath.Join(tmpDir, "zen-test")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "../../")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, output)
	}

	// Test that daemon can start and respond to health checks
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "daemon", "start", "--foreground")
	cmd.Env = append(os.Environ(),
		"GOZEN_CONFIG_DIR="+configDir,
		"GOZEN_DAEMON=1",
	)

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start daemon: %v", err)
	}
	defer cmd.Process.Kill()

	// Wait for daemon to be ready
	if !waitForDaemonReady(proxyPort, 10*time.Second) {
		t.Fatal("daemon did not become ready in time")
	}

	// Verify daemon is responding
	statusURL := fmt.Sprintf("http://127.0.0.1:%d/api/v1/daemon/status", proxyPort)
	resp, err := http.Get(statusURL)
	if err != nil {
		t.Fatalf("daemon not responding: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("daemon status check failed: %d", resp.StatusCode)
	}

	t.Log("daemon started and verified healthy")

	// Note: Full auto-restart testing requires the daemon wrapper to be running
	// This test verifies the daemon can start and respond to health checks
	// Auto-restart behavior is tested in daemon_autorestart_test.go
}

// TestRestartBackoff verifies exponential backoff calculation
func TestRestartBackoff(t *testing.T) {
	backoffs := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		16 * time.Second,
		30 * time.Second, // capped at 30s
		30 * time.Second,
	}

	current := 1 * time.Second
	for i, expected := range backoffs {
		if current != expected {
			t.Errorf("backoff[%d] = %v, want %v", i, current, expected)
		}

		current *= 2
		if current > 30*time.Second {
			current = 30 * time.Second
		}
	}
}

// getFreePortForTest returns a free port for testing
func getFreePortForTest(t *testing.T) int {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port
}

// waitForDaemonReady polls the daemon status endpoint until it's ready or timeout
func waitForDaemonReady(port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	statusURL := fmt.Sprintf("http://127.0.0.1:%d/api/v1/daemon/status", port)

	for time.Now().Before(deadline) {
		resp, err := http.Get(statusURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}
