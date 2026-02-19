package adapters

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// LarkAdapter implements the Adapter interface for Lark/Feishu.
type LarkAdapter struct {
	config        *LarkConfig
	client        *http.Client
	accessToken   string
	tokenExpiry   time.Time
	tokenMu       sync.RWMutex
	botUserID     string
	msgHandler    func(*Message)
	buttonHandler func(*ButtonClick)
	ctx           context.Context
	cancel        context.CancelFunc
	server        *http.Server
}

// NewLarkAdapter creates a new Lark adapter.
func NewLarkAdapter(config *LarkConfig) *LarkAdapter {
	return &LarkAdapter{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (a *LarkAdapter) Platform() Platform {
	return PlatformLark
}

func (a *LarkAdapter) Start(ctx context.Context) error {
	a.ctx, a.cancel = context.WithCancel(ctx)

	// Get initial access token
	if err := a.refreshToken(); err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Start token refresh goroutine
	go a.tokenRefreshLoop()

	return nil
}

func (a *LarkAdapter) Stop() error {
	if a.cancel != nil {
		a.cancel()
	}
	if a.server != nil {
		return a.server.Shutdown(context.Background())
	}
	return nil
}

func (a *LarkAdapter) BotUserID() string {
	return a.botUserID
}

func (a *LarkAdapter) SetMessageHandler(handler func(*Message)) {
	a.msgHandler = handler
}

func (a *LarkAdapter) SetButtonHandler(handler func(*ButtonClick)) {
	a.buttonHandler = handler
}

func (a *LarkAdapter) SendMessage(chatID string, msg *OutgoingMessage) (string, error) {
	return a.sendMessage(chatID, "", msg)
}

func (a *LarkAdapter) SendReply(chatID, replyTo string, msg *OutgoingMessage) (string, error) {
	return a.sendMessage(chatID, replyTo, msg)
}

func (a *LarkAdapter) sendMessage(chatID, replyTo string, msg *OutgoingMessage) (string, error) {
	content := a.buildMessageContent(msg)

	payload := map[string]interface{}{
		"receive_id": chatID,
		"msg_type":   "interactive",
		"content":    content,
	}

	resp, err := a.apiCall("POST", "/im/v1/messages?receive_id_type=chat_id", payload)
	if err != nil {
		return "", err
	}

	var result struct {
		Data struct {
			MessageID string `json:"message_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", err
	}

	return result.Data.MessageID, nil
}

func (a *LarkAdapter) buildMessageContent(msg *OutgoingMessage) string {
	card := map[string]interface{}{
		"config": map[string]interface{}{
			"wide_screen_mode": true,
		},
		"elements": []interface{}{
			map[string]interface{}{
				"tag": "markdown",
				"content": msg.Text,
			},
		},
	}

	if len(msg.Buttons) > 0 {
		var actions []interface{}
		for _, btn := range msg.Buttons {
			btnType := "default"
			switch btn.Style {
			case "primary":
				btnType = "primary"
			case "danger":
				btnType = "danger"
			}
			actions = append(actions, map[string]interface{}{
				"tag": "button",
				"text": map[string]interface{}{
					"tag":     "plain_text",
					"content": btn.Label,
				},
				"type": btnType,
				"value": map[string]string{
					"id":   btn.ID,
					"data": btn.Data,
				},
			})
		}
		card["elements"] = append(card["elements"].([]interface{}), map[string]interface{}{
			"tag":     "action",
			"actions": actions,
		})
	}

	content, _ := json.Marshal(card)
	return string(content)
}

func (a *LarkAdapter) EditMessage(chatID, msgID string, msg *OutgoingMessage) error {
	content := a.buildMessageContent(msg)

	_, err := a.apiCall("PATCH", "/im/v1/messages/"+msgID, map[string]interface{}{
		"content": content,
	})
	return err
}

func (a *LarkAdapter) DeleteMessage(chatID, msgID string) error {
	_, err := a.apiCall("DELETE", "/im/v1/messages/"+msgID, nil)
	return err
}

func (a *LarkAdapter) refreshToken() error {
	payload := map[string]string{
		"app_id":     a.config.AppID,
		"app_secret": a.config.AppSecret,
	}

	data, _ := json.Marshal(payload)
	resp, err := a.client.Post(
		"https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal",
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Code              int    `json:"code"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.Code != 0 {
		return fmt.Errorf("failed to get token: code %d", result.Code)
	}

	a.tokenMu.Lock()
	a.accessToken = result.TenantAccessToken
	a.tokenExpiry = time.Now().Add(time.Duration(result.Expire-60) * time.Second)
	a.tokenMu.Unlock()

	return nil
}

func (a *LarkAdapter) tokenRefreshLoop() {
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-time.After(time.Minute):
			a.tokenMu.RLock()
			needRefresh := time.Now().After(a.tokenExpiry)
			a.tokenMu.RUnlock()

			if needRefresh {
				a.refreshToken()
			}
		}
	}
}

func (a *LarkAdapter) getToken() string {
	a.tokenMu.RLock()
	defer a.tokenMu.RUnlock()
	return a.accessToken
}

func (a *LarkAdapter) apiCall(method, path string, payload interface{}) ([]byte, error) {
	url := "https://open.feishu.cn/open-apis" + path

	var body io.Reader
	if payload != nil {
		data, _ := json.Marshal(payload)
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(a.ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.getToken())

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// HandleWebhook handles incoming webhook events from Lark.
func (a *LarkAdapter) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	var event struct {
		Challenge string `json:"challenge"`
		Type      string `json:"type"`
		Event     struct {
			Message struct {
				MessageID   string `json:"message_id"`
				ChatID      string `json:"chat_id"`
				Content     string `json:"content"`
				MessageType string `json:"message_type"`
			} `json:"message"`
			Sender struct {
				SenderID struct {
					UserID string `json:"user_id"`
				} `json:"sender_id"`
			} `json:"sender"`
			Action struct {
				Value map[string]string `json:"value"`
			} `json:"action"`
		} `json:"event"`
		Header struct {
			EventType string `json:"event_type"`
		} `json:"header"`
	}

	json.Unmarshal(body, &event)

	// URL verification
	if event.Challenge != "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"challenge": event.Challenge})
		return
	}

	// Handle message
	if event.Header.EventType == "im.message.receive_v1" && a.msgHandler != nil {
		var content struct {
			Text string `json:"text"`
		}
		json.Unmarshal([]byte(event.Event.Message.Content), &content)

		isMention := strings.Contains(content.Text, "@_user_")
		text := content.Text
		// Remove mention
		for strings.Contains(text, "@_user_") {
			start := strings.Index(text, "@_user_")
			end := strings.Index(text[start:], " ")
			if end == -1 {
				text = text[:start]
			} else {
				text = text[:start] + text[start+end:]
			}
		}

		msg := &Message{
			ID:        event.Event.Message.MessageID,
			Platform:  PlatformLark,
			ChatID:    event.Event.Message.ChatID,
			UserID:    event.Event.Sender.SenderID.UserID,
			Content:   strings.TrimSpace(text),
			IsMention: isMention,
		}
		a.msgHandler(msg)
	}

	// Handle button click
	if event.Header.EventType == "card.action.trigger" && a.buttonHandler != nil {
		click := &ButtonClick{
			Platform: PlatformLark,
			ButtonID: event.Event.Action.Value["id"],
			Data:     event.Event.Action.Value["data"],
		}
		a.buttonHandler(click)
	}

	w.WriteHeader(http.StatusOK)
}

// VerifySignature verifies the webhook signature.
func (a *LarkAdapter) VerifySignature(timestamp, nonce, signature string, body []byte) bool {
	content := timestamp + nonce + a.config.AppSecret + string(body)
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:]) == signature
}
