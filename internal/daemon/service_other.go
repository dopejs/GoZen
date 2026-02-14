//go:build !darwin && !linux && !windows

package daemon

// findServicePid is a no-op on unsupported platforms.
func findServicePid() (int, bool) {
	return 0, false
}
