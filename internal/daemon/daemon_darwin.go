package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

const launchdLabel = "com.dopejs.zen-web"
const legacyLaunchdLabel = "com.dopejs.opencc-web"

func launchdPlistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", launchdLabel+".plist")
}

const plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>{{.Label}}</string>
  <key>ProgramArguments</key>
  <array>
    <string>{{.Executable}}</string>
    <string>web</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>{{.LogPath}}</string>
  <key>StandardErrorPath</key>
  <string>{{.LogPath}}</string>
  <key>EnvironmentVariables</key>
  <dict>
    <key>GOZEN_WEB_DAEMON</key>
    <string>1</string>
  </dict>
</dict>
</plist>
`

// EnableService installs and loads the launchd plist on macOS.
func EnableService() error {
	// Clean up legacy opencc-web plist if it exists
	legacyPlistPath := filepath.Join(filepath.Dir(launchdPlistPath()), legacyLaunchdLabel+".plist")
	if _, err := os.Stat(legacyPlistPath); err == nil {
		exec.Command("launchctl", "unload", legacyPlistPath).Run()
		os.Remove(legacyPlistPath)
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}

	plistPath := launchdPlistPath()
	dir := filepath.Dir(plistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(plistPath)
	if err != nil {
		return err
	}

	tmpl := template.Must(template.New("plist").Parse(plistTemplate))
	if err := tmpl.Execute(f, struct {
		Label      string
		Executable string
		LogPath    string
	}{
		Label:      launchdLabel,
		Executable: exe,
		LogPath:    LogPath(),
	}); err != nil {
		f.Close()
		return err
	}
	f.Close() // Close before launchctl load

	out, err := exec.Command("launchctl", "load", plistPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("launchctl load failed: %s: %w", string(out), err)
	}

	return nil
}

// DisableService unloads and removes the launchd plist on macOS.
func DisableService() error {
	plistPath := launchdPlistPath()

	out, err := exec.Command("launchctl", "unload", plistPath).CombinedOutput()
	if err != nil {
		// Ignore error if not loaded
		_ = out
	}

	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plist: %w", err)
	}

	return nil
}

// findServicePid checks launchd for the daemon's PID.
func findServicePid() (int, bool) {
	out, err := exec.Command("launchctl", "list", launchdLabel).Output()
	if err != nil {
		return 0, false
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "\"PID\"") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			pidStr := strings.TrimSuffix(strings.TrimSpace(parts[1]), ";")
			pid, err := strconv.Atoi(strings.TrimSpace(pidStr))
			if err == nil && pid > 0 {
				return pid, true
			}
		}
	}
	return 0, false
}
