package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/middleware"
)

// MiddlewareConfigResponse is the API response for middleware config.
type MiddlewareConfigResponse struct {
	Enabled     bool                       `json:"enabled"`
	Middlewares []*MiddlewareEntryResponse `json:"middlewares"`
}

// MiddlewareEntryResponse is the API response for a single middleware entry.
type MiddlewareEntryResponse struct {
	Name        string          `json:"name"`
	Enabled     bool            `json:"enabled"`
	Source      string          `json:"source"`
	Version     string          `json:"version,omitempty"`
	Description string          `json:"description,omitempty"`
	Priority    int             `json:"priority,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
}

// handleMiddleware routes GET and PUT requests for middleware config.
// GET/PUT /api/v1/middleware
func (s *Server) handleMiddleware(w http.ResponseWriter, r *http.Request) {
	// Handle specific middleware by name
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/middleware")
	if path != "" && path != "/" {
		name := strings.TrimPrefix(path, "/")
		// Handle sub-routes
		if strings.HasSuffix(name, "/enable") {
			name = strings.TrimSuffix(name, "/enable")
			s.handleMiddlewareEnable(w, r, name)
			return
		}
		if strings.HasSuffix(name, "/disable") {
			name = strings.TrimSuffix(name, "/disable")
			s.handleMiddlewareDisable(w, r, name)
			return
		}
		s.handleMiddlewareByName(w, r, name)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetMiddleware(w, r)
	case http.MethodPut:
		s.handleSetMiddleware(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetMiddleware returns the middleware configuration.
// GET /api/v1/middleware
func (s *Server) handleGetMiddleware(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetMiddleware()
	if cfg == nil {
		cfg = &config.MiddlewareConfig{
			Enabled:     false,
			Middlewares: []*config.MiddlewareEntry{},
		}
	}

	// Get available builtins from registry
	registry := middleware.GetGlobalRegistry()
	availableBuiltins := make(map[string]bool)
	if registry != nil {
		for _, name := range registry.AvailableBuiltins() {
			availableBuiltins[name] = true
		}
	}

	// Build response with middleware info
	resp := MiddlewareConfigResponse{
		Enabled:     cfg.Enabled,
		Middlewares: make([]*MiddlewareEntryResponse, 0),
	}

	// Add configured middlewares
	configuredNames := make(map[string]bool)
	for _, entry := range cfg.Middlewares {
		configuredNames[entry.Name] = true
		entryResp := &MiddlewareEntryResponse{
			Name:    entry.Name,
			Enabled: entry.Enabled,
			Source:  entry.Source,
			Config:  entry.Config,
		}
		if entryResp.Source == "" {
			entryResp.Source = "builtin"
		}

		// Get version and description from loaded middleware
		if registry != nil {
			if m := registry.Pipeline().Get(entry.Name); m != nil {
				entryResp.Version = m.Version()
				entryResp.Description = m.Description()
				entryResp.Priority = m.Priority()
			}
		}

		resp.Middlewares = append(resp.Middlewares, entryResp)
	}

	// Add available but not configured builtins
	for name := range availableBuiltins {
		if !configuredNames[name] {
			resp.Middlewares = append(resp.Middlewares, &MiddlewareEntryResponse{
				Name:    name,
				Enabled: false,
				Source:  "builtin",
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleSetMiddleware updates the middleware configuration.
// PUT /api/v1/middleware
func (s *Server) handleSetMiddleware(w http.ResponseWriter, r *http.Request) {
	var req MiddlewareConfigResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	cfg := &config.MiddlewareConfig{
		Enabled:     req.Enabled,
		Middlewares: make([]*config.MiddlewareEntry, len(req.Middlewares)),
	}

	for i, entry := range req.Middlewares {
		cfg.Middlewares[i] = &config.MiddlewareEntry{
			Name:    entry.Name,
			Enabled: entry.Enabled,
			Source:  entry.Source,
			Config:  entry.Config,
		}
	}

	if err := config.SetMiddleware(cfg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reload middleware pipeline
	if registry := middleware.GetGlobalRegistry(); registry != nil {
		if err := registry.Reload(); err != nil {
			s.logger.Printf("Failed to reload middleware: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleMiddlewareByName handles GET for a specific middleware.
// GET /api/v1/middleware/{name}
func (s *Server) handleMiddlewareByName(w http.ResponseWriter, r *http.Request, name string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := config.GetMiddleware()
	if cfg == nil {
		http.Error(w, "middleware not found", http.StatusNotFound)
		return
	}

	for _, entry := range cfg.Middlewares {
		if entry.Name == name {
			resp := &MiddlewareEntryResponse{
				Name:    entry.Name,
				Enabled: entry.Enabled,
				Source:  entry.Source,
				Config:  entry.Config,
			}

			// Get version and description from loaded middleware
			if registry := middleware.GetGlobalRegistry(); registry != nil {
				if m := registry.Pipeline().Get(entry.Name); m != nil {
					resp.Version = m.Version()
					resp.Description = m.Description()
					resp.Priority = m.Priority()
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
	}

	http.Error(w, "middleware not found", http.StatusNotFound)
}

// handleMiddlewareEnable enables a middleware.
// POST /api/v1/middleware/{name}/enable
func (s *Server) handleMiddlewareEnable(w http.ResponseWriter, r *http.Request, name string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	registry := middleware.GetGlobalRegistry()
	if registry == nil {
		http.Error(w, "middleware registry not initialized", http.StatusInternalServerError)
		return
	}

	if err := registry.EnableMiddleware(name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleMiddlewareDisable disables a middleware.
// POST /api/v1/middleware/{name}/disable
func (s *Server) handleMiddlewareDisable(w http.ResponseWriter, r *http.Request, name string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	registry := middleware.GetGlobalRegistry()
	if registry == nil {
		http.Error(w, "middleware registry not initialized", http.StatusInternalServerError)
		return
	}

	if err := registry.DisableMiddleware(name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleMiddlewareReload reloads all middlewares.
// POST /api/v1/middleware/reload
func (s *Server) handleMiddlewareReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	registry := middleware.GetGlobalRegistry()
	if registry == nil {
		http.Error(w, "middleware registry not initialized", http.StatusInternalServerError)
		return
	}

	if err := registry.Reload(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
