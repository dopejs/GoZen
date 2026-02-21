package middleware

import (
	"fmt"
	"log"
	"sort"
	"sync"
)

// Pipeline manages the middleware execution chain.
type Pipeline struct {
	middlewares []Middleware
	enabled     bool
	mu          sync.RWMutex
	logger      *log.Logger
}

// NewPipeline creates a new middleware pipeline.
func NewPipeline(logger *log.Logger) *Pipeline {
	return &Pipeline{
		middlewares: make([]Middleware, 0),
		enabled:     false,
		logger:      logger,
	}
}

// SetEnabled enables or disables the pipeline.
func (p *Pipeline) SetEnabled(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = enabled
}

// IsEnabled returns whether the pipeline is enabled.
func (p *Pipeline) IsEnabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.enabled
}

// Add adds a middleware to the pipeline.
func (p *Pipeline) Add(m Middleware) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.middlewares = append(p.middlewares, m)
	p.sortMiddlewares()
}

// Remove removes a middleware by name.
func (p *Pipeline) Remove(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, m := range p.middlewares {
		if m.Name() == name {
			p.middlewares = append(p.middlewares[:i], p.middlewares[i+1:]...)
			return
		}
	}
}

// Get returns a middleware by name.
func (p *Pipeline) Get(name string) Middleware {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, m := range p.middlewares {
		if m.Name() == name {
			return m
		}
	}
	return nil
}

// List returns all middlewares in priority order.
func (p *Pipeline) List() []Middleware {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]Middleware, len(p.middlewares))
	copy(result, p.middlewares)
	return result
}

// Clear removes all middlewares.
func (p *Pipeline) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	// Close all middlewares
	for _, m := range p.middlewares {
		if err := m.Close(); err != nil && p.logger != nil {
			p.logger.Printf("[middleware] error closing %s: %v", m.Name(), err)
		}
	}
	p.middlewares = make([]Middleware, 0)
}

// sortMiddlewares sorts middlewares by priority (lower = earlier).
func (p *Pipeline) sortMiddlewares() {
	sort.Slice(p.middlewares, func(i, j int) bool {
		return p.middlewares[i].Priority() < p.middlewares[j].Priority()
	})
}

// ProcessRequest runs all middlewares' ProcessRequest in priority order.
// Returns the modified context or an error if any middleware fails.
func (p *Pipeline) ProcessRequest(ctx *RequestContext) (*RequestContext, error) {
	if !p.IsEnabled() {
		return ctx, nil
	}

	p.mu.RLock()
	middlewares := make([]Middleware, len(p.middlewares))
	copy(middlewares, p.middlewares)
	p.mu.RUnlock()

	var err error
	for _, m := range middlewares {
		if p.logger != nil {
			p.logger.Printf("[middleware] %s.ProcessRequest", m.Name())
		}
		ctx, err = m.ProcessRequest(ctx)
		if err != nil {
			return nil, fmt.Errorf("middleware %s: %w", m.Name(), err)
		}
	}
	return ctx, nil
}

// ProcessResponse runs all middlewares' ProcessResponse in reverse priority order.
// Returns the modified context or an error if any middleware fails.
func (p *Pipeline) ProcessResponse(ctx *ResponseContext) (*ResponseContext, error) {
	if !p.IsEnabled() {
		return ctx, nil
	}

	p.mu.RLock()
	middlewares := make([]Middleware, len(p.middlewares))
	copy(middlewares, p.middlewares)
	p.mu.RUnlock()

	// Process in reverse order for response
	var err error
	for i := len(middlewares) - 1; i >= 0; i-- {
		m := middlewares[i]
		if p.logger != nil {
			p.logger.Printf("[middleware] %s.ProcessResponse", m.Name())
		}
		ctx, err = m.ProcessResponse(ctx)
		if err != nil {
			return nil, fmt.Errorf("middleware %s: %w", m.Name(), err)
		}
	}
	return ctx, nil
}

// MiddlewareInfo contains information about a middleware.
type MiddlewareInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	Source      string `json:"source"` // "builtin", "local", "remote"
}

// Info returns information about all loaded middlewares.
func (p *Pipeline) Info() []MiddlewareInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	info := make([]MiddlewareInfo, len(p.middlewares))
	for i, m := range p.middlewares {
		info[i] = MiddlewareInfo{
			Name:        m.Name(),
			Version:     m.Version(),
			Description: m.Description(),
			Priority:    m.Priority(),
			Source:      "builtin", // TODO: track source
		}
	}
	return info
}
