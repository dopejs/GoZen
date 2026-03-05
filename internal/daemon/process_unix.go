//go:build !windows

package daemon

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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

// StopDaemonProcess stops the zend daemon gracefully.
// It first tries the HTTP shutdown API (works even without a PID file),
// then falls back to SIGTERM if the PID is known.
func StopDaemonProcess(timeout time.Duration) error {
	pid, running := IsDaemonRunning()
	if !running && pid == 0 {
		RemoveDaemonPid()
		return fmt.Errorf("zend is not running")
	}

	// Try HTTP shutdown first — this works regardless of PID file state
	if shutdownViaHTTP(timeout) {
		RemoveDaemonPid()
		return nil
	}

	// HTTP failed. If we don't have a PID, we can't do anything else.
	if pid <= 0 {
		port := config.GetProxyPort()
		return fmt.Errorf("zend is running on port %d but could not be stopped via API; use 'lsof -i :%d' to find and kill it manually", port, port)
	}

	// Fallback: SIGTERM
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop zend (PID %d): %w", pid, err)
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if proc.Signal(syscall.Signal(0)) != nil {
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

// shutdownViaHTTP sends a POST to the daemon shutdown API and waits for the
// port to close. Returns true if the daemon stopped successfully.
func shutdownViaHTTP(timeout time.Duration) bool {
	webPort := config.GetWebPort()
	proxyPort := config.GetProxyPort()

	// Try both web port and proxy port (both register the shutdown endpoint)
	urls := []string{
		fmt.Sprintf("http://127.0.0.1:%d/api/v1/daemon/shutdown", webPort),
		fmt.Sprintf("http://127.0.0.1:%d/api/v1/daemon/shutdown", proxyPort),
	}

	client := &http.Client{Timeout: 2 * time.Second}
	sent := false
	for _, u := range urls {
		resp, err := client.Post(u, "application/json", nil)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			sent = true
			break
		}
	}
	if !sent {
		return false
	}

	// Wait for the proxy port to close
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !IsDaemonPortListening(proxyPort) {
			return true
		}
		time.Sleep(300 * time.Millisecond)
	}
	return false
}

// GetProcessOnPort returns the PID and process name of the process listening on
// the given port. Returns an error if no process is found.
func GetProcessOnPort(port int) (pid int, name string, err error) {
	out, err := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-sTCP:LISTEN", "-t").Output()
	if err != nil {
		return 0, "", fmt.Errorf("no process found on port %d", port)
	}

	pidStr := strings.TrimSpace(strings.Split(string(out), "\n")[0])
	pid, err = strconv.Atoi(pidStr)
	if err != nil {
		return 0, "", fmt.Errorf("invalid PID %q from lsof: %w", pidStr, err)
	}

	out2, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=").Output()
	if err != nil {
		return pid, "", nil // PID found but can't get name
	}
	name = strings.TrimSpace(string(out2))
	return pid, name, nil
}

// IsZenProcess returns true if the given process name is a zen/gozen binary.
// Matches exact binary names "zen" or "gozen" (with or without path prefix).
func IsZenProcess(processName string) bool {
	if processName == "" {
		return false
	}
	// Extract basename from path
	base := filepath.Base(processName)
	return base == "zen" || base == "gozen" || base == "zen-dev" || base == "gozen-dev"
}
