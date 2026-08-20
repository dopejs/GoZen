---
sidebar_position: 13
title: Webhooks
---

# Webhooks

透過 Slack、Discord 或自訂 webhook 接收預算警報、提供商狀態變化和每日摘要的即時通知。

## 功能特性

- **多種格式** — Slack、Discord 或通用 JSON
- **事件過濾** — 訂閱特定事件類型
- **自訂標頭** — 新增身份驗證或自訂標頭
- **非同步分發** — 非阻塞 webhook 傳遞
- **自動格式化** — 帶有表情符號和顏色的豐富訊息
- **測試功能** — 在啟用前驗證 webhook 配置

## 配置

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

## 事件類型

| 事件 | 描述 | 觸發時機 |
|------|------|----------|
| `budget_warning` | 預算閾值已達到 | 當支出達到限制的 80% 時 |
| `budget_exceeded` | 預算限制已超出 | 當支出超過配置的限制時 |
| `provider_down` | 提供商變為不健康 | 當成功率低於 70% 時 |
| `provider_up` | 提供商恢復 | 當不健康的提供商再次變為健康時 |
| `failover` | 請求故障轉移 | 當請求切換到備用提供商時 |
| `daily_summary` | 每日使用摘要 | 每天 UTC 午夜一次 |

## Webhook 格式

### Slack

當 URL 包含 `slack.com` 時自動偵測。

**範例訊息：**
```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

**格式：**
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

當 URL 包含 `discord.com` 時自動偵測。

**範例嵌入：**
- **標題：** budget_warning
- **描述：** ⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
- **顏色：** 琥珀色 (#FBBF24)
- **時間戳：** 2026-03-05T10:30:00Z

**格式：**
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

### 通用 JSON

用於所有其他 URL。

**格式：**
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

## 事件資料結構

### 預算警告 / 超出

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

### 提供商宕機 / 恢復

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

### 故障轉移

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

### 每日摘要

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

## 平台設定

### Slack

1. 造訪 [Slack API](https://api.slack.com/apps)
2. 建立新應用程式或選擇現有應用程式
3. 啟用 "Incoming Webhooks"
4. 將 webhook 新增到工作區
5. 複製 webhook URL（以 `https://hooks.slack.com/` 開頭）

**配置：**
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

1. 開啟 Discord 伺服器設定
2. 進入 Integrations → Webhooks
3. 點選 "New Webhook"
4. 選擇頻道並複製 webhook URL

**配置：**
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

### 自訂 Webhook

對於自訂整合，使用通用 JSON 格式：

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

## Web UI 配置

在 `http://localhost:19840/settings` 存取 webhook 設定：

1. 導覽到 "Webhooks" 標籤
2. 點選 "Add Webhook"
3. 輸入 webhook URL
4. 選擇要訂閱的事件
5. （可選）新增自訂標頭
6. 點選 "Test" 驗證配置
7. 點選 "Save"

## API 端點

### 列出 Webhook

```bash
GET /api/v1/webhooks
```

### 新增 Webhook

```bash
POST /api/v1/webhooks
Content-Type: application/json

{
  "enabled": true,
  "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
  "events": ["budget_warning", "provider_down"]
}
```

### 更新 Webhook

```bash
PUT /api/v1/webhooks/{id}
Content-Type: application/json

{
  "enabled": false
}
```

### 刪除 Webhook

```bash
DELETE /api/v1/webhooks/{id}
```

### 測試 Webhook

```bash
POST /api/v1/webhooks/{id}/test
```

傳送測試訊息以驗證配置。

## 訊息範例

### 預算警告（Slack）

```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

### 預算超出（Discord）

```
🚫 Budget Exceeded: monthly limit of $200.00 reached (spent: $205.50). Action: block
```

### 提供商宕機（Slack）

```
🔴 Provider Down: anthropic-primary is unhealthy. Error: connection timeout
```

### 提供商恢復（Discord）

```
🟢 Provider Up: anthropic-primary is healthy again (latency: 1250ms)
```

### 故障轉移（Slack）

```
🔄 Failover: Switched from anthropic-primary to anthropic-backup. Reason: rate limit exceeded
```

### 每日摘要（Discord）

```
📊 Daily Summary (2026-03-04): 150 requests, $25.50 total cost, 125000 input / 35000 output tokens
```

## 最佳實務

1. **使用單獨的 webhook** — 為不同的事件類型建立不同的 webhook
2. **啟用前測試** — 儲存前始終測試 webhook 配置
3. **保護自訂 webhook** — 使用 HTTPS 和身份驗證標頭
4. **監控 webhook 失敗** — 如果通知停止，檢查守護程式日誌
5. **避免敏感資料** — 不要在 webhook URL 中包含 API 金鑰或權杖
6. **設定警報** — 訂閱關鍵事件，如 `budget_exceeded` 和 `provider_down`

## 疑難排解

### Webhook 未接收訊息

1. 驗證配置中的 webhook 已啟用
2. 檢查 URL 是否正確（使用 curl 測試）
3. 驗證事件配置是否正確
4. 檢查守護程式日誌中的 webhook 錯誤：`tail -f ~/.zen/zend.log`
5. 透過 API 測試 webhook：`POST /api/v1/webhooks/{id}/test`

### Slack webhook 失敗

1. 驗證 webhook URL 以 `https://hooks.slack.com/` 開頭
2. 檢查 webhook 在 Slack 設定中未被撤銷
3. 確保工作區未停用傳入 webhook
4. 使用 curl 測試：
   ```bash
   curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"text":"test"}'
   ```

### Discord webhook 失敗

1. 驗證 webhook URL 以 `https://discord.com/api/webhooks/` 開頭
2. 檢查 webhook 在 Discord 設定中未被刪除
3. 確保機器人有權限在頻道中發布
4. 使用 curl 測試：
   ```bash
   curl -X POST https://discord.com/api/webhooks/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"content":"test"}'
   ```

### 自訂 webhook 不工作

1. 驗證端點是否可存取（使用 curl 測試）
2. 檢查身份驗證標頭是否正確
3. 確保端點接受 POST 請求
4. 驗證端點傳回 2xx 狀態碼
5. 檢查端點日誌中的錯誤

## 安全考量

1. **保護 webhook URL** — 將 webhook URL 視為機密
2. **使用 HTTPS** — 始終對 webhook 端點使用 HTTPS
3. **驗證簽章** — 為自訂 webhook 實作簽章驗證
4. **速率限制** — 在 webhook 端點上實作速率限制
5. **不要記錄敏感資料** — 避免記錄完整的 webhook 負載

## 進階配置

### 條件 Webhook

將不同的事件傳送到不同的 webhook：

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

### 用於身份驗證的自訂標頭

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
