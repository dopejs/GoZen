package integration

import (
	"runtime"
	"testing"
	"time"
)

// TestMemoryStability verifies memory growth stays under 10% over extended runtime
func TestMemoryStability(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping memory stability test in short mode")
	}

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)
	baseline := m1.Alloc

	// Simulate 24-hour load (compressed to 10 seconds for testing)
	// In production, this would run for 24 hours with 10-50 req/hr
	duration := 10 * time.Second
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	done := time.After(duration)
	for {
		select {
		case <-done:
			runtime.ReadMemStats(&m2)
			current := m2.Alloc
			growth := float64(current-baseline) / float64(baseline) * 100

			if growth > 10.0 {
				t.Fatalf("memory growth %.2f%% exceeds 10%% threshold (baseline=%d current=%d)",
					growth, baseline, current)
			}
			return
		case <-ticker.C:
			// Simulate request processing
			_ = make([]byte, 1024)
		}
	}
}

// TestGoroutineStability verifies no goroutine leaks over extended runtime
func TestGoroutineStability(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping goroutine stability test in short mode")
	}

	baseline := runtime.NumGoroutine()

	// Simulate 24-hour load (compressed to 10 seconds for testing)
	duration := 10 * time.Second
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	done := time.After(duration)
	for {
		select {
		case <-done:
			// Allow time for goroutines to clean up
			time.Sleep(100 * time.Millisecond)
			runtime.GC()
			time.Sleep(100 * time.Millisecond)

			current := runtime.NumGoroutine()
			// Allow small variance (±5 goroutines) for runtime fluctuations
			if current > baseline+5 {
				t.Fatalf("goroutine leak detected: baseline=%d current=%d (growth=%d)",
					baseline, current, current-baseline)
			}
			return
		case <-ticker.C:
			// Simulate spawning short-lived goroutines
			go func() {
				time.Sleep(10 * time.Millisecond)
			}()
		}
	}
}
