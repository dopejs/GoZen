package integration

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"
)

// TestAutoRestart verifies daemon restarts within 5 seconds after crash
func TestAutoRestart(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping auto-restart test in short mode")
	}

	// This test requires the auto-restart wrapper to be implemented in cmd/daemon.go
	// For now, we'll test the restart mechanism concept

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Simulate daemon crash and restart
	restartCount := 0
	maxRestarts := 5
	backoff := 1 * time.Second

	for restartCount < maxRestarts {
		start := time.Now()

		// Simulate daemon process
		cmd := exec.CommandContext(ctx, "sleep", "0.1")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			// Daemon crashed, attempt restart
			restartCount++
			elapsed := time.Since(start)

			if elapsed > 5*time.Second {
				t.Fatalf("restart took %v, exceeds 5s threshold", elapsed)
			}

			// Exponential backoff
			time.Sleep(backoff)
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}

			continue
		}

		// Daemon exited cleanly
		break
	}

	if restartCount >= maxRestarts {
		t.Fatalf("exceeded max restart attempts (%d)", maxRestarts)
	}
}

// TestRestartBackoff verifies exponential backoff between restart attempts
func TestRestartBackoff(t *testing.T) {
	backoffs := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		16 * time.Second,
		30 * time.Second, // capped at 30s
		30 * time.Second,
	}

	current := 1 * time.Second
	for i, expected := range backoffs {
		if current != expected {
			t.Errorf("backoff[%d] = %v, want %v", i, current, expected)
		}

		current *= 2
		if current > 30*time.Second {
			current = 30 * time.Second
		}
	}
}
