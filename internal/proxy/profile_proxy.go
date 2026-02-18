package proxy

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/dopejs/gozen/internal/config"
)

// TempProfileProvider supplies temporary profile data (from zen pick).
type TempProfileProvider interface {
	GetTempProfileProviders(id string) []string
}

// ProfileProxy is an HTTP handler that routes requests based on profile and session
// extracted from the URL path. It dynamically builds provider chains from config.
type ProfileProxy struct {
	Logger       *log.Logger
	TempProfiles TempProfileProvider // optional, for _tmp_ profiles

	mu    sync.RWMutex
	cache map[string]*ProxyServer // profile name -> cached proxy server
}

// NewProfileProxy creates a new profile-based proxy router.
func NewProfileProxy(logger *log.Logger) *ProfileProxy {
	return &ProfileProxy{
		Logger: logger,
		cache:  make(map[string]*ProxyServer),
	}
}

func (pp *ProfileProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse route from URL path
	route, err := ParseRoutePath(r.URL.Path)
	if err != nil {
		pp.writeError(w, http.StatusBadRequest, "invalid_path",
			fmt.Sprintf("Invalid proxy path: %s. Expected /<profile>/<session>/v1/...", err))
		return
	}

	// Extract and strip X-Zen-Client header (from original request)
	clientType := r.Header.Get("X-Zen-Client")
	r.Header.Del("X-Zen-Client")

	pp.Logger.Printf("[route] profile=%s session=%s client=%s path=%s",
		route.Profile, route.SessionID, clientType, route.Remainder)

	// Resolve provider names for this profile
	providerNames, err := pp.resolveProviderNames(route)
	if err != nil {
		pp.writeError(w, http.StatusNotFound, "profile_not_found", err.Error())
		return
	}

	// Build providers from config
	providers, err := pp.buildProviders(providerNames)
	if err != nil {
		pp.writeError(w, http.StatusInternalServerError, "provider_error", err.Error())
		return
	}

	// Determine client format from the client type header
	clientFormat := config.ProviderTypeAnthropic
	if clientType == "codex" {
		clientFormat = config.ProviderTypeOpenAI
	}

	// Get or create a proxy server for this profile
	srv := pp.getOrCreateProxy(route.Profile, providers, clientFormat)

	// Rewrite the request URL to strip profile/session prefix
	r.URL.Path = route.Remainder
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}

	// Override session ID extraction: use the route's cache key instead of body parsing
	// We do this by setting a context value or header that ServeHTTP can pick up
	r.Header.Set("X-Zen-Session", route.CacheKey())

	// Pass client type to ProxyServer for logging
	if clientType != "" {
		r.Header.Set("X-Zen-Client", clientType)
	}

	srv.ServeHTTP(w, r)
}

// resolveProviderNames looks up provider names for a profile.
func (pp *ProfileProxy) resolveProviderNames(route *RouteInfo) ([]string, error) {
	if route.IsTempProfile() {
		if pp.TempProfiles == nil {
			return nil, fmt.Errorf("temporary profile %q: temp profiles not supported", route.Profile)
		}
		names := pp.TempProfiles.GetTempProfileProviders(route.Profile)
		if len(names) == 0 {
			return nil, fmt.Errorf("temporary profile %q not found or expired", route.Profile)
		}
		return names, nil
	}

	// Look up from config
	store := config.DefaultStore()
	pc := store.GetProfileConfig(route.Profile)
	if pc == nil {
		return nil, fmt.Errorf("profile %q not found", route.Profile)
	}
	if len(pc.Providers) == 0 {
		return nil, fmt.Errorf("profile %q has no providers configured", route.Profile)
	}
	return pc.Providers, nil
}

// buildProviders converts provider names to Provider objects.
func (pp *ProfileProxy) buildProviders(names []string) ([]*Provider, error) {
	store := config.DefaultStore()
	var providers []*Provider

	for _, name := range names {
		pc := store.GetProvider(name)
		if pc == nil {
			return nil, fmt.Errorf("provider %q not found in config", name)
		}

		baseURL, err := url.Parse(pc.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("provider %q: invalid base URL: %w", name, err)
		}

		model := pc.Model
		if model == "" {
			model = "claude-sonnet-4-5"
		}

		providers = append(providers, &Provider{
			Name:            name,
			Type:            pc.GetType(),
			BaseURL:         baseURL,
			Token:           pc.AuthToken,
			Model:           model,
			ReasoningModel:  pc.ReasoningModel,
			HaikuModel:      pc.HaikuModel,
			OpusModel:       pc.OpusModel,
			SonnetModel:     pc.SonnetModel,
			EnvVars:         pc.EnvVars,
			ClaudeEnvVars:   pc.ClaudeEnvVars,
			CodexEnvVars:    pc.CodexEnvVars,
			OpenCodeEnvVars: pc.OpenCodeEnvVars,
			Healthy:         true,
		})
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no valid providers")
	}
	return providers, nil
}

// getOrCreateProxy returns a cached ProxyServer for the profile, or creates one.
func (pp *ProfileProxy) getOrCreateProxy(profile string, providers []*Provider, clientFormat string) *ProxyServer {
	pp.mu.RLock()
	if srv, ok := pp.cache[profile]; ok {
		pp.mu.RUnlock()
		return srv
	}
	pp.mu.RUnlock()

	pp.mu.Lock()
	defer pp.mu.Unlock()

	// Double-check after acquiring write lock
	if srv, ok := pp.cache[profile]; ok {
		return srv
	}

	srv := NewProxyServerWithClientFormat(providers, clientFormat, pp.Logger)
	pp.cache[profile] = srv
	return srv
}

// InvalidateCache removes a profile from the proxy cache.
// Called when config is reloaded.
func (pp *ProfileProxy) InvalidateCache() {
	pp.mu.Lock()
	defer pp.mu.Unlock()
	pp.cache = make(map[string]*ProxyServer)
}

func (pp *ProfileProxy) writeError(w http.ResponseWriter, status int, errType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"type":    errType,
			"message": message,
		},
	})
}
