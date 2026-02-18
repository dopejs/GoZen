package daemon

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

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

func (d *Daemon) handleDaemonSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

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
