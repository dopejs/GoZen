package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/dopejs/gozen/internal/config"
)

// Registry manages middleware registration and lifecycle.
type Registry struct {
	pipeline *Pipeline
	builtins map[string]func() Middleware // factory functions for built-in middleware
	loader   *PluginLoader
	logger   *log.Logger
	mu       sync.RWMutex
}

// Global registry instance
var (
	globalRegistry     *Registry
	globalRegistryOnce sync.Once
	globalRegistryMu   sync.RWMutex
)

// InitGlobalRegistry initializes the global middleware registry.
func InitGlobalRegistry(logger *log.Logger) {
	globalRegistryOnce.Do(func() {
		globalRegistryMu.Lock()
		globalRegistry = NewRegistry(logger)
		globalRegistryMu.Unlock()
	})
}

// GetGlobalRegistry returns the global middleware registry.
func GetGlobalRegistry() *Registry {
	globalRegistryMu.RLock()
	defer globalRegistryMu.RUnlock()
	return globalRegistry
}

// GetGlobalPipeline returns the global middleware pipeline.
func GetGlobalPipeline() *Pipeline {
	reg := GetGlobalRegistry()
	if reg == nil {
		return nil
	}
	return reg.Pipeline()
}

// NewRegistry creates a new middleware registry.
func NewRegistry(logger *log.Logger) *Registry {
	r := &Registry{
		pipeline: NewPipeline(logger),
		builtins: make(map[string]func() Middleware),
		loader:   NewPluginLoader(""),
		logger:   logger,
	}
	r.registerBuiltins()
	return r
}

// Pipeline returns the middleware pipeline.
func (r *Registry) Pipeline() *Pipeline {
	return r.pipeline
}

// registerBuiltins registers all built-in middleware factories.
func (r *Registry) registerBuiltins() {
	r.builtins["request-logger"] = NewRequestLogger
	r.builtins["context-injection"] = NewContextInjection
	r.builtins["session-memory"] = NewSessionMemory
	r.builtins["orchestration"] = NewOrchestration
}

// RegisterBuiltin registers a built-in middleware factory.
func (r *Registry) RegisterBuiltin(name string, factory func() Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.builtins[name] = factory
}

// AvailableBuiltins returns the names of all available built-in middlewares.
func (r *Registry) AvailableBuiltins() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.builtins))
	for name := range r.builtins {
		names = append(names, name)
	}
	return names
}

// LoadFromConfig loads middlewares from the configuration.
func (r *Registry) LoadFromConfig() error {
	cfg := config.GetMiddleware()
	if cfg == nil {
		r.pipeline.SetEnabled(false)
		return nil
	}

	r.pipeline.SetEnabled(cfg.Enabled)
	if !cfg.Enabled {
		return nil
	}

	// Clear existing middlewares
	r.pipeline.Clear()

	// Load each configured middleware
	for _, entry := range cfg.Middlewares {
		if !entry.Enabled {
			continue
		}

		if err := r.loadMiddleware(entry); err != nil {
			if r.logger != nil {
				r.logger.Printf("[middleware] failed to load %s: %v", entry.Name, err)
			}
			// Continue loading other middlewares
			continue
		}
	}

	return nil
}

// loadMiddleware loads a single middleware based on its entry configuration.
func (r *Registry) loadMiddleware(entry *config.MiddlewareEntry) error {
	var m Middleware
	var err error

	switch entry.Source {
	case "", "builtin":
		factory, ok := r.builtins[entry.Name]
		if !ok {
			return fmt.Errorf("unknown builtin middleware: %s", entry.Name)
		}
		m = factory()

	case "local":
		if entry.Path == "" {
			return fmt.Errorf("local middleware requires 'path' field")
		}
		m, err = r.loader.LoadLocal(entry.Path)
		if err != nil {
			return fmt.Errorf("failed to load local plugin: %w", err)
		}

	case "remote":
		if entry.URL == "" {
			return fmt.Errorf("remote middleware requires 'url' field")
		}
		m, err = r.loader.LoadRemote(entry.URL)
		if err != nil {
			return fmt.Errorf("failed to load remote plugin: %w", err)
		}

	default:
		return fmt.Errorf("unknown middleware source: %s", entry.Source)
	}

	// Initialize the middleware with its config
	if err := m.Init(entry.Config); err != nil {
		return fmt.Errorf("init failed: %w", err)
	}

	r.pipeline.Add(m)
	if r.logger != nil {
		r.logger.Printf("[middleware] loaded %s v%s (source=%s, priority=%d)", m.Name(), m.Version(), entry.Source, m.Priority())
	}

	return nil
}

// Reload reloads all middlewares from configuration.
func (r *Registry) Reload() error {
	return r.LoadFromConfig()
}

// EnableMiddleware enables a middleware by name in the config.
func (r *Registry) EnableMiddleware(name string) error {
	cfg := config.GetMiddleware()
	if cfg == nil {
		cfg = &config.MiddlewareConfig{
			Enabled:     true,
			Middlewares: []*config.MiddlewareEntry{},
		}
	}

	// Find and enable the middleware
	found := false
	for _, entry := range cfg.Middlewares {
		if entry.Name == name {
			entry.Enabled = true
			found = true
			break
		}
	}

	// If not found, add it as a new builtin
	if !found {
		if _, ok := r.builtins[name]; !ok {
			return fmt.Errorf("unknown middleware: %s", name)
		}
		cfg.Middlewares = append(cfg.Middlewares, &config.MiddlewareEntry{
			Name:    name,
			Enabled: true,
			Source:  "builtin",
		})
	}

	if err := config.SetMiddleware(cfg); err != nil {
		return err
	}

	return r.Reload()
}

// DisableMiddleware disables a middleware by name in the config.
func (r *Registry) DisableMiddleware(name string) error {
	cfg := config.GetMiddleware()
	if cfg == nil {
		return nil
	}

	for _, entry := range cfg.Middlewares {
		if entry.Name == name {
			entry.Enabled = false
			break
		}
	}

	if err := config.SetMiddleware(cfg); err != nil {
		return err
	}

	return r.Reload()
}

// GetMiddlewareConfig returns the config for a specific middleware.
func (r *Registry) GetMiddlewareConfig(name string) json.RawMessage {
	cfg := config.GetMiddleware()
	if cfg == nil {
		return nil
	}

	for _, entry := range cfg.Middlewares {
		if entry.Name == name {
			return entry.Config
		}
	}
	return nil
}

// SetMiddlewareConfig updates the config for a specific middleware.
func (r *Registry) SetMiddlewareConfig(name string, mwConfig json.RawMessage) error {
	cfg := config.GetMiddleware()
	if cfg == nil {
		cfg = &config.MiddlewareConfig{
			Enabled:     true,
			Middlewares: []*config.MiddlewareEntry{},
		}
	}

	found := false
	for _, entry := range cfg.Middlewares {
		if entry.Name == name {
			entry.Config = mwConfig
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("middleware not found: %s", name)
	}

	if err := config.SetMiddleware(cfg); err != nil {
		return err
	}

	return r.Reload()
}
