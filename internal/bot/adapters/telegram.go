package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TelegramAdapter implements the Adapter interface for Telegram.
type TelegramAdapter struct {
	config        *TelegramConfig
	client        *http.Client
	apiBase       string
	botUserID     string
	botUsername   string
	msgHandler    func(*Message)
	buttonHandler func(*ButtonClick)
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	lastUpdateID  int64
}

// NewTelegramAdapter creates a new Telegram adapter.
func NewTelegramAdapter(config *TelegramConfig) *TelegramAdapter {
	return &TelegramAdapter{
		config:  config,
		client:  &http.Client{Timeout: 60 * time.Second},
		apiBase: "https://api.telegram.org/bot" + config.Token,
	}
}

func (a *TelegramAdapter) Platform() Platform {
	return PlatformTelegram
}

func (a *TelegramAdapter) Start(ctx context.Context) error {
	// Get bot info
	info, err := a.getMe()
	if err != nil {
		return fmt.Errorf("failed to get bot info: %w", err)
	}
	a.botUserID = strconv.FormatInt(info.ID, 10)
	a.botUsername = info.Username

	a.ctx, a.cancel = context.WithCancel(ctx)
	a.wg.Add(1)
	go a.pollUpdates()

	return nil
}

func (a *TelegramAdapter) Stop() error {
	if a.cancel != nil {
		a.cancel()
	}
	a.wg.Wait()
	return nil
}

func (a *TelegramAdapter) BotUserID() string {
	return a.botUserID
}

func (a *TelegramAdapter) SetMessageHandler(handler func(*Message)) {
	a.msgHandler = handler
}

func (a *TelegramAdapter) SetButtonHandler(handler func(*ButtonClick)) {
	a.buttonHandler = handler
}

func (a *TelegramAdapter) SendMessage(chatID string, msg *OutgoingMessage) (string, error) {
	return a.sendMessage(chatID, "", msg)
}

func (a *TelegramAdapter) SendReply(chatID, replyTo string, msg *OutgoingMessage) (string, error) {
	return a.sendMessage(chatID, replyTo, msg)
}

func (a *TelegramAdapter) sendMessage(chatID, replyTo string, msg *OutgoingMessage) (string, error) {
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    msg.Text,
	}

	if msg.Format == "markdown" {
		payload["parse_mode"] = "Markdown"
	}

	if replyTo != "" {
		if msgID, err := strconv.ParseInt(replyTo, 10, 64); err == nil {
			payload["reply_to_message_id"] = msgID
		}
	}

	// Add inline keyboard if buttons present
	if len(msg.Buttons) > 0 {
		var buttons [][]map[string]string
		var row []map[string]string
		for _, btn := range msg.Buttons {
			row = append(row, map[string]string{
				"text":          btn.Label,
				"callback_data": btn.ID + ":" + btn.Data,
			})
			if len(row) >= 2 {
				buttons = append(buttons, row)
				row = nil
			}
		}
		if len(row) > 0 {
			buttons = append(buttons, row)
		}
		payload["reply_markup"] = map[string]interface{}{
			"inline_keyboard": buttons,
		}
	}

	resp, err := a.apiCall("sendMessage", payload)
	if err != nil {
		return "", err
	}

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			MessageID int64 `json:"message_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", err
	}

	return strconv.FormatInt(result.Result.MessageID, 10), nil
}

func (a *TelegramAdapter) EditMessage(chatID, msgID string, msg *OutgoingMessage) error {
	messageID, _ := strconv.ParseInt(msgID, 10, 64)
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       msg.Text,
	}

	if msg.Format == "markdown" {
		payload["parse_mode"] = "Markdown"
	}

	// Update inline keyboard
	if len(msg.Buttons) > 0 {
		var buttons [][]map[string]string
		var row []map[string]string
		for _, btn := range msg.Buttons {
			row = append(row, map[string]string{
				"text":          btn.Label,
				"callback_data": btn.ID + ":" + btn.Data,
			})
			if len(row) >= 2 {
				buttons = append(buttons, row)
				row = nil
			}
		}
		if len(row) > 0 {
			buttons = append(buttons, row)
		}
		payload["reply_markup"] = map[string]interface{}{
			"inline_keyboard": buttons,
		}
	} else {
		// Remove keyboard
		payload["reply_markup"] = map[string]interface{}{
			"inline_keyboard": [][]map[string]string{},
		}
	}

	_, err := a.apiCall("editMessageText", payload)
	return err
}

func (a *TelegramAdapter) DeleteMessage(chatID, msgID string) error {
	messageID, _ := strconv.ParseInt(msgID, 10, 64)
	_, err := a.apiCall("deleteMessage", map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
	})
	return err
}

func (a *TelegramAdapter) pollUpdates() {
	defer a.wg.Done()

	for {
		select {
		case <-a.ctx.Done():
			return
		default:
		}

		updates, err := a.getUpdates(a.lastUpdateID + 1)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		for _, update := range updates {
			a.lastUpdateID = update.UpdateID
			a.handleUpdate(&update)
		}
	}
}

func (a *TelegramAdapter) handleUpdate(update *tgUpdate) {
	// Handle callback query (button click)
	if update.CallbackQuery != nil {
		a.handleCallbackQuery(update.CallbackQuery)
		return
	}

	// Handle message
	if update.Message != nil {
		a.handleMessage(update.Message)
	}
}

func (a *TelegramAdapter) handleMessage(msg *tgMessage) {
	if a.msgHandler == nil {
		return
	}

	chatID := strconv.FormatInt(msg.Chat.ID, 10)
	userID := strconv.FormatInt(msg.From.ID, 10)

	// Check permissions
	if !a.config.IsUserAllowed(userID) || !a.config.IsChatAllowed(chatID) {
		return
	}

	// Check if bot is mentioned
	isMention := false
	content := msg.Text

	// Check for @mention
	if a.botUsername != "" && strings.Contains(content, "@"+a.botUsername) {
		isMention = true
		content = strings.ReplaceAll(content, "@"+a.botUsername, "")
		content = strings.TrimSpace(content)
	}

	// Check entities for mention
	for _, entity := range msg.Entities {
		if entity.Type == "mention" && entity.Offset == 0 {
			isMention = true
		}
	}

	isDirectMsg := msg.Chat.Type == "private"

	botMsg := &Message{
		ID:          strconv.FormatInt(msg.MessageID, 10),
		Platform:    PlatformTelegram,
		ChatID:      chatID,
		UserID:      userID,
		UserName:    msg.From.Username,
		Content:     content,
		Timestamp:   time.Unix(int64(msg.Date), 0),
		IsMention:   isMention,
		IsDirectMsg: isDirectMsg,
	}

	if msg.ReplyToMessage != nil {
		botMsg.ReplyTo = strconv.FormatInt(msg.ReplyToMessage.MessageID, 10)
	}

	a.msgHandler(botMsg)
}

func (a *TelegramAdapter) handleCallbackQuery(query *tgCallbackQuery) {
	// Answer callback to remove loading state
	a.apiCall("answerCallbackQuery", map[string]interface{}{
		"callback_query_id": query.ID,
	})

	if a.buttonHandler == nil {
		return
	}

	chatID := strconv.FormatInt(query.Message.Chat.ID, 10)
	userID := strconv.FormatInt(query.From.ID, 10)

	// Check permissions
	if !a.config.IsUserAllowed(userID) || !a.config.IsChatAllowed(chatID) {
		return
	}

	// Parse callback data: "buttonID:data"
	parts := strings.SplitN(query.Data, ":", 2)
	buttonID := parts[0]
	data := ""
	if len(parts) > 1 {
		data = parts[1]
	}

	click := &ButtonClick{
		Platform:  PlatformTelegram,
		ChatID:    chatID,
		UserID:    userID,
		MessageID: strconv.FormatInt(query.Message.MessageID, 10),
		ButtonID:  buttonID,
		Data:      data,
	}

	a.buttonHandler(click)
}

func (a *TelegramAdapter) getMe() (*tgUser, error) {
	resp, err := a.apiCall("getMe", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		OK     bool   `json:"ok"`
		Result tgUser `json:"result"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result.Result, nil
}

func (a *TelegramAdapter) getUpdates(offset int64) ([]tgUpdate, error) {
	resp, err := a.apiCall("getUpdates", map[string]interface{}{
		"offset":  offset,
		"timeout": 30,
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		OK     bool       `json:"ok"`
		Result []tgUpdate `json:"result"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result.Result, nil
}

func (a *TelegramAdapter) apiCall(method string, payload interface{}) ([]byte, error) {
	url := a.apiBase + "/" + method

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(a.ctx, "POST", url, body)
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

// Telegram API types
type tgUpdate struct {
	UpdateID      int64            `json:"update_id"`
	Message       *tgMessage       `json:"message"`
	CallbackQuery *tgCallbackQuery `json:"callback_query"`
}

type tgMessage struct {
	MessageID      int64       `json:"message_id"`
	From           *tgUser     `json:"from"`
	Chat           *tgChat     `json:"chat"`
	Date           int         `json:"date"`
	Text           string      `json:"text"`
	Entities       []tgEntity  `json:"entities"`
	ReplyToMessage *tgMessage  `json:"reply_to_message"`
}

type tgUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

type tgChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type tgEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

type tgCallbackQuery struct {
	ID      string     `json:"id"`
	From    *tgUser    `json:"from"`
	Message *tgMessage `json:"message"`
	Data    string     `json:"data"`
}
