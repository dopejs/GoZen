---
sidebar_position: 12
title: 健康监控与负载均衡
---

# 健康监控与负载均衡

实时监控提供商健康状况，并自动将请求路由到最佳可用提供商。

## 功能特性

- **实时健康检查** — 定期健康监控，可配置检查间隔
- **成功率跟踪** — 基于请求成功率计算提供商健康状况
- **延迟监控** — 跟踪每个提供商的平均响应时间
- **多种策略** — 故障转移、轮询、最低延迟、最低成本
- **自动故障转移** — 当主提供商不健康时切换到备用提供商
- **健康仪表板** — Web UI 中的可视化状态指示器

## 配置

### 启用健康监控

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

**选项：**
- `interval` — 检查提供商健康状况的频率（默认：5 分钟）
- `timeout` — 健康检查的请求超时时间（默认：10 秒）
- `endpoint` — 要测试的 API 端点（默认：`/v1/messages`）
- `method` — 健康检查的 HTTP 方法（默认：`POST`）

### 配置负载均衡

```json
{
  "load_balancing": {
    "strategy": "least-latency",
    "health_aware": true,
    "cache_ttl": "30s"
  }
}
```

## 负载均衡策略

### 1. 故障转移（默认）

按顺序使用提供商，失败时切换到下一个。

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

**行为：**
1. 尝试 `anthropic-primary`
2. 如果失败，尝试 `anthropic-backup`
3. 如果失败，尝试 `openai`
4. 如果全部失败，返回错误

**最适合：** 具有明确主/备层次结构的生产工作负载

### 2. 轮询

在所有健康的提供商之间均匀分配请求。

```json
{
  "load_balancing": {
    "strategy": "round-robin"
  }
}
```

**行为：**
- 请求 1 → 提供商 A
- 请求 2 → 提供商 B
- 请求 3 → 提供商 C
- 请求 4 → 提供商 A（循环重复）

**最适合：** 在多个账户之间分配负载以避免速率限制

### 3. 最低延迟

路由到平均延迟最低的提供商。

```json
{
  "load_balancing": {
    "strategy": "least-latency"
  }
}
```

**行为：**
- 跟踪每个提供商的平均响应时间
- 路由到最快的提供商
- 每 30 秒更新一次指标（可通过 `cache_ttl` 配置）

**最适合：** 延迟敏感的应用程序、实时交互

### 4. 最低成本

路由到请求模型的最便宜提供商。

```json
{
  "load_balancing": {
    "strategy": "least-cost"
  }
}
```

**行为：**
- 比较提供商之间的定价
- 路由到最便宜的选项
- 同时考虑输入和输出 token 成本

**最适合：** 成本优化、批量处理

## 健康状态

提供商被分为四种健康状态：

| 状态 | 成功率 | 行为 |
|------|--------|------|
| **健康** | ≥ 95% | 正常优先级 |
| **降级** | 70-95% | 较低优先级，仍可使用 |
| **不健康** | < 70% | 跳过，除非没有健康的提供商 |
| **未知** | 无数据 | 初始时视为健康 |

### 健康感知路由

当 `health_aware: true`（默认）时：
- 健康的提供商优先
- 降级的提供商用作后备
- 不健康的提供商被跳过，除非所有其他提供商都失败

## Web UI 仪表板

在 `http://localhost:19840/health` 访问健康仪表板：

### 提供商状态

- **状态指示器** — 绿色（健康）、黄色（降级）、红色（不健康）
- **成功率** — 成功请求的百分比
- **平均延迟** — 平均响应时间（毫秒）
- **最后检查** — 最近一次健康检查的时间戳
- **错误计数** — 最近失败的次数

### 指标时间线

- **延迟图表** — 响应时间随时间的趋势
- **成功率图表** — 健康状况随时间的趋势
- **请求量** — 每个提供商的请求数

## API 端点

### 获取提供商健康状况

```bash
GET /api/v1/health/providers
```

响应：
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

### 获取提供商指标

```bash
GET /api/v1/health/providers/{name}/metrics?period=1h
```

响应：
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

### 触发手动健康检查

```bash
POST /api/v1/health/check
Content-Type: application/json

{
  "provider": "anthropic-primary"
}
```

## Webhook 通知

当提供商状态变化时接收警报：

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

**事件类型：**
- `provider_down` — 提供商变为不健康
- `provider_up` — 提供商恢复到健康状态
- `failover` — 请求故障转移到备用提供商

## 基于场景的路由

将健康监控与场景路由结合，实现智能请求分发：

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

详见[场景路由](./routing.md)。

## 最佳实践

1. **设置适当的间隔** — 大多数情况下 5 分钟即可，关键系统使用 1 分钟
2. **使用健康感知路由** — 生产工作负载始终启用
3. **监控降级的提供商** — 当成功率低于 95% 时进行调查
4. **组合策略** — 主/备使用故障转移，负载分配使用轮询
5. **启用 webhook** — 提供商宕机时立即收到通知
6. **定期检查仪表板** — 查看健康趋势以识别模式

## 故障排除

### 健康检查失败

1. 验证提供商 API 密钥是否有效
2. 检查到提供商端点的网络连接
3. 如果提供商响应慢，增加超时时间：`"timeout": "30s"`
4. 查看守护进程日志中的具体错误消息

### 延迟指标不正确

1. 延迟包括网络时间 + API 处理时间
2. 检查代理或 VPN 是否增加了开销
3. 指标默认缓存 30 秒（可通过 `cache_ttl` 配置）

### 故障转移不工作

1. 验证负载均衡配置中的 `health_aware: true`
2. 检查配置文件中是否配置了备用提供商
3. 确保健康检查已启用并正在运行
4. 在 Web UI 或日志中查看故障转移事件

### 提供商卡在不健康状态

1. 通过 API 手动触发健康检查
2. 检查提供商是否真的宕机（使用 curl 测试）
3. 重启守护进程以重置健康状态：`zen daemon restart`
4. 查看错误日志以找出根本原因

## 性能影响

- **健康检查** — 最小开销，在后台 goroutine 中运行
- **指标缓存** — 30 秒 TTL 减少数据库查询
- **原子操作** — 并发请求的线程安全计数器
- **无阻塞** — 健康检查不会阻塞请求处理

## 高级配置

### 自定义健康检查负载

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

### 按提供商的健康设置

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
