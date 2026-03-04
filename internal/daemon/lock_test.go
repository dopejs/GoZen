package daemon

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

func TestDaemonLockPath(t *testing.T) {
	expected := filepath.Join(config.ConfigDirPath(), "zend.lock")
	got := DaemonLockPath()
	if got != expected {
		t.Errorf("DaemonLockPath() = %q, want %q", got, expected)
	}
}

func TestAcquireDaemonLock(t *testing.T) {
	// Use a temp dir as config dir so we don't interfere with real config
	tmpDir := t.TempDir()
	t.Setenv("GOZEN_CONFIG_DIR", tmpDir)

	t.Run("lock acquired successfully", func(t *testing.T) {
		f, err := AcquireDaemonLock()
		if err != nil {
			t.Fatalf("AcquireDaemonLock() error: %v", err)
		}
		if f == nil {
			t.Fatal("AcquireDaemonLock() returned nil file")
		}
		defer ReleaseDaemonLock(f)

		// Verify the lock file exists
		lockPath := DaemonLockPath()
		if _, err := os.Stat(lockPath); os.IsNotExist(err) {
			t.Error("lock file does not exist after acquire")
		}
	})

	t.Run("second lock attempt returns ErrLockContention", func(t *testing.T) {
		f1, err := AcquireDaemonLock()
		if err != nil {
			t.Fatalf("first AcquireDaemonLock() error: %v", err)
		}
		defer ReleaseDaemonLock(f1)

		// Second attempt should fail with ErrLockContention
		f2, err := AcquireDaemonLock()
		if err != ErrLockContention {
			if f2 != nil {
				ReleaseDaemonLock(f2)
			}
			t.Fatalf("second AcquireDaemonLock() error = %v, want ErrLockContention", err)
		}
		if f2 != nil {
			t.Error("expected nil file on contention, got non-nil")
			ReleaseDaemonLock(f2)
		}
	})

	t.Run("lock released on file close", func(t *testing.T) {
		f1, err := AcquireDaemonLock()
		if err != nil {
			t.Fatalf("AcquireDaemonLock() error: %v", err)
		}

		// Release the lock
		ReleaseDaemonLock(f1)

		// Should be able to acquire again
		f2, err := AcquireDaemonLock()
		if err != nil {
			t.Fatalf("AcquireDaemonLock() after release error: %v", err)
		}
		defer ReleaseDaemonLock(f2)
	})
}

func TestAcquireDaemonLockBlocking(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("GOZEN_CONFIG_DIR", tmpDir)

	// Acquire the lock
	f1, err := AcquireDaemonLock()
	if err != nil {
		t.Fatalf("AcquireDaemonLock() error: %v", err)
	}

	// Try non-blocking — should get contention and a file handle for blocking wait
	lockPath := DaemonLockPath()
	f2, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		ReleaseDaemonLock(f1)
		t.Fatalf("OpenFile error: %v", err)
	}

	err = syscall.Flock(int(f2.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != syscall.EWOULDBLOCK {
		f2.Close()
		ReleaseDaemonLock(f1)
		t.Fatalf("expected EWOULDBLOCK, got: %v", err)
	}

	// Release the first lock
	ReleaseDaemonLock(f1)

	// Now blocking acquire should succeed
	err = syscall.Flock(int(f2.Fd()), syscall.LOCK_EX)
	if err != nil {
		f2.Close()
		t.Fatalf("blocking Flock error: %v", err)
	}

	// Clean up
	syscall.Flock(int(f2.Fd()), syscall.LOCK_UN)
	f2.Close()
}
