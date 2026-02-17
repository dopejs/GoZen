package daemon

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/dopejs/gozen/internal/config"
)

// WriteDaemonPid writes the zend daemon PID file atomically with 0600 permissions.
func WriteDaemonPid(pid int) error {
	dir := config.ConfigDirPath()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	pidPath := DaemonPidPath()
	tmp := pidPath + ".tmp"
	if err := os.WriteFile(tmp, []byte(strconv.Itoa(pid)+"\n"), 0600); err != nil {
		return err
	}
	return os.Rename(tmp, pidPath)
}

// ReadDaemonPid reads the zend daemon PID from the PID file.
func ReadDaemonPid() (int, error) {
	data, err := os.ReadFile(DaemonPidPath())
	if err != nil {
		return 0, fmt.Errorf("PID file not found")
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid PID file: %w", err)
	}
	return pid, nil
}

// RemoveDaemonPid removes the zend daemon PID file.
func RemoveDaemonPid() {
	os.Remove(DaemonPidPath())
}

// CleanupLegacyPidFiles removes old web daemon PID files from previous versions.
func CleanupLegacyPidFiles() {
	dir := config.ConfigDirPath()
	// Remove legacy web.pid
	os.Remove(dir + "/web.pid")
	// Remove hash-based web-*.pid files
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "web-") && strings.HasSuffix(e.Name(), ".pid") {
			os.Remove(dir + "/" + e.Name())
		}
	}
}

// IsDaemonPortListening checks if the given port is being listened on.
// Used for PID-port validation to ensure the PID file corresponds to an actual zend process.
func IsDaemonPortListening(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 500*1e6) // 500ms
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
