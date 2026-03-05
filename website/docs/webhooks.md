---
sidebar_position: 13
title: Webhooks
---

# Webhooks

Receive real-time notifications for budget alerts, provider status changes, and daily summaries via Slack, Discord, or custom webhooks.

## Features

- **Multiple formats** — Slack, Discord, or generic JSON
- **Event filtering** — Subscribe to specific event types
- **Custom headers** — Add authentication or custom headers
- **Async dispatch** — Non-blocking webhook delivery
- **Automatic formatting** — Rich messages with emojis and colors
- **Test functionality** — Verify webhook configuration before enabling

## Configuration

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

## Event Types

| Event | Description | When Triggered |
|-------|-------------|----------------|
| `budget_warning` | Budget threshold reached | When spending reaches 80% of limit |
| `budget_exceeded` | Budget limit exceeded | When spending exceeds configured limit |
| `provider_down` | Provider becomes unhealthy | When success rate drops below 70% |
| `provider_up` | Provider recovers | When unhealthy provider becomes healthy again |
| `failover` | Request failed over | When request switches to backup provider |
| `daily_summary` | Daily usage summary | Once per day at midnight UTC |

## Webhook Formats

### Slack

Automatically detected when URL contains `slack.com`.

**Example message:**
```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

**Format:**
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

Automatically detected when URL contains `discord.com`.

**Example embed:**
- **Title:** budget_warning
- **Description:** ⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
- **Color:** Amber (#FBBF24)
- **Timestamp:** 2026-03-05T10:30:00Z

**Format:**
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

### Generic JSON

Used for all other URLs.

**Format:**
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

## Event Data Structures

### Budget Warning / Exceeded

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

### Provider Down / Up

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

### Failover

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

### Daily Summary

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

## Platform Setup

### Slack

1. Go to [Slack API](https://api.slack.com/apps)
2. Create a new app or select existing
3. Enable "Incoming Webhooks"
4. Add webhook to workspace
5. Copy webhook URL (starts with `https://hooks.slack.com/`)

**Configuration:**
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

1. Open Discord server settings
2. Go to Integrations → Webhooks
3. Click "New Webhook"
4. Select channel and copy webhook URL

**Configuration:**
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

### Custom Webhook

For custom integrations, use generic JSON format:

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

## Web UI Configuration

Access webhook settings at `http://localhost:19840/settings`:

1. Navigate to "Webhooks" tab
2. Click "Add Webhook"
3. Enter webhook URL
4. Select events to subscribe
5. (Optional) Add custom headers
6. Click "Test" to verify configuration
7. Click "Save"

## API Endpoints

### List Webhooks

```bash
GET /api/v1/webhooks
```

### Add Webhook

```bash
POST /api/v1/webhooks
Content-Type: application/json

{
  "enabled": true,
  "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
  "events": ["budget_warning", "provider_down"]
}
```

### Update Webhook

```bash
PUT /api/v1/webhooks/{id}
Content-Type: application/json

{
  "enabled": false
}
```

### Delete Webhook

```bash
DELETE /api/v1/webhooks/{id}
```

### Test Webhook

```bash
POST /api/v1/webhooks/{id}/test
```

Sends a test message to verify configuration.

## Message Examples

### Budget Warning (Slack)

```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

### Budget Exceeded (Discord)

```
🚫 Budget Exceeded: monthly limit of $200.00 reached (spent: $205.50). Action: block
```

### Provider Down (Slack)

```
🔴 Provider Down: anthropic-primary is unhealthy. Error: connection timeout
```

### Provider Up (Discord)

```
🟢 Provider Up: anthropic-primary is healthy again (latency: 1250ms)
```

### Failover (Slack)

```
🔄 Failover: Switched from anthropic-primary to anthropic-backup. Reason: rate limit exceeded
```

### Daily Summary (Discord)

```
📊 Daily Summary (2026-03-04): 150 requests, $25.50 total cost, 125000 input / 35000 output tokens
```

## Best Practices

1. **Use separate webhooks** — Create different webhooks for different event types
2. **Test before enabling** — Always test webhook configuration before saving
3. **Secure custom webhooks** — Use HTTPS and authentication headers
4. **Monitor webhook failures** — Check daemon logs if notifications stop
5. **Avoid sensitive data** — Don't include API keys or tokens in webhook URLs
6. **Set up alerts** — Subscribe to critical events like `budget_exceeded` and `provider_down`

## Troubleshooting

### Webhook not receiving messages

1. Verify webhook is enabled in configuration
2. Check URL is correct (test with curl)
3. Verify events are configured correctly
4. Check daemon logs for webhook errors: `tail -f ~/.zen/zend.log`
5. Test webhook via API: `POST /api/v1/webhooks/{id}/test`

### Slack webhook failing

1. Verify webhook URL starts with `https://hooks.slack.com/`
2. Check webhook is not revoked in Slack settings
3. Ensure workspace has not disabled incoming webhooks
4. Test with curl:
   ```bash
   curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"text":"test"}'
   ```

### Discord webhook failing

1. Verify webhook URL starts with `https://discord.com/api/webhooks/`
2. Check webhook is not deleted in Discord settings
3. Ensure bot has permission to post in channel
4. Test with curl:
   ```bash
   curl -X POST https://discord.com/api/webhooks/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"content":"test"}'
   ```

### Custom webhook not working

1. Verify endpoint is accessible (test with curl)
2. Check authentication headers are correct
3. Ensure endpoint accepts POST requests
4. Verify endpoint returns 2xx status code
5. Check endpoint logs for errors

## Security Considerations

1. **Protect webhook URLs** — Treat webhook URLs as secrets
2. **Use HTTPS** — Always use HTTPS for webhook endpoints
3. **Validate signatures** — Implement signature validation for custom webhooks
4. **Rate limiting** — Implement rate limiting on webhook endpoints
5. **Don't log sensitive data** — Avoid logging full webhook payloads

## Advanced Configuration

### Conditional Webhooks

Send different events to different webhooks:

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

### Custom Headers for Authentication

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
