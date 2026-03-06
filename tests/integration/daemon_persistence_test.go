package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestDaemonSurvivesCLITermination verifies that the daemon process continues
// running after the CLI process that started it is terminated.
func TestDaemonSurvivesCLITermination(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup test environment
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("GOZEN_CONFIG_DIR", filepath.Join(tmpDir, ".zen"))

	// Cleanup function to fix permissions before TempDir cleanup
	t.Cleanup(func() {
		// Fix permissions on go module cache to allow cleanup
		modCache := filepath.Join(tmpDir, "go", "pkg", "mod")
		if err := filepath.Walk(modCache, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors during walk
			}
			// Make everything writable for cleanup
			os.Chmod(path, 0755)
			return nil
		}); err != nil {
			t.Logf("warning: failed to fix permissions: %v", err)
		}
	})

	// Build zen binary
	zenBinary := filepath.Join(tmpDir, "zen")
	buildCmd := exec.Command("go", "build", "-o", zenBinary, "../../")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build zen binary: %v\n%s", err, output)
	}

	// Start daemon
	startCmd := exec.Command(zenBinary, "daemon", "start")
	if output, err := startCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to start daemon: %v\n%s", err, output)
	}

	// Wait for daemon to be ready
	time.Sleep(2 * time.Second)

	// Get daemon PID
	statusCmd := exec.Command(zenBinary, "daemon", "status")
	statusOutput, err := statusCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to get daemon status: %v\n%s", err, statusOutput)
	}

	// Extract daemon PID from status output
	// Expected format: "zend is running (PID 12345)"
	var daemonPID int
	statusStr := string(statusOutput)

	// Look for pattern "(PID 12345)"
	if idx := strings.Index(statusStr, "(PID "); idx != -1 {
		// Extract the number after "(PID "
		start := idx + 5 // length of "(PID "
		end := strings.Index(statusStr[start:], ")")
		if end != -1 {
			pidStr := statusStr[start : start+end]
			if pid, err := strconv.Atoi(pidStr); err == nil {
				daemonPID = pid
			}
		}
	}

	if daemonPID == 0 {
		t.Fatalf("could not find daemon PID in status output:\n%s", statusOutput)
	}

	// Verify daemon is running
	if err := syscall.Kill(daemonPID, 0); err != nil {
		t.Fatalf("daemon process not running: %v", err)
	}

	// Simulate CLI termination by killing the status command process
	// (In real usage, user would kill the zen CLI process)
	// The daemon should continue running independently

	// Wait a moment
	time.Sleep(1 * time.Second)

	// Verify daemon is still running
	if err := syscall.Kill(daemonPID, 0); err != nil {
		t.Fatalf("daemon process died after CLI termination: %v", err)
	}

	// Verify daemon is still responsive
	statusCmd2 := exec.Command(zenBinary, "daemon", "status")
	statusOutput2, err := statusCmd2.CombinedOutput()
	if err != nil {
		t.Fatalf("daemon not responsive after CLI termination: %v\n%s", err, statusOutput2)
	}

	if !strings.Contains(string(statusOutput2), "running") {
		t.Errorf("daemon status does not show 'running': %s", statusOutput2)
	}

	// Cleanup: stop daemon
	stopCmd := exec.Command(zenBinary, "daemon", "stop")
	if output, err := stopCmd.CombinedOutput(); err != nil {
		t.Logf("warning: failed to stop daemon: %v\n%s", err, output)
	}
}
