package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dopejs/gozen/internal/bot/adapters"
)

// GatewayConfig is the configuration for the bot gateway.
type GatewayConfig struct {
	Enabled       bool              `json:"enabled"`
	SocketPath    string            `json:"socket_path,omitempty"`
	Profile       string            `json:"profile,omitempty"` // Profile for NLU
	Platforms     PlatformsConfig   `json:"platforms"`
	Interaction   InteractionConfig `json:"interaction"`
	Aliases       map[string]string `json:"aliases,omitempty"`
	Notifications NotifyConfig      `json:"notifications,omitempty"`
}

// PlatformsConfig contains configuration for all platforms.
type PlatformsConfig struct {
	Telegram    *adapters.TelegramConfig    `json:"telegram,omitempty"`
	Discord     *adapters.DiscordConfig     `json:"discord,omitempty"`
	Slack       *adapters.SlackConfig       `json:"slack,omitempty"`
	Lark        *adapters.LarkConfig        `json:"lark,omitempty"`
	FBMessenger *adapters.FBMessengerConfig `json:"fbmessenger,omitempty"`
}

// InteractionConfig controls how the bot responds to messages.
type InteractionConfig struct {
	RequireMention  bool     `json:"require_mention"`
	MentionKeywords []string `json:"mention_keywords,omitempty"`
	DirectMsgMode   string   `json:"direct_message_mode,omitempty"` // always, mention
	ChannelMode     string   `json:"channel_mode,omitempty"`        // always, mention
}

// NotifyConfig controls notification behavior.
type NotifyConfig struct {
	DefaultChat *struct {
		Platform Platform `json:"platform"`
		ChatID   string   `json:"chat_id"`
	} `json:"default_chat,omitempty"`
	QuietHours *struct {
		Enabled  bool   `json:"enabled"`
		Start    string `json:"start"`
		End      string `json:"end"`
		Timezone string `json:"timezone"`
	} `json:"quiet_hours,omitempty"`
}

// Gateway is the central bot gateway that manages adapters and routes messages.
type Gateway struct {
	config   *GatewayConfig
	logger   *log.Logger
	adapters []adapters.Adapter

	registry    *Registry
	sessions    *SessionManager
	approvals   *ApprovalManager
	nlu         *NLUParser
	listener    net.Listener
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	mu          sync.RWMutex
	connections map[string]net.Conn // processID -> connection
}

// NewGateway creates a new bot gateway.
func NewGateway(config *GatewayConfig, logger *log.Logger) *Gateway {
	if config.SocketPath == "" {
		config.SocketPath = filepath.Join(os.TempDir(), "zen-gateway.sock")
	}

	keywords := config.Interaction.MentionKeywords
	if len(keywords) == 0 {
		keywords = []string{"@zen", "/zen", "zen"}
	}

	return &Gateway{
		config:      config,
		logger:      logger,
		registry:    NewRegistry(config.Aliases),
		sessions:    NewSessionManager(),
		approvals:   NewApprovalManager(),
		nlu:         NewNLUParser(keywords),
		connections: make(map[string]net.Conn),
	}
}

// Start starts the gateway and all enabled adapters.
func (g *Gateway) Start(ctx context.Context) error {
	g.ctx, g.cancel = context.WithCancel(ctx)

	// Start IPC listener
	if err := g.startIPCListener(); err != nil {
		return fmt.Errorf("failed to start IPC listener: %w", err)
	}

	// Initialize and start adapters
	if err := g.initAdapters(); err != nil {
		return fmt.Errorf("failed to initialize adapters: %w", err)
	}

	// Start cleanup goroutines
	g.wg.Add(1)
	go g.cleanupLoop()

	g.logger.Printf("Bot gateway started (socket: %s)", g.config.SocketPath)
	return nil
}

// Stop stops the gateway and all adapters.
func (g *Gateway) Stop() error {
	g.logger.Println("Stopping bot gateway...")

	if g.cancel != nil {
		g.cancel()
	}

	// Stop all adapters
	for _, adapter := range g.adapters {
		adapter.Stop()
	}

	// Close IPC listener
	if g.listener != nil {
		g.listener.Close()
	}

	// Close all connections
	g.mu.Lock()
	for _, conn := range g.connections {
		conn.Close()
	}
	g.mu.Unlock()

	// Remove socket file
	os.Remove(g.config.SocketPath)

	g.wg.Wait()
	g.logger.Println("Bot gateway stopped")
	return nil
}

func (g *Gateway) initAdapters() error {
	cfg := g.config.Platforms

	// Telegram
	if cfg.Telegram != nil && cfg.Telegram.Enabled && cfg.Telegram.Token != "" {
		adapter := adapters.NewTelegramAdapter(cfg.Telegram)
		adapter.SetMessageHandler(g.handleMessage)
		adapter.SetButtonHandler(g.handleButtonClick)
		if err := adapter.Start(g.ctx); err != nil {
			g.logger.Printf("Failed to start Telegram adapter: %v", err)
		} else {
			g.adapters = append(g.adapters, adapter)
			g.logger.Println("Telegram adapter started")
		}
	}

	// Discord
	if cfg.Discord != nil && cfg.Discord.Enabled && cfg.Discord.Token != "" {
		adapter := adapters.NewDiscordAdapter(cfg.Discord)
		adapter.SetMessageHandler(g.handleMessage)
		adapter.SetButtonHandler(g.handleButtonClick)
		if err := adapter.Start(g.ctx); err != nil {
			g.logger.Printf("Failed to start Discord adapter: %v", err)
		} else {
			g.adapters = append(g.adapters, adapter)
			g.logger.Println("Discord adapter started")
		}
	}

	// Slack
	if cfg.Slack != nil && cfg.Slack.Enabled && cfg.Slack.BotToken != "" {
		adapter := adapters.NewSlackAdapter(cfg.Slack)
		adapter.SetMessageHandler(g.handleMessage)
		adapter.SetButtonHandler(g.handleButtonClick)
		if err := adapter.Start(g.ctx); err != nil {
			g.logger.Printf("Failed to start Slack adapter: %v", err)
		} else {
			g.adapters = append(g.adapters, adapter)
			g.logger.Println("Slack adapter started")
		}
	}

	// Lark
	if cfg.Lark != nil && cfg.Lark.Enabled && cfg.Lark.AppID != "" {
		adapter := adapters.NewLarkAdapter(cfg.Lark)
		adapter.SetMessageHandler(g.handleMessage)
		adapter.SetButtonHandler(g.handleButtonClick)
		if err := adapter.Start(g.ctx); err != nil {
			g.logger.Printf("Failed to start Lark adapter: %v", err)
		} else {
			g.adapters = append(g.adapters, adapter)
			g.logger.Println("Lark adapter started")
		}
	}

	// FB Messenger
	if cfg.FBMessenger != nil && cfg.FBMessenger.Enabled && cfg.FBMessenger.PageToken != "" {
		adapter := adapters.NewFBMessengerAdapter(cfg.FBMessenger)
		adapter.SetMessageHandler(g.handleMessage)
		adapter.SetButtonHandler(g.handleButtonClick)
		if err := adapter.Start(g.ctx); err != nil {
			g.logger.Printf("Failed to start FB Messenger adapter: %v", err)
		} else {
			g.adapters = append(g.adapters, adapter)
			g.logger.Println("FB Messenger adapter started")
		}
	}

	if len(g.adapters) == 0 {
		g.logger.Println("Warning: No chat adapters enabled")
	}

	return nil
}

func (g *Gateway) startIPCListener() error {
	// Remove existing socket
	os.Remove(g.config.SocketPath)

	listener, err := net.Listen("unix", g.config.SocketPath)
	if err != nil {
		return err
	}
	g.listener = listener

	g.wg.Add(1)
	go g.acceptConnections()

	return nil
}

func (g *Gateway) acceptConnections() {
	defer g.wg.Done()

	for {
		conn, err := g.listener.Accept()
		if err != nil {
			select {
			case <-g.ctx.Done():
				return
			default:
				g.logger.Printf("Accept error: %v", err)
				continue
			}
		}

		g.wg.Add(1)
		go g.handleConnection(conn)
	}
}

func (g *Gateway) handleConnection(conn net.Conn) {
	defer g.wg.Done()
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	var processID string

	for {
		select {
		case <-g.ctx.Done():
			return
		default:
		}

		var msg IPCMessage
		if err := decoder.Decode(&msg); err != nil {
			if processID != "" {
				g.registry.Unregister(processID)
				g.mu.Lock()
				delete(g.connections, processID)
				g.mu.Unlock()
				g.logger.Printf("Process disconnected: %s", processID)
			}
			return
		}

		switch msg.Type {
		case IPCRegister:
			var payload RegisterPayload
			json.Unmarshal(msg.Payload, &payload)
			processID = payload.ProcessID

			info := &ProcessInfo{
				ID:         payload.ProcessID,
				Path:       payload.ProcessPath,
				SocketPath: payload.SocketPath,
				PID:        payload.PID,
				Status:     "idle",
				StartTime:  time.Now(),
			}
			g.registry.Register(info, conn)

			g.mu.Lock()
			g.connections[processID] = conn
			g.mu.Unlock()

			g.logger.Printf("Process registered: %s (%s)", info.Name, info.Path)

		case IPCHeartbeat:
			var payload HeartbeatPayload
			json.Unmarshal(msg.Payload, &payload)
			g.registry.UpdateStatus(payload.ProcessID, payload.Status, payload.CurrentTask)

		case IPCNotification:
			var payload NotificationPayload
			json.Unmarshal(msg.Payload, &payload)
			g.handleNotification(processID, &payload)

		case IPCApproval:
			var payload ApprovalPayload
			json.Unmarshal(msg.Payload, &payload)
			g.handleApprovalRequest(processID, &payload)

		case IPCResponse:
			// Response to a command - handled via request ID
			var payload ResponsePayload
			json.Unmarshal(msg.Payload, &payload)
			g.handleProcessResponse(msg.RequestID, &payload)
		}
	}
}

func (g *Gateway) cleanupLoop() {
	defer g.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-g.ctx.Done():
			return
		case <-ticker.C:
			// Cleanup stale processes
			removed := g.registry.CleanupStale(30 * time.Second)
			for _, name := range removed {
				g.logger.Printf("Removed stale process: %s", name)
			}

			// Cleanup stale sessions
			g.sessions.Cleanup(24 * time.Hour)

			// Cleanup expired approvals
			g.approvals.Cleanup()
		}
	}
}
