package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
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

// IsDaemonRunning is not fully supported on Windows.
// Returns false — Windows users should use the scheduled task.
func IsDaemonRunning() (int, bool) {
	return 0, false
}

// StopDaemonProcess is not fully supported on Windows.
func StopDaemonProcess(timeout time.Duration) error {
	return fmt.Errorf("stopping zend is not supported on Windows; disable the scheduled task instead")
}

const _CREATE_NEW_PROCESS_GROUP = 0x00000200

// DaemonSysProcAttr returns SysProcAttr for detaching the child process on Windows.
func DaemonSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: _CREATE_NEW_PROCESS_GROUP}
}
