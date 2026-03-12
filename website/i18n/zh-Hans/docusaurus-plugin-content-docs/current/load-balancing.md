---
title: 负载均衡
---

# 负载均衡

除了基础的 failover 之外，GoZen 还支持多种 provider 选择策略。你可以为每个 profile 选择不同策略，并结合健康检查，根据可用性、延迟或成本来分配流量。

## 可用策略

### Failover

按顺序尝试 provider，直到有一个成功。这是默认策略，适合主备场景。

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

把请求均匀分发到多个等价 provider。

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

优先选择最近响应时间最低的 provider。

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

优先选择所请求模型下成本最低的 provider。

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

## 健康感知路由

所有策略都可以与健康监控一起工作。启用 `health_aware` 后，不健康的 provider 会在恢复之前被自动跳过。

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

## 策略选择建议

- 优先可靠性时使用 `failover`
- provider 可互换时使用 `round-robin`
- 交互式或时延敏感场景使用 `least-latency`
- 预算比速度更重要时使用 `least-cost`

## 相关文档

- [Profiles](/docs/profiles) 介绍如何定义 provider 组。
- [Routing](/docs/routing) 介绍基于场景的 provider 选择。
- [健康监控](/docs/health-monitoring) 说明健康检查如何影响路由。
