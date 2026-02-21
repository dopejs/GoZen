package adapters

import (
	"context"
	"time"
)

// Platform represents a chat platform.
type Platform string

const (
	PlatformTelegram    Platform = "telegram"
	PlatformDiscord     Platform = "discord"
	PlatformSlack       Platform = "slack"
	PlatformLark        Platform = "lark"
	PlatformFBMessenger Platform = "fbmessenger"
)

// Message represents an incoming message from any platform.
type Message struct {
	ID          string            `json:"id"`
	Platform    Platform          `json:"platform"`
	ChatID      string            `json:"chat_id"`
	UserID      string            `json:"user_id"`
	UserName    string            `json:"user_name"`
	Content     string            `json:"content"`
	ReplyTo     string            `json:"reply_to,omitempty"`
	ThreadID    string            `json:"thread_id,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
	IsMention   bool              `json:"is_mention"`
	IsDirectMsg bool              `json:"is_direct_msg"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// OutgoingMessage represents a message to be sent.
type OutgoingMessage struct {
	Text    string   `json:"text"`
	Format  string   `json:"format"` // plain, markdown
	Buttons []Button `json:"buttons,omitempty"`
}

// Button represents an interactive button.
type Button struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Style string `json:"style"` // primary, secondary, danger
	Data  string `json:"data"`
}

// ButtonClick represents a button click event.
type ButtonClick struct {
	Platform  Platform `json:"platform"`
	ChatID    string   `json:"chat_id"`
	UserID    string   `json:"user_id"`
	MessageID string   `json:"message_id"`
	ButtonID  string   `json:"button_id"`
	Data      string   `json:"data"`
}

// Adapter is the interface for chat platform adapters.
type Adapter interface {
	// Platform returns the platform identifier.
	Platform() Platform

	// Start starts the adapter and begins listening for messages.
	Start(ctx context.Context) error

	// Stop stops the adapter.
	Stop() error

	// SendMessage sends a message to a chat.
	SendMessage(chatID string, msg *OutgoingMessage) (string, error)

	// SendReply sends a reply to a specific message.
	SendReply(chatID, replyTo string, msg *OutgoingMessage) (string, error)

	// EditMessage edits an existing message.
	EditMessage(chatID, msgID string, msg *OutgoingMessage) error

	// DeleteMessage deletes a message.
	DeleteMessage(chatID, msgID string) error

	// SetMessageHandler sets the handler for incoming messages.
	SetMessageHandler(handler func(*Message))

	// SetButtonHandler sets the handler for button clicks.
	SetButtonHandler(handler func(*ButtonClick))

	// BotUserID returns the bot's user ID on this platform.
	BotUserID() string
}

// AdapterConfig is the base configuration for adapters.
type AdapterConfig struct {
	Enabled         bool     `json:"enabled"`
	AllowedUsers    []string `json:"allowed_users,omitempty"`
	AllowedChannels []string `json:"allowed_channels,omitempty"`
	AllowedChats    []string `json:"allowed_chats,omitempty"`
}

// IsUserAllowed checks if a user is in the allowed list.
func (c *AdapterConfig) IsUserAllowed(userID string) bool {
	if len(c.AllowedUsers) == 0 {
		return true
	}
	for _, u := range c.AllowedUsers {
		if u == userID {
			return true
		}
	}
	return false
}

// IsChatAllowed checks if a chat/channel is in the allowed list.
func (c *AdapterConfig) IsChatAllowed(chatID string) bool {
	// Check both channels and chats
	allowed := append(c.AllowedChannels, c.AllowedChats...)
	if len(allowed) == 0 {
		return true
	}
	for _, ch := range allowed {
		if ch == chatID {
			return true
		}
	}
	return false
}

// TelegramConfig is the configuration for Telegram adapter.
type TelegramConfig struct {
	AdapterConfig
	Token string `json:"token"`
}

// DiscordConfig is the configuration for Discord adapter.
type DiscordConfig struct {
	AdapterConfig
	Token         string   `json:"token"`
	AllowedGuilds []string `json:"allowed_guilds,omitempty"`
}

// SlackConfig is the configuration for Slack adapter.
type SlackConfig struct {
	AdapterConfig
	BotToken string `json:"bot_token"`
	AppToken string `json:"app_token"`
}

// LarkConfig is the configuration for Lark/Feishu adapter.
type LarkConfig struct {
	AdapterConfig
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

// FBMessengerConfig is the configuration for Facebook Messenger adapter.
type FBMessengerConfig struct {
	AdapterConfig
	PageToken   string `json:"page_token"`
	VerifyToken string `json:"verify_token"`
	AppSecret   string `json:"app_secret,omitempty"`
}
