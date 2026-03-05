package web

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/dopejs/gozen/internal/proxy"
)

// handleSessions handles GET /api/v1/sessions - returns all active sessions.
func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	insights := proxy.GetAllSessionInsights()
	if insights == nil {
		insights = []*proxy.SessionInsight{}
	}

	// Get cache stats
	size, maxSize := proxy.GetCacheStats()

	response := struct {
		Sessions  []*proxy.SessionInsight `json:"sessions"`
		CacheSize int                     `json:"cache_size"`
		MaxSize   int                     `json:"max_size"`
	}{
		Sessions:  insights,
		CacheSize: size,
		MaxSize:   maxSize,
	}

	writeJSON(w, http.StatusOK, response)
}

// handleSession handles GET /api/v1/sessions/{id} - returns a specific session.
func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract session ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/sessions/")
	sessionID := strings.TrimSuffix(path, "/")

	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "session ID required")
		return
	}

	insight := proxy.GetSessionInsight(sessionID)
	if insight == nil {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}

	// Check for context warning
	threshold := 100000 // Default 100k tokens
	if t := r.URL.Query().Get("threshold"); t != "" {
		if n, err := strconv.Atoi(t); err == nil && n > 0 {
			threshold = n
		}
	}

	warning := proxy.GetContextWarning(sessionID, threshold)

	response := struct {
		Insight *proxy.SessionInsight   `json:"insight"`
		Warning *proxy.ContextWarning   `json:"warning,omitempty"`
	}{
		Insight: insight,
		Warning: warning,
	}

	writeJSON(w, http.StatusOK, response)
}
