package web

import (
	"net/http"
	"sort"
	"strings"

	"github.com/dopejs/gozen/internal/config"
)

// providerResponse is the JSON shape returned for a single provider.
type providerResponse struct {
	Name            string                     `json:"name"`
	Type            string                     `json:"type,omitempty"`
	BaseURL         string                     `json:"base_url"`
	AuthToken       string                     `json:"auth_token"`
	ProxyURL        string                     `json:"proxy_url,omitempty"`
	Model           string                     `json:"model,omitempty"`
	ReasoningModel  string                     `json:"reasoning_model,omitempty"`
	HaikuModel      string                     `json:"haiku_model,omitempty"`
	OpusModel       string                     `json:"opus_model,omitempty"`
	SonnetModel     string                     `json:"sonnet_model,omitempty"`
	EnvVars         map[string]string          `json:"env_vars,omitempty"`
	ClaudeEnvVars   map[string]string          `json:"claude_env_vars,omitempty"`
	CodexEnvVars    map[string]string          `json:"codex_env_vars,omitempty"`
	OpenCodeEnvVars map[string]string          `json:"opencode_env_vars,omitempty"`
	Disabled        *config.UnavailableMarking `json:"disabled,omitempty"`
}

type createProviderRequest struct {
	Name          string                `json:"name"`
	Config        config.ProviderConfig `json:"config"`
	AddToProfiles []string              `json:"add_to_profiles,omitempty"`
}

func toProviderResponse(name string, p *config.ProviderConfig, mask bool) providerResponse {
	token := p.AuthToken
	if mask {
		token = maskToken(token)
	}
	resp := providerResponse{
		Name:            name,
		Type:            p.Type,
		BaseURL:         p.BaseURL,
		AuthToken:       token,
		ProxyURL:        p.ProxyURL,
		Model:           p.Model,
		ReasoningModel:  p.ReasoningModel,
		HaikuModel:      p.HaikuModel,
		OpusModel:       p.OpusModel,
		SonnetModel:     p.SonnetModel,
		EnvVars:         p.EnvVars,
		ClaudeEnvVars:   p.ClaudeEnvVars,
		CodexEnvVars:    p.CodexEnvVars,
		OpenCodeEnvVars: p.OpenCodeEnvVars,
	}
	// Include active disabled status
	disabled := config.DefaultStore().GetDisabledProviders()
	if m, ok := disabled[name]; ok {
		resp.Disabled = m
	}
	return resp
}

// handleProviders handles GET /api/v1/providers and POST /api/v1/providers.
func (s *Server) handleProviders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listProviders(w, r)
	case http.MethodPost:
		s.createProvider(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleProvider handles GET/PUT/DELETE /api/v1/providers/{name}
// and POST /api/v1/providers/{name}/disable, /api/v1/providers/{name}/enable.
// Also handles GET /api/v1/providers/disabled (list disabled providers).
func (s *Server) handleProvider(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/providers/")
	if path == "" {
		writeError(w, http.StatusBadRequest, "provider name required")
		return
	}

	// Handle GET /api/v1/providers/disabled
	if path == "disabled" && r.Method == http.MethodGet {
		s.handleDisabledProviders(w, r)
		return
	}

	// Check for /disable and /enable sub-paths
	if strings.HasSuffix(path, "/disable") {
		name := strings.TrimSuffix(path, "/disable")
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		s.handleProviderDisable(w, r, name)
		return
	}
	if strings.HasSuffix(path, "/enable") {
		name := strings.TrimSuffix(path, "/enable")
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		s.handleProviderEnable(w, r, name)
		return
	}

	name := path
	switch r.Method {
	case http.MethodGet:
		s.getProvider(w, r, name)
	case http.MethodPut:
		s.updateProvider(w, r, name)
	case http.MethodDelete:
		s.deleteProvider(w, r, name)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) listProviders(w http.ResponseWriter, r *http.Request) {
	store := config.DefaultStore()
	names := store.ProviderNames()
	providers := make([]providerResponse, 0, len(names))
	for _, name := range names {
		p := store.GetProvider(name)
		if p != nil {
			providers = append(providers, toProviderResponse(name, p, true))
		}
	}
	writeJSON(w, http.StatusOK, providers)
}

func (s *Server) getProvider(w http.ResponseWriter, r *http.Request, name string) {
	store := config.DefaultStore()
	p := store.GetProvider(name)
	if p == nil {
		writeError(w, http.StatusNotFound, "provider not found")
		return
	}
	writeJSON(w, http.StatusOK, toProviderResponse(name, p, false))
}

func (s *Server) createProvider(w http.ResponseWriter, r *http.Request) {
	var req createProviderRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Decrypt auth token if encrypted
	if s.keys != nil && req.Config.AuthToken != "" {
		decrypted, err := s.keys.MaybeDecryptToken(req.Config.AuthToken)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to decrypt auth token")
			return
		}
		req.Config.AuthToken = decrypted
	}

	store := config.DefaultStore()
	if store.GetProvider(req.Name) != nil {
		writeError(w, http.StatusConflict, "provider already exists")
		return
	}

	// Validate proxy URL if provided
	if err := config.ValidateProxyURL(req.Config.ProxyURL); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := store.SetProvider(req.Name, &req.Config); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Add provider to requested profiles
	for _, profile := range req.AddToProfiles {
		order := store.GetProfileOrder(profile)
		if order != nil {
			order = append(order, req.Name)
			store.SetProfileOrder(profile, order)
		}
	}

	writeJSON(w, http.StatusCreated, toProviderResponse(req.Name, &req.Config, false))
}

func (s *Server) updateProvider(w http.ResponseWriter, r *http.Request, name string) {
	store := config.DefaultStore()
	existing := store.GetProvider(name)
	if existing == nil {
		writeError(w, http.StatusNotFound, "provider not found")
		return
	}

	var update config.ProviderConfig
	if err := readJSON(r, &update); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Decrypt auth token if encrypted
	if s.keys != nil && update.AuthToken != "" {
		decrypted, err := s.keys.MaybeDecryptToken(update.AuthToken)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to decrypt auth token")
			return
		}
		update.AuthToken = decrypted
	}

	// If token is empty, keep the original.
	if update.AuthToken == "" {
		update.AuthToken = existing.AuthToken
	}

	if update.BaseURL != "" {
		existing.BaseURL = update.BaseURL
	}
	if update.AuthToken != "" {
		existing.AuthToken = update.AuthToken
	}
	existing.Type = update.Type
	existing.Model = update.Model
	existing.ReasoningModel = update.ReasoningModel
	existing.HaikuModel = update.HaikuModel
	existing.OpusModel = update.OpusModel
	existing.SonnetModel = update.SonnetModel
	existing.EnvVars = update.EnvVars
	existing.ClaudeEnvVars = update.ClaudeEnvVars
	existing.CodexEnvVars = update.CodexEnvVars
	existing.OpenCodeEnvVars = update.OpenCodeEnvVars

	// Validate and apply proxy URL
	if err := config.ValidateProxyURL(update.ProxyURL); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	existing.ProxyURL = update.ProxyURL

	if err := store.SetProvider(name, existing); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toProviderResponse(name, existing, false))
}

func (s *Server) deleteProvider(w http.ResponseWriter, r *http.Request, name string) {
	store := config.DefaultStore()
	if store.GetProvider(name) == nil {
		writeError(w, http.StatusNotFound, "provider not found")
		return
	}

	if err := store.DeleteProvider(name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// handleProviderDisable handles POST /api/v1/providers/{name}/disable.
func (s *Server) handleProviderDisable(w http.ResponseWriter, r *http.Request, name string) {
	store := config.DefaultStore()
	if store.GetProvider(name) == nil {
		writeError(w, http.StatusNotFound, "provider '"+name+"' not found")
		return
	}

	var req struct {
		Type string `json:"type"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Validate marking type
	switch req.Type {
	case config.MarkingTypeToday, config.MarkingTypeMonth, config.MarkingTypePermanent:
		// valid
	default:
		writeError(w, http.StatusBadRequest, "invalid type: must be 'today', 'month', or 'permanent'")
		return
	}

	if err := store.DisableProvider(name, req.Type); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get the created marking for response
	disabled := store.GetDisabledProviders()
	marking := disabled[name]

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"provider":   name,
		"disabled":   true,
		"type":       marking.Type,
		"created_at": marking.CreatedAt,
		"expires_at": marking.ExpiresAt,
	})
}

// handleProviderEnable handles POST /api/v1/providers/{name}/enable.
func (s *Server) handleProviderEnable(w http.ResponseWriter, r *http.Request, name string) {
	store := config.DefaultStore()
	if store.GetProvider(name) == nil {
		writeError(w, http.StatusNotFound, "provider '"+name+"' not found")
		return
	}

	if err := store.EnableProvider(name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"provider": name,
		"disabled": false,
	})
}

// handleDisabledProviders handles GET /api/v1/providers/disabled.
func (s *Server) handleDisabledProviders(w http.ResponseWriter, r *http.Request) {
	disabled := config.GetDisabledProviders()

	type disabledEntry struct {
		Provider  string `json:"provider"`
		Type      string `json:"type"`
		CreatedAt string `json:"created_at"`
		ExpiresAt string `json:"expires_at,omitempty"`
	}

	entries := make([]disabledEntry, 0, len(disabled))
	// Sort by provider name for consistent output
	names := make([]string, 0, len(disabled))
	for name := range disabled {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		m := disabled[name]
		entry := disabledEntry{
			Provider:  name,
			Type:      m.Type,
			CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
		}
		if !m.ExpiresAt.IsZero() {
			entry.ExpiresAt = m.ExpiresAt.Format("2006-01-02T15:04:05-07:00")
		}
		entries = append(entries, entry)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"disabled_providers": entries,
	})
}
