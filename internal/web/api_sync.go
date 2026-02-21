package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dopejs/gozen/internal/config"
	gosync "github.com/dopejs/gozen/internal/sync"
)

// handleSyncConfig handles GET/PUT /api/v1/sync/config
func (s *Server) handleSyncConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := config.GetSyncConfig()
		if cfg == nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{"configured": false})
			return
		}
		// Mask sensitive fields
		resp := *cfg
		if resp.Token != "" {
			resp.Token = maskToken(resp.Token)
		}
		if resp.AccessKey != "" {
			resp.AccessKey = maskToken(resp.AccessKey)
		}
		if resp.SecretKey != "" {
			resp.SecretKey = maskToken(resp.SecretKey)
		}
		if resp.Passphrase != "" {
			resp.Passphrase = "********"
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"configured": true,
			"config":     resp,
		})

	case http.MethodPut:
		var cfg config.SyncConfig
		if err := readJSON(r, &cfg); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		// Preserve existing secrets if masked values are sent back
		existing := config.GetSyncConfig()
		if existing != nil {
			if cfg.Token == maskToken(existing.Token) || cfg.Token == "" {
				cfg.Token = existing.Token
			}
			if cfg.AccessKey == maskToken(existing.AccessKey) || cfg.AccessKey == "" {
				cfg.AccessKey = existing.AccessKey
			}
			if cfg.SecretKey == maskToken(existing.SecretKey) || cfg.SecretKey == "" {
				cfg.SecretKey = existing.SecretKey
			}
			if cfg.Passphrase == "********" || cfg.Passphrase == "" {
				cfg.Passphrase = existing.Passphrase
			}
		}
		if err := config.SetSyncConfig(&cfg); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		// Reinitialize sync manager if server has one
		s.syncMu.Lock()
		if s.syncMgr != nil {
			if mgr, err := gosync.NewSyncManager(&cfg); err == nil {
				s.syncMgr = mgr
			}
		}
		s.syncMu.Unlock()
		writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleSyncPull handles POST /api/v1/sync/pull
func (s *Server) handleSyncPull(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	mgr, err := s.getOrCreateSyncManager()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	if err := mgr.Pull(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, "pull failed: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "pulled"})
}

// handleSyncPush handles POST /api/v1/sync/push
func (s *Server) handleSyncPush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	mgr, err := s.getOrCreateSyncManager()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	if err := mgr.Push(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, "push failed: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "pushed"})
}

// handleSyncStatus handles GET /api/v1/sync/status
func (s *Server) handleSyncStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	mgr, err := s.getOrCreateSyncManager()
	if err != nil {
		writeJSON(w, http.StatusOK, &gosync.SyncStatus{Configured: false})
		return
	}
	writeJSON(w, http.StatusOK, mgr.Status())
}

// handleSyncTest handles POST /api/v1/sync/test
func (s *Server) handleSyncTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	// Allow testing with provided config (not yet saved)
	var cfg config.SyncConfig
	if err := readJSON(r, &cfg); err == nil && cfg.Backend != "" {
		// Preserve secrets from existing config if masked
		existing := config.GetSyncConfig()
		if existing != nil {
			if cfg.Token == maskToken(existing.Token) {
				cfg.Token = existing.Token
			}
			if cfg.AccessKey == maskToken(existing.AccessKey) {
				cfg.AccessKey = existing.AccessKey
			}
			if cfg.SecretKey == maskToken(existing.SecretKey) {
				cfg.SecretKey = existing.SecretKey
			}
			if cfg.Passphrase == "********" {
				cfg.Passphrase = existing.Passphrase
			}
		}
		mgr, err := gosync.NewSyncManager(&cfg)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		if err := mgr.TestConnection(ctx); err != nil {
			writeError(w, http.StatusBadGateway, "connection failed: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}

	mgr, err := s.getOrCreateSyncManager()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	if err := mgr.TestConnection(ctx); err != nil {
		writeError(w, http.StatusBadGateway, "connection failed: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleSyncCreateGist handles POST /api/v1/sync/create-gist
func (s *Server) handleSyncCreateGist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Token string `json:"token"`
	}
	if err := readJSON(r, &req); err != nil || req.Token == "" {
		writeError(w, http.StatusBadRequest, "token is required")
		return
	}
	// Preserve existing token if masked
	existing := config.GetSyncConfig()
	if existing != nil && req.Token == maskToken(existing.Token) {
		req.Token = existing.Token
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	gistID, err := gosync.CreateGist(ctx, req.Token)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"gist_id": gistID})
}

// getOrCreateSyncManager returns the server's sync manager, creating one lazily if needed.
func (s *Server) getOrCreateSyncManager() (*gosync.SyncManager, error) {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()

	if s.syncMgr != nil {
		return s.syncMgr, nil
	}
	cfg := config.GetSyncConfig()
	if cfg == nil || cfg.Backend == "" {
		return nil, fmt.Errorf("sync not configured")
	}
	mgr, err := gosync.NewSyncManager(cfg)
	if err != nil {
		return nil, err
	}
	s.syncMgr = mgr
	return mgr, nil
}
