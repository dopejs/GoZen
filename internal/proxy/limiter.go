package proxy

import (
	"context"
	"fmt"
	"time"
)

// Limiter implements a semaphore-based concurrency limiter
type Limiter struct {
	sem     chan struct{}
	timeout time.Duration
}

// NewLimiter creates a new concurrency limiter with the specified limit.
// A limit of 0 means unlimited (no blocking).
// Default timeout is 30 seconds - requests exceeding this will be rejected.
func NewLimiter(limit int) *Limiter {
	if limit <= 0 {
		// Unlimited: use nil channel (never blocks)
		return &Limiter{sem: nil, timeout: 0}
	}
	return &Limiter{
		sem:     make(chan struct{}, limit),
		timeout: 30 * time.Second,
	}
}

// SetTimeout configures the maximum wait time for acquiring a slot.
// A timeout of 0 means unlimited waiting (original behavior).
func (l *Limiter) SetTimeout(timeout time.Duration) {
	l.timeout = timeout
}

// Acquire blocks until a slot is available or timeout is reached.
// Returns an error if the timeout is exceeded.
func (l *Limiter) Acquire() error {
	if l.sem == nil {
		// Unlimited mode
		return nil
	}

	// No timeout configured - block indefinitely (original behavior)
	if l.timeout == 0 {
		l.sem <- struct{}{}
		return nil
	}

	// Try to acquire with timeout
	ctx, cancel := context.WithTimeout(context.Background(), l.timeout)
	defer cancel()

	select {
	case l.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("concurrency limit reached: request timed out after %v", l.timeout)
	}
}

// Release releases a slot
func (l *Limiter) Release() {
	if l.sem == nil {
		// Unlimited mode
		return
	}
	<-l.sem
}
