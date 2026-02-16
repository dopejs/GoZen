package daemon

import (
	"fmt"
	"os"
	"os/exec"
)

const taskName = "zend"

// EnableService creates a Windows scheduled task that runs at logon.
func EnableService() error {
	// Clean up legacy tasks
	exec.Command("schtasks", "/delete", "/tn", "opencc-web", "/f").Run()
	exec.Command("schtasks", "/delete", "/tn", "zen-web", "/f").Run()

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}

	out, err := exec.Command("schtasks", "/create",
		"/tn", taskName,
		"/sc", "onlogon",
		"/tr", fmt.Sprintf(`"%s" daemon start --foreground`, exe),
		"/f",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("schtasks create failed: %s: %w", string(out), err)
	}

	return nil
}

// DisableService removes the Windows scheduled task.
func DisableService() error {
	out, err := exec.Command("schtasks", "/delete",
		"/tn", taskName,
		"/f",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("schtasks delete failed: %s: %w", string(out), err)
	}

	return nil
}
