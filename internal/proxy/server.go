package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/middleware"
	"github.com/dopejs/gozen/internal/proxy/transform"
)

// ProxyError represents a categorized error from the proxy
type ProxyError struct {
	Provider string
	ErrType  string
	Err      error
}

func (e *ProxyError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s [%s]: %v", e.Provider, e.ErrType, e.Err)
	}
	return fmt.Sprintf("%s [%s]", e.Provider, e.ErrType)
}

func (e *ProxyError) Unwrap() error {
	return e.Err
}

// Type returns the error type for metrics classification
func (e *ProxyError) Type() string {
	return e.ErrType
}

// TransformError represents an error during request/response transformation
type TransformError struct {
	Op  string // "request" or "response"
	Err error
}

func (e *TransformError) Error() string {
	return fmt.Sprintf("transform %s failed: %v", e.Op, e.Err)
}

func (e *TransformError) Unwrap() error {
	return e.Err
}

// Error type constants for metrics classification
const (
	ErrorTypeAuth       = "auth"
	ErrorTypeRateLimit  = "rate_limit"
	ErrorTypeRequest    = "request"
	ErrorTypeServer     = "server"
	ErrorTypeNetwork    = "network"
	ErrorTypeTimeout    = "timeout"
	ErrorTypeConcurrency = "concurrency"
)

var (
	globalLogger     *StructuredLogger
	globalLogDB      *LogDB
	globalLoggerOnce sync.Once
	globalLoggerMu   sync.RWMutex
)

// InitGlobalLogger initializes the global structured logger with SQLite storage.
func InitGlobalLogger(logDir string) error {
	var initErr error
	globalLoggerOnce.Do(func() {
		logDB, err := OpenLogDB(logDir)
		if err != nil {
			initErr = err
			return
		}
		logger, err := NewStructuredLogger(logDir, 2000, logDB)
		if err != nil {
			logDB.Close()
			initErr = err
			return
		}
		globalLoggerMu.Lock()
		globalLogger = logger
		globalLogDB = logDB
		globalLoggerMu.Unlock()
	})
	return initErr
}

// GetGlobalLogger returns the global structured logger.
func GetGlobalLogger() *StructuredLogger {
	globalLoggerMu.RLock()
	defer globalLoggerMu.RUnlock()
	return globalLogger
}

// GetGlobalLogDB returns the global log database.
func GetGlobalLogDB() *LogDB {
	globalLoggerMu.RLock()
	defer globalLoggerMu.RUnlock()
	return globalLogDB
}

// RoutingConfig holds the default provider chain and optional scenario routes.
type RoutingConfig struct {
	DefaultProviders     []*Provider
	ScenarioRoutes       map[string]*ScenarioProviders
	LongContextThreshold int // threshold for longContext scenario detection
}

// ScenarioProviders defines the providers and routing policy for a scenario.
type ScenarioProviders struct {
	Providers            []*Provider
	Models               map[string]string // provider name → model override
	Strategy             *config.LoadBalanceStrategy
	ProviderWeights      map[string]int
	LongContextThreshold *int
	FallbackToDefault    *bool
}

// providerFailure tracks details of a failed provider attempt.
type providerFailure struct {
	Name       string
	StatusCode int
	Body       string
	Elapsed    time.Duration
}

type ProxyServer struct {
	Providers        []*Provider
	Routing          *RoutingConfig // optional; nil means use Providers as-is
	Logger           *log.Logger
	StructuredLogger *StructuredLogger
	Client           *http.Client
	Limiter          *Limiter        // optional; nil means unlimited
	MetricsRecorder  MetricsRecorder // optional; for recording request metrics
	Strategy         config.LoadBalanceStrategy // load balancing strategy
	LoadBalancer     *LoadBalancer              // for strategy-based provider selection
	Profile          string                     // profile name for per-profile strategy state
}

func (s *ProxyServer) Close() {
	seenClients := make(map[*http.Client]struct{})
	closeClient := func(client *http.Client) {
		if client == nil {
			return
		}
		if _, ok := seenClients[client]; ok {
			return
		}
		seenClients[client] = struct{}{}
		closeHTTPClientIdleConnections(client)
	}

	closeClient(s.Client)
	for _, provider := range s.allProviders() {
		closeClient(provider.Client)
	}
}

func (s *ProxyServer) allProviders() []*Provider {
	providers := make([]*Provider, 0, len(s.Providers))
	providers = append(providers, s.Providers...)
	if s.Routing == nil {
		return providers
	}
	for _, scenarioProviders := range s.Routing.ScenarioRoutes {
		if scenarioProviders == nil {
			continue
		}
		providers = append(providers, scenarioProviders.Providers...)
	}
	return providers
}

func NewProxyServer(providers []*Provider, logger *log.Logger, strategy config.LoadBalanceStrategy, lb *LoadBalancer) *ProxyServer {
	return &ProxyServer{
		Providers:        providers,
		Logger:           logger,
		StructuredLogger: GetGlobalLogger(),
		Client:           newHTTPClient(10 * time.Minute),
		Strategy:         strategy,
		LoadBalancer:     lb,
	}
}

// NewProxyServerWithRouting creates a proxy server with scenario-based routing.
func NewProxyServerWithRouting(routing *RoutingConfig, logger *log.Logger, strategy config.LoadBalanceStrategy, lb *LoadBalancer) *ProxyServer {
	return &ProxyServer{
		Providers:        routing.DefaultProviders,
		Routing:          routing,
		Logger:           logger,
		StructuredLogger: GetGlobalLogger(),
		Client:           newHTTPClient(10 * time.Minute),
		Strategy:         strategy,
		LoadBalancer:     lb,
	}
}

// isProviderDisabled checks if a provider is manually marked as unavailable via config.
// Uses lazy evaluation — expiration is handled by the config layer.
func (s *ProxyServer) isProviderDisabled(name string) bool {
	return config.IsProviderDisabled(name)
}

// filterDisabledProviders partitions providers into available and disabled lists.
// Returns (available, disabledNames).
func (s *ProxyServer) filterDisabledProviders(providers []*Provider) ([]*Provider, []string) {
	var available []*Provider
	var disabledNames []string
	for _, p := range providers {
		if s.isProviderDisabled(p.Name) {
			disabledNames = append(disabledNames, p.Name)
		} else {
			available = append(available, p)
		}
	}
	return available, disabledNames
}

// writeAllProvidersUnavailableError writes a 503 JSON error response when all providers are disabled.
func (s *ProxyServer) writeAllProvidersUnavailableError(w http.ResponseWriter, disabledNames []string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	errResp := map[string]interface{}{
		"error": map[string]interface{}{
			"type":               "all_providers_unavailable",
			"message":            "All providers are manually marked as unavailable. Please re-enable a provider via Web UI or 'zen enable <provider>'.",
			"disabled_providers": disabledNames,
		},
	}
	json.NewEncoder(w).Encode(errResp)
}

func (s *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestStart := time.Now()

	// Acquire concurrency slot if limiter is configured
	// Pass request context so limiter respects client cancellation
	if s.Limiter != nil {
		if err := s.Limiter.Acquire(r.Context()); err != nil {
			// Record concurrency limit error in metrics
			// Use empty provider to indicate system-level error (not provider-specific)
			if s.MetricsRecorder != nil {
				s.MetricsRecorder.RecordRequest("", time.Since(requestStart), &ProxyError{
					Provider: "",
					ErrType:  ErrorTypeConcurrency,
					Err:      err,
				})
			}
			http.Error(w, "service unavailable: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
		defer s.Limiter.Release()
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadGateway)
		return
	}
	r.Body.Close()

	// Determine session ID:
	// 1. X-Zen-Session header (set by ProfileProxy with <profile>:<session> key)
	// 2. Fallback: extract from request body metadata (legacy per-invocation proxy)
	sessionID := r.Header.Get("X-Zen-Session")
	r.Header.Del("X-Zen-Session")
	if sessionID == "" {
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &bodyMap); err == nil {
			sessionID = extractSessionID(bodyMap)
		}
	}

	// Extract client type for logging (set by ProfileProxy)
	clientType := r.Header.Get("X-Zen-Client")
	r.Header.Del("X-Zen-Client")

	// Mark session as busy in bot bridge
	if bridge := GetBotBridge(); bridge != nil && sessionID != "" {
		bridge.MarkSessionBusy(sessionID, clientType)
	}

	// Extract request format (detected per-request by ProfileProxy)
	requestFormat := r.Header.Get("X-Zen-Request-Format")
	r.Header.Del("X-Zen-Request-Format")
	if requestFormat == "" {
		requestFormat = config.ProviderTypeAnthropic // Default
	}

	// Detect protocol and normalize request for routing (T023-T024)
	var bodyMap map[string]interface{}
	var normalized *NormalizedRequest
	var features *RequestFeatures
	if err := json.Unmarshal(bodyBytes, &bodyMap); err == nil {
		// Detect protocol using priority: URL path → header → body structure
		detectedProtocol := DetectProtocol(r.URL.Path, r.Header, bodyMap)

		// Normalize request based on detected protocol
		var normErr error
		switch detectedProtocol {
		case "anthropic":
			normalized, normErr = NormalizeAnthropicMessages(bodyMap)
		case "openai_chat":
			normalized, normErr = NormalizeOpenAIChat(bodyMap)
		case "openai_responses":
			normalized, normErr = NormalizeOpenAIResponses(bodyMap)
		default:
			// Unknown protocol, try anthropic as fallback
			normalized, normErr = NormalizeAnthropicMessages(bodyMap)
		}

		// Log normalization error but continue (T025: route to default on failure)
		if normErr != nil {
			s.Logger.Printf("[routing] normalization error for protocol %s: %v", detectedProtocol, normErr)
		}

		// Extract features for routing classification
		if normalized != nil {
			features = ExtractFeatures(normalized)
			// T077: Log request features for observability
			if features != nil {
				s.Logger.Printf("[routing] features: has_image=%v, has_tools=%v, is_long_context=%v, total_tokens=%d, message_count=%d",
					features.HasImage, features.HasTools, features.IsLongContext, features.TotalTokens, features.MessageCount)
			}
		}
	}

	// [BETA] Apply context compression if enabled
	if compressor := GetGlobalCompressor(); compressor != nil && compressor.IsEnabled() {
		compressedBody, compressed, err := compressor.CompressRequestBody(bodyBytes)
		if err != nil {
			s.Logger.Printf("[compression] error: %v", err)
		} else if compressed {
			s.Logger.Printf("[compression] compressed request body from %d to %d bytes", len(bodyBytes), len(compressedBody))
			bodyBytes = compressedBody
		}
	}

	// [BETA] Apply middleware pipeline if enabled
	var processedCtx *middleware.RequestContext
	if pipeline := middleware.GetGlobalPipeline(); pipeline != nil && pipeline.IsEnabled() {
		reqCtx := &middleware.RequestContext{
			SessionID:         sessionID,
			ClientType:        clientType,
			Method:            r.Method,
			Path:              r.URL.Path,
			Headers:           r.Header.Clone(),
			Body:              bodyBytes,
			Metadata:          make(map[string]interface{}),
			RequestFormat:     requestFormat,
			NormalizedRequest: normalized,
			Profile:           s.Profile,
		}

		// Parse model and messages for middleware
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &bodyMap); err == nil {
			if model, ok := bodyMap["model"].(string); ok {
				reqCtx.Model = model
			}
			if msgs, ok := bodyMap["messages"].([]interface{}); ok {
				for _, m := range msgs {
					if msgMap, ok := m.(map[string]interface{}); ok {
						role, _ := msgMap["role"].(string)
						reqCtx.Messages = append(reqCtx.Messages, middleware.Message{
							Role:    role,
							Content: msgMap["content"],
						})
					}
				}
			}
		}

		var err error
		processedCtx, err = pipeline.ProcessRequest(reqCtx)
		if err != nil {
			s.Logger.Printf("[middleware] request processing error: %v", err)
			http.Error(w, fmt.Sprintf("middleware error: %v", err), http.StatusBadRequest)
			return
		}
		bodyBytes = processedCtx.Body
	}

	// T034-T036: Extract routing decision and hints from middleware context
	var middlewareDecision *RoutingDecision
	var routingHints *RoutingHints
	if processedCtx != nil {
		if rd, ok := processedCtx.RoutingDecision.(*RoutingDecision); ok {
			middlewareDecision = rd
		}
		if rh, ok := processedCtx.RoutingHints.(*RoutingHints); ok {
			routingHints = rh
		}
	}

	// T035: Resolve routing decision (middleware > builtin classifier)
	threshold := defaultLongContextThreshold
	if s.Routing != nil && s.Routing.LongContextThreshold > 0 {
		threshold = s.Routing.LongContextThreshold
	}

	decision := ResolveRoutingDecision(
		middlewareDecision,
		normalized,
		features,
		routingHints,
		threshold,
		sessionID,
		bodyMap,
	)

	// T036: Log routing decision
	s.Logger.Printf("[routing] scenario=%s, source=%s, reason=%s, confidence=%.2f",
		decision.Scenario, decision.Source, decision.Reason, decision.Confidence)

	// T044-T045: Look up scenario route (with fallback to default)
	providers := s.Providers
	var modelOverrides map[string]string
	var usingScenarioRoute bool
	var scenarioProviders *ScenarioProviders

	if s.Routing != nil && len(s.Routing.ScenarioRoutes) > 0 {
		// Try to find route for the detected scenario
		normalizedScenario := NormalizeScenarioKey(decision.Scenario)

		// Try normalized key first, then original key
		if sp, ok := s.Routing.ScenarioRoutes[normalizedScenario]; ok {
			scenarioProviders = sp
		} else if sp, ok := s.Routing.ScenarioRoutes[decision.Scenario]; ok {
			scenarioProviders = sp
		}

		if scenarioProviders != nil {
			providers = scenarioProviders.Providers
			modelOverrides = scenarioProviders.Models
			usingScenarioRoute = true
			s.Logger.Printf("[routing] using scenario route: providers=%d, model_overrides=%d",
				len(providers), len(modelOverrides))
		} else if decision.Scenario != string(config.ScenarioDefault) {
			s.Logger.Printf("[routing] no route configured for scenario=%s, using default providers", decision.Scenario)
		}
	}

	// Apply RoutingDecision overrides (ModelHint, ProviderAllowlist, ProviderDenylist)
	if decision.ModelHint != nil && *decision.ModelHint != "" {
		// Apply model hint as override for all providers
		if modelOverrides == nil {
			modelOverrides = make(map[string]string)
		}
		for _, p := range providers {
			if _, exists := modelOverrides[p.Name]; !exists {
				modelOverrides[p.Name] = *decision.ModelHint
			}
		}
	}

	// Apply provider allowlist/denylist filters
	if len(decision.ProviderAllowlist) > 0 {
		allowSet := make(map[string]bool)
		for _, name := range decision.ProviderAllowlist {
			allowSet[name] = true
		}
		filtered := make([]*Provider, 0, len(providers))
		for _, p := range providers {
			if allowSet[p.Name] {
				filtered = append(filtered, p)
			}
		}
		providers = filtered
		if len(providers) == 0 {
			s.Logger.Printf("[routing] provider allowlist resulted in no providers")
		}
	}

	if len(decision.ProviderDenylist) > 0 {
		denySet := make(map[string]bool)
		for _, name := range decision.ProviderDenylist {
			denySet[name] = true
		}
		filtered := make([]*Provider, 0, len(providers))
		for _, p := range providers {
			if !denySet[p.Name] {
				filtered = append(filtered, p)
			}
		}
		providers = filtered
		if len(providers) == 0 {
			s.Logger.Printf("[routing] provider denylist resulted in no providers")
		}
	}

	// Filter disabled providers BEFORE strategy selection to avoid polluting
	// round-robin counters, weighted distribution, and least-* rankings.
	availableProviders, disabledNames := s.filterDisabledProviders(providers)
	if len(availableProviders) == 0 && len(disabledNames) > 0 {
		// If using scenario route, try falling back to default providers first
		if usingScenarioRoute && len(s.Providers) > 0 {
			defaultAvailable, defaultDisabledNames := s.filterDisabledProviders(s.Providers)
			if len(defaultAvailable) == 0 && len(defaultDisabledNames) > 0 {
				allDisabled := append(disabledNames, defaultDisabledNames...)
				s.Logger.Printf("[proxy] all providers unavailable (manually disabled): %v", allDisabled)
				s.writeAllProvidersUnavailableError(w, allDisabled)
				return
			}
			// Fall through — default providers have some available, will be tried below
		} else {
			s.Logger.Printf("[proxy] all providers unavailable (manually disabled): %v", disabledNames)
			s.writeAllProvidersUnavailableError(w, disabledNames)
			return
		}
	}
	// Use only non-disabled providers for strategy selection and routing
	providers = availableProviders

	// T055: Apply load balancing strategy to reorder providers
	if s.LoadBalancer != nil && len(providers) > 1 {
		// Extract model from request body for strategy decisions
		var model string
		if bodyMap != nil {
			if m, ok := bodyMap["model"].(string); ok {
				model = m
			}
		}

		// Use per-scenario strategy if available, otherwise use profile default
		strategy := s.Strategy
		var weights map[string]int
		if usingScenarioRoute && scenarioProviders != nil {
			if scenarioProviders.Strategy != nil {
				strategy = *scenarioProviders.Strategy
			}
			if len(scenarioProviders.ProviderWeights) > 0 {
				weights = scenarioProviders.ProviderWeights
			}
		}

		// Apply RoutingDecision strategy override (highest priority)
		if decision.StrategyOverride != nil {
			strategy = *decision.StrategyOverride
		}

		providers = s.LoadBalancer.Select(providers, strategy, model, s.Profile, modelOverrides, weights)
	}

	// Track provider failure details for error reporting
	var failures []providerFailure

	// Try scenario providers first, then fallback to default if all fail
	success := s.tryProviders(w, r, providers, modelOverrides, bodyBytes, sessionID, clientType, requestFormat, &failures, requestStart)
	if success {
		// Log request_received only if duration >1s (selective logging per T067)
		duration := time.Since(requestStart)
		if duration > time.Second {
			s.logRequestReceived(r.Method, r.URL.Path, sessionID, clientType, duration, nil)
		}
		return
	}

	// If scenario route failed and we have default providers to fallback to
	if usingScenarioRoute && len(s.Providers) > 0 {
		s.Logger.Printf("[routing] scenario=%s all providers failed, falling back to default providers", decision.Scenario)
		// Filter disabled providers from defaults
		defaultAvailable, defaultDisabledNames := s.filterDisabledProviders(s.Providers)
		if len(defaultAvailable) == 0 && len(defaultDisabledNames) > 0 {
			s.Logger.Printf("[proxy] all default providers also unavailable (manually disabled): %v", defaultDisabledNames)
			allDisabled := append(disabledNames, defaultDisabledNames...)
			s.writeAllProvidersUnavailableError(w, allDisabled)
			// Log request_received for error (selective logging per T067)
			duration := time.Since(requestStart)
			s.logRequestReceived(r.Method, r.URL.Path, sessionID, clientType, duration, fmt.Errorf("all providers unavailable"))
			return
		}
		// Filter disabled, then apply strategy to defaults (no model overrides for defaults)
		defaultProviders := defaultAvailable
		if s.LoadBalancer != nil && len(defaultProviders) > 1 {
			var model string
			var bodyMap map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &bodyMap); err == nil {
				if m, ok := bodyMap["model"].(string); ok {
					model = m
				}
			}
			defaultProviders = s.LoadBalancer.Select(defaultProviders, s.Strategy, model, s.Profile, nil, nil)
		}
		success = s.tryProviders(w, r, defaultProviders, nil, bodyBytes, sessionID, clientType, requestFormat, &failures, requestStart)
		if success {
			// Log request_received only if duration >1s (selective logging per T067)
			duration := time.Since(requestStart)
			if duration > time.Second {
				s.logRequestReceived(r.Method, r.URL.Path, sessionID, clientType, duration, nil)
			}
			return
		}
	}

	// Build detailed error message with all provider failures
	var errMsg strings.Builder
	errMsg.WriteString("all providers failed\n")
	for _, f := range failures {
		if f.StatusCode > 0 {
			errMsg.WriteString(fmt.Sprintf("[%s] %d %s (%dms)\n", f.Name, f.StatusCode, f.Body, f.Elapsed.Milliseconds()))
		} else {
			errMsg.WriteString(fmt.Sprintf("[%s] error: %s (%dms)\n", f.Name, f.Body, f.Elapsed.Milliseconds()))
		}
	}

	errStr := errMsg.String()
	s.Logger.Printf("%s", errStr)
	if s.StructuredLogger != nil {
		s.StructuredLogger.Error("", errStr)
	}
	// Log request_received for error (selective logging per T067)
	duration := time.Since(requestStart)
	s.logRequestReceived(r.Method, r.URL.Path, sessionID, clientType, duration, fmt.Errorf("all providers failed"))
	http.Error(w, errStr, http.StatusBadGateway)
}

// tryProviders attempts to forward the request to each provider in order.
// Returns true if a provider successfully handled the request.
func (s *ProxyServer) tryProviders(w http.ResponseWriter, r *http.Request, providers []*Provider, modelOverrides map[string]string, bodyBytes []byte, sessionID, clientType, requestFormat string, failures *[]providerFailure, requestStart time.Time) bool {
	// Generate request ID for monitoring
	requestID := generateRequestID()

	for i, p := range providers {
		isLast := i == len(providers)-1

		// Skip manually disabled providers (checked via config, lazy evaluation)
		if s.isProviderDisabled(p.Name) {
			msg := "skipping (manually disabled)"
			s.Logger.Printf("[%s] %s", p.Name, msg)
			s.logStructured(p.Name, r.Method, r.URL.Path, 0, LogLevelInfo, msg, sessionID, clientType)
			continue
		}

		if !p.IsHealthy() && !isLast {
			msg := fmt.Sprintf("skipping (unhealthy, backoff %v)", p.Backoff)
			s.Logger.Printf("[%s] %s", p.Name, msg)
			s.logStructured(p.Name, r.Method, r.URL.Path, 0, LogLevelInfo, msg, sessionID, clientType)
			continue
		}

		if !p.IsHealthy() && isLast {
			s.Logger.Printf("[%s] last provider, forcing request despite unhealthy (backoff %v)", p.Name, p.Backoff)
		}

		// Get model override for this specific provider
		var modelOverride string
		if modelOverrides != nil {
			modelOverride = modelOverrides[p.Name]
		}

		if p.ProxyURL != "" {
			s.Logger.Printf("[%s] trying %s %s via proxy %s", p.Name, r.Method, r.URL.Path, config.MaskProxyURL(p.ProxyURL))
		} else {
			s.Logger.Printf("[%s] trying %s %s", p.Name, r.Method, r.URL.Path)
		}
		start := time.Now()
		resp, err := s.forwardRequest(r, p, bodyBytes, modelOverride, requestFormat)
		elapsed := time.Since(start)
		if err != nil {
			// Check if this is a transform error - don't mark provider unhealthy
			var transformErr *TransformError
			if errors.As(err, &transformErr) {
				msg := fmt.Sprintf("transform error: %v", transformErr)
				s.Logger.Printf("[%s] %s", p.Name, msg)
				s.logStructured(p.Name, r.Method, r.URL.Path, 0, LogLevelError, msg, sessionID, clientType)

				// Return proper JSON error response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				errResp := map[string]interface{}{
					"error": map[string]interface{}{
						"type":    "transform_error",
						"message": transformErr.Error(),
					},
				}
				json.NewEncoder(w).Encode(errResp)
				return true
			}

			// Check if client canceled the request - don't mark provider unhealthy
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				msg := fmt.Sprintf("request canceled by client: %v", err)
				s.Logger.Printf("[%s] %s", p.Name, msg)
				s.logStructured(p.Name, r.Method, r.URL.Path, 0, LogLevelInfo, msg, sessionID, clientType)
				// Return true to stop processing - client is gone
				return true
			}
			msg := fmt.Sprintf("request error: %v", err)
			s.Logger.Printf("[%s] %s", p.Name, msg)
			s.logStructuredError(p.Name, r.Method, r.URL.Path, err, sessionID, clientType)
			// Log provider_failed event (T068)
			s.logProviderFailed(sessionID, p.Name, err.Error(), elapsed)
			*failures = append(*failures, providerFailure{Name: p.Name, StatusCode: 0, Body: err.Error(), Elapsed: elapsed})
			// Record daemon-level metrics for error
			if s.MetricsRecorder != nil {
				// Classify network/timeout errors
				errType := ErrorTypeNetwork
				if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "timeout") {
					errType = ErrorTypeTimeout
				}
				s.MetricsRecorder.RecordRequest(p.Name, elapsed, &ProxyError{
					Provider: p.Name,
					ErrType:  errType,
					Err:      err,
				})
			}
			p.MarkFailed()
			continue
		}

		// Auth/account errors → failover with long backoff
		if resp.StatusCode == 401 || resp.StatusCode == 402 || resp.StatusCode == 403 {
			errBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			msg := fmt.Sprintf("got %d (auth/account error), failing over", resp.StatusCode)
			s.Logger.Printf("[%s] %s response=%s", p.Name, msg, string(errBody))
			s.logStructuredWithResponse(p.Name, r.Method, r.URL.Path, resp.StatusCode, msg, errBody, sessionID, clientType)
			// Log provider_failed event (T068)
			s.logProviderFailed(sessionID, p.Name, fmt.Sprintf("auth error: %d", resp.StatusCode), elapsed)
			*failures = append(*failures, providerFailure{Name: p.Name, StatusCode: resp.StatusCode, Body: string(errBody), Elapsed: elapsed})
			// Record daemon-level metrics for error
			if s.MetricsRecorder != nil {
				s.MetricsRecorder.RecordRequest(p.Name, elapsed, &ProxyError{
					Provider: p.Name,
					ErrType:  ErrorTypeAuth,
					Err:      fmt.Errorf("status %d", resp.StatusCode),
				})
			}
			p.MarkAuthFailed()
			continue
		}

		// Rate limit → failover with short backoff
		if resp.StatusCode == 429 {
			errBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			msg := fmt.Sprintf("got %d (rate limited), failing over", resp.StatusCode)
			s.Logger.Printf("[%s] %s response=%s", p.Name, msg, string(errBody))
			s.logStructuredWithResponse(p.Name, r.Method, r.URL.Path, resp.StatusCode, msg, errBody, sessionID, clientType)
			// Log provider_failed event (T068)
			s.logProviderFailed(sessionID, p.Name, "rate limited", elapsed)
			*failures = append(*failures, providerFailure{Name: p.Name, StatusCode: resp.StatusCode, Body: string(errBody), Elapsed: elapsed})
			// Record daemon-level metrics for error
			if s.MetricsRecorder != nil {
				s.MetricsRecorder.RecordRequest(p.Name, elapsed, &ProxyError{
					Provider: p.Name,
					ErrType:  ErrorTypeRateLimit,
					Err:      fmt.Errorf("status 429"),
				})
			}
			p.MarkFailed()
			continue
		}

		// Server errors → check if request-related or server-side issue
		if resp.StatusCode >= 500 {
			// Read body to check error type
			errBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			// Check if provider expects Responses API format (not Chat Completions)
			if isResponsesAPIRequired(errBody) && p.GetType() == config.ProviderTypeOpenAI {
				s.Logger.Printf("[%s] got 'input is required', retrying with Responses API format", p.Name)
				retryResp, retryErr := s.retryWithResponsesAPI(r, p, bodyBytes, modelOverride, requestFormat)
				if retryErr != nil {
					s.Logger.Printf("[%s] Responses API retry error: %v", p.Name, retryErr)
					*failures = append(*failures, providerFailure{Name: p.Name, StatusCode: 0, Body: retryErr.Error(), Elapsed: time.Since(start)})
					continue
				}
				if retryResp.StatusCode >= 200 && retryResp.StatusCode < 300 {
					p.MarkHealthy()
					s.Logger.Printf("[%s] Responses API retry success %d", p.Name, retryResp.StatusCode)
					s.logStructured(p.Name, r.Method, r.URL.Path, retryResp.StatusCode, LogLevelInfo, fmt.Sprintf("success %d (Responses API)", retryResp.StatusCode), sessionID, clientType)

					// Update session cache with token usage from response
					s.updateSessionCache(sessionID, retryResp)

					// Record usage and metrics
					s.recordUsageAndMetrics(p.Name, sessionID, clientType, bodyBytes, retryResp, requestID, requestStart, requestFormat, failures)

					// Record daemon-level metrics if recorder is available
					if s.MetricsRecorder != nil {
						s.MetricsRecorder.RecordRequest(p.Name, time.Since(requestStart), nil)
					}

					s.copyResponseFromResponsesAPI(w, retryResp, p, requestFormat)
					return true
				}
				// Retry failed — record the Responses API error, not the Chat Completions error
				retryBody, _ := io.ReadAll(retryResp.Body)
				retryResp.Body.Close()
				s.Logger.Printf("[%s] Responses API retry failed %d: %s", p.Name, retryResp.StatusCode, string(retryBody))
				*failures = append(*failures, providerFailure{Name: p.Name, StatusCode: retryResp.StatusCode, Body: string(retryBody), Elapsed: time.Since(start)})
				continue
			}

			if isRequestRelatedError(errBody) {
				// Request-related error (e.g., context too long) - failover without marking unhealthy
				msg := fmt.Sprintf("got %d (request-related error), failing over without backoff, request_body_size=%d", resp.StatusCode, len(bodyBytes))
				s.Logger.Printf("[%s] %s response=%s", p.Name, msg, string(errBody))
				s.logStructuredWithResponse(p.Name, r.Method, r.URL.Path, resp.StatusCode, msg, errBody, sessionID, clientType)
				// Log provider_failed event (T068)
				s.logProviderFailed(sessionID, p.Name, fmt.Sprintf("request error: %d", resp.StatusCode), elapsed)
				*failures = append(*failures, providerFailure{Name: p.Name, StatusCode: resp.StatusCode, Body: string(errBody), Elapsed: elapsed})
				// Record daemon-level metrics for error (request-related, not provider issue)
				if s.MetricsRecorder != nil {
					s.MetricsRecorder.RecordRequest(p.Name, elapsed, &ProxyError{
						Provider: p.Name,
						ErrType:  ErrorTypeRequest,
						Err:      fmt.Errorf("status %d", resp.StatusCode),
					})
				}
				continue
			}

			// Server-side issue - mark as failed with backoff
			msg := fmt.Sprintf("got %d (server error), failing over", resp.StatusCode)
			s.Logger.Printf("[%s] %s response=%s", p.Name, msg, string(errBody))
			s.logStructuredWithResponse(p.Name, r.Method, r.URL.Path, resp.StatusCode, msg, errBody, sessionID, clientType)
			// Log provider_failed event (T068)
			s.logProviderFailed(sessionID, p.Name, fmt.Sprintf("server error: %d", resp.StatusCode), elapsed)
			*failures = append(*failures, providerFailure{Name: p.Name, StatusCode: resp.StatusCode, Body: string(errBody), Elapsed: elapsed})
			// Record daemon-level metrics for error
			if s.MetricsRecorder != nil {
				s.MetricsRecorder.RecordRequest(p.Name, elapsed, &ProxyError{
					Provider: p.Name,
					ErrType:  ErrorTypeServer,
					Err:      fmt.Errorf("status %d", resp.StatusCode),
				})
			}
			p.MarkFailed()
			continue
		}

		p.MarkHealthy()
		msg := fmt.Sprintf("success %d", resp.StatusCode)
		s.Logger.Printf("[%s] %s", p.Name, msg)
		s.logStructured(p.Name, r.Method, r.URL.Path, resp.StatusCode, LogLevelInfo, msg, sessionID, clientType)

		// Update session cache with token usage from response
		s.updateSessionCache(sessionID, resp)

		// Record usage and metrics
		s.recordUsageAndMetrics(p.Name, sessionID, clientType, bodyBytes, resp, requestID, requestStart, requestFormat, failures)

		// Record daemon-level metrics if recorder is available
		if s.MetricsRecorder != nil {
			s.MetricsRecorder.RecordRequest(p.Name, time.Since(requestStart), nil)
		}

		s.copyResponse(w, resp, p, requestFormat)
		return true
	}

	return false
}

// logStructured logs to the structured logger if available.
func (s *ProxyServer) logStructured(provider, method, path string, statusCode int, level LogLevel, message, sessionID, clientType string) {
	if s.StructuredLogger == nil {
		return
	}
	s.StructuredLogger.Log(LogEntry{
		Level:      level,
		Provider:   provider,
		Method:     method,
		Path:       path,
		StatusCode: statusCode,
		Message:    message,
		SessionID:  sessionID,
		ClientType: clientType,
	})
}

// logStructuredError logs an error to the structured logger.
func (s *ProxyServer) logStructuredError(provider, method, path string, err error, sessionID, clientType string) {
	if s.StructuredLogger == nil {
		return
	}
	s.StructuredLogger.Log(LogEntry{
		Level:      LogLevelError,
		Provider:   provider,
		Method:     method,
		Path:       path,
		Message:    "request failed",
		Error:      err.Error(),
		SessionID:  sessionID,
		ClientType: clientType,
	})
}

// logStructuredWithResponse logs an error with response body to the structured logger.
func (s *ProxyServer) logStructuredWithResponse(provider, method, path string, statusCode int, message string, responseBody []byte, sessionID, clientType string) {
	if s.StructuredLogger == nil {
		return
	}
	bodyStr := string(responseBody)
	if len(bodyStr) > 500 {
		bodyStr = bodyStr[:500] + "..."
	}
	s.StructuredLogger.Log(LogEntry{
		Level:        LogLevelError,
		Provider:     provider,
		Method:       method,
		Path:         path,
		StatusCode:   statusCode,
		Message:      message,
		ResponseBody: bodyStr,
		SessionID:    sessionID,
		ClientType:   clientType,
	})
}

func (s *ProxyServer) forwardRequest(r *http.Request, p *Provider, body []byte, modelOverride string, requestFormat string) (*http.Response, error) {
	var modifiedBody []byte
	if modelOverride != "" {
		// Scenario routing: skip model mapping, use the override model directly
		modifiedBody = s.applyModelOverride(body, modelOverride, p.Name)
	} else {
		// Normal: apply per-provider model mapping
		modifiedBody = s.applyModelMapping(body, p)
	}

	// Apply request transformation if needed
	providerFormat := p.GetType()
	if transform.NeedsTransform(requestFormat, providerFormat) {
		transformer := transform.GetTransformer(providerFormat)
		transformed, err := transformer.TransformRequest(modifiedBody, requestFormat)
		if err != nil {
			s.Logger.Printf("[%s] transform request error: %v", p.Name, err)
			return nil, &TransformError{Op: "request", Err: err}
		}
		s.Logger.Printf("[%s] transformed request: %s → %s", p.Name, requestFormat, providerFormat)
		modifiedBody = transformed
	}

	// Transform path if needed (e.g., /responses → /v1/messages)
	targetPath := r.URL.Path
	if transform.NeedsTransform(requestFormat, providerFormat) {
		targetPath = transform.TransformPath(requestFormat, providerFormat, r.URL.Path)
		if targetPath != r.URL.Path {
			s.Logger.Printf("[%s] path transform: %s → %s", p.Name, r.URL.Path, targetPath)
		}
	}

	// Deduplicate /v1 prefix when base_url already ends with /v1
	// e.g., base_url "https://host/v1" + targetPath "/v1/chat/completions"
	// should produce "https://host/v1/chat/completions", not "https://host/v1/v1/chat/completions"
	basePath := strings.TrimSuffix(p.BaseURL.Path, "/")
	if strings.HasSuffix(basePath, "/v1") && strings.HasPrefix(targetPath, "/v1") {
		originalTarget := targetPath
		targetPath = targetPath[3:] // strip "/v1", keep e.g. "/chat/completions"
		s.Logger.Printf("[%s] path dedup: %s → %s (base_url has /v1)", p.Name, originalTarget, targetPath)
	}

	targetURL := singleJoiningSlash(p.BaseURL.String(), targetPath)
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, bytes.NewReader(modifiedBody))
	if err != nil {
		return nil, err
	}

	// Copy headers
	for k, vv := range r.Header {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	// Override auth
	req.Header.Set("x-api-key", p.Token)
	req.Header.Set("Authorization", "Bearer "+p.Token)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(modifiedBody)))

	// Apply environment variable headers
	s.applyEnvVarsHeaders(req, p.EnvVars)

	// Use per-provider client if available, otherwise fall back to shared client
	client := s.Client
	if p.Client != nil {
		client = p.Client
	}
	return client.Do(req)
}

// retryWithResponsesAPI re-sends a request using the Responses API format
// when the provider returned "input is required" for a Chat Completions request.
func (s *ProxyServer) retryWithResponsesAPI(r *http.Request, p *Provider, originalBody []byte, modelOverride string, requestFormat string) (*http.Response, error) {
	// First apply the same model mapping/override as forwardRequest
	var modifiedBody []byte
	if modelOverride != "" {
		modifiedBody = s.applyModelOverride(originalBody, modelOverride, p.Name)
	} else {
		modifiedBody = s.applyModelMapping(originalBody, p)
	}

	// Apply Anthropic→Chat Completions transform if needed
	providerFormat := p.GetType()
	if transform.NeedsTransform(requestFormat, providerFormat) {
		transformer := transform.GetTransformer(providerFormat)
		transformed, err := transformer.TransformRequest(modifiedBody, requestFormat)
		if err != nil {
			s.Logger.Printf("[%s] Responses API retry: transform request error: %v", p.Name, err)
		} else {
			modifiedBody = transformed
		}
	}

	// Now transform Chat Completions → Responses API
	responsesBody, err := transform.ChatCompletionsToResponsesAPI(modifiedBody)
	if err != nil {
		return nil, fmt.Errorf("transform to Responses API: %w", err)
	}

	// Build the Responses API path
	targetPath := "/v1/responses"

	// Deduplicate /v1 prefix when base_url already ends with /v1
	basePath := strings.TrimSuffix(p.BaseURL.Path, "/")
	if strings.HasSuffix(basePath, "/v1") {
		targetPath = "/responses"
		s.Logger.Printf("[%s] Responses API retry: path dedup /v1/responses → /responses", p.Name)
	}

	targetURL := singleJoiningSlash(p.BaseURL.String(), targetPath)
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	s.Logger.Printf("[%s] Responses API retry: %s %s", p.Name, r.Method, targetURL)

	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, bytes.NewReader(responsesBody))
	if err != nil {
		return nil, err
	}

	// Copy headers
	for k, vv := range r.Header {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	// Override auth
	req.Header.Set("x-api-key", p.Token)
	req.Header.Set("Authorization", "Bearer "+p.Token)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(responsesBody)))

	s.applyEnvVarsHeaders(req, p.EnvVars)

	client := s.Client
	if p.Client != nil {
		client = p.Client
	}
	return client.Do(req)
}

// copyResponseFromResponsesAPI transforms a Responses API response to the client's
// expected format and writes it to the ResponseWriter.
func (s *ProxyServer) copyResponseFromResponsesAPI(w http.ResponseWriter, resp *http.Response, p *Provider, requestFormat string) {
	defer resp.Body.Close()

	// Non-streaming response
	if !strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "failed to read response", http.StatusBadGateway)
			return
		}

		// Transform Responses API → Anthropic
		// Check if client expects Anthropic format (anthropic-messages or legacy anthropic)
		if transform.NormalizeFormat(requestFormat) == config.ProviderTypeAnthropic {
			transformed, err := transform.ResponsesAPIToAnthropic(body)
			if err != nil {
				s.Logger.Printf("[%s] Responses API response transform error: %v", p.Name, err)
			} else {
				s.Logger.Printf("[%s] transformed Responses API → Anthropic", p.Name)
				body = transformed
			}
		}

		// Copy headers (except Content-Length which may have changed)
		for k, vv := range resp.Header {
			if strings.ToLower(k) == "content-length" {
				continue
			}
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
		return
	}

	// Streaming response — transform Responses API SSE → Anthropic SSE
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	flusher, ok := w.(http.Flusher)

	var reader io.Reader = resp.Body
	// Check if client expects Anthropic format (anthropic-messages or legacy anthropic)
	if transform.NormalizeFormat(requestFormat) == config.ProviderTypeAnthropic {
		st := &transform.StreamTransformer{
			ClientFormat:   "anthropic",
			ProviderFormat: "openai-responses",
		}
		reader = st.TransformSSEStream(resp.Body)
		s.Logger.Printf("[%s] transforming Responses API SSE stream → Anthropic", p.Name)
	}

	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				s.Logger.Printf("[%s] streaming response write error: %v", p.Name, writeErr)
				return
			}
			if ok {
				flusher.Flush()
			}
		}
		if err != nil {
			break
		}
	}
}

func (s *ProxyServer) copyResponse(w http.ResponseWriter, resp *http.Response, p *Provider, requestFormat string) {
	defer resp.Body.Close()

	// Check if response transformation is needed
	providerFormat := p.GetType()
	needsTransform := transform.NeedsTransform(requestFormat, providerFormat)

	// Stream SSE responses
	if strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)

		flusher, ok := w.(http.Flusher)

		// Apply stream transformation if needed
		var reader io.Reader = resp.Body
		if needsTransform {
			st := &transform.StreamTransformer{
				ClientFormat:   requestFormat,
				ProviderFormat: providerFormat,
			}
			reader = st.TransformSSEStream(resp.Body)
			s.Logger.Printf("[%s] transforming SSE stream: %s → %s", p.Name, providerFormat, requestFormat)
		}

		buf := make([]byte, 4096)
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				if _, writeErr := w.Write(buf[:n]); writeErr != nil {
					s.Logger.Printf("[%s] streaming response write error: %v", p.Name, writeErr)
					return
				}
				if ok {
					flusher.Flush()
				}
			}
			if err != nil {
				break
			}
		}
		return
	}

	// Non-streaming response - can apply transformation
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read response", http.StatusBadGateway)
		return
	}

	// Apply response transformation if needed
	if needsTransform && len(body) > 0 {
		transformer := transform.GetTransformer(providerFormat)
		transformed, err := transformer.TransformResponse(body, requestFormat)
		if err != nil {
			s.Logger.Printf("[%s] transform response error: %v", p.Name, err)

			// Return proper JSON error response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errResp := map[string]interface{}{
				"error": map[string]interface{}{
					"type":    "transform_error",
					"message": fmt.Sprintf("response transformation failed: %v", err),
				},
			}
			json.NewEncoder(w).Encode(errResp)
			return
		}
		s.Logger.Printf("[%s] transformed response: %s → %s", p.Name, providerFormat, requestFormat)
		body = transformed
	}

	// Copy headers (except Content-Length which may have changed)
	for k, vv := range resp.Header {
		if strings.ToLower(k) == "content-length" {
			continue
		}
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

// applyModelOverride replaces the model in the request body with the given override.
func (s *ProxyServer) applyModelOverride(body []byte, override string, providerName string) []byte {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return body
	}

	originalModel, _ := data["model"].(string)
	if originalModel == override {
		return body
	}

	s.Logger.Printf("[%s] model override: %s → %s", providerName, originalModel, override)
	data["model"] = override
	modified, err := json.Marshal(data)
	if err != nil {
		return body
	}
	return modified
}

// applyModelMapping detects the model type in the request and maps it to
// the provider's corresponding model. This ensures each provider gets the
// correct model name during failover.
//
// Mapping priority:
//  1. Thinking mode enabled → ReasoningModel
//  2. Model name contains "haiku" → HaikuModel
//  3. Model name contains "opus" → OpusModel
//  4. Model name contains "sonnet" → SonnetModel
//  5. Fallback → Model (default model)
func (s *ProxyServer) applyModelMapping(body []byte, p *Provider) []byte {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return body
	}

	originalModel, ok := data["model"].(string)
	if !ok || originalModel == "" {
		return body
	}

	mapped := s.mapModel(originalModel, data, p)
	if mapped == originalModel {
		return body
	}

	s.Logger.Printf("[%s] model mapping: %s → %s", p.Name, originalModel, mapped)
	data["model"] = mapped
	modified, err := json.Marshal(data)
	if err != nil {
		return body
	}
	return modified
}

// mapModel determines which provider model to use based on the request.
func (s *ProxyServer) mapModel(original string, body map[string]interface{}, p *Provider) string {
	// 1. Thinking mode → reasoning model
	if hasThinkingEnabled(body) && p.ReasoningModel != "" {
		return p.ReasoningModel
	}

	// 2. Match by model type (case-insensitive)
	lower := strings.ToLower(original)

	// OpenAI reasoning models (o1, o1-mini, o1-pro, o3, o3-mini, o4-mini)
	if isOpenAIReasoningModel(lower) && p.ReasoningModel != "" {
		return p.ReasoningModel
	}

	// Anthropic haiku or OpenAI mini models
	if (strings.Contains(lower, "haiku") || isOpenAIMiniModel(lower)) && p.HaikuModel != "" {
		return p.HaikuModel
	}

	// Anthropic opus
	if strings.Contains(lower, "opus") && p.OpusModel != "" {
		return p.OpusModel
	}

	// Anthropic sonnet or OpenAI standard models (gpt-4o, gpt-4.1, etc.)
	if (strings.Contains(lower, "sonnet") || isOpenAIStandardModel(lower)) && p.SonnetModel != "" {
		return p.SonnetModel
	}

	// 3. Default model
	if p.Model != "" {
		return p.Model
	}

	// 4. No mapping — keep original
	return original
}

// isOpenAIReasoningModel checks if the model is an OpenAI reasoning model.
func isOpenAIReasoningModel(model string) bool {
	// o1, o1-mini, o1-pro, o3, o3-mini, o4-mini, etc.
	// Match: starts with "o" followed by a digit
	if len(model) >= 2 && model[0] == 'o' && model[1] >= '0' && model[1] <= '9' {
		return true
	}
	return false
}

// isOpenAIMiniModel checks if the model is an OpenAI mini model (small/fast).
func isOpenAIMiniModel(model string) bool {
	// Known OpenAI mini models: gpt-4o-mini, gpt-4.1-mini, gpt-4.1-nano
	knownMiniModels := []string{"gpt-4o-mini", "gpt-4.1-mini", "gpt-4.1-nano", "gpt-4-mini"}
	for _, m := range knownMiniModels {
		if strings.HasPrefix(model, m) {
			return true
		}
	}
	return false
}

// isOpenAIStandardModel checks if the model is a known OpenAI standard model.
func isOpenAIStandardModel(model string) bool {
	// Known OpenAI standard models (not mini/nano, not reasoning)
	knownStandardModels := []string{"gpt-4o", "gpt-4.1", "gpt-4-turbo", "gpt-4.5"}
	for _, m := range knownStandardModels {
		if strings.HasPrefix(model, m) && !strings.Contains(model, "-mini") && !strings.Contains(model, "-nano") {
			return true
		}
	}
	return false
}

// updateSessionCache extracts token usage from the response and updates the session cache.
// Only works for non-streaming (non-SSE) responses.
func (s *ProxyServer) updateSessionCache(sessionID string, resp *http.Response) {
	if sessionID == "" {
		return
	}

	// Skip SSE streaming responses — body is consumed by copyResponse
	if strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		return
	}

	// Read response body to extract usage information
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	// Restore body for copyResponse
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	var respData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &respData); err != nil {
		return
	}

	// Extract usage from response
	usage, ok := respData["usage"].(map[string]interface{})
	if !ok {
		return
	}

	inputTokens, _ := usage["input_tokens"].(float64)
	outputTokens, _ := usage["output_tokens"].(float64)

	if inputTokens > 0 || outputTokens > 0 {
		UpdateSessionUsage(sessionID, &SessionUsage{
			InputTokens:  int(inputTokens),
			OutputTokens: int(outputTokens),
		})
		s.Logger.Printf("[session] updated cache for %s: input=%d, output=%d",
			sessionID, int(inputTokens), int(outputTokens))
	}
}

// recordUsageAndMetrics records usage data and provider metrics after a successful request.
func (s *ProxyServer) recordUsageAndMetrics(providerName, sessionID, clientType string, requestBody []byte, resp *http.Response, requestID string, requestStart time.Time, requestFormat string, failures *[]providerFailure) {
	// Extract model from request
	var reqData map[string]interface{}
	model := ""
	if err := json.Unmarshal(requestBody, &reqData); err == nil {
		model, _ = reqData["model"].(string)
	}

	// Calculate total duration
	duration := time.Since(requestStart)

	// We need to peek at the response body for usage info
	// Note: For non-streaming responses, the body was already read by updateSessionCache
	// and restored. For streaming, we skip usage tracking.
	if strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		// Streaming response - record metric without usage details
		if db := GetGlobalLogDB(); db != nil {
			db.RecordMetric(providerName, 0, resp.StatusCode, false, false)
		}

		// Record request to monitor (streaming, no token info yet)
		monitor := GetGlobalRequestMonitor()
		monitor.Add(RequestRecord{
			ID:            requestID,
			Timestamp:     requestStart,
			SessionID:     sessionID,
			ClientType:    clientType,
			Provider:      providerName,
			Model:         model,
			RequestFormat: requestFormat,
			StatusCode:    resp.StatusCode,
			DurationMs:    duration.Milliseconds(),
			RequestSize:   len(requestBody),
			FailoverChain: buildFailoverChain(failures),
		})
		return
	}

	// Get usage from session cache (was just updated by updateSessionCache)
	usage := GetSessionUsage(sessionID)
	if usage == nil {
		// Record request without token info
		monitor := GetGlobalRequestMonitor()
		monitor.Add(RequestRecord{
			ID:            requestID,
			Timestamp:     requestStart,
			SessionID:     sessionID,
			ClientType:    clientType,
			Provider:      providerName,
			Model:         model,
			RequestFormat: requestFormat,
			StatusCode:    resp.StatusCode,
			DurationMs:    duration.Milliseconds(),
			RequestSize:   len(requestBody),
			FailoverChain: buildFailoverChain(failures),
		})
		return
	}

	// Calculate cost
	tracker := GetGlobalUsageTracker()
	if tracker == nil {
		return
	}

	cost := tracker.CalculateCost(model, usage.InputTokens, usage.OutputTokens)

	// Record usage entry
	entry := UsageEntry{
		Timestamp:    time.Now(),
		SessionID:    sessionID,
		Provider:     providerName,
		Model:        model,
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		CostUSD:      cost,
		ClientType:   clientType,
	}
	tracker.Record(entry)

	// Record provider metric
	if db := GetGlobalLogDB(); db != nil {
		db.RecordMetric(providerName, 0, resp.StatusCode, false, false)
	}

	// Update session with turn info
	AddTurnToSession(sessionID, TurnUsage{
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		Cost:         cost,
		Model:        model,
		Provider:     providerName,
		Timestamp:    time.Now(),
	})

	// Update bot bridge with session status
	if bridge := GetBotBridge(); bridge != nil {
		bridge.MarkSessionIdle(sessionID)
	}

	// Record request to monitor with full details
	monitor := GetGlobalRequestMonitor()
	monitor.Add(RequestRecord{
		ID:            requestID,
		Timestamp:     requestStart,
		SessionID:     sessionID,
		ClientType:    clientType,
		Provider:      providerName,
		Model:         model,
		RequestFormat: requestFormat,
		StatusCode:    resp.StatusCode,
		DurationMs:    duration.Milliseconds(),
		InputTokens:   usage.InputTokens,
		OutputTokens:  usage.OutputTokens,
		Cost:          cost,
		RequestSize:   len(requestBody),
		FailoverChain: buildFailoverChain(failures),
	})
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

// generateRequestID generates a unique request ID for monitoring.
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// buildFailoverChain converts providerFailure slice to ProviderAttempt slice for monitoring.
func buildFailoverChain(failures *[]providerFailure) []ProviderAttempt {
	if failures == nil || len(*failures) == 0 {
		return nil
	}

	chain := make([]ProviderAttempt, len(*failures))
	for i, f := range *failures {
		chain[i] = ProviderAttempt{
			Provider:     f.Name,
			StatusCode:   f.StatusCode,
			ErrorMessage: f.Body,
			DurationMs:   f.Elapsed.Milliseconds(),
		}
	}
	return chain
}

// isRequestRelatedError checks if a 5xx error is caused by the request itself
// (e.g., context too long) rather than a server-side issue.
// These errors should trigger failover but not mark the provider as unhealthy.
func isRequestRelatedError(body []byte) bool {
	var errResp struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err != nil {
		return false
	}

	// Check for known request-related error types
	errType := strings.ToLower(errResp.Error.Type)
	errMsg := strings.ToLower(errResp.Error.Message)

	// invalid_request_error with context/token related messages
	if errType == "invalid_request_error" {
		return true
	}

	// Check message for context/token length issues
	contextKeywords := []string{
		"too long", "too large", "exceeds maximum", "context length",
		"exceeds the maximum", "maximum context",
	}
	for _, kw := range contextKeywords {
		if strings.Contains(errMsg, kw) {
			return true
		}
	}

	return false
}

// isResponsesAPIRequired checks if an error response indicates the provider
// expects Responses API format instead of Chat Completions.
// Detects "input is required" in the response body, which occurs when a
// provider only supports the Responses API (/v1/responses with input field).
func isResponsesAPIRequired(body []byte) bool {
	return strings.Contains(string(body), "input is required")
}

// applyEnvVarsHeaders converts environment variables to HTTP headers.
// Environment variable names are converted to lowercase and prefixed with "x-env-".
// For example: CLAUDE_CODE_MAX_OUTPUT_TOKENS -> x-env-claude-code-max-output-tokens
func (s *ProxyServer) applyEnvVarsHeaders(req *http.Request, envVars map[string]string) {
	if envVars == nil {
		return
	}

	for k, v := range envVars {
		if k == "" || v == "" {
			continue
		}
		// Convert env var name to HTTP header format
		// CLAUDE_CODE_MAX_OUTPUT_TOKENS -> x-env-claude-code-max-output-tokens
		headerName := "x-env-" + strings.ToLower(strings.ReplaceAll(k, "_", "-"))
		req.Header.Set(headerName, v)
	}
}

// StartProxy starts the proxy server and returns the port.
// Note: clientFormat parameter is kept for backward compatibility but is now ignored.
// Request format is detected per-request from the X-Zen-Request-Format header.
func StartProxy(providers []*Provider, clientFormat string, listenAddr string, logger *log.Logger) (int, error) {
	srv := NewProxyServer(providers, logger, config.LoadBalanceFailover, GetGlobalLoadBalancer())

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return 0, fmt.Errorf("listen: %w", err)
	}

	port := ln.Addr().(*net.TCPAddr).Port

	go http.Serve(ln, srv)

	return port, nil
}

// StartProxyWithRouting starts the proxy server with scenario-based routing.
// Note: clientFormat parameter is kept for backward compatibility but is now ignored.
// Request format is detected per-request from the X-Zen-Request-Format header.
func StartProxyWithRouting(routing *RoutingConfig, clientFormat string, listenAddr string, logger *log.Logger) (int, error) {
	srv := NewProxyServerWithRouting(routing, logger, config.LoadBalanceFailover, GetGlobalLoadBalancer())

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return 0, fmt.Errorf("listen: %w", err)
	}

	port := ln.Addr().(*net.TCPAddr).Port

	go http.Serve(ln, srv)

	return port, nil
}

// logRequestReceived logs request_received event (T067: only if error or duration >1s)
func (s *ProxyServer) logRequestReceived(method, path, sessionID, clientType string, duration time.Duration, err error) {
	// Get daemon structured logger if available
	daemonLogger := GetDaemonLogger()
	if daemonLogger == nil {
		return
	}

	fields := map[string]interface{}{
		"method":       method,
		"path":         path,
		"session":      sessionID,
		"client_type":  clientType,
		"duration_ms":  duration.Milliseconds(),
	}

	if err != nil {
		fields["error"] = err.Error()
		daemonLogger.Error("request_received", fields)
	} else {
		daemonLogger.Info("request_received", fields)
	}
}

// logProviderFailed logs provider_failed event (T068)
func (s *ProxyServer) logProviderFailed(sessionID, provider, errorMsg string, duration time.Duration) {
	// Get daemon structured logger if available
	daemonLogger := GetDaemonLogger()
	if daemonLogger == nil {
		return
	}

	daemonLogger.Error("provider_failed", map[string]interface{}{
		"session":     sessionID,
		"provider":    provider,
		"error":       errorMsg,
		"duration_ms": duration.Milliseconds(),
	})
}

// daemonStructuredLogger holds the daemon's structured logger
var (
	daemonStructuredLogger     *daemonLogger
	daemonStructuredLoggerOnce sync.Once
	daemonStructuredLoggerMu   sync.RWMutex
)

// daemonLogger interface matches daemon.StructuredLogger methods
type daemonLogger interface {
	Error(event string, fields map[string]interface{})
	Info(event string, fields map[string]interface{})
}

// SetDaemonLogger sets the daemon's structured logger for proxy logging
func SetDaemonLogger(logger daemonLogger) {
	daemonStructuredLoggerMu.Lock()
	defer daemonStructuredLoggerMu.Unlock()
	daemonStructuredLogger = &logger
}

// GetDaemonLogger returns the daemon's structured logger if available
func GetDaemonLogger() daemonLogger {
	daemonStructuredLoggerMu.RLock()
	defer daemonStructuredLoggerMu.RUnlock()
	if daemonStructuredLogger != nil {
		return *daemonStructuredLogger
	}
	return nil
}
