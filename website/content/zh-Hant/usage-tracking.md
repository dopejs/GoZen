---
sidebar_position: 11
title: 使用追蹤與預算控制
---

# 使用追蹤與預算控制

跨提供商、模型和專案追蹤 token 使用量和成本。設定支出限制並自動執行操作。

## 功能特性

- **即時追蹤** — 監控每個請求的 token 使用量和成本
- **多維度聚合** — 按提供商、模型、專案和時間段追蹤
- **預算限制** — 設定每日、每週和每月支出上限
- **自動操作** — 超出限制時警告、降級或阻止請求
- **成本估算** — 所有主要 AI 模型的準確定價
- **歷史資料** — SQLite 儲存，按小時聚合以提高效能

## 配置

### 啟用使用追蹤

```json
{
  "usage_tracking": {
    "enabled": true,
    "db_path": "~/.zen/usage.db"
  }
}
```

### 配置模型定價

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

**模型匹配**：首先匹配精確的模型名稱，然後回退到模型系列前綴。

### 設定預算限制

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

## 預算操作

| 操作 | 行為 |
|------|------|
| `warn` | 記錄警告並傳送 webhook 通知，但允許請求 |
| `downgrade` | 切換到更便宜的模型（例如 opus → sonnet → haiku） |
| `block` | 以 429 狀態碼拒絕請求 |

## Web UI

在 `http://localhost:19840/usage` 存取使用儀表板：

- **概覽** — 目前週期的總成本、請求和 token
- **按提供商** — 每個提供商的成本細分
- **按模型** — 每個模型的使用統計
- **按專案** — 追蹤每個專案的成本（透過專案繫結）
- **時間線** — 每小時/每日成本趨勢
- **預算狀態** — 每日/每週/每月限制的視覺化指示器

## API 端點

### 取得使用摘要

```bash
GET /api/v1/usage/summary?period=daily
```

回應：
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

### 取得預算狀態

```bash
GET /api/v1/budget/status
```

回應：
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

### 更新預算限制

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

## 專案級追蹤

使用目錄繫結追蹤每個專案的成本：

```bash
# 將目前目錄繫結到設定檔
zen bind work-profile

# 來自此目錄的所有請求都會標記專案路徑
# 在 Web UI 的 "By Project" 下檢視成本
```

## Webhook 通知

當預算超出時接收警報：

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

詳見 [Webhooks](./webhooks.md) 取得完整配置。

## 最佳實務

1. **從警告開始** — 最初使用 `warn` 操作以了解使用模式
2. **設定實際限制** — 基於歷史使用資料設定限制
3. **開發時使用降級** — 測試時自動切換到更便宜的模型
4. **生產環境保留阻止** — 僅對硬性支出上限使用 `block` 操作
5. **每日監控** — 定期檢查使用儀表板以避免意外
6. **啟用 webhook** — 接近限制時取得即時警報

## 疑難排解

### 使用未被追蹤

1. 驗證配置中的 `usage_tracking.enabled` 為 `true`
2. 檢查資料庫路徑是否可寫：`~/.zen/usage.db`
3. 重啟守護程式：`zen daemon restart`

### 成本不正確

1. 驗證配置中的模型定價與目前費率匹配
2. 檢查模型名稱匹配（精確匹配 vs 系列前綴）
3. 如果提供商更改費率，更新定價配置

### 預算未執行

1. 檢查預算配置是否已啟用
2. 驗證操作是否已設定（`warn`、`downgrade` 或 `block`）
3. 檢查守護程式日誌中的預算檢查器錯誤

## 效能

- **按小時聚合** — 原始資料按小時聚合以減少查詢負載
- **索引查詢** — 資料庫索引提供商、模型、專案、時間戳
- **高效儲存** — 每個請求約 1KB，30,000 個請求約 30MB
- **快速儀表板** — 典型使用模式的查詢時間低於一秒
