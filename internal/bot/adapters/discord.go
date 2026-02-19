package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// DiscordAdapter implements the Adapter interface for Discord.
type DiscordAdapter struct {
	config        *DiscordConfig
	client        *http.Client
	botUserID     string
	msgHandler    func(*Message)
	buttonHandler func(*ButtonClick)
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	gatewayURL    string
	sessionID     string
	sequence      int64
	heartbeatInt  int
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

	// Note: Full Discord bot requires WebSocket gateway connection
	// This is a simplified HTTP-only implementation for sending messages
	// For receiving messages, you would need to set up webhooks or use the gateway

	return nil
}

func (a *DiscordAdapter) Stop() error {
	if a.cancel != nil {
		a.cancel()
	}
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
