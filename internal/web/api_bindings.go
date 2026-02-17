package web

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/dopejs/gozen/internal/config"
)

// bindingResponse is the JSON shape for a single project binding.
type bindingResponse struct {
	Path    string `json:"path"`
	Profile string `json:"profile"`
	Client  string `json:"client"`
}

// bindingsResponse is the JSON shape for listing all bindings.
type bindingsResponse struct {
	Bindings []bindingResponse `json:"bindings"`
	Profiles []string          `json:"profiles"`
	Clients  []string          `json:"clients"`
}

// bindingRequest is the JSON shape for creating/updating a binding.
type bindingRequest struct {
	Path    string `json:"path"`
	Profile string `json:"profile"`
	Client  string `json:"client"`
}

func (s *Server) handleBindings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listBindings(w, r)
	case http.MethodPost:
		s.createBinding(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleBinding(w http.ResponseWriter, r *http.Request) {
	// Extract path from URL: /api/v1/bindings/{encoded-path}
	pathPart := strings.TrimPrefix(r.URL.Path, "/api/v1/bindings/")
	if pathPart == "" {
		writeError(w, http.StatusBadRequest, "path required")
		return
	}

	// URL decode the path
	decodedPath, err := url.PathUnescape(pathPart)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid path encoding")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getBinding(w, r, decodedPath)
	case http.MethodPut:
		s.updateBinding(w, r, decodedPath)
	case http.MethodDelete:
		s.deleteBinding(w, r, decodedPath)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) listBindings(w http.ResponseWriter, r *http.Request) {
	store := config.DefaultStore()
	allBindings := store.GetAllProjectBindings()
	profiles := store.ListProfiles()

	bindings := make([]bindingResponse, 0, len(allBindings))
	for path, b := range allBindings {
		bindings = append(bindings, bindingResponse{
			Path:    path,
			Profile: b.Profile,
			Client:  b.Client,
		})
	}

	writeJSON(w, http.StatusOK, bindingsResponse{
		Bindings: bindings,
		Profiles: profiles,
		Clients:  config.AvailableClients,
	})
}

func (s *Server) getBinding(w http.ResponseWriter, r *http.Request, path string) {
	store := config.DefaultStore()
	binding := store.GetProjectBinding(path)
	if binding == nil {
		writeError(w, http.StatusNotFound, "binding not found")
		return
	}

	writeJSON(w, http.StatusOK, bindingResponse{
		Path:    path,
		Profile: binding.Profile,
		Client:  binding.Client,
	})
}

func (s *Server) createBinding(w http.ResponseWriter, r *http.Request) {
	var req bindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.Path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	store := config.DefaultStore()

	// Verify profile exists if specified
	if req.Profile != "" {
		if store.GetProfileOrder(req.Profile) == nil {
			writeError(w, http.StatusBadRequest, "profile not found")
			return
		}
	}

	// Verify CLI is valid if specified
	if req.Client != "" {
		if !config.IsValidClient(req.Client) {
			writeError(w, http.StatusBadRequest, "invalid CLI")
			return
		}
	}

	if err := store.BindProject(req.Path, req.Profile, req.Client); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, bindingResponse{
		Path:    req.Path,
		Profile: req.Profile,
		Client:  req.Client,
	})
}

func (s *Server) updateBinding(w http.ResponseWriter, r *http.Request, path string) {
	var req bindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	store := config.DefaultStore()

	// Verify profile exists if specified
	if req.Profile != "" {
		if store.GetProfileOrder(req.Profile) == nil {
			writeError(w, http.StatusBadRequest, "profile not found")
			return
		}
	}

	// Verify CLI is valid if specified
	if req.Client != "" {
		if !config.IsValidClient(req.Client) {
			writeError(w, http.StatusBadRequest, "invalid CLI")
			return
		}
	}

	if err := store.BindProject(path, req.Profile, req.Client); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, bindingResponse{
		Path:    path,
		Profile: req.Profile,
		Client:  req.Client,
	})
}

func (s *Server) deleteBinding(w http.ResponseWriter, r *http.Request, path string) {
	store := config.DefaultStore()

	// Check if binding exists
	if store.GetProjectBinding(path) == nil {
		writeError(w, http.StatusNotFound, "binding not found")
		return
	}

	if err := store.UnbindProject(path); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
