package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

func systemdUnitPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "systemd", "user", "zend.service")
}

const unitTemplate = `[Unit]
Description=GoZen Daemon
After=network.target

[Service]
Type=simple
ExecStart={{.Executable}} daemon start --foreground
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
`

// EnableService installs and enables the systemd user unit on Linux.
func EnableService() error {
	// Clean up legacy opencc-web service
	exec.Command("systemctl", "--user", "stop", "opencc-web.service").Run()
	exec.Command("systemctl", "--user", "disable", "opencc-web.service").Run()
	home, _ := os.UserHomeDir()
	os.Remove(filepath.Join(home, ".config", "systemd", "user", "opencc-web.service"))

	// Clean up legacy zen-web service
	exec.Command("systemctl", "--user", "stop", "zen-web.service").Run()
	exec.Command("systemctl", "--user", "disable", "zen-web.service").Run()
	os.Remove(filepath.Join(home, ".config", "systemd", "user", "zen-web.service"))

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

	tmpl := template.Must(template.New("unit").Parse(unitTemplate))
	if err := tmpl.Execute(f, struct {
		Executable string
	}{
		Executable: exe,
	}); err != nil {
		f.Close()
		return err
	}
	f.Close()

	if out, err := exec.Command("systemctl", "--user", "daemon-reload").CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl daemon-reload failed: %s: %w", string(out), err)
	}
	if out, err := exec.Command("systemctl", "--user", "enable", "--now", "zend.service").CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl enable failed: %s: %w", string(out), err)
	}

	return nil
}

// DisableService disables and removes the systemd user unit on Linux.
func DisableService() error {
	unitPath := systemdUnitPath()

	exec.Command("systemctl", "--user", "stop", "zend.service").Run()
	exec.Command("systemctl", "--user", "disable", "zend.service").Run()

	if err := os.Remove(unitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove unit file: %w", err)
	}

	exec.Command("systemctl", "--user", "daemon-reload").Run()

	return nil
}
