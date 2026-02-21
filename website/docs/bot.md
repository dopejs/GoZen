---
sidebar_position: 10
title: Bot Gateway
---

# Bot Gateway

Monitor and control your Claude Code sessions remotely via chat platforms. The bot connects to running `zen` processes via IPC and lets you:

- View connected processes and their status
- Send tasks to specific processes
- Receive notifications for approvals, errors, and completions
- Control tasks (pause/resume/cancel)

## Supported Platforms

| Platform | Setup Required |
|----------|----------------|
| [Telegram](#telegram) | BotFather token |
| [Discord](#discord) | Bot application token |
| [Slack](#slack) | Bot + App tokens (Socket Mode) |
| [Lark/Feishu](#larkfeishu) | App ID + Secret |
| [Facebook Messenger](#facebook-messenger) | Page token + Verify token |

## Basic Configuration

```json
{
  "bot": {
    "enabled": true,
    "socket_path": "/tmp/zen-bot.sock",
    "platforms": {
      // Platform-specific config (see below)
    },
    "interaction": {
      "require_mention": true,
      "mention_keywords": ["@zen", "/zen"],
      "direct_message_mode": "always",
      "channel_mode": "mention"
    },
    "aliases": {
      "api": "/path/to/api-project",
      "web": "/path/to/web-project"
    },
    "notify": {
      "default_platform": "telegram",
      "default_chat_id": "-100123456789",
      "quiet_hours_start": "23:00",
      "quiet_hours_end": "07:00",
      "quiet_hours_zone": "Asia/Shanghai"
    }
  }
}
```

## Bot Commands

| Command | Description |
|---------|-------------|
| `list` | List all connected processes |
| `status [name]` | Show process status |
| `bind <name>` | Bind to a process for subsequent commands |
| `pause [name]` | Pause the current task |
| `resume [name]` | Resume a paused task |
| `cancel [name]` | Cancel the current task |
| `<name> <task>` | Send a task to a process |
| `help` | Show available commands |

### Natural Language Support

The bot understands natural language queries in multiple languages:

- "show me the status of gozen"
- "帮我看看 gozen 的状态"
- "list all processes"
- "pause the api project"

## Interaction Modes

### Direct Messages

Set `direct_message_mode` to control how the bot responds in DMs:

- `"always"` — Always respond (no mention required)
- `"mention"` — Only respond when mentioned

### Channel Messages

Set `channel_mode` to control behavior in group chats:

- `"always"` — Respond to all messages
- `"mention"` — Only respond when mentioned (recommended)

### Mention Keywords

Configure what triggers the bot:

```json
{
  "interaction": {
    "require_mention": true,
    "mention_keywords": ["@zen", "/zen", "zen"]
  }
}
```

## Project Aliases

Define short names for your projects:

```json
{
  "aliases": {
    "api": "/Users/john/projects/api-server",
    "web": "/Users/john/projects/web-app",
    "backend": "/Users/john/work/backend"
  }
}
```

Then use them in commands:

```
api run tests
web build production
status backend
```

## Platform Setup

### Telegram

1. Create a bot via [@BotFather](https://t.me/botfather):
   - Send `/newbot` and follow the prompts
   - Copy the token (e.g., `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)

2. Get your user ID:
   - Send a message to [@userinfobot](https://t.me/userinfobot)
   - Copy your numeric user ID

3. Configure:

```json
{
  "platforms": {
    "telegram": {
      "enabled": true,
      "token": "123456789:ABCdefGHIjklMNOpqrsTUVwxyz",
      "allowed_users": ["your_username", "123456789"],
      "allowed_chats": ["-100123456789"]
    }
  }
}
```

**Security options:**
- `allowed_users` — Usernames or user IDs that can interact with the bot
- `allowed_chats` — Group chat IDs where the bot responds (get via [@getidsbot](https://t.me/getidsbot))

### Discord

1. Create a Discord Application:
   - Go to [Discord Developer Portal](https://discord.com/developers/applications)
   - Click "New Application" and give it a name
   - Go to "Bot" section and click "Add Bot"
   - Copy the token

2. Enable required intents:
   - In the Bot section, enable "Message Content Intent"
   - Enable "Server Members Intent" if using user filtering

3. Invite the bot to your server:
   - Go to OAuth2 → URL Generator
   - Select scopes: `bot`
   - Select permissions: `Send Messages`, `Read Message History`
   - Use the generated URL to invite

4. Configure:

```json
{
  "platforms": {
    "discord": {
      "enabled": true,
      "token": "MTIzNDU2Nzg5MDEyMzQ1Njc4.XXXXXX.XXXXXXXXXXXXXXXXXXXXXXXX",
      "allowed_users": ["user_id_1", "user_id_2"],
      "allowed_channels": ["channel_id_1"],
      "allowed_guilds": ["guild_id_1"]
    }
  }
}
```

**Get IDs:** Enable Developer Mode in Discord settings, then right-click users/channels/servers to copy IDs.

### Slack

1. Create a Slack App:
   - Go to [Slack API](https://api.slack.com/apps)
   - Click "Create New App" → "From scratch"
   - Name your app and select workspace

2. Enable Socket Mode:
   - Go to "Socket Mode" and enable it
   - Generate an App-Level Token with `connections:write` scope
   - Copy the token (starts with `xapp-`)

3. Configure Bot Token:
   - Go to "OAuth & Permissions"
   - Add scopes: `chat:write`, `channels:history`, `groups:history`, `im:history`, `mpim:history`
   - Install to workspace and copy the Bot Token (starts with `xoxb-`)

4. Enable Events:
   - Go to "Event Subscriptions" and enable
   - Subscribe to: `message.channels`, `message.groups`, `message.im`, `message.mpim`

5. Configure:

```json
{
  "platforms": {
    "slack": {
      "enabled": true,
      "bot_token": "xoxb-xxx-xxx-xxx",
      "app_token": "xapp-xxx-xxx-xxx",
      "allowed_users": ["U12345678"],
      "allowed_channels": ["C12345678"]
    }
  }
}
```

### Lark/Feishu

1. Create a Lark App:
   - Go to [Lark Open Platform](https://open.larksuite.com/) or [Feishu Open Platform](https://open.feishu.cn/)
   - Create a new app
   - Copy App ID and App Secret

2. Configure permissions:
   - Add `im:message:receive_v1` event
   - Add `im:message:send_v1` permission

3. Configure webhook:
   - Set up event subscription URL (or use WebSocket mode)

4. Configure:

```json
{
  "platforms": {
    "lark": {
      "enabled": true,
      "app_id": "cli_xxxxx",
      "app_secret": "xxxxxxxxxxxxx",
      "allowed_users": ["ou_xxxxx"],
      "allowed_chats": ["oc_xxxxx"]
    }
  }
}
```

### Facebook Messenger

1. Create a Facebook App:
   - Go to [Facebook Developers](https://developers.facebook.com/)
   - Create a new app with "Business" type
   - Add "Messenger" product

2. Configure Messenger:
   - Generate a Page Access Token
   - Set up webhook with verify token
   - Subscribe to `messages` event

3. Configure:

```json
{
  "platforms": {
    "fbmessenger": {
      "enabled": true,
      "page_token": "EAAxxxxx",
      "verify_token": "your_verify_token",
      "app_secret": "xxxxx",
      "allowed_users": ["psid_1", "psid_2"]
    }
  }
}
```

**Note:** Facebook Messenger requires a publicly accessible webhook URL. Consider using a service like ngrok for development.

## Notifications

Configure where the bot sends notifications:

```json
{
  "notify": {
    "default_platform": "telegram",
    "default_chat_id": "-100123456789",
    "quiet_hours_start": "23:00",
    "quiet_hours_end": "07:00",
    "quiet_hours_zone": "UTC"
  }
}
```

### Notification Types

- **Approval requests** — When Claude Code needs permission for an action
- **Task completion** — When a task finishes successfully
- **Errors** — When a task fails or encounters an error
- **Status changes** — When a process connects/disconnects

### Quiet Hours

During quiet hours, non-urgent notifications are suppressed. Approval requests are always sent.

## Security Best Practices

1. **Restrict users** — Always configure `allowed_users` to limit who can control your sessions
2. **Use private channels** — Avoid using the bot in public channels
3. **Protect tokens** — Never commit bot tokens to version control
4. **Review approvals** — Carefully review approval requests before accepting

## Troubleshooting

### Bot not responding

1. Check if the daemon is running: `zen daemon status`
2. Verify bot configuration in Web UI
3. Check daemon logs: `tail -f ~/.zen/zend.log`

### Connection issues

1. Verify tokens are correct
2. Check network connectivity
3. For Slack/Discord, ensure required intents are enabled

### Process not showing in list

1. Ensure the process was started with `zen` (not directly with `claude`)
2. Check if the socket path matches the bot configuration
