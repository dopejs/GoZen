---
sidebar_position: 12
title: 健康監控與負載平衡
---

# 健康監控與負載平衡

即時監控提供商健康狀況，並自動將請求路由到最佳可用提供商。

## 功能特性

- **即時健康檢查** — 定期健康監控，可配置檢查間隔
- **成功率追蹤** — 基於請求成功率計算提供商健康狀況
- **延遲監控** — 追蹤每個提供商的平均回應時間
- **多種策略** — 故障轉移、輪詢、最低延遲、最低成本
- **自動故障轉移** — 當主提供商不健康時切換到備用提供商
- **健康儀表板** — Web UI 中的視覺化狀態指示器

## 配置

### 啟用健康監控

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

**選項：**
- `interval` — 檢查提供商健康狀況的頻率（預設：5 分鐘）
- `timeout` — 健康檢查的請求逾時時間（預設：10 秒）
- `endpoint` — 要測試的 API 端點（預設：`/v1/messages`）
- `method` — 健康檢查的 HTTP 方法（預設：`POST`）

### 配置負載平衡

```json
{
  "load_balancing": {
    "strategy": "least-latency",
    "health_aware": true,
    "cache_ttl": "30s"
  }
}
```

## 負載平衡策略

### 1. 故障轉移（預設）

按順序使用提供商，失敗時切換到下一個。

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

**行為：**
1. 嘗試 `anthropic-primary`
2. 如果失敗，嘗試 `anthropic-backup`
3. 如果失敗，嘗試 `openai`
4. 如果全部失敗，傳回錯誤

**最適合：** 具有明確主/備層次結構的生產工作負載

### 2. 輪詢

在所有健康的提供商之間均勻分配請求。

```json
{
  "load_balancing": {
    "strategy": "round-robin"
  }
}
```

**行為：**
- 請求 1 → 提供商 A
- 請求 2 → 提供商 B
- 請求 3 → 提供商 C
- 請求 4 → 提供商 A（循環重複）

**最適合：** 在多個帳戶之間分配負載以避免速率限制

### 3. 最低延遲

路由到平均延遲最低的提供商。

```json
{
  "load_balancing": {
    "strategy": "least-latency"
  }
}
```

**行為：**
- 追蹤每個提供商的平均回應時間
- 路由到最快的提供商
- 每 30 秒更新一次指標（可透過 `cache_ttl` 配置）

**最適合：** 延遲敏感的應用程式、即時互動

### 4. 最低成本

路由到請求模型的最便宜提供商。

```json
{
  "load_balancing": {
    "strategy": "least-cost"
  }
}
```

**行為：**
- 比較提供商之間的定價
- 路由到最便宜的選項
- 同時考慮輸入和輸出 token 成本

**最適合：** 成本最佳化、批次處理

## 健康狀態

提供商被分為四種健康狀態：

| 狀態 | 成功率 | 行為 |
|------|--------|------|
| **健康** | ≥ 95% | 正常優先順序 |
| **降級** | 70-95% | 較低優先順序，仍可使用 |
| **不健康** | < 70% | 跳過，除非沒有健康的提供商 |
| **未知** | 無資料 | 初始時視為健康 |

### 健康感知路由

當 `health_aware: true`（預設）時：
- 健康的提供商優先
- 降級的提供商用作後備
- 不健康的提供商被跳過，除非所有其他提供商都失敗

## Web UI 儀表板

在 `http://localhost:19840/health` 存取健康儀表板：

### 提供商狀態

- **狀態指示器** — 綠色（健康）、黃色（降級）、紅色（不健康）
- **成功率** — 成功請求的百分比
- **平均延遲** — 平均回應時間（毫秒）
- **最後檢查** — 最近一次健康檢查的時間戳
- **錯誤計數** — 最近失敗的次數

### 指標時間線

- **延遲圖表** — 回應時間隨時間的趨勢
- **成功率圖表** — 健康狀況隨時間的趨勢
- **請求量** — 每個提供商的請求數

## API 端點

### 取得提供商健康狀況

```bash
GET /api/v1/health/providers
```

回應：
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

### 取得提供商指標

```bash
GET /api/v1/health/providers/{name}/metrics?period=1h
```

回應：
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

### 觸發手動健康檢查

```bash
POST /api/v1/health/check
Content-Type: application/json

{
  "provider": "anthropic-primary"
}
```

## Webhook 通知

當提供商狀態變化時接收警報：

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

**事件類型：**
- `provider_down` — 提供商變為不健康
- `provider_up` — 提供商恢復到健康狀態
- `failover` — 請求故障轉移到備用提供商

## 基於場景的路由

將健康監控與場景路由結合，實現智慧請求分發：

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

詳見[場景路由](./routing.md)。

## 最佳實務

1. **設定適當的間隔** — 大多數情況下 5 分鐘即可，關鍵系統使用 1 分鐘
2. **使用健康感知路由** — 生產工作負載始終啟用
3. **監控降級的提供商** — 當成功率低於 95% 時進行調查
4. **組合策略** — 主/備使用故障轉移，負載分配使用輪詢
5. **啟用 webhook** — 提供商宕機時立即收到通知
6. **定期檢查儀表板** — 檢視健康趨勢以識別模式

## 疑難排解

### 健康檢查失敗

1. 驗證提供商 API 金鑰是否有效
2. 檢查到提供商端點的網路連接
3. 如果提供商回應慢，增加逾時時間：`"timeout": "30s"`
4. 查看守護程式日誌中的具體錯誤訊息

### 延遲指標不正確

1. 延遲包括網路時間 + API 處理時間
2. 檢查代理或 VPN 是否增加了開銷
3. 指標預設快取 30 秒（可透過 `cache_ttl` 配置）

### 故障轉移不運作

1. 驗證負載平衡配置中的 `health_aware: true`
2. 檢查設定檔中是否配置了備用提供商
3. 確保健康檢查已啟用並正在執行
4. 在 Web UI 或日誌中檢視故障轉移事件

### 提供商卡在不健康狀態

1. 透過 API 手動觸發健康檢查
2. 檢查提供商是否真的宕機（使用 curl 測試）
3. 重啟守護程式以重設健康狀態：`zen daemon restart`
4. 查看錯誤日誌以找出根本原因

## 效能影響

- **健康檢查** — 最小開銷，在背景 goroutine 中執行
- **指標快取** — 30 秒 TTL 減少資料庫查詢
- **原子操作** — 並行請求的執行緒安全計數器
- **無阻塞** — 健康檢查不會阻塞請求處理

## 進階配置

### 自訂健康檢查負載

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

### 按提供商的健康設定

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
