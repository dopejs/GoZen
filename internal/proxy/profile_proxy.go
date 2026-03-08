package proxy

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

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
	MetricsRecorder MetricsRecorder   // optional, for recording request metrics

	mu    sync.RWMutex
	cache map[string]*ProxyServer // profile name -> cached proxy server
}

// MetricsRecorder is an interface for recording request metrics
type MetricsRecorder interface {
	RecordRequest(provider string, latency time.Duration, err error)
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

	// Auto-detect client format from request path if not explicitly set
	clientFormat := detectClientFormat(route.Remainder, clientType)

	pp.Logger.Printf("[route] profile=%s session=%s client=%s format=%s path=%s",
		route.Profile, route.SessionID, clientType, clientFormat, route.Remainder)

	// Register session with bot bridge (for task list visibility)
	if bridge := GetBotBridge(); bridge != nil {
		bridge.MarkSessionBusy(route.CacheKey(), clientType)
	}

	// Resolve profile config (providers + routing)
	profileCfg, err := pp.resolveProfileConfig(route)
	if err != nil {
		pp.writeError(w, http.StatusNotFound, "profile_not_found", err.Error())
		return
	}

	// Build default providers from config
	providers, err := pp.buildProviders(profileCfg.providers)
	if err != nil {
		pp.writeError(w, http.StatusInternalServerError, "provider_error", err.Error())
		return
	}

	// Build routing config if scenario routing is configured
	var routing *RoutingConfig
	if profileCfg.routing != nil && len(profileCfg.routing) > 0 {
		scenarioRoutes := make(map[config.Scenario]*ScenarioProviders)
		for scenario, sr := range profileCfg.routing {
			scenarioProviders, err := pp.buildProviders(sr.ProviderNames())
			if err != nil {
				pp.Logger.Printf("[routing] warning: failed to build providers for scenario %s: %v", scenario, err)
				continue
			}
			models := make(map[string]string)
			for _, pr := range sr.Providers {
				if pr != nil && pr.Model != "" {
					models[pr.Name] = pr.Model
				}
			}
			scenarioRoutes[scenario] = &ScenarioProviders{
				Providers: scenarioProviders,
				Models:    models,
			}
		}
		if len(scenarioRoutes) > 0 {
			routing = &RoutingConfig{
				DefaultProviders:     providers,
				ScenarioRoutes:       scenarioRoutes,
				LongContextThreshold: profileCfg.longContextThreshold,
			}
		}
	}

	// Get or create a proxy server for this profile
	srv := pp.getOrCreateProxy(route.Profile, providers, routing)

	// Rewrite the request URL to strip profile/session prefix
	r.URL.Path = route.Remainder
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}

	// Override session ID extraction: use the route's cache key instead of body parsing
	r.Header.Set("X-Zen-Session", route.CacheKey())

	// Pass request format to ProxyServer (detected per-request, not cached)
	r.Header.Set("X-Zen-Request-Format", clientFormat)

	// Pass client type to ProxyServer for logging
	if clientType != "" {
		r.Header.Set("X-Zen-Client", clientType)
	}

	// Wrap response writer to capture status code
	mrw := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	srv.ServeHTTP(mrw, r)

	// Note: Metrics are recorded by the underlying ProxyServer with the correct provider name.
	// We don't record here to avoid double-counting and incorrect provider attribution.
}

// profileInfo holds resolved profile data for proxy construction.
type profileInfo struct {
	providers            []string
	routing              map[config.Scenario]*config.ScenarioRoute
	longContextThreshold int
}

// resolveProfileConfig looks up provider names and routing config for a profile.
func (pp *ProfileProxy) resolveProfileConfig(route *RouteInfo) (*profileInfo, error) {
	if route.IsTempProfile() {
		if pp.TempProfiles == nil {
			return nil, fmt.Errorf("temporary profile %q: temp profiles not supported", route.Profile)
		}
		names := pp.TempProfiles.GetTempProfileProviders(route.Profile)
		if len(names) == 0 {
			return nil, fmt.Errorf("temporary profile %q not found or expired", route.Profile)
		}
		return &profileInfo{providers: names}, nil
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
	return &profileInfo{
		providers:            pc.Providers,
		routing:              pc.Routing,
		longContextThreshold: pc.LongContextThreshold,
	}, nil
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

		// Only fill Anthropic default model names for Anthropic providers.
		// OpenAI providers should leave empty tier fields as-is so mapModel()
		// falls through to the provider's default model.
		isAnthropic := pc.GetType() == config.ProviderTypeAnthropic

		model := pc.Model
		if model == "" && isAnthropic {
			model = "claude-sonnet-4-5"
		}
		reasoningModel := pc.ReasoningModel
		if reasoningModel == "" && isAnthropic {
			reasoningModel = "claude-sonnet-4-5-thinking"
		}
		haikuModel := pc.HaikuModel
		if haikuModel == "" && isAnthropic {
			haikuModel = "claude-haiku-4-5"
		}
		opusModel := pc.OpusModel
		if opusModel == "" && isAnthropic {
			opusModel = "claude-opus-4-5"
		}
		sonnetModel := pc.SonnetModel
		if sonnetModel == "" && isAnthropic {
			sonnetModel = "claude-sonnet-4-5"
		}

		if !isAnthropic {
			pp.Logger.Printf("[%s] openai provider: using model=%q, skipping Anthropic tier defaults", name, model)
		}

		p := &Provider{
			Name:            name,
			Type:            pc.GetType(),
			BaseURL:         baseURL,
			Token:           pc.AuthToken,
			Model:           model,
			ReasoningModel:  reasoningModel,
			HaikuModel:      haikuModel,
			OpusModel:       opusModel,
			SonnetModel:     sonnetModel,
			EnvVars:         pc.EnvVars,
			ClaudeEnvVars:   pc.ClaudeEnvVars,
			CodexEnvVars:    pc.CodexEnvVars,
			OpenCodeEnvVars: pc.OpenCodeEnvVars,
			ProxyURL:        pc.ProxyURL,
			Healthy:         true,
		}

		// Create per-provider HTTP client if proxy is configured
		if pc.ProxyURL != "" {
			client, err := NewHTTPClientWithProxy(pc.ProxyURL, 10*time.Minute)
			if err != nil {
				pp.Logger.Printf("[%s] warning: failed to create proxy client: %v", name, err)
			} else {
				p.Client = client
			}
		}

		providers = append(providers, p)
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no valid providers")
	}
	return providers, nil
}

// getOrCreateProxy returns a cached ProxyServer for the profile, or creates one.
func (pp *ProfileProxy) getOrCreateProxy(profile string, providers []*Provider, routing *RoutingConfig) *ProxyServer {
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

	var srv *ProxyServer
	if routing != nil {
		srv = NewProxyServerWithRouting(routing, pp.Logger)
	} else {
		srv = NewProxyServer(providers, pp.Logger)
	}
	// Set concurrency limiter (100 concurrent requests as per spec)
	srv.Limiter = NewLimiter(100)
	pp.cache[profile] = srv
	return srv
}

// InvalidateCache removes a profile from the proxy cache.
// Called when config is reloaded.
func (pp *ProfileProxy) InvalidateCache() {
	pp.mu.Lock()
	defer pp.mu.Unlock()
	for _, srv := range pp.cache {
		if srv != nil {
			srv.Close()
		}
	}
	pp.cache = make(map[string]*ProxyServer)
}

func (pp *ProfileProxy) Close() {
	pp.InvalidateCache()
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

// detectClientFormat determines the client API format based on request path and client type.
// OpenAI clients use /responses or /v1/chat/completions endpoints.
// Anthropic clients use /v1/messages endpoint.
func detectClientFormat(path, clientType string) string {
	// If client type is explicitly set, use it
	if clientType == "codex" {
		return config.ProviderTypeOpenAI
	}

	// Auto-detect from path
	// OpenAI Responses API: /responses
	if strings.HasSuffix(path, "/responses") || strings.Contains(path, "/responses/") {
		return config.ProviderTypeOpenAI
	}
	// OpenAI Chat Completions API: /v1/chat/completions
	if strings.HasSuffix(path, "/chat/completions") {
		return config.ProviderTypeOpenAI
	}

	// Default to Anthropic
	return config.ProviderTypeAnthropic
}

// metricsResponseWriter wraps http.ResponseWriter to capture status code
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (m *metricsResponseWriter) WriteHeader(code int) {
	m.statusCode = code
	m.ResponseWriter.WriteHeader(code)
}

// metricsError represents an error for metrics recording
type metricsError struct {
	statusCode int
}

func (e *metricsError) Error() string {
	return fmt.Sprintf("HTTP %d", e.statusCode)
}
