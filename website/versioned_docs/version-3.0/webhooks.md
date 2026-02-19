---
sidebar_position: 13
title: Webhooks
---

# Webhook Notifications

Get notified about important events via webhooks.

## Overview

GoZen can send HTTP POST notifications to external services like Slack, Discord, or custom endpoints.

## Configuration

```json
{
  "webhooks": [
    {
      "name": "slack-alerts",
      "url": "https://hooks.slack.com/services/xxx",
      "events": ["budget_warning", "provider_down"],
      "enabled": true
    },
    {
      "name": "discord-notifications",
      "url": "https://discord.com/api/webhooks/xxx",
      "events": ["daily_summary"],
      "enabled": true
    }
  ]
}
```

## Supported Events

| Event | Description |
|-------|-------------|
| `budget_warning` | Budget threshold reached (e.g., 80%) |
| `budget_exceeded` | Budget limit exceeded |
| `provider_down` | Provider became unhealthy |
| `provider_up` | Provider recovered |
| `failover` | Request failed over to backup provider |
| `daily_summary` | Daily usage summary |

## Payload Format

### Generic Format

```json
{
  "event": "budget_warning",
  "timestamp": "2025-02-19T10:30:00Z",
  "data": {
    "budget_type": "daily",
    "limit": 10.0,
    "current": 8.5,
    "percentage": 85
  }
}
```

### Slack Format

GoZen automatically formats messages for Slack webhooks with rich formatting.

### Discord Format

GoZen automatically formats messages for Discord webhooks with embeds.

## Custom Headers

Add custom headers for authentication:

```json
{
  "webhooks": [
    {
      "name": "custom-endpoint",
      "url": "https://api.example.com/webhook",
      "events": ["budget_exceeded"],
      "headers": {
        "Authorization": "Bearer xxx",
        "X-Custom-Header": "value"
      },
      "enabled": true
    }
  ]
}
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/webhooks` | GET | List all webhooks |
| `/api/v1/webhooks` | POST | Create webhook |
| `/api/v1/webhooks/{name}` | PUT | Update webhook |
| `/api/v1/webhooks/{name}` | DELETE | Delete webhook |
| `/api/v1/webhooks/test` | POST | Test a webhook |
