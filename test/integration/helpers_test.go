// Package integration contains shared test helpers for integration tests.
//
//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// BaseTestConfig holds shared test configuration used by all integration test types.
type BaseTestConfig struct {
	BinaryPath string
	ConfigDir  string
	ProxyPort  int
	WebPort    int
}

// findProjectRoot locates the project root by walking up from cwd to find go.mod.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

// findFreePort allocates an ephemeral TCP port and returns it.
func findFreePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}

// setupBaseTest builds the zen binary and creates an isolated config environment.
func setupBaseTest(t *testing.T) *BaseTestConfig {
	t.Helper()

	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("failed to find project root: %v", err)
	}

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "zen")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = projectRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, out)
	}

	proxyPort := findFreePort(t)
	webPort := findFreePort(t)

	configDir := filepath.Join(tmpDir, ".zen")
	os.MkdirAll(configDir, 0755)

	return &BaseTestConfig{
		BinaryPath: binaryPath,
		ConfigDir:  configDir,
		ProxyPort:  proxyPort,
		WebPort:    webPort,
	}
}

// writeMinimalConfig writes a minimal zen.json config to the test config directory.
func (b *BaseTestConfig) writeMinimalConfig(t *testing.T) {
	t.Helper()
	config := fmt.Sprintf(`{"version":6,"proxy_port":%d,"web_port":%d,"providers":{},"profiles":{}}`, b.ProxyPort, b.WebPort)
	configPath := filepath.Join(b.ConfigDir, "zen.json")
	os.WriteFile(configPath, []byte(config), 0644)
}

// writeJSONConfig writes a JSON config map to zen.json in the test config directory.
func (b *BaseTestConfig) writeJSONConfig(t *testing.T, config map[string]interface{}) {
	t.Helper()
	if _, ok := config["version"]; !ok {
		config["version"] = 6
	}
	if _, ok := config["proxy_port"]; !ok {
		config["proxy_port"] = b.ProxyPort
	}
	if _, ok := config["web_port"]; !ok {
		config["web_port"] = b.WebPort
	}
	data, _ := json.Marshal(config)
	configPath := filepath.Join(b.ConfigDir, "zen.json")
	os.WriteFile(configPath, data, 0644)
}

// startDaemonCmd creates and starts a daemon foreground process.
func (b *BaseTestConfig) startDaemonCmd(t *testing.T) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(b.BinaryPath, "daemon", "start", "--foreground")
	cmd.Env = append(os.Environ(),
		"HOME="+filepath.Dir(b.ConfigDir),
		"GOZEN_DAEMON=1",
	)
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start daemon: %v", err)
	}
	return cmd
}

// isPortListening checks if a TCP port is accepting connections.
func isPortListening(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// waitForDaemonReady waits until both proxy and web ports respond.
func (b *BaseTestConfig) waitForDaemonReady(t *testing.T, timeout time.Duration) error {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if isPortListening(b.ProxyPort) && isPortListening(b.WebPort) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("daemon not ready after %v", timeout)
}
