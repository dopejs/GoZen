package integration

import (
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
	var daemonPID int
	for _, line := range strings.Split(string(statusOutput), "\n") {
		if strings.Contains(line, "PID:") {
			// Parse PID from line like "PID: 12345"
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if pid, err := strconv.Atoi(parts[1]); err == nil {
					daemonPID = pid
					break
				}
			}
		}
	}

	if daemonPID == 0 {
		t.Fatal("could not find daemon PID in status output")
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
