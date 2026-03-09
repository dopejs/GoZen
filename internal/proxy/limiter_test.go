package proxy

import (
	"sync"
	"testing"
	"time"
)

// TestLimiterBasic verifies basic acquire/release behavior
func TestLimiterBasic(t *testing.T) {
	limiter := NewLimiter(2)

	// Acquire first slot
	if err := limiter.Acquire(); err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}

	// Acquire second slot
	if err := limiter.Acquire(); err != nil {
		t.Fatalf("second acquire failed: %v", err)
	}

	// Release first slot
	limiter.Release()

	// Should be able to acquire again
	if err := limiter.Acquire(); err != nil {
		t.Fatalf("acquire after release failed: %v", err)
	}

	// Cleanup
	limiter.Release()
	limiter.Release()
}

// TestLimiterBlocking verifies that Acquire blocks when limit is reached
func TestLimiterBlocking(t *testing.T) {
	limiter := NewLimiter(2)

	// Acquire both slots
	limiter.Acquire()
	limiter.Acquire()

	// Third acquire should block
	blocked := make(chan bool, 1)
	go func() {
		blocked <- true
		limiter.Acquire()
		blocked <- false
	}()

	// Wait for goroutine to start
	<-blocked

	// Give it time to block
	time.Sleep(50 * time.Millisecond)

	// Should still be blocked
	select {
	case <-blocked:
		t.Fatal("acquire should have blocked but didn't")
	default:
		// Good, still blocked
	}

	// Release one slot
	limiter.Release()

	// Now it should unblock
	select {
	case <-blocked:
		// Good, unblocked
	case <-time.After(100 * time.Millisecond):
		t.Fatal("acquire didn't unblock after release")
	}

	// Cleanup
	limiter.Release()
	limiter.Release()
}

// TestLimiterConcurrent verifies limiter works correctly under concurrent load
func TestLimiterConcurrent(t *testing.T) {
	const limit = 10
	const workers = 50
	limiter := NewLimiter(limit)

	var wg sync.WaitGroup
	var mu sync.Mutex
	maxConcurrent := 0
	currentConcurrent := 0

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			limiter.Acquire()
			defer limiter.Release()

			// Track concurrent executions
			mu.Lock()
			currentConcurrent++
			if currentConcurrent > maxConcurrent {
				maxConcurrent = currentConcurrent
			}
			mu.Unlock()

			// Simulate work
			time.Sleep(10 * time.Millisecond)

			mu.Lock()
			currentConcurrent--
			mu.Unlock()
		}()
	}

	wg.Wait()

	if maxConcurrent > limit {
		t.Errorf("max concurrent = %d, want <= %d", maxConcurrent, limit)
	}
	if maxConcurrent < limit {
		t.Logf("warning: max concurrent = %d, expected %d (may indicate timing issue)", maxConcurrent, limit)
	}
}

// TestLimiterZeroLimit verifies behavior with zero limit (should allow unlimited)
func TestLimiterZeroLimit(t *testing.T) {
	limiter := NewLimiter(0)

	// Should be able to acquire many times without blocking
	for i := 0; i < 100; i++ {
		if err := limiter.Acquire(); err != nil {
			t.Fatalf("acquire %d failed: %v", i, err)
		}
	}

	// Cleanup
	for i := 0; i < 100; i++ {
		limiter.Release()
	}
}

// TestLimiterTimeout verifies that Acquire times out when limit is reached
func TestLimiterTimeout(t *testing.T) {
	limiter := NewLimiter(2)
	limiter.SetTimeout(100 * time.Millisecond)

	// Acquire both slots
	if err := limiter.Acquire(); err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	if err := limiter.Acquire(); err != nil {
		t.Fatalf("second acquire failed: %v", err)
	}

	// Third acquire should timeout
	start := time.Now()
	err := limiter.Acquire()
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	if elapsed < 90*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Errorf("timeout took %v, expected ~100ms", elapsed)
	}

	// Cleanup
	limiter.Release()
	limiter.Release()
}

// TestLimiterNoTimeout verifies that SetTimeout(0) disables timeout
func TestLimiterNoTimeout(t *testing.T) {
	limiter := NewLimiter(1)
	limiter.SetTimeout(0) // Disable timeout

	// Acquire the only slot
	if err := limiter.Acquire(); err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}

	// Second acquire should block indefinitely (we'll release after a delay)
	done := make(chan error, 1)
	go func() {
		done <- limiter.Acquire()
	}()

	// Wait a bit to ensure it's blocking
	time.Sleep(50 * time.Millisecond)

	// Should still be blocked
	select {
	case err := <-done:
		t.Fatalf("acquire should have blocked, but got: %v", err)
	default:
		// Good, still blocked
	}

	// Release the slot
	limiter.Release()

	// Now it should unblock
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("acquire failed after release: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("acquire didn't unblock after release")
	}

	// Cleanup
	limiter.Release()
}

// TestLimiterTimeoutConcurrent verifies timeout behavior under concurrent load
func TestLimiterTimeoutConcurrent(t *testing.T) {
	const limit = 5
	const workers = 20
	limiter := NewLimiter(limit)
	limiter.SetTimeout(50 * time.Millisecond)

	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0
	timeoutCount := 0

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := limiter.Acquire()
			if err != nil {
				mu.Lock()
				timeoutCount++
				mu.Unlock()
				return
			}
			defer limiter.Release()

			mu.Lock()
			successCount++
			mu.Unlock()

			// Simulate work
			time.Sleep(100 * time.Millisecond)
		}()
	}

	wg.Wait()

	if successCount+timeoutCount != workers {
		t.Errorf("success(%d) + timeout(%d) = %d, want %d", successCount, timeoutCount, successCount+timeoutCount, workers)
	}

	if timeoutCount == 0 {
		t.Error("expected some timeouts, got none")
	}

	t.Logf("success: %d, timeout: %d", successCount, timeoutCount)
}
