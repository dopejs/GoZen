package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/proxy"
	gosync "github.com/dopejs/gozen/internal/sync"
)

// Server is the web configuration management server.
type Server struct {
	httpServer *http.Server
	mux        *http.ServeMux
	logger     *log.Logger
	version    string
	port       int
	auth       *AuthManager
	keys       *KeyPair
	syncMu     sync.RWMutex
	syncMgr    *gosync.SyncManager
}

// NewServer creates a new web server bound to 127.0.0.1 on the configured port.
// If portOverride > 0, it is used instead of the configured port.
func NewServer(version string, logger *log.Logger, portOverride int) *Server {
	port := config.GetWebPort()
	if portOverride > 0 {
		port = portOverride
	}
	s := &Server{
		logger:  logger,
		version: version,
		port:    port,
		auth:    NewAuthManager(),
	}

	// Generate RSA key pair for encrypted token transport
	keys, err := GenerateKeyPair()
	if err != nil {
		if logger != nil {
			logger.Printf("Warning: failed to generate RSA key pair: %v", err)
		}
	}
	s.keys = keys

	s.mux = http.NewServeMux()

	// Auth routes (accessible without authentication)
	s.mux.HandleFunc("/api/v1/auth/login", s.handleLogin)
	s.mux.HandleFunc("/api/v1/auth/logout", s.handleLogout)
	s.mux.HandleFunc("/api/v1/auth/check", s.handleAuthCheck)
	s.mux.HandleFunc("/api/v1/auth/pubkey", s.handlePubKey)

	// API routes
	s.mux.HandleFunc("/api/v1/health", s.handleHealth)
	s.mux.HandleFunc("/api/v1/reload", s.handleReload)
	s.mux.HandleFunc("/api/v1/providers", s.handleProviders)
	s.mux.HandleFunc("/api/v1/providers/", s.handleProvider)
	s.mux.HandleFunc("/api/v1/profiles", s.handleProfiles)
	s.mux.HandleFunc("/api/v1/profiles/", s.handleProfile)
	s.mux.HandleFunc("/api/v1/logs", s.handleLogs)
	s.mux.HandleFunc("/api/v1/settings", s.handleSettings)
	s.mux.HandleFunc("/api/v1/settings/password", s.handlePasswordChange)
	s.mux.HandleFunc("/api/v1/bindings", s.handleBindings)
	s.mux.HandleFunc("/api/v1/bindings/", s.handleBinding)

	// Sync routes
	s.mux.HandleFunc("/api/v1/sync/config", s.handleSyncConfig)
	s.mux.HandleFunc("/api/v1/sync/pull", s.handleSyncPull)
	s.mux.HandleFunc("/api/v1/sync/push", s.handleSyncPush)
	s.mux.HandleFunc("/api/v1/sync/status", s.handleSyncStatus)
	s.mux.HandleFunc("/api/v1/sync/test", s.handleSyncTest)
	s.mux.HandleFunc("/api/v1/sync/create-gist", s.handleSyncCreateGist)

	// Usage & Budget routes
	s.mux.HandleFunc("/api/v1/usage", s.handleUsage)
	s.mux.HandleFunc("/api/v1/usage/summary", s.handleUsageSummary)
	s.mux.HandleFunc("/api/v1/usage/hourly", s.handleUsageHourly)
	s.mux.HandleFunc("/api/v1/budget", s.handleBudget)
	s.mux.HandleFunc("/api/v1/budget/status", s.handleBudgetStatus)

	// Health monitoring routes
	s.mux.HandleFunc("/api/v1/health/providers", s.handleHealthProviders)
	s.mux.HandleFunc("/api/v1/health/providers/", s.handleHealthProvider)

	// Session routes
	s.mux.HandleFunc("/api/v1/sessions", s.handleSessions)
	s.mux.HandleFunc("/api/v1/sessions/", s.handleSession)

	// Webhook routes
	s.mux.HandleFunc("/api/v1/webhooks", s.handleWebhooks)
	s.mux.HandleFunc("/api/v1/webhooks/test", s.handleWebhookTest)
	s.mux.HandleFunc("/api/v1/webhooks/", s.handleWebhook)

	// Pricing routes
	s.mux.HandleFunc("/api/v1/pricing", s.handlePricing)
	s.mux.HandleFunc("/api/v1/pricing/reset", s.handlePricingReset)

	// Compression routes (BETA)
	s.mux.HandleFunc("/api/v1/compression", s.handleCompression)
	s.mux.HandleFunc("/api/v1/compression/stats", s.handleGetCompressionStats)

	// Middleware routes (BETA)
	s.mux.HandleFunc("/api/v1/middleware", s.handleMiddleware)
	s.mux.HandleFunc("/api/v1/middleware/", s.handleMiddleware)
	s.mux.HandleFunc("/api/v1/middleware/reload", s.handleMiddlewareReload)

	// Agent routes (BETA)
	s.mux.HandleFunc("/api/v1/agent/config", s.handleAgentConfig)
	s.mux.HandleFunc("/api/v1/agent/stats", s.handleAgentStats)
	s.mux.HandleFunc("/api/v1/agent/sessions", s.handleAgentSessions)
	s.mux.HandleFunc("/api/v1/agent/sessions/", s.handleAgentSessions)
	s.mux.HandleFunc("/api/v1/agent/locks", s.handleAgentLocks)
	s.mux.HandleFunc("/api/v1/agent/locks/", s.handleAgentLocks)
	s.mux.HandleFunc("/api/v1/agent/changes", s.handleAgentChanges)
	s.mux.HandleFunc("/api/v1/agent/tasks", s.handleAgentTasks)
	s.mux.HandleFunc("/api/v1/agent/tasks/", s.handleAgentTasks)
	s.mux.HandleFunc("/api/v1/agent/runtime", s.handleAgentRuntime)
	s.mux.HandleFunc("/api/v1/agent/runtime/", s.handleAgentRuntime)
	s.mux.HandleFunc("/api/v1/agent/guardrails", s.handleAgentGuardrails)
	s.mux.HandleFunc("/api/v1/agent/guardrails/", s.handleAgentGuardrails)

	// Static files
	staticSub, _ := fs.Sub(staticFS, "dist")
	fileServer := http.FileServer(http.FS(staticSub))
	s.mux.Handle("/", fileServer)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: s.securityHeaders(s.authMiddleware(s.mux)),
	}

	return s
}

// HandleFunc registers an additional handler on the server's mux.
// Must be called before Start().
func (s *Server) HandleFunc(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, handler)
}

// SetSyncManager sets the sync manager for the web server.
func (s *Server) SetSyncManager(mgr *gosync.SyncManager) {
	s.syncMu.Lock()
	s.syncMgr = mgr
	s.syncMu.Unlock()
}

// Start begins listening. Returns an error if the port is already in use.
// Returns nil on graceful shutdown (http.ErrServerClosed).
func (s *Server) Start() error {
	// Start periodic session cleanup
	go s.auth.sessionCleanupLoop()

	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("port %d is already in use: %w", s.port, err)
	}
	s.logger.Printf("Web server listening on %s", s.httpServer.Addr)
	err = s.httpServer.Serve(ln)
	if err == http.ErrServerClosed {
		return nil // graceful shutdown
	}
	return err
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	// Stop the session cleanup loop
	s.auth.StopCleanup()
	return s.httpServer.Shutdown(ctx)
}

// securityHeaders adds security response headers.
func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

// --- health & reload ---

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": s.version,
	})
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	store := config.DefaultStore()
	if err := store.Reload(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "reloaded"})
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func readJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

// maskToken masks an auth token for display: "sk-abc...xyz" style.
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:5] + "..." + token[len(token)-4:]
}

// --- logs ---

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	filter := proxy.LogFilter{
		Provider:   query.Get("provider"),
		SessionID:  query.Get("session_id"),
		ClientType: query.Get("client_type"),
	}

	if query.Get("errors_only") == "true" {
		filter.ErrorsOnly = true
	}

	if level := query.Get("level"); level != "" {
		filter.Level = proxy.LogLevel(level)
	}

	if statusCode := query.Get("status_code"); statusCode != "" {
		if code, err := strconv.Atoi(statusCode); err == nil {
			filter.StatusCode = code
		}
	}

	if statusMin := query.Get("status_min"); statusMin != "" {
		if code, err := strconv.Atoi(statusMin); err == nil {
			filter.StatusMin = code
		}
	}

	if statusMax := query.Get("status_max"); statusMax != "" {
		if code, err := strconv.Atoi(statusMax); err == nil {
			filter.StatusMax = code
		}
	}

	if limit := query.Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	if filter.Limit <= 0 {
		filter.Limit = 100 // default limit
	}

	// Try in-memory logger first (same process as proxy), then SQLite (cross-process).
	var entries []proxy.LogEntry
	var providers []string

	logger := proxy.GetGlobalLogger()
	if logger != nil && logger.HasEntries() {
		entries = logger.GetEntries(filter)
		providers = logger.GetProviders()
	} else if db := proxy.GetGlobalLogDB(); db != nil {
		var err error
		entries, err = db.Query(filter)
		if err != nil {
			s.logger.Printf("Failed to query log database: %v", err)
			entries = []proxy.LogEntry{}
		}
		providers, err = db.GetProviders()
		if err != nil {
			s.logger.Printf("Failed to query log providers: %v", err)
			providers = []string{}
		}
	}

	writeJSON(w, http.StatusOK, proxy.LogsResponse{
		Entries:   entries,
		Total:     len(entries),
		Providers: providers,
	})
}
