//go:build !windows

package daemon

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

// TestDaemonLifecycle tests all real-world daemon lifecycle scenarios.
// These tests simulate what users experience when managing the daemon.

// setupTestEnv creates an isolated test environment with a unique port.
func setupTestEnv(t *testing.T, port int) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	config.ResetDefaultStore()
	t.Cleanup(func() { config.ResetDefaultStore() })

	// Create config directory
	configDir := dir + "/.zen"
	os.MkdirAll(configDir, 0755)

	// Set a unique port for this test to avoid conflicts
	config.SetProxyPort(port)

	return dir
}

// startMockServer starts a TCP server on the given port to simulate a running daemon.
// Returns a cleanup function to stop the server.
func startMockServer(t *testing.T, port int) (net.Listener, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("failed to start mock server on port %d: %v", port, err)
	}
	return ln, func() { ln.Close() }
}

// startRealProcess starts a real background process that we can track.
// Returns the PID and a cleanup function.
func startRealProcess(t *testing.T) (int, func()) {
	t.Helper()
	// Start a simple sleep process that we can signal
	cmd := exec.Command("sleep", "60")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start test process: %v", err)
	}
	pid := cmd.Process.Pid
	return pid, func() {
		cmd.Process.Kill()
		cmd.Wait()
	}
}

// =============================================================================
// Scenario 1: Clean state - no daemon, no PID file
// =============================================================================

func TestScenario_CleanState_NoDaemon(t *testing.T) {
	setupTestEnv(t, 51001)

	pid, running := IsDaemonRunning()

	if running {
		t.Error("IsDaemonRunning should return false when no daemon is running")
	}
	if pid != 0 {
		t.Errorf("PID should be 0 when no daemon is running, got %d", pid)
	}
}

// =============================================================================
// Scenario 2: Normal running daemon with valid PID file
// =============================================================================

func TestScenario_NormalRunning_WithPidFile(t *testing.T) {
	setupTestEnv(t, 51002)

	// Start a real process to simulate daemon
	realPid, cleanup := startRealProcess(t)
	defer cleanup()

	// Start mock server on the port
	_, stopServer := startMockServer(t, 51002)
	defer stopServer()

	// Write PID file
	WriteDaemonPid(realPid)

	pid, running := IsDaemonRunning()

	if !running {
		t.Error("IsDaemonRunning should return true when daemon is running normally")
	}
	if pid != realPid {
		t.Errorf("PID should be %d, got %d", realPid, pid)
	}
}

// =============================================================================
// Scenario 3: Stale PID file - process is dead
// =============================================================================

func TestScenario_StalePidFile_ProcessDead(t *testing.T) {
	setupTestEnv(t, 51003)

	// Write a PID for a non-existent process
	WriteDaemonPid(999999999)

	pid, running := IsDaemonRunning()

	if running {
		t.Error("IsDaemonRunning should return false when process is dead")
	}
	if pid != 0 {
		t.Errorf("PID should be 0 when process is dead, got %d", pid)
	}

	// PID file should be cleaned up
	_, err := ReadDaemonPid()
	if err == nil {
		t.Error("Stale PID file should be removed")
	}
}

// =============================================================================
// Scenario 4: Orphaned daemon - port in use but no PID file
// This is the critical upgrade scenario!
// =============================================================================

func TestScenario_OrphanedDaemon_NoPidFile(t *testing.T) {
	setupTestEnv(t, 51004)

	// Start mock server to simulate orphaned daemon occupying the port
	_, stopServer := startMockServer(t, 51004)
	defer stopServer()

	// No PID file exists

	pid, running := IsDaemonRunning()

	if !running {
		t.Error("IsDaemonRunning should return true when port is in use (orphaned daemon)")
	}
	if pid != -1 {
		t.Errorf("PID should be -1 for orphaned daemon, got %d", pid)
	}
}

// =============================================================================
// Scenario 5: Stale PID file + port taken by different process
// Process died but another process took the port
// =============================================================================

func TestScenario_StalePidFile_PortTakenByOther(t *testing.T) {
	setupTestEnv(t, 51005)

	// Write a PID for a non-existent process
	WriteDaemonPid(999999999)

	// But the port is in use by something else
	_, stopServer := startMockServer(t, 51005)
	defer stopServer()

	pid, running := IsDaemonRunning()

	if !running {
		t.Error("IsDaemonRunning should return true when port is in use")
	}
	if pid != -1 {
		t.Errorf("PID should be -1 when original process is dead but port is taken, got %d", pid)
	}

	// Stale PID file should be cleaned up
	_, err := ReadDaemonPid()
	if err == nil {
		t.Error("Stale PID file should be removed even when port is taken by other process")
	}
}

// =============================================================================
// Scenario 6: Process alive but not listening yet (startup phase)
// =============================================================================

func TestScenario_ProcessAlive_NotListeningYet(t *testing.T) {
	setupTestEnv(t, 51006)

	// Start a real process
	realPid, cleanup := startRealProcess(t)
	defer cleanup()

	// Write PID file
	WriteDaemonPid(realPid)

	// But don't start the server - simulating startup phase

	pid, running := IsDaemonRunning()

	// Process is alive but not listening
	if running {
		t.Error("IsDaemonRunning should return false when process is not listening")
	}
	if pid != realPid {
		t.Errorf("PID should be %d (process is alive), got %d", realPid, pid)
	}

	// PID file should NOT be removed (process is still alive)
	savedPid, err := ReadDaemonPid()
	if err != nil {
		t.Error("PID file should NOT be removed when process is alive")
	}
	if savedPid != realPid {
		t.Errorf("Saved PID should be %d, got %d", realPid, savedPid)
	}
}

// =============================================================================
// Scenario 7: Stop a normally running daemon
// =============================================================================

func TestScenario_StopNormalDaemon(t *testing.T) {
	setupTestEnv(t, 51007)

	// Start a real process that responds to signals
	cmd := exec.Command("sleep", "60")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start test process: %v", err)
	}
	realPid := cmd.Process.Pid

	// Channel to track if process exited
	procDone := make(chan error, 1)
	go func() {
		procDone <- cmd.Wait()
	}()

	defer func() {
		// Ensure process is cleaned up
		cmd.Process.Kill()
		select {
		case <-procDone:
		case <-time.After(time.Second):
		}
	}()

	// Start mock server
	ln, _ := startMockServer(t, 51007)

	// Write PID file
	WriteDaemonPid(realPid)

	// Verify daemon is detected as running
	pid, running := IsDaemonRunning()
	if !running || pid != realPid {
		t.Fatalf("Setup failed: daemon should be running with PID %d", realPid)
	}

	// Stop the mock server first (simulate daemon shutdown)
	ln.Close()

	// Now stop the daemon
	err := StopDaemonProcess(2 * time.Second)
	if err != nil {
		t.Errorf("StopDaemonProcess should succeed, got error: %v", err)
	}

	// Wait for process to exit via the Wait() goroutine
	select {
	case <-procDone:
		// Process exited successfully
	case <-time.After(3 * time.Second):
		t.Error("Process should have exited after StopDaemonProcess")
	}

	// PID file should be removed
	_, err = ReadDaemonPid()
	if err == nil {
		t.Error("PID file should be removed after stopping daemon")
	}
}

// =============================================================================
// Scenario 8: Stop when daemon is not running
// =============================================================================

func TestScenario_StopNotRunning(t *testing.T) {
	setupTestEnv(t, 51008)

	err := StopDaemonProcess(time.Second)

	if err == nil {
		t.Error("StopDaemonProcess should return error when daemon is not running")
	}
	if err.Error() != "zend is not running" {
		t.Errorf("Error message should be 'zend is not running', got: %v", err)
	}
}

// =============================================================================
// Scenario 9: Stop orphaned daemon (unknown PID)
// =============================================================================

func TestScenario_StopOrphanedDaemon(t *testing.T) {
	setupTestEnv(t, 51009)

	// Start mock server to simulate orphaned daemon
	_, stopServer := startMockServer(t, 51009)
	defer stopServer()

	// No PID file - orphaned state

	err := StopDaemonProcess(time.Second)

	if err == nil {
		t.Error("StopDaemonProcess should return error for orphaned daemon")
	}

	// Should suggest using lsof
	errMsg := err.Error()
	if !contains(errMsg, "lsof") || !contains(errMsg, "51009") {
		t.Errorf("Error should mention lsof and port number, got: %v", err)
	}
}

// =============================================================================
// Scenario 10: Upgrade scenario simulation
// Old daemon running -> upgrade -> new binary tries to restart
// =============================================================================

func TestScenario_UpgradeWithRunningDaemon(t *testing.T) {
	setupTestEnv(t, 51010)

	// Simulate old daemon: process running + port listening + PID file
	realPid, cleanupProc := startRealProcess(t)
	defer cleanupProc()

	_, stopServer := startMockServer(t, 51010)
	defer stopServer()

	WriteDaemonPid(realPid)

	// New binary checks if daemon is running
	pid, running := IsDaemonRunning()

	if !running {
		t.Error("New binary should detect old daemon is running")
	}
	if pid != realPid {
		t.Errorf("New binary should get correct PID %d, got %d", realPid, pid)
	}

	// New binary should be able to stop old daemon
	// (In real scenario, this allows restart to work)
}

// =============================================================================
// Scenario 11: Upgrade scenario - PID file lost during upgrade
// This is the problematic case that caused issues!
// =============================================================================

func TestScenario_UpgradeWithLostPidFile(t *testing.T) {
	setupTestEnv(t, 51011)

	// Old daemon is running on the port
	_, stopServer := startMockServer(t, 51011)
	defer stopServer()

	// But PID file was somehow lost (e.g., deleted during upgrade)
	// This simulates the orphaned daemon scenario

	pid, running := IsDaemonRunning()

	// New binary should still detect something is running
	if !running {
		t.Error("CRITICAL: New binary should detect daemon via port check even without PID file")
	}
	if pid != -1 {
		t.Errorf("PID should be -1 (unknown) when PID file is lost, got %d", pid)
	}

	// Attempting to start should fail gracefully
	// (In real code, this would show "already running" message)
}

// =============================================================================
// Scenario 12: Multiple rapid start attempts
// =============================================================================

func TestScenario_RapidStartAttempts(t *testing.T) {
	setupTestEnv(t, 51012)

	// First "start" - daemon starts and occupies port
	_, stopServer := startMockServer(t, 51012)
	defer stopServer()

	realPid, cleanupProc := startRealProcess(t)
	defer cleanupProc()
	WriteDaemonPid(realPid)

	// Multiple checks should all return consistent results
	for i := 0; i < 5; i++ {
		pid, running := IsDaemonRunning()
		if !running {
			t.Errorf("Attempt %d: should detect daemon is running", i)
		}
		if pid != realPid {
			t.Errorf("Attempt %d: PID should be %d, got %d", i, realPid, pid)
		}
	}
}

// =============================================================================
// Scenario 13: Port released but PID file remains (graceful shutdown incomplete)
// =============================================================================

func TestScenario_PortReleasedPidFileRemains(t *testing.T) {
	setupTestEnv(t, 51013)

	// Process is alive but stopped listening (shutdown in progress)
	realPid, cleanup := startRealProcess(t)
	defer cleanup()

	WriteDaemonPid(realPid)

	// Port is NOT in use (server stopped)

	pid, running := IsDaemonRunning()

	// Process is alive but not listening - should return pid but running=false
	if running {
		t.Error("Should not report running when port is not listening")
	}
	if pid != realPid {
		t.Errorf("Should return actual PID %d, got %d", realPid, pid)
	}
}

// =============================================================================
// Helper functions
// =============================================================================

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
