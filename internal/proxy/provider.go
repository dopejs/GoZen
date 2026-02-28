package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/dopejs/gozen/internal/config"
	"golang.org/x/net/proxy"
)

const (
	InitialBackoff     = 60 * time.Second
	MaxBackoff         = 5 * time.Minute
	AuthInitialBackoff = 30 * time.Minute
	AuthMaxBackoff     = 2 * time.Hour
)

type Provider struct {
	Name            string
	Type            string // "anthropic" or "openai"
	BaseURL         *url.URL
	Token           string
	Model           string
	ReasoningModel  string
	HaikuModel      string
	OpusModel       string
	SonnetModel     string
	EnvVars         map[string]string // Legacy env vars (for backward compat)
	ClaudeEnvVars   map[string]string // Claude Code specific
	CodexEnvVars    map[string]string // Codex specific
	OpenCodeEnvVars map[string]string // OpenCode specific
	ProxyURL        string            // Proxy server URL (http/https/socks5)
	Client          *http.Client      // Per-provider HTTP client (nil = use shared)
	Healthy         bool
	AuthFailed      bool
	FailedAt        time.Time
	Backoff         time.Duration
	mu              sync.Mutex
}

// GetType returns the provider type, defaulting to "anthropic".
func (p *Provider) GetType() string {
	if p.Type == "" {
		return config.ProviderTypeAnthropic
	}
	return p.Type
}

// GetEnvVarsForClient returns the environment variables for a specific client.
func (p *Provider) GetEnvVarsForClient(client string) map[string]string {
	switch client {
	case "codex":
		if len(p.CodexEnvVars) > 0 {
			return p.CodexEnvVars
		}
	case "opencode":
		if len(p.OpenCodeEnvVars) > 0 {
			return p.OpenCodeEnvVars
		}
	default: // claude
		if len(p.ClaudeEnvVars) > 0 {
			return p.ClaudeEnvVars
		}
	}
	return p.EnvVars
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

func (p *Provider) MarkAuthFailed() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Healthy = false
	p.AuthFailed = true
	p.FailedAt = time.Now()
	if p.Backoff < AuthInitialBackoff {
		p.Backoff = AuthInitialBackoff
	} else {
		p.Backoff *= 2
		if p.Backoff > AuthMaxBackoff {
			p.Backoff = AuthMaxBackoff
		}
	}
}

func (p *Provider) MarkHealthy() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Healthy = true
	p.AuthFailed = false
	p.Backoff = 0
}

// NewHTTPClientWithProxy creates an *http.Client that routes requests through
// the given proxy URL. Supports http, https (via http.ProxyURL) and socks5
// (via golang.org/x/net/proxy.SOCKS5 dialer). Returns an error for empty or
// invalid proxy URLs.
func NewHTTPClientWithProxy(proxyURL string, timeout time.Duration) (*http.Client, error) {
	if proxyURL == "" {
		return nil, fmt.Errorf("proxy URL is empty")
	}
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	transport := &http.Transport{}

	switch u.Scheme {
	case "http", "https":
		transport.Proxy = http.ProxyURL(u)
	case "socks5":
		var auth *proxy.Auth
		if u.User != nil {
			auth = &proxy.Auth{
				User: u.User.Username(),
			}
			if p, ok := u.User.Password(); ok {
				auth.Password = p
			}
		}
		dialer, err := proxy.SOCKS5("tcp", u.Host, auth, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
		}
		transport.DialContext = dialer.(proxy.ContextDialer).DialContext
	default:
		return nil, fmt.Errorf("unsupported proxy scheme: %s", u.Scheme)
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}, nil
}
