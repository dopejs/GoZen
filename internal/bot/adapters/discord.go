package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// DiscordAdapter implements the Adapter interface for Discord.
type DiscordAdapter struct {
	config        *DiscordConfig
	client        *http.Client
	botUserID     string
	botUsername   string
	msgHandler    func(*Message)
	buttonHandler func(*ButtonClick)
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	wsConn        *websocket.Conn
	wsMu          sync.Mutex
	sessionID     string
	sequence      atomic.Int64
	resumeURL     string
}

// NewDiscordAdapter creates a new Discord adapter.
func NewDiscordAdapter(config *DiscordConfig) *DiscordAdapter {
	return &DiscordAdapter{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (a *DiscordAdapter) Platform() Platform {
	return PlatformDiscord
}

func (a *DiscordAdapter) Start(ctx context.Context) error {
	a.ctx, a.cancel = context.WithCancel(ctx)

	// Get bot user info
	user, err := a.getCurrentUser()
	if err != nil {
		return fmt.Errorf("failed to get bot user: %w", err)
	}
	a.botUserID = user.ID
	a.botUsername = user.Username

	// Start Gateway WebSocket connection
	a.wg.Add(1)
	go a.gatewayLoop()

	return nil
}

func (a *DiscordAdapter) Stop() error {
	if a.cancel != nil {
		a.cancel()
	}
	a.wsMu.Lock()
	if a.wsConn != nil {
		a.wsConn.Close()
	}
	a.wsMu.Unlock()
	a.wg.Wait()
	return nil
}

func (a *DiscordAdapter) BotUserID() string {
	return a.botUserID
}

func (a *DiscordAdapter) SetMessageHandler(handler func(*Message)) {
	a.msgHandler = handler
}

func (a *DiscordAdapter) SetButtonHandler(handler func(*ButtonClick)) {
	a.buttonHandler = handler
}

func (a *DiscordAdapter) SendMessage(chatID string, msg *OutgoingMessage) (string, error) {
	return a.sendMessage(chatID, "", msg)
}

func (a *DiscordAdapter) SendReply(chatID, replyTo string, msg *OutgoingMessage) (string, error) {
	return a.sendMessage(chatID, replyTo, msg)
}

func (a *DiscordAdapter) sendMessage(chatID, replyTo string, msg *OutgoingMessage) (string, error) {
	payload := map[string]interface{}{
		"content": msg.Text,
	}

	if replyTo != "" {
		payload["message_reference"] = map[string]string{
			"message_id": replyTo,
		}
	}

	// Add buttons as components
	if len(msg.Buttons) > 0 {
		var buttons []map[string]interface{}
		for _, btn := range msg.Buttons {
			style := 1 // Primary
			switch btn.Style {
			case "secondary":
				style = 2
			case "danger":
				style = 4
			}
			buttons = append(buttons, map[string]interface{}{
				"type":      2, // Button
				"label":     btn.Label,
				"style":     style,
				"custom_id": btn.ID + ":" + btn.Data,
			})
		}
		payload["components"] = []map[string]interface{}{
			{
				"type":       1, // Action Row
				"components": buttons,
			},
		}
	}

	resp, err := a.apiCall("POST", "/channels/"+chatID+"/messages", payload)
	if err != nil {
		return "", err
	}

	var result struct {
		ID string `json:"id"`
	}
	json.Unmarshal(resp, &result)
	return result.ID, nil
}

func (a *DiscordAdapter) EditMessage(chatID, msgID string, msg *OutgoingMessage) error {
	payload := map[string]interface{}{
		"content": msg.Text,
	}

	if len(msg.Buttons) > 0 {
		var buttons []map[string]interface{}
		for _, btn := range msg.Buttons {
			style := 1
			switch btn.Style {
			case "secondary":
				style = 2
			case "danger":
				style = 4
			}
			buttons = append(buttons, map[string]interface{}{
				"type":      2,
				"label":     btn.Label,
				"style":     style,
				"custom_id": btn.ID + ":" + btn.Data,
			})
		}
		payload["components"] = []map[string]interface{}{
			{
				"type":       1,
				"components": buttons,
			},
		}
	} else {
		payload["components"] = []interface{}{}
	}

	_, err := a.apiCall("PATCH", "/channels/"+chatID+"/messages/"+msgID, payload)
	return err
}

func (a *DiscordAdapter) DeleteMessage(chatID, msgID string) error {
	_, err := a.apiCall("DELETE", "/channels/"+chatID+"/messages/"+msgID, nil)
	return err
}

func (a *DiscordAdapter) getCurrentUser() (*discordUser, error) {
	resp, err := a.apiCall("GET", "/users/@me", nil)
	if err != nil {
		return nil, err
	}

	var user discordUser
	if err := json.Unmarshal(resp, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (a *DiscordAdapter) apiCall(method, path string, payload interface{}) ([]byte, error) {
	url := "https://discord.com/api/v10" + path

	var body io.Reader
	if payload != nil {
		data, _ := json.Marshal(payload)
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(a.ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bot "+a.config.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("discord API error: %s", string(respBody))
	}

	return respBody, nil
}

// HandleInteraction handles incoming Discord interactions (for webhook mode).
func (a *DiscordAdapter) HandleInteraction(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	var interaction struct {
		Type int `json:"type"`
		Data struct {
			CustomID string `json:"custom_id"`
		} `json:"data"`
		ChannelID string `json:"channel_id"`
		Message   struct {
			ID string `json:"id"`
		} `json:"message"`
		Member *struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"member"`
		User *struct {
			ID string `json:"id"`
		} `json:"user"`
	}

	if err := json.Unmarshal(body, &interaction); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Ping
	if interaction.Type == 1 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"type": 1})
		return
	}

	// Message Component (button click)
	if interaction.Type == 3 && a.buttonHandler != nil {
		userID := ""
		if interaction.Member != nil {
			userID = interaction.Member.User.ID
		} else if interaction.User != nil {
			userID = interaction.User.ID
		}

		if !a.config.IsUserAllowed(userID) || !a.config.IsChatAllowed(interaction.ChannelID) {
			w.WriteHeader(http.StatusOK)
			return
		}

		parts := strings.SplitN(interaction.Data.CustomID, ":", 2)
		buttonID := parts[0]
		data := ""
		if len(parts) > 1 {
			data = parts[1]
		}

		click := &ButtonClick{
			Platform:  PlatformDiscord,
			ChatID:    interaction.ChannelID,
			UserID:    userID,
			MessageID: interaction.Message.ID,
			ButtonID:  buttonID,
			Data:      data,
		}
		a.buttonHandler(click)
	}

	// Acknowledge
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"type": 6}) // Deferred update
}

type discordUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// Discord Gateway opcodes
const (
	discordOpDispatch        = 0
	discordOpHeartbeat       = 1
	discordOpIdentify        = 2
	discordOpResume          = 6
	discordOpReconnect       = 7
	discordOpInvalidSession  = 9
	discordOpHello           = 10
	discordOpHeartbeatAck    = 11
)

// Discord Gateway intents
const discordIntents = 33281 // GUILDS | GUILD_MESSAGES | MESSAGE_CONTENT | DIRECT_MESSAGES

type discordGatewayPayload struct {
	Op   int             `json:"op"`
	D    json.RawMessage `json:"d,omitempty"`
	S    *int64          `json:"s,omitempty"`
	T    string          `json:"t,omitempty"`
}

type discordHelloData struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type discordReadyData struct {
	SessionID string `json:"session_id"`
	ResumeURL string `json:"resume_gateway_url"`
}

type discordMessageCreate struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	GuildID   string `json:"guild_id,omitempty"`
	Author    struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Bot      bool   `json:"bot"`
	} `json:"author"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	Mentions  []struct {
		ID string `json:"id"`
	} `json:"mentions"`
	ReferencedMessage *struct {
		ID string `json:"id"`
	} `json:"referenced_message"`
}

type discordInteractionCreate struct {
	Type      int    `json:"type"`
	ChannelID string `json:"channel_id"`
	GuildID   string `json:"guild_id,omitempty"`
	Data      struct {
		CustomID string `json:"custom_id"`
	} `json:"data"`
	Message struct {
		ID string `json:"id"`
	} `json:"message"`
	Member *struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	} `json:"member"`
	User *struct {
		ID string `json:"id"`
	} `json:"user"`
}

func (a *DiscordAdapter) gatewayLoop() {
	defer a.wg.Done()

	backoff := time.Second
	maxBackoff := 5 * time.Minute

	for {
		select {
		case <-a.ctx.Done():
			return
		default:
		}

		err := a.connectGateway()
		if err != nil {
			select {
			case <-a.ctx.Done():
				return
			case <-time.After(backoff):
				backoff = min(backoff*2, maxBackoff)
				continue
			}
		}
		backoff = time.Second
	}
}

func (a *DiscordAdapter) connectGateway() error {
	// Get gateway URL
	gatewayURL := a.resumeURL
	if gatewayURL == "" {
		gatewayURL = "wss://gateway.discord.gg/?v=10&encoding=json"
	}

	conn, _, err := websocket.DefaultDialer.DialContext(a.ctx, gatewayURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}

	a.wsMu.Lock()
	a.wsConn = conn
	a.wsMu.Unlock()

	defer func() {
		a.wsMu.Lock()
		if a.wsConn == conn {
			a.wsConn.Close()
			a.wsConn = nil
		}
		a.wsMu.Unlock()
	}()

	// Read HELLO
	var hello discordGatewayPayload
	if err := conn.ReadJSON(&hello); err != nil {
		return fmt.Errorf("failed to read hello: %w", err)
	}
	if hello.Op != discordOpHello {
		return fmt.Errorf("expected HELLO, got op %d", hello.Op)
	}

	var helloData discordHelloData
	json.Unmarshal(hello.D, &helloData)

	// Send IDENTIFY or RESUME
	if a.sessionID != "" && a.resumeURL != "" {
		// Resume
		resume := map[string]interface{}{
			"token":      a.config.Token,
			"session_id": a.sessionID,
			"seq":        a.sequence.Load(),
		}
		resumeData, _ := json.Marshal(resume)
		if err := conn.WriteJSON(discordGatewayPayload{Op: discordOpResume, D: resumeData}); err != nil {
			return fmt.Errorf("failed to send resume: %w", err)
		}
	} else {
		// Identify
		identify := map[string]interface{}{
			"token":   a.config.Token,
			"intents": discordIntents,
			"properties": map[string]string{
				"os":      "linux",
				"browser": "gozen",
				"device":  "gozen",
			},
		}
		identifyData, _ := json.Marshal(identify)
		if err := conn.WriteJSON(discordGatewayPayload{Op: discordOpIdentify, D: identifyData}); err != nil {
			return fmt.Errorf("failed to send identify: %w", err)
		}
	}

	// Start heartbeat goroutine
	heartbeatCtx, heartbeatCancel := context.WithCancel(a.ctx)
	defer heartbeatCancel()

	go a.heartbeatLoop(heartbeatCtx, conn, time.Duration(helloData.HeartbeatInterval)*time.Millisecond)

	// Event loop
	for {
		select {
		case <-a.ctx.Done():
			return nil
		default:
		}

		var payload discordGatewayPayload
		if err := conn.ReadJSON(&payload); err != nil {
			return fmt.Errorf("read error: %w", err)
		}

		if payload.S != nil {
			a.sequence.Store(*payload.S)
		}

		switch payload.Op {
		case discordOpDispatch:
			a.handleDispatch(payload.T, payload.D)
		case discordOpReconnect:
			return fmt.Errorf("server requested reconnect")
		case discordOpInvalidSession:
			// Check if resumable
			var resumable bool
			json.Unmarshal(payload.D, &resumable)
			if !resumable {
				a.sessionID = ""
				a.resumeURL = ""
			}
			return fmt.Errorf("invalid session (resumable: %v)", resumable)
		case discordOpHeartbeatAck:
			// Heartbeat acknowledged
		case discordOpHeartbeat:
			// Server requesting heartbeat
			a.sendHeartbeat(conn)
		}
	}
}

func (a *DiscordAdapter) heartbeatLoop(ctx context.Context, conn *websocket.Conn, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.sendHeartbeat(conn)
		}
	}
}

func (a *DiscordAdapter) sendHeartbeat(conn *websocket.Conn) {
	seq := a.sequence.Load()
	seqData, _ := json.Marshal(seq)
	a.wsMu.Lock()
	conn.WriteJSON(discordGatewayPayload{Op: discordOpHeartbeat, D: seqData})
	a.wsMu.Unlock()
}

func (a *DiscordAdapter) handleDispatch(eventType string, data json.RawMessage) {
	switch eventType {
	case "READY":
		var ready discordReadyData
		json.Unmarshal(data, &ready)
		a.sessionID = ready.SessionID
		a.resumeURL = ready.ResumeURL
		log.Printf("[discord] connected, session_id=%s", a.sessionID)

	case "MESSAGE_CREATE":
		var msg discordMessageCreate
		if err := json.Unmarshal(data, &msg); err != nil {
			return
		}
		a.handleMessageCreate(&msg)

	case "INTERACTION_CREATE":
		var interaction discordInteractionCreate
		if err := json.Unmarshal(data, &interaction); err != nil {
			return
		}
		a.handleInteractionCreate(&interaction)
	}
}

func (a *DiscordAdapter) handleMessageCreate(msg *discordMessageCreate) {
	if a.msgHandler == nil {
		return
	}

	// Skip bot messages
	if msg.Author.Bot {
		return
	}

	// Check permissions
	if !a.config.IsUserAllowed(msg.Author.ID) || !a.config.IsChatAllowed(msg.ChannelID) {
		return
	}
	if msg.GuildID != "" && !a.config.IsGuildAllowed(msg.GuildID) {
		return
	}

	// Check if bot is mentioned
	isMention := false
	content := msg.Content
	for _, mention := range msg.Mentions {
		if mention.ID == a.botUserID {
			isMention = true
			content = strings.ReplaceAll(content, "<@"+a.botUserID+">", "")
			content = strings.ReplaceAll(content, "<@!"+a.botUserID+">", "")
			content = strings.TrimSpace(content)
			break
		}
	}

	isDirectMsg := msg.GuildID == ""

	botMsg := &Message{
		ID:          msg.ID,
		Platform:    PlatformDiscord,
		ChatID:      msg.ChannelID,
		UserID:      msg.Author.ID,
		UserName:    msg.Author.Username,
		Content:     content,
		IsMention:   isMention,
		IsDirectMsg: isDirectMsg,
		Metadata:    map[string]string{"guild_id": msg.GuildID},
	}

	if msg.ReferencedMessage != nil {
		botMsg.ReplyTo = msg.ReferencedMessage.ID
	}

	a.msgHandler(botMsg)
}

func (a *DiscordAdapter) handleInteractionCreate(interaction *discordInteractionCreate) {
	// Only handle message component interactions (buttons)
	if interaction.Type != 3 || a.buttonHandler == nil {
		return
	}

	userID := ""
	if interaction.Member != nil {
		userID = interaction.Member.User.ID
	} else if interaction.User != nil {
		userID = interaction.User.ID
	}

	if !a.config.IsUserAllowed(userID) || !a.config.IsChatAllowed(interaction.ChannelID) {
		return
	}
	if interaction.GuildID != "" && !a.config.IsGuildAllowed(interaction.GuildID) {
		return
	}

	parts := strings.SplitN(interaction.Data.CustomID, ":", 2)
	buttonID := parts[0]
	data := ""
	if len(parts) > 1 {
		data = parts[1]
	}

	click := &ButtonClick{
		Platform:  PlatformDiscord,
		ChatID:    interaction.ChannelID,
		UserID:    userID,
		MessageID: interaction.Message.ID,
		ButtonID:  buttonID,
		Data:      data,
	}
	a.buttonHandler(click)
}
