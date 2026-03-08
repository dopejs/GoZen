package daemon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/dopejs/gozen/internal/agent"
	"github.com/dopejs/gozen/internal/bot"
	"github.com/dopejs/gozen/internal/bot/adapters"
	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/httpx"
	"github.com/dopejs/gozen/internal/middleware"
	"github.com/dopejs/gozen/internal/proxy"
	gosync "github.com/dopejs/gozen/internal/sync"
	"github.com/dopejs/gozen/internal/web"
)

// FatalError represents an unrecoverable error that should not trigger auto-restart
type FatalError struct {
	Err error
}

func (e *FatalError) Error() string {
	return e.Err.Error()
}

func (e *FatalError) Unwrap() error {
	return e.Err
}

// IsFatalError checks if an error is a fatal error
func IsFatalError(err error) bool {
	var fatalErr *FatalError
	return errors.As(err, &fatalErr)
}

// Daemon is the zend main server that hosts both the proxy and web UI.
type Daemon struct {
	webServer      *web.Server
	proxyServer    *http.Server
	proxyMux       *http.ServeMux
	profileProxy   *proxy.ProfileProxy
	botGateway     *bot.Gateway
	logger         *log.Logger
	structuredLog  *StructuredLogger
	version        string
	watcher        *ConfigWatcher

	// Session tracking
	mu       sync.RWMutex
	sessions map[string]*SessionInfo // session ID -> info

	// Temporary profiles (for zen pick)
	tmpMu       sync.RWMutex
	tmpProfiles map[string]*TempProfile

	// Sync
	syncMgr    *gosync.SyncManager
	syncCancel context.CancelFunc // cancels auto-pull ticker
	pushTimer  *time.Timer        // debounced auto-push

	startTime time.Time
	proxyPort int
	webPort   int

	// Feature gates tracking (for detecting changes on reload)
	currentGates *config.FeatureGates

	// Shutdown channel - closed when shutdown is requested via API
	shutdownCh chan struct{}
	runCtx     context.Context
	runCancel  context.CancelFunc
	bgWG       sync.WaitGroup

	// Goroutine leak detection
	baselineGoroutines int
	leakCheckTicker    *time.Ticker

	// Metrics collection
	metrics *Metrics
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
	runCtx, runCancel := context.WithCancel(context.Background())

	// Create structured logger that writes to stderr
	structuredLog := NewStructuredLogger(os.Stderr)

	return &Daemon{
		version:       version,
		logger:        logger,
		structuredLog: structuredLog,
		sessions:      make(map[string]*SessionInfo),
		tmpProfiles:   make(map[string]*TempProfile),
		shutdownCh:    make(chan struct{}),
		runCtx:        runCtx,
		runCancel:     runCancel,
		metrics:       NewMetrics(),
	}
}

// Start initializes and starts both the proxy and web servers.
func (d *Daemon) Start() error {
	d.startTime = time.Now()

	// Ensure proxy port is persisted on first run
	if err := config.EnsureProxyPort(); err != nil {
		d.logger.Printf("Warning: failed to persist proxy port: %v", err)
	}

	d.proxyPort = config.GetProxyPort()
	d.webPort = config.GetWebPort()

	// Initialize current feature gates for change detection
	d.currentGates = config.GetFeatureGates()

	// Initialize structured logger for proxy logs (SQLite)
	if err := proxy.InitGlobalLogger(config.ConfigDirPath()); err != nil {
		d.logger.Printf("Warning: failed to initialize structured logger: %v", err)
	}

	// Set daemon structured logger for proxy and httpx selective logging
	proxy.SetDaemonLogger(d.structuredLog)
	httpx.SetDaemonLogger(d.structuredLog)

	// Initialize usage tracker, budget checker, health checker, and load balancer
	logDB := proxy.GetGlobalLogDB()
	proxy.InitGlobalUsageTracker(logDB)
	proxy.InitGlobalBudgetChecker(proxy.GetGlobalUsageTracker())
	proxy.InitGlobalHealthChecker(logDB)
	proxy.InitGlobalLoadBalancer(logDB)

	// Initialize context compressor (BETA)
	proxy.InitGlobalCompressor(nil) // providers will be set per-request

	// Initialize middleware registry (BETA)
	middleware.InitGlobalRegistry(d.logger)
	if registry := middleware.GetGlobalRegistry(); registry != nil {
		if err := registry.LoadFromConfig(); err != nil {
			d.logger.Printf("Warning: failed to load middleware config: %v", err)
		}
	}

	// Initialize agent infrastructure (BETA)
	agent.InitGlobalObservatory()
	agent.InitGlobalGuardrails()
	agent.InitGlobalCoordinator()
	agent.InitGlobalTaskQueue()
	agent.InitGlobalRuntime(d.proxyPort)

	// Start health checker if enabled
	proxy.StartGlobalHealthChecker()

	// Generate web password on first start if not configured
	if config.GetWebPasswordHash() == "" {
		if password, err := web.GeneratePassword(); err == nil {
			d.logger.Printf("Web UI password generated: %s (change in Web UI Settings)", password)
		} else {
			d.logger.Printf("Warning: failed to generate web password: %v", err)
		}
	}

	// Start proxy server
	if err := d.startProxy(); err != nil {
		return fmt.Errorf("proxy server: %w", err)
	}

	// Start web server (includes daemon API routes)
	d.webServer = web.NewServer(d.version, d.logger, 0)

	// Register daemon API routes on the web server
	d.webServer.HandleFunc("/api/v1/daemon/status", d.handleDaemonStatus)
	d.webServer.HandleFunc("/api/v1/daemon/health", d.handleDaemonHealth)
	d.webServer.HandleFunc("/api/v1/daemon/metrics", d.handleDaemonMetrics)
	d.webServer.HandleFunc("/api/v1/daemon/shutdown", d.handleDaemonShutdown)
	d.webServer.HandleFunc("/api/v1/daemon/reload", d.handleDaemonReload)
	d.webServer.HandleFunc("/api/v1/daemon/sessions", d.handleDaemonSessions)
	d.webServer.HandleFunc("/api/v1/profiles/temp", d.handleTempProfiles)
	d.webServer.HandleFunc("/api/v1/profiles/temp/", d.handleTempProfile)

	// Start config watcher
	d.watcher = NewConfigWatcher(d.logger, d.onConfigReload)
	go d.watcher.Start()

	// Start session cleanup goroutine
	d.bgWG.Add(1)
	go d.sessionCleanupLoop(d.runCtx)

	// Start goroutine leak detection monitor
	d.baselineGoroutines = runtime.NumGoroutine()
	d.leakCheckTicker = time.NewTicker(1 * time.Minute)
	d.bgWG.Add(1)
	go d.goroutineLeakMonitor(d.runCtx)

	// Initialize sync if configured
	d.initSync()

	// Initialize bot bridge for session tracking
	proxy.InitBotBridge("")

	// Initialize bot gateway if configured
	d.initBot()

	d.logger.Printf("zend started: proxy=:%d web=:%d", d.proxyPort, d.webPort)

	// Start web server (blocks)
	err := d.webServer.Start()

	// If web server fails to start, clean up resources
	if err != nil {
		d.logger.Printf("web server failed to start, cleaning up: %v", err)

		// Cancel background goroutines
		d.runCancel()

		// Stop proxy server
		if d.proxyServer != nil {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			d.proxyServer.Shutdown(shutdownCtx)
			cancel()
		}

		// Stop watcher
		if d.watcher != nil {
			d.watcher.Stop()
		}

		// Stop leak check ticker
		if d.leakCheckTicker != nil {
			d.leakCheckTicker.Stop()
		}

		// Wait for background goroutines to finish
		d.bgWG.Wait()

		// Wrap as FatalError since web port conflict is unrecoverable
		return &FatalError{Err: fmt.Errorf("web server: %w", err)}
	}

	return nil
}

// startProxy creates and starts the proxy HTTP server on the configured port.
func (d *Daemon) startProxy() error {
	d.proxyMux = http.NewServeMux()

	// Create profile-based proxy router
	d.profileProxy = proxy.NewProfileProxy(d.logger)
	d.profileProxy.TempProfiles = d
	d.profileProxy.MetricsRecorder = d.metrics

	// Daemon API routes on the proxy mux (for internal use)
	d.proxyMux.HandleFunc("/api/v1/daemon/status", d.handleDaemonStatus)
	d.proxyMux.HandleFunc("/api/v1/daemon/health", d.handleDaemonHealth)
	d.proxyMux.HandleFunc("/api/v1/daemon/metrics", d.handleDaemonMetrics)
	d.proxyMux.HandleFunc("/api/v1/daemon/shutdown", d.handleDaemonShutdown)
	d.proxyMux.HandleFunc("/api/v1/daemon/reload", d.handleDaemonReload)
	d.proxyMux.HandleFunc("/api/v1/daemon/sessions", d.handleDaemonSessions)
	d.proxyMux.HandleFunc("/api/v1/profiles/temp", d.handleTempProfiles)
	d.proxyMux.HandleFunc("/api/v1/profiles/temp/", d.handleTempProfile)

	// Default handler: profile-based proxy routing
	// URL format: /<profile>/<session>/v1/messages
	d.proxyMux.HandleFunc("/", d.profileProxy.ServeHTTP)

	addr := fmt.Sprintf("127.0.0.1:%d", d.proxyPort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		// Port is busy — identify who's using it
		pid, procName, identErr := GetProcessOnPort(d.proxyPort)
		if identErr != nil {
			return &FatalError{Err: fmt.Errorf("port %d is already in use (cannot identify process): %w", d.proxyPort, err)}
		}

		if IsZenProcess(procName) {
			// Stale zen daemon — kill it and retry
			d.logger.Printf("Port %d occupied by stale zen process (PID %d: %s), killing...", d.proxyPort, pid, procName)
			proc, findErr := os.FindProcess(pid)
			if findErr != nil {
				return &FatalError{Err: fmt.Errorf("port %d occupied by stale zen process (PID %d) but cannot find process: %w", d.proxyPort, pid, err)}
			}
			if killErr := proc.Signal(syscall.SIGTERM); killErr != nil {
				return &FatalError{Err: fmt.Errorf("port %d occupied by zen process (PID %d) but cannot kill (permission denied). Try: sudo kill %d", d.proxyPort, pid, pid)}
			}
			// Wait briefly for process to die
			for i := 0; i < 10; i++ {
				time.Sleep(300 * time.Millisecond)
				if proc.Signal(syscall.Signal(0)) != nil {
					break // process is dead
				}
			}
			d.logger.Printf("Daemon restarted (replaced stale process %d)", pid)

			// Retry bind
			ln, err = net.Listen("tcp", addr)
			if err != nil {
				return &FatalError{Err: fmt.Errorf("port %d still in use after killing stale process: %w", d.proxyPort, err)}
			}
		} else {
			return &FatalError{Err: fmt.Errorf("port %d is occupied by %s (PID %d) — not a zen process. Use 'zen config set proxy_port <port>' to change the proxy port", d.proxyPort, procName, pid)}
		}
	}

	d.proxyServer = &http.Server{
		Handler:           httpx.Recover(d.logger, "proxy", d.proxyMux),
		ReadHeaderTimeout: 15 * time.Second,
		ReadTimeout:       2 * time.Minute,
		WriteTimeout:      10 * time.Minute,
		IdleTimeout:       90 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		if err := d.proxyServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			d.logger.Printf("proxy server error: %v", err)
		}
	}()

	d.logger.Printf("proxy server listening on %s", addr)

	// Log daemon_started event with structured logging
	d.structuredLog.Info("daemon_started", map[string]interface{}{
		"pid":         os.Getpid(),
		"version":     d.version,
		"proxy_port":  d.proxyPort,
		"web_port":    d.webPort,
		"config_path": config.ConfigDirPath(),
	})

	return nil
}

// Shutdown gracefully stops the daemon.
func (d *Daemon) Shutdown(ctx context.Context) error {
	d.logger.Println("shutting down zend...")

	// Stop bot gateway
	if d.botGateway != nil {
		d.botGateway.Stop()
	}

	// Stop health checker
	proxy.StopGlobalHealthChecker()

	// Stop sync auto-pull ticker
	if d.syncCancel != nil {
		d.syncCancel()
	}
	if d.pushTimer != nil {
		d.pushTimer.Stop()
	}
	if d.leakCheckTicker != nil {
		d.leakCheckTicker.Stop()
	}
	if d.runCancel != nil {
		d.runCancel()
	}

	// Stop config watcher
	if d.watcher != nil {
		d.watcher.Stop()
	}
	d.bgWG.Wait()

	// Shutdown proxy server
	if d.proxyServer != nil {
		if err := d.proxyServer.Shutdown(ctx); err != nil {
			d.logger.Printf("proxy shutdown error: %v", err)
		}
	}
	if d.profileProxy != nil {
		d.profileProxy.Close()
	}

	// Shutdown web server
	if d.webServer != nil {
		if err := d.webServer.Shutdown(ctx); err != nil {
			d.logger.Printf("web shutdown error: %v", err)
		}
	}

	// Remove PID file
	os.Remove(DaemonPidPath())

	// Log daemon_shutdown event with structured logging
	uptime := time.Since(d.startTime)
	d.structuredLog.Info("daemon_shutdown", map[string]interface{}{
		"uptime_seconds": uptime.Seconds(),
		"reason":         "graceful_shutdown",
	})

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
func (d *Daemon) sessionCleanupLoop(ctx context.Context) {
	defer d.bgWG.Done()
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
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

// gatesChanged compares two FeatureGates and returns true if any field differs.
func gatesChanged(old, new *config.FeatureGates) bool {
	// Treat nil as empty struct
	if old == nil {
		old = &config.FeatureGates{}
	}
	if new == nil {
		new = &config.FeatureGates{}
	}

	return old.Bot != new.Bot ||
		old.Compression != new.Compression ||
		old.Middleware != new.Middleware ||
		old.Agent != new.Agent
}

// logFeatureGateChanges logs each changed field in FeatureGates.
func (d *Daemon) logFeatureGateChanges(old, new *config.FeatureGates) {
	// Treat nil as empty struct
	if old == nil {
		old = &config.FeatureGates{}
	}
	if new == nil {
		new = &config.FeatureGates{}
	}

	if old.Bot != new.Bot {
		d.logger.Printf("Feature gate changed: bot: %v → %v", old.Bot, new.Bot)
	}
	if old.Compression != new.Compression {
		d.logger.Printf("Feature gate changed: compression: %v → %v", old.Compression, new.Compression)
	}
	if old.Middleware != new.Middleware {
		d.logger.Printf("Feature gate changed: middleware: %v → %v", old.Middleware, new.Middleware)
	}
	if old.Agent != new.Agent {
		d.logger.Printf("Feature gate changed: agent: %v → %v", old.Agent, new.Agent)
	}
}

// onConfigReload is called when the config file changes.
func (d *Daemon) onConfigReload() {
	d.logger.Println("config file changed, reloading...")

	// Capture old feature gates from daemon state (before reload)
	oldGates := d.currentGates

	config.ResetDefaultStore()

	// Detect and log feature gate changes
	newGates := config.GetFeatureGates()
	if gatesChanged(oldGates, newGates) {
		d.logFeatureGateChanges(oldGates, newGates)
	}

	// Update current gates
	d.currentGates = newGates

	// Protect running ports: revert any port changes to preserve active sessions.
	// Port changes only take effect on daemon restart.
	if newProxy := config.GetProxyPort(); newProxy != d.proxyPort {
		d.logger.Printf("WARNING: proxy_port changed %d→%d in config, reverting (active sessions depend on port %d; restart daemon to apply)", newProxy, d.proxyPort, d.proxyPort)
		config.SetProxyPort(d.proxyPort)
	}
	if newWeb := config.GetWebPort(); newWeb != d.webPort {
		d.logger.Printf("WARNING: web_port changed %d→%d in config, reverting (restart daemon to apply)", newWeb, d.webPort)
		config.SetWebPort(d.webPort)
	}

	// Invalidate proxy cache so new config takes effect
	if d.profileProxy != nil {
		d.profileProxy.InvalidateCache()
	}
	if checker := proxy.GetGlobalHealthChecker(); checker != nil {
		checker.ReloadConfig()
		proxy.StartGlobalHealthChecker()
	}
	// Reinitialize sync if config changed
	d.initSync()
	// Reinitialize bot gateway if config changed
	d.reinitBot()
	d.logger.Println("config reloaded successfully")
}

// initSync initializes or reinitializes the sync manager from current config.
func (d *Daemon) initSync() {
	// Stop existing auto-pull
	if d.syncCancel != nil {
		d.syncCancel()
		d.syncCancel = nil
	}

	cfg := config.GetSyncConfig()
	if cfg == nil || cfg.Backend == "" {
		d.syncMgr = nil
		if d.webServer != nil {
			d.webServer.SetSyncManager(nil)
		}
		return
	}

	mgr, err := gosync.NewSyncManager(cfg)
	if err != nil {
		d.logger.Printf("sync init failed: %v", err)
		return
	}
	d.syncMgr = mgr

	// Pass to web server
	if d.webServer != nil {
		d.webServer.SetSyncManager(mgr)
	}

	// Register auto-push hook (debounced)
	store := config.DefaultStore()
	store.SetOnSave(func() {
		if mgr.IsPulling() {
			return
		}
		if d.pushTimer != nil {
			d.pushTimer.Stop()
		}
		d.pushTimer = time.AfterFunc(2*time.Second, func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := mgr.Push(ctx); err != nil {
				d.logger.Printf("sync auto-push failed: %v", err)
			} else {
				d.logger.Println("sync auto-push completed")
			}
		})
	})

	// Start auto-pull ticker if enabled
	if cfg.AutoPull {
		interval := time.Duration(cfg.PullInterval) * time.Second
		if interval < 60*time.Second {
			interval = 5 * time.Minute // default 5 min
		}
		ctx, cancel := context.WithCancel(context.Background())
		d.syncCancel = cancel
		go func() {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					pullCtx, pullCancel := context.WithTimeout(ctx, 30*time.Second)
					err := mgr.Pull(pullCtx)
					pullCancel()
					if err != nil {
						d.logger.Printf("sync auto-pull failed: %v", err)
					}
				}
			}
		}()
		d.logger.Printf("sync auto-pull enabled (interval: %s)", interval)
	}

	d.logger.Printf("sync initialized (backend: %s)", cfg.Backend)
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

// GetTempProfileProviders implements proxy.TempProfileProvider.
func (d *Daemon) GetTempProfileProviders(id string) []string {
	d.tmpMu.RLock()
	defer d.tmpMu.RUnlock()
	if tp, ok := d.tmpProfiles[id]; ok {
		return tp.Providers
	}
	return nil
}

// randomID generates a short random hex ID.
func randomID() string {
	b := make([]byte, 4)
	f, err := os.Open("/dev/urandom")
	if err == nil {
		_, readErr := f.Read(b)
		f.Close()
		if readErr == nil {
			return fmt.Sprintf("%x", b)
		}
	}
	// Fallback: use time-based
	t := time.Now().UnixNano()
	b[0] = byte(t >> 24)
	b[1] = byte(t >> 16)
	b[2] = byte(t >> 8)
	b[3] = byte(t)
	return fmt.Sprintf("%x", b)
}

// reinitBot stops the existing bot gateway (if any) and starts a new one if configured.
func (d *Daemon) reinitBot() {
	if d.botGateway != nil {
		d.botGateway.Stop()
		d.botGateway = nil
		d.logger.Println("Bot gateway stopped for reload")
	}
	d.initBot()
}

// initBot initializes the bot gateway if configured.
func (d *Daemon) initBot() {
	cfg := config.GetBot()
	if cfg == nil || !cfg.Enabled {
		return
	}

	// Convert config types to bot gateway config
	gwConfig := &bot.GatewayConfig{
		Enabled:     cfg.Enabled,
		SocketPath:  cfg.SocketPath,
		Profile:     cfg.Profile,
		Model:       cfg.Model,
		ProxyPort:   d.proxyPort,
		Aliases:     cfg.Aliases,
		HistorySize: cfg.HistorySize,
	}

	// Convert platform configs
	if cfg.Platforms != nil {
		gwConfig.Platforms = bot.PlatformsConfig{}

		if cfg.Platforms.Telegram != nil {
			gwConfig.Platforms.Telegram = &adapters.TelegramConfig{
				AdapterConfig: adapters.AdapterConfig{
					Enabled:      cfg.Platforms.Telegram.Enabled,
					AllowedUsers: cfg.Platforms.Telegram.AllowedUsers,
					AllowedChats: cfg.Platforms.Telegram.AllowedChats,
				},
				Token: cfg.Platforms.Telegram.Token,
			}
		}

		if cfg.Platforms.Discord != nil {
			gwConfig.Platforms.Discord = &adapters.DiscordConfig{
				AdapterConfig: adapters.AdapterConfig{
					Enabled:         cfg.Platforms.Discord.Enabled,
					AllowedUsers:    cfg.Platforms.Discord.AllowedUsers,
					AllowedChannels: cfg.Platforms.Discord.AllowedChannels,
				},
				Token:         cfg.Platforms.Discord.Token,
				AllowedGuilds: cfg.Platforms.Discord.AllowedGuilds,
			}
		}

		if cfg.Platforms.Slack != nil {
			gwConfig.Platforms.Slack = &adapters.SlackConfig{
				AdapterConfig: adapters.AdapterConfig{
					Enabled:         cfg.Platforms.Slack.Enabled,
					AllowedUsers:    cfg.Platforms.Slack.AllowedUsers,
					AllowedChannels: cfg.Platforms.Slack.AllowedChannels,
				},
				BotToken: cfg.Platforms.Slack.BotToken,
				AppToken: cfg.Platforms.Slack.AppToken,
			}
		}

		if cfg.Platforms.Lark != nil {
			gwConfig.Platforms.Lark = &adapters.LarkConfig{
				AdapterConfig: adapters.AdapterConfig{
					Enabled:      cfg.Platforms.Lark.Enabled,
					AllowedUsers: cfg.Platforms.Lark.AllowedUsers,
					AllowedChats: cfg.Platforms.Lark.AllowedChats,
				},
				AppID:     cfg.Platforms.Lark.AppID,
				AppSecret: cfg.Platforms.Lark.AppSecret,
			}
		}

		if cfg.Platforms.FBMessenger != nil {
			gwConfig.Platforms.FBMessenger = &adapters.FBMessengerConfig{
				AdapterConfig: adapters.AdapterConfig{
					Enabled:      cfg.Platforms.FBMessenger.Enabled,
					AllowedUsers: cfg.Platforms.FBMessenger.AllowedUsers,
				},
				PageToken:   cfg.Platforms.FBMessenger.PageToken,
				VerifyToken: cfg.Platforms.FBMessenger.VerifyToken,
				AppSecret:   cfg.Platforms.FBMessenger.AppSecret,
			}
		}
	}

	// Convert interaction config
	if cfg.Interaction != nil {
		gwConfig.Interaction = bot.InteractionConfig{
			RequireMention:  cfg.Interaction.RequireMention,
			MentionKeywords: cfg.Interaction.MentionKeywords,
			DirectMsgMode:   cfg.Interaction.DirectMsgMode,
			ChannelMode:     cfg.Interaction.ChannelMode,
		}
	} else {
		gwConfig.Interaction = bot.InteractionConfig{
			RequireMention: true,
		}
	}

	// Convert notification config
	if cfg.Notify != nil && cfg.Notify.DefaultPlatform != "" {
		gwConfig.Notifications = bot.NotifyConfig{
			DefaultChat: &struct {
				Platform bot.Platform `json:"platform"`
				ChatID   string       `json:"chat_id"`
			}{
				Platform: bot.Platform(cfg.Notify.DefaultPlatform),
				ChatID:   cfg.Notify.DefaultChatID,
			},
		}
		if cfg.Notify.QuietHoursStart != "" {
			gwConfig.Notifications.QuietHours = &struct {
				Enabled  bool   `json:"enabled"`
				Start    string `json:"start"`
				End      string `json:"end"`
				Timezone string `json:"timezone"`
			}{
				Enabled:  true,
				Start:    cfg.Notify.QuietHoursStart,
				End:      cfg.Notify.QuietHoursEnd,
				Timezone: cfg.Notify.QuietHoursZone,
			}
		}
	}

	d.botGateway = bot.NewGateway(gwConfig, d.logger)
	if err := d.botGateway.Start(context.Background()); err != nil {
		d.logger.Printf("Failed to start bot gateway: %v", err)
		d.botGateway = nil
		return
	}

	// Connect bot bridge to gateway for session tracking
	if bridge := proxy.GetBotBridge(); bridge != nil {
		d.botGateway.SetSessionProvider(bridge)
		d.logger.Println("Bot bridge connected to gateway")
	}

	d.logger.Println("Bot gateway started")

	// Expose gateway to web server for skill management API
	if d.webServer != nil {
		d.webServer.SetBotGateway(d.botGateway)
	}
}

// goroutineLeakMonitor checks for goroutine leaks every minute
func (d *Daemon) goroutineLeakMonitor(ctx context.Context) {
	defer d.bgWG.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case <-d.leakCheckTicker.C:
			current := runtime.NumGoroutine()

			// Update resource peaks for metrics
			if d.metrics != nil {
				var mem runtime.MemStats
				runtime.ReadMemStats(&mem)
				memoryMB := int64(mem.Alloc / 1024 / 1024)
				d.metrics.UpdateResourcePeaks(current, memoryMB)
			}

			// Allow 20% growth tolerance for normal fluctuations
			threshold := d.baselineGoroutines + (d.baselineGoroutines / 5)
			if current > threshold {
				// Potential leak detected, dump stack traces
				buf := make([]byte, 1<<20) // 1MB buffer
				stackLen := runtime.Stack(buf, true)
				d.logger.Printf("[goroutine-leak] detected: baseline=%d current=%d threshold=%d\n%s",
					d.baselineGoroutines, current, threshold, buf[:stackLen])

				// Log structured event
				d.structuredLog.Warn("goroutine_leak_detected", map[string]interface{}{
					"baseline_goroutines": d.baselineGoroutines,
					"current_goroutines":  current,
					"threshold":           threshold,
					"growth_percent":      float64(current-d.baselineGoroutines) / float64(d.baselineGoroutines) * 100,
				})
			}
		}
	}
}
