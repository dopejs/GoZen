---
title: Agents
---

# Agents

GoZen 可以作為 Claude Code、Codex 與其他 CLI 助手的編碼 Agent 營運層。在不改變既有工作流程的前提下，它可以協助你協調 Agent 工作、監控會話，並施加執行期安全控制。

## GoZen 帶來的能力

- **協作協調**：減少多個 Agent 同時處理同一專案時的衝突。
- **可觀測性**：集中查看會話、成本、錯誤與活動。
- **安全護欄**：對支出、請求頻率與敏感操作施加限制。
- **任務路由**：把不同類型的工作分配給不同 provider 或 profile。

## 設定範例

```json
{
  "agent": {
    "enabled": true,
    "coordinator": {
      "enabled": true,
      "lock_timeout_sec": 300,
      "inject_warnings": true
    },
    "observatory": {
      "enabled": true,
      "stuck_threshold": 5,
      "idle_timeout_min": 30
    },
    "guardrails": {
      "enabled": true,
      "session_spending_cap": 5.0,
      "request_rate_limit": 30
    }
  }
}
```

## 常見工作流程

### 多 Agent 協調

當多個 Agent 在同一個儲存庫中工作時，GoZen 可以追蹤檔案活動、提示警告，並幫助避免衝突。

### 會話監控

你可以透過儀表板與 API 查看活躍會話、Token 使用量、錯誤次數與執行時長。

### 安全控制

護欄可以暫停失控會話、標記高風險操作，並在重試迴圈變得昂貴前先行抑制。

## 相關文件

- [Agent 基礎設施](/docs/agent-infrastructure) 更詳細介紹新的 runtime、observatory、coordinator 與 guardrails 架構。
- [Bot 閘道](/docs/bot) 說明如何從 Telegram、Slack、Discord 等聊天平台控制執行中的會話。
- [用量追蹤](/docs/usage-tracking) 與 [健康監控](/docs/health-monitoring) 介紹支撐 Agent 營運的指標。
