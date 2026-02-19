package web

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

func setupTestServerWithBot(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	config.ResetDefaultStore()
	t.Cleanup(func() { config.ResetDefaultStore() })

	configDir := filepath.Join(dir, config.ConfigDir)
	os.MkdirAll(configDir, 0755)
	cfg := &config.OpenCCConfig{
		Version: config.CurrentConfigVersion,
		Bot: &config.BotConfig{
			Enabled:    true,
			Profile:    "default",
			SocketPath: "/tmp/test.sock",
			Platforms: &config.BotPlatformsConfig{
				Telegram: &config.BotTelegramConfig{
					Enabled:      true,
					Token:        "telegram-secret-token-12345678",
					AllowedUsers: []string{"user1", "user2"},
					AllowedChats: []string{"-100123"},
				},
				Discord: &config.BotDiscordConfig{
					Enabled:       true,
					Token:         "discord-secret-token-87654321",
					AllowedGuilds: []string{"guild1"},
				},
				Slack: &config.BotSlackConfig{
					Enabled:  true,
					BotToken: "xoxb-slack-bot-token",
					AppToken: "xapp-slack-app-token",
				},
				Lark: &config.BotLarkConfig{
					Enabled:   true,
					AppID:     "lark-app-id",
					AppSecret: "lark-secret-12345678",
				},
				FBMessenger: &config.BotFBMessengerConfig{
					Enabled:     true,
					PageToken:   "fb-page-token-secret",
					VerifyToken: "fb-verify-token",
					AppSecret:   "fb-app-secret-12345",
				},
			},
			Interaction: &config.BotInteractionConfig{
				RequireMention:  true,
				MentionKeywords: []string{"@zen", "/zen"},
				DirectMsgMode:   "always",
				ChannelMode:     "mention",
			},
			Aliases: map[string]string{
				"api":     "/path/to/api",
				"backend": "/path/to/backend",
			},
			Notify: &config.BotNotifyConfig{
				DefaultPlatform: "telegram",
				DefaultChatID:   "-100123",
				QuietHoursStart: "23:00",
				QuietHoursEnd:   "07:00",
				QuietHoursZone:  "UTC",
			},
		},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	configPath := filepath.Join(configDir, config.ConfigFile)
	os.WriteFile(configPath, data, 0600)

	// Force reload by getting the store
	_ = config.DefaultStore()

	logger := log.New(io.Discard, "", 0)
	return NewServer("1.0.0-test", logger, 0)
}

func TestGetBot(t *testing.T) {
	s := setupTestServerWithBot(t)
	w := doRequest(s, "GET", "/api/v1/bot", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp botResponse
	decodeJSON(t, w, &resp)

	if !resp.Enabled {
		t.Error("expected enabled to be true")
	}
	if resp.Profile != "default" {
		t.Errorf("expected profile 'default', got '%s'", resp.Profile)
	}
	if resp.SocketPath != "/tmp/test.sock" {
		t.Errorf("expected socket path '/tmp/test.sock', got '%s'", resp.SocketPath)
	}
}

func TestGetBot_TokensMasked(t *testing.T) {
	s := setupTestServerWithBot(t)
	w := doRequest(s, "GET", "/api/v1/bot", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp botResponse
	decodeJSON(t, w, &resp)

	// Telegram token should be masked
	if resp.Platforms.Telegram.Token == "telegram-secret-token-12345678" {
		t.Error("telegram token should be masked")
	}
	if resp.Platforms.Telegram.Token != "teleg...5678" {
		t.Errorf("unexpected masked token: %s", resp.Platforms.Telegram.Token)
	}

	// Discord token should be masked
	if resp.Platforms.Discord.Token == "discord-secret-token-87654321" {
		t.Error("discord token should be masked")
	}

	// Slack tokens should be masked
	if resp.Platforms.Slack.BotToken == "xoxb-slack-bot-token" {
		t.Error("slack bot token should be masked")
	}
	if resp.Platforms.Slack.AppToken == "xapp-slack-app-token" {
		t.Error("slack app token should be masked")
	}

	// Lark app secret should be masked
	if resp.Platforms.Lark.AppSecret == "lark-secret-12345678" {
		t.Error("lark app secret should be masked")
	}

	// FB Messenger tokens should be masked
	if resp.Platforms.FBMessenger.PageToken == "fb-page-token-secret" {
		t.Error("fb page token should be masked")
	}
	if resp.Platforms.FBMessenger.AppSecret == "fb-app-secret-12345" {
		t.Error("fb app secret should be masked")
	}
}

func TestGetBot_Platforms(t *testing.T) {
	s := setupTestServerWithBot(t)
	w := doRequest(s, "GET", "/api/v1/bot", nil)

	var resp botResponse
	decodeJSON(t, w, &resp)

	if resp.Platforms == nil {
		t.Fatal("platforms should not be nil")
	}

	// Telegram
	if resp.Platforms.Telegram == nil {
		t.Fatal("telegram should not be nil")
	}
	if !resp.Platforms.Telegram.Enabled {
		t.Error("telegram should be enabled")
	}
	if len(resp.Platforms.Telegram.AllowedUsers) != 2 {
		t.Errorf("expected 2 allowed users, got %d", len(resp.Platforms.Telegram.AllowedUsers))
	}

	// Discord
	if resp.Platforms.Discord == nil {
		t.Fatal("discord should not be nil")
	}
	if len(resp.Platforms.Discord.AllowedGuilds) != 1 {
		t.Errorf("expected 1 allowed guild, got %d", len(resp.Platforms.Discord.AllowedGuilds))
	}

	// Lark
	if resp.Platforms.Lark == nil {
		t.Fatal("lark should not be nil")
	}
	if resp.Platforms.Lark.AppID != "lark-app-id" {
		t.Errorf("expected app id 'lark-app-id', got '%s'", resp.Platforms.Lark.AppID)
	}

	// FB Messenger
	if resp.Platforms.FBMessenger == nil {
		t.Fatal("fbmessenger should not be nil")
	}
	if resp.Platforms.FBMessenger.VerifyToken != "fb-verify-token" {
		t.Errorf("expected verify token 'fb-verify-token', got '%s'", resp.Platforms.FBMessenger.VerifyToken)
	}
}

func TestGetBot_Interaction(t *testing.T) {
	s := setupTestServerWithBot(t)
	w := doRequest(s, "GET", "/api/v1/bot", nil)

	var resp botResponse
	decodeJSON(t, w, &resp)

	if resp.Interaction == nil {
		t.Fatal("interaction should not be nil")
	}
	if !resp.Interaction.RequireMention {
		t.Error("require_mention should be true")
	}
	if len(resp.Interaction.MentionKeywords) != 2 {
		t.Errorf("expected 2 mention keywords, got %d", len(resp.Interaction.MentionKeywords))
	}
	if resp.Interaction.DirectMsgMode != "always" {
		t.Errorf("expected direct_message_mode 'always', got '%s'", resp.Interaction.DirectMsgMode)
	}
	if resp.Interaction.ChannelMode != "mention" {
		t.Errorf("expected channel_mode 'mention', got '%s'", resp.Interaction.ChannelMode)
	}
}

func TestGetBot_Aliases(t *testing.T) {
	s := setupTestServerWithBot(t)
	w := doRequest(s, "GET", "/api/v1/bot", nil)

	var resp botResponse
	decodeJSON(t, w, &resp)

	if len(resp.Aliases) != 2 {
		t.Errorf("expected 2 aliases, got %d", len(resp.Aliases))
	}
	if resp.Aliases["api"] != "/path/to/api" {
		t.Errorf("expected alias 'api' -> '/path/to/api', got '%s'", resp.Aliases["api"])
	}
}

func TestGetBot_Notify(t *testing.T) {
	s := setupTestServerWithBot(t)
	w := doRequest(s, "GET", "/api/v1/bot", nil)

	var resp botResponse
	decodeJSON(t, w, &resp)

	if resp.Notify == nil {
		t.Fatal("notify should not be nil")
	}
	if resp.Notify.DefaultPlatform != "telegram" {
		t.Errorf("expected default_platform 'telegram', got '%s'", resp.Notify.DefaultPlatform)
	}
	if resp.Notify.QuietHoursStart != "23:00" {
		t.Errorf("expected quiet_hours_start '23:00', got '%s'", resp.Notify.QuietHoursStart)
	}
}

func TestGetBot_Empty(t *testing.T) {
	s := setupTestServer(t)
	w := doRequest(s, "GET", "/api/v1/bot", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp botResponse
	decodeJSON(t, w, &resp)

	if resp.Enabled {
		t.Error("expected enabled to be false for empty config")
	}
}

func TestUpdateBot_General(t *testing.T) {
	s := setupTestServerWithBot(t)

	update := map[string]interface{}{
		"enabled":     false,
		"profile":     "work",
		"socket_path": "/tmp/new.sock",
	}

	w := doRequest(s, "PUT", "/api/v1/bot", update)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp botResponse
	decodeJSON(t, w, &resp)

	if resp.Enabled {
		t.Error("expected enabled to be false")
	}
	if resp.Profile != "work" {
		t.Errorf("expected profile 'work', got '%s'", resp.Profile)
	}
	if resp.SocketPath != "/tmp/new.sock" {
		t.Errorf("expected socket path '/tmp/new.sock', got '%s'", resp.SocketPath)
	}
}

func TestUpdateBot_Platforms(t *testing.T) {
	s := setupTestServerWithBot(t)

	update := map[string]interface{}{
		"platforms": map[string]interface{}{
			"telegram": map[string]interface{}{
				"enabled":       false,
				"token":         "new-telegram-token-12345678",
				"allowed_users": []string{"newuser"},
			},
		},
	}

	w := doRequest(s, "PUT", "/api/v1/bot", update)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp botResponse
	decodeJSON(t, w, &resp)

	if resp.Platforms.Telegram.Enabled {
		t.Error("telegram should be disabled")
	}
	if len(resp.Platforms.Telegram.AllowedUsers) != 1 {
		t.Errorf("expected 1 allowed user, got %d", len(resp.Platforms.Telegram.AllowedUsers))
	}
	// Token should be masked in response
	if resp.Platforms.Telegram.Token == "new-telegram-token-12345678" {
		t.Error("new token should be masked in response")
	}
}

func TestUpdateBot_PreservesExistingToken(t *testing.T) {
	s := setupTestServerWithBot(t)

	// Update without providing token - should preserve existing
	update := map[string]interface{}{
		"platforms": map[string]interface{}{
			"telegram": map[string]interface{}{
				"enabled":       true,
				"allowed_users": []string{"newuser"},
			},
		},
	}

	w := doRequest(s, "PUT", "/api/v1/bot", update)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify token was preserved by checking it's still masked (not empty)
	var resp botResponse
	decodeJSON(t, w, &resp)

	if resp.Platforms.Telegram.Token == "" {
		t.Error("token should be preserved when not provided in update")
	}
}

func TestUpdateBot_Interaction(t *testing.T) {
	s := setupTestServerWithBot(t)

	update := map[string]interface{}{
		"interaction": map[string]interface{}{
			"require_mention":     false,
			"mention_keywords":    []string{"@bot"},
			"direct_message_mode": "mention",
			"channel_mode":        "always",
		},
	}

	w := doRequest(s, "PUT", "/api/v1/bot", update)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp botResponse
	decodeJSON(t, w, &resp)

	if resp.Interaction.RequireMention {
		t.Error("require_mention should be false")
	}
	if len(resp.Interaction.MentionKeywords) != 1 {
		t.Errorf("expected 1 mention keyword, got %d", len(resp.Interaction.MentionKeywords))
	}
	if resp.Interaction.DirectMsgMode != "mention" {
		t.Errorf("expected direct_message_mode 'mention', got '%s'", resp.Interaction.DirectMsgMode)
	}
}

func TestUpdateBot_Aliases(t *testing.T) {
	s := setupTestServerWithBot(t)

	update := map[string]interface{}{
		"aliases": map[string]string{
			"web": "/path/to/web",
		},
	}

	w := doRequest(s, "PUT", "/api/v1/bot", update)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp botResponse
	decodeJSON(t, w, &resp)

	if len(resp.Aliases) != 1 {
		t.Errorf("expected 1 alias, got %d", len(resp.Aliases))
	}
	if resp.Aliases["web"] != "/path/to/web" {
		t.Errorf("expected alias 'web' -> '/path/to/web', got '%s'", resp.Aliases["web"])
	}
}

func TestUpdateBot_Notify(t *testing.T) {
	s := setupTestServerWithBot(t)

	update := map[string]interface{}{
		"notify": map[string]interface{}{
			"default_platform":  "discord",
			"default_chat_id":   "channel123",
			"quiet_hours_start": "22:00",
			"quiet_hours_end":   "08:00",
			"quiet_hours_zone":  "America/New_York",
		},
	}

	w := doRequest(s, "PUT", "/api/v1/bot", update)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp botResponse
	decodeJSON(t, w, &resp)

	if resp.Notify.DefaultPlatform != "discord" {
		t.Errorf("expected default_platform 'discord', got '%s'", resp.Notify.DefaultPlatform)
	}
	if resp.Notify.QuietHoursZone != "America/New_York" {
		t.Errorf("expected quiet_hours_zone 'America/New_York', got '%s'", resp.Notify.QuietHoursZone)
	}
}

func TestUpdateBot_InvalidJSON(t *testing.T) {
	s := setupTestServerWithBot(t)

	w := doRequestRaw(s, "PUT", "/api/v1/bot", []byte("invalid json"))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleBot_MethodNotAllowed(t *testing.T) {
	s := setupTestServerWithBot(t)

	w := doRequest(s, "DELETE", "/api/v1/bot", nil)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestToBotResponse_NilBot(t *testing.T) {
	resp := toBotResponse(nil, true)
	if resp == nil {
		t.Fatal("response should not be nil")
	}
	if resp.Enabled {
		t.Error("enabled should be false for nil bot")
	}
}

func TestToBotResponse_NilPlatforms(t *testing.T) {
	bot := &config.BotConfig{
		Enabled: true,
	}
	resp := toBotResponse(bot, true)
	if resp.Platforms != nil {
		t.Error("platforms should be nil when bot.Platforms is nil")
	}
}

func TestMergeBotConfig_NilExisting(t *testing.T) {
	update := &config.BotConfig{
		Enabled: true,
		Profile: "test",
	}
	result := mergeBotConfig(nil, update)
	if result != update {
		t.Error("should return update when existing is nil")
	}
}

func TestMergeBotConfig_PreservesExistingPlatforms(t *testing.T) {
	existing := &config.BotConfig{
		Platforms: &config.BotPlatformsConfig{
			Telegram: &config.BotTelegramConfig{
				Enabled: true,
				Token:   "existing-token",
			},
		},
	}
	update := &config.BotConfig{
		Enabled: true,
	}
	result := mergeBotConfig(existing, update)
	if result.Platforms == nil || result.Platforms.Telegram == nil {
		t.Fatal("should preserve existing platforms")
	}
	if result.Platforms.Telegram.Token != "existing-token" {
		t.Error("should preserve existing token")
	}
}

func TestMergeBotConfig_PreservesTokenWhenEmpty(t *testing.T) {
	existing := &config.BotConfig{
		Platforms: &config.BotPlatformsConfig{
			Discord: &config.BotDiscordConfig{
				Token: "existing-discord-token",
			},
			Slack: &config.BotSlackConfig{
				BotToken: "existing-bot-token",
				AppToken: "existing-app-token",
			},
			Lark: &config.BotLarkConfig{
				AppSecret: "existing-lark-secret",
			},
			FBMessenger: &config.BotFBMessengerConfig{
				PageToken: "existing-page-token",
				AppSecret: "existing-fb-secret",
			},
		},
	}
	update := &config.BotConfig{
		Platforms: &config.BotPlatformsConfig{
			Discord:     &config.BotDiscordConfig{Enabled: true},
			Slack:       &config.BotSlackConfig{Enabled: true},
			Lark:        &config.BotLarkConfig{Enabled: true},
			FBMessenger: &config.BotFBMessengerConfig{Enabled: true},
		},
	}
	result := mergeBotConfig(existing, update)

	if result.Platforms.Discord.Token != "existing-discord-token" {
		t.Error("should preserve discord token")
	}
	if result.Platforms.Slack.BotToken != "existing-bot-token" {
		t.Error("should preserve slack bot token")
	}
	if result.Platforms.Slack.AppToken != "existing-app-token" {
		t.Error("should preserve slack app token")
	}
	if result.Platforms.Lark.AppSecret != "existing-lark-secret" {
		t.Error("should preserve lark app secret")
	}
	if result.Platforms.FBMessenger.PageToken != "existing-page-token" {
		t.Error("should preserve fb page token")
	}
	if result.Platforms.FBMessenger.AppSecret != "existing-fb-secret" {
		t.Error("should preserve fb app secret")
	}
}

func TestDecryptBotTokens_NilPlatforms(t *testing.T) {
	bot := &config.BotConfig{}
	// Should not panic
	decryptBotTokens(nil, bot)
}
