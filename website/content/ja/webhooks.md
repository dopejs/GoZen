---
sidebar_position: 13
title: Webhook
---

# Webhook

Slack、Discord、またはカスタム Webhook を通じて、予算アラート、プロバイダーステータス変更、日次サマリーのリアルタイム通知を受信します。

## 機能

- **複数のフォーマット** — Slack、Discord、または汎用JSON
- **イベントフィルタリング** — 特定のイベントタイプを購読
- **カスタムヘッダー** — 認証またはカスタムヘッダーを追加
- **非同期配信** — ノンブロッキングな Webhook 配信
- **自動フォーマット** — 絵文字と色を使用したリッチメッセージ
- **テスト機能** — 有効化前に Webhook 設定を検証

## 設定

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
      "events": [
        "budget_warning",
        "budget_exceeded",
        "provider_down",
        "provider_up",
        "failover",
        "daily_summary"
      ],
      "headers": {
        "Authorization": "Bearer YOUR_TOKEN"
      }
    }
  ]
}
```

## イベントタイプ

| イベント | 説明 | トリガー時 |
|------|------|----------|
| `budget_warning` | 予算閾値に到達 | 支出が制限の80%に達した時 |
| `budget_exceeded` | 予算制限を超過 | 支出が設定された制限を超えた時 |
| `provider_down` | プロバイダーが不健全になった | 成功率が70%を下回った時 |
| `provider_up` | プロバイダーが回復 | 不健全なプロバイダーが再び健全になった時 |
| `failover` | リクエストがフェイルオーバー | リクエストがバックアッププロバイダーに切り替わった時 |
| `daily_summary` | 日次使用状況サマリー | 毎日UTC午前0時に1回 |

## Webhookフォーマット

### Slack

URLに`slack.com`が含まれる場合、自動的に検出されます。

**メッセージ例：**
```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

**フォーマット：**
```json
{
  "text": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)",
  "blocks": [
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)"
      }
    }
  ]
}
```

### Discord

URLに`discord.com`が含まれる場合、自動的に検出されます。

**埋め込み例：**
- **タイトル：** budget_warning
- **説明：** ⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
- **色：** アンバー (#FBBF24)
- **タイムスタンプ：** 2026-03-05T10:30:00Z

**フォーマット：**
```json
{
  "content": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)",
  "embeds": [
    {
      "title": "budget_warning",
      "description": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)",
      "timestamp": "2026-03-05T10:30:00Z",
      "color": 16432932
    }
  ]
}
```

### 汎用JSON

その他すべてのURLに使用されます。

**フォーマット：**
```json
{
  "event": "budget_warning",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "period": "daily",
    "spent": 8.5,
    "limit": 10.0,
    "percent": 85.0,
    "project": ""
  }
}
```

## イベントデータ構造

### 予算警告 / 超過

```json
{
  "event": "budget_warning",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "period": "daily",
    "spent": 8.5,
    "limit": 10.0,
    "percent": 85.0,
    "action": "warn",
    "project": "my-project"
  }
}
```

### プロバイダーダウン / アップ

```json
{
  "event": "provider_down",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "provider": "anthropic-primary",
    "status": "unhealthy",
    "error": "connection timeout",
    "latency_ms": 0
  }
}
```

### フェイルオーバー

```json
{
  "event": "failover",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "from_provider": "anthropic-primary",
    "to_provider": "anthropic-backup",
    "reason": "rate limit exceeded",
    "session_id": "sess_abc123"
  }
}
```

### 日次サマリー

```json
{
  "event": "daily_summary",
  "timestamp": "2026-03-05T00:00:00Z",
  "data": {
    "date": "2026-03-04",
    "total_cost": 25.50,
    "total_requests": 150,
    "total_input_tokens": 125000,
    "total_output_tokens": 35000,
    "by_provider": {
      "anthropic": 18.20,
      "openai": 7.30
    }
  }
}
```

## プラットフォーム設定

### Slack

1. [Slack API](https://api.slack.com/apps)にアクセス
2. 新しいアプリを作成または既存のアプリを選択
3. "Incoming Webhooks"を有効化
4. ワークスペースに Webhook を追加
5. Webhook URLをコピー（`https://hooks.slack.com/`で始まる）

**設定：**
```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
      "events": ["budget_warning", "provider_down"]
    }
  ]
}
```

### Discord

1. Discordサーバー設定を開く
2. Integrations → Webhooksに移動
3. "New Webhook"をクリック
4. チャンネルを選択して Webhook URLをコピー

**設定：**
```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://discord.com/api/webhooks/123456789/XXXXXXXXXXXXXXXXXXXX",
      "events": ["budget_exceeded", "failover"]
    }
  ]
}
```

### カスタムWebhook

カスタム統合には汎用JSONフォーマットを使用：

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://your-server.com/webhook",
      "events": ["budget_warning", "daily_summary"],
      "headers": {
        "Authorization": "Bearer YOUR_SECRET_TOKEN",
        "X-Custom-Header": "value"
      }
    }
  ]
}
```

## Web UI 設定

`http://localhost:19840/settings`で Webhook 設定にアクセス：

1. "Webhooks"タブに移動
2. "Add Webhook"をクリック
3. Webhook URLを入力
4. 購読するイベントを選択
5. （オプション）カスタムヘッダーを追加
6. "Test"をクリックして設定を検証
7. "Save"をクリック

## APIエンドポイント

### Webhookのリスト表示

```bash
GET /api/v1/webhooks
```

### Webhookの追加

```bash
POST /api/v1/webhooks
Content-Type: application/json

{
  "enabled": true,
  "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
  "events": ["budget_warning", "provider_down"]
}
```

### Webhookの更新

```bash
PUT /api/v1/webhooks/{id}
Content-Type: application/json

{
  "enabled": false
}
```

### Webhookの削除

```bash
DELETE /api/v1/webhooks/{id}
```

### Webhookのテスト

```bash
POST /api/v1/webhooks/{id}/test
```

設定を検証するためのテストメッセージを送信します。

## メッセージ例

### 予算警告（Slack）

```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

### 予算超過（Discord）

```
🚫 Budget Exceeded: monthly limit of $200.00 reached (spent: $205.50). Action: block
```

### プロバイダーダウン（Slack）

```
🔴 Provider Down: anthropic-primary is unhealthy. Error: connection timeout
```

### プロバイダーアップ（Discord）

```
🟢 Provider Up: anthropic-primary is healthy again (latency: 1250ms)
```

### フェイルオーバー（Slack）

```
🔄 Failover: Switched from anthropic-primary to anthropic-backup. Reason: rate limit exceeded
```

### 日次サマリー（Discord）

```
📊 Daily Summary (2026-03-04): 150 requests, $25.50 total cost, 125000 input / 35000 output tokens
```

## ベストプラクティス

1. **個別の Webhook を使用** — 異なるイベントタイプに対して異なる Webhook を作成
2. **有効化前にテスト** — 保存前に常に Webhook 設定をテスト
3. **カスタム Webhook を保護** — HTTPSと認証ヘッダーを使用
4. **Webhook 失敗を監視** — 通知が停止した場合、デーモンログを確認
5. **機密データを避ける** — Webhook URLにAPIキーやトークンを含めない
6. **アラートを設定** — `budget_exceeded`や`provider_down`などの重要なイベントを購読

## トラブルシューティング

### Webhookがメッセージを受信しない

1. 設定で Webhook が有効化されていることを確認
2. URLが正しいことを確認（curlでテスト）
3. イベント設定が正しいことを確認
4. デーモンログで Webhook エラーを確認：`tail -f ~/.zen/zend.log`
5. API 経由で Webhook をテスト：`POST /api/v1/webhooks/{id}/test`

### Slack Webhook が失敗する

1. Webhook URLが`https://hooks.slack.com/`で始まることを確認
2. Slack設定で Webhook が取り消されていないか確認
3. ワークスペースが受信 Webhook を無効化していないことを確認
4. curlでテスト：
   ```bash
   curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"text":"test"}'
   ```

### Discord Webhook が失敗する

1. Webhook URLが`https://discord.com/api/webhooks/`で始まることを確認
2. Discord設定で Webhook が削除されていないか確認
3. Botがチャンネルに投稿する権限を持っていることを確認
4. curlでテスト：
   ```bash
   curl -X POST https://discord.com/api/webhooks/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"content":"test"}'
   ```

### カスタム Webhook が機能しない

1. エンドポイントがアクセス可能であることを確認（curlでテスト）
2. 認証ヘッダーが正しいことを確認
3. エンドポイントがPOSTリクエストを受け入れることを確認
4. エンドポイントが2xxステータスコードを返すことを確認
5. エンドポイントログでエラーを確認

## セキュリティ考慮事項

1. **Webhook URLを保護** — Webhook URLを機密情報として扱う
2. **HTTPSを使用** — Webhook エンドポイントには常にHTTPSを使用
3. **署名を検証** — カスタム Webhook に署名検証を実装
4. **レート制限** — Webhook エンドポイントにレート制限を実装
5. **機密データをログに記録しない** — 完全な Webhook ペイロードのログ記録を避ける

## 高度な設定

### 条件付きWebhook

異なるイベントを異なる Webhook に送信：

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/CRITICAL/ALERTS",
      "events": ["budget_exceeded", "provider_down"]
    },
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/DAILY/REPORTS",
      "events": ["daily_summary"]
    },
    {
      "enabled": true,
      "url": "https://discord.com/api/webhooks/MONITORING",
      "events": ["failover", "provider_up"]
    }
  ]
}
```

### 認証用のカスタムヘッダー

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://your-server.com/webhook",
      "events": ["budget_warning"],
      "headers": {
        "Authorization": "Bearer YOUR_SECRET_TOKEN",
        "X-API-Key": "your-api-key",
        "X-Webhook-Source": "gozen"
      }
    }
  ]
}
```
