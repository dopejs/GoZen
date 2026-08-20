---
title: 負載平衡
---

# 負載平衡

除了基礎的 failover 之外，GoZen 也支援多種 provider 選擇策略。你可以為每個 profile 指定不同策略，並結合健康檢查，根據可用性、延遲或成本來分配流量。

## 可用策略

### Failover

依序嘗試 provider，直到其中一個成功。這是預設策略，適合主備情境。

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

將請求平均分散到多個等價 provider。

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

優先選擇最近回應時間最低的 provider。

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

優先選擇所請求模型下成本最低的 provider。

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

所有策略都可以與健康監控一起運作。啟用 `health_aware` 後，不健康的 provider 會在恢復之前自動被跳過。

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

## 策略選擇建議

- 優先可靠性時使用 `failover`
- provider 可互換時使用 `round-robin`
- 互動式或對延遲敏感的場景使用 `least-latency`
- 預算比速度更重要時使用 `least-cost`

## 相關文件

- [Profiles](/docs/profiles) 介紹如何定義 provider 群組。
- [Routing](/docs/routing) 介紹基於情境的 provider 選擇。
- [健康監控](/docs/health-monitoring) 說明健康檢查如何影響路由。
