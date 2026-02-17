---
sidebar_position: 4
title: 場景路由
---

# 場景路由

根據請求特徵自動路由到不同 provider。

## 支援的場景

| Scenario | Description |
|----------|-------------|
| `think` | 啟用 thinking 模式 |
| `image` | 包含圖片內容 |
| `longContext` | 內容超過閾值 |
| `webSearch` | 使用 web_search 工具 |
| `background` | 使用 Haiku 模型 |

## Fallback 機制

如果場景設定的 providers 全部失敗，會自動 fallback 到 profile 的預設 providers。

## 設定範例

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
