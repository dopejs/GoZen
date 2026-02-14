//go:build !windows

package daemon

import (
	"fmt"
	"os"
	"syscall"
)

// IsRunning checks if the daemon process is still alive.
// It first checks the PID file, then falls back to the platform
// service manager (launchd/systemd) to find the actual PID.
func IsRunning() (int, bool) {
	pid, err := ReadPid()
	if err == nil {
		proc, err := os.FindProcess(pid)
		if err == nil {
			if proc.Signal(syscall.Signal(0)) == nil {
				return pid, true
			}
		}
	}

	// Fallback: check platform service manager for the actual PID
	if foundPid, ok := findServicePid(); ok {
		WritePid(foundPid)
		return foundPid, true
	}

	return 0, false
}

// StopDaemon sends SIGTERM to the daemon process.
func StopDaemon() error {
	pid, running := IsRunning()
	if !running {
		RemovePid()
		return fmt.Errorf("daemon is not running")
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop daemon (PID %d): %w", pid, err)
	}
	RemovePid()
	return nil
}

// DaemonSysProcAttr returns SysProcAttr for detaching the child process on Unix.
func DaemonSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
