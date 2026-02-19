---
sidebar_position: 11
title: Health Monitoring
---

# Provider Health Monitoring

Monitor the health and performance of your API providers in real-time.

## Overview

- Automatic health checks for all providers
- Latency and error rate tracking
- Rate limit detection
- Real-time status dashboard

## Configuration

```json
{
  "health_check": {
    "enabled": true,
    "interval_secs": 60,
    "timeout_secs": 10
  }
}
```

## Health Status

Each provider can have one of these statuses:

| Status | Description |
|--------|-------------|
| `healthy` | Provider is responding normally |
| `degraded` | High latency or occasional errors |
| `unhealthy` | Provider is down or rate limited |
| `unknown` | No recent health data |

## Metrics Tracked

- **Latency**: Average response time (ms)
- **Error Rate**: Percentage of failed requests
- **Rate Limit**: Whether provider is rate limiting
- **Uptime**: Availability over time

## Web UI

The Providers page shows real-time health indicators:

- 🟢 Green: Healthy
- 🟡 Yellow: Degraded
- 🔴 Red: Unhealthy

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/health/providers` | GET | Get all provider health status |
| `/api/v1/health/providers/{name}` | GET | Get specific provider health |

## Smart Failover

When a provider becomes unhealthy, GoZen automatically:

1. Detects the failure
2. Routes requests to the next healthy provider
3. Continues monitoring the failed provider
4. Resumes using it when it recovers
