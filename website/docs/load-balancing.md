---
title: Load Balancing
---

# Load Balancing

GoZen supports several provider selection strategies beyond basic failover. You can choose a strategy per profile, combine it with health checks, and steer traffic based on availability, latency, or cost.

## Available strategies

### Failover

Try providers in order until one succeeds. This is the default strategy and works well for primary/backup setups.

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

### Round robin

Distribute requests evenly across multiple equivalent providers.

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

### Least latency

Prefer the provider with the lowest recent response time.

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

### Least cost

Prefer the cheapest provider for the requested model.

```json
{
  "profiles": {
    "budget": {
      "providers": ["cheap-provider", "premium-provider"],
      "strategy": "least-cost"
    }
  }
}
```

## Health-aware routing

All strategies can work with health monitoring. When `health_aware` is enabled, unhealthy providers are skipped automatically until they recover.

```json
{
  "profiles": {
    "production": {
      "providers": ["primary", "secondary", "tertiary"],
      "strategy": "least-latency",
      "health_aware": true
    }
  }
}
```

## Choosing a strategy

- Use `failover` for reliability-first routing.
- Use `round-robin` when providers are interchangeable.
- Use `least-latency` for interactive or time-sensitive workloads.
- Use `least-cost` when budget matters more than raw speed.

## Related docs

- [Profiles](/docs/profiles) explains how provider groups are defined.
- [Routing](/docs/routing) covers scenario-based provider selection.
- [Health Monitoring](/docs/health-monitoring) explains how health checks affect routing.
