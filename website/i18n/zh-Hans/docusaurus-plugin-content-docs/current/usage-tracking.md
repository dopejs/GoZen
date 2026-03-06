---
sidebar_position: 11
title: 使用跟踪与预算控制
---

# 使用跟踪与预算控制

跨提供商、模型和项目跟踪 token 使用量和成本。设置支出限制并自动执行操作。

## 功能特性

- **实时跟踪** — 监控每个请求的 token 使用量和成本
- **多维度聚合** — 按提供商、模型、项目和时间段跟踪
- **预算限制** — 设置每日、每周和每月支出上限
- **自动操作** — 超出限制时警告、降级或阻止请求
- **成本估算** — 所有主要 AI 模型的准确定价
- **历史数据** — SQLite 存储，按小时聚合以提高性能

## 配置

### 启用使用跟踪

```json
{
  "usage_tracking": {
    "enabled": true,
    "db_path": "~/.zen/usage.db"
  }
}
```

### 配置模型定价

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

**模型匹配**：首先匹配精确的模型名称，然后回退到模型系列前缀。

### 设置预算限制

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

## 预算操作

| 操作 | 行为 |
|------|------|
| `warn` | 记录警告并发送 webhook 通知，但允许请求 |
| `downgrade` | 切换到更便宜的模型（例如 opus → sonnet → haiku） |
| `block` | 以 429 状态码拒绝请求 |

## Web UI

在 `http://localhost:19840/usage` 访问使用仪表板：

- **概览** — 当前周期的总成本、请求和 token
- **按提供商** — 每个提供商的成本细分
- **按模型** — 每个模型的使用统计
- **按项目** — 跟踪每个项目的成本（通过项目绑定）
- **时间线** — 每小时/每日成本趋势
- **预算状态** — 每日/每周/每月限制的可视化指示器

## API 端点

### 获取使用摘要

```bash
GET /api/v1/usage/summary?period=daily
```

响应：
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

### 获取预算状态

```bash
GET /api/v1/budget/status
```

响应：
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

### 更新预算限制

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

## 项目级跟踪

使用目录绑定跟踪每个项目的成本：

```bash
# 将当前目录绑定到配置文件
zen bind work-profile

# 来自此目录的所有请求都会标记项目路径
# 在 Web UI 的 "By Project" 下查看成本
```

## Webhook 通知

当预算超出时接收警报：

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

详见 [Webhooks](./webhooks.md) 获取完整配置。

## 最佳实践

1. **从警告开始** — 最初使用 `warn` 操作以了解使用模式
2. **设置实际限制** — 基于历史使用数据设置限制
3. **开发时使用降级** — 测试时自动切换到更便宜的模型
4. **生产环境保留阻止** — 仅对硬性支出上限使用 `block` 操作
5. **每日监控** — 定期检查使用仪表板以避免意外
6. **启用 webhook** — 接近限制时获得实时警报

## 故障排除

### 使用未被跟踪

1. 验证配置中的 `usage_tracking.enabled` 为 `true`
2. 检查数据库路径是否可写：`~/.zen/usage.db`
3. 重启守护进程：`zen daemon restart`

### 成本不正确

1. 验证配置中的模型定价与当前费率匹配
2. 检查模型名称匹配（精确匹配 vs 系列前缀）
3. 如果提供商更改费率，更新定价配置

### 预算未执行

1. 检查预算配置是否已启用
2. 验证操作是否已设置（`warn`、`downgrade` 或 `block`）
3. 检查守护进程日志中的预算检查器错误

## 性能

- **按小时聚合** — 原始数据按小时聚合以减少查询负载
- **索引查询** — 数据库索引提供商、模型、项目、时间戳
- **高效存储** — 每个请求约 1KB，30,000 个请求约 30MB
- **快速仪表板** — 典型使用模式的查询时间低于一秒
