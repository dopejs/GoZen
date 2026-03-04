//go:build !windows

package daemon

import (
	"net"
	"os"
	"strconv"
	"testing"
)

func TestGetProcessOnPort(t *testing.T) {
	t.Run("returns PID and name for listening port", func(t *testing.T) {
		// Start a TCP listener on a random port
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("net.Listen error: %v", err)
		}
		defer ln.Close()

		// Extract port
		port := ln.Addr().(*net.TCPAddr).Port

		pid, name, err := GetProcessOnPort(port)
		if err != nil {
			t.Fatalf("GetProcessOnPort(%d) error: %v", port, err)
		}
		if pid != os.Getpid() {
			t.Errorf("PID = %d, want %d (current process)", pid, os.Getpid())
		}
		if name == "" {
			t.Error("process name should not be empty")
		}
	})

	t.Run("returns error for unused port", func(t *testing.T) {
		// Use a port that's almost certainly not in use
		_, _, err := GetProcessOnPort(19999)
		if err == nil {
			t.Error("expected error for unused port, got nil")
		}
	})
}

func TestIsZenProcess(t *testing.T) {
	tests := []struct {
		name     string
		procName string
		want     bool
	}{
		{"zen binary", "zen", true},
		{"gozen binary", "gozen", true},
		{"zen path", "/usr/local/bin/zen", true},
		{"gozen path", "/usr/local/bin/gozen", true},
		{"zen-dev binary", "zen-dev", true},
		{"not zen - python", "python3", false},
		{"not zen - node", "node", false},
		{"not zen - empty", "", false},
		{"not zen - partial", "zenith", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsZenProcess(tt.procName)
			if got != tt.want {
				t.Errorf("IsZenProcess(%q) = %v, want %v", tt.procName, got, tt.want)
			}
		})
	}
}

func TestGetProcessOnPort_Integration(t *testing.T) {
	// Start a listener, verify we can find it
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen error: %v", err)
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port

	pid, name, err := GetProcessOnPort(port)
	if err != nil {
		t.Fatalf("GetProcessOnPort error: %v", err)
	}

	// The PID should match our test process
	expectedPID := os.Getpid()
	if pid != expectedPID {
		t.Errorf("PID = %d, want %d", pid, expectedPID)
	}

	// The name should be the test binary
	t.Logf("Process on port %d: PID=%d name=%s", port, pid, name)

	// Verify the PID string is valid
	if strconv.Itoa(pid) == "" {
		t.Error("PID should be a valid number")
	}
}
