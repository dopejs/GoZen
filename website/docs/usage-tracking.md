---
sidebar_position: 11
title: Usage Tracking & Budget Control
---

# Usage Tracking & Budget Control

Track token usage and costs across providers, models, and projects. Set spending limits with automatic enforcement actions.

## Features

- **Real-time tracking** — Monitor token usage and costs per request
- **Multi-dimensional aggregation** — Track by provider, model, project, and time period
- **Budget limits** — Set daily, weekly, and monthly spending caps
- **Automatic actions** — Warn, downgrade, or block requests when limits are exceeded
- **Cost estimation** — Accurate pricing for all major AI models
- **Historical data** — SQLite storage with hourly aggregation for performance

## Configuration

### Enable Usage Tracking

```json
{
  "usage_tracking": {
    "enabled": true,
    "db_path": "~/.zen/usage.db"
  }
}
```

### Configure Model Pricing

```json
{
  "pricing": {
    "models": {
      "claude-opus-4": {
        "input_per_mtok": 15.0,
        "output_per_mtok": 75.0
      },
      "claude-sonnet-4": {
        "input_per_mtok": 3.0,
        "output_per_mtok": 15.0
      },
      "gpt-4o": {
        "input_per_mtok": 2.5,
        "output_per_mtok": 10.0
      }
    },
    "model_families": {
      "claude-opus": {
        "input_per_mtok": 15.0,
        "output_per_mtok": 75.0
      },
      "claude-sonnet": {
        "input_per_mtok": 3.0,
        "output_per_mtok": 15.0
      },
      "gpt-4": {
        "input_per_mtok": 2.5,
        "output_per_mtok": 10.0
      }
    }
  }
}
```

**Model matching**: Exact model names are matched first, then falls back to model family prefixes.

### Set Budget Limits

```json
{
  "budget": {
    "daily": {
      "enabled": true,
      "limit": 10.0,
      "action": "warn"
    },
    "weekly": {
      "enabled": true,
      "limit": 50.0,
      "action": "downgrade"
    },
    "monthly": {
      "enabled": true,
      "limit": 200.0,
      "action": "block"
    }
  }
}
```

## Budget Actions

| Action | Behavior |
|--------|----------|
| `warn` | Log warning and send webhook notification, but allow request |
| `downgrade` | Switch to cheaper model (e.g., opus → sonnet → haiku) |
| `block` | Reject request with 429 status code |

## Web UI

Access usage dashboard at `http://localhost:19840/usage`:

- **Overview** — Total cost, requests, and tokens for current period
- **By Provider** — Cost breakdown per provider
- **By Model** — Usage statistics per model
- **By Project** — Track costs per project (via project bindings)
- **Timeline** — Hourly/daily cost trends
- **Budget Status** — Visual indicators for daily/weekly/monthly limits

## API Endpoints

### Get Usage Summary

```bash
GET /api/v1/usage/summary?period=daily
```

Response:
```json
{
  "period": "daily",
  "start": "2026-03-05T00:00:00Z",
  "end": "2026-03-05T23:59:59Z",
  "total_cost": 8.45,
  "total_requests": 42,
  "total_input_tokens": 125000,
  "total_output_tokens": 35000,
  "by_provider": {
    "anthropic": 6.20,
    "openai": 2.25
  },
  "by_model": {
    "claude-sonnet-4": 5.10,
    "claude-opus-4": 1.10,
    "gpt-4o": 2.25
  }
}
```

### Get Budget Status

```bash
GET /api/v1/budget/status
```

Response:
```json
{
  "daily": {
    "enabled": true,
    "limit": 10.0,
    "spent": 8.45,
    "percent": 84.5,
    "action": "warn",
    "exceeded": false
  },
  "weekly": {
    "enabled": true,
    "limit": 50.0,
    "spent": 32.10,
    "percent": 64.2,
    "action": "downgrade",
    "exceeded": false
  },
  "monthly": {
    "enabled": true,
    "limit": 200.0,
    "spent": 145.80,
    "percent": 72.9,
    "action": "block",
    "exceeded": false
  }
}
```

### Update Budget Limits

```bash
PUT /api/v1/budget/limits
Content-Type: application/json

{
  "daily": {
    "enabled": true,
    "limit": 15.0,
    "action": "warn"
  }
}
```

## Project-Level Tracking

Track costs per project using directory bindings:

```bash
# Bind current directory to a profile
zen bind work-profile

# All requests from this directory are tagged with the project path
# View costs in Web UI under "By Project"
```

## Webhook Notifications

Receive alerts when budgets are exceeded:

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
      "events": ["budget_warning", "budget_exceeded"]
    }
  ]
}
```

See [Webhooks](./webhooks.md) for full configuration.

## Best Practices

1. **Start with warnings** — Use `warn` action initially to understand usage patterns
2. **Set realistic limits** — Base limits on historical usage data
3. **Use downgrade for development** — Automatically switch to cheaper models when testing
4. **Reserve block for production** — Use `block` action only for hard spending caps
5. **Monitor daily** — Check usage dashboard regularly to avoid surprises
6. **Enable webhooks** — Get real-time alerts when approaching limits

## Troubleshooting

### Usage not tracked

1. Verify `usage_tracking.enabled` is `true` in config
2. Check database path is writable: `~/.zen/usage.db`
3. Restart daemon: `zen daemon restart`

### Incorrect costs

1. Verify model pricing in config matches current rates
2. Check model name matching (exact match vs family prefix)
3. Update pricing configuration if providers change rates

### Budget not enforced

1. Check budget configuration is enabled
2. Verify action is set (`warn`, `downgrade`, or `block`)
3. Check daemon logs for budget checker errors

## Performance

- **Hourly aggregation** — Raw data aggregated hourly to reduce query load
- **Indexed queries** — Database indexes on provider, model, project, timestamp
- **Efficient storage** — ~1KB per request, ~30MB per 30,000 requests
- **Fast dashboard** — Sub-second query times for typical usage patterns
