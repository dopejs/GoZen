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
// Default timeout is 5 seconds - requests exceeding this will be rejected.
func NewLimiter(limit int) *Limiter {
	if limit <= 0 {
		// Unlimited: use nil channel (never blocks)
		return &Limiter{sem: nil, timeout: 0}
	}
	return &Limiter{
		sem:     make(chan struct{}, limit),
		timeout: 5 * time.Second, // Conservative default for fast failure
	}
}

// SetTimeout configures the maximum wait time for acquiring a slot.
// A timeout of 0 means unlimited waiting (original behavior).
func (l *Limiter) SetTimeout(timeout time.Duration) {
	l.timeout = timeout
}

// Acquire blocks until a slot is available, timeout is reached, or context is cancelled.
// Returns an error if the timeout is exceeded or context is cancelled.
func (l *Limiter) Acquire(ctx context.Context) error {
	if l.sem == nil {
		// Unlimited mode
		return nil
	}

	// No timeout configured - use request context only
	if l.timeout == 0 {
		select {
		case l.sem <- struct{}{}:
			return nil
		case <-ctx.Done():
			return fmt.Errorf("request cancelled while waiting for concurrency slot: %w", ctx.Err())
		}
	}

	// Combine request context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, l.timeout)
	defer cancel()

	select {
	case l.sem <- struct{}{}:
		return nil
	case <-timeoutCtx.Done():
		// Check which context was cancelled
		if ctx.Err() != nil {
			return fmt.Errorf("request cancelled while waiting for concurrency slot: %w", ctx.Err())
		}
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
