package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// FBMessengerAdapter implements the Adapter interface for Facebook Messenger.
type FBMessengerAdapter struct {
	config        *FBMessengerConfig
	client        *http.Client
	botUserID     string
	msgHandler    func(*Message)
	buttonHandler func(*ButtonClick)
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.RWMutex
}

// NewFBMessengerAdapter creates a new Facebook Messenger adapter.
func NewFBMessengerAdapter(config *FBMessengerConfig) *FBMessengerAdapter {
	return &FBMessengerAdapter{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (a *FBMessengerAdapter) Platform() Platform {
	return PlatformFBMessenger
}

func (a *FBMessengerAdapter) Start(ctx context.Context) error {
	a.ctx, a.cancel = context.WithCancel(ctx)
	return nil
}

func (a *FBMessengerAdapter) Stop() error {
	if a.cancel != nil {
		a.cancel()
	}
	return nil
}

func (a *FBMessengerAdapter) BotUserID() string {
	return a.botUserID
}

func (a *FBMessengerAdapter) SetMessageHandler(handler func(*Message)) {
	a.msgHandler = handler
}

func (a *FBMessengerAdapter) SetButtonHandler(handler func(*ButtonClick)) {
	a.buttonHandler = handler
}

func (a *FBMessengerAdapter) SendMessage(chatID string, msg *OutgoingMessage) (string, error) {
	return a.sendMessage(chatID, msg)
}

func (a *FBMessengerAdapter) SendReply(chatID, replyTo string, msg *OutgoingMessage) (string, error) {
	// FB Messenger doesn't have native reply, just send a message
	return a.sendMessage(chatID, msg)
}

func (a *FBMessengerAdapter) sendMessage(recipientID string, msg *OutgoingMessage) (string, error) {
	payload := map[string]interface{}{
		"recipient": map[string]string{
			"id": recipientID,
		},
		"messaging_type": "RESPONSE",
	}

	if len(msg.Buttons) > 0 {
		// Use button template
		var buttons []map[string]interface{}
		for _, btn := range msg.Buttons {
			buttons = append(buttons, map[string]interface{}{
				"type":    "postback",
				"title":   btn.Label,
				"payload": btn.ID + ":" + btn.Data,
			})
		}
		payload["message"] = map[string]interface{}{
			"attachment": map[string]interface{}{
				"type": "template",
				"payload": map[string]interface{}{
					"template_type": "button",
					"text":          msg.Text,
					"buttons":       buttons,
				},
			},
		}
	} else {
		payload["message"] = map[string]interface{}{
			"text": msg.Text,
		}
	}

	resp, err := a.apiCall("POST", "/me/messages", payload)
	if err != nil {
		return "", err
	}

	var result struct {
		MessageID string `json:"message_id"`
	}
	json.Unmarshal(resp, &result)

	return result.MessageID, nil
}

func (a *FBMessengerAdapter) EditMessage(chatID, msgID string, msg *OutgoingMessage) error {
	// FB Messenger doesn't support editing messages
	return fmt.Errorf("editing messages is not supported on Facebook Messenger")
}

func (a *FBMessengerAdapter) DeleteMessage(chatID, msgID string) error {
	// FB Messenger doesn't support deleting messages
	return fmt.Errorf("deleting messages is not supported on Facebook Messenger")
}

func (a *FBMessengerAdapter) apiCall(method, path string, payload interface{}) ([]byte, error) {
	url := "https://graph.facebook.com/v18.0" + path + "?access_token=" + a.config.PageToken

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

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// HandleWebhook handles incoming webhook events from Facebook.
func (a *FBMessengerAdapter) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Verification request
	if r.Method == "GET" {
		mode := r.URL.Query().Get("hub.mode")
		token := r.URL.Query().Get("hub.verify_token")
		challenge := r.URL.Query().Get("hub.challenge")

		if mode == "subscribe" && token == a.config.VerifyToken {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(challenge))
			return
		}
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Webhook event
	body, _ := io.ReadAll(r.Body)

	var event struct {
		Object string `json:"object"`
		Entry  []struct {
			Messaging []struct {
				Sender struct {
					ID string `json:"id"`
				} `json:"sender"`
				Recipient struct {
					ID string `json:"id"`
				} `json:"recipient"`
				Timestamp int64 `json:"timestamp"`
				Message   *struct {
					MID  string `json:"mid"`
					Text string `json:"text"`
				} `json:"message"`
				Postback *struct {
					Title   string `json:"title"`
					Payload string `json:"payload"`
				} `json:"postback"`
			} `json:"messaging"`
		} `json:"entry"`
	}

	if err := json.Unmarshal(body, &event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Process events
	for _, entry := range event.Entry {
		for _, messaging := range entry.Messaging {
			// Handle message
			if messaging.Message != nil && a.msgHandler != nil {
				if !a.config.IsUserAllowed(messaging.Sender.ID) {
					continue
				}

				msg := &Message{
					ID:          messaging.Message.MID,
					Platform:    PlatformFBMessenger,
					ChatID:      messaging.Sender.ID,
					UserID:      messaging.Sender.ID,
					Content:     messaging.Message.Text,
					Timestamp:   time.Unix(messaging.Timestamp/1000, 0),
					IsDirectMsg: true, // FB Messenger is always direct
				}
				a.msgHandler(msg)
			}

			// Handle postback (button click)
			if messaging.Postback != nil && a.buttonHandler != nil {
				if !a.config.IsUserAllowed(messaging.Sender.ID) {
					continue
				}

				// Parse payload: "buttonID:data"
				payload := messaging.Postback.Payload
				buttonID := payload
				data := ""
				if idx := len(payload) - len(payload); idx > 0 {
					// Find colon
					for i, c := range payload {
						if c == ':' {
							buttonID = payload[:i]
							data = payload[i+1:]
							break
						}
					}
				}

				click := &ButtonClick{
					Platform: PlatformFBMessenger,
					ChatID:   messaging.Sender.ID,
					UserID:   messaging.Sender.ID,
					ButtonID: buttonID,
					Data:     data,
				}
				a.buttonHandler(click)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}
