---
sidebar_position: 10
title: Usage Tracking
---

# Usage Tracking & Budget Control

GoZen v3.0 introduces comprehensive usage tracking and budget control features.

## Overview

- Track token usage and costs per request
- Set budget limits (daily/weekly/monthly)
- Automatic actions when limits are reached
- Per-project cost tracking via project bindings

## Configuration

Add to your `zen.json`:

```json
{
  "pricing": {
    "claude-sonnet-4-20250514": {
      "input_per_million": 3.0,
      "output_per_million": 15.0
    }
  },
  "budgets": {
    "daily": {
      "amount": 10.0,
      "action": "warn"
    },
    "monthly": {
      "amount": 100.0,
      "action": "block"
    }
  }
}
```

## Budget Actions

| Action | Description |
|--------|-------------|
| `warn` | Log a warning but continue |
| `downgrade` | Switch to a cheaper model |
| `block` | Block the request |

## Web UI

Access the Usage dashboard at `http://127.0.0.1:19840` to view:

- Total cost and token usage
- Usage breakdown by provider
- Usage breakdown by model
- Budget status and alerts

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/usage/summary` | GET | Get usage summary |
| `/api/v1/budget` | GET | Get budget config |
| `/api/v1/budget` | PUT | Update budget config |
| `/api/v1/budget/status` | GET | Get current budget status |

## Default Model Pricing

GoZen includes built-in pricing for popular models:

| Model | Input (per 1M) | Output (per 1M) |
|-------|----------------|-----------------|
| claude-sonnet-4-20250514 | $3.00 | $15.00 |
| claude-opus-4-20250514 | $15.00 | $75.00 |
| claude-haiku-3-5-20241022 | $0.80 | $4.00 |
| gpt-4o | $2.50 | $10.00 |
| gpt-4o-mini | $0.15 | $0.60 |

You can override these or add custom models in the `pricing` config.
