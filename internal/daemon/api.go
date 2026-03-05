package daemon

import (
	"encoding/json"
	"net/http"
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
	Status         string `json:"status"`
	Version        string `json:"version"`
	Uptime         string `json:"uptime"`
	UptimeSeconds  int64  `json:"uptime_seconds"`
	ProxyPort      int    `json:"proxy_port"`
	WebPort        int    `json:"web_port"`
	ActiveSessions int    `json:"active_sessions"`
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
	})
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
	var req registerSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.SessionID == "" || req.Profile == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "session_id and profile are required"})
		return
	}

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
