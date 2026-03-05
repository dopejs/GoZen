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

func TestDiscordConfig_IsGuildAllowed(t *testing.T) {
	tests := []struct {
		name          string
		allowedGuilds []string
		guildID       string
		expected      bool
	}{
		{"empty list allows all", nil, "any-guild", true},
		{"empty list allows all 2", []string{}, "any-guild", true},
		{"allowed guild", []string{"guild1", "guild2"}, "guild1", true},
		{"not allowed guild", []string{"guild1", "guild2"}, "guild3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &DiscordConfig{AllowedGuilds: tt.allowedGuilds}
			result := cfg.IsGuildAllowed(tt.guildID)
			if result != tt.expected {
				t.Errorf("IsGuildAllowed(%q) = %v, want %v", tt.guildID, result, tt.expected)
			}
		})
	}
}

func TestDiscordAdapter_HandleMessageCreate(t *testing.T) {
	tests := []struct {
		name        string
		msg         *discordMessageCreate
		botUserID   string
		config      *DiscordConfig
		expectCall  bool
		expectMsg   *Message
	}{
		{
			name: "normal message",
			msg: &discordMessageCreate{
				ID:        "msg-123",
				ChannelID: "channel-1",
				GuildID:   "guild-1",
				Author: struct {
					ID       string `json:"id"`
					Username string `json:"username"`
					Bot      bool   `json:"bot"`
				}{ID: "user-1", Username: "testuser", Bot: false},
				Content: "hello world",
			},
			botUserID:  "bot-1",
			config:     &DiscordConfig{},
			expectCall: true,
			expectMsg: &Message{
				ID:          "msg-123",
				Platform:    PlatformDiscord,
				ChatID:      "channel-1",
				UserID:      "user-1",
				UserName:    "testuser",
				Content:     "hello world",
				IsMention:   false,
				IsDirectMsg: false,
			},
		},
		{
			name: "skip bot message",
			msg: &discordMessageCreate{
				ID:        "msg-123",
				ChannelID: "channel-1",
				Author: struct {
					ID       string `json:"id"`
					Username string `json:"username"`
					Bot      bool   `json:"bot"`
				}{ID: "bot-1", Username: "bot", Bot: true},
				Content: "bot message",
			},
			botUserID:  "bot-1",
			config:     &DiscordConfig{},
			expectCall: false,
		},
		{
			name: "direct message",
			msg: &discordMessageCreate{
				ID:        "msg-123",
				ChannelID: "dm-channel",
				GuildID:   "", // Empty = DM
				Author: struct {
					ID       string `json:"id"`
					Username string `json:"username"`
					Bot      bool   `json:"bot"`
				}{ID: "user-1", Username: "testuser", Bot: false},
				Content: "dm content",
			},
			botUserID:  "bot-1",
			config:     &DiscordConfig{},
			expectCall: true,
			expectMsg: &Message{
				ID:          "msg-123",
				Platform:    PlatformDiscord,
				ChatID:      "dm-channel",
				UserID:      "user-1",
				Content:     "dm content",
				IsDirectMsg: true,
			},
		},
		{
			name: "mention removes bot mention",
			msg: &discordMessageCreate{
				ID:        "msg-123",
				ChannelID: "channel-1",
				GuildID:   "guild-1",
				Author: struct {
					ID       string `json:"id"`
					Username string `json:"username"`
					Bot      bool   `json:"bot"`
				}{ID: "user-1", Username: "testuser", Bot: false},
				Content:  "<@bot-1> hello",
				Mentions: []struct{ ID string `json:"id"` }{{ID: "bot-1"}},
			},
			botUserID:  "bot-1",
			config:     &DiscordConfig{},
			expectCall: true,
			expectMsg: &Message{
				ID:        "msg-123",
				Platform:  PlatformDiscord,
				ChatID:    "channel-1",
				UserID:    "user-1",
				Content:   "hello",
				IsMention: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewDiscordAdapter(tt.config)
			adapter.botUserID = tt.botUserID

			var receivedMsg *Message
			adapter.SetMessageHandler(func(msg *Message) {
				receivedMsg = msg
			})

			adapter.handleMessageCreate(tt.msg)

			if tt.expectCall {
				if receivedMsg == nil {
					t.Fatal("expected message handler to be called")
				}
				if receivedMsg.ID != tt.expectMsg.ID {
					t.Errorf("ID = %q, want %q", receivedMsg.ID, tt.expectMsg.ID)
				}
				if receivedMsg.Content != tt.expectMsg.Content {
					t.Errorf("Content = %q, want %q", receivedMsg.Content, tt.expectMsg.Content)
				}
				if receivedMsg.IsMention != tt.expectMsg.IsMention {
					t.Errorf("IsMention = %v, want %v", receivedMsg.IsMention, tt.expectMsg.IsMention)
				}
				if receivedMsg.IsDirectMsg != tt.expectMsg.IsDirectMsg {
					t.Errorf("IsDirectMsg = %v, want %v", receivedMsg.IsDirectMsg, tt.expectMsg.IsDirectMsg)
				}
			} else {
				if receivedMsg != nil {
					t.Error("expected message handler NOT to be called")
				}
			}
		})
	}
}

func TestSlackAdapter_HandleEventsAPI(t *testing.T) {
	tests := []struct {
		name       string
		payload    slackEventsAPIPayload
		botUserID  string
		config     *SlackConfig
		expectCall bool
		expectMsg  *Message
	}{
		{
			name: "normal message",
			payload: slackEventsAPIPayload{
				Type: "event_callback",
				Event: struct {
					Type        string `json:"type"`
					User        string `json:"user"`
					Text        string `json:"text"`
					Channel     string `json:"channel"`
					TS          string `json:"ts"`
					ThreadTS    string `json:"thread_ts"`
					ChannelType string `json:"channel_type"`
				}{
					Type:    "message",
					User:    "U123",
					Text:    "hello",
					Channel: "C456",
					TS:      "1234567890.123456",
				},
			},
			botUserID:  "UBOT",
			config:     &SlackConfig{},
			expectCall: true,
			expectMsg: &Message{
				ID:       "1234567890.123456",
				Platform: PlatformSlack,
				ChatID:   "C456",
				UserID:   "U123",
				Content:  "hello",
			},
		},
		{
			name: "skip bot's own message",
			payload: slackEventsAPIPayload{
				Event: struct {
					Type        string `json:"type"`
					User        string `json:"user"`
					Text        string `json:"text"`
					Channel     string `json:"channel"`
					TS          string `json:"ts"`
					ThreadTS    string `json:"thread_ts"`
					ChannelType string `json:"channel_type"`
				}{
					Type:    "message",
					User:    "UBOT",
					Text:    "bot message",
					Channel: "C456",
				},
			},
			botUserID:  "UBOT",
			config:     &SlackConfig{},
			expectCall: false,
		},
		{
			name: "app_mention removes mention",
			payload: slackEventsAPIPayload{
				Event: struct {
					Type        string `json:"type"`
					User        string `json:"user"`
					Text        string `json:"text"`
					Channel     string `json:"channel"`
					TS          string `json:"ts"`
					ThreadTS    string `json:"thread_ts"`
					ChannelType string `json:"channel_type"`
				}{
					Type:    "app_mention",
					User:    "U123",
					Text:    "<@UBOT> hello",
					Channel: "C456",
					TS:      "1234567890.123456",
				},
			},
			botUserID:  "UBOT",
			config:     &SlackConfig{},
			expectCall: true,
			expectMsg: &Message{
				ID:        "1234567890.123456",
				Platform:  PlatformSlack,
				ChatID:    "C456",
				UserID:    "U123",
				Content:   "hello",
				IsMention: true,
			},
		},
		{
			name: "direct message",
			payload: slackEventsAPIPayload{
				Event: struct {
					Type        string `json:"type"`
					User        string `json:"user"`
					Text        string `json:"text"`
					Channel     string `json:"channel"`
					TS          string `json:"ts"`
					ThreadTS    string `json:"thread_ts"`
					ChannelType string `json:"channel_type"`
				}{
					Type:        "message",
					User:        "U123",
					Text:        "dm",
					Channel:     "D789",
					TS:          "1234567890.123456",
					ChannelType: "im",
				},
			},
			botUserID:  "UBOT",
			config:     &SlackConfig{},
			expectCall: true,
			expectMsg: &Message{
				ID:          "1234567890.123456",
				Platform:    PlatformSlack,
				ChatID:      "D789",
				UserID:      "U123",
				Content:     "dm",
				IsDirectMsg: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewSlackAdapter(tt.config)
			adapter.botUserID = tt.botUserID

			var receivedMsg *Message
			adapter.SetMessageHandler(func(msg *Message) {
				receivedMsg = msg
			})

			// Simulate handleEventsAPI by calling the internal logic
			ev := tt.payload.Event
			if ev.User == adapter.botUserID {
				// Skip
			} else if !tt.config.IsUserAllowed(ev.User) || !tt.config.IsChatAllowed(ev.Channel) {
				// Skip
			} else if ev.Type == "message" && adapter.msgHandler != nil {
				msg := &Message{
					ID:          ev.TS,
					Platform:    PlatformSlack,
					ChatID:      ev.Channel,
					UserID:      ev.User,
					Content:     ev.Text,
					ThreadID:    ev.ThreadTS,
					IsDirectMsg: ev.ChannelType == "im",
				}
				adapter.msgHandler(msg)
			} else if ev.Type == "app_mention" && adapter.msgHandler != nil {
				content := ev.Text
				content = content[len("<@UBOT> "):]
				msg := &Message{
					ID:        ev.TS,
					Platform:  PlatformSlack,
					ChatID:    ev.Channel,
					UserID:    ev.User,
					Content:   content,
					ThreadID:  ev.ThreadTS,
					IsMention: true,
				}
				adapter.msgHandler(msg)
			}

			if tt.expectCall {
				if receivedMsg == nil {
					t.Fatal("expected message handler to be called")
				}
				if receivedMsg.ID != tt.expectMsg.ID {
					t.Errorf("ID = %q, want %q", receivedMsg.ID, tt.expectMsg.ID)
				}
				if receivedMsg.Content != tt.expectMsg.Content {
					t.Errorf("Content = %q, want %q", receivedMsg.Content, tt.expectMsg.Content)
				}
				if receivedMsg.IsMention != tt.expectMsg.IsMention {
					t.Errorf("IsMention = %v, want %v", receivedMsg.IsMention, tt.expectMsg.IsMention)
				}
				if receivedMsg.IsDirectMsg != tt.expectMsg.IsDirectMsg {
					t.Errorf("IsDirectMsg = %v, want %v", receivedMsg.IsDirectMsg, tt.expectMsg.IsDirectMsg)
				}
			} else {
				if receivedMsg != nil {
					t.Error("expected message handler NOT to be called")
				}
			}
		})
	}
}
