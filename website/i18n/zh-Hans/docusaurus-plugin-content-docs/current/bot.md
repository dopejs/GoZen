---
sidebar_position: 10
title: Bot 网关
---

# Bot 网关

通过聊天平台远程监控和控制您的 Claude Code 会话。Bot 通过 IPC 连接到运行中的 `zen` 进程，让您可以：

- 查看已连接的进程及其状态
- 向特定进程发送任务
- 接收审批、错误和完成通知
- 控制任务（暂停/恢复/取消）

## 支持的平台

| 平台 | 所需设置 |
|----------|----------------|
| [Telegram](#telegram) | BotFather token |
| [Discord](#discord) | Bot 应用程序 token |
| [Slack](#slack) | Bot + App tokens (Socket Mode) |
| [Lark/飞书](#lark飞书) | App ID + Secret |
| [Facebook Messenger](#facebook-messenger) | Page token + Verify token |

## 基本配置

```json
{
  "bot": {
    "enabled": true,
    "socket_path": "/tmp/zen-bot.sock",
    "platforms": {
      // 平台特定配置（见下文）
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

## Bot 命令

| 命令 | 描述 |
|---------|-------------|
| `list` | 列出所有已连接的进程 |
| `status [name]` | 显示进程状态 |
| `bind <name>` | 绑定到一个进程以执行后续命令 |
| `pause [name]` | 暂停当前任务 |
| `resume [name]` | 恢复暂停的任务 |
| `cancel [name]` | 取消当前任务 |
| `<name> <task>` | 向进程发送任务 |
| `help` | 显示可用命令 |

### 自然语言支持

Bot 理解多种语言的自然语言查询：

- "show me the status of gozen"
- "帮我看看 gozen 的状态"
- "list all processes"
- "pause the api project"

## 交互模式

### 私聊消息

设置 `direct_message_mode` 来控制 bot 在私聊中的响应方式：

- `"always"` — 始终响应（无需提及）
- `"mention"` — 仅在被提及时响应

### 频道消息

设置 `channel_mode` 来控制群聊中的行为：

- `"always"` — 响应所有消息
- `"mention"` — 仅在被提及时响应（推荐）

### 提及关键词

配置触发 bot 的关键词：

```json
{
  "interaction": {
    "require_mention": true,
    "mention_keywords": ["@zen", "/zen", "zen"]
  }
}
```

## 项目别名

为您的项目定义简短名称：

```json
{
  "aliases": {
    "api": "/Users/john/projects/api-server",
    "web": "/Users/john/projects/web-app",
    "backend": "/Users/john/work/backend"
  }
}
```

然后在命令中使用它们：

```
api run tests
web build production
status backend
```

## 平台设置

### Telegram

1. 通过 [@BotFather](https://t.me/botfather) 创建 bot：
   - 发送 `/newbot` 并按照提示操作
   - 复制 token（例如：`123456789:ABCdefGHIjklMNOpqrsTUVwxyz`）

2. 获取您的用户 ID：
   - 向 [@userinfobot](https://t.me/userinfobot) 发送消息
   - 复制您的数字用户 ID

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

**安全选项：**
- `allowed_users` — 可以与 bot 交互的用户名或用户 ID
- `allowed_chats` — bot 响应的群聊 ID（通过 [@getidsbot](https://t.me/getidsbot) 获取）

### Discord

1. 创建 Discord 应用程序：
   - 访问 [Discord 开发者门户](https://discord.com/developers/applications)
   - 点击 "New Application" 并命名
   - 进入 "Bot" 部分并点击 "Add Bot"
   - 复制 token

2. 启用所需的 intents：
   - 在 Bot 部分，启用 "Message Content Intent"
   - 如果使用用户过滤，启用 "Server Members Intent"

3. 邀请 bot 到您的服务器：
   - 进入 OAuth2 → URL Generator
   - 选择 scopes：`bot`
   - 选择权限：`Send Messages`、`Read Message History`
   - 使用生成的 URL 邀请

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

**获取 ID：** 在 Discord 设置中启用开发者模式，然后右键点击用户/频道/服务器以复制 ID。

### Slack

1. 创建 Slack 应用：
   - 访问 [Slack API](https://api.slack.com/apps)
   - 点击 "Create New App" → "From scratch"
   - 命名您的应用并选择工作区

2. 启用 Socket Mode：
   - 进入 "Socket Mode" 并启用
   - 生成具有 `connections:write` 范围的应用级 Token
   - 复制 token（以 `xapp-` 开头）

3. 配置 Bot Token：
   - 进入 "OAuth & Permissions"
   - 添加 scopes：`chat:write`、`channels:history`、`groups:history`、`im:history`、`mpim:history`
   - 安装到工作区并复制 Bot Token（以 `xoxb-` 开头）

4. 启用事件：
   - 进入 "Event Subscriptions" 并启用
   - 订阅：`message.channels`、`message.groups`、`message.im`、`message.mpim`

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

### Lark/飞书

1. 创建 Lark 应用：
   - 访问 [Lark 开放平台](https://open.larksuite.com/) 或 [飞书开放平台](https://open.feishu.cn/)
   - 创建新应用
   - 复制 App ID 和 App Secret

2. 配置权限：
   - 添加 `im:message:receive_v1` 事件
   - 添加 `im:message:send_v1` 权限

3. 配置 webhook：
   - 设置事件订阅 URL（或使用 WebSocket 模式）

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

1. 创建 Facebook 应用：
   - 访问 [Facebook 开发者](https://developers.facebook.com/)
   - 创建类型为 "Business" 的新应用
   - 添加 "Messenger" 产品

2. 配置 Messenger：
   - 生成 Page Access Token
   - 使用 verify token 设置 webhook
   - 订阅 `messages` 事件

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

**注意：** Facebook Messenger 需要公开可访问的 webhook URL。开发时可考虑使用 ngrok 等服务。

## 通知

配置 bot 发送通知的位置：

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

### 通知类型

- **审批请求** — 当 Claude Code 需要操作权限时
- **任务完成** — 当任务成功完成时
- **错误** — 当任务失败或遇到错误时
- **状态变更** — 当进程连接/断开时

### 免打扰时间

在免打扰时间内，非紧急通知会被抑制。审批请求始终会发送。

## 安全最佳实践

1. **限制用户** — 始终配置 `allowed_users` 以限制谁可以控制您的会话
2. **使用私有频道** — 避免在公共频道中使用 bot
3. **保护 token** — 切勿将 bot token 提交到版本控制
4. **审查审批** — 在接受之前仔细审查审批请求

## 故障排除

### Bot 无响应

1. 检查守护进程是否运行：`zen daemon status`
2. 在 Web UI 中验证 bot 配置
3. 检查守护进程日志：`tail -f ~/.zen/zend.log`

### 连接问题

1. 验证 token 是否正确
2. 检查网络连接
3. 对于 Slack/Discord，确保启用了所需的 intents

### 进程未显示在列表中

1. 确保进程是用 `zen` 启动的（而不是直接用 `claude`）
2. 检查 socket 路径是否与 bot 配置匹配
