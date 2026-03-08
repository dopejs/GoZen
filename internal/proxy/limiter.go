package proxy

// Limiter implements a semaphore-based concurrency limiter
type Limiter struct {
	sem chan struct{}
}

// NewLimiter creates a new concurrency limiter with the specified limit.
// A limit of 0 means unlimited (no blocking).
func NewLimiter(limit int) *Limiter {
	if limit <= 0 {
		// Unlimited: use nil channel (never blocks)
		return &Limiter{sem: nil}
	}
	return &Limiter{
		sem: make(chan struct{}, limit),
	}
}

// Acquire blocks until a slot is available
func (l *Limiter) Acquire() error {
	if l.sem == nil {
		// Unlimited mode
		return nil
	}
	l.sem <- struct{}{}
	return nil
}

// Release releases a slot
func (l *Limiter) Release() {
	if l.sem == nil {
		// Unlimited mode
		return
	}
	<-l.sem
}
