package proxy

import (
	"net/url"
	"sync"
	"time"
)

const (
	InitialBackoff = 60 * time.Second
	MaxBackoff     = 5 * time.Minute
)

type Provider struct {
	Name     string
	BaseURL  *url.URL
	Token    string
	Model    string
	Healthy  bool
	FailedAt time.Time
	Backoff  time.Duration
	mu       sync.Mutex
}

func (p *Provider) IsHealthy() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Healthy {
		return true
	}
	if time.Since(p.FailedAt) >= p.Backoff {
		p.Healthy = true
		return true
	}
	return false
}

func (p *Provider) MarkFailed() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Healthy = false
	p.FailedAt = time.Now()
	if p.Backoff == 0 {
		p.Backoff = InitialBackoff
	} else {
		p.Backoff *= 2
		if p.Backoff > MaxBackoff {
			p.Backoff = MaxBackoff
		}
	}
}

func (p *Provider) MarkHealthy() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Healthy = true
	p.Backoff = 0
}
