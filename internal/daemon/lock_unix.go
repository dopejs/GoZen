//go:build !windows

package daemon

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"

	"github.com/dopejs/gozen/internal/config"
)

// ErrLockContention is returned when another process holds the daemon lock.
var ErrLockContention = errors.New("daemon lock is held by another process")

// DaemonLockPath returns the path to the daemon lock file.
func DaemonLockPath() string {
	return filepath.Join(config.ConfigDirPath(), "zend.lock")
}

// AcquireDaemonLock tries to acquire an exclusive, non-blocking lock on the
// daemon lock file. Returns the lock file handle on success, or
// ErrLockContention if another process holds the lock. The caller must call
// ReleaseDaemonLock when done.
func AcquireDaemonLock() (*os.File, error) {
	lockPath := DaemonLockPath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		f.Close()
		if err == syscall.EWOULDBLOCK {
			return nil, ErrLockContention
		}
		return nil, err
	}

	return f, nil
}

// ReleaseDaemonLock releases the daemon lock and closes the file handle.
func ReleaseDaemonLock(f *os.File) {
	if f == nil {
		return
	}
	syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	f.Close()
}
