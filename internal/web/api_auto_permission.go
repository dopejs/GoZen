package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dopejs/gozen/internal/config"
)

// autoPermissionResponse is the JSON shape for auto-permission settings.
type autoPermissionResponse struct {
	Claude   *config.AutoPermissionConfig `json:"claude"`
	Codex    *config.AutoPermissionConfig `json:"codex"`
	OpenCode *config.AutoPermissionConfig `json:"opencode"`
}

// autoPermissionRequest is the JSON shape for updating a single client's auto-permission.
type autoPermissionRequest struct {
	Enabled bool   `json:"enabled"`
	Mode    string `json:"mode"`
}

func (s *Server) handleAutoPermission(w http.ResponseWriter, r *http.Request) {
	// Extract client from path: /api/v1/auto-permission/{client}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/auto-permission")
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")

	if path == "" {
		// /api/v1/auto-permission — list all
		switch r.Method {
		case http.MethodGet:
			s.getAllAutoPermission(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	// /api/v1/auto-permission/{client}
	client := path
	if !config.IsValidClient(client) {
		writeError(w, http.StatusBadRequest, "invalid client: "+client)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getAutoPermission(w, r, client)
	case http.MethodPut:
		s.updateAutoPermission(w, r, client)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) getAllAutoPermission(w http.ResponseWriter, r *http.Request) {
	store := config.DefaultStore()
	resp := autoPermissionResponse{
		Claude:   store.GetAutoPermission(config.ClientClaude),
		Codex:    store.GetAutoPermission(config.ClientCodex),
		OpenCode: store.GetAutoPermission(config.ClientOpenCode),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) getAutoPermission(w http.ResponseWriter, r *http.Request, client string) {
	store := config.DefaultStore()
	ap := store.GetAutoPermission(client)
	if ap == nil {
		ap = &config.AutoPermissionConfig{}
	}
	writeJSON(w, http.StatusOK, ap)
}

func (s *Server) updateAutoPermission(w http.ResponseWriter, r *http.Request, client string) {
	var req autoPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	ap := &config.AutoPermissionConfig{
		Enabled: req.Enabled,
		Mode:    req.Mode,
	}

	store := config.DefaultStore()
	if err := store.SetAutoPermission(client, ap); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ap)
}
