---
sidebar_position: 12
title: Load Balancing
---

# Smart Load Balancing

GoZen v3.0 supports multiple load balancing strategies beyond simple failover.

## Strategies

### Failover (Default)

Try providers in order until one succeeds.

```json
{
  "profiles": {
    "default": {
      "providers": ["primary", "backup"],
      "strategy": "failover"
    }
  }
}
```

### Round Robin

Distribute requests evenly across providers.

```json
{
  "profiles": {
    "balanced": {
      "providers": ["provider-a", "provider-b", "provider-c"],
      "strategy": "round-robin"
    }
  }
}
```

### Least Latency

Route to the provider with the lowest recent latency.

```json
{
  "profiles": {
    "fast": {
      "providers": ["us-east", "us-west", "eu"],
      "strategy": "least-latency"
    }
  }
}
```

### Least Cost

Route to the cheapest provider for the requested model.

```json
{
  "profiles": {
    "budget": {
      "providers": ["cheap-provider", "expensive-provider"],
      "strategy": "least-cost"
    }
  }
}
```

## Strategy Comparison

| Strategy | Best For |
|----------|----------|
| `failover` | High availability, primary/backup setup |
| `round-robin` | Even distribution, multiple equivalent providers |
| `least-latency` | Performance-critical applications |
| `least-cost` | Cost optimization |

## Health-Aware Routing

All strategies automatically skip unhealthy providers. If a provider is marked as unhealthy by the health checker, it will be excluded from selection until it recovers.
