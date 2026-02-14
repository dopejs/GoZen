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

func systemdUnitPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "systemd", "user", "zen-web.service")
}

const unitTemplate = `[Unit]
Description=GoZen Web Config Server
After=network.target

[Service]
Type=simple
ExecStart={{.Executable}} web
Environment=GOZEN_WEB_DAEMON=1
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
`

// EnableService installs and enables the systemd user unit on Linux.
func EnableService() error {
	// Clean up legacy opencc-web service if it exists
	exec.Command("systemctl", "--user", "stop", "opencc-web.service").Run()
	exec.Command("systemctl", "--user", "disable", "opencc-web.service").Run()
	home, _ := os.UserHomeDir()
	legacyUnitPath := filepath.Join(home, ".config", "systemd", "user", "opencc-web.service")
	os.Remove(legacyUnitPath)

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}

	unitPath := systemdUnitPath()
	dir := filepath.Dir(unitPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(unitPath)
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl := template.Must(template.New("unit").Parse(unitTemplate))
	if err := tmpl.Execute(f, struct {
		Executable string
	}{
		Executable: exe,
	}); err != nil {
		return err
	}

	if out, err := exec.Command("systemctl", "--user", "daemon-reload").CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl daemon-reload failed: %s: %w", string(out), err)
	}
	if out, err := exec.Command("systemctl", "--user", "enable", "--now", "zen-web.service").CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl enable failed: %s: %w", string(out), err)
	}

	return nil
}

// DisableService disables and removes the systemd user unit on Linux.
func DisableService() error {
	unitPath := systemdUnitPath()

	exec.Command("systemctl", "--user", "stop", "zen-web.service").Run()
	exec.Command("systemctl", "--user", "disable", "zen-web.service").Run()

	if err := os.Remove(unitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove unit file: %w", err)
	}

	exec.Command("systemctl", "--user", "daemon-reload").Run()

	return nil
}

// findServicePid checks systemd for the daemon's PID.
func findServicePid() (int, bool) {
	out, err := exec.Command("systemctl", "--user", "show", "zen-web.service", "-p", "MainPID", "--value").Output()
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err == nil && pid > 0 {
		return pid, true
	}
	return 0, false
}
