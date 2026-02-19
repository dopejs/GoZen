package adapters

import (
	"testing"
	"time"
)

func TestPlatformConstants(t *testing.T) {
	platforms := []Platform{
		PlatformTelegram,
		PlatformDiscord,
		PlatformSlack,
		PlatformLark,
		PlatformFBMessenger,
	}

	for _, p := range platforms {
		if p == "" {
			t.Error("Platform constant should not be empty")
		}
	}
}

func TestMessage(t *testing.T) {
	msg := &Message{
		ID:          "msg-1",
		Platform:    PlatformTelegram,
		ChatID:      "chat-1",
		UserID:      "user-1",
		UserName:    "John",
		Content:     "Hello",
		ReplyTo:     "msg-0",
		ThreadID:    "thread-1",
		Timestamp:   time.Now(),
		IsMention:   true,
		IsDirectMsg: false,
		Metadata:    map[string]string{"key": "value"},
	}

	if msg.ID != "msg-1" {
		t.Errorf("expected ID 'msg-1', got '%s'", msg.ID)
	}
	if msg.Platform != PlatformTelegram {
		t.Errorf("expected Platform Telegram, got %v", msg.Platform)
	}
}

func TestOutgoingMessage(t *testing.T) {
	msg := &OutgoingMessage{
		Text:   "Hello World",
		Format: "markdown",
		Buttons: []Button{
			{ID: "btn-1", Label: "OK", Style: "primary", Data: "ok"},
		},
	}

	if msg.Text != "Hello World" {
		t.Errorf("expected Text 'Hello World', got '%s'", msg.Text)
	}
	if len(msg.Buttons) != 1 {
		t.Errorf("expected 1 button, got %d", len(msg.Buttons))
	}
}

func TestButton(t *testing.T) {
	btn := Button{
		ID:    "btn-1",
		Label: "Click Me",
		Style: "primary",
		Data:  "action:click",
	}

	if btn.ID != "btn-1" {
		t.Errorf("expected ID 'btn-1', got '%s'", btn.ID)
	}
	if btn.Style != "primary" {
		t.Errorf("expected Style 'primary', got '%s'", btn.Style)
	}
}

func TestButtonClick(t *testing.T) {
	click := &ButtonClick{
		Platform:  PlatformDiscord,
		ChatID:    "channel-1",
		UserID:    "user-1",
		MessageID: "msg-1",
		ButtonID:  "btn-1",
		Data:      "action:approve",
	}

	if click.Platform != PlatformDiscord {
		t.Errorf("expected Platform Discord, got %v", click.Platform)
	}
	if click.ButtonID != "btn-1" {
		t.Errorf("expected ButtonID 'btn-1', got '%s'", click.ButtonID)
	}
}

func TestAdapterConfig(t *testing.T) {
	cfg := AdapterConfig{
		Enabled:         true,
		AllowedUsers:    []string{"user1", "user2"},
		AllowedChannels: []string{"channel1"},
		AllowedChats:    []string{"chat1"},
	}

	if !cfg.Enabled {
		t.Error("expected Enabled to be true")
	}
	if len(cfg.AllowedUsers) != 2 {
		t.Errorf("expected 2 allowed users, got %d", len(cfg.AllowedUsers))
	}
}

func TestAdapterConfig_IsUserAllowed(t *testing.T) {
	tests := []struct {
		name         string
		allowedUsers []string
		userID       string
		expected     bool
	}{
		{"empty list allows all", nil, "anyone", true},
		{"empty list allows all 2", []string{}, "anyone", true},
		{"allowed user", []string{"user1", "user2"}, "user1", true},
		{"not allowed user", []string{"user1", "user2"}, "user3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &AdapterConfig{AllowedUsers: tt.allowedUsers}
			result := cfg.IsUserAllowed(tt.userID)
			if result != tt.expected {
				t.Errorf("IsUserAllowed(%q) = %v, want %v", tt.userID, result, tt.expected)
			}
		})
	}
}

func TestAdapterConfig_IsChatAllowed(t *testing.T) {
	tests := []struct {
		name            string
		allowedChannels []string
		allowedChats    []string
		chatID          string
		expected        bool
	}{
		{"empty lists allow all", nil, nil, "any", true},
		{"allowed channel", []string{"channel1"}, nil, "channel1", true},
		{"allowed chat", nil, []string{"chat1"}, "chat1", true},
		{"not allowed", []string{"channel1"}, []string{"chat1"}, "other", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &AdapterConfig{
				AllowedChannels: tt.allowedChannels,
				AllowedChats:    tt.allowedChats,
			}
			result := cfg.IsChatAllowed(tt.chatID)
			if result != tt.expected {
				t.Errorf("IsChatAllowed(%q) = %v, want %v", tt.chatID, result, tt.expected)
			}
		})
	}
}

func TestTelegramConfig(t *testing.T) {
	cfg := &TelegramConfig{
		AdapterConfig: AdapterConfig{
			Enabled:      true,
			AllowedUsers: []string{"123456"},
		},
		Token: "bot123:ABC",
	}

	if !cfg.Enabled {
		t.Error("expected Enabled to be true")
	}
	if cfg.Token != "bot123:ABC" {
		t.Errorf("expected Token 'bot123:ABC', got '%s'", cfg.Token)
	}
}

func TestDiscordConfig(t *testing.T) {
	cfg := &DiscordConfig{
		AdapterConfig: AdapterConfig{
			Enabled: true,
		},
		Token:         "discord-token",
		AllowedGuilds: []string{"guild1", "guild2"},
	}

	if cfg.Token != "discord-token" {
		t.Errorf("expected Token 'discord-token', got '%s'", cfg.Token)
	}
	if len(cfg.AllowedGuilds) != 2 {
		t.Errorf("expected 2 allowed guilds, got %d", len(cfg.AllowedGuilds))
	}
}

func TestSlackConfig(t *testing.T) {
	cfg := &SlackConfig{
		AdapterConfig: AdapterConfig{
			Enabled: true,
		},
		BotToken: "xoxb-token",
		AppToken: "xapp-token",
	}

	if cfg.BotToken != "xoxb-token" {
		t.Errorf("expected BotToken 'xoxb-token', got '%s'", cfg.BotToken)
	}
	if cfg.AppToken != "xapp-token" {
		t.Errorf("expected AppToken 'xapp-token', got '%s'", cfg.AppToken)
	}
}

func TestLarkConfig(t *testing.T) {
	cfg := &LarkConfig{
		AdapterConfig: AdapterConfig{
			Enabled: true,
		},
		AppID:     "cli_xxx",
		AppSecret: "secret",
	}

	if cfg.AppID != "cli_xxx" {
		t.Errorf("expected AppID 'cli_xxx', got '%s'", cfg.AppID)
	}
	if cfg.AppSecret != "secret" {
		t.Errorf("expected AppSecret 'secret', got '%s'", cfg.AppSecret)
	}
}

func TestFBMessengerConfig(t *testing.T) {
	cfg := &FBMessengerConfig{
		AdapterConfig: AdapterConfig{
			Enabled: true,
		},
		PageToken:   "page-token",
		VerifyToken: "verify-token",
		AppSecret:   "app-secret",
	}

	if cfg.PageToken != "page-token" {
		t.Errorf("expected PageToken 'page-token', got '%s'", cfg.PageToken)
	}
	if cfg.VerifyToken != "verify-token" {
		t.Errorf("expected VerifyToken 'verify-token', got '%s'", cfg.VerifyToken)
	}
}
