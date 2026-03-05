package web

import (
	"net/http"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/proxy"
)

// botResponse is the JSON shape returned for bot configuration.
type botResponse struct {
	Enabled     bool                         `json:"enabled"`
	Profile     string                       `json:"profile,omitempty"`
	Model       string                       `json:"model,omitempty"`
	SocketPath  string                       `json:"socket_path,omitempty"`
	Platforms   *botPlatformsResponse        `json:"platforms,omitempty"`
	Interaction *config.BotInteractionConfig `json:"interaction,omitempty"`
	Aliases     map[string]string            `json:"aliases,omitempty"`
	Notify      *config.BotNotifyConfig      `json:"notify,omitempty"`
	RecentPaths []string                     `json:"recent_paths,omitempty"`
	HistorySize int                          `json:"history_size,omitempty"`
}

type botPlatformsResponse struct {
	Telegram    *botTelegramResponse    `json:"telegram,omitempty"`
	Discord     *botDiscordResponse     `json:"discord,omitempty"`
	Slack       *botSlackResponse       `json:"slack,omitempty"`
	Lark        *botLarkResponse        `json:"lark,omitempty"`
	FBMessenger *botFBMessengerResponse `json:"fbmessenger,omitempty"`
}

type botTelegramResponse struct {
	Enabled      bool     `json:"enabled"`
	Token        string   `json:"token"`
	AllowedUsers []string `json:"allowed_users,omitempty"`
	AllowedChats []string `json:"allowed_chats,omitempty"`
}

type botDiscordResponse struct {
	Enabled         bool     `json:"enabled"`
	Token           string   `json:"token"`
	AllowedUsers    []string `json:"allowed_users,omitempty"`
	AllowedChannels []string `json:"allowed_channels,omitempty"`
	AllowedGuilds   []string `json:"allowed_guilds,omitempty"`
}

type botSlackResponse struct {
	Enabled         bool     `json:"enabled"`
	BotToken        string   `json:"bot_token"`
	AppToken        string   `json:"app_token"`
	AllowedUsers    []string `json:"allowed_users,omitempty"`
	AllowedChannels []string `json:"allowed_channels,omitempty"`
}

type botLarkResponse struct {
	Enabled      bool     `json:"enabled"`
	AppID        string   `json:"app_id"`
	AppSecret    string   `json:"app_secret"`
	AllowedUsers []string `json:"allowed_users,omitempty"`
	AllowedChats []string `json:"allowed_chats,omitempty"`
}

type botFBMessengerResponse struct {
	Enabled      bool     `json:"enabled"`
	PageToken    string   `json:"page_token"`
	VerifyToken  string   `json:"verify_token"`
	AppSecret    string   `json:"app_secret,omitempty"`
	AllowedUsers []string `json:"allowed_users,omitempty"`
}

// handleBot handles GET/PUT /api/v1/bot.
func (s *Server) handleBot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getBot(w, r)
	case http.MethodPut:
		s.updateBot(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) getBot(w http.ResponseWriter, r *http.Request) {
	store := config.DefaultStore()
	bot := store.GetBot()

	resp := toBotResponse(bot)

	// Add recent paths from usage tracker (excluding already aliased paths)
	if tracker := proxy.GetGlobalUsageTracker(); tracker != nil {
		if paths, err := tracker.GetRecentPaths(20); err == nil {
			// Filter out paths that already have aliases
			aliased := make(map[string]bool)
			if bot != nil && bot.Aliases != nil {
				for _, path := range bot.Aliases {
					aliased[path] = true
				}
			}
			var unaliased []string
			for _, path := range paths {
				if !aliased[path] {
					unaliased = append(unaliased, path)
				}
			}
			resp.RecentPaths = unaliased
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) updateBot(w http.ResponseWriter, r *http.Request) {
	var update config.BotConfig
	if err := readJSON(r, &update); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Decrypt tokens if encrypted
	if s.keys != nil {
		decryptBotTokens(s.keys, &update)
	}

	store := config.DefaultStore()
	existing := store.GetBot()

	// Merge update with existing config
	merged := mergeBotConfig(existing, &update)

	if err := store.SetBot(merged); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toBotResponse(merged))
}

func toBotResponse(bot *config.BotConfig) *botResponse {
	if bot == nil {
		return &botResponse{}
	}

	resp := &botResponse{
		Enabled:     bot.Enabled,
		Profile:     bot.Profile,
		Model:       bot.Model,
		SocketPath:  bot.SocketPath,
		Interaction: bot.Interaction,
		Aliases:     bot.Aliases,
		Notify:      bot.Notify,
		HistorySize: bot.HistorySize,
	}

	if bot.Platforms != nil {
		resp.Platforms = &botPlatformsResponse{}

		if bot.Platforms.Telegram != nil {
			resp.Platforms.Telegram = &botTelegramResponse{
				Enabled:      bot.Platforms.Telegram.Enabled,
				Token:        bot.Platforms.Telegram.Token,
				AllowedUsers: bot.Platforms.Telegram.AllowedUsers,
				AllowedChats: bot.Platforms.Telegram.AllowedChats,
			}
		}

		if bot.Platforms.Discord != nil {
			resp.Platforms.Discord = &botDiscordResponse{
				Enabled:         bot.Platforms.Discord.Enabled,
				Token:           bot.Platforms.Discord.Token,
				AllowedUsers:    bot.Platforms.Discord.AllowedUsers,
				AllowedChannels: bot.Platforms.Discord.AllowedChannels,
				AllowedGuilds:   bot.Platforms.Discord.AllowedGuilds,
			}
		}

		if bot.Platforms.Slack != nil {
			resp.Platforms.Slack = &botSlackResponse{
				Enabled:         bot.Platforms.Slack.Enabled,
				BotToken:        bot.Platforms.Slack.BotToken,
				AppToken:        bot.Platforms.Slack.AppToken,
				AllowedUsers:    bot.Platforms.Slack.AllowedUsers,
				AllowedChannels: bot.Platforms.Slack.AllowedChannels,
			}
		}

		if bot.Platforms.Lark != nil {
			resp.Platforms.Lark = &botLarkResponse{
				Enabled:      bot.Platforms.Lark.Enabled,
				AppID:        bot.Platforms.Lark.AppID,
				AppSecret:    bot.Platforms.Lark.AppSecret,
				AllowedUsers: bot.Platforms.Lark.AllowedUsers,
				AllowedChats: bot.Platforms.Lark.AllowedChats,
			}
		}

		if bot.Platforms.FBMessenger != nil {
			resp.Platforms.FBMessenger = &botFBMessengerResponse{
				Enabled:      bot.Platforms.FBMessenger.Enabled,
				PageToken:    bot.Platforms.FBMessenger.PageToken,
				VerifyToken:  bot.Platforms.FBMessenger.VerifyToken,
				AppSecret:    bot.Platforms.FBMessenger.AppSecret,
				AllowedUsers: bot.Platforms.FBMessenger.AllowedUsers,
			}
		}
	}

	return resp
}

func decryptBotTokens(keys *KeyPair, bot *config.BotConfig) {
	if bot.Platforms == nil {
		return
	}

	if bot.Platforms.Telegram != nil && bot.Platforms.Telegram.Token != "" {
		if decrypted, err := keys.MaybeDecryptToken(bot.Platforms.Telegram.Token); err == nil {
			bot.Platforms.Telegram.Token = decrypted
		}
	}

	if bot.Platforms.Discord != nil && bot.Platforms.Discord.Token != "" {
		if decrypted, err := keys.MaybeDecryptToken(bot.Platforms.Discord.Token); err == nil {
			bot.Platforms.Discord.Token = decrypted
		}
	}

	if bot.Platforms.Slack != nil {
		if bot.Platforms.Slack.BotToken != "" {
			if decrypted, err := keys.MaybeDecryptToken(bot.Platforms.Slack.BotToken); err == nil {
				bot.Platforms.Slack.BotToken = decrypted
			}
		}
		if bot.Platforms.Slack.AppToken != "" {
			if decrypted, err := keys.MaybeDecryptToken(bot.Platforms.Slack.AppToken); err == nil {
				bot.Platforms.Slack.AppToken = decrypted
			}
		}
	}

	if bot.Platforms.Lark != nil && bot.Platforms.Lark.AppSecret != "" {
		if decrypted, err := keys.MaybeDecryptToken(bot.Platforms.Lark.AppSecret); err == nil {
			bot.Platforms.Lark.AppSecret = decrypted
		}
	}

	if bot.Platforms.FBMessenger != nil {
		if bot.Platforms.FBMessenger.PageToken != "" {
			if decrypted, err := keys.MaybeDecryptToken(bot.Platforms.FBMessenger.PageToken); err == nil {
				bot.Platforms.FBMessenger.PageToken = decrypted
			}
		}
		if bot.Platforms.FBMessenger.AppSecret != "" {
			if decrypted, err := keys.MaybeDecryptToken(bot.Platforms.FBMessenger.AppSecret); err == nil {
				bot.Platforms.FBMessenger.AppSecret = decrypted
			}
		}
	}
}

func mergeBotConfig(existing, update *config.BotConfig) *config.BotConfig {
	if existing == nil {
		return update
	}

	result := &config.BotConfig{
		Enabled:     update.Enabled,
		Profile:     update.Profile,
		Model:       update.Model,
		SocketPath:  update.SocketPath,
		HistorySize: update.HistorySize,
	}

	// Merge platforms
	if update.Platforms != nil {
		result.Platforms = &config.BotPlatformsConfig{}

		// Telegram
		if update.Platforms.Telegram != nil {
			result.Platforms.Telegram = update.Platforms.Telegram
			// Keep existing token if update token is empty
			if result.Platforms.Telegram.Token == "" && existing.Platforms != nil && existing.Platforms.Telegram != nil {
				result.Platforms.Telegram.Token = existing.Platforms.Telegram.Token
			}
		} else if existing.Platforms != nil {
			result.Platforms.Telegram = existing.Platforms.Telegram
		}

		// Discord
		if update.Platforms.Discord != nil {
			result.Platforms.Discord = update.Platforms.Discord
			if result.Platforms.Discord.Token == "" && existing.Platforms != nil && existing.Platforms.Discord != nil {
				result.Platforms.Discord.Token = existing.Platforms.Discord.Token
			}
		} else if existing.Platforms != nil {
			result.Platforms.Discord = existing.Platforms.Discord
		}

		// Slack
		if update.Platforms.Slack != nil {
			result.Platforms.Slack = update.Platforms.Slack
			if existing.Platforms != nil && existing.Platforms.Slack != nil {
				if result.Platforms.Slack.BotToken == "" {
					result.Platforms.Slack.BotToken = existing.Platforms.Slack.BotToken
				}
				if result.Platforms.Slack.AppToken == "" {
					result.Platforms.Slack.AppToken = existing.Platforms.Slack.AppToken
				}
			}
		} else if existing.Platforms != nil {
			result.Platforms.Slack = existing.Platforms.Slack
		}

		// Lark
		if update.Platforms.Lark != nil {
			result.Platforms.Lark = update.Platforms.Lark
			if result.Platforms.Lark.AppSecret == "" && existing.Platforms != nil && existing.Platforms.Lark != nil {
				result.Platforms.Lark.AppSecret = existing.Platforms.Lark.AppSecret
			}
		} else if existing.Platforms != nil {
			result.Platforms.Lark = existing.Platforms.Lark
		}

		// FBMessenger
		if update.Platforms.FBMessenger != nil {
			result.Platforms.FBMessenger = update.Platforms.FBMessenger
			if existing.Platforms != nil && existing.Platforms.FBMessenger != nil {
				if result.Platforms.FBMessenger.PageToken == "" {
					result.Platforms.FBMessenger.PageToken = existing.Platforms.FBMessenger.PageToken
				}
				if result.Platforms.FBMessenger.AppSecret == "" {
					result.Platforms.FBMessenger.AppSecret = existing.Platforms.FBMessenger.AppSecret
				}
			}
		} else if existing.Platforms != nil {
			result.Platforms.FBMessenger = existing.Platforms.FBMessenger
		}
	} else if existing.Platforms != nil {
		result.Platforms = existing.Platforms
	}

	// Merge interaction
	if update.Interaction != nil {
		result.Interaction = update.Interaction
	} else {
		result.Interaction = existing.Interaction
	}

	// Merge aliases
	if update.Aliases != nil {
		result.Aliases = update.Aliases
	} else {
		result.Aliases = existing.Aliases
	}

	// Merge notify
	if update.Notify != nil {
		result.Notify = update.Notify
	} else {
		result.Notify = existing.Notify
	}

	return result
}
