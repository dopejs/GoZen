---
sidebar_position: 13
title: Webhooks
---

# Webhooks

通过 Slack、Discord 或自定义 webhook 接收预算警报、提供商状态变化和每日摘要的实时通知。

## 功能特性

- **多种格式** — Slack、Discord 或通用 JSON
- **事件过滤** — 订阅特定事件类型
- **自定义头部** — 添加身份验证或自定义头部
- **异步分发** — 非阻塞 webhook 传递
- **自动格式化** — 带有表情符号和颜色的丰富消息
- **测试功能** — 在启用前验证 webhook 配置

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

## 事件类型

| 事件 | 描述 | 触发时机 |
|------|------|----------|
| `budget_warning` | 预算阈值已达到 | 当支出达到限制的 80% 时 |
| `budget_exceeded` | 预算限制已超出 | 当支出超过配置的限制时 |
| `provider_down` | 提供商变为不健康 | 当成功率低于 70% 时 |
| `provider_up` | 提供商恢复 | 当不健康的提供商再次变为健康时 |
| `failover` | 请求故障转移 | 当请求切换到备用提供商时 |
| `daily_summary` | 每日使用摘要 | 每天 UTC 午夜一次 |

## Webhook 格式

### Slack

当 URL 包含 `slack.com` 时自动检测。

**示例消息：**
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

当 URL 包含 `discord.com` 时自动检测。

**示例嵌入：**
- **标题：** budget_warning
- **描述：** ⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
- **颜色：** 琥珀色 (#FBBF24)
- **时间戳：** 2026-03-05T10:30:00Z

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

用于所有其他 URL。

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

## 事件数据结构

### 预算警告 / 超出

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

### 提供商宕机 / 恢复

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

### 故障转移

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

## 平台设置

### Slack

1. 访问 [Slack API](https://api.slack.com/apps)
2. 创建新应用或选择现有应用
3. 启用 "Incoming Webhooks"
4. 将 webhook 添加到工作区
5. 复制 webhook URL（以 `https://hooks.slack.com/` 开头）

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

1. 打开 Discord 服务器设置
2. 进入 Integrations → Webhooks
3. 点击 "New Webhook"
4. 选择频道并复制 webhook URL

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

### 自定义 Webhook

对于自定义集成，使用通用 JSON 格式：

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

在 `http://localhost:19840/settings` 访问 webhook 设置：

1. 导航到 "Webhooks" 标签
2. 点击 "Add Webhook"
3. 输入 webhook URL
4. 选择要订阅的事件
5. （可选）添加自定义头部
6. 点击 "Test" 验证配置
7. 点击 "Save"

## API 端点

### 列出 Webhook

```bash
GET /api/v1/webhooks
```

### 添加 Webhook

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

### 删除 Webhook

```bash
DELETE /api/v1/webhooks/{id}
```

### 测试 Webhook

```bash
POST /api/v1/webhooks/{id}/test
```

发送测试消息以验证配置。

## 消息示例

### 预算警告（Slack）

```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

### 预算超出（Discord）

```
🚫 Budget Exceeded: monthly limit of $200.00 reached (spent: $205.50). Action: block
```

### 提供商宕机（Slack）

```
🔴 Provider Down: anthropic-primary is unhealthy. Error: connection timeout
```

### 提供商恢复（Discord）

```
🟢 Provider Up: anthropic-primary is healthy again (latency: 1250ms)
```

### 故障转移（Slack）

```
🔄 Failover: Switched from anthropic-primary to anthropic-backup. Reason: rate limit exceeded
```

### 每日摘要（Discord）

```
📊 Daily Summary (2026-03-04): 150 requests, $25.50 total cost, 125000 input / 35000 output tokens
```

## 最佳实践

1. **使用单独的 webhook** — 为不同的事件类型创建不同的 webhook
2. **启用前测试** — 保存前始终测试 webhook 配置
3. **保护自定义 webhook** — 使用 HTTPS 和身份验证头部
4. **监控 webhook 失败** — 如果通知停止，检查守护进程日志
5. **避免敏感数据** — 不要在 webhook URL 中包含 API 密钥或令牌
6. **设置警报** — 订阅关键事件，如 `budget_exceeded` 和 `provider_down`

## 故障排除

### Webhook 未接收消息

1. 验证配置中的 webhook 已启用
2. 检查 URL 是否正确（使用 curl 测试）
3. 验证事件配置是否正确
4. 检查守护进程日志中的 webhook 错误：`tail -f ~/.zen/zend.log`
5. 通过 API 测试 webhook：`POST /api/v1/webhooks/{id}/test`

### Slack webhook 失败

1. 验证 webhook URL 以 `https://hooks.slack.com/` 开头
2. 检查 webhook 在 Slack 设置中未被撤销
3. 确保工作区未禁用传入 webhook
4. 使用 curl 测试：
   ```bash
   curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"text":"test"}'
   ```

### Discord webhook 失败

1. 验证 webhook URL 以 `https://discord.com/api/webhooks/` 开头
2. 检查 webhook 在 Discord 设置中未被删除
3. 确保机器人有权限在频道中发布
4. 使用 curl 测试：
   ```bash
   curl -X POST https://discord.com/api/webhooks/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"content":"test"}'
   ```

### 自定义 webhook 不工作

1. 验证端点是否可访问（使用 curl 测试）
2. 检查身份验证头部是否正确
3. 确保端点接受 POST 请求
4. 验证端点返回 2xx 状态码
5. 检查端点日志中的错误

## 安全考虑

1. **保护 webhook URL** — 将 webhook URL 视为机密
2. **使用 HTTPS** — 始终对 webhook 端点使用 HTTPS
3. **验证签名** — 为自定义 webhook 实现签名验证
4. **速率限制** — 在 webhook 端点上实现速率限制
5. **不要记录敏感数据** — 避免记录完整的 webhook 负载

## 高级配置

### 条件 Webhook

将不同的事件发送到不同的 webhook：

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

### 用于身份验证的自定义头部

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
