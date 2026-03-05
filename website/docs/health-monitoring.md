---
sidebar_position: 12
title: Health Monitoring & Load Balancing
---

# Health Monitoring & Load Balancing

Monitor provider health in real-time and automatically route requests to the best available provider.

## Features

- **Real-time health checks** — Periodic health monitoring with configurable intervals
- **Success rate tracking** — Calculate provider health based on request success rates
- **Latency monitoring** — Track average response times per provider
- **Multiple strategies** — Failover, round-robin, least-latency, least-cost
- **Automatic failover** — Switch to backup providers when primary is unhealthy
- **Health dashboard** — Visual status indicators in Web UI

## Configuration

### Enable Health Monitoring

```json
{
  "health_check": {
    "enabled": true,
    "interval": "5m",
    "timeout": "10s",
    "endpoint": "/v1/messages",
    "method": "POST"
  }
}
```

**Options:**
- `interval` — How often to check provider health (default: 5 minutes)
- `timeout` — Request timeout for health checks (default: 10 seconds)
- `endpoint` — API endpoint to test (default: `/v1/messages`)
- `method` — HTTP method for health check (default: `POST`)

### Configure Load Balancing

```json
{
  "load_balancing": {
    "strategy": "least-latency",
    "health_aware": true,
    "cache_ttl": "30s"
  }
}
```

## Load Balancing Strategies

### 1. Failover (Default)

Use providers in order, switch to next on failure.

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-primary", "anthropic-backup", "openai"],
      "load_balancing": {
        "strategy": "failover"
      }
    }
  }
}
```

**Behavior:**
1. Try `anthropic-primary`
2. If fails, try `anthropic-backup`
3. If fails, try `openai`
4. If all fail, return error

**Best for:** Production workloads with clear primary/backup hierarchy

### 2. Round-Robin

Distribute requests evenly across all healthy providers.

```json
{
  "load_balancing": {
    "strategy": "round-robin"
  }
}
```

**Behavior:**
- Request 1 → Provider A
- Request 2 → Provider B
- Request 3 → Provider C
- Request 4 → Provider A (cycle repeats)

**Best for:** Distributing load across multiple accounts to avoid rate limits

### 3. Least-Latency

Route to the provider with lowest average latency.

```json
{
  "load_balancing": {
    "strategy": "least-latency"
  }
}
```

**Behavior:**
- Tracks average response time per provider
- Routes to fastest provider
- Updates metrics every 30 seconds (configurable via `cache_ttl`)

**Best for:** Latency-sensitive applications, real-time interactions

### 4. Least-Cost

Route to the cheapest provider for the requested model.

```json
{
  "load_balancing": {
    "strategy": "least-cost"
  }
}
```

**Behavior:**
- Compares pricing across providers
- Routes to cheapest option
- Considers both input and output token costs

**Best for:** Cost optimization, batch processing

## Health Status

Providers are classified into four health states:

| Status | Success Rate | Behavior |
|--------|--------------|----------|
| **Healthy** | ≥ 95% | Normal priority |
| **Degraded** | 70-95% | Lower priority, still usable |
| **Unhealthy** | < 70% | Skipped unless no healthy providers |
| **Unknown** | No data | Treated as healthy initially |

### Health-Aware Routing

When `health_aware: true` (default):
- Healthy providers are prioritized
- Degraded providers used as fallback
- Unhealthy providers skipped unless all others fail

## Web UI Dashboard

Access health dashboard at `http://localhost:19840/health`:

### Provider Status

- **Status indicator** — Green (healthy), yellow (degraded), red (unhealthy)
- **Success rate** — Percentage of successful requests
- **Average latency** — Mean response time in milliseconds
- **Last check** — Timestamp of most recent health check
- **Error count** — Number of recent failures

### Metrics Timeline

- **Latency graph** — Response time trends over time
- **Success rate graph** — Health trends over time
- **Request volume** — Requests per provider

## API Endpoints

### Get Provider Health

```bash
GET /api/v1/health/providers
```

Response:
```json
{
  "providers": [
    {
      "name": "anthropic-primary",
      "status": "healthy",
      "success_rate": 98.5,
      "avg_latency_ms": 1250,
      "last_check": "2026-03-05T10:30:00Z",
      "error_count": 2,
      "total_requests": 150
    },
    {
      "name": "openai-backup",
      "status": "degraded",
      "success_rate": 85.0,
      "avg_latency_ms": 2100,
      "last_check": "2026-03-05T10:29:00Z",
      "error_count": 15,
      "total_requests": 100
    }
  ]
}
```

### Get Provider Metrics

```bash
GET /api/v1/health/providers/{name}/metrics?period=1h
```

Response:
```json
{
  "provider": "anthropic-primary",
  "period": "1h",
  "metrics": [
    {
      "timestamp": "2026-03-05T10:00:00Z",
      "latency_ms": 1200,
      "success_rate": 99.0,
      "requests": 25
    },
    {
      "timestamp": "2026-03-05T10:05:00Z",
      "latency_ms": 1300,
      "success_rate": 98.0,
      "requests": 28
    }
  ]
}
```

### Trigger Manual Health Check

```bash
POST /api/v1/health/check
Content-Type: application/json

{
  "provider": "anthropic-primary"
}
```

## Webhook Notifications

Receive alerts when provider status changes:

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
      "events": ["provider_down", "provider_up", "failover"]
    }
  ]
}
```

**Event types:**
- `provider_down` — Provider becomes unhealthy
- `provider_up` — Provider recovers to healthy state
- `failover` — Request failed over to backup provider

## Scenario-Based Routing

Combine health monitoring with scenario routing for intelligent request distribution:

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-primary", "anthropic-backup"],
      "scenarios": {
        "thinking": {
          "providers": ["anthropic-thinking"],
          "load_balancing": {
            "strategy": "least-latency"
          }
        },
        "image": {
          "providers": ["anthropic-vision", "openai-vision"],
          "load_balancing": {
            "strategy": "failover"
          }
        }
      }
    }
  }
}
```

See [Scenario Routing](./routing.md) for details.

## Best Practices

1. **Set appropriate intervals** — 5 minutes is good for most cases, 1 minute for critical systems
2. **Use health-aware routing** — Always enable for production workloads
3. **Monitor degraded providers** — Investigate when success rate drops below 95%
4. **Combine strategies** — Use failover for primary/backup, round-robin for load distribution
5. **Enable webhooks** — Get notified immediately when providers go down
6. **Check dashboard regularly** — Review health trends to identify patterns

## Troubleshooting

### Health checks failing

1. Verify provider API keys are valid
2. Check network connectivity to provider endpoints
3. Increase timeout if providers are slow: `"timeout": "30s"`
4. Review daemon logs for specific error messages

### Incorrect latency metrics

1. Latency includes network time + API processing time
2. Check if proxy or VPN is adding overhead
3. Metrics are cached for 30 seconds by default (configurable via `cache_ttl`)

### Failover not working

1. Verify `health_aware: true` in load balancing config
2. Check that backup providers are configured in profile
3. Ensure health checks are enabled and running
4. Review failover events in Web UI or logs

### Provider stuck in unhealthy state

1. Manually trigger health check via API
2. Check if provider is actually down (test with curl)
3. Restart daemon to reset health state: `zen daemon restart`
4. Review error logs for root cause

## Performance Impact

- **Health checks** — Minimal overhead, runs in background goroutine
- **Metrics caching** — 30-second TTL reduces database queries
- **Atomic operations** — Thread-safe counters for concurrent requests
- **No blocking** — Health checks don't block request processing

## Advanced Configuration

### Custom Health Check Payload

```json
{
  "health_check": {
    "enabled": true,
    "custom_payload": {
      "model": "claude-3-haiku-20240307",
      "max_tokens": 10,
      "messages": [
        {
          "role": "user",
          "content": "ping"
        }
      ]
    }
  }
}
```

### Per-Provider Health Settings

```json
{
  "providers": {
    "anthropic-primary": {
      "health_check": {
        "interval": "1m",
        "timeout": "5s"
      }
    },
    "openai-backup": {
      "health_check": {
        "interval": "5m",
        "timeout": "10s"
      }
    }
  }
}
```
