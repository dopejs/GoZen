---
sidebar_position: 10
title: Botゲートウェイ
---

# Botゲートウェイ

チャットプラットフォームを通じてClaude Codeセッションをリモート監視・制御できます。BotはIPCを介して実行中の`zen`プロセスに接続し、以下が可能です：

- 接続されたプロセスとそのステータスの表示
- 特定のプロセスへのタスク送信
- 承認、エラー、完了通知の受信
- タスク制御（一時停止/再開/キャンセル）

## サポートされているプラットフォーム

| プラットフォーム | 必要な設定 |
|----------|----------------|
| [Telegram](#telegram) | BotFather token |
| [Discord](#discord) | Bot アプリケーション token |
| [Slack](#slack) | Bot + App tokens (Socket Mode) |
| [Lark/飛書](#lark飛書) | App ID + Secret |
| [Facebook Messenger](#facebook-messenger) | Page token + Verify token |

## 基本設定

```json
{
  "bot": {
    "enabled": true,
    "socket_path": "/tmp/zen-bot.sock",
    "platforms": {
      // プラットフォーム固有の設定（以下を参照）
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

## Botコマンド

| コマンド | 説明 |
|---------|-------------|
| `list` | すべての接続されたプロセスをリスト表示 |
| `status [name]` | プロセスステータスを表示 |
| `bind <name>` | 後続のコマンドを実行するプロセスにバインド |
| `pause [name]` | 現在のタスクを一時停止 |
| `resume [name]` | 一時停止したタスクを再開 |
| `cancel [name]` | 現在のタスクをキャンセル |
| `<name> <task>` | プロセスにタスクを送信 |
| `help` | 利用可能なコマンドを表示 |

### 自然言語サポート

Botは複数の言語での自然言語クエリを理解します：

- "show me the status of gozen"
- "gozenのステータスを見せて"
- "list all processes"
- "pause the api project"

## インタラクションモード

### ダイレクトメッセージ

`direct_message_mode`を設定してダイレクトメッセージでのBot応答方法を制御：

- `"always"` — 常に応答（メンション不要）
- `"mention"` — メンションされた時のみ応答

### チャンネルメッセージ

`channel_mode`を設定してグループチャットでの動作を制御：

- `"always"` — すべてのメッセージに応答
- `"mention"` — メンションされた時のみ応答（推奨）

### メンションキーワード

Botをトリガーするキーワードを設定：

```json
{
  "interaction": {
    "require_mention": true,
    "mention_keywords": ["@zen", "/zen", "zen"]
  }
}
```

## プロジェクトエイリアス

プロジェクトの短縮名を定義：

```json
{
  "aliases": {
    "api": "/Users/john/projects/api-server",
    "web": "/Users/john/projects/web-app",
    "backend": "/Users/john/work/backend"
  }
}
```

コマンドで使用：

```
api run tests
web build production
status backend
```

## プラットフォーム設定

### Telegram

1. [@BotFather](https://t.me/botfather)でBotを作成：
   - `/newbot`を送信してプロンプトに従う
   - tokenをコピー（例：`123456789:ABCdefGHIjklMNOpqrsTUVwxyz`）

2. ユーザーIDを取得：
   - [@userinfobot](https://t.me/userinfobot)にメッセージを送信
   - 数字のユーザーIDをコピー

3. 設定：

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

**セキュリティオプション：**
- `allowed_users` — Botと対話できるユーザー名またはユーザーID
- `allowed_chats` — Botが応答するグループチャットID（[@getidsbot](https://t.me/getidsbot)で取得）

### Discord

1. Discordアプリケーションを作成：
   - [Discord Developer Portal](https://discord.com/developers/applications)にアクセス
   - "New Application"をクリックして名前を付ける
   - "Bot"セクションに移動して"Add Bot"をクリック
   - tokenをコピー

2. 必要なintentsを有効化：
   - Botセクションで"Message Content Intent"を有効化
   - ユーザーフィルタリングを使用する場合は"Server Members Intent"を有効化

3. Botをサーバーに招待：
   - OAuth2 → URL Generatorに移動
   - scopesを選択：`bot`
   - 権限を選択：`Send Messages`、`Read Message History`
   - 生成されたURLを使用して招待

4. 設定：

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

**IDの取得：** Discord設定で開発者モードを有効にし、ユーザー/チャンネル/サーバーを右クリックしてIDをコピー。

### Slack

1. Slackアプリを作成：
   - [Slack API](https://api.slack.com/apps)にアクセス
   - "Create New App" → "From scratch"をクリック
   - アプリに名前を付けてワークスペースを選択

2. Socket Modeを有効化：
   - "Socket Mode"に移動して有効化
   - `connections:write`スコープでアプリレベルTokenを生成
   - tokenをコピー（`xapp-`で始まる）

3. Bot Tokenを設定：
   - "OAuth & Permissions"に移動
   - scopesを追加：`chat:write`、`channels:history`、`groups:history`、`im:history`、`mpim:history`
   - ワークスペースにインストールしてBot Tokenをコピー（`xoxb-`で始まる）

4. イベントを有効化：
   - "Event Subscriptions"に移動して有効化
   - 購読：`message.channels`、`message.groups`、`message.im`、`message.mpim`

5. 設定：

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

1. Larkアプリを作成：
   - [Lark Open Platform](https://open.larksuite.com/)または[飛書開放平台](https://open.feishu.cn/)にアクセス
   - 新しいアプリを作成
   - App IDとApp Secretをコピー

2. 権限を設定：
   - `im:message:receive_v1`イベントを追加
   - `im:message:send_v1`権限を追加

3. webhookを設定：
   - イベント購読URLを設定（またはWebSocketモードを使用）

4. 設定：

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

1. Facebookアプリを作成：
   - [Facebook Developers](https://developers.facebook.com/)にアクセス
   - タイプ"Business"の新しいアプリを作成
   - "Messenger"製品を追加

2. Messengerを設定：
   - Page Access Tokenを生成
   - verify tokenでwebhookを設定
   - `messages`イベントを購読

3. 設定：

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

**注意：** Facebook Messengerには公開アクセス可能なwebhook URLが必要です。開発時はngrokなどのサービスの使用を検討してください。

## 通知

Botが通知を送信する場所を設定：

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

### 通知タイプ

- **承認リクエスト** — Claude Codeが操作の許可を必要とする時
- **タスク完了** — タスクが正常に完了した時
- **エラー** — タスクが失敗またはエラーに遭遇した時
- **ステータス変更** — プロセスが接続/切断された時

### 静寂時間

静寂時間中は、緊急でない通知が抑制されます。承認リクエストは常に送信されます。

## セキュリティベストプラクティス

1. **ユーザーを制限** — 常に`allowed_users`を設定してセッションを制御できるユーザーを制限
2. **プライベートチャンネルを使用** — 公開チャンネルでのBot使用を避ける
3. **tokenを保護** — Bot tokenをバージョン管理にコミットしない
4. **承認をレビュー** — 承認リクエストを受け入れる前に慎重にレビュー

## トラブルシューティング

### Botが応答しない

1. デーモンが実行中か確認：`zen daemon status`
2. Web UIでBot設定を確認
3. デーモンログを確認：`tail -f ~/.zen/zend.log`

### 接続の問題

1. tokenが正しいことを確認
2. ネットワーク接続を確認
3. Slack/Discordの場合、必要なintentsが有効化されていることを確認

### プロセスがリストに表示されない

1. プロセスが`zen`で起動されていることを確認（`claude`で直接起動していない）
2. socketパスがBot設定と一致していることを確認
