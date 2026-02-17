package web

import (
	"encoding/json"
	"net/http"

	"github.com/dopejs/gozen/internal/config"
)

// settingsResponse is the JSON shape for global settings.
type settingsResponse struct {
	DefaultProfile string   `json:"default_profile"`
	DefaultClient  string   `json:"default_client"`
	WebPort        int      `json:"web_port"`
	Profiles       []string `json:"profiles"`
	Clients        []string `json:"clients"`
}

// settingsRequest is the JSON shape for updating settings.
type settingsRequest struct {
	DefaultProfile string `json:"default_profile,omitempty"`
	DefaultClient  string `json:"default_client,omitempty"`
	WebPort        int    `json:"web_port,omitempty"`
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getSettings(w, r)
	case http.MethodPut:
		s.updateSettings(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) getSettings(w http.ResponseWriter, r *http.Request) {
	store := config.DefaultStore()
	profiles := store.ListProfiles()

	resp := settingsResponse{
		DefaultProfile: store.GetDefaultProfile(),
		DefaultClient:  store.GetDefaultClient(),
		WebPort:        store.GetWebPort(),
		Profiles:       profiles,
		Clients:        config.AvailableClients,
	}
	writeJSON(w, http.StatusOK, resp)
}
func (s *Server) updateSettings(w http.ResponseWriter, r *http.Request) {
	var req settingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	store := config.DefaultStore()

	if req.DefaultProfile != "" {
		if store.GetProfileOrder(req.DefaultProfile) == nil {
			writeError(w, http.StatusBadRequest, "profile not found")
			return
		}
		if err := store.SetDefaultProfile(req.DefaultProfile); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	if req.DefaultClient != "" {
		if !config.IsValidClient(req.DefaultClient) {
			writeError(w, http.StatusBadRequest, "invalid client")
			return
		}
		if err := store.SetDefaultClient(req.DefaultClient); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	if req.WebPort > 0 {
		if req.WebPort < 1024 || req.WebPort > 65535 {
			writeError(w, http.StatusBadRequest, "port must be between 1024 and 65535")
			return
		}
		if err := store.SetWebPort(req.WebPort); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	s.getSettings(w, r)
}
