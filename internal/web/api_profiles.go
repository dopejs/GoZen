package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dopejs/gozen/internal/config"
)

// providerRouteResponse is the JSON shape for a provider route.
type providerRouteResponse struct {
	Name  string `json:"name"`
	Model string `json:"model,omitempty"`
}

// scenarioRouteResponse is the JSON shape for a scenario route.
type scenarioRouteResponse struct {
	Providers            []*providerRouteResponse `json:"providers"`
	Strategy             *string                  `json:"strategy,omitempty"`
	ProviderWeights      map[string]int           `json:"provider_weights,omitempty"`
	LongContextThreshold *int                     `json:"long_context_threshold,omitempty"`
	FallbackToDefault    *bool                    `json:"fallback_to_default,omitempty"`
}

// profileResponse is the JSON shape returned for a single profile.
type profileResponse struct {
	Name             string                             `json:"name"`
	Providers        []string                           `json:"providers"`
	Routing          map[string]*scenarioRouteResponse `json:"routing,omitempty"`
	ScenarioPriority []string                           `json:"scenario_priority,omitempty"`
}

type createProfileRequest struct {
	Name             string                             `json:"name"`
	Providers        []string                           `json:"providers"`
	Routing          map[string]*scenarioRouteResponse `json:"routing,omitempty"`
	ScenarioPriority []string                           `json:"scenario_priority,omitempty"`
}

type updateProfileRequest struct {
	Providers        []string                           `json:"providers"`
	Routing          map[string]*scenarioRouteResponse `json:"routing,omitempty"`
	ScenarioPriority []string                           `json:"scenario_priority,omitempty"`
}

// profileConfigToResponse converts a ProfileConfig to a profileResponse.
func profileConfigToResponse(name string, pc *config.ProfileConfig) profileResponse {
	providers := pc.Providers
	if providers == nil {
		providers = []string{}
	}
	resp := profileResponse{
		Name:             name,
		Providers:        providers,
		ScenarioPriority: pc.ScenarioPriority,
	}
	if len(pc.Routing) > 0 {
		resp.Routing = make(map[string]*scenarioRouteResponse)
		for scenario, route := range pc.Routing {
			var providerRoutes []*providerRouteResponse
			for _, pr := range route.Providers {
				providerRoutes = append(providerRoutes, &providerRouteResponse{
					Name:  pr.Name,
					Model: pr.Model,
				})
			}

			scenarioResp := &scenarioRouteResponse{
				Providers: providerRoutes,
			}

			// Serialize strategy (convert LoadBalanceStrategy to string)
			if route.Strategy != "" {
				strategyStr := string(route.Strategy)
				scenarioResp.Strategy = &strategyStr
			}

			// Serialize provider weights
			if len(route.ProviderWeights) > 0 {
				scenarioResp.ProviderWeights = route.ProviderWeights
			}

			// Serialize long context threshold
			if route.LongContextThreshold != nil {
				scenarioResp.LongContextThreshold = route.LongContextThreshold
			}

			// Serialize fallback to default
			if route.FallbackToDefault != nil {
				scenarioResp.FallbackToDefault = route.FallbackToDefault
			}

			resp.Routing[scenario] = scenarioResp
		}
	}
	return resp
}

// routingResponseToConfig converts routing response data to config RoutePolicy map.
func routingResponseToConfig(routing map[string]*scenarioRouteResponse) map[string]*config.RoutePolicy {
	if len(routing) == 0 {
		return nil
	}
	result := make(map[string]*config.RoutePolicy)
	for scenario, route := range routing {
		if len(route.Providers) > 0 {
			var providerRoutes []*config.ProviderRoute
			for _, pr := range route.Providers {
				providerRoutes = append(providerRoutes, &config.ProviderRoute{
					Name:  pr.Name,
					Model: pr.Model,
				})
			}

			policy := &config.RoutePolicy{
				Providers: providerRoutes,
			}

			// Deserialize strategy (convert string to LoadBalanceStrategy)
			if route.Strategy != nil && *route.Strategy != "" {
				policy.Strategy = config.LoadBalanceStrategy(*route.Strategy)
			}

			// Deserialize provider weights
			if len(route.ProviderWeights) > 0 {
				policy.ProviderWeights = route.ProviderWeights
			}

			// Deserialize long context threshold
			if route.LongContextThreshold != nil {
				policy.LongContextThreshold = route.LongContextThreshold
			}

			// Deserialize fallback to default
			if route.FallbackToDefault != nil {
				policy.FallbackToDefault = route.FallbackToDefault
			}

			result[scenario] = policy
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// handleProfiles handles GET /api/v1/profiles and POST /api/v1/profiles.
func (s *Server) handleProfiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listProfiles(w, r)
	case http.MethodPost:
		s.createProfile(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleProfile handles GET/PUT/DELETE /api/v1/profiles/{name}.
func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/v1/profiles/")
	if name == "" {
		writeError(w, http.StatusBadRequest, "profile name required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getProfile(w, r, name)
	case http.MethodPut:
		s.updateProfile(w, r, name)
	case http.MethodDelete:
		s.deleteProfile(w, r, name)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) listProfiles(w http.ResponseWriter, r *http.Request) {
	store := config.DefaultStore()
	names := store.ListProfiles()
	profiles := make([]profileResponse, 0, len(names))
	for _, name := range names {
		pc := store.GetProfileConfig(name)
		if pc == nil {
			pc = &config.ProfileConfig{Providers: []string{}}
		}
		profiles = append(profiles, profileConfigToResponse(name, pc))
	}
	writeJSON(w, http.StatusOK, profiles)
}

func (s *Server) getProfile(w http.ResponseWriter, r *http.Request, name string) {
	store := config.DefaultStore()
	pc := store.GetProfileConfig(name)
	if pc == nil {
		writeError(w, http.StatusNotFound, "profile not found")
		return
	}
	writeJSON(w, http.StatusOK, profileConfigToResponse(name, pc))
}

func (s *Server) createProfile(w http.ResponseWriter, r *http.Request) {
	var req createProfileRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	store := config.DefaultStore()
	existing := store.GetProfileConfig(req.Name)
	if existing != nil {
		writeError(w, http.StatusConflict, "profile already exists")
		return
	}

	providers := req.Providers
	if providers == nil {
		providers = []string{}
	}

	pc := &config.ProfileConfig{
		Providers:        providers,
		Routing:          routingResponseToConfig(req.Routing),
		ScenarioPriority: req.ScenarioPriority,
	}

	if err := store.SetProfileConfig(req.Name, pc); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, profileConfigToResponse(req.Name, pc))
}

func (s *Server) updateProfile(w http.ResponseWriter, r *http.Request, name string) {
	store := config.DefaultStore()
	existing := store.GetProfileConfig(name)
	if existing == nil {
		writeError(w, http.StatusNotFound, "profile not found")
		return
	}

	var req updateProfileRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	providers := req.Providers
	if providers == nil {
		providers = []string{}
	}

	existing.Providers = providers
	existing.Routing = routingResponseToConfig(req.Routing)
	existing.ScenarioPriority = req.ScenarioPriority

	if err := store.SetProfileConfig(name, existing); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, profileConfigToResponse(name, existing))
}

func (s *Server) deleteProfile(w http.ResponseWriter, r *http.Request, name string) {
	store := config.DefaultStore()

	// Check if this is the default profile
	defaultProfile := store.GetDefaultProfile()
	if name == defaultProfile {
		writeError(w, http.StatusForbidden, fmt.Sprintf("cannot delete the default profile '%s'", name))
		return
	}

	existing := store.GetProfileOrder(name)
	if existing == nil {
		writeError(w, http.StatusNotFound, "profile not found")
		return
	}

	if err := store.DeleteProfile(name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
