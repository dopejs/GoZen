---
sidebar_position: 14
title: 上下文壓縮 (BETA)
---

# 上下文壓縮 (BETA)

:::warning BETA 功能
上下文壓縮目前處於測試階段。預設情況下已停用，需要明確配置才能啟用。
:::

當 token 數量超過閾值時自動壓縮對話上下文，在保持對話品質的同時降低成本。

## 功能特性

- **自動壓縮** — 當 token 數量超過閾值時觸發
- **智慧摘要** — 使用廉價模型（claude-3-haiku）總結舊訊息
- **保留最近訊息** — 保持最近的訊息完整以保證上下文連續性
- **Token 估算** — 在 API 呼叫前準確計算 token 數量
- **統計追蹤** — 監控壓縮效果
- **透明操作** — 與所有 AI 客戶端無縫協作

## 工作原理

1. **Token 估算** — 計算對話歷史中的 token 數量
2. **閾值檢查** — 與配置的閾值比較（預設：50,000）
3. **訊息選擇** — 識別需要壓縮的舊訊息
4. **摘要產生** — 使用廉價模型建立簡潔摘要
5. **上下文替換** — 用摘要替換舊訊息
6. **請求轉發** — 將壓縮後的上下文傳送到目標模型

## 配置

### 啟用壓縮

```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 50000,
    "target_tokens": 20000,
    "summarizer_model": "claude-3-haiku-20240307",
    "preserve_recent_messages": 5,
    "tokens_per_char": 0.25
  }
}
```

**選項：**

| 選項 | 預設值 | 描述 |
|------|--------|------|
| `enabled` | `false` | 啟用上下文壓縮 |
| `threshold_tokens` | `50000` | 當上下文超過此值時觸發壓縮 |
| `target_tokens` | `20000` | 壓縮後的目標 token 數量 |
| `summarizer_model` | `claude-3-haiku-20240307` | 用於摘要的模型 |
| `preserve_recent_messages` | `5` | 保持完整的最近訊息數量 |
| `tokens_per_char` | `0.25` | Token 計數的估算比率 |

### 按設定檔配置

為特定設定檔啟用壓縮：

```json
{
  "profiles": {
    "long-context": {
      "providers": ["anthropic"],
      "compression": {
        "enabled": true,
        "threshold_tokens": 100000,
        "target_tokens": 40000
      }
    },
    "short-context": {
      "providers": ["openai"],
      "compression": {
        "enabled": false
      }
    }
  }
}
```

## Token 估算

GoZen 使用基於字元的估算進行快速 token 計數：

```
estimated_tokens = character_count * tokens_per_char
```

**預設比率：** 每字元 0.25 個 token（1 個 token ≈ 4 個字元）

**準確度：** 英文文字 ±10%，其他語言可能有所不同

對於精確的 token 計數，GoZen 在可用時使用 `tiktoken-go` 函式庫。

## 壓縮策略

### 訊息選擇

1. **系統訊息** — 始終保留
2. **最近訊息** — 保留最後 N 條訊息（預設：5）
3. **舊訊息** — 壓縮候選

### 摘要提示

```
簡潔地總結以下對話歷史，同時保留關鍵資訊、決策和上下文：

[舊訊息]

提供一個捕捉要點的簡短摘要。
```

### 結果

```
原始：45,000 tokens（30 條訊息）
壓縮後：22,000 tokens（摘要 + 5 條最近訊息）
節省：23,000 tokens（51%）
```

## Web UI

在 `http://localhost:19840/settings` 存取壓縮設定：

1. 導覽到 "Compression" 標籤（標有 BETA 徽章）
2. 切換 "Enable Compression"
3. 調整閾值和目標 token 數量
4. 選擇摘要模型
5. 設定要保留的最近訊息數量
6. 點選 "Save"

### 統計儀表板

檢視壓縮統計：

- **總壓縮次數** — 觸發壓縮的次數
- **節省的 Token** — 所有壓縮中節省的總 token 數
- **平均節省** — 每次壓縮的平均 token 減少量
- **壓縮率** — 觸發壓縮的請求百分比

## API 端點

### 取得壓縮統計

```bash
GET /api/v1/compression/stats
```

回應：
```json
{
  "enabled": true,
  "total_compressions": 42,
  "tokens_saved": 1250000,
  "average_savings": 29761,
  "compression_rate": 0.15,
  "last_compression": "2026-03-05T10:30:00Z"
}
```

### 更新壓縮設定

```bash
PUT /api/v1/compression/settings
Content-Type: application/json

{
  "enabled": true,
  "threshold_tokens": 60000,
  "target_tokens": 25000
}
```

### 重設統計

```bash
POST /api/v1/compression/stats/reset
```

## 使用場景

### 長時間編碼會話

**場景：** 使用 Claude Code 進行多小時編碼會話

**配置：**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 80000,
    "target_tokens": 30000,
    "preserve_recent_messages": 10
  }
}
```

**優勢：** 在不觸及上下文限制的情況下保持對話連續性

### 批次處理

**場景：** 使用 AI 處理多個文件

**配置：**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 40000,
    "target_tokens": 15000,
    "preserve_recent_messages": 3
  }
}
```

**優勢：** 在處理大型文件集時降低成本

### 研究與分析

**場景：** 涉及多個主題的長時間研究會話

**配置：**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 100000,
    "target_tokens": 40000,
    "preserve_recent_messages": 8
  }
}
```

**優勢：** 保持對話專注於最近的主題，同時保留早期上下文

## 最佳實務

1. **從預設值開始** — 預設設定適用於大多數使用場景
2. **監控統計** — 定期檢查壓縮率和節省情況
3. **調整閾值** — 對於長上下文模型（Claude Opus）增加，對於短上下文減少
4. **保留足夠的訊息** — 保留 5-10 條最近訊息以保證上下文連續性
5. **使用廉價摘要器** — Haiku 快速且成本效益高，適合摘要
6. **生產前測試** — 使用您的特定用例驗證壓縮品質

## 限制

1. **品質損失** — 摘要可能會遺失細微的細節
2. **延遲增加** — 增加摘要 API 呼叫開銷
3. **成本權衡** — 摘要成本 vs. token 節省
4. **語言支援** — 最適合英語，其他語言可能有所不同
5. **上下文視窗** — 不能超過模型的最大上下文視窗

## 疑難排解

### 壓縮未觸發

1. 驗證 `compression.enabled` 為 `true`
2. 檢查 token 數量是否超過閾值
3. 確保對話有足夠的訊息可壓縮
4. 查看守護程式日誌中的壓縮錯誤

### 摘要品質差

1. 嘗試不同的摘要模型（例如 claude-3-sonnet）
2. 增加 `preserve_recent_messages` 以保留更多上下文
3. 調整 `target_tokens` 以允許更長的摘要
4. 檢查摘要模型是否可用且正常運作

### 延遲增加

1. 壓縮會增加一次額外的 API 呼叫（摘要）
2. 使用更快的摘要模型（haiku 最快）
3. 增加閾值以減少壓縮頻率
4. 考慮對延遲敏感的應用程式停用壓縮

### 意外成本

1. 在使用儀表板中監控摘要成本
2. 比較節省 vs. 摘要成本
3. 調整閾值以減少壓縮頻率
4. 使用最便宜的可用模型進行摘要

## 效能影響

- **Token 估算** — 每個請求約 1ms（可忽略）
- **摘要產生** — 1-3 秒（取決於模型和訊息數量）
- **記憶體開銷** — 最小（每次壓縮約 1KB）
- **成本節省** — 通常減少 30-50% 的 token

## 進階配置

### 自訂摘要提示

```json
{
  "compression": {
    "enabled": true,
    "custom_prompt": "建立以下對話的技術摘要，重點關注程式碼變更、決策和行動項：\n\n{messages}\n\n摘要："
  }
}
```

### 條件壓縮

僅為特定場景啟用壓縮：

```json
{
  "profiles": {
    "default": {
      "scenarios": {
        "longContext": {
          "providers": ["anthropic"],
          "compression": {
            "enabled": true,
            "threshold_tokens": 100000
          }
        },
        "default": {
          "providers": ["anthropic"],
          "compression": {
            "enabled": false
          }
        }
      }
    }
  }
}
```

### 多階段壓縮

對於非常長的對話進行多次壓縮：

```json
{
  "compression": {
    "enabled": true,
    "stages": [
      {
        "threshold_tokens": 50000,
        "target_tokens": 30000
      },
      {
        "threshold_tokens": 80000,
        "target_tokens": 40000
      }
    ]
  }
}
```

## 未來增強

- 用於智慧訊息選擇的語義相似度配對
- 用於品質比較的多模型摘要
- 壓縮品質指標和回饋
- 針對每個用例的自訂壓縮策略
- 與 RAG 整合以進行外部上下文儲存
