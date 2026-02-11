package proxy

import (
	"net/url"
	"testing"
	"time"
)

func newTestProvider(name string) *Provider {
	u, _ := url.Parse("https://api.example.com")
	return &Provider{
		Name:    name,
		BaseURL: u,
		Token:   "tok",
		Model:   "test-model",
		Healthy: true,
	}
}

func TestProviderIsHealthy(t *testing.T) {
	p := newTestProvider("a")
	if !p.IsHealthy() {
		t.Error("new provider should be healthy")
	}
}

func TestProviderMarkFailed(t *testing.T) {
	p := newTestProvider("a")
	p.MarkFailed()

	if p.Healthy {
		t.Error("should be unhealthy after MarkFailed")
	}
	if p.Backoff != InitialBackoff {
		t.Errorf("Backoff = %v, want %v", p.Backoff, InitialBackoff)
	}
	if !p.IsHealthy() == true {
		// IsHealthy should return false since backoff hasn't elapsed
	}
}

func TestProviderBackoffDoubles(t *testing.T) {
	p := newTestProvider("a")

	p.MarkFailed()
	if p.Backoff != InitialBackoff {
		t.Errorf("first fail: Backoff = %v, want %v", p.Backoff, InitialBackoff)
	}

	p.MarkFailed()
	if p.Backoff != InitialBackoff*2 {
		t.Errorf("second fail: Backoff = %v, want %v", p.Backoff, InitialBackoff*2)
	}

	p.MarkFailed()
	if p.Backoff != InitialBackoff*4 {
		t.Errorf("third fail: Backoff = %v, want %v", p.Backoff, InitialBackoff*4)
	}
}

func TestProviderBackoffCapsAtMax(t *testing.T) {
	p := newTestProvider("a")

	// Fail many times to exceed max
	for i := 0; i < 20; i++ {
		p.MarkFailed()
	}

	if p.Backoff > MaxBackoff {
		t.Errorf("Backoff = %v, should not exceed %v", p.Backoff, MaxBackoff)
	}
	if p.Backoff != MaxBackoff {
		t.Errorf("Backoff = %v, want %v", p.Backoff, MaxBackoff)
	}
}

func TestProviderMarkHealthyResetsBackoff(t *testing.T) {
	p := newTestProvider("a")
	p.MarkFailed()
	p.MarkFailed()
	p.MarkHealthy()

	if !p.Healthy {
		t.Error("should be healthy after MarkHealthy")
	}
	if p.Backoff != 0 {
		t.Errorf("Backoff = %v, want 0", p.Backoff)
	}
}

func TestProviderIsHealthyAfterBackoffElapsed(t *testing.T) {
	p := newTestProvider("a")
	p.mu.Lock()
	p.Healthy = false
	p.FailedAt = time.Now().Add(-2 * InitialBackoff) // well in the past
	p.Backoff = InitialBackoff
	p.mu.Unlock()

	if !p.IsHealthy() {
		t.Error("should be healthy after backoff elapsed (half-open)")
	}
	// After IsHealthy returns true, Healthy should be set to true
	if !p.Healthy {
		t.Error("Healthy flag should be set to true after half-open recovery")
	}
}

func TestProviderIsHealthyDuringBackoff(t *testing.T) {
	p := newTestProvider("a")
	p.mu.Lock()
	p.Healthy = false
	p.FailedAt = time.Now() // just now
	p.Backoff = InitialBackoff
	p.mu.Unlock()

	if p.IsHealthy() {
		t.Error("should not be healthy during backoff period")
	}
}
