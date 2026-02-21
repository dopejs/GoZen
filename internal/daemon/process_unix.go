//go:build !windows

package daemon

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

// DaemonSysProcAttr returns SysProcAttr for detaching the child process on Unix.
func DaemonSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}

// IsDaemonRunning checks if the zend daemon is running.
// Performs PID-port validation: checks PID file, verifies process is alive,
// and confirms the process is actually listening on the expected port.
// Also detects orphaned daemons (port in use but no valid PID file).
func IsDaemonRunning() (int, bool) {
	proxyPort := config.GetProxyPort()

	pid, err := ReadDaemonPid()
	if err != nil {
		// No PID file - check if port is in use (orphaned daemon)
		if IsDaemonPortListening(proxyPort) {
			// Port is in use but no PID file - orphaned daemon
			// Return -1 to indicate unknown PID but daemon is running
			return -1, true
		}
		return 0, false
	}

	// Check if process is alive
	proc, err := os.FindProcess(pid)
	if err != nil {
		return 0, false
	}
	if proc.Signal(syscall.Signal(0)) != nil {
		// Process is dead, clean up stale PID file
		RemoveDaemonPid()
		// Check if port is still in use (different process took over)
		if IsDaemonPortListening(proxyPort) {
			return -1, true
		}
		return 0, false
	}

	// PID-port validation: verify the process is listening on the proxy port
	// Note: We don't remove the PID file here even if port check fails, because
	// the process is confirmed alive. The port check might fail due to timeout
	// or the daemon still starting up. Removing the PID file would make it
	// impossible to stop the daemon later.
	if !IsDaemonPortListening(proxyPort) {
		// Process is alive but not listening — could be starting up or wrong process
		// Return the PID anyway so caller can decide what to do
		return pid, false
	}

	return pid, true
}

// StopDaemonProcess sends SIGTERM to the zend daemon and waits for it to exit.
// timeout specifies the maximum time to wait for graceful shutdown.
func StopDaemonProcess(timeout time.Duration) error {
	pid, running := IsDaemonRunning()
	if !running && pid == 0 {
		// No PID file or process is dead
		RemoveDaemonPid()
		return fmt.Errorf("zend is not running")
	}

	// If pid == -1, daemon is running but we don't know the PID (orphaned)
	if pid == -1 {
		return fmt.Errorf("zend is running on port %d but PID is unknown; use 'lsof -i :%d' to find and kill it manually", config.GetProxyPort(), config.GetProxyPort())
	}

	// If pid > 0 but running == false, the process is alive but not listening
	// on the expected port. We should still try to stop it.

	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	// Send SIGTERM for graceful shutdown
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop zend (PID %d): %w", pid, err)
	}

	// Wait for process to exit
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if proc.Signal(syscall.Signal(0)) != nil {
			// Process has exited
			RemoveDaemonPid()
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Force kill if still running
	proc.Signal(syscall.SIGKILL)
	RemoveDaemonPid()
	return nil
}
