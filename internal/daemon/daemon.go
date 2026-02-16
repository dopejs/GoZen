package daemon

import (
	"fmt"
	"hash/fnv"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dopejs/gozen/internal/config"
)

// exeHash returns a short hash of the current executable's resolved path.
// Each distinct binary path gets a unique hash, so multiple binaries
// (e.g. installed + local dev build) use separate PID files.
func exeHash() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		resolved = exe
	}
	h := fnv.New32a()
	h.Write([]byte(resolved))
	return fmt.Sprintf("%08x", h.Sum32())
}

// PidPath returns the path to the PID file for the current executable (legacy web daemon).
func PidPath() string {
	if hash := exeHash(); hash != "" {
		return filepath.Join(config.ConfigDirPath(), fmt.Sprintf("web-%s.pid", hash))
	}
	return filepath.Join(config.ConfigDirPath(), config.WebPidFile)
}

// legacyPidPath returns the path to the legacy PID file (web.pid).
func legacyPidPath() string {
	return filepath.Join(config.ConfigDirPath(), config.WebPidFile)
}

// LogPath returns the path to the web log file.
func LogPath() string {
	return filepath.Join(config.ConfigDirPath(), config.WebLogFile)
}

// WritePid writes the given PID to the PID file atomically with 0600 permissions.
func WritePid(pid int) error {
	dir := config.ConfigDirPath()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	pidPath := PidPath()
	tmp := pidPath + ".tmp"
	if err := os.WriteFile(tmp, []byte(strconv.Itoa(pid)+"\n"), 0600); err != nil {
		return err
	}
	// Clean up legacy PID file if we're using a new hash-based path
	if legacy := legacyPidPath(); legacy != pidPath {
		os.Remove(legacy)
	}
	return os.Rename(tmp, pidPath)
}

// ReadPid reads the PID from the PID file.
// It checks the hash-based PID file first, then falls back to the legacy
// web.pid for migration from older versions.
func ReadPid() (int, error) {
	pidPath := PidPath()
	data, err := os.ReadFile(pidPath)
	if err == nil {
		pid, parseErr := strconv.Atoi(strings.TrimSpace(string(data)))
		if parseErr == nil {
			return pid, nil
		}
	}

	// Fallback: try legacy web.pid for migration
	legacy := legacyPidPath()
	if legacy == pidPath {
		return 0, fmt.Errorf("PID file not found")
	}
	data, err = os.ReadFile(legacy)
	if err != nil {
		return 0, fmt.Errorf("PID file not found")
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid PID file: %w", err)
	}
	return pid, nil
}

// RemovePid removes the PID file.
func RemovePid() {
	os.Remove(PidPath())
}

// --- zend PID management ---

// WriteDaemonPid writes the zend daemon PID file atomically.
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
