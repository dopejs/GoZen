---
sidebar_position: 4
title: 场景路由
---

# 场景路由

根据请求特征自动路由到不同 provider。

## 支持的场景

| Scenario | Description |
|----------|-------------|
| `think` | 启用 thinking 模式 |
| `image` | 包含图片内容 |
| `longContext` | 内容超过阈值 |
| `webSearch` | 使用 web_search 工具 |
| `background` | 使用 Haiku 模型 |

## Fallback 机制

如果场景配置的 providers 全部失败，会自动 fallback 到 profile 的默认 providers。

## 配置示例

```json
{
  "profiles": {
    "smart": {
      "providers": ["main-api"],
      "long_context_threshold": 60000,
      "routing": {
        "think": {
          "providers": [{"name": "thinking-api", "model": "claude-opus-4-5"}]
        },
        "longContext": {
          "providers": [{"name": "long-context-api"}]
        }
      }
    }
  }
}
```
