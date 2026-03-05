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
	"time"

	"github.com/gorilla/websocket"
)

// SlackAdapter implements the Adapter interface for Slack.
type SlackAdapter struct {
	config        *SlackConfig
	client        *http.Client
	botUserID     string
	msgHandler    func(*Message)
	buttonHandler func(*ButtonClick)
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	wsConn        *websocket.Conn
	wsMu          sync.Mutex
}

// NewSlackAdapter creates a new Slack adapter.
func NewSlackAdapter(config *SlackConfig) *SlackAdapter {
	return &SlackAdapter{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (a *SlackAdapter) Platform() Platform {
	return PlatformSlack
}

func (a *SlackAdapter) Start(ctx context.Context) error {
	a.ctx, a.cancel = context.WithCancel(ctx)

	// Get bot user ID via auth.test
	resp, err := a.apiCall("POST", "auth.test", nil)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	var authResp struct {
		OK     bool   `json:"ok"`
		UserID string `json:"user_id"`
		Error  string `json:"error"`
	}
	if err := json.Unmarshal(resp, &authResp); err != nil {
		return err
	}
	if !authResp.OK {
		return fmt.Errorf("slack auth failed: %s", authResp.Error)
	}
	a.botUserID = authResp.UserID

	// Start Socket Mode if app_token is configured
	if a.config.AppToken != "" {
		a.wg.Add(1)
		go a.socketModeLoop()
	} else {
		log.Printf("[slack] app_token not configured, Socket Mode disabled (bot can only send messages)")
	}

	return nil
}

func (a *SlackAdapter) Stop() error {
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

func (a *SlackAdapter) BotUserID() string {
	return a.botUserID
}

func (a *SlackAdapter) SetMessageHandler(handler func(*Message)) {
	a.msgHandler = handler
}

func (a *SlackAdapter) SetButtonHandler(handler func(*ButtonClick)) {
	a.buttonHandler = handler
}

func (a *SlackAdapter) SendMessage(chatID string, msg *OutgoingMessage) (string, error) {
	return a.sendMessage(chatID, "", msg)
}

func (a *SlackAdapter) SendReply(chatID, replyTo string, msg *OutgoingMessage) (string, error) {
	return a.sendMessage(chatID, replyTo, msg)
}

func (a *SlackAdapter) sendMessage(chatID, threadTS string, msg *OutgoingMessage) (string, error) {
	payload := map[string]interface{}{
		"channel": chatID,
		"text":    msg.Text,
	}

	if threadTS != "" {
		payload["thread_ts"] = threadTS
	}

	// Add buttons as blocks
	if len(msg.Buttons) > 0 {
		var elements []map[string]interface{}
		for _, btn := range msg.Buttons {
			style := "default"
			if btn.Style == "primary" {
				style = "primary"
			} else if btn.Style == "danger" {
				style = "danger"
			}
			element := map[string]interface{}{
				"type":      "button",
				"action_id": btn.ID + ":" + btn.Data,
				"text": map[string]interface{}{
					"type": "plain_text",
					"text": btn.Label,
				},
			}
			if style != "default" {
				element["style"] = style
			}
			elements = append(elements, element)
		}
		payload["blocks"] = []map[string]interface{}{
			{
				"type":     "section",
				"text":     map[string]string{"type": "mrkdwn", "text": msg.Text},
			},
			{
				"type":     "actions",
				"elements": elements,
			},
		}
	}

	resp, err := a.apiCall("POST", "chat.postMessage", payload)
	if err != nil {
		return "", err
	}

	var result struct {
		OK    bool   `json:"ok"`
		TS    string `json:"ts"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", err
	}
	if !result.OK {
		return "", fmt.Errorf("slack error: %s", result.Error)
	}

	return result.TS, nil
}

func (a *SlackAdapter) EditMessage(chatID, msgID string, msg *OutgoingMessage) error {
	payload := map[string]interface{}{
		"channel": chatID,
		"ts":      msgID,
		"text":    msg.Text,
	}

	if len(msg.Buttons) > 0 {
		var elements []map[string]interface{}
		for _, btn := range msg.Buttons {
			style := "default"
			if btn.Style == "primary" {
				style = "primary"
			} else if btn.Style == "danger" {
				style = "danger"
			}
			element := map[string]interface{}{
				"type":      "button",
				"action_id": btn.ID + ":" + btn.Data,
				"text": map[string]interface{}{
					"type": "plain_text",
					"text": btn.Label,
				},
			}
			if style != "default" {
				element["style"] = style
			}
			elements = append(elements, element)
		}
		payload["blocks"] = []map[string]interface{}{
			{
				"type": "section",
				"text": map[string]string{"type": "mrkdwn", "text": msg.Text},
			},
			{
				"type":     "actions",
				"elements": elements,
			},
		}
	} else {
		payload["blocks"] = []interface{}{}
	}

	resp, err := a.apiCall("POST", "chat.update", payload)
	if err != nil {
		return err
	}

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	json.Unmarshal(resp, &result)
	if !result.OK {
		return fmt.Errorf("slack error: %s", result.Error)
	}
	return nil
}

func (a *SlackAdapter) DeleteMessage(chatID, msgID string) error {
	payload := map[string]interface{}{
		"channel": chatID,
		"ts":      msgID,
	}

	resp, err := a.apiCall("POST", "chat.delete", payload)
	if err != nil {
		return err
	}

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	json.Unmarshal(resp, &result)
	if !result.OK {
		return fmt.Errorf("slack error: %s", result.Error)
	}
	return nil
}

func (a *SlackAdapter) apiCall(method, endpoint string, payload interface{}) ([]byte, error) {
	url := "https://slack.com/api/" + endpoint

	var body io.Reader
	if payload != nil {
		data, _ := json.Marshal(payload)
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(a.ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+a.config.BotToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// HandleEvent handles incoming Slack events (for Events API).
func (a *SlackAdapter) HandleEvent(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	// Check for URL verification challenge
	var challenge struct {
		Type      string `json:"type"`
		Challenge string `json:"challenge"`
	}
	if err := json.Unmarshal(body, &challenge); err == nil && challenge.Type == "url_verification" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(challenge.Challenge))
		return
	}

	// Parse event
	var event struct {
		Type  string `json:"type"`
		Event struct {
			Type      string `json:"type"`
			User      string `json:"user"`
			Text      string `json:"text"`
			Channel   string `json:"channel"`
			TS        string `json:"ts"`
			ThreadTS  string `json:"thread_ts"`
			ChannelType string `json:"channel_type"`
		} `json:"event"`
	}

	if err := json.Unmarshal(body, &event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Handle message events
	if event.Type == "event_callback" {
		ev := event.Event

		// Skip bot's own messages
		if ev.User == a.botUserID {
			w.WriteHeader(http.StatusOK)
			return
		}

		if !a.config.IsUserAllowed(ev.User) || !a.config.IsChatAllowed(ev.Channel) {
			w.WriteHeader(http.StatusOK)
			return
		}

		if ev.Type == "message" && a.msgHandler != nil {
			msg := &Message{
				ID:          ev.TS,
				Platform:    PlatformSlack,
				ChatID:      ev.Channel,
				UserID:      ev.User,
				Content:     ev.Text,
				ThreadID:    ev.ThreadTS,
				IsDirectMsg: ev.ChannelType == "im",
			}
			a.msgHandler(msg)
		}

		if ev.Type == "app_mention" && a.msgHandler != nil {
			content := strings.ReplaceAll(ev.Text, "<@"+a.botUserID+">", "")
			content = strings.TrimSpace(content)

			msg := &Message{
				ID:        ev.TS,
				Platform:  PlatformSlack,
				ChatID:    ev.Channel,
				UserID:    ev.User,
				Content:   content,
				ThreadID:  ev.ThreadTS,
				IsMention: true,
			}
			a.msgHandler(msg)
		}
	}

	w.WriteHeader(http.StatusOK)
}

// HandleInteraction handles incoming Slack interactions (button clicks).
func (a *SlackAdapter) HandleInteraction(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	payloadStr := r.FormValue("payload")

	var payload struct {
		Type    string `json:"type"`
		User    struct {
			ID string `json:"id"`
		} `json:"user"`
		Channel struct {
			ID string `json:"id"`
		} `json:"channel"`
		Message struct {
			TS string `json:"ts"`
		} `json:"message"`
		Actions []struct {
			ActionID string `json:"action_id"`
			Value    string `json:"value"`
		} `json:"actions"`
	}

	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if payload.Type == "block_actions" && a.buttonHandler != nil {
		if !a.config.IsUserAllowed(payload.User.ID) || !a.config.IsChatAllowed(payload.Channel.ID) {
			w.WriteHeader(http.StatusOK)
			return
		}

		for _, action := range payload.Actions {
			parts := strings.SplitN(action.ActionID, ":", 2)
			buttonID := parts[0]
			data := ""
			if len(parts) > 1 {
				data = parts[1]
			}

			click := &ButtonClick{
				Platform:  PlatformSlack,
				ChatID:    payload.Channel.ID,
				UserID:    payload.User.ID,
				MessageID: payload.Message.TS,
				ButtonID:  buttonID,
				Data:      data,
			}
			a.buttonHandler(click)
		}
	}

	w.WriteHeader(http.StatusOK)
}

// Socket Mode types
type slackSocketEnvelope struct {
	Type       string          `json:"type"`
	EnvelopeID string          `json:"envelope_id"`
	Payload    json.RawMessage `json:"payload"`
}

type slackEventsAPIPayload struct {
	Type  string `json:"type"`
	Event struct {
		Type        string `json:"type"`
		User        string `json:"user"`
		Text        string `json:"text"`
		Channel     string `json:"channel"`
		TS          string `json:"ts"`
		ThreadTS    string `json:"thread_ts"`
		ChannelType string `json:"channel_type"`
	} `json:"event"`
}

type slackInteractivePayload struct {
	Type    string `json:"type"`
	User    struct {
		ID string `json:"id"`
	} `json:"user"`
	Channel struct {
		ID string `json:"id"`
	} `json:"channel"`
	Message struct {
		TS string `json:"ts"`
	} `json:"message"`
	Actions []struct {
		ActionID string `json:"action_id"`
	} `json:"actions"`
}

func (a *SlackAdapter) socketModeLoop() {
	defer a.wg.Done()

	backoff := time.Second
	maxBackoff := 5 * time.Minute

	for {
		select {
		case <-a.ctx.Done():
			return
		default:
		}

		err := a.connectSocketMode()
		if err != nil {
			log.Printf("[slack] socket mode error: %v", err)
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

func (a *SlackAdapter) connectSocketMode() error {
	// Get WebSocket URL via apps.connections.open
	req, err := http.NewRequestWithContext(a.ctx, "POST", "https://slack.com/api/apps.connections.open", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+a.config.AppToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var connResp struct {
		OK    bool   `json:"ok"`
		URL   string `json:"url"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &connResp); err != nil {
		return err
	}
	if !connResp.OK {
		return fmt.Errorf("failed to get socket URL: %s", connResp.Error)
	}

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.DialContext(a.ctx, connResp.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
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

	log.Printf("[slack] socket mode connected")

	// Event loop
	for {
		select {
		case <-a.ctx.Done():
			return nil
		default:
		}

		var envelope slackSocketEnvelope
		if err := conn.ReadJSON(&envelope); err != nil {
			return fmt.Errorf("read error: %w", err)
		}

		// ACK immediately
		if envelope.EnvelopeID != "" {
			a.wsMu.Lock()
			conn.WriteJSON(map[string]string{"envelope_id": envelope.EnvelopeID})
			a.wsMu.Unlock()
		}

		switch envelope.Type {
		case "events_api":
			a.handleEventsAPI(envelope.Payload)
		case "interactive":
			a.handleInteractive(envelope.Payload)
		case "hello":
			// Connection established
		case "disconnect":
			return fmt.Errorf("server requested disconnect")
		}
	}
}

func (a *SlackAdapter) handleEventsAPI(data json.RawMessage) {
	var payload slackEventsAPIPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return
	}

	ev := payload.Event

	// Skip bot's own messages
	if ev.User == a.botUserID {
		return
	}

	if !a.config.IsUserAllowed(ev.User) || !a.config.IsChatAllowed(ev.Channel) {
		return
	}

	if ev.Type == "message" && a.msgHandler != nil {
		msg := &Message{
			ID:          ev.TS,
			Platform:    PlatformSlack,
			ChatID:      ev.Channel,
			UserID:      ev.User,
			Content:     ev.Text,
			ThreadID:    ev.ThreadTS,
			IsDirectMsg: ev.ChannelType == "im",
		}
		a.msgHandler(msg)
	}

	if ev.Type == "app_mention" && a.msgHandler != nil {
		content := strings.ReplaceAll(ev.Text, "<@"+a.botUserID+">", "")
		content = strings.TrimSpace(content)

		msg := &Message{
			ID:        ev.TS,
			Platform:  PlatformSlack,
			ChatID:    ev.Channel,
			UserID:    ev.User,
			Content:   content,
			ThreadID:  ev.ThreadTS,
			IsMention: true,
		}
		a.msgHandler(msg)
	}
}

func (a *SlackAdapter) handleInteractive(data json.RawMessage) {
	var payload slackInteractivePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return
	}

	if payload.Type != "block_actions" || a.buttonHandler == nil {
		return
	}

	if !a.config.IsUserAllowed(payload.User.ID) || !a.config.IsChatAllowed(payload.Channel.ID) {
		return
	}

	for _, action := range payload.Actions {
		parts := strings.SplitN(action.ActionID, ":", 2)
		buttonID := parts[0]
		data := ""
		if len(parts) > 1 {
			data = parts[1]
		}

		click := &ButtonClick{
			Platform:  PlatformSlack,
			ChatID:    payload.Channel.ID,
			UserID:    payload.User.ID,
			MessageID: payload.Message.TS,
			ButtonID:  buttonID,
			Data:      data,
		}
		a.buttonHandler(click)
	}
}
