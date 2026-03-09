package integration

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/daemon"
)

// raceEnabled is set to true by race_on.go when built with -race flag
var raceEnabled = false

// TestDaemonAutoRestart tests the real auto-restart behavior in cmd/daemon.go
// Note: These tests build and run the actual binary, which may be flaky in CI
// environments. They are skipped with race detector and in CI environments.
func TestDaemonAutoRestart(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping daemon auto-restart test in short mode")
	}

	// Skip in race detector mode - these tests spawn real processes
	// which can trigger false positives in race detection
	if raceEnabled {
		t.Skip("skipping daemon auto-restart test with race detector")
	}

	// Skip in CI environment - these tests are flaky on GitHub runners
	if os.Getenv("CI") != "" {
		t.Skip("skipping daemon auto-restart test in CI environment")
	}

	// Create isolated test environment
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".zen-test")
	os.MkdirAll(configDir, 0755)
	os.Setenv("GOZEN_CONFIG_DIR", configDir)
	defer os.Unsetenv("GOZEN_CONFIG_DIR")

	// Initialize minimal config using DefaultStore
	config.ResetDefaultStore()
	store := config.DefaultStore()
	if err := store.SetProxyPort(19999); err != nil {
		t.Fatalf("failed to set proxy port: %v", err)
	}
	if err := store.SetWebPort(29999); err != nil {
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

	// Test 1: Verify daemon starts and runs
	t.Run("daemon_starts", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, binaryPath, "daemon", "start", "--foreground")
		cmd.Env = append(os.Environ(), "GOZEN_CONFIG_DIR="+configDir, "GOZEN_DAEMON=1")

		// Start daemon in background
		if err := cmd.Start(); err != nil {
			t.Fatalf("failed to start daemon: %v", err)
		}

		// Wait for daemon to be ready
		time.Sleep(2 * time.Second)

		// Check if PID file exists
		pidPath := filepath.Join(configDir, "daemon.pid")
		if _, err := os.Stat(pidPath); os.IsNotExist(err) {
			t.Errorf("PID file not created: %s", pidPath)
		}

		// Stop daemon
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	})

	// Test 2: Verify fatal error prevents restart
	t.Run("fatal_error_no_restart", func(t *testing.T) {
		// Start a process on the target port to cause port conflict
		listener, err := startTestListener(19999)
		if err != nil {
			t.Fatalf("failed to start test listener: %v", err)
		}
		defer listener.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, binaryPath, "daemon", "start", "--foreground")
		cmd.Env = append(os.Environ(), "GOZEN_CONFIG_DIR="+configDir, "GOZEN_DAEMON=1")

		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatal("expected daemon to fail due to port conflict, but it succeeded")
		}

		// Verify it's a fatal error (not a restart loop)
		outputStr := string(output)
		if !strings.Contains(outputStr, "fatal error") && !strings.Contains(outputStr, "port") {
			t.Errorf("expected fatal error message, got: %s", outputStr)
		}

		// Verify it didn't retry multiple times (should fail quickly)
		if strings.Count(outputStr, "restarting") > 0 {
			t.Errorf("daemon should not restart on fatal error, but found restart attempts: %s", outputStr)
		}
	})

	// Test 3: Verify signal stops daemon without restart
	t.Run("signal_stop_no_restart", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, binaryPath, "daemon", "start", "--foreground")
		cmd.Env = append(os.Environ(), "GOZEN_CONFIG_DIR="+configDir, "GOZEN_DAEMON=1")

		// Capture output
		outputCh := make(chan string, 1)
		go func() {
			output, _ := cmd.CombinedOutput()
			outputCh <- string(output)
		}()

		if err := cmd.Start(); err != nil {
			t.Fatalf("failed to start daemon: %v", err)
		}

		// Wait for daemon to be ready
		time.Sleep(2 * time.Second)

		// Send interrupt signal
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			t.Fatalf("failed to send interrupt: %v", err)
		}

		// Wait for process to exit
		select {
		case output := <-outputCh:
			// Verify no restart attempts after signal
			if strings.Contains(output, "restarting") {
				t.Errorf("daemon should not restart after signal, but found restart attempts: %s", output)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("daemon did not exit within 5 seconds after signal")
		}
	})
}

// TestDaemonCrashRecovery tests that daemon recovers from crashes
func TestDaemonCrashRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping daemon crash recovery test in short mode")
	}

	// This test would require injecting a crash into the daemon
	// For now, we verify the crash detection logic exists
	t.Run("crash_detection_exists", func(t *testing.T) {
		// Verify IsFatalError function exists and works
		normalErr := fmt.Errorf("normal error")
		if daemon.IsFatalError(normalErr) {
			t.Error("normal error should not be fatal")
		}

		fatalErr := &daemon.FatalError{Err: fmt.Errorf("port conflict")}
		if !daemon.IsFatalError(fatalErr) {
			t.Error("FatalError should be detected as fatal")
		}
	})
}

// startTestListener starts a TCP listener on the given port for testing
func startTestListener(port int) (*testListener, error) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &testListener{listener: listener}, nil
}

type testListener struct {
	listener net.Listener
}

func (l *testListener) Close() error {
	if l.listener != nil {
		return l.listener.Close()
	}
	return nil
}
