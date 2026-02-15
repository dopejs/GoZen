package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"syscall"
)

// IsRunning checks if the daemon process is still alive on Windows.
func IsRunning() (int, bool) {
	pid, err := ReadPid()
	if err != nil {
		return 0, false
	}
	// On Windows, FindProcess always succeeds. Use tasklist to verify.
	out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH").Output()
	if err != nil {
		return 0, false
	}
	// Check if output contains the PID as a whole number (word boundary match)
	pattern := regexp.MustCompile(`\b` + fmt.Sprintf("%d", pid) + `\b`)
	if pattern.Match(out) {
		return pid, true
	}
	return 0, false
}

// StopDaemon terminates the daemon process on Windows.
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
	if err := proc.Kill(); err != nil {
		return fmt.Errorf("failed to stop daemon (PID %d): %w", pid, err)
	}
	RemovePid()
	return nil
}

const _CREATE_NEW_PROCESS_GROUP = 0x00000200

// DaemonSysProcAttr returns SysProcAttr for detaching the child process on Windows.
func DaemonSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: _CREATE_NEW_PROCESS_GROUP}
}
