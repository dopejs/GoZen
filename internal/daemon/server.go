package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/proxy"
	"github.com/dopejs/gozen/internal/web"
)

// Daemon is the zend main server that hosts both the proxy and web UI.
type Daemon struct {
	webServer   *web.Server
	proxyServer *http.Server
	proxyMux    *http.ServeMux
	logger      *log.Logger
	version     string
	watcher     *ConfigWatcher

	// Session tracking
	mu       sync.RWMutex
	sessions map[string]*SessionInfo // session ID -> info

	// Temporary profiles (for zen pick)
	tmpMu       sync.RWMutex
	tmpProfiles map[string]*TempProfile

	startTime time.Time
	proxyPort int
	webPort   int
}

// SessionInfo tracks an active client session.
type SessionInfo struct {
	ID        string    `json:"id"`
	Profile   string    `json:"profile"`
	Client    string    `json:"client"`
	CreatedAt time.Time `json:"created_at"`
	LastSeen  time.Time `json:"last_seen"`
}

// TempProfile is a temporary profile registered by zen pick.
type TempProfile struct {
	ID        string   `json:"id"`
	Providers []string `json:"providers"`
	CreatedAt time.Time
}

// DaemonPidPath returns the path to the zend PID file.
func DaemonPidPath() string {
	return filepath.Join(config.ConfigDirPath(), config.DaemonPidFile)
}

// DaemonLogPath returns the path to the zend log file.
func DaemonLogPath() string {
	return filepath.Join(config.ConfigDirPath(), config.DaemonLogFile)
}

// NewDaemon creates a new zend daemon instance.
func NewDaemon(version string, logger *log.Logger) *Daemon {
	return &Daemon{
		version:     version,
		logger:      logger,
		sessions:    make(map[string]*SessionInfo),
		tmpProfiles: make(map[string]*TempProfile),
	}
}

// Start initializes and starts both the proxy and web servers.
func (d *Daemon) Start() error {
	d.startTime = time.Now()
	d.proxyPort = config.GetProxyPort()
	d.webPort = config.GetWebPort()

	// Initialize structured logger for proxy logs (SQLite)
	if err := proxy.InitGlobalLogger(config.ConfigDirPath()); err != nil {
		d.logger.Printf("Warning: failed to initialize structured logger: %v", err)
	}

	// Start proxy server
	if err := d.startProxy(); err != nil {
		return fmt.Errorf("proxy server: %w", err)
	}

	// Start web server (includes daemon API routes)
	d.webServer = web.NewServer(d.version, d.logger, 0)

	// Register daemon API routes on the web server
	d.webServer.HandleFunc("/api/v1/daemon/status", d.handleDaemonStatus)
	d.webServer.HandleFunc("/api/v1/daemon/reload", d.handleDaemonReload)
	d.webServer.HandleFunc("/api/v1/daemon/sessions", d.handleDaemonSessions)
	d.webServer.HandleFunc("/api/v1/profiles/temp", d.handleTempProfiles)
	d.webServer.HandleFunc("/api/v1/profiles/temp/", d.handleTempProfile)

	// Start config watcher
	d.watcher = NewConfigWatcher(d.logger, d.onConfigReload)
	go d.watcher.Start()

	// Start session cleanup goroutine
	go d.sessionCleanupLoop()

	d.logger.Printf("zend started: proxy=:%d web=:%d", d.proxyPort, d.webPort)

	// Start web server (blocks)
	return d.webServer.Start()
}

// startProxy creates and starts the proxy HTTP server on the configured port.
func (d *Daemon) startProxy() error {
	d.proxyMux = http.NewServeMux()

	// Daemon API routes on the proxy mux (for internal use)
	d.proxyMux.HandleFunc("/api/v1/daemon/status", d.handleDaemonStatus)
	d.proxyMux.HandleFunc("/api/v1/daemon/reload", d.handleDaemonReload)
	d.proxyMux.HandleFunc("/api/v1/daemon/sessions", d.handleDaemonSessions)
	d.proxyMux.HandleFunc("/api/v1/profiles/temp", d.handleTempProfiles)
	d.proxyMux.HandleFunc("/api/v1/profiles/temp/", d.handleTempProfile)

	// Default handler: proxy requests (profile-based routing will be added in Task #4)
	// For now, return 502 since profile-based routing isn't implemented yet
	d.proxyMux.HandleFunc("/", d.handleProxyPlaceholder)

	addr := fmt.Sprintf("127.0.0.1:%d", d.proxyPort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("port %d is already in use: %w", d.proxyPort, err)
	}

	d.proxyServer = &http.Server{
		Handler: d.proxyMux,
	}

	go func() {
		if err := d.proxyServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			d.logger.Printf("proxy server error: %v", err)
		}
	}()

	d.logger.Printf("proxy server listening on %s", addr)
	return nil
}

// handleProxyPlaceholder returns 502 until profile-based routing is implemented.
func (d *Daemon) handleProxyPlaceholder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadGateway)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"type":    "proxy_not_configured",
			"message": "Profile-based routing not yet configured. Use zen to start a session.",
		},
	})
}

// Shutdown gracefully stops the daemon.
func (d *Daemon) Shutdown(ctx context.Context) error {
	d.logger.Println("shutting down zend...")

	// Stop config watcher
	if d.watcher != nil {
		d.watcher.Stop()
	}

	// Shutdown proxy server
	if d.proxyServer != nil {
		if err := d.proxyServer.Shutdown(ctx); err != nil {
			d.logger.Printf("proxy shutdown error: %v", err)
		}
	}

	// Shutdown web server
	if d.webServer != nil {
		if err := d.webServer.Shutdown(ctx); err != nil {
			d.logger.Printf("web shutdown error: %v", err)
		}
	}

	// Remove PID file
	os.Remove(DaemonPidPath())

	d.logger.Println("zend stopped")
	return nil
}

// ActiveSessionCount returns the number of active sessions.
func (d *Daemon) ActiveSessionCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.sessions)
}

// RegisterSession registers a new client session.
func (d *Daemon) RegisterSession(id, profile, client string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.sessions[id] = &SessionInfo{
		ID:        id,
		Profile:   profile,
		Client:    client,
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
	}
}

// TouchSession updates the last-seen time for a session.
func (d *Daemon) TouchSession(id string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if s, ok := d.sessions[id]; ok {
		s.LastSeen = time.Now()
	}
}

// RemoveSession removes a session.
func (d *Daemon) RemoveSession(id string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.sessions, id)
}

// sessionCleanupLoop periodically removes stale sessions.
func (d *Daemon) sessionCleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		d.mu.Lock()
		now := time.Now()
		for id, s := range d.sessions {
			// Remove sessions not seen for 2 hours
			if now.Sub(s.LastSeen) > 2*time.Hour {
				delete(d.sessions, id)
				d.logger.Printf("cleaned up stale session: %s", id)
			}
		}
		d.mu.Unlock()

		// Also clean up proxy session cache
		proxy.CleanupOldSessions(2 * time.Hour)
	}
}

// onConfigReload is called when the config file changes.
func (d *Daemon) onConfigReload() {
	d.logger.Println("config file changed, reloading...")
	config.ResetDefaultStore()
	d.logger.Println("config reloaded successfully")
}

// --- Temporary Profiles ---

// RegisterTempProfile creates a temporary profile and returns its ID.
func (d *Daemon) RegisterTempProfile(providers []string) string {
	d.tmpMu.Lock()
	defer d.tmpMu.Unlock()

	id := fmt.Sprintf("_tmp_%s", randomID())
	d.tmpProfiles[id] = &TempProfile{
		ID:        id,
		Providers: providers,
		CreatedAt: time.Now(),
	}
	return id
}

// GetTempProfile returns a temporary profile by ID.
func (d *Daemon) GetTempProfile(id string) *TempProfile {
	d.tmpMu.RLock()
	defer d.tmpMu.RUnlock()
	return d.tmpProfiles[id]
}

// RemoveTempProfile removes a temporary profile.
func (d *Daemon) RemoveTempProfile(id string) {
	d.tmpMu.Lock()
	defer d.tmpMu.Unlock()
	delete(d.tmpProfiles, id)
}

// randomID generates a short random hex ID.
func randomID() string {
	b := make([]byte, 4)
	f, _ := os.Open("/dev/urandom")
	if f != nil {
		f.Read(b)
		f.Close()
	} else {
		// Fallback: use time-based
		t := time.Now().UnixNano()
		b[0] = byte(t >> 24)
		b[1] = byte(t >> 16)
		b[2] = byte(t >> 8)
		b[3] = byte(t)
	}
	return fmt.Sprintf("%x", b)
}
