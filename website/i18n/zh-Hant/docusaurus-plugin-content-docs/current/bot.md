---
sidebar_position: 10
title: Bot 閘道
---

# Bot 閘道

透過聊天平台遠端監控和控制您的 Claude Code 會話。Bot 透過 IPC 連接到執行中的 `zen` 程序，讓您可以：

- 檢視已連接的程序及其狀態
- 向特定程序傳送任務
- 接收審批、錯誤和完成通知
- 控制任務（暫停/繼續/取消）

## 支援的平台

| 平台 | 所需設定 |
|----------|----------------|
| [Telegram](#telegram) | BotFather token |
| [Discord](#discord) | Bot 應用程式 token |
| [Slack](#slack) | Bot + App tokens (Socket Mode) |
| [Lark/飛書](#lark飛書) | App ID + Secret |
| [Facebook Messenger](#facebook-messenger) | Page token + Verify token |

## 基本配置

```json
{
  "bot": {
    "enabled": true,
    "socket_path": "/tmp/zen-bot.sock",
    "platforms": {
      // 平台特定配置（見下文）
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
      "quiet_hours_zone": "Asia/Taipei"
    }
  }
}
```

## Bot 命令

| 命令 | 描述 |
|---------|-------------|
| `list` | 列出所有已連接的程序 |
| `status [name]` | 顯示程序狀態 |
| `bind <name>` | 繫結到一個程序以執行後續命令 |
| `pause [name]` | 暫停目前任務 |
| `resume [name]` | 繼續暫停的任務 |
| `cancel [name]` | 取消目前任務 |
| `<name> <task>` | 向程序傳送任務 |
| `help` | 顯示可用命令 |

### 自然語言支援

Bot 理解多種語言的自然語言查詢：

- "show me the status of gozen"
- "幫我看看 gozen 的狀態"
- "list all processes"
- "pause the api project"

## 互動模式

### 私訊

設定 `direct_message_mode` 來控制 bot 在私訊中的回應方式：

- `"always"` — 始終回應（無需提及）
- `"mention"` — 僅在被提及時回應

### 頻道訊息

設定 `channel_mode` 來控制群聊中的行為：

- `"always"` — 回應所有訊息
- `"mention"` — 僅在被提及時回應（建議）

### 提及關鍵字

配置觸發 bot 的關鍵字：

```json
{
  "interaction": {
    "require_mention": true,
    "mention_keywords": ["@zen", "/zen", "zen"]
  }
}
```

## 專案別名

為您的專案定義簡短名稱：

```json
{
  "aliases": {
    "api": "/Users/john/projects/api-server",
    "web": "/Users/john/projects/web-app",
    "backend": "/Users/john/work/backend"
  }
}
```

然後在命令中使用它們：

```
api run tests
web build production
status backend
```

## 平台設定

### Telegram

1. 透過 [@BotFather](https://t.me/botfather) 建立 bot：
   - 傳送 `/newbot` 並按照提示操作
   - 複製 token（例如：`123456789:ABCdefGHIjklMNOpqrsTUVwxyz`）

2. 取得您的使用者 ID：
   - 向 [@userinfobot](https://t.me/userinfobot) 傳送訊息
   - 複製您的數字使用者 ID

3. 配置：

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

**安全選項：**
- `allowed_users` — 可以與 bot 互動的使用者名稱或使用者 ID
- `allowed_chats` — bot 回應的群聊 ID（透過 [@getidsbot](https://t.me/getidsbot) 取得）

### Discord

1. 建立 Discord 應用程式：
   - 造訪 [Discord 開發者入口](https://discord.com/developers/applications)
   - 點選 "New Application" 並命名
   - 進入 "Bot" 部分並點選 "Add Bot"
   - 複製 token

2. 啟用所需的 intents：
   - 在 Bot 部分，啟用 "Message Content Intent"
   - 如果使用使用者篩選，啟用 "Server Members Intent"

3. 邀請 bot 到您的伺服器：
   - 進入 OAuth2 → URL Generator
   - 選擇 scopes：`bot`
   - 選擇權限：`Send Messages`、`Read Message History`
   - 使用產生的 URL 邀請

4. 配置：

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

**取得 ID：** 在 Discord 設定中啟用開發者模式，然後右鍵點選使用者/頻道/伺服器以複製 ID。

### Slack

1. 建立 Slack 應用程式：
   - 造訪 [Slack API](https://api.slack.com/apps)
   - 點選 "Create New App" → "From scratch"
   - 命名您的應用程式並選擇工作區

2. 啟用 Socket Mode：
   - 進入 "Socket Mode" 並啟用
   - 產生具有 `connections:write` 範圍的應用程式級 Token
   - 複製 token（以 `xapp-` 開頭）

3. 配置 Bot Token：
   - 進入 "OAuth & Permissions"
   - 新增 scopes：`chat:write`、`channels:history`、`groups:history`、`im:history`、`mpim:history`
   - 安裝到工作區並複製 Bot Token（以 `xoxb-` 開頭）

4. 啟用事件：
   - 進入 "Event Subscriptions" 並啟用
   - 訂閱：`message.channels`、`message.groups`、`message.im`、`message.mpim`

5. 配置：

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

### Lark/飛書

1. 建立 Lark 應用程式：
   - 造訪 [Lark 開放平台](https://open.larksuite.com/) 或 [飛書開放平台](https://open.feishu.cn/)
   - 建立新應用程式
   - 複製 App ID 和 App Secret

2. 配置權限：
   - 新增 `im:message:receive_v1` 事件
   - 新增 `im:message:send_v1` 權限

3. 配置 webhook：
   - 設定事件訂閱 URL（或使用 WebSocket 模式）

4. 配置：

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

1. 建立 Facebook 應用程式：
   - 造訪 [Facebook 開發者](https://developers.facebook.com/)
   - 建立類型為 "Business" 的新應用程式
   - 新增 "Messenger" 產品

2. 配置 Messenger：
   - 產生 Page Access Token
   - 使用 verify token 設定 webhook
   - 訂閱 `messages` 事件

3. 配置：

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

**注意：** Facebook Messenger 需要公開可存取的 webhook URL。開發時可考慮使用 ngrok 等服務。

## 通知

配置 bot 傳送通知的位置：

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

### 通知類型

- **審批請求** — 當 Claude Code 需要操作權限時
- **任務完成** — 當任務成功完成時
- **錯誤** — 當任務失敗或遇到錯誤時
- **狀態變更** — 當程序連接/中斷連接時

### 免打擾時間

在免打擾時間內，非緊急通知會被抑制。審批請求始終會傳送。

## 安全最佳實務

1. **限制使用者** — 始終配置 `allowed_users` 以限制誰可以控制您的會話
2. **使用私有頻道** — 避免在公開頻道中使用 bot
3. **保護 token** — 切勿將 bot token 提交到版本控制
4. **審查審批** — 在接受之前仔細審查審批請求

## 疑難排解

### Bot 無回應

1. 檢查守護程式是否執行：`zen daemon status`
2. 在 Web UI 中驗證 bot 配置
3. 檢查守護程式日誌：`tail -f ~/.zen/zend.log`

### 連接問題

1. 驗證 token 是否正確
2. 檢查網路連接
3. 對於 Slack/Discord，確保啟用了所需的 intents

### 程序未顯示在清單中

1. 確保程序是用 `zen` 啟動的（而不是直接用 `claude`）
2. 檢查 socket 路徑是否與 bot 配置相符
