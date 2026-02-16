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
func IsDaemonRunning() (int, bool) {
	pid, err := ReadDaemonPid()
	if err != nil {
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
		return 0, false
	}

	// PID-port validation: verify the process is listening on the proxy port
	proxyPort := config.GetProxyPort()
	if !IsDaemonPortListening(proxyPort) {
		// PID is alive but not listening on expected port — stale PID or wrong process
		RemoveDaemonPid()
		return 0, false
	}

	return pid, true
}

// StopDaemonProcess sends SIGTERM to the zend daemon and waits for it to exit.
// timeout specifies the maximum time to wait for graceful shutdown.
func StopDaemonProcess(timeout time.Duration) error {
	pid, running := IsDaemonRunning()
	if !running {
		RemoveDaemonPid()
		return fmt.Errorf("zend is not running")
	}

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
