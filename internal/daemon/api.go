package daemon

import (
	"encoding/json"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/proxy"
)

// getBotBridge returns the global bot bridge.
func getBotBridge() *proxy.BotBridge {
	return proxy.GetBotBridge()
}

// --- Daemon Status API ---

type daemonStatusResponse struct {
	Status         string               `json:"status"`
	Version        string               `json:"version"`
	Uptime         string               `json:"uptime"`
	UptimeSeconds  int64                `json:"uptime_seconds"`
	ProxyPort      int                  `json:"proxy_port"`
	WebPort        int                  `json:"web_port"`
	ActiveSessions int                  `json:"active_sessions"`
	FeatureGates   *config.FeatureGates `json:"feature_gates,omitempty"`
}

type daemonMemoryStats struct {
	AllocBytes     uint64 `json:"alloc_bytes"`
	SysBytes       uint64 `json:"sys_bytes"`
	HeapAllocBytes uint64 `json:"heap_alloc_bytes"`
	HeapObjects    uint64 `json:"heap_objects"`
	NumGC          uint32 `json:"num_gc"`
}

type daemonProviderHealth struct {
	Name        string             `json:"name"`
	Status      proxy.HealthStatus `json:"status"`
	LastCheck   *time.Time         `json:"last_check,omitempty"`
	LatencyMs   int                `json:"latency_ms,omitempty"`
	SuccessRate float64            `json:"success_rate"`
	CheckCount  int                `json:"check_count"`
	FailCount   int                `json:"fail_count"`
}

type daemonHealthResponse struct {
	Status             string                 `json:"status"`
	Version            string                 `json:"version"`
	UptimeSeconds      int64                  `json:"uptime_seconds"`
	Goroutines         int                    `json:"goroutines"`
	Memory             daemonMemoryStats      `json:"memory"`
	ActiveSessions     int                    `json:"active_sessions"`
	HealthCheckEnabled bool                   `json:"health_check_enabled"`
	HealthCheckRunning bool                   `json:"health_check_running"`
	Providers          []daemonProviderHealth `json:"providers"`
}

func (d *Daemon) handleDaemonStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	uptime := time.Since(d.startTime)
	writeJSON(w, http.StatusOK, daemonStatusResponse{
		Status:         "running",
		Version:        d.version,
		Uptime:         uptime.Truncate(time.Second).String(),
		UptimeSeconds:  int64(uptime.Seconds()),
		ProxyPort:      d.proxyPort,
		WebPort:        d.webPort,
		ActiveSessions: d.ActiveSessionCount(),
		FeatureGates:   config.GetFeatureGates(),
	})
}

func (d *Daemon) handleDaemonHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	cfg := config.GetHealthCheck()
	checker := proxy.GetGlobalHealthChecker()
	running := checker != nil && checker.IsRunning()

	providers := make([]daemonProviderHealth, 0)
	unhealthyCount := 0
	degradedCount := 0
	if checker != nil {
		for _, status := range checker.GetAllStatus() {
			providers = append(providers, daemonProviderHealth{
				Name:        status.Provider,
				Status:      status.Status,
				LastCheck:   status.LastCheck,
				LatencyMs:   status.LatencyMs,
				SuccessRate: status.SuccessRate,
				CheckCount:  status.CheckCount,
				FailCount:   status.FailCount,
			})
			switch status.Status {
			case proxy.HealthStatusUnhealthy:
				unhealthyCount++
			case proxy.HealthStatusDegraded:
				degradedCount++
			}
		}
	}

	overallStatus := "healthy"
	if runtime.NumGoroutine() > 1000 || mem.Alloc > 500*1024*1024 {
		overallStatus = "degraded"
	}
	if degradedCount > 0 || unhealthyCount > 0 {
		overallStatus = "degraded"
	}
	if len(providers) > 0 && unhealthyCount == len(providers) {
		overallStatus = "unhealthy"
	}

	writeJSON(w, http.StatusOK, daemonHealthResponse{
		Status:        overallStatus,
		Version:       d.version,
		UptimeSeconds: int64(time.Since(d.startTime).Seconds()),
		Goroutines:    runtime.NumGoroutine(),
		Memory: daemonMemoryStats{
			AllocBytes:     mem.Alloc,
			SysBytes:       mem.Sys,
			HeapAllocBytes: mem.HeapAlloc,
			HeapObjects:    mem.HeapObjects,
			NumGC:          mem.NumGC,
		},
		ActiveSessions:     d.ActiveSessionCount(),
		HealthCheckEnabled: cfg != nil && cfg.Enabled,
		HealthCheckRunning: running,
		Providers:          providers,
	})
}

// --- Daemon Metrics API ---

func (d *Daemon) handleDaemonMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	if d.metrics == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "metrics not initialized"})
		return
	}

	// Update resource peaks before returning stats
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	memoryMB := int64(mem.Alloc / 1024 / 1024)
	d.metrics.UpdateResourcePeaks(runtime.NumGoroutine(), memoryMB)

	stats := d.metrics.GetStats()
	writeJSON(w, http.StatusOK, stats)
}

// --- Daemon Shutdown API ---

func (d *Daemon) handleDaemonShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "shutting down"})

	// Trigger shutdown in background so the response is sent first
	go func() {
		time.Sleep(100 * time.Millisecond)
		select {
		case <-d.shutdownCh:
			// Already closed
		default:
			close(d.shutdownCh)
		}
	}()
}

// ShutdownCh returns a channel that is closed when shutdown is requested via API.
func (d *Daemon) ShutdownCh() <-chan struct{} {
	return d.shutdownCh
}

// --- Daemon Reload API ---

func (d *Daemon) handleDaemonReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	d.onConfigReload()
	writeJSON(w, http.StatusOK, map[string]string{"status": "reloaded"})
}

// --- Daemon Sessions API ---

type registerSessionRequest struct {
	SessionID  string `json:"session_id"`
	Profile    string `json:"profile"`
	ClientType string `json:"client_type"`
}

func (d *Daemon) handleDaemonSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		d.handleGetSessions(w, r)
	case http.MethodPost:
		d.handleRegisterSession(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (d *Daemon) handleGetSessions(w http.ResponseWriter, r *http.Request) {
	d.mu.RLock()
	sessions := make([]*SessionInfo, 0, len(d.sessions))
	for _, s := range d.sessions {
		sessions = append(sessions, s)
	}
	d.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

func (d *Daemon) handleRegisterSession(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req registerSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.SessionID == "" || req.Profile == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "session_id and profile are required"})
		return
	}

	d.RegisterSession(req.SessionID, req.Profile, req.ClientType)

	// Register with bot bridge
	if bridge := getBotBridge(); bridge != nil {
		cacheKey := req.Profile + ":" + req.SessionID
		bridge.UpdateSession(cacheKey, req.ClientType, nil, "", "", "input")
		d.logger.Printf("[session] registered %s (client=%s)", cacheKey, req.ClientType)
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "registered"})
}

// --- Temporary Profile API ---

type tempProfileRequest struct {
	Providers []string `json:"providers"`
}

type tempProfileResponse struct {
	ID string `json:"id"`
}

func (d *Daemon) handleTempProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req tempProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.Body.Close()
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	_ = r.Body.Close()

	if len(req.Providers) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "providers required"})
		return
	}

	// Validate provider names exist
	store := config.DefaultStore()
	for _, name := range req.Providers {
		if store.GetProvider(name) == nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "provider not found: " + name,
			})
			return
		}
	}

	id := d.RegisterTempProfile(req.Providers)
	writeJSON(w, http.StatusCreated, tempProfileResponse{ID: id})
}

func (d *Daemon) handleTempProfile(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL: /api/v1/profiles/temp/{id}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/profiles/temp/")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}

	switch r.Method {
	case http.MethodDelete:
		d.RemoveTempProfile(id)
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
